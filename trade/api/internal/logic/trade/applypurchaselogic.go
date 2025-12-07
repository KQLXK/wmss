// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"WMSS/trade/api/internal/svc"
	"WMSS/trade/api/internal/types"
	"WMSS/trade/api/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ApplyPurchaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApplyPurchaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyPurchaseLogic {
	return &ApplyPurchaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApplyPurchaseLogic) ApplyPurchase(req *types.PurchaseRequest) (resp *types.BaseResponse, err error) {
	// 1. 获取当前系统状态（这里需要根据你的系统状态表查询）
	//if err := l.checkSystemStatus(); err != nil {
	//	return &types.BaseResponse{
	//		Code:    1001,
	//		Message: "系统已停止交易，无法提交申购申请",
	//	}, nil
	//}

	// 2. 校验客户风险等级与产品风险等级是否匹配
	riskCheck, err := l.checkRiskMatch(req.CustomerId, req.ProductId)
	if err != nil {
		l.Errorf("风险校验失败: %v", err)
		return &types.BaseResponse{
			Code:    500,
			Message: "风险校验失败: " + err.Error(),
		}, nil
	}

	if !riskCheck.IsMatch && !req.RiskConfirmed {
		return &types.BaseResponse{
			Code:    1002,
			Message: "客户风险等级与产品风险等级不匹配，请客户确认后提交",
			Data:    riskCheck,
		}, nil
	}

	// 3. 校验银行卡余额是否充足
	balanceCheck, err := l.checkBalance(req.CardId, req.Amount)
	if err != nil {
		l.Errorf("余额校验失败: %v", err)
		return &types.BaseResponse{
			Code:    500,
			Message: "余额校验失败: " + err.Error(),
		}, nil
	}

	if !balanceCheck.IsSufficient {
		return &types.BaseResponse{
			Code:    1003,
			Message: "银行卡余额不足",
			Data:    balanceCheck,
		}, nil
	}

	// 4. 获取产品信息，计算申购费用
	product, err := l.getProductInfo(req.ProductId)
	if err != nil {
		l.Errorf("获取产品信息失败: %v", err)
		return &types.BaseResponse{
			Code:    500,
			Message: "获取产品信息失败: " + err.Error(),
		}, nil
	}

	if product.ProductStatus != "正常" {
		return &types.BaseResponse{
			Code:    1004,
			Message: "产品状态异常，无法申购",
		}, nil
	}

	// 5. 计算申购费用和净申购金额
	var purchaseFee float64
	if product.PurchaseFeeRate.Valid {
		purchaseFee = req.Amount * product.PurchaseFeeRate.Float64
	} else {
		// 如果费率为null，则默认费率为0
		purchaseFee = 0
	}
	netAmount := req.Amount - purchaseFee

	// 6. 生成申购申请编号
	applicationId, err := l.generateApplicationId()
	if err != nil {
		l.Errorf("生成申请编号失败: %v", err)
		return &types.BaseResponse{
			Code:    500,
			Message: "生成申请编号失败: " + err.Error(),
		}, nil
	}

	// 7. 获取预计确认日期（T+1）
	expectedDate := l.getNextWorkday()

	// 8. 使用事务保存申购申请并更新银行卡余额
	err = l.savePurchaseTransaction(req, applicationId, product, purchaseFee, netAmount, expectedDate)
	if err != nil {
		l.Errorf("保存申购申请失败: %v", err)
		return &types.BaseResponse{
			Code:    500,
			Message: "保存申购申请失败: " + err.Error(),
		}, nil
	}

	// 9. 返回申购申请结果
	response := &types.PurchaseResponse{
		ApplicationId:   applicationId,
		ApplicationDate: time.Now().Format("2006-01-02"),
		ApplicationTime: time.Now().Format("15:04:05"),
		ExpectedDate:    expectedDate.Format("2006-01-02"),
		Status:          "未确认",
		Amount:          req.Amount,
		PurchaseFee:     purchaseFee,
		NetAmount:       netAmount,
	}

	return &types.BaseResponse{
		Code:    200,
		Message: "申购申请提交成功",
		Data:    response,
	}, nil
}

