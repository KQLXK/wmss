package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ CustomerInfoModel = (*customCustomerInfoModel)(nil)

type (
	// CustomerInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCustomerInfoModel.
	CustomerInfoModel interface {
		customerInfoModel
		withSession(session sqlx.Session) CustomerInfoModel
	}

	customCustomerInfoModel struct {
		*defaultCustomerInfoModel
	}
)

// NewCustomerInfoModel returns a model for the database table.
func NewCustomerInfoModel(conn sqlx.SqlConn) CustomerInfoModel {
	return &customCustomerInfoModel{
		defaultCustomerInfoModel: newCustomerInfoModel(conn),
	}
}

func (m *customCustomerInfoModel) withSession(session sqlx.Session) CustomerInfoModel {
	return NewCustomerInfoModel(sqlx.NewSqlConnFromSession(session))
}
