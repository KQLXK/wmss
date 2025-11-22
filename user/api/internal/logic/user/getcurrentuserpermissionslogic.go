// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"WMSS/user/api/internal/svc"
	"WMSS/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCurrentUserPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取当前用户权限
func NewGetCurrentUserPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentUserPermissionsLogic {
	return &GetCurrentUserPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCurrentUserPermissionsLogic) GetCurrentUserPermissions() (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
