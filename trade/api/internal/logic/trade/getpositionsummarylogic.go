// getpositionsummarylogic.go
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

type GetPositionSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPositionSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionSummaryLogic {
	return &GetPositionSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPositionSummaryLogic) GetPositionSummary(customerId string) (resp *types.BaseResponse, err error) {
	// 记录请求参数
	l.Infof("收到客户持仓汇总请求: customerId=%s", customerId)

	// 1. 验证客户是否存在
	customer, err := l.svcCtx.CustomerInfoModel.FindOne(l.ctx, customerId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return l.errorResponse(404, fmt.Sprintf("客户不存在: %s", customerId), nil), nil
		}
		l.Errorf("查询客户信息失败: %v", err)
		return l.errorResponse(500, "查询客户信息失败: "+err.Error(), nil), nil
	}

	// 2. 查询客户的持仓汇总统计
	summary, err := l.queryPositionSummary(customerId)
	if err != nil {
		l.Errorf("查询持仓汇总失败: %v", err)
		return l.errorResponse(500, "查询持仓汇总失败: "+err.Error(), nil), nil
	}

	// 3. 设置客户信息
	summary.CustomerId = customerId
	summary.CustomerName = customer.CustomerName

	return l.successResponse("持仓汇总查询成功", summary), nil
}

// queryPositionSummary 查询持仓汇总统计
func (l *GetPositionSummaryLogic) queryPositionSummary(customerId string) (*types.PositionSummary, error) {
	// 查询当前日期
	currentDate := time.Now().Format("2006-01-02")

	// 主查询：获取基本持仓统计 - 使用 COALESCE 处理 NULL 值
	query := `
		SELECT 
			COUNT(*) as total_positions,
			COUNT(DISTINCT cp.product_id) as total_products,
			COUNT(DISTINCT cp.card_id) as total_cards,
			COALESCE(SUM(cp.total_shares), 0) as total_shares,
			COALESCE(SUM(cp.available_shares), 0) as total_available_shares,
			COALESCE(SUM(cp.frozen_shares), 0) as total_frozen_shares,
			COALESCE(AVG(cp.average_cost), 0) as avg_cost
		FROM customer_position cp
		WHERE cp.customer_id = ? 
		AND cp.position_date = ?
	`

	var baseStats struct {
		TotalPositions       int64   `db:"total_positions"`
		TotalProducts        int64   `db:"total_products"`
		TotalCards           int64   `db:"total_cards"`
		TotalShares          float64 `db:"total_shares"`
		TotalAvailableShares float64 `db:"total_available_shares"`
		TotalFrozenShares    float64 `db:"total_frozen_shares"`
		AvgCost              float64 `db:"avg_cost"`
	}

	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &baseStats, query, customerId, currentDate)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, fmt.Errorf("查询基础持仓统计失败: %w", err)
	}

	// 如果当天没有持仓记录，尝试查找最近一天的持仓
	if baseStats.TotalPositions == 0 {
		l.Infof("客户 %s 当天没有持仓记录，尝试查找最近持仓", customerId)
		return l.queryLatestPositionSummary(customerId)
	}

	// 查询持仓市值和盈亏信息 - 使用 COALESCE 处理 NULL 值
	marketValueQuery := `
		SELECT 
			COALESCE(SUM(cp.total_shares * COALESCE(pnv.unit_net_value, 0)), 0) as total_market_value,
			COALESCE(SUM(cp.total_shares * COALESCE(cp.average_cost, 0)), 0) as total_cost_value,
			COALESCE(SUM(cp.total_shares * COALESCE(pnv.unit_net_value, 0) - cp.total_shares * COALESCE(cp.average_cost, 0)), 0) as total_profit_loss
		FROM customer_position cp
		LEFT JOIN product_net_value pnv ON cp.product_id = pnv.product_id 
			AND pnv.stat_date = (
				SELECT MAX(stat_date) 
				FROM product_net_value 
				WHERE product_id = cp.product_id 
				AND stat_date <= ?
			)
		WHERE cp.customer_id = ? 
		AND cp.position_date = ?
	`

	var marketStats struct {
		TotalMarketValue float64 `db:"total_market_value"`
		TotalCostValue   float64 `db:"total_cost_value"`
		TotalProfitLoss  float64 `db:"total_profit_loss"`
	}

	err = l.svcCtx.Conn.QueryRowCtx(l.ctx, &marketStats, marketValueQuery, currentDate, customerId, currentDate)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, fmt.Errorf("查询市值盈亏统计失败: %w", err)
	}

	// 计算平均盈亏比例
	var avgProfitLossRate float64
	if marketStats.TotalCostValue > 0 {
		avgProfitLossRate = (marketStats.TotalProfitLoss / marketStats.TotalCostValue) * 100
	}

	// 构建返回结果
	summary := &types.PositionSummary{
		TotalPositions:    baseStats.TotalPositions,
		TotalProducts:     baseStats.TotalProducts,
		TotalCards:        baseStats.TotalCards,
		TotalMarketValue:  roundFloat(marketStats.TotalMarketValue, 2),
		TotalCost:         roundFloat(marketStats.TotalCostValue, 2),
		TotalProfitLoss:   roundFloat(marketStats.TotalProfitLoss, 2),
		AvgProfitLossRate: roundFloat(avgProfitLossRate, 2),
	}

	return summary, nil
}

