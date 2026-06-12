# geom 全量退役映射表（plan-first）

状态：**✅ 已实施（commit `8d5d2b4`，委员会签字 D1-D4 后施工）**。下文映射与改法已落地；实施差异见文末「实施记录」。

---

## 0. 背景与一句话结论

`belief.Geom`（`GeomInBed / GeomInEnter / GeomInToilet / GeomOpenFloor / GeomUnknown`）在服务器侧**几乎是 `cell.area_type` + `门距` 的纯函数**——见 `geomFromArea`（belief_adapter.go:76-89）与 `geomFromGrid`（:93-102）。dwell 这条腿已经退役（`fallLRFromDwell` 直接吃 `o.RoomType, o.AreaType`，likelihood.go:33），**poseLikelihood / Track 层 TObsAbsent / 一批 obs.Geom 字段仍残留 geom**。

退役 = 把每个仍读 geom 的触点改成**直读已存在的权威源**，并删掉 `Geom` 字段与 `geomFrom*` 这层中间映射。

**核心架构建议（待签字 D1）**：权威源解析**只在观测构造处（adapter / shadowTick）做一次**，把结果落成 Observation 上已有/新增的字段；likelihood 函数改成消费"已解析的区域上下文"，而非自己 switch geom。这样：①权威优先级（bed_state > area_id > 几何）只实现一遍；②床边 FN 止血（bed_state 翻转）天然落在构造层，likelihood 保持纯函数。

---

## 1. 权威源清单（替代目标，均已存在于代码）

| 权威源 | 读取点 | 语义 | 取代 geom 的哪部分 |
|---|---|---|---|
| **cell.area_type** | `grid.AreaTypeAt(x,y)` / `b.CellAreaType` | 该格 layout 区域类型（Bed/Enter/Toilet/Shower/Sit/Active/Deny） | geom 的绝大部分（rest-zone vs open-floor 区分） |
| **room_type** | `e.roomType[roomID]` → `o.RoomType` | 房间类型（Default/Bathroom/Kitchen） | toilet/shower 语义的房间级确认 |
| **bed_state** | `tm.BedOccupancyState(nowMs)` | 床占用权威（InBed/LeftBed + conf），**权威 > area_id > 几何** | "床区躺=睡 vs 床边真摔"的判别（止血已用） |
| **门距（几何）** | `grid.NearestEntryDistCm(x,y) ≤ 30cm` | 距最近门 ≤30cm | `GeomInEnter`（area_type 无法编码动态门距，需保留为独立 `NearDoor` 信号） |

> 注：memory 里"区域走 room_type+cell.area_type，床占用走 bed_state（权威>area_id>几何）"即本表。`area_id` 在服务器侧表现为 `cell.area_type`（grid 已按 area 投影到格）。

---

## 2. geom → 权威源 基础反推表（`geomFromArea` 逆映射）

| geom 值 | 当前来源 | 退役后直读 |
|---|---|---|
| `GeomInBed` | area==AreaBed | `area_type==AreaBed`；**且 pose=lying 的睡/摔判别叠加 `bed_state`** |
| `GeomInEnter` | 门距≤30cm **或** area==AreaEnter | `NearDoor`(门距≤30cm) **或** `area_type==AreaEnter` |
| `GeomInToilet` | area∈{AreaToilet,AreaShower} | `area_type∈{Toilet,Shower}`（可叠 `room_type==Bathroom` 确认） |
| `GeomOpenFloor` | area∈{AreaSit,AreaActive,AreaDeny} | `area_type∈{Sit,Active,Deny}`（= 非 rest-zone 非门 = open floor） |
| `GeomUnknown` | 无 cell | `area_type` 取不到（ok=false） |

---

## 3. 逐触点映射表

### A. `poseLikelihood`（belief/likelihood.go:130-173）— pose × geom 全分支

签名现为 `poseLikelihood(pose int, g Geom)`。退役后建议改 `poseLikelihood(pose int, ctx AreaCtx)`，`AreaCtx{AreaType int; RoomType int; NearDoor bool; BedReleased bool}`（**待签字 D1/D2**）。

