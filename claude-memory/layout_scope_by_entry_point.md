---
name: layout-scope-by-entry-point
description: 2026-05-16 拍板 — room_visual_layout 的 scope (/80 unit / /88 room / /128 device) 由 URL 切入点决定，不让用户在 UI 选；当前 radar-trajectory URL 是 device 切入故 save target = device /128
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 73e2018d-7ab1-4a06-b638-4f14d5bf7746
---

room_visual_layout 三档作用域 (/80 unit / /88 room / /128 device) 的判定规则：**URL 切入点 = layout scope**。

**Why:** 用户 2026-05-16 否决了"在 import 时弹 scope picker 让用户选 Unit/Room/Device"的设计。理由是简化：内容推断 vs 用户选 vs 入口决定 三种方案里，入口决定最稳——零歧义、零选择项、code 路径单一。

**How to apply:**
- 当前 URL `/monitoring/radar-trajectory/:cardId/:deviceId` 是 **device 切入** → save target = 该 device 的 /128 CIDR
- backend `ListCardDevicesByDeviceID` 返响应字段 `spatial_prefix` = device 的 /128 CIDR；FE 存在 `canvasStore.params.spatialPrefix` 用于 save 和 export filename
- 未来若新增 room 切入 URL (`/admin/room/:id/layout` 类) → 该 handler 返 /88 CIDR；unit 切入 → /80
- import 只换 `objectsStore.objects`，**不动 scope**：哪怕导入文件是 Unit /80 scope，进到 device canvas 后存的仍是当前 /128
- modal 上文件 scope ≠ 当前 canvas scope 仅作 warning（"objects will be saved to current canvas's prefix"），不阻塞

**红线：永远不要在 import 时改 `params.spatialPrefix` / `params.roomId` 切 scope** — scope 是 URL 的属性，不是文件的属性。

**与 [[platform_agent_addressing]] 一致**：IPv6 prefix containment 是权威，URL 是 scope 的唯一来源。

**实现位置：**
- backend：`wisefido-data/internal/service/radar_install_service.go` `ListCardDevicesByDeviceID`（返 `spatialPrefix` = dev.DeviceIPv6 + "/128"）
- backend：`wisefido-data/internal/http/radar_handler.go` `PutRoomLayout`（优先读 body.spatial_prefix 覆盖 path roomID，支持 /80 unit save）
- FE：`utils/radar/types.ts` `CanvasParams.spatialPrefix`；`Toolbar.vue` `unifiedSave` 用 `params.spatialPrefix || params.roomId` 作 saveTarget
- FE：`utils/radar/layoutFilename.ts` encode/parse `layout-<ipv6>_<mask>.json`
