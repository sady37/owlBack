# v1 ↔ v2 Schema 字段对齐 — Final State Checklist

> 2026-05-10 — 全链路 rename 完成（DB + BE + FE + IPAM），**0 转换层**，三层字段名完全一致。
> 本文档反映 **post-rename 终态**，作为后续 audit / 新功能开发的对照表。

## 设计原则

1. **业务语义一致** → 字段名沿用 v1（admin/manager/FE 用户认知不变）
2. **type 升级** → UUID/string 改 IPv6 INET CIDR，但**名字稳定**
3. **不留 adapter / shim / 字段名映射层** —— DB ↔ BE ↔ FE 字段名严格一致
4. **v2 新概念**（IPv6 spatial、HoA、`device_factory_meta` 拆表等）保留 v2 命名

---

## 表-级 Final State

### 1. tenants
| 列 | type | 说明 |
|---|---|---|
| `tenant_id` (PK) | INET /48 | IPv6 prefix；v1 是 UUID |
| `tenant_slot` | SMALLINT | uint16，UI 显示 |
| `tenant_name / contact_name / contact_email / contact_phone / timezone / status` | — | |

### 2. branches
| 列 | type | 说明 |
|---|---|---|
| `branch_id` (PK) | INET /56 | parent tenant 靠 INET prefix-match |
| `branch_slot / branch_name / address / timezone / status` | — | |

### 3. sites *(v1 buildings 改造)*
| 列 | type | 说明 |
|---|---|---|
| `site_id` (PK) | INET /64 | v1 `buildings` 拆为 `sites` (building+floor) |
| `site_slot / building / floor / site_name / address` | — | site_slot = (building<<4 \| floor) |

### 4. units
| 列 | type | 说明 |
|---|---|---|
| `unit_id` (PK) | INET /80 | |
| `unit_slot / unit_name / unit_property(0=Home/1=Facility) / unit_type(0..3) / unit_layout_type / timezone` | — | v1 `is_public/is_shared_unit` 派生自 unit_type |

### 5. rooms
| 列 | type | 说明 |
|---|---|---|
| `room_id` (PK) | INET /88 | |
| `room_slot / room_name / room_type / is_primary / description` | — | |

### 6. beds
| 列 | type | 说明 |
|---|---|---|
| `bed_id` (PK) | INET /96 | |
| `bed_slot / bed_name / description` | — | v1 mattress_* 字段已删 |

### 7. devices
| 列 | type | 说明 |
|---|---|---|
| `device_ipv6` (PK) | INET /128 | 完整 IPv6 地址（业务空间绑定 + MAC 末 32bit）|
| `device_id` (UNIQUE) | UUID | 外部稳定 ID，关联 `device_factory_meta` |
| `monitoring_enabled` | BOOL | per-device 业务开关；**FE Allow Access toggle 写入此列** |

> v1 单表 `devices` 拆为 v2 四表：
> - `devices` (业务空间绑定 + monitoring_enabled)
> - `device_factory_meta` (device_uid / mac_wifi / mcu_model / device_code 等静态)
> - `device_runtime_state` (firmware_version / online / last_position_addr 等运行态)
> - `device_ota` (target_firmware_version / status / progress 等升级计划)
>
> **v1 `business_access`** 概念在 v2 已删除；权限层面由 role + spatial scope 表达，
> per-device 业务开关由 `devices.monitoring_enabled` 承载（tenant_admin 可改）。

### 8. residents
| 列 | type | 说明 |
|---|---|---|
| `resident_id` (PK) | INET /128 | = HoA (Mobile IPv6 Home Address)；v1 是 UUID |
| `resident_slot` | INTEGER | uint16 in (tenant, kind=Resident) |
| `resident_account` | VARCHAR(50) | 人类可读，per-tenant UNIQUE |
| `nickname (NOT NULL) / gender / birth_year / admission_date / discharge_date / service_level / status / note` | — | v1 命名沿用 |

### 9. resident_phi
| 列 | type | 说明 |
|---|---|---|
| `resident_id` (PK FK) | INET | 与 residents 同 |
| `full_name_enc/iv/tag / phone_enc/iv/tag / address_enc/iv/tag / home_address_enc/iv/tag / medical_summary_enc/iv/tag` | BYTEA | AES-256-GCM；K 服务 + MW 双因素 |
| `mapcode / mapcode_kind / otp_code/purpose/issued_at/expires_at/used_at/used_by` | — | home care 紧急上门 |
| `accessed_count / last_accessed_at / last_accessed_by / encryption_version / encrypted_at` | — | PHI 审计 |

