# case2-quilt-bedside-fall-0604

实测 false-negative（真摔无报警）fixture。201 卧室 / Yang.R / radar `9D8A32A1CD2B`(cd2b) + sleepad `BM87224601641`(1641)。窗口 UTC 22:14–22:31（Denver MDT 16:14–16:31）2026-06-04。

## 文件（tools/export/export.sh canonical 格式，unit 级，owl_v2）
- `window.json` — unit 内两台 Radar(cd2b + 333b) 的 track/heart + event（含被子 lying + 心跳 /
  InBed×3 / LeftBed×2 / Enter×2 / Exit×2 / number_people / activity）
- `window_sleepad.json` — sleepad 1641 的 track（含 InBed/LeftBed 床态）
- `alarm.json` — 设备直发 alarm；本窗口为空 `[]`（两道抑制都生效故无 fire，正是本 case 的 oracle）
- `meta.json` — 3 设备 uid→addr→type（replay 纯文件解 addr，零 DB）
- `room_layout.json` / `room_layout_333b.json` — 两台雷达 /128 canvas（含 bed area）

重新生成：`cd tools && ./export/export.sh case-cd2b-0604-16141631 --tz America/Denver --year 2026`
（案名内 `1614` = 源记录 Denver 本地时刻；内部存 UTC epoch ms，replay 全程 UTC）

## 真实时间线
- 22:15:43 radar InBed(area bed) → 22:16:27 sleepad InBed（差 44s，>15s；单床房同床）
- 在床 >6min
- 22:23:20 sleepad **LeftBed**（人跌床边，雷达不可见）
- 之后 radar 持续在床报 lying（**被子/静止反射**伪装成在床的人）
- 22:25:49 silent_leftbed → **exempt_human_bed**(x80,y230 人工床格) → **抑制**
- lying 消失 → lost_fall pending
- 22:29:44 lost_fall **cancelled_by_recovery**（人坐到椅子 -40,220 在等待窗内）→ **抑制**
- 22:30:22 ExitRoom（离房）

## 这个 fixture 是用来验什么的（oracle）
两道抑制都错：human-bed 豁免让 radar 的"床上 lying"否决了 sleepad 的权威 LeftBed（违反 sleepad!=radar→取 sleepad）。**正确结果 = 应在 ~22:25(LeftBed+窗) 报一条 Fall**（被子是静止反射,不该豁免）。

修"被子静止反射不豁免"后,**回放本 fixture 必须 fire**——这才是"完成"判据,不是我自己编的合成单测。

## 状态
- 由 tools/export/export.sh 导出（unit /64 全活跃设备：cd2b+333b radar + 1641 sleepad，monitor+event+alarm 全流）；replay 纯文件、零 DB、全程 UTC。
- 回归：`TestGateReplay_Case2Quilt_FiresAfterFastEvict`（gate_replay_test.go）用本 fixture 驱动真实 TrackManager
  gate 路径，断言 1+2 驱逐修复下报 silent_leftbed vanish。已证非 vacuous：退回旧驱逐(仅 MissCount)则 FAIL
  (reported=0, cancelled=1=陈旧被子命中 human-bed 豁免)。
