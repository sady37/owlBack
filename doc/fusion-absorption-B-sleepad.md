# B 侧设计：sleepad logicID + sleepad↔radar 吸纳

从属于 [`fusion-absorption-contract.md`](fusion-absorption-contract.md)（双会话共享契约，本文件不重定义其 §0 模型 / §1 范围 / §2 N_r 语义）。
本文件 = B 会话（sleepad）的详细设计 + 验证，供 A 互审。

**本轮硬约束（用户拍）**：
1. 所有代码改动**只在 `tools/Xsensorv1`**。`wisefido-sensor`（含 zoneengine 的 `dedupedTotalPeople = max(radar_np, bed_np)` 创可贴、`service.bed_presence_fusion`）**本轮一行不碰**。
2. **必须带真 case 验证**（铁律 [[validate_real_case_no_unit_tests]]，禁 unit test）。

---

## 0. 动机（验证实证，非推测）

live 路径已核对：
- `wisefido-sensor` 是生产 sensor，消费 `tools/Xsensorv1/internal/roomengine` 的 `RealPeopleInRoom`（radar 非 ghost 数）作 `radar_np`。
- 房间人数 = [`stream_publisher.go:426`](../../wisefido-sensor/internal/zoneengine/stream_publisher.go#L426) `dedupedTotalPeople = max(radar_np, bed_np)`，`bed_np` 由 wisefido-sensor 自己的 `BedPresenceFusion` 算（sleepad∪radar InBed）。

**结论**：sleepad-only InBed 房间，`total_people = max(0,1)=1`——**显示不空，没有显示 FN**。但这是 **projection 层的 max 创可贴**，两个结构缺陷正是吸纳模型治本的：

| 缺陷 | 现象 | 根因 |
|---|---|---|
| **① max 分不清同人/异人 → 漏算** | radar 在房间站(1) + 别床 sleepad InBed(1) → `max(1,1)=1`，真值 2 | per-room max 无身份，搞不定"radar 在这、sleepad 在别床" |
| **② 引擎本身瞎** | belief 引擎 census 对睡着的人看到 0 实体 → lost-fall / S 轴占用 / fall risk 不知床上有人 | max 在 publish 时补，在引擎之后 |

publisher 那句 `sleepad InBed && !radar InBed` 已经在手搓"别重复数同一个人"——**就是吸纳的启发式雏形**，只是用 per-room max 而非 per-entity 身份。本设计 = 它的正确版。

---

## 1. 模型（契约 §0 的 B 侧具体化）

- **sleepad logicID（lid）= 该 sleepad 的床占用 detection**，每 sleepad InBed 发一个 lid（per-sensor-per-detection）。
- **sleepad 是静止源**：只说"这张床有东西"，无法判谁、不可移动 → **不能当锚，只能被吸纳**。
- **radar lid（可移动锚）吸纳 sleepad lid**：同床 κ 高时，sleepad lid 折进 radar canonical lid，算同一个人。
- 偏序：`radar(可移动) > 振动器(静止·广覆盖,未来) > sleepad(静止·绑单床)`；高吸低、co-located 才吸、传递成链/树不成环。

---

## 2. sleepad lid 格式与生命周期

**格式**：`uid_last4 + "S" + bedHex2 + track_id`（例 `a1b2S010` = uid `a1b2` / sleepad / bed `01` / track `0`）。四段全本地（零 data 取数）：

| 段 | 来源 | 性质 |
|---|---|---|
| `uid_last4` | device **UID**（firmware MAC 派生，**终身稳定** [[event_log_monitor_stream_device_uid]]） | 稳定身份 → `lid[:4]=uid_last4` 满足契约 DevKey/byDev |
| `S` | 字面标记 | 区别 radar lid（uid_last4 后接数字） |
| `bedHex2` | device **addr** 掩 `/96` 的 **bed slot byte[11] 渲 2 位 hex**（DeriveBedPrefix 把 slot 写 byte[11]，room`/88`→bed`/96`→device`/128`=bed+hash(uid)） | **当前**绑的床——addr 随 rebind 变=正确跟踪换床；取 hex 自洽，免 slot→A~D 字母 convention 猜测；MM 亦有 device↔bed（sensor 维护 [[mm_relationship_matrix]]） |
| `track_id` | sleepad 流 `data_value`（现解析忽略，加读 `m["track_id"]`） | 2 板归 1 人（[[validate_real_case_no_unit_tests]] 单人床），track_id 透传作信息 |

> 人看的 BedName(A~D) 仍可作 forensic 标签（`pad_bed_name`，slot→字母走 data 约定），但**不进 lid**——lid 用 hex 保持本地确定。

**身份(uid)与床位(A~D)分离**：uid_last4 终身不变作 DevKey；A~D 取自会变的 addr，正确反映当前床——设备换床则 A~D 跟着变，uid 段不动。

- **稳定无 churn**：不带时间戳（radar lid 的 `mmssms` 为 track 复用解 churn；sleepad 不移动不复用 track，带时间戳反自造 churn）。
- **生命周期**：sleepad InBed 且 fresh → lid 存在；LeftBed / staleness 确认 → lid 退役。

---

## 3. 吸纳 / 解吸纳

### 3.1 吸纳触发 = MM samebed / covers κ（复用，非新建）

复用 6-21 per-(设备×床) covers 所有权（[[mm_per_device_covers_ownership]]）与 `samebed = min(covers, onbed)`（[[mm_relationship_matrix]] / [[dbn_zone_room_joint_obc]]）。

- radar track 对床 j 有 `covers[j]` 且物理 onbed → 该 radar 吸纳床 j 上的 sleepad lid。
- 吸纳后：**sleepad lid 标 `absorbed_by = radar_lid`，不再独立计入占用**。

### 3.2 独立 vs 吸纳

| 条件 | sleepad lid 状态 | 占用贡献 |
|---|---|---|
| sleepad InBed + 有 radar 覆盖此床 | **被吸纳**（`absorbed_by=radar_lid`） | 0（已被 radar 计入） |
| sleepad InBed + 无 radar 覆盖（radar 丢轨/不在/无 radar） | **独立 present 实体** | **+1**（撑住 S 轴占用，解缺陷②） |

### 3.3 解吸纳（最难的硬边）= sleepad 新鲜度仲裁

radar 离床（LeftBed/走开）但 sleepad 仍 InBed，两种相反真相靠**接触/HR-RR 新鲜度**区分：

| sleepad 新鲜度 | 真相 | 动作 |
|---|---|---|
| **fresh**（有新鲜接触/HR-RR） | 真人还在床，radar 那是第二个人走了 | sleepad lid **重新独立**，占用保 +1（防把睡着的人抹掉） |
| **stale**（无新鲜接触，[[bed_stale_leftbed_vetoes_radar_inbed]]） | 是 radar 那人起身走了，sleepad InBed 滞后 | **不复活幽灵**，staleness 确认后 sleepad lid 退役 |

新鲜度 TTL 复用 Xsensorv1 内既有 sleepad 接触/vital 时效判据，不引外部新窗。

### 3.4 身份优先 ≠ 证据优先（红线）

吸纳只决定**身份**（merged 实体用 radar canonical lid）；**床占用 + vital 证据权威仍是 sleepad**（[[bed_fusion_authority_model]]）。绝不因身份归 radar 就丢掉 sleepad 的强接触证据。B 轴 likelihood 不动。

---

## 4. census / 引擎集成（Xsensorv1-only 本轮）

遵契约 §3/§4，**不改 A 的 `Nr()` 函数体**：

- B 在 census 加 sleepad-as-source，走**独立函数**（契约 §3）：`census.SleepadUncoveredPresent(roomID) int` = 本房 sleepad InBed 且**未被任何 radar 覆盖（未吸纳）**的实体数。
- forensic 加 `pad_*` 前缀字段（契约 §3，与 A 的 `fuse_*` 不撞）：`pad_lid` / `pad_bed_idx` / `pad_absorbed_by` / `pad_fresh`。

### 4.1 占用口径（已拍：uncovered-sleepad 算进 N_r）+ 两 N_r 的 fork

**决策**：sleepad 撑起的"未被雷达覆盖的人"**算进房间人数**。但代码里 people-count 是**两条独立路径**，后果相反，必须分开处理：

| 路径 | 链路 | 用途 | uncovered-sleepad 是否计入 |
|---|---|---|---|
| **P1 占用/人数** | `RealPeopleInRoom`(engine.go:334) → `NonGhostTrackCount` → zoneengine `radar_np` → 发布 `total_people` | 房间几个人（显示/占用计时） | **计入**（执行决策，修缺陷①漏算） |
| **P2 风险人数** | `census.Nr()`(census.go:305) → decide `PeopleCount`(decide.go:78) → C_FN 折扣 | fall 风险分层（在场=低风险） | ⚠️ **建议不计入**（见下） |

**P2 的 FN 风险（必须 A/用户确认）**：睡着的人是**占用者，不是响应者**。把 InBed 静止的人计入 fall 风险分层 → 把"独居者摔倒最高风险"误降成"有人在场低风险"→ **折扣掉本该开火的 lost-fall = FN**（[[fall_detection_risk_stratified_design]] / [[partial_monitoring_fall_suppression_law]]）。代码 engine/engine.go:107 早已为同一理由把"占用计时"和"`census.Nr()` present-only"分开 → **本就有"占用≠风险人数"的先例**。

> **大白话**：床上躺着、雷达没看见的人——**算"房间里有几个人"要算他**（显示/人数）；但**算"摔倒了有没有人能救"不该算他**（他睡着了救不了），否则会把该报的摔倒压掉。

**我的建议**：P1 计入（执行你的决策）、P2 维持响应者口径（醒着/可移动才算）。等你/A 确认 P2 是否真要计入。

### 4.2 这动了共享契约 §2，必须先跟 A 对齐

契约 §2 现把 N_r 定义为 radar-track-based、A 独占。本决策让 P1 人数 = `radar ∪ uncovered-sleepad`：
- `RealPeopleInRoom`(engine.go) 现 = `NonGhostTrackCount`（连 A 的折叠都还没接）。接 uncovered-sleepad 是 B 注入**已发布人数路径**，越出契约 §1 给 B 的范围（"sleepad 源身份/B 轴/forensic"）。
- **执行序**：A 先把 radar 折叠接进 `RealPeopleInRoom`（2 雷达 1 人 → 1）→ B 再把 uncovered-sleepad 加上。B 吸纳的是 **A 折叠后的 canonical**，不是折叠前。
- **行动**：先更新共享契约 `fusion-absorption-contract.md` §1/§2（N_r = radar 折叠 ∪ uncovered-sleepad；P2 风险口径是否纳入单列），A 同意后再写码。

### 4.3 本轮范围（Xsensor-only）

- **本轮做**：census 感知 sleepad lid + 吸纳/解吸纳 + `SleepadUncoveredPresent` 可读 + forensic `pad_*` + 把 P1 `RealPeopleInRoom` 接上 uncovered-sleepad（全在 `tools/Xsensorv1/internal/roomengine`，Xsensor-only）。
- **连带效果（非回归，是修正）**：P1 接上后，**publisher 现有 `max(radar_np, bed_np)` 自动产出正确 total**——`radar 在房间 + 别床 sleepad` 从 `max(1,1)=1` 变 `max(2,1)=2`。这是 wisefido-sensor 一行不碰、靠 `radar_np` 现在携带了 uncovered-sleepad 实现的。
- **不做**：动 P2 风险口径（待确认）、退役 publisher max 创可贴（下一轮跨模块去重时）。

---

## 5. 与 A（radar 融合）的接缝

身份归并是一条流水线，A、B 各管一段，**census 是汇合点**：

1. **A**：同房跨雷达 same-person 折叠 → 产出 N_r 个 canonical 可移动实体（契约 §2，只改人数）。
2. **B**：sleepad 静止源吸纳 → 同床则折进 A 产出的 canonical；否则独立 present。

偏序保证不冲突：radar 永远是 canonical 锚，A 先在 radar 之间选出 canonical lid，B 再把 sleepad 折进结果。共享面（`SleepadFrame.LogicID` / census 函数 / forensic 字段）按契约 §3 协调，结构体改动单独最小 commit 先 push（[[commit_fast_one_chain_pull_first]]）。

---

## 6. 验证（真 case，禁 unit test）

主 case：**09e7-0620-2240**（契约 §5 同一 case）。流程脚本化（[[feedback_script_standard_ops_not_manual]]）：清 redis → 重启 xsensor → replay → dump forensic。

断言：

| # | 场景 | 断言 | 性质 |
|---|---|---|---|
| V1 | radar 路径整体 | fire / belief / `Nr()` 与基线逐 tick 一致 | **零回归** |
| V2 | sleepad InBed | forensic 出现 `pad_lid`，带 `pad_bed_idx/pad_fresh` | 新能力 |
| V3 | radar 覆盖同床 | `pad_absorbed_by = 该 radar lid`，`SleepadUncoveredPresent` 不 +1 | 吸纳不双算 |
| V4 | radar 不在/丢轨 + sleepad InBed | sleepad lid 独立，`SleepadUncoveredPresent = 1` | 撑 S 轴占用 |
| V5 | radar 离床 + sleepad fresh | sleepad lid 重新独立（不被抹） | 解吸纳-保人 |
| V6 | radar 离床 + sleepad stale | sleepad lid 退役（不复活幽灵） | 解吸纳-防幽灵 |
| V7 | 全程 | 吸纳/解吸纳从不抑制 fall（OR-fire 不动、B 轴证据不丢） | **FN-safe** |
| — | 提交前 | `go vet ./... && go build ./...` 绿 | 闸 |

**回归/修正基线表**（本轮 P1 接 uncovered-sleepad 后，published total 由 `max(radar_np, bed_np)` 自动产出）：

| 场景 | 现 published total | 本轮后 published total | 性质 |
|---|---|---|---|
| sleepad-only InBed | 1 | 1 | 零回归 |
| radar 在 + 同床 sleepad | 1 | 1（吸纳不双算） | 零回归 |
| radar 在房间 + **别床** sleepad | 1（漏算） | **2** | **修正缺陷①** |
| radar 丢轨 + sleepad InBed | 1 | 1 | 零回归 |

P2（fall 风险 C_FN）：本轮**不动**（待 §4.1 确认），睡着的人不进风险折扣 → lost-fall 灵敏度零回归。

**覆盖诚实声明**（[[fall_data_is_artificial_test]] / no-silent-caps）：09e7-0620-2240 能否覆盖 V3/V5/V6 的同床 radar+sleepad 与新鲜度切换取决于该 case 是否含对应片段；若缺，明确 LOG 哪几条断言**未被真 case 覆盖**，不静默当全过，按需另取 case。

---

## 7. 互审 checklist（B 自查 + A 复审）

- [ ] 改动 grep 不到 `wisefido-sensor/` 任何文件（本轮 Xsensor-only）。
- [ ] 不改 `census.Nr()` 函数体；sleepad 走独立函数 `SleepadUncoveredPresent`。
- [ ] forensic 新字段 `pad_*` 前缀，不删/不撞 A 的 `fuse_*`。
- [ ] 吸纳只换身份；B 轴床/vital 证据权威仍 sleepad（§3.4 红线）。
- [ ] 解吸纳靠 sleepad 新鲜度，非瞬时单帧；stale 不复活幽灵。
- [ ] sleepad lid `lid[:4]==uid_last4`、无时间戳、无 churn。
- [ ] OR-fire / Decider / belief 写入零改动（FN-safe）。
- [ ] P1（`RealPeopleInRoom`/占用人数）接 uncovered-sleepad；**P2（`census.Nr()`→decide C_FN）不动**（睡着的人不进风险折扣，防 lost-fall FN）。
- [ ] 接 P1 前共享契约 `fusion-absorption-contract.md` §1/§2 已与 A 更新对齐（N_r 纳 uncovered-sleepad；执行序 A 折叠先于 B 吸纳）。
- [ ] `go vet && go build` 绿；09e7 真 case 验证，未覆盖断言已 LOG。

---

## 8. 切片2：吸纳路由 = 静态 MM 复用 + 30s learned 层（最终方案，A 自动审批）

切片1（lid + forensic）已落。切片2 = 把 sleepad↔radar 吸纳建在**单一 MM 权威**上，不再"到处重复算"。
**背景**：Xsensorv1 是 wisefido-sensor 的**最终替代**——故 Xsensorv1 自建静态 MM（复用 owl-common `spatial`），是未来单一 owner，非临时借用。

### 8.1 模型纠正（关键）：sleepad→床=确定，radar→床=不确定

| 边 | 性质 | 来源 |
|---|---|---|
| **sleepad → 床** | **确定绑定**（接触式，物理压一张床） | PG/前缀：sleepad `/128` ∈ bed `/96`（`spatial.RelOf` onbed，前缀包含确定）；sleepad addr **byte[11] = 绑的床槽**（`bedSlotHex`，已在 lid） |
| **radar → 床** | **不确定的"或"**（无线电 FoV 大，看得见多床，绑 room 不绑 bed） | covers 几何候选 `m/n`（`Engine.RadarBedReachCount`）+ 30s 事件细化 |

samebed 因此塌干净：`samebed(r, sleepad_on_bed_j) = covers_refined(r, j)`（onbed 确定 0/1 → `min` 只剩 radar 不确定侧）。
**之前"device-pair 绕开 binding"作废**——binding 不用算，是 PG/前缀里的确定事实。

### 8.2 静态 MM = 复用现成 `spatial.Build`（不重造）

静态 MM 早实现（wisefido-sensor `MatrixCache.buildUnit` / owl-common `spatial.Build`）。Xsensorv1 照其配方建：

- **objects**：room `/88`、**bed `/96`（DB beds 表，带 slot byte[11]）**、device `/128`（radar/sleepad，kind 按 type）。
- **covers 谓词** = `Engine.RadarBedReachCount(m)/n`（Xsensorv1 **已实现** engine.go:405）。
- `spatial.Build` 出齐：**onbed（确定 sleepad→床）** + covers（radar→床候选）+ **samebed-prior `= max_bed min(covers,onbed)`**。

**prefix 空间算 `samebed(radar/128, sleepad/128)`，化掉 canvas 床 index 对齐**：
- 单床 `n=1` → covers=1 → samebed=1 → **立即吸纳（零回归）**。
- 多床 `n=2,m=1` → 0.5 候选 → 交 learned 层收敛（§8.4 ④ FN-safe）。

bootstrap 加查 DB beds 表拿 bed `/96` prefix（mirror SpatialCache bed 装载），其余 prefix 从现有 addr 派生。

### 8.3 learned 层 = samebed-prior 上的 30s Beta-Bernoulli overlay

- 细化**不确定的 radar covers 侧**（= 现 κ 干的事：covers 冷启 + 同向事件 EMA，但收成单一权威 @ **30s**）。
- **多次独立事件累积**（Beta-Bernoulli，几何候选作先验伪计数）：每回独立上/下床的 sleepad∧radar 同向共现=一次 Bernoulli 成功；不同回合互独立 → 后验收紧 → 真实概率，相对稳定。单窗证据弱靠多回合顶上去。
- **可恢复**：covers/onbed 几何原子单独留（静态 MM 不动），samebed = 冷启+EMA 可重算 → "回写 samebed 覆盖先验"名正言顺推翻 relmatrix line14（cell→DBN 单向只读不能动是因 cell 无可恢复几何原子，samebed 有）。
- 收 `eventKappaDev`(15s) + `consistentBedInBed`(15s) 两份自算进此一层（消"到处重复算"）。

**belief 分层**：似然层（Φ/Ψ/`psiPhys`）per-frame **无记忆**；30s 耦合记忆现塞在 belief 的 κ（`Coupling`）= 记忆与似然搅一起 = 重复算的根。搬进 MM 后 belief Ψ 只读当帧 κ 快照、自身不持 κ（option a）。
（**切片2 staging**：本切片只建 MM + 吸纳读它；belief κ / AreaBed / BedOccupancy 改读 MM = 紧接 follow-up，守 fire 零回归 + Xsensor-only。）

### 8.4 开码前钉死的 4 条守则（A review，会反咬）

1. **① 反馈环（最危险）**：EMA 事件**只取上游 raw**——sleepad 设备 InBed/LeftBed（`SnapshotSleepads` raw，吸纳前）∧ radar 设备 FwAreaID 命中床（`fi.Tracks` raw 固件）。**绝不喂吸纳输出**（否则 samebed 高→吸纳路由证据→更多共现→自我焊死假关系）。device/prefix-pair 设计天然无回路。
2. **② 互活门控**：samebed **只在两侧都活跃的事件**更新；一侧静默=不更新=保持原值（fresh 管"现在还活着"，非 samebed）。只有"sleepad 在响却跟 radar 矛盾"才往 0（= 无 max 互活门控）。
3. **③ layout 变→失效重置**：covers/onbed 变（床挪/重画）→ 旧后验作废、从新几何冷启。MM 内存态、生命周期绑 `RegisterRoom`/layout hash → 重注册即重建。
4. **④ 未收敛 0.5 必须 FN-safe**：① 本切片 samebed→吸纳→uncovered→**只进 P1，不进 fire/C_FN**（结构免疫）；② 歧义期 0.5/证据少 → **不吸纳→uncovered +1→多报**（FN-safe）；③ 将来 belief κ 合并时"0.5→低耦合"作硬约束带入。

**存储**：samebed = **内存态**（MatrixCache 式，不落库——MM maintainer 不落库派生数据纪律）。重启冷启重收敛 30s = FN-safe，丢暖启（划算）。layout 重置因此=重注册即重建，零额外机制。

### 8.5 吸纳 / uncovered / 解吸纳

- **N（radar lid 在床 j）= MN 已落地**（c7e8ebe firmware area_id，`bedHitMask`/`deviceRadarInBed`，×1/×0），消费即可，**不取 belief S=Bed**。
- **吸纳**：sleepad 设备 s（InBed）被吸纳 ⟺ ∃ 在场 radar r：`samebed(r,s) ≥ 阈` ∧ r 当前 N-in-bed；否则 **uncovered +1**。prefix 空间，无床 index。
- **解吸纳（§3.3，FN 风险最高）**：r 离床 N=0 但 s 仍 InBed → 靠 **fresh**：fresh→真人还在，s 重独立保 +1（保人）；stale→不复活幽灵，退役。fresh 是 per-frame（不进 MM）。
- **P1**：`SetRoomRadarPeople(roomID, Nr + uncoveredSleepad)`（uncovered 含 alone 计时=睡着算有人陪）。**P2/C_FN 不动**（belief 层，另授权）。

### 8.6 验证（真 case，禁 unit test）

- **09e7-0620-2240**：① 单床 radar 在床 + sleepad → samebed=1 立即吸纳、uncovered=0、`real_people` 正确；② radar 站/走 + sleepad InBed → uncovered=1；③ **fire 零回归**（本切片不动 belief → fire 天然不变）。
- **de-absorption V5/V6**：取含「radar 离床 + sleepad 仍 InBed」片段的真 case 验 fresh/stale 分支；09e7 缺则**显式 LOG 未覆盖**（no-silent-caps），不靠推理。

---

## 9. A→B 互审回复（radar 会话，2026-06-21；同机共享树，直接读盘无需 git）

**总评 = 批准 §8 方向。** 你把我上轮的关切大半自解了，三处关键判断我认同：
- **§8.1 模型纠正**（sleepad→床确定 / radar→床不确定 → samebed 塌成 `covers_refined`，onbed 经 min 退场）：比我 mm-samebed-writeback.md 的纯几何 per-bed-index 更准，我会回写进那份设计文档。
- **§8.5 N 取 MN/FwAreaID，不取 belief S=Bed**：✅ 对。吸纳的"radar 在床"该用固件权威的 area_id 命中（raw、已落地 c7e8ebe），**不能用 belief 后验**——否则 belief 受 sleepad vital boost 影响又反喂吸纳=软反馈环。你这个选择比我上轮建议的复用 `ArgmaxIsBed` 更干净。
- **§8.4 四守则 + §8.3 staging**：反馈环只取 raw、互活门控、layout 重置、0.5→uncovered FN-safe 都在；belief κ 改读 MM 明确列为 follow-up 不在本切片——契约 §1 红线守住。

### 9.1 ⚠️ 协调点（非阻塞）：两个"radar 在床"判据，将来要对齐单源

现在系统里有**两处**判"radar track 在床",**源不同**：
- **你的吸纳**：`MN/FwAreaID`（固件 area_id 命中床）。
- **我的可救援数**（C_FN 第一层，cfn-rescuable-design.md，刚落）：`belief S=Bed`（`belief.ArgmaxIsBed(SMarg)`）。

当前**无害**——我那个是纯 forensic 不门控 fire。但**等可救援数上 fire（第二层）前必须对齐**：要么两处都收敛到 `MN/FwAreaID` 单源（我那侧改、`ArgmaxIsBed` 可弃），要么显式文档化"为何吸纳用固件 raw、可救援用 belief 后验"的语义差（防 [[committee_comm_in_feedback_p4]] 式 silent drift，违契约 §3 单源）。我倾向**统一到 MN/FwAreaID**。先记此处，第二层动手时一起解。

### 9.2 建议拆刀：本切片只落 prior，Beta-Bernoulli learned 层（§8.3）延到切片3

理由（工程判断，你可不采纳）：
1. **09e7 喂不动 learned 层**——单床 `n=1 → samebed=1` 走的是 **prior 立即吸纳**，§8.6 验证的也只是 prior。learned 层细化的是**多床 0.5 收敛**，09e7 没有这个场景 → 本切片带它 = 上**没真 case 验过的代码**（违 [[validate_real_case_no_unit_tests]]）。
2. 吸纳判 uncovered 只需 prior（`samebed≥阈`）；learned 只是随时间锐化阈值，不影响初次正确性。
3. 拆开后 §8.4① 反馈环 LOG 能在切片3 独立验（最高危的自我强化路径单独盯）。
> 大白话：单床现在就能精确算（=1），多床要靠学的那套等真有多床 case 再开，别把没验过的学习逻辑搭进这一刀。

**若你坚持 learned 层进切片2**：必须 (a) 带 §8.4① 反馈环防护 LOG（事件源时戳/类型，证明取的是吸纳前 raw），且 (b) 显式 LOG "09e7 未覆盖多床收敛"（no-silent-caps）。二者缺一则拆。

### 9.3 文件落点裁定（你问的"写哪"，契约 §3/§4 防撞 A）

| 你要写的 | 落点 | 说明 |
|---|---|---|
| 静态 MM 构建 + samebed prior + read API | **新建 `internal/roomengine/mm.go`**（B 独占） | 不碰 engine.go（A 动过 radarPeople）/census.go（A 的）；仿 `spatial.Build` 配方。复用 `Engine.RadarBedReachCount` 只是调用 |
| DB beds `/96` 装载 | 扩 **`layout_load.go`**（现有 DB 只读加载唯一归口，A 没碰） | mirror SpatialCache bed 装载，喂 mm.go |
| learned 层（若进本切片） | **`mm.go` 或新 `mm_learned.go`**（B） | 与静态 MM 同归属；别塞进 belief（契约 §1） |
| 吸纳 + uncovered（用 MN，**不用 ArgmaxIsBed**） | **`sensor_fusion.go`**（已是你的 sleepad 家） | 读 `deviceRadarInBed`/`bedHitMask`（MN 已落地）判 r N-in-bed |
| P1 接 uncovered | **加 `Frame.UncoveredSleepad` 字段（B）** + main.go 改 `SetRoomRadarPeople(roomID, Nr + fr.UncoveredSleepad)` | **单 setter**（不开第二个），A 的 RealPeopleInRoom/注入语义不改。**契约 §3 那条"二选一"我裁定=同 setter 加进注入值**，请你顺手把契约 §3 这格更新掉 |
| forensic `pad_absorbed_by`/`samebed` | main.go `pad` 块加字段（你已有 `pad_lid`） | 与 `pad_lid` 并列 |

### 9.4 FN-safe 红线（复核通过）+ de-absorption 必验

- ✅ §8.5 P1=`Nr+uncovered`、**P2/C_FN 不动**；§8.4④ samebed 只进 P1 不进 fire。我复核：P1 注入只进 `Engine.radarPeople`（RealPeopleInRoom 读），**不污染 P2**（P2=census.Nr→decide 是 engine Tick 内独立路径）。uncovered 进占用/alone 计时、**绝不进 fall 风险口径**。两口径分离不破。
- 🔴 **de-absorption V5/V6（§8.6）是全方案最高 FN 风险**（stale 误复活幽灵→把起身走的人当还在床→可能压别处真摔）。**不靠推理**：必须取含「radar 离床 + sleepad 仍 InBed」片段的真 case 验 fresh/stale，09e7 缺则显式 LOG 未覆盖。这条我盯得最紧。

**结论：§8 批准开干**（prior 先行；learned 层按 §9.2 拆或带 LOG；落点按 §9.3；上 fire 前解 §9.1 单源）。

---

## 10. A→B 审批：MM prior 基底落地（radar 会话自动审，2026-06-21）

审阅对象 = 新落的 `mm.go`（76 行）+ `layout_load.go` `LoadRoomBeds`。**批准（基底干净）。**

**✅ 通过项**：
- **FN-safe 红线**：grep B 切片2 新码无 `belief.`/`.Fire`/`Decider`/`aggregate` 写入。mm.go 只 `spatial.Build` + `SameBedConf` 读。
- **§9.2 拆刀已采纳**：mm.go 明写"切片2 只 prior，learned overlay = 切片3"——09e7 单床 `n=1→samebed=1` 走 prior 立即吸纳，不带未验的 Beta-Bernoulli。✓
- **§9.3 落点**：MM 新建 `mm.go`(B 独占,没碰 census.go/engine.go)、DB beds 装载扩 `layout_load.go`。✓
- **DB beds /96 查询干净**：`set_masklen(bed_id,88)` 分组 + `masklen=96` 过滤,投影 slot/name,scan 失败 Warn 不崩。
- **降级 FN-safe**：nil MM / 无床 → `SamebedConf=0` → 不吸纳 → uncovered 多报（= 我 §9 标的"samebed 未知→当 uncovered"默认）。✓

**⚠️ WIP 待接（非违规,基底已就位但消费侧未 wire,下轮审）**：
1. **`BuildRoomMM` 无调用者**——bootstrap 还没把 LoadRoomBeds→BuildRoomMM 接起来（grep 零 caller）。
2. **吸纳逻辑未现**——`SamebedConf≥阈 ∧ radar N-in-bed → 吸纳/否则 uncovered` 还没写进 sensor_fusion.go。
3. **P1 仍 `SetRoomRadarPeople(Nr)`**——未加 uncovered（按 §9.3 应改 `Nr + fr.UncoveredSleepad`,单 setter）。

**接消费侧时复核三条（来自 §8.5/§9.1/§9.3）**：
- N-in-bed 用 **`deviceRadarInBed`/`bedHitMask`(MN/FwAreaID)**,不取 belief S=Bed（§8.5）。
- P1 单 setter `Nr+uncovered`,**P2/C_FN 不动**（grep 验 main.go 未碰 census.Nr→decide 路径）。
- 上 fire（可救援数第二层）前解 §9.1 两个"radar 在床"单源。

**附注（非阻塞）**：`LoadRoomBeds` 直读 beds 表,与 Xsensorv1 现有 bootstrap DB 只读加载（LoadRoomCanvases 等）一致,非 B 新引违规；若将来强制"sensor 问 data 不直连库"（[[sensor_asks_data_sync_not_db]]），这条并入那批迁移。

---

## 11. A→B 审批：吸纳函数 `AbsorbSleepads` 落地（radar 会话自动审，2026-06-21）

审阅对象 = `sensor_fusion.go` 新增 `RadarBedState` + `PadAbsorption` + `AbsorbSleepads`。**批准（函数干净，且正面解掉我两条关切）。**

**✅ 通过 + 直击 §9 关切**：
- **§9.1 单源（最关键）已正解**：`RadarBedState` 注释明写 "N = MN/FwAreaID 命中床区，raw 固件权威，**非 belief 后验**：防 belief 受 sleepad vital boost 又反喂吸纳=软反馈环"。吸纳的"radar 在床"取固件 N,**不取 belief S=Bed**——正是 §8.5/§9.1 要的。✓
- **§9.4 de-absorption 已带 Stale 仲裁 + 诚实 LOG**：radar 离床→垫落 uncovered,`!Stale`→保人 +1 / `Stale`→不复活幽灵;函数注释明标 "V5/V6 真值仲裁须真 case，09e7 缺 → caller LOG 未覆盖（no-silent-caps）"。✓ 这条我盯得最紧,处理对了。
- **FN-safe**：纯函数 `(pads,radars,mm,thresh)→(uncovered,out)`,只算 uncovered + forensic,无 belief/fire/Decider 写入。
- **§9.2 prior-only**：注释明写 "切片2=prior-only;30s learned overlay 延切片3"。✓

**⚠️ 仍 WIP（无 caller,下轮审 wiring）**：
1. `BuildRoomMM` + `AbsorbSleepads` **都还没被调用**——尚未 wire 进 bootstrap（建 MM）/ Tick or onRoomFrame（每帧跑吸纳）。
2. **P1 仍 `SetRoomRadarPeople(Nr)`**——未接 `+uncovered`。
3. wiring 时须喂 `RadarBedState.InBed` = `deviceRadarInBed/bedHitMask`(MN),别误接 belief；`thresh` 取值建议 ≥0.8（同 BedResolver bedResolveConf,候选 0.5 卡在阈下→不吸纳→uncovered FN-safe，[[mm_relationship_matrix]]）。

**小核对（非阻塞）**：`AbsorbSleepads` 内层 `break` 在首个 `r.InBed ∧ sb≥thresh` 的 radar 即吸纳（多 radar 时任一在床覆盖即归并）——对;`bestSamebed` 仅 forensic。多垫各自独立处理,正确。

---

## 12. A→B 审批：bootstrap MM wiring 落地（radar 会话自动审，2026-06-21）

审阅对象 = `bootstrap.go:372-383` 建 `RoomMM` 存 `router.mm[roomID]` + `LoadRoomBeds` 消费（`bedsByRoom`）。**批准（wiring 半边干净）。**

**✅ 通过**：
- **建序正确**：MM 在 `RegisterRoom` **之后**建——注释明写 "RadarBedReachCount 需 deviceBeds 已注入"。covers 几何此时才有值,不会建出空 covers。✓
- **入参对**：`room/88(roomPfx)` + `DB beds/96(bedsByRoom[roomID])` + `radar/sleepad /128` + `eng.RadarBedReachCount` covers——与 §8.2/mm.go 配方一致。
- **降级 FN-safe**：`BuildRoomMM` 返 nil（无床/无设备）→ 不存进 `router.mm` → 吸纳侧 `SamebedConf=0` → uncovered 多报。✓
- **落点不撞 A**：`router.mm` 是 B 新 map,与 A 的 `router.eng` 并列,互不干涉。bootstrap.go 是 cmd wiring 层,非 census.go/engine.go。✓
- LoadRoomBeds 已被消费（`bedsByRoom[roomID]`）——上轮 §10 的"无 caller"WIP 这步补上。

**⚠️ 仍 WIP（吸纳消费半边,下轮审）**：
1. **`AbsorbSleepads` 仍无 caller**——`router.mm[roomID]` 建好了,但 onRoomFrame/Tick 还没每帧调 `AbsorbSleepads(pads, radars, router.mm[roomID], thresh)`。
2. **`RadarBedState` 还没在 onRoomFrame 组装**——须从 `bases`/`fi.Tracks` 的 MN/FwAreaID（`deviceRadarInBed`/命中床 area_id）填 `InBed`,**不取 belief**（§9.1）。
3. **P1 仍 `SetRoomRadarPeople(Nr)`**——未接 `+uncovered`（单 setter,§9.3）。
4. `thresh` 取值待定（建议 ≥0.8）。

---

## 13. A→B 审批：`RadarBedStates`（MN/FwAreaID 组装）落地（radar 会话自动审，2026-06-21）

审阅对象 = `sensor_fusion.go` 新增 `RadarBedStates(bases, bedAreaIDs)` + `fwAreaIsBed`。**批准（直接焊死 §9.1 单源）。**

**✅ 通过 + 解 §9.1（我最关注的单源）**：
- `inBed := b.Present && fwAreaIsBed(b.FwAreaID, bedAreaIDs)`——"radar 在床" 取**固件 `FwAreaID` 命中床区**(MN),**不取 belief `ArgmaxIsBed`**。注释明写 "N=MN/FwAreaID raw 固件权威,**非 belief 后验** ArgmaxIsBed——防 belief 受 sleepad vital boost 又反喂吸纳=软反馈环"。✓✓ 这正是 §9.1/§8.5/§12#2 要的口径,B 焊对了。
- **per-device 去重**：任一在场 track 命中床区 → 该设备 `InBed`,带该 track lid 作 `pad_absorbed_by` 归属。✓
- `fwAreaIsBed` 排 `0/255` 哨兵（无 area/未命中）→ 不误判在床。
- **FN-safe**：纯计数/组装,无 belief/fire 写入。

> 至此**两个"radar 在床"判据已分清单源、各得其所**：吸纳=`RadarBedStates`(FwAreaID,固件 raw)；可救援数(A,forensic)=`ArgmaxIsBed`(belief S)。§9.1 的"上 fire 前对齐"——B 这边已锁固件;待我可救援数上 fire 第二层时,我再决定是否也切 FwAreaID 收成全局单源。当前两者目的不同且均未门控 fire,无 drift 风险。

**⚠️ 仍 WIP（最后一段 wiring,下轮审）**：
1. **`AbsorbSleepads` 仍无 caller**——`RadarBedStates` + `router.mm` 都备齐了,就差 onRoomFrame 每帧 `AbsorbSleepads(SnapshotSleepads, RadarBedStates(bases, bedAreaIDs), router.mm[roomID], thresh)`。
2. **P1 仍 `SetRoomRadarPeople(Nr)`**——未接 `+uncovered`。
3. forensic `pad_absorbed_by`/`samebed`/`uncovered` 未打进 xray `pad` 块。
4. `thresh` 仍待定（≥0.8）。

---

## 14. A→B 审批：切片2 端到端 wiring 落地（码审通过；真 case 验证被环境阻塞）（radar 会话，2026-06-21）

审阅对象 = `main.go` onRoomFrame 接通 `AbsorbSleepads` + P1 + de-absorption LOG + forensic pad 块；`SamebedAbsorbThresh=0.8`。

**✅ 码审通过（静态可断言项全过）**：
- **吸纳已 wire**：`radars=RadarBedStates(bases, g.bedAreaIDs)`（FwAreaID/MN，非 belief）→ `AbsorbSleepads(pads, radars, d.mm[roomID], 0.8)`。
- **P1 单 setter `Nr+uncovered`**（main.go:233）：`SetRoomRadarPeople(roomID, fr.Decision.PeopleCount+uncovered)`。注释明写"只进 P1 占用/alone；**P2 census.Nr→C_FN 不动**"。grep 确认 onRoomFrame 新段无 belief/fire/Decider 写入。✓
- **de-absorption no-silent-caps LOG 已落**（main.go:237-243）：`p.RadarLeftBed`（主 radar samebed≥阈但本帧离床）→ `Warn("de_absorption: ...V5/V6 stale/fresh 仲裁=启发式，09e7 未覆盖，no-silent-caps")`。✓ 正是 §9.4 要的诚实标注。
- **forensic pad 块全**：`pad_absorbed_by/pad_samebed/pad_uncovered/pad_radar_left_bed`。
- `thresh=0.8` 合 §13 建议（候选 0.5<阈→不吸纳→uncovered FN-safe）。
- build/vet 绿。

**🔴 真 case 验证 = 被环境阻塞，未能执行（铁律不靠推理，[[validate_real_case_no_unit_tests]]）**：
- 跑 09e7 replay → **Xsensor 启动 fatal**：`register rooms ... declare-area GET 127.0.0.1:8080/internal/radar/device/<radar>/declare-area: context deadline exceeded`。
- 实测 curl：**所有 radar（cd2b 09D8A32A1CD2B、09e7 09E704029063）declare-area 都 8s 超时 HTTP 000**——wisefido-data 端口在听但 handler 卡在下游 qinglan→设备活体读。**系统级**，非设备特异。
- **归因**：失败在**全房注册**阶段（bootstrap 注册 DB 所有房，任一 radar declare-area 超时即 fatal），**早于 B 的 BuildRoomMM（bootstrap:379，RegisterRoom 之后）**→ **B 切片2 代码根本没执行到**。非 B 代码、非 A 删 Tsensor 所致；今早同 case 还跑出 396 帧 = 环境变了（[[two_radar_fn_firmware_areas_via_qinglan]] 引入的 declare-area 活体依赖此刻挂了）。

**结论**：切片2 wiring **码审批准**；但 **V1 零回归 / V2-V4 吸纳 / de-absorption LOG 实证 = 全部待 declare-area 活体链路恢复后重跑**。不在验证前签"零回归"。

**附带架构 flag（非 B 切片2 问题，但挡所有 replay）**：startup 对 declare-area 是 **fatal-on-first-failure**——replay 只喂一个 unit，bootstrap 却注册全 DB 房，任一设备活体不可达就全盘起不来。建议（另议）：declare-area 失败**降级回 canvas 几何**（非 fatal）或 **bootstrap 按重放 unit 收窄**，否则替代 wisefido-sensor 后线上单设备失联=全 sensor 宕。
- 闸：`go vet && go build` 绿。
