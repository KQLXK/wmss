package svc

import (
	"WMSS/trade/api/internal/config"
	"WMSS/trade/api/model"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                       config.Config
	Conn                         sqlx.SqlConn
	PurchaseApplicationModel     model.PurchaseApplicationModel
	RedemptionApplicationModel   model.RedemptionApplicationModel
	TransactionConfirmationModel model.TransactionConfirmationModel
	CustomerPositionModel        model.CustomerPositionModel
	ProductInfoModel             model.ProductInfoModel
	CustomerInfoModel            model.CustomerInfoModel
	CustomerBankCardModel        model.CustomerBankCardModel
	ProductNetValueModel         model.ProductNetValueModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:                       c,
		Conn:                         conn,
		PurchaseApplicationModel:     model.NewPurchaseApplicationModel(conn),
		RedemptionApplicationModel:   model.NewRedemptionApplicationModel(conn),
		TransactionConfirmationModel: model.NewTransactionConfirmationModel(conn),
		CustomerPositionModel:        model.NewCustomerPositionModel(conn),
		ProductInfoModel:             model.NewProductInfoModel(conn),
		CustomerInfoModel:            model.NewCustomerInfoModel(conn),
		CustomerBankCardModel:        model.NewCustomerBankCardModel(conn),
		ProductNetValueModel:         model.NewProductNetValueModel(conn),
	}
}
