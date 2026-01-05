# Facility 型 Unit 的 Unoccupied 逻辑

## 实现逻辑

### Facility 型 Unit (非 publicSpace)

#### 1. SharedUnit (多人共享)
- **ActiveBed Card**: 检查床上是否有人
  - 床上无人 (`bed.ResidentID == nil`) → CardName: `"Unoccupied"`
  - 床上有人 → CardName: resident nickname

#### 2. 独享型 (非 SharedUnit)
- **ActiveBed Card**: 检查 unit 内是否有人
  - unit 内无人 (`len(residents) == 0`) → CardName: `"Unoccupied"`
  - unit 内有人 → CardName: 第一个 resident nickname
- **UnitCard**: 检查 unit 内是否有人
  - unit 内无人 (`len(residents) == 0`) → CardName: `"Unoccupied"`
  - unit 内有人 → CardName: 第一个 resident nickname

## 代码实现

### `calculateActiveBedCardName`

```go
// 2. Bed is not bound to resident
// Check if this is a Facility type unit (not publicSpace)
isFacility := (unitInfo.UnitType == "Facility" || unitInfo.UnitType == "Institutional") && !unitInfo.IsPublic

if isFacility {
	// Facility 型 unit (非 publicSpace)
	if unitInfo.IsSharedUnit {
		// SharedUnit (多人共享): 检查床上是否有人
		// bed.ResidentID == nil 表示床上无人
		return "Unoccupied", nil
	} else {
		// 独享型 (非 SharedUnit): 检查 unit 内是否有人
		residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
		if err != nil {
			return "", fmt.Errorf("failed to get unit residents: %w", err)
		}
		if len(residents) == 0 {
			// unit 内无人
			return "Unoccupied", nil
		}
		// unit 内有人，显示第一个 resident nickname
		return residents[0].Nickname, nil
	}
}
```

### `calculateUnitCardName`

```go
// Priority 2: Facility 型 unit (非 publicSpace)
isFacility := unitInfo.UnitType == "Facility" || unitInfo.UnitType == "Institutional"
if isFacility && !unitInfo.IsSharedUnit {
	// 独享型 (非 SharedUnit): 检查 unit 内是否有人
	residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
	if err != nil {
		return "", fmt.Errorf("failed to get unit residents: %w", err)
	}
	if len(residents) == 0 {
		// unit 内无人
		return "Unoccupied", nil
	}
	// unit 内有人，继续后续逻辑
}
```

## 逻辑流程图

```
Facility 型 Unit (非 publicSpace)
│
├─ SharedUnit (多人共享)
│  └─ ActiveBed Card
│     ├─ 床上有人 → resident nickname
│     └─ 床上无人 → "Unoccupied"
│
└─ 独享型 (非 SharedUnit)
   ├─ ActiveBed Card
   │  ├─ unit 内有人 → 第一个 resident nickname
   │  └─ unit 内无人 → "Unoccupied"
   │
   └─ UnitCard
      ├─ unit 内有人 → 第一个 resident nickname
      └─ unit 内无人 → "Unoccupied"
```

## 业务意义

**目的**：避免院方陷入风险：未有效监护

- 当 bed 或 unit 无人时，CardName 显示 `"Unoccupied"`，提示院方应该关闭监控
- 降低法律风险：避免在无人监护的情况下继续监控

## 测试场景

### 场景 1: Facility 型 SharedUnit，bed 未绑人
- **期望**: CardName = `"Unoccupied"`

### 场景 2: Facility 型 独享型，unit 内无人
- **期望**: ActiveBed CardName = `"Unoccupied"`, UnitCard CardName = `"Unoccupied"`

### 场景 3: Facility 型 独享型，unit 内有人但 bed 未绑人
- **期望**: ActiveBed CardName = 第一个 resident nickname（不是 "Unoccupied"）

### 场景 4: Facility 型 publicSpace
- **期望**: 按原逻辑处理（不检查 Unoccupied）

### 场景 5: HomeCare 型
- **期望**: 按原逻辑处理（不检查 Unoccupied）

