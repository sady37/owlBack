# Card Name 状态提示讨论

## 当前逻辑分析

### 当前 Card Name 生成规则

#### ActiveBed Card Name

```go
// 规则：
// 1. If bed is bound to resident → use resident's nickname
// 2. If bed is not bound to resident:
//   - Non-multi-person room → show first resident's nickname under unit
//   - Multi-person room → show 'disable monitor'
// 3. If unit has no residents → show 'Unknown'
```

**当前行为**：
- ✅ Bed 绑定 resident → 显示 resident nickname（如：`"Smith"`）
- ⚠️ Bed 未绑定 resident，非多人房间，unit 下有 resident → 显示 unit 下第一个 resident nickname（如：`"Smith"`）
  - **问题**：用户可能不知道这个 bed 实际上没有绑定 resident
- ✅ Bed 未绑定 resident，多人房间 → 显示 `"disable monitor"`
- ⚠️ Bed 未绑定 resident，unit 也没有 resident → 显示 `"Unknown"`
  - **问题**：不够明确，用户不知道是 bed 未绑定还是 unit 没有 resident

#### UnitCard Name

```go
// 规则：
// 1. is_public = TRUE → unit_name
// 2. is_shared_unit = TRUE → unit_name
// 3. unit_type = 'HomeCare' and unit has residents → first resident's nickname
// 4. is_shared_unit = FALSE and unit has residents → first resident's nickname
// 5. If no residents → unit_name
```

**当前行为**：
- ✅ 公共空间/多人房间 → 显示 unit_name（如：`"E203"`）
- ✅ Unit 下有 resident → 显示第一个 resident nickname（如：`"Smith"`）
- ✅ Unit 下没有 resident → 显示 unit_name（如：`"E203"`）

## 需要检查的状态

### 1. Bed 未绑定 Resident

**当前处理**：
- 多人房间：显示 `"disable monitor"` ✅ 明确
- 非多人房间，unit 下有 resident：显示 unit 下第一个 resident nickname ⚠️ 不够明确
- 非多人房间，unit 下没有 resident：显示 `"Unknown"` ⚠️ 不够明确

**建议**：
- 非多人房间，bed 未绑定 resident，unit 下有 resident：
  - 选项 A：`"Smith [未绑定床位]"` 或 `"Smith (Bed Not Assigned)"`
  - 选项 B：`"BedA [未绑定住户]"` 或 `"BedA (No Resident)"`
  - 选项 C：保持现状，显示 `"Smith"`（因为 unit 下有 resident，可以认为这个 bed 属于该 resident）

- 非多人房间，bed 未绑定 resident，unit 下也没有 resident：
  - 选项 A：`"BedA [未绑定住户]"` 或 `"BedA (No Resident)"`
  - 选项 B：`"Unknown [未绑定住户]"` 或 `"Unknown (No Resident)"`
  - 选项 C：保持现状，显示 `"Unknown"`

### 2. 设备 monitoring_enabled = FALSE

**当前处理**：
- 设备 `monitoring_enabled = FALSE` 不会被计入 ActiveBed
- 不会出现在 card 的 devices 列表中
- Card name 没有提示

**建议**：
- 如果 bed 有部分设备 `monitoring_enabled = FALSE`：
  - 选项 A：`"Smith [部分设备未启用]"` 或 `"Smith (Partial Monitoring)"`
  - 选项 B：`"Smith [监控未启用]"` 或 `"Smith (Monitoring Disabled)"`
  - 选项 C：保持现状，不提示（因为只有启用的设备才会被监控）

- 如果 bed 的所有设备 `monitoring_enabled = FALSE`：
  - 当前：不会生成 card
  - 建议：是否需要生成 card 并提示？如：`"Smith [监控未启用]"` 或 `"Smith (Monitoring Disabled)"`

### 3. Unit 下没有 Resident

**当前处理**：
- ActiveBed card：显示 `"Unknown"`
- UnitCard：显示 `unit_name`

**建议**：
- ActiveBed card，unit 下没有 resident：
  - 选项 A：`"BedA [未绑定住户]"` 或 `"BedA (No Resident)"`
  - 选项 B：保持现状，显示 `"Unknown"`

- UnitCard，unit 下没有 resident：
  - 当前：显示 `unit_name` ✅ 合理（公共空间或空房间）