// queryLatestPositionSummary 查询最近一天的持仓汇总
func (l *GetPositionSummaryLogic) queryLatestPositionSummary(customerId string) (*types.PositionSummary, error) {
	// 查找客户最近有持仓记录的日期 - 使用 COALESCE 处理 NULL 值
	dateQuery := `
		SELECT COALESCE(MAX(position_date), '') as latest_date
		FROM customer_position
		WHERE customer_id = ?
	`

	var latestDate string
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &latestDate, dateQuery, customerId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			// 客户没有任何持仓记录
			return &types.PositionSummary{
				CustomerId:        customerId,
				TotalPositions:    0,
				TotalProducts:     0,
				TotalCards:        0,
				TotalMarketValue:  0,
				TotalCost:         0,
				TotalProfitLoss:   0,
				AvgProfitLossRate: 0,
			}, nil
		}
		return nil, fmt.Errorf("查询最近持仓日期失败: %w", err)
	}

	if latestDate == "" {
		// 客户没有任何持仓记录
		return &types.PositionSummary{
			CustomerId:        customerId,
			TotalPositions:    0,
			TotalProducts:     0,
			TotalCards:        0,
			TotalMarketValue:  0,
			TotalCost:         0,
			TotalProfitLoss:   0,
			AvgProfitLossRate: 0,
		}, nil
	}

	// 使用最近日期重新查询 - 使用 COALESCE 处理 NULL 值
	l.Infof("使用最近持仓日期: %s", latestDate)

	// 重新查询汇总信息
	query := `
		SELECT 
			COUNT(*) as total_positions,
			COUNT(DISTINCT product_id) as total_products,
			COUNT(DISTINCT card_id) as total_cards
		FROM customer_position
		WHERE customer_id = ? 
		AND position_date = ?
	`

	var baseStats struct {
		TotalPositions int64 `db:"total_positions"`
		TotalProducts  int64 `db:"total_products"`
		TotalCards     int64 `db:"total_cards"`
	}

	err = l.svcCtx.Conn.QueryRowCtx(l.ctx, &baseStats, query, customerId, latestDate)
	if err != nil {
		if err == sqlx.ErrNotFound {
			// 即使找到日期，但可能没有记录
			return &types.PositionSummary{
				CustomerId:        customerId,
				TotalPositions:    0,
				TotalProducts:     0,
				TotalCards:        0,
				TotalMarketValue:  0,
				TotalCost:         0,
				TotalProfitLoss:   0,
				AvgProfitLossRate: 0,
			}, nil
		}
		return nil, fmt.Errorf("查询最近持仓统计失败: %w", err)
	}

	// 查询最近日期的市值和盈亏 - 使用 COALESCE 处理 NULL 值
	marketValueQuery := `
		SELECT 
			COALESCE(SUM(cp.total_shares * COALESCE(pnv.unit_net_value, 0)), 0) as total_market_value,
			COALESCE(SUM(cp.total_shares * COALESCE(cp.average_cost, 0)), 0) as total_cost_value,
			COALESCE(SUM(cp.total_shares * COALESCE(pnv.unit_net_value, 0) - cp.total_shares * COALESCE(cp.average_cost, 0)), 0) as total_profit_loss
		FROM customer_position cp
		LEFT JOIN product_net_value pnv ON cp.product_id = pnv.product_id 
			AND pnv.stat_date = (
				SELECT MAX(stat_date) 
				FROM product_net_value 
				WHERE product_id = cp.product_id 
				AND stat_date <= ?
			)
		WHERE cp.customer_id = ? 
		AND cp.position_date = ?
	`

	var marketStats struct {
		TotalMarketValue float64 `db:"total_market_value"`
		TotalCostValue   float64 `db:"total_cost_value"`
		TotalProfitLoss  float64 `db:"total_profit_loss"`
	}

	err = l.svcCtx.Conn.QueryRowCtx(l.ctx, &marketStats, marketValueQuery, latestDate, customerId, latestDate)
	if err != nil && err != sqlx.ErrNotFound {
		// 如果查询失败，只返回基础统计，不返回市值
		l.Errorf("查询最近持仓市值失败: %v", err)
		marketStats.TotalMarketValue = 0
		marketStats.TotalCostValue = 0
		marketStats.TotalProfitLoss = 0
	}

	// 计算平均盈亏比例
	var avgProfitLossRate float64
	if marketStats.TotalCostValue > 0 {
		avgProfitLossRate = (marketStats.TotalProfitLoss / marketStats.TotalCostValue) * 100
	}

	// 返回汇总信息
	summary := &types.PositionSummary{
		CustomerId:        customerId,
		TotalPositions:    baseStats.TotalPositions,
		TotalProducts:     baseStats.TotalProducts,
		TotalCards:        baseStats.TotalCards,
		TotalMarketValue:  roundFloat(marketStats.TotalMarketValue, 2),
		TotalCost:         roundFloat(marketStats.TotalCostValue, 2),
		TotalProfitLoss:   roundFloat(marketStats.TotalProfitLoss, 2),
		AvgProfitLossRate: roundFloat(avgProfitLossRate, 2),
	}

	return summary, nil
}

