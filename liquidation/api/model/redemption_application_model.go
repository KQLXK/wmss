package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ RedemptionApplicationModel = (*customRedemptionApplicationModel)(nil)

type (
	// RedemptionApplicationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRedemptionApplicationModel.
	RedemptionApplicationModel interface {
		redemptionApplicationModel
		withSession(session sqlx.Session) RedemptionApplicationModel
	}

	customRedemptionApplicationModel struct {
		*defaultRedemptionApplicationModel
	}
)

// NewRedemptionApplicationModel returns a model for the database table.
func NewRedemptionApplicationModel(conn sqlx.SqlConn) RedemptionApplicationModel {
	return &customRedemptionApplicationModel{
		defaultRedemptionApplicationModel: newRedemptionApplicationModel(conn),
	}
}

func (m *customRedemptionApplicationModel) withSession(session sqlx.Session) RedemptionApplicationModel {
	return NewRedemptionApplicationModel(sqlx.NewSqlConnFromSession(session))
}
