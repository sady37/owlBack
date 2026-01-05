# Caregiver (c1) 获取 Card 的完整流程

## 1. 登录阶段

### 1.1 用户登录
- **用户**: c1 (caregiver)
- **账户**: c1
- **密码**: [用户输入]

### 1.2 登录 API 处理
**Endpoint**: `POST /auth/api/v1/login`

**流程**:
1. 前端发送登录请求（包含 `userType: "staff"`）
2. `AuthHandler.Login` 处理请求
3. `AuthService.Login` 验证账户和密码
4. 查询 `users` 表获取用户信息：
   - `user_id`: `c3b9bde2-4be4-4671-af81-f79694486932`
   - `user_account`: `c1`
   - `tenant_id`: `bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c`
   - `role`: `Caregiver`

### 1.3 登录响应
```json
{
  "accessToken": "...",
  "refreshToken": "...",
  "userId": "c3b9bde2-4be4-4671-af81-f79694486932",  // user_id
  "user_account": "c1",
  "userType": "staff",
  "role": "Caregiver",
  "tenant_id": "bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c",
  ...
}
```

### 1.4 前端存储
前端将以下信息存储到 `userStore`:
- `userId`: `c3b9bde2-4be4-4671-af81-f79694486932`
- `userType`: `staff`
- `role`: `Caregiver`
- `tenant_id`: `bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c`

---

## 2. 获取 Card 列表阶段

### 2.1 前端发送请求
**Endpoint**: `GET /admin/api/v1/card-overview`

**Headers** (由前端 axios interceptor 自动添加):
```
X-User-Id: c3b9bde2-4be4-4671-af81-f79694486932
X-User-Type: staff
X-User-Role: Caregiver
X-Tenant-Id: bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c
Authorization: Bearer <accessToken>
```

### 2.2 Handler 层处理
**文件**: `card_overview_handler.go`

**流程**:
1. `CardOverviewHandler.GetCardOverview` 接收请求
2. 从 Header 提取:
   - `tenantID`: `bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c`
   - `currentUserID`: `c3b9bde2-4be4-4671-af81-f79694486932`
   - `currentUserType`: `staff`
   - `currentUserRole`: `Caregiver`
3. 记录调试日志
4. 构建 `GetCardOverviewRequest` 传递给 Service 层

---

## 3. Service 层处理
**文件**: `card_service.go`

### 3.1 权限检查
**函数**: `getResourcePermission(ctx, "Caregiver", "cards", "R")`

**SQL 查询**:
```sql
SELECT permission_scope
FROM role_permissions
WHERE tenant_id = '00000000-0000-0000-0000-000000000001'  -- System tenant
  AND role_code = 'Caregiver'
  AND resource_type = 'cards'
  AND permission_type = 'R'
```

**结果**: `permission_scope = 'S'` (AssignedOnly)

**转换**:
- `AssignedOnly = true`
- `BranchOnly = false`

### 3.2 用户验证
**函数**: `GetCardOverview` → 验证 user_id

**SQL 查询**:
```sql
SELECT EXISTS(
  SELECT 1 FROM users u
  WHERE u.tenant_id = 'bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c'
    AND u.user_id::text = 'c3b9bde2-4be4-4671-af81-f79694486932'
)
```

**结果**: `true` (用户存在且属于当前 tenant)

### 3.3 设置权限过滤
```go
repoReq.PermissionFilter = &repository.PermissionFilter{
  AssignedOnly: true,
  UserIDForAssignment: "c3b9bde2-4be4-4671-af81-f79694486932",
}
```

### 3.4 调用 Repository 层
调用 `cardsRepo.ListCards(ctx, repoReq)`

---

## 4. Repository 层处理
**文件**: `cards_repository.go`

### 4.1 构建 SQL 查询

