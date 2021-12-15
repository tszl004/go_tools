package db

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"

	"github.com/tszl004/go_tools/core_errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type SqlType string

const (
	MySQL     = "Mysql"
	Postgre   = "Postgre"
	SqlServer = "SqlServer"
)

func GetSqlType(sqlType string) (SqlType, error) {
	switch strings.ToLower(sqlType) {
	case "mysql":
		return MySQL, nil
	case "sqlserver", "mssql":
		return SqlServer, nil
	case "postgres", "postgresql", "postgre":
		return Postgre, nil
	default:
		return "", errors.New(core_errors.ErrDbDriverNotExists.Error() + sqlType)
	}
}

func GetDial(sqlType SqlType, confDetail ConfigParamsDetail) (gorm.Dialector, error) {
	var dbDial gorm.Dialector
	dsn := getDsn(sqlType, confDetail)
	switch sqlType {
	case MySQL:
		dbDial = mysql.Open(dsn)
	case SqlServer:
		dbDial = sqlserver.Open(dsn)
	case Postgre:
		dbDial = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("%s: %s", core_errors.ErrDbDriverNotExists.Error(), sqlType)
	}
	return dbDial, nil
}

func GetSqlDriver(sqlType SqlType, readOpen bool, dbConf ConfigParams) (*gorm.DB, error) {
	var dbDial gorm.Dialector
	if val, err := GetDial(sqlType, dbConf.Write); err != nil {
		return nil, err
	} else {
		dbDial = val
	}
	gormDb, err := gorm.Open(dbDial, &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 dbConf.Logger, // 使用配置中的Logger
		// Logger:                 redefineLog(sqlType), //拦截、接管 gorm v2 自带日志
	})
	if err != nil {
		// gorm 数据库驱动初始化失败
		return nil, err
	}

	// 如果开启了读写分离，配置读数据库（resource、read、replicas）
	// 读写分离配置只
	if readOpen {
		if val, err := GetDial(sqlType, dbConf.Write); err != nil {
			return nil, err
		} else {
			dbDial = val
		}
		resolverConf := dbresolver.Config{
			Replicas: []gorm.Dialector{dbDial},  //  读 操作库，查询类
			Policy:   dbresolver.RandomPolicy{}, // sources/replicas 负载均衡策略适用于
		}
		solver := dbresolver.Register(resolverConf).
			SetConnMaxIdleTime(dbConf.Read.ConnMaxIdleTime).
			SetConnMaxLifetime(dbConf.Read.ConnMaxLifetime).
			SetMaxIdleConns(dbConf.Read.MaxIdleConn).
			SetMaxOpenConns(dbConf.Read.MaxOpenConn)
		err = gormDb.Use(solver)
		if err != nil {
			return nil, err
		}
	}

	// 查询没有数据，屏蔽 gorm v2 包中会爆出的错误
	// https://github.com/go-gorm/gorm/issues/3789  此 issue 所反映的问题就是我们本次解决掉的
	_ = gormDb.Callback().Query().Before("gorm:query").Register("disable_raise_record_not_found", func(d *gorm.DB) {
		d.Statement.RaiseErrorOnNotFound = false
	})

	// 为主连接设置连接池(43行返回的数据库驱动指针)
	if rawDb, err := gormDb.DB(); err != nil {
		return nil, err
	} else {
		rawDb.SetConnMaxIdleTime(dbConf.Write.ConnMaxIdleTime)
		rawDb.SetConnMaxLifetime(dbConf.Write.ConnMaxLifetime)
		rawDb.SetMaxIdleConns(dbConf.Write.MaxIdleConn)
		rawDb.SetMaxOpenConns(dbConf.Write.MaxOpenConn)
		return gormDb, nil
	}
}

//  根据配置参数生成数据库驱动 dsn
func getDsn(sqlType SqlType, confDetail ConfigParamsDetail) string {
	var (
		User     = confDetail.User
		Pass     = confDetail.Pass
		Host     = confDetail.Host
		Port     = confDetail.Port
		DataBase = confDetail.DataBase
		Charset  = confDetail.Charset
	)

	switch sqlType {
	case MySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", User, Pass, Host, Port, DataBase, Charset)
	case SqlServer:
		return fmt.Sprintf("server=%s;port=%d;database=%s;user id=%s;password=%s;encrypt=disable", Host, Port, DataBase, User, Pass)
	case Postgre:
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable TimeZone=Asia/Shanghai", Host, Port, DataBase, User, Pass)
	}
	return ""
}
