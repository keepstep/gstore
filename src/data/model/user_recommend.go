package model

import "gstore/src/data/config"

//配置
var CfgUserRecommend = config.ZSetDataConfig{
	UniqName:   config.Name_user_recommend,
	ExpireTime: 60 * 60 * 12,
	Tdata:      UserRecommend{},
	DBName:     "test",
	TableName:  "user_recommend",
}

//数据结构
type UserRecommend struct {
	Uid uint64 `bson:"_id" json:"id"`
	Sex int32  `bson:"sex" json:"sex"`
}
