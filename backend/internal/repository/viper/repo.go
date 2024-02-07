package viper

import (
	"fmt"
	"reflect"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Reloadable 接口定义了支持配置重载的组件
type Reloadable interface {
	// Reload 根据提供的配置更新组件
	Reload(config interface{}) error
}

type ComponentName string

// Repo 是配置仓库
type Repo struct {
	file *viper.Viper
	path string
	// 存储上一次的组件配置
	reloadables map[ComponentName]Reloadable
	lastConfigs map[ComponentName]interface{}
}

func NewRepo(path string) (*Repo, error) {
	file := viper.New()
	file.SetConfigFile(path)
	if err := file.ReadInConfig(); err != nil {
		return nil, err
	}
	repo := &Repo{
		file:        file,
		path:        path,
		lastConfigs: make(map[ComponentName]interface{}),
		reloadables: make(map[ComponentName]Reloadable),
	}
	return repo, nil
}

func (r *Repo) RegisterReloadable(name ComponentName, reloadable Reloadable) {
	r.reloadables[name] = reloadable
}

func (r *Repo) SetComponentConfig(name ComponentName, newConfig interface{}) error {
	if err := defaults.Set(newConfig); err != nil {
		return err
	}
	lastConfig, exists := r.lastConfigs[name]
	if !exists || !r.configEquals(lastConfig, newConfig) {
		reloadable, ok := r.reloadables[name]
		if ok {
			// 配置发生变化，触发重载
			if err := reloadable.Reload(newConfig); err != nil {
				return err
			}
		}
		// 将结构体转换为map，保留yaml tag信息
		var configMap map[string]interface{}
		if err := mapstructure.Decode(newConfig, &configMap); err != nil {
			return fmt.Errorf("convert config to map failed: %w", err)
		}
		r.file.Set(string(name), configMap)
		return r.file.WriteConfig()
	}
	return nil
}

func (r *Repo) configEquals(oldConfig, newConfig interface{}) bool {
	return reflect.DeepEqual(oldConfig, newConfig)
}

func (r *Repo) GetComponentConfig(name ComponentName, config interface{}) error {
	if !r.file.IsSet(string(name)) {
		r.SetComponentConfig(name, config)
	}

	return r.file.UnmarshalKey(string(name), &config)
}