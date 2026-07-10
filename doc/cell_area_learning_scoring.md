# Cell Area 学习重设计：log-odds 打分 + 人控边界

状态：设计稿（待评审）。触发 case：E598A2ACD523 2026-06-23 11:28:39 floor 假摔。

## 0. 一句话

把 cell area 学习从「硬阈值直升」改成「**per-cell log-odds 证据竞争**」；只增强 **Sit** 并轻量统一 **Sit/Active/Unknown** 三态；**Deny/Enter/Toilet/Shower 不动**；**Bed/Lying 算-但-哑、人授权才生效（永久 pin AreaBed / dismiss）**。cell-learn 仍是 area 权威单源，DBN 只读其投影；human feedback 落 cell，不注入 DBN。

## 1. 病根（实测）

- D523 角装（installModel=corn, (-300,10), rotation45, height200），用户常坐的 (0,520) 在**最大斜距后墙角**。
- 固件把"坐"误读成 **pose=Standing(4) + z=0**（既非 Sitting(3) 也非 Lying(6)）；位置抖动 ±30~50cm。
- 现有 Sit 学习全失效：
  - RegionStatic A 路径要 `z>0`（z-jump）→ z=0 废。
  - RegionStatic B 路径 / PR-7.2 stand-static 用**位置-region 累加器**（连续帧 |dx|,|dy|≤15cm）→ ±30~50 抖动反复 reset，累不到 12min → 不触发。
  - firmware-sit（pose=Sitting）也不在（固件报 Standing）。
- 走过 2 次 → Walk 直升 **AreaActive(4)**（[roomengine_grid_snapshot] 实测三日均 type=4）。
- AreaActive → floor 默认阈 **12min**（floor.go `MuDefaultSec=480` → μ+1.5σ=720s）；用户坐满 12min → **floor 假摔**。

**关键**：固件 stillbox 累到 720（floor 信它），而位置-region 被抖动打散（现有 sit 学习信它）→ 两套"静止"打架。**stillbox 是唯一抗抖动可信源**（且全系统单算单源）。

## 2. 架构决策（已批 Option A 增强版）

| 方案 | 裁决 |
|---|---|
| A. cell-learn 不变，增强 sit（+轻量统一 Sit/Active/Unknown） | **采纳** |
| B. sit 独立 DBN | 否（标量+衰减信念，DBN 过度建模；多一条时标/wiring） |
| C. 五类全进 DBN | 否（毁掉 cell↔DBN 时标/单向/权威分离；mislearn 直入每帧 fire 环；爆炸半径大） |

不变量：
- **时标分离**：cell-learn 慢/批/持久（每5min dump `roomengine_grid_snapshot` = `owlRD/dbv2/28_room_cell_state`），DBN 每帧只读。
- **单源**：cell 是 area-type 唯一权威；DBN 经 `BuildObservation` 读 `cell.Belief[0].Type → obs.AreaType` 当先验。
- **feedback 落 cell**：human 改"地图"(cell, Source 分档/持久/可审计)，DBN 是"读图人"，绝不往 DBN 塞临时记号。

## 3. 五类型 + 权威/风险分层

| 类型(AreaType) | 涉 fall fire | 学习/激活 |
|---|---|---|
| Unknown(0) | 否 | 基线 |
| Enter(1)/Toilet(7)/Shower(6) | 几何 | layout 画布锁定（不动） |
| **Sit(3)** | 否（仅抬 floor 阈 12→90min） | **auto log-odds（本稿增强）** |
| **Active(4)** | 否 | auto（transit；并入 log-odds 竞争） |
| Deny(5) | 是（拒 track） | auto 但保守：15天 BFS + 并存（不动） |
| **Bed/Lying(2)** | **是（压制真摔）** | **auto 打分但对 DBN 哑；人 handle 授权 永久 pin AreaBed / dismiss** |

红线：**Bed 错标 = 漏真摔**，故永不 auto 生效。

## 4. 判别轴：static+recurring 不足

Sit 与 curtain/绿植杂波**都"静止+易重复"** → 必须第三轴区分（否则杂波误学 Sit→软真摔，或误学 Deny→拒真摔）：

