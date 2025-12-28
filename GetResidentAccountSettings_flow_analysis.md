# GetResidentAccountSettings 业务流程追踪

## 数据库查询结果

### 1. 查询所有该 contact_id 的记录

**Contact ID**: `f943c983-0a36-479c-bf3c-494bbdbda7be`

查询结果：
```
contact_id: f943c983-0a36-479c-bf3c-494bbdbda7be
resident_id: 5201a246-963a-4048-861d-243ada6bb0ff
resident_account: r2
can_view_status: true
slot: C
contact_email: r1a@demo.com
contact_phone: (空)
phone_hash: NULL
email_hash: b60f36f83ff6404b695168d3c4251aac6082f5c1d1c4fee2a52d07c3771df55b
is_enabled: true
tenant_id: 095c1ab6-5143-47ea-8670-5476158b6cad
```

### 2. 查询所有使用相同 email_hash 的记录

**Email Hash**: `b60f36f83ff6404b695168d3c4251aac6082f5c1d1c4fee2a52d07c3771df55b`

查询结果（2条记录）：
1. **记录1** (r2, can_view_status=true):
   - contact_id: `f943c983-0a36-479c-bf3c-494bbdbda7be`
   - resident_id: `5201a246-963a-4048-861d-243ada6bb0ff`
   - resident_account: `r2`
   - can_view_status: `true`
   - slot: `C`

2. **记录2** (r1, can_view_status=false):
   - contact_id: `d6352f70-8a73-4bec-aeca-c93694ff43af`
   - resident_id: `85ddc499-eea5-4b92-b625-546819a841e7`
   - resident_account: `r1`
   - can_view_status: `false`
   - slot: `A`

## Service 业务流程

### Step 1: 权限检查
```go
if req.CurrentUserID != req.ResidentID {
    return nil, fmt.Errorf("permission denied: can only view own account settings")
```
**结果**: ✅ 通过（假设 `CurrentUserID == ResidentID == f943c983-0a36-479c-bf3c-494bbdbda7be`）

### Step 2: 判断是 resident 还是 contact
```go
if req.CurrentUserRole == "Family" {
    isContact = true
} else {
    isContact = false
}
```
**结果**: ✅ `isContact = true`（假设 `CurrentUserRole == "Family"`）

### Step 3: 获取 phone_hash 和 email_hash
```sql
SELECT phone_hash, email_hash 
FROM resident_contacts 
WHERE tenant_id = $1 AND contact_id::text = $2
```
**结果**:
- `phone_hash`: `NULL` (空)
- `email_hash`: `b60f36f83ff6404b695168d3c4251aac6082f5c1d1c4fee2a52d07c3771df55b`

### Step 4: 选择 accountHash
```go
if len(phoneHash) > 0 {
    accountHash = phoneHash
} else if len(emailHash) > 0 {
    accountHash = emailHash
}
```
**结果**: ✅ `accountHash = email_hash`（因为 `phone_hash` 为空）

### Step 5: 找到第一个可用的 resident_id

**查询 SQL**:
```sql
SELECT r.resident_id::text
FROM resident_contacts rc
JOIN residents r ON r.resident_id = rc.resident_id AND r.tenant_id = rc.tenant_id
WHERE rc.tenant_id = $1
  AND (rc.phone_hash = $2 OR rc.email_hash = $2)
  AND COALESCE(rc.is_enabled,true) = true
  AND EXISTS (
    SELECT 1 FROM residents r2
    WHERE r2.resident_id = rc.resident_id
      AND r2.tenant_id = rc.tenant_id
      AND COALESCE(r2.can_view_status,true) = true
  )
ORDER BY 
  CASE WHEN COALESCE(r.can_view_status,true) = true THEN 0 ELSE 1 END ASC,
  CASE
    WHEN rc.phone_hash = $2 THEN 1
    WHEN rc.email_hash = $2 THEN 2
    ELSE 3
  END ASC
LIMIT 1
```

**参数**:
- `$1` (tenant_id): `095c1ab6-5143-47ea-8670-5476158b6cad`
- `$2` (accountHash): `b60f36f83ff6404b695168d3c4251aac6082f5c1d1c4fee2a52d07c3771df55b` (bytea)

**预期结果**:
- 应该返回 `5201a246-963a-4048-861d-243ada6bb0ff` (r2, can_view_status=true)

**问题**: 
- 如果查询返回 0 行，可能是 `bytea` 比较的问题
- 需要检查 `rc.email_hash = $2` 的比较是否正确

### Step 6: 查询 resident_account

**查询 SQL**:
```sql
SELECT resident_account
FROM residents
WHERE tenant_id = $1 AND resident_id::text = $2 
  AND resident_account IS NOT NULL AND resident_account != ''
```

**参数**:
- `$1` (tenant_id): `095c1ab6-5143-47ea-8670-5476158b6cad`
- `$2` (resident_id): `5201a246-963a-4048-861d-243ada6bb0ff`

**预期结果**: ✅ `r2`

### Step 7: 查询 Contact 的其他信息

**查询 SQL**:
```sql
SELECT 
  COALESCE(rc.contact_email, '') as contact_email,
  COALESCE(rc.contact_phone, '') as contact_phone,
  COALESCE(rc.contact_first_name || ' ' || rc.contact_last_name, '') as nickname
FROM resident_contacts rc
WHERE rc.tenant_id = $1 AND rc.contact_id::text = $2
```

**参数**:
- `$1` (tenant_id): `095c1ab6-5143-47ea-8670-5476158b6cad`
- `$2` (contact_id): `f943c983-0a36-479c-bf3c-494bbdbda7be`

**预期结果**:
- `contact_email`: `r1a@demo.com`
- `contact_phone`: (空)
- `nickname`: (取决于 `contact_first_name` 和 `contact_last_name`)

## 问题分析

### 可能的问题点

1. **Step 5 查询返回 0 行**:
   - 可能原因：`bytea` 比较问题
   - 检查：`rc.email_hash = $2` 的比较是否正确
   - 解决方案：确保 `accountHash` 是 `[]byte` 类型，并且与数据库中的 `email_hash` 类型匹配

2. **Step 7 查询返回 0 行**:
   - 可能原因：`contact_id` 或 `tenant_id` 不匹配
   - 检查：确认 `req.ResidentID` 和 `req.TenantID` 的值是否正确

## 建议的修复方案

1. **检查 `bytea` 比较**:
   - 确保 `accountHash` 是 `[]byte` 类型
   - 确保数据库中的 `email_hash` 和 `phone_hash` 是 `bytea` 类型

2. **添加调试日志**:
   - 在 Step 5 查询前后添加日志，输出 `accountHash` 的值和查询结果
   - 在 Step 7 查询前后添加日志，输出 `req.ResidentID` 和 `req.TenantID` 的值

3. **简化查询逻辑**:
   - 如果 `EXISTS` 子查询有问题，可以简化为直接检查 `COALESCE(r.can_view_status,true) = true`

