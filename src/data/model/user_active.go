package model

import "gstore/src/data/config"

//配置
var CfgUserActive = config.ZSetDataConfig{
	UniqName:    config.Name_user_active,
	ExpireTime:  60 * 60 * 12,
	Tdata:       uint64(10000),
	DBName:      "test",
	TableName:   "user_active",
	SyncTimeout: 5,
	SyncCount:   100,
	SyncDisable: false,

	CountLimit: 10000, //总数
	Sort:       -1,    //降序
}
