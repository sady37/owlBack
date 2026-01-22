# 设备卡片创建问题分析

## 设备信息
- **设备类型**: Sleepad BM8701-2
- **序列号**: BM87224700978
- **Device UID**: 8amzqonkfyfyy
- **状态**: offline
- **Monitor**: enable
- **Business Access**: Approved
- **绑定位置**: DV-A-E102 (Unit E102)

## 卡片创建规则

### ActiveBed 卡片创建条件
设备必须满足以下**所有条件**：
1. ✅ 设备绑定到床：`devices.bound_bed_id IS NOT NULL`
2. ✅ 设备监护已启用：`devices.monitoring_enabled = TRUE`
3. ✅ 设备状态正常：`devices.status <> 'disabled'`
4. ✅ 设备类型：Radar 或 SleepPad（通过 device_store.device_type 判断）

### UnitCard 卡片创建条件
如果设备**只绑定到 unit 而未绑床**，需要满足以下条件：
1. ✅ 设备通过 room 关联到 unit：`devices.bound_room_id IS NOT NULL`
2. ✅ 设备未绑床：`devices.bound_bed_id IS NULL`
3. ✅ 设备监护已启用：`devices.monitoring_enabled = TRUE`
4. ✅ 场景 C：该 unit 下**没有 ActiveBed**（无设备绑床）
5. ✅ 场景 B：该 unit 下有**多个 ActiveBed**（≥2），且存在未绑床的设备

## 可能的问题原因

### 问题 1：设备未绑定到房间 (bound_room_id IS NULL)
**症状**：
- 设备只绑定到 unit（`devices.unit_id`），但没有通过 room 关联
- 查询 `devices` 表时，`bound_room_id` 为 NULL

**原因**：
- 卡片创建逻辑通过 `bound_room_id` 关联到 `rooms.unit_id` 来查找设备
- 如果 `bound_room_id IS NULL`，设备无法被 `GetUnboundDevicesByUnit` 查询到

**解决方案**：
1. 将设备绑定到 room：设置 `bound_room_id`（该 room 必须属于目标 unit E102）
2. 确保 room 的 `unit_id` 指向正确的 unit

**检查 SQL**：
```sql
SELECT 
    d.device_id,
    d.device_name,
    d.bound_room_id,
    d.unit_id,
    r.room_name,
    u.unit_name
FROM devices d
LEFT JOIN rooms r ON d.bound_room_id = r.room_id
LEFT JOIN units u ON r.unit_id = u.unit_id OR d.unit_id = u.unit_id
WHERE d.device_uid = '8amzqonkfyfyy';
```

### 问题 2：设备未绑定到床 (bound_bed_id IS NULL) 且 unit 下有 ActiveBed
**症状**：
- 设备只绑定到 unit，没有绑定到具体的 bed
- Unit E102 下有其他设备已绑床（ActiveBed）
- 根据规则，场景 B 下只有未绑床的设备数量 > 0 时才会创建 UnitCard

**解决方案**：
1. 将设备绑定到具体的 bed：设置 `bound_bed_id`（推荐）
2. 或者确保该 unit 下没有 ActiveBed（场景 C）

### 问题 3：设备状态为 'offline'
**症状**：
- 设备状态为 `offline`
- 虽然 `monitoring_enabled = TRUE`，但 `status = 'offline'` 可能影响卡片创建

**解决方案**：
- 检查 `status` 字段是否等于 'disabled'
- 如果 `status = 'offline'` 且不等于 'disabled'，应该不影响卡片创建
- 但建议将状态更新为 'active' 或 'online'

### 问题 4：设备未在 devices 表中创建
**症状**：
- 设备在 `device_store` 表中存在
- 但 `devices` 表中没有对应记录

**解决方案**：
- 设备需要从 `device_store` 出库到 `devices` 表
- 当设备分配给租户时，应该自动创建 `devices` 记录

## 诊断步骤

### 步骤 1：检查设备基本信息
运行诊断 SQL：`check_device_card_issue.sql`

或手动执行：
```sql
SELECT 
    d.device_id,
    d.device_name,
    d.device_uid,
    d.unit_id,
    d.bound_room_id,
    d.bound_bed_id,
    d.monitoring_enabled,
    d.business_access,
    d.status,
    ds.device_type,
    u.unit_name,
    r.room_name,
    b.bed_name
FROM devices d
JOIN device_store ds ON d.device_id = ds.device_id
LEFT JOIN units u ON d.unit_id = u.unit_id
LEFT JOIN rooms r ON d.bound_room_id = r.room_id
LEFT JOIN beds b ON d.bound_bed_id = b.bed_id
WHERE d.device_uid = '8amzqonkfyfyy';
```

### 步骤 2：检查 Unit E102 下的 ActiveBed
```sql
SELECT 
    b.bed_id,
    b.bed_name,
    COUNT(DISTINCT d.device_id) AS device_count
FROM beds b
INNER JOIN rooms r ON b.room_id = r.room_id
INNER JOIN units u ON r.unit_id = u.unit_id
INNER JOIN devices d ON d.bound_bed_id = b.bed_id
WHERE u.unit_name = 'E102'
  AND d.monitoring_enabled = TRUE
  AND d.status <> 'disabled'
GROUP BY b.bed_id, b.bed_name;
```

### 步骤 3：检查未绑床的设备
```sql
SELECT 
    d.device_id,
    d.device_name,
    d.bound_room_id,
    d.monitoring_enabled,
    d.status,
    r.room_name,
    u.unit_name
FROM devices d
LEFT JOIN rooms r ON d.bound_room_id = r.room_id
LEFT JOIN units u ON r.unit_id = u.unit_id
WHERE u.unit_name = 'E102'
  AND d.bound_bed_id IS NULL
  AND d.bound_room_id IS NOT NULL
  AND d.monitoring_enabled = TRUE;
```

## 修复建议

根据诊断结果，按以下优先级修复：

1. **如果 `bound_room_id IS NULL`**：
   - 将设备绑定到 room：设置 `bound_room_id`（该 room 必须属于 unit E102）
   - 这是**最可能的原因**

2. **如果 `bound_bed_id IS NULL` 且 unit 下有 ActiveBed**：
   - 将设备绑定到具体的 bed：设置 `bound_bed_id`
   - 或者确保该 unit 下没有 ActiveBed（场景 C）

3. **如果 `monitoring_enabled = FALSE`**：
   - 更新设备：设置 `monitoring_enabled = TRUE`

4. **如果 `status = 'disabled'`**：
   - 更新设备状态：设置 `status = 'active'` 或 'online'

5. **如果设备不在 `devices` 表中**：
   - 检查设备是否已从 `device_store` 出库
   - 确保设备已分配给正确的租户

## 手动触发卡片创建

修复后，可以手动触发卡片创建：

```bash
# 通过 API 触发卡片创建
curl -X POST http://localhost:8082/api/v1/cards/create \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "YOUR_TENANT_ID",
    "unit_id": "UNIT_ID_FOR_E102"
  }'
```

## 相关代码位置

- 卡片创建逻辑：`owl-common/card/creator.go`
- ActiveBed 查询：`wisefido-card-manage/internal/repository/card.go:GetActiveBedsByUnit`
- 设备查询：`wisefido-card-manage/internal/repository/card.go:GetUnboundDevicesByUnit`
- 设备更新触发：`wisefido-data/internal/service/device_service.go:UpdateDevice`
- 卡片创建规则：`owlRD/docs/20_Card_Creation_Rules_Final.md`
