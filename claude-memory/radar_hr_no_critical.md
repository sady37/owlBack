---
name: radar-hr-no-critical
description: Radar HR 阈值告警不能设 CRITICAL（mmWave 精度 ±10 BPM 不支持临床判断）；FE select 需主动收窄；接触式 sleepad 不受此限
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 0ef86edd-98ed-4293-bb19-0c02b8af18a4
---

雷达 HR 告警等级**只能 DISABLED/WARNING**，不能 CRITICAL。Sleepad（接触式压电）仍可 CRITICAL。

**Why:** mmWave HR 是从胸廓位移相位差解出来的，BPF 通带 0.8-2 Hz + 算法精度上限：
- TI mmWave Vital Signs Lab demo：HR ±10 BPM
- Calterah 60GHz（含现役 CAL60S244-AB）官方：理论上限 HR ±5 / RR ±2 BPM（先进 ESPRIT/IAA 算法、lab condition），实际部署常规算法 HR ±10 量级
- 实际场景受距离/角度/穿透/动作影响误差更大

用这种精度做 CRITICAL（=临床决策级）会产生大量误报，护士台直接信任 → 出事。Sleepad 接触式压电传感精度高得多，CRITICAL 是合理的。

**How to apply:**
- alarm.go 改 Radar HR 默认 AlarmLevel：用 `AlarmLevelWarn` 不用 `AlarmLevelCrit`（[Sleepad 那段](../../../owl/owlBack/owl-common/alarm/alarm.go) 不动）
- FE radar 的 alarm-level select 不能用通用 `[DISABLED, WARNING, CRITICAL]` 三选，要给 HR 加单独 case 只返回 `[DISABLED, WARNING]`（[radar-monitor-settings.vue getRadarSelectOptions](../../../owl/owlFront/src/views/settings/radar-monitor-settings.vue)）
- load 路径 normalize：老 DB 里 HR=CRITICAL 的数据 load 时降级成 WARNING（save 一次就修），否则 select 找不到匹配选项会崩
- Registry 表（alarm.go 全局 fallback）的 HR `DefaultLevel: AlarmLevelCrit` **不要动** — 那是 per-AlarmType 全局，sleepad 也共用

阈值范围（TI/Calterah mmWave 实测有效）：HR 60-110 BPM、RR 6-30 rpm。设到此范围外的阈值 BPF 滤掉根本触不到，FE 用 Effective 做 input :min/:max。详见 [[sleepace-vendor-param-limits]]（同样的"协议允许 vs 实测能用"两层范围思路）。
