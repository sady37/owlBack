# 数据库触发器 vs 应用层同步：决策指南

## 当前系统现状

### 已使用触发器的场景
1. **Tag 同步** (`22_tags_catalog.sql`):
   - `family_tag` → `tags_catalog.tag_objects` (行 515-519)
   - `branch_tag` → `tags_catalog.tag_objects` (行 557-561)
   - `area_tag` → `tags_catalog.tag_objects` (行 599-603)
   - `user_tag` → `tags_catalog.tag_objects` (行 694-698)
   - `alarm_tag` → `tags_catalog.tag_objects` (行 648-652)

2. **数据清理** (`22_tags_catalog.sql`):
   - 对象删除时自动清理 tag_objects (行 705-763)

3. **完整性约束** (`25_integrity_constraints.sql`):
   - 跨租户验证 (10个触发器)
   - 关系一致性验证 (5个触发器)
   - 业务规则验证 (2个触发器)

### 已使用应用层的场景
1. **权限控制** (`admin_tags_handlers.go`):
   - 删除 family_tag member 时，同步清空 `residents.family_tag` (行 315-329)
   - 删除 user_tag member 时，同步更新 `users.tags` (行 297-313)

2. **业务逻辑**:
   - 角色权限检查
   - 数据验证
   - 错误处理

---

## 两种方案对比

### 方案 A: 数据库触发器

#### ✅ 优点
1. **数据一致性保证**
   - 无论从哪个入口修改数据，都会自动同步
   - 防止遗漏同步逻辑
   - 数据库层面保证 ACID 特性

2. **性能优势**
   - 在数据库内部执行，减少网络往返
   - 事务内完成，无需额外事务管理

3. **维护简单**
   - 逻辑集中在一个地方（数据库函数）
   - 不受应用层代码重构影响
   - 支持多应用访问同一数据库

4. **安全性**
   - 即使应用层有 bug，数据库层面也能保证一致性
   - 防止绕过应用层直接操作数据库导致的数据不一致

#### ❌ 缺点
1. **调试困难**
   - 触发器执行过程不直观
   - 错误信息可能不够详细
   - 需要数据库日志才能追踪

2. **测试复杂**
   - 需要数据库环境才能测试
   - 单元测试困难
   - 集成测试依赖数据库状态

3. **版本控制**
   - SQL 脚本版本管理
   - 迁移脚本需要仔细设计
   - 回滚可能复杂

4. **灵活性限制**
   - 难以实现复杂的业务逻辑
   - 错误处理能力有限
   - 难以实现条件分支逻辑

5. **可观测性差**
   - 难以添加日志和监控
   - 性能问题难以追踪
   - 业务指标难以统计

6. **跨服务问题**
   - 如果未来需要跨服务同步，触发器无法实现
   - 无法调用外部 API
   - 无法发送消息队列事件

---

### 方案 B: 应用层同步

#### ✅ 优点
1. **灵活性强**
   - 可以实现复杂的业务逻辑
   - 支持条件分支和错误处理
   - 可以调用外部服务

2. **可测试性**
   - 易于单元测试
   - 可以 mock 依赖
   - 测试覆盖率高

3. **可观测性**
   - 可以添加详细的日志
   - 可以添加监控指标
   - 可以追踪性能问题

4. **版本控制**
   - 代码版本控制更直观
   - 代码审查更容易
   - 回滚更简单

5. **跨服务支持**
   - 可以调用其他服务 API
   - 可以发送消息队列事件
   - 可以集成外部系统

#### ❌ 缺点
1. **数据一致性风险**
   - 容易遗漏同步逻辑
   - 多个入口修改数据时，需要每个入口都实现同步
   - 如果应用层崩溃，可能导致数据不一致

2. **性能开销**
   - 需要额外的数据库查询
   - 网络往返增加延迟
   - 事务管理更复杂

3. **维护成本**
   - 需要在多个地方实现相同逻辑
   - 代码重复风险
   - 重构时容易遗漏

4. **安全性**
   - 如果绕过应用层直接操作数据库，不会触发同步
   - 需要额外的权限控制

---

## 决策建议

### 推荐使用触发器的场景

#### ✅ 适合触发器的场景
1. **数据一致性要求高**
   - 必须保证数据同步的场景
   - 防止数据不一致会导致严重问题的场景
   - 例如：tag_objects 同步

2. **简单的一对一同步**
   - 逻辑简单，不需要复杂判断
   - 例如：A 表字段变化 → B 表字段同步

3. **性能敏感**
   - 高频操作，需要最小延迟
   - 例如：每次更新都需要同步

4. **多应用访问**
   - 多个应用访问同一数据库
   - 需要统一的数据同步逻辑

5. **数据清理**
   - 对象删除时自动清理关联数据
   - 例如：resident 删除时清理 tag_objects

#### 当前系统中的触发器使用 ✅ 合理
- `family_tag` → `tag_objects`: ✅ 适合（数据一致性要求高）
- `branch_tag` → `tag_objects`: ✅ 适合（数据一致性要求高）
- `area_tag` → `tag_objects`: ✅ 适合（数据一致性要求高）
- `user_tag` → `tag_objects`: ✅ 适合（数据一致性要求高）
- 对象删除时清理: ✅ 适合（防止数据残留）

---

### 推荐使用应用层的场景

#### ✅ 适合应用层的场景
1. **复杂业务逻辑**
   - 需要条件判断、权限检查
   - 需要调用外部服务
   - 需要发送通知

