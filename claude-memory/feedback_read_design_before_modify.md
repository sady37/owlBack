---
name: feedback-read-design-before-modify
description: 改代码前先读设计/实施文档对齐业务逻辑；改完复查是否符合业务逻辑
metadata: 
  node_type: memory
  type: feedback
  originSessionId: d23e371f-aa78-4087-ba7b-63ba1a80ec04
---

修改任何业务代码前必走：

1. **改前**：先读对应的设计文档 / 实施 doc（owlBack/doc/, /home/wisefido/owl/owlRD/, memory/* 等），把"这块代码的业务逻辑是什么"理清楚
2. **改时**：每一处改动对照业务逻辑
3. **改后**：复查 — 我的改动有没有破坏其它业务逻辑？特别是隐含的不变量

**Why**：

2026-05-20 实测教训：[[bed_presence_fusion]] Bug 2 dedup 添加 `applyRoomDedupInPlace` 用 `cache.GetZ()` 覆写 `rs.TotalPeople`。
当时没复查 `applyTransitionToState` 的 Vacant 路径行为（engine 显 set `Count=0`），导致 `OnZoneEvent(TransitionVacant)`
时 cache.Z（radar 最后 NumberPeople）反向覆写 engine 的 ground truth 0 → bathroom 离场后永远卡 InBath（实测
f61e8o 6min stale）。如果改前读过 `engine.go applyTransitionToState` + `translator.go TranslateRoomState`，
立即看出 Vacant case Count=0 是权威，dedup 不能 trump 它。

**How to apply**：

- 改 sensor zoneengine / 任何 state machine：先读 `owlBack/wisefido-sensor/internal/zoneengine/types.go` 三态状态机
  设计 + `engine.go applyTransitionToState` 字段更新规则
- 改 cardagg display builder：先读 `owlBack/doc/card_display.md` spec
- 改 alarm 决策路径：先读 `owl-common/alarm/*` + 相关 memory（[[feedback_no_dynamic_threshold_modulation]] 铁律）
- 改 API ID 形式：先读 [[feedback_api_ids_ipv6_only]] + [[feedback_ipv6_prefix_set_membership]]
- 改完跨业务路径（如 dedup 影响多种 transition）：枚举所有 transition / 边界场景，逐个走一遍逻辑

**变体 — "孤儿 publisher / 0-callsite interface 不是 bug"**（2026-05-24 zonealarm 重构教训）：

发现 `PublishAlarmDeviceMessage` interface 已声明 + impl 已存在但 0 callsite 时，**不要直觉地"补上 callsite"**。先做 reader 全貌检查：

```
grep -rn '<config_key>\|<stream_name>' --include='*.go'
```

枚举所有读者，确认这个 publish 真该被消费。

zonealarm 案例：cloud_config 0 callsite 不是漏 wire，是**snapshot model 设计**——cloud_config 只是 UI 模板（reset / new device bind 取 snapshot），sensor/cardagg 运行时只读 device_config /128 精确匹配，cloud_config 在运行时**没人订阅**。

误判后果（实际返工 4 处）：
- alarmCloudService 强行加 publisher 依赖 + interface method + 25 行 fan-out → 全删
- sensor enablementCache 加"device 优先 + cloud LPM fallback"复杂分支 → 简化为只读 device
- cardagg enablementCache 同样复杂化 → 同样简化
- main.go wiring 多 wire → 删

**规则**：写新代码补丁前，先 grep 现有 reader/writer 拓扑。0 callsite 是 signal（设计上不需要）而非 bug（漏 wire）。补 callsite 前必须证明"有 reader 等着用这个 publish"。
