// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"WMSS/liquidation/api/internal/logic"
	"WMSS/liquidation/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetLiquidationStepStatisticsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetLiquidationStepStatisticsLogic(r.Context(), svcCtx)
		resp, err := l.GetLiquidationStepStatistics()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
