# iot_timeseries 表结构优化方案

## 📋 问题

如果将所有统计数据字段都放在 `iot_timeseries` 表中，会导致表字段过多。

**当前字段数**：约 30+ 个字段

**需要新增的统计字段**：
- `stat_breath_state`、`stat_heart_state`、`stat_vital_signs_state`、`stat_sleep_state`（4个）
- `stat_timestamp`（1个）
- `track_version`、`track_person_count`、`track_walk_distance`、`track_walk_duration`、`track_sit_duration`、`track_lie_duration`、`track_stand_duration`、`track_multi_person_duration`（8个）
- `track_raw`（1个）

**总计新增**：约 14 个字段

**最终字段数**：约 44+ 个字段

## 🔍 方案对比

### 方案 A：所有字段都在 iot_timeseries 表（当前策略）

**表结构**：

```sql
CREATE TABLE iot_timeseries (
    -- 基础字段（~10个）
    id, tenant_id, device_id, timestamp, data_type, category, ...
    
    -- 轨迹数据（4个）
    tracking_id, radar_pos_x, radar_pos_y, radar_pos_z
    
    -- 姿态数据（2个）
    posture_snomed_code, posture_display
    
    -- 事件数据（4个）
    event_type, event_snomed_code, event_display, area_id
    
    -- 生命体征（6个）
    heart_rate_code, heart_rate_display, heart_rate,
    respiratory_rate_code, respiratory_rate_display, respiratory_rate
    
    -- 睡眠状态（2个）
    sleep_state_snomed_code, sleep_state_display
    
    -- 位置信息（2个）
    unit_id, room_id
    
    -- 统计数据 ⭐（14个新字段）
    stat_breath_state,
    stat_heart_state,
    stat_vital_signs_state,
    stat_sleep_state,
    stat_timestamp,
    track_version,
    track_person_count,
    track_walk_distance,
    track_walk_duration,
    track_sit_duration,
    track_lie_duration,
    track_stand_duration,
    track_multi_person_duration,
    track_raw
    
    -- 其他字段（~5个）
    alarm_event_id, confidence, remaining_time,
    raw_original, raw_format, raw_compression, metadata, created_at
);
```

**✅ 优点**：
- 查询简单：所有数据都在一个表中，无需 JOIN
- 性能好：直接查询字段，索引直接使用
- 符合现有策略：与其他字段（如 `posture_*`、`sleep_state_*`）一致

**❌ 缺点**：
- 字段过多：约 44+ 个字段，表结构复杂
- 维护困难：字段定义和管理复杂
- 查询性能可能下降：PostgreSQL 虽然支持很多字段，但过多的 NULL 值会影响性能
- 存储开销：每个字段都有存储开销（即使是 NULL）

---

### 方案 B：统计数据存储在 JSONB 字段 ✅ **推荐**

**表结构**：

```sql
CREATE TABLE iot_timeseries (
    -- 基础字段（保持不变）
    id, tenant_id, device_id, timestamp, data_type, category, ...
    
    -- 现有字段（保持不变）
    tracking_id, radar_pos_x, radar_pos_y, radar_pos_z,
    posture_snomed_code, posture_display,
    event_type, event_snomed_code, event_display, area_id,
    heart_rate_code, heart_rate_display, heart_rate,
    respiratory_rate_code, respiratory_rate_display, respiratory_rate,
    sleep_state_snomed_code, sleep_state_display,
    unit_id, room_id,
    alarm_event_id, confidence, remaining_time,
    raw_original, raw_format, raw_compression,
    
    -- ⭐ 新增：统计数据 JSONB 字段
    statistics JSONB,  -- 存储所有统计数据
    
    metadata JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

**statistics JSONB 结构**：

```json
{
  "breath_state": "NORMAL",      // 'NORMAL', 'LOW', 'HIGH', 'APNEA'
  "heart_state": "NORMAL",       // 'NORMAL', 'LOW', 'HIGH', 'UNDEFINED'
  "vital_signs_state": "NORMAL", // 'NORMAL', 'WEAK', 'UNDEFINED'
  "sleep_state": "AWAKE",        // 'UNDEFINED', 'LIGHT_SLEEP', 'DEEP_SLEEP', 'AWAKE'
  "stat_timestamp": "2024-01-01T12:00:00Z",
  
  "track": {
    "version": 1,
    "person_count": 1,
    "walk_distance": 10,        // 米
    "walk_duration": 30,        // 秒
    "sit_duration": 0,
    "lie_duration": 180,
    "stand_duration": 60,
    "multi_person_duration": 0  // 多人时长（访客标志）
  },
  
  "track_raw": "base64_encoded_16_bytes"  // 原始 track 数据（可选）
}
```

**✅ 优点**：
- 字段数量少：只增加 1 个字段（`statistics`）
- 灵活性高：可以轻松添加新的统计字段，无需修改表结构
- 存储效率：JSONB 压缩存储，NULL 值不占用空间
- 查询支持：PostgreSQL JSONB 支持索引和查询（GIN 索引）
- 符合现有模式：与 `metadata` JSONB 字段一致

**❌ 缺点**：
- 查询稍微复杂：需要使用 JSONB 操作符（如 `->`, `->>`）
- 索引需要特殊处理：需要创建 JSONB GIN 索引
- 类型检查：需要在应用层进行类型验证（JSONB 不强制类型）

**查询示例**：

```sql
-- 查询呼吸状态为 'LOW' 的数据
SELECT * FROM iot_timeseries
WHERE statistics->>'breath_state' = 'LOW';

