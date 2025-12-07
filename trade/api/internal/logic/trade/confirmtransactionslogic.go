// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"
	"fmt"
	"time"

	"WMSS/trade/api/internal/svc"
	"WMSS/trade/api/internal/types"
	"WMSS/trade/api/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ConfirmTransactionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmTransactionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmTransactionsLogic {
	return &ConfirmTransactionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmTransactionsLogic) ConfirmTransactions(req *types.ConfirmRequest) (resp *types.BaseResponse, err error) {
	// 记录请求参数
	l.Infof("收到交易确认请求: confirmationDate=%s, operatorId=%s, batchSize=%d",
		req.ConfirmationDate, req.OperatorId, req.BatchSize)

	// 1. 解析确认日期
	confirmationDate, err := time.Parse("2006-01-02", req.ConfirmationDate)
	if err != nil {
		return l.errorResponse(400, "日期格式错误，请使用YYYY-MM-DD格式", nil), nil
	}

	// 2. 获取需要确认的申购申请
	purchaseApplications, err := l.getPurchaseApplicationsToConfirm(confirmationDate, req.ApplicationIds)
	if err != nil {
		l.Errorf("获取申购申请失败: %v", err)
		return l.errorResponse(500, "获取申购申请失败: "+err.Error(), nil), nil
	}

	// 3. 获取需要确认的赎回申请
	redemptionApplications, err := l.getRedemptionApplicationsToConfirm(confirmationDate, req.ApplicationIds)
	if err != nil {
		l.Errorf("获取赎回申请失败: %v", err)
		return l.errorResponse(500, "获取赎回申请失败: "+err.Error(), nil), nil
	}

	// 4. 处理申购确认
	purchaseResults := l.confirmPurchaseApplications(purchaseApplications, confirmationDate, req.OperatorId)

	// 5. 处理赎回确认
	redemptionResults := l.confirmRedemptionApplications(redemptionApplications, confirmationDate, req.OperatorId)

	// 6. 合并结果
	totalProcessed := len(purchaseApplications) + len(redemptionApplications)
	successCount := purchaseResults.SuccessCount + redemptionResults.SuccessCount
	failedCount := len(purchaseResults.FailedApplications) + len(redemptionResults.FailedApplications)

	allConfirmationIds := append(purchaseResults.ConfirmationIds, redemptionResults.ConfirmationIds...)
	allFailedApplications := append(purchaseResults.FailedApplications, redemptionResults.FailedApplications...)
	totalFees := purchaseResults.TotalFees + redemptionResults.TotalFees

	// 7. 返回结果
	response := &types.ConfirmResponse{
		ConfirmationDate:   req.ConfirmationDate,
		TotalProcessed:     int64(totalProcessed),
		SuccessCount:       int64(successCount),
		FailedCount:        int64(failedCount),
		ConfirmationIds:    allConfirmationIds,
		FailedApplications: allFailedApplications,
		TotalFees:          totalFees,
	}

	return l.successResponse("交易确认处理完成", response), nil
}

// getPurchaseApplicationsToConfirm 获取需要确认的申购申请
func (l *ConfirmTransactionsLogic) getPurchaseApplicationsToConfirm(confirmationDate time.Time, specifiedIds []string) ([]model.PurchaseApplication, error) {
	// 计算申请日期（T日）
	//applicationDate := confirmationDate.AddDate(0, 0, -1)

	query := `
		SELECT * FROM purchase_application 
		WHERE application_status = '未确认' 
		AND expected_confirmation_date = ? 
	`

	var args []interface{}
	args = append(args, confirmationDate.Format("2006-01-02"))

	// 如果指定了申请ID，则只处理指定的申请
	if len(specifiedIds) > 0 {
		query += " AND application_id IN ("
		for i, id := range specifiedIds {
			if i > 0 {
				query += ","
			}
			query += "?"
			args = append(args, id)
		}
		query += ")"
	}

	query += " ORDER BY application_date, application_time LIMIT 1000"

	var applications []model.PurchaseApplication
	err := l.svcCtx.Conn.QueryRowsCtx(l.ctx, &applications, query, args...)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, fmt.Errorf("查询申购申请失败: %w", err)
	}

	l.Infof("获取到 %d 个待确认的申购申请", len(applications))
	return applications, nil
}

