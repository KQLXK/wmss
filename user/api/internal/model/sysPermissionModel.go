package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SysPermissionModel = (*customSysPermissionModel)(nil)

type (
	// SysPermissionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSysPermissionModel.
	SysPermissionModel interface {
		sysPermissionModel
		withSession(session sqlx.Session) SysPermissionModel
	}

	customSysPermissionModel struct {
		*defaultSysPermissionModel
	}
)

// NewSysPermissionModel returns a model for the database table.
func NewSysPermissionModel(conn sqlx.SqlConn) SysPermissionModel {
	return &customSysPermissionModel{
		defaultSysPermissionModel: newSysPermissionModel(conn),
	}
}

func (m *customSysPermissionModel) withSession(session sqlx.Session) SysPermissionModel {
	return NewSysPermissionModel(sqlx.NewSqlConnFromSession(session))
}
