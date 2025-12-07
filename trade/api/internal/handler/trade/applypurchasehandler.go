// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"net/http"

	"WMSS/trade/api/internal/logic/trade"
	"WMSS/trade/api/internal/svc"
	"WMSS/trade/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ApplyPurchaseHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PurchaseRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := trade.NewApplyPurchaseLogic(r.Context(), svcCtx)
		resp, err := l.ApplyPurchase(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