// getRedemptionApplicationsToConfirm 获取需要确认的赎回申请
func (l *ConfirmTransactionsLogic) getRedemptionApplicationsToConfirm(confirmationDate time.Time, specifiedIds []string) ([]model.RedemptionApplication, error) {
	// 计算申请日期（T日）
	//applicationDate := confirmationDate.AddDate(0, 0, -1)

	query := `
		SELECT * FROM redemption_application 
		WHERE application_status = '未确认' 
		AND expected_confirmation_date = ? 
	`

	var args []interface{}
	args = append(args, confirmationDate.Format("2006-01-02"))

	// 如果指定了申请ID，则只处理指定的申请
	if len(specifiedIds) > 0 {
		query += " AND application_id IN ("
		for i, id := range specifiedIds {
			if i > 0 {
				query += ","
			}
			query += "?"
			args = append(args, id)
		}
		query += ")"
	}

	query += " ORDER BY application_date, application_time LIMIT 1000"

	var applications []model.RedemptionApplication
	err := l.svcCtx.Conn.QueryRowsCtx(l.ctx, &applications, query, args...)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, fmt.Errorf("查询赎回申请失败: %w", err)
	}

	l.Infof("获取到 %d 个待确认的赎回申请", len(applications))
	return applications, nil
}

// confirmPurchaseApplications 确认申购申请
func (l *ConfirmTransactionsLogic) confirmPurchaseApplications(applications []model.PurchaseApplication, confirmationDate time.Time, operatorId string) ConfirmResult {
	var result ConfirmResult
	result.FailedApplications = []types.FailedApplication{}
	result.ConfirmationIds = []string{}

	for _, app := range applications {
		confirmationId, err := l.confirmSinglePurchase(app, confirmationDate, operatorId)
		if err != nil {
			l.Errorf("申购确认失败: applicationId=%s, error=%v", app.ApplicationId, err)
			result.FailedApplications = append(result.FailedApplications, types.FailedApplication{
				ApplicationId: app.ApplicationId,
				ErrorReason:   err.Error(),
				CustomerId:    app.CustomerId,
				ProductId:     app.ProductId,
			})
		} else {
			result.SuccessCount++
			result.ConfirmationIds = append(result.ConfirmationIds, confirmationId)
			// 累加费用
			if app.PurchaseFee.Valid {
				result.TotalFees += app.PurchaseFee.Float64
			}
		}
	}

	return result
}

