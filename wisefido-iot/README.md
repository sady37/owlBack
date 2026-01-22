# wisefido-iot 服务

IoT 时序数据服务，从 Redis Streams 消费数据并存储到 TimescaleDB。

## 排障工具：iot-inspect

排障工具统一位于 `owlBack/tools/iot-inspect`，使用方法如下：

### 使用方法

#### 从 Redis Stream 读取数据

```bash
# 从 Redis Streams 读取最新的 monitor 数据
cd owlBack/tools/iot-inspect
go run main.go --source redis --stream iot:monitor:stream --count 1

# 或者使用编译后的二进制文件
./iot-inspect --source redis --stream iot:monitor:stream --count 1

# 从 Redis Streams 读取最新的 stat 数据
go run main.go --source redis --stream iot:stat:stream --count 1

# 从 Redis Streams 读取最新的 event 数据
go run main.go --source redis --stream iot:event:stream --count 1

# 从 Redis Streams 读取最新的 alarm 数据
go run main.go --source redis --stream iot:alarm:stream --count 1

# 读取多条数据
go run main.go --source redis --stream iot:monitor:stream --count 5
```

#### 从数据库读取数据

```bash
# 从数据库读取最新的 monitor 数据
go run main.go --source db --topic-type monitor --limit 1

# 从数据库读取最新的 stat 数据
go run main.go --source db --topic-type stat --limit 1

# 从数据库读取最新的 event 数据
go run main.go --source db --topic-type event --limit 1

# 从数据库读取最新的 alarm 数据
go run main.go --source db --topic-type alarm --limit 1

# 从数据库读取所有类型的数据（不指定 topic-type）
go run main.go --source db --limit 10

# 按设备ID过滤
go run main.go --source db --device-id <device_id> --limit 5
```

#### 环境变量配置

如果需要连接 Redis（带密码）：

```bash
export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=TeLunSu-36kr
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