#### 4.1.1 SELECT 子句
```sql
SELECT 
  c.card_id::text,
  c.tenant_id::text,
  c.card_type,
  c.bed_id::text,
  c.unit_id::text,
  c.card_name,
  c.card_address,
  c.resident_id::text,
  c.devices,
  c.residents,
  ...
FROM cards c
LEFT JOIN units u ON c.unit_id = u.unit_id
LEFT JOIN branches br ON u.branch_id = br.branch_id
LEFT JOIN buildings bld ON u.building_id = bld.building_id
```

#### 4.1.2 JOIN resident_caregivers 表
```sql
LEFT JOIN resident_caregivers rc ON (
  rc.tenant_id = $1  -- 'bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c'
  AND (
    -- ActiveBed 卡片：直接匹配 resident_id
    (c.card_type = 'ActiveBed' AND rc.resident_id = c.resident_id)
    OR
    -- Unit 卡片（数据库中使用 'Location'）：检查 residents JSONB 数组
    (c.card_type = 'Location' 
      AND (
        -- 第一个住户
        rc.resident_id::text = (c.residents->0->>'resident_id')::text
        OR
        -- 第二个住户（如果存在）
        (jsonb_array_length(c.residents) >= 2 
          AND rc.resident_id::text = (c.residents->1->>'resident_id')::text)
      )
    )
  )
  AND (
    -- 检查 user_list JSONB 是否包含 userID
    rc.user_list::text LIKE '%"c3b9bde2-4be4-4671-af81-f79694486932"%'
    OR
    -- 检查 group_list JSONB 是否匹配用户的 tags
    EXISTS (
      SELECT 1 FROM users u2
      WHERE u2.tenant_id = $1
        AND u2.user_id::text = $2
        AND u2.user_tags ?| (
          SELECT ARRAY(SELECT jsonb_array_elements_text(rc.group_list))
        )
    )
  )
)
```

#### 4.1.3 WHERE 子句
```sql
WHERE c.tenant_id = $3  -- 'bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c'
  AND rc.resident_id IS NOT NULL  -- 确保 JOIN 成功（即该卡片关联的住户被分配给当前用户）
  AND (
    -- ActiveBed 卡片：已通过 JOIN 过滤
    c.card_type = 'ActiveBed'
    OR
    -- Unit 卡片：检查权限
    (c.card_type = 'Location' 
      AND (u.is_public = FALSE AND u.is_shared_unit = FALSE)
      AND (
        -- 第一个住户（不需要额外检查）
        rc.resident_id::text = (c.residents->0->>'resident_id')::text
        OR
        -- 第二个住户（需要检查 is_access_enabled）
        (
          jsonb_array_length(c.residents) >= 2
          AND rc.resident_id::text = (c.residents->1->>'resident_id')::text
          AND EXISTS (
            SELECT 1 FROM residents r
            WHERE r.tenant_id = c.tenant_id
              AND r.resident_id = rc.resident_id
              AND r.is_access_enabled = TRUE
          )
        )
      )
    )
  )
```

### 4.2 实际数据匹配

#### 4.2.1 resident_caregivers 表数据
```sql
SELECT * FROM resident_caregivers 
WHERE user_list::text LIKE '%"c3b9bde2-4be4-4671-af81-f79694486932"%'
```

**结果**:
```
resident_id: b266ec84-c626-41ad-8bdd-3a72b4194fc0
user_list: ["c3b9bde2-4be4-4671-af81-f79694486932", "029f6ad1-aaf5-4904-b430-e8e76e727718"]
group_list: []
```

#### 4.2.2 匹配的 Card
- **ActiveBed 卡片**: `c.resident_id = 'b266ec84-c626-41ad-8bdd-3a72b4194fc0'`
- **Unit 卡片**: `c.residents->0->>'resident_id' = 'b266ec84-c626-41ad-8bdd-3a72b4194fc0'` 或 `c.residents->1->>'resident_id' = 'b266ec84-c626-41ad-8bdd-3a72b4194fc0'`

