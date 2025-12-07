// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"WMSS/liquidation/api/model"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"WMSS/liquidation/api/internal/svc"
	"WMSS/liquidation/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestNetValuesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLatestNetValuesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestNetValuesLogic {
	return &GetLatestNetValuesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLatestNetValuesLogic) GetLatestNetValues(req *types.NetValueQueryRequest) (resp *types.BaseResponse, err error) {
	// 参数处理
	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size

	// 解析日期
	var startDate, endDate time.Time
	if req.StartDate != "" {
		startDate, err = time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return &types.BaseResponse{
				Code:    400,
				Message: "开始日期格式错误，应为 YYYY-MM-DD",
			}, nil
		}
	}
	if req.EndDate != "" {
		endDate, err = time.Parse("2006-01-02", req.EndDate)
		if endDate.Hour() == 0 && endDate.Minute() == 0 && endDate.Second() == 0 {
			// 如果是日期，设置为当天的最后一刻
			endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())
		}
		if err != nil {
			return &types.BaseResponse{
				Code:    400,
				Message: "结束日期格式错误，应为 YYYY-MM-DD",
			}, nil
		}
	}

	// 查询净值数据
	netValues, total, err := l.queryNetValues(req.ProductId, startDate, endDate, int(offset), int(size))
	if err != nil {
		l.Errorf("查询净值失败: %v", err)
		return &types.BaseResponse{
			Code:    500,
			Message: "查询净值失败: " + err.Error(),
		}, nil
	}

	// 如果没有数据
	if len(netValues) == 0 {
		return &types.BaseResponse{
			Code:    200,
			Message: "查询成功",
			Data: &types.NetValuePageResponse{
				Total:   0,
				Records: []types.NetValueRecord{},
			},
		}, nil
	}

	// 收集所有需要查询的产品ID
	productIdSet := make(map[string]bool)
	for _, nv := range netValues {
		productIdSet[nv.ProductId] = true
	}

	// 批量查询产品名称
	productIds := make([]string, 0, len(productIdSet))
	for id := range productIdSet {
		productIds = append(productIds, id)
	}
	productIdMap := l.getProductNames(l.ctx, productIds)

	// 转换数据并计算增长率
	records := make([]types.NetValueRecord, 0, len(netValues))
	for _, nv := range netValues {
		// 查询前一日净值用于计算增长率
		previousNetValue := l.getPreviousNetValue(l.ctx, nv.ProductId, nv.StatDate)

		// 计算增长率
		dailyGrowthRate := 0.0
		if previousNetValue > 0 {
			dailyGrowthRate = ((nv.UnitNetValue - previousNetValue) / previousNetValue) * 100
		}

		// 计算增长金额
		growthAmount := nv.UnitNetValue - previousNetValue

		record := types.NetValueRecord{
			ProductId:          nv.ProductId,
			ProductName:        productIdMap[nv.ProductId],
			StatDate:           nv.StatDate.Format("2006-01-02"),
			UnitNetValue:       nv.UnitNetValue,
			CumulativeNetValue: nv.CumulativeNetValue,
			DailyGrowthRate:    dailyGrowthRate,
			PreviousNetValue:   previousNetValue,
			GrowthAmount:       growthAmount,
		}

		// 如果数据库中有日增长率，优先使用数据库的值
		if nv.DailyGrowthRate.Valid {
			record.DailyGrowthRate = nv.DailyGrowthRate.Float64
		}

		records = append(records, record)
	}

	response := &types.NetValuePageResponse{
		Total:   total,
		Records: records,
	}

	return &types.BaseResponse{
		Code:    200,
		Message: "查询成功",
		Data:    response,
	}, nil
}

