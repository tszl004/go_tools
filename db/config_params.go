package db

import "gorm.io/gorm/logger"

type ConfigParams struct {
	Write  ConfigParamsDetail
	Read   ConfigParamsDetail
	Logger logger.Interface
}