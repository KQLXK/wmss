// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission

import (
	"WMSS/user/api/internal/common"
	"context"

	"WMSS/user/api/internal/svc"
	"WMSS/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeletePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除权限
func NewDeletePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePermissionLogic {
	return &DeletePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePermissionLogic) DeletePermission(req *types.DeletePermissionReq) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	userId, ok := l.ctx.Value("user_id").(string)
	if !ok || userId == "" {
		l.Logger.Errorf("登录过期：用户ID为空或类型错误")
		return &types.Response{
			Code: 401,
			Msg:  "登录过期，请重新登录",
		}, nil
	}

	user, err := l.svcCtx.SysUserModel.FindOne(l.ctx, userId)
	if user == nil || err != nil {
		l.Logger.Errorf("用户[%s]不存在", userId)
		return &types.Response{
			Code: 404,
			Msg:  "用户不存在",
		}, nil
	}

	if user.RoleId != common.SUPER_ADMIN_ROLE_ID {
		l.Logger.Errorf("用户[%s]权限不足", userId)
		return &types.Response{
			Code: 403,
			Msg:  "权限不足，仅超级管理员可删除权限",
		}, nil
	}

	err = l.svcCtx.PermissionRepository.DeleteLogical(l.ctx, req.PermissionID)
	if err != nil {
		l.Logger.Errorf("逻辑删除失败，%v", err)
		return &types.Response{
			Code: 400,
			Msg:  "删除失败",
		}, nil
	}
	return &types.Response{
		Code: 200,
		Msg:  "删除成功",
	}, nil
}
