# 简化更新策略讨论

## 用户提出的方案

**核心思路**：
1. 每 60 分钟计算一次期望的 cards（基于当前数据）
2. 与数据库中的现有 cards 比对
3. **如果没有变化，不更新**（跳过）
4. **如果有变化，统一更新**（全量重建）

**轮询间隔选择：60 分钟（3600 秒）**
- 原因：有事件驱动触发更新机制（主要更新方式）
- 即使触发更新失败，新入户第1小时没有生效也很正常
- 配置更新需要时间，病人可能也没有这么快即时入住

## 方案分析

### 优点 ✅

1. **实现简单**
   - 只需要添加一个"比较"步骤
   - 不需要区分 CREATE/UPDATE/DELETE
   - 仍然使用现有的全量重建逻辑（DELETE + INSERT）

2. **避免无意义的操作**
   - 如果数据没有变化，完全跳过更新
   - 大幅减少数据库操作（从每 10 分钟全量重建 → 只在有变化时重建）

3. **保持简单性**
   - 不需要实现复杂的增量更新逻辑
   - 不需要处理 UPDATE 操作
   - 仍然保证数据一致性（全量重建）

4. **轮询间隔合理**
   - 10 分钟（600秒）比 60 秒更合理
   - 对实时性影响可接受
   - 减少系统负载

### 需要实现的功能

1. **获取现有 cards**
   ```go
   GetCardsByUnit(tenantID, unitID) ([]CardInfo, error)
   ```
   - 从数据库查询该 unit 下的所有现有 cards
   - 包括 card_type, bed_id, card_name, card_address, devices, residents 等字段

2. **计算期望的 cards**
   - 使用现有的 card 创建逻辑（但不实际创建）
   - 生成期望的 cards 列表（包括所有字段）

3. **比较逻辑**
   ```go
   compareCards(existing []CardInfo, expected []CardInfo) bool
   ```
   - 比较两个 cards 列表是否相同
   - 需要比较的关键字段：
     - card_type
     - bed_id（对于 ActiveBed cards）
     - card_name
     - card_address
     - devices（JSONB，需要比较内容）
     - residents（JSONB，需要比较内容）
   - 如果相同返回 true，不同返回 false

4. **修改 CreateCardsForUnit**
   ```go
   func (c *CardCreator) CreateCardsForUnit(tenantID, unitID string) error {
       // 1. 获取现有 cards
       existingCards := c.repo.GetCardsByUnit(tenantID, unitID)
       
       // 2. 计算期望的 cards（不实际创建）
       expectedCards := c.calculateExpectedCards(tenantID, unitID)
       
       // 3. 比较
       if c.compareCards(existingCards, expectedCards) {
           // 没有变化，跳过更新
           c.logger.Debug("Cards unchanged, skipping update", zap.String("unit_id", unitID))
           return nil
       }
       
       // 4. 有变化，执行全量重建（现有逻辑）
       // ... 现有的删除和创建逻辑 ...
   }
   ```

### 实现复杂度评估

**简单部分**：
- ✅ `GetCardsByUnit`：标准的数据库查询，类似现有的 `GetAllCards`
- ✅ 计算期望 cards：可以复用现有的创建逻辑，但不实际执行 INSERT

**中等复杂度部分**：
- ⚠️ **比较逻辑**：需要比较 JSONB 字段（devices, residents）
  - JSONB 比较：需要解析 JSON 并比较内容
  - 需要考虑 JSON 字段顺序（Go 的 json.Marshal 可能顺序不同）
  - 可以使用 `json.Marshal` 后比较字符串，或者解析后深度比较

**潜在问题**：

1. **JSONB 比较的准确性**
   - PostgreSQL 的 JSONB 存储时可能重新排序键
   - 需要确保比较逻辑能正确处理
   - 建议：解析 JSON 后深度比较，而不是比较字符串

2. **性能考虑**
   - 每次轮询都需要查询现有 cards 和计算期望 cards
   - 但相比全量重建，这个开销很小
   - 如果数据没有变化，可以节省大量 DELETE + INSERT 操作

3. **边界情况**
   - 首次运行（没有现有 cards）：应该创建
   - cards 数量变化：需要比较数量
   - cards 内容变化：需要比较内容

## 与之前方案的对比

### 方案对比

| 方案 | 实现复杂度 | 性能提升 | 数据一致性 |
|------|-----------|---------|-----------|
| **当前（60秒全量重建）** | ⭐ 简单 | ❌ 无优化 | ✅ 保证 |
| **方案A（10分钟全量重建）** | ⭐ 简单 | ⭐⭐ 减少80%操作 | ✅ 保证 |
| **方案B（增量更新）** | ⭐⭐⭐⭐ 复杂 | ⭐⭐⭐⭐ 最优 | ✅ 保证 |
| **简化方案（10分钟+比较）** | ⭐⭐ 中等 | ⭐⭐⭐ 大幅减少 | ✅ 保证 |

### 简化方案的优势

相比**方案A（只增加间隔）**：
- ✅ 进一步优化：只在有变化时执行
- ✅ 避免无意义的删除重建

相比**方案B（增量更新）**：
- ✅ 实现简单得多
- ✅ 不需要处理 UPDATE 操作
- ✅ 不需要区分 CREATE/UPDATE/DELETE
- ⚠️ 性能略差（仍然全量重建，但只在有变化时）

## 建议

### 推荐采用简化方案

**理由**：
1. ✅ 实现简单：只需要添加比较逻辑
2. ✅ 效果显著：避免无变化时的全量重建
3. ✅ 风险可控：仍然使用全量重建，保证一致性
4. ✅ 轮询间隔合理：10 分钟比 60 秒更合理

### 实现步骤

1. **第一步**：添加 `GetCardsByUnit` 方法
2. **第二步**：实现 `calculateExpectedCards`（复用现有逻辑，但不创建）
3. **第三步**：实现 `compareCards` 比较逻辑
4. **第四步**：修改 `CreateCardsForUnit`，添加比较步骤
5. **第五步**：修改轮询间隔为 10 分钟（600秒）

### 需要注意的点

1. **JSONB 比较**：确保能正确比较 devices 和 residents 字段
2. **日志记录**：记录是否跳过了更新，便于调试
3. **首次运行**：确保首次运行时（没有现有 cards）能正常创建

## 总结

**简化方案的核心价值**：
- 用简单的比较逻辑，避免无意义的全量重建
- 保持全量重建的简单性和一致性保证
- 大幅减少数据库操作（只在有变化时执行）

**这是一个很好的平衡点**：
- 比只增加间隔更优化
- 比增量更新更简单
- 适合当前的需求和资源

---

## 决策点

请确认：

1. **是否采用简化方案**？
   - [ ] 是，采用简化方案（10分钟+比较）
   - [ ] 否，采用其他方案

2. **如果采用，轮询间隔**？
   - [x] 60 分钟（3600秒，已确定）
   - [ ] 其他：______ 秒

3. **实现优先级**？
   - [ ] 立即实现
   - [ ] 后续实现
   - [ ] 先讨论，暂不实现

