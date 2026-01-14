# 雷达数据解码测试工具

## 功能

从 Redis Streams 或数据库读取雷达数据，解码 base64 编码的 track 字段，并打印所有解码后的字段。

## 使用方法

### 1. 从 Redis Streams 读取并解码

```bash
# 读取 monitor 数据
go run ./cmd/decode-track --source redis --stream radar:monitor:stream --count 1

# 读取 stat 数据
go run ./cmd/decode-track --source redis --stream radar:stat:stream --count 1

# 读取 event 数据（event 数据不包含 track 字段）
go run ./cmd/decode-track --source redis --stream radar:event:stream --count 1
```

### 2. 直接解码 base64 track 字符串

```bash
go run ./cmd/decode-track --decode "AA8mAAoAAAAAAAAAAAQA/w=="
```

### 3. 从数据库读取（注意：数据库中的 raw_original 不包含 track 数据）

```bash
go run ./cmd/decode-track --source db --topic-type monitor --limit 1
```

## 环境变量

需要设置以下环境变量：

```bash
export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=TeLunSu-36kr
export DB_HOST=127.0.0.1
export DB_PORT=5433
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd
```

## 输出示例

### Monitor Track 解码

```
=== monitor Track Decoded ===
Device ID: 56534328-7960-4e74-a278-3f60c2cb0ce6
Timestamp: 2026-01-11T20:47:31+08:00
Topic Type: monitor

Track (Base64): AA8mAAoAAAAAAAAAAAQA/w==
Track (Hex): 000f26000a00000000000000000400ff

Person 0:
  target_id: 0
  position_x: 150 cm (原始值: 15 dm)
  position_y: 380 cm (原始值: 38 dm)
  position_z: 0 cm
  remaining_time: 0 sec
  pose: 4 (站立)
  event: 0 (无事件)
  area_id: 255
```

### Stat Track 解码

```
=== stat Track Decoded ===
Device ID: 56534328-7960-4e74-a278-3f60c2cb0ce6
Timestamp: 2026-01-11T20:47:14+08:00
Topic Type: stat

Track (Base64): AgAAAAAAAAAAAAAAAAAAAA==
Track (Hex): 02000000000000000000000000000000

Statistics Track Data:
  version: 2
  people_count: 0
  walk_distance: 0 cm (原始值: 0 m)
  walk_duration: 0 sec
  sit_duration: 0 sec (未开放使用)
  lie_duration: 0 sec
  stand_duration: 0 sec
  multi_person_duration: 0 sec
```

## 支持的 Track 格式

### Monitor Track
- 16 字节 * N（N 为人数）
- 每 16 字节代表 1 个人
- 包含：target_id, position_x/y/z, pose, event, remaining_time, area_id

### Stat Track
- 固定 16 字节
- 包含：version, people_count, walk_distance, walk_duration, lie_duration, stand_duration, multi_person_duration

## 注意事项

1. 数据库中的 `raw_original` 字段只包含基本审计信息，不包含 track 数据
2. 如需解码数据库中的数据，请使用 `--source redis` 从 Redis Streams 读取
3. Event 类型的数据不包含 track 字段
