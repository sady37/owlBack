# DBN CD2B Case 测试报告

测试日期: 2026-06-10 ~ 06-11 | DBN_FIRE=1 | gate-list 已删除 | 测试员:sady37

## 总表

| Case | 窗口(MDT) | 固件检测 | gate-list | DBN | 结果 |
|---|---|---|---|---|---|
| case-cd2b-0427 | 04-27 11:35-11:48 | — | — | ❌ | data expired |
| case-cd2b-0606 | 06-06 10:27-10:37 | 0 fire | 0 fire | **0 fire ✅** | **正确不fire**:人动了3次(57/76cm→StillBox破),最长静止仅188s,pose=4站立 |
| case-cd2b-0604-1514 | 06-04 16:14-16:31 | 0 fire | 0 fire | **0 fire** | bed_occupied_suppress bug |
| case-cd2b-0604-1614 | 06-04 16:14-16:31 | 0 fire | 0 fire | **0 fire** | 同窗,无 bed_occupied |

## DBN live 验证(非 replay)
- ✅ 真人坠落(D523,P=0.985,silent,CRITICAL) — 过
- ✅ D523 person_silent FP(DBN 不报) — 过

## case-cd2b-0427 — ❌ 数据过期

- 窗口: 2026-04-27 11:35-11:48 MDT
- 设备: CD2B radar + 1641 sleepad
- DB: monitor_stream 0 track(45 天,已过期)
- **不可测**

## case-cd2b-0606 — ✅ DBN 正确不 fire(人动了3次,不是纯静止)

- 窗口: 2026-06-06 10:27-10:37 MDT
- 设备: CD2B radar + 1641 sleepad
- DB: 605 track 帧
- 原始: 0 alarm(固件+gate-list 均未报)
- DBN: **0 fire → 正确**
- DBN replay Pmax=0.131,0 Floor-Fallen(clean room replay,437 traces in window)
- 场景: 两次 LeftBed(10:28:45 cd2b+10:28:48 1641)。床正确翻 NotInBed
- **★纠正(2026-06-11):之前误判"6min静止",逐帧位移实算:**
  - 16:30:12 14cm微动→16:31:20 **57cm跳**→16:31:22 58cm回→16:34:30 **76cm起身**
  - StillBox(30s滚动窗,50cm阈值)被57cm/76cm两次击破
  - 最长连续StillBox=188s,pose=4站立,DwellStill累积不够
- **★纠正2:之前说"sleepad LeftBed 没进 bed scorer 导致压制"是错的。即使 bed scorer 正确、床态正确,人动了就达不到 fire**
- **结论:DBN 不 fire 正确。人动了。**

## case-cd2b-0604-1514 — ⚠️ DBN 不 fire(bed_occupied_suppress bug)

- 窗口: 2026-06-04 16:14-16:31 MDT
- 设备: CD2B radar + 1641 sleepad
- DB: 595 track 帧
- 原始: 0 alarm(固件+gate-list 均未报,false-negative,"棉被旁摔")
- DBN: **0 fire**
- DBN 压制: `bed_occupied_suppress`×195 + `nodetect_gated`×194 + `v认同_evidence`×194
- bed_decision: **LeftBed P=0.01**(床空!),但 `BedStatus` 字段可能 stale
- 场景: sleepad LeftBed 16:23:20;放射 7 分钟后 ExitRoom。人掉进棉被,雷达看不见摔
- **bug: 床是空(LeftBed P=0.01),`bedAdapter` 读 `BedStatus==0`(InBed)才设 Value=1 → LeftBed 时 Value=0 不该触发压制。195×bed_occupied_suppress 与床态矛盾 —— `BedStatus` 字段与决策不同源**
- **gap: `BedStatus` 字段未与 `bed_decision_trace` 同步。bed_decision=LeftBed 但 BedStatus 仍是 InBed 旧值 → 错误压制**

## case-cd2b-0604-1614 — ⚠️ DBN 不 fire(无 bed_occupied,仍不fire)

- 窗口: 2026-06-04 16:14-16:31 MDT(同窗,第二 run)
- DBN: **0 fire**
- DBN 压制: `nodetect_gated`×83 + `v认同_evidence`×83,**0** `bed_occupied_suppress`
- 相比 1514: 床占用压制消失(第二 run 床态干净),但 nodetect+v认同 仍不 fire
- **结论: 即使床态正确,棉被旁摔的 track 在场+无 dwell → DBN 不 fire。根本 gap = 棉被遮挡后雷达丢 track,但 replay 数据 track 在,所以 DBN 看到的是在场→不触发 nodetect。**

## case-9e7-2311 (快速下床)

- 窗口: 2026-06-05 23:11-23:16 MDT
- 设备: 9e7 radar + 978 sleepad
- DB: 281 track 帧
- 原始: 0 alarm(固件+gate-list 未报)
- DBN: **0 fire**
- 场景: EnterRoom→InBed(9e7+978)→978 LeftBed 23:15:00→9e7 Walking 23:15:01 **(1秒差)**
- **结论: ✅ DBN 正确不 fire。这是快速起床,不是摔。LeftBed+Walking 同时=正常行为**

## case-d5f7-bathroom (浴室摔,05-24)

- 窗口: 2026-05-24 13:35-13:57 MDT, DB: 55 track 帧(稀疏)
- DBN: **0 fire**(room 111:3:300 0 belief 活跃)
- **结论: 数据太旧/太稀疏,不可靠。待用例文件 format 统一后重测**

## case-d523-mirror-ghost (镜像 ghost)

- 窗口: 2026-05-16 21:00-21:40 MDT
- DB: 0 track(数据过期)
- **结论: ❌ 不可测**

## 关键发现汇总

1. **bedAdapter BedStatus 不同源 bug**: `bed_decision_trace`=LeftBed 时,`BedStatus` 可能 stale(InBed) → 错误 `bed_occupied_suppress`。`bedAdapter` 应读 Bayesian 的 `Occupied` 概率而非 `BedStatus` 整数。
2. **LeftBed 回溯 gap**: sleepad LeftBed 晚于 Fall,DBN 不回溯修正(已在 #2 讨论)。
3. **dwell 时长**: 床边摔 27s dwell 不够触发 silent(需要 ≥30-60s),快下床 1s 够。
4. **DBN 不 fire 床旁摔 = 结构问题**:track 在场→无丢轨→不触发 lost-fall;床占用→压制 silent-fall;LeftBed 破床占用但晚到→不回溯。三道门全关 → 床旁摔 FN。
5. **DBN 对真人坠落(D523 bathroom,P=0.985)正确 fire** — 非床旁摔场景,无床占用干扰。
