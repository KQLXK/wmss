// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"net/http"

	"WMSS/user/api/internal/logic/role"
	"WMSS/user/api/internal/svc"
	"WMSS/user/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取角色列表
func ListRolesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListRolesReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := role.NewListRolesLogic(r.Context(), svcCtx)
		resp, err := l.ListRoles(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