### 4.3 返回结果
返回所有匹配的 Card 列表（只包含 c1 被分配的住户相关的卡片）

---

## 5. 数据聚合
**文件**: `card_service.go` → `aggregateCardData`

### 5.1 聚合设备信息
- 从 `devices` 表批量查询设备信息
- 填充 `CardDevice` 结构（device_type, status, serial_number 等）

### 5.2 聚合住户信息
- 从 `residents` 表批量查询住户信息
- 填充 `CardResident` 结构（nickname, service_level 等）

### 5.3 计算权限字段
- `ResidentAccess`: 是否允许住户访问
- `CaregiverCount`: 护理人员数量
- `CaregiverGroups`: 护理人员组列表

---

## 6. 返回响应

### 6.1 Service 层返回
```go
&GetCardOverviewResponse{
  Items: []*domain.CardOverviewItem{...},  // 只包含 c1 被分配的住户相关的卡片
  Total: len(items),
}
```

### 6.2 Handler 层返回
```json
{
  "code": 200,
  "result": {
    "items": [
      {
        "card_id": "...",
        "card_name": "...",
        "card_address": "...",
        "devices": [...],
        "residents": [...],
        ...
      },
      ...
    ],
    "pagination": {
      "total": 1,
      "page": 1,
      "size": 10,
      ...
    }
  }
}
```

---

## 7. 安全保证

### 7.1 Service 层验证
- ✅ 验证 `user_id` 是否存在且属于当前 `tenant_id`
- ✅ 如果验证失败，直接返回空结果，不继续执行

### 7.2 Repository 层过滤
- ✅ JOIN 条件中检查 `rc.tenant_id = $tenantIDParam`
- ✅ JOIN 条件中检查 `u2.tenant_id = $tenantIDParam` (EXISTS 子查询)
- ✅ WHERE 子句中检查 `c.tenant_id = $argIdx`
- ✅ WHERE 子句中检查 `rc.resident_id IS NOT NULL` (确保 JOIN 成功)

### 7.3 双重保护
- **第一道防线**: Service 层验证 user_id
- **第二道防线**: Repository 层 SQL JOIN 过滤

---

## 8. 关键数据表

### 8.1 users 表
```
user_id: c3b9bde2-4be4-4671-af81-f79694486932
user_account: c1
tenant_id: bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c
role: Caregiver
```

### 8.2 role_permissions 表
```
role_code: Caregiver
resource_type: cards
permission_type: R
permission_scope: S  (AssignedOnly)
```

### 8.3 resident_caregivers 表
```
resident_id: b266ec84-c626-41ad-8bdd-3a72b4194fc0
user_list: ["c3b9bde2-4be4-4671-af81-f79694486932", ...]
group_list: []
```

### 8.4 cards 表
- ActiveBed 卡片: `resident_id = 'b266ec84-c626-41ad-8bdd-3a72b4194fc0'`
- Unit 卡片: `residents` JSONB 包含 `resident_id = 'b266ec84-c626-41ad-8bdd-3a72b4194fc0'`

---

## 9. 总结

1. **登录**: c1 登录，获取 `user_id`, `userType`, `role`
2. **权限检查**: 查询 `role_permissions` 表，确认 `AssignedOnly = true`
3. **用户验证**: 验证 `user_id` 存在且属于当前 `tenant_id`
4. **SQL JOIN**: 直接 JOIN `resident_caregivers` 表，匹配 `user_list` 或 `group_list`
5. **结果过滤**: 只返回 c1 被分配的住户相关的卡片
6. **数据聚合**: 聚合设备和住户信息
7. **返回响应**: 返回过滤后的卡片列表

**关键点**:
- ✅ Service 层和 Repository 层双重验证
- ✅ 直接 JOIN `resident_caregivers` 表，性能更好
- ✅ 支持 `user_list` 和 `group_list` 两种分配方式
- ✅ 支持 ActiveBed 和 Unit 两种卡片类型

