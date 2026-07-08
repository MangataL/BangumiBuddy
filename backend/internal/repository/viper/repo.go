package viper

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Reloadable 接口定义了支持配置重载的组件
type Reloadable interface {
	// Reload 根据提供的配置更新组件
	Reload(config interface{}) error
}

type ComponentName string

// Repo 是配置仓库
type Repo struct {
	mu   sync.Mutex
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
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reloadables[name] = reloadable
}

func (r *Repo) SetComponentConfig(name ComponentName, newConfig interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.setComponentConfigLocked(name, newConfig)
}

func (r *Repo) setComponentConfigLocked(name ComponentName, newConfig interface{}) error {
	if err := defaults.Set(newConfig); err != nil {
		return err
	}
	lastConfig, exists := r.lastConfigs[name]
	if !exists || !r.configEquals(lastConfig, newConfig) {
		var configMap map[string]interface{}
		if err := mapstructure.Decode(newConfig, &configMap); err != nil {
			return fmt.Errorf("convert config to map failed: %w", err)
		}
		oldValue, oldExists := r.file.Get(string(name)), r.file.IsSet(string(name))
		r.file.Set(string(name), configMap)
		if err := r.file.WriteConfig(); err != nil {
			_ = r.reloadFileLocked()
			return err
		}
		reloadable, ok := r.reloadables[name]
		if ok {
			// 配置发生变化，触发重载
			if err := reloadable.Reload(newConfig); err != nil {
				if rollbackErr := r.rollbackComponentLocked(name, oldValue, oldExists); rollbackErr != nil {
					return fmt.Errorf("reload config failed: %w, rollback failed: %v", err, rollbackErr)
				}
				return err
			}
		}
		r.lastConfigs[name] = newConfig
	}
	return nil
}

func (r *Repo) configEquals(oldConfig, newConfig interface{}) bool {
	return reflect.DeepEqual(oldConfig, newConfig)
}

func (r *Repo) GetComponentConfig(name ComponentName, config interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.file.IsSet(string(name)) {
		if err := r.setComponentConfigLocked(name, config); err != nil {
			return err
		}
	}

	return r.file.UnmarshalKey(string(name), &config)
}

func (r *Repo) rollbackComponentLocked(name ComponentName, oldValue interface{}, oldExists bool) error {
	settings := r.file.AllSettings()
	if oldExists {
		setDottedValue(settings, string(name), oldValue)
	} else {
		deleteDottedValue(settings, string(name))
	}
	data, err := yaml.Marshal(settings)
	if err != nil {
		return err
	}
	if err := os.WriteFile(r.path, data, 0644); err != nil {
		return err
	}
	return r.reloadFileLocked()
}

func (r *Repo) reloadFileLocked() error {
	file := viper.New()
	file.SetConfigFile(r.path)
	if err := file.ReadInConfig(); err != nil {
		return err
	}
	r.file = file
	return nil
}

func setDottedValue(settings map[string]interface{}, dottedKey string, value interface{}) {
	parts := strings.Split(dottedKey, ".")
	target := settings
	for _, part := range parts[:len(parts)-1] {
		next, ok := target[part].(map[string]interface{})
		if !ok {
			next = make(map[string]interface{})
			target[part] = next
		}
		target = next
	}
	target[parts[len(parts)-1]] = value
}

func deleteDottedValue(settings map[string]interface{}, dottedKey string) {
	parts := strings.Split(dottedKey, ".")
	target := settings
	for _, part := range parts[:len(parts)-1] {
		next, ok := target[part].(map[string]interface{})
		if !ok {
			return
		}
		target = next
	}
	delete(target, parts[len(parts)-1])
}
