---
name: NightAbsence Implementation
description: 整夜未归检测，cardagg 定时查 iot_timeseries + sleepace_report，已部署待明早验证
type: project
originSessionId: 765631c1-fe23-4b51-87d0-5feb18148e27
---
NightAbsence（整夜未归）检测：每天 OutBedTime 后检查昨晚 InBedTime~OutBedTime 有无 InBed/LeftBed 事件，无则发 alarm。

**已完成（未编译）：**
- `wisefido-cardagg/internal/service/alarm_service.go` — CheckNightAbsence 方法
- `wisefido-cardagg/main.go` — runNightAbsenceCheck goroutine + HGet adapter

**待修：**
1. SQL: `card_devices` 表不存在，需用 `cards.devices` JSONB 提取 device_id
2. 时区 fallback 改为 `America/Los_Angeles`（不是 Denver）
3. ticker 间隔改 10 分钟
4. 编译测试

**Why:** Sleepace 和 Radar 均有 Night Out of Bed 9PM-7AM 需求，检测老人是否整夜未归

**How to apply:** 修完 SQL + 时区后编译部署，用 Sleepace 设备 BM87224601641 测试
