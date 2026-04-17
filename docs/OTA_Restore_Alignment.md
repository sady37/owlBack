# OTA 全栈恢复 — 对齐提示词

## 当前状态
- **owlBack** commit `c71b2be` on main — OTA后端 + device command endpoints
- **owlFront** commit `bfece86` on main — OTA前端 + radar toolbar改进 + refresh按钮
- 基于 `00ae2f0`(card_id重构) + `10f8e16`(PHI加密) 之上
- 日期: 2026-04-17 (更新)

---

## 1. 数据库 (已执行)

```sql
-- device_store 表 OTA 字段
ALTER TABLE device_store
  ADD COLUMN IF NOT EXISTS ota_permit VARCHAR(10) DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS ota_way VARCHAR(10) DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS ota_schedule VARCHAR(20) DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS ota_status VARCHAR(20) DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS ota_progress INTEGER DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS ota_error TEXT DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS ota_updated_at TIMESTAMPTZ DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS ota_firmware_url TEXT DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS ota_firmware_sha256 VARCHAR(64) DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS ota_firmware_size BIGINT DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS ota_tenant_approved BOOLEAN DEFAULT FALSE;

ALTER TABLE device_store ALTER COLUMN allow_access SET DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS idx_device_store_ota_status ON device_store (ota_status) WHERE ota_status IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_device_store_ota_permit_way ON device_store (ota_permit, ota_way) WHERE ota_permit IS NOT NULL;

-- Unallocated 租户已删除
DELETE FROM iot_timeseries WHERE tenant_id = '00000000-0000-0000-0000-000000000002';
DELETE FROM tenants WHERE tenant_id = '00000000-0000-0000-0000-000000000002';
```

---

## 2. 架构概要

### OTA 两阶段设计
1. **TCP OTA (P0)**: 老MCU通过TCP连接 → cmux分流 → protobuf OTAPush(type=16) → 设备下载固件 → 自动重启
2. **MQTT OTA (P1)**: 新MCU通过MQTT → cmd:"ota" → 设备下载固件 → 自动重启

### 端口复用 (cmux)
- 8443 端口: TLS → HTTPS Auth, 原始TCP → 设备OTA连接
- `/firmware/` 静态文件服务（固件下载）

### 租户迁移规则
- Unallocated 已删除，默认租户为 System
- 业务租户间不能直接迁移，必须经 System 或 Trash
- Delete = Move to Trash（非物理删除）
- 迁移时清除 bound_room_id / bound_bed_id

### ESP32 端口行为
- 老MCU: OTA后 web server port 回 443，由外部 BLE 重配，owlBack 不处理
- 新MCU: 自动读取 webserver:port，无需干预

---

## 3. owlBack — wisefido-data (10个文件)

| 文件 | 改动 |
|------|------|
| `domain/device_store.go` | 11 OTA 字段 + 4 Set flags + OTATenantApproved, ToJSON 输出 |
| `domain/device.go` | 8 OTA 字段(从 device_store JOIN), ToJSON 输出 |
| `repository/device_store_repo.go` | DeviceStoreFilters 增加 DeviceUID/DeviceCode/DeviceName/FirmwareVersion/AllowAccess/OTAStatus/OTAPermit/OTAWay |
| `repository/postgres_device_store.go` | SELECT/Scan 3处扩展; OTA auto-reset; 迁移规则(business→business禁止); 默认tenant=System; isDeviceStorePivotTenant 去掉 Unallocated |
| `repository/postgres_devices.go` | SELECT/Scan 3处扩展(8个OTA字段) |
| `http/device_store_handler.go` | ListDeviceStores/ExportDeviceStores 读取 OTA 过滤参数 |
| `http/device_store_payload.go` | payloadToDeviceStore/Patch 解析 OTA 字段 + Set flags |
| `http/device_handler.go` | SetDB 方法; ApproveOTA: POST `/admin/api/v1/devices/{id}/ota-approve`; 路由在 PUT/DELETE 之前 |
| `cmd/wisefido-data/main.go` | deviceHandler.SetDB(db) |

---

## 4. owlBack — wisefido-qinglan (13个文件)

