// getcustomerproductslogic.go
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

type GetCustomerProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCustomerProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCustomerProductsLogic {
	return &GetCustomerProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCustomerProductsLogic) GetCustomerProducts(req *types.GetCustomerProductsReq) (resp *types.BaseResponse, err error) {
	// 记录请求参数
	l.Infof("获取客户产品持仓: customerId=%s", req.CustomerId)

	// 1. 验证客户是否存在
	customer, err := l.svcCtx.CustomerInfoModel.FindOne(l.ctx, req.CustomerId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return l.errorResponse(404, fmt.Sprintf("客户不存在: %s", req.CustomerId), nil), nil
		}
		l.Errorf("查询客户信息失败: %v", err)
		return l.errorResponse(500, "查询客户信息失败: "+err.Error(), nil), nil
	}

	// 2. 获取当前日期
	//currentDate := time.Now().Format("2006-01-02")

	// 3. 查询客户的所有产品持仓
	products, totalMarketValue, totalProfitLoss, err := l.queryCustomerProducts(req.CustomerId)
	if err != nil {
		l.Errorf("查询客户产品持仓失败: %v", err)
		return l.errorResponse(500, "查询客户产品持仓失败: "+err.Error(), nil), nil
	}

	// 4. 构建响应
	response := types.CustomerProductsResp{
		CustomerId:       req.CustomerId,
		CustomerName:     customer.CustomerName,
		TotalProducts:    int64(len(products)),
		TotalMarketValue: totalMarketValue,
		TotalProfitLoss:  totalProfitLoss,
		Products:         products,
	}

	return l.successResponse("获取客户产品持仓成功", response), nil
}

// queryCustomerProducts 查询客户的产品持仓
func (l *GetCustomerProductsLogic) queryCustomerProducts(customerId string) ([]types.CustomerProductPosition, float64, float64, error) {
	// 查询客户当天的产品持仓
	query := `
		SELECT 
			cp.product_id,
			COALESCE(SUM(cp.total_shares), 0) as total_shares,
			COALESCE(SUM(cp.available_shares), 0) as available_shares,
			COALESCE(SUM(cp.frozen_shares), 0) as frozen_shares,
			COALESCE(AVG(cp.average_cost), 0) as average_cost
		FROM customer_position cp
		WHERE cp.customer_id = ? 
		AND cp.total_shares > 0
		GROUP BY cp.product_id
	`

	var basePositions []struct {
		ProductId       string  `db:"product_id"`
		TotalShares     float64 `db:"total_shares"`
		AvailableShares float64 `db:"available_shares"`
		FrozenShares    float64 `db:"frozen_shares"`
		AverageCost     float64 `db:"average_cost"`
	}

	err := l.svcCtx.Conn.QueryRowsCtx(l.ctx, &basePositions, query, customerId)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, 0, 0, fmt.Errorf("查询基础持仓失败: %w", err)
	}

	// 如果没有当天的持仓记录，尝试查找最近一天的持仓
	if len(basePositions) == 0 {
		return l.queryLatestCustomerProducts(customerId)
	}

	var products []types.CustomerProductPosition
	var totalMarketValue, totalProfitLoss float64

	for _, base := range basePositions {
		// 获取产品信息
		product, err := l.svcCtx.ProductInfoModel.FindOne(l.ctx, base.ProductId)
		if err != nil {
			l.Infof("获取产品信息失败: productId=%s, error=%v", base.ProductId, err)
			continue
		}

		// 获取最新净值
		latestNetValue, err := l.getLatestNetValue(base.ProductId)
		if err != nil {
			l.Infof("获取最新净值失败: productId=%s, error=%v", base.ProductId, err)
			latestNetValue = 0
		}

		// 获取持有天数
		holdingDays, err := l.getHoldingDays(customerId, base.ProductId)
		if err != nil {
			l.Infof("获取持有天数失败: customerId=%s, productId=%s, error=%v", customerId, base.ProductId, err)
			holdingDays = 0
		}

		// 计算市值和盈亏
		marketValue := base.TotalShares * latestNetValue
		costValue := base.TotalShares * base.AverageCost
		profitLoss := marketValue - costValue
		var profitLossRate float64
		if costValue > 0 {
			profitLossRate = (profitLoss / costValue) * 100
		}

		// 构建产品持仓信息
		position := types.CustomerProductPosition{
			ProductId:       base.ProductId,
			ProductName:     product.ProductName,
			ProductType:     product.ProductType,
			RiskLevel:       product.RiskLevel,
			ProductStatus:   product.ProductStatus,
			TotalShares:     base.TotalShares,
			AvailableShares: base.AvailableShares,
			FrozenShares:    base.FrozenShares,
			AverageCost:     base.AverageCost,
			LatestNetValue:  latestNetValue,
			MarketValue:     marketValue,
			CostValue:       costValue,
			ProfitLoss:      profitLoss,
			ProfitLossRate:  profitLossRate,
			HoldingDays:     holdingDays,
		}

		products = append(products, position)
		totalMarketValue += marketValue
		totalProfitLoss += profitLoss
	}

	return products, totalMarketValue, totalProfitLoss, nil
}