### 10. resident_unit
| 列 | type | 说明 |
|---|---|---|
| `record_id` (PK) | UUID | |
| `resident_id` (FK) | INET | |
| `spatial_prefix` | INET | **多态**：/48 tenant / /56 branch / /80 unit / /88 room / /96 bed |
| `valid_from / valid_to / move_reason / moved_by / notes` | — | 历史可追溯 |

### 11. resident_caregivers
| 列 | type | 说明 |
|---|---|---|
| `assignment_id` (PK) | UUID | |
| `resident_id` (FK) | INET | |
| `caregiver_id` 或 `care_team_id` | INET 或 UUID | 二选一；caregiver_id 指向 users.hoa |
| `role / is_primary / care_priority / valid_from/to / notes` | — | |

### 12. resident_contacts
| 列 | type | 说明 |
|---|---|---|
| `contact_id` (PK) | UUID | |
| `resident_id` (FK) | INET | |
| `linked_user_id` | UUID | 可选，关联 users |
| `relationship / is_emergency / is_primary / contact_priority / preferred_channel / notify_alarm_severity / notify_quiet_hours` | — | |
| `full_name_enc/iv/tag / phone_enc/iv/tag / email_enc/iv/tag / address_enc/iv/tag` | BYTEA | 加密 |

### 13. users
| 列 | type | 说明 |
|---|---|---|
| `user_id` (PK) | UUID | |
| `tenant_id` (FK) | INET /48 | NULL = platform-level admin |
| `user_account` (per-tenant UNIQUE) | VARCHAR(100) | v1 名沿用；FE 字段名 `user_account` |
| `password_hash` | VARCHAR(255) bcrypt(sha256(plain)) | |
| `password_check_hash` | BYTEA sha256(plain) | 反向定位 + admin 全局唯一 |
| `pin_hash` | VARCHAR(255) bcrypt | mobile PIN |
| `hoa` (UNIQUE) | INET /128 | caregiver kind；admin/manager NULL |
| `subject_slot` | INTEGER | 配对 hoa |
| `nickname / full_name / email / phone / employee_code / role / hire_date / leave_date / status` | — | |
| `notify_mode / work_days / work_time_start / work_time_end` | — | |
| `failed_login_count / locked_until / last_failed_at / last_login_at / last_login_ip / last_active_at` | — | |
| `relegation` | VARCHAR(20) | 'all'/'branch'/'assigned_only'；**保留 v2 命名**（区别于 v1 alarm_scope）|

### 14. user_branches
| 列 | type | 说明 |
|---|---|---|
| `user_id + branch_id` (复合 PK) | UUID + INET | |
| `is_primary / valid_from / valid_to / notes` | — | |

### 15. roles
| 列 | type | 说明 |
|---|---|---|
| `role_id` (PK) | UUID | |
| `tenant_id` (FK) | INET | NULL = system role |
| `role_code` | VARCHAR(50) | 'platform_admin'/'tenant_admin'/'manager'/'nurse'/'caregiver'/'family'/'viewer' |
| `role_name` (NOT NULL) | VARCHAR(100) | 显示名（**独立列**，不再拼进 description）|
| `description` | TEXT | 详细描述 body（不含显示名前缀）|
| `is_system` | BOOL | 系统预置不可删 |
| `scope_prefix_len` | SMALLINT | 32/48/56/64/80/88/96/128 — 可重置 device 字节段层级 |

### 16. role_permissions
| 列 | type | 说明 |
|---|---|---|
| `role_permission_id` (PK) | UUID | |
| `role_id` (FK) | UUID | |
| `permission` | VARCHAR(100) | 'devices.read' / 'alarms.process' / '*' 等 |
| `resource_scope` | INET | NULL = tenant 全范围 |
| `granted_at / granted_by` | — | |

### 17. care_teams / care_team_members
| 列 | type | 说明 |
|---|---|---|
| `team_id` (PK) | UUID | |
| `tenant_id` (FK) | INET | |
| `team_code / team_name / team_kind / spatial_scope / parent_team_id / description / color_hex / is_active` | — | spatial_scope 多态 INET |

