# TagService 删除策略分析

## 📋 问题描述

当删除某个标签（tag）时，需要从所有使用该标签的实体中移除：
- **User**: `users.tags` (JSONB 数组)
- **Resident**: `residents.family_tag` (VARCHAR)
- **Unit**: `units.branch_tag`, `units.area_tag`, `units.groupList` (JSONB)
- **ResidentCaregiver**: `resident_caregivers.groupList` (JSONB)
- **Card**: `cards.routing_alarm_tags` (VARCHAR[])

**核心问题**：
1. 如果直接删除（不调用其他 Service），很简单
2. 如果要调用 User、Resident、Unit 的 Service，会怎样？

---

## 🔍 当前数据库实现

### 数据库函数：`drop_tag()`

数据库层面已经提供了 `drop_tag(tenant_id, tag_name)` 函数（`22_tags_catalog.sql` 行 305-426），该函数会：

1. **检查是否可以删除**：
   - 系统预定义类型（`branch_tag`, `family_tag`, `area_tag`）不能删除
   - 如果 tag 还在源表中使用，不能删除

2. **自动清理所有使用该 tag 的地方**：
   ```sql
   -- family_tag: 清除 residents.family_tag
   UPDATE residents SET family_tag = NULL WHERE family_tag = p_tag_name;
   
   -- area_tag: 清除 units.area_tag
   UPDATE units SET area_tag = NULL WHERE area_tag = p_tag_name;
   
   -- user_tag: 清除多个地方
   -- 1. units.groupList JSONB 数组
   UPDATE units SET groupList = ... WHERE groupList 包含该 tag;
   
   -- 2. resident_caregivers.groupList JSONB 数组
   UPDATE resident_caregivers SET groupList = ... WHERE groupList 包含该 tag;
   
   -- 3. cards.routing_alarm_tags 数组
   UPDATE cards SET routing_alarm_tags = array_remove(...) WHERE ...;
   ```

3. **删除 tags_catalog 记录**：
   ```sql
   DELETE FROM tags_catalog WHERE tag_id = v_tag_id;
   ```

---

## 🎯 实现方案对比

### 方案1：直接调用数据库函数（简单）✅ 推荐

**实现方式**：
```go
// TagService.DeleteTag
func (s *tagService) DeleteTag(ctx context.Context, req DeleteTagRequest) error {
    // 1. 权限检查
    if !s.hasPermission(ctx, req.UserRole, "tags", "D") {
        return fmt.Errorf("permission denied")
    }
    
    // 2. 调用数据库函数（自动清理所有关联）
    _, err := s.tagRepo.DeleteTag(ctx, req.TenantID, req.TagName)
    return err
}

// Repository.DeleteTag
func (r *postgresTagsRepository) DeleteTag(ctx context.Context, tenantID string, tagName string) error {
    _, err := r.db.ExecContext(ctx, 
        `SELECT drop_tag($1, $2)`, 
        tenantID, tagName)
    return err
}
```

**优点**：
- ✅ **简单**：只需调用一个数据库函数
- ✅ **性能好**：数据库层面批量更新，效率高
- ✅ **原子性**：数据库事务保证一致性
- ✅ **无循环依赖**：不依赖其他 Service
- ✅ **符合数据库设计**：数据库已经提供了完整的清理逻辑

**缺点**：
- ⚠️ **绕过业务规则**：如果 User/Resident/Unit 有业务逻辑需要处理（如事件通知、缓存清理），可能被绕过
- ⚠️ **测试困难**：需要 Mock 数据库函数

**适用场景**：
- ✅ 标签删除是纯数据操作，不需要业务逻辑
- ✅ 数据库函数已经处理了所有清理逻辑
- ✅ 性能要求高

---

### 方案2：调用其他 Service（复杂）❌ 不推荐

**实现方式**：
```go
// TagService.DeleteTag
func (s *tagService) DeleteTag(ctx context.Context, req DeleteTagRequest) error {
    // 1. 权限检查
    if !s.hasPermission(ctx, req.UserRole, "tags", "D") {
        return fmt.Errorf("permission denied")
    }
    
    // 2. 查询所有使用该 tag 的实体
    users, err := s.userService.ListUsersByTag(ctx, req.TagName)
    residents, err := s.residentService.ListResidentsByTag(ctx, req.TagName)
    units, err := s.unitService.ListUnitsByTag(ctx, req.TagName)
    
    // 3. 逐个更新实体（移除 tag）
    for _, user := range users {
        err := s.userService.RemoveTag(ctx, user.UserID, req.TagName)
        if err != nil {
            return err
        }
    }
    for _, resident := range residents {
        err := s.residentService.RemoveTag(ctx, resident.ResidentID, req.TagName)
        if err != nil {
            return err
        }
    }
    for _, unit := range units {
        err := s.unitService.RemoveTag(ctx, unit.UnitID, req.TagName)
        if err != nil {
            return err
        }
    }
    
    // 4. 删除 tags_catalog 记录
    return s.tagRepo.DeleteTag(ctx, req.TenantID, req.TagName)
}
```

