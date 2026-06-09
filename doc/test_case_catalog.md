# 基础测试模块 —— 真碎片案例目录（ground-truth 确认）

> 用途：① Tier-1 真碎片复合测的积木来源；② 未来 AI 模拟生成器的素材库。
> 硬约束（用户）：只收 **① 验证过的 test case**（真摔/假fall，人工确认 ground-truth）+ **② 无报警的真实正常数据**（enter/left room·bed、HR/RR；「没报警」=可信「正常」）。
> 本目录 = 用户 2026-06-09 确认信息录入；`⚠` 标施工方待确认项。**裁前不建**（模块/harness 待委员会受理）。

## 一、Unit × Device 编组（从案例反推，⚠ 完整 roster 待用户确认）

| Unit | 房间 | Device（radar/sleepad）| 来源案例 |
|---|---|---|---|
| **101** | bedroom | Radar **09e7**(9e7/e97)、Radar **D523**、sleepad **0978**(978)/**0977**⚠ | #1/#2/#4/#6/#13 |
| **101** | bathroom | ⚠ 待确认有无 radar | — |
| **201** | bedroom | Radar **CD2B**、sleepad **1641**⚠ | #7/#8/#10/#11/#12 |
| **201** | bathroom | Radar **333B** | #9 |
| **hunzi** | bathroom | Radar **CABB** | #5 |
| **hunzi** | bedroom | ⚠ 待确认 | — |
| **Ton** | （房间⚠）| Radar **FFDB** | #3 |

⚠ `0977` vs `0978`：#1 NOTES 写 sleepad 978；#2 用户写 Radar 0977——待厘清 0977 是 radar 还是 sleepad、与 0978 关系。

## 二、案例目录（Trigger/alarm | real/false | unit | 摘要 | unit 全 device | fixture）

| # | 日期·时间(trigger) | real/false | unit·房·device | 分析摘要 | fixture 路径 |
|---|---|---|---|---|---|
| 1 | 0605 11:37–11:46 | **REAL·漏报** | 101 bedroom · 09e7+sleepad978 | 床边跌倒；雷达误识别旁边 chair；firmware 全程 pose≈3 **未报**；段4c 自救取消零记录 | `bedtest-0605-1-bedside-fall-no-fw-detect` |
| 2 | 0606 19:08–19:18 | **REAL** | 101 bedroom · 0977⚠+sleepad | 床边跌倒；身子仍靠床边→sleepad 仍检 HR/RR；firmware **判出**(pose=5) | `bedtest-0605-2-bedside-fall-fw-detect` |
| 3 | 0608 16:37 | **FALSE** | Ton · FFDB | 坐沙发（sofa 比椅矮→重心降）→ 频报跌倒 | ⚠ 无 fixture（待导出） |
| 4 | 0605 09:59 | **FALSE** | 101 bedroom · D523 | 雷达 **edge fall**，丢轨距 500~550cm（**距离闸 G-3 案**） | ⚠ `d523-bedroom-lost-0933`? 待核(09:59 vs 09:33) |
| 5 | 0605 06:08:44–06:16 | **FALSE** | hunzi bathroom · CABB | **lost track** 误报 | `hunzi-cabb-lost-*`（0529/0530/0601 三窗） |
| 6 | 0605 17:38 | **REAL** | 101 bedroom · e97(09e7) | bed fall test | ⚠ 疑 = #1 的 UTC 时刻（11:37MDT=17:37UTC，同 09e7）→ 待厘清是否同案 |
| 7 | 0607 01:27:29 | **FALSE** | 201 bedroom · CD2B | bed 上报 fall（在床误报） | `cd2b-fall-0607-0127` |
| 8 | 0606 09:17 | **FALSE** | 201 bedroom · CD2B | Bathroom_enter fall；后又有 track 进入并经过该处 | `cd2b-fall-0606-0917` |
| 9 | **0609 07:18** | **REAL·recovered** | 201 bathroom · 333B | tub 旁 fall，**firmware fire**→起身 auto-resolved→**走进 201 bedroom**（★真 hand-off：bathroom→bedroom，Neighbor recall 黄金案） | ⚠ 无 fixture（**今日新，待导出**） |
| 10 | 0604 16:14–16:31 | **REAL** | 201 bedroom · CD2B | bedside-fall | ⚠ `cd2b-fall-0604-2233`? 待核(16:14 vs 22:33) |
| 11 | 0606 10:27–10:37 | **REAL** | 201 bedroom · CD2B | bedside-fall | `bedroom201-bedside-1027` |
| 12 | 0504 11:35–11:48 | **REAL** | 201 bedroom · CD2B | bedside-fall（cabb+last30min） | `case_lostfall_cd2b_11351148` |
| 13 | 0526 21:00–21:24 | **FALSE·ghost** | 101 bedroom · D523 | 镜面 ghost 误报 | `d523-mirror-ghost-0526` |

## 三、按积木类（taxonomy）归并 —— AI 生成器/复合测取用

- **真摔（recall oracle）**：#1(漏报·最硬)、#2(fw判出)、#9(recovered+hand-off)、#10、#11、#12 ＝ 6 例
- **假fall（precision oracle）**：#3(沙发矮坐)、#4(edge/距离闸)、#5(lost)、#7(在床)、#8(bathroom_enter+穿越)、#13(ghost) ＝ 6 例
- **★跨房真 hand-off（Neighbor recall 黄金）**：#9（201 bathroom fall→走进 201 bedroom）= 真实「人从一房到另一房」，**不是拼接、是真因果**——若导出含 bathroom(333B)+bedroom(CD2B) 两 device 窗口，即 Tier-2 级真验素材
- **无报警正常背景**（待抽）：⚠ 用户点的「101 bedroom enter/left bed·room、hunzi sleepad 上床、HR/RR」需指定**从哪些无报警窗口抽**（这些案例的 fixture 窗口里凡未触警的 enter/left/InBed/HR-RR 片段都可切作 benign 积木）

## 四、待用户/委员会
1. ⚠ 厘清：#1 vs #6 是否同案（MDT/UTC）；#4/#10 fixture 时间对不上；0977/0978 device 性质。
2. ⚠ 补全各 unit 完整 device roster（101 有无 bathroom radar；201 sleepad；hunzi/Ton 其余）。
3. **#3(Ton)、#9(201 0609) 无 fixture → 待导出**（#9 尤其值——真 hand-off）。
4. **无报警正常片段**：指定提取源窗口（101/hunzi 的哪些 fixture 的哪段）。
5. 委员会受理后：写 manifest schema + test_record.txt→bRecord 转换器 + 切片工具（裁前不建）。
