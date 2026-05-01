# AI 派生 iot stream 报文格式（Ghost / Fall）

wisefido-ai 通过 `iot:event:stream` 与 `iot:alarm:stream` 发布"事后裁决"
（ghost）与"AI 派生 alarm"（fall）。本文档 freeze wire 格式 + 字段语义，
作为下游消费方（cardagg / owlfront / 其它 future consumer）的契约。

任何 schema 变更先改本文档，再改代码；评审 PR 时以本文档为基准。

---

## 1. envelope 公共部分

由 `IoTStreamMessage.ToStreamMap()` 写出，所有 AI 派生消息共享同一组 flat keys：

| key          | 值                                  | 备注 |
|--------------|-------------------------------------|------|
| `device_uid` | 源 sensor UID（如 `"D5F7"`）         | 业务主键，反查源设备 |
| `device_id`  | 源 sensor UUID                       | log/debug 用，可空 |
| `device_type`| `"Radar"` / `"Sleepad"`              | **不带 `.AI` 后缀**；AI 派生身份由 payload.source 表达 |
| `card_id`    | `roomCards[roomID]`                  | 可空 |
| `tenant_id`  | `roomTenants[roomID]`                |  |
| `timestamp`  | `"1746086400000"`                    | string-formatted ms（XADD 协议要求） |
| `topic_type` | `"event"` / `"alarm"`                | 决定走哪条 stream + 下游分发 |
| `category`   | `"track_verdict"` / `"Fall"` / ...   | ★ envelope 唯一权威类型字段 |
| `dataValue`  | `<JSON 字符串>`                      | 单 element array：`[{...}]` |

**envelope 不重复 payload 字段**，反之亦然（架构原则 #9：envelope 路由+寻址+粗
分类，payload 纯业务）。

---

## 2. Ghost — `iot:event:stream`

**envelope.category = `"track_verdict"`，topic_type = `"event"`**

dataValue payload：

```json
[
  {
    "track_id": 3,
    "position_x": 245, "position_y": 180, "position_z": 95,
    "pose": 1,
    "track_confidence": 20,        // 100 - GhostPenalty (≤30 → ghost UI)
    "ts": 1746086400000,
    "area_type": "monitor_bed",    // engine 自算，仅 grid 命中时才有
    "source": "AI.Caregiver01",    // = cfg.AIPublish.Source
    "reason": "ghost_penalty",     // ReasonGhostPenalty / GhostLowScore / GhostPostReal
    "evidence": {
      "score": 35,                 // ts.Score（独立维度，见 §4.1）
      "birth_score": 60,
      "ghost_penalty": 85,
      "frame_count": 42,
      "context": "ghost_penalty_accumulated"
    }
  }
]
```

注：ghost 路径 `bed_status` 默认 0 时被 `engine.go:570-572` `delete(fields,
"bed_status")` 删掉，wire 上不出现（与 fall 路径 D 形成对比，见 §4.2）。

---

## 3. Fall — `iot:alarm:stream`

**envelope.category = `"Fall"`（= `alarm.Fall` 常量），topic_type = `"alarm"`**

子类型由 payload.reason 区分（不污染 envelope）。`track_confidence` 不发
（fall 已确认，`payloadFromTrack` 显式置 0，`ToFieldMap` 跳过零值字段）。

### (A) lost_fall — track 失锁

```json
[{
  "track_id": 3, "position_x": 245, "position_y": 180, "position_z": 95,
  "ts": 1746086400000, "area_type": "custom",
  "source": "AI.Caregiver01",
  "reason": "lost_track",
  "evidence": {
    "context": "track_lost_no_exit_room_no_recovery",
    "fall_score": 78, "frozen_start_ms": 1746086380000,
    "spatial_jump": false, "cell_area_type": 1,
    "wait_ms": 18000, "last_verdict": 2
  }
}]
```

### (B) still_fall — 浴室长时间静止

```json
[{
  "track_id": 5, "position_x": 110, "position_y": 320, "position_z": 80,
  "pose": 1, "ts": "...", "area_type": "monitor_bed",
  "source": "AI.Caregiver01",
  "reason": "still_in_bathroom",
  "evidence": {
    "context": "still_in_bathroom_over_threshold",
    "still_seconds": 95, "still_timeout_sec": 60,
    "cell_area_type": 4
  }
}]
```

### (C) bedside_fall — R4 床边晕倒

```json
[{
  "track_id": 2, "position_x": 200, "position_y": 220, "position_z": 30,
  "pose": 0, "ts": "...", "area_type": "bed",
  "source": "AI.Caregiver01",
  "reason": "bedside_silent",
  "evidence": {
    "context": "bedside_still_after_leftbed",
    "still_seconds": 920, "window_sec": 1800,
    "still_timeout_sec": 900, "bedside_margin_cm": 50,
    "leftbed_at_ms": 1746085500000
  }
}]
```

### (D) silent_leftbed_fall — sleepad LeftBed + radar 仍在床