// queryProductDistribution 查询产品分布详情
func (l *GetPositionSummaryLogic) queryProductDistribution(customerId, date string) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			cp.product_id,
			pi.product_name,
			pi.product_type,
			pi.risk_level,
			COALESCE(SUM(cp.total_shares), 0) as total_shares,
			COALESCE(SUM(cp.available_shares), 0) as available_shares,
			COALESCE(SUM(cp.frozen_shares), 0) as frozen_shares,
			COALESCE(AVG(cp.average_cost), 0) as avg_cost,
			COALESCE(pnv.unit_net_value, 0) as latest_net_value
		FROM customer_position cp
		LEFT JOIN product_info pi ON cp.product_id = pi.product_id
		LEFT JOIN product_net_value pnv ON cp.product_id = pnv.product_id 
			AND pnv.stat_date = (
				SELECT MAX(stat_date) 
				FROM product_net_value 
				WHERE product_id = cp.product_id 
				AND stat_date <= ?
			)
		WHERE cp.customer_id = ? 
		AND cp.position_date = ?
		GROUP BY cp.product_id, pi.product_name, pi.product_type, pi.risk_level, pnv.unit_net_value
		ORDER BY total_shares DESC
	`

	var distribution []struct {
		ProductId       string  `db:"product_id"`
		ProductName     string  `db:"product_name"`
		ProductType     string  `db:"product_type"`
		RiskLevel       string  `db:"risk_level"`
		TotalShares     float64 `db:"total_shares"`
		AvailableShares float64 `db:"available_shares"`
		FrozenShares    float64 `db:"frozen_shares"`
		AvgCost         float64 `db:"avg_cost"`
		LatestNetValue  float64 `db:"latest_net_value"`
	}

	err := l.svcCtx.Conn.QueryRowsCtx(l.ctx, &distribution, query, date, customerId, date)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, fmt.Errorf("查询产品分布失败: %w", err)
	}

	// 转换为map格式
	var result []map[string]interface{}
	for _, item := range distribution {
		marketValue := item.TotalShares * item.LatestNetValue
		costValue := item.TotalShares * item.AvgCost
		profitLoss := marketValue - costValue
		var profitLossRate float64
		if costValue > 0 {
			profitLossRate = (profitLoss / costValue) * 100
		}

		product := map[string]interface{}{
			"productId":       item.ProductId,
			"productName":     item.ProductName,
			"productType":     item.ProductType,
			"riskLevel":       item.RiskLevel,
			"totalShares":     roundFloat(item.TotalShares, 4),
			"availableShares": roundFloat(item.AvailableShares, 4),
			"frozenShares":    roundFloat(item.FrozenShares, 4),
			"avgCost":         roundFloat(item.AvgCost, 4),
			"latestNetValue":  roundFloat(item.LatestNetValue, 4),
			"marketValue":     roundFloat(marketValue, 2),
			"costValue":       roundFloat(costValue, 2),
			"profitLoss":      roundFloat(profitLoss, 2),
			"profitLossRate":  roundFloat(profitLossRate, 2),
		}
		result = append(result, product)
	}

	return result, nil
}

// roundFloat 四舍五入浮点数
func roundFloat(value float64, precision int) float64 {
	// 简单的四舍五入实现
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10.0
	}
	return float64(int64(value*multiplier+0.5)) / multiplier
}

// successResponse 成功响应
func (l *GetPositionSummaryLogic) successResponse(message string, data interface{}) *types.BaseResponse {
	return &types.BaseResponse{
		Code:    200,
		Message: message,
		Data:    data,
	}
}

// errorResponse 错误响应
func (l *GetPositionSummaryLogic) errorResponse(code int64, message string, data interface{}) *types.BaseResponse {
	return &types.BaseResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
