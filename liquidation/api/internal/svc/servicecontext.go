// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"WMSS/liquidation/api/internal/config"
	"WMSS/liquidation/api/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

	// 数据库连接
	Conn sqlx.SqlConn

	// 数据库模型
	WorkCalendarModel            model.WorkCalendarModel
	ProductNetValueModel         model.ProductNetValueModel
	TransactionConfirmationModel model.TransactionConfirmationModel
	LiquidationLogModel          model.LiquidationLogModel
	PurchaseApplicationModel     model.PurchaseApplicationModel
	RedemptionApplicationModel   model.RedemptionApplicationModel
	CustomerPositionModel        model.CustomerPositionModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)

	// 使用空缓存配置（不使用 Redis，使用内存缓存）
	// 如果后续需要 Redis，可以在配置文件中添加 Redis 配置
	//cacheConf := cache.CacheConf{}

	return &ServiceContext{
		Config: c,
		Conn:   conn,

		// 初始化所有模型（使用缓存）
		WorkCalendarModel:            model.NewWorkCalendarModel(conn),
		ProductNetValueModel:         model.NewProductNetValueModel(conn),
		TransactionConfirmationModel: model.NewTransactionConfirmationModel(conn),
		LiquidationLogModel:          model.NewLiquidationLogModel(conn),
		PurchaseApplicationModel:     model.NewPurchaseApplicationModel(conn),
		RedemptionApplicationModel:   model.NewRedemptionApplicationModel(conn),
		CustomerPositionModel:        model.NewCustomerPositionModel(conn),
	}
}