// queryNetValues 查询净值数据
func (l *GetLatestNetValuesLogic) queryNetValues(productId string, startDate, endDate time.Time, offset, size int) ([]*model.ProductNetValue, int64, error) {
	// 构建查询条件
	whereClause := "WHERE 1=1"
	args := make([]interface{}, 0)

	l.Infof("构建查询条件: productId='%s'", productId)

	// 当 productId 不为空时，只查询该产品
	if productId != "" {
		whereClause += " AND product_id = ?"
		args = append(args, productId)
		l.Infof("添加产品ID条件: %s", productId)
	} else {
		l.Infof("产品ID为空，查询所有产品")
	}

	// 日期条件 - 注意：参数必须与占位符匹配
	if !startDate.IsZero() {
		whereClause += " AND stat_date >= ?"
		args = append(args, startDate.Format("2006-01-02"))
		l.Infof("添加开始日期条件: %s", startDate.Format("2006-01-02"))
	}

	if !endDate.IsZero() {
		whereClause += " AND stat_date <= ?"
		args = append(args, endDate.Format("2006-01-02"))
		l.Infof("添加结束日期条件: %s", endDate.Format("2006-01-02"))
	}

	// 构建完整的查询语句
	baseQuery := "SELECT * FROM product_net_value " + whereClause

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM product_net_value " + whereClause
	l.Infof("总数查询SQL: %s", countQuery)
	l.Infof("总数查询参数: %v (数量: %d)", args, len(args))

	// 如果没有任何条件，可能会有问题，但 1=1 应该能处理
	var total int64
	if len(args) > 0 {
		err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &total, countQuery, args...)
		if err != nil {
			l.Errorf("查询总数失败: %v", err)
			return nil, 0, err
		}
	} else {
		// 如果没有参数，直接执行查询
		err := l.svcCtx.Conn.QueryRowCtx(l.ctx, &total, countQuery)
		if err != nil {
			l.Errorf("查询总数失败（无参数）: %v", err)
			return nil, 0, err
		}
	}

	l.Infof("查询到总数: %d", total)

	// 查询数据（按日期倒序排列）
	dataQuery := baseQuery + " ORDER BY stat_date DESC, product_id LIMIT ? OFFSET ?"

	// 注意：LIMIT 和 OFFSET 需要额外的参数
	dataArgs := make([]interface{}, 0)
	if len(args) > 0 {
		// 如果有条件参数，先添加它们
		dataArgs = append(dataArgs, args...)
	}
	// 然后添加分页参数
	dataArgs = append(dataArgs, size, offset)

	l.Infof("数据查询SQL: %s", dataQuery)
	l.Infof("数据查询参数: %v (数量: %d)", dataArgs, len(dataArgs))

	type NetValueRow struct {
		ProductId          string          `db:"product_id"`
		StatDate           time.Time       `db:"stat_date"`
		UnitNetValue       float64         `db:"unit_net_value"`
		CumulativeNetValue float64         `db:"cumulative_net_value"`
		DailyGrowthRate    sql.NullFloat64 `db:"daily_growth_rate"`
		CreateTime         time.Time       `db:"create_time"`
	}

	var rows []NetValueRow

	// 根据参数数量执行查询
	if len(dataArgs) > 0 {
		err := l.svcCtx.Conn.QueryRowsCtx(l.ctx, &rows, dataQuery, dataArgs...)
		if err != nil {
			l.Errorf("查询数据失败: %v", err)
			return nil, 0, err
		}
	} else {
		// 理论上不会到这里，因为至少有分页参数
		err := l.svcCtx.Conn.QueryRowsCtx(l.ctx, &rows, dataQuery)
		if err != nil {
			l.Errorf("查询数据失败（无参数）: %v", err)
			return nil, 0, err
		}
	}

	l.Infof("查询到 %d 条记录", len(rows))

	// 转换为 ProductNetValue
	netValues := make([]*model.ProductNetValue, len(rows))
	for i, row := range rows {
		netValues[i] = &model.ProductNetValue{
			ProductId:          row.ProductId,
			StatDate:           row.StatDate,
			UnitNetValue:       row.UnitNetValue,
			CumulativeNetValue: row.CumulativeNetValue,
			DailyGrowthRate:    row.DailyGrowthRate,
			CreateTime:         row.CreateTime,
		}
		l.Infof("记录 %d: 产品ID=%s, 日期=%s, 净值=%f",
			i, row.ProductId, row.StatDate.Format("2006-01-02"), row.UnitNetValue)
	}

	return netValues, total, nil
}

// getProductNames 批量获取产品名称
func (l *GetLatestNetValuesLogic) getProductNames(ctx context.Context, productIds []string) map[string]string {
	if len(productIds) == 0 {
		return make(map[string]string)
	}

	result := make(map[string]string)

	// 构建查询，使用 IN 子句批量查询
	placeholders := strings.Repeat("?,", len(productIds)-1) + "?"
	// 修改：去掉 AND `deleted_at` IS NULL 条件
	query := fmt.Sprintf("SELECT `product_id`, `product_name` FROM `product_info` WHERE `product_id` IN (%s)", placeholders)

	args := make([]interface{}, len(productIds))
	for i, id := range productIds {
		args[i] = id
	}

	// 使用 ServiceContext 中的数据库连接查询
	type ProductNameRow struct {
		ProductId   string `db:"product_id"`
		ProductName string `db:"product_name"`
	}

	var rows []ProductNameRow
	err := l.svcCtx.Conn.QueryRowsCtx(ctx, &rows, query, args...)
	if err != nil {
		l.Errorf("批量查询产品名称失败: %v", err)
		// 如果批量查询失败，回退到单个查询
		for _, id := range productIds {
			name := l.getProductNameSingle(ctx, id)
			if name != "" {
				result[id] = name
			}
		}
		return result
	}

	for _, row := range rows {
		result[row.ProductId] = row.ProductName
	}

	return result
}

// getProductNameSingle 单个查询产品名称（备用方法）
func (l *GetLatestNetValuesLogic) getProductNameSingle(ctx context.Context, productId string) string {
	var productName string
	// 修改：去掉 AND `deleted_at` IS NULL 条件
	query := "SELECT `product_name` FROM `product_info` WHERE `product_id` = ? LIMIT 1"

	err := l.svcCtx.Conn.QueryRowCtx(ctx, &productName, query, productId)
	if err != nil {
		l.Errorf("查询产品名称失败: productId=%s, error=%v", productId, err)
		return ""
	}
	return productName
}

// getPreviousNetValue 获取前一日净值
func (l *GetLatestNetValuesLogic) getPreviousNetValue(ctx context.Context, productId string, currentDate time.Time) float64 {
	var previousNetValue sql.NullFloat64
	query := "SELECT `unit_net_value` FROM `product_net_value` WHERE `product_id` = ? AND `stat_date` < ? ORDER BY `stat_date` DESC LIMIT 1"

	err := l.svcCtx.Conn.QueryRowCtx(ctx, &previousNetValue, query, productId, currentDate)
	if err != nil {
		// 如果查询失败（可能没有前一日数据），返回0
		return 0
	}

	if previousNetValue.Valid {
		return previousNetValue.Float64
	}
	return 0
}