```json
[{
  "track_id": 0,                      // pickActiveTrackNearBed 选床附近 track
  "position_x": 195, "position_y": 215, "position_z": 28,
  "bed_status": 1,                    // 仅此路径保留（见 §4.2）
  "ts": "...", "area_type": "bed",
  "source": "AI.Caregiver01",
  "reason": "sleepad_radar_conflict",
  "evidence": {
    "context": "sleepad_leftbed_radar_still_on_bed",
    "fall_score": 65, "radar_verdict": 1,
    "sleepad_uid": "SLP-XX", "had_hr_rr": true, "max_people": 1
  }
}]
```

---

## 4. Schema 设计决定（pinned，不要再讨论）

### 4.1 `track_confidence` vs `evidence.score` 是双轨，不是同一个量

| 字段 | 来源 | 含义 | 用途 |
|---|---|---|---|
| `track_confidence` | `100 - ts.GhostPenalty`（`engine.go:310`）| AI 对 track 的整体信任度 | **UI 渲染**：前端按数值阈值染色饱满度（≤30 灰阶/30-79 半饱和/≥80 全饱和）|
| `evidence.score` | `ts.Score`（独立累积，birth_score 起点 + 每帧加减） | track 真实性历史评分 | **审计追溯**：和 birth_score / ghost_penalty 一起说明"为什么判 ghost" |

举例：`track_confidence=20` 同时 `evidence.score=35` 是合理的——
- birth 给 60 分起步，老化降到 35（仍 > `ScoreGhostTh`，没靠 score 路径）
- `GhostPenalty` 累积到 85（≥ `GhostPenaltyThreshold=80`），主路径触发 ghost
- payload 把派生信号 `track_confidence=20` 暴露给 UI；evidence 保留三个原始量
  （score / birth_score / ghost_penalty）回答审计问题

**不重命名 evidence.score** —— 改名会回归 fall-score-replay 工具 + 所有 case
fixture，价值低。文档对齐即可。

### 4.2 `bed_status` 不配 `bedstatus_confidence` 是有意的

observation.Track schema 中其它"状态+置信度"成对：
- `pose` ↔ `pose_confidence`
- `track_id/position` ↔ `track_confidence`
- `heart_rate/respiratory_rate` ↔ `vital_confidence`

但 `bed_status` 没有配套 confidence 字段，原因：
- **firmware/sleepad 上报时本来就不带**——sensor 自己也没有"在床置信度"概念
- **cardagg 聚合层兜底**：`owl-common/card/card_types.go:244` 的 `BedConfidence`
  字段表达多源融合可信度（`60=雷达单源 / 90=Sleepad 单源 / 100=双源一致`）。
  前端要看在床置信度，读 cardagg 输出即可

**只有 fall (D) 路径保留 `bed_status=1`**，因为它的语义就是"sleepad 报 LeftBed
但 radar 还说在床"——`bed_status=1` 表达的是 sleepad 视角的离床事件，不是 AI
对在床状态的整体判断，无需配 confidence。

如果未来真有"基于 bed_status 单源置信度"的渲染需求，不在 wire 加新字段，而是
在 cardagg 那侧扩展 `BedConfidence` 的语义维度。

### 4.3 `area_type` 在 ghost 路径可能缺失

engine 用 `g.CellAt(px, py)` 算 area_type，仅 grid 已注册且坐标在 grid 范围内
才填。下游必须容忍此字段缺失（cardagg 端目前不读 ghost 路径的 area_type）。

### 4.4 `evidence` 是自由 schema，按 reason 派生 keys

每个 reason 对应一组特征 evidence keys（fall_score 仅 lost/silent_leftbed 有，
still_seconds 仅 still/bedside 有等），不强制统一。`evidence.context` 是所有路
径共有的自然语言审计串，下游展示用。

### 4.5 `reason` 不加 `fall:` / `ghost:` 前缀

类型已由 envelope.category 区分（`"Fall"` vs `"track_verdict"`），reason 加前
缀是冗余且违反原则 #9。reason 命名空间靠"在哪个 category 下出现"上下文化。

---

## 5. 速查表

| 维度 | Ghost | Fall |
|---|---|---|
| stream | `iot:event:stream` | `iot:alarm:stream` |
| envelope.category | `track_verdict` | `Fall` (= `alarm.Fall`) |
| envelope.topic_type | `event` | `alarm` |
| 子类型表达 | `reason ∈ {ghost_penalty, ghost_low_score, ghost_post_real}` | `reason ∈ {lost_track, still_in_bathroom, bedside_silent, sleepad_radar_conflict}` |
| `track_confidence` | 必填 0-100（ghost 路径自然 ≤30） | **不发**（fall 已确认） |
| `bed_status` | 不发（`engine.go:570-572` delete） | 仅 (D) 路径保留 = 1 |
| `evidence` 必有 keys | `score / birth_score / ghost_penalty / frame_count / context` | `context` + 各 reason 派生 keys |
| dataValue 形态 | 单 element `[{...}]` | 单 element `[{...}]` |
