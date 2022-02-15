package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"gstore/src/data/logger"
)

type CacheZSet struct {
	options Options
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

	d.options.needSync = false
	return nil
}
func (d *CacheZSet) Name() string {
	return d.options.uniqName
}
func (d *CacheZSet) Sync() error {
	return nil
}
func (d *CacheZSet) IsNeedSync() bool {
	return false
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
func (d *CacheZSet) ZRemRangeByRank(start, stop int64) (int64, error) {
	return d.ZRemRangeByRankEx(nil, start, stop)
}
func (d *CacheZSet) ZRemRangeByScore(min, max string) (int64, error) {
	return d.ZRemRangeByScoreEx(nil, min, max)
}
func (d *CacheZSet) ZPopMin(count int64) (interface{}, []float64, error) {
	return d.ZPopMinEx(nil, count)
}
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
	saveData, err := d.convertData(member)
	if err != nil {
		logger.Errorf("ZRem convert err key:%v err:%v val:%+v", d.options.uniqName, err, member)
		return err
	}
	key := d.getCacheKey(id)
	err = d.options.cache.ZRem(context.Background(), key, saveData).Err()
	if err != nil {
		logger.Errorf("ZRem err key:%v err:%v val:%+v", d.options.uniqName, err, member)
	}
	return err
}
func (d *CacheZSet) ZCardEx(id interface{}) (int64, error) {
	key := d.getCacheKey(id)
	count, err := d.options.cache.ZCard(context.Background(), key).Result()
	if err != nil {
		logger.Errorf("ZCard err key:%v err:%v", d.options.uniqName, err)
	}
	return count, err
}
func (d *CacheZSet) ZCountEx(id interface{}, min, max string) (int64, error) {
	key := d.getCacheKey(id)
	count, err := d.options.cache.ZCount(context.Background(), key, min, max).Result()
	if err != nil {
		logger.Errorf("ZCard err key:%v err:%v", d.options.uniqName, err)
	}
	return count, err
}
func (d *CacheZSet) ZScoreEx(id, member interface{}) (float64, error) {
	saveData, err := d.convertData(member)
	if err != nil {
		return 0, err
	}
	key := d.getCacheKey(id)
	score, err := d.options.cache.ZScore(context.Background(), key, saveData).Result()
	if err != nil {
		logger.Errorf("ZScore err key:%v err:%v val:%+v", d.options.uniqName, err, member)
	}
	return score, err
}
func (d *CacheZSet) ZRankEx(id, member interface{}) (int64, error) {
	saveData, err := d.convertData(member)
	if err != nil {
		return 0, err
	}
	key := d.getCacheKey(id)
	rank, err := d.options.cache.ZRank(context.Background(), key, saveData).Result()
	if err != nil {
		logger.Errorf("ZRank err key:%v err:%v val:%+v", d.options.uniqName, err, member)
	}
	return rank, err
}
func (d *CacheZSet) ZRevRankEx(id, member interface{}) (int64, error) {
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
	key := d.getCacheKey(id)
	count, err := d.options.cache.ZRemRangeByRank(context.Background(), key, start, stop).Result()
	if err != nil {
		logger.Errorf("ZRemRangeByRank err key:%v err:%v start:%v stop:%v", d.options.uniqName, err, start, stop)
	}
	return count, err
}
func (d *CacheZSet) ZRemRangeByScoreEx(id interface{}, min, max string) (int64, error) {
	key := d.getCacheKey(id)
	count, err := d.options.cache.ZRemRangeByScore(context.Background(), key, min, max).Result()
	if err != nil {
		logger.Errorf("ZRemRangeByScore err key:%v err:%v min:%v max:%v", d.options.uniqName, err, min, max)
	}
	return count, err
}
func (d *CacheZSet) ZPopMinEx(id interface{}, count int64) (interface{}, []float64, error) {
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
	}

	return score, nil
}
