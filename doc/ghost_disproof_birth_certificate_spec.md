# ghost 判断反证法重做 —— 出生证 spec（实现工单）

> 本文是实现工单。先在 **Xsensorv1** 改 + replay 验证，再逐字镜像 wisefido-sensor（[[sensor_xsensor_one_to_one_env_differs]]）。
> 验证用 case：`doc/cases/case-ffdb-0707-02080247/`（含 Xsensor.log / track_event_timeline.md / alarm_and_events.md；生产三次 dbn.Fall 02:35 / 02:47 / 02:59，全 auto_resolved 误报）。
> 取代 [[unified_ghost_score_lid_lifecycle_fp_root]] 的多轴 ghost 打分与 [[realness_axis_redefined_real_vs_mirror]] 的 mirror 几何轴。

## 背景 / 根因（case-ffdb-0707-02080247 复盘定论）

冻结干扰 phantom（风扇/反射，固件残迹）在房中央 (135,205) 反复重生，被判 real、攒满 floor 兜底 720s、误火三次。当前 ghost 判断是**一堆窄口径捕手 + 一个把万物拉向 real 的重力**，phantom 从所有缝里漏光：

| 现有机制 | 抓什么 | phantom 的逃逸口 |
|---|---|---|
| mirror 轴（realness 2 态几何：sep/rho/wall） | 反射（需 coexist + 墙外几何） | 非镜像 → p_mirror≈0 |
| split 定罪 | 锚点 glued 孤迹 | **per-lid**，lid churn 即清零 |
| floor-artifact latch | immature(<5s) coast 孤儿 | 成熟轨豁免（"可能真 faller"） |
| teleport purge | ≥200cm/s 瞬移 | 原地冻结没跳 → 不触发 |
| area 自查 | 声明的 AreaInterfer/Reflector | 该格是 AreaDeny/Unknown → 白查 |

外加**重力**：realness 的 `rcRealBase` 每 tick 无条件把 `bR` 拉向 real。

### 三个结构缺陷（所有 FP 的公因子）

- **D1 baseline-real 重力**：非镜像、无定罪的轨靠 `rcRealBase` 单调飘到 real。默认"真、除非证伪"，而证伪只有窄口径的 mirror/split。
- **D2 定罪挂 lid，churn 清零**：FFDB00811215 在 (135,205) 被 split 定罪 → 冻结点 churn 出新 lid FFDB02150835（`birth_x=135 birth_y=205 door_d=65`，全新一身，`real_proven=false`）→ 逃罪 → `rcRealBase` 重新拉到 `p_real=0.73` → real。**per-lid 状态扛不住 lid 重生。**
- **D3 捕手的 FN-safe 逃逸口正好叠在 phantom 栖息地**：成熟 + 非镜像 + 没瞬移 + 格子没声明 —— 五个逃逸口的交集就是这条冻结 phantom。

实证对照：`birth_verdict` 里真人 FFDB22017442（(75,85)）`p_real=1.0` 与 phantom FFDB02150835（(135,205)）`p_real=0.75`，**都判 real**——当前判据从出生就分不开真人与冻结 phantom。

## 设计原则 —— 反证法（default ghost）

> **默认 ghost。real 必须挣得「出生证」，一次挣到即 latch 死；一条轨活满生命周期零出生证 → 就是 ghost。**

- **删掉所有几何判据**（mirror / 墙外 / sep / rho / wall_margin / split 几何打分）。不再每帧算"像不像反射"。
- **删掉 `rcRealBase` baseline 恢复**（D1 的重力源）。real 不再自动飘上去。
- 判据塌缩成"**出生登记一次 + 少量物理事实兜底**"，全部是**真人才有的物理行为/事件**，不是几何猜测。

## 判决轴 ⊥ 显示轴（拆开双职，是治 D1 的架构前提）

`PReal` 旧设计一人干两活——既当 fire/N_r 判决门（≥0.5），又当 FE confidence（×100）。**混用正是 D1 能把连续漂移灌进决策的根**。拆成两个字段：

| 轴 | 字段 | 类型 | 谁读 |
|---|---|---|---|
| **判决** | `RealProven`（cert latch） | **bool** | N_r 计数、F1 floor 占用门（改读它，不再读 `PReal≥0.5`） |
| **显示** | `PReal`（confidence） | **连续 0-1** | 只喂 FE ×100，**gate 任何决策皆无** |

