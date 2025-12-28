# 表结构强规范检查报告：alarm_events

## 1. 数据库表结构（PostgreSQL 实际）

```sql
CREATE TABLE alarm_events (
    event_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    device_id         UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE,
    event_type        VARCHAR(50) NOT NULL,
    category          VARCHAR(50) CHECK (category IN ('safety', 'clinical', 'behavioral', 'device')),
    alarm_level       VARCHAR(20) NOT NULL CHECK (...),
    alarm_status      VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (...),
    triggered_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    hand_time         TIMESTAMPTZ,
    iot_timeseries_id BIGINT,
    trigger_data      JSONB,
    handler           UUID REFERENCES users(user_id) ON DELETE SET NULL,
    operation         VARCHAR(30) CHECK (...),
    notes             TEXT,
    notified_users    JSONB DEFAULT '[]'::JSONB,
    metadata          JSONB DEFAULT '{}'::JSONB,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 字段列表：
- `event_id` (UUID, PK, NOT NULL, DEFAULT gen_random_uuid())
- `tenant_id` (UUID, NOT NULL, FK → tenants.tenant_id)
- `device_id` (UUID, NOT NULL, FK → devices.device_id)
- `event_type` (VARCHAR(50), NOT NULL)
- `category` (VARCHAR(50), nullable, CHECK IN ('safety', 'clinical', 'behavioral', 'device'))
- `alarm_level` (VARCHAR(20), NOT NULL, CHECK IN ('0'-'7', 'EMERG', 'ALERT', 'CRIT', 'ERR', 'WARNING', 'NOTICE', 'INFO', 'DEBUG'))
- `alarm_status` (VARCHAR(20), NOT NULL, DEFAULT 'active', CHECK IN ('active', 'acknowledged'))
- `triggered_at` (TIMESTAMPTZ, NOT NULL, DEFAULT CURRENT_TIMESTAMP)
- `hand_time` (TIMESTAMPTZ, nullable)
- `iot_timeseries_id` (BIGINT, nullable)
- `trigger_data` (JSONB, nullable)
- `handler` (UUID, nullable, FK → users.user_id)
- `operation` (VARCHAR(30), nullable, CHECK IN ('verified_and_processed', 'false_alarm', 'test', 'auto_relieved'))
- `notes` (TEXT, nullable)
- `notified_users` (JSONB, nullable, DEFAULT '[]'::JSONB)
- `metadata` (JSONB, nullable, DEFAULT '{}'::JSONB)
- `created_at` (TIMESTAMPTZ, NOT NULL, DEFAULT CURRENT_TIMESTAMP)
- `updated_at` (TIMESTAMPTZ, NOT NULL, DEFAULT CURRENT_TIMESTAMP)

---

## 2. Domain 模型检查

**状态：✅ 字段匹配正确**

当前文件：`internal/domain/alarm_event.go`

### Domain 模型：
```go
type AlarmEvent struct {
    EventID          string          `db:"event_id"`          // ✅ UUID, PRIMARY KEY
    TenantID         string          `db:"tenant_id"`          // ✅ UUID, NOT NULL
    DeviceID         string          `db:"device_id"`         // ✅ UUID, NOT NULL
    EventType        string          `db:"event_type"`        // ✅ VARCHAR(50), NOT NULL
    Category         string          `db:"category"`          // ✅ VARCHAR(50), nullable
    AlarmLevel       string          `db:"alarm_level"`       // ✅ VARCHAR(20), NOT NULL
    AlarmStatus      string          `db:"alarm_status"`      // ✅ VARCHAR(20), NOT NULL
    TriggeredAt      time.Time       `db:"triggered_at"`      // ✅ TIMESTAMPTZ, NOT NULL
    HandTime         *time.Time      `db:"hand_time"`        // ✅ TIMESTAMPTZ, nullable
    IoTTimeSeriesID  *int64          `db:"iot_timeseries_id"` // ✅ BIGINT, nullable
    TriggerData      json.RawMessage `db:"trigger_data"`      // ✅ JSONB, nullable
    Handler          *string         `db:"handler"`           // ✅ UUID, nullable
    Operation        *string         `db:"operation"`         // ✅ VARCHAR(30), nullable
    Notes            *string         `db:"notes"`             // ✅ TEXT, nullable
    NotifiedUsers    json.RawMessage `db:"notified_users"`    // ✅ JSONB, nullable
    Metadata         json.RawMessage `db:"metadata"`           // ✅ JSONB, nullable
    CreatedAt        time.Time       `db:"created_at"`        // ✅ TIMESTAMPTZ, NOT NULL
    UpdatedAt        time.Time       `db:"updated_at"`        // ✅ TIMESTAMPTZ, NOT NULL
}
```

**所有字段的 db tag 与数据库表结构一致。**

---

## 3. Repository 接口检查

**状态：✅ 已存在**

文件：`internal/repository/alarm_events_repo.go`

接口方法：
- `ListAlarmEvents` - 列表查询（支持多条件过滤、分页）
- `GetAlarmEvent` - 获取单个报警事件
- `CreateAlarmEvent` - 创建报警事件
- `AcknowledgeAlarmEvent` - 确认报警
- `UpdateAlarmEventOperation` - 更新操作结果
- `UpdateAlarmEvent` - 更新报警事件（部分更新）
- `DeleteAlarmEvent` - 软删除报警事件
- `GetRecentAlarmEvent` - 获取最近的报警事件（用于去重检查）
- `CountAlarmEvents` - 统计报警事件数量

---

## 4. Repository 实现检查

**状态：⚠️ 发现一个问题需要修复**

文件：`internal/repository/postgres_alarm_events.go`

### 问题 1：BranchTag 过滤使用了错误的字段

**位置**：第 184 行和第 895 行

**问题**：
- 代码使用了 `u.branch_name`，但 `units` 表只有 `branch_id` 字段
- 需要通过 JOIN `branches` 表来获取 `branch_name`

**当前代码**（错误）：
```go
// 分支标签过滤
if filters.BranchTag != nil {
    where = append(where, fmt.Sprintf("u.branch_name = $%d", argN))
    args = append(args, *filters.BranchTag)
    argN++
}
```

**应该修复为**：
```go
// 分支标签过滤（通过 JOIN branches 表获取 branch_name）
if filters.BranchTag != nil {
    // 确保 JOIN branches 表
    needBranchesJoin := true
    for _, join := range joins {
        if strings.Contains(join, "branches") {
            needBranchesJoin = false
            break
        }
    }
    if needBranchesJoin {
        joins = append(joins, "LEFT JOIN branches br ON u.branch_id = br.branch_id")
    }
    where = append(where, fmt.Sprintf("br.branch_name = $%d", argN))
    args = append(args, *filters.BranchTag)
    argN++
}
```

### 其他检查结果：

**ListAlarmEvents** (第 230-255 行)：
- ✅ 查询字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段和 JSONB 字段

**GetAlarmEvent** (第 342-380 行)：
- ✅ 查询字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段和 JSONB 字段

**CreateAlarmEvent** (第 486-508 行)：
- ✅ INSERT 字段：所有字段与数据库表结构一致
- ✅ 正确处理 JSONB 字段

**UpdateAlarmEvent** (第 646-649 行)：
- ✅ UPDATE 字段：动态构建，字段名与数据库表结构一致

**DeleteAlarmEvent** (第 680-718 行)：
- ✅ 软删除逻辑：使用 metadata 字段标记删除时间

---

## 5. 问题总结

### ⚠️ 需要修复：
1. **BranchTag 过滤**：`postgres_alarm_events.go` 第 184 行和第 895 行使用了 `u.branch_name`，应该通过 JOIN `branches` 表获取 `branch_name`。

### ✅ 其他内容正确：
1. Domain 模型：所有字段的 db tag 与数据库表结构一致
2. Repository 接口：已定义，方法完整
3. Repository 实现：除 BranchTag 过滤外，所有 SQL 查询字段与数据库表结构一致

---

## 6. 修复方案

需要修复 `postgres_alarm_events.go` 中的两处 `BranchTag` 过滤逻辑：
1. 第 182-187 行（`ListAlarmEvents` 方法中）
2. 第 893-898 行（`CountAlarmEvents` 方法中）

修复方法：在 JOIN `units` 表后，再 JOIN `branches` 表，然后使用 `br.branch_name` 进行过滤。

---

## 7. 相关代码位置

- Domain 模型：`internal/domain/alarm_event.go`
- Repository 接口：`internal/repository/alarm_events_repo.go`
- Repository 实现：`internal/repository/postgres_alarm_events.go`

