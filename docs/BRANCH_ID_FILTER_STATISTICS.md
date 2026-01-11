# Branch_id 过滤统计报告

## 统计范围
统计 `owlBack/wisefido-data` 目录下所有涉及 `branch_id` 过滤的代码位置。

## 统计结果

### 总体统计
- **总匹配数**：461 处（包含所有 branch_id/BranchID 相关的 WHERE、filter、AND、= 操作）
- **SQL 查询中的过滤**：119 处（WHERE/AND/filter/IN 条件）
- **涉及的文件数**：39 个文件

### 按目录分类

#### Repository 层（数据访问层）
- **匹配数**：84 处
- **涉及文件**：11 个文件
  - `postgres_user_branches.go`: 10 处
  - `postgres_users.go`: 15 处
  - `postgres_units.go`: 41 处
  - `postgres_residents.go`: 4 处
  - `postgres_alarm_events.go`: 2 处
  - `postgres_branches.go`: 13 处
  - `postgres_card.go`: 1 处
  - `units_repo.go`: 4 处
  - `cards_repository.go`: 1 处
  - `memory_units.go`: 1 处
  - `user_branches_repo.go`: 1 处

#### Service 层（业务逻辑层）
- **匹配数**：41 处
- **涉及文件**：9 个文件
  - `unit_service.go`: 10 处
  - `user_service.go`: 5 处
  - `resident_service.go`: 11 处
  - `alarm_event_service.go`: 3 处
  - `card_service_vital_focus.go`: 2 处
  - `auth_service.go`: 2 处
  - `branch_service.go`: 3 处
  - `card_service.go`: 2 处
  - `resident_service_test.go`: 3 处

#### HTTP Handler 层（API 层）
- **匹配数**：约 10 处
- **涉及文件**：多个文件
  - `user_handler.go`: 10 处
  - `unit_handler.go`: 7 处
  - `resident_handler.go`: 3 处
  - `alarm_event_handler.go`: 1 处
  - `permission_utils.go`: 2 处

### 主要过滤场景

1. **用户权限验证**：
   - 从 `user_branches` 表查询用户所属的 branch_id
   - 验证用户是否有权限访问指定的 branch

2. **数据查询过滤**：
   - Units、Rooms、Beds 按 branch_id 过滤
   - Residents 按 branch_id 过滤
   - Devices 按 branch_id 过滤（通过 unit 关联）
   - Alarm Events 按 branch_id 过滤

3. **权限控制**：
   - `branch_only` 权限：只允许访问用户所属 branch 的数据
   - `assigned_only` 权限：只允许访问分配给用户的数据
   - `all` 权限：可以访问所有数据（不进行 branch 过滤）

### 关键函数

1. **getUserBranchID**：获取用户的第一个 branch_id
2. **getUserBranchIDs**：获取用户的所有 branch_id 列表
3. **getBranchIDForPermission**：统一获取 branch_id 用于权限验证
4. **verifyUnitPermission**：验证 unit 的权限（tenant + branch）
5. **verifyRoomPermission**：验证 room 的权限（通过 unit）
6. **verifyBedPermission**：验证 bed 的权限（通过 room -> unit）

### 特殊处理

- ***(ALL) 选项**：当 `branch_ids` 包含 `'*'` 时，表示用户可以访问所有分支
  - 不创建 `user_branches` 记录
  - `users.branch_id` 设置为 NULL
  - 权限验证时跳过 branch 检查

### 统计命令

```bash
# 统计所有 branch_id/BranchID 相关的 WHERE/filter/AND/IN 操作
cd owlBack/wisefido-data
grep -r "branch_id\|BranchID" --include="*.go" | grep -E "WHERE|AND|filter|IN" | wc -l
# 结果：119 处

# 统计所有 branch_id/BranchID 相关的代码行
grep -r "branch_id\|BranchID" --include="*.go" | wc -l
# 结果：1315 处（包含所有引用）
```

## 总结

在 `owlBack/wisefido-data` 中，共有 **119 处**涉及 `branch_id` 过滤的 SQL 查询或业务逻辑，分布在：
- **Repository 层**：84 处（主要在 `postgres_units.go` 和 `postgres_users.go`）
- **Service 层**：41 处（主要在 `unit_service.go` 和 `resident_service.go`）
- **HTTP Handler 层**：约 10 处

这些过滤主要用于：
1. 权限验证（验证用户是否有权限访问指定的 branch）
2. 数据查询过滤（按 branch_id 过滤查询结果）
3. 多租户数据隔离（确保用户只能访问自己 branch 的数据）
