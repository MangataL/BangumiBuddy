package viper

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	Name    string `mapstructure:"name"`
	AgeNum  int    `mapstructure:"age_num"`
	Enabled bool   `mapstructure:"enabled"`
}

type failingReloadable struct {
	err error
}

func (r failingReloadable) Reload(config interface{}) error {
	return r.err
}

func TestRepo_SetAndGetComponentConfig(t *testing.T) {
	// 创建临时配置文件
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// 初始化空的yaml文件
	err = os.WriteFile(tmpFile.Name(), []byte("{}"), 0644)
	require.NoError(t, err)

	// 创建配置仓库
	repo, err := NewRepo(tmpFile.Name())
	require.NoError(t, err)

	// 测试写入配置
	testConfig := &TestConfig{
		Name:    "test",
		AgeNum:  18,
		Enabled: true,
	}

	err = repo.SetComponentConfig("test_component", testConfig)
	require.NoError(t, err)

	// 测试读取配置
	var readConfig TestConfig
	err = repo.GetComponentConfig("test_component", &readConfig)
	require.NoError(t, err)

	// 验证配置是否一致
	assert.Equal(t, testConfig.Name, readConfig.Name)
	assert.Equal(t, testConfig.AgeNum, readConfig.AgeNum)
	assert.Equal(t, testConfig.Enabled, readConfig.Enabled)

	// 测试更新配置
	updatedConfig := &TestConfig{
		Name:    "updated",
		AgeNum:  20,
		Enabled: false,
	}

	err = repo.SetComponentConfig("test_component", updatedConfig)
	require.NoError(t, err)

	// 测试读取更新后的配置
	var readUpdatedConfig TestConfig
	err = repo.GetComponentConfig("test_component", &readUpdatedConfig)
	require.NoError(t, err)

	// 验证更新后的配置是否一致
	assert.Equal(t, updatedConfig.Name, readUpdatedConfig.Name)
	assert.Equal(t, updatedConfig.AgeNum, readUpdatedConfig.AgeNum)
	assert.Equal(t, updatedConfig.Enabled, readUpdatedConfig.Enabled)
}

func TestRepo_SetComponentConfig_RollbackWhenReloadFails(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(tmpFile, []byte("{}"), 0644))

	repo, err := NewRepo(tmpFile)
	require.NoError(t, err)

	oldConfig := &TestConfig{Name: "old", AgeNum: 18, Enabled: true}
	require.NoError(t, repo.SetComponentConfig("test_component", oldConfig))
	repo.RegisterReloadable("test_component", failingReloadable{err: errors.New("reload failed")})

	newConfig := &TestConfig{Name: "new", AgeNum: 20, Enabled: false}
	err = repo.SetComponentConfig("test_component", newConfig)
	require.Error(t, err)

	var readConfig TestConfig
	require.NoError(t, repo.GetComponentConfig("test_component", &readConfig))
	assert.Equal(t, *oldConfig, readConfig)

	reloadedRepo, err := NewRepo(tmpFile)
	require.NoError(t, err)
	var diskConfig TestConfig
	require.NoError(t, reloadedRepo.GetComponentConfig("test_component", &diskConfig))
	assert.Equal(t, *oldConfig, diskConfig)
}

func TestRepo_SetComponentConfig_ConcurrentAccess(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(tmpFile, []byte("{}"), 0644))

	repo, err := NewRepo(tmpFile)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			config := &TestConfig{Name: "name", AgeNum: index, Enabled: index%2 == 0}
			require.NoError(t, repo.SetComponentConfig("test_component", config))
			var readConfig TestConfig
			require.NoError(t, repo.GetComponentConfig("test_component", &readConfig))
		}(i)
	}
	wg.Wait()
}
