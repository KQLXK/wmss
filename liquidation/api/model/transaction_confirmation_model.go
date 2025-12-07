package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TransactionConfirmationModel = (*customTransactionConfirmationModel)(nil)

type (
	// TransactionConfirmationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTransactionConfirmationModel.
	TransactionConfirmationModel interface {
		transactionConfirmationModel
		withSession(session sqlx.Session) TransactionConfirmationModel
	}

	customTransactionConfirmationModel struct {
		*defaultTransactionConfirmationModel
	}
)

// NewTransactionConfirmationModel returns a model for the database table.
func NewTransactionConfirmationModel(conn sqlx.SqlConn) TransactionConfirmationModel {
	return &customTransactionConfirmationModel{
		defaultTransactionConfirmationModel: newTransactionConfirmationModel(conn),
	}
}

func (m *customTransactionConfirmationModel) withSession(session sqlx.Session) TransactionConfirmationModel {
	return NewTransactionConfirmationModel(sqlx.NewSqlConnFromSession(session))
}
