# OwlBack 统一配置文件说明

## 配置文件位置

统一配置文件位于 `owlBack/.env`（根目录）

## 使用方法

### 1. 创建配置文件

```bash
cd /Users/sady3721/project/owlBack
cp .env.example .env
```

### 2. 编辑配置文件

编辑 `.env` 文件，修改配置值：

```bash
vim .env
# 或
nano .env
```

### 3. 启动服务

启动脚本会自动加载 `.env` 文件：

```bash
# 启动所有后台服务
./start_owlback.sh

# 启动雷达服务
cd wisefido-radar
./start-radar.sh
```

## 配置优先级

1. **环境变量**（最高优先级）- 在命令行中设置的变量
2. **`.env` 文件** - 统一配置文件
3. **脚本默认值**（最低优先级）- 如果以上都没有，使用脚本中的默认值

## 配置项说明

### 数据库配置
- `DB_HOST` - 数据库主机地址
- `DB_PORT` - 数据库端口（默认 5433）
- `DB_USER` - 数据库用户名
- `DB_PASSWORD` - 数据库密码
- `DB_NAME` - 数据库名称
- `DB_SSLMODE` - SSL 模式

### Redis 配置
- `REDIS_ADDR` - Redis 地址（格式：host:port）
- `REDIS_PASSWORD` - Redis 密码
- `REDIS_DB` - Redis 数据库编号

### MQTT 配置
- `MQTT_BROKER` - MQTT Broker 地址（服务连接）
- `MQTT_CLIENT_ID` - MQTT 客户端 ID
- `RADAR_MQTT_SERVER` - 返回给设备的 MQTT 服务器地址
- `RADAR_MQTT_PORT` - 返回给设备的 MQTT 端口

### 日志配置
- `LOG_LEVEL` - 日志级别（debug, info, warn, error）
- `LOG_FORMAT` - 日志格式（json, text）

## 注意事项

1. `.env` 文件包含敏感信息，已添加到 `.gitignore`，不会被提交到 Git
2. `.env.example` 是模板文件，可以提交到 Git
3. 如果 `.env` 文件不存在，脚本会使用默认值
4. 可以通过环境变量覆盖 `.env` 文件中的配置

## 示例

```bash
# 使用 .env 文件中的配置
./start_owlback.sh

# 覆盖特定配置
DB_HOST=192.168.1.100 ./start_owlback.sh

# 使用不同的配置文件
ENV_FILE=/path/to/custom.env source load_env.sh
```
