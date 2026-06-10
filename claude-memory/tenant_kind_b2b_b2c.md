---
name: tenants.kind B2B/B2C + Family 权限分流
description: tenants 表加 kind 字段控制 family role 行为；B2C 时 family 即 Manager(自家 scope, 不能创建 user)，signup 自动建同名 resident_account 并绑定
type: project
originSessionId: 45c27d41-bfc6-45c1-bff5-228bbb490b7e
---
## DB schema 改动（2026-05-10 已落）

```sql
ALTER TABLE tenants ADD COLUMN kind VARCHAR(10) NOT NULL DEFAULT 'B2B'
    CHECK (kind IN ('B2B','B2C'));
UPDATE tenants SET kind='B2C' WHERE tenant_slot = 4;  -- wisefido
```

源文件：`owlRD/dbv2/10_tenants.sql` + `90_seed_tenants.sql` 同步。

## 业务流程分流

### B2B + Family
- tenant_admin 在 admin 后台**创建 resident**、绑定 family user（role=Family）
- family 登录后**仅查看**：alarm push、resident 状态（无 PHI / 无写权限）
- family role_permissions 维持现状（alarms.read, residents.read.basic, reports.read）

### B2C + Family
- 用户从 wisefido 设备开机 → 扫码 → app 自助 signup
- signup 完成后：
  - 创建 family user（role=Family，tenant=wisefido kind=B2C）
  - **自动**用同名创建 resident + resident_phi 行
  - **自动** insert resident_caregivers 行（resident_id=新建, caregiver_id=family.hoa, role='primary'）
- Family 权限提升到 Manager 等价：
  - 可改 resident 信息（PHI 编辑）
  - 可绑定 device 到自家 resident
  - 可改 alarm threshold
  - 可处理 alarm（true/false 判定）
- **限制**：
  - `relegation='assigned_only'` 强制（只能看自家 resident scope）
  - 不能创建 user（无 `users.create`）

## BE 实施待办（下次会话）

1. `findActiveUser` 或权限检查处加 tenant.kind 读取
2. 权限 elevation 逻辑：role='Family' && tenant.kind='B2C' → 加 Manager 权限子集，去 users.create
3. Signup API：`POST /auth/api/v2/signup`
   - body: `{user_account, password_hash, phone, nickname, resident_nickname, device_qr}`
   - 校验 device_qr（从 device_factory_meta 或 IPv6 QR 解出 device_uid）
   - 落 family user + resident + resident_caregivers + bind device
   - 返回 access token
4. FE 注册页 `/signup` + login 页加 Sign Up 链接

## 关键约束

- B2C tenant 共用一个 `fd00:0:4::/48` (wisefido) 还是每户独立 tenant？
  - 共用：简单，但 60k user 上限（足够）
  - 独立：cleaner 但 IPAM slot 消耗快
  - 建议：B2C 全共用 wisefido tenant，user 通过 spatial.assigned_branch /56 隔离
- Signup OTP 校验避免恶意注册：phone OTP / email OTP / device_qr 三因素至少一个
- B2B → B2C 不可改 tenant.kind（数据模型不兼容）
