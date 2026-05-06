---
name: AI publish 二期 + 三期 backlog（任务 A 完成；任务 B/C 待做）
description: 二期主体 + B1/B2/B3 + 三期任务 A（wire 三层冗余清理）已完成；三期任务 B（cardagg ↔ AI 对接梳理）+ C（owlfront 展示）待做
type: project
originSessionId: 48426fae-acda-4d8d-ae6b-946fa83fcc2e
---

## 已完成（feat/radar-fall-verifier）

| Commit | 主题 |
|---|---|
| 2c6b730 | PR1-5c 一期主体（AI publish toggle + Radar.AI<n> + track_verdict + alarm.Fall）|
| dbfc642 | doc/AI_fall_detect.md §17 + lost_fall ExitDistMinCm 100→30cm |
| 2bf9dec | bedside_fall vs lost_fall 双报 dedup（BedsideFallReported flag）|
| ae8ee38 | PR6 cardagg aiOverrides cache + sandbox/release 双模式 |
| 2d6cef2 | refactor: 撤 device_type AI 后缀方案（删 BaseDeviceType + 4 处 cardagg 调用）|
| 6621f20 | fix+feat: wisefido-ai 稳定化（3 critical bug + wire schema + config simplify + 工具）|
| **f1e94cf** | **refactor(wire) Stage 1a**: drop event_name from iot stream payload (9 文件 +25/-25) |
| **1cc2a5d** | **refactor(wire) Stage 1b + 2**: drop dataCategory + EventItem schema cleanup (15 文件 +89/-156) |

### 三期任务 A 协议简化（2026-05-01 完成）

iot stream 三层冗余事件类型表达 → 单一权威：

```
Pre:    m.Category + dataValue[i].dataCategory + dataValue[i].event_name   (3 重)
Post:   m.Category                                                          (1 重)
```

**三阶段执行序列**：
1. Stage 1a：producer 边界 strip event_name + consumer 切 m.Category
2. Stage 1b：producer 边界 strip dataCategory + 3 个 parser 接 envelope 参数
3. Stage 2：observation.EventItem 删 DataCategory/EventName 字段 + 19 处 populate 站点清理

**执行哲学**：consumer 先切（m.Category 优先 + payload fallback），producer 后砍。
helper 边界 strip 让 producer 调用代码 0 改动（除 wisefido-ai engine.go 直构旁路）。

**三次端到端 Fall 实证**：
- 12:41 (Stage 1a): trigger_data `{event_since, dataCategory:"Fall", event_status, event_payload}` 删 event_name
- 13:03 (Stage 1b): trigger_data `{event_since, event_status, event_payload}` 再删 dataCategory
- 13:35 (Stage 2):  trigger_data 同 1b（schema 已干净）

`event_payload` 内 firmware 原始 JSON 字符串里的 `event_name` 是 qinglan 固件不透明审计字段，
不属于协议层管辖（不剥离）。

### 三大 bug 实证修复（2026-04-30 + 2026-05-01）

| Bug | 后果 | 修复 |
|---|---|---|
| B1 主循环不 XAck | 4.86 天积压 1.51M pending | engine.go 三处 loop 加 XAck + 一次性清理脚本 `scripts/clean-roomengine-pending.sh` |
| B2 路由静默 + 启动后绑定不可恢复 | 启动后才绑 device 永远沉默 | tm==nil warn + 60s 路由热加载（reloader 闭包注入）|
| **B3 (root cause)** msg.Values["data"] vs 实际 flat key | engine 长期"接收但不处理" | 三个 handler 全改用 `rediscommon.FromStreamMap` |

## 核心架构原则

