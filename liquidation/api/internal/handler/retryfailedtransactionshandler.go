// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"WMSS/liquidation/api/internal/logic"
	"WMSS/liquidation/api/internal/svc"
	"WMSS/liquidation/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func RetryFailedTransactionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RetryFailedTransactionRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewRetryFailedTransactionsLogic(r.Context(), svcCtx)
		resp, err := l.RetryFailedTransactions(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