**优点**：
- ✅ **符合业务规则**：可以触发业务逻辑（如事件通知、缓存清理）
- ✅ **可测试性好**：可以 Mock 各个 Service

**缺点**：
- ❌ **循环依赖风险**：TagService 依赖 UserService、ResidentService、UnitService
  - 如果 UserService 也需要 TagService（如查询标签），就会形成循环依赖
- ❌ **性能差**：需要多次查询和更新，效率低
- ❌ **事务复杂**：需要跨 Service 事务管理
- ❌ **复杂度高**：需要实现 `ListUsersByTag`、`RemoveTag` 等方法
- ❌ **重复实现**：数据库函数已经实现了清理逻辑

**适用场景**：
- ❌ 不推荐使用（除非有特殊业务需求）

---

### 方案3：Repository 层调用数据库函数（折中）✅ 推荐

**实现方式**：
```go
// TagService.DeleteTag
func (s *tagService) DeleteTag(ctx context.Context, req DeleteTagRequest) error {
    // 1. 权限检查
    if !s.hasPermission(ctx, req.UserRole, "tags", "D") {
        return fmt.Errorf("permission denied")
    }
    
    // 2. 业务规则验证（在 Service 层）
    tag, err := s.tagRepo.GetTagByName(ctx, req.TenantID, req.TagName)
    if err != nil {
        return err
    }
    
    // 系统预定义类型不能删除
    if tag.TagType == "branch_tag" || tag.TagType == "family_tag" || tag.TagType == "area_tag" {
        return fmt.Errorf("cannot delete system predefined tag type: %s", tag.TagType)
    }
    
    // 3. 调用 Repository（Repository 调用数据库函数）
    return s.tagRepo.DeleteTag(ctx, req.TenantID, req.TagName)
}

// Repository.DeleteTag
func (r *postgresTagsRepository) DeleteTag(ctx context.Context, tenantID string, tagName string) error {
    // 调用数据库函数（自动清理所有关联）
    _, err := r.db.ExecContext(ctx, 
        `SELECT drop_tag($1, $2)`, 
        tenantID, tagName)
    return err
}
```

**优点**：
- ✅ **简单**：只需调用数据库函数
- ✅ **性能好**：数据库层面批量更新
- ✅ **原子性**：数据库事务保证一致性
- ✅ **无循环依赖**：不依赖其他 Service
- ✅ **业务规则验证**：在 Service 层验证，Repository 层执行

**缺点**：
- ⚠️ **绕过业务逻辑**：如果 User/Resident/Unit 有业务逻辑需要处理，可能被绕过

**适用场景**：
- ✅ **推荐使用**：平衡了简单性和业务规则验证

---

### 方案4：事件驱动（未来扩展）🔮

**实现方式**：
```go
// TagService.DeleteTag
func (s *tagService) DeleteTag(ctx context.Context, req DeleteTagRequest) error {
    // 1. 权限检查
    // 2. 业务规则验证
    // 3. 调用 Repository 删除
    err := s.tagRepo.DeleteTag(ctx, req.TenantID, req.TagName)
    if err != nil {
        return err
    }
    
    // 4. 发布事件
    s.eventBus.Publish(ctx, &TagDeletedEvent{
        TenantID: req.TenantID,
        TagName:  req.TagName,
        TagType:  tag.TagType,
    })
    
    return nil
}

// UserService 监听事件
func (s *userService) OnTagDeleted(ctx context.Context, event *TagDeletedEvent) {
    // 从所有用户的 tags 中移除该 tag
    s.userRepo.RemoveTagFromAllUsers(ctx, event.TenantID, event.TagName)
}
```

**优点**：
- ✅ **解耦**：TagService 不依赖其他 Service
- ✅ **可扩展**：可以添加更多监听者
- ✅ **符合领域驱动设计**：事件驱动架构

**缺点**：
- ❌ **复杂度高**：需要事件系统
- ❌ **最终一致性**：不是强一致性
- ❌ **当前未实现**：需要额外开发

**适用场景**：
- 🔮 未来扩展（如果系统需要更复杂的业务逻辑）

---

## 📊 方案对比表

| 方案 | 复杂度 | 性能 | 循环依赖 | 业务规则 | 推荐度 |
|------|--------|------|---------|---------|--------|
| 方案1：直接调用数据库函数 | ⭐ | ⭐⭐⭐ | ✅ 无 | ⚠️ 绕过 | ⭐⭐⭐ |
| 方案2：调用其他 Service | ⭐⭐⭐ | ⭐ | ❌ 有风险 | ✅ 符合 | ❌ |
| 方案3：Repository 层调用 | ⭐⭐ | ⭐⭐⭐ | ✅ 无 | ✅ 部分 | ⭐⭐⭐⭐ |
| 方案4：事件驱动 | ⭐⭐⭐ | ⭐⭐ | ✅ 无 | ✅ 符合 | ⭐⭐ |

