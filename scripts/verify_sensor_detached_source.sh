#!/usr/bin/env bash
# 源头判定：看 MQTT 报文的 deviceId(device_code) + dataKey + timeStamp 即可
# dataKey=deviceSenSor → 告警来自 device 或 sleepace-service；cardagg 只落库不产生
# 用法: ./verify_sensor_detached_source.sh
echo "MQTT 看 deviceId + dataKey + timeStamp。dataKey=deviceSenSor 即设备/sleepace-service 上报；wisefido-cardagg 只落库。"
