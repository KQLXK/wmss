// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"net/http"

	"WMSS/trade/api/internal/logic/trade"
	"WMSS/trade/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetTransactionDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := trade.NewGetTransactionDetailLogic(r.Context(), svcCtx)
		resp, err := l.GetTransactionDetail()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
