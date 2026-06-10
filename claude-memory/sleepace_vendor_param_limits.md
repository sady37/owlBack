---
name: sleepace-vendor-param-limits
description: sleepace updatealarmnotifyconfig 厂家硬限范围 + 超限语义（HR 40-130 / BR 9-30），来源 = 反编译 AlarmConfigController.class
metadata: 
  node_type: memory
  type: reference
  originSessionId: 0ef86edd-98ed-4293-bb19-0c02b8af18a4
---

`/sleepace/updatealarmnotifyconfig` 厂家 hard validation（jar 反编译来源；不在文档里）：

| 字段 | 范围 | 备注 |
|---|---|---|
| `maxHeartRate` | **40 – 130** | bpm |
| `minHeartRate` | **40 – 130** | bpm |
| `maxBreathRate` | **9 – 30** | rpm |
| `minBreathRate` | **9 – 30** | rpm |

**超范围语义 = 拒绝整个请求**（`checkParameter` 返回 false 后 controller 直接 return error response），不是 silent clamp。一项超限 → 所有阈值都不会更新 → UI 看着 save 成功但厂家其实没动。

时长字段（`heartRateFastDuration` / `breathPauseDuration` 等）**无** hard validation。

来源（jar 反编译）：
```
/home/wisefido/owl/sleepace/sleepace-service/lib/sleepace-tb-3.x-impl-0.0.1-SNAPSHOT.jar
  → com/medica/xxbio/controller/AlarmConfigController.class
  → updateAlarmNotifyConfig() → checkParameter(resp, name, value, min, max)
javap -c -p AlarmConfigController.class | grep -B2 -A6 'String maxHeartRate'
```

写新 sleepace 集成或调阈值前查这里。FE clamp 实现：[[respiratory-heart-rate-fe-clamp]]（见 [sleepace-monitor-settings.vue](../../../owl/owlFront/src/views/settings/sleepace-monitor-settings.vue) `clampRespRateRpm` / `clampHrBpm`）。

相关：[[v2-cutover-type-assert-silent-drop]]（v2 sleepace 下发被 type assertion 默默吞掉的同源 bug）
