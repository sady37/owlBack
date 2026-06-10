---
name: Resident v2 设计原则
description: residents/resident_phi/resident_unit/resident_caregivers v2 schema + HIPAA minimum-necessary + Caregiver vs CareTeam FE 二选一
type: project
originSessionId: 6a5f2016-6082-4856-b91a-b95fb56f8a66
---
## v2 schema 关键差异（vs v1）

| 维度 | v1 | v2 |
|---|---|---|
| PK | `resident_id` UUID | **`hoa` INET** (`fd00:tenant:ff01:slot::`) |
| 空间归属 | `residents.{unit_id, room_id, bed_id, branch_id}` UUID 列 | `resident_unit (resident_hoa, spatial_prefix INET, valid_to)` 关联表 + hoa /48 反推 tenant、/56 反推 branch |
| 业务字段 | `resident_account/service_level/admission_date/discharge_date/is_access_enabled` | `resident_account` (新加) / `service_tier` / `move_in_date` / `move_out_date` |
| PHI | 明文列 | `resident_phi.{full_name,phone,address,home_address}_enc/iv/tag` AES-256-GCM |
| Slot 范围 | `resident_slot >= 0 AND <= 65534` | **`>= 1 AND <= 65534`**（slot 0 保留 unbound 哨兵；2026-05-10 收紧） |

## resident_account 设计

人类可读账号 ID（如 `R0001`），per-tenant UNIQUE (case-insensitive)：
```sql
CREATE UNIQUE INDEX idx_residents_account
  ON residents (network(set_masklen(hoa, 48)), LOWER(resident_account))
  WHERE resident_account IS NOT NULL;
```
2026-05-10 demo tenant 10 行 backfill `R0001..R0010`（slot 0 padding）。
不在 hoa 里占位（hoa 主机位预留 48 bit 给未来扩展，不引入 fixed-32-bit residentID）。

**Why**: IPv6 hoa 不可记忆；resident_account 给业务/UI 用。
**How to apply**: List/Get/Create 都用 resident_account 作 user-facing 标识；hoa 仅内部 spatial 路由。

## HIPAA minimum-necessary 原则

业务规则：**存的 PHI 越少，泄漏风险越小**。
- `residents.nickname` NOT NULL — 日常 UI/alarm 标识；免依赖 PHI 即可工作
- `resident_phi.full_name_enc / phone_enc / address_enc / home_address_enc` 全 nullable
- FE 所有 PHI 字段 placeholder *"Optional, leave blank if not needed"*
- Save 时空值不发后端 → 不写 resident_phi 行 / 字段保持 NULL
- 加密字段 UI 用 `••••••` mask 显示；仅当用户显式输入新值才 rotate

## Caregiver vs CareTeam — FE 二选一

后端 schema **保持中立**（不加 enum 列）：
- `resident_caregivers` 表（直接绑定 caregiver/nurse user）和 `care_team_members` 表（user 通过 careteam 间接关联 resident）**都保留**
- 前端 Edit Resident 加 radio: `◯ Caregiver  ◯ CareTeam`
- 切换时把另一方的关联清空提交（PUT 一次完成 add+remove）
- 未来要并存改 radio→checkbox 即可，**无 schema migration**

**Why**: 简化业务规则（每 resident 只挂一种关联）；同时保留扩展灵活性。
**How to apply**: GetResident v2 返 caregiver_list + careteam_list；FE 按当前哪个非空决定 radio 默认；切换时提交清空另一方。

## 业务规则（Forward Design 用）— 仅参考 v1 业务，不复用 v1 代码

### 角色权限
- **Admin / Manager**：full CRUD（含 admission/discharge）
- **Nurse**：可改 Resident 所有字段；**不能处理 admission/discharge**（涉及财务）
- **Resident**：只能 `GET /residents/{自己 hoa}`

### 软删 + 硬删
- 默认 **soft delete**：`UPDATE residents SET status = 'deleted'`，列表默认隐藏
- **Hard delete (Clear)** 条件：自该 resident 创建（或 bind device）起，三表均无 `resident_hoa` 记录：
  - `alarm_events`
  - `event_log`
  - `monitor_stream`
- 满足 → `DELETE FROM residents WHERE hoa = $1`（FK ON DELETE CASCADE 自动清 resident_unit/caregivers/phi/contacts）
- 否则 → 拒绝硬删，回退到 soft

### 转院
跨 tenant = **discharge from A + admission to B**（两步），不 in-place 改 hoa
（hoa 含 tenant prefix，转 tenant 必生新 hoa）

### HIPAA min-necessary
PHI 字段全 optional；空值不发后端 / 不写 resident_phi 行

### Caregiver / CareTeam — FE 二选一
schema 中立（`resident_caregivers` 一表 `caregiver_hoa` 或 `care_team_id` 二选一）；FE radio 切换时清空对方

### Unit type 联动
- `unit_type=1 (Private)` → 只 bind unit（spatial_prefix /80）
- `unit_type=2 (Share)` → 可绑 room (/88) 或 bed (/96)
- `unit_type=3 (Public)` → resident 不可 bind

## v2 实施 — Forward Design（2026-05-10 起）

❌ **禁止**: short-circuit / 复用 v1 DTO / wrap v1 service
✅ **必须**: from scratch 新建 v2 文件
  - `internal/domain/resident_v2.go` — 纯 v2 types（hoa, resident_account, nickname, status, service_tier, move_in/out, unit_prefix, branch_prefix, caregivers[], teams[], phi）
  - `internal/repository/postgres_residents_v2.go` — 唯一 repo，hoa PK
  - `internal/service/resident_v2_service.go` — 业务规则
  - `internal/http/resident_v2_handler.go` — 新路由 `/admin/api/v2/residents`
  - main.go wire；FE 切 `/v2/` endpoint
v1 代码原地保留（git ref），下次清理 PR 整段删除

## 相关文档/代码

- 后端 v2 短路: `wisefido-data/internal/service/resident_service_v2.go`
- v1 SQL reference 仍保留: `resident_service.go::ListResidents` / `GetResident`（不进入）
- spatial 反推规范: [doc/spatial_query_patterns.md](../../../owl/owlBack/doc/spatial_query_patterns.md)
