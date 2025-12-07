// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"WMSS/trade/api/internal/svc"
	"WMSS/trade/api/internal/types"
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
	// 查询总数
	countQuery := "SELECT COUNT(*) FROM purchase_application " + whereClause
	var total int64
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询申购总数失败: %w", err)
	}

	// 查询数据
	query := `
		SELECT 
			application_id,
			customer_id,
			product_id,
			application_amount,
			application_date,
			application_time,
			application_status,
			operator_id
		FROM purchase_application
	` + whereClause + " ORDER BY application_date DESC, application_time DESC"

	// 添加分页
	if req.Size > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", req.Size, (req.Page-1)*req.Size)
	}

	// 使用临时结构体接收查询结果
	type tempPurchaseRecord struct {
		ApplicationId     string  `db:"application_id"`
		CustomerId        string  `db:"customer_id"`
		ProductId         string  `db:"product_id"`
		ApplicationAmount float64 `db:"application_amount"`
		ApplicationDate   string  `db:"application_date"`
		ApplicationTime   string  `db:"application_time"`
		ApplicationStatus string  `db:"application_status"`
		OperatorId        string  `db:"operator_id"`
	}

	var tempRecords []tempPurchaseRecord
	err = l.svcCtx.Conn.QueryRowsCtx(l.ctx, &tempRecords, query, args...)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, 0, fmt.Errorf("查询申购记录失败: %w", err)
	}

	// 转换为 TransactionRecord
	records := make([]types.TransactionRecord, len(tempRecords))
	for i, temp := range tempRecords {
		records[i] = types.TransactionRecord{
			ApplicationId:   temp.ApplicationId,
			Type:            "PURCHASE",
			CustomerId:      temp.CustomerId,
			ProductId:       temp.ProductId,
			Amount:          temp.ApplicationAmount,
			ApplicationDate: temp.ApplicationDate,
			ApplicationTime: temp.ApplicationTime, // 只取时间部分
			Status:          temp.ApplicationStatus,
			OperatorId:      temp.OperatorId,
		}
	}

	return records, total, nil
}

func (l *QueryTransactionsLogic) queryRedemptionRecords(req *types.TransactionQuery, whereClause string, args []interface{}) ([]types.TransactionRecord, int64, error) {
	// 查询总数
	countQuery := "SELECT COUNT(*) FROM redemption_application " + whereClause
	var total int64
	err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询赎回总数失败: %w", err)
	}

	// 查询数据
	query := `
		SELECT 
			application_id,
			customer_id,
			product_id,
			application_shares,
			application_date,
			application_time,
			application_status,
			operator_id
		FROM redemption_application
	` + whereClause + " ORDER BY application_date DESC, application_time DESC"

	// 添加分页
	if req.Size > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", req.Size, (req.Page-1)*req.Size)
	}

	// 使用临时结构体接收查询结果
	type tempRedemptionRecord struct {
		ApplicationId     string  `db:"application_id"`
		CustomerId        string  `db:"customer_id"`
		ProductId         string  `db:"product_id"`
		ApplicationShares float64 `db:"application_shares"`
		ApplicationDate   string  `db:"application_date"`
		ApplicationTime   string  `db:"application_time"`
		ApplicationStatus string  `db:"application_status"`
		OperatorId        string  `db:"operator_id"`
	}

	var tempRecords []tempRedemptionRecord
	err = l.svcCtx.Conn.QueryRowsCtx(l.ctx, &tempRecords, query, args...)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, 0, fmt.Errorf("查询赎回记录失败: %w", err)
	}

	// 转换为 TransactionRecord
	records := make([]types.TransactionRecord, len(tempRecords))
	for i, temp := range tempRecords {
		records[i] = types.TransactionRecord{
			ApplicationId:   temp.ApplicationId,
			Type:            "REDEMPTION",
			CustomerId:      temp.CustomerId,
			ProductId:       temp.ProductId,
			Shares:          temp.ApplicationShares,
			ApplicationDate: temp.ApplicationDate,
			ApplicationTime: temp.ApplicationTime, // 只取时间部分
			Status:          temp.ApplicationStatus,
			OperatorId:      temp.OperatorId,
		}
	}

	return records, total, nil
}