1. **宁可误报不可漏报**：alarm 路径不被 AI verdict 压制；mode 仅控制 publish/UI 合并。
2. **observation.Track 是公共 schema**：AI 写"事后裁决"用同一份 Track 形态，不污染上游。
3. **Source 是一等公民**：`fields["source"]` 表达数据生产者节点身份（"AI.Caregiver01" 等）；空 = firmware 直传或非 AI 派生。`reason` 是机器可分类的决策路径，`evidence.context` 是自然语言审计文本。
4. ~~device_type 后缀 Radar.AI01~~ **已废弃**：device_type 始终为源 sensor 物理类型；AI 派生身份归 source 字段。
5. **wire 永远 flat**：data_value[0] 是 flat object，新字段加 sibling key（不嵌套），向后兼容。
6. **cardagg 端"宁可不接"**：sandbox 默认开，release 才生效——担心 AI 判断质量时不破坏现有 UI。
7. **消费端必须 XAck**：read 后无 ACK 让 redis pending 无限增长（B1 教训）。所有 stream subscriber 必须 ACK。
8. **stream 消息解析用 `redis.FromStreamMap`**：消息是 flat key 不是 "data" 单 key 嵌套（B3 教训）。
9. **envelope/payload 分层（任务 A 沉淀）**：
   - **envelope**（IoTStreamMessage flat keys）= 路由 + 寻址 + 粗分类（device_uid / card_id / tenant_id / topic_type / **category** / device_type / timestamp）
   - **payload**（dataValue[i]）= 纯业务字段（position / pose / evidence / event_payload / event_since / event_status 等）
   - envelope.Category 是事件类型唯一权威；payload 不重复 envelope 字段。
   - 多 item 合发场景如未来回归，新增 `kind` 字段（不要叫 dataCategory），envelope.category="mixed" 派生——不为假设场景预留化石。
10. **EventNameKey / DataCategoryKey 常量保留**：作历史读路径兼容（DB 老 trigger_data 含这些 key），producer 永不再写。

## 三期任务 B：cardagg ↔ AI 对接全梳理（✅ 已实证 2026-05-06）

**实证状态（2026-05-06 实查）**：
- 系统每日多次 ai_emit（D523 + D5F7 等多设备产 verdict）
- 08:47:50 D5F7 Fall verdict（topic_type=alarm，published=true）走通 alarm 链路
- 08:39 / 08:42 / 08:46 / 08:49 D523 track_verdict ghost emit（confidence=20/30）
- B1（XAck 修复）+ B2（路由热加载 + tm==nil warn）+ B3（FromStreamMap）线上行为正常

## 三期任务 B 历史梳理（2026-05-01）

**目标**：确认 cardagg 完整消费 AI emit + alarm + 前端展示路径。

**当前状态**：
- cardagg PR6 ai_overrides cache 已就位（`event_handler case "track_verdict"`）
- cardagg sandbox/release 双模式（默认 sandbox，Apply 不动 fields）
- alarm 路径独立，AI 不 gate 任何 alarm fire（"宁可误报不可漏报"原则）
- 任务 A 完成后 wire 协议已干净，可以专注业务逻辑验证

### B1 + B2 实证准备（2026-05-01 完成代码梳理 + 文档 + 脚本）

**已交付**：
- Wire 格式契约：`doc/AI_Iot_stream.md`（envelope + ghost/fall payload + §4 五条 Schema 设计决定 pinned）
- Fall 链路代码梳理：`alarm_handler.go:106` → `alarm_service.go:227 PersistAlarmAndPublish` → trigger_data 完整存 dataValue[0]
- Ghost 链路代码梳理：`event_handler.go:154` Set → `monitor_handler.go:152` Apply（release 才覆写 fields）
- 验证脚本：`scripts/verify-ai-cardagg-link.sh`（4 tap 点 + 一致性诊断）

**Schema 设计 pinned**（不要再讨论，详 doc/AI_Iot_stream.md §4）：
- track_confidence (=100-GhostPenalty) 与 evidence.score (=ts.Score) 是双轨，不重命名
- bed_status 不配 confidence 是有意（cardagg.BedConfidence 60/90/100 兜底融合可信度）
- evidence 自由 schema（按 reason 派生 keys）；reason 不加 fall:/ghost: 前缀
- area_type 在 ghost 路径可能缺失（grid 未命中时不填），下游必须容忍

**Fall 路径 3 挂点**（drop 时看哪个）：① 30s stale（时钟漂移）② ResolveDeviceID 失败 ③ enablement 未配（**静默 if 不进，无 drop log**）

