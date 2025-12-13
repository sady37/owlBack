# wisefido-sensor-fusion 问题检查报告

## 📋 检查依据

- `owlRD/db/21_cards.sql` - 卡片表结构定义
- `owlRD/docs/20_Card_Creation_Rules_Final.md` - 卡片创建规则

## ✅ 已解决的问题

### 1. 设备 JSONB 格式匹配 ✅
- **cards.sql 要求**：`[{"device_id": "...", "device_name": "...", "device_type": "...", "device_model": "...", "binding_type": "direct|indirect"}, ...]`
- **wisefido-card-aggregator 实现**：`DeviceJSON` 结构体完全匹配
- **wisefido-sensor-fusion 解析**：`DeviceInfo` 结构体完全匹配
- **状态**：✅ 格式一致

### 2. 设备类型过滤 ✅
- **要求**：支持 Radar、Sleepace、SleepPad
- **实现**：`FuseCardData` 中已过滤 `deviceType == "Radar" || deviceType == "Sleepace" || deviceType == "SleepPad"`
- **状态**：✅ 已支持

### 3. 卡片类型支持 ✅
- **cards.sql 要求**：`card_type` 为 `'ActiveBed'` 或 `'Location'`
- **实现**：`CardInfo.CardType` 字段支持两种类型
- **状态**：✅ 已支持

### 4. 从 JSONB 读取设备列表 ✅
- **要求**：从 `cards.devices` JSONB 字段读取设备列表
- **实现**：`GetCardDevices` 正确解析 JSONB
- **状态**：✅ 实现正确

## ⚠️ 潜在问题

### 问题 1：GetCardByDeviceID 查询逻辑

**当前实现**：
```go
// 查询逻辑：
// 1. 如果设备绑定到 bed（bound_bed_id IS NOT NULL）：
//    - 查询 ActiveBed 类型的卡片（cards.bed_id = bound_bed_id）
// 2. 如果设备绑定到 room（bound_room_id IS NOT NULL）：
//    - 查询 Location 类型的卡片（cards.unit_id = room.unit_id）
```

**cards.sql 说明**：
- ActiveBed 卡片：`bed_id` 不为 NULL，`unit_id` 可为 NULL（冗余）
- Location 卡片：`unit_id` 不为 NULL，`bed_id` 为 NULL

**卡片创建规则**：
- 设备绑定优先级：床 > 门牌号
- 如果设备同时绑定床和门牌号，优先归属到床
- Room 仅用于组织结构，卡片创建时不使用

**分析**：
- ✅ 查询逻辑基本正确
- ⚠️ 但根据卡片创建规则，设备可能通过 `unit_id` 绑定到 Location 卡片（未绑床的设备）
- ⚠️ 当前查询只考虑了 `bound_room_id`，但卡片创建时使用的是 `unit_id`

**建议检查**：
- 如果设备 `bound_bed_id IS NULL` 且 `unit_id IS NOT NULL`，应该查询 Location 卡片
- 当前实现通过 `bound_room_id` 查询 `room.unit_id`，这是正确的
- 但还需要考虑直接通过 `devices.unit_id` 查询的情况

### 问题 2：设备绑定到 unit 的情况 ❌ **不需要处理**

**前端绑定规则**（`owlFront/src/views/units/composables/useDevice.ts`）：
- **设备不能直接绑定到 Unit**，必须绑定到 Room 或 Bed
- 当设备绑定到 Unit 时，前端会先调用 `ensureUnitRoom(unit)` 创建 `unit_room`（`room_name === unit_name`），然后绑定到 room
- 所有 Bed 都绑定在 Room 下

**结论**：
- ✅ 设备总是通过 `bound_bed_id` 或 `bound_room_id` 绑定
- ✅ 不需要直接查询 `devices.unit_id`
- ✅ 当前实现（只查询 `bound_bed_id` 和 `bound_room_id`）是正确的

## 🔍 需要验证的场景

### 场景 1：设备绑定到床
- 设备：`bound_bed_id = 'bed-123'`, `unit_id = 'unit-456'`
- 预期：查询到 ActiveBed 卡片（`cards.bed_id = 'bed-123'`）
- 当前实现：✅ 应该能正确查询