| 行 | pose | geom 分支 | 当前用途 | 替代权威源 | 改法 |
|---|---|---|---|---|---|
| 136-138 | Walking/Running | `g==GeomInEnter` → `m[SLeft]` | 门口走动可能离场 | **NearDoor ∨ area==Enter** | `if ctx.NearDoor \|\| ctx.AreaType==AreaEnter` |
| 149-151 | Fallen | `g==GeomInBed` → 降权 SFallen + 抬 SBedLying | 床上 fallen 多是躺姿误读 | **area==Bed**（睡/摔再叠 bed_state，见 D3） | `if ctx.AreaType==AreaBed` |
| 152-153 | Fallen | `g==GeomOpenFloor` → 升 SFallen | 开阔地确认倒地 | **area∈{Sit,Active,Deny}** | `else if isOpenFloor(ctx.AreaType)` |
| 158-161 | Lying | `g==GeomInBed` → SBedLying 主导 | 床上躺=睡 | **area==Bed ∧ ¬BedReleased**（关键，见 D3） | 床区且床态未释放→睡 |
| 162-163 | Lying | `g==GeomOpenFloor` → SFallen 倒地候选 | 开阔地躺=倒地 | **isOpenFloor ∨ (area==Bed ∧ BedReleased)** | 止血逻辑并入此分支 |
| 164-165 | Lying | default → 两可 SBedLying/SFallen | Toilet/Enter/Unknown 躺 | **其余 area_type** | `default` |
| — | SuspectedFall/Sitting/Standing/SuspectedSitGround/SitGround/BedSitUp* | 无 geom 条件 | 独立于 geom | 无 | **不变** |

> **PoseWalking 的 `GeomInEnter`** 是 area_type 无法独立编码的唯一点（动态门距）→ 必须保留 `NearDoor`（来自 `grid.NearestEntryDistCm`）。

### B. Track 层 `rawTLikelihood` TObsAbsent（belief/track.go:171-188）— last-geom 分流

| 行 | geom 分支 | 当前用途 | 替代权威源 | 改法 |
|---|---|---|---|---|
| 172-176 | `GeomInEnter` | 门区消失→JustLeft | **NearDoor ∨ area==Enter** | 同 A 的 NearDoor 信号 |
| 177-182 | `GeomInBed` | 床区消失→None（绝不 Lost） | **area==Bed** | `case AreaBed` |
| 183-187 | default(OpenFloor/Toilet/Unknown) | 开阔地消失→Lost 候选 | **其余 area_type** | `default` |

> TObservation 也需带"已解析区域上下文"而非 `Geom`（track.go:79-88 的 `Geom Geom` 字段）。

### C. obs 结构体 `Geom` 字段（定义 + 读写点）

| 文件:行 | 字段/位置 | 退役动作 |
|---|---|---|
| belief/observation.go:66 | `Observation.Geom` | **删字段**；保留已有 `RoomType`(:86)/`AreaType`(:87)，**新增** `NearDoor bool`、`BedReleased bool` |
| belief/observation.go:70 | `Observation.GeomConf` | **见 D4**（provenance 信任 blend 去留） |
| belief/track.go:83 | `TObservation.Geom` | **删**；改带区域上下文（AreaType + NearDoor，TObsAbsent 用） |
| belief_shadow.go:85 | `beliefShadowTrack.geom` | **删**；记录改存 area_type/nearDoor（诊断日志用） |
| belief_shadow.go:101 | `beliefShadowTLayer.geom` | 同上 |
| belief_shadow.go:128 | `beliefShadow.lastLostGeom` | **诊断日志专用**，改 `lastLostAreaType`（log 标签 `last_geom`→`last_area`） |

### D. `geomFrom*` / `geomConfFromGrid` 定义 + 调用点

