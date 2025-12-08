package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WorkCalendarModel = (*customWorkCalendarModel)(nil)

type (
	// WorkCalendarModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWorkCalendarModel.
	WorkCalendarModel interface {
		workCalendarModel
		withSession(session sqlx.Session) WorkCalendarModel
		FindCurrentOrNextTradingDay(ctx context.Context, now time.Time) (*WorkCalendar, error)
		FindPreviousTradingDay(ctx context.Context, date time.Time) (*WorkCalendar, error)
		FindNextTradingDay(ctx context.Context, date time.Time) (*WorkCalendar, error)
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

// FindCurrentOrNextTradingDay 查找当前或下一个交易日
func (m *customWorkCalendarModel) FindCurrentOrNextTradingDay(ctx context.Context, now time.Time) (*WorkCalendar, error) {
	today := now.Format("2006-01-02")

	// 先查询今天是否是交易日
	var result WorkCalendar
	query := "SELECT * FROM work_calendar WHERE calendar_date = ? AND is_trading_day = 1 LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &result, query, today)
	if err == nil {
		return &result, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// 如果今天不是交易日，查找下一个交易日
	nextQuery := `
		SELECT * FROM work_calendar 
		WHERE calendar_date > ? AND is_trading_day = 1 
		ORDER BY calendar_date ASC 
		LIMIT 1
	`
	err = m.conn.QueryRowCtx(ctx, &result, nextQuery, today)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// FindPreviousTradingDay 查找指定日期前最近的交易日
func (m *customWorkCalendarModel) FindPreviousTradingDay(ctx context.Context, date time.Time) (*WorkCalendar, error) {
	dateStr := date.Format("2006-01-02")

	var result WorkCalendar
	query := `
		SELECT * FROM work_calendar 
		WHERE calendar_date < ? AND is_trading_day = 1 
		ORDER BY calendar_date DESC 
		LIMIT 1
	`
	err := m.conn.QueryRowCtx(ctx, &result, query, dateStr)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// FindNextTradingDay 查找指定日期后最近的交易日
func (m *customWorkCalendarModel) FindNextTradingDay(ctx context.Context, date time.Time) (*WorkCalendar, error) {
	dateStr := date.Format("2006-01-02")

	var result WorkCalendar
	query := `
		SELECT * FROM work_calendar 
		WHERE calendar_date > ? AND is_trading_day = 1 
		ORDER BY calendar_date ASC 
		LIMIT 1
	`
	err := m.conn.QueryRowCtx(ctx, &result, query, dateStr)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