// confirmSinglePurchase 确认单个申购申请
func (l *ConfirmTransactionsLogic) confirmSinglePurchase(app model.PurchaseApplication, confirmationDate time.Time, operatorId string) (string, error) {
	// 1. 获取T日净值
	netValue, err := l.getNetValueForDate(app.ProductId, confirmationDate.AddDate(0, 0, -1))
	if err != nil {
		return "", fmt.Errorf("获取产品净值失败: %w", err)
	}

	// 2. 检查净申购金额是否有有效值
	if !app.NetPurchaseAmount.Valid {
		return "", fmt.Errorf("净申购金额无效")
	}

	// 3. 计算确认份额
	var confirmedShares float64
	if netValue > 0 {
		confirmedShares = app.NetPurchaseAmount.Float64 / netValue
	} else {
		return "", fmt.Errorf("产品净值为0或负数")
	}

	// 4. 生成确认编号
	confirmationId, err := l.generateConfirmationId(confirmationDate)
	if err != nil {
		return "", fmt.Errorf("生成确认编号失败: %w", err)
	}

	// 5. 获取申购费用（如果有）
	var purchaseFee float64
	if app.PurchaseFee.Valid {
		purchaseFee = app.PurchaseFee.Float64
	}

	// 6. 创建确认记录
	now := time.Now()
	insertQuery := `
		INSERT INTO transaction_confirmation 
		(confirmation_id, application_id, transaction_type, customer_id, product_id,
		 confirmation_date, confirmed_shares, confirmed_amount, net_value, fee,
		 confirmation_status, failure_reason, create_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = l.svcCtx.Conn.ExecCtx(l.ctx, insertQuery,
		confirmationId,
		app.ApplicationId,
		"申购",
		app.CustomerId,
		app.ProductId,
		confirmationDate.Format("2006-01-02"),
		confirmedShares,
		app.NetPurchaseAmount.Float64,
		netValue,
		purchaseFee,
		"确认成功",
		nil,
		now,
	)
	if err != nil {
		return "", fmt.Errorf("创建确认记录失败: %w", err)
	}

	// 7. 更新客户持仓
	err = l.updatePositionForPurchase(app, confirmedShares, confirmationDate, netValue)
	if err != nil {
		// 如果更新持仓失败，删除确认记录
		l.deleteConfirmationRecord(confirmationId)
		return "", fmt.Errorf("更新持仓失败: %w", err)
	}

	// 8. 更新申购申请状态
	err = l.updateApplicationStatus(app.ApplicationId, "申购", "已确认")
	if err != nil {
		// 如果更新状态失败，标记为部分成功（确认记录已创建，但状态未更新）
		l.Errorf("更新申购申请状态失败: %s, 确认记录已创建", app.ApplicationId)
	}

	l.Infof("申购确认成功: applicationId=%s, confirmationId=%s, shares=%.4f, netValue=%.4f",
		app.ApplicationId, confirmationId, confirmedShares, netValue)

	return confirmationId, nil
}

// confirmRedemptionApplications 确认赎回申请
func (l *ConfirmTransactionsLogic) confirmRedemptionApplications(applications []model.RedemptionApplication, confirmationDate time.Time, operatorId string) ConfirmResult {
	var result ConfirmResult
	result.FailedApplications = []types.FailedApplication{}
	result.ConfirmationIds = []string{}

	for _, app := range applications {
		confirmationId, err := l.confirmSingleRedemption(app, confirmationDate, operatorId)
		if err != nil {
			l.Errorf("赎回确认失败: applicationId=%s, error=%v", app.ApplicationId, err)
			result.FailedApplications = append(result.FailedApplications, types.FailedApplication{
				ApplicationId: app.ApplicationId,
				ErrorReason:   err.Error(),
				CustomerId:    app.CustomerId,
				ProductId:     app.ProductId,
			})
		} else {
			result.SuccessCount++
			result.ConfirmationIds = append(result.ConfirmationIds, confirmationId)
			// 累加费用
			if app.RedemptionFee.Valid {
				result.TotalFees += app.RedemptionFee.Float64
			}
		}
	}

	return result
}

// confirmSingleRedemption 确认单个赎回申请
func (l *ConfirmTransactionsLogic) confirmSingleRedemption(app model.RedemptionApplication, confirmationDate time.Time, operatorId string) (string, error) {
	// 1. 获取T日净值
	netValue, err := l.getNetValueForDate(app.ProductId, confirmationDate.AddDate(0, 0, -1))
	if err != nil {
		return "", fmt.Errorf("获取产品净值失败: %w", err)
	}

	// 2. 计算确认金额
	confirmedAmount := app.ApplicationShares*netValue - app.RedemptionFee.Float64
	if confirmedAmount < 0 {
		confirmedAmount = 0
	}

	// 3. 生成确认编号
	confirmationId, err := l.generateConfirmationId(confirmationDate)
	if err != nil {
		return "", fmt.Errorf("生成确认编号失败: %w", err)
	}

	// 4. 创建确认记录
	now := time.Now()
	insertQuery := `
		INSERT INTO transaction_confirmation 
		(confirmation_id, application_id, transaction_type, customer_id, product_id,
		 confirmation_date, confirmed_shares, confirmed_amount, net_value, fee,
		 confirmation_status, failure_reason, create_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = l.svcCtx.Conn.ExecCtx(l.ctx, insertQuery,
		confirmationId,
		app.ApplicationId,
		"赎回",
		app.CustomerId,
		app.ProductId,
		confirmationDate.Format("2006-01-02"),
		app.ApplicationShares,
		confirmedAmount,
		netValue,
		app.RedemptionFee.Float64,
		"确认成功",
		nil,
		now,
	)
	if err != nil {
		return "", fmt.Errorf("创建确认记录失败: %w", err)
	}

	// 5. 更新客户持仓
	err = l.updatePositionForRedemption(app, confirmationDate)
	if err != nil {
		// 如果更新持仓失败，删除确认记录
		l.deleteConfirmationRecord(confirmationId)
		return "", fmt.Errorf("更新持仓失败: %w", err)
	}

	// 6. 更新赎回申请状态
	err = l.updateApplicationStatus(app.ApplicationId, "赎回", "已确认")
	if err != nil {
		// 如果更新状态失败，标记为部分成功
		l.Error("更新赎回申请状态失败: %s, 确认记录已创建", app.ApplicationId)
	}

	l.Infof("赎回确认成功: applicationId=%s, confirmationId=%s, amount=%.2f, netValue=%.4f",
		app.ApplicationId, confirmationId, confirmedAmount, netValue)

	return confirmationId, nil
}

// getNetValueForDate 获取指定日期的产品净值
func (l *ConfirmTransactionsLogic) getNetValueForDate(productId string, date time.Time) (float64, error) {
	query := `
		SELECT unit_net_value FROM product_net_value 
		WHERE product_id = ? AND stat_date = ?
	`

	var netValue float64
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &netValue, query, productId, date.Format("2006-01-02"))
	if err != nil {
		if err == sqlx.ErrNotFound {
			// 尝试获取最近一天的净值
			fallbackQuery := `
				SELECT unit_net_value FROM product_net_value 
				WHERE product_id = ? 
				ORDER BY stat_date DESC LIMIT 1
			`
			err = l.svcCtx.Conn.QueryRowCtx(l.ctx, &netValue, fallbackQuery, productId)
			if err != nil {
				return 0, fmt.Errorf("无法获取产品净值: %w", err)
			}
		} else {
			return 0, fmt.Errorf("查询净值失败: %w", err)
		}
	}

	return netValue, nil
}

