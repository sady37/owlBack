# P4 反馈日志 — dwell 尾表 + K 参数 + FP 回归调参 — 项目组A+B ↔ 委员会

> 本文件从 `feedback_p3.md` 拆出(P4 讨论已达 ~2000 行,文件过大)。倒序,最新在上。
> P3 文件仍维护 P1/P2/P3/cutover 历史;P4(dwell tail + K)在此讨论。
>
> **协作协议**:项目组提案 → 委员会裁 → 裁后建；裁前不建 / 需新岔口列选项不擅决 / 直接 main / cd wisefido-sensor 跑 go / 放行 bar = build/vet/belief 绿 + 0 FAIL。

---

## 当前状态(2026-06-11)

**已建(项目组A,待委员会 R6 签字)**:
- `ea459c9` feat(belief): `dwellTailFor`→roomType×areaType 尾表(60s/12min/20min/90min/排除)
- `e467be8` fix(adapter_sleepace): 删 SubjectEntity gate

**委员会 R6 亲跑结果**:build/vet ✅ belief ✅ roomengine **3 FAIL FP 回归**——tail 表 default→60s 让已知 FP 被 DwellStill 推 fire:

| 测试 | 旧(tail 前) | 新(tail 后) |
|---|---|---|
| case-5(hunzi CABB lost FP) | P=0.022 | **P=0.993** |
| cabb-fall-A(静止站立 FP) | P=0.024 | **P=0.989** |
| 高 tol 开阔地久站 | 不 confirm | P=0.896 |

**根因**:旧 `dwellTailFor` default→false 让 GeomUnknown 的 DwellStill 数学上零作用→FP 靠没有 dwell 贡献被抑制。新 tail 表「其余→60s」让这些 zone 的 DwellStill 全面再生效→FP 被推 fire。

**待讨论**(本文件):
1. tail 表 default 60s 太激进?改为更大的 scale(120s/180s)?用 cell tolerance learning?
2. K 参数(45-120s 按 unit)怎么与 tail 表交互?
3. case-5/cabb-fall-A 的 FP 是该被 suppress 还是该 fire?是数据问题还是模型问题?
4. 高 tol 开阔地 tolerance gate 是否应单独修复?

---

## 审查记录（倒序）

(项目组/委员会讨论日志从下轮开始追加至此)