**Ghost 路径 5 挂点**：① event 30s stale ② ResolveDeviceID ③ **monitor 6s stale**（更严）④ EnterRoom/ExitRoom 误清缓存 ⑤ **sandbox 默认开启，UI 看不到效果——必须切 release**

**待做（等人到现场）**：
- 远程跑 `verify-ai-cardagg-link.sh` 拿 baseline 数据
- 现场触发 fall（lost_track 最易复现）+ SQL `SELECT trigger_data FROM alarm_events WHERE source LIKE 'AI.%'`
- 现场触发 ghost：人走动，看 release 模式 monitor 流的 track_confidence 是否被覆写到 ≤30
- 根据日志数据决定 5 挂点是否需要新 PR 修

3. ~~**AI override TTL 刷新策略（PR7b）**~~ **已撤销，无需做**（2026-05-01 纠正）：
   - 之前误以为"AI 只首次 emit + 60s 后 ghost 复活"是 bug，建议 PR7b 让 AI 周
     期重发——**全部基于错误代码理解**。
   - 实际代码（`track_manager.go:968-993`）：`tm.emitGhostVerdict` 在 GhostPenalty
     ≥ 阈值时**每帧都 emit**；`LoggedGhost` flag 只控制**本地结构化 log 去重**，
     完全不挡 emit。
   - cardagg `ai_overrides.go` Set 时整条覆盖含 UpdatedMs，每次都刷新 TTL；
     GC 只删 60s 内未刷新的条目。
   - 实际行为是健康 fail-safe：ghost 持续存在 → 每帧刷新 → 永不过期；ghost 消失
     → 60s 后自动 GC → **真人误判最多被压制 60s**（这是有意设计，不是缺陷）。
   - **教训**：以后涉及"是不是只 emit 一次"这类怀疑，先 grep 实际 emit 路径而不
     是看 flag 命名瞎猜。

4. ~~**LogicID 由 AI 分配**~~ **推迟到 3.0 版本**（2026-05-01 决定）：
   - 当前 `Track.LogicID` 字段在 owl-common 定义但全代码库 0 模块读写——保留 schema 不动
   - 3.0 再做：AI 在 emit verdict 时填 logic_id；cardagg 落 LogicID→physical track 映射；
     monitor 流回写让前端 group by 同一逻辑目标
   - 涉及 AI engine track 关联 + cardagg state_service + 新 stream `iot:logic:state`

## 三期任务 C：owlfront 前端展示

### C1 三档饱和度渲染（✅ 已部署 2026-05-06 确认）

**部署状态确认（2026-05-06 实查）**：
- qinglan/internal/consumer/mqtt_consumer.go:712,876 TrackConfidence=80 ✅
- owlFront/src/components/Radar/RadarCanvas.vue:195-199 `CONFIDENCE_DROP_MAX=20 / CONFIDENCE_HIGH_MIN=80 / FALL_POSTURES`，isLowConf 逻辑全部落地 ✅
- 当日 sensor 多次 ai_emit 含 Fall 验证（如 08:47:50 D5F7 Fall track_confidence=0）

### C1 历史设计（2026-05-01）

**简化决定**：原 5 项待动作（角标/tooltip/详情页/演示/多角色）砍到只做"confidence × 饱和度"
一项；其它若有需求再单独排。

**最终三档矩阵**（含 fall 例外）：

| confidence | 非 fall posture | fall 类 posture |
|---|---|---|
| ≤20  | **不画**（drop） | **半饱和**（强制显示） |
| 21-79 | 半饱和 + alpha 0.65 + 灰标签 | 同左 |
| ≥80   | 全饱和默认 | 全饱和默认 |

**Fall 例外原因**：宁可误报不可漏报。若 AI 把 conf 写到 ≤20 但 firmware/AI 同时
报 fall 姿态，UI 必须显示位置否则运营看不到人。

**实施改动**：
- `wisefido-qinglan/internal/consumer/mqtt_consumer.go` L766+L930: TrackConfidence 60→80
  （让 firmware 默认落 ≥80 全饱和档；qinglan 注释自承"无 signal_quality 所以填默认"，
  60 是任意占位符；改 80 不损失语义）
