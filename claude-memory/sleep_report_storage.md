---
name: Sleep Report 存储位置
description: sleep report 在 sleepace 服务自身 DB，key=user_id==device_id(UUID)，跨 owl 主库 rebuild 不需要迁移
type: reference
originSessionId: 53d81dc2-b99e-4371-95da-247c2cf5c46c
---
**Sleep report 不在 owl 主库**，存在 sleepace 服务自己的 DB 里。

- 索引 key = `user_id`
- sleepace 体系内：`user_id == device_id` (UUID, == logMAC)
- device_id 跨 owl spatial schema rebuild 保持不变（stateless 派生）
- owl 主库 drop+recreate 不影响 sleep report 可查询性

**查询链路：**
```
spatial_addr (新 IPv6)
  → device_id (UUID, 稳定)
  → sleepace.user_id
  → sleep_report
```

**影响：**
- 任何 owl 主库迁移/rebuild 计划讨论中，sleep report 都**不在迁移范围**
- wisefido-data 侧若要适配新 schema 只是查询路径多一步 lookup，不是数据迁移
- 不要再提议"sleep report 要迁移外键"——data ownership 在 sleepace 不在 owl
