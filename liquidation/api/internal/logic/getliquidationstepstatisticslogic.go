// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"WMSS/liquidation/api/internal/svc"
	"WMSS/liquidation/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLiquidationStepStatisticsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLiquidationStepStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLiquidationStepStatisticsLogic {
	return &GetLiquidationStepStatisticsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLiquidationStepStatisticsLogic) GetLiquidationStepStatistics() (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
