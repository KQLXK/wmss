package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ CustomerBankCardModel = (*customCustomerBankCardModel)(nil)

type (
	// CustomerBankCardModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCustomerBankCardModel.
	CustomerBankCardModel interface {
		customerBankCardModel
		withSession(session sqlx.Session) CustomerBankCardModel
	}

	customCustomerBankCardModel struct {
		*defaultCustomerBankCardModel
	}
)

// NewCustomerBankCardModel returns a model for the database table.
func NewCustomerBankCardModel(conn sqlx.SqlConn) CustomerBankCardModel {
	return &customCustomerBankCardModel{
		defaultCustomerBankCardModel: newCustomerBankCardModel(conn),
	}
}

func (m *customCustomerBankCardModel) withSession(session sqlx.Session) CustomerBankCardModel {
	return NewCustomerBankCardModel(sqlx.NewSqlConnFromSession(session))
}
