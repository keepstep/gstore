package data

import (
	"gstore/src/data/config"
	"gstore/src/data/model"
)

var dataConfigList = map[string]*config.DataConfig{
	config.Name_userinfo:              &model.CfgUserinfo,
}

var zsetDataConfigList = map[string]*config.ZSetDataConfig{

}

var memoryDataConfigList = map[string]*config.MemoryDataConfig{

}

func GetDataConfig(name string) *config.DataConfig {
	if r, ok := dataConfigList[name]; ok {
		return r
	}
	return nil
}

func GetZSetDataConfig(name string) *config.ZSetDataConfig {
	if r, ok := zsetDataConfigList[name]; ok {
		return r
	}
	return nil
}

func GetMemoryDataConfig(name string) *config.MemoryDataConfig {
	if r, ok := memoryDataConfigList[name]; ok {
		return r
	}
	return nil
}
