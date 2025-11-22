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

type GetPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取权限详情
func NewGetPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPermissionLogic {
	return &GetPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPermissionLogic) GetPermission(req *types.GetPermissionReq) (resp *types.Response, err error) {
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
			Msg:  "权限不足，仅超级管理员可查看权限",
		}, nil
	}

	permission, err := l.svcCtx.SysPermissionModel.FindOne(l.ctx, req.PermissionID)
	if err != nil {
		l.Logger.Errorf("查询权限信息失败，%v", err)
		return &types.Response{
			Code: 200,
			Msg:  "查询失败",
		}, nil
	}
	var permissonInfo types.PermissionInfo
	permissonInfo = types.PermissionInfo{
		PermissionID:       permission.PermissionId,
		PermissionName:     permission.PermissionName,
		PermissionCode:     permission.PermissionCode,
		PermissionType:     permission.PermissionType,
		ParentPermissionID: permission.ParentPermissionId.String,
		CreateTime:         permission.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:         permission.UpdateTime.Format("2006-01-02 15:04:05"),
		DeletedAt:          permission.DeletedAt.Time.Format("2006-01-02 15:04:05"),
	}
	return &types.Response{
		Code: 200,
		Msg:  "查询成功",
		Data: permissonInfo,
	}, nil
}