| 轴 | 人·Sit | curtain/绿植 → Deny |
|---|---|---|
| 进出轨迹 | walk-in + walk-out | 原地凭空生灭，无进出路径 |
| **并存测试** | 它就是那个人（无并存真人） | **真人在别处被追时它同时在原地 → reflection（杀手判据）** |
| z/vitals | z∈[40,70]、可能 HR/RR | z=0/乱跳、无 vitals |

并存判据来自 per-帧 **DBN realness/co-existence 轴**（见 [[realness_axis_redefined_real_vs_mirror]]）→ 作为**一路输入**喂慢速 cell-learn 的 Deny 通道。非"仅用 DBN"。

## 5. 打分模型：per-cell log-odds 竞争

每次 track 占用/经过 cell 产 1 条 **occupancy episode**，抽特征 → 累进 per-cell 分；argmax 过 τ 升格；都弱 = Unknown。

### 5.1 Sit 的 4 条正向通道（log-odds 相加）

```
ΔL_sit(episode) = LLR_resolution  +  LLR_z  +  LLR_dwell  +  g_fwSit
```

| 通道 | 公式 | 参数（oracle，文献/实测支撑） | 备注 |
|---|---|---|---|
| **resolution（硬判据）** | walk-away → +a；lost/coast/无exit → **否决（不计/负大）** | a 大 | 倒地者不会自己走开；这是 sit↔fall 的硬分割 |
| **z（纯正向平滑，缺失中性）** | `m(z)=exp(−(z−zSitCenter)²/(2·zSitSigma²))`，episode 内取 max；**z=0→0(中性,不扣分)**；**无负分支** | `zSitCenter=80`,`zSitSigma=12`,权 +0.4 | 高斯隶属度替硬框：峰=实测 Sitting z>0 中位 80，σ=12 使 [73,87] 近满、~105 自然趋零。**中心可调**（165cm 实测→80；175cm 均值→~83–85；老人偏矮 80 代表性好；per-deployment 可标）。z=0 占 70%=缺失，绝不用 z 反向否定 |
| **dwell（平滑 bump）** | `σ((t−8)/3)·σ((55−t)/10)`；<5min→Active；峰~20min；30+渐降（过久偏躺）；~90→0 | rise-mid 8 / fall-mid 55min | MBD 12–16min（PMC8679788/9166254）；升降双 sigmoid 替分段 |
| **firmware-sit（pose=3，权重>z）** | `g=fwMax≥1?clamp(0.6+0.04(fwMaxMin−1),0,1.5):0`，×3s软门 | base 0.6, Tmin 1min, Gmax 1.5 | pose=3 = 固件融合「z+质心变化幅度」比 z 富 → **base 0.6 恒 > z 的 0.4(pose>z)**；与 z **各算各的(独立相加,不互门控)**；安全靠 episode walk-away 门控(near-field 假 sit 不自走→veto)，非高 Tmin |

> **z 实测标定（2026-06-23 monitor_stream 24h，track_id≤7）**：青澜固件静态姿态 z 多为 0（Sitting 70%、Standing 93% 帧 z=0=缺失）。**z>0 时**各姿态中位：Lying 69 < Standing 72 < **Sitting 80** < Walking 89；Sitting z>0 又紧又居中（IQR 73–87，90% 落 [62,100]，91% 在 60–100 两档）。故 z 作**纯正向弱佐证**：z∈[65,100] 时小幅加分（episode 内已排除 Walking），缺失/区间外不扣分。原 [40,70] 已按实测改 [65,100]。厂家 TI/Infineon 用 centroid 高度，定义≠position_z，不可借。

**键控用 stillbox（StillSec），不用位置-region**（治本：抗 ±30~50 角装抖动；与 floor 同源）。空间聚合用 cell±50cm，不用 ±15。

### 5.2 Active / Unknown

- **Active** ← transit：Move 穿过 / dwell<5min / `walk_duration`,`stand_duration` 分钟统计（[fields.go] firmware）。
- **Unknown** ← 两类都 <τ 的基线。
- **优先级 Sit > Active**：一次确认 settle-and-leave 权重 ≫ 一次 transit pass → 既被坐又被走过的椅子位正确判 Sit（治"Active 赢"老 bug；与现有 cell_learning 注释"2. Sit 阈值最强"在 Walk 前查一致）。