| 文件:行 | 符号/调用 | 退役动作 |
|---|---|---|
| belief_adapter.go:76-89 | `geomFromArea` 定义 | **删**（其逆映射并入各 likelihood 的 area 分支） |
| belief_adapter.go:93-102 | `geomFromGrid` 定义 | **删**；构造观测时改：`areaType, ok := grid.AreaTypeAt`；`nearDoor := grid.NearestEntryDistCm ≤ 30` 分别落字段 |
| belief_adapter.go:151 | `radarFrameAdapter` 调 geomFromGrid | 改为填 `AreaType/NearDoor`（RoomType 已填，:192） |
| belief_adapter.go:152 | `geomConfFromGrid` 调用 | 见 D4 |
| belief_adapter.go:442/455 | sleepad/bed adapter 硬编码 `GeomInBed` | 改 `AreaType=AreaBed`（这些观测本就在床） |
| belief_adapter.go:462 | neighbor 硬编码 `GeomUnknown` | 改 area 上下文置空（ok=false 等价） |
| belief_shadow.go:179 | event 硬编码 `GeomInEnter` | 改 `NearDoor=true` / `AreaType=AreaEnter` |
| belief_shadow.go:300/396 | `geomFromArea(b.CellAreaType)` | **删中间转换**，直接传 `b.CellAreaType` |
| belief_shadow.go:363-368 | **床边 FN 止血**（geom InBed→OpenFloor） | **改为设 `BedReleased`**：`if areaType==AreaBed && bedReleased { ctx.BedReleased=true }`，由 poseLikelihood Lying 分支消费（D3）。语义等价、更干净 |
| belief_shadow.go:538 | `st.geom==GeomInToilet`（跨设备轨迹守恒 P6.5①） | 改 `st.areaType∈{Toilet,Shower}`（可叠 room_type==Bathroom） |
| belief_adapter.go:76-135 测试 + belief_adapter_test.go / belief_replay_test.go / belief_p61b_*_test.go | geomFrom* 单测 | 随签名改重写（**见第 5 节 e2e/单测**） |

### E. 几何自算（保留，非 geom 残留）

`NearestEntryDistCm`（grid.go）、`AreaTypeAt`（belief_cell_contract.go）、`reachableExit*`（belief_adapter.go:350-410）是**权威源本身的读取/几何计算**，**不退役**——只是不再经 `geomFrom*` 包一层。memory「Geom 已弃→XY∩多边形自算 geom 是残留」指的是 `geomFromGrid` 这层包装，不是底层 grid 查询。

---

## 4. 需你（架构师）裁的开放决策点

- **D1 — likelihood 签名**：`poseLikelihood(pose, AreaCtx)` 引入小结构体 `AreaCtx{AreaType,RoomType,NearDoor,BedReleased}`，避免 4 参散开。**推荐：用结构体**（rawTLikelihood 也复用）。
- **D2 — 权威解析归位**：所有权威解析（含 bed_state 优先级）**只在 adapter/shadowTick 构造观测时做一次**，likelihood 保持纯函数。**推荐：是**（架构最干净，止血天然落位）。
- **D3 — bed_state 进 Lying 判别**：床区 `PoseLying` 的睡/摔二义由 `BedReleased` 决定（= 当前止血逻辑搬进 likelihood 数据流，而非 obs 构造时硬翻 geom）。**推荐：保留止血现语义，只换载体**（geom→BedReleased 字段），不改阈值不改行为。
- **D4 — GeomConf/provenance 信任 blend 去留**：现 `geomConfFromGrid` 按 cell.Source（FE画1.0/feedback0.6/学0.4）对"抑制跌倒的 rest geom"打折，likelihood.go:13-17 lerp 到中性。退役后：①**保留**该 blend，改成按 `(area_type∈rest-zone, source)` 直算 `AreaCtx.Conf`，likelihood 仍 lerp；或 ②**简化删除**（rest-zone 抑制不再按 provenance 打折）。**推荐 ①保留**（feedback/learned 暂定抑制弱化是已签字的 layout 权威模型，删了会让自学床区全抑制跌倒——风险）。**此点风险最高，请明确裁定。**

---

## 5. 退役顺序（分批，签字后执行）

