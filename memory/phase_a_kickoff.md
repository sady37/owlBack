---
name: 协议层 4 Phase 进度跟踪
description: Phase A/B/C/D 的实施进度（A/B 完成，C P0 完成 2026-05-03，C P1/P2/P3 + D 待做）
type: project
originSessionId: 2540d0c0-e995-418d-a1a4-3e2d3075ad3a
---

[doc/TODO.md 第 0 项](../../../owl/owlBack/doc/TODO.md) 是完整设计；本文件是当前进度跟踪。

## Phase A — 协议层 v2 envelope 重构（✅ 完成）

- envelope schema：删 CardID，加 Producer/SubjectEntity/SemanticLocation/SequenceNumber → [owl-common/redis/message_types.go](../../../owl/owlBack/owl-common/redis/message_types.go) 已 first-class
- device:status:{deviceID} 拆出（替代 card:state.DeviceStatus）
- IsUUID 守卫 / mapRoomsToCards / StreamHeadCardID 残留清理
- qinglan / sleepace publisher 填 Producer + SubjectEntity；wisefido-sensor 填 Producer 留空 SubjectEntity
- DeviceMetaCache 反向索引 O(1) 反查

## Phase B — 命名规范化（✅ 完成 + 残留清理 2026-05-06）

- `wisefido-ai/` → `wisefido-sensor/`（目录 / systemd unit / 二进制全改）
- producer 命名约定：`sensor.caregiver01` / `device:UUID`
- 2026-05-06 清理残留：
  - `systemctl reset-failed owlback.ai.service` 清 ghost 状态（未来不再出现在 `--failed`）
  - stop-owlback-full.sh / stop-owlback.sh 列表中的 `owlback.ai` 删除
  - clean-roomengine-pending.sh 推荐流程注释 `owlback.ai` → `owlback.sensor`
  - verify-ai-cardagg-link.sh `journalctl -t owlback-ai` → `owlback-sensor`
  - scripts/systemd/owlback-run-service.sh 删除 `wisefido-ai` 兼容分支（过渡期结束）
  - wisefido-sensor/stop-sensor.sh 删除遗留 `wisefido-ai` 进程兼容 pgrep
  - .gitignore / .dockerignore 注释中的 wisefido-ai → wisefido-sensor
  - doc/ 多文件 wisefido-ai → wisefido-sensor 路径修正（保留 doc/TODO.md 历史背景描述不动）
- `wisefido-ai-health/` 仍叫 ai-health（layer 2 候选位，独立模块，不在 B 范围）

## Phase C — alarm 三件套 + dedup + Tags（部分完成）

- ✅ **P0 — DedupWhileActive + InsertAlarmAndUpdateCard 守卫**（2026-05-03 完成）
  - `AlarmDef.DedupWhileActive` 字段 + 5 设备类置 true（Offline/SignalPoor/AngleException/SensorDetached/DeviceFailure）
  - `AlarmInsertResult.Deduped` 字段 + `InsertAlarmAndUpdateCard` BeginTx 前守卫
  - 7 个 caller 全部加 `if result.Deduped { return nil }` 短路 push
  - 配合 Layer 1 (qinglan/sleepace 源头 transition dedup) 形成完整防御链
- ⏸ **P1+P2 — alarm_events `producer` + `sequence_number` 列 + `PersistAlarmAndPublish` 二件套签名**（2026-05-03 决策暂缓）
  - 实质：把 envelope 已 first-class 的 Producer / SequenceNumber 持久化到 DB，并让 InsertAlarmAndUpdateCard 接收
  - 暂缓理由：短期无实质损失；observability 痛点能从 trigger_data 反查 source 标签兜底；SequenceNumber 单调策略应和 Phase D session token 一起设计（D 阶段 producer 会从 `device:UUID` 换成 `xai:session:<token>`，现在做必重做）
  - 启动信号：cognitive agent 真接入产 alarm / 同毫秒 timestamp 撞号 / trigger_data 反查痛点持续
- ❌ **P3 — `Tags []Tag` 替代 `Category`**（wontfix；FHIR `GetFHIRCategory(eventName)` 静态映射已够用）

## Phase D — device 端 LITE 直发 + 隐私（远期）

- 触发条件：firmware protobuf nano + DTLS-PSK 协议栈成熟 / session token 颁发机制 / 新硬件批次
- 当前不安排时间，记录作为远期目标

## 一刀切原则（不向后兼容）

「不向后兼容，当前仍在开发测试阶段，如果一味向后兼容，永远到不了终点」（用户原话）。
旧字段直接删除，所有 publisher + consumer 同时改完。

## 部署红线

部署是修改共享系统，必须用户明确确认时机才动手。代码改完编译/测试通过 → 停下等用户拍板部署窗口。
