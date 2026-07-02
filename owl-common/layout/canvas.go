// Package layout 定义 room_visual_layout.canvas 的跨服务权威 schema——
// types.ts BaseObject 的 BE 语义镜像（去掉纯 FE 渲染/交互态 visual/interactive/zIndex/measureType/view）。
//
// data/sensor/Xsensorv1 各自的解析 struct 直接用 CanvasObject，令 canvas 字段的 key 与命名只此一处、
// 不再三处各写各的漂移；FE(types.ts)↔BE(本文件) 靠字段名逐字对齐（rules.md #S1/#S2）。
//
// 分层 = 正交三轴（rules.md #S2）：物理(是什么) / 几何(什么形状朝向) / 空间(什么区域语义)，
// 用 Go 匿名嵌入组合 —— Go 层三轴分离清晰，JSON 层自动 flatten 成扁平、与 FE 逐字段一致。
//
// 强类型 vs raw 的界线：**跨对象共享的身份/语义**（ObjectType/AreaType/geometry.type/device 身份）
// 强类型；**每个 object 类型各不相同的私有属性**（radar installModel/boundary/hfov/vfov、
// geometry 各形状顶点 data、furniture 子属性）一律 json.RawMessage，由各消费方按 ObjectType 自解析。
package layout

import "encoding/json"

// PhysicalProps 物理轴：这个对象"是什么"——种类 + 承载的设备 + 设备绑定身份。
type PhysicalProps struct {
	ObjectType       string       `json:"ObjectType"` // 物理种类枚举（Radar/Sleepad/Bed/Chair/Curtain...）
	Device           CanvasDevice `json:"device"`
	DeviceAddr       string       `json:"device_addr,omitempty"`       // 顶层绑定设备 addr（/128）
	BindedDeviceAddr string       `json:"bindedDeviceAddr,omitempty"`  // 画布对象↔设备关联
}

// CanvasDevice 设备属性：category/type + IoT 身份（共享强类型）+ 私有配置（raw）。
type CanvasDevice struct {
	Category  string          `json:"category"` // iot / furniture / structure
	Type      string          `json:"type,omitempty"`
	IOT       *CanvasIoT      `json:"iot,omitempty"`
	Furniture json.RawMessage `json:"furniture,omitempty"` // 私有命名空间：enter{target,kind} / bed 子类型 等，每个 object 类型各不相同，各消费方按 ObjectType 自解析
}

// CanvasIoT IoT 设备身份（共享强类型）+ 各设备私有配置（raw）。
type CanvasIoT struct {
	DeviceUID  string          `json:"device_uid,omitempty"`  // 终身身份 logMAC：declare 下发认它（免疫 addr 漂移）
	DeviceAddr string          `json:"deviceAddr,omitempty"`  // 实测绑定 addr
	IsOnline   bool            `json:"isOnline,omitempty"`
	Radar      json.RawMessage `json:"radar,omitempty"`   // 私有配置 raw：installModel/rotation/boundary/hfov/vfov/baseline...
	Sleepad    json.RawMessage `json:"sleepad,omitempty"` // 私有配置 raw
	Sensor     json.RawMessage `json:"sensor,omitempty"`  // 私有配置 raw
}

// GeometryProps 几何轴：形状 / 朝向 / 高度。
type GeometryProps struct {
	Geometry CanvasGeometry `json:"geometry"`
	Rotation *float64       `json:"angle,omitempty"` // 平面旋转角（0-360°）
	Height   *int           `json:"height,omitempty"`
}

// CanvasGeometry 几何：type 判别（共享强类型）+ 各形状顶点（私有 raw）。
type CanvasGeometry struct {
	Type string          `json:"type"` // point/line/circle/rectangle/polygon/measure_dots
	Data json.RawMessage `json:"data"` // 私有：各形状顶点结构不同，各消费方按 Type 解析
}

// SpatialProps 空间轴：区域语义（判据唯一源）。只放"人人共享"的 AreaType；
// object 类型私有的空间语义（Enter 的通向/kind、Bed 子类型等）走 device.furniture 私有 raw。
type SpatialProps struct {
	AreaType int `json:"AreaType"` // 0-9 / -1 非区域（经 observation.AreaTypeFrom 转）
}

// CanvasObject = 元信息 + 物理/几何/空间三轴（嵌入 → JSON 扁平，对齐 FE types.ts BaseObject）。
type CanvasObject struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Source string `json:"source,omitempty"`
	PhysicalProps
	GeometryProps
	SpatialProps
}
