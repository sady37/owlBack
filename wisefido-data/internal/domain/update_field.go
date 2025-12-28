package domain

import (
	"encoding/json"
	"time"
)

// UpdateAction 更新操作类型
// 用于区分三种状态：保持不变、更新值、删除（设置为 NULL）
type UpdateAction int

const (
	UpdateActionKeep    UpdateAction = iota // 保持不变（不更新）
	UpdateActionUpdate                      // 更新值
	UpdateActionDelete                      // 删除（设置为 NULL）
)

// UpdateString 字符串更新字段
type UpdateString struct {
	Action UpdateAction
	Value  string
}

// UpdateBytes 字节数组更新字段
type UpdateBytes struct {
	Action UpdateAction
	Value  []byte
}

// UpdateBool 布尔更新字段
type UpdateBool struct {
	Action UpdateAction
	Value  bool
}

// UpdateInt 整数更新字段
type UpdateInt struct {
	Action UpdateAction
	Value  int
}

// UpdateInt64 64位整数更新字段
type UpdateInt64 struct {
	Action UpdateAction
	Value  int64
}

// UpdateFloat64 浮点数更新字段
type UpdateFloat64 struct {
	Action UpdateAction
	Value  float64
}

// UpdateTime 时间更新字段
type UpdateTime struct {
	Action UpdateAction
	Value  *time.Time
}

// UpdateJSON JSON更新字段（json.RawMessage）
type UpdateJSON struct {
	Action UpdateAction
	Value  json.RawMessage
}

