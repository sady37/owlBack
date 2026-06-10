---
name: IPAM slot 0/0xFF 保留语义
description: 各 spatial 层 slot 0 = unbound 哨兵，0xFF/0xFFFF = wildcard/subject namespace 保留；分配权威是 owl-common/ipam (kea + PG)，不在 app 层重复 MAX+1
type: project
originSessionId: 6a5f2016-6082-4856-b91a-b95fb56f8a66
---
## 设计原则

每个 spatial 层 slot 数值都保留两端：
| 层 | 字节 | slot 范围 | 0 含义 | 上界保留 |
|---|---|---|---|---|
| branch | byte 6 | 1..254 | unbound（device 在 tenant pool 未绑 branch） | 0xFF = subject namespace (resident HoA) |
| site | byte 7 = (building<<4|floor) | building 1..14 / floor 1..14 | unbound | 0xF = wildcard 保留 |
| unit | bytes 8-9 | 1..65534 | unbound | 0xFFFF = wildcard |
| room | byte 10 | 1..254 | unbound | 0xFF |
| bed | byte 11 | 1..254 | unbound | 0xFF |
| device | bytes 12-15 | MAC 末 32-bit | (无意义；MAC 真实值) | — |

**Why**：让 `network(set_masklen(addr, N))` + slot 字节检测能**唯一**判定绑定深度，无需 JOIN 验证存在性。
- byte 10 = 0 → device 必然未绑 room（不会与"绑到默认 Main room"歧义，因为 room_slot=0 不再合法）
- 同理 byte 11 / byte 6 / bytes 8-9

## 分配权威（不要在 app 层做 MAX+1）

正典分配器在 `owl-common/ipam`：
- `pg_backend.go::Allocate{Branch,Site,Unit,Room,Bed}` — PG-first，CHECK constraint 兜底
- `kea_client.go` — 后续接 kea-DHCPv6-PD 时把分配源改 kea
- `netbox_backend.go` — Future NetBox 集成

`wisefido-data/internal/repository/postgres_*.go::Create*` 也有平行 MAX+1 实现，是 v1 legacy FE 路径，**应统一切到 IPAM**（Phase 待办）。

kea 配置端可在 pool range 排除 0 / 0xFF：
```
"pools": [{"prefix": "fd00:0:T:01::/56", "delegated-len": 56, "...": "exclude T:00 / T:FF"}]
```
当前 PG-first 模式，由 PG CHECK + allocator `COALESCE(MAX,0)+1` 保证。

## DB CHECK 约束

5 张表 schema 加 CHECK（`slot BETWEEN 1 AND <max>`），2026-05-09 已 VALIDATE（不再是 NOT VALID）。
源文件：`owlRD/dbv2/{11_branches,12_sites,13_units,14_rooms,15_beds}.sql`

## Legacy 数据迁移 — 已完成

✅ 2026-05-09 cleanup（migration: `owlRD/dbv2/migrations/2026-05-09_legacy_slot0_cleanup.sql`）：
- 1 branch (fd00:0:7::) + 4 sites + 5 units + 10 rooms + 8 beds slot=0 全部迁到 ≥1
- + 5 cascade rooms / 2 cascade units / 1 cascade bed 跟随 parent prefix 改写
- + 1 bound device (fcaedf95...) spatial_addr 跟随 cascade（device_ota FK auto cascade）
- VALIDATE 5 张表 CHECK 通过；行数和唯一性 sanity 全过

迁移完成后已删除 `postgres_devices.go::ListDevices` 的 LEFT JOIN rooms/beds，改为纯位掩码（byte 10/11 != 0 检测）。后续新查询代码必须遵循 [doc/spatial_query_patterns.md](../../../owl/owlBack/doc/spatial_query_patterns.md) 位掩码优先原则。