// checkSystemStatus 检查系统状态
//func (l *ApplyPurchaseLogic) checkSystemStatus() error {
//	// 这里需要查询系统状态表，判断是否停止柜台交易
//	// 假设有一个系统配置表存储柜台状态
//	// 如果系统停止交易，返回错误
//
//	// 示例：直接返回nil表示系统正常
//	// 在实际应用中，应该从数据库查询系统状态
//	return nil
//}

// checkRiskMatch 校验客户风险等级与产品风险等级是否匹配
func (l *ApplyPurchaseLogic) checkRiskMatch(customerId, productId string) (*RiskCheckResponse, error) {
	// 获取客户风险等级
	customer, err := l.svcCtx.CustomerInfoModel.FindOne(l.ctx, customerId)
	if err != nil {
		log.Println("获取客户风险等级失败:", err)
		return nil, err
	}

	// 获取产品风险等级
	product, err := l.svcCtx.ProductInfoModel.FindOne(l.ctx, productId)
	if err != nil {
		log.Println("获取产品风险等级失败:", err)
		return nil, err
	}

	// 风险等级匹配规则
	// R1-R5 数字越大风险越高
	// 通常规则：客户风险等级 >= 产品风险等级 才允许购买
	isMatch := true
	riskMismatchRule := ""
	allowed := true
	message := ""

	// 简单的风险匹配规则
	if customer.RiskLevel < product.RiskLevel {
		isMatch = false
		allowed = false
		riskMismatchRule = "客户风险等级低于产品风险等级"
		message = "您的风险等级不足以购买此产品"
	}

	// 如果客户风险等级高于产品风险等级，允许购买但需要提示
	if customer.RiskLevel > product.RiskLevel {
		isMatch = false
		allowed = true
		riskMismatchRule = "客户风险等级高于产品风险等级"
		message = "您购买的产品风险低于您的风险承受能力"
	}

	return &RiskCheckResponse{
		CustomerRiskLevel: customer.RiskLevel,
		ProductRiskLevel:  product.RiskLevel,
		IsMatch:           isMatch,
		RiskMismatchRule:  riskMismatchRule,
		Allowed:           allowed,
		Message:           message,
	}, nil
}

// checkBalance 校验银行卡余额是否充足
func (l *ApplyPurchaseLogic) checkBalance(cardId int64, amount float64) (*BalanceCheckResponse, error) {
	// 获取银行卡信息
	card, err := l.svcCtx.CustomerBankCardModel.FindOne(l.ctx, cardId)
	if err != nil {
		return nil, err
	}

	if card.BindStatus != "正常" {
		return &BalanceCheckResponse{
			CardBalance:    card.CardBalance,
			RequiredAmount: amount,
			IsSufficient:   false,
			Message:        "银行卡状态异常",
		}, nil
	}

	isSufficient := card.CardBalance >= amount
	message := ""
	if !isSufficient {
		message = fmt.Sprintf("银行卡余额不足，当前余额: %.2f, 需要: %.2f", card.CardBalance, amount)
	}

	return &BalanceCheckResponse{
		CardBalance:    card.CardBalance,
		RequiredAmount: amount,
		IsSufficient:   isSufficient,
		Message:        message,
	}, nil
}

// getProductInfo 获取产品信息
func (l *ApplyPurchaseLogic) getProductInfo(productId string) (*model.ProductInfo, error) {
	product, err := l.svcCtx.ProductInfoModel.FindOne(l.ctx, productId)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// generateApplicationId 生成申购申请编号（格式: YYYYMMDD+5位序列号）
func (l *ApplyPurchaseLogic) generateApplicationId() (string, error) {
	// 获取当前日期
	date := time.Now().Format("20060102")

	// 使用原始的sqlx连接查询当日最大序列号
	query := "SELECT COALESCE(MAX(CAST(SUBSTRING(application_id, 9) AS UNSIGNED)), 0) FROM purchase_application WHERE application_id LIKE ?"
	var maxSeq int64

	// 使用SqlConn而不是PurchaseApplicationModel
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &maxSeq, query, date+"%")
	if err != nil && err != sqlx.ErrNotFound {
		return "", err
	}

	// 生成新序列号
	seq := maxSeq + 1
	return fmt.Sprintf("%s%05d", date, seq), nil
}

