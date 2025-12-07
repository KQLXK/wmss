// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"WMSS/liquidation/api/internal/svc"
	"WMSS/liquidation/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPendingTransactionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPendingTransactionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPendingTransactionsLogic {
	return &GetPendingTransactionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPendingTransactionsLogic) GetPendingTransactions(req *types.PendingTransactionRequest) (resp *types.BaseResponse, err error) {
	// 参数校验
	applicationDate, err := time.Parse("2006-01-02", req.ApplicationDate)
	if err != nil {
		return &types.BaseResponse{
			Code:    400,
			Message: "申请日期格式错误，应为 YYYY-MM-DD",
		}, nil
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 50
	}
	if size > 200 {
		size = 200
	}
	offset := (page - 1) * size

	// 构建查询条件
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "application_date = ?")
	args = append(args, applicationDate)

	conditions = append(conditions, "application_status = '未确认'")

	if req.ProductId != "" {
		conditions = append(conditions, "product_id = ?")
		args = append(args, req.ProductId)
	}
	if req.CustomerId != "" {
		conditions = append(conditions, "customer_id = ?")
		args = append(args, req.CustomerId)
	}

	whereClause := strings.Join(conditions, " AND ")

	// 查询申购申请
	var purchaseRecords []struct {
		ApplicationId     string    `db:"application_id"`
		CustomerId        string    `db:"customer_id"`
		ProductId         string    `db:"product_id"`
		ApplicationAmount float64   `db:"application_amount"`
		ApplicationDate   time.Time `db:"application_date"`
		ApplicationStatus string    `db:"application_status"`
	}

	purchaseQuery := fmt.Sprintf(`
		SELECT application_id, customer_id, product_id, application_amount, application_date, application_status
		FROM purchase_application
		WHERE %s
		LIMIT %d OFFSET %d
	`, whereClause, size*2, offset)

	err = l.svcCtx.Conn.QueryRowsCtx(l.ctx, &purchaseRecords, purchaseQuery, args...)
	if err != nil {
		l.Errorf("查询申购申请失败: %v", err)
	}

	// 查询赎回申请
	var redemptionRecords []struct {
		ApplicationId     string    `db:"application_id"`
		CustomerId        string    `db:"customer_id"`
		ProductId         string    `db:"product_id"`
		ApplicationShares float64   `db:"application_shares"`
		ApplicationDate   time.Time `db:"application_date"`
		ApplicationStatus string    `db:"application_status"`
	}

	redemptionQuery := fmt.Sprintf(`
		SELECT application_id, customer_id, product_id, application_shares, application_date, application_status
		FROM redemption_application
		WHERE %s
		LIMIT %d OFFSET %d
	`, whereClause, size*2, offset)

	err = l.svcCtx.Conn.QueryRowsCtx(l.ctx, &redemptionRecords, redemptionQuery, args...)
	if err != nil {
		l.Errorf("查询赎回申请失败: %v", err)
	}

	// 如果指定了申请类型，进行过滤
	if req.ApplicationType == "PURCHASE" {
		redemptionRecords = []struct {
			ApplicationId     string    `db:"application_id"`
			CustomerId        string    `db:"customer_id"`
			ProductId         string    `db:"product_id"`
			ApplicationShares float64   `db:"application_shares"`
			ApplicationDate   time.Time `db:"application_date"`
			ApplicationStatus string    `db:"application_status"`
		}{}
	} else if req.ApplicationType == "REDEMPTION" {
		purchaseRecords = []struct {
			ApplicationId     string    `db:"application_id"`
			CustomerId        string    `db:"customer_id"`
			ProductId         string    `db:"product_id"`
			ApplicationAmount float64   `db:"application_amount"`
			ApplicationDate   time.Time `db:"application_date"`
			ApplicationStatus string    `db:"application_status"`
		}{}
	}

	// 先查询总数（使用 UNION）
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as cnt FROM (
			SELECT application_id FROM purchase_application WHERE %s
			UNION ALL
			SELECT application_id FROM redemption_application WHERE %s
		) as combined
	`, whereClause, whereClause)

	var total int64
	err = l.svcCtx.Conn.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		l.Errorf("查询总数失败: %v", err)
		total = int64(len(purchaseRecords) + len(redemptionRecords))
	}

	// 合并结果
	allRecords := make([]types.PendingTransactionRecord, 0, len(purchaseRecords)+len(redemptionRecords))

	// 处理申购记录
	for _, pr := range purchaseRecords {
		allRecords = append(allRecords, types.PendingTransactionRecord{
			ApplicationId:     pr.ApplicationId,
			ApplicationType:   "PURCHASE",
			CustomerId:        pr.CustomerId,
			ProductId:         pr.ProductId,
			ApplicationAmount: pr.ApplicationAmount,
			ApplicationDate:   pr.ApplicationDate.Format("2006-01-02"),
			ApplicationStatus: pr.ApplicationStatus,
		})
	}

	// 处理赎回记录
	for _, rr := range redemptionRecords {
		allRecords = append(allRecords, types.PendingTransactionRecord{
			ApplicationId:     rr.ApplicationId,
			ApplicationType:   "REDEMPTION",
			CustomerId:        rr.CustomerId,
			ProductId:         rr.ProductId,
			ApplicationShares: rr.ApplicationShares,
			ApplicationDate:   rr.ApplicationDate.Format("2006-01-02"),
			ApplicationStatus: rr.ApplicationStatus,
		})
	}

	// 分页处理
	start := int(offset)
	end := start + int(size)
	if start > len(allRecords) {
		start = len(allRecords)
	}
	if end > len(allRecords) {
		end = len(allRecords)
	}

	var pagedRecords []types.PendingTransactionRecord
	if start < len(allRecords) {
		pagedRecords = allRecords[start:end]
	} else {
		pagedRecords = []types.PendingTransactionRecord{}
	}

	// 批量查询产品信息和客户信息
	productIds := make(map[string]bool)
	customerIds := make(map[string]bool)
	for _, record := range pagedRecords {
		productIds[record.ProductId] = true
		customerIds[record.CustomerId] = true
	}

	// 查询产品信息
	productInfoMap := l.getProductInfos(l.ctx, getKeys(productIds))
	// 查询客户信息
	customerInfoMap := l.getCustomerInfos(l.ctx, getKeys(customerIds))

	// 填充详细信息
	for i := range pagedRecords {
		record := &pagedRecords[i]
		if productInfo, ok := productInfoMap[record.ProductId]; ok {
			record.ProductName = productInfo.ProductName
			record.ProductRiskLevel = productInfo.RiskLevel
		}
		if customerInfo, ok := customerInfoMap[record.CustomerId]; ok {
			record.CustomerName = customerInfo.CustomerName
			record.RiskLevel = customerInfo.RiskLevel
		}
		// 判断风险等级是否匹配
		record.RiskMatch = record.RiskLevel == record.ProductRiskLevel
	}

	response := &types.PendingTransactionResponse{
		Total:   total,
		Records: pagedRecords,
	}

	return &types.BaseResponse{
		Code:    200,
		Message: "查询成功",
		Data:    response,
	}, nil
}

// getProductInfos 批量获取产品信息
func (l *GetPendingTransactionsLogic) getProductInfos(ctx context.Context, productIds []string) map[string]struct {
	ProductName    string
	RiskLevel      string
} {
	if len(productIds) == 0 {
		return make(map[string]struct {
			ProductName string
			RiskLevel   string
		})
	}

	result := make(map[string]struct {
		ProductName string
		RiskLevel   string
	})

	placeholders := strings.Repeat("?,", len(productIds)-1) + "?"
	query := fmt.Sprintf("SELECT product_id, product_name, risk_level FROM product_info WHERE product_id IN (%s) AND deleted_at IS NULL", placeholders)

	args := make([]interface{}, len(productIds))
	for i, id := range productIds {
		args[i] = id
	}

	type ProductRow struct {
		ProductId   string `db:"product_id"`
		ProductName string `db:"product_name"`
		RiskLevel   string `db:"risk_level"`
	}

	var rows []ProductRow
	err := l.svcCtx.Conn.QueryRowsCtx(ctx, &rows, query, args...)
	if err != nil {
		l.Errorf("批量查询产品信息失败: %v", err)
		return result
	}

	for _, row := range rows {
		result[row.ProductId] = struct {
			ProductName string
			RiskLevel   string
		}{
			ProductName: row.ProductName,
			RiskLevel:   row.RiskLevel,
		}
	}

	return result
}

// getCustomerInfos 批量获取客户信息
func (l *GetPendingTransactionsLogic) getCustomerInfos(ctx context.Context, customerIds []string) map[string]struct {
	CustomerName string
	RiskLevel    string
} {
	if len(customerIds) == 0 {
		return make(map[string]struct {
			CustomerName string
			RiskLevel    string
		})
	}

	result := make(map[string]struct {
		CustomerName string
		RiskLevel    string
	})

	placeholders := strings.Repeat("?,", len(customerIds)-1) + "?"
	query := fmt.Sprintf("SELECT customer_id, customer_name, risk_level FROM customer_info WHERE customer_id IN (%s)", placeholders)

	args := make([]interface{}, len(customerIds))
	for i, id := range customerIds {
		args[i] = id
	}

	type CustomerRow struct {
		CustomerId   string `db:"customer_id"`
		CustomerName string `db:"customer_name"`
		RiskLevel    string `db:"risk_level"`
	}

	var rows []CustomerRow
	err := l.svcCtx.Conn.QueryRowsCtx(ctx, &rows, query, args...)
	if err != nil {
		l.Errorf("批量查询客户信息失败: %v", err)
		return result
	}

	for _, row := range rows {
		result[row.CustomerId] = struct {
			CustomerName string
			RiskLevel    string
		}{
			CustomerName: row.CustomerName,
			RiskLevel:    row.RiskLevel,
		}
	}

	return result
}

// getKeys 从 map 中提取键
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
