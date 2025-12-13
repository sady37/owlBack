package database

import (
	"database/sql"
	"fmt"
	"owl-common/config"
	
	_ "github.com/lib/pq"
)

// NewPostgresDB 创建PostgreSQL数据库连接
func NewPostgresDB(cfg *config.DatabaseConfig) (*sql.DB, error) {
	dsn := cfg.GetDSN()
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// 设置连接池参数
	if cfg.MaxConns > 0 {
		db.SetMaxOpenConns(cfg.MaxConns)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	}
	
	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	return db, nil
}

// Close 关闭数据库连接
func Close(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