- `owlFront/src/components/Radar/RadarCanvas.vue`：~12 处编辑
  - 常量 `CONFIDENCE_DROP_MAX=20` / `CONFIDENCE_HIGH_MIN=80` / `FALL_POSTURES[]`
    （含 4 个：FallSuspect/FallConfirm/SitGroundSuspect/SitGroundConfirm）
  - `isGhost: boolean` 全部重命名 `isLowConf: boolean`（语义从"二元 ghost"→"中等置信"）
  - drawPersons forEach: drop early-return + fall 例外
  - 渲染从 `grayscale(80%) + alpha 0.35` → `saturate(50%) + alpha 0.65`
    （"评估中"语义比"灰化"更准）
  - 轨迹线保留调色板色（不再统一灰）；标签去掉 `?` 后缀（颜色/alpha 已传达低置信信号）

**firmware 默认值现状（已对齐三档）**：
- qinglan track_confidence: 80（改后）
- sleepad track_confidence/pose_confidence: 90（不动，已在 ≥80 档）
- AI 派生 ghost: 100 - GhostPenalty（emit 触发 GhostPenalty ≥ 80，所以 ≤ 20）

**已知二阶现象**（不影响 C1 上线）：21-79 半饱和档当前 wire 上**永远不会出现**——
firmware 给 80（qinglan）/90（sleepad），AI 只在 GhostPenalty ≥ 80 时 emit（覆写
到 ≤20），中间值无来源。中档代码作为未来扩展位保留（如未来产品需求出现"评估中
半饱和"概念再用），现状下视觉效果实际是"≤20 不画 / ≥80 全饱和"二元（fall 例外
保护层不变）。**不要因此回头改 PR7b 想点亮中档**——见 B3 教训。

**待做**：
- 部署 wisefido-qinglan 重编译（否则线上 firmware 还是发 60，前端把所有雷达 track
  渲染成半饱和）
- 前端 dev server 测试 fall 例外（mock posture=FallConfirm + track_confidence=20，
  验证人形不消失而是半饱和）

### C2 后续待排（暂不做）

原任务 C 的另外 4 项（AI 角标/tooltip / Fall 详情页 trigger_data 展示 / sandbox
演示模式 / AI.Doctor 扩展位）已从 backlog 移除——3.0 或有产品需求时再排。

**owlfront 仓库路径**：`/home/wisefido/owl/owlFront/`（已确认）

## 不要再做的事（避免来回讨论）

- 不要把 verdict 名字（"ghost"/"real"）作为 wire 字段——下游按 confidence 数值阈值自己派生
- 不要给 observation.Track schema 加 sandbox/release / source / reason 等审计字段——它们都是 wire flat sibling，不是观测属性
- 不要嵌套 data_value 结构——破坏现有 flat parser
- 不要给 alarm 路径加 AI override gate——违反"宁可误报不可漏报"
- 不要恢复 device_type AI 后缀方案——已废弃，与 source 字段二选一（架构原则 3+4）
- 不要新写不带 XAck 的 stream consumer——必须 ACK（架构原则 7）
- 不要用 `msg.Values["data"]` 解析 stream 消息——用 `redis.FromStreamMap`（架构原则 8）
- 不要在 owl-common 集中定义 Source/Reason 常量——目前 wisefido-ai 是唯一 producer，本地 const 即可；出现第二 producer 再上移
- 不要把 NodeID 与 Source 拼字符串——config.yaml `source: "AI.Caregiver01"` 直传
- **不要在 dataValue payload 内重复 envelope 字段**（任务 A 沉淀）——envelope.Category 是唯一权威；想加新字段先问"是路由还是业务"
- **不要重新引入 V2 风格 m.Category 集合字符串**（如 "track.vital"）——拆条 producer 是 V4 策略；多 item 真回归用新 `kind` 字段而非复活 dataCategory
- **不要砍 EventNameKey/DataCategoryKey 常量**——保留作历史 DB trigger_data 读路径兼容；producer 永不再写
