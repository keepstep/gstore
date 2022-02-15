package store

const syncCacheScript = `
	local sync_key = KEYS[1]
	local base_key = ARGV[1]
	local version = ARGV[2]
	if redis.call('hget',sync_key,base_key) == version then
		redis.call('hdel',sync_key,base_key)
		return 1
	end
	return 0
`

const syncCacheScriptByScore = `
	local sync_key = KEYS[1]
	local base_key = ARGV[1]
	local score = ARGV[2]
	if redis.call('zscore',sync_key,base_key) == score then
		redis.call('zrem',sync_key,base_key)
		return 1
	end
	return 0
`

type StoreSync interface {
	Name() string
	IsNeedSync() bool
	Sync() error
}

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

type ZSetStore interface {
	Init(opts ...Option) error
	Name() string
	Sync() error
	IsNeedSync() bool

	ZAdd(score float64, member interface{}) error
	ZRem(member interface{}) error
	ZCard() (int64, error)
	ZCount(min, max string) (int64, error)
	ZScore(member interface{}) (float64, error)
	ZRank(member interface{}) (int64, error)
	ZRevRank(member interface{}) (int64, error)
	ZRange(start, stop int64) (interface{}, error)
	ZRevRange(start, stop int64) (interface{}, error)
	ZRangeByScore(min, max string, offset, count int64) (interface{}, error)
	ZRevRangeByScore(min, max string, offset, count int64) (interface{}, error)
	ZRemRangeByRank(start, stop int64) (int64, error)
	ZRemRangeByScore(min, max string) (int64, error)
	ZPopMin(count int64) (interface{}, []float64, error)
	ZPopMax(count int64) (interface{}, []float64, error)
	ZIncrBy(increment float64, member interface{}) (float64, error)

	ZRangeWithScores(start, stop int64) (interface{}, []float64, error)
	ZRevRangeWithScores(start, stop int64) (interface{}, []float64, error)
	ZRangeByScoreWithScores(min, max string, offset, count int64) (interface{}, []float64, error)
	ZRevRangeByScoreWithScores(min, max string, offset, count int64) (interface{}, []float64, error)

	ZAddEx(id interface{}, score float64, member interface{}) error
	ZRemEx(id, member interface{}) error
	ZCardEx(id interface{}) (int64, error)
	ZCountEx(id interface{}, min, max string) (int64, error)
	ZScoreEx(id, member interface{}) (float64, error)
	ZRankEx(id, member interface{}) (int64, error)
	ZRevRankEx(id, member interface{}) (int64, error)
	ZRangeEx(id interface{}, start, stop int64) (interface{}, error)
	ZRevRangeEx(id interface{}, start, stop int64) (interface{}, error)
	ZRangeByScoreEx(id interface{}, min, max string, offset, count int64) (interface{}, error)
	ZRevRangeByScoreEx(id interface{}, min, max string, offset, count int64) (interface{}, error)
	ZRemRangeByRankEx(id interface{}, start, stop int64) (int64, error)
	ZRemRangeByScoreEx(id interface{}, min, max string) (int64, error)
	ZPopMinEx(id interface{}, count int64) (interface{}, []float64, error)
	ZPopMaxEx(id interface{}, count int64) (interface{}, []float64, error)

	ZRangeWithScoresEx(id interface{}, start, stop int64) (interface{}, []float64, error)
	ZRevRangeWithScoresEx(id interface{}, start, stop int64) (interface{}, []float64, error)
	ZRangeByScoreWithScoresEx(id interface{}, min, max string, offset, count int64) (interface{}, []float64, error)
	ZRevRangeByScoreWithScoresEx(id interface{}, min, max string, offset, count int64) (interface{}, []float64, error)
	ZIncrByEx(id interface{}, increment float64, member interface{}) (float64, error)
}
type MemoryStoreMapIter func(key, member interface{}) (stop bool)
type MemoryStore interface {
	Init(opts ...Option) error
	Name() string
	Sync() error
	IsNeedSync() bool

	MapSet(key, member interface{}) error
	MapGet(key interface{}) (interface{}, bool)
	MapDel(key interface{}) error
	MapCount() int
	MapAutoDel() (int, error)
	MapClean() error
	MapIter(iter MemoryStoreMapIter) int

	ListInsert(index int, member interface{}) error
	ListDel(index int) error
	ListGet(index int) (interface{}, error)
	ListCount() int
	ListAutoDel() (int, error)
	ListClean() error

	Clean() error
}