-- 查询多人时长 > 0 的数据（访客）
SELECT * FROM iot_timeseries
WHERE (statistics->'track'->>'multi_person_duration')::int > 0;

-- 创建 GIN 索引（提高查询性能）
CREATE INDEX idx_iot_timeseries_statistics ON iot_timeseries USING GIN (statistics);
```

---

### 方案 C：单独的统计表（一对一关系）

**表结构**：

```sql
-- 主表（保持不变）
CREATE TABLE iot_timeseries (
    -- 现有字段（不变）
    id, tenant_id, device_id, timestamp, ...
);

-- ⭐ 新建：统计数据表
CREATE TABLE iot_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    iot_timeseries_id BIGINT NOT NULL REFERENCES iot_timeseries(id) ON DELETE CASCADE,
    
    -- 呼吸心率事件统计
    breath_state      VARCHAR(20),  -- 'NORMAL', 'LOW', 'HIGH', 'APNEA'
    heart_state       VARCHAR(20),  -- 'NORMAL', 'LOW', 'HIGH', 'UNDEFINED'
    vital_signs_state VARCHAR(20),  -- 'NORMAL', 'WEAK', 'UNDEFINED'
    sleep_state       VARCHAR(20),  -- 'UNDEFINED', 'LIGHT_SLEEP', 'DEEP_SLEEP', 'AWAKE'
    stat_timestamp    TIMESTAMPTZ,  -- 统计周期结束时间戳
    
    -- track 统计数据
    track_version          INTEGER,
    track_person_count     INTEGER,
    track_walk_distance    INTEGER,
    track_walk_duration    INTEGER,
    track_sit_duration     INTEGER,
    track_lie_duration     INTEGER,
    track_stand_duration   INTEGER,
    track_multi_person_duration INTEGER,
    track_raw              BYTEA,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    -- 唯一约束：一个 iot_timeseries 记录只有一条统计记录
    UNIQUE(iot_timeseries_id)
);

-- 索引
CREATE INDEX idx_iot_statistics_breath_state ON iot_statistics(breath_state);
CREATE INDEX idx_iot_statistics_heart_state ON iot_statistics(heart_state);
CREATE INDEX idx_iot_statistics_sleep_state ON iot_statistics(sleep_state);
CREATE INDEX idx_iot_statistics_multi_person ON iot_statistics(track_multi_person_duration);
```

**✅ 优点**：
- 表结构清晰：主表保持简洁，统计数据独立
- 字段类型明确：每个字段都有明确的类型约束
- 查询灵活：可以根据需要 JOIN 或单独查询
- 性能好：统计查询可以单独优化

**❌ 缺点**：
- 需要 JOIN：查询统计数据需要 JOIN，性能稍差
- 表数量增加：需要管理两个表
- 数据一致性：需要确保统计表与主表的数据一致性
- 不符合现有策略：与其他字段（如 `posture_*`）的存储方式不一致

---

## 🎯 推荐方案：方案 B（JSONB 存储）✅

### 理由

1. **字段数量少**：只增加 1 个字段，而不是 14 个
2. **灵活性高**：可以轻松添加新的统计字段，无需修改表结构
3. **符合现有模式**：与 `metadata` JSONB 字段一致
4. **查询性能可接受**：PostgreSQL JSONB 支持 GIN 索引，查询性能良好
5. **存储效率**：JSONB 压缩存储，NULL 值不占用空间

### 实施建议

#### 1. 表结构修改

```sql
-- 添加 statistics JSONB 字段
ALTER TABLE iot_timeseries
ADD COLUMN IF NOT EXISTS statistics JSONB;

-- 创建 GIN 索引（提高查询性能）
CREATE INDEX IF NOT EXISTS idx_iot_timeseries_statistics 
    ON iot_timeseries USING GIN (statistics);

-- 创建表达式索引（针对常用查询）
CREATE INDEX IF NOT EXISTS idx_iot_timeseries_stat_breath_state
    ON iot_timeseries((statistics->>'breath_state'))
    WHERE statistics IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_iot_timeseries_stat_sleep_state
    ON iot_timeseries((statistics->>'sleep_state'))
    WHERE statistics IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_iot_timeseries_stat_multi_person
    ON iot_timeseries(((statistics->'track'->>'multi_person_duration')::int))
    WHERE statistics IS NOT NULL AND (statistics->'track'->>'multi_person_duration') IS NOT NULL;
