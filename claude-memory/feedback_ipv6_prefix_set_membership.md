---
name: feedback-ipv6-prefix-set-membership
description: 集合/包含关系（Unit/CardID/Room/Bed/Device 之间）首选 IPv6 掩码计算，不查 DB JOIN
metadata: 
  node_type: memory
  type: feedback
  originSessionId: d23e371f-aa78-4087-ba7b-63ba1a80ec04
---

判 Unit↔Room↔Bed↔Device 任意两层的"属于/包含/同 X 下"关系时，首选 **IPv6 prefix mask 计算**
（`netip.Prefix.Contains` / SQL `<<=`），不要先 DB JOIN 查 room_id/bed_id。

层级（spatial_prefix 自描述）：

| 层 | mask | 关系 |
|----|------|------|
| Tenant | /48 | 包含所有 unit |
| Unit   | /80 | 包含本 unit 所有 room |
| Room   | /88 | 包含本 room 所有 bed |
| Bed    | /96 | 包含本 bed 所有 device |
| Device | /128 | 叶子 |

典型用法：
- "/88 room 下有几张 /96 bed" → 扫已知 /96 集合 `room88.Contains(bed96.Addr())`
- "device 属哪个 room" → `prefixOf(deviceAddr, 88)` 直接得 room CIDR，不查 devices.bound_room_id
- "bed 属哪个 unit" → `prefixOf(bedAddr, 80)`

**Why**：spatial_prefix 是自描述地址（[[platform_agent_addressing]] / [[subject_addressing]]）—
DB JOIN 是绕远路，慢且需 tenant_id + null-handle + 软删 filter。owl_v2 IPAM
([[ipam_slot_reservation]]) 保证前缀正确分配后，containment 即真相，无 drift 风险。

**How to apply**：
- 写跨实体逻辑（dedup/聚合/权限/枚举）时第一反应 = `netip.Prefix.Contains` 或 SQL `INET <<= CIDR`
- 仅当查的是 *业务字段*（room_name / unit_kind / device_type）才走 DB
- LPM（最长前缀匹配）找祖先卡：参见 [[card_display_projector_handoff]] / cardagg `LookupCardByPrefix`
- 反例：`devices JOIN beds JOIN rooms WHERE bed_id=X` 三表 JOIN 拿 room_id — 用 prefixOf 一行替代

落地示例：[[bed_presence_fusion]] `ExtraPeopleInRoom(roomCIDR)` 扫 /96 bed 集合直接
`roomPfx.Contains(bedPfx.Addr())`，零 DB 查询。
