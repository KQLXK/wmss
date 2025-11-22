// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"WMSS/user/api/internal/common"
	"WMSS/user/api/internal/model"
	"context"

	"WMSS/user/api/internal/svc"
	"WMSS/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssignRolePermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 分配角色权限
func NewAssignRolePermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRolePermissionsLogic {
	return &AssignRolePermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssignRolePermissionsLogic) AssignRolePermissions(req *types.AssignRolePermissionsReq) (resp *types.Response, err error) {
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
		l.Logger.Errorf("用户[%s]权限不足，尝试更新角色[%s]", userId, req.RoleID)
		return &types.Response{
			Code: 403,
			Msg:  "权限不足，仅超级管理员可更新角色",
		}, nil
	}
	var data []*model.SysRolePermission
	for _, permissionId := range req.PermissionIDs {
		var permission model.SysRolePermission
		permission.PermissionId = permissionId
		permission.RoleId = req.RoleID
		data = append(data, &permission)
	}
	_, err = l.svcCtx.SysRolePermissionModel.InsertBatch(l.ctx, data)
	if err != nil {
		l.Logger.Errorf("分配权限失败")
		return &types.Response{
			Code: 500,
			Msg:  "分配权限失败",
		}, nil
	}
	return &types.Response{
		Code: 200,
		Msg:  "分配权限成功",
	}, err
}
