package role

import (
	"WMSS/user/api/internal/model"
	"context"

	"WMSS/user/api/internal/svc"
	"WMSS/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRolesLogic {
	return &ListRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListRolesLogic) ListRoles(req *types.ListRolesReq) (resp *types.Response, err error) {
	userId, ok := l.ctx.Value("user_id").(string)
	if !ok || userId == "" {
		l.Logger.Errorf("登录过期：用户ID为空或类型错误")
		return &types.Response{
			Code: 401,
			Msg:  "登录过期，请重新登录",
		}, nil
	}

	var roles []*model.SysRole
	var total int64
	roles, total, err = l.svcCtx.RoleRepository.FindByPage(l.ctx, int(req.Page), int(req.PageSize), req.RoleName)
	if err != nil {
		l.Logger.Errorf("查询角色列表失败: %v", err)
		return &types.Response{
			Code: 500,
			Msg:  "查询角色列表失败",
		}, nil
	}
	var rolesInfo []types.RoleInfo
	for _, role := range roles {
		var roleInfo types.RoleInfo
		var permissionIds []string
		permissionIds, err = l.svcCtx.SysRolePermissionModel.FindPermissionsByRoleId(l.ctx, role.RoleId)
		var permissions []model.SysPermission
		permissions, err = l.svcCtx.SysPermissionModel.FindPermissionsByIds(l.ctx, permissionIds)
		if err != nil {
			l.Logger.Errorf("查询角色列表失败: %v", err)
			return &types.Response{
				Code: 500,
				Msg:  "查询角色列表失败",
			}, nil
		}
		var permissionsInfo []types.PermissionInfo
		for _, permission := range permissions {
			var permissionInfo types.PermissionInfo
			permissionInfo = types.PermissionInfo{
				PermissionID:       permission.PermissionId,
				PermissionName:     permission.PermissionName,
				PermissionCode:     permission.PermissionCode,
				PermissionType:     permission.PermissionType,
				ParentPermissionID: permission.PermissionId,
				CreateTime:         permission.CreateTime.Format("2006-01-02 15:04:05"),
				UpdateTime:         permission.UpdateTime.Format("2006-01-02 15:04:05"),
			}
			permissionsInfo = append(permissionsInfo, permissionInfo)
		}
		roleInfo = types.RoleInfo{
			RoleID:      role.RoleId,
			RoleName:    role.RoleName,
			RoleDesc:    role.RoleDesc.String,
			CreateTime:  role.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateTime:  role.UpdateTime.Format("2006-01-02 15:04:05"),
			Permissions: permissionsInfo,
		}
		rolesInfo = append(rolesInfo, roleInfo)
	}
	var listRolesResp types.ListRolesResp
	listRolesResp = types.ListRolesResp{
		Total: total,
		List:  rolesInfo,
	}
	return &types.Response{
		Code: 200,
		Msg:  "查询角色列表成功",
		Data: listRolesResp,
	}, nil
}
