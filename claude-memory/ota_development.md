---
name: OTA Development — Restored State
description: OTA升级+tenant迁移+UI改造已恢复到00ae2f0之上，owlBack 3dee949 / owlFront fffa354
type: project
originSessionId: 28e1aa2c-aa66-4a17-96ae-51bffb6a19d5
---
## 状态：P0/P1/P2 全部完成（2026-05-06 实查确认）

- owlBack: `3dee949` 起源 + 后续多次提交，main 持续推进
- owlFront: `fffa354` 起源 + 后续多次提交，main 持续推进
- 基于 00ae2f0 (card_id 重构) + 10f8e16 (PHI encryption)

**P2 调度器实现**（确认存在于 [wisefido-qinglan/cmd/wisefido-qinglan/main.go:223-288](../../../owl/owlBack/wisefido-qinglan/cmd/wisefido-qinglan/main.go#L223-L288)）：
- `otaScheduler = ota.NewScheduler(db, mqttPushFn, tcpPushFn, 1*time.Minute, fwDir, fwURL)`
- `otaScheduler.Start()` 守护线程；`Stop()` 优雅退出
- MQTTPublisher.PublishOTA 已 wire 为 schedule callback
- TCP path 走 httpsServer.OTAManager().PushToDevice（老 MCU）

## 架构概要

### OTA 两阶段设计
1. **TCP OTA (P0)**: 老MCU通过TCP连接→cmux分流→protobuf OTAPush(type=16)→设备下载固件→自动重启
2. **MQTT OTA (P1)**: 新MCU通过MQTT→cmd:"ota"→设备下载固件→自动重启

### 端口复用 (cmux)
- 8443 端口: TLS→HTTPS Auth, 原始TCP→设备OTA连接
- `/firmware/` 静态文件服务（固件下载）

### 租户迁移规则
- Unallocated 已删除，默认租户为 System
- 业务租户间不能直接迁移，必须经 System 或 Trash
- Delete = Move to Trash（非物理删除）
- 迁移时清除 bound_room_id/bound_bed_id

### ESP32 端口行为
- 老MCU: OTA后 web server port 回 443，由外部BLE重配，owlBack不处理
- 新MCU: 自动读取 webserver:port，无需干预

## 关键约束（踩坑记录）

1. **protobuf 必须用 protoc 生成的文件**，不能手写 — 否则 proto.Unmarshal 不兼容
2. **OTAReq 字段名**: Espsfver / ESPFileUrl / ESPFileSize / ESPFileSHA256 / Radarsfver / RadarFileUrl / RadarFileSize / RadarFileSHA256（大小写混合，跟 proto 定义一致）
3. **Frame 字段名**: Data（不是 Payload）
4. **OTAProgressCallback 签名**: `func(uid string, progress int, message string)` — progress 编码状态: -1=失败, 0=接受, 10=雷达下载完, 25=雷达升级完, 56=ESP下载完, 100=全部完成
5. **tcp.NewServer 第二参数**: uint32（不是 int32）
6. **PushToDevice 签名**: `func(req PushRequest) PushResult`（不是 3 参数版本）
7. **scheduler 发 MQTT OTA push**（不是 reboot/restart）
8. **disconnect 不报告 OTA 失败**（defer 中不调 OnProgress(-1)）
9. **Reboot 命令格式**: `dev:0`（不是 `restart:true`）