- 判决 = 硬 latch：无证 = ghost，`PReal` 显示多高都不进 fire → **D1 彻底死**（连续值再飘也飘不进决策）。
- 显示 = 连续梯度：随便喂 FE，不怕 ≥0.5，因为没人拿它当门。
- **`PReal` 分档（显示，gates 无，可调 UX）**：

| R（无证，离 born 最远净距 cm） | 档 | `PReal` 显示 |
|---|---|---|
| ≤ 50 | GLUED | 0（0%） |
| 50 – 100 | ramp | 0 → 0.50 线性 |
| 100 – 200 | HOLD | 0.50（50%） |
| ≥ 200 | WALKOUT | 1.00（同刻 `RealProven`→true） |

## 判据 —— 出生证白名单（命中任一 → `RealProven`=true，sticky）

```
RealProven ⟸ E1 门口出生：出生同刻有 EnterRoom 且 raw 自查落 AreaEnter（严 door_d≈0）
           E2 双源床出生：raw 自查落 AreaBed/MonBed 且 sleepad InBed
           E3 handoff 继承：新铸 logicID 从「proven-real ∧ 离场/lost」源轨继承证
           E4 relost 重抓：房级 np 0→1 且 stillbox 在跑 且 stillbox 起点→now 无 Enter/Exit
           E5 vital：该轨/床测到心跳/呼吸（仅独处无证 R<200 这一格启用）
           R  walkout：离 born 最远净距曾 ≥200cm（纯距离，sticky）
ghost ⟸ 以上全无（房中央凭空生、冻着不走、非续轨、无体征）
```

### E1 门口出生
出生瞬时（born hook）raw 末点落 AreaEnter，且近窗有 EnterRoom（[[real_by_provenance_latch_optionb_area_drift]] 门第通道）。**严 E1**：要求 door_d≈0，不要"只要有 EnterRoom"（phantom 出生也常伴 np++/事件，松了放 phantom）。

### E2 双源床出生
raw 末点落声明床区（AreaBed=2 / MonBed=5）**且** sleepad InBed（接触占用）。**双源缺一不可**——光落床格几何不够（phantom 也能生在床格），须 sleepad 坐实真人躺着。复用 [[bed_fusion_authority_model]] / [[mn_bed_firmware_areaid_exec_order]]。

### E3 handoff 继承（防固件重号杀真人，高频）
固件给同一真人反复换 tid，新 tid **生在人当前位置（房中央）、不在门口/床**。本 case 实证 FFDB00811215 在 tid=0↔tid=1 反复 rebind。字面执行①② → 真人每重号变 ghost → 漏摔。
- 出生证**挂 logicID、随最小作功关联传递**（现有 nearestAliveTrack，非新判据）：真人换 tid → 新 raw 关联老 logicID → 证跟着走。
- 关联失手铸**真新 logicID** 时才走 E3：**允许从「proven-real ∧ 正在离场/lost」源轨继承证**。
- **锁**：源轨须 `RealProven ∧ (exit / lost coast 中)`，**非** co-present 冻结 ghost → phantom 的"源"是它自己那条 co-present 冻结轨，继承不到。E3 **FP-中性**。

### E4 relost 重抓（治独处睡者丢-重抓，补 E3 的洞）
lone 静止睡者被 radar 丢-重抓时，新 tid 生在房中央、无门第、**无别的轨可 handoff**（E3 够不着）。E4 专收这条：
```
房级 np 0→1  ∧  stillbox 在跑  ∧  [stillbox 起点 → now] 全程无 Enter/Exit Room
                                ∧  [stillbox 起点 → now] 无 sleepad LeftBed 事件
```
- **房级 np=0**：跨设备聚合，任一雷达 np≥1 即非空（[[per_device_np_latch_no_crossradar_clobber]]），不认单雷达瞎报 0。
- **判别 ffdb phantom**：phantom 在 (135,205) 出生时 np 已 ≥2（真人在场），非 0→1 → E4 不成立 → ghost。✓
- **挡空房 Exit 残迹**：真人 Exit 走后留的冻结残迹，那个 Exit 落在其 stillbox 窗内 → "无 Enter/Exit"违反 → E4 不成立 → ghost。✓
- **无 sleepad LeftBed 守卫（防跌床被洗白）**：窗内若有 LeftBed，说明人刚离床——此时床边的静止轨可能是**跌床真摔者**，绝不能被 E4 当"安详重抓睡者"发证放行、掩盖跌床。有 LeftBed → E4 不成立，该轨走跌倒路径（且其真人身份本就有 E2 床 latch 兜着，不靠 E4）。✓

