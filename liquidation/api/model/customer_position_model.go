package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ CustomerPositionModel = (*customCustomerPositionModel)(nil)

type (
	// CustomerPositionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCustomerPositionModel.
	CustomerPositionModel interface {
		customerPositionModel
		withSession(session sqlx.Session) CustomerPositionModel
	}

	customCustomerPositionModel struct {
		*defaultCustomerPositionModel
	}
)

// NewCustomerPositionModel returns a model for the database table.
func NewCustomerPositionModel(conn sqlx.SqlConn) CustomerPositionModel {
	return &customCustomerPositionModel{
		defaultCustomerPositionModel: newCustomerPositionModel(conn),
	}
}

func (m *customCustomerPositionModel) withSession(session sqlx.Session) CustomerPositionModel {
	return NewCustomerPositionModel(sqlx.NewSqlConnFromSession(session))
}