### 5.3 Deny（不动）

15天 BFS + 5-cell 软共识 + 并存(realness) → AreaDeny。本稿不改。

### 5.4 Bed/Lying：算-但-哑 + 人授权

- **auto 打分**：长时 Lying episode（pose=Lying 持续 + dwell≥长阈 + walk-away 收尾）→ 累 Bed 分，但 `Source=SourceShadow`。
- **DBN 哑**：`BuildObservation` 对 `SourceShadow` 的 Bed cell → `obs.AreaType=AreaUnknown`（floor 照 12min，**真摔照报**）。分数仅驱动 handle 界面"候选建议"。
- **人 handle 授权**（[feedback.go] AlarmFeedbackIngester 扩展）：

| 选项 | 机制 |
|---|---|
| **永久 pin（AreaBed）** | `Source=SourceHuman` + **AreaBed**（非 Sit：沙发/躺椅会斜躺，AreaSit 不压"躺"→固件按 lying 续 fire；AreaBed 才压躺，与 [[bugB_arealying_handoff]] 一致），写入 layout、不衰减、DBN 生效 |
| **dismiss（none）** | 仅消本次告警，保持 Unknown（哑），**不抑制** → 告警继续 |

**无 2H（已删）**：时间盒抑制会静默跨交接班 → 隐性漏报；且给了 snooze 逃生门会让管理者拖延 layout 更新。

**权限-升级催办模型（= "逼管理者更新 layout"）**：
- 永久 pin = 改 layout = **管理者权限**；护理只能 `dismiss` 本次（不抑制）。
- 未 pin 前，该 furniture rest zone **持续 fire** → 升级催办**定向推给管理者**（非无限砸护理，防告警疲劳）→ 管理者**一键**永久 pin。
- 安全单调：auto-score 只"建议"绝不降灵敏度；放松只由管理者永久 pin 触发，无时间盒、无静默到期。

## 6. 数据结构 / Source 档

```
Source: Unset0 / Human1 / Learned2 / Feedback4   （现有）
      + Shadow（新）：算+存+显示，BuildObservation 映射成 Unknown，DBN 不消费
Cell 增：SitScore/ActiveScore/BedScore（log-odds 累加，沿用 Confidence 衰减半衰期）
        RZC（RestZoneConfirmed，已有 schema_v5）→ 升格阶梯门
（无 2H TTL/AuthorizedUntilMs：Bed 只有永久 pin 或 dismiss）
```

## 7. 接口（最小爆炸半径）

| 文件 | 改动 |
|---|---|
| `cell_learning.go::LearnCellAreas` | Sit/Active/Unknown 段换 log-odds 竞争；Deny/Enter 段不动 |
| `track_manager.go::updateRegionStatic` | 累加器从位置-region 换 **stillbox 驱动**；保留 A 路径（z-jump，z>0 正向）作可选加速 |
| `cell.go::MarkRestZoneByFeedback` | 加 `SourceShadow` 分支；Bed auto 走 shadow |
| `belief/floor.go` | 零改（tFloorFor 已读 obs.AreaType+RoomType，μ/σ 单源复用） |
| `adapter BuildObservation` | `SourceShadow` Bed → obs.AreaType=Unknown 映射 |
| `feedback.go::AlarmFeedbackIngester` | 坐类→AreaSit；furniture-lying(沙发/躺椅)永久 pin→AreaBed/Human(管理者权限)；dismiss=不抑制；升级催办定向管理者 |
| `cmd/xsensor bootstrap` | （独立工单）加载 28 grid snapshot，replay 同源 [[replay_floor_timer_needs_warmup]] |

## 8. 回归红线 / 安全

- **Bed 永不 auto 生效**；`verified` 真摔擦非-Human（Sit/Bed-Feedback/Shadow 全擦）。
- z 只正向：低/零 z 永不否定/不断言 fall。
- floor 首发不可压（12min 照报；学习只防复发；不得为压首发延后 floor → 违 FN-safe）。
- 验证走真 case 跑 replay 看机制（owlBack 规则#3 + [[validate_real_case_no_unit_tests]]），且 **replay 必加载 28** 否则 area 维度无效。

## 9. Oracle 参数（无真摔数据，留调）

