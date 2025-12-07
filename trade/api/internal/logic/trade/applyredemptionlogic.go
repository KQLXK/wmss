// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"WMSS/trade/api/internal/svc"
	"WMSS/trade/api/internal/types"
	"WMSS/trade/api/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ApplyRedemptionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApplyRedemptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyRedemptionLogic {
	return &ApplyRedemptionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApplyRedemptionLogic) ApplyRedemption(req *types.RedemptionRequest) (resp *types.BaseResponse, err error) {
	// 记录请求参数
	l.Infof("收到赎回请求: customerId=%s, productId=%s, cardId=%d, shares=%.4f, operatorId=%s",
		req.CustomerId, req.ProductId, req.CardId, req.Shares, req.OperatorId)

	// 1. 检查系统状态（如果实现）
	// if err := l.checkSystemStatus(); err != nil {
	//     return l.errorResponse(1001, "系统已停止交易，无法提交赎回申请", nil), nil
	// }

	// 2. 校验客户信息
	customer, err := l.checkCustomer(req.CustomerId)
	if err != nil {
		return l.errorResponse(500, "客户信息校验失败: "+err.Error(), nil), nil
	}

	// 3. 校验产品信息
	product, err := l.checkProduct(req.ProductId)
	if err != nil {
		return l.errorResponse(500, "产品信息校验失败: "+err.Error(), nil), nil
	}

	// 4. 校验银行卡和持仓信息（同卡进出原则）
	position, err := l.checkPositionAndCard(req.CustomerId, req.ProductId, req.CardId, req.Shares)
	if err != nil {
		if err.Error() == "持仓不足" {
			return l.errorResponse(1005, "赎回份额超过可用份额", nil), nil
		}
		if err.Error() == "银行卡不存在持仓" {
			return l.errorResponse(1006, "该银行卡未持有此产品，请遵循同卡进出原则", nil), nil
		}
		return l.errorResponse(500, "持仓校验失败: "+err.Error(), nil), nil
	}

	// 5. 获取产品净值
	netValue, err := l.getLatestNetValue(req.ProductId)
	if err != nil {
		return l.errorResponse(500, "获取产品净值失败: "+err.Error(), nil), nil
	}

	// 6. 计算持有期限（天）
	holdingPeriod, err := l.calculateHoldingPeriod(req.CustomerId, req.ProductId, req.CardId)
	if err != nil {
		// 如果计算失败，使用默认值0
		//l.("计算持有期限失败，使用默认值0: %v", err)
		holdingPeriod = 0
	}

	// 7. 计算赎回费用
	redemptionFee, feeRate := l.calculateRedemptionFee(req.Shares, netValue, holdingPeriod, product.RedemptionFeeRule.String)

	// 8. 计算预计赎回金额
	expectedAmount := req.Shares*netValue - redemptionFee

	// 9. 生成赎回申请编号
	applicationId, err := l.generateRedemptionId()
	if err != nil {
		l.Errorf("生成赎回申请编号失败: %v", err)
		return l.errorResponse(500, "生成赎回申请编号失败: "+err.Error(), nil), nil
	}

	// 10. 获取预计确认日期（T+1）
	expectedDate := l.getNextWorkday()

	// 11. 使用事务保存赎回申请并更新持仓
	err = l.saveRedemptionTransaction(req, applicationId, product, position, holdingPeriod, redemptionFee, expectedAmount, netValue, expectedDate)
	if err != nil {
		l.Errorf("保存赎回申请失败: %v", err)
		return l.errorResponse(500, "保存赎回申请失败: "+err.Error(), nil), nil
	}

	// 12. 返回赎回申请结果
	response := &types.RedemptionResponse{
		ApplicationId:   applicationId,
		ApplicationDate: time.Now().Format("2006-01-02"),
		ApplicationTime: time.Now().Format("15:04:05"),
		ExpectedDate:    expectedDate.Format("2006-01-02"),
		Status:          "未确认",
		Shares:          req.Shares,
		ExpectedAmount:  expectedAmount,
		RedemptionFee:   redemptionFee,
	}

	// 添加额外信息
	data := map[string]interface{}{
		"redemption": response,
		"details": map[string]interface{}{
			"netValue":      netValue,
			"holdingPeriod": holdingPeriod,
			"feeRate":       feeRate,
			"positionInfo": map[string]interface{}{
				"totalShares":     position.TotalShares,
				"availableShares": position.AvailableShares,
				"frozenShares":    position.FrozenShares,
				"averageCost":     position.AverageCost.Float64,
			},
			"customerInfo": map[string]interface{}{
				"customerName": customer.CustomerName,
				"riskLevel":    customer.RiskLevel,
			},
			"productInfo": map[string]interface{}{
				"productName": product.ProductName,
				"productType": product.ProductType,
				"riskLevel":   product.RiskLevel,
			},
		},
	}

	return l.successResponse("赎回申请提交成功", data), nil
}

