package service

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

// StaffPopPendingCounter 统计某 staff 可见卡中「待 pop」卡数：与 CardAlarmState.ToAlarmState 一致
//（pop_alarm_level 非空且 pop_alarm_event_id 非空），每张卡最多计 1。
type StaffPopPendingCounter struct {
	db      *sql.DB
	allowed AllowedCardIDsProvider
}

func NewStaffPopPendingCounter(db *sql.DB, allowed AllowedCardIDsProvider) *StaffPopPendingCounter {
	return &StaffPopPendingCounter{db: db, allowed: allowed}
}

// CountForStaff 返回 tenant 下该 staff 可见、且当前应出 pop 的卡数量。
func (c *StaffPopPendingCounter) CountForStaff(ctx context.Context, tenantID, staffUserID string) (int, error) {
	if c == nil || c.db == nil || c.allowed == nil {
		return 0, nil
	}
	cl, err := c.allowed.GetCardList(ctx, tenantID, staffUserID, "staff")
	if err != nil {
		return 0, err
	}
	if cl == nil {
		return 0, nil
	}
	ids := dedupeCardIDs(cl.AllCardIDs())
	if len(ids) == 0 {
		return 0, nil
	}
	var n int
	err = c.db.QueryRowContext(ctx, `
		SELECT COUNT(*)::int
		FROM cards
		WHERE tenant_id = $1::uuid
		  AND card_id = ANY($2::uuid[])
		  AND COALESCE(TRIM(pop_alarm_level), '') <> ''
		  AND pop_alarm_event_id IS NOT NULL
	`, tenantID, pq.Array(ids)).Scan(&n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func dedupeCardIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
