// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission

import (
	"WMSS/user/api/internal/common"
	"WMSS/user/api/internal/model"
	"context"

	"WMSS/user/api/internal/svc"
	"WMSS/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取权限列表
func NewListPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPermissionsLogic {
	return &ListPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPermissionsLogic) ListPermissions(req *types.ListPermissionsReq) (resp *types.Response, err error) {
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
			Msg:  "权限不足，仅超级管理员可查询权限",
		}, nil
	}

	// 数据库查询
	var sysPermissions []*model.SysPermission
	var sum int64
	sysPermissions, sum, err = l.svcCtx.PermissionRepository.FindByPage(l.ctx, int(req.Page), int(req.PageSize), req.PermissionType, req.PermissionName)
	if err != nil {
		l.Logger.Errorf("数据库查询失败，%v", err)
		return &types.Response{
			Code: 400,
			Msg:  "查询失败",
		}, nil
	}

	// 数据转换
	var permissionsInfo []types.PermissionInfo
	for _, sysPermission := range sysPermissions {
		permissionsInfo = append(permissionsInfo, types.PermissionInfo{
			PermissionID:       sysPermission.PermissionId,
			PermissionName:     sysPermission.PermissionName,
			PermissionCode:     sysPermission.PermissionCode,
			PermissionType:     sysPermission.PermissionType,
			ParentPermissionID: sysPermission.ParentPermissionId.String,
			CreateTime:         sysPermission.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateTime:         sysPermission.UpdateTime.Format("2006-01-02 15:04:05"),
			DeletedAt:          sysPermission.DeletedAt.Time.Format("2006-01-02 15:04:05"),
		})
	}
	return &types.Response{
		Code: 200,
		Msg:  "查询成功",
		Data: &types.ListPermissionsResp{
			Total: sum,
			List:  permissionsInfo,
		},
	}, nil
}
