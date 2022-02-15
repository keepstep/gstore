package data

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gstore/src/data/config"
	"gstore/src/data/logger"
	"gstore/src/data/model"
	"gstore/src/data/store"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var testDuration = flag.Duration("test_duration", 2*time.Second, "the duration for test")

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
		fmt.Println("newCache err:", err)
		return nil
	}
	//-----------------------------------------------------
	userCache, err := NewCacheData(
		config.Name_userinfo,
		redis,
		mongo)
	if err != nil {
		fmt.Println("newCache err:", err)
		return nil
	}
	return userCache
}

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

func TestData() {
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
