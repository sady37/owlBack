# Branch Tag 逻辑说明

## 1. 创建 Tag Name (branch_tag)

### 前端逻辑 (TagList.vue)
- **函数**: `handleCreateBranch()` (行 563-627)
- **触发**: 用户输入 Branch Name 并点击 "Create Branch" 按钮
- **流程**:
  1. 验证输入：检查 `createBranchData.value.trim()` 是否为空
  2. 检查是否存在 `tag_type='branch_tag'` 且 `tag_name='Branch'` 的 tag
  3. 如果不存在，调用 `createTagApi()` 创建：
     - `tag_type: 'branch_tag'`
     - `tag_name: 'Branch'` (固定值)
  4. 刷新 tag 列表 (`fetchTags()`)
  5. 生成 UUID 作为 branch_id
  6. 调用 `addTagObjectsApi()` 添加 branch member：
     - `tag_id`: 刚创建的 Branch tag 的 ID
     - `object_type: 'branch'`
     - `objects: [{ object_id: branchId, object_name: branchName }]`

### 后端逻辑 (admin_tags_handlers.go)
- **API**: `POST /admin/api/v1/tags`
- **处理**: 行 117-172
- **流程**:
  1. 验证 `tag_name` 和 `tag_type`
  2. 调用 PostgreSQL 函数 `upsert_tag_to_catalog(tenant_id, tag_name, tag_type)`
  3. 返回 `tag_id`

### 特点
- **tag_name 固定为 "Branch"**，不能修改
- 所有 branch 名称存储在 `tag_objects.branch` JSONB 中
- 前端显示时会将所有 `branch_tag` 类型的 tag 合并为一行，tag_name 显示为 "Branch"

---

## 2. 创建 Member (branch)

### 前端逻辑 (TagList.vue)
- **函数**: `handleCreateBranch()` (行 607-618)
- **流程**:
  1. 生成 UUID: `crypto.randomUUID()`
  2. 获取 branch 名称: `createBranchData.value.trim()`
  3. 调用 `addTagObjectsApi()`:
     ```typescript
     {
       tag_id: branchTag.tag_id,
       object_type: 'branch',
       objects: [{ object_id: branchId, object_name: branchName }]
     }
     ```

### 后端逻辑 (admin_tags_handlers.go)
- **API**: `POST /admin/api/v1/tags/{tag_id}/objects`
- **处理**: 行 174-226
- **流程**:
  1. 从 URL 路径提取 `tag_id`
  2. 解析 payload: `object_type` 和 `objects` 数组
  3. 遍历 `objects`，对每个 object 调用:
     ```sql
     SELECT update_tag_objects($1::uuid, $2, $3::uuid, $4, 'add')
     ```
  4. 参数: `tag_id`, `object_type`, `object_id`, `object_name`, `'add'`

### 特点
- 每个 branch 有唯一的 UUID (`object_id`)
- branch 名称存储在 `object_name` 字段
- 存储在 `tag_objects.branch` JSONB 中，结构: `{ "<uuid>": "<branch_name>" }`

---

## 3. 更新 Member (branch)

### 前端逻辑
- **不支持直接更新**
- branch_tag 的 member 显示为删除按钮，不支持 checkbox 编辑
- 如需修改，需要删除后重新创建

### 后端逻辑
- **无更新逻辑**
- 只能通过删除后重新添加来实现更新

---

## 4. 删除 Member (branch)

### 前端逻辑 (TagList.vue)
- **函数**: `deleteObjectFromTag()` (行 996-1028)
- **触发**: 点击 branch member 旁边的删除按钮 (×)
- **权限检查**:
  - 只有 Admin 可以删除 branch (`canManageBranch.value`)
  - Manager 和其他角色不能删除
- **流程**:
  1. 权限检查: `if (record.tag_type === 'branch_tag' && !canManageBranch.value)`
  2. 调用 `removeTagObjectsApi()`:
     ```typescript
     {
       tag_id: record.tag_id,
       object_type: objectType,  // 'branch'
       object_ids: [objectId]
     }
     ```

