---
name: Device unbind via prefix reset (capability-as-prefix)
description: device→unit 解绑模型：UPDATE spatial_addr 把绑定段清零（保留 MAC 末 32 bit + 父级前缀），权限边界 == 角色可重置的字节段；不要硬删 devices 行
type: project
originSessionId: 6a5f2016-6082-4856-b91a-b95fb56f8a66
---
## 核心思想

把 device 绑定关系表达为 IPv6 /128 的字节段，"解绑"= UPDATE 清零相应字节段；不删行（R-002 + HIPAA 友好）。

## 字节布局

```
fd00:owl:0001:0112:0042:0301:A2AC:D523
         │ │  │    │ │   └────┬─────┘
         │ │  │    │ │        └── Device-host (MAC 末 32 bit, 不可改)
         │ │  │    │ └────── Bed     (byte 11)
         │ │  │    └──────── Room    (byte 10)
         │ │  └──────────── Unit    (bytes 8-9)
         │ └──────────────── Site    (byte 7 = building<<4 | floor)
         └────────────────── Branch  (byte 6)
   Tenant /48 = bytes 0-5
   Device MAC = bytes 12-15
```

## 权限层级 ↔ 可重置字节段

- `platform_admin`: bytes 12-15 (MAC)；注册/退役物理设备
- `tenant_admin`: bytes 6-11 within own /48；branch/site/unit/room/bed 全可重置
- `branch_manager`: bytes 7-11 within own /56；site/unit/room/bed 重置
- `unit_manager`: bytes 10-11 within own /80；room/bed 重置
- 重置 = UPDATE 字节为 0；天然解除该层级以下所有绑定

## 例子

device `fd00:owl:0001:0112:0042:0301:A2AC:D523` 解绑 unit（保留 branch）：
- branch_manager 操作：UPDATE spatial_addr → `fd00:owl:0001:0100:0000:0000:A2AC:D523`
- bytes 7-11 全清零；MAC 保留；行不删

## 当前 schema 阻碍

1. **PK = spatial_addr**：UPDATE PK 允许但要小心 FK（device_ota 有 CASCADE FK，会跟随）
2. **Trigger `devices_check_within_unit`** 要求 `spatial_addr << units.spatial_prefix`：清零后落在 `tenant:branch:0:0::/80` 不属于任何 unit → 触发器拒
3. **Check `masklen(spatial_addr)=128`**：保留

## 落地方案（未实施）

- 每个 tenant/branch 自动建 sentinel 哨兵 unit：`fd00:owl:tenant:branch:0:0::/80` (unit_name="_unbound_pool")，让重置后的 device 落在这里
- 或：放宽 `devices_check_within_unit` 触发器为"<< 任意 layout 节点（branch /56 也行）"
- 新 RPC `POST /admin/api/v1/devices/{spatial_addr}/unbind?scope=unit|room|bed`
- 权限校验：调用者角色对应可清零的字节段必须 ⊇ 请求 scope

## 当前临时方案

Schema 没改之前，"清空所有 device 绑定"= DELETE FROM devices（21 行 → 0）+ device_factory_meta 保留；schema migration 完成后，所有此类操作走 UPDATE。
