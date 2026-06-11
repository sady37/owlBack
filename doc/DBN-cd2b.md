# DBN CD2B Case 测试报告

测试日期: 2026-06-10 ~ 06-11 | DBN_FIRE=1 | gate-list 已删除 | 测试员:sady37

## 总表

| Case | 窗口(MDT) | 固件检测 | gate-list | DBN | 结果 |
|---|---|---|---|---|---|
| case-cd2b-0427 | 04-27 11:35-11:48 | — | — | ❌ | data expired |
| case-cd2b-0606 | 06-06 10:27-10:37 | 0 fire | 0 fire | **0 fire** | 不 fire(nodetect+dwell) |
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

## case-cd2b-0606 — ⚠️ DBN 不 fire(与 gate-list 同)

- 窗口: 2026-06-06 10:27-10:37 MDT
- 设备: CD2B radar + 1641 sleepad
- DB: 605 track 帧
- 原始: 0 alarm(固件+gate-list 均未报,false-negative)
- DBN: **0 fire**
- DBN 压制: `nodetect_gated`×85 + `v认同_evidence`×27
- bed_decision: LeftBed P=0.01(床空)
- 场景: 两次 sleepad LeftBed→InBed。10:35:21 LeftBed→10:35:49 InBed(~27s 摔)。无固件 Fall。track 在场无丢轨—So nodetect 门控不触发,silent/dwell 累积不到阈值(27s < dwell)。
- **gap: dwell 太短,27s 不够。不是 DBN bug。**

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