2. **需要可观测性**
   - 需要详细的日志和监控
   - 需要性能追踪
   - 需要业务指标统计

3. **跨服务同步**
   - 需要调用其他服务 API
   - 需要发送消息队列事件
   - 需要集成外部系统

4. **需要灵活的错误处理**
   - 需要详细的错误信息
   - 需要重试机制
   - 需要降级策略

5. **测试要求高**
   - 需要高测试覆盖率
   - 需要单元测试
   - 需要集成测试

#### 当前系统中的应用层使用 ✅ 合理
- 删除 member 时同步 `residents.family_tag`: ✅ 适合（需要权限检查）
- 删除 member 时同步 `users.tags`: ✅ 适合（需要权限检查）

---

## 混合方案（推荐）

### 最佳实践：触发器 + 应用层

#### 原则
1. **触发器负责数据一致性**
   - 简单的、必须保证的同步逻辑
   - 例如：A 表 → B 表的字段同步

2. **应用层负责业务逻辑**
   - 复杂的业务规则
   - 权限检查
   - 错误处理和日志

#### 当前系统的混合使用 ✅ 合理

**触发器层**（数据一致性）:
```sql
-- 当 residents.family_tag 变化时，自动同步到 tags_catalog.tag_objects
CREATE TRIGGER trigger_sync_family_tag
    AFTER INSERT OR UPDATE OF family_tag ON residents
    FOR EACH ROW
    EXECUTE FUNCTION sync_family_tag_to_catalog();
```

**应用层**（业务逻辑）:
```go
// 当从 Tags Management 删除 member 时，需要权限检查后同步
if objectType == "resident" && tagType == "family_tag" {
    // 权限检查已在前面完成
    // 同步清空 resident.family_tag
    _, err = s.DB.ExecContext(...)
}
```

---

## 统一策略建议

### 建议：根据场景具体分析，但遵循以下原则

#### 1. 数据一致性层（触发器）
- **职责**: 保证数据一致性，防止数据不一致
- **场景**: 
  - 简单的字段同步
  - 对象删除时的自动清理
  - 必须保证的同步逻辑

#### 2. 业务逻辑层（应用层）
- **职责**: 实现业务规则，权限控制，错误处理
- **场景**:
  - 需要权限检查的操作
  - 需要复杂判断的逻辑
  - 需要日志和监控的操作
  - 需要调用外部服务的操作

#### 3. 决策流程

```
是否需要数据一致性保证？
├─ 是 → 是否需要复杂业务逻辑？
│   ├─ 是 → 混合方案（触发器 + 应用层）
│   └─ 否 → 使用触发器
└─ 否 → 使用应用层
```

---

## 当前系统的问题和建议

### 问题 1: family_tag 删除时触发器不执行

**当前状态**:
```sql
CREATE TRIGGER trigger_sync_family_tag
    AFTER INSERT OR UPDATE OF family_tag ON residents
    FOR EACH ROW
    WHEN (NEW.family_tag IS NOT NULL)  -- ❌ Bug: 设置为 NULL 时不执行
    EXECUTE FUNCTION sync_family_tag_to_catalog();
```

**修复建议**:
```sql
-- 方案 1: 移除 WHEN 条件（推荐）
CREATE TRIGGER trigger_sync_family_tag
    AFTER INSERT OR UPDATE OF family_tag ON residents
    FOR EACH ROW
    EXECUTE FUNCTION sync_family_tag_to_catalog();

-- 方案 2: 修改 WHEN 条件
CREATE TRIGGER trigger_sync_family_tag
    AFTER INSERT OR UPDATE OF family_tag ON residents
    FOR EACH ROW
    WHEN (OLD.family_tag IS NOT NULL OR NEW.family_tag IS NOT NULL)
    EXECUTE FUNCTION sync_family_tag_to_catalog();
```

**原因**: 函数 `sync_family_tag_to_catalog()` 已经有删除逻辑（行 484），只需要让触发器在设置为 NULL 时也能执行。

---

### 问题 2: 双向同步的不一致性

**当前状态**:
- ✅ 从 Resident Profile 设置 family_tag → 触发器自动添加到 tag_objects
- ❌ 从 Resident Profile 删除 family_tag → 触发器不执行（Bug）
- ✅ 从 Tags Management 删除 member → 应用层自动清空 resident.family_tag

**建议**: 
- 修复触发器 Bug，让删除也能自动同步
- 保持混合方案：触发器负责数据一致性，应用层负责权限检查

---

## 总结

### 推荐策略

1. **数据一致性保证** → 使用触发器
   - 简单的字段同步
   - 对象删除时的自动清理
   - 必须保证的同步逻辑

2. **业务逻辑** → 使用应用层
   - 权限检查
   - 复杂判断
   - 错误处理和日志

3. **混合方案** → 最佳实践
   - 触发器保证数据一致性
   - 应用层实现业务逻辑
   - 两者互补，各司其职

### 当前系统评估

✅ **整体设计合理**:
- 触发器用于数据一致性（tag_objects 同步）
- 应用层用于业务逻辑（权限检查、错误处理）
- 混合方案符合最佳实践

⚠️ **需要修复**:
- family_tag 删除时触发器不执行（Bug）
- 建议统一所有 tag_type 的触发器逻辑

### 建议

1. **保持混合方案**，但明确职责划分
2. **修复触发器 Bug**，确保删除操作也能自动同步
3. **统一触发器逻辑**，所有 tag_type 使用相同的触发器模式
4. **文档化决策**，记录每个同步逻辑使用触发器还是应用层的原因

