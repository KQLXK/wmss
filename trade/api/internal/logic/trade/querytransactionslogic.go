// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"WMSS/trade/api/internal/svc"
	"WMSS/trade/api/internal/types"
	"WMSS/trade/api/model"
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type QueryTransactionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryTransactionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryTransactionsLogic {
	return &QueryTransactionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryTransactionsLogic) QueryTransactions(req *types.TransactionQuery) (resp *types.BaseResponse, err error) {
	// 1. 构建查询条件
	whereClause, args := l.buildWhereClause(req)

	// 2. 查询申购记录
	purchaseRecords, purchaseTotal, err := l.queryPurchaseRecords(req, whereClause, args)
	if err != nil {
		l.Errorf("查询申购记录失败: %v", err)
		return &types.BaseResponse{
			Code:    500,
			Message: "查询申购记录失败: " + err.Error(),
		}, nil
	}

	// 3. 查询赎回记录
	redemptionRecords, redemptionTotal, err := l.queryRedemptionRecords(req, whereClause, args)
	if err != nil {
		l.Errorf("查询赎回记录失败: %v", err)
		return &types.BaseResponse{
			Code:    500,
			Message: "查询赎回记录失败: " + err.Error(),
		}, nil
	}

	// 4. 合并结果
	allRecords := append(purchaseRecords, redemptionRecords...)
	totalRecords := purchaseTotal + redemptionTotal

	// 5. 批量查询客户和产品信息
	recordsWithDetails, err := l.enrichRecordsWithDetails(allRecords)
	if err != nil {
		l.Errorf("获取详细信息失败: %v", err)
		return &types.BaseResponse{
			Code:    500,
			Message: "获取详细信息失败: " + err.Error(),
		}, nil
	}

	// 6. 分页处理
	startIndex := (req.Page - 1) * req.Size
	endIndex := startIndex + req.Size
	if endIndex > int64(len(recordsWithDetails)) {
		endIndex = int64(len(recordsWithDetails))
	}

	pagedRecords := recordsWithDetails
	if startIndex < int64(len(recordsWithDetails)) {
		pagedRecords = recordsWithDetails[startIndex:endIndex]
	} else {
		pagedRecords = []types.TransactionRecord{}
	}

	// 7. 返回结果
	return &types.BaseResponse{
		Code:    200,
		Message: "查询成功",
		Data: types.PageResponse{
			Total:   totalRecords,
			Records: pagedRecords,
			Page:    req.Page,
			Size:    req.Size,
		},
	}, nil
}

// buildWhereClause 构建查询条件
func (l *QueryTransactionsLogic) buildWhereClause(req *types.TransactionQuery) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if req.ApplicationId != "" {
		conditions = append(conditions, "application_id = ?")
		args = append(args, req.ApplicationId)
	}

	if req.CustomerId != "" {
		conditions = append(conditions, "customer_id = ?")
		args = append(args, req.CustomerId)
	}

	if req.ProductId != "" {
		conditions = append(conditions, "product_id = ?")
		args = append(args, req.ProductId)
	}

	if req.StartDate != "" {
		conditions = append(conditions, "application_date >= ?")
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		conditions = append(conditions, "application_date <= ?")
		args = append(args, req.EndDate)
	}

	if req.Status != "" {
		conditions = append(conditions, "application_status = ?")
		args = append(args, req.Status)
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

// queryPurchaseRecords 查询申购记录
func (l *QueryTransactionsLogic) queryPurchaseRecords(req *types.TransactionQuery, whereClause string, args []interface{}) ([]types.TransactionRecord, int64, error) {
	baseQuery := `
		SELECT application_id, customer_id, product_id, application_amount, 
		       application_date, application_time, application_status, operator_id
		FROM purchase_application
	`

	countQuery := "SELECT COUNT(*) FROM purchase_application " + whereClause
	var total int64
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := baseQuery + whereClause + " ORDER BY application_date DESC, application_time DESC"
	if req.Size > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", req.Size, (req.Page-1)*req.Size)
	}

	var records []types.TransactionRecord
	err = l.svcCtx.Conn.QueryRowsCtx(l.ctx, &records, query, args...)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, 0, err
	}

	// 设置交易类型为PURCHASE
	for i := range records {
		records[i].Type = "PURCHASE"
		records[i].Amount = records[i].Amount // 申购金额
	}

	return records, total, nil
}

