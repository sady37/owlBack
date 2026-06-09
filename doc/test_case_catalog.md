# 基础测试模块 —— 真碎片案例目录（ground-truth 确认）

> 用途：① Tier-1 真碎片复合测的积木来源；② 未来 AI 模拟生成器的素材库。
> 硬约束（用户）：只收 **① 验证过的 test case**（真摔/假fall，人工确认 ground-truth）+ **② 无报警的真实正常数据**（enter/left room·bed、HR/RR；「没报警」=可信「正常」）。
> 本目录 = 用户 2026-06-09 确认信息录入；`⚠` 标施工方待确认项。**裁前不建**（模块/harness 待委员会受理）。

## 一、Unit × Device 编组（★replay 核心：知道 unit 全 device 才能喂全 unit 实时流+事件）

> 「以 case 里说明为准」（用户）——下表据 case NOTES + measure-* 目录核；⚠=待用户/NOTES 补。
> 寻址：09e7=`9D8A326309E7` / D523=`E598A2ACD523` / d5f7=`E598A2ACD5F7` / CD2B=`9D8A32A1CD2B` / 1641=sleepad`BM87224601641` / 0978=sleepad`BM87224700978` / CABB=`…10D5CABB` / 333B=`…333B`。

| Unit | 房间 | Device（radar/sleepad）| 备注 |
|---|---|---|---|
| **101 / John.Y**（`fd00:0:3:111:3::/80`）| bedroom | Radar **09e7**、Radar **D523**（双雷达房）、sleepad **0978** | #1/#2/#13 |
| | bathroom | Radar **d5f7** | #14（d5f7 多径干扰误报） |
| **201**（`fd00:0:3:111:…`⚠）| bedroom | Radar **CD2B**、sleepad **1641** | #7/#8/#10/#11/#12 |
| | bathroom | Radar **333B** | #9 |
| **hunzi**（**深圳**，与 101/201 **完全独立**，**不可跨 unit 拼**）| bathroom | Radar **CABB** | #5 |
| | bedroom | sleepad **0672** | — |
| ~~Ton~~ | — | ~~FFDB~~ | **#3 drop（用户）** |

> ★地理隔离（用户）：**101/John.Y + 201 在 Denver；hunzi 在 Shenzhen**。Neighbor/全-unit 拼接**只能同 unit 内**，三方互不为邻。

## 二、案例目录（Trigger/alarm | real/false | unit | 摘要 | unit 全 device | fixture）

**筛选铁律（用户）**：只留**清晰无歧义**的；**有 NOTES 的优先**（当时分析过、场景较真）；ground-truth 未定 / fixture 对不上 / 无据 → 不入核心库。

### 核心库（清晰 + ground-truth 明确；NOTES 优先）

| # | 时间(trigger) | real/false | unit·房·**触发 device** | **同 unit 其余 device(neighbor)** | 摘要 | fixture | NOTES |
|---|---|---|---|---|---|---|---|
| 1 | 0605 11:37–11:46 MDT | **REAL·漏报** | 101 bedroom · **09e7** | D523(bedroom)、d5f7(bathroom)、sleepad 0978 | 床边跌倒；雷达误识旁 chair；pose≈3 firmware **未报**；自救取消零记录 | `bedtest-0605-1-bedside-fall-no-fw-detect` | ✓ |
| 2 | 0606 19:08–19:18 | **REAL·fw判出** | 101 bedroom · **(radar)** | 09e7/D523、d5f7、sleepad 0978 | 床边跌倒；身仍靠床→sleepad 检 HR/RR；firmware pose=5 | `bedtest-0605-2-bedside-fall-fw-detect` | ✓ |
| 13 | 0526 21:00–21:24 | **FALSE·ghost** | 101 bedroom · **D523** | 09e7、d5f7、sleepad 0978 | 镜面 ghost 误报 | `d523-mirror-ghost-0526` | ✓ |
| 14 | 0524 13:35–13:57 | **FALSE·多径** | 101/John.Y bathroom · **d5f7** | 09e7、D523、sleepad 0978 | bathroom 多径干扰；pose=3 sit 却报 Fall（缺 bathroom card） | `d5f7-bathroom-fall-1335` | ✓ |
| 5 | 0529/0530/0601 | **FALSE·lost** | hunzi bathroom · **CABB** | ⚠ hunzi 其余 device 待补 | lost track 误报（3 窗） | `hunzi-cabb-lost-{0529,0530,0601}-FP` | ✗(用户确认) |

### ★ #9 黄金真 hand-off —— ✅ 已导出（2026-06-09，DB owl_v2）