// checkCustomer 校验客户信息
func (l *ApplyRedemptionLogic) checkCustomer(customerId string) (*model.CustomerInfo, error) {
	customer, err := l.svcCtx.CustomerInfoModel.FindOne(l.ctx, customerId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, fmt.Errorf("客户不存在: %s", customerId)
		}
		return nil, fmt.Errorf("查询客户信息失败: %w", err)
	}
	return customer, nil
}

// checkProduct 校验产品信息
func (l *ApplyRedemptionLogic) checkProduct(productId string) (*model.ProductInfo, error) {
	product, err := l.svcCtx.ProductInfoModel.FindOne(l.ctx, productId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, fmt.Errorf("产品不存在: %s", productId)
		}
		return nil, fmt.Errorf("查询产品信息失败: %w", err)
	}

	if product.ProductStatus != "正常" && product.ProductStatus != "暂停申购" {
		return nil, fmt.Errorf("产品状态异常，无法赎回: %s", product.ProductStatus)
	}

	return product, nil
}

// checkPositionAndCard 校验持仓和银行卡（同卡进出原则）
func (l *ApplyRedemptionLogic) checkPositionAndCard(customerId, productId string, cardId int64, shares float64) (*model.CustomerPosition, error) {
	// 查询该客户、产品、银行卡的持仓信息
	query := `
		SELECT * FROM customer_position 
		WHERE customer_id = ? AND product_id = ? AND card_id = ?
		AND position_date = CURDATE()
	`
	var position model.CustomerPosition
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &position, query, customerId, productId, cardId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, fmt.Errorf("银行卡不存在持仓")
		}
		return nil, fmt.Errorf("查询持仓失败: %w", err)
	}

	// 校验可用份额是否足够
	if position.AvailableShares < shares {
		return nil, fmt.Errorf("持仓不足")
	}

	return &position, nil
}

// getLatestNetValue 获取最新净值
func (l *ApplyRedemptionLogic) getLatestNetValue(productId string) (float64, error) {
	query := `
		SELECT unit_net_value FROM product_net_value 
		WHERE product_id = ? 
		ORDER BY stat_date DESC 
		LIMIT 1
	`
	var netValue float64
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &netValue, query, productId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return 0, fmt.Errorf("产品净值未找到")
		}
		return 0, fmt.Errorf("查询净值失败: %w", err)
	}
	return netValue, nil
}

// calculateHoldingPeriod 计算持有期限（天）
func (l *ApplyRedemptionLogic) calculateHoldingPeriod(customerId, productId string, cardId int64) (int, error) {
	// 查询该持仓的最早购买确认日期
	query := `
		SELECT MIN(confirmation_date) as earliest_date 
		FROM transaction_confirmation tc
		JOIN purchase_application pa ON tc.application_id = pa.application_id
		WHERE tc.customer_id = ? 
		AND tc.product_id = ? 
		AND pa.card_id = ?
		AND tc.transaction_type = '申购'
		AND tc.confirmation_status = '确认成功'
	`
	var earliestDate sql.NullTime
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &earliestDate, query, customerId, productId, cardId)
	if err != nil {
		return 0, fmt.Errorf("查询持有期限失败: %w", err)
	}

	if !earliestDate.Valid {
		// 如果没有购买记录，可能是初始持仓，使用持仓创建时间
		query = `
			SELECT create_time 
			FROM customer_position 
			WHERE customer_id = ? AND product_id = ? AND card_id = ?
			ORDER BY create_time ASC 
			LIMIT 1
		`
		var createTime time.Time
		err = l.svcCtx.Conn.QueryRowCtx(l.ctx, &createTime, query, customerId, productId, cardId)
		if err != nil {
			return 0, err
		}
		earliestDate = sql.NullTime{Time: createTime, Valid: true}
	}

	// 计算持有天数
	holdingDays := int(time.Since(earliestDate.Time).Hours() / 24)
	if holdingDays < 0 {
		holdingDays = 0
	}

	return holdingDays, nil
}

// calculateRedemptionFee 计算赎回费用
func (l *ApplyRedemptionLogic) calculateRedemptionFee(shares, netValue float64, holdingDays int, feeRuleJson string) (fee float64, feeRate float64) {
	// 默认赎回费率为0
	feeRate = 0.0

	// 如果产品有赎回费率规则，则解析
	if feeRuleJson != "" {
		var feeRule map[string]float64
		if err := json.Unmarshal([]byte(feeRuleJson), &feeRule); err == nil {
			// 根据持有天数匹配费率
			for period, rate := range feeRule {
				// 简单解析持有期限，如 "0-7": 0.015
				var minDays, maxDays int
				if _, err := fmt.Sscanf(period, "%d-%d", &minDays, &maxDays); err == nil {
					if holdingDays >= minDays && holdingDays <= maxDays {
						feeRate = rate
						break
					}
				} else {
					// 尝试解析单个天数，如 "7": 0.0075
					var days int
					if _, err := fmt.Sscanf(period, "%d", &days); err == nil && holdingDays >= days {
						feeRate = rate
						break
					}
				}
			}
		}
	}

	// 计算赎回费用
	fee = shares * netValue * feeRate
	return fee, feeRate
}

