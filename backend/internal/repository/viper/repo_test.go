package viper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	Name    string `mapstructure:"name"`
	AgeNum  int    `mapstructure:"age_num"`
	Enabled bool   `mapstructure:"enabled"`
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
