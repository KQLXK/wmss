// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"WMSS/liquidation/api/internal/svc"
	"WMSS/liquidation/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TransactionConfirmLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTransactionConfirmLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TransactionConfirmLogic {
	return &TransactionConfirmLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TransactionConfirmLogic) TransactionConfirm(req *types.TransactionConfirmRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
