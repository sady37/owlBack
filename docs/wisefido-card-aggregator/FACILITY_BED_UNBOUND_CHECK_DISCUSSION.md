# Facility 型 Unit 的 Bed 未绑人检查讨论

## 业务需求

**核心问题**：
- 当床上无人时，院方应关闭监控，避免院方陷入风险：未有效监护
- 当前只有多人房间（`is_shared_unit = TRUE`）时，bed 未绑人显示 `"disable monitor"`
- **是否需要为 Facility 型（机构场景）的非多人房间也添加此提示？**

## 当前逻辑分析

### 当前代码：`calculateActiveBedCardName`

```go
func (c *CardCreator) calculateActiveBedCardName(
	tenantID string,
	bed repository.ActiveBedInfo,
	unitInfo *repository.UnitInfo,
) (string, error) {
	// 1. Bed 绑定 resident → 返回 resident nickname
	if bed.ResidentID != nil {
		// ...
		return resident.Nickname, nil
	}

	// 2. Bed 未绑定 resident
	if unitInfo.IsSharedUnit {
		// 多人房间 → 显示 "disable monitor" ✅
		return "disable monitor", nil
	}

	// 3. 非多人房间
	// ⚠️ 当前：显示 unit 下第一个 resident nickname（没有提示 bed 未绑人）
	residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
	if len(residents) > 0 {
		return residents[0].Nickname, nil  // ⚠️ 没有提示
	}

	// 4. Unit 下没有 resident
	return "Unknown", nil  // ⚠️ 没有明确提示
}
```

### 当前问题

**Facility 型（机构场景）的非多人房间**：
- Bed 未绑人，但 unit 下有 resident → 显示第一个 resident nickname
- **问题**：用户不知道这个 bed 实际上没有绑定 resident
- **风险**：如果监控开启，但 bed 未绑人，院方可能陷入"未有效监护"的风险

## Unit Type 说明

根据文档，`unit_type` 的可能值：
- `'Facility'` 或 `'Institutional'` - 机构场景（养老院、护理院等）
- `'HomeCare'` - 家庭场景

**Facility 型的特点**：
- 机构场景，需要严格监护
- 如果 bed 未绑人，应该关闭监控，避免风险
- 与 HomeCare 不同，HomeCare 可能更灵活

## 建议方案

### 方案：增加 Facility 型检测

**逻辑**：
```go
func (c *CardCreator) calculateActiveBedCardName(
	tenantID string,
	bed repository.ActiveBedInfo,
	unitInfo *repository.UnitInfo,
) (string, error) {
	// 1. Bed 绑定 resident → 返回 resident nickname
	if bed.ResidentID != nil {
		return resident.Nickname, nil
	}

	// 2. Bed 未绑定 resident
	// 检查是否是 Facility 型（机构场景）
	isFacility := unitInfo.UnitType == "Facility" || unitInfo.UnitType == "Institutional"
	
	if unitInfo.IsSharedUnit {
		// 多人房间 → 显示 "disable monitor"
		return "disable monitor", nil
	} else if isFacility {
		// ⭐ 新增：Facility 型非多人房间，bed 未绑人 → 也显示 "disable monitor"
		// 原因：避免院方陷入风险：未有效监护
		return "disable monitor", nil
	}

	// 3. 非 Facility 型（如 HomeCare），非多人房间
	residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
	if len(residents) > 0 {
		return residents[0].Nickname, nil
	}

	// 4. Unit 下没有 resident
	return "Unknown", nil
}
```

### 方案优势

1. **符合业务需求**：Facility 型 bed 未绑人时提示关闭监控
2. **降低风险**：避免院方陷入"未有效监护"的风险
3. **逻辑清晰**：
   - 多人房间 bed 未绑人 → "disable monitor"
   - Facility 型 bed 未绑人 → "disable monitor"
   - HomeCare 型 bed 未绑人 → 显示 unit 下第一个 resident nickname（更灵活）

### 需要考虑的问题

1. **Unit Type 的值**：
   - 确认 `unit_type` 的值是 `"Facility"` 还是 `"Institutional"`？
   - 还是两者都需要检查？

2. **HomeCare 场景**：
   - HomeCare 型是否也需要此检查？
   - 还是 HomeCare 更灵活，允许 bed 未绑人但显示 unit 下第一个 resident？

3. **提示文本**：
   - 保持 `"disable monitor"`？
   - 还是使用更明确的提示，如 `"BedA [未绑定住户] - 请关闭监控"`？

## 决策点

请确认：

1. **是否需要为 Facility 型添加 bed 未绑人检查**？
   - [ ] 是，Facility 型 bed 未绑人时也显示 "disable monitor"
   - [ ] 否，保持现状

2. **Unit Type 的值**？
   - [ ] `"Facility"`
   - [ ] `"Institutional"`
   - [ ] 两者都需要检查
   - [ ] 其他：______

3. **HomeCare 场景**？
   - [ ] HomeCare 型也需要此检查
   - [ ] HomeCare 型保持现状（显示 unit 下第一个 resident nickname）

4. **提示文本**？
   - [ ] 保持 `"disable monitor"`
   - [ ] 使用更明确的提示，如 `"BedA [未绑定住户]"` 或 `"BedA [请关闭监控]"`
   - [ ] 其他：______

---

## 总结

**当前逻辑**：
- ✅ 多人房间 bed 未绑人 → "disable monitor"
- ⚠️ Facility 型非多人房间 bed 未绑人 → 显示 unit 下第一个 resident nickname（没有提示）

**建议**：
- ✅ 增加 Facility 型检测
- ✅ Facility 型 bed 未绑人时也显示 "disable monitor"
- ✅ 降低院方风险：避免未有效监护

