# API 前后端一致性总表（owlFront ↔ owlBack）

目标：**owlFront 的每个 Vue/模块调用的 API（URL + Method + 入参 + 返回包装）在 owlBack 必须一一对应**，避免上线后 404/字段不一致。

## 统一约定（与 owlFront axios 拦截器一致）

- **统一返回包装**（`owlFront/types/axios.d.ts`）：
  - `{"code":2000,"type":"success","message":"ok","result": ... }`
- **前端会自动携带的 Header**（`owlFront/src/utils/http/axios/index.ts`）：
  - `Authorization: <token>`
  - `X-User-Id: <userId>`（可为空）
  - `X-User-Role: <role>`（可为空）

## 路由前缀分组（owlFront 当前使用）

- `/data/api/v1/...`：数据展示（监控卡片等）
- `/admin/api/v1/...`：后台管理（units/rooms/beds/devices/residents/tags/users/roles/permissions/alarm config 等）
- `/auth/api/v1/...`：登录/找回密码
- `/settings/api/v1/...`：设备监控配置（sleepace/radar）
- `/device/api/v1/...`：设备关系/详情（device relations）
- `/sleepace/api/v1/...`：睡眠报表

> 说明：这些路径在 owlFront 已写死，因此 owlBack 侧必须保持完全一致（哪怕内部服务拆分）。

---

## 总表（按 owlFront `src/api` 模块）

状态说明：
- ✅ 已实现：owlBack 已有对应路由 + 返回结构对齐
- 🟡 已占位：owlBack 有路由骨架/临时实现（可能仅依赖 Redis，不连 DB）
- ❌ 缺失：owlBack 还未提供该路由