| 参数 | 初值 | 来源 |
|---|---|---|
| dwell rise-mid/fall-mid (slope) | 8 / 55 min (3 / 10) | 平滑 bump 峰~20min;MBD 12–16 |
| **z 中心 zSitCenter / σ / 权重** | **80 / 12 / +0.4**（高斯隶属度，中心可调） | 实测 Sitting z>0 中位80 IQR73–87（2026-06-23）；纯正向缺失中性；175cm 人群→~83–85 |
| Active 下闸 | <5min | 短 bout<10min，保守 |
| firmware-sit base/slope/Tmin/Gmax/软门 | 0.6 / 0.04·min / 1min / 1.5 / 3s | pose=3 含质心变化幅度,base>z(0.4)即 pose>z;与 z 独立相加 |
| floor μ/σ（复用 floor.go 单源） | Default 8/2.67min；Sit 60/20min；Bath 12/4min | 已在库 |
| 升格 τ / N episode | **6.0 / ≈2**（见 §11.4，oracle） | log-odds 累积 |
| LLR 权重 resolution/z/dwell-peak | **+2.0 / +1.0 / +1.0** | §11.3 oracle |
| firmware-sit Gmax/k | **1.5 / 0.0556 per-min** | Gmax/(Dcap−Tmin) |
| SitScore 衰减 HL | **≈4d** | 隔离 episode 自然褪 |
| 锚半径 / activeCutoff | 50cm / 5min | 抗角装抖动 / 短 bout |

## 10. 待评审决策点（§11 已给 oracle 初值，红字处待你拍）

1. ~~SitScore τ/N~~ → §11 定 τ=6.0 / ≈2 episode（可调）。
2. ~~firmware-sit Gmax/k~~ → §11 定 1.5 / 0.0556（可调）。
3. SourceShadow 是否独立持久列 vs 复用 Source 值。
4. 升级催办的载体/权限模型：furniture-lying 持续 fire → 定向管理者 → 一键永久 pin AreaBed（走哪条通知/审批链待定）。
5. **SitScore 是否落 28 payload 持久化**（建议是：重启不丢累积，与 cell-learn 持久化一致）。
6. 升格 Source 用 `SourceLearned`（auto 可被 human 覆盖）vs 现有 RegionStatic 写的 `SourceFeedback`——建议统一为 Learned。

## 11. Sit 4 通道实现细化（episode 状态机 + LLR 数值）

> **🟢 已实现（2026-06-23,branch 已并 main,Xsensorv1+wisefido-sensor 两引擎）**：`emitSitEpisode`（sit_learning.go + track_manager.go）。
> - **单源 = 区域创建（LEARNING）**：仅 emitSitEpisode 从零造 AreaSit；旧 updateRegionStatic/markBothCellsAreaSit/stand-static 全删。PR-11 的 5min-sit refresh（gated on 已是 AreaSit）是**已确认区 upkeep**（与 lying-4h→AreaBed 对称），不造新区、不破坏单源。
> - **键控用现成 `StillBoxRunStart`（30s 滚动 50×50 box，≈±50cm 锚）**：box-break(updateContinuousIndicators)=walk-away→emit；ResetStillBox(lost/evict)=不 emit（fall 域）。
> - **config 可调**：HL=`DecayParams.SitScoreSec`(默认4d) / τ=`LearnParams.SitPromoteTau`(6.0) / 半径=`SitSpreadCm`(30)，经 `SetSitLearnParams` 注入 TrackManager；cap=τ×`sitScoreCapRatio`(1.5)。SitScore 不持久化（重启清零，低影响；要续学进 CellSnapshot+升 v11）。

替换现有 `updateRegionStatic` 的「位置-region（|dx|,|dy|≤15cm + 帧比）」为「**锚-停留（stillbox 驱动 + ±50cm 锚 + break 分类）**」——治本：位置-region 被 ±30~50cm 角装抖动反复 reset，而 stillbox 抗抖动。

### 11.1 Episode 生命周期（挂 TrackState，每 present 帧驱动）

