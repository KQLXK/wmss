// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"WMSS/trade/api/internal/svc"
	"WMSS/trade/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTransactionDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTransactionDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTransactionDetailLogic {
	return &GetTransactionDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTransactionDetailLogic) GetTransactionDetail() (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
