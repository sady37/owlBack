# owlOps — 运维自动化

与核心在线服务（cardagg / iot / sleepace / ai 等）解耦的运维任务集合。

## 背景（简）

2026-04 发现 NightAbsence 批量误报，追溯后发现两个独立问题：
1. **Sleepace 时区 / 报告时间下发路径**：已修——改用 per-device alarm_params（随 monitor settings save 自动下发），DST 由 IANA 库在调用时换算。不再需要独立调度器。
2. **NightAbsence 判定逻辑**：仍待修——还在查有 bug 的 sleepace_report 表（见 §2）。

架构结论：
- **厂家按设备本地时钟**触发 `reportUploadTime`（不是服务器 OS 时钟）
- 设备本地时钟 = bind 时传的 `timezone` 秒数；**厂家不做 DST**，所以我们每次 save 用 IANA 当前 offset 重算
- Sleepace 服务器的 `config.properties.timeZone=-25200` 是 channel 默认（仅影响 report summary 字段），与设备行为无关

---

## 1. Sleepace 时区 / 报告时间（已完成）

**数据模型**：
```
alarm_cloud.metadata.tenant_sleepreport_time   (int 1-24)  ← tenant 默认
alarm_cloud.metadata.TenantResetTime           (ResetTime/NapTime)
alarm_device.monitor_config.items[SleepadSetting].alarm_params:
    timezone           (IANA 字符串，"" 表示跟随 unit.timezone)
    sleep_report_time  (int 1-24)
```

**save 流程**（`device_monitor_settings_service.pushDeviceSettings` SleepadSetting 分支）：
1. `normalizeSleepadTimezone`：若 `alarm_params.timezone == unit.timezone` → 清空（让 unit.tz 变更自动跟随）
2. `BindDevice(tzSec, gender, age)` 下发 `/sleepace/bind`：
   - tzSec = `IANAToOffsetSeconds(timezone)`，IANA 库按 call-time 处理 DST
   - gender/age = 查 resident_phi 解密（无 resident → 1 / 65；age 向上取整 5 倍数，最小 60）
3. `SetReportUploadTime(hour)` 下发 `/sleepace/setReportUploadTime`

**DST 触发入口**（外部脚本调，service 内部自行查当前 effective 值下发）：
- `POST /internal/sleepace/device/{device_id}/resync-timezone`
- `POST /internal/sleepace/device/{device_id}/resync-report-time`

典型用法：DST 切换日 cron 遍历所有 Sleepad 设备各调一次这两个端点即可。不需要 Python 常驻 scheduler。

---

## 2. NightAbsence 逻辑改造（P1，待做）

依赖 sleepace_report 对齐，但当前 cardagg 查询层有两处 bug，必须先修。

### 2.1 sleepace_report 查询 bug（cardagg）