TrackState 新增字段（删旧 `RegionStatic*`）：
```
sitAnchorX, sitAnchorY int     // 当前停留锚（±50cm 内算同锚）
sitAnchorMs            int64    // 锚建立时刻 = dwell 起点
sitFwContigMs          int64    // 当前 pose=Sitting 连续段起点（断则重置 = 3s 软门载体）
sitFwMaxMs             int64    // 本 episode 内最长连续 Sitting
sitZBest               float64  // episode 内 z 隶属度 m(z)=exp(−(z−zSitCenter)²/(2·zSitSigma²)) 的最大值（纯正向；z=0 不更新 = 缺失中性）
```
常量：`zSitCenter=80`（坐姿 z 中心，oracle 可调）、`zSitSigma=12`。

每帧（取 x,y,z,pose,StillSec）：
1. `d = |xy − anchor|`
2. **d ≤ 50cm（同锚）**：`dwell = now − sitAnchorMs`；`pose==Sitting(3)` → 延 `sitFwContigMs`、`sitFwMax=max`；否则断 contig；z：`z>0 → sitZBest = max(sitZBest, exp(−(z−zSitCenter)²/(2·zSitSigma²)))`（z=0 不更新 = 中性）。
3. **d > 50cm（移动）**：`classifyBreak()` → 若 walk-away 且 `dwell≥5min` → `emitEpisode()`；随后 re-anchor 到新 xy（开新 episode）。
4. **track 消失帧**：`classifyBreak()`（lost/exit）→ emit 或 veto；清 episode。

### 11.2 Break 分类（= 现有"人走 vs 人摔"判定，同信号同阈，FN-safe）

```
classifyBreak(ts):
  present 本帧?
   ├─ 是 + 位移>50cm(同 track_id)                       → walk-away   (resume-move:人起身走)
   └─ 否(track 消失):
        recentRadarEvents 有 ExitRoom(本 tid/lid, ≤12s)   → walk-away
        ExitLogOdds(lid) ≥ exitFlipLogOdds(=logit(0.85))   → walk-away  (离房对数几率翻转)
        logic_id 仍活在别 track_id(renumber 守恒)           → 不算 break(重关联续 episode)
        else (凭空消失 + lostExitInfo 非朝门趋势)            → lost → VETO(不 emit;fall 域,告警留)
```
- **walk-away = 任一正向离场/移动证据**（resume-move / ExitRoom / ExitLogOdds 翻转）→ emit。
- **lost = 无离场证据的消失**（真摔特征）→ veto。含糊一律当 lost（FN-safe）。
- 信号全来自 `track_manager.go` 现成：`recentRadarEvents`(12s buffer) / `ExitLogOdds`(12s claim) / `exitFlipLogOdds=logit(0.85)`([belief/engine.go]) / `lostExitInfo`(末2s朝门强度) / logic_id 守恒（[[logicid_unified_census_refs_tm]]）。

### 11.3 4 通道 LLR（emit 时算一次；仅 dwell≥5min ∧ walk-away 才到这）
```
LLR_resolution = +2.0
LLR_z          = 0.4 · sitZBest                 // 平滑高斯隶属度（峰 zSitCenter=80, σ=12），纯正向；z=0 不更新=中性，绝不扣分、无负分支
LLR_dwell      = σ((dwellMin−8)/3)·σ((55−dwellMin)/10)   // 平滑 bump:峰~20min≈0.95,14min≈0.86,45→0.73,60→0.38,~90→0;<5min 不到此(→Active)。[bedLean 已 deferred 未接线:Bed 本就 human-gated,30+ 久坐自然由 dwell 渐降表达]
g_fwSit        = fwMaxMin≥1 ? clamp(0.6 + 0.04·(fwMaxMin−1), 0, 1.5) : 0   // base 0.6（pose=3 出现即 >z 的 0.4,恒 pose>z）+ 时长加成,封顶 1.5;3s 软门滤闪烁;Tmin=1min（安全靠 episode walk-away 门控,非高 Tmin）
ΔL_sit         = LLR_resolution + LLR_z + LLR_dwell + g_fwSit
```
3s 软门由累积期 `sitFwContigMs` 实现：连续 Sitting 须 ≥3s 才计入 `sitFwMax`，sub-3s 段被 contig 重置丢弃（等价 `min(1, contig/3)`）。

