package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ProductNetValueModel = (*customProductNetValueModel)(nil)

type (
	// ProductNetValueModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProductNetValueModel.
	ProductNetValueModel interface {
		productNetValueModel
		withSession(session sqlx.Session) ProductNetValueModel
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