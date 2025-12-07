// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"time"

	"WMSS/liquidation/api/internal/svc"
	"WMSS/liquidation/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCurrentTradingDayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCurrentTradingDayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentTradingDayLogic {
	return &GetCurrentTradingDayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCurrentTradingDayLogic) GetCurrentTradingDay() (resp *types.BaseResponse, err error) {
	now := time.Now()
	
	// 查找当前交易日
	currentTradingDay, err := l.svcCtx.WorkCalendarModel.FindCurrentOrNextTradingDay(l.ctx, now)
	if err != nil {
		return &types.BaseResponse{
			Code:    500,
			Message: "查找交易日失败: " + err.Error(),
		}, nil
	}

	// 查找上一个交易日
	previousTradingDay, err := l.svcCtx.WorkCalendarModel.FindPreviousTradingDay(l.ctx, currentTradingDay.CalendarDate)
	previousDate := ""
	if err == nil {
		previousDate = previousTradingDay.CalendarDate.Format("2006-01-02")
	}

	// 查找下一个交易日
	nextTradingDay, err := l.svcCtx.WorkCalendarModel.FindNextTradingDay(l.ctx, currentTradingDay.CalendarDate)
	nextDate := ""
	if err == nil {
		nextDate = nextTradingDay.CalendarDate.Format("2006-01-02")
	}

	response := &types.TradingDayResponse{
		TradingDay:         currentTradingDay.CalendarDate.Format("2006-01-02"),
		IsWorkday:          currentTradingDay.IsWorkday == 1,
		WorkdayType:        currentTradingDay.WorkdayType,
		PreviousTradingDay: previousDate,
		NextTradingDay:     nextDate,
	}

	return &types.BaseResponse{
		Code:    200,
		Message: "获取成功",
		Data:    response,
	}, nil
}
