package core_vars

import "time"

var (
	RPCLoc, _ = time.LoadLocation("Asia/Shanghai")
	DateTimeLayout = "2006-01-02 15:04:05"
	DateLayout = "2006-01-02"
	DateIntLayout = "20060102"
)
