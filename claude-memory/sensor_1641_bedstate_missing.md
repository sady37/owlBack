---
name: sensor-1641-bedstate-missing
description: 2026-05-23 sensor 不写 bed_state 现象不止 1641；2520:* MAC 工作，2460/2470:* 不工作；sleepace 20:17 UTC 才重启 cutover 到 7826455
metadata: 
  node_type: memory
  type: project
  originSessionId: 36f14658-2d5a-470c-b075-31726dab854a
---

**症状**：cardagg 显示 Yang.R 卡 (5oi0n5 / fd00:0:3:112:3::/80) bed_status=1 是 cardagg 修复后的默认值（commit 7826455），不是 sensor 写的真 BedState。

**Why**: sensor 端 bed FSM 漏处理 1641 → cardagg 拿不到 BedState → cardagg 走 "缺 BedState 强制 NotInBed" 兜底（[[bed_status_default_not_in_bed]]）。兜底是临时方案，sensor 真处理才是修复。

**最新发现（2026-05-23 23:00 UTC 接续会话）**：

**1. 现象不止 1641，是一组设备共同问题**

`HKEYS card:state:<unit>/80` 结果：

| Unit /80 | sleepad UID | MAC range | bed_state in Redis | sensor log hits |
|---|---|---|---|---|
| 411:1 (Hunzi) | BM87225200672 | 2520:* | ✓ ts=17:30 UTC | 20 hits |
| 411:3 (MoM) | BM87225200677 | 2520:* | ✓ ts=18:10 UTC | (assumed yes) |
| 411:2 (Ton) | BM87224601335 | 2460:* | MISSING | 0 hits |
| 112:1 (202) | BM87224700865 + BM87224601897 | 2470:* / 2460:* | MISSING | 0 hits |
| 112:3 (Yang.R) | BM87224601641 | 2460:* | MISSING | 0 hits |

模式：**MAC 2520:* 工作 / 2460:* + 2470:* 不工作**。
firmware/model 全同（6.89 / BM8701-2），non-difference。
访问权限全 t（access=t, monitoring_enabled=t）。

**2. sleepace binary cutover 时间线**

- sleepace binary build/start: 2026-05-23 13:17 PDT = **20:17 UTC** today
- 部署的是 commit 7826455 — envelope.SubjectEntity 从 cardID 改成 deviceUID
- 1641 today 全部 bed events 08:01-16:17 UTC = OLD binary period (subject_entity=cardID 格式)
- 672 ts=17:30 UTC 也 OLD binary period — **但成功写到 Redis**。说明 OLD format 本身不阻断流程。

**3. 1641 sleepace 行为**（不为零，否定"sleepace 没发"假说）

| Event | Count | Last ts |
|---|---|---|
| connectionStatus online=true | 10 | 20:17:41Z |
| bedStatus change → event | 3 | 16:17:39Z (LeftBed) |
| inBedStatus | 4 | 16:17:39Z (LeftBed) |
| sleepStage (stage∈{1,2,4}) | 72 | 16:17:39Z |
| (alarm) OfflineRecover | 99 (/80s, health_check 心跳) | 20:34Z (持续) |

sleepace LOG 显示 publish 成功（info 级 "bedStatus change → event"），而非 warn 级 publish failure。

**4. sensor 1641 完全不可见**

`grep -c 1641 /home/wisefido/owl/log/wisefido-sensor.log` = 0。
对比 672 = 20 hits。同样 OLD binary period，同样 sleepace publish 成功——可 sensor 只看到 672。
旧 log（log-20260523 历史）只在 2026-05-22 23:05 出现过 1641 旧 placeholder addr `fd00:0:3:112:3:101:0:1`（rebind 前 slot 0:1）。

**5. SleepaceAdapter silent return gates（每条都不 log）**

`internal/zoneengine/adapter_sleepace.go:124-156`：
- L131: fitness gate → `IsFit(deviceAddr)` false 时静默 return（1641 alarm 流全是 OfflineRecover，没 Offline，应是 fit）
- L133: SubjectEntity == ""（OLD binary 填 cardID `fd00:0:3:112:3::/80` 非空 ✓）
- L137: device_type 不含 "sleepad"（=Sleepad，lowercased "sleepad" 含 ✓）
- L141: ts 超 30s 旧（sleepace 用 nowMs，应 fresh）
- L149: event_status="end" → return（OLD binary publish bedStatus 用 "start"，inBedStatus 用 "instant"，应 ✓）
- L153: bedPref == ""（addr 有效→/96 prefix 有效 ✓）

**结论**：纸面分析所有 gate 都 pass。1641 应该被处理，但实际 0 log。silent drop 在某处发生但找不到原因。

**未排除假说**：

- A. **sleepace 端 baselineCache stale**: 1641 rebind 前缓存 OLD address `0:1`，rebind 后 cache 没 invalidate（config:card:stream 失效路径漏 1641 device_uid）。OLD binary 用 stale addr publish → sensor adapter prefixOf(stale_addr, 96) 仍是同样 /96 prefix → 不影响。**否定**。
- B. **sensor zoneengine fitness silent**: AlarmConsumer.fan-out 误把 1641 当 unfit。但 1641 alarm 流 99 全 OfflineRecover 无 Offline，fitness 不会被 mark。**否定**。
- C. **MAC range 是真因**: 2520:* vs 2460/2470:* — 但 firmware/model 全同，sleepace 解析路径相同，不应解析差异。**待验证**。
- D. **iot:event:stream MAXLEN evict 太快**：stream 1003 条 ≈ 5-10 min window；1641 7 bed events 间隔几小时一次，每次发布到读取间隔 < 1s 应没问题；consumer group lag=0 confirm 全消费。**否定**。
- E. **设备 access/monitoring_enabled flip 期间发生**：1641 当时 monitoring_enabled=false → sleepace canMonitor=false → inBedStatus 跳过（line 479），但 bedStatus 走 canIoT 路径仍发。也不解释完全空白。**部分否定**。
- F. **roomengine bedSession 累计 silent 拒绝**: 早 log 有 1641 旧 addr 0:1 的 fitness 处理，可能 roomengine in-memory 残留 stale state。但 roomengine 走 ProcessSleepadBedEvent 跟 zoneengine 独立，不解释 zoneengine 也漏。**部分**。

**下一步建议**：

1. 重启 sleepace（已 20:17 重启完）后让 Yang.R 上床触发新 bed event，确认 NEW binary 路径下 1641 是否能产生 sensor log。
2. 若新 binary 仍漏，**临时加 debug log 在 SleepaceAdapter.handleMsg 入口**（不进 silent gate 前先打一条），确认事件是否真到了 adapter。
3. 若 adapter 都收不到，看 consumer group `wisefido-zoneengine-sleepace` 的 PEL（pending entries）是否积压 1641 entry。

**入口查询**:
```bash
# 后续 1641 上床后 — sensor 是否 log 任何 1641 痕迹
grep -E "BM87224601641|2460:1641|112:3:101" /home/wisefido/owl/log/wisefido-sensor.log
# Redis 是否最终写入 bed_state
REDISCLI_AUTH='TeLunSu-36kr' redis-cli HGET "card:state:fd00:0:3:112:3::/80" bed_state
# 对比同时段 411:1（已确定工作）
REDISCLI_AUTH='TeLunSu-36kr' redis-cli HGET "card:state:fd00:0:3:411:1::/80" bed_state
```

**前置上下文**：本轮 cardagg 修复见 [[bed_status_default_not_in_bed]] + commit 7826455 (envelope SubjectEntity convention + cardagg display fix)。
