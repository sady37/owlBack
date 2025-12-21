# Family Tag 逻辑说明

## 1. 创建 Tag Name (family_tag)

### 前端逻辑 (TagList.vue)
- **函数**: `handleCreateFamily()` (行 701-732)
- **触发**: 用户输入 Family Name 并点击 "Create Familytag" 按钮
- **流程**:
  1. 验证输入：检查 `createFamilyData.value.trim()` 是否为空
  2. 调用 `createTagApi()` 创建 tag：
     - `tag_type: 'family_tag'`
     - `tag_name: createFamilyData.value.trim()` (用户输入的 family 名称)
  3. 刷新 tag 列表 (`fetchTags()`)

### 后端逻辑 (admin_tags_handlers.go)
- **API**: `POST /admin/api/v1/tags`
- **处理**: 行 117-172
- **流程**:
  1. 验证 `tag_name` 和 `tag_type`
  2. 调用 PostgreSQL 函数 `upsert_tag_to_catalog(tenant_id, tag_name, tag_type)`
  3. 返回 `tag_id`

### 特点
- **tag_name 是用户输入的 family 名称**（如 "F002"）
- 每个 family 有独立的 tag_name
- 前端显示时，每个 family_tag 显示为独立的一行

---

## 2. 创建 Member (resident)

### 前端逻辑 (TagList.vue)
- **重要**: Tag Management 页面**不支持添加** member
- **只能删除**: 页面只显示**已经存在的** member（从 `record.tag_objects` 中读取），通过 checkbox 控制删除
- **添加方式**: Member 的添加应该通过其他方式完成：
  - 当 resident 的 `family_tag` 字段被设置时，可能通过数据库触发器或后端逻辑自动添加到 family_tag 的 `tag_objects`
  - 或者通过其他 API/页面手动添加

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
- **Tag Management 页面不支持添加 member**
- Member 的添加需要通过其他机制完成（可能是数据库触发器、后端逻辑或其他页面）
- resident 的 `family_tag` 字段和 family_tag 的 `tag_objects` 的同步关系需要进一步确认

---

## 3. 更新 Member (resident)

### 前端逻辑 (TagList.vue)
- **不支持更新**: Tag Management 页面**不支持更新** member
- **只支持删除**: 只能通过取消勾选 checkbox 来删除 member
- **函数**: `handleObjectCheckChange()` (行 828-900)
- **触发**: 用户取消勾选 resident checkbox
- **流程**:
  1. 用户取消勾选 checkbox
  2. `handleObjectCheckChange()` 被调用
  3. 检测到取消勾选，将 object 添加到 `objectsToRemove[tagId]` 列表
  4. 如果重新勾选，从 `objectsToRemove[tagId]` 列表中移除（取消删除操作）
  5. 用户点击 "Save" 后，`handleSaveAll()` 处理所有删除操作

### 后端逻辑
- **无更新逻辑**
- 只能删除，不能更新

### 特点
- **只能删除**，不能添加或更新
- 通过 checkbox 控制 member 的删除
- 支持批量操作（一次可以取消勾选多个 resident）

---

## 4. 删除 Member (resident)

### 前端逻辑 (TagList.vue)
- **函数**: `handleObjectCheckChange()` (行 828-900) + `handleSaveAll()` (行 925-1000)
- **触发**: 用户取消勾选 resident checkbox，然后点击 "Save"
- **流程**:
  1. 用户取消勾选 resident checkbox
  2. `handleObjectCheckChange()` 检测到取消勾选，将 object 添加到 `objectsToRemove[tagId]` 列表
  3. 用户点击 "Save" 按钮
  4. `handleSaveAll()` 遍历 `objectsToRemove`，调用 `removeTagObjectsApi()`:
     ```typescript
     {
       tag_id: tagId,
       object_type: 'resident',
       object_ids: [residentId]
     }
     ```

### 后端逻辑 (admin_tags_handlers.go)
- **API**: `DELETE /admin/api/v1/tags/{tag_id}/objects`
- **处理**: 行 231-330 (object_ids 格式) 和 335-410 (objects 格式)
- **流程**:
  1. 从 URL 路径提取 `tag_id`
  2. 解析 payload: `object_type` 和 `object_ids` 数组
  3. 查询 tag 信息: `SELECT tag_name, tag_type FROM tags_catalog WHERE tag_id = $1`
  4. 遍历 `object_ids`，对每个 object_id 调用:
     ```sql
     SELECT update_tag_objects($1::uuid, $2, $3::uuid, '', 'remove')
     ```
  5. **重要**: 如果 `objectType == "resident" && tagType == "family_tag"`，同步更新 `residents.family_tag`:
     ```sql
     UPDATE residents 
     SET family_tag = NULL
     WHERE resident_id = $1::uuid
       AND family_tag = $2
     ```
  6. 参数: `objectID`, `tagName`