### E5 vital（仅独处无证 R<200 启用）
独处 + 无证 + R<200 是真·二义：R 零信息（睡者与 phantom 都 R≈0）。此格唯 vital 能判——风扇 phantom 无心跳，静卧真人有。vital 是"活体才有"的物理事实，非几何判据。**仅在这一格启用**（有共存 real 轨时不需要，共存已足判 phantom）。

### R walkout（纯距离，非速度）
`R = max over logicID 生命周期 of dist(当前位, BirthPos)`。`R ≥ 200` → 补证、`RealProven` sticky true。
- **纯距离、取历史最大、不过期、无时间窗口**。
- **速度归 teleport 轴单判**：任一帧 ≥200cm/s 跳 = teleport → 直接 ghost + purge 该帧（[[teleport_interference_purge_mechanism]]），**不并入轨迹、喂不到 R**。故快跳/单帧噪声尖峰上游先滤，R 只在存活连续帧上量净距 → 无需自带速度守卫。
- **实现排序**：teleport 轴须**先于** R 跑（purge 快帧），R 再量距。
- churn 免疫：新 lid 从自己 BirthPos 起算 R=0 → 冻结点重生 lid 永远 R≈0 → 拿不到 walkout 证。

## 独处无证 R<200 的判决（共存门定，非 R 定）

R<200 分不出"独处真睡者"与"phantom"，故 **ghost score 由共存定，不由 R 定**：

| 场景 | 判据 | `RealProven` |
|---|---|---|
| **有共存 certified-real 轨**（人已在别处有着落） | 这条冻结无证轨 = phantom | **false（ghost，PReal 显示按档）** |
| **独处** + vital 有 | 活体 | true（救睡者/静卧真人） |
| **独处** + E4 命中 | 丢-重抓睡者 | true |
| **独处** + 以上皆无 | 空房风扇 phantom | **false（默认 ghost，FP 优先）** |

ffdb phantom 共存（np≥2）→ ghost 铁；cd2b 独处睡者靠 vital/E4 救；空房独处 phantom 无 vital/E4 → ghost，不再走老 `ForceReal` 独处永发（那条恰会让空房 phantom 误火）。

## Latch 语义 + churn 免疫（治 D2）

- **sticky latch**：命中 E1~E5/R 任一 → `RealProven=true`，**单调不撤**。躺平/静止不撤 latch → **静止真摔者受保护**（他摔前已挣过证，latch 焊死）。
- **锚定 logicID + cell，不锚 lid**：
  - real latch 随 logicID 走（E3 传递）。
  - **ghost 态也持久化到冻结锚点 cell**：某 cell 上判定 ghost → 该 cell "hot"，其上新生的 glued 且无 E1~E5 的 lid **继承 ghost 默认**，不给它"从先验重新起步"的机会（堵死 FFDB00811215→FFDB02150835 洗白）。
  - **release**：仅 lid 从该 cell walkout≥200（R）解 hot——自主位移证明这里现在是真人。
- 解闩唯一路径 = split / swap 重分配（沿用 [[real_by_provenance_latch_optionb_area_drift]]）。

## 明确接受的 FN 边界（2026-07-07 定，FP 优先）

**接受**：独处真人从盲区/门外**未触发 EnterRoom** 地冒出、**立即静卧零移动、又读不到 vital**——E1~E5/R 全空 → 判 ghost。
- 依据：当前 FP 太多（90% 告警是 ghost 误报），必须先降 FP；此残余罕见（真人入房绝大多数经门 → E1，或先走动 → R，或有续轨 → E3，或有心跳 → E5，或丢-重抓 → E4）。
- 若 radar 仍能看到并经其它路 fire，无妨；看不到则认了。
- **不接受**且必须守住的是**高频**场景：固件重号真人（E3）、独处睡者丢-重抓（E4）、静卧真人有体征（E5）、真人静止（sticky latch）——这些是主干，不是边界。

## 现有机制收编映射

