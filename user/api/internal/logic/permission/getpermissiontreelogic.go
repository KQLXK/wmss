// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission

import (
	"context"

	"WMSS/user/api/internal/svc"
	"WMSS/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPermissionTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取权限树
func NewGetPermissionTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPermissionTreeLogic {
	return &GetPermissionTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPermissionTreeLogic) GetPermissionTree(req *types.GetPermissionTreeReq) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
