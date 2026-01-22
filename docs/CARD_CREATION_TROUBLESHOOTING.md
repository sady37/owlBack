# 卡片创建问题排查指南

## 问题：设备绑定到 Unit 后没有创建卡片

### 场景描述
- 设备：Sleepad-yfyy
- 绑定到：DV-A Unit E102
- 问题：没有创建卡片

## 卡片创建规则

### ActiveBed 卡片创建条件
设备必须满足以下**所有条件**才会被识别为 ActiveBed，从而创建 ActiveBed 卡片：

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

## 常见问题排查

### 问题 1：设备没有绑定到床（bound_bed_id IS NULL）

**症状**：
- 设备只绑定到 unit，没有绑定到具体的 bed
- 查询 `devices` 表时，`bound_bed_id` 为 NULL

**解决方案**：
1. 如果该 unit 下没有 ActiveBed（场景 C）：
   - 确保 `monitoring_enabled = TRUE`
   - 确保 `bound_room_id IS NOT NULL`（通过 room 关联到 unit）
   - 应该会创建 UnitCard

2. 如果该 unit 下有 ActiveBed：
   - 需要将设备绑定到具体的 bed（设置 `bound_bed_id`）
   - 或者确保该 unit 下没有 ActiveBed，才会创建 UnitCard

**检查 SQL**：
```sql
SELECT 
    d.device_id,
    d.device_name,
    d.bound_bed_id,
    d.bound_room_id,
    d.monitoring_enabled,
    d.status
FROM devices d
WHERE d.device_name LIKE '%Sleepad-yfyy%';
```

### 问题 2：设备监护未启用（monitoring_enabled = FALSE）

**症状**：
- 设备已绑定到 unit/bed，但 `monitoring_enabled = FALSE`

**解决方案**：
- 更新设备：设置 `monitoring_enabled = TRUE`
- 触发卡片创建：设备更新后会自动调用 `CreateCardsForUnit`

**检查 SQL**：
```sql
SELECT 
    d.device_id,
    d.device_name,
    d.monitoring_enabled
FROM devices d
WHERE d.device_name LIKE '%Sleepad-yfyy%';
```

### 问题 3：设备状态为 disabled

**症状**：
- 设备状态为 `disabled`，即使 `monitoring_enabled = TRUE` 也不会创建卡片

**解决方案**：
- 更新设备状态：设置 `status = 'active'` 或其他非 'disabled' 状态

**检查 SQL**：
```sql
SELECT 
    d.device_id,
    d.device_name,
    d.status,
    d.monitoring_enabled
FROM devices d
WHERE d.device_name LIKE '%Sleepad-yfyy%';
```

### 问题 4：设备没有通过 room 关联到 unit

**症状**：
- 设备 `bound_room_id IS NULL`，无法通过 room 关联到 unit

**解决方案**：
- 将设备绑定到 room：设置 `bound_room_id`（该 room 必须属于目标 unit）

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
LEFT JOIN units u ON r.unit_id = u.unit_id
WHERE d.device_name LIKE '%Sleepad-yfyy%';
```

### 问题 5：Unit 下已有 ActiveBed，但设备未绑床

**症状**：
- Unit 下有 ActiveBed（其他设备已绑床）
- 当前设备只绑定到 unit，未绑床
- 根据规则，场景 B 下只有未绑床的设备数量 > 0 时才会创建 UnitCard

**解决方案**：
1. 将设备绑定到具体的 bed（推荐）
2. 或者确保该 unit 下没有 ActiveBed（场景 C）

**检查 SQL**：
```sql
-- 检查 Unit 下的 ActiveBed 数量
SELECT 
    COUNT(DISTINCT b.bed_id) AS active_bed_count
FROM beds b
INNER JOIN rooms r ON b.room_id = r.room_id
INNER JOIN units u ON r.unit_id = u.unit_id
INNER JOIN devices d ON d.bound_bed_id = b.bed_id
WHERE u.unit_name = 'E102'
  AND d.monitoring_enabled = TRUE
  AND d.status <> 'disabled'
GROUP BY u.unit_id;
```

## 诊断步骤

### 步骤 1：检查设备基本信息
```sql
SELECT 
    d.device_id,
    d.device_name,
    d.device_uid,
    d.unit_id,
    d.bound_room_id,
    d.bound_bed_id,
    d.monitoring_enabled,
    d.status,
    ds.device_type
FROM devices d
JOIN device_store ds ON d.device_id = ds.device_id
WHERE d.device_name LIKE '%Sleepad-yfyy%';
```

### 步骤 2：检查 Unit 下的 ActiveBed
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
    d.monitoring_enabled,
    d.status
FROM devices d
LEFT JOIN rooms r ON d.bound_room_id = r.room_id
LEFT JOIN units u ON r.unit_id = u.unit_id
WHERE u.unit_name = 'E102'
  AND d.bound_bed_id IS NULL
  AND d.bound_room_id IS NOT NULL
  AND d.monitoring_enabled = TRUE;
```

### 步骤 4：手动触发卡片创建
如果设备配置正确但仍未创建卡片，可以手动触发：

```bash
# 通过 API 触发卡片创建
curl -X POST http://localhost:8082/api/v1/cards/create \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "YOUR_TENANT_ID",
    "unit_id": "UNIT_ID_FOR_E102"
  }'
```

## 修复建议

根据诊断结果，按以下优先级修复：

1. **如果 `monitoring_enabled = FALSE`**：
   - 更新设备：设置 `monitoring_enabled = TRUE`

2. **如果 `status = 'disabled'`**：
   - 更新设备状态：设置 `status = 'active'`

3. **如果 `bound_bed_id IS NULL` 且 unit 下有 ActiveBed**：
   - 将设备绑定到具体的 bed：设置 `bound_bed_id`

4. **如果 `bound_room_id IS NULL`**：
   - 将设备绑定到 room：设置 `bound_room_id`（该 room 必须属于目标 unit）

5. **如果所有配置都正确但仍未创建卡片**：
   - 检查 `wisefido-card-manage` 服务是否正常运行
   - 查看服务日志：`Failed to create cards for unit`
   - 手动触发卡片创建 API

## 删除设备与 device_store.tenant_id

**DeleteDevice（硬删除）**：

1. **devices**：物理删除对应行。
2. **device_store**：`tenant_id` 改为 **未分配** `00000000-0000-0000-0000-000000000000`，**不是** System 租户 `00000000-0000-0000-0000-000000000001`。

若需「删后迁回 System」，须改 `wisefido-data/internal/repository/postgres_devices.go` 中 `DeleteDevice` 的 `unallocatedTenantID` 为 System 租户 UUID。

## 相关代码位置

- 卡片创建逻辑：`owl-common/card/creator.go`
- ActiveBed 查询：`wisefido-card-manage/internal/repository/card.go:GetActiveBedsByUnit`
- 设备更新触发：`wisefido-data/internal/service/device_service.go:UpdateDevice`
- 删除设备与 device_store 回迁：`wisefido-data/internal/repository/postgres_devices.go:DeleteDevice`
- 卡片创建规则：`owlRD/docs/20_Card_Creation_Rules_Final.md`
