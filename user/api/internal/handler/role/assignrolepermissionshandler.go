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

// 分配角色权限
func AssignRolePermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AssignRolePermissionsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := role.NewAssignRolePermissionsLogic(r.Context(), svcCtx)
		resp, err := l.AssignRolePermissions(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