// queryLatestCustomerProducts 查询最近的产品持仓
func (l *GetCustomerProductsLogic) queryLatestCustomerProducts(customerId string) ([]types.CustomerProductPosition, float64, float64, error) {
	// 查找最近有持仓记录的日期
	dateQuery := `
		SELECT MAX(position_date) as latest_date
		FROM customer_position
		WHERE customer_id = ?
	`

	var latestDate string
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &latestDate, dateQuery, customerId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			// 没有任何持仓记录
			return []types.CustomerProductPosition{}, 0, 0, nil
		}
		return nil, 0, 0, fmt.Errorf("查询最近持仓日期失败: %w", err)
	}

	if latestDate == "" {
		// 没有任何持仓记录
		return []types.CustomerProductPosition{}, 0, 0, nil
	}

	l.Infof("使用最近持仓日期: %s", latestDate)
	// 使用最近日期查询
	return l.queryCustomerProducts(customerId)
}

// getLatestNetValue 获取最新净值
func (l *GetCustomerProductsLogic) getLatestNetValue(productId string) (float64, error) {
	query := `
		SELECT unit_net_value 
		FROM product_net_value 
		WHERE product_id = ? 
		ORDER BY stat_date DESC 
		LIMIT 1
	`
	var netValue float64
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &netValue, query, productId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return 0, fmt.Errorf("产品净值未找到: %s", productId)
		}
		return 0, fmt.Errorf("查询净值失败: %w", err)
	}
	return netValue, nil
}

// getHoldingDays 获取持有天数
func (l *GetCustomerProductsLogic) getHoldingDays(customerId, productId string) (int64, error) {
	query := `
		SELECT MIN(confirmation_date) as first_date
		FROM transaction_confirmation 
		WHERE customer_id = ? 
		AND product_id = ? 
		AND confirmation_status = '确认成功'
		AND transaction_type = '申购'
	`

	var firstDate string
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &firstDate, query, customerId, productId)
	if err != nil && err != sqlx.ErrNotFound {
		return 0, fmt.Errorf("查询首次购买日期失败: %w", err)
	}

	if firstDate == "" {
		// 如果没有购买记录，可能是初始持仓或没有交易
		return 0, nil
	}

	// 计算持有天数
	firstTime, err := time.Parse("2006-01-02", firstDate)
	if err != nil {
		return 0, fmt.Errorf("解析日期失败: %w", err)
	}

	holdingDays := int64(time.Since(firstTime).Hours() / 24)
	if holdingDays < 0 {
		holdingDays = 0
	}

	return holdingDays, nil
}

// successResponse 成功响应
func (l *GetCustomerProductsLogic) successResponse(message string, data interface{}) *types.BaseResponse {
	return &types.BaseResponse{
		Code:    200,
		Message: message,
		Data:    data,
	}
}

// errorResponse 错误响应
func (l *GetCustomerProductsLogic) errorResponse(code int64, message string, data interface{}) *types.BaseResponse {
	return &types.BaseResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
