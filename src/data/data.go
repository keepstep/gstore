package data

import (
	"errors"
	"flag"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"gstore/src/data/logger"
	"gstore/src/data/store"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var syncStoreMap = map[string]store.StoreSync{}
var syncMutex sync.RWMutex
var syncDuration = flag.Duration("data_sync_duration", 5*time.Second, "the duration for synchronization")
var syncRunFlag = false

func AddToSyncList(item store.StoreSync) bool {
	syncMutex.Lock()
	defer syncMutex.Unlock()
	if item != nil {
		if _, ok := syncStoreMap[item.Name()]; ok {
			return false
		}
		syncStoreMap[item.Name()] = item
	}
	return true
}

func DelFromSyncList(item store.StoreSync) bool {
	syncMutex.Lock()
	defer syncMutex.Unlock()
	delete(syncStoreMap, item.Name())
	return true
}

func Sync() {
	syncMutex.RLock()
	defer syncMutex.RUnlock()
	for _, item := range syncStoreMap {
		item.Sync()
	}
}

func RunSync(exit chan string) {
	syncMutex.Lock()
	if syncRunFlag {
		syncMutex.Unlock()
		return
	} else {
		syncRunFlag = true
		syncMutex.Unlock()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)

	t := time.NewTicker(*syncDuration)
	for {
		select {
		case <-t.C:
			Sync()
		case s := <-quit:
			t.Stop()
			logger.Errorf("quit received signal:%v", s.String())
			return
		case s := <-exit:
			t.Stop()
			logger.Errorf("exit :%v", s)
			return
		}
	}
}

//----------------------------------------
func NewCacheData(name string, cache *redis.Client, db *mongo.Client) (*store.CacheStore, error) {
	cfg := GetDataConfig(name)
	if cfg == nil {
		return nil, errors.New("no data config name :" + name)
	}
	cacheKey := cfg.CacheKey
	if cacheKey == "" {
		cacheKey = cfg.UniqName
	}
	st := &store.CacheStore{}
	err := st.Init(
		store.Name(cfg.UniqName),
		store.TData(cfg.Tdata),
		store.Expired(cfg.ExpireTime),
		store.ReidsClient(cache),
		store.CacheKey(cacheKey),
		store.MongoClient(db),
		store.MongoDB(cfg.DBName),
		store.MongoTable(cfg.TableName),
		store.SyncParam(cfg.SyncTimeout, cfg.SyncCount),
	)
	if err != nil {
		return nil, errors.New("Init err name :" + name)
	}
	if st.IsNeedSync() {
		AddToSyncList(st)
	}
	return st, nil
}

func NewCacheZSet(name string, cache *redis.Client) (store.ZSetStore, error) {
	cfg := GetZSetDataConfig(name)
	if cfg == nil {
		return nil, errors.New("no data config name :" + name)
	}
	cacheKey := cfg.CacheKey
	if cacheKey == "" {
		cacheKey = cfg.UniqName
	}
	st := &store.CacheZSet{}
	err := st.Init(
		store.Name(cfg.UniqName),
		store.TZSetData(cfg.Tdata),
		store.Expired(cfg.ExpireTime),
		store.ReidsClient(cache),
		store.CacheKey(cacheKey),
	)
	if err != nil {
		return nil, errors.New("Init err name :" + name)
	}
	return st, nil
}

func NewCacheMemory(name string) (store.MemoryStore, error) {
	cfg := GetMemoryDataConfig(name)
	if cfg == nil {
		return nil, errors.New("no data config name :" + name)
	}
	st := &store.CacheMemory{}
	err := st.Init(
		store.Name(cfg.UniqName),
		store.TMemoryData(cfg.Tkey, cfg.Tdata),
	)
	if err != nil {
		return nil, errors.New("Init err name :" + name)
	}
	if st.IsNeedSync() {
		AddToSyncList(st)
	}
	return st, nil
}