[alarm_service.go:1051-1083](owlBack/wisefido-cardagg/internal/service/alarm_service.go#L1051-L1083)：
- `sleepace_report.date` 列是 INT（`20260422` 形式），但 cardagg 查询用 `nightStart.Format("20060102")` 字符串 → 类型不匹配
- `date` 写入时是 **UTC**（`timeToDate()` 用 `.UTC()`），cardagg 查询用 **Denver 本地日期** → 时区错位
- 修：改 int + 改用 UTC 日期换算

### 2.2 改用实时流为主、报告为辅

- 主判据：`iot_timeseries` 在 `nightStart~nightEnd` 窗口内有 `bed_status=0` 的 monitor track（或 HR>0 / RR>0）
- 兜底判据：`alarm_events` 同窗口有 `InBed` 事件
- `sleepace_report` 仅作交叉验证（报告存在且 `sleep_state>=1` → 有人）
- 三者 OR = 有人；三者全无 → NightAbsence

### 2.3 Sleepad 灵敏度异常独立事件（P2）

如 202 Room 整夜只有 ~2 分钟在床信号——不是"整夜未归"而是传感器问题。加新事件 `DeviceInsensitivity`（或并入 `DeviceFailure`）：
- 夜间窗口 bed_status=0 累计 < N 分钟（默认 30）
- 但 monitor 在线
- → 灵敏度/摆放问题

### 2.4 ResetTime 两层 resolver（仍需做）

当前 `cardagg/alarm_service.go:512 getResetTimeParamsForTenant` 只读 `alarm_cloud.metadata`；`alarm_device.metadata.ResetTime` 没人读也没人写。要补：
- wisefido-data 补 device-level ResetTime CRUD + 初建从 tenant 拷贝
- owl-common 加 `EffectiveResetTime(tenantID, deviceID)` resolver（device > tenant）
- cardagg 切换为 consumer

---

## 3. 健康快照（P2，未实现）

每日一份 `/home/wisefido/owl/log/maintenance-YYYYMMDD.json`：
- Sleepad 灵敏度异常候选（像 202 Room）
- NightAbsence 报警日/周趋势
- sleepace_report 最新 createTime 漂移（> N 小时未上传的设备）
- Sleepace analysis MQ 事件延迟

---

## 4. 备份（未实现，占位）

**Target 待定：**
- 本机冷备 `/var/backups/owl/` 还是推 S3/NAS
- 保留策略（日备 14 天 / 周备 8 周）
- 加密层

**范围待定：**
- `owlrd` PG（核心）
- `sleepace_tb_data` MySQL（历史报告，厂家也存，可选）
- `sleepace_pro_user` MySQL（设备映射）
- Redis AOF（cardStatus 可重建，可选）
- `/home/wisefido/owl/log/*.log`
- 各服务 `config.yaml`

---

## 5. 日志轮转（未实现，占位）

- 归档 `/home/wisefido/owl/log/*.log.YYYYMMDD`，压缩
- 14 天解压 + 90 天压缩 + 之后删除
- 关键错误 tail 到 `maintenance-YYYYMM.log`

---

## 实施优先级

| P | 任务 | 状态 |
|---|------|------|
| ~~P0~~ | Sleepace 时区 / 报告时间下发 | ✅ 已完成（§1） |
| **P1** | 2.1 cardagg sleepace_report 查询 bug | 待做 |
| **P1** | 2.2 NightAbsence 改用实时流 | 待做（依赖 2.1） |
| P2 | 2.3 灵敏度异常事件 | 待做 |
| P2 | 2.4 ResetTime device-level resolver | 待做 |
| P2 | 3 健康快照 | 待做 |
| P3 | 4 备份 | 明确 target 后 |
| P3 | 5 日志轮转 | 空间告急前 |

---

## 6. DST 自动化（建议简化版）

不需要常驻 Python scheduler，只需 **2 个 DST 切换日**的 cron：

```cron
# 每年 3 月第二个周日 + 11 月第一个周日 03:30 local 触发
# （或直接每天凌晨 03:30 跑一次——幂等，每日重复无副作用）
30 3 * * *  /usr/local/bin/owl-dst-resync.sh
```

脚本伪代码：
```bash
#!/bin/bash
# 遍历所有 Sleepad 设备，各调一次 resync endpoint。
# /internal/* 跳过 auth，无需 token。
for device_id in $(psql -At -c "SELECT d.device_id FROM devices d JOIN device_store ds ON ds.device_id=d.device_id WHERE ds.device_type='Sleepad' AND ds.device_code<>''"); do
  curl -X POST "http://127.0.0.1:8080/internal/sleepace/device/$device_id/resync-timezone"
done
```

**新装设备不需要这个脚本**——device_store 创建时 `DeviceStoreService.InitializeDevice` 静默自动调 bind + monitor save 路径都会推 timezone。

---

## 开放问题

1. 多 branch 全球化后：DST resync cron 用 UTC 触发还是各 branch 本地？当前单 branch（Denver）简单直接
2. `setReportUploadTime` 冬/夏切换当天的报告覆盖不足 24h（-1h）或重叠（+1h），UI 如何展示？（低优先，厂家行为）
