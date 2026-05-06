---
name: Track Fusion & Gate Card-ID Architecture
description: Radar-Sleepace track融合方案 + device-gate填card_id架构重构，含5轮实测数据验证
type: project
originSessionId: 35a944e5-1093-45ec-b924-44491ffec0db
---
## 背景

2026-04-16 实测发现 Canvas 上床区出现重复人形（P0+P1），通过5轮实测定位根因。

## 实测结论

- Sleepace 不发 position_x/y，前端 fallback 到 (0,0) 画出假人
- Radar track 分裂：人躺下后雷达检测出2个目标（非金属杆反射）
- 缩小雷达边界（L300→L200, F200→F100）后分裂消失
- 最终安全边界：L200 R200 F100 B170（测试房间 Denver-LakeW-101）

## alarm_events 问题

- BM87224601897 (Sleepad_1897) 的 LeftBed alarm triggered→created 延迟45分钟
- 根因：cardagg 频繁重启 + stream 积压 + card_id 为空导致14个临时card_id
- iot_timeseries 中同一 LeftBed 写了3条（sleepace realtime + inBedStatus + cardagg alarm 三路）

## 架构决策

### 核心原则
- data 是 card 绑定关系的唯一权威源
- device-gate 负责贴 card_id 标签（从 data 查询 device_uid→card_id 映射，内存 map）
- cardagg 只验证 card_id，不解析/不猜测，card_id="" 直接 drop + error
- 未来 Edge gate (ESP32) 同样逻辑，baseline 存本地 flash

### 分层
- device-gate：协议转换 + 贴 card_id 标签（只查 map，不理解业务语义）
- cardagg：业务聚合（融合、告警、状态推导），启动时从 data 加载 card 白名单
- unit/room/bed 绑定关系是业务语义，由 cardagg 从 baseline 解析

### data 重启
- XTRIM 所有 stream 到当前时间点
- 发 configCard reset 通知
- 所有服务收到 reset 后回 data 重查映射

### 消息规范
- 所有 iot:*:stream 消息必须带 card_id
- 未绑卡设备 card_id = device_id

**Why:** cardagg 重启时 baseline 不稳定导致45分钟延迟和幽灵 card_id，需要从根本上解决 card_id 来源问题。

**How to apply:** 后续实施时按此架构执行，不要让 cardagg 自行解析 card_id。