// queryRedemptionRecords 查询赎回记录
func (l *QueryTransactionsLogic) queryRedemptionRecords(req *types.TransactionQuery, whereClause string, args []interface{}) ([]types.TransactionRecord, int64, error) {
	baseQuery := `
		SELECT application_id, customer_id, product_id, shares, 
		       application_date, application_time, application_status, operator_id
		FROM redemption_application
	`

	countQuery := "SELECT COUNT(*) FROM redemption_application " + whereClause
	var total int64
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := baseQuery + whereClause + " ORDER BY application_date DESC, application_time DESC"
	if req.Size > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", req.Size, (req.Page-1)*req.Size)
	}

	var records []types.TransactionRecord
	err = l.svcCtx.Conn.QueryRowsCtx(l.ctx, &records, query, args...)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, 0, err
	}

	// 设置交易类型为REDEMPTION
	for i := range records {
		records[i].Type = "REDEMPTION"
		records[i].Shares = records[i].Shares // 赎回份额
	}

	return records, total, nil
}

// enrichRecordsWithDetails 批量获取客户和产品信息
func (l *QueryTransactionsLogic) enrichRecordsWithDetails(records []types.TransactionRecord) ([]types.TransactionRecord, error) {
	if len(records) == 0 {
		return records, nil
	}

	// 获取客户ID和产品ID列表
	customerIds := make(map[string]bool)
	productIds := make(map[string]bool)
	for _, record := range records {
		customerIds[record.CustomerId] = true
		productIds[record.ProductId] = true
	}

	// 批量查询客户信息
	customerInfos, err := l.getCustomerInfos(l.getKeys(customerIds))
	if err != nil {
		return nil, err
	}

	// 批量查询产品信息
	productInfos, err := l.getProductInfos(l.getKeys(productIds))
	if err != nil {
		return nil, err
	}

	// 填充详细信息
	for i := range records {
		if customer, ok := customerInfos[records[i].CustomerId]; ok {
			records[i].CustomerName = customer.CustomerName
		}
		if product, ok := productInfos[records[i].ProductId]; ok {
			records[i].ProductName = product.ProductName
		}
	}

	return records, nil
}

// getCustomerInfos 批量获取客户信息
func (l *QueryTransactionsLogic) getCustomerInfos(customerIds []string) (map[string]*model.CustomerInfo, error) {
	if len(customerIds) == 0 {
		return map[string]*model.CustomerInfo{}, nil
	}

	placeholders := strings.Repeat("?,", len(customerIds))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf("SELECT customer_id, customer_name FROM customer_info WHERE customer_id IN (%s)", placeholders)

	// 将 []string 转换为 []interface{}
	args := make([]interface{}, len(customerIds))
	for i, id := range customerIds {
		args[i] = id
	}

	var customers []*model.CustomerInfo
	err := l.svcCtx.Conn.QueryRowsCtx(l.ctx, &customers, query, args...)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	}

	result := make(map[string]*model.CustomerInfo)
	for _, customer := range customers {
		result[customer.CustomerId] = customer
	}

	return result, nil
}

// getProductInfos 批量获取产品信息
func (l *QueryTransactionsLogic) getProductInfos(productIds []string) (map[string]*model.ProductInfo, error) {
	if len(productIds) == 0 {
		return map[string]*model.ProductInfo{}, nil
	}

	placeholders := strings.Repeat("?,", len(productIds))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf("SELECT product_id, product_name FROM product_info WHERE product_id IN (%s)", placeholders)

	// 将 []string 转换为 []interface{}
	args := make([]interface{}, len(productIds))
	for i, id := range productIds {
		args[i] = id
	}

	var products []*model.ProductInfo
	err := l.svcCtx.Conn.QueryRowsCtx(l.ctx, &products, query, args...)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	}

	result := make(map[string]*model.ProductInfo)
	for _, product := range products {
		result[product.ProductId] = product
	}

	return result, nil
}

// getKeys 从map中提取键值
func (l *QueryTransactionsLogic) getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
