package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultValues(t *testing.T) {
	// 清除环境变量
	os.Clearenv()

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// 验证默认值
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "postgres", cfg.Database.User)
	assert.Equal(t, "postgres", cfg.Database.Password)
	assert.Equal(t, "owlrd", cfg.Database.Database)
	assert.Equal(t, "disable", cfg.Database.SSLMode)

	assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
	assert.Equal(t, "", cfg.Redis.Password)
	assert.Equal(t, 0, cfg.Redis.DB)

	assert.Equal(t, "vital-focus:card:", cfg.Alarm.Cache.RealtimeKeyPrefix)
	assert.Equal(t, ":realtime", cfg.Alarm.Cache.RealtimeSuffix)
	assert.Equal(t, "vital-focus:card:", cfg.Alarm.Cache.AlarmKeyPrefix)
	assert.Equal(t, ":alarms", cfg.Alarm.Cache.AlarmSuffix)
	assert.Equal(t, 30, cfg.Alarm.Cache.AlarmTTL)
	assert.Equal(t, "alarm:state:", cfg.Alarm.Cache.StateKeyPrefix)

	assert.Equal(t, 5, cfg.Alarm.PollInterval)
	assert.Equal(t, 10, cfg.Alarm.Evaluation.BatchSize)

	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// 设置环境变量
	os.Setenv("DB_HOST", "test-host")
	os.Setenv("DB_USER", "test-user")
	os.Setenv("DB_PASSWORD", "test-password")
	os.Setenv("DB_NAME", "test-db")
	os.Setenv("REDIS_ADDR", "test-redis:6380")
	os.Setenv("REDIS_PASSWORD", "test-redis-password")
	os.Setenv("TENANT_ID", "test-tenant")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "text")

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// 验证环境变量覆盖
	assert.Equal(t, "test-host", cfg.Database.Host)
	assert.Equal(t, "test-user", cfg.Database.User)
	assert.Equal(t, "test-password", cfg.Database.Password)
	assert.Equal(t, "test-db", cfg.Database.Database)

	assert.Equal(t, "test-redis:6380", cfg.Redis.Addr)
	assert.Equal(t, "test-redis-password", cfg.Redis.Password)

	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "text", cfg.Log.Format)

	// 清理环境变量
	os.Clearenv()
}

func TestGetEnv(t *testing.T) {
	// 测试默认值
	os.Clearenv()
	value := getEnv("TEST_KEY", "default-value")
	assert.Equal(t, "default-value", value)

	// 测试环境变量存在
	os.Setenv("TEST_KEY", "env-value")
	value = getEnv("TEST_KEY", "default-value")
	assert.Equal(t, "env-value", value)

	// 清理
	os.Unsetenv("TEST_KEY")
}
