package model

import "gstore/src/data/config"

//配置
var CfgUserinfo = config.DataConfig{
	UniqName:   config.Name_userinfo,
	ExpireTime: 60 * 60 * 12,
	Tdata:      UserInfo{},
	DBName:     "user_data_db",
	TableName:  "user_store",
}

//数据结构
type UserInfo struct {
	Id   string `bson:"_id" json:"id"`
	Name string `bson:"name" json:"name"`
	Age  int    `bson:"age" json:"age"`
}

//表名(如需分表,根据id返回表名,不需要则返回空）
func (d UserInfo) TableName(id interface{}) string {
	return ""
}

//默认数据(数据不存在,生成一份默认数据,避免每次请求到查询数据库,可以返回nil)
func (d UserInfo) DefaultData(id interface{}) interface{} {
	if key, ok := id.(string); ok {
		return &UserInfo{
			Id: key,
		}
	}
	return nil
}
