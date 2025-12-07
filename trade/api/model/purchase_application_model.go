package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ PurchaseApplicationModel = (*customPurchaseApplicationModel)(nil)

type (
	// PurchaseApplicationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPurchaseApplicationModel.
	PurchaseApplicationModel interface {
		purchaseApplicationModel
		withSession(session sqlx.Session) PurchaseApplicationModel
	}

	customPurchaseApplicationModel struct {
		*defaultPurchaseApplicationModel
	}
)

// NewPurchaseApplicationModel returns a model for the database table.
func NewPurchaseApplicationModel(conn sqlx.SqlConn) PurchaseApplicationModel {
	return &customPurchaseApplicationModel{
		defaultPurchaseApplicationModel: newPurchaseApplicationModel(conn),
	}
}

func (m *customPurchaseApplicationModel) withSession(session sqlx.Session) PurchaseApplicationModel {
	return NewPurchaseApplicationModel(sqlx.NewSqlConnFromSession(session))
}
