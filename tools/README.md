# OwlBack 排障工具

统一在 `owlBack/tools/` 目录下的检测工具。

## iot-inspect

IoT 时序数据排障工具，用于查看 Redis Stream 和数据库中的数据。

### 使用方法

#### 从 Redis Stream 读取数据

```bash
# 进入工具目录
cd owlBack/tools

# 从 Redis Streams 读取最新的 monitor 数据
go run iot-inspect.go --source redis --stream iot:monitor:stream --count 1

# 从 Redis Streams 读取最新的 stat 数据
go run iot-inspect.go --source redis --stream iot:stat:stream --count 1

# 从 Redis Streams 读取最新的 event 数据
go run iot-inspect.go --source redis --stream iot:event:stream --count 1

# 从 Redis Streams 读取最新的 alarm 数据
go run iot-inspect.go --source redis --stream iot:alarm:stream --count 1

# 读取多条数据
go run iot-inspect.go --source redis --stream iot:monitor:stream --count 5

# 或使用编译后的二进制文件
./iot-inspect --source redis --stream iot:monitor:stream --count 1
```

#### 从数据库读取数据

```bash
# 从数据库读取最新的 monitor 数据
go run iot-inspect.go --source db --topic-type monitor --limit 1

# 从数据库读取最新的 stat 数据
go run iot-inspect.go --source db --topic-type stat --limit 1

# 从数据库读取最新的 event 数据
go run iot-inspect.go --source db --topic-type event --limit 1

# 从数据库读取最新的 alarm 数据
go run iot-inspect.go --source db --topic-type alarm --limit 1

# 从数据库读取所有类型的数据（不指定 topic-type）
go run iot-inspect.go --source db --limit 10

# 按设备ID过滤
go run iot-inspect.go --source db --device-id <device_id> --limit 5
```

#### 环境变量配置

工具默认使用以下配置，可通过环境变量覆盖：

```bash
# Redis 配置（默认值：localhost:6379，密码：TeLunSu-36kr）
export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=TeLunSu-36kr  # 如果密码不同，需要设置此变量
```

数据库配置（默认值）：

```bash
export DB_HOST=localhost
export DB_PORT=5433
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd
```

#### 使用 Docker 直接查看 Redis Stream

如果 Redis 要求密码，可以使用 docker 命令直接查看：

```bash
docker exec owl-redis redis-cli -a TeLunSu-36kr --raw XRANGE iot:monitor:stream - + COUNT 1
docker exec owl-redis redis-cli -a TeLunSu-36kr --raw XRANGE iot:stat:stream - + COUNT 1
docker exec owl-redis redis-cli -a TeLunSu-36kr --raw XRANGE iot:event:stream - + COUNT 1
docker exec owl-redis redis-cli -a TeLunSu-36kr --raw XRANGE iot:alarm:stream - + COUNT 1
```

#### 命令行参数说明

- `--source`: 数据源，`redis` 或 `db`（默认：`redis`）
- `--stream`: Redis Stream 名称（仅当 `source=redis` 时使用，默认：`iot:monitor:stream`）
- `--count`: 读取的消息数量（仅当 `source=redis` 时使用，默认：1）
- `--topic-type`: topic_type 过滤（仅当 `source=db` 时使用，可选：`monitor`, `stat`, `event`, `alarm`）
- `--device-id`: 设备ID 过滤（仅当 `source=db` 时使用，可选）
- `--limit`: 读取的记录数量（仅当 `source=db` 时使用，默认：1）
