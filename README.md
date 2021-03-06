# gstore
golang redis+mongo store auto sync
### 说明
- 这是一个redis+mongodb数据存取同步模块
- 使用反射实现
- 定时同步到mongo(默认数据被修改60s后)
- 使用例子 test.go

### 配置
- 全局唯一名字 data/config/config.go 
- 某一项数据 data/model/userinfo.go
- 全局map data/data_config.go 

### 例子
```go
//接口定义
type Store interface {
	Init(opts ...Option) error
	Name() string
	Get(id interface{}) (interface{}, error)
	Set(id, value interface{}) error
	SetAlways(id, value interface{}) error

	Read(id interface{}) (interface{}, error)
	Save(id, value interface{}) error
	Sync() error

	ReadMany(tableName string, filter interface{}, option interface{}) (interface{}, error)
	IsNeedSync() bool
	Remove(id interface{}) error
}

//创建实例
func newCache() *store.CacheStore {
	//-----------------------------------------------------
	redisCfg := &store.RedisCfg{
		Host: "192.168.0.118",
		Port: 6379,
		DB:   1,
		Auth: "",
	}
	mongoCfg := &store.MongoCfg{
		Url: "mongodb://192.168.0.118:27017",
	}

	redis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisCfg.Host, redisCfg.Port),
		Password: redisCfg.Auth,
		DB:       redisCfg.DB,
	})
	mongo, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoCfg.Url).SetMaxPoolSize(20))
	if err != nil {
		return nil
	}
	//-----------------------------------------------------
	userCache, err := NewCacheData(
		config.Name_userinfo,
		redis,
		mongo)
	if err != nil {
		return nil
	}
	return userCache
}
//读写
func example(user *store.CacheStore) {
	id := "1000"
	item, err := user.Get(id)
	//不存在新建
	if err == store.Nil {
		d := &model.UserInfo{
			Id:   id,
			Name: "lisa",
			Age:  18,
		}
		err = user.Set(d.Id, d)
		fmt.Println("user.New", err)
	} else if err != nil {
		fmt.Println("user.Get", err)
	} else {
		//修改
		d := item.(*model.UserInfo)
		d.Age++
		err := user.Set(d.Id, d)
		if err != nil {
			fmt.Println("user.Set", err)
		}
	}
}
func main() {
	//创建
	userCache := newCache()
	//启动同步
	exit := make(chan string)
	go RunSync(exit)
	//-----------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	t := time.NewTicker(*testDuration)
	for {
		select {
		case <-t.C:
			//读写操作
			example(userCache)
		case s := <-quit:
			t.Stop()
			logger.Errorf("quit received signal:%v", s.String())
			close(exit)
			return
		}
	}
}
```


