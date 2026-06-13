# OwlBack 排障工具

统一在 `owlBack/tools/` 目录下的检测工具。

## case fixture 三件套（导出再重放）

围绕统一 fixture 契约（`window.json` + `meta.json{device_uid,device_addr,device_type}`），导出与重放分离，fixture 有两条来源（PG 真实 / 合成），重放只吃文件、不管来源。DB/Redis 密码单源 `owlBack/.env`。

**职责铁律**：export 触库、replay 只读文件，两者绝不互越。`replay/` 不 import 任何 DB 包（编译期证明），uid→addr→type 全经 `meta.json`（导出侧权威写入）。回放 alarm 流走 `alarm.json`（export 已导设备直发 alarm）。

| 工具 | 职责 | 来源/去向 |
|------|------|-----------|
| `export/export.sh` | PG → fixture（**unit 级**：从主设备推 /64，导整个 unit 全活跃设备：多 radar + sleepad + alarm） | PG → 文件 |
| `simulate-make/` | lego 块合成 → fixture（AI 灵活造 fall/ghost 场景，不连 DB） | 合成 → 文件 |
| `replay/` | fixture → rebase ts → 灌 Redis → live sensor 真消费跑完整链（track + belief shadow） | 文件 → live |

```bash
cd owlBack/tools
# 导出（case 名自动解析 uid/时段；产出 window+window_sleepad+alarm+meta+room_layout）：
./export/export.sh case-cd2b-0606-10271037 --tz America/Denver
# 合成：
go run ./simulate-make/ --scenario open-floor-fall --uid SIM0000FALL --addr <真实绑定addr> --out ../doc/cases/sim-fall
# 重放（只读文件，灌 live；含 alarm 流）：
go run ./replay/ --fixture ../doc/cases/case-cd2b-0606-10271037 --streams monitor,event,alarm
# 加速重放（仅验数据流连通性，不能验 fire）：
go run ./replay/ --fixture ../doc/cases/case-cd2b-0606-10271037 --speed 30
```

> `replay/` 原名 `redis-replay`，2026-06-12 改名加 `--fixture`，并删除 DB 直查模式（only 文件源，符合 export/replay 职责分离）。
>
> ⚠️ **`--speed` 与 fall confirm 的时间语义**：belief `Decider` 的确认窗 `confirmMs=90s` 按 frame TMs 计，而 replay 把 frame TMs 重写成灌入墙钟（`/speed` 压缩间隔）。`--speed 30` 会把 confirmMs 压成 1/30，**任何 case 加速重放都测不到 fire confirm**。**验 fire 行为必须 `--speed 1`**（默认值）；加速重放仅用于验 track / belief 演化（p_fallen 轨迹、ghost_veto），不能验 confirm。

## iot-inspect

IoT 时序数据排障工具，用于查看 Redis Stream 和数据库中的数据。

### 使用方法

#### 从 Redis Stream 读取数据

```bash
# 进入工具目录（源码在 iot-inspect-dir/）
cd owlBack/tools

# 从 Redis Streams 读取最新的 monitor 数据
go run ./iot-inspect-dir/ --source redis --stream iot:monitor:stream --count 1

# 从 Redis Streams 读取最新的 stat 数据
go run ./iot-inspect-dir/ --source redis --stream iot:stat:stream --count 1

# 从 Redis Streams 读取最新的 event 数据
go run ./iot-inspect-dir/ --source redis --stream iot:event:stream --count 1

# 从 Redis Streams 读取最新的 alarm 数据
go run ./iot-inspect-dir/ --source redis --stream iot:alarm:stream --count 1

# 读取多条数据
go run ./iot-inspect-dir/ --source redis --stream iot:monitor:stream --count 5

# 或 go build -o iot-inspect ./iot-inspect-dir/ && ./iot-inspect ...
```

#### 从数据库读取数据

```bash
# 从数据库读取最新的 monitor 数据
go run ./iot-inspect-dir/ --source db --topic-type monitor --limit 1

# 从数据库读取最新的 stat 数据
go run ./iot-inspect-dir/ --source db --topic-type stat --limit 1

# 从数据库读取最新的 event 数据
go run ./iot-inspect-dir/ --source db --topic-type event --limit 1

# 从数据库读取最新的 alarm 数据
go run ./iot-inspect-dir/ --source db --topic-type alarm --limit 1

# 从数据库读取所有类型的数据（不指定 topic-type）
go run ./iot-inspect-dir/ --source db --limit 10

# 按设备ID过滤
go run ./iot-inspect-dir/ --source db --device-id <device_id> --limit 5
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
export DB_PORT=5432
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
