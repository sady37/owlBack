package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// 从环境变量获取数据库连接信息，如果没有则使用默认值
	cfg := &database.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:      parseInt(getEnv("DB_PORT", "5432"), 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Database: getEnv("DB_NAME", "owlrd"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// 连接数据库
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// 查询 done 和 smith 的 can_view_status
	query := `
		SELECT 
			r.resident_id,
			r.tenant_id,
			r.resident_account,
			r.nickname,
			r.can_view_status,
			r.status,
			r.service_level,
			r.admission_date,
			r.family_tag
		FROM residents r
		WHERE LOWER(r.nickname) IN ('done', 'smith')
		   OR LOWER(r.resident_account) IN ('done', 'smith')
		ORDER BY r.nickname;
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer rows.Close()

	fmt.Println("查询结果：")
	fmt.Println("=" + string(make([]byte, 100)))
	fmt.Printf("%-40s %-40s %-20s %-20s %-15s %-15s %-15s %-15s %-20s\n",
		"resident_id", "tenant_id", "resident_account", "nickname", "can_view_status", "status", "service_level", "admission_date", "family_tag")
	fmt.Println("=" + string(make([]byte, 100)))

	var count int
	for rows.Next() {
		var residentID, tenantID, residentAccount, nickname, status, serviceLevel, familyTag sql.NullString
		var canViewStatus bool
		var admissionDate sql.NullTime

		err := rows.Scan(&residentID, &tenantID, &residentAccount, &nickname, &canViewStatus, &status, &serviceLevel, &admissionDate, &familyTag)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		fmt.Printf("%-40s %-40s %-20s %-20s %-15v %-15s %-15s %-15s %-20s\n",
			getString(residentID), getString(tenantID), getString(residentAccount), getString(nickname),
			canViewStatus, getString(status), getString(serviceLevel), getDate(admissionDate), getString(familyTag))
		count++
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	if count == 0 {
		fmt.Println("未找到匹配的记录（nickname 或 resident_account 为 'done' 或 'smith'）")
	} else {
		fmt.Printf("\n共找到 %d 条记录\n", count)
	}
}

func getString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return "NULL"
}

func getDate(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format("2006-01-02")
	}
	return "NULL"
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseInt(s string, def int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return i
}