// generateRedemptionId 生成赎回申请编号
func (l *ApplyRedemptionLogic) generateRedemptionId() (string, error) {
	// 获取当前日期
	date := time.Now().Format("20060102")

	// 查询当日最大序列号
	query := "SELECT COALESCE(MAX(CAST(SUBSTRING(application_id, 9) AS UNSIGNED)), 0) FROM redemption_application WHERE application_id LIKE ?"
	var maxSeq int64

	// 使用SqlConn查询
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &maxSeq, query, date+"%")
	if err != nil && err != sqlx.ErrNotFound {
		return "", err
	}

	// 生成新序列号
	seq := maxSeq + 1
	return fmt.Sprintf("%s%05d", date, seq), nil
}

// getNextWorkday 获取下一个工作日
func (l *ApplyRedemptionLogic) getNextWorkday() time.Time {
	now := time.Now()

	// 简单实现：跳过周末
	nextDay := now.AddDate(0, 0, 1)
	for nextDay.Weekday() == time.Saturday || nextDay.Weekday() == time.Sunday {
		nextDay = nextDay.AddDate(0, 0, 1)
	}

	return nextDay
}

// saveRedemptionTransaction 保存赎回申请（简化版本，不使用事务）
func (l *ApplyRedemptionLogic) saveRedemptionTransaction(req *types.RedemptionRequest, applicationId string, product *model.ProductInfo, position *model.CustomerPosition, holdingPeriod int, redemptionFee, expectedAmount, netValue float64, expectedDate time.Time) error {
	now := time.Now()

	// 1. 插入赎回申请记录
	insertQuery := `
		INSERT INTO redemption_application 
		(application_id, customer_id, product_id, card_id, application_shares, 
		 holding_period, redemption_fee, expected_redemption_amount, 
		 application_date, application_time, expected_confirmation_date, 
		 application_status, operator_id, create_time, update_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := l.svcCtx.Conn.ExecCtx(l.ctx, insertQuery,
		applicationId,
		req.CustomerId,
		req.ProductId,
		req.CardId,
		req.Shares,
		holdingPeriod,
		redemptionFee,
		expectedAmount,
		now.Format("2006-01-02"), // application_date: 日期部分
		now,                      // application_time: 完整的日期时间
		expectedDate.Format("2006-01-02"),
		"未确认",
		req.OperatorId,
		now, // create_time: 完整的日期时间
		now, // update_time: 完整的日期时间
	)
	if err != nil {
		return fmt.Errorf("插入赎回申请失败: %w", err)
	}

	// 2. 更新客户持仓，冻结赎回份额
	updateQuery := `
		UPDATE customer_position 
		SET available_shares = available_shares - ?, 
			frozen_shares = frozen_shares + ?,
			update_time = NOW()
		WHERE customer_id = ? AND product_id = ? AND card_id = ?
		AND position_date = CURDATE()
	`
	result, err := l.svcCtx.Conn.ExecCtx(l.ctx, updateQuery,
		req.Shares,
		req.Shares,
		req.CustomerId,
		req.ProductId,
		req.CardId,
	)
	if err != nil {
		// 如果更新失败，尝试删除插入的赎回申请
		deleteQuery := "DELETE FROM redemption_application WHERE application_id = ?"
		l.svcCtx.Conn.ExecCtx(l.ctx, deleteQuery, applicationId)
		return fmt.Errorf("更新持仓失败: %w", err)
	}

	// 检查是否成功更新了记录
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// 没有更新到任何记录，删除赎回申请
		deleteQuery := "DELETE FROM redemption_application WHERE application_id = ?"
		l.svcCtx.Conn.ExecCtx(l.ctx, deleteQuery, applicationId)
		return fmt.Errorf("未找到对应的持仓记录，可能持仓信息不存在或已更新")
	}

	return nil
}

// successResponse 成功响应
func (l *ApplyRedemptionLogic) successResponse(message string, data interface{}) *types.BaseResponse {
	return &types.BaseResponse{
		Code:    200,
		Message: message,
		Data:    data,
	}
}

// errorResponse 错误响应
func (l *ApplyRedemptionLogic) errorResponse(code int64, message string, data interface{}) *types.BaseResponse {
	return &types.BaseResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