| 现有 | 处置 |
|---|---|
| realness mirror 几何轴（sep/rho/wall_margin/LaterBorn/Coexist mEv） | **删**（[[realness_axis_redefined_real_vs_mirror]] 退役） |
| `rcRealBase` baseline 恢复 | **删**（D1 重力源） |
| `rcWAuto` Displaced 真化 | **删**（并入 R walkout，判据同但不吃"伪 autonomous"） |
| split 定罪（几何打分 → mEv） | **收编**为 cell hot 的一个 seed（不再直接调 p_real） |
| floor-artifact latch / teleport purge | **保留**（track-lifecycle 轴，与本判据正交；immature/瞬移仍各治其段） |
| area 自查（raw newer-covers） | **降级**为 E1/E2 的区型来源 + cell hot 的先验 seed（AreaInterfer/Reflector 声明 → 直接 seed hot） |
| provenance latch（RealProven） | **升为唯一真化通道**（E1~E5/R 汇入它） |
| `PReal`（旧：判决 + 显示双职） | **拆轴**：`RealProven` bool 管判决；`PReal` 连续降为 FE 显示派生量（分档，gate 无） |
| `ForceReal` 独处永发（§G六 Coexist==0→real） | **删/改**：独处不再无条件 force real（那让空房 phantom 误火）；独处走 vital/E4 门（见「独处无证 R<200 判决」表） |

## fire 耦合（FP 真正降的落点）

- **F1 floor 占用门改读 `RealProven` bool**（不再读 `PReal≥0.5`）：`F1 占用 ⟺ RealProven ∧ S∉{E,L}`（[[engine_aggregation_floor_gate_f1_occupancy]]）→ 无证 phantom `RealProven=false` → 不计 floor 占用 → **floor 兜底不火**。这是三次误火被掐断的机制点。
- **§79 不破**：`RealProven` 仍只喂 N_r / F1 占用门，**不直接否决 fall**；静止真摔者靠 sticky latch 保持 `RealProven=true` → 照进 F1 → 照火。改的是"谁算 real"，不是"real 能不能否决 fall"。
- `PReal` 连续值**不参与**任何 fire 判决，纯 FE ×100 显示。

### 边界声明：本设计只作用于「引擎 floor 兜底占用路」
- **固件即时 fire（pose=5/6 Fallen → SFallen 发射）不受 `RealProven` gate**（[belief/filter.go:145] 铁律 [[realness_never_vetoes_fall]]：realness 只喂 N_r，不调制 SFallen 发射）→ 真人真摔固件报了，照火，§79 守住。
- 反面：phantom 若拿到固件 pose=5/6 也照火，**ghost 大改降不了固件误火**——那是**另一条独立问题**，走 auto-recover / verifier 95% 门（[[dbn_mode2_veto_via_autorecover]]），不在本 spec 范围。
- 本 case 三次 dbn.Fall 均 band=floor / pose=4（**非固件 Fallen**）/ producer=云端引擎 → 全在本设计命中的 floor 兜底路，不误伤固件路。

## 验收判据（case-ffdb-0707-02080247，replay fire 必须由 3→0）

| 对象 | 出生地 | 期望判定 | 验收点 |
|---|---|---|---|
| 真人 FFDB22017442 | (75,85) 关联/门第 | **real 全程** | latch 不掉 |
| 真人 FFDB00811215 | (45,45) 门口 | **real 全程**，换 tid 靠 E3 不掉证 | tid=0↔1 rebind 后仍 RealProven |
| phantom FFDB02150835 及 (135,205)/(125,205)/(145,195) 一串火点 lid | 房中央冻结 | **ghost 全程**，cell hot 后新生 lid 不洗白 | `PReal<0.5`，不进 F1，floor 不火 |
| 三次 dbn.Fall（02:35/02:47/02:59） | — | **replay fire=0** | grep xsensor_dbn_fire 空 |

## 实现顺序

