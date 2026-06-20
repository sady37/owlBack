# 执行批令:radar 床→SBed 用 firmware area_id 算 M×N

> 新会话照此执行。方向已由架构师拍定,**不要重开设计讨论**,直接实现 6 处 + build 验证。
> 工作目录:`/home/wisefido/owl/owlBack/tools/Xsensorv1`

## 目标(一句话)
radar 判"人在床上"→抬 SBed 的那一项,把 **N(在不在床)从 cell 几何换成 firmware area_id**;
**M(雷达和这张床的关系)= 已有的 covers**(emission 里就是权重 w),不动。
最终形态:`床→SBed boost = lArea × M(covers) × N(firmware area_id 命中床)`。

## 为什么用 firmware area_id 而不是 cell
- cell 是 canvas 几何,会 drift(真人漂到画歪的床框外 → cell 误判)。见 memory [[layout_drift_wall_expansion_policy]]。
- firmware area_id 是雷达自己的地面真值,不受 canvas drift 影响。
- 两者编号**本是同一套**:FE Save 时经 `declare_area` 用 `area.areaId` 下发设备(owlFront `radarDeviceApi.ts:233`),
  固件回报同一 id。(之前我比出"两套编号"是因为取了**事发后的 layout 快照**,非当时的——已澄清,是我的错。)

## 已确认的事实(直接用,不用再查)
- **radar.areas 位置**:canvas 里 `objects[].device.iot.radar.areas[]`,每项 `{areaId, areaType, objectId, vertices}`。
  **床 = areaType==2**(协议 §3.4.3:0无效/1自定义/2床/3干扰/4门/5监护床)。
- **firmware area_id 字段**:`owl-common/observation` 的 `Track.AreaID`(`track.go:14`,255=声明区外,0=无区/sleepad)。
  固件 track 帧**只带 area_id,不带 area_type**(window.json 实测),所以**必须有 areaId→是不是床 的映射**。
- **firmware area_id 已 plumb 一半**:`track_manager.go:45 TrackFrame.AreaType` 注释"雷达给的 area_id,engine 不信,用 cell"
  —— 即 ParseRadarTracks 解析后**没存进 track state**,snapshot(`TrackStatusBase`)只带 `CellAreaType`。要补存+暴露。
- **M 已存在**:emission `radarLogS` 的 `w = geom0Covers()` = 床 covers,床 boost 已是 `lArea × w`。N 现在隐式 = `o.AreaType==areaBed`(cell)。

## 6 处改动落点
1. **`internal/roomengine/layout_parser.go`**(`ParseLayoutConfig`):
   解析 `objects[].device.iot.radar.areas`,挑 `areaType==2` 的 `areaId`。按 objectId 对到对应床,
   产出与 `cfg.Beds` 对齐的 `BedAreaIDs []int`。(单床直接取那个 areaId。)
2. **`internal/roomengine/engine.go`**(`RoomConfig`,~行27):加 `BedAreaIDs []int`(与 `Beds` 一一对齐)。
3. **`internal/roomengine/track_manager.go`**:
   - `ParseRadarTracks` 把 firmware area_id 存进 track state(present 帧更新,**lost 用冻结值**)。
   - `TrackStatusBase`(:783)加 `FwAreaID int`,`SnapshotTrackStatuses`(~:853)填它。
4. **`cmd/xsensor/main.go`**:
   - roomGeom 带 `bedAreaIDs`,`FrameInput` 带上(新增字段或塞进 BedGeom 平行切片)。
   - TrackObs(:159 附近)加 `FwAreaID: b.FwAreaID`。合成 bed-track(:173)给 FwAreaID = 床的 areaId(自洽)。
5. **`internal/roomengine/adapter/adapter.go`**:
   - `RadarTrack` 加 `FwAreaID int`;`FrameInput` 带 `BedAreaIDs`。
   - `BuildObservation` 算 **N**:`FwAreaID` 命中某床的 areaId → 该床 N=1(255/未命中=0)。
     新增 observation 字段(建议 `RadarInBedFw bool` 或逐床 `RadarBedHitMask []bool`)。
