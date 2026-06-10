---
name: visitor-belongs-to-cardagg
description: 2026-05-19 v2 双路径全部落地（commit ac9f64d）：bed-bound radar 解锁 Share unit visitor；room 兜底 Private；NumberPeople 走 iot:event:stream
metadata: 
  node_type: memory
  type: project
  originSessionId: 3803fdc4-b793-4d26-9019-59b8c61adfa9
---

## 架构原则（防再次跑偏）

**card_id 不是物理实体地址**，是 1 unit + x rooms + y beds + devices 的**组合视图**。Sensor 永远不跨物理实体合并（一旦合并细节丢失），合并是 cardagg 职责。

**unit / room / bed / device** 都是物理实体；**card** 是 UI/展示逻辑层概念，零物理存在。

## Visitor 职责分工

| 层 | 职责 |
|---|---|
| **Sensor** | 不做 visitor 累加；发 RoomState (/88 total_people) + monitor 流 (per radar number_people) 喂给 cardagg |
| **Cardagg `VisitorDeriver`** | 60s tick + 双路径判定（bed level 优先 / room level 兜底）+ 5min 阈值 |
| **FE** | dumb render card 的 target.visitor_* 字段 |

## 双路径矩阵（2026-05-19 升级）

| unit_type | room level (/88 room card) | bed level (/96 bed card + bed-bound radar) | 净结果 |
|---|---|---|---|
| Private (1) | ✅ 兜底 | ✅ 优先（写 /96 bed card） | 双路径 |
| **Share (2)** | ❌（多 resident 常态）| ✅ 唯一路径（解锁 share）| **bed level 解锁** |
| Public (3) | ❌ | ❌ | 跳过 |

### 关键升级：bed-bound radar 路径

**物理基础**：
- bed-bound radar = device IPv6 绑到 /96 bed prefix 的 Radar 设备
- firmware 自带 boundary 物理裁剪视野到该床区域（雷达自身不知道，由部署 / 配置决定）
- share 双床各自 boundary 设到自己床区——雷达看到 2 人=床主+访客（无歧义）

**cardagg 判定**：`card.CardType == "bed" AND card.Devices` 含 `device_type == "Radar"`

### 已接受的边界限制

| 边界 | 处理 |
|---|---|
| 视野中央盲区漏报（visitor 站两床之间）| 接受；5min 阈值过滤短时，visitor 长留必走床边 |
| segment 不连续（短离开又回来）| 严格 5min 连续；不做间断容忍 |
| 路过他床 / 巡房 | 5min 阈值天然过滤 |

### 优先级冲突

bed level 触发了 /96 bed card visitor → 对应 /88 父 room card 本轮跳过（避免父子双显）

## 显示位置

**显示在触发的卡上**（card target）：
- bed level 触发 → /96 bed card target
- room level 触发 → /88 room card target
- 每张卡独立 tracking 自己的 visitor 状态

不显示在 /80 unit card——unit picker 自己从子卡派生 display。

## 算法（cardagg VisitorDeriver 60s tick）

```python
for each tracked card:
    if cardType == "bed" AND card.Devices 含 Radar:
        # 路径 A: bed-bound radar
        peopleCount = bedPeopleCountTracker.get(card.id)  # cardagg monitor handler 维护
    elif cardType == "room" AND parent unit_type == Private:
        # 路径 B: room level
        peopleCount = redis.HGET(f"card:state:{card.id}", "room_state.total_people")
    else:
        continue  # Share /88 / Public 跳过
    
    if peopleCount >= 2:
        if segment_start_ts == 0:
            segment_start_ts = now
        segment_duration_min = (now - segment_start_ts) / 60_000
        
        if segment_duration_min >= 5:                              # ⭐ 5min 阈值
            target.visitor_start_ts = segment_start_ts
            target.has_visitor_today = true
            target.today_max_visitor_min = max(prev, segment_duration_min)
    else:
        segment_start_ts = 0  # 段结束

# 冲突处理: bed level 触发 → 同轮父 /88 room card 跳过
# midnight reset (parent unit timezone): 清三字段 + 重置 segment
```

## bed level 数据流（新增 2026-05-19）

```
radar monitor frame (含 number_people)
    ↓ iot:monitor:stream
cardagg monitor handler
    ↓ 检查 device 是否 bed-bound（card.CardType=="bed" + Radar device）
cardagg per-bed PeopleCount in-memory tracker (含 UpdatedAt)
    ↓
VisitorDeriver 60s tick 读
    ↓
写 /96 bed card target.visitor_*
```

## Sensor 端职责（极简）

- 维持现有 RoomState publish（/88 total_people）
- 维持现有 monitor 流（per radar number_people）
- **不参与 visitor 累加** —— 数据 owner 全在 cardagg

## 相关 memory 关联

- [[target_state_weak_bio_signal_design]] — weakBio 是单实体内事实判断，仍在 sensor
- [[cardagg_sensor_responsibility_split]] — 老 memory 需修正"sensor 推导 + 时间窗"措辞，应该是"单实体内"推导
- [[card_display_projector_handoff]] — 老 §Task 2.4 "sensor visitor 累加器" 作废
- [[p4_next_session_prompt]] — Visitor-v2 实施任务清单

## 实施 checkpoint

- 2026-05-18: doc/card_display.md §4.4 初稿（仅 Private /88 路径）
- 2026-05-19: cardagg `visitor_deriver.go` room 路径 P4-C 已落地 commit `06fb20b`
- 2026-05-19: **方案升级——加 bed-bound radar 路径（Share 解锁）**；doc/§4.4 修订 commit `487cf12`
- 2026-05-19: **Visitor-v2 实施完成 commit `ac9f64d`**
  - `CardMeta.HasBedBoundRadar()` + `DeviceMetaCache.ListBedCardsWithBedBoundRadar()` SQL
  - `BedPeopleTracker` (service) — per-device snapshot + offline 过滤 + out-of-order 丢弃
  - `EventHandler` (consumer) — 订阅 iot:event:stream filter category=number_people
  - stream_subscriber.go + main.go wire 完整
  - `VisitorDeriver` 双路径 tick：bed 优先 + parent room 跳过 + Private room 兜底
  - 修正 doc §4.4 数据流（number_people 走 event 流不是 monitor 流）
- **关键事实修正**：NumberPeople 是 firmware type=3 event（qinglan radar_decoder.go:578 `buildPeopleNumber`），走 `iot:event:stream` 不是 monitor 流；监控盲区 / 5min 阈值 / 静态人数不变都靠 event-driven 特性自然处理，不加时间窗 staleness
- **遗留 followup**（不本 PR 解决）：
  - FU6（中）VisitorDeriver 直接写 hash —— 当前仅 inject TargetMerger，stale 1 tick 等下次 target.state 触发才落
  - FU11（低）VisitorDeriver 单测 —— metaCache/reader 接口化 + mock
