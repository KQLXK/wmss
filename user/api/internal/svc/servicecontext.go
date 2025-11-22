package svc

import (
	"WMSS/user/api/internal/config"
	"WMSS/user/api/internal/middleware"
	"WMSS/user/api/internal/model"
	"WMSS/user/api/internal/repository"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config         config.Config
	DB             *gorm.DB
	AuthMiddleware rest.Middleware

	SysUserModel           model.SysUserModel
	SysRoleModel           model.SysRoleModel
	SysPermissionModel     model.SysPermissionModel
	SysRolePermissionModel model.SysRolePermissionModel
	RoleRepository         repository.RoleRepository
	PermissionRepository   repository.PermissionRepository
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)

	return &ServiceContext{
		Config:         c,
		AuthMiddleware: middleware.NewAuthMiddleware().Handle,

		// 初始化 Model（无缓存版本）
		SysUserModel:           model.NewSysUserModel(conn),
		SysRoleModel:           model.NewSysRoleModel(conn),
		SysPermissionModel:     model.NewSysPermissionModel(conn),
		SysRolePermissionModel: model.NewSysRolePermissionModel(conn),
		RoleRepository:         *repository.NewRoleRepository(conn, model.NewSysRoleModel(conn)),
		PermissionRepository:   *repository.NewPermissionRepository(conn, model.NewSysPermissionModel(conn)),
	}
}
