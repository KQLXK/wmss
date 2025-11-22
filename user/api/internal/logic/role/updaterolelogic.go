// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"WMSS/user/api/internal/common"
	"WMSS/user/api/internal/model"
	"context"
	"database/sql"
	"time"

	"WMSS/user/api/internal/svc"
	"WMSS/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新角色信息
func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRoleLogic) UpdateRole(req *types.UpdateRoleReq) (resp *types.Response, err error) {
	userId, ok := l.ctx.Value("user_id").(string)
	if !ok || userId == "" {
		l.Logger.Errorf("登录过期：用户ID为空或类型错误")
		return &types.Response{
			Code: 401,
			Msg:  "登录过期，请重新登录",
		}, nil
	}

	user, err := l.svcCtx.SysUserModel.FindOne(l.ctx, userId)
	if err != nil {
		l.Logger.Errorf("查询用户[%s]失败: %v", userId, err)
		return &types.Response{
			Code: 500,
			Msg:  "系统错误，查询用户信息失败",
		}, err
	}
	if user == nil {
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

	if req.RoleID == "" {
		l.Logger.Errorf("更新角色失败：角色ID为空")
		return &types.Response{
			Code: 400,
			Msg:  "角色ID不能为空",
		}, nil
	}

	// 5. 查询要更新的角色是否存在
	targetRole, err := l.svcCtx.SysRoleModel.FindOne(l.ctx, req.RoleID)
	if err != nil {
		l.Logger.Errorf("查询角色[%s]失败: %v", req.RoleID, err)
		return &types.Response{
			Code: 500,
			Msg:  "系统错误，查询角色信息失败",
		}, err
	}
	if targetRole == nil {
		l.Logger.Errorf("角色[%s]不存在，无法更新", req.RoleID)
		return &types.Response{
			Code: 404,
			Msg:  "目标角色不存在",
		}, nil
	}

	// 6. 构造更新数据（仅更新非空字段，适配 sql.NullString）
	updateData := &model.SysRole{
		RoleId:     req.RoleID,            // 角色ID为查询条件，必须传入
		RoleName:   targetRole.RoleName,   // 默认沿用原有值
		RoleDesc:   targetRole.RoleDesc,   // 默认沿用原有值
		CreateTime: targetRole.CreateTime, // 创建时间不允许修改
		UpdateTime: time.Now(),            // 更新时间设为当前时间
	}

	// 6.1 角色名称：仅当请求中传入非空值时更新
	if req.RoleName != "" {
		updateData.RoleName = req.RoleName
	}

	// 6.2 角色描述：适配 sql.NullString（区分「传空字符串」和「未传值」）
	if req.RoleDesc != "" {
		// 传入非空描述：设为有效字符串
		updateData.RoleDesc = sql.NullString{
			String: req.RoleDesc,
			Valid:  true,
		}
	} else if req.RoleDesc == "" && req.RoleDesc != targetRole.RoleDesc.String {
		// 传入空字符串：明确要清空描述（Valid 设为 true，String 为空）
		updateData.RoleDesc = sql.NullString{
			String: "",
			Valid:  true,
		}
	}
	// 若 req.RoleDesc 未传（omitempty），则沿用原有 RoleDesc，无需修改

	// 7. 执行更新操作
	err = l.svcCtx.SysRoleModel.Update(l.ctx, updateData)
	if err != nil {
		l.Logger.Errorf("更新角色[%s]失败: %v", req.RoleID, err)
		return &types.Response{
			Code: 500,
			Msg:  "系统错误，更新角色信息失败",
		}, err
	}

	// 8. 成功响应
	l.Logger.Infof("用户[%s]成功更新角色[%s]，更新内容：RoleName=%s, RoleDesc=%s",
		userId, req.RoleID, updateData.RoleName, updateData.RoleDesc.String)
	return &types.Response{
		Code: 200,
		Msg:  "角色更新成功",
	}, nil
}