---

## 🎯 推荐方案

### 当前阶段：方案3（Repository 层调用数据库函数）

**理由**：
1. ✅ **简单**：只需调用数据库函数，不需要实现复杂的跨 Service 调用
2. ✅ **性能好**：数据库层面批量更新，效率高
3. ✅ **无循环依赖**：不依赖其他 Service
4. ✅ **业务规则验证**：在 Service 层验证（如系统预定义类型不能删除）
5. ✅ **符合数据库设计**：数据库已经提供了完整的清理逻辑

**实现要点**：
```go
// Service 层：业务规则验证
func (s *tagService) DeleteTag(ctx context.Context, req DeleteTagRequest) error {
    // 1. 权限检查
    // 2. 业务规则验证（系统预定义类型不能删除）
    // 3. 调用 Repository
    return s.tagRepo.DeleteTag(ctx, req.TenantID, req.TagName)
}

// Repository 层：调用数据库函数
func (r *postgresTagsRepository) DeleteTag(ctx context.Context, tenantID string, tagName string) error {
    _, err := r.db.ExecContext(ctx, `SELECT drop_tag($1, $2)`, tenantID, tagName)
    return err
}
```

### 未来扩展：方案4（事件驱动）

如果未来需要更复杂的业务逻辑（如事件通知、缓存清理），可以考虑：
1. 在 Repository 层调用数据库函数删除
2. 在 Service 层发布事件
3. 其他 Service 监听事件，执行业务逻辑

---

## 📝 实现建议

### 1. TagService.DeleteTag 实现

```go
// DeleteTagRequest 删除标签请求
type DeleteTagRequest struct {
    TenantID string
    UserRole string
    TagName  string
}

// DeleteTag 删除标签
func (s *tagService) DeleteTag(ctx context.Context, req DeleteTagRequest) error {
    // 1. 参数验证
    if req.TenantID == "" {
        return fmt.Errorf("tenant_id is required")
    }
    if req.TagName == "" {
        return fmt.Errorf("tag_name is required")
    }
    
    // 2. 权限检查
    if !s.hasPermission(ctx, req.UserRole, "tags", "D") {
        return fmt.Errorf("permission denied: cannot delete tag")
    }
    
    // 3. 业务规则验证：查询 tag 信息
    tag, err := s.tagRepo.GetTagByName(ctx, req.TenantID, req.TagName)
    if err != nil {
        if err == sql.ErrNoRows {
            return fmt.Errorf("tag not found: %s", req.TagName)
        }
        return fmt.Errorf("failed to get tag: %w", err)
    }
    
    // 4. 系统预定义类型不能删除
    if tag.TagType == "branch_tag" || tag.TagType == "family_tag" || tag.TagType == "area_tag" {
        return fmt.Errorf("cannot delete system predefined tag type: %s", tag.TagType)
    }
    
    // 5. 调用 Repository（Repository 调用数据库函数 drop_tag）
    // 数据库函数会自动清理所有使用该 tag 的地方
    err = s.tagRepo.DeleteTag(ctx, req.TenantID, req.TagName)
    if err != nil {
        // 数据库函数会检查是否还在使用，如果还在使用会返回错误
        return fmt.Errorf("failed to delete tag: %w", err)
    }
    
    return nil
}
```

### 2. Repository.DeleteTag 实现

```go
// DeleteTag 删除标签（调用数据库函数 drop_tag）
func (r *postgresTagsRepository) DeleteTag(ctx context.Context, tenantID string, tagName string) error {
    if tenantID == "" {
        return fmt.Errorf("tenant_id is required")
    }
    if tagName == "" {
        return fmt.Errorf("tag_name is required")
    }
    
    // 调用数据库函数 drop_tag
    // 该函数会：
    // 1. 检查是否可以删除（系统预定义类型、是否还在使用）
    // 2. 自动清理所有使用该 tag 的地方
    // 3. 删除 tags_catalog 记录
    _, err := r.db.ExecContext(ctx, 
        `SELECT drop_tag($1, $2)`, 
        tenantID, tagName)
    
    if err != nil {
        // 数据库函数会返回详细的错误信息
        return fmt.Errorf("failed to delete tag: %w", err)
    }
    
    return nil
}
```

---

## ✅ 结论

**推荐使用方案3（Repository 层调用数据库函数）**，因为：

1. ✅ **简单**：只需调用数据库函数，不需要实现复杂的跨 Service 调用
2. ✅ **性能好**：数据库层面批量更新，效率高
3. ✅ **无循环依赖**：不依赖其他 Service
4. ✅ **业务规则验证**：在 Service 层验证（如系统预定义类型不能删除）
5. ✅ **符合数据库设计**：数据库已经提供了完整的清理逻辑（`drop_tag` 函数）

**如果未来需要更复杂的业务逻辑**，可以考虑方案4（事件驱动），但当前阶段方案3已经足够。

