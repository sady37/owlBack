# monitoring_enabled = FALSE 时的 Card Name 处理讨论

## 当前逻辑分析

### 当前行为

1. **ActiveBed 判断**：
   - 只查询 `monitoring_enabled = TRUE` 的设备
   - 如果 bed 的所有设备都是 `monitoring_enabled = FALSE`，这个 bed **不会成为 ActiveBed**
   - 因此**不会生成 card**

2. **Card 中的设备**：
   - 只有 `monitoring_enabled = TRUE` 的设备会被绑定到 card
   - `monitoring_enabled = FALSE` 的设备**不会出现在 card 中**

3. **Card Name 规则**（当前）：
   - ActiveBed card：使用 resident nickname 或 "disable monitor"（针对 shared unit）
   - UnitCard：使用 unit_name 或 resident nickname
   - **没有针对 `monitoring_enabled = FALSE` 的特殊提示**

## 问题场景分析

### 场景 1：Bed 有设备，但所有设备 monitoring_enabled = FALSE

**当前行为**：
- Bed 不会成为 ActiveBed
- 不会生成 card
- 前端看不到这个 bed

**问题**：
- 如果 bed 有 resident，但设备监控未启用，用户可能不知道
- 是否需要生成 card 并提示监控未启用？

### 场景 2：Bed 有部分设备 monitoring_enabled = TRUE，部分为 FALSE

**当前行为**：
- Bed 会成为 ActiveBed（因为有 `monitoring_enabled = TRUE` 的设备）
- 只绑定 `monitoring_enabled = TRUE` 的设备到 card
- Card name 使用 resident nickname（正常显示）

**问题**：
- 用户可能不知道还有 `monitoring_enabled = FALSE` 的设备
- 是否需要在 card_name 中提示？

### 场景 3：Bed 有设备且 monitoring_enabled = TRUE，但后来被禁用

**当前行为**：
- 当设备 `monitoring_enabled` 从 TRUE 变为 FALSE 时
- Bed 可能不再满足 ActiveBed 条件
- Card 会被删除（如果所有设备都变为 FALSE）

**问题**：
- 如果 bed 有 resident，但所有设备监控被禁用，card 会消失
- 是否应该保留 card 并提示监控已禁用？

## 可能的方案

### 方案 A：在 Card Name 中添加监控状态提示

**思路**：
- 检查 bed 上是否有 `monitoring_enabled = FALSE` 的设备
- 如果有，在 card_name 后添加提示，如：
  - `"ResidentName [监控未启用]"`
  - `"ResidentName [部分设备未启用]"`
  - `"ResidentName (Monitoring Disabled)"`

**优点**：
- ✅ 用户一眼就能看到监控状态
- ✅ 不需要改变 card 生成逻辑
- ✅ 实现简单

**缺点**：
- ⚠️ Card name 可能变长
- ⚠️ 需要检查所有设备的 monitoring 状态

### 方案 B：为 monitoring_enabled = FALSE 的设备生成特殊 Card

**思路**：
- 即使设备 `monitoring_enabled = FALSE`，如果 bed 有 resident，也生成 card
- Card name 明确提示：`"ResidentName [监控未启用]"` 或 `"ResidentName (Monitoring Disabled)"`
- Card 中绑定所有设备（包括 `monitoring_enabled = FALSE` 的）

**优点**：
- ✅ 确保有 resident 的 bed 都有 card
- ✅ 明确提示监控状态
- ✅ 用户可以看到所有设备

**缺点**：
- ⚠️ 需要修改 ActiveBed 判断逻辑
- ⚠️ 需要修改 card 生成逻辑
- ⚠️ 可能生成大量"监控未启用"的 cards

### 方案 C：在 Card 的 devices 字段中标记监控状态

**思路**：
- Card 中绑定所有设备（包括 `monitoring_enabled = FALSE` 的）
- 在 devices JSONB 中添加 `monitoring_enabled` 字段
- Card name 保持正常（不添加提示）
- 前端根据 `monitoring_enabled` 字段显示不同图标或样式

**优点**：
- ✅ Card name 保持简洁
- ✅ 前端可以灵活显示监控状态
- ✅ 信息更完整

**缺点**：
- ⚠️ 需要修改 card 生成逻辑（绑定所有设备）
- ⚠️ 前端需要处理监控状态显示

### 方案 D：保持现状，不添加特殊提示

**理由**：
- `monitoring_enabled = FALSE` 的设备不应该被监控
- 如果用户需要监控，应该启用设备的 `monitoring_enabled`
- Card 只显示可监控的设备，符合业务逻辑

**优点**：
- ✅ 逻辑简单清晰
- ✅ 不需要修改代码
- ✅ Card 只显示可用的监控设备

**缺点**：
- ❌ 用户可能不知道有设备但监控未启用
- ❌ 如果 bed 有 resident 但所有设备监控未启用，不会生成 card

## 建议

### 推荐：方案 A（在 Card Name 中添加提示）

**理由**：
1. **实现简单**：只需要在 `calculateActiveBedCardName` 中添加检查逻辑
2. **用户友好**：一眼就能看到监控状态
3. **不影响现有逻辑**：仍然只绑定 `monitoring_enabled = TRUE` 的设备

**实现思路**：
```go
func (c *CardCreator) calculateActiveBedCardName(
    tenantID string,
    bed repository.ActiveBedInfo,
    unitInfo *repository.UnitInfo,
) (string, error) {
    // ... 现有的 name 计算逻辑 ...
    baseName := resident.Nickname  // 或 "disable monitor" 等
    
    // 检查是否有 monitoring_enabled = FALSE 的设备
    allDevices, err := c.repo.GetAllDevicesByBed(tenantID, bed.BedID)  // 包括 FALSE 的
    if err != nil {
        return baseName, nil  // 如果查询失败，返回基础名称
    }
    
    hasDisabledMonitoring := false
    for _, device := range allDevices {
        if !device.MonitoringEnabled {
            hasDisabledMonitoring = true
            break
        }
    }
    
    if hasDisabledMonitoring {
        return baseName + " [监控未启用]", nil
    }
    
    return baseName, nil
}
```

**需要实现**：
- `GetAllDevicesByBed` - 获取 bed 的所有设备（包括 `monitoring_enabled = FALSE` 的）

### 备选：方案 C（在 devices 字段中标记）

如果希望更灵活，可以考虑方案 C，让前端处理监控状态的显示。

## 决策点

请确认：

1. **是否需要为 `monitoring_enabled = FALSE` 的设备添加提示**？
   - [ ] 是，在 card_name 中添加提示（方案 A）
   - [ ] 是，在 devices 字段中标记（方案 C）
   - [ ] 是，生成特殊 card（方案 B）
   - [ ] 否，保持现状（方案 D）

2. **如果选择方案 A，提示文本**？
   - [ ] `"ResidentName [监控未启用]"`
   - [ ] `"ResidentName (Monitoring Disabled)"`
   - [ ] `"ResidentName [部分设备未启用]"`（如果有部分设备启用）
   - [ ] 其他：______

3. **是否需要为所有设备 monitoring_enabled = FALSE 的 bed 生成 card**？
   - [ ] 是，生成 card 并提示监控未启用
   - [ ] 否，不生成 card（当前行为）

---

## 总结

**当前问题**：
- 如果设备 `monitoring_enabled = FALSE`，不会出现在 card 中
- 如果 bed 的所有设备都是 `monitoring_enabled = FALSE`，不会生成 card
- Card name 没有提示监控状态

**建议方案**：
- 在 card_name 中添加监控状态提示（方案 A）
- 实现简单，用户友好
- 不影响现有逻辑

