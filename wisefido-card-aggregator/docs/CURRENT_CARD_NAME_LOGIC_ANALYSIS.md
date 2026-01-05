# 当前 Card Name 生成逻辑分析

## 代码检查结果

### 当前实现：`calculateActiveBedCardName`

**位置**：`internal/aggregator/card_creator.go:329-368`

**当前逻辑**：

```go
func (c *CardCreator) calculateActiveBedCardName(
	tenantID string,
	bed repository.ActiveBedInfo,
	unitInfo *repository.UnitInfo,
) (string, error) {
	// 1. Check if bed is bound to resident
	if bed.ResidentID != nil {
		resident, err := c.repo.GetResidentByBed(tenantID, bed.BedID)
		if err != nil {
			return "", fmt.Errorf("failed to get resident: %w", err)
		}
		if resident != nil {
			return resident.Nickname, nil  // ✅ 正常情况，无提示
		}
	}

	// 2. Bed is not bound to resident, decide based on unit's is_shared_unit
	if unitInfo.IsSharedUnit {
		return "disable monitor", nil  // ✅ 有提示：多人房间 bed 未绑人
	}

	// 3. Non-multi-person room, get first resident's nickname under unit
	residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
	if err != nil {
		return "", fmt.Errorf("failed to get unit residents: %w", err)
	}

	if len(residents) > 0 {
		return residents[0].Nickname, nil  // ⚠️ 没有提示 bed 未绑定
	}

	// 4. If unit has no residents, return default value
	return "Unknown", nil  // ⚠️ 没有明确提示原因
}
```

## 当前检查状态总结

### ✅ 已有提示的情况

1. **多人房间 bed 未绑定 resident**
   - 提示：`"disable monitor"`
   - 位置：`calculateActiveBedCardName` 第 353 行
   - 说明：明确提示多人房间的 bed 未绑人时，设备应该关闭 monitor

### ⚠️ 没有提示的情况

1. **非多人房间 bed 未绑定 resident，但 unit 下有 resident**
   - 当前：显示 unit 下第一个 resident nickname（如：`"Smith"`）
   - 问题：用户不知道这个 bed 实际上没有绑定 resident
   - 位置：`calculateActiveBedCardName` 第 362-363 行

2. **Bed 未绑定 resident，unit 下也没有 resident**
   - 当前：显示 `"Unknown"`
   - 问题：不够明确，用户不知道是 bed 未绑定还是其他原因
   - 位置：`calculateActiveBedCardName` 第 367 行

3. **设备 monitoring_enabled = FALSE**
   - 当前：没有检查，没有提示
   - 问题：用户不知道有设备但监控未启用
   - 位置：整个 card 生成流程中都没有检查

## 代码流程分析

### Card 生成流程（`CreateCardsForUnit`）

```go
func (c *CardCreator) CreateCardsForUnit(tenantID, unitID string) error {
	// 1. Get unit information
	unitInfo, err := c.repo.GetUnitInfo(tenantID, unitID)
	
	// 2. Get all ActiveBeds under this unit
	//    ⚠️ 只查询 monitoring_enabled = TRUE 的设备
	activeBeds, err := c.repo.GetActiveBedsByUnit(tenantID, unitID)
	
	// 3. Delete all old cards under this unit (recreate)
	
	// 4. Determine scenario and create cards
	//    - Scenario A: 1 ActiveBed
	//    - Scenario B: Multiple ActiveBeds
	//    - Scenario C: No ActiveBed
}
```

### ActiveBed 判断（`GetActiveBedsByUnit`）

```sql
-- 只查询 monitoring_enabled = TRUE 的设备
WHERE d.monitoring_enabled = TRUE
  AND d.status <> 'disabled'
```

**结果**：
- ⚠️ 设备 `monitoring_enabled = FALSE` 不会被计入 ActiveBed
- ⚠️ 如果 bed 的所有设备都是 `monitoring_enabled = FALSE`，不会生成 card

### Card Name 计算（`calculateActiveBedCardName`）

**检查项**：
1. ✅ Bed 是否绑定 resident（`bed.ResidentID != nil`）
2. ✅ Unit 是否是 shared unit（`unitInfo.IsSharedUnit`）
3. ✅ Unit 下是否有 resident（`len(residents) > 0`）
4. ❌ **没有检查**：设备 monitoring_enabled 状态
5. ❌ **没有检查**：bed 上是否有 monitoring_enabled = FALSE 的设备

**提示情况**：
- ✅ 多人房间 bed 未绑人 → `"disable monitor"`
- ❌ 非多人房间 bed 未绑人 → 显示 unit 下第一个 resident nickname（无提示）
- ❌ Unit 下没有 resident → `"Unknown"`（无明确提示）
- ❌ 设备 monitoring_enabled = FALSE → 无提示（甚至不会生成 card）

## 刷新逻辑

### 轮询刷新（`createAllCards`）

```go
func (s *AggregatorService) createAllCards(ctx context.Context) error {
	// 获取所有 unit
	unitIDs, err := s.cardRepo.GetAllUnits(tenantID)
	
	// 为每个 unit 创建卡片
	for _, unitID := range unitIDs {
		if err := s.cardCreator.CreateCardsForUnit(tenantID, unitID); err != nil {
			// 错误处理
		}
	}
}
```

**刷新时的检查**：
- 与生成时相同，调用 `CreateCardsForUnit`
- 使用相同的 `calculateActiveBedCardName` 逻辑
- **没有额外的状态检查**

### 事件驱动刷新（`processEvent`）

```go
func (c *EventConsumer) processEvent(ctx context.Context, msg rediscommon.StreamMessage) error {
	switch event.EventType {
	case "device.bound", "device.unbound", "device.monitoring_changed":
		return c.cardCreator.CreateCardsForUnit(event.TenantID, event.UnitID)
	case "resident.bound", "resident.unbound", "resident.status_changed":
		return c.cardCreator.CreateCardsForUnit(event.TenantID, event.UnitID)
	// ...
	}
}
```

**刷新时的检查**：
- 与生成时相同，调用 `CreateCardsForUnit`
- 使用相同的逻辑
- **没有额外的状态检查**

## 总结

### 当前已有的检查

1. ✅ **Bed 是否绑定 resident** - 检查了，但只在多人房间时有提示
2. ✅ **Unit 是否是 shared unit** - 检查了，用于决定 card name
3. ✅ **Unit 下是否有 resident** - 检查了，用于决定 card name

### 当前缺失的检查

1. ❌ **设备 monitoring_enabled 状态** - 没有检查，没有提示
2. ❌ **Bed 未绑定 resident 的明确提示** - 非多人房间时没有提示
3. ❌ **Unit 下没有 resident 的明确提示** - 只显示 "Unknown"

### 当前已有的提示

1. ✅ `"disable monitor"` - 多人房间 bed 未绑人时

### 当前缺失的提示

1. ❌ Bed 未绑定 resident（非多人房间）
2. ❌ Unit 下没有 resident
3. ❌ 设备 monitoring_enabled = FALSE

---

## 建议

基于当前代码分析，建议添加以下检查：

1. **Bed 未绑定 resident 的提示**（非多人房间）
   - 当前：显示 unit 下第一个 resident nickname
   - 建议：添加 `[未绑定床位]` 提示

2. **设备 monitoring_enabled = FALSE 的检查**
   - 当前：没有检查
   - 建议：检查并添加 `[监控未启用]` 提示

3. **Unit 下没有 resident 的明确提示**
   - 当前：显示 "Unknown"
   - 建议：改为 `"BedA [未绑定住户]"` 或类似格式