### 11.4 cell 累积 + 升格 + 衰减
- emit → 把 `ΔL_sit` 加到**锚 cell + 50cm 内邻 cell** 的 `Cell.SitScore`（float log-odds）。
- 衰减：`SitScore` 每 decay tick 按 HL≈4d 指数衰减（隔离 episode 自然褪；2–3 次近日 episode 才够 τ）。
- 升格：`SitScore ≥ τ(=6.0)` → `MarkRestZoneByFeedback(AreaSit)` 传 **SourceLearned**（auto，可被 human feedback 覆盖）。
- N 隐含（dwell≈0.9 取 ~18min）：**角落 episode**（resolution+2.0, z=0, dwell+0.9, fwSit=0[固件报 Standing 非 Sitting]）ΔL≈**+2.9** → τ=6 ≈ **2~3 次**自走长坐升格（纯姿态无关两通道）；**正常椅子**（+z 0.4, fwSit ~1.1[pose=3 14–18min]）ΔL≈**+4.4** → ~2 次（pose 报对则更快）。τ 可调以定 N。

### 11.5 接口（最小爆炸半径）
| 点 | 改动 |
|---|---|
| `track_manager.go` TrackState | 加 §11.1 字段；删 `RegionStatic*` 旧字段 |
| `track_manager.go::updateRegionStatic` | 重写为锚-停留+break+emit；A 路径(z-jump,z>0)并入 LLR_z 正向加速 |
| `cell.go` Cell | 加 `SitScore float64`；建议落 28 payload 持久化 |
| `classifyBreak`（新） | 复用 lostExitInfo / ExitLogOdds / recentRadarEvents |
| `cell.go::MarkRestZoneByFeedback` | 升格调用传 `SourceLearned` |
| `cell.go Decay` | 加 SitScore 指数衰减分支 |

### 11.6 红线
- **lost-break 永不 emit**（fall 域）；dwell<5min 永不计 Sit（→Active）；z 只正向（z=0 中性）；**Bed 永远 human-gated**（算法绝不 auto 升 Bed；bedLean 自动建议已 deferred 未实现）。

## 12. AreaSit 学习接通 floor：影子 chair + 分级封顶（2026-07-09 定稿）

> **病根（本次排查实测，§11 之后暴露）**：§11 的 SitScore 学习**只产出 AreaSit 标签，floor 兜底完全不消费它** → 学成 Sit 也治不了它自己的立项 case（"用户坐满 12min → floor 假摔"）。三处断裂：
> 1. `floor.go::tFloorFor` **无 `case areaSit`** → 学到的 Sit 掉进 `default → tActive(12min)`；90min 松绑只挂 `case inChair`（几何 pin）。
> 2. `emission.go::cellMuSigma`（返回 sit 90min μ/σ）**无任何调用方 = 死码**（floor.go 注释 §I "规划" 从未接线）；且 `radarLogS` 里 **still 时长根本不进 emission**（静止→fall 已单源到 FloorGuard）。
> 3. `emission.go` 的 `case areaSit → 抬 SSit` 是弱 redirect，**明确不压 SFallen**。
>
> 而 FloorGuard 是**独立 OR 腿**（`engine.go` `fg.Step()` → `d.Fire=true`，不看 pF），12min 照火。所以在 OR 架构下，**抑制只能在 floor 抬阈实现**，belief 侧压 SFallen 全被 OR 盖过 = 无效。

### 12.1 决策：learned Sit 与 pin chair 收敛成同一结构，只差 Source 与封顶

| | 载体 | Source | floor 封顶 | dwell 学习 |
|---|---|---|---|---|
| **自动学习 Sit** | **影子 chair object**（合成 object_id） | `SourceLearned` | **30min** | μ/σ/maxSit 后台学（同 ChairDwell） |
| **管理员 pin** | 同一 object 翻正 | `SourceHuman` | **90min** | **继承**影子已学的 μ/σ（warm-start） |

tFloor 公式两档共用，只换 cap：
```
tFloor = min( max(chairMu + 1.5·chairSigma + 600, chairMaxSit), cap )
cap = 1800(30min) if Source==Learned ; 5400(90min) if Source==Human
```

