# Ghost 排险:z-物理矛盾 realness 先验(O(1)/帧)

便宜地抓"无高度反射假人"这一类 ghost——一条轨**自报直立却长期贴地**(多径反射把身高塌成 0)。
专治 mirror_detect(O(T²))/static_reflector(O(T×30)+300s 门)抓不到、或够不到的廉价场景。
**不替代**它们;它们留给"z 也被伪造成真高度"的同步镜像。

实锤样本:case-b197-0626-15371540 tid1(反射 ghost)vs tid0(真人),直立 pose 下:

| | 直立帧 | z<30cm | z 中位 |
|---|---|---|---|
| tid0 真人 | 143 | **40%** | 63 |
| tid1 ghost | 91 | **96%** | 0 |

## 1. 信号:直立 pose ∧ 持续贴地

直立 pose(Walking/Sitting/Standing)**必有身高**(躯干 z>~30cm)。"站着却 z≈0" = 物理矛盾 = 反射证据。

⚠️ **单帧 z<30 不干净**:真人也有 40% 低 z 帧(雷达 z 噪声/丢测)。**干净的是持续比例**——
ghost 是**几乎总**贴地(96%),真人只是**偶尔**(40%)。所以判据是"长期占比"非"这一帧"。

## 2. 机制:O(1)/帧累加器

每轨一个 EMA(无 history 扫、无配对、无几何):

```
# 每帧摄入(已有 pose/z 在手处)，仅直立 pose 参与:
if pose in {Walking, Sitting, Standing}:
    sample = 1.0 if z < zFloorCm else 0.0        # zFloorCm ≈ 30
    ts.UprightFloorEMA = α*sample + (1-α)*ts.UprightFloorEMA   # α≈0.1
    ts.UprightFrames++                            # 门控最小样本
```

- **ghost 判据**:`UprightFrames ≥ minFrames(≈15)` ∧ `UprightFloorEMA ≥ τ(≈0.8)`。
- 字段挂 TrackState(同 Prev2Pose 一档),更新点在现有 per-track 摄入循环,**零额外遍历**。

## 3. 自校准门(防冤枉)

z 可靠性因雷达/固件而异——有的对**所有人**报 z=0。规则只在"本台雷达已证明会报 z"时启用:

- 维护 per-device 标志 `RadarReportsZ`:近窗内**任一**直立轨出现过 z≥zFloorCm(像 tid0 的 86 帧)→ true。
- `RadarReportsZ==false`(全房 z=0,无对比)→ 信号**自动失效**,不产出 ghost 证据。
- 即"有真人 z 作对照,反射 z≈0 才有意义";没对照不猜。

## 4. 输出 = census realness 软先验,**不是 fall veto**

- 喂给已有 **census realness**(ghost 单一权威),与几何/连续性证据**软合并**,**不单独硬否决**。
- **FN-safety / 铁律 [[realness_never_vetoes_fall]]**:**只统计直立 pose**。`z<30 ∧ {Fallen/Lying/SitGround}`
  = 真摔正常形态 → **永不计、永不碰**。真摔者(贴地+摔 pose)对这条规则完全透明。
- 大白话:"站着的人脚底板不会贴地飘——谁老这样八成是镜子/金属反的假人;但真摔倒贴地是正常的,这规矩碰都不碰他。"

## 5. 计算量对比

| 机制 | 成本 | 抓什么 |
|---|---|---|
| mirror_detect | O(T²)+5帧窗+3不变量 | 同步运动镜像 |
| static_reflector | O(T×30)+300s 门 | 久钉金属(本 ghost 才 89s,够不到) |
| census 几何 | 每对 O(墙) | 连线穿墙 |
| **z-持续比例(本稿)** | **O(1)/帧,一计数器** | 无高度反射假人 |

## 6. Oracle(待真数据标定,铁律 [[fall_data_is_artificial_test]])

`zFloorCm(30)` / `τ(0.8)` / `minFrames(15)` / `α(0.1)` / 自校准窗 —— 经验初值,真 case replay 调,
**不靠人为测试数据标定 fire**(本机制不进 fire,只进 realness,可证伪部分 = "ghost 该判 ghost 吗")。

## 7. 验证(若实施)

simulate-make 加场景:真人(直立 z 真)+ 反射假人(直立 z≈0 抖)并存 → 确认 ghost 轨 EMA 越阈被 census 判低
realness、真人不受影响;并造"摔倒贴地"轨确认**不被碰**(realness 不动、fall 照常)。回归用本 b197 case。
