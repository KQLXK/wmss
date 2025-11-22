// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"WMSS/user/api/internal/common"
	"WMSS/user/api/internal/model"
	"context"
	"time"

	"WMSS/user/api/internal/svc"
	"WMSS/user/api/internal/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.Response, err error) {
	var sysUser *model.SysUser
	sysUser, err = l.svcCtx.SysUserModel.FindOneByUserName(l.ctx, req.UserName)
	if err != nil {
		l.Logger.Errorf("登录失败: %v", err)
		return &types.Response{
			Code: 500,
			Msg:  "登录失败",
		}, err
	}
	if !common.VerifyPassword(req.Password, sysUser.Password) {
		l.Logger.Errorf("密码错误")
		return &types.Response{
			Code: 400,
			Msg:  "密码错误",
		}, nil
	}

	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	accessSecret := l.svcCtx.Config.Auth.AccessSecret

	claims := make(jwt.MapClaims)
	claims["exp"] = now + accessExpire
	claims["iat"] = now
	claims["user_id"] = sysUser.UserId
	claims["username"] = sysUser.UserName

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	tokenString, err := token.SignedString([]byte(accessSecret))
	if err != nil {
		return nil, err
	}

	return &types.Response{
		Code: 200,
		Msg:  "登录成功",
		Data: tokenString,
	}, nil
}