- **90min = 通用 FN 天花板**：现状 `tFloorFor` 已把 pin+人工 maxSit 棘轮硬顶 90min（floor.go §chair）；任何 Sit（人钉/自学）都不破 90。
- **30min = 自主降灵敏的克制上限**：比已上线的**浴室自主学习顶 45min**（`BathCapSec`，per-room 无 pin）**更保守** → 不在新增风险类别里。
- **顶切换只认正式 pin（Source→Human）**：护理的 false-alarm "Sit on Chair" 反馈可棘轮影子 maxSit，**但被当前 cap 夹住，绝不解锁 90**；解锁是管理员动作（doc §5.4/§10.4 一键 pin）。

### 12.2 为什么保留 μ/σ（不 flat-30）：warm-start = 催办闭环的手感

μ/σ 在自动档**对 30 阈是被夹平的废值，对 pin 继承不废**——两职分离：
- **当前容忍** = 30 顶（自动档保守）；
- **历史记录** = μ/σ 后台一直学，**只为 pin 那一刻被继承**。

催办 UX 命门：若 flat-30、pin 后冷启重学几天 → "我都 pin 了怎么还在报" → 催办砸信任。保留 μ/σ 后 **pin 只把同一份已学分布的顶从 30 抬到 90 → 瞬间生效、FP 立停**。

### 12.3 影子 chair：白嫖现成 chair 通道解决持久化

学习区无 object_id、dwell 跨 layout 编辑/重启如何存活 —— **实现成合成 object_id 的影子 chair，整条复用 chair 机器**：
- **mint**：`SitScore ≥ τ` 升格时，造影子 chair：rect = anchor ± `sitSpreadCm`，`Source=Learned`，合成 object_id（如 `learned_sit_<cellX>_<cellY>`）。
- **持久化**：走**现成 `chair_dwell_state`**（object_id 锚定，[[chair_dwell_learning_migrate_to_object]] 当初就为抗 layout 编辑才迁 object_id）→ canvas 编辑不碰此表 → **影子天然存活**；新增 `source` 列。
- **hydrate**：`chair_dwell_state` 已存 `CX/CY` → 合成 rect，不依赖 canvas。
- **floor 读**：`InChair` 分支**按 object 的 Source 选 cap**；`ChairRect` 加 `Source` 字段。
- **pin**：管理员一键 pin = 把影子 `Source` 翻 `Human`、cap 30→90、**dwell 原地继承**；影子 object 本身**就是催办 surface 给管理员的那个 pin 建议**。

### 12.4 安全网（四重，全部现成或结构性）

1. **90min 通用天花板**：任何 Sit 都不破（误学最多拖 90min，且那要先经人 pin）。
2. **30min 自主顶**：自动档误学最多拖 30min < 浴室已上线 45min。
3. **leak 自愈**：影子不再被坐 → `SitScore`(HL 4d)/`Belief.Confidence`(<10 降级) 衰 → 影子 chair 消失、cap 随之没。
4. **walk-away-only（最强）**：`emitSitEpisode` lost/fall 一律 VETO 不 emit（§11.2）→ 真摔点（lost）**结构上学不进 Sit** → 影子只可能长在"真有人自走长坐"的格。

### 12.5 边角

- **首次长坐仍 12min**：影子要 2-3 次 walk-away 长坐才升格；某人第一次就坐>30min 时影子未生成 → floor 走 default 12min 照报（对的，全新点第一次本就该催办）。
- **删死码**：`emission.go::cellMuSigma`/`stillMuSigma`（OR 架构下永不生效）删除。
- **催办 vs 减扰取舍**：本次定**催办优先** → 自动档 30min 顶偏紧，长坐区早触发"去 pin"提示；不追求"尽量别烦护理"（那会松到对齐浴室 45）。

### 12.6 接口（最小爆炸半径，两引擎 1:1）

| 文件 | 改动 |
|---|---|
| `belief/floor.go::tFloorFor` | 新增 cap 参数（由 Source 决定 30/90）；`inChair` 分支按 cap 夹顶替死 `tSit`；删 `case area==areaSit` 无（走 inChair）。删 `cellMuSigma`/`stillMuSigma`。 |
| `track_manager.go` `ChairRect` | 加 `Source` 字段（Learned/Human）。 |
| `track_manager.go::emitSitEpisode` | `SitScore≥τ` 升格时 mint 影子 chair（合成 object_id + rect + Source=Learned）注入 `tm.chairs`+`tm.chairDwell`。 |
| `track_manager.go` floor 喂参 | box-start/break 处 `TFloorFor` 调用按命中 chair 的 Source 传 cap。 |
| `chair_dwell.go` / `chair_dwell_state` | 加 `source` 列；Snapshot/Hydrate 带 Source；hydrate 从 CX/CY 合成影子 rect。 |
| pin 落地（data `handle_pin_zones` / sensor `SetChairs`） | 管理员 pin 落到既有影子 object → Source 翻 Human（cap 升 90，dwell 继承）。 |

