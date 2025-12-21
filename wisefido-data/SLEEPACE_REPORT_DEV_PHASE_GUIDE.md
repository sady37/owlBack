# SleepaceReportService 开发阶段指南

## 📋 当前状态

### ✅ 已完成功能

1. **查询功能**（完全可用）
   - ✅ 获取报告列表 (`GET /sleepace/api/v1/sleepace/reports/:id`)
   - ✅ 获取报告详情 (`GET /sleepace/api/v1/sleepace/reports/:id/detail?date=YYYYMMDD`)
   - ✅ 获取有效日期列表 (`GET /sleepace/api/v1/sleepace/reports/:id/dates`)

2. **数据库层**
   - ✅ PostgreSQL 表结构
   - ✅ Repository 层（支持 `device_code` 匹配）

3. **业务层**
   - ✅ Service 层（查询功能）
   - ✅ Handler 层
   - ✅ 路由注册

### ⏸️ 暂停功能

- **数据下载功能**：待设备接入后再实现
  - 原因：开发阶段无设备，厂家服务无数据
  - 状态：功能设计已完成，代码待实现

---

## 🧪 开发阶段测试方案

### 方案 1：使用测试数据（推荐）

**步骤**：

1. **准备测试设备**
   ```sql
   -- 确保 devices 表中有测试设备
   -- device_code 可以是 serial_number 或 uid
   SELECT device_id, serial_number, uid, device_name 
   FROM devices 
   WHERE tenant_id = 'your-tenant-id'::uuid
     AND status <> 'disabled'
   LIMIT 5;
   ```

2. **加载测试数据**
   ```bash
   # 执行测试数据脚本
   psql -h localhost -U postgres -d owlrd -f db/test_data_sleepace_report.sql
   ```

3. **验证数据**
   ```sql
   -- 查看插入的报告
   SELECT * FROM sleepace_report ORDER BY date DESC LIMIT 10;
   ```

4. **测试 API**
   ```bash
   # 获取报告列表
   curl -X GET "http://localhost:8080/sleepace/api/v1/sleepace/reports/{device_id}?startDate=20240819&endDate=20240820" \
     -H "Authorization: Bearer {token}"

   # 获取报告详情
   curl -X GET "http://localhost:8080/sleepace/api/v1/sleepace/reports/{device_id}/detail?date=20240820" \
     -H "Authorization: Bearer {token}"

   # 获取有效日期列表
   curl -X GET "http://localhost:8080/sleepace/api/v1/sleepace/reports/{device_id}/dates" \
     -H "Authorization: Bearer {token}"
   ```

### 方案 2：手动插入测试数据

**步骤**：

1. **插入单条测试报告**
   ```sql
   INSERT INTO sleepace_report (
       tenant_id,
       device_id,
       device_code,
       record_count,
       start_time,
       end_time,
       date,
       stop_mode,
       time_step,
       timezone,
       sleep_state,
       report,
       created_at,
       updated_at
   ) VALUES (
       'your-tenant-id'::uuid,
       'your-device-id'::uuid,
       'SP001',  -- 对应 devices.serial_number 或 devices.uid
       1440,
       EXTRACT(EPOCH FROM '2024-08-20 00:00:00'::timestamptz)::bigint,
       EXTRACT(EPOCH FROM '2024-08-21 00:00:00'::timestamptz)::bigint,
       20240820,
       0,
       60,
       28800,
       '[1,1,1,2,2,2,3,3,3,2,2,1,1,1]',
       '[{"summary":{"recordCount":1440,"startTime":1721491200,"stopMode":0,"timeStep":60,"timezone":28800},"analysis":{"sleepStateStr":[1,1,1,2,2,2,3,3,3,2,2,1,1,1]}}]',
       NOW(),
       NOW()
   );
   ```

---

## 📝 测试数据说明

### 字段说明

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `device_code` | 设备编码（对应 `devices.serial_number` 或 `devices.uid`） | `'SP001'` |
| `date` | 日期（YYYYMMDD 格式） | `20240820` |
| `sleep_state` | 睡眠状态数组（JSON 字符串） | `'[1,1,1,2,2,2,3,3,3]'` |
| `report` | 完整报告数据（JSON 字符串） | `'[{...}]'` |

### 睡眠状态值

- `1` = 清醒（Awake）
- `2` = 浅睡眠（Light sleep）
- `3` = 深睡眠（Deep sleep）

### 报告数据格式

`report` 字段存储完整的 JSON 字符串，格式如下：
```json
[{
  "summary": {
    "recordCount": 1440,
    "startTime": 1721491200,
    "stopMode": 0,
    "timeStep": 60,
    "timezone": 28800
  },
  "analysis": {
    "sleepStateStr": [1,1,1,2,2,2,3,3,3,2,2,1,1,1]
  }
}]
```

---

## 🔍 常见问题

### Q1: 如何获取 device_id？

**A**: 通过 `device_code`（serial_number 或 uid）查询：
```sql
SELECT device_id 
FROM devices 
WHERE tenant_id = 'your-tenant-id'::uuid
  AND (serial_number = 'SP001' OR uid = 'SP001')
  AND status <> 'disabled'
LIMIT 1;
```

### Q2: 测试数据插入失败？

**A**: 检查以下几点：
1. `tenant_id` 是否存在
2. `device_id` 是否存在且 `status <> 'disabled'`
3. `device_code` 是否与 `devices.serial_number` 或 `devices.uid` 匹配
4. 唯一性约束：`(tenant_id, device_id, date)` 不能重复

### Q3: 如何清理测试数据？

**A**: 
```sql
-- 删除所有测试数据
DELETE FROM sleepace_report 
WHERE tenant_id = 'your-tenant-id'::uuid;

-- 或删除特定设备的报告
DELETE FROM sleepace_report 
WHERE tenant_id = 'your-tenant-id'::uuid
  AND device_id = 'your-device-id'::uuid;
```

---

## 🚀 设备接入后的工作

当设备接入后，需要实现：

1. **数据下载功能**（阶段 1）
   - 实现 `DownloadReport` Service 方法
   - 实现 Sleepace 厂家 API 客户端
   - 实现 `DownloadReport` Handler

2. **后台任务**（可选）
   - 定时任务自动下载报告
   - MQTT 触发下载

详细实现计划见：`SLEEPACE_REPORT_NEXT_STEPS.md`

---

## 📚 相关文档

- `SLEEPACE_REPORT_SERVICE_IMPLEMENTATION.md` - 实现总结
- `SLEEPACE_REPORT_NEXT_STEPS.md` - 后续规划
- `SLEEPACE_REPORT_DEVICE_CODE_CLARIFICATION.md` - device_code 说明
- `db/test_data_sleepace_report.sql` - 测试数据脚本

