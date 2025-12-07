package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ WorkCalendarModel = (*customWorkCalendarModel)(nil)

type (
	// WorkCalendarModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWorkCalendarModel.
	WorkCalendarModel interface {
		workCalendarModel
		withSession(session sqlx.Session) WorkCalendarModel
	}

	customWorkCalendarModel struct {
		*defaultWorkCalendarModel
	}
)

// NewWorkCalendarModel returns a model for the database table.
func NewWorkCalendarModel(conn sqlx.SqlConn) WorkCalendarModel {
	return &customWorkCalendarModel{
		defaultWorkCalendarModel: newWorkCalendarModel(conn),
	}
}

func (m *customWorkCalendarModel) withSession(session sqlx.Session) WorkCalendarModel {
	return NewWorkCalendarModel(sqlx.NewSqlConnFromSession(session))
}