| # | 时间(MDT) | label | unit·房·device | fixture |
|---|---|---|---|---|
| 9 | 0609 07:14–07:28 | **REAL·recovered + 真跨房 hand-off** | 201 · bathroom **333B** + bedroom **CD2B**（同 `/80`，neighbor sleepad 1641） | `unit201-handoff-0609-bathroom-333B`（264行）+ `unit201-handoff-0609-bedroom-CD2B`（93行） |

**确认的真因果时间链（event_log 实查）**：333B bathroom `EnterRoom 07:15:30 → Walking 07:15:41 → Fall 07:16:11 → Fall 07:17:32 → ExitRoom 07:17:50` ‖ CD2B bedroom `EnterRoom 07:17:46 → ExitRoom 07:18:08`。**两房事件差 ~4s = 真实人 bathroom→bedroom 移动**，非拼接 → **Tier-2 级真验素材**（Neighbor recall 黄金：真摔+firmware fire+起身 recovered+真 hand-off 一体）。

> ★★ **数据墙解封（2026-06-09）**：event_log/monitor_stream 在本地 `owl_v2`（`DB_PASSWORD=postgres` 在 .env）**全可查**；`tools/redis-replay`（多 device-uids→重放回 redis 实时流）即委员会要的「整单元 replay 工具」；`scripts/export_case_v2.sh` 导任意窗口。**「blocked on unit201」实为缺工具/密码,非真 block** → #9 已导，benign 挖掘 + 任意真案导出 + Tier-2 整单元 replay 均**现在可做**。

### 排除（歧义/无据/不要轻易用）

| 原# | 案 | 排除理由 |
|---|---|---|
| 3 | Ton FFDB 沙发矮坐 | **用户 drop**；无 fixture |
| 4 | 101 D523 edge fall 09:59 | **用户去**；fixture 时间对不上 |
| 6 | 101 17:38 bed fall | **= #1**（11:37MDT=17:37UTC 同案，已并） |
| 7/8 | `cd2b-fall-0607-0127`/`-0606-0917` | **用户：cd2b-fall-060x 是 false，不要轻易用**；无 NOTES |
| 10 | 0604 CD2B bedside | fixture 对不上 cd2b-fall-0604（那是 false）；无据 |
| 11 | `bedroom201-bedside-1027` | 无 NOTES，label 未经分析 |
| 12 | `case_lostfall_cd2b_11351148` | **README 自述「需用户判定到底发生了什么」= ground-truth 未定**，歧义出局 |

## 三、按积木类（taxonomy）—— 核心库 + 待导出

- **真摔（recall oracle）**：#1(漏报·最硬·NOTES)、#2(fw判出·NOTES)、#9(待导出·recovered+真hand-off) ＝ 3
- **假fall（precision oracle）**：#13(ghost·NOTES)、#14(多径·NOTES)、#5(lost·用户确认) ＝ 3
- **★跨房真 hand-off（Neighbor 黄金）**：#9（201 bathroom→bedroom）= **真因果非拼接**，待导出 333B+CD2B 窗口
- **★replay 关键三要素（用户）= 时间 + unit + neighbor device** → 上表每行已带「同 unit 其余 device」，使可喂**全 unit 实时流+事件**演示，非只单设备

## 三.5、benign（无报警正常）多设备积木 —— event_log 挖掘配方（用户给）

委员会实质2 问「无报警片段哪些是正向确认正常」→ 用户给可执行配方（**从 event_log 挖，不靠『碰巧没报』**）：

- **判据**：一段窗口里**同一 Room 的所有 device 都有 event** 且**全程无 fall** → 这段是干净的多设备同步正常生活片。
  - 例：15min 内 **101 bedroom 的 d523 + 09e7 + sleepad 0978 都有 event** → 优质 benign 积木（各 event 间含真实时间顺序）。
- **锚点**：以 **sleepad 上床/下床（InBed/LeftBed）为锚** → 取锚点 **±15min** 的同 Room radar 数据。
  - 理由：上/下床是强生活事件，前后必有真实的 enter/walk/sit 序列；同 room 多 radar 在这窗内的 track/event 是天然对齐的真实流。
- **ground-truth**：这类窗口的「正常」由**结构正向确认**（有完整生活事件链 + 全程无 fall），非「沉默」→ 满足委员会实质2（可作 benign 背景；是否可当 fall negative control 仍以「该窗确无摔」为准）。
- **产物**：每个 benign 积木 = 一个 `causality: real-contiguous` 的多设备同 room 窗口（manifest 记 unit/room/devices/锚点 event/window ts）。