// getCustomerInfos 批量获取客户信息
func (l *QueryTransactionsLogic) getCustomerInfos(customerIds []string) (map[string]string, error) {
	if len(customerIds) == 0 {
		return map[string]string{}, nil
	}

	placeholders := strings.Repeat("?,", len(customerIds))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf("SELECT customer_id, customer_name FROM customer_info WHERE customer_id IN (%s)", placeholders)

	// 将 []string 转换为 []interface{}
	args := make([]interface{}, len(customerIds))
	for i, id := range customerIds {
		args[i] = id
	}

	// 使用自定义结构体接收结果
	type customerResult struct {
		CustomerId   string `db:"customer_id"`
		CustomerName string `db:"customer_name"`
	}

	var results []customerResult
	err := l.svcCtx.Conn.QueryRowsCtx(l.ctx, &results, query, args...)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, fmt.Errorf("查询客户信息失败: %w", err)
	}

	// 转换为 map
	result := make(map[string]string)
	for _, r := range results {
		result[r.CustomerId] = r.CustomerName
	}

	return result, nil
}

// getProductInfos 批量获取产品信息
func (l *QueryTransactionsLogic) getProductInfos(productIds []string) (map[string]string, error) {
	if len(productIds) == 0 {
		return map[string]string{}, nil
	}

	placeholders := strings.Repeat("?,", len(productIds))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf("SELECT product_id, product_name FROM product_info WHERE product_id IN (%s)", placeholders)

	// 将 []string 转换为 []interface{}
	args := make([]interface{}, len(productIds))
	for i, id := range productIds {
		args[i] = id
	}

	// 使用自定义结构体接收结果
	type productResult struct {
		ProductId   string `db:"product_id"`
		ProductName string `db:"product_name"`
	}

	var results []productResult
	err := l.svcCtx.Conn.QueryRowsCtx(l.ctx, &results, query, args...)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, fmt.Errorf("查询产品信息失败: %w", err)
	}

	// 转换为 map
	result := make(map[string]string)
	for _, r := range results {
		result[r.ProductId] = r.ProductName
	}

	return result, nil
}

// enrichRecordsWithDetails 批量获取客户和产品信息
func (l *QueryTransactionsLogic) enrichRecordsWithDetails(records []types.TransactionRecord) ([]types.TransactionRecord, error) {
	if len(records) == 0 {
		return records, nil
	}

	// 获取客户ID和产品ID列表
	customerIds := make([]string, 0, len(records))
	productIds := make([]string, 0, len(records))

	customerSet := make(map[string]bool)
	productSet := make(map[string]bool)

	for _, record := range records {
		if !customerSet[record.CustomerId] {
			customerSet[record.CustomerId] = true
			customerIds = append(customerIds, record.CustomerId)
		}
		if !productSet[record.ProductId] {
			productSet[record.ProductId] = true
			productIds = append(productIds, record.ProductId)
		}
	}

	// 批量查询客户信息
	customerNames, err := l.getCustomerInfos(customerIds)
	if err != nil {
		return nil, err
	}

	// 批量查询产品信息
	productNames, err := l.getProductInfos(productIds)
	if err != nil {
		return nil, err
	}

	// 填充详细信息
	for i := range records {
		if name, ok := customerNames[records[i].CustomerId]; ok {
			records[i].CustomerName = name
		} else {
			records[i].CustomerName = "未知客户"
		}

		if name, ok := productNames[records[i].ProductId]; ok {
			records[i].ProductName = name
		} else {
			records[i].ProductName = "未知产品"
		}
	}

	return records, nil
}

// getKeys 从map中提取键值
func (l *QueryTransactionsLogic) getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
