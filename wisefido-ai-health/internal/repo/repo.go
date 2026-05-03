// Package repo 提供 owlrd 数据库读写。
//
// 设计原则：所有数据从 PG 表读，绝不修改 cardagg / sleepace 实时通路（§0.3 原则 1）。
// 本包只做骨架级 helper：连接管理、ETL watermark 读写、ETL error 落库。
// 各指标计算的 SELECT 留到 Phase 1+ 实现。
package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	commonconfig "owl-common/config"

	_ "github.com/lib/pq"
)

// Repo 持有 owlrd 连接池
type Repo struct {
	db *sql.DB
}

// Open 建立连接 + Ping 验证
func Open(ctx context.Context, cfg commonconfig.DatabaseConfig) (*Repo, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	if cfg.MaxConns > 0 {
		db.SetMaxOpenConns(cfg.MaxConns)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	}
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping owlrd: %w", err)
	}
	return &Repo{db: db}, nil
}

// Close 关闭连接池
func (r *Repo) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}

// DB 暴露底层 *sql.DB 供 ETL 包做具体查询（Phase 1+ 用）
func (r *Repo) DB() *sql.DB { return r.db }

// ---------------------------------------------------------------------------
// health_etl_state — watermark 读写
// ---------------------------------------------------------------------------

// Watermark 描述某个 task 的运行进度
//
//	Global watermark：CardID == "" 时 task 全局粒度（如 cohort 月聚合）
//	Per-card watermark：CardID 非空时按 card 粒度
type Watermark struct {
	TaskName        string
	CardID          string // 空字符串 = global
	LastRunAt       time.Time
	LastWatermarkMs int64
	LastCompleteDay time.Time // date 类型，零值代表未跑过
	LastStatus      string    // running | success | failed | ""
	ErrorMessage    string
}

// GetWatermark 读取指定 task + card 的水位线；不存在返回 (zero-Watermark, nil)
func (r *Repo) GetWatermark(ctx context.Context, taskName, cardID string) (Watermark, error) {
	var (
		wm                Watermark
		lastRun           sql.NullTime
		lastWmMs          sql.NullInt64
		lastCompleteDay   sql.NullTime
		lastStatus        sql.NullString
		errMsg            sql.NullString
		row               *sql.Row
	)
	wm.TaskName = taskName
	wm.CardID = cardID

	if cardID == "" {
		row = r.db.QueryRowContext(ctx, `
			SELECT last_run_at, last_watermark_ms, last_complete_date, last_status, error_message
			FROM health_etl_state
			WHERE task_name = $1 AND card_id IS NULL
		`, taskName)
	} else {
		row = r.db.QueryRowContext(ctx, `
			SELECT last_run_at, last_watermark_ms, last_complete_date, last_status, error_message
			FROM health_etl_state
			WHERE task_name = $1 AND card_id = $2
		`, taskName, cardID)
	}
	err := row.Scan(&lastRun, &lastWmMs, &lastCompleteDay, &lastStatus, &errMsg)
	if errors.Is(err, sql.ErrNoRows) {
		return wm, nil
	}
	if err != nil {
		return wm, fmt.Errorf("scan watermark: %w", err)
	}
	if lastRun.Valid {
		wm.LastRunAt = lastRun.Time
	}
	if lastWmMs.Valid {
		wm.LastWatermarkMs = lastWmMs.Int64
	}
	if lastCompleteDay.Valid {
		wm.LastCompleteDay = lastCompleteDay.Time
	}
	if lastStatus.Valid {
		wm.LastStatus = lastStatus.String
	}
	if errMsg.Valid {
		wm.ErrorMessage = errMsg.String
	}
	return wm, nil
}

// UpsertWatermark 写入或更新水位线
func (r *Repo) UpsertWatermark(ctx context.Context, wm Watermark) error {
	var (
		cardArg          interface{}
		lastCompleteArg  interface{}
		errMsgArg        interface{}
		conflictTarget   string
	)
	if wm.CardID == "" {
		cardArg = nil
		conflictTarget = "task_name) WHERE card_id IS NULL"
	} else {
		cardArg = wm.CardID
		conflictTarget = "task_name, card_id) WHERE card_id IS NOT NULL"
	}
	if !wm.LastCompleteDay.IsZero() {
		lastCompleteArg = wm.LastCompleteDay
	}
	if wm.ErrorMessage != "" {
		errMsgArg = wm.ErrorMessage
	}

	q := fmt.Sprintf(`
		INSERT INTO health_etl_state
			(task_name, card_id, last_run_at, last_watermark_ms, last_complete_date,
			 last_status, error_message, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now())
		ON CONFLICT (%s
		DO UPDATE SET
			last_run_at        = EXCLUDED.last_run_at,
			last_watermark_ms  = EXCLUDED.last_watermark_ms,
			last_complete_date = EXCLUDED.last_complete_date,
			last_status        = EXCLUDED.last_status,
			error_message      = EXCLUDED.error_message,
			updated_at         = now()
	`, conflictTarget)
	_, err := r.db.ExecContext(ctx, q,
		wm.TaskName, cardArg, time.Now().UTC(), wm.LastWatermarkMs,
		lastCompleteArg, nullableStr(wm.LastStatus), errMsgArg,
	)
	if err != nil {
		return fmt.Errorf("upsert watermark: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// health_etl_errors — 失败日志（不阻塞批次，§8.5）
// ---------------------------------------------------------------------------

// LogError 记录单 card 单天 ETL 失败；调用方应继续处理下一个 card
func (r *Repo) LogError(ctx context.Context, taskName, cardID string, targetDate time.Time, err error, stack string) {
	if err == nil {
		return
	}
	var cardArg, dateArg, stackArg interface{}
	if cardID != "" {
		cardArg = cardID
	}
	if !targetDate.IsZero() {
		dateArg = targetDate
	}
	if stack != "" {
		stackArg = stack
	}
	_, _ = r.db.ExecContext(ctx, `
		INSERT INTO health_etl_errors
			(task_name, card_id, target_date, error_message, error_stack, occurred_at)
		VALUES ($1, $2, $3, $4, $5, now())
	`, taskName, cardArg, dateArg, err.Error(), stackArg)
	// 故意忽略错误：error log 失败不应该影响 ETL 主流程
}

func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
