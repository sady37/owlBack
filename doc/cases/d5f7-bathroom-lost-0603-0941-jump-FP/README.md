# D5F7 浴室 lost_track 误报 — track 跳变变 ghost（2026-06-03 09:41–09:53 MDT）

## 一句话

真人 track（id=0）在浴室洗漱，**09:43:19 发生空间跳变**漂到反射区 (180,40) 变成 ghost、被冻在 ~160cm；firmware 把真人重编为 track 1。真人 09:53:30 ExitRoom 离开（np→0，房已空），但 09:53:40 那个 ghost track 消失被判 lost_track → 10:01:11 误报 Fall CRITICAL。

## 文件

| 文件 | 内容 |
|---|---|
| `timeline_0941_0953.txt` | 1310 行逐秒 monitor(radar.track) + 全部 event，时序合并，未压缩 |
| `room_layout.json` | D5F7 /128 layout（雷达 canvas(0,240) z=270；门 Enter 中心(81,312)；mirror Interfere x0–120,y160–170；shower；curtain）|
| `alarm_record.json` | 误报 alarm_events 行（reason=lost_track, last_verdict=1=Real, cell_area_type=0, still_box_start_ms=0, wait_ms=300000）|

## 关键证据（timeline 摘）

```
09:41:00–09:43:15  track0 在 (30~70,0) p3        真人，近雷达洗漱
09:43:19–23        track0 (80,0)→(180,40)        ★空间跳变 ~140cm/4s 漂到反射区
09:43:36           >>EVENT number_people np=2     固件认成 2 人
09:43:36→          track1=(40,0) 真人 / track0 冻 (160,-30) p4   分裂：真人重编 track1，track0 成 ghost
09:43:39–09:48:40  track0 长期冻在 (160,-30)      frozen ghost（一段 188 帧≈3min 不动）
09:53:30           >>EVENT ExitRoom               真人 track1 走出
09:53:31/41        np 2→1→0                       房间确认空
09:53:40           track0 ghost 末帧 (160,50) 死亡
10:01:11           lost_track Fall CRITICAL       报在已死的 ghost 上（锚点 09:55:41 是 fire−wait−30s 合成值）
```

## 为什么四道防线全漏（出生地为何没限制住）

| 防线 | 为何漏 |
|---|---|
| rule1 出生地（远离 entry→ghost）| 只查 `newBorn`；track0 出生在原点(真人,合法)早通过，跳变非新生 → 不复查 |
| 速度 ghost（>200cm/s）| 跳变摊 4s ≈40–60cm/s，未超阈 |
| mirror_detect 镜像配对 | track0 跳过去即冻、track1 真人乱动，不构成刚性镜像对 |
| anchor-reject | track0 是 Real(=1) 非 Anchored，本可翻 ghost，但无规则去翻它 |

根因：**firmware track_id 不稳定**（复用/跳变/分裂），ghost 判定都挂在 raw track_id 上；真人的 track 整条"漂"成 ghost，出生合法→一路 Real 到死。room_type=1（Bathroom，rule1 在跑）、entry 不在盲区——都不是原因。

## 解题思路：logicID（逻辑轨迹身份）+ 跳变即重判

firmware track_id 不可信，引入 sensor 侧稳定身份 `logicID`，ghost/连续性/lost-fall 全部基于 logicID 判定。

1. **分配**：sensor 收 `iot:*:stream` 时 firmware 不带 logicID（为空）。为每个**新建 track_id** 分配：

   ```
   logicID = device_uid + 出生时刻(ms) + track_id
   ```

   出生戳锚定，全局唯一、稳定不随 firmware 复用而变。

2. **跨 track_id 数据关联**：当出现**新 track_id 且无 EnterRoom 事件** → 遍历所有现存 track，取**上一秒位置最近**的那条 → 新 track_id **沿用其原 logicID**（判为同一逻辑实体的延续/relabel，而非新实体）。

3. **用 logicID 判定**：ghost、连续性、lost-fall pending 都按 logicID 聚合，不按 raw track_id。

4. **跳变即重判**：同一 logicID 位置发生跳变（无 enter）→ 触发 verdict **重评**（等效再过一次出生地判据），不沿用旧 Real。

### 这套如何修本 case

- track0 跳变到 (180)：logicID 不变但位置跳变 → 触发重判；落点远离 entry（在反射区）→ 判 **ghost** → 其后消失不进 lost-fall。
- track1 真人出现在原点、无 enter → 关联到最近的原真人 logicID → 真人身份延续，不被当新实体、也不被误判。
- 净效果：ghost 与真人身份分离正确，ghost 消失不报，真人正常 ExitRoom。

## 最终实现态（2026-06-03，安全版，铁律 ghost 不进 Fall）

设计迭代后落地的是**两道不读 ghost 的身份/守恒闸**（先前"远端反射判 ghost + 门控 lost_fall"的 sticky 方案已回退——它用 ghost 门控 Fall 违铁律，且 firmware id 横跳会把真人也标 ghost 漏真摔）：

1. **id-swap 守恒闸** `hasOtherLiveTrackWithLogicID`：track 失锁但其 logicID 仍活在另一条 track 上 = firmware 换 ID（数量守恒，无人真消失）→ 不入 lost-fall pending。
2. **空房账闸** `roomLedgerEmpty` = `lastExitMs > lastEnterMs`（**只信 EnterRoom/ExitRoom 过门空间证据，绝不信 np=0**——铁律 count=0≠离开，人摔倒丢锁也 np=0）→ 房已空，失锁 track 是"人走后残影"→ 抑制。
3. **Fall 只与 pose**（firmware pose=5 走 qinglan 不受影响）；ghost verdict 退回纯显示信号。
   replay 实证：reported 1→0（track1 id_swap 跳过 + track0 room_empty 跳过，全程无 ghost）；oracle 全 13 案真跌倒 confirm=true 不变。

## 显示层残影修复（cardagg 88 清 track）

**现象**：人走出后 web 仍显示 track >60s（超过前端残影上限，故非残影，是上游一直在续命）。

**根因**（不是 qinglan——qinglan 必须发 `track_id=88` 作 device online 心跳）：cardagg `monitor_handler` 收到 88 时**没按约定**（monitor_buffer.go:13「88 不入库」）丢弃+清 track，反而照写（line 142），且每 1s 用**新 `ts_ms`**（writer.go:256）重发旧 track 位置 → 前端按 ts_ms 判新鲜永不过期。

**修法**（纯显示层，不碰 Fall）：cardagg 收到 88 → `MonitorBuffer.ClearDeviceTracks(cardID, deviceAddr)` 立即清掉该 device 全部 track + 不写 88 + 保留 `TouchLastSeen`（88 仍维护 device online）。前端 6s presence 窗（`DEFAULT_PRESENCE_SEC`，跌倒姿态 30s `FALL_PRESENCE_SEC_DEFAULT`，owlFront/src/utils/radar/types.ts）平滑帧间偶发 88 不闪；持续 88（人真走）→ ≤6s 消失。受控注入验证通过（真 track 在 card:realtime → 注入 88 → ~1s 清空）。

> 关联 `target_state_per_device`（v3 logicID 真融合）、`track_fusion_and_gate_cardid`、memory `lost_track_fall_detection_envelope_gate`（距离闸只治边缘 FP；浴室近距 FP 靠 id-swap 守恒闸 + ExitRoom 空房闸；显示残影靠 cardagg 88 清 track）。
