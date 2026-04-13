package repository

import (
	"database/sql"
	"fmt"
	"log"

	commonconfig "owl-common/config"

	_ "github.com/lib/pq"
)

// NewDatabaseConnection 创建数据库连接
func NewDatabaseConnection(cfg *commonconfig.DatabaseConfig) (*sql.DB, error) {
	dsn := cfg.GetDSN()
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 设置连接池
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * 60) // 5分钟

	log.Printf("Database connected: %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)
	return db, nil
}