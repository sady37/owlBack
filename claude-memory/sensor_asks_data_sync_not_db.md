---
name: sensor_asks_data_sync_not_db
description: "sensor 不直连 DB 取 device 信息;需要时同步问 data 要(请求-响应,成功/失败,单 device ≤3s)"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: a821a2af-8448-4576-87bb-f1a826e65a17
---

**铁律(用户 2026-06-05)**:wisefido-sensor **不直连 DB** 取 device 相关信息(device 绑定/spatial_addr、room_visual_layout 几何与 radar.areas、rooms/beds、spatial_config、radar↔bed↔sleepad 关系图)。需要时**同步问 wisefido-data 要**:请求-响应,返回成功或失败,**查一个 device 最多 3 秒**。data 是这些信息的唯一 owner(只它写,也由它查)。

**Why(成本论,用户)**:data 是 owner,**本来就有查询+写库代码**(库里没有它还得主动查源,这份固定、省不掉)。若 sensor **自己连库读**:等于再写一份 sensor 查询代码 + 还得加一层**通知机制**(库变了通知 sensor 失效)→ 在 data 固定代码之上**平白多出 sensor查询 + 通知两摊,不划算**。sensor 直接问 data **复用其查询能力**最省——不重复写、不要缓存失效通知层。

**模式**:**同步阻塞**(非异步/订阅/缓存失效),成功或失败二态,单 device ≤3s 上界。

**约束**:sensor 对这些**只读不写**(写归 data,单源真相 #1.3)。

**★ 2026-06-08 用户纠正(MM)**:**MM 本来就是 sensor 里 create 的**——neighbor/MM 关系图是 **sensor 本地构造**,**不问 data**。推翻本note原 line "MM 关系图由 sensor 问 data 拿"的措辞(那是误记)。即:device 绑定/layout 几何/spatial_config 等**原始**信息仍同步问 data(上文铁律不变),但 MM 关系**在 sensor 内 create**,本地可读。#3 Neighbor wire 据此走本地 MM,详 [[neighbor_wire_build_spec]]。原 [[mm_relationship_matrix]] 关联保留。

**现状违规(待迁移,2026-06-05 审计)**:sensor 多处直连 DB——`internal/roomengine/layout_load.go`/`feedback.go`(room_visual_layout)、`internal/playback/db.go`、`internal/zoneengine/wiring/spatial_cache.go`(rooms/beds/devices)、`internal/service/alarm_enablement.go`(spatial_config)、`cmd/wisefido-sensor/engine_bootstrap.go`/`initial_publish.go`(devices/rooms/beds)→ 都该改成**同步问 data 要**。

**应用**:bed-presence 取床区(`room_visual_layout` radar 对象 `device.iot.radar.areas[]` 里 areaType==2 的 areaId)/ MM 关系 → sensor 同步问 data,不连库。详 [[bed_stale_leftbed_vetoes_radar_inbed]]。关联 [[cardagg_sensor_responsibility_split.md]]、[[feedback_low_freq_invalidate_simple]]。