**Why（本节）**：§11 学了个花架子（AreaSit 标签无人消费）；本节让学习真正接通唯一有效的抑制层（floor OR 腿），并用"影子 chair"把 learned/pinned 收敛成一套结构、白嫖持久化、天然实现 warm-start 与催办 surface。分级封顶（30 自主 / 90 人授权）守 FN 红线，且比已上线浴室机制更保守。

### 12.7 实现状态（2026-07-09）

**Phase 1（已实现，两引擎 build+vet 全绿、核心逻辑 1:1）**：floor 分级封顶机构 + 影子 chair 内存 mint。
- `belief/floor.go`：`tFloorFor`/`TFloorFor` 加 `chairCap` 参；chair 分支冷启与硬顶都用 `chairCap`（≤0/越界回退 90min）。删死码 `cellMuSigma`/`stillMuSigma`/`roomMuSigma`。
- `ChairRect.Source`（chair_dwell.go）；`chairCapForSource`（Learned→30min / 其余→90min）；`chairHumanCapSec`/`chairLearnedCapSec` 常量。
- `ChairMatch`/`chairMatch`（返回命中 ChairRect 含 Source）；`ChairObjectAt` 转调之。
- `SetChairs`：layout chair 标 `SourceHuman`，layout reload 保留 `SourceLearned` 影子。
- `mintShadowChair`：`SitScore≥τ` 升格且锚点无 chair → 造合成 object_id（`learned_sit_<col>_<row>`）影子 chair 入 `tm.chairs`（正本另灌 dwell，Xsensorv1 不灌=冷启走 30min）。
- `ChairCap` 全层线：`StillBoxChairCap`→`TrackStatusBase.ChairCap`→`adapter.RadarTrack.ChairCap`→`Observation.ChairCap`→floor；present 经 dbn_router/main.go、lost 经 engine.go 构造。box-start 按 `chairMatch.Source` 设 cap。

**行为影响**：既有 layout chair（含 feedback pin）全部保持 90min（`SourceHuman`）；仅新增自动学习影子 chair 走 30min。既有挙动不变。

**持久化（已实现）**：影子 chair 直接落 `chair_dwell_state`（不搞 AreaSit-cell 派生），跨重启存活：
- schema：`chair_dwell_state` 加 `src SMALLINT DEFAULT 1`（1=Human/2=Learned），既有行默认 Human（cap 90 不变）。migration `2026-07-09_chair_dwell_src.sql`。
- 写：`ChairDwellRow.Source`；`SnapshotChairDwell` 从 `tm.chairs` 查 Source 落行；`SaveRoom`/`LoadRoom` 带 `src`。
- 读/hydrate：`src=Learned` 且 layout 无此 object → 按 `cx/cy±20cm` **重建 ChairRect** 注入 `tm.chairs`（sensor `hydrateChairDwell` / Xsensorv1 `ChairDwellRead`+`HydrateChairDwell`），dwell 一并灌回。影子跨重启即恢复、不用重学。

**Phase 2（未实现，待做）**：
- 数据层 pin 适配（真 warm-start）：管理员 pin 落到既有影子 object（复用其 object_id 与 dwell）→ `Source` 翻 `Human`、cap 30→90、dwell 原地继承；`layout_parser` 解析 `source` → `SetChairs` 传 provenance。
- 未做前：管理员 pin 走现有 `feedback_sit_<eventID>` 新对象（`SourceHuman`，冷启即 90min，FP 立停但不继承影子 dwell）。
- 验证：按 owlBack 规则 #3 用真 learned-sit case 跑 replay 看机制（影子 mint / box-start cap=30 / hydrate 重建 / leak 降级），需具体 case。
