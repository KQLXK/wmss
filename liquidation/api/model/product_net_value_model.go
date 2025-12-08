package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ProductNetValueModel = (*customProductNetValueModel)(nil)

type (
	// ProductNetValueModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProductNetValueModel.
	ProductNetValueModel interface {
		productNetValueModel
		withSession(session sqlx.Session) ProductNetValueModel
		FindLatestBeforeDate(ctx context.Context, productId string, date time.Time) (*ProductNetValue, error)
	}

	customProductNetValueModel struct {
		*defaultProductNetValueModel
	}
)

// NewProductNetValueModel returns a model for the database table.
func NewProductNetValueModel(conn sqlx.SqlConn) ProductNetValueModel {
	return &customProductNetValueModel{
		defaultProductNetValueModel: newProductNetValueModel(conn),
	}
}

func (m *customProductNetValueModel) withSession(session sqlx.Session) ProductNetValueModel {
	return NewProductNetValueModel(sqlx.NewSqlConnFromSession(session))
}

// FindLatestBeforeDate 查找指定日期前最新的净值记录
func (m *customProductNetValueModel) FindLatestBeforeDate(ctx context.Context, productId string, date time.Time) (*ProductNetValue, error) {
	dateStr := date.Format("2006-01-02")

	var result ProductNetValue
	query := `
		SELECT * FROM product_net_value 
		WHERE product_id = ? AND stat_date <= ? 
		ORDER BY stat_date DESC 
		LIMIT 1
	`
	err := m.conn.QueryRowCtx(ctx, &result, query, productId, dateStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, err
	}

	return &result, nil
}
