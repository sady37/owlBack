package domain

// CardSyncAffected 单次同步中受影响的卡片（用于同步后发送 config.card.*）
// 与 wisefido-card-manage 的 CreateCardsForUnit 行为对应：create/update/delete
type CardSyncAffected struct {
	TenantID string // 租户
	CardID   string // 卡片 ID
	UnitID   string // 单元 ID
	Op       string // "created" | "updated" | "deleted"
}