```

#### 2. 数据模型更新

```go
// internal/models/standardized_data.go
type StandardizedData struct {
    // ... 现有字段 ...
    
    // 统计数据（JSONB）
    Statistics *StatisticsData `json:"statistics,omitempty"`
}

type StatisticsData struct {
    // 呼吸心率事件统计
    BreathState      string     `json:"breath_state,omitempty"`      // 'NORMAL', 'LOW', 'HIGH', 'APNEA'
    HeartState       string     `json:"heart_state,omitempty"`       // 'NORMAL', 'LOW', 'HIGH', 'UNDEFINED'
    VitalSignsState  string     `json:"vital_signs_state,omitempty"` // 'NORMAL', 'WEAK', 'UNDEFINED'
    SleepState       string     `json:"sleep_state,omitempty"`       // 'UNDEFINED', 'LIGHT_SLEEP', 'DEEP_SLEEP', 'AWAKE'
    StatTimestamp    *time.Time `json:"stat_timestamp,omitempty"`    // 统计周期结束时间戳
    
    // track 统计数据
    Track *TrackStatistics `json:"track,omitempty"`
    
    // 原始 track 数据（base64 编码）
    TrackRaw string `json:"track_raw,omitempty"` // 可选，用于审计追溯
}

type TrackStatistics struct {
    Version            int `json:"version"`              // 版本标识符（1 或 2）
    PersonCount        int `json:"person_count"`         // 人数（0-8）
    WalkDistance       int `json:"walk_distance"`        // 行走距离（米）
    WalkDuration       int `json:"walk_duration"`        // 行走时长（秒）
    SitDuration        int `json:"sit_duration"`         // 静坐时长（秒，未开放使用）
    LieDuration        int `json:"lie_duration"`         // 躺卧时长（秒）
    StandDuration      int `json:"stand_duration"`       // 站立时长（秒）
    MultiPersonDuration int `json:"multi_person_duration"` // 多人时长（秒）⭐ 访客标志
}
```

#### 3. Repository 更新

```go
// internal/repository/iot_timeseries.go
func (r *IoTTimeSeriesRepository) Insert(data *models.StandardizedData) (int64, error) {
    // 序列化 statistics
    var statisticsJSON []byte
    if data.Statistics != nil {
        var err error
        statisticsJSON, err = json.Marshal(data.Statistics)
        if err != nil {
            return 0, fmt.Errorf("failed to marshal statistics: %w", err)
        }
    }
    
    query := `
        INSERT INTO iot_timeseries (
            -- ... 现有字段 ...
            statistics
        ) VALUES (
            -- ... 现有参数 ...
            $N,  -- statistics (JSONB)
            NULL  -- 如果 statistics 为空，插入 NULL
        )
        RETURNING id
    `
    // ... 执行插入 ...
}
```

#### 4. 查询示例

```sql
-- 查询呼吸状态为 'LOW' 的数据
SELECT * FROM iot_timeseries
WHERE statistics->>'breath_state' = 'LOW'
  AND timestamp >= NOW() - INTERVAL '1 hour';

-- 查询有访客的数据（多人时长 > 0）
SELECT 
    ts.*,
    (ts.statistics->'track'->>'multi_person_duration')::int as multi_person_duration
FROM iot_timeseries ts
WHERE (ts.statistics->'track'->>'multi_person_duration')::int > 0;

-- 查询睡眠状态为 'DEEP_SLEEP' 的统计数据
SELECT 
    ts.*,
    ts.statistics->>'sleep_state' as sleep_state
FROM iot_timeseries ts
WHERE ts.statistics->>'sleep_state' = 'DEEP_SLEEP';
```

## 📊 方案对比总结

| 方案 | 字段数 | 查询复杂度 | 灵活性 | 性能 | 推荐度 |
|------|--------|------------|--------|------|--------|
| **方案 A**：所有字段 | 44+ | 简单 | 低 | 好 | ⭐⭐⭐ |
| **方案 B**：JSONB | 31+ | 中等 | 高 | 良好 | ⭐⭐⭐⭐⭐ |
| **方案 C**：单独表 | 30+（主表） | 复杂（需要 JOIN） | 中等 | 良好 | ⭐⭐⭐⭐ |

## ✅ 最终推荐

**采用方案 B（JSONB 存储）**：

1. ✅ **字段数量少**：只增加 1 个字段
2. ✅ **灵活性高**：可以轻松扩展
3. ✅ **符合现有模式**：与 `metadata` JSONB 字段一致
4. ✅ **性能可接受**：JSONB GIN 索引性能良好
5. ✅ **存储效率**：压缩存储，NULL 值不占用空间

**实施步骤**：
1. 添加 `statistics JSONB` 字段
2. 创建 JSONB GIN 索引和表达式索引
3. 更新数据模型和 Repository
4. 更新 Transformer 处理统计数据

