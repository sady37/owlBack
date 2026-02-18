package card

import (
	"context"
	"database/sql"
	"fmt"
)

// InitializeCardAlarmCounts 在 card INSERT 之后调用，初始化报警计数。
// 内部委托给 recalcAndUpdateCards：从 cards.devices 获取设备列表，
// 统计 active alarms，更新 unhandled_alarm_0~4 + pop_alarm_*。
//
// 语义上表示"首次初始化"，与 RecalcCardAlarmState 逻辑相同。
func InitializeCardAlarmCounts(ctx context.Context, db *sql.DB, cardID, tenantID string) (*CardAlarmState, error) {
	if cardID == "" {
		return nil, fmt.Errorf("cardID is required")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	state, err := recalcAndUpdateCards(ctx, tx, cardID, tenantID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return state, nil
}
