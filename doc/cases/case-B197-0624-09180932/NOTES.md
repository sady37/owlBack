# case-B197-0624-09180932

Denver-LivingRoom · radar `9923003AB197` (unit `fd00:0:3:111:1::/80`, room `fd00:0:3:111:1:100::/88`)
窗口 2026-06-24 09:18–09:32 (America/Denver) · 用户标注「有 Ghost、2 次 fall(Fall-Cloud)」

## 产物
- `window.json` — 原始 fixture（567 行 track/heart/event；固件只发 EnterRoom×1+ExitRoom×1，无 Fall→纯 Fall-Cloud）
- `window_sleepad.json` `[]` 客厅无 sleepad · `alarm.json` `[]` 无固件直发 alarm
- `room_layout.json` / `meta.json`
- `Xsensor.log` / `track_event_timeline.md` — **gate-list 退役后**重放产物（`timeline_from_xray.py`）

## 现场（raw track，09:18:41）
track0(B197.0) 真人在 (-220,490)；track1(B197.1) 凭空出现于 (-220,340) z 抖 54/19/0/37，与 track0 同 x、y 差 150cm。number_people 一度爬到 3，含 track_id=2。

## 调查结论（铁律 #3：看机制非 fire）
1. **track1 不是穿墙反射 ghost**。曾被 `reflectSep` 判 ghost(p_real 0.09) 的唯一依据是 `sep=405/wall_margin=1.00`，但该房那个位置**无墙可穿**。
2. **真因 = radar 踩墙 bbox 边线**：radar canvas x=250 == 墙 bbox 右边 x=250 → radar→ghost 线段在起点 t=0(radar 自身)被判"穿墙"，sep=radar→track 全长。两连带缺陷：`segSeg` 闭区间 t∈[0,1] 接受端点交点；`wallsFromPolygon` 零厚度线段使 `insideRect`「房内点→非反射」护栏对内部点恒失效。
3. **出生地评分**：track1 door_d=306 ≥ rcDoorScaleCm(120) → census 出生先验 `bR₀=rcRealFloor=0.50`（far-born 可疑下限，FN-safe 不为 0）——即时间线 09:18:41 那个 `p_real=0.50` 的起点；非穿墙、非 sync(rho=0)。

## 本 case 触发的修复（详见记忆 gatelist_retired_ghost_single_source_census）
- **reflectSep**（两树 census.go）：改 bbox 实心判房内（ghost 在房 bbox 内直接 sep=0）+ 排除距 radar <10cm 的自穿交点。修后 track1 sep 405→0、判 real。
- **maxBeds 3→4**（两树 joint.go）：LongSofa 折进 Beds 轴致 numBeds 含沙发撞 bound（阻塞 Ton/Bedroom1 启动）。
- **gate-list 退役**（cutover 任务①）：ghost 单源到 census realness（Xsensorv1=PReal 回灌 / 生产=DBNConfidence）；删 legacy TrackManager BirthScore/Verdict/GhostPenalty/StartupGrace/symmetry + 两树皆死的 computeRisk/TrackOutput.Risk 链。

## 修复后机制（B197 验通）
- track1：09:18:41 起 p_real 0.50→0.51→0.78→0.92，判 real，不再被幻影墙冤判 ghost。
- census p_real 序列、still-box→floor(top_s=Fallen 段) 全保留 = 零回归。
- 读 ghost 看 `dbn[].p_real` 或 `target[].bf_preal`（xray 旧 `verdict` 列已随 gate-list 删除）。