6. **`internal/roomengine/belief/observation.go` + `emission.go`**:
   - observation 加上一步的新字段。
   - `radarLogS` 的 `switch o.AreaType { case areaBed: ... }`(:161-163)**床这一支改由 firmware N 驱动**:
     N=1 才 `addLogLk(SFallen 无关;SBed: lArea, w, SBed)`。`w` 仍是 covers=M。
     **sit/bath/active 三支保持走 cell `o.AreaType`**(只换床)。floor 阈也仍用 cell,不动。

## 细节约束
- area_id=255 → N=0(声明区外)。
- 多床:FwAreaID 命中哪张床的 areaId,就用**那张床的 covers** 作 M(不是 max)。单床无所谓。
- lost track 无新帧 → 用冻结的 FwAreaID(同 cell 的冻结处理)。
- 守 CLAUDE.md:删即删不留兼容;不写 WHAT 注释;`go build ./... && go vet ./...` 全绿才算完。

## 验证
- `go build ./... && go vet ./...`。
- 明天用户**重做测试**;export 时**同步导出 layout**(确保 fixture layout 与测试时固件同时点)。老 case 回放不用管(layout 已是未来快照,N 对不上是预期)。

## ⚠️ 重要:这一步**不修 cd2b FN**
cd2b 反射 track 几何/firmware 都在床区内 → **N=1**,M×N 仍高 → 反射照样钉 SBed → cd2b 还是漏。
M×N 是把"雷达床判定"做成正确的概率加权(架构师要的"不管用不用先实现"),**不是 cd2b 的解**。

## cd2b 真解(M×N 之后,2026-06-20 A+C 对码收敛 —— 权威见 `doc/DBN-Zone-Room.md` §11)

**收敛:cd2b 真解只新建两件,其余现成结构在干活。**

1. **① $\kappa$ 变活(wire `UpdateKappa`)**:现 orphan,$\kappa$ 死在几何冷启。落点=每帧每房 per-track `LogPsi` 前调,`matched/live` per-bed 从 15s K 窗事件算。$\kappa$=**纯归属权重**(决定 sleepad 床信息贴哪个雷达 logicID、贴多强),**不碰 SBed 维持**;无需防睡眠者衰减 gate(承重不变量见 DBN §4)。
2. **③ LeftBed→SOpenFloor(主路径)**:emission 的 LeftBed 现只推 $B{\to}$vac,补:$S$ 轴**主动满幅**抬 SOpenFloor,按 $a_j$ 归属($g^{xy}$ 几何门控多住户假摔)。落点按 track 态:在场→SOpenFloor / lost→SBlindRest / 先 Open 再 lost→SBlindOpen。

**不用单建(我曾当补丁加、已砍回)**:

3. ~~**牙齿 latch**~~ **删**:「LeftBed 一次性 vs 反射每帧抬 SBed」的持久压制**已是 $B$ 轴结构自带**——`bed_axis.go` `kObs` $B\text{vac}{\to}B\text{vac}{\approx}0.99$ 自持 + 只 InBed($L_{\text{in}}{=}20$)能扳回,正是"咬住、只 InBed 能松";配 $\Psi(\text{SBed},\text{vac}){=}1{-}o_j$ 每帧压。LeftBed 经 $B$ 轴是持久、每帧施加的,非一次性。
4. ~~**Q3 状态驱动兜底**~~ **作废,floor 不改**:`floor.go` `StillSec >= tFloorFor(area)` 是 **belief-独立天花板**(专为 belief 判错兜底)。"状态驱动 / per-state tFloor"会把兜底焊回它该兜的 belief 上=纵深防御塌一层,**不做**。床边摔(床矩形外=open cell)→ floor 12min 独立兜底,belief 卡 SBed 也照响;床矩形内+③失败=残留,靠实测验证③不靠改 floor。(`SBath→SEmpty=0` 是独立的转移完整性小修,有间接路 SBath→Left→Empty,低优可单独定。)
5. **sleepad-only(无 radar track)天生测不了 fall**(没 pose 抬不了 SFall)—— 接受,不设计。

**M×N 与 cd2b 的关系**:M×N(本批令)把"雷达床判定"做成正确概率加权,**不修 cd2b**(反射 firmware 也在床区→N=1→照钉 SBed);cd2b 靠上面 ①+③。
