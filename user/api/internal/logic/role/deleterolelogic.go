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

type DeleteRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除角色
func NewDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteRoleLogic) DeleteRole(req *types.DeleteRoleReq) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	userId := l.ctx.Value("user_id").(string)
	var user *model.SysUser
	user, err = l.svcCtx.SysUserModel.FindOne(l.ctx, userId)
	if err != nil {
		l.Logger.Errorf("删除角色失败: %v", err)
		return &types.Response{
			Code: 500,
			Msg:  "删除角色失败",
		}, err
	}
	err = l.svcCtx.SysRoleModel.Delete(l.ctx, req.RoleID)
	if err != nil || user.RoleId != common.SUPER_ADMIN_ROLE_ID {
		l.Logger.Errorf("删除角色失败，权限不足: %v", err)
		return &types.Response{
			Code: 500,
			Msg:  "删除角色失败，权限不足",
		}, err
	}
	l.Logger.Info("删除角色成功,roleId：%s", req.RoleID)
	return &types.Response{
		Code: 200,
		Msg:  "删除角色成功",
	}, err
}
