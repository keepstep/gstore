package store

type ITData interface {
	TableName(id interface{}) string
	DefaultData(id interface{}) interface{}
}

type IZSetTData interface {

}

type IMemoryTData interface {
	MapItemExpired(key, value interface{}) bool
	ListItemExpired(index int, value interface{}) bool
}
