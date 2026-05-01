package observation

import "encoding/json"

// EventItem 表示 iot:event:stream / iot:alarm:stream 的 dataValue 中单条事件/报警 payload。
// Stage 2：dataCategory / event_name 字段已删，事件类型权威由 envelope.Category（IoTStreamMessage.Category）承担。
// 必填：event_since、event_status；其它字段按事件子类按需填写。
type EventItem struct {
	EventID      string `json:"event_id,omitempty"`
	EventSince   int64  `json:"event_since"`  // 必填，发生时间戳 ms
	EventStatus  string `json:"event_status"` // 必填：start / end / instant / pending
	EventEnd     int64  `json:"event_end,omitempty"`
	EventReason  string `json:"event_reason,omitempty"` // 解除原因，与 event_value 不同
	EventLevel   string `json:"event_level,omitempty"`
	EventValue   int64  `json:"event_value,omitempty"`   // 触发时单一测量值（如 pose、心率），非整条 payload
	EventPayload string `json:"event_payload,omitempty"` // 可选，扩展字符串（如整条 payload 的 JSON、备注等）

	TrackID int `json:"track_id,omitempty"`
}

// EventItemToDataMap 将 EventItem 转为 map[string]interface{}（键与 json 标签一致），供 NewSingleItemMessage 等使用。
func EventItemToDataMap(item *EventItem) (map[string]interface{}, error) {
	if item == nil {
		return nil, nil
	}
	b, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}


/*
纯 event（发 iot:event:stream）：平铺 —— EventItem 只填基础字段（dataCategory、event_name、event_since、event_status、track_id），业务字段直接写在 data 顶层，不塞进 EventPayload。
alarm（发 iot:alarm:stream）：EventPayload —— 把原始/完整 payload 打成 JSON 放进 event_payload，方便排查和审计，必要时再在顶层补少量字段
*/