// getNextWorkday 获取下一个工作日
func (l *ApplyPurchaseLogic) getNextWorkday() time.Time {
	now := time.Now()

	// 简单实现：跳过周末
	nextDay := now.AddDate(0, 0, 1)
	for nextDay.Weekday() == time.Saturday || nextDay.Weekday() == time.Sunday {
		nextDay = nextDay.AddDate(0, 0, 1)
	}

	return nextDay
}

// savePurchaseTransaction 简化版本，不使用事务
func (l *ApplyPurchaseLogic) savePurchaseTransaction(req *types.PurchaseRequest, applicationId string, product *model.ProductInfo, purchaseFee, netAmount float64, expectedDate time.Time) error {
	// 1. 创建申购申请记录
	purchaseApp := &model.PurchaseApplication{
		ApplicationId:     applicationId,
		CustomerId:        req.CustomerId,
		ProductId:         req.ProductId,
		CardId:            req.CardId,
		ApplicationAmount: req.Amount,
		PurchaseFee: sql.NullFloat64{ // 转换为 sql.NullFloat64
			Float64: purchaseFee,
			Valid:   true, // 如果 purchaseFee 为 0，但仍然是有效值
		},
		NetPurchaseAmount: sql.NullFloat64{ // 转换为 sql.NullFloat64
			Float64: netAmount,
			Valid:   true, // 如果 netAmount 为 0，但仍然是有效值
		},
		ApplicationDate:          time.Now(),
		ApplicationTime:          time.Now(),
		ExpectedConfirmationDate: expectedDate,
		ApplicationStatus:        "未确认",
		RiskMismatchRemark:       sql.NullString{}, // 如果有风险不匹配且客户确认，可以记录
		OperatorId:               req.OperatorId,
		CreateTime:               time.Now(),
		UpdateTime:               time.Now(),
	}

	// 2. 插入申购申请
	_, err := l.svcCtx.PurchaseApplicationModel.Insert(l.ctx, purchaseApp)
	if err != nil {
		return fmt.Errorf("插入申购申请失败: %w", err)
	}

	// 3. 更新银行卡余额
	card, err := l.svcCtx.CustomerBankCardModel.FindOne(l.ctx, req.CardId)
	if err != nil {
		return fmt.Errorf("查询银行卡失败: %w", err)
	}

	if card.CardBalance < req.Amount {
		return fmt.Errorf("银行卡余额不足")
	}

	// 直接执行 SQL 更新余额
	updateQuery := "UPDATE customer_bank_card SET card_balance = card_balance - ?, update_time = NOW() WHERE card_id = ?"
	_, err = l.svcCtx.Conn.ExecCtx(l.ctx, updateQuery, req.Amount, req.CardId)
	if err != nil {
		return fmt.Errorf("更新银行卡余额失败: %w", err)
	}

	return nil
}

// RiskCheckResponse 风险校验响应
type RiskCheckResponse struct {
	CustomerRiskLevel string `json:"customerRiskLevel"`
	ProductRiskLevel  string `json:"productRiskLevel"`
	IsMatch           bool   `json:"isMatch"`
	RiskMismatchRule  string `json:"riskMismatchRule,optional"`
	Allowed           bool   `json:"allowed"`
	Message           string `json:"message,optional"`
}

// BalanceCheckResponse 余额校验响应
type BalanceCheckResponse struct {
	CardBalance    float64 `json:"cardBalance"`
	RequiredAmount float64 `json:"requiredAmount"`
	IsSufficient   bool    `json:"isSufficient"`
	Message        string  `json:"message,optional"`
}