### 特点
- **自动同步**: 删除 member 时，**会自动清空** resident 的 `family_tag` 字段
- 确保 `residents.family_tag` 和 family_tag 的 `tag_objects` 保持一致
- 这是与 branch_tag 的主要区别（branch_tag 不自动同步）

---

## 5. 删除 Tag Name (family_tag)

### 前端逻辑 (TagList.vue)
- **函数**: `deleteTagName()` (行 784-823)
- **显示条件**: 
  - `!hasObjects(record)` (没有 member)
  - `record.tag_type === 'user_tag' || record.tag_type === 'family_tag'` (只对 user_tag 和 family_tag 显示删除按钮)
  - `canManageOtherTags` (Manager 或 Admin)
- **流程**:
  1. 检查权限: `canManageOtherTags` (Manager 或 Admin)
  2. 检查是否有 objects: `hasObjects(record)`
  3. 验证 `tag_name` 不为空
  4. 调用 `deleteTagApi()`:
     ```typescript
     {
       tenant_id: tenantId,
       tag_name: record.tag_name.trim()
     }
     ```

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
- 只有没有 member 的 family_tag 才能删除
- 删除 tag_name 后，所有关联的 member 也会被删除
- **注意**: 删除 tag_name 时，**不会自动清空** resident 的 `family_tag` 字段（需要通过删除 member 来触发）

---

## 6. 数据同步

### 删除 Member 时的同步
- **family_tag**: **自动同步**
  - 删除 resident member 时，会自动执行:
    ```sql
    UPDATE residents 
    SET family_tag = NULL
    WHERE resident_id = $1::uuid
      AND family_tag = $2
    ```
  - 确保 `residents.family_tag` 和 family_tag 的 `tag_objects` 保持一致

### 创建 Member 时的同步
- **family_tag**: **通过数据库触发器自动同步**
  - 当在 Resident Profile 中设置 `family_tag` 时，数据库触发器 `trigger_sync_family_tag` 会自动执行
  - 触发器函数 `sync_family_tag_to_catalog()` 会：
    1. 从旧 family_tag 的 `tag_objects` 中删除该 resident（如果存在）
    2. 确保新 family_tag 存在于 `tags_catalog` 表
    3. 将 resident 添加到新 family_tag 的 `tag_objects` 中
  - **实现位置**: `owlRD/db/22_tags_catalog.sql` 行 477-519

### 对比其他 tag_type
- **branch_tag**: 删除 member 时**不自动同步** `units.branch_tag`
- **user_tag**: 删除 member 时会同步更新 `users.tags` JSONB
- **family_tag**: 删除 member 时会同步清空 `residents.family_tag`

---

## 7. 权限控制

### 前端权限
- **创建 Family Tag**: `canManageOtherTags` (Manager 或 Admin)
- **添加/删除 Member**: `canManageOtherTags` (Manager 或 Admin)
- **删除 Tag Name**: `canManageOtherTags` (Manager 或 Admin)

### 后端权限
- 后端不检查权限，依赖前端控制
- 建议后端也添加权限检查

---

## 8. 数据流向

### 从 Resident Profile 设置 family_tag
1. 用户在 Resident Profile 页面设置 `family_tag = "F002"`
2. 更新 `residents.family_tag` 字段
3. **数据库触发器自动同步**: 通过 `trigger_sync_family_tag` 触发器自动执行：
   - 调用 `sync_family_tag_to_catalog()` 函数
   - 如果旧 `family_tag` 存在，从旧 family_tag 的 `tag_objects` 中删除该 resident
   - 确保新 family_tag 存在于 `tags_catalog` 表（调用 `upsert_tag_to_catalog`）
   - 将 resident 添加到新 family_tag 的 `tag_objects` 中（调用 `update_tag_objects`）
4. **实现位置**: `owlRD/db/22_tags_catalog.sql` 行 477-519

### 从 Resident Profile 删除 family_tag
1. 用户在 Resident Profile 页面删除 `family_tag`（设置为空或 NULL）
2. 更新 `residents.family_tag = NULL`
3. **问题**: 触发器有 `WHEN (NEW.family_tag IS NOT NULL)` 条件，当设置为 NULL 时**触发器不会执行**
4. **函数逻辑支持删除**: `sync_family_tag_to_catalog()` 函数中有删除逻辑（行 484），但由于触发器条件限制，不会执行
5. **需要修复**: 触发器应该移除 `WHEN` 条件或修改为 `WHEN (OLD.family_tag IS NOT NULL OR NEW.family_tag IS NOT NULL)`，以支持删除操作

### 从 Tags Management 删除 resident
1. 用户在 Tags Management 页面取消勾选 resident checkbox
2. 点击 "Save"
3. 从 family_tag 的 `tag_objects` 中删除
4. **自动清空** resident 的 `family_tag` 字段

