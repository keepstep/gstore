package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gstore/src/data/logger"

	"github.com/go-redis/redis/v8"
)

const DEFAULT_ZSET_SYNC_COUNT_LIMIT = 2000
const DEFAULT_ZSET_SYNC_SORT = -1

type CacheZSetListItem struct {
	Member string `bson:"_id" json:"member"`
	Score  int64  `bson:"score" json:"score"`
}

type CacheZSetData struct {
	Id         string               `bson:"_id" json:"_id"`
	List       []*CacheZSetListItem `bson:"list" json:"list"`
	CreateTime int64                `bson:"create_time" json:"create_time"`
	UpdateTime int64                `bson:"update_time" json:"update_time"`
}

type CacheZSet struct {
	options          Options
	syncCountPerTime int64 //每批次落地数量
}

func (d *CacheZSet) Init(opts ...Option) error {
	for _, o := range opts {
		o(&d.options)
	}
	if d.options.uniqName == "" {
		msg := "CacheZSet Init Err uniqName"
		logger.Error(msg)
		return errors.New(msg)
	}
	if d.options.cacheKey == "" {
		d.options.cacheKey = d.options.uniqName
	}
	if d.options.cache == nil {
		msg := "CacheZSet Init Err cache nil name:" + d.options.uniqName
		logger.Error(msg)
		return errors.New(msg)
	}
	//if d.options.expireSecond <= 0 {
	//	msg := "CacheZSet Init Err expireSecond name:" + d.options.uniqName
	//	logger.Error(msg)
	//	return errors.New(msg)
	//}
	if d.options.syncTimeout <= 0 {
		d.options.syncTimeout = DEFAULT_SYNC_TIMEOUT
	}
	if d.options.syncCountPerTime <= 0 {
		d.options.syncCountPerTime = DEFAULT_SYNC_COUNT
	}
	d.syncCountPerTime = d.options.syncCountPerTime

	if d.options.tdata == nil {
		msg := "CacheZSet Init Err tdata nil name:" + d.options.uniqName
		logger.Error(msg)
		return errors.New(msg)
	}

	if d.options.tdataKind == reflect.Struct || d.options.tdataKind == reflect.String ||
		d.options.tdataKind == reflect.Bool ||
		d.options.tdataKind == reflect.Int32 ||
		d.options.tdataKind == reflect.Int64 ||
		d.options.tdataKind == reflect.Uint32 ||
		d.options.tdataKind == reflect.Uint64 ||
		d.options.tdataKind == reflect.Float32 ||
		d.options.tdataKind == reflect.Float64 {

	} else {
		msg := "CacheZSet Init Err tdataKind no valid reflect.Type name:" + d.options.uniqName
		logger.Error(msg)
		return errors.New(msg)
	}

	if d.options.mongo == nil {
		d.options.needSync = false
		msg := "CacheZSet Init mongo nil name:" + d.options.uniqName
		logger.Error(msg)
	} else {
		if d.options.mongoCfg.DB == "" {
			msg := "CacheZSet Init mongoCfg DB nil name:" + d.options.uniqName
			logger.Error(msg)
			return errors.New(msg)
		}
		if d.options.mongoCfg.Table == "" {
			msg := "CacheZSet Init mongoCfg Table nil name:" + d.options.uniqName
			logger.Error(msg)
			return errors.New(msg)
		}
		_, err := d.options.mongo.Database(d.options.mongoCfg.DB).Collection(d.options.mongoCfg.Table).Indexes().CreateOne(context.Background(), mongo.IndexModel{
			//Keys:    bson.D{{Key: "list._id", Value: 1}},//这样是不同行之间，如果子文档list中有重复的这行没法插入
			Keys:    bson.D{{Key: "_id", Value: 1}, {Key: "list._id", Value: 1}},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			msg := fmt.Sprintf("CreateIndex------Err: %v %v ", err, d.options.uniqName)
			logger.Error(msg)
			return errors.New(msg)
		}
		d.options.needSync = true
	}
	if d.options.syncDisable {
		d.options.needSync = false
	}
	if d.options.tzsetCountLimit <= 0 {
		d.options.tzsetCountLimit = DEFAULT_ZSET_SYNC_COUNT_LIMIT
	}
	if d.options.tzsetSort == 0 {
		d.options.tzsetSort = DEFAULT_ZSET_SYNC_SORT
	}
	return nil
}
func (d *CacheZSet) Name() string {
	return d.options.uniqName
}
func (d *CacheZSet) Sync() error {
	if !d.options.needSync {
		return nil
	}
	if d.options.mongo == nil {
		msg := fmt.Sprintf("Sync Err:no db name:%v", d.options.uniqName)
		logger.Error(msg)
		return errors.New(msg)
	}
	//syncKey := d.getSyncKey()
	//ssm, err := d.options.cache.HGetAll(context.Background(), syncKey).Result()
	total, need, ssm, err := d.getSyncItems()
	if err == redis.Nil {
		return nil
	} else if err != nil {
		msg := fmt.Sprintf("Sync Err:get all key name:%v err:%v", d.options.uniqName, err)
		logger.Error(msg)
		return err
	}
	allCount := 0
	newCount := 0
	for key, version := range ssm {
		insert, err := d.syncOne(key, version)
		if err != nil {
			msg := fmt.Sprintf("Sync Err:syncOne name:%v key:%v err:%v", d.options.uniqName, key, err)
			logger.Error(msg)
			continue
		}
		allCount++
		newCount += insert
		logger.Infof("Sync succ key:%v version:%v", key, version)
	}
	logger.Errorf("Sync succ name:%v total:%v need:%v all:%v new:%v", d.options.uniqName, total, need, allCount, newCount)
	return nil
}
func (d *CacheZSet) IsNeedSync() bool {
	return d.options.needSync
}

func (d *CacheZSet) convertRedisData(val string) (interface{}, error) {
	switch d.options.tdataKind {
	case reflect.Struct:
		var ii = reflect.New(d.options.tdataType)
		rst := ii.Interface()
		err := json.Unmarshal([]byte(val), rst)
		return rst, err
	case reflect.String:
		return val, nil
	case reflect.Int:
	case reflect.Int32:
	case reflect.Int64:
		i, err := strconv.ParseInt(val, 10, 64)
		return i, err
	case reflect.Uint:
	case reflect.Uint32:
	case reflect.Uint64:
		i, err := strconv.ParseUint(val, 10, 64)
		return i, err
	case reflect.Float32:
	case reflect.Float64:
		i, err := strconv.ParseFloat(val, 64)
		return i, err
	}
	return nil, errors.New("convert err")
}
func (d *CacheZSet) getCacheKey(id interface{}) string {
	if id == nil {
		return d.options.cacheKey
	}
	key := fmt.Sprintf("%s:%v", d.options.cacheKey, id)
	return key
}
func (d *CacheZSet) getTableName() string {
	nn := d.options.mongoCfg.Table
	return nn
}

func (d *CacheZSet) convertData(value interface{}) (string, error) {
	if value == nil {
		msg := "convertData err value nil name:" + d.options.uniqName
		logger.Error(msg)
		return "", errors.New(msg)
	}
	saveData := ""
	switch reflect.ValueOf(value).Kind() {
	case reflect.Ptr:
		if d.options.tdataType == reflect.ValueOf(value).Type().Elem() {
			bs, err := json.Marshal(value)
			if err != nil {
				return "", err
			}
			saveData = string(bs)
		} else {
			msg := "convertData err ptr value type!=tdataType name:" + d.options.uniqName
			logger.Error(msg)
			return "", errors.New(msg)
		}
	case reflect.Struct:
		if d.options.tdataType != reflect.ValueOf(value).Type() {
			msg := "convertData err struct value type!=tdataType name:" + d.options.uniqName
			logger.Error(msg)
			return "", errors.New(msg)
		}
		val := reflect.ValueOf(value)
		bs, err := json.Marshal(val)
		if err != nil {
			return "", err
		}
		saveData = string(bs)
	case reflect.String, reflect.Bool, reflect.Int32, reflect.Int64, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		if d.options.tdataType != reflect.ValueOf(value).Type() {
			msg := "convertData err value type!=tdataType name:" + d.options.uniqName
			logger.Error(msg)
			return "", errors.New(msg)
		}
		saveData = fmt.Sprintf("%v", value)
	default:
		msg := "convertData err value type err name:" + d.options.uniqName
		logger.Error(msg)
		return "", errors.New(msg)
	}
	return saveData, nil
}
func (d *CacheZSet) parseData(dataList []string) (interface{}, error) {
	if d.options.tdataKind == reflect.Struct {
		stp := reflect.SliceOf(reflect.PtrTo(d.options.tdataType))
		var ii = reflect.New(stp).Elem()
		for _, str := range dataList {
			val, err := d.convertRedisData(str)
			if err != nil {
				return nil, err
			}
			ii = reflect.Append(ii, reflect.ValueOf(val))
		}
		return ii.Interface(), nil
	} else {
		stp := reflect.SliceOf(d.options.tdataType)
		var ii = reflect.New(stp).Elem()
		for _, str := range dataList {
			val, err := d.convertRedisData(str)
			if err != nil {
				return nil, err
			}
			ii = reflect.Append(ii, reflect.ValueOf(val))
		}
		return ii.Interface(), nil
	}
}
func (d *CacheZSet) setByScore(id interface{}, score float64, saveData string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	key := d.getCacheKey(id)
	err := d.options.cache.ZAdd(ctx, key, &redis.Z{
		Score:  score,
		Member: saveData,
	}).Err()
	if err != nil {
		logger.Errorf("setEx err name=%v key=%v valule=%v err:%v", d.options.uniqName, key, saveData, err)
	} else {
		if d.options.expireSecond <= 0 {
			d.options.cache.Expire(ctx, key, redis.KeepTTL)
		} else {
			d.options.cache.Expire(ctx, key, time.Duration(d.options.expireSecond)*time.Second)
		}
		d.setSync(id, saveData)
	}
	return err
}

func (d *CacheZSet) ZAdd(score float64, member interface{}) error {
	return d.ZAddEx(nil, score, member)
}
func (d *CacheZSet) ZRem(member interface{}) error {
	return d.ZRemEx(nil, member)
}
func (d *CacheZSet) ZCard() (int64, error) {
	return d.ZCardEx(nil)
}
func (d *CacheZSet) ZCount(min, max string) (int64, error) {
	return d.ZCountEx(nil, min, max)
}
func (d *CacheZSet) ZScore(member interface{}) (float64, error) {
	return d.ZScoreEx(nil, member)
}
func (d *CacheZSet) ZRank(member interface{}) (int64, error) {
	return d.ZRankEx(nil, member)
}
func (d *CacheZSet) ZRevRank(member interface{}) (int64, error) {
	return d.ZRevRankEx(nil, member)
}
func (d *CacheZSet) ZRange(start, stop int64) (interface{}, error) {
	return d.ZRangeEx(nil, start, stop)
}
func (d *CacheZSet) ZRevRange(start, stop int64) (interface{}, error) {
	return d.ZRevRangeEx(nil, start, stop)
}
func (d *CacheZSet) ZRangeByScore(min, max string, offset, count int64) (interface{}, error) {
	return d.ZRangeByScoreEx(nil, min, max, offset, count)
}
func (d *CacheZSet) ZRevRangeByScore(min, max string, offset, count int64) (interface{}, error) {
	return d.ZRevRangeByScoreEx(nil, min, max, offset, count)
}

//no support sync
func (d *CacheZSet) ZRemRangeByRank(start, stop int64) (int64, error) {
	return d.ZRemRangeByRankEx(nil, start, stop)
}

//sync just now
func (d *CacheZSet) ZRemRangeByScore(min, max string) (int64, error) {
	return d.ZRemRangeByScoreEx(nil, min, max)
}

//no support sync
func (d *CacheZSet) ZPopMin(count int64) (interface{}, []float64, error) {
	return d.ZPopMinEx(nil, count)
}

//no support sync
func (d *CacheZSet) ZPopMax(count int64) (interface{}, []float64, error) {
	return d.ZPopMaxEx(nil, count)
}
func (d *CacheZSet) ZIncrBy(increment float64, member interface{}) (float64, error) {
	return d.ZIncrByEx(nil, increment, member)
}

func (d *CacheZSet) ZRangeWithScores(start, stop int64) (interface{}, []float64, error) {
	return d.ZRangeWithScoresEx(nil, start, stop)
}
func (d *CacheZSet) ZRevRangeWithScores(start, stop int64) (interface{}, []float64, error) {
	return d.ZRevRangeWithScoresEx(nil, start, stop)
}
func (d *CacheZSet) ZRangeByScoreWithScores(min, max string, offset, count int64) (interface{}, []float64, error) {
	return d.ZRangeByScoreWithScoresEx(nil, min, max, offset, count)
}
func (d *CacheZSet) ZRevRangeByScoreWithScores(min, max string, offset, count int64) (interface{}, []float64, error) {
	return d.ZRevRangeByScoreWithScoresEx(nil, min, max, offset, count)
}

func (d *CacheZSet) ZAddEx(id interface{}, score float64, member interface{}) error {
	d.syncLoadFromDB(id)
	saveData, err := d.convertData(member)
	if err != nil {
		logger.Errorf("ZAdd convert err key:%v err:%v score:%v val:%+v", d.options.uniqName, err, score, member)
		return err
	}
	err = d.setByScore(id, score, saveData)
	if err != nil {
		logger.Errorf("ZAdd err key:%v err:%v score:%v val:%+v", d.options.uniqName, err, score, member)
	}

	return err
}
func (d *CacheZSet) ZRemEx(id, member interface{}) error {
	d.syncLoadFromDB(id)
	saveData, err := d.convertData(member)
	if err != nil {
		logger.Errorf("ZRem convert err key:%v err:%v val:%+v", d.options.uniqName, err, member)
		return err
	}
	key := d.getCacheKey(id)
	err = d.options.cache.ZRem(context.Background(), key, saveData).Err()
	if err != nil {
		logger.Errorf("ZRem err key:%v err:%v val:%+v", d.options.uniqName, err, member)
	} else {
		d.setSync(id, saveData)
	}
	return err
}
func (d *CacheZSet) ZCardEx(id interface{}) (int64, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	count, err := d.options.cache.ZCard(context.Background(), key).Result()
	if err != nil {
		logger.Errorf("ZCard err key:%v err:%v", d.options.uniqName, err)
	}
	if err == redis.Nil {
		err = Nil
	}
	return count, err
}
func (d *CacheZSet) ZCountEx(id interface{}, min, max string) (int64, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	count, err := d.options.cache.ZCount(context.Background(), key, min, max).Result()
	if err != nil {
		logger.Errorf("ZCard err key:%v err:%v", d.options.uniqName, err)
	}
	return count, err
}
func (d *CacheZSet) ZScoreEx(id, member interface{}) (float64, error) {
	d.syncLoadFromDB(id)
	saveData, err := d.convertData(member)
	if err != nil {
		return 0, err
	}
	key := d.getCacheKey(id)
	score, err := d.options.cache.ZScore(context.Background(), key, saveData).Result()
	if err != nil {
		logger.Errorf("ZScore err key:%v err:%v val:%+v", d.options.uniqName, err, member)
	}
	if err == redis.Nil {
		err = Nil
	}
	return score, err
}
func (d *CacheZSet) ZRankEx(id, member interface{}) (int64, error) {
	d.syncLoadFromDB(id)
	saveData, err := d.convertData(member)
	if err != nil {
		return 0, err
	}
	key := d.getCacheKey(id)
	rank, err := d.options.cache.ZRank(context.Background(), key, saveData).Result()
	if err != nil {
		logger.Errorf("ZRank err key:%v err:%v val:%+v", d.options.uniqName, err, member)
	}
	if err == redis.Nil {
		err = Nil
	}
	return rank, err
}
func (d *CacheZSet) ZRevRankEx(id, member interface{}) (int64, error) {
	d.syncLoadFromDB(id)
	saveData, err := d.convertData(member)
	if err != nil {
		return 0, err
	}
	key := d.getCacheKey(id)
	rank, err := d.options.cache.ZRevRank(context.Background(), key, saveData).Result()
	if err != nil {
		logger.Errorf("ZRevRank err key:%v err:%v val:%+v", d.options.uniqName, err, member)
	}
	return rank, err
}
func (d *CacheZSet) ZRangeEx(id interface{}, start, stop int64) (interface{}, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	dataList, err := d.options.cache.ZRange(context.Background(), key, start, stop).Result()
	if err != nil {
		logger.Errorf("ZRange err key:%v err:%v start:%v stop:%v", d.options.uniqName, err, start, stop)
		return nil, err
	}
	ii, err := d.parseData(dataList)
	if err != nil {
		logger.Errorf("ZRange err parseData key:%v err:%v start:%v stop:%v", d.options.uniqName, err, start, stop)
	}
	return ii, err
}
func (d *CacheZSet) ZRevRangeEx(id interface{}, start, stop int64) (interface{}, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	dataList, err := d.options.cache.ZRevRange(context.Background(), key, start, stop).Result()
	if err != nil {
		logger.Errorf("ZRevRange err key:%v err:%v start:%v stop:%v", d.options.uniqName, err, start, stop)
		return nil, err
	}
	ii, err := d.parseData(dataList)
	if err != nil {
		logger.Errorf("ZRevRange err parseData key:%v err:%v start:%v stop:%v", d.options.uniqName, err, start, stop)
	}
	return ii, err
}
func (d *CacheZSet) ZRangeByScoreEx(id interface{}, min, max string, offset, count int64) (interface{}, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	dataList, err := d.options.cache.ZRangeByScore(context.Background(), key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}).Result()
	if err != nil {
		logger.Errorf("ZRangeByScore err key:%v err:%v min:%v max:%v", d.options.uniqName, err, min, max)
		return nil, err
	}
	ii, err := d.parseData(dataList)
	if err != nil {
		logger.Errorf("ZRangeByScore err parseData key:%v err:%v min:%v max:%v", d.options.uniqName, err, min, max)
	}
	return ii, err
}
func (d *CacheZSet) ZRevRangeByScoreEx(id interface{}, min, max string, offset, count int64) (interface{}, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	dataList, err := d.options.cache.ZRevRangeByScore(context.Background(), key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}).Result()
	if err != nil {
		logger.Errorf("ZRevRangeByScore err key:%v err:%v min:%v max:%v", d.options.uniqName, err, min, max)
		return nil, err
	}
	ii, err := d.parseData(dataList)
	if err != nil {
		logger.Errorf("ZRevRangeByScore err parseData key:%v err:%v min:%v max:%v", d.options.uniqName, err, min, max)
	}
	return ii, err
}
func (d *CacheZSet) ZRemRangeByRankEx(id interface{}, start, stop int64) (int64, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	count, err := d.options.cache.ZRemRangeByRank(context.Background(), key, start, stop).Result()
	if err != nil {
		logger.Errorf("ZRemRangeByRank err key:%v err:%v start:%v stop:%v", d.options.uniqName, err, start, stop)
	}
	return count, err
}
func (d *CacheZSet) ZRemRangeByScoreEx(id interface{}, min, max string) (int64, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	count, err := d.options.cache.ZRemRangeByScore(context.Background(), key, min, max).Result()
	if err != nil {
		logger.Errorf("ZRemRangeByScore err key:%v err:%v min:%v max:%v", d.options.uniqName, err, min, max)
	}
	if d.options.needSync {
		d.syncRemRangeByScore(id, min, max)
	}
	return count, err
}
func (d *CacheZSet) ZPopMinEx(id interface{}, count int64) (interface{}, []float64, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	zlist, err := d.options.cache.ZPopMin(context.Background(), key, count).Result()
	if err != nil {
		logger.Errorf("ZPopMin err key:%v count:%v err:%v", d.options.uniqName, count, err)
		return nil, nil, err
	}
	scoreList := []float64{}
	dataList := []string{}
	for _, item := range zlist {
		scoreList = append(scoreList, item.Score)
		dataList = append(dataList, fmt.Sprintf("%v", item.Member))
	}
	ii, err := d.parseData(dataList)
	if err != nil {
		logger.Errorf("ZPopMin err parseData key:%v count:%v err:%v", d.options.uniqName, count, err)
		return nil, nil, err
	}
	return ii, scoreList, nil
}
func (d *CacheZSet) ZPopMaxEx(id interface{}, count int64) (interface{}, []float64, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	zlist, err := d.options.cache.ZPopMax(context.Background(), key, count).Result()
	if err != nil {
		logger.Errorf("ZPopMax err key:%v count:%v err:%v", d.options.uniqName, count, err)
		return nil, nil, err
	}
	scoreList := []float64{}
	dataList := []string{}
	for _, item := range zlist {
		scoreList = append(scoreList, item.Score)
		dataList = append(dataList, fmt.Sprintf("%v", item.Member))
	}
	ii, err := d.parseData(dataList)
	if err != nil {
		logger.Errorf("ZPopMax err parseData key:%v count:%v err:%v", d.options.uniqName, count, err)
		return nil, nil, err
	}
	return ii, scoreList, nil
}

