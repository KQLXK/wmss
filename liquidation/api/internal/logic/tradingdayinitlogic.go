// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"database/sql"
	"time"

	"WMSS/liquidation/api/internal/svc"
	"WMSS/liquidation/api/internal/types"
	"WMSS/liquidation/api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type TradingDayInitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTradingDayInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TradingDayInitLogic {
	return &TradingDayInitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TradingDayInitLogic) TradingDayInit(req *types.TradingDayInitRequest) (resp *types.BaseResponse, err error) {
	now := time.Now()
	var targetDate time.Time

	// 确定目标日期
	if req.TargetDate != "" {
		targetDate, err = time.Parse("2006-01-02", req.TargetDate)
		if err != nil {
			return &types.BaseResponse{
				Code:    400,
				Message: "目标日期格式错误，应为 YYYY-MM-DD",
			}, nil
		}
	} else {
		targetDate = now
	}

	// 查找当前交易日
	currentTradingDay, err := l.svcCtx.WorkCalendarModel.FindCurrentOrNextTradingDay(l.ctx, targetDate)
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

	// 记录清算日志
	logData := &model.LiquidationLog{
		LiquidationDate: currentTradingDay.CalendarDate,
		LiquidationStep: "交易日初始化",
		StepStatus:      "成功",
		StartTime:       sql.NullTime{Time: now, Valid: true},
		EndTime:         sql.NullTime{Time: now, Valid: true},
		ProcessedCount:  1,
		FailureCount:    0,
		OperatorId:      sql.NullString{String: req.OperatorId, Valid: true},
		CreateTime:      now,
	}
	result, err := l.svcCtx.LiquidationLogModel.Insert(l.ctx, logData)
	var logId int64
	if err == nil && result != nil {
		logId, _ = result.LastInsertId()
	}

	response := &types.TradingDayInitResponse{
		PreviousTradingDay: previousDate,
		CurrentTradingDay:  currentTradingDay.CalendarDate.Format("2006-01-02"),
		NextTradingDay:     nextDate,
		InitTime:           now.Format("2006-01-02 15:04:05"),
		LogId:              logId,
	}

	return &types.BaseResponse{
		Code:    200,
		Message: "交易日初始化成功",
		Data:    response,
	}, nil
}
