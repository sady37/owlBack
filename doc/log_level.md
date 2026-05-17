# owlBack 日志级别约定

zap 替换 stdlib log.Printf，输出 RFC3339 ms UTC + 结构化字段，机器/人都能读。

## 通用分级原则

| 级别 | 含义 | 例子 |
|---|---|---|
| **Fatal** | 服务无法启动/运行 | `connect db failed` 启动期 |
| **Error** | 业务流程中断，需立刻处理 | DB INSERT 失败、Redis publish 失败、解码崩溃 |
| **Warn** | 业务可继续但出了异常情况 | 单设备订阅失败、payload parse 失败、stale cache |
| **Info** | **真正变化的业务事件**（人需要看的） | 服务启动完成、auth 成功、状态跃迁、真实告警 |
| **Debug** | 周期性 state tick、调试细节（机器要看的） | monitor 高频流、未变化的健康检查、内部去重命中 |

**核心规则**：稳态周期性 tick 必须 Debug。Info 只记**变化** + **真业务事件**。频率以"60 设备 × 健康检查周期"为上限想象——若每条都 Info 就是 noise。

## 各服务具体记什么

### wisefido-qinglan

**Info**（人看：手动测试 + alarm pipeline 可见）
- 启动/关闭、auth request/response（auth_service.go）
- mqtt rx：`prop` topic（设备属性查询/设置应答）
- mqtt rx：`event`/`alarm` topic + firmware `type=1`（EnterRoom/ExitRoom/InBed/LeftBed）
- mqtt rx：`event`/`alarm` topic + firmware `type=2`（Fall/SittingOnGround，含 Suspected\*）
  - 带 `info` 字段白话摘要（如 `"SuspectedFall pose=2(suspected_fall) last_pose=1(walking) track=1"`）
- Device healthcheck **transitions**：Offline / SignalPoor / AngleAbnormal 的 fail ↔ recover 跃迁（`health_check.go publishHealthIfChanged "device health transition"`）

**Debug**（机器看：周期性 + 高频）
- mqtt rx：`monitor` topic（2Hz × N radar track 数据）
- mqtt rx：`stat` topic（周期性 sleep_trajectory）
- mqtt rx：`event`/`alarm` topic + firmware `type=3`（number_people 变化）
- mqtt rx：`event`/`alarm` topic + firmware `type=5/7/8`（device-status，由健康检查 transition 接管）
- mqtt rx：`func` topic（命令应答）

### wisefido-sleepace

**Info**
- 启动/关闭、redis/db/mqtt connected、订阅 ready
- `connectionStatus transition`：online↔offline 跃迁 + first_observation（重启冷启动一次）
- `device_store version synced`：仅版本真改变了才记（UpdateDeviceStoreReportedVersion 返 changed=true）
- `sleepStage` / `inBedStatus` / `bedStatus change`：Sleepad 推上来的真业务事件
- `deviceSenSor` transition：传感器脱落/插上跃迁

**Debug**
- `connectionStatus tick`：稳态健康检查 tick（值未变）
- `deviceSenSor dedup`：未变化重复推送被去重时

### wisefido-iot

**Info**
- 启动/关闭、流订阅 ready
- 单错误（per-message INSERT 失败）

**Debug**
- 不应有 — iot 是纯持久化边界，没有"业务事件"概念，只有错误（Error/Warn）

### wisefido-cardagg

**Info**
- 启动/关闭、stream 订阅 ready
- alarm_events 真实写入（每条 alarm row 一行 Info）
- 状态投影变化（card:status hash 字段跃迁）

**Debug**
- monitor buffer 入队、redis pipeline 命中、未变化的 device_status

### wisefido-sensor

**Info**
- 启动/关闭、roomengine 路由就绪
- Fall verifier verdict、AI emit（带 trace_id）
- alarm_back_channel forward（真转发到 iot:alarm:stream 的那刻）

**Debug**
- `radar suspected (event_log only, no alarm forward)` — Suspected\* 早返
- track/grid 内部细节

### wisefido-data

**Info**
- 启动/关闭、HTTP server ready、playback request entry

**Debug**
- 单 SQL 查询参数、cache 命中

## 编码 checklist

写新日志前问自己：

1. **每 N 秒/分钟会被触发吗？** → Debug
2. **同一信号已被另一处 Info 记过吗？** → 不写（避免 dual-log）
3. **运维收到 PagerDuty 时希望看到这条吗？** → Info / Warn / Error
4. **payload 是原始 JSON？** → 走 Debug 带 ByteString；Info 只写人话 summary 字段
5. **跃迁/初次观测？** → Info；**稳态 tick？** → Debug

## 反模式（违规清单）

| 反模式 | 例子 | 改法 |
|---|---|---|
| 周期 tick Info | 健康检查每设备每 tick info | transition-only |
| dual-log | log.Printf + logger.Info 同一事件 | 删 log.Printf |
| no-op 也 log | UpdateXxx 短路返回还 log "synced" | 函数返 `changed bool` |
| raw JSON 给人看 | payload 直接当 Info 字段 | 加 `info` 白话摘要 |
| 注释指责调用方 | `log.Printf("xxx (from Y called by Z)")` | 注释只写 WHY 不显然时 |
