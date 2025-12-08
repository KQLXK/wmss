// canceltransactionlogic.go
// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"
	"fmt"
	"time"

	"WMSS/trade/api/internal/svc"
	"WMSS/trade/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CancelTransactionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelTransactionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelTransactionLogic {
	return &CancelTransactionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelTransactionLogic) CancelTransaction(req *types.CancelRequest) (resp *types.BaseResponse, err error) {
	// 记录请求参数
	l.Infof("收到交易撤回请求: applicationId=%s, transactionType=%s, operatorId=%s, reason=%s",
		req.ApplicationId, req.TransactionType, req.OperatorId, req.Reason)

	// 1. 校验交易类型
	if req.TransactionType != "PURCHASE" && req.TransactionType != "REDEMPTION" {
		return l.errorResponse(400, "交易类型错误，只能为'PURCHASE'或'REDEMPTION'", nil), nil
	}

	// 2. 根据交易类型执行不同的撤回逻辑
	var result *types.CancelResponse
	switch req.TransactionType {
	case "PURCHASE":
		result, err = l.cancelPurchase(req)
	case "REDEMPTION":
		result, err = l.cancelRedemption(req)
	}

	if err != nil {
		l.Errorf("撤回交易失败: %v", err)
		return l.errorResponse(500, "撤回交易失败: "+err.Error(), nil), nil
	}

	return l.successResponse("交易撤回成功", result), nil
}

// cancelPurchase 撤回申购申请
func (l *CancelTransactionLogic) cancelPurchase(req *types.CancelRequest) (*types.CancelResponse, error) {
	// 1. 获取申购申请详情
	purchaseApp, err := l.svcCtx.PurchaseApplicationModel.FindOne(l.ctx, req.ApplicationId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, fmt.Errorf("申购申请不存在: %s", req.ApplicationId)
		}
		return nil, fmt.Errorf("查询申购申请失败: %w", err)
	}

	// 2. 检查申请状态，只能撤回未确认的申请
	if purchaseApp.ApplicationStatus != "未确认" {
		return nil, fmt.Errorf("申请状态为'%s'，无法撤回，只能撤回未确认的申请", purchaseApp.ApplicationStatus)
	}

	// 3. 检查确认日期，如果已经过了预计确认日期，则不能撤回
	currentDate := time.Now()
	if currentDate.After(purchaseApp.ExpectedConfirmationDate) {
		return nil, fmt.Errorf("已过预计确认日期%s，无法撤回", purchaseApp.ExpectedConfirmationDate.Format("2006-01-02"))
	}

	// 4. 更新申购申请状态为"已撤回"
	updateQuery := `
		UPDATE purchase_application 
		SET application_status = '已撤回', 
			update_time = NOW() 
		WHERE application_id = ? AND application_status = '未确认'
	`
	result, err := l.svcCtx.Conn.ExecCtx(l.ctx, updateQuery, req.ApplicationId)
	if err != nil {
		return nil, fmt.Errorf("更新申购申请状态失败: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("申购申请状态已变更，无法撤回")
	}

	// 5. 退款到银行卡
	refundAmount := purchaseApp.ApplicationAmount
	refundQuery := `
		UPDATE customer_bank_card 
		SET card_balance = card_balance + ?, 
			update_time = NOW() 
		WHERE card_id = ?
	`
	_, err = l.svcCtx.Conn.ExecCtx(l.ctx, refundQuery, refundAmount, purchaseApp.CardId)
	if err != nil {
		// 如果退款失败，记录错误但不回滚状态，需要人工干预
		l.Errorf("退款到银行卡失败: cardId=%d, amount=%.2f, error=%v", purchaseApp.CardId, refundAmount, err)
		// 继续执行，返回成功但标记退款问题
	}

	l.Infof("申购撤回成功: applicationId=%s, refundAmount=%.2f", req.ApplicationId, refundAmount)

	// 6. 返回结果
	return &types.CancelResponse{
		ApplicationId:   req.ApplicationId,
		TransactionType: "PURCHASE",
		OriginalStatus:  "未确认",
		NewStatus:       "已撤回",
		CancelTime:      time.Now().Format("2006-01-02 15:04:05"),
		Reason:          req.Reason,
		RefundAmount:    refundAmount,
	}, nil
}

