# resident_service.go 问题检查报告

## 发现的问题

### 1. ❌ 第 91 行：`getUserBranchTag` 方法使用了不存在的字段

**问题**：
```go
func (s *residentService) getUserBranchTag(ctx context.Context, tenantID, userID string) (string, error) {
    var branchTag sql.NullString
    err := s.db.QueryRowContext(ctx,
        `SELECT branch_tag FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
        tenantID, userID,
    ).Scan(&branchTag)
    // ...
}
```

**原因**：
- `users` 表已经没有 `branch_tag` 字段
- 现在应该通过 `user_branches` 表 JOIN `branches` 表来获取 `branch_name`
- 或者通过 `users.branch_id` JOIN `branches` 表来获取 `branch_name`

**修复方案**：
需要修改为通过 `user_branches` 表或 `branches` 表 JOIN 来获取 `branch_name`。

---

### 2. ❌ 第 1925 行：引用了已删除的 `tags_catalog` 表

**问题**：
```go
SELECT 1 FROM tags_catalog tc
WHERE tc.tenant_id = $1
  AND tc.tag_type = 'user_tag'
  AND tc.tag_name = user_tag_name
  AND tc.tag_id::text = ANY(
    SELECT jsonb_array_elements_text(rc.group_list)::text
  )
```

**原因**：
- `tags_catalog` 表已经被删除
- 现在 tags 直接存储在 `users.user_tags` JSONB 字段中

**修复方案**：
需要修改逻辑，直接比较 `users.user_tags` 中的 tag 值，而不需要通过 `tags_catalog` 表映射。

---

### 3. ❌ 第 2346 行：使用了错误的字段名 `userList`

**问题**：
```go
AND (rc.userList::text LIKE $3 OR rc.userList::text LIKE $4)
```

**原因**：
- 数据库字段名是 `user_list`（下划线命名），不是 `userList`（驼峰命名）

**修复方案**：
改为 `rc.user_list::text`

---

## 需要修复的位置

1. **第 87-104 行**：`getUserBranchTag` 方法
   - 需要改为通过 `user_branches` 表或 `branches` 表 JOIN 来获取 `branch_name`

2. **第 1925 行**：`tags_catalog` 表引用
   - 需要移除对 `tags_catalog` 表的依赖，直接使用 `users.user_tags` 进行比较

3. **第 2346 行**：字段名 `userList`
   - 改为 `user_list`

