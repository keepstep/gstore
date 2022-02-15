package config

type DataConfig struct {
	UniqName    string      `json:"uniq_name"`    //全局唯一名字(必填)
	ExpireTime  int64       `json:"expire_time"`  //redis过期时间(必填)
	Tdata       interface{} `json:"tdata"`        //模板结构体变量(必填)
	DBName      string      `json:"db_name"`      //mongo库名字(不填表示不落地)
	TableName   string      `json:"table_name"`   //mongo表名(不填表示不落地)
	CacheKey    string      `json:"cache_key"`    //rediskey前缀(不填使用UniqName)
	SyncTimeout int64       `json:"sync_timeout"` //超过多久落地(不填使用默认)
	SyncCount   int64       `json:"sync_count"`   //每次落地数量(不填使用默认)
}

type ZSetDataConfig struct {
	UniqName   string      `json:"uniq_name"`
	ExpireTime int64       `json:"expire_time"`
	Tdata      interface{} `json:"tdata"`
	CacheKey   string      `json:"cache_key"`
}

type MemoryDataConfig struct {
	UniqName string      `json:"uniq_name"`
	Tkey     interface{} `json:"tkey"`
	Tdata    interface{} `json:"tdata"`
}

const (
	Name_userinfo = "userinfo"
)

//zset
const (
	Name_user_recommend_t1 = "user_recommend_t1"
	Name_user_recommend_t2 = "user_recommend_t2"
)

//memory
const (
	Name_city_match_t1 = "city_match_t1"
	Name_city_match_t2 = "city_match_t2"
)