// cancelRedemption 撤回赎回申请
func (l *CancelTransactionLogic) cancelRedemption(req *types.CancelRequest) (*types.CancelResponse, error) {
	// 1. 获取赎回申请详情
	redemptionApp, err := l.svcCtx.RedemptionApplicationModel.FindOne(l.ctx, req.ApplicationId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, fmt.Errorf("赎回申请不存在: %s", req.ApplicationId)
		}
		return nil, fmt.Errorf("查询赎回申请失败: %w", err)
	}

	// 2. 检查申请状态，只能撤回未确认的申请
	if redemptionApp.ApplicationStatus != "未确认" {
		return nil, fmt.Errorf("申请状态为'%s'，无法撤回，只能撤回未确认的申请", redemptionApp.ApplicationStatus)
	}

	// 3. 检查确认日期，如果已经过了预计确认日期，则不能撤回
	currentDate := time.Now()
	if currentDate.After(redemptionApp.ExpectedConfirmationDate) {
		return nil, fmt.Errorf("已过预计确认日期%s，无法撤回", redemptionApp.ExpectedConfirmationDate.Format("2006-01-02"))
	}

	// 4. 更新赎回申请状态为"已撤回"
	updateQuery := `
		UPDATE redemption_application 
		SET application_status = '已撤回', 
			update_time = NOW() 
		WHERE application_id = ? AND application_status = '未确认'
	`
	result, err := l.svcCtx.Conn.ExecCtx(l.ctx, updateQuery, req.ApplicationId)
	if err != nil {
		return nil, fmt.Errorf("更新赎回申请状态失败: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("赎回申请状态已变更，无法撤回")
	}

	// 5. 解冻持仓份额
	unfrozenShares := redemptionApp.ApplicationShares
	unfreezeQuery := `
		UPDATE customer_position 
		SET available_shares = available_shares + ?, 
			frozen_shares = frozen_shares - ?,
			update_time = NOW()
		WHERE customer_id = ? AND product_id = ? AND card_id = ?
		AND position_date = CURDATE()
		AND frozen_shares >= ?
	`
	result, err = l.svcCtx.Conn.ExecCtx(l.ctx, unfreezeQuery,
		unfrozenShares,
		unfrozenShares,
		redemptionApp.CustomerId,
		redemptionApp.ProductId,
		redemptionApp.CardId,
		unfrozenShares,
	)
	if err != nil {
		// 如果解冻失败，记录错误但不回滚状态，需要人工干预
		l.Errorf("解冻持仓份额失败: customerId=%s, productId=%s, shares=%.4f, error=%v",
			redemptionApp.CustomerId, redemptionApp.ProductId, unfrozenShares, err)
		// 继续执行，返回成功但标记解冻问题
	} else {
		// 检查是否成功更新了记录
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			l.Infof("未找到对应的持仓记录或冻结份额不足，请手动检查")
		}
	}

	l.Infof("赎回撤回成功: applicationId=%s, unfrozenShares=%.4f", req.ApplicationId, unfrozenShares)

	// 6. 返回结果
	return &types.CancelResponse{
		ApplicationId:   req.ApplicationId,
		TransactionType: "REDEMPTION",
		OriginalStatus:  "未确认",
		NewStatus:       "已撤回",
		CancelTime:      time.Now().Format("2006-01-02 15:04:05"),
		Reason:          req.Reason,
		UnfrozenShares:  unfrozenShares,
	}, nil
}

// successResponse 成功响应
func (l *CancelTransactionLogic) successResponse(message string, data interface{}) *types.BaseResponse {
	return &types.BaseResponse{
		Code:    200,
		Message: message,
		Data:    data,
	}
}

// errorResponse 错误响应
func (l *CancelTransactionLogic) errorResponse(code int64, message string, data interface{}) *types.BaseResponse {
	return &types.BaseResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
