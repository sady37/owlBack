# Spatial 查询模式 — 位掩码 vs JOIN

> Code review checklist for unit / room / bed / device 查询。
> 前提：所有 spatial slot 已遵循 [IPAM slot 保留语义](../../.claude/projects/-home-wisefido-owl/memory/ipam_slot_reservation.md) — slot 0 = unbound 哨兵，不分配给真实节点。

## v2 IPv6 字节布局回顾

```
fd00 : 0    : tttt : bbss : uuuu : rrbb : MAC2 : MAC1
       (0-1) (2-5) (6-7)  (8-9) (10-11)(12-13)(14-15)
       sub   tenant branch  unit  room   MAC last 32 bit
                    +site         +bed
```

| Slot | bytes | 0 含义 | wildcard 保留 |
|---|---|---|---|
| branch | 6 | unbound | 0xFF |
| site (building<<4|floor) | 7 | unbound | 0xF/0xF |
| unit | 8-9 | unbound | 0xFFFF |
| room | 10 | unbound | 0xFF |
| bed | 11 | unbound | 0xFF |
| device MAC | 12-15 | n/a | n/a |

## 推荐模式 — 位掩码（首选）

### 正向：从 device 反推绑定深度

| 想知道什么 | SQL 片段 | 说明 |
|---|---|---|
| 设备所属 tenant /48 | `network(set_masklen(addr, 48))` | 总能反推 |
| 设备所属 branch /56 | `network(set_masklen(addr, 56))` IF `byte 6 != 0` | 否则未绑 branch |
| 设备所属 site /64 | `network(set_masklen(addr, 64))` IF `byte 7 != 0` | 否则未绑 site |
| 设备所属 unit /80 | `network(set_masklen(addr, 80))` IF `(addr & '::ffff:0:0:0'::inet) <> '::'::inet` | bytes 8-9 != 0 |
| 设备所属 room /88 | `network(set_masklen(addr, 88))` IF `(addr & '::ff00:0:0'::inet) <> '::'::inet` | byte 10 != 0 |
| 设备所属 bed /96 | `network(set_masklen(addr, 96))` IF `(addr & '::ff:0:0'::inet) <> '::'::inet` | byte 11 != 0 |

**位掩码字面量速查**：
- byte 6: `'0:0:0:ff::'::inet`
- byte 7: `'0:0:0:ff::'::inet` (低 8 bit) — 实际用 `'::ff00::'`
- bytes 8-9: `'::ffff:0:0:0'::inet`
- byte 10: `'::ff00:0:0'::inet`
- byte 11: `'::ff:0:0'::inet`
- bytes 12-15 (MAC): `'::ffff:ffff'::inet`

### 反向：按节点查它包含的下层

| 任务 | SQL |
|---|---|
| 某 branch 下所有 devices | `WHERE addr <<= '<branch>/56'::inet` |
| 某 unit 下所有 devices | `WHERE addr <<= '<unit>/80'::inet` |
| 某 room 下绑定的 devices（room-level only，不含 bed） | `WHERE network(set_masklen(addr, 88)) = '<room>/88'::inet AND (addr & '::ff:0:0'::inet) = '::'::inet` |
| 某 bed 下绑定的 devices | `WHERE network(set_masklen(addr, 96)) = '<bed>/96'::inet` |
| 某 tenant 下未绑 device 的 device（tenant 池） | `WHERE addr = network(set_masklen(addr, 48)) | (addr & '::ffff:ffff'::inet)` |

### 反向：构造新地址（绑定操作）

```sql
-- 把 device 绑到 room（保留 MAC 末 4 字节）
new_addr = set_masklen(network(set_masklen(room_prefix, 88)), 128) | (current_addr & '::ffff:ffff'::INET)

-- 把 device 绑到 bed
new_addr = set_masklen(network(set_masklen(bed_prefix, 96)), 128) | (current_addr & '::ffff:ffff'::INET)

-- 解绑回 tenant 池（清 bytes 6-11）
new_addr = reset_device_prefix(addr, 'fd00::/32', 32, 'branch')
```

## 反模式 — JOIN（仅用于 legacy 兼容）

```sql
-- ❌ 在 owl_v2 完成 slot=0 cleanup 后，禁用以下模式：
LEFT JOIN rooms r ON r.spatial_prefix = network(set_masklen(d.spatial_addr, 88))
LEFT JOIN beds  b ON b.spatial_prefix = network(set_masklen(d.spatial_addr, 96))
```

**当前为什么还需要 JOIN**：
- legacy slot=0 数据让 `byte 10=0` 既可能是「真未绑」也可能是「绑到 slot=0 的 legacy room」
- JOIN 验证 prefix 真实存在 → 可以区分

**何时可以删 JOIN**：
- legacy slot=0 数据全部迁移到 slot 1+ 后
- 此时 `byte 10=0` 唯一含义 = 未绑

**Code review 红线**：
- 任何**新写**的 device/unit/room/bed 查询代码都应**优先用位掩码模式**
- 仅当方法被旧代码调用，且需要兼容 legacy slot=0 数据时才允许 JOIN
- JOIN 必须加注释解释「兼容 legacy slot=0；cleanup 后改位掩码」

## JOIN 残留状态（已清理）

✅ 2026-05-09 legacy slot=0 全部迁移（详见 `owlRD/dbv2/migrations/2026-05-09_legacy_slot0_cleanup.sql`），CHECK constraints 已 VALIDATE，以下方法已切纯位掩码：

| 文件 | 函数 | 状态 |
|---|---|---|
| `postgres_devices.go::ListDevices` | byte 10/11 != 0 位掩码 | ✅ 已删 JOIN rooms/beds |
| `postgres_devices.go::GetDevicesByRoomIDs` | 本来就纯位掩码 | ✅ 无 rooms JOIN |
| `postgres_devices.go::GetDevicesByBedIDs` | 本来就纯位掩码 | ✅ 无 beds JOIN |

后续新增 device/unit/room/bed 查询代码必须遵循"位掩码优先"原则，禁止再 JOIN 对应 spatial 表做存在性验证。
