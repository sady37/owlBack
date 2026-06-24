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

| 通道 | 公式 | 参数（oracle，文献支撑） | 备注 |
|---|---|---|---|
| **resolution（硬判据）** | walk-away → +a；lost/coast/无exit → **否决（不计/负大）** | a 大 | 倒地者不会自己走开；这是 sit↔fall 的硬分割 |
| **z（缺失门控）** | z∈[40,70]→正；z≥80→负（合法高读数）；**z=0/缺失→0(中性)** | μz=55 | 绝不用低 z 反向否定（[[geom_abandoned_use_roomtype_areatype]] z 单向原则） |
| **dwell（对数正态）** | <5min→Active(不计sit)；峰~14min；10–30 强；30–90 偏Bed | 峰 ln(14min)，核心[10,30] | MBD 12–16min（PMC8679788/9166254） |
| **firmware-sit** | `g=clamp(k·(dur−Tmin),0,Gmax)·min(1, contigSit_s/3)` | Tmin≈3min, Dcap≈30min, 3s软平滑 | pose=Sitting(3) 直证；3s 连续软门滤闪烁；封顶防失控（治旧 PR-15 禁因） |

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
| dwell 峰 / 核心 | 14min / [10,30] | MBD 12–16min |
| Active 下闸 | <5min | 短 bout<10min，保守 |
| firmware-sit Tmin/Dcap/软门 | 3min / 30min / 3s | 防闪烁+封顶 |
| floor μ/σ（复用 floor.go 单源） | Default 8/2.67min；Sit 60/20min；Bath 12/4min | 已在库 |
| 升格 τ / N episode | 待标 / 2~3 | oracle |

## 10. 待评审决策点

1. SitScore 升格 τ 与 N（几次 self-resolved episode 升 Sit）。
2. firmware-sit `Gmax`/`k` 标定（与 dwell/z 通道权重平衡）。
3. SourceShadow 是否独立持久列 vs 复用 Source 值。
4. 升级催办的载体/权限模型：furniture-lying 持续 fire → 定向管理者 → 一键永久 pin AreaBed（走哪条通知/审批链待定）。
