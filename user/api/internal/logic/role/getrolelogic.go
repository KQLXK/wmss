// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"WMSS/user/api/internal/model"
	"context"

	"WMSS/user/api/internal/svc"
	"WMSS/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取角色详情
func NewGetRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleLogic {
	return &GetRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRoleLogic) GetRole(req *types.GetRoleReq) (resp *types.Response, err error) {

	var sysRole *model.SysRole
	sysRole, err = l.svcCtx.SysRoleModel.FindOne(l.ctx, req.RoleID)
	if err != nil {
		l.Logger.Errorf("查询角色失败: %v", err)
		return &types.Response{
			Code: 500,
			Msg:  "查询角色失败",
		}, err
	}
	var sysPermissionsIds []string
	sysPermissionsIds, err = l.svcCtx.SysRolePermissionModel.FindPermissionsByRoleId(l.ctx, req.RoleID)
	var sysPermissions []model.SysPermission
	sysPermissions, err = l.svcCtx.SysPermissionModel.FindPermissionsByIds(l.ctx, sysPermissionsIds)
	if err != nil {
		l.Logger.Errorf("查询角色失败: %v", err)
		return &types.Response{
			Code: 500,
			Msg:  "查询角色失败",
		}, err
	}
	var permissions []types.PermissionInfo
	for _, sysPermission := range sysPermissions {
		permissions = append(permissions, types.PermissionInfo{
			PermissionID:       sysPermission.PermissionId,
			PermissionName:     sysPermission.PermissionName,
			PermissionCode:     sysPermission.PermissionCode,
			PermissionType:     sysPermission.PermissionType,
			ParentPermissionID: sysPermission.PermissionId,
			CreateTime:         sysPermission.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateTime:         sysPermission.UpdateTime.Format("2006-01-02 15:04:05"),
		})
	}
	return &types.Response{
		Code: 200,
		Msg:  "查询成功",
		Data: types.RoleInfo{
			RoleID:      sysRole.RoleId,
			RoleName:    sysRole.RoleName,
			RoleDesc:    sysRole.RoleDesc.String,
			CreateTime:  sysRole.CreateTime.String(),
			UpdateTime:  sysRole.UpdateTime.String(),
			Permissions: permissions,
		},
	}, nil
}
