# Sensor 变更审查日志 (feedback)

**用途**:每 10min `git pull`,审查另一 AI agent 对 `wisefido-sensor/` 的进度与变更,对照下列原则 audit,结果倒序追加到本文件。

**审查基准**(来源:`doc/belief_dbn_signal_map.md` + `doc/belief_dbn_proposal.html` + `CLAUDE.md`):

*信号可信性原则*
1. **pose/z 对 fall 只正向不负向** —— lying 抬 fall;stand/sit、z 任何值**永不压制 fall**。任何用 pose/z 抑制/否决跌倒的代码 = 违规(制造 FN)。
2. **z 三档 posture**:z>80→stand / 30–80→sit / <30→噪声(不进 fall)。z<30 不能当 fall 正/负向。
3. **enter/exit event**:present=可信正向;**absence≠负向**(信号丢失不发 event ≠ 人还在)。不得用"无 exit event"推"还在房间"。
4. **realness(R_i)用 XY 运动学**:室内-老人速度天花板(~110–130,非 200)、空间跳跃近确定性 ghost、冻结零方差+pose/z 锁定=ghost(≠ box 内真静止)。
5. **fall 压制只走 reliable**:realness / cell rest-zone / sleepad / recapture / human-bed;不走 pose/z。
6. **recapture 软恢复非硬 cancel**:人返回可能是跌后自救,不得抹掉真摔。

*架构原则*
7. **cell engine 独立**:DBN 只读其 AreaType(只读先验);cell 学习(dwell/z 档/护栏)不混进 DBN。
8. **still-box 单源**:track 层算一次 → cell engine + DBN 双消费,谁都不重算(防 drift)。
9. **单源真相**(CLAUDE.md #1.3):多源风险字段唯一写入口;派生数据唯一权威源,下游只读。
10. **常量化**(#1.1):事件/类别字面量唯一来源 = observation 包常量,禁字面量复写。
11. **不向后兼容**(#1.2):删即删,无 stub/no-op/deprecated alias/TODO 迁移。
12. **错误处理只在边界**(#1.4):内部调用 trust caller,禁防御性 nil 兜底吞 bug。

*流分配原则*(#2)
13. 持久化事件→`iot:*:stream`;瞬态状态→专用流不入库;producer 维护的流 consumer 只读不重做。

**审查方法**:`git fetch && diff <last-audited>..origin/main -- wisefido-sensor/`,逐变更对照上列;命中违规标 ⚠️,良好标 ✅,存疑标 ❓。

---

## 进度基准

- **审查起点 commit**:`3da5dfe`(本日志创建时 HEAD)
- 此前 sensor 主线:`5aacad1`(still-box 50×50)、`4245f14`(lost-fall 读 room_type)、`d867c62`(risk 窗 22:00-06:30)、`96c69bd`(bed bayesian decay+standby)。
- **last-audited**:`f0bb43f`(下次从此 commit 起算 delta)

---

## 审查记录(倒序,最新在上)

<!-- 每次 audit 追加一条:
### [YYYY-MM-DD HH:MM TZ] <last>..<new>
**变更摘要**:...
**对照审查**:✅/⚠️/❓ 逐条
**建议**:...
-->

### [2026-06-06 21:06 MDT] 3da5dfe..f0bb43f — 施工计划 belief_dbn_impl_plan.md(10 commit,纯 doc,无 sensor 代码)

**变更摘要**:新增 `doc/belief_dbn_impl_plan.md`(520 行),把 signal_map/proposal 拆成 P2–P9 + P9' 可施工/验收/灰度的 P-task,含铁律 R0–R7、DoD 模板、依赖 DAG、灰度门。**无生产代码改动**(明示代码待签字后另起 commit)。

**总评**:✅ **高质量,忠实对照 signal_map**。R0 shadow-first / R1 不碰 alarm 决策 / R2-R3 cell engine 独立只读 / R5 pose-z 正向-only / R7 常量化 全部入铁律;逐行审 likelihood.go 标出待改点;P9 设为 go/no-go 数学闸并诚实写明 §11.2 残差对可能 no-go-but-shadow。可作为施工基线。

**⚠️ 实质问题(建议施工方消化后再开 P-task)**:

1. **⚠️ ObsNoDetect→Fallen×1.6 必须由 realness+door-distance 门控**(P6.1 提到"不重复计"但未点破根因)。"消失→抬 fall"是**用 absence 当正向**,正是历史 dropout-FP 的来源。必须:仅当 R_i=real ∧ 非门区消失(door-distance 未↓)才允许 ObsNoDetect 抬 Fallen;ghost 消失 / 门区消失 → 不抬。否则 P3 还没判出 ghost,no-detect 已先把 Fallen 抬起来。**建议 P6.1 显式写此门控,与 P3/P2.5 联锁。**

2. **⚠️ P3.2 冻结判据不能只靠"零方差"** —— 真机反例:bedroom201-cd2b 那个冻结椅子 ghost **位置在 (-170,330)/(-150,340)/(-120,300) 间晃 ~50cm,不是零方差**。静止反射体会带多径抖动。判据应是**复合签名**:`不可能跳变出生 + pose/z 锁死(pose=4∧z=0 恒定 N tick)+ 位置受限于小区 + 距门远`,**不是单纯方差阈**。只用方差既会漏(ghost 有抖)也会误(真人微抖)。**建议 P3.2 改成复合签名,方差只是其一。**

3. **⚠️ S_vol 尾形标定样本不足**(P4.1)。7-case fixture 标不出生存函数尾形,尤其 bedside benign-外出尾(cd2b 两例都 ~5.5–6min 返回,n=2)。8min 站立尾、bedside 尾都需更多样本,否则参数只能粗档。**建议 P4.1 标注"尾形粗档 + 待样本收紧",P9 报告里把样本量列为 margin 置信的限制项。**

4. **❓ O_b→S_Fallen(1−0.7p) 抑制勿掩 bedside-fall**(P5.4)。床占用压 Fallen 合规(sleepad 可信),但要确认"leftBed→床边静止"窗内 O_b 已低(否则陈旧 O_b 压掉床边晕倒)。**建议 P5.4 加边界:O_b 抑制 fall 仅在 O_b fresh∧高时;leftBed 后 O_b 必须及时落。**

5. **❓ firmware-fall 占位 ×3–4 偏高**(P2.4)。pose=5 是 pose 派生(R5 域),即便有 firmware 限定,×3–4 仍是强正向。**建议**:shadow 期先取更保守档(≤×2)直到 P9 用 firmware_fall TP/FP 真机率标定;现网 Device_ALARM 直发不动(计划已守 R1 ✓)。

**小记**:P9' 命名(前置却带撇号)建议改 P1;P6.4 R_i–S_i 先因子化后按 oracle 升联合 ✓(与委员会一致)。

**裁决**:**计划通过,可按 DAG 开工**;P-task 落地时把上述 1/2 作**阻塞项**(它们直接关系 cd2b 漏报与 dropout-FP 两类核心错误),3/4/5 作**注意项**。每个 P-task 仍 shadow-first,P9 oracle 出 go/no-go 前不进 canary。

---
