---
name: qinglan-sleepace-access-gate-consolidation
description: S5 已完成 — qinglan/sleepace mqtt_consumer access gate 由 v1 双字段（allow_access+business_access）合并为单一 devices.access bool；monitoring_enabled 故意保留作为独立语义
metadata: 
  node_type: memory
  type: project
  originSessionId: 5a73f32f-5389-4b8e-a637-f38a229d262b
---

## 终态（2026-05-22 完成）

**两字段，非三字段**：
- `devices.access BOOL`（platform_admin 审批位）= v1 `allow_access + business_access` 合并
- `devices.monitoring_enabled BOOL`（tenant 业务监控开关）= 故意独立保留，语义不同

memory 早先描述 "3 字段合并为 1 字段" 不准确——实际只合并了审批侧双字段，monitoring_enabled 始终独立。

## Gate 表达式

- `canIoT = dev.Access`（向 iot:event/alarm/stat 流发送的前提）
- `canMonitor = dev.Access && dev.MonitoringEnabled`（向 iot:monitor 流发送的前提）

实现：
- [wisefido-qinglan/internal/consumer/mqtt_consumer.go](owlBack/wisefido-qinglan/internal/consumer/mqtt_consumer.go) `resolveIotPolicy`
- [wisefido-sleepace/internal/consumer/mqtt_consumer.go](owlBack/wisefido-sleepace/internal/consumer/mqtt_consumer.go) `dispatch`
- [owl-common/card/device_baseline.go](owlBack/owl-common/card/device_baseline.go) `DeviceBaseline.Access / MonitoringEnabled`

## 默认值规则

- **factory_meta 已登记设备 INSERT**：不指定 → DB DEFAULT TRUE（已批准，待分配）
- **MQTT runtime auto-detect 陌生 device_uid**：显式 INSERT access=FALSE + 落入 Trash tenant `fd00:0:2::/48`，platform_admin 手动审批后才能改 TRUE

## 残留作用域外（不在 S5 范围）

[[v2_alias_shim_removal]]（S10）涵盖 wisefido-data 服务 `AllowAccess` 字段 — 它是 v1 FE wire-format compat shim（device_store_payload / postgres_device_store 的 v1 列名映射层），改它要 FE 联调，不属本任务。

**Why**：access gate 是流出口的决策点，监控开关是业务策略，混在一起会让 platform 审批和租户选择互相耦合。
**How to apply**：碰到 access gate 相关代码，记住"审批 vs 业务"两层独立，不要试图再合并 monitoring_enabled。
