// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"net/http"

	"WMSS/user/api/internal/logic/user"
	"WMSS/user/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取当前用户权限
func GetCurrentUserPermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewGetCurrentUserPermissionsLogic(r.Context(), svcCtx)
		resp, err := l.GetCurrentUserPermissions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
