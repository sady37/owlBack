package service

import (
	"context"

	"owl-common/card"
)

// PublishFields 待发布的 CardStatus 各块，均为可选；只传需要发布的块（一项或多项）。
// device:status:{ipv6} 独立 Hash 由 monitor/event/alarm 流事件驱动 + 看门狗 fail-safe，不在此发布。
type PublishFields struct {
	Target     *card.TargetState
	RoomState  *card.RoomState
	BedState   *card.BedState
	AlarmState *card.AlarmState
	Message    map[string]interface{}
}

// PublishCardStatus 统一发布入口：按 PublishFields 只写传入的非 nil 块到 Hash 并推 card:status:stream。
func PublishCardStatus(ctx context.Context, w *card.Writer, cardID string, fields PublishFields) error {
	status := &card.CardStatus{CardID: cardID}
	if fields.Target != nil {
		status.Target = fields.Target
	}
	if fields.RoomState != nil {
		status.RoomState = fields.RoomState
	}
	if fields.BedState != nil {
		status.BedState = fields.BedState
	}
	if fields.AlarmState != nil {
		status.AlarmState = fields.AlarmState
	}
	if fields.Message != nil {
		status.Message = fields.Message
	}
	return w.WriteCardStatus(ctx, status)
}

// PublishCardStatusSilent 只写 Hash，不推 stream（如启动全量同步）。
func PublishCardStatusSilent(ctx context.Context, w *card.Writer, cardID string, fields PublishFields) error {
	status := &card.CardStatus{CardID: cardID}
	if fields.Target != nil {
		status.Target = fields.Target
	}
	if fields.RoomState != nil {
		status.RoomState = fields.RoomState
	}
	if fields.BedState != nil {
		status.BedState = fields.BedState
	}
	if fields.AlarmState != nil {
		status.AlarmState = fields.AlarmState
	}
	if fields.Message != nil {
		status.Message = fields.Message
	}
	return w.WriteCardStatusSilent(ctx, status)
}
