package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrSlotExhausted —— 某 parent 容器内 [1, ceiling] 全部 slot 已占用，无可分配。
var ErrSlotExhausted = errors.New("slot exhausted")

// slotQuerier 同时被 *sql.Tx 和 *sql.DB 满足。
type slotQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// AllocSlotReclaim 在 [1, ceiling] 内为某容器分配 spatial slot：
//
//  1. 先 MAX+1（maxQ）—— 单调递增、不复用空号；slot 稳定，删了再建不会撞到刚释放的旧地址。
//  2. MAX+1 撞顶（> ceiling，即可用号段已用完）→ 回收：generate_series(1, ceiling) 扣掉
//     usedQ 返回的已占用集，取最低的空闲 slot（复用被删留下的空号）。
//  3. 全满 → ErrSlotExhausted。
//
// maxQ 与 usedQ 接收相同的 args（容器 parent；tenant 这类全局分配时 args 为空）。
// ceiling 是编译期常量，内联进 SQL（int，无注入）。保留位天然规避：slot 0（unbound 哨兵）和
// ceiling 之上的顶值（byte 0xFF / uint16 0xFFFF = wildcard / subject namespace）都不在 [1, ceiling]。
func AllocSlotReclaim(ctx context.Context, q slotQuerier, ceiling int, maxQ, usedQ string, args ...any) (int, error) {
	var n int
	if err := q.QueryRowContext(ctx, maxQ, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("slot MAX+1: %w", err)
	}
	if n >= 1 && n <= ceiling {
		return n, nil
	}
	reclaimQ := fmt.Sprintf(
		`SELECT s FROM generate_series(1, %d) AS s WHERE s NOT IN (%s) ORDER BY s LIMIT 1`,
		ceiling, usedQ)
	switch err := q.QueryRowContext(ctx, reclaimQ, args...).Scan(&n); err {
	case nil:
		return n, nil
	case sql.ErrNoRows:
		return 0, ErrSlotExhausted
	default:
		return 0, fmt.Errorf("slot reclaim: %w", err)
	}
}
