package repository

import (
	"database/sql"
	"fmt"

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

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * 60)

	return db, nil
}