package db

import "time"

type ConfigParamsDetail struct {
	Host            string
	DataBase        string
	Port            int
	Prefix          string
	User            string
	Pass            string
	Charset         string
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
	MaxIdleConn     int
	MaxOpenConn     int
}