func (d *CacheZSet) ZRangeWithScoresEx(id interface{}, start, stop int64) (interface{}, []float64, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	zlist, err := d.options.cache.ZRangeWithScores(context.Background(), key, start, stop).Result()
	if err != nil {
		logger.Errorf("ZRangeWithScores err key:%v err:%v start:%v stop:%v", d.options.uniqName, err, start, stop)
		return nil, nil, err
	}
	scoreList := []float64{}
	dataList := []string{}
	for _, item := range zlist {
		scoreList = append(scoreList, item.Score)
		dataList = append(dataList, fmt.Sprintf("%v", item.Member))
	}
	ii, err := d.parseData(dataList)
	if err != nil {
		logger.Errorf("ZRangeWithScores err parseData key:%v err:%v", d.options.uniqName, err)
		return nil, nil, err
	}
	return ii, scoreList, nil
}
func (d *CacheZSet) ZRevRangeWithScoresEx(id interface{}, start, stop int64) (interface{}, []float64, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	zlist, err := d.options.cache.ZRevRangeWithScores(context.Background(), key, start, stop).Result()
	if err != nil {
		logger.Errorf("ZRevRangeWithScores err key:%v err:%v start:%v stop:%v", d.options.uniqName, err, start, stop)
		return nil, nil, err
	}
	scoreList := []float64{}
	dataList := []string{}
	for _, item := range zlist {
		scoreList = append(scoreList, item.Score)
		dataList = append(dataList, fmt.Sprintf("%v", item.Member))
	}
	ii, err := d.parseData(dataList)
	if err != nil {
		logger.Errorf("ZRevRangeWithScores err parseData key:%v err:%v", d.options.uniqName, err)
		return nil, nil, err
	}
	return ii, scoreList, nil
}
func (d *CacheZSet) ZRangeByScoreWithScoresEx(id interface{}, min, max string, offset, count int64) (interface{}, []float64, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	zlist, err := d.options.cache.ZRangeByScoreWithScores(context.Background(), key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}).Result()
	if err != nil {
		logger.Errorf("ZRangeByScoreWithScores err key:%v err:%v min:%v max:%v", d.options.uniqName, err, min, max)
		return nil, nil, err
	}
	scoreList := []float64{}
	dataList := []string{}
	for _, item := range zlist {
		scoreList = append(scoreList, item.Score)
		dataList = append(dataList, fmt.Sprintf("%v", item.Member))
	}
	ii, err := d.parseData(dataList)
	if err != nil {
		logger.Errorf("ZRangeByScoreWithScores err parseData key:%v err:%v min:%v max:%v", d.options.uniqName, err, min, max)
		return nil, nil, err
	}
	return ii, scoreList, nil
}
func (d *CacheZSet) ZRevRangeByScoreWithScoresEx(id interface{}, min, max string, offset, count int64) (interface{}, []float64, error) {
	d.syncLoadFromDB(id)
	key := d.getCacheKey(id)
	zlist, err := d.options.cache.ZRevRangeByScoreWithScores(context.Background(), key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}).Result()
	if err != nil {
		logger.Errorf("ZRevRangeByScoreWithScores err key:%v err:%v min:%v max:%v", d.options.uniqName, err, min, max)
		return nil, nil, err
	}
	scoreList := []float64{}
	dataList := []string{}
	for _, item := range zlist {
		scoreList = append(scoreList, item.Score)
		dataList = append(dataList, fmt.Sprintf("%v", item.Member))
	}
	ii, err := d.parseData(dataList)
	if err != nil {
		logger.Errorf("ZRevRangeByScoreWithScores err parseData key:%v err:%v min:%v max:%v", d.options.uniqName, err, min, max)
		return nil, nil, err
	}
	return ii, scoreList, nil
}
func (d *CacheZSet) ZIncrByEx(id interface{}, increment float64, member interface{}) (float64, error) {
	d.syncLoadFromDB(id)
	saveData, err := d.convertData(member)
	if err != nil {
		logger.Errorf("ZIncrByEx convert err key:%v val:%v err:%v ", d.options.uniqName, member, err)
		return 0, err
	}
	key := d.getCacheKey(id)
	score, err := d.options.cache.ZIncrBy(context.Background(), key, increment, saveData).Result()
	if err != nil {
		logger.Errorf("ZIncrBy err id:%v key:%v increment:%v member:%v err:%v", id, d.options.uniqName, increment, member, err)
		return 0, err
	} else {
		d.setSync(id, saveData)
	}

	return score, nil
}

