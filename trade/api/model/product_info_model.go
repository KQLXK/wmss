package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ProductInfoModel = (*customProductInfoModel)(nil)

type (
	// ProductInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProductInfoModel.
	ProductInfoModel interface {
		productInfoModel
		withSession(session sqlx.Session) ProductInfoModel
	}

	customProductInfoModel struct {
		*defaultProductInfoModel
	}
)

// NewProductInfoModel returns a model for the database table.
func NewProductInfoModel(conn sqlx.SqlConn) ProductInfoModel {
	return &customProductInfoModel{
		defaultProductInfoModel: newProductInfoModel(conn),
	}
}

func (m *customProductInfoModel) withSession(session sqlx.Session) ProductInfoModel {
	return NewProductInfoModel(sqlx.NewSqlConnFromSession(session))
}
