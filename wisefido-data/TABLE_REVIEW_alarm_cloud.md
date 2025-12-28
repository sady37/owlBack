# 表结构强规范检查报告：alarm_cloud

## 1. 数据库表结构（PostgreSQL 实际）

```sql
CREATE TABLE alarm_cloud (
    tenant_id          UUID PRIMARY KEY,
    offlinealarm       VARCHAR(20) DEFAULT 'ERROR',
    lowbattery         VARCHAR(20) DEFAULT 'WARNING',
    devicefailure      VARCHAR(20) DEFAULT 'ERROR',
    device_alarms      JSONB NOT NULL DEFAULT '{}'::jsonb,
    conditions         JSONB,
    notification_rules JSONB,
    metadata           JSONB
);
```

### 字段列表：
- `tenant_id` (UUID, PK, NOT NULL, FK → tenants.tenant_id)
- `offlinealarm` (VARCHAR(20), nullable, DEFAULT 'ERROR')
- `lowbattery` (VARCHAR(20), nullable, DEFAULT 'WARNING')
- `devicefailure` (VARCHAR(20), nullable, DEFAULT 'ERROR')
- `device_alarms` (JSONB, NOT NULL, DEFAULT '{}'::jsonb)
- `conditions` (JSONB, nullable)
- `notification_rules` (JSONB, nullable)
- `metadata` (JSONB, nullable)

### 约束：
- 主键：`tenant_id`（一个租户只有一条记录）
- 外键：`tenant_id` → `tenants.tenant_id`

### 业务规则：
1. 一个租户只有一条记录，包含所有设备类型的报警配置
2. `tenant_id = SystemTenantID` 表示系统默认模板（用于新租户初始化）
3. 通用报警（OfflineAlarm, LowBattery, DeviceFailure）：所有设备类型都支持
4. 设备特定报警：存储在 `device_alarms` JSONB 中

---

## 2. Domain 模型检查

**状态：⚠️ 存在字段名不匹配问题**

当前文件：`internal/domain/alarm_cloud.go`

### 问题：
- Domain 模型的 db tag 使用了 `offline_alarm`、`low_battery`、`device_failure`（带下划线）
- 但数据库表实际字段名是 `offlinealarm`、`lowbattery`、`devicefailure`（无下划线）

### 当前 Domain 模型：
```go
OfflineAlarm  string `db:"offline_alarm"`  // ❌ 应该是 "offlinealarm"
LowBattery    string `db:"low_battery"`    // ❌ 应该是 "lowbattery"
DeviceFailure string `db:"device_failure"` // ❌ 应该是 "devicefailure"
```

### 需要修复：
```go
OfflineAlarm  string `db:"offlinealarm"`  // ✅
LowBattery    string `db:"lowbattery"`    // ✅
DeviceFailure string `db:"devicefailure"` // ✅
```

**注意**：PostgreSQL 会将未加引号的标识符转换为小写，所以 SQL 查询中使用 `OfflineAlarm` 会被转换为 `offlinealarm`，这应该是可以工作的。但为了保持一致性，建议修复 db tag。

---

## 3. Repository 接口检查

**状态：✅ 已存在**

文件：`internal/repository/alarm_cloud_repo.go`

接口方法：
- `GetAlarmCloud` - 获取租户的告警策略配置
- `UpsertAlarmCloud` - 创建或更新租户的告警策略配置
- `GetSystemAlarmCloud` - 获取系统默认告警策略模板
- `DeleteAlarmCloud` - 删除租户的告警策略配置

---

## 4. Repository 实现检查

**状态：⚠️ 存在字段名不匹配问题**

文件：`internal/repository/postgres_alarm_cloud.go`

### 问题：
1. **GetAlarmCloud** (第 36-38 行)：
   - SQL 查询使用了 `OfflineAlarm`、`LowBattery`、`DeviceFailure`（大写）
   - PostgreSQL 会将未加引号的标识符转换为小写，所以这应该是可以工作的
   - 但为了保持一致性，建议使用小写字段名

2. **UpsertAlarmCloud** (第 102-104, 111-113 行)：
   - INSERT/UPDATE 语句使用了 `OfflineAlarm`、`LowBattery`、`DeviceFailure`（大写）
   - 同样，PostgreSQL 会转换为小写，但建议使用小写字段名

### 建议修复：
- 将所有 SQL 查询中的字段名改为小写：`offlinealarm`、`lowbattery`、`devicefailure`

---

## 5. 问题总结

### ⚠️ 需要修复的内容：
1. **Domain 模型** (`domain/alarm_cloud.go`)：
   - `OfflineAlarm` 的 db tag 从 `offline_alarm` 改为 `offlinealarm`
   - `LowBattery` 的 db tag 从 `low_battery` 改为 `lowbattery`
   - `DeviceFailure` 的 db tag 从 `device_failure` 改为 `devicefailure`

2. **Repository 实现** (`postgres_alarm_cloud.go`)：
   - SQL 查询中的字段名改为小写（可选，但建议保持一致）

### ✅ 已存在的内容：
- Repository 接口已定义
- Repository 实现已存在（但需要修复字段名）

---

## 6. 修复方案

### 方案 1：修复 db tag（推荐）
- Domain 模型：将 db tag 改为小写字段名
- Repository：SQL 查询使用小写字段名

### 方案 2：保持现状
- PostgreSQL 会自动将未加引号的标识符转换为小写，所以当前代码应该可以工作
- 但为了保持代码清晰和一致性，建议修复

推荐使用方案 1，保持代码与数据库表结构完全一致。