### 18. alarm_events
| 列 | type | 说明 |
|---|---|---|
| `alarm_id` (PK) | UUID | |
| `device_ipv6` + `device_id` | INET + UUID | 双 ref |
| `resident_id` (FK) | INET | |
| `tenant_name / branch_name / unit_name / room_name / bed_name` | — | denormalized snapshot |
| `alarm_kind / severity / reason / trace_id / parent_span / process_status / handle_type / handler_notes / payload jsonb / evidence jsonb` | — | v2 完全重设计 |

---

## 多态列（不命名特定父表）

这些列**保留 spatial_prefix 命名**——本质多态、polymorphic INET：

| 表 | 列 | 长度 |
|---|---|---|
| `resident_unit` | `spatial_prefix` | /48~/96 |
| `room_visual_layout` | `spatial_prefix` | /80 (unit) 或 /88 (room) — drop FK + CHECK |
| `roomengine_grid_snapshot[_history]` | `spatial_prefix` | /80 或 /88 — 同上 |
| `spatial_config` | `spatial_prefix` | /48~/128 任意 |
| `cards` | `spatial_prefix` | /48~/128 决定 card 类型 |
| `visitor_episode` | `spatial_prefix` | room/unit prefix |

`subject_binding_cache.hoa` — polymorphic subject (resident HoA + caregiver HoA)。

---

## v2 终态约束（已审计通过）

- ✅ DB schema 所有表 column rename 完成
- ✅ owl-common/ipam SQL 字符串同步
- ✅ owlBack 7 个服务（wisefido-data / qinglan / cardagg / iot / sensor / sleepace / ai-health）`go build` 全绿
- ✅ owlFront TS interfaces + Vue 组件用 v1 字段名，**无 adapter shim**：
  - `auth.ts`: `loginApi` 直接传 `user_account / tenant_id`
  - `resident.ts`: 删除 `uiToV2UpdateInput / v2ToUIResident` 转换函数
  - HTTP header `X-Tenant-Id`（不再用 `X-Tenant-Prefix`）

## 历史 bug — 已修

| 文件 | 问题 | 修复 |
|---|---|---|
| `45_tenant_holidays.sql:63` / live `should_notify_user` | `tenants WHERE spatial_prefix=...` | → `tenant_id` |
| `postgres_units.go` | 多处 `s/u/r/b.spatial_prefix` 漏改 | 按 alias 分别 → site_id/unit_id/room_id/bed_id |
| `postgres_branches.go isBranchEmpty` | `residents WHERE hoa` | → `resident_id` |
| `postgres_residents_v2.go` | join alias spatial_prefix 漏改 + `r.service_tier/move_in_date` | → 各 `*_id` + 新 residents 列名 |
| `resident_service_v2.go` | 同上 | 同上 |
| `postgres_roles.go rolesSelectCols` | `role_name + "\n" + description` 拼回单字段 → 重复前缀 | 改 SELECT 返双列；service 用 role_name 作 display_name |
| `service/role_service.go UpdateRole/CreateRole` | 拼 `display_name + "\n" + description` 写回 description | 改双列独立写入 |
| `postgres_device_store.go` | `allow_access` 写入被忽略 | → `UPDATE devices.monitoring_enabled` |
| `owlFront/src/api/admin/role/role.ts` | `params:` (query string) 替代 `data:` (body) | 改 3 处 `data:` |

---

## DeviceStore stack — 当前关注点

业务模型（v1 → v2）：
- v1 `devices.business_access` → **已删**
- v2 替代：`devices.monitoring_enabled BOOL` + role/spatial scope 权限
- 管理界面：`/admin/devicestore`（仅 SystemAdmin / SystemOperator）
- FE Allow Access toggle → BE `allow_access` 字段 → repo 写 `devices.monitoring_enabled`

API endpoints (v1 路径，BE 落 v2 schema)：
- `GET  /admin/api/v1/device-store` — list（含 allow_access 字段）
- `PUT  /admin/api/v1/device-store/batch` — 批量更新 `{updates:[{device_uid, data:{...}}]}`
- `GET  /admin/api/v1/device-store/export` — xlsx 导出
- `POST /admin/api/v1/device-store/import` — xlsx 导入
- `DELETE /admin/api/v1/device-store/{device_uid}` — 删除
