---
name: radar_area_vertices_must_apply_angle
description: "区域生成/边界判定必须按 obj.angle 旋转顶点;存储顶点是 PRE-rotation,旋转的矩形(门/床/电视)之前被误删或下发错朝向"
metadata: 
  node_type: memory
  type: project
  originSessionId: 4107a0ef-baff-4b14-88ec-6b94eff6144e
---

2026-06-04 修复:`owlFront/src/utils/radar/radarUtils.ts` 的 `getObjectVertices` 旧实现直接返回 `geometry.data.vertices`,**从不应用 `obj.angle`**。该函数同时喂给边界判定 `isObjectInBoundaryWithTolerance`(决定对象是否进雷达探测框)和区域下发 `getObjectVerticesInRadar`(算 declare_area 坐标),所以两处对旋转对象都用错轮廓。

**铁律:存储的 rectangle 顶点是 PRE-rotation;真实轮廓 = 绕矩形中心按 `obj.angle` 旋转。** 旋转约定与渲染端 `drawObjects.ts`(`ctx.rotate(-(angle*π/180))` 绕中心)和 `getRadarBoundaryVertices` 同式:
`x' = cx + dx·cosA + dy·sinA, y' = cy − dx·sinA + dy·cosA`(cx/cy=矩形中心,a=angle*π/180)。

症状:门/电视画成窄高矩形再转 90° 横躺贴墙,旧逻辑拿没转的竖条判定,竖条底边假性戳出探测框 → 整个对象被“全顶点入框+20cm容差”规则误删,不下发(CD2B 的 Enter201/TV)。0°/180° 的对象因 bbox 不变而侥幸正确。270°/90° 的床(Bed)虽进框但下发的是**没转的矩形(宽高对调)→ 床区朝向一直错**,影响 fall 的床上排除。

修复=在 `getObjectVertices` rectangle 分支按 angle 绕中心旋转(circle 对称不变、line/polygon 暂未处理)。改后 CD2B 6 对象全进:1 Bed+2 Custom(Furniture)+2 Enter+1 Interfere。

全库受影响只有 2 台(同 unit fd00:0:3:112:3::):**9D8A32A1CD2B**(Bed270/Enter270/Enter90/Interfere90)和 **25A859B8333B**(Enter-90)。修公共函数后两台都需重新 Save 把正确区域重下发;其余房间区域对象全是 0/180° 不受影响。

注意:declare_area 坐标全是雷达坐标系 (h,v) 不是画布 (x,y);换算 CD2B 实测 `h=60−canvasY, v=250−canvasX`(含 corn −45° 补偿,随 radar 位置/angle/installModel 变)。相关:[[radar_geometry_point_invariant]] [[radar_measure_flow_design]] [[radar_layout_device_invariant]]