// generateConfirmationId 生成确认编号
func (l *ConfirmTransactionsLogic) generateConfirmationId(date time.Time) (string, error) {
	dateStr := date.Format("20060102")

	query := "SELECT COALESCE(MAX(CAST(SUBSTRING(confirmation_id, 9) AS UNSIGNED)), 0) FROM transaction_confirmation WHERE confirmation_id LIKE ?"
	var maxSeq int64

	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &maxSeq, query, dateStr+"%")
	if err != nil && err != sqlx.ErrNotFound {
		return "", err
	}

	seq := maxSeq + 1
	return fmt.Sprintf("%s%05d", dateStr, seq), nil
}

// updatePositionForPurchase 更新申购持仓
func (l *ConfirmTransactionsLogic) updatePositionForPurchase(app model.PurchaseApplication, confirmedShares float64, confirmationDate time.Time, netValue float64) error {
	// 1. 检查是否存在持仓记录
	query := `
		SELECT position_id FROM customer_position 
		WHERE customer_id = ? AND product_id = ? AND card_id = ?
		AND position_date = ?
	`

	var positionId int64
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &positionId, query,
		app.CustomerId, app.ProductId, app.CardId, confirmationDate.Format("2006-01-02"))

	if err != nil && err != sqlx.ErrNotFound {
		return fmt.Errorf("检查持仓失败: %w", err)
	}

	now := time.Now()
	if err == sqlx.ErrNotFound {
		// 创建新的持仓记录
		insertQuery := `
			INSERT INTO customer_position 
			(customer_id, product_id, card_id, total_shares, available_shares, 
			 frozen_shares, average_cost, position_date, create_time, update_time)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		// 计算平均成本
		averageCost := netValue

		_, err = l.svcCtx.Conn.ExecCtx(l.ctx, insertQuery,
			app.CustomerId,
			app.ProductId,
			app.CardId,
			confirmedShares,
			confirmedShares,
			0.0000,
			averageCost,
			confirmationDate.Format("2006-01-02"),
			now,
			now,
		)
		if err != nil {
			return fmt.Errorf("创建持仓记录失败: %w", err)
		}
	} else {
		// 更新现有持仓记录
		updateQuery := `
			UPDATE customer_position 
			SET total_shares = total_shares + ?,
				available_shares = available_shares + ?,
				-- 重新计算平均成本
				average_cost = (average_cost * total_shares + ? * ?) / (total_shares + ?),
				update_time = ?
			WHERE customer_id = ? AND product_id = ? AND card_id = ?
			AND position_date = ?
		`

		_, err = l.svcCtx.Conn.ExecCtx(l.ctx, updateQuery,
			confirmedShares,
			confirmedShares,
			netValue,
			confirmedShares,
			confirmedShares,
			now,
			app.CustomerId,
			app.ProductId,
			app.CardId,
			confirmationDate.Format("2006-01-02"),
		)
		if err != nil {
			return fmt.Errorf("更新持仓记录失败: %w", err)
		}
	}

	return nil
}

// updatePositionForRedemption 更新赎回持仓
func (l *ConfirmTransactionsLogic) updatePositionForRedemption(app model.RedemptionApplication, confirmationDate time.Time) error {
	updateQuery := `
		UPDATE customer_position 
		SET total_shares = total_shares - ?,
			frozen_shares = frozen_shares - ?,
			update_time = NOW()
		WHERE customer_id = ? AND product_id = ? AND card_id = ?
		AND position_date = ?
	`

	result, err := l.svcCtx.Conn.ExecCtx(l.ctx, updateQuery,
		app.ApplicationShares,
		app.ApplicationShares,
		app.CustomerId,
		app.ProductId,
		app.CardId,
		confirmationDate.AddDate(0, 0, -1).Format("2006-01-02"),
	)

	if err != nil {
		return fmt.Errorf("更新赎回持仓失败: %w", err)
	}

	// 检查是否成功更新
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("未找到对应的持仓记录")
	}

	return nil
}

// updateApplicationStatus 更新申请状态
func (l *ConfirmTransactionsLogic) updateApplicationStatus(applicationId, applicationType, status string) error {
	var tableName string
	if applicationType == "申购" {
		tableName = "purchase_application"
	} else if applicationType == "赎回" {
		tableName = "redemption_application"
	} else {
		return fmt.Errorf("未知的申请类型: %s", applicationType)
	}

	query := fmt.Sprintf("UPDATE %s SET application_status = ?, update_time = NOW() WHERE application_id = ?", tableName)

	_, err := l.svcCtx.Conn.ExecCtx(l.ctx, query, status, applicationId)
	if err != nil {
		return fmt.Errorf("更新申请状态失败: %w", err)
	}

	return nil
}

// deleteConfirmationRecord 删除确认记录（错误处理用）
func (l *ConfirmTransactionsLogic) deleteConfirmationRecord(confirmationId string) {
	query := "DELETE FROM transaction_confirmation WHERE confirmation_id = ?"
	_, err := l.svcCtx.Conn.ExecCtx(l.ctx, query, confirmationId)
	if err != nil {
		l.Errorf("删除确认记录失败: confirmationId=%s, error=%v", confirmationId, err)
	}
}

// ConfirmResult 确认结果结构体
type ConfirmResult struct {
	SuccessCount       int
	TotalFees          float64
	ConfirmationIds    []string
	FailedApplications []types.FailedApplication
}

// successResponse 成功响应
func (l *ConfirmTransactionsLogic) successResponse(message string, data interface{}) *types.BaseResponse {
	return &types.BaseResponse{
		Code:    200,
		Message: message,
		Data:    data,
	}
}

// errorResponse 错误响应
func (l *ConfirmTransactionsLogic) errorResponse(code int64, message string, data interface{}) *types.BaseResponse {
	return &types.BaseResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
