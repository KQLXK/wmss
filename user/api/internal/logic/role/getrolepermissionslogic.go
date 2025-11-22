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

type GetRolePermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取角色权限列表
func NewGetRolePermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRolePermissionsLogic {
	return &GetRolePermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRolePermissionsLogic) GetRolePermissions(req *types.GetRolePermissionsReq) (resp *types.Response, err error) {
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

	var permissionsIds []string
	permissionsIds, err = l.svcCtx.SysRolePermissionModel.FindPermissionsByRoleId(l.ctx, req.RoleID)
	if err != nil {
		l.Logger.Errorf("查询permissionsIds失败：%v", err)
		return &types.Response{
			Code: 500,
			Msg:  "查询permissionsIds失败",
		}, nil
	}
	var permissions []model.SysPermission
	permissions, err = l.svcCtx.SysPermissionModel.FindPermissionsByIds(l.ctx, permissionsIds)
	if err != nil {
		l.Logger.Errorf("查询permissions失败：%v", err)
		return &types.Response{
			Code: 500,
			Msg:  "查询permissions失败",
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
	return &types.Response{
		Code: 200,
		Msg:  "查询权限成功",
		Data: types.ListPermissionsResp{
			Total: int64(len(permissions)),
			List:  permissionsInfo,
		},
	}, nil
}
