# 统一配置迁移指南

## 目标
将所有 `owlBack` 项目的配置统一到 `owl-common/config` 包中，减少重复代码，提高一致性。

## 当前状态

### 已统一到 `owl-common/config` 的配置
1. **DatabaseConfig** - 数据库配置
2. **RedisConfig** - Redis 配置
3. **MQTTConfig** - MQTT 配置
4. **HTTPConfig** - HTTP 服务器配置
5. **LogConfig** - 日志配置
6. **CacheConfig** - 缓存配置
7. **StreamConfig** - Redis Stream 配置
8. **AlarmConfig** - 报警配置

### 各项目特有配置
- **wisefido-data**: SleepaceConfig, RadarConfig, CardManageConfig, IoTTimeSeriesConfig
- **wisefido-radar**: 雷达设备 MQTT 配置，HTTPS 配置
- **wisefido-qinglan**: 雷达设备 MQTT 主题配置，订阅配置

## 迁移步骤

### 步骤1：更新依赖
确保项目的 `go.mod` 包含：
```go
require owl-common v0.0.0
replace owl-common => ../owl-common
```

### 步骤2：更新导入
```go
import commonconfig "owl-common/config"
```

### 步骤3：更新配置结构
```go
// 之前（各项目自定义）
type Config struct {
    DB struct {
        Host     string
        Port     int
        User     string
        // ...
    }
    Redis struct {
        Addr     string
        Password string
        DB       int
    }
    // ...
}

// 之后（使用统一配置）
type Config struct {
    DB    commonconfig.DatabaseConfig
    Redis commonconfig.RedisConfig
    HTTP  commonconfig.HTTPConfig
    Log   commonconfig.LogConfig
    // 项目特有配置
    ProjectSpecific ProjectSpecificConfig
}
```

### 步骤4：更新配置加载
```go
// 使用统一的环境变量加载方法
cfg.DB.LoadFromEnv("DB")
cfg.Redis.LoadFromEnv("REDIS")
cfg.HTTP.LoadFromEnv("HTTP")
cfg.Log.LoadFromEnv("LOG")
```

### 步骤5：更新使用方式
```go
// 之前
db, err := sql.Open("postgres", 
    fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
        cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.DBName))

// 之后
db, err := sql.Open("postgres", cfg.DB.GetDSN())
```

## 示例：wisefido-qinglan 迁移

### 迁移前
```go
// internal/config/config.go
type DBConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    Password string `yaml:"password"`
    DBName   string `yaml:"dbname"`
    SSLMode  string `yaml:"sslmode"`
}

type RedisConfig struct {
    Address  string `yaml:"address"`
    Password string `yaml:"password"`
    DB       int    `yaml:"db"`
}
```

### 迁移后
```go
// internal/config/config.go
import commonconfig "owl-common/config"

type Config struct {
    DB    commonconfig.DatabaseConfig  `yaml:"database"`
    Redis commonconfig.RedisConfig     `yaml:"redis"`
    HTTP  commonconfig.HTTPConfig      `yaml:"http"`
    Log   commonconfig.LogConfig       `yaml:"log"`
    // 青兰特有配置
    MQTT         MQTTConfig
    Subscription SubscriptionConfig
}
```

## 环境变量命名规范

使用统一的环境变量前缀：
- `DB_` - 数据库配置
- `REDIS_` - Redis 配置
- `HTTP_` - HTTP 服务器配置
- `LOG_` - 日志配置
- `MQTT_` - MQTT 配置
- 项目特定配置使用项目前缀，如 `QINGLAN_`

## 配置优先级
1. 环境变量（最高优先级）
2. 配置文件（YAML）
3. 代码默认值（最低优先级）

## 验证步骤

### 1. 编译验证
```bash
go build ./...
```

### 2. 配置加载验证
```bash
# 测试环境变量加载
export DB_HOST=localhost
export DB_PORT=5432
export REDIS_ADDR=localhost:6379
go run main.go

# 测试配置文件加载
CONFIG_PATH=config.yaml go run main.go
```

### 3. 功能验证
- 数据库连接正常
- Redis 连接正常
- HTTP 服务器正常启动
- 日志输出正常

## 后续优化建议

### 1. 添加配置验证
在 `owl-common/config` 中添加配置验证方法：
```go
func (c *DatabaseConfig) Validate() error {
    if c.Host == "" {
        return errors.New("database host is required")
    }
    if c.Port <= 0 {
        return errors.New("database port must be positive")
    }
    // ...
    return nil
}
```

### 2. 添加配置合并
支持从多个源合并配置：
```go
func MergeConfigs(defaults, fileConfig, envConfig *Config) *Config {
    // 合并逻辑
}
```

### 3. 添加配置热重载
支持配置热重载，无需重启服务。

## 注意事项

1. **向后兼容**：确保现有配置文件和代码仍然工作
2. **测试充分**：迁移后进行全面测试
3. **文档更新**：更新项目文档中的配置说明
4. **团队沟通**：确保所有开发人员了解新的配置结构

## 参考项目

- `wisefido-qinglan` - 已完成迁移
- `wisefido-data` - 部分使用统一配置
- `wisefido-radar` - 部分使用统一配置

通过统一配置，我们可以：
1. 减少代码重复
2. 提高配置一致性
3. 简化新项目配置
4. 便于维护和更新