### 后端逻辑 (admin_tags_handlers.go)
- **API**: `DELETE /admin/api/v1/tags/{tag_id}/objects`
- **处理**: 行 231-330
- **流程**:
  1. 从 URL 路径提取 `tag_id`
  2. 解析 payload: `object_type` 和 `object_ids` 数组
  3. 查询 tag 信息: `SELECT tag_name, tag_type FROM tags_catalog WHERE tag_id = $1`
  4. 遍历 `object_ids`，对每个 object_id 调用:
     ```sql
     SELECT update_tag_objects($1::uuid, $2, $3::uuid, '', 'remove')
     ```
  5. 参数: `tag_id`, `object_type`, `object_id`, `''`, `'remove'`
  6. **注意**: branch_tag 删除 member 时，**不会同步更新其他表**（与 family_tag 不同）

### 特点
- 只有 Admin 可以删除
- 删除后不会自动更新 `units.branch_tag` 字段（需要手动处理或通过其他机制）

---

## 5. 删除 Tag Name (branch_tag)

### 前端逻辑 (TagList.vue)
- **函数**: `deleteTagName()` (行 784-823)
- **显示条件**: 
  - `!hasObjects(record)` (没有 member)
  - `record.tag_type === 'user_tag' || record.tag_type === 'family_tag'` (只对 user_tag 和 family_tag 显示删除按钮)
  - `canManageOtherTags` (Manager 或 Admin)
- **branch_tag 特点**:
  - **不显示删除按钮**（因为 tag_name 固定为 "Branch"，且是系统预定义类型）
  - 即使没有 member，也不会显示删除按钮

### 后端逻辑 (admin_tags_handlers.go)
- **API**: `DELETE /admin/api/v1/tags?tenant_id=xxx&tag_name=xxx`
- **处理**: 行 14-43
- **流程**:
  1. 从 query 参数获取 `tag_name`
  2. 验证 `tag_name` 不为空
  3. 调用 PostgreSQL 函数:
     ```sql
     SELECT drop_tag($1::uuid, $2)
     ```
  4. 参数: `tenant_id`, `tag_name`

### 特点
- branch_tag 是系统预定义类型，通常不允许删除
- 前端不提供删除 branch_tag tag_name 的入口
- 即使通过 API 删除，也会影响所有使用该 tag 的资源

---

## 6. 数据合并逻辑 (前端显示)

### 前端逻辑 (TagList.vue)
- **函数**: `sortedDataSource` computed (行 307-423)
- **处理**: 行 327-358
- **流程**:
  1. 将所有 `tag_type='branch_tag'` 的 tag 合并为一行
  2. 查找 `tag_name='Branch'` 的 tag（如果不存在，使用第一个）
  3. 合并所有 branch_tag 的 `tag_objects` 到单个 tag
  4. 固定显示 `tag_name='Branch'`

### 特点
- 所有 branch_tag 类型的 tag 在 UI 中显示为一行
- tag_name 固定显示为 "Branch"
- 所有 branch member 显示在 objects 列中

---

## 7. 权限控制

### 前端权限
- **创建 Branch**: `canManageBranch` (只有 Admin)
- **删除 Branch Member**: `canManageBranch` (只有 Admin)
- **删除 Tag Name**: branch_tag 不支持删除

### 后端权限
- 后端不检查权限，依赖前端控制
- 建议后端也添加权限检查

---

## 8. 数据同步

### 删除 Member 时的同步
- **branch_tag**: **不自动同步**
  - 删除 branch member 后，不会自动更新 `units.branch_tag` 字段
  - 需要手动处理或通过其他机制更新

### 对比其他 tag_type
- **family_tag**: 删除 member 时会同步更新 `residents.family_tag = NULL`
- **user_tag**: 删除 member 时会同步更新 `users.tags` JSONB

---

## 总结

| 操作 | 前端函数 | 后端 API | 权限 | 特点 |
|------|---------|---------|------|------|
| 创建 Tag Name | `handleCreateBranch()` | `POST /admin/api/v1/tags` | Admin | tag_name 固定为 "Branch" |
| 创建 Member | `handleCreateBranch()` | `POST /admin/api/v1/tags/{tag_id}/objects` | Admin | 生成 UUID，存储在 tag_objects.branch |
| 更新 Member | 不支持 | 无 | - | 只能删除后重新创建 |
| 删除 Member | `deleteObjectFromTag()` | `DELETE /admin/api/v1/tags/{tag_id}/objects` | Admin | 不自动同步 units.branch_tag |
| 删除 Tag Name | 不支持 | `DELETE /admin/api/v1/tags` | - | 前端不提供入口，系统预定义类型 |