| 前端模块（文件） | API 名称 | Method | URL | owlBack 归属建议 | 当前状态 | 备注 |
|---|---|---:|---|---|---|---|
| `src/api/monitors/monitor.ts` | GetVitalFocusCards | GET | `/data/api/v1/data/vital-focus/cards` | `wisefido-data` | ✅ | 读 Redis `vital-focus:card:{card_id}:full`，统一 Result 包装 |
|  | GetVitalFocusCardByResident | GET | `/data/api/v1/data/vital-focus/card/:residentId` | `wisefido-data` | ✅ | 同一路由兼容 residentId/cardId：先按 card_id，未命中再按 resident 扫描 |
|  | GetVitalFocusCardDetail | GET | `/data/api/v1/data/vital-focus/card/:cardId` | `wisefido-data` | ✅ | 同上 |
|  | SaveVitalFocusSelection | POST | `/data/api/v1/data/vital-focus/selection` | `wisefido-data` | ✅ | 临时存 Redis：`vital-focus:selection:user:{X-User-Id}` |
| `src/api/alarm/alarm.ts` | GetConfig | GET | `/admin/api/v1/alarm-cloud` | API 层（待定） | ❌ | 需要 DB：`alarm_cloud` |
|  | UpdateConfig | PUT | `/admin/api/v1/alarm-cloud` | API 层（待定） | ❌ | 需要 DB |
|  | GetEvents | GET | `/admin/api/v1/alarm-events` | API 层（待定） | ❌ | 需要 DB：`alarm_events` |
|  | HandleEvent | PUT | `/admin/api/v1/alarm-events/:id/handle` | API 层（待定） | ❌ | 需要 DB 更新状态 |
| `src/api/devices/device.ts` | GetList | GET | `/admin/api/v1/devices` | API 层（待定） | ❌ | 需要 DB：`devices/device_store/...` |
|  | GetDetail | GET | `/admin/api/v1/devices/:id` | API 层（待定） | ❌ |  |
|  | Update | PUT | `/admin/api/v1/devices/:id` | API 层（待定） | ❌ | 绑定关系变更后需要发布 card 更新事件（后续） |
|  | Delete | DELETE | `/admin/api/v1/devices/:id` | API 层（待定） | ❌ |  |
|  | GetDeviceRelations | GET | `/device/api/v1/device/:id/relations` | API 层（待定） | ❌ |  |
| `src/api/units/unit.ts` | CreateBuilding | POST | `/admin/api/v1/buildings` | API 层（待定） | ❌ |  |
|  | GetBuildings | GET | `/admin/api/v1/buildings` | API 层（待定） | ❌ |  |
|  | UpdateBuilding | PUT | `/admin/api/v1/buildings/:id` | API 层（待定） | ❌ |  |
|  | DeleteBuilding | DELETE | `/admin/api/v1/buildings/:id` | API 层（待定） | ❌ |  |
|  | CreateUnit | POST | `/admin/api/v1/units` | API 层（待定） | ❌ |  |
|  | GetUnits | GET | `/admin/api/v1/units` | API 层（待定） | ❌ |  |
|  | GetUnitDetail | GET | `/admin/api/v1/units/:id` | API 层（待定） | ❌ |  |
|  | UpdateUnit | PUT | `/admin/api/v1/units/:id` | API 层（待定） | ❌ |  |
|  | DeleteUnit | DELETE | `/admin/api/v1/units/:id` | API 层（待定） | ❌ |  |
|  | GetRooms | GET | `/admin/api/v1/rooms` | API 层（待定） | ❌ |  |
|  | CreateRoom | POST | `/admin/api/v1/rooms` | API 层（待定） | ❌ |  |
|  | UpdateRoom | PUT | `/admin/api/v1/rooms/:id` | API 层（待定） | ❌ |  |
|  | DeleteRoom | DELETE | `/admin/api/v1/rooms/:id` | API 层（待定） | ❌ |  |
|  | GetBeds | GET | `/admin/api/v1/beds` | API 层（待定） | ❌ |  |
|  | CreateBed | POST | `/admin/api/v1/beds` | API 层（待定） | ❌ |  |
|  | UpdateBed | PUT | `/admin/api/v1/beds/:id` | API 层（待定） | ❌ |  |
|  | DeleteBed | DELETE | `/admin/api/v1/beds/:id` | API 层（待定） | ❌ |  |
| `src/api/resident/resident.ts` | GetList | GET | `/admin/api/v1/residents` | API 层（待定） | ❌ |  |
|  | GetDetail | GET | `/admin/api/v1/residents/:id` | API 层（待定） | ❌ |  |
|  | Create | POST | `/admin/api/v1/residents` | API 层（待定） | ❌ |  |
|  | Update | PUT | `/admin/api/v1/residents/:id` | API 层（待定） | ❌ |  |
|  | Delete | DELETE | `/admin/api/v1/residents/:id` | API 层（待定） | ❌ |  |
|  | UpdatePHI | PUT | `/admin/api/v1/residents/:id/phi` | API 层（待定） | ❌ |  |
|  | UpdateContact | PUT | `/admin/api/v1/residents/:id/contacts` | API 层（待定） | ❌ |  |
| `src/api/admin/tags/tags.ts` | GetList | GET | `/admin/api/v1/tags` | API 层（待定） | ❌ |  |
|  | Create | POST | `/admin/api/v1/tags` | API 层（待定） | ❌ |  |
|  | Update | PUT | `/admin/api/v1/tags/:id` | API 层（待定） | ❌ |  |
|  | Delete | DELETE | `/admin/api/v1/tags` | API 层（待定） | ❌ |  |
|  | AddObjects | POST | `/admin/api/v1/tags/:id/objects` | API 层（待定） | ❌ |  |
|  | RemoveObjects | DELETE | `/admin/api/v1/tags/:id/objects` | API 层（待定） | ❌ |  |
|  | DeleteTagType | DELETE | `/admin/api/v1/tags/types` | API 层（待定） | ❌ |  |
|  | GetTagsForObject | GET | `/admin/api/v1/tags/for-object` | API 层（待定） | ❌ |  |
| `src/api/admin/user/user.ts` | GetList | GET | `/admin/api/v1/users` | API 层（待定） | ❌ |  |
|  | Create | POST | `/admin/api/v1/users` | API 层（待定） | ❌ |  |
|  | Update | PUT | `/admin/api/v1/users/:id` | API 层（待定） | ❌ |  |
|  | Delete | DELETE | `/admin/api/v1/users/:id` | API 层（待定） | ❌ |  |
|  | ResetPassword | POST | `/admin/api/v1/users/:id/reset-password` | API 层（待定） | ❌ |  |
|  | ResetPin | POST | `/admin/api/v1/users/:id/reset-pin` | API 层（待定） | ❌ |  |
| `src/api/admin/role/role.ts` | GetList | GET | `/admin/api/v1/roles` | API 层（待定） | ❌ |  |
|  | Create | POST | `/admin/api/v1/roles` | API 层（待定） | ❌ |  |
|  | Update | PUT | `/admin/api/v1/roles/:id` | API 层（待定） | ❌ |  |
|  | Delete | DELETE | `/admin/api/v1/roles/:id` | API 层（待定） | ❌ |  |
|  | UpdateStatus | PUT | `/admin/api/v1/roles/:id/status` | API 层（待定） | ❌ |  |
| `src/api/admin/role-permission/rolePermission.ts` | GetList | GET | `/admin/api/v1/role-permissions` | API 层（待定） | ❌ |  |
|  | Create | POST | `/admin/api/v1/role-permissions` | API 层（待定） | ❌ |  |
|  | BatchCreate | POST | `/admin/api/v1/role-permissions/batch` | API 层（待定） | ❌ |  |
|  | Update | PUT | `/admin/api/v1/role-permissions/:id` | API 层（待定） | ❌ |  |
|  | Delete | DELETE | `/admin/api/v1/role-permissions/:id` | API 层（待定） | ❌ |  |
|  | UpdateStatus | PUT | `/admin/api/v1/role-permissions/:id/status` | API 层（待定） | ❌ |  |
|  | GetResourceTypes | GET | `/admin/api/v1/role-permissions/resource-types` | API 层（待定） | ❌ |  |
| `src/api/service-level/serviceLevel.ts` | GetList | GET | `/admin/api/v1/service-levels` | API 层（待定） | ❌ |  |
| `src/api/settings/settings.ts` | GetSleepaceSettings | GET | `/settings/api/v1/monitor/sleepace/:deviceId` | API 层（待定） | ❌ |  |
|  | UpdateSleepaceSettings | PUT | `/settings/api/v1/monitor/sleepace/:deviceId` | API 层（待定） | ❌ |  |
|  | GetRadarSettings | GET | `/settings/api/v1/monitor/radar/:deviceId` | API 层（待定） | ❌ |  |
|  | UpdateRadarSettings | PUT | `/settings/api/v1/monitor/radar/:deviceId` | API 层（待定） | ❌ |  |
| `src/api/report/report.ts` | SleepaceReports | GET | `/sleepace/api/v1/sleepace/reports/:id` | API 层（待定） | ❌ |  |
|  | SleepaceReportDetail | GET | `/sleepace/api/v1/sleepace/reports/:id/detail` | API 层（待定） | ❌ |  |
|  | SleepaceReportsDates | GET | `/sleepace/api/v1/sleepace/reports/:id/dates` | API 层（待定） | ❌ |  |
| `src/api/card-overview/cardOverview.ts` | GetList | GET | `/admin/api/v1/card-overview` | API 层（待定） | ❌ |  |
| `src/api/address/address.ts` | CreateAddress | POST | `/admin/api/v1/addresses` | API 层（待定） | ❌ |  |
|  | GetAddresses | GET | `/admin/api/v1/addresses` | API 层（待定） | ❌ |  |
|  | GetAddressDetail | GET | `/admin/api/v1/addresses/:id` | API 层（待定） | ❌ |  |
|  | UpdateAddress | PUT | `/admin/api/v1/addresses/:id` | API 层（待定） | ❌ |  |
|  | DeleteAddress | DELETE | `/admin/api/v1/addresses/:id` | API 层（待定） | ❌ |  |
|  | AllocateCarrier | POST | `/admin/api/v1/addresses/:id/allocate/carrier` | API 层（待定） | ❌ |  |
|  | AllocateResident | POST | `/admin/api/v1/addresses/:id/allocate/resident` | API 层（待定） | ❌ |  |
|  | AllocateDevice | POST | `/admin/api/v1/addresses/:id/allocate/device` | API 层（待定） | ❌ |  |
| `src/api/auth/auth.ts` | Login | POST | `/auth/api/v1/login` | API 层（待定） | ❌ |  |
|  | SearchInstitutions | GET | `/auth/api/v1/institutions/search` | API 层（待定） | ❌ |  |
|  | SendVerificationCode | POST | `/auth/api/v1/forgot-password/send-code` | API 层（待定） | ❌ |  |
|  | VerifyCode | POST | `/auth/api/v1/forgot-password/verify-code` | API 层（待定） | ❌ |  |
|  | ResetPassword | POST | `/auth/api/v1/forgot-password/reset` | API 层（待定） | ❌ |  |

---

## 下一步建议（确保“每个 Vue 都能跑起来”）

1. 先补齐 **当前 UI 页面会调用的管理端 API**（units/devices/residents/tags/users/roles/permissions/alarm cloud/alarm events）。
2. 把这些路由统一落在一个 **HTTP API 服务**（建议继续扩展 `wisefido-data` 作为 API 层），保持路径不变。
3. DB/Redis 未起时：可以先用 **占位实现**（返回 `code=-1` 的明确错误），避免前端 silent failure；等 DB 建好再逐个替换为真实实现。