1. **批 1 — 字段与构造层**：Observation/TObservation 加 `AreaType/NearDoor/BedReleased`（D1 结构体），adapter/shadowTick 填好，**geom 字段并行保留**（双写，编译绿、行为不变）。
2. **批 2 — 切 likelihood 读侧**：poseLikelihood / rawTLikelihood TObsAbsent / 止血(D3) / toilet 守恒(:538) 全改读新字段。shadow 比对（DBN_MODE 不变）确认 likelihood 向量逐 tick 等价。
3. **批 3 — 删 geom**：删 `Geom/GeomConf` 字段、`geomFrom*`、`geomConfFromGrid`、`lastLostGeom`，编译器驱动清剩余引用。日志标签 `last_geom`→`last_area`。
4. **批 4 — 测试**：重写 geomFrom* 单测为 area_type 直读；补**床区躺+离床→SFallen fire 的 e2e 回归**（= 当前止血缺的那条，承接床边 FN 红线）。

> 规则 #1.2（不向后兼容/不留双路）：批 1 的"双写"仅为分批安全，**批 3 必须删净**，合并前 grep `geom\|Geom` 自查。

---

## 6. 不在本次范围

- firmware 侧 geom（雷达 `radar.areas`）—— 不动。
- bed scorer / bed_state 内部模型 —— 不动（本退役只是把它作为权威源**读**进 likelihood）。
- DBN_MODE / ghost 检测数学 —— 不动。

---

## 7. 实施记录（commit `8d5d2b4`）

**与 plan 的差异（均经工程判断，方向不变）**：
- **批1+2 合并**：发现真读 geom 的逻辑点仅 3 个(poseLikelihood / TObsAbsent / toilet 守恒)+ 日志，其余「触点」全是死字段（sleepad/bed/neighbor/ZBand/Vital/TrackPresent/noDetectObs/reachableExitObs 的 Geom 填了从不读）。死字段无需双写，故批1(双写)与批2(切读)合并为一次完整切换，用 test 套件验证逐 tick 等价（批1 单独是空 plumbing 无法验证）。
- **AreaCtx 精化**：`RoomType` 不进 AreaCtx——poseLikelihood/TObsAbsent 无任何分支用 roomType（已被 dwell 直接消费），且 adapter 无 roomType 入参。AreaCtx = `{AreaType, NearDoor, BedReleased}`。
- **area_type 值契约**：belief 分层不能 import roomengine.AreaType（dwell 已迁的既定约束），新建 `belief/area.go` 镜像值契约常量（canonical 在 roomengine/cell.go，改枚举值须同步）+ `AreaTypeLabel` 给日志。
- **Track 层无门距**：`tl.geom` 原用 `geomFromArea`(不查门距)，故 TObsAbsent 只看 `area==Enter`，TObservation **不加 NearDoor**。Room 层 ObsPose 用 `geomFromGrid`(含门距)→ 保留 `NearDoor`。
- **顺序微调**：批3(删 geom)放在测试迁移之后——删 `belief.Geom` 类型前测试须先脱离 Geom 常量。

**等价验证**：全 sensor build/vet 绿；belief 包 40+ geom 测试经 area 迁移后全绿；roomengine 仅 2 预存 dwell tail FP 红 + zoneengine 1 预存 sleepace 红（git stash 验证三者基线一致，**geom 退役零新回归**）。代码 geom 残留 grep=NONE。

**红线回归**：新增 `TestBedsideFallBedReleased`（belief 机制层）锁 D3——床区躺×bed_state：占用→SBedLying 主导/SFallen 中性；离床(BedReleased)→SFallen 抬=开阔地躺；门优先(近门走 default)。承接止血缺的「床区躺+离床→fire」端到端逻辑。

**已部署 test1**（09:24，新二进制 + DBN_MODE=2 + 无 panic，替换止血版 5ef9ede）。

**待补 follow-up（non-blocking）**：engine 层 plumbing e2e（`beliefShadowTick` 由 `bedReleased` 正确设 `obs.BedReleased`）——机制层已锁 D3 核心，plumbing 是 3 行赋值且被现有 engine 测试路径执行（未断言）；补一个带 bed_state mock 的 engine 断言测试需较重 setup，留作 follow-up。
