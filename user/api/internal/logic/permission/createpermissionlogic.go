// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission

import (
	"WMSS/user/api/internal/common"
	"WMSS/user/api/internal/model"
	"WMSS/user/api/internal/svc"
	"WMSS/user/api/internal/types"
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建权限
func NewCreatePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePermissionLogic {
	return &CreatePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePermissionLogic) CreatePermission(req *types.CreatePermissionReq) (resp *types.Response, err error) {
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
			Msg:  "权限不足，仅超级管理员可新增权限",
		}, nil
	}

	var sysPermission *model.SysPermission
	sysPermission = &model.SysPermission{
		PermissionId:   req.PermissionID,
		PermissionName: req.PermissionName,
		PermissionCode: req.PermissionCode,
		PermissionType: req.PermissionType,
		ParentPermissionId: sql.NullString{
			String: req.ParentPermissionID,
			Valid:  false,
		},
	}
	_, err = l.svcCtx.SysPermissionModel.Insert(l.ctx, sysPermission)
	if err != nil {
		l.Logger.Errorf("新增权限失败：%v", err)
		return &types.Response{
			Code: 500,
			Msg:  fmt.Sprintf("新增权限失败：%v", err),
		}, nil
	}
	l.Logger.Errorf("新增权限成功")
	return &types.Response{
		Code: 200,
		Msg:  "新增权限成功",
	}, nil
}
