---
name: OTA ESP32 Port Behavior
description: ESP32 OTA后端口重置行为：老MCU用BLE重配，新MCU自动读取webserver:port
type: project
originSessionId: 28e1aa2c-aa66-4a17-96ae-51bffb6a19d5
---
## ESP32 Web Server Port 在 OTA 升级后的行为

**老 MCU (TCP 阶段)**:
- OTA 刷新后 ESP32 自动重启，web server port 会自动回 443（NVS 默认值）
- **不需要 owlBack 处理**，不需要 nginx 代理 443
- 由外部 **BLE 工具重配** web server port（ip_port 属性）
- 这是已知的硬件行为，属于运维操作

**新 MCU (MQTT 阶段)**:
- 固件升级后 MCU 能自动读取刷新时的 webserver:port
- 无需人工 BLE 干预

**Why:** ESP32 NVS 存储在 OTA 分区擦写后会丢失自定义端口配置，老固件无法保持 port 设置
**How to apply:** TCP OTA 推送后不要尝试在代码中修正端口，等运维人员 BLE 重配即可
