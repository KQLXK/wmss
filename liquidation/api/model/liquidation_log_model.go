package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ LiquidationLogModel = (*customLiquidationLogModel)(nil)

type (
	// LiquidationLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customLiquidationLogModel.
	LiquidationLogModel interface {
		liquidationLogModel
		withSession(session sqlx.Session) LiquidationLogModel
	}

	customLiquidationLogModel struct {
		*defaultLiquidationLogModel
	}
)

// NewLiquidationLogModel returns a model for the database table.
func NewLiquidationLogModel(conn sqlx.SqlConn) LiquidationLogModel {
	return &customLiquidationLogModel{
		defaultLiquidationLogModel: newLiquidationLogModel(conn),
	}
}

func (m *customLiquidationLogModel) withSession(session sqlx.Session) LiquidationLogModel {
	return NewLiquidationLogModel(sqlx.NewSqlConnFromSession(session))
}