| 文件 | 改动 |
|------|------|
| `proto/qinglan.proto` + `gen/qinglan.pb.go` | **protoc 生成**，不可手写。关键类型: GetServerReq(1), RegisterReq(3), Heartbeat(7/8), OTAReq(16), OtaResponse(17), OTAProgress(18) |
| `internal/tcp/protocol.go` | Frame{Type byte, Data []byte}, ReadFrame/WriteFrame, 帧格式: [2B长度LE][1B类型][data] |
| `internal/tcp/session.go` | DeviceSession(UID/Conn/Type/SfVer/HwVer/ConnectAt/LastHeart), SessionManager(Connect/Disconnect/GetByUID/OnDisconnect回调) |
| `internal/tcp/handlers.go` | HandleFrame: type 1→GetServer分发, 3→Register, 7→Heartbeat, 17→OTAResp, 18→OTAProgress(-1/10/25/56/100) |
| `internal/tcp/server.go` | Server(Sessions/ServerAddr/ServerPort/**OnProgress**), Serve(listener), handleConnection(5s首帧/90s超时), PushOTA(uid, *Frame) |
| `internal/ota/ota_service.go` | Manager(TCPServer/FirmwareDir/FirmwareURL), PushRequest(UID/EspFirmware/EspVersion/EspSHA256/RadarFirmware/RadarVersion/RadarSHA256), PushToDevice(req) → PushResult, getFirmwareInfo(SHA256), ListFirmwareFiles |
| `internal/ota/scheduler.go` | Scheduler(db/otaPushFn/interval/fwDir/fwURL), scan: 查 device_store permit+way+schedule+approved, MQTT推送, ParseSchedule/IsInScheduleWindow(MMDDHH+xH) |
| `internal/http/ota_handler.go` | 6 API: POST trigger/{uid}, POST trigger/batch, GET status, GET/POST/DELETE firmware/{vendor} |
| `internal/http/https_server.go` | cmux: TLS→HTTPS Auth, Any→TCP; /firmware/ 静态; TCPServer()/OTAManager()/AuthService() getter |
| `internal/http/server.go` | otaHandler 字段 + SetOTAHandler |
| `internal/http/auth_service.go` | 去掉 Unallocated 拒绝, System 允许 auth |
| `internal/publisher/mqtt_publisher.go` | PublishOTA(uid, data) + PublishReboot(uid, dev:0) |
| `internal/consumer/mqtt_consumer.go` | db 字段 + SetDB + handleOTAReturn(ota_return 进度→device_store) |
| `cmd/wisefido-qinglan/main.go` | TCP OnProgress 回调→DB; OTAHandler 接线; Scheduler 创建 |

---

## 5. owlFront (10个文件)

| 文件 | 改动 |
|------|------|
| `api/admin/device-store/model/deviceStoreModel.ts` | DeviceStore: 11 OTA字段; GetDeviceStoresParams: tenant_id + OTA filters |
| `api/admin/device-store/deviceStore.ts` | Api enum: OTATrigger/OTAStatus/OTAFirmware → `/qinglan/api/v1/ota/...`; 6个函数(trigger/batch/status/list/upload/delete), 全部 isTransformResponse:false |
| `api/devices/model/deviceModel.ts` | Device: 8 OTA字段 |
| `api/devices/device.ts` | approveOTAApi: POST `/admin/api/v1/devices/{id}/ota-approve` |
| `views/admin/devicestore.vue` | OTA模式切换(otaMode); 行选择(selectedRowKeys); 工具栏(Delete→Trash/Restart/IoTServer/Firmware); OTA列(permit/way/schedule/status/progress/approved)用_otaOnly标记; 固件管理模态框(vendor选择/上传/列表/删除); Save包含ota_permit/ota_way/ota_schedule |
| `views/devices/DeviceList.vue` | OTA模式切换; 行选择; 工具栏(Delete→Trash/Restart/WiFi); OTA列(status/way/approve)用_otaOnly; handleApproveOTA调approveOTAApi |
| `components/layout/Menu.vue` | svgIconMap 映射(overview/alarm/cloud/resident/device/unit/user/branch/tenant); overview用menu-svg-icon-overview(filter:none保留原色) |
| `types/menu.ts` | 菜单顺序: Overview→Alarms→Cloud→Residents(无divider)→Branches(divider)→Units→Devices→Users→Tenants→DeviceStore→Role→Permission; SVG图标引用 |
| `vite.config.ts` | `/qinglan/` proxy → localhost:8081 |

---

## 6. 基础设施 (已配置，无需修改)

- **nginx**: `/etc/nginx/sites-available/owlfront` 已有 `/qinglan/` proxy (rewrite strip prefix → 8081)
- **sudoers**: `wisefido-systemctl`(systemctl免密) + `wisefido-docker`(docker免密)

---

## 7. 关键约束（必须遵守）

1. **protobuf 必须用 ota-main 的 protoc 生成文件** — 手写会导致 proto.Unmarshal 不兼容
2. **OTAReq 字段名大小写混合**: `Espsfver` / `ESPFileUrl` / `ESPFileSize` / `ESPFileSHA256`（跟 proto 定义一致）
3. **Frame.Data** 不是 Frame.Payload
4. **OTAProgressCallback**: `func(uid string, progress int, message string)` — progress 编码: -1=失败, 0=接受, 10/25/56/100=阶段
5. **tcp.NewServer 第二参数 uint32**（不是 int32）
6. **PushToDevice(req PushRequest) PushResult** — 单参数
7. **scheduler 发 MQTT OTA push**（不是 reboot）
8. **disconnect 不报告 OTA 失败**（defer 中无 OnProgress(-1)）
9. **Reboot 命令**: `dev:0`（不是 `restart:true`）
10. **ESP32 老MCU OTA后 port 回443** — BLE外部重配，代码不处理

---

## 8. 已修复的 card_id 一致性问题 (2026-04-17)

`postgres_devices.go` 中 `GetDevice()` 和 `GetDeviceByUID()` 缺少 card_id 子查询，
导致设备列表有 card_id 但详情接口返回 null。已补齐与 `ListDevices()` 一致的子查询：

```sql
(SELECT c.card_id::text FROM cards c,
   jsonb_array_elements(COALESCE(c.devices, '[]'::jsonb)) AS j
 WHERE (j->>'device_id') = d.device_id::text
   AND c.tenant_id = d.tenant_id
 LIMIT 1) AS card_id
```

---

## 9. 待验证 / 待实现

- [ ] scheduler.go SQL 字段名运行时验证（ota_mode vs ota_way, ota_target_version 等列名是否匹配实际DB）
- [ ] Restart/WiFi/IoTServer 工具栏按钮目前是 placeholder (message.info)
- [ ] 固件目录结构: `owlBack/ota/{vendor}/{firmware.bin}`
- [ ] OTA 端到端测试: TCP OTA 推送 → 设备下载 → 进度回报 → DB 更新
- [ ] MQTT OTA scheduler 端到端测试
- [ ] ant-design 原生列过滤（checkbox filter）未在本次恢复中实现（原版有，当前版本列头无过滤下拉）
