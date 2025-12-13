package config

import (
	"os"
	"testing"
)

func TestLoad_DefaultValues(t *testing.T) {
	// 清除环境变量
	os.Clearenv()
	
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// 检查默认值
	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected DB_HOST default 'localhost', got '%s'", cfg.Database.Host)
	}
	
	if cfg.Database.Port != 5432 {
		t.Errorf("Expected DB_PORT default 5432, got %d", cfg.Database.Port)
	}
	
	if cfg.Database.User != "postgres" {
		t.Errorf("Expected DB_USER default 'postgres', got '%s'", cfg.Database.User)
	}
	
	if cfg.Database.Database != "owlrd" {
		t.Errorf("Expected DB_NAME default 'owlrd', got '%s'", cfg.Database.Database)
	}
	
	if cfg.Redis.Addr != "localhost:6379" {
		t.Errorf("Expected REDIS_ADDR default 'localhost:6379', got '%s'", cfg.Redis.Addr)
	}
	
	if cfg.Aggregator.TriggerMode != "polling" {
		t.Errorf("Expected CARD_TRIGGER_MODE default 'polling', got '%s'", cfg.Aggregator.TriggerMode)
	}
	
	if cfg.Aggregator.Polling.Interval != 60 {
		t.Errorf("Expected polling interval default 60, got %d", cfg.Aggregator.Polling.Interval)
	}
	
	if cfg.Log.Level != "info" {
		t.Errorf("Expected LOG_LEVEL default 'info', got '%s'", cfg.Log.Level)
	}
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// 设置环境变量
	os.Setenv("DB_HOST", "test-host")
	os.Setenv("DB_USER", "test-user")
	os.Setenv("DB_PASSWORD", "test-password")
	os.Setenv("DB_NAME", "test-db")
	os.Setenv("TENANT_ID", "test-tenant-id")
	os.Setenv("CARD_TRIGGER_MODE", "events")
	os.Setenv("LOG_LEVEL", "debug")
	
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("TENANT_ID")
		os.Unsetenv("CARD_TRIGGER_MODE")
		os.Unsetenv("LOG_LEVEL")
	}()
	
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// 检查环境变量值
	if cfg.Database.Host != "test-host" {
		t.Errorf("Expected DB_HOST 'test-host', got '%s'", cfg.Database.Host)
	}
	
	if cfg.Database.User != "test-user" {
		t.Errorf("Expected DB_USER 'test-user', got '%s'", cfg.Database.User)
	}
	
	if cfg.Database.Password != "test-password" {
		t.Errorf("Expected DB_PASSWORD 'test-password', got '%s'", cfg.Database.Password)
	}
	
	if cfg.Database.Database != "test-db" {
		t.Errorf("Expected DB_NAME 'test-db', got '%s'", cfg.Database.Database)
	}
	
	if cfg.Aggregator.TenantID != "test-tenant-id" {
		t.Errorf("Expected TENANT_ID 'test-tenant-id', got '%s'", cfg.Aggregator.TenantID)
	}
	
	if cfg.Aggregator.TriggerMode != "events" {
		t.Errorf("Expected CARD_TRIGGER_MODE 'events', got '%s'", cfg.Aggregator.TriggerMode)
	}
	
	if cfg.Log.Level != "debug" {
		t.Errorf("Expected LOG_LEVEL 'debug', got '%s'", cfg.Log.Level)
	}
}

func TestGetEnv(t *testing.T) {
	// 测试环境变量存在
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")
	
	value := getEnv("TEST_VAR", "default")
	if value != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", value)
	}
	
	// 测试环境变量不存在，使用默认值
	value = getEnv("NON_EXISTENT_VAR", "default-value")
	if value != "default-value" {
		t.Errorf("Expected 'default-value', got '%s'", value)
	}
}

