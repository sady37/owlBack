---
name: PHI Encryption Restore Prompt
description: resident_phi AES-256-GCM加密恢复提示词，含架构设计、文件清单、实现步骤，供新会话使用
type: project
originSessionId: 9bcbb98f-46ff-4434-937d-0f2abb4643b0
---
## PHI 加密恢复任务

PHI 加密功能在 2026-04-14 曾完整实现（14条数据已迁移），在 4/16 的 card_id 架构重构 (00ae2f0) 中丢失。需重新实现。

### 当前代码状态（00ae2f0 之后）
- `wisefido-data/internal/domain/resident_phi.go` — 结构体存在，全字段定义
- `wisefido-data/internal/domain/resident_phi_update.go` — 更新模型（update/delete/keep 三态）
- `wisefido-data/internal/repository/postgres_residents.go` — GetResidentPHI / UpsertResidentPHIFields **明文读写**
- `wisefido-data/internal/service/resident_service.go` — 业务层，含 SHA256 hash 查询
- `wisefido-data/internal/http/resident_handler.go` — PUT /admin/api/v1/residents/:id/phi
- **不存在**: crypto模块、K服务、migrate_phi、_enc列、MW

### 架构设计

**K 服务（独立进程）**
- 独立 Go 进程，Unix socket `/tmp/owl-kms.sock`
- 存储/派发 tenant_key（每 tenant 一个 AES-256 密钥）
- 启动时从 MASTER_PIN 解密本地密钥库（JSON）
- 不在热路径：data 启动时调一次，缓存 tenant_key 在内存
- API: `GET /key/{tenant_id}`, `POST /key/{tenant_id}`

**加密层**
- `internal/crypto/phi_crypto.go` — AES-256-GCM
  - `Encrypt(plaintext, key) → base64(nonce+ciphertext+tag)`
  - `Decrypt(ciphertext, key) → plaintext`
  - 空字符串/零值不加密
- 加解密在 repository 层调用（postgres_residents.go）
- tenant_key 通过 service 传入 repo

**数据库**
- resident_phi 所有数据字段改为 `_enc TEXT` 列（first_name_enc, last_name_enc, gender_enc, date_of_birth_enc, resident_phone_enc, resident_email_enc, weight_lb_enc, height_ft_enc, height_in_enc, mobility_level_enc, tremor_status_enc, mobility_aid_enc, adl_assistance_enc, comm_status_enc, plus_code_enc, medical_conditions_enc, medications_enc, allergies_enc, primary_physician_enc, emergency_contact_*_enc, home_address_enc, home_city_enc, home_state_enc, home_zip_enc）
- bool/int/float 也加密（先 strconv 转字符串）
- 保留原列名过渡，后续 DROP

**MW（Matrix Wallet）**
- 12×7 查询表，K 服务丢失 MASTER_PIN 时恢复
- `cmd/generate_mw/main.go`
- master_pin 单独交付，不在 MW 打印件

**migrate_phi**
- `cmd/migrate_phi/main.go` — 明文→加密迁移
- 已知 bug: ~L152 fmt.Sprintf 参数不匹配需修复

**plus_code 限流**: 5次/天/用户，service 层实现

### 实现顺序
1. `internal/crypto/phi_crypto.go` — AES-256-GCM
2. `cmd/owl-kms/main.go` — K 服务骨架
3. DB migration — _enc 列（owlRD/db/）
4. repository 层 — 读 _enc+解密，写加密+_enc
5. service 层 — 启动获取 tenant_key
6. migrate_phi — 读明文→加密→写 _enc
7. 测试验证

### 关键约束
- 全字段加密，不区分敏感级别
- K 只在启动时调用
- 14 条现有数据需迁移
- 加密后明文列暂保留