### 场景 2：设备绑定到房间
- 设备：`bound_room_id = 'room-789'`, `unit_id = 'unit-456'`（从 room 查询得到）
- 预期：查询到 Location 卡片（`cards.unit_id = 'unit-456'`）
- 当前实现：✅ 应该能正确查询（通过 `room.unit_id`）

### 场景 3：设备只绑定到 unit（未绑床、未绑房间）
- 设备：`bound_bed_id IS NULL`, `bound_room_id IS NULL`, `unit_id = 'unit-456'`
- **前端规则**：❌ **不会出现这种情况**（前端确保设备必须绑定到 Room 或 Bed）
- **当前实现**：✅ **不需要处理**（前端已确保设备总是通过 room 绑定）

## 📝 建议修复

### 修复 GetCardByDeviceID 查询逻辑

添加对 `devices.unit_id` 的直接查询：

```go
func (r *CardRepository) GetCardByDeviceID(tenantID, deviceID string) (*CardInfo, error) {
	query := `
		WITH device_info AS (
			SELECT 
				d.device_id,
				d.tenant_id,
				d.bound_bed_id,
				d.bound_room_id,
				d.unit_id  -- 添加 unit_id
			FROM devices d
			WHERE d.device_id = $1 AND d.tenant_id = $2
		),
		bed_card AS (
			SELECT 
				c.card_id,
				c.tenant_id,
				c.card_type,
				c.bed_id,
				c.unit_id
			FROM cards c
			INNER JOIN device_info di ON c.bed_id = di.bound_bed_id AND c.tenant_id = di.tenant_id
			WHERE di.bound_bed_id IS NOT NULL
			LIMIT 1
		),
		room_card AS (
			SELECT 
				c.card_id,
				c.tenant_id,
				c.card_type,
				c.bed_id,
				c.unit_id
			FROM cards c
			INNER JOIN device_info di ON c.unit_id = (
				SELECT r.unit_id FROM rooms r WHERE r.room_id = di.bound_room_id AND r.tenant_id = di.tenant_id
			) AND c.tenant_id = di.tenant_id
			WHERE di.bound_room_id IS NOT NULL
			LIMIT 1
		),
		unit_card AS (
			-- 新增：直接通过 unit_id 查询 Location 卡片
			SELECT 
				c.card_id,
				c.tenant_id,
				c.card_type,
				c.bed_id,
				c.unit_id
			FROM cards c
			INNER JOIN device_info di ON c.unit_id = di.unit_id AND c.tenant_id = di.tenant_id
			WHERE di.bound_bed_id IS NULL
			  AND di.bound_room_id IS NULL
			  AND di.unit_id IS NOT NULL
			  AND c.card_type = 'Location'
			LIMIT 1
		)
		SELECT card_id, tenant_id, card_type, bed_id, unit_id
		FROM bed_card
		UNION ALL
		SELECT card_id, tenant_id, card_type, bed_id, unit_id
		FROM room_card
		UNION ALL
		SELECT card_id, tenant_id, card_type, bed_id, unit_id
		FROM unit_card
		LIMIT 1
	`
	// ... 后续代码
}
```

## 📊 检查总结

### ✅ 已正确实现
1. 设备 JSONB 格式解析
2. 设备类型过滤（Radar、Sleepace、SleepPad）
3. 卡片类型支持（ActiveBed、Location）
4. 从 JSONB 读取设备列表

### ✅ 已验证
1. **GetCardByDeviceID** 查询逻辑正确
   - 前端确保设备不能直接绑定到 Unit，必须绑定到 Room 或 Bed
   - 当设备绑定到 Unit 时，前端会先创建 `unit_room`，然后绑定到 room
   - 当前实现（只查询 `bound_bed_id` 和 `bound_room_id`）是正确的，不需要直接查询 `devices.unit_id`

### 🔍 需要验证
1. 测试设备绑定到床的场景
2. 测试设备绑定到房间的场景
3. 测试设备只绑定到 unit 的场景（未绑床、未绑房间）

