# 表结构强规范检查报告：iot_timeseries

## 1. 数据库表结构（PostgreSQL 实际）

```sql
CREATE TABLE iot_timeseries (
    id                       BIGSERIAL PRIMARY KEY,
    tenant_id                UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    device_id                UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE,
    timestamp                TIMESTAMPTZ NOT NULL,
    data_type                VARCHAR(20) NOT NULL DEFAULT 'observation' CHECK (...),
    category                 VARCHAR(50) CHECK (...),
    tracking_id              INTEGER,
    radar_pos_x              INTEGER,
    radar_pos_y              INTEGER,
    radar_pos_z              INTEGER,
    posture_snomed_code      VARCHAR(50),
    posture_display          VARCHAR(100),
    event_type               VARCHAR(50),
    event_snomed_code        VARCHAR(50),
    event_display            VARCHAR(100),
    area_id                  INTEGER,
    heart_rate_code          VARCHAR(50),
    heart_rate_display       VARCHAR(100),
    heart_rate               INTEGER,
    respiratory_rate_code    VARCHAR(50),
    respiratory_rate_display VARCHAR(100),
    respiratory_rate         INTEGER,
    sleep_state_snomed_code  VARCHAR(50),
    sleep_state_display      VARCHAR(100),
    unit_id                  UUID REFERENCES units(unit_id) ON DELETE SET NULL,
    room_id                  UUID REFERENCES rooms(room_id) ON DELETE SET NULL,
    alarm_event_id           UUID,
    confidence                INTEGER,
    remaining_time            INTEGER,
    raw_original              BYTEA NOT NULL,
    raw_format                VARCHAR(50) NOT NULL,
    raw_compression           VARCHAR(50),
    metadata                  JSONB DEFAULT '{}'::JSONB,
    created_at                TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

### 字段列表（共 34 个字段）：
- `id` (BIGSERIAL, PK, NOT NULL)
- `tenant_id` (UUID, NOT NULL, FK → tenants.tenant_id)
- `device_id` (UUID, NOT NULL, FK → devices.device_id)
- `timestamp` (TIMESTAMPTZ, NOT NULL)
- `data_type` (VARCHAR(20), NOT NULL, DEFAULT 'observation', CHECK IN ('observation', 'alarm'))
- `category` (VARCHAR(50), nullable, CHECK IN ('vital-signs', 'activity', 'safety', 'clinical', 'behavioral', 'device', 'social-history'))
- `tracking_id` (INTEGER, nullable)
- `radar_pos_x` (INTEGER, nullable)
- `radar_pos_y` (INTEGER, nullable)
- `radar_pos_z` (INTEGER, nullable)
- `posture_snomed_code` (VARCHAR(50), nullable)
- `posture_display` (VARCHAR(100), nullable)
- `event_type` (VARCHAR(50), nullable)
- `event_snomed_code` (VARCHAR(50), nullable)
- `event_display` (VARCHAR(100), nullable)
- `area_id` (INTEGER, nullable)
- `heart_rate_code` (VARCHAR(50), nullable)
- `heart_rate_display` (VARCHAR(100), nullable)
- `heart_rate` (INTEGER, nullable)
- `respiratory_rate_code` (VARCHAR(50), nullable)
- `respiratory_rate_display` (VARCHAR(100), nullable)
- `respiratory_rate` (INTEGER, nullable)
- `sleep_state_snomed_code` (VARCHAR(50), nullable)
- `sleep_state_display` (VARCHAR(100), nullable)
- `unit_id` (UUID, nullable, FK → units.unit_id)
- `room_id` (UUID, nullable, FK → rooms.room_id)
- `alarm_event_id` (UUID, nullable, FK → alarm_events.event_id)
- `confidence` (INTEGER, nullable)
- `remaining_time` (INTEGER, nullable)
- `raw_original` (BYTEA, NOT NULL)
- `raw_format` (VARCHAR(50), NOT NULL)
- `raw_compression` (VARCHAR(50), nullable)
- `metadata` (JSONB, nullable, DEFAULT '{}'::JSONB)
- `created_at` (TIMESTAMPTZ, nullable, DEFAULT CURRENT_TIMESTAMP)

---

## 2. Domain 模型检查

**状态：⚠️ 发现多个问题**

当前文件：`internal/domain/iot_timeseries.go`

### Domain 模型：
```go
type IoTTimeSeries struct {
    ID                      int64  `db:"id"`                      // ✅ BIGSERIAL, PRIMARY KEY
    TenantID                string `db:"tenant_id"`               // ✅ UUID, NOT NULL
    DeviceID                string `db:"device_id"`               // ✅ UUID, NOT NULL
    Timestamp               time.Time `db:"timestamp"`            // ✅ TIMESTAMPTZ, NOT NULL
    DataType                string `db:"data_type"`              // ✅ VARCHAR(20), NOT NULL
    Category                string `db:"category"`                // ⚠️ VARCHAR(50), nullable - 应该是 *string 或 sql.NullString
    TrackingID              *int   `db:"tracking_id"`            // ✅ INTEGER, nullable
    RadarPosX               *int   `db:"radar_pos_x"`            // ✅ INTEGER, nullable
    RadarPosY               *int   `db:"radar_pos_y"`             // ✅ INTEGER, nullable
    RadarPosZ               *int   `db:"radar_pos_z"`             // ✅ INTEGER, nullable
    PostureSNOMEDCode       string `db:"posture_snomed_code"`    // ⚠️ VARCHAR(50), nullable - 应该是 *string 或 sql.NullString
    PostureDisplay          string `db:"posture_display"`         // ⚠️ VARCHAR(100), nullable - 应该是 *string 或 sql.NullString
    EventType               string `db:"event_type"`              // ⚠️ VARCHAR(50), nullable - 应该是 *string 或 sql.NullString
    EventSNOMEDCode         string `db:"event_snomed_code"`       // ⚠️ VARCHAR(50), nullable - 应该是 *string 或 sql.NullString
    EventDisplay            string `db:"event_display"`           // ⚠️ VARCHAR(100), nullable - 应该是 *string 或 sql.NullString
    AreaID                  *int   `db:"area_id"`                // ✅ INTEGER, nullable
    HeartRateCode           string `db:"heart_rate_code"`         // ⚠️ VARCHAR(50), nullable - 应该是 *string 或 sql.NullString
    HeartRateDisplay        string `db:"heart_rate_display"`      // ⚠️ VARCHAR(100), nullable - 应该是 *string 或 sql.NullString
    HeartRate               *int   `db:"heart_rate"`             // ✅ INTEGER, nullable
    RespiratoryRateCode     string `db:"respiratory_rate_code"`  // ⚠️ VARCHAR(50), nullable - 应该是 *string 或 sql.NullString
    RespiratoryRateDisplay  string `db:"respiratory_rate_display"` // ⚠️ VARCHAR(100), nullable - 应该是 *string 或 sql.NullString
    RespiratoryRate         *int   `db:"respiratory_rate"`        // ✅ INTEGER, nullable
    SleepStateSNOMEDCode    string `db:"sleep_state_snomed_code"` // ⚠️ VARCHAR(50), nullable - 应该是 *string 或 sql.NullString
    SleepStateDisplay       string `db:"sleep_state_display"`     // ⚠️ VARCHAR(100), nullable - 应该是 *string 或 sql.NullString
    UnitID                  string `db:"unit_id"`                 // ⚠️ UUID, nullable - 应该是 *string 或 sql.NullString
    RoomID                  string `db:"room_id"`                 // ⚠️ UUID, nullable - 应该是 *string 或 sql.NullString
    AlarmEventID            string `db:"alarm_event_id"`          // ⚠️ UUID, nullable - 应该是 *string 或 sql.NullString
    Confidence              *int   `db:"confidence"`              // ✅ INTEGER, nullable
    RemainingTime           *int   `db:"remaining_time"`          // ✅ INTEGER, nullable
    RawOriginal             []byte `db:"raw_original"`           // ✅ BYTEA, NOT NULL
    RawFormat               string `db:"raw_format"`              // ✅ VARCHAR(50), NOT NULL
    RawCompression          string `db:"raw_compression"`         // ⚠️ VARCHAR(50), nullable - 应该是 *string 或 sql.NullString
    Metadata                map[string]interface{} `db:"metadata"` // ⚠️ JSONB, nullable - 应该是 json.RawMessage 或 *json.RawMessage
    CreatedAt               time.Time `db:"created_at"`           // ⚠️ TIMESTAMPTZ, nullable - 应该是 *time.Time 或 sql.NullTime
    // 关联数据（通过 JOIN 获取）
    DeviceSN                string `db:"device_sn"`               // ✅ 从 devices 表获取
    DeviceUID               string `db:"device_uid"`               // ✅ 从 devices 表获取
    DeviceType               string `db:"device_type"`            // ✅ 从 device_store 表获取
    FirmwareVersion          string `db:"firmware_version"`       // ✅ 从 devices 表获取
}
```

**问题**：
- 多个可空字段使用了 `string` 类型，应该使用 `*string` 或 `sql.NullString` 来表示可空字段
- `Metadata` 字段类型是 `map[string]interface{}`，但数据库字段是 JSONB，应该使用 `json.RawMessage`
- `CreatedAt` 字段类型是 `time.Time`，但数据库字段是 nullable，应该使用 `*time.Time` 或 `sql.NullTime`

---

## 3. Repository 接口检查

**状态：✅ 已存在**

文件：`internal/repository/iot_timeseries_repo.go`

接口方法：
- `GetTimeSeriesData` - 获取时序数据（按ID）
- `GetLatestData` - 获取最新数据（按设备）
- `GetDataByDevice` - 按设备查询（支持过滤）
- `GetDataByResident` - 按住户查询（通过device关联）
- `GetDataByTimeRange` - 时间范围查询
- `GetDataByLocation` - 按位置查询（unit_id/room_id）

**注意**：此 Repository 只提供查询方法，数据写入由 `wisefido-data-transformer` 服务负责。

---

## 4. Repository 实现检查

**状态：✅ 字段匹配正确（但需要修复 Domain 模型中的字段类型）**

文件：`internal/repository/postgres_iot_timeseries.go`

### 检查结果：

**buildBaseQuery** (第 28-79 行)：
- ✅ 查询字段：所有字段与数据库表结构一致
- ✅ 正确处理 JOIN 查询（devices, device_store, alarm_events）

**scanIoTTimeSeries** (第 138-290 行)：
- ✅ 使用 `sql.NullString`, `sql.NullInt64`, `sql.NullTime` 正确处理可空字段
- ✅ 正确转换类型（INTEGER → *int, UUID → string）
- ⚠️ `metadata` 字段使用 `[]byte` 扫描，但 domain 模型使用 `map[string]interface{}`，需要解析 JSONB

---

## 5. 问题总结

### ⚠️ 需要修复：
1. **多个可空字段类型**：`domain.IoTTimeSeries` 中多个可空字段使用了 `string` 类型，应该使用 `*string` 或 `sql.NullString`：
   - `Category`, `PostureSNOMEDCode`, `PostureDisplay`, `EventType`, `EventSNOMEDCode`, `EventDisplay`
   - `HeartRateCode`, `HeartRateDisplay`, `RespiratoryRateCode`, `RespiratoryRateDisplay`
   - `SleepStateSNOMEDCode`, `SleepStateDisplay`
   - `UnitID`, `RoomID`, `AlarmEventID`
   - `RawCompression`

2. **Metadata 字段类型**：`domain.IoTTimeSeries.Metadata` 应该是 `json.RawMessage`，而不是 `map[string]interface{}`，以正确处理 JSONB 数据。

3. **CreatedAt 字段类型**：`domain.IoTTimeSeries.CreatedAt` 应该是 `*time.Time` 或 `sql.NullTime`，而不是 `time.Time`，因为数据库字段是 nullable。

### ✅ 其他内容正确：
1. Repository 接口：已定义，方法完整
2. Repository 实现：所有 SQL 查询字段与数据库表结构一致，正确处理可空字段

---

## 6. 修复方案

需要修复 `domain/iot_timeseries.go` 中的多个字段类型：

1. 将所有可空的 `string` 字段改为 `*string` 或 `sql.NullString`
2. 将 `Metadata` 字段改为 `json.RawMessage`
3. 将 `CreatedAt` 字段改为 `*time.Time` 或 `sql.NullTime`

然后需要更新 `postgres_iot_timeseries.go` 中的相关代码，确保正确处理这些字段类型。

---

## 7. 相关代码位置

- Domain 模型：`internal/domain/iot_timeseries.go`
- Repository 接口：`internal/repository/iot_timeseries_repo.go`
- Repository 实现：`internal/repository/postgres_iot_timeseries.go`