//sync
func (d *CacheZSet) syncLoadFromDB(id interface{}) error {
	if !d.options.needSync {
		return nil
	}
	ctx := context.Background()
	_id := d.getCacheKey(id)
	tp, typeErr := d.options.cache.Type(ctx, _id).Result()
	if typeErr != nil {
		if typeErr != redis.Nil {
			logger.Errorf("syncLoadFromDB type error name:%v id:%v err:%v", d.options.uniqName, id, typeErr)
			return typeErr
		}
	} else {
		if tp == "none" {

		} else if tp != "zset" {
			logger.Errorf("syncLoadFromDB type error name:%v id:%v err:%v-%v", d.options.uniqName, id, "type not zset", tp)
			return errors.New("type not zset")
		} else {
			return nil
		}
	}
	opts := options.Update()
	tableName := d.getTableName()
	coll := d.options.mongo.Database(d.options.mongoCfg.DB).Collection(tableName)
	filter := bson.M{"_id": _id}
	list := bson.M{
		"$each":  []bson.M{},
		"$sort":  bson.M{"score": d.options.tzsetSort},
		"$slice": d.options.tzsetCountLimit,
	}
	//不限额
	if d.options.tzsetCountLimit <= 0 {
		list["$slice"] = nil
	}
	update := bson.M{
		"$push": bson.M{
			"list": list,
		},
	}
	_, sortErr := coll.UpdateOne(ctx, filter, update, opts)
	if sortErr != nil {
		logger.Errorf("syncLoadFromDB sort error name:%v id:%v err:%v", d.options.uniqName, id, sortErr)
	}

	data := new(CacheZSetData)
	findErr := coll.FindOne(ctx, filter).Decode(&data)
	if findErr != nil {
		msg := fmt.Sprintf("syncLoadFromDB find error name:%v id:%v err:%v", d.options.uniqName, id, findErr)
		logger.Error(msg)
		return findErr
	}
	var err error = nil
	allCount := int64(0)
	zs := []*redis.Z{}
	for _, item := range data.List {
		zs = append(zs, &redis.Z{
			Score:  float64(item.Score),
			Member: item.Member,
		})
		if len(zs) == 500 {
			//存在的不添加
			count, addErr := d.options.cache.ZAddNX(ctx, _id, zs...).Result()
			if addErr != nil {
				msg := fmt.Sprintf("syncLoadFromDB ZAddNX error name:%v id:%v count:%v-%v err:%v", d.options.uniqName, id, allCount, count, addErr)
				logger.Error(msg)
				err = addErr
			} else {
				allCount += count
			}
			zs = []*redis.Z{}
		}
	}
	if len(zs) > 0 {
		//存在的不添加
		count, addErr := d.options.cache.ZAddNX(ctx, _id, zs...).Result()
		if addErr != nil {
			msg := fmt.Sprintf("syncLoadFromDB ZAddNX error name:%v id:%v count:%v-%v err:%v", d.options.uniqName, id, allCount, count, addErr)
			logger.Error(msg)
			err = addErr
		} else {
			allCount += count
		}
	}
	msg := fmt.Sprintf("syncLoadFromDB info name:%v id:%v count:%v-%v err:%v", d.options.uniqName, id, len(data.List), allCount, err)
	logger.Info(msg)
	return err
}

