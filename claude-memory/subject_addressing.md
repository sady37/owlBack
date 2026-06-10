---
name: Subject 寻址 - Mobile IPv6 风格 (resident/caregiver)
description: resident 走 Mobile IPv6 HoA+CoA；caregiver 占位不接线；admin/visitor 不入体系；编码与 spatial 同 tenant 内 branch=0xFF 区分
type: project
originSessionId: 53d81dc2-b99e-4371-95da-247c2cf5c46c
---
resident / caregiver 当作 RFC 6275 Mobile IPv6 mobile node 处理（决策于 2026-05-08，落进 owl-common/spatial/subject.go + doc/datagram_envelope.md §2.7）。

**Why**：业务里大量 device↔room↔resident↔caregiver 关系判断，传统 SQL JOIN 写一堆。Mobile IPv6 的 HoA+CoA 模型 + IPv6 prefix-match 标准原语，把这些查询全部退化成 `prefix.Contains()` 一行。零自定义关系判断代码。

**How to apply**：
- 业务"X 在 Y 内吗 / X 关联 Y 吗"全走 `netip.Prefix.Contains()` / `LongestPrefixMatch()`，不要写 multi-table JOIN
- resident 移动 = Mobile IP 重新绑定 CoA，不改 HoA；HoA 永久不变
- 任何"找当前在某区的 caregiver" 类查询都该走 binding cache + prefix-match，不该 SQL JOIN devices→rooms→assignments

## 编码布局

```
fd00 : 0000 : TTTT : FF KK : SSSS : 0000 : 0000 : 0000
              tenant FF=     kind  subject_id (uint16)   reserved 48 bit
                    subject  =01 resident
                    namespace=02 caregiver
                             0x03..0xFE 保留
```

字段宽度与 spatial 对齐（subject_id = uint16 同 unit_slot 宽度，65k/tenant×kind 远超总人数）。

## Phase A 状态（2026-05-08 已 push, commit ec1733a）

| 类型 | 进体系 | encoder 状态 | 业务接线 |
|---|---|---|---|
| resident | ✓ | `BuildResidentHoA` 完整暴露 | HoA 已可用；CoA binding cache 待二期 |
| caregiver | ✓ kind=0x02 永久占位 | `BuildSubjectHoA(t, Caregiver, id)` 显式调用，**便利封装故意不暴露** | **不接线**：alarm 路由 / 访问权限沿用 user role；DB 应预留 subject_id/home_addr 列（NULL 允许） |
| admin/manager | ✗ | — | UUID + role 在 user 表，永远不进 IPv6 |
| visitor | ✗ | — | monitor track 流 anonymous tracking，不持久化 |

## caregiver 后期接入零迁移

HoA 编码方案 frozen。将来引入 caregiver tracking 时只需：
- 加 `caregiver_binding(subject_id, coa, coa_since, on_shift)` 表
- 加 binding-update service（消费 monitor track 流识别 caregiver → 更新 CoA）
- alarm 路由切到 `LookupCaregiverByCoAPrefix`
- 访问权限切到 HoA-based ACL

不动 spatial 包，不动现有 schema 列。

## branch=0xFF 是约定不是强制

encoder 不拒分配 0xFF（保持纯函数）；spatial branch allocator **必须主动跳过 0xFF**，留给 subject namespace。0xFE 个空间 branch 槽位对 owl 业务已足够。
