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

func newCache() store.Store {
	//-----------------------------------------------------
	redisCfg := &store.RedisCfg{
		Host: "192.168.0.118",
		Port: 6379,
		DB:   1,
		Auth: "",
	}
	mongoCfg := &store.MongoCfg{
		Url: "mongodb://192.168.0.185:27017",
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
func example(user store.Store) {
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

func newZSet() store.ZSetStore {
	//-----------------------------------------------------
	redisCfg := &store.RedisCfg{
		Host: "192.168.0.118",
		Port: 6379,
		DB:   1,
		Auth: "",
	}
	mongoCfg := &store.MongoCfg{
		Url: "mongodb://192.168.0.185:27017",
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
	userActive, err := NewCacheZSet(
		config.Name_user_active,
		redis,
		mongo)
	if err != nil {
		return nil
	}
	return userActive
}
func exampleZSet(user store.ZSetStore) {
	id := "1000"
	var member uint64 = 100000
	score, err := user.ZScoreEx(id, member)
	//不存在新建
	if err == store.Nil {
		err = user.ZAddEx(id, 1000, member)
		fmt.Println("user_active.Add", err)
	} else if err != nil {
		fmt.Println("user_active.Score Err", err)
	} else {
		fmt.Println("user_active.Score", score)
		//修改
		user.ZIncrByEx(id, 10, member)
		user.ZIncrByEx(id, 10, member+100000)
		user.ZIncrByEx(id, 10, member+2*100000)
	}
}
func exampleZSet2(user store.ZSetStore) {
	var member uint64 = 100000
	score, err := user.ZScore(member)
	//不存在新建
	if err == store.Nil {
		err = user.ZAdd(1000, member)
		fmt.Println("user_active.Add", err)
	} else if err != nil {
		fmt.Println("user_active.Score Err", err)
	} else {
		fmt.Println("user_active.Score", score)
		//修改
		user.ZIncrBy(10, member)
		user.ZIncrBy(10, member+100000)
		user.ZIncrBy(10, member+2*100000)
	}
}
func exampleZSet2Rem(user store.ZSetStore, min, max string) {
	user.ZRemRangeByScore(min, max)
}

func TestStoreData() {
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
func TestZSetData() {
	//创建
	userActive := newZSet()
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
			//exampleZSet(userActive)
			exampleZSet2(userActive)
		case s := <-quit:
			exampleZSet2Rem(userActive, "0", "4000")
			t.Stop()
			logger.Errorf("quit received signal:%v", s.String())
			close(exit)
			return
		}
	}
}

func TestData() {
	TestZSetData()
}
