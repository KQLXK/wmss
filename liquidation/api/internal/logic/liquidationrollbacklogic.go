// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"WMSS/liquidation/api/internal/svc"
	"WMSS/liquidation/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LiquidationRollbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLiquidationRollbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LiquidationRollbackLogic {
	return &LiquidationRollbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LiquidationRollbackLogic) LiquidationRollback(req *types.LiquidationRollbackRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