func (d *CacheZSet) syncRemRangeByScore(id interface{}, min, max string) error {
	if !d.options.needSync {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_id := d.getCacheKey(id)
	opts := options.FindOneAndUpdate()
	tableName := d.getTableName()
	coll := d.options.mongo.Database(d.options.mongoCfg.DB).Collection(tableName)
	filter := bson.M{"_id": _id}
	minInt, _ := strconv.ParseInt(min, 10, 64)
	maxInt, _ := strconv.ParseInt(max, 10, 64)
	err := coll.FindOneAndUpdate(ctx, filter, bson.M{
		"$pull": bson.M{"list": bson.M{"score": bson.M{"$gte": minInt, "$lte": maxInt}}},
		"$set":  bson.M{"update_time": time.Now().Unix()},
	}, opts).Err()
	if err != nil {
		msg := fmt.Sprintf("syncRemRangeByScore fail name:%v id:%v min:%v max:%v err:%v", d.options.uniqName, id, min, max, err)
		logger.Error(msg)
	}
	return err
}

func (d *CacheZSet) getScoreByKeyForSync(id interface{}, member string) (score int64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	key := d.getCacheKey(id)
	val, err := d.options.cache.ZScore(ctx, key, member).Result()
	if err == redis.Nil {
		return 0, Nil
	} else if err != nil {
		return 0, err
	} else {
		return int64(val), nil
	}
}
func (d *CacheZSet) getSyncKey() string {
	syncKey := fmt.Sprintf("SYNC_ZSET:%s", d.options.uniqName)
	return syncKey
}
func (d *CacheZSet) combineSyncKey(id interface{}, member string) string {
	if len(member) == 0 {
		return ""
	}
	if id == nil {
		return member
	}
	return fmt.Sprintf("%v#@#%v", member, id)
}
func (d *CacheZSet) parseSyncKey(key string) (id interface{}, member string) {
	keys := strings.Split(key, "#@#")
	if len(keys) == 1 {
		return nil, key
	} else if len(keys) == 2 {
		return keys[1], keys[0]
	}
	return nil, ""
}
func (d *CacheZSet) setSync(id interface{}, member string) error {
	if !d.options.needSync {
		return nil
	}
	if nil == d.options.mongo {
		msg := fmt.Sprintf("setSync Err db nil name:%v,id:%v,member:%v", d.options.uniqName, id, member)
		logger.Error(msg)
		return errors.New(msg)
	}
	syncKey := d.getSyncKey()
	key := d.combineSyncKey(id, member)
	//毫秒级
	score := time.Now().UnixNano() / int64(time.Millisecond)
	memberAdd := &redis.Z{
		Score:  float64(score),
		Member: key,
	}
	memberInc := &redis.Z{
		Score:  1,
		Member: key,
	}
	//ZAdd 会导致刷新频繁数据 一直得不到落地机会
	//ZAddNX 存在score不变 可能会丢失最后一次数据 syncCacheScriptByScore
	//ZIncrXX 旧数据每次加1毫秒 新数据不处理
	err := d.options.cache.ZIncrXX(context.Background(), syncKey, memberInc).Err()
	if err != nil {
		msg := fmt.Sprintf("setSync step ZIncrXX name:%v,key:%v err:%v", d.options.uniqName, key, err)
		logger.Info(msg)
		err = d.options.cache.ZAdd(context.Background(), syncKey, memberAdd).Err()
		if err != nil {
			msg := fmt.Sprintf("setSync Err ZAdd name:%v,key:%v err:%v", d.options.uniqName, key, err)
			logger.Error(msg)
			return errors.New(msg)
		}
	}
	return err
}
func (d *CacheZSet) delSyncItem(key string) error {
	syncKey := d.getSyncKey()
	return d.options.cache.ZRem(context.Background(), syncKey, key).Err()
}
func (d *CacheZSet) getSyncItems() (int64, int64, map[string]string, error) {
	count := d.syncCountPerTime
	if count == 0 {
		count = d.options.syncCountPerTime
	}
	syncKey := d.getSyncKey()
	min := "0"
	score := time.Now().UnixNano()/int64(time.Millisecond) - d.options.syncTimeout*1000
	max := strconv.FormatInt(score, 10)
	ctx := context.Background()
	dataList, err := d.options.cache.ZRangeByScoreWithScores(ctx, syncKey, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: 0,
		Count:  count,
	}).Result()
	if err != nil {
		return 0, 0, map[string]string{}, err
	}
	mmap := map[string]string{}
	for _, item := range dataList {
		key := fmt.Sprintf("%v", item.Member)
		score := strconv.FormatInt(int64(item.Score), 10)
		mmap[key] = score
	}
	all, _ := d.options.cache.ZCard(ctx, syncKey).Result()
	need, _ := d.options.cache.ZCount(ctx, syncKey, min, max).Result()
	//------------------------------------------
	//----调节数量----
	if need > count {
		d.syncCountPerTime = count + d.options.syncCountPerTime
		if d.syncCountPerTime > DEFAULT_SYNC_COUNT_MAX {
			d.syncCountPerTime = DEFAULT_SYNC_COUNT_MAX
		}
	} else {
		d.syncCountPerTime = count - d.options.syncCountPerTime
		if d.syncCountPerTime < d.options.syncCountPerTime {
			d.syncCountPerTime = d.options.syncCountPerTime
		}
	}
	//msg := fmt.Sprintf("getSyncItemsZSet count %v:%v name:%v", d.syncCountPerTime, d.options.syncCountPerTime, d.options.uniqName)
	//logger.Errorf(msg)
	//------------------------------------------
	return all, need, mmap, nil
}
func (d *CacheZSet) syncOne(key string, version string) (insert int, err error) {
	insert = 0
	opt := 0
	needRemoveIfErr := true
	defer func() {
		if err != nil && needRemoveIfErr {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			syncKey := d.getSyncKey()
			_, err2 := d.options.cache.Eval(ctx, syncCacheScriptByScore, []string{syncKey}, key, version).Int()
			if err2 != nil {
				msg := fmt.Sprintf("syncOne fail to remove fail name:%v key:%v opt:%v err:%v err2:%v", d.options.uniqName, key, opt,
					err, err2)
				logger.Error(msg)
			} else {
				msg := fmt.Sprintf("syncOne fail to remove succ name:%v key:%v opt:%v err:%v", d.options.uniqName, key, opt, err)
				logger.Error(msg)
			}
		}
	}()

	id, member := d.parseSyncKey(key)
	if member == "" {
		err = errors.New("member empty")
		return
	}
	operate := "insert"
	score, err := d.getScoreByKeyForSync(id, member)
	if err == Nil {
		operate = "delete"
	} else if err != nil {
		return
	}
	logger.Infof("parseSyncKey %v %v %v %v", key, id, member, operate)
	_id := d.getCacheKey(id)
	//--------------------------------------------------------------------------------
	opts := options.Update()
	opts.ArrayFilters = &options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem._id": member},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	tableName := d.getTableName()
	coll := d.options.mongo.Database(d.options.mongoCfg.DB).Collection(tableName)
	if operate == "delete" {
		insert = 4
		filter := bson.M{"_id": _id}
		_, err = coll.UpdateOne(ctx, filter, bson.M{
			"$pull": bson.M{"list": bson.M{"_id": member}},
			"$set":  bson.M{"update_time": time.Now().Unix()},
		}, opts)
	} else {
		filter := bson.M{"_id": _id}
		rst, updErr := coll.UpdateOne(ctx, filter, bson.M{
			"$set": bson.M{
				"list.$[elem].score": score,
				"update_time":        time.Now().Unix(),
			},
		}, opts)

		if updErr != nil {
			opt = -3
			msg := fmt.Sprintf("syncOne updateOne Err key:%v err:%v", key, updErr.Error())
			logger.Error(msg)
		} else if rst.ModifiedCount > 0 {
			opt = 3
		} else {
			filter := bson.M{"_id": _id, "list._id": bson.M{"$ne": member}}
			item := &CacheZSetListItem{Member: member, Score: score}
			rst, addErr := coll.UpdateOne(ctx, filter, bson.M{
				"$addToSet": bson.M{"list": item},
				"$set":      bson.M{"update_time": time.Now().Unix()},
			})
			if addErr != nil {
				opt = -2
				msg := fmt.Sprintf("syncOne addToSet Err key:%v err:%v", key, addErr.Error())
				logger.Error(msg)
			} else if rst.ModifiedCount > 0 {
				opt = 2
				insert = 1
			} else {
				opt = 1
				data := &CacheZSetData{
					Id: _id,
					List: []*CacheZSetListItem{
						item,
					},
					CreateTime: time.Now().Unix(),
					UpdateTime: time.Now().Unix(),
				}
				_, intErr := coll.InsertOne(ctx, data)
				if err != nil {
					opt = -1
					msg := fmt.Sprintf("syncOne insert Err key:%v err:%v", key, intErr.Error())
					logger.Error(msg)
					err = errors.New(msg)
					return
				} else {
					insert = 1
				}
			}
		}
	}
	logger.Infof("syncOne name=%v key=%v member=%v score:%v opt:%v err:%v", d.options.uniqName, key, member, score, opt, err)
	needRemoveIfErr = false
	syncKey := d.getSyncKey()
	_, err = d.options.cache.Eval(ctx, syncCacheScriptByScore, []string{syncKey}, key, version).Int()
	if err != nil {
		msg := fmt.Sprintf("syncOne fail to eval name:%v key:%v err:%v", d.options.uniqName, key, err)
		logger.Error(msg)
		err = errors.New(msg)
		return
	}
	return
}
