// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"WMSS/liquidation/api/internal/svc"
	"WMSS/liquidation/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryFailedTransactionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRetryFailedTransactionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryFailedTransactionsLogic {
	return &RetryFailedTransactionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetryFailedTransactionsLogic) RetryFailedTransactions(req *types.RetryFailedTransactionRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