**✅ 已跑（2026-06-09，owl_v2）—— 410 个合格 benign 窗**：地址 /88 高字节编房（101 bedroom=`…111:3:01xx`=0978+09E7+D523；201 bedroom=`…112:3:01xx`=1641+CD2B；bathroom=`02xx/03xx`）。挖掘 SQL（可复现）：

```sql
-- anchor=sleepad InBed/LeftBed；同 /88 room 全 device 有 event(dev>=2) 且 ±15min 无 Fall
WITH anchors AS (
  SELECT e.ts a_ts, network(set_masklen(e.device_addr,88)) room88
  FROM event_log e JOIN devices d ON e.device_addr=d.device_addr
  WHERE d.device_uid IN ('BM87224601641','BM87224700978') AND e.event_kind IN ('InBed','LeftBed'))
SELECT a.* FROM anchors a
WHERE (SELECT count(DISTINCT e2.device_addr) FROM event_log e2
        WHERE network(set_masklen(e2.device_addr,88))=a.room88
          AND e2.ts BETWEEN a.a_ts-interval '15 min' AND a.a_ts+interval '15 min')>=2
  AND (SELECT count(*) FROM event_log e3 WHERE network(set_masklen(e3.device_addr,88))=a.room88
          AND e3.event_kind='Fall' AND e3.ts BETWEEN a.a_ts-interval '15 min' AND a.a_ts+interval '15 min')=0;
```

- **结果：410 窗**（101 bedroom 多为 **dev=3** 三设备同步=最富；201 bedroom dev=2）。
- **按需挖不复制**（委员会 index-not-copy）：manifest 只记每个 benign 窗的 `unit/room/devices/anchor_ts/window` 坐标，loader 用 `export_case_v2.sh` 现挖；410 是池子，不全导。
- **ground-truth = 正向确认**：完整 InBed→生活事件→LeftBed 链 + 全程无 Fall = 结构正常，非「沉默」（满足委员会实质2；当 fall negative control 时该窗确无摔已由 fall_cnt=0 保证）。
- **无报警正常背景**（待抽，★委员会实质2 修正 ground-truth）：用户点的「101 bedroom enter/left bed·room、hunzi sleepad 上床、HR/RR」可抽，**但「无报警」不等于「确认正常」**——本系统存在的理由就是生产会**漏摔**（#1 即真摔+firmware 漏），一段无报警可能藏未报真摔。故：
  - **enter/left room·bed 事件 + HR/RR**：本就不产报警 → 「无报警」对 label **零信息**；其可信来自**受控/标注**，可作 benign 事件积木（room-ledger/bed-state/vital 纹理）。
  - **「这段没摔」当 fall negative control（断言 DBN 不该 fire）**：**必须正向确认**（住户自述/视频/受控实验），不是「碰巧没报」。无法正向确认的段 → manifest 标 `groundtruth: unverified-benign`，**只当背景纹理不当 negative control**（否则藏真摔→惩罚 DBN 正确 fire = 反 R5）。

## 四、待用户/委员会
1. ⚠ 厘清：#1 vs #6 是否同案（MDT/UTC）；#4/#10 fixture 时间对不上；0977/0978 device 性质。
2. ⚠ 补全各 unit 完整 device roster（101 有无 bathroom radar；201 sleepad；hunzi/Ton 其余）。
3. **#3(Ton)、#9(201 0609) 无 fixture → 待导出**（#9 尤其值——真 hand-off）。
4. **无报警正常片段**：指定提取源窗口（101/hunzi 的哪些 fixture 的哪段）。
5. **委员会已裁（feedback_p3.md）**：① 方向认可；② **乐高格式 = 生产 StreamMessage 原字段名**（`device_addr/device_type/topic_type/category/timestamp/dataValue`），不另造 bRecord；③ **manifest 落 `doc/cases/legos/`（索引指向真源不复制，#1.3）+ loader/切片器落 `testkit/`**；④ manifest schema = `id/class/labels/source_fixture/window/device_type/duration/groundtruth(verified-fall|verified-FP|verified-benign|unverified-benign)/causality(real-contiguous|synthetic-composite)`；⑤ **建序：先 manifest + 一个真碎片 Tier-1 recall 闭环测（喂真摔→断言 P(Fallen) fire 没被压 Vacant）走通，再批量铺**。
6. **本目录将转写为 `doc/cases/legos/` 的 manifest**（受理已具备，按上 schema）；当前文件作人读种子台账。