## 综合建议

### 方案：统一的状态提示格式

**建议格式**：`"基础名称 [状态提示]"`

**状态提示类型**：
1. `[未绑定床位]` - Bed 未绑定 resident，但 unit 下有 resident
2. `[未绑定住户]` - Bed 或 Unit 下没有 resident
3. `[监控未启用]` - 设备 monitoring_enabled = FALSE
4. `[部分设备未启用]` - 部分设备 monitoring_enabled = FALSE

**示例**：
- `"Smith"` - 正常情况（bed 绑定 resident，设备监控启用）
- `"Smith [未绑定床位]"` - Bed 未绑定 resident，但 unit 下有 resident Smith
- `"BedA [未绑定住户]"` - Bed 未绑定 resident，unit 下也没有 resident
- `"Smith [监控未启用]"` - Bed 有设备但 monitoring_enabled = FALSE
- `"Smith [部分设备未启用]"` - Bed 有部分设备 monitoring_enabled = FALSE

### 优先级规则

当有多个状态需要提示时，优先级：
1. **监控状态**（最高优先级）
   - `[监控未启用]` > `[部分设备未启用]`
2. **绑定状态**（次优先级）
   - `[未绑定住户]` > `[未绑定床位]`

**示例**：
- Bed 未绑定 resident + 设备监控未启用 → `"BedA [未绑定住户] [监控未启用]"`
- Bed 绑定 resident + 部分设备监控未启用 → `"Smith [部分设备未启用]"`

## 实现复杂度

### 简单实现（推荐）

只添加最关键的提示：
1. **Bed 未绑定 resident**（非多人房间，unit 下有 resident）→ `"Smith [未绑定床位]"`
2. **设备 monitoring_enabled = FALSE**（所有设备）→ `"Smith [监控未启用]"`

**优点**：
- 实现简单
- 覆盖主要场景
- 用户友好

### 完整实现

添加所有状态提示：
1. Bed 未绑定 resident（所有情况）
2. 设备 monitoring_enabled = FALSE（所有/部分）
3. Unit 下没有 resident

**优点**：
- 信息完整
- 用户清楚所有状态

**缺点**：
- 实现复杂
- Card name 可能很长

## 决策点

请确认：

1. **是否需要为 Bed 未绑定 resident 添加提示**？
   - [ ] 是，添加 `[未绑定床位]` 或 `[未绑定住户]` 提示
   - [ ] 否，保持现状

2. **如果需要，提示文本格式**？
   - [ ] `"ResidentName [未绑定床位]"`
   - [ ] `"BedName [未绑定住户]"`
   - [ ] `"ResidentName (Bed Not Assigned)"`
   - [ ] 其他：______

3. **是否需要为设备 monitoring_enabled = FALSE 添加提示**？
   - [ ] 是，添加 `[监控未启用]` 或 `[部分设备未启用]` 提示
   - [ ] 否，保持现状

4. **如果需要，提示文本格式**？
   - [ ] `"ResidentName [监控未启用]"`
   - [ ] `"ResidentName [部分设备未启用]"`
   - [ ] `"ResidentName (Monitoring Disabled)"`
   - [ ] 其他：______

5. **是否需要为所有设备 monitoring_enabled = FALSE 的 bed 生成 card**？
   - [ ] 是，生成 card 并提示 `[监控未启用]`
   - [ ] 否，不生成 card（当前行为）

6. **多人房间的 "disable monitor" 是否需要修改**？
   - [ ] 保持 `"disable monitor"`
   - [ ] 改为 `"BedName [未绑定住户]"` 或类似格式
   - [ ] 其他：______

---

## 总结

**当前状态**：
- ✅ 多人房间 bed 未绑定 resident → `"disable monitor"`（明确）
- ⚠️ 非多人房间 bed 未绑定 resident → 显示 unit 下第一个 resident nickname（不够明确）
- ⚠️ 设备 monitoring_enabled = FALSE → 没有提示
- ⚠️ Unit 下没有 resident → `"Unknown"`（不够明确）

**建议**：
- 统一使用 `[状态提示]` 格式
- 优先实现：Bed 未绑定 resident + 设备 monitoring_enabled = FALSE 的提示
- 根据实际需求决定是否需要为所有设备 monitoring_enabled = FALSE 的 bed 生成 card

