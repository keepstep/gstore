package store

import (
	"errors"
	"gstore/src/data/logger"
	"reflect"
	"sync"
)

type CacheMemory struct {
	options Options

	mapMutex sync.RWMutex
	mmap     map[interface{}]interface{}

	listMutex sync.RWMutex
	mlist     []interface{}

	canAutoDel bool
}

func (d *CacheMemory) Init(opts ...Option) error {
	for _, o := range opts {
		o(&d.options)
	}
	if d.options.uniqName == "" {
		msg := "CacheMemory Init Err uniqName"
		logger.Error(msg)
		return errors.New(msg)
	}
	if d.options.tdata == nil {
		msg := "CacheMemory Init Err tdata nil name:" + d.options.uniqName
		logger.Error(msg)
		return errors.New(msg)
	}
	if d.options.tkeyKind == reflect.Invalid {
		msg := "CacheMemory Init Err tkeyKind invalid name:" + d.options.uniqName
		logger.Error(msg)
		return errors.New(msg)
	}
	if d.options.tdataKind == reflect.Invalid {
		msg := "CacheMemory Init Err tdataKind invalid name:" + d.options.uniqName
		logger.Error(msg)
		return errors.New(msg)
	}
	if _, ok := d.options.tdata.(IMemoryTData); ok {
		d.canAutoDel = true
	} else {
		d.canAutoDel = false
		msg := "CacheMemory Init no support audoDel name:" + d.options.uniqName
		logger.Error(msg)
	}

	d.options.needSync = false
	d.mmap = map[interface{}]interface{}{}
	d.mlist = []interface{}{}
	return nil
}

func (d *CacheMemory) Name() string {
	return d.options.uniqName
}
func (d *CacheMemory) Sync() error {
	if !d.canAutoDel {
		return nil
	}
	d.MapAutoDel()
	d.ListAutoDel()
	return nil
}
func (d *CacheMemory) IsNeedSync() bool {
	return d.canAutoDel
}

func (d *CacheMemory) checkKeyType(key interface{}) error {
	if reflect.TypeOf(key) != d.options.tkeyType {
		return errors.New("keyType diff")
	}
	return nil
}
func (d *CacheMemory) checkDataType(data interface{}) error {
	if reflect.TypeOf(data).Kind() == reflect.Ptr {
		//必须是结构体指针
		if reflect.ValueOf(data).Elem().Kind() != reflect.Struct {
			return errors.New("data kind must be ptr Struct")
		}
		if reflect.ValueOf(data).Elem().Type() != d.options.tdataType {
			return errors.New("dataType diff ptr")
		}
	} else if reflect.TypeOf(data) != d.options.tdataType {
		return errors.New("dataType diff")
	}
	return nil
}

func (d *CacheMemory) MapSet(key, data interface{}) error {
	err := d.checkKeyType(key)
	if err != nil {
		return err
	}
	err = d.checkDataType(data)
	if err != nil {
		return err
	}
	d.mapMutex.Lock()
	defer d.mapMutex.Unlock()
	d.mmap[key] = data
	return nil
}
func (d *CacheMemory) MapGet(key interface{}) (interface{}, bool) {
	err := d.checkKeyType(key)
	if err != nil {
		return nil, false
	}
	d.mapMutex.RLock()
	defer d.mapMutex.RUnlock()
	ii, ok := d.mmap[key]
	return ii, ok
}
func (d *CacheMemory) MapDel(key interface{}) error {
	err := d.checkKeyType(key)
	if err != nil {
		return err
	}
	d.mapMutex.Lock()
	defer d.mapMutex.Unlock()
	delete(d.mmap, key)
	return nil
}
func (d *CacheMemory) MapCount() int {
	d.mapMutex.RLock()
	defer d.mapMutex.RUnlock()
	return len(d.mmap)
}
func (d *CacheMemory) MapAutoDel() (int, error) {
	if !d.canAutoDel {
		return 0, nil
	}
	d.mapMutex.Lock()
	defer d.mapMutex.Unlock()
	count := 0
	for key, val := range d.mmap {
		if d.options.tdata.(IMemoryTData).MapItemExpired(key, val) {
			count++
			delete(d.mmap, key)
		}
	}
	return count, nil
}
func (d *CacheMemory) MapClean() error {
	d.mapMutex.Lock()
	defer d.mapMutex.Unlock()
	d.mmap = map[interface{}]interface{}{}
	return nil
}

func (d *CacheMemory) MapIter(iter MemoryStoreMapIter) int {
	d.mapMutex.RLock()
	defer d.mapMutex.RUnlock()
	count := 0
	for key, val := range d.mmap {
		count++
		if iter(key, val) {
			break
		}
	}
	return count
}

func (d *CacheMemory) ListInsert(index int, member interface{}) error {
	return nil
}
func (d *CacheMemory) ListDel(index int) error {
	return nil
}
func (d *CacheMemory) ListCount() int {
	return 0
}
func (d *CacheMemory) ListGet(index int) (interface{}, error) {
	return 0, nil
}
func (d *CacheMemory) ListAutoDel() (int, error) {
	return 0, nil
}
func (d *CacheMemory) ListClean() error {
	return nil
}

func (d *CacheMemory) Clean() error {
	return nil
}