1. **Phase A（拆轴 + 塌缩）**：删 mirror 几何轴 + `rcRealBase`/`rcWAuto`；`PReal` 判决职拆给 `RealProven` bool（N_r / F1 floor 门改读它）；`PReal` 降为显示派生量（按 R 分档）。realness 塌成"latch 驱动"：默认 `RealProven=false`，仅 E1/E2/R 置 true。replay 看 FFDB02150835 落 ghost（RealProven=false，不进 F1，fire 掐断）、门第/床真人保 real。
2. **Phase B（cell 锚定）**：ghost 态 cell hot（继承 + walkout release），治 lid churn 洗白（FFDB00811215→FFDB02150835）。
3. **Phase C（续轨/独处证）**：E3 handoff 继承（源 real∧离场 锁）验固件重号真人不掉证；E4 relost（房级 np0→1 ∧ stillbox ∧ 窗内无 Enter/Exit）+ E5 vital（独处无证 R<200 启用）验 cd2b 独处睡者不漏。
4. **Phase D**：逐字镜像 wisefido-sensor。

> 铁律：验证 = 跑真 case 看 replay fire / track_event_timeline，**禁 unit test**（[[validate_real_case_no_unit_tests]]）；看机制不看单帧 fire 阈（[[validate_real_case_no_unit_tests]] / 规则 #3）。

---

## 术语表（大白话，时间久了防忘）

### `rcRealBase` —— "把每条轨往真人方向拉的重力"（本次要删）
realness 滤波里的一个**无条件恢复率常量**。含义 = 默认偏见"没证据说你是假的，那就每 tick 悄悄把你往‘真人’挪一点点"。每帧加一丢丢，时间一长，**任何没被明确判死的轨都会慢慢飘到 real**。这就是冻结 phantom FFDB02150835 能从 0.75 爬过 0.5 变 real 的**地心引力（D1）**。本设计删它 → 改成"不挣出生证就永远不许往上飘"。

### teleport purge / floor-artifact latch —— 两个"删/压伪迹轨"的机制（生命周期轴，本次保留）
- **teleport purge（瞬移删）**：固件偶尔让一条轨**单帧瞬间跳几米**（≥200cm/s，人不可能）→ 一看就是雷达伪迹 → 直接删轨。删前两道闸（跳前姿态非倒地前兆 + 无质心骤降），防把"真人被干扰带跳"误删。本设计里它**先于 R 跑**，把快跳帧滤掉，R 只在存活连续帧量净距。
- **floor-artifact latch（地板伪迹闩）**：把"**刚出生没几秒（<5s）就冻住、悬在地板**"的短命孤儿标记压制、别触发 floor。**成熟度闸：只管新生的（<5s），活久了的不碰**（怕是真摔的人躺着不动）。→ 本 case 冻结 phantom **活了好几分钟=成熟轨**，正从这个成熟度闸的缝漏过去，所以它没被压住。

### area 自查（raw newer-covers）—— "引擎自己算这条轨踩在哪个区，不等固件"（本次降级为 seed 来源）
轨"在什么区（床/坐/门/干扰区）"本靠**固件 area 事件**定，但 ghost/丢轨轨**没有该事件**（area_id=255=无区）→ 区域失明 → floor 没床豁免误火。area 自查补这个：
- **自查**：引擎拿轨坐标去查 layout 网格**自己算**踩哪个声明区，不等固件。
- **raw（非 Kalman 平滑位）**：用**末次真观测点**——丢轨期 Kalman 会外推乱跑，raw 才是"人最后在哪"。
- **newer-covers（新盖旧）**：自查值 vs 固件事件值，**时间戳新的赢**（单调取最新），防旧盖新。
- **触发**：new-tid 出生 / enter-exit / np>0 时，该雷达每条轨各自查一遍。
- 出处 [[cd2b0707_ghost_nr0_area_selfcheck_raw]]；本设计里**降级**：不再主判据，只当 E1/E2 的区型来源 + 给 cell hot 播种（声明 AreaInterfer/Reflector 的格 → 直接种 hot）。

### 本 spec 其它记号
- **`RealProven`**：判决轴，bool，cert latch。N_r / F1 floor 门读它。硬 latch。
- **`PReal`**：显示轴，连续 0-1，纯 FE ×100 confidence，**gate 任何决策皆无**（拆轴后）。
- **cell hot**：某冻结锚点 cell 被判 ghost 后置"热"，其上新生 glued 无证 lid 继承 ghost 默认（治 lid churn 洗白）；仅 walkout≥200 解热。
- **R（walkout）**：离 born 最远净距（cm），纯距离、sticky、无时间窗口；≥200 补证。
- **D1 / D2 / D3**：三结构缺陷 = baseline-real 重力 / 定罪挂 lid churn 清零 / 五捕手 FN 逃逸口交集。