### 数据同步问题

#### 当前状态
- **从 Tags Management 删除**: ✅ 自动同步（删除 member 时自动清空 resident.family_tag）
- **从 Resident Profile 设置**: ❌ 不同步（设置 family_tag 时不会自动添加到 tag_objects）
- **从 Resident Profile 删除**: ❌ 不同步（删除 family_tag 时不会自动从 tag_objects 中删除）

#### 已实现的逻辑（数据库触发器）
1. **当在 Resident Profile 中设置 family_tag 时**:
   - ✅ **已实现**: 数据库触发器 `trigger_sync_family_tag` 自动执行
   - 触发器函数 `sync_family_tag_to_catalog()` 会：
     - 从旧 family_tag 的 `tag_objects` 中删除该 resident（如果存在）
     - 确保新 family_tag 存在于 `tags_catalog` 表（调用 `upsert_tag_to_catalog`）
     - 将 resident 添加到新 family_tag 的 `tag_objects` 中（调用 `update_tag_objects`）
   - **实现位置**: `owlRD/db/22_tags_catalog.sql` 行 477-519

#### 缺失的逻辑（触发器 Bug）
2. **当在 Resident Profile 中删除 family_tag 时**:
   - ❌ **未实现**: 触发器有 `WHEN (NEW.family_tag IS NOT NULL)` 条件
   - 当 `family_tag` 被设置为 NULL 时，触发器**不会执行**
   - 虽然函数 `sync_family_tag_to_catalog()` 中有删除逻辑（行 484），但由于触发器条件限制，不会执行
   - **需要修复**: 触发器应该移除 `WHEN` 条件或修改为 `WHEN (OLD.family_tag IS NOT NULL OR NEW.family_tag IS NOT NULL)`

### 添加 Member 的方式
- **Tag Management 页面不支持添加 member**
- Member 的添加可能通过以下方式完成：
  - 数据库触发器：当 resident 的 `family_tag` 字段被设置时自动添加（需要实现）
  - 后端逻辑：在更新 resident 的 `family_tag` 时同步更新 tag_objects（当前缺失）
  - 其他 API/页面：可能有专门的接口或页面用于添加 member

---

## 总结

| 操作 | 前端函数 | 后端 API | 权限 | 特点 |
|------|---------|---------|------|------|
| 创建 Tag Name | `handleCreateFamily()` | `POST /admin/api/v1/tags` | Manager/Admin | tag_name 是用户输入的 family 名称 |
| 创建 Member | **数据库触发器** | `PUT /residents/:id` (设置 family_tag) | Manager/Admin | ✅ 通过数据库触发器自动同步到 tag_objects |
| 更新 Member | **数据库触发器** | `PUT /residents/:id` (修改 family_tag) | Manager/Admin | ✅ 通过数据库触发器自动同步（删除旧，添加新） |
| 删除 Member (Tags Management) | `handleObjectCheckChange()` + `handleSaveAll()` | `DELETE /admin/api/v1/tags/{tag_id}/objects` | Manager/Admin | **自动清空** resident.family_tag |
| 删除 Member (Resident Profile) | **数据库触发器** | `PUT /residents/:id` (删除 family_tag) | Manager/Admin | ❌ **Bug**: 触发器条件限制，不会执行删除逻辑 |
| 删除 Tag Name | `deleteTagName()` | `DELETE /admin/api/v1/tags` | Manager/Admin | 只有没有 member 的才能删除 |

### 关键区别

1. **与 branch_tag 的区别**:
   - branch_tag: tag_name 固定为 "Branch"，所有 branch 存储在 `tag_objects.branch` 中
   - family_tag: 每个 family 有独立的 tag_name，member 存储在 `tag_objects.resident` 中

2. **数据同步**:
   - 删除 member 时，family_tag **会自动清空** resident 的 `family_tag` 字段
   - 添加 member 时，family_tag **不会自动**更新 resident 的 `family_tag` 字段

3. **数据同步状态**:
   - **从 Tags Management 删除**: ✅ 自动同步（删除 member 时自动清空 resident.family_tag）
   - **从 Resident Profile 设置**: ✅ 自动同步（通过数据库触发器自动添加到 tag_objects）
   - **从 Resident Profile 删除**: ❌ **Bug**（触发器条件限制，不会执行删除逻辑）
   - **需要修复**: 触发器 `trigger_sync_family_tag` 的 `WHEN` 条件需要修改

4. **Tag Management 页面限制**:
   - **不支持添加 member**: 只能删除已存在的 member
   - **不支持更新 member**: 只能通过删除后重新添加（但页面不支持添加）
   - Member 的添加需要通过其他机制完成（可能是数据库触发器、后端逻辑或其他页面）

