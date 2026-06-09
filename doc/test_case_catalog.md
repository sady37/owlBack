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
| **hunzi** | bathroom | Radar **CABB** | #5；其余 device ⚠ |
| ~~Ton~~ | — | ~~FFDB~~ | **#3 drop（用户）** |

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

### ★ 待导出（黄金，无 fixture）

| # | 时间 | label | unit·房·device | 价值 |
|---|---|---|---|---|
| 9 | **0609 07:18** | **REAL·recovered** | 201 bathroom **333B** → 走进 201 bedroom **CD2B**（neighbor: sleepad 1641） | **真跨房 hand-off**（非拼接、真因果）：tub 旁 fall→firmware fire→起身 auto-resolved→进 bedroom。**Neighbor recall 黄金 + Tier-2 级真验**（含 333B+CD2B 两 device 窗口即可） |

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
