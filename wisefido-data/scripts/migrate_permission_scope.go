//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "owlrd")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping: %v", err)
	}

	fmt.Println("✅ 数据库连接成功")
	fmt.Println("")

	// 1. 检查当前表结构
	fmt.Println("📊 检查当前表结构...")
	hasAssignedOnly := checkColumnExists(db, "role_permissions", "assigned_only")
	hasBranchOnly := checkColumnExists(db, "role_permissions", "branch_only")
	hasPermissionScope := checkColumnExists(db, "role_permissions", "permission_scope")

	fmt.Printf("  - assigned_only 字段: %v\n", hasAssignedOnly)
	fmt.Printf("  - branch_only 字段: %v\n", hasBranchOnly)
	fmt.Printf("  - permission_scope 字段: %v\n", hasPermissionScope)
	fmt.Println("")

	if !hasAssignedOnly && !hasBranchOnly && hasPermissionScope {
		fmt.Println("✅ 表结构已经是新格式，无需迁移")
		return
	}

	if !hasAssignedOnly || !hasBranchOnly {
		log.Fatalf("❌ 表结构不完整: assigned_only=%v, branch_only=%v", hasAssignedOnly, hasBranchOnly)
	}

	// 2. 统计当前数据
	fmt.Println("📊 统计当前数据...")
	totalCount := countRecords(db, "role_permissions")
	assignedOnlyCount := countRecordsWithCondition(db, "role_permissions", "assigned_only = TRUE")
	branchOnlyCount := countRecordsWithCondition(db, "role_permissions", "branch_only = TRUE")
	bothCount := countRecordsWithCondition(db, "role_permissions", "assigned_only = TRUE AND branch_only = TRUE")
	noneCount := countRecordsWithCondition(db, "role_permissions", "assigned_only = FALSE AND branch_only = FALSE")

	fmt.Printf("  - 总记录数: %d\n", totalCount)
	fmt.Printf("  - assigned_only=TRUE: %d\n", assignedOnlyCount)
	fmt.Printf("  - branch_only=TRUE: %d\n", branchOnlyCount)
	fmt.Printf("  - assigned_only=TRUE AND branch_only=TRUE: %d\n", bothCount)
	fmt.Printf("  - assigned_only=FALSE AND branch_only=FALSE: %d\n", noneCount)
	fmt.Println("")

	// 3. 执行迁移
	fmt.Println("🔄 开始执行迁移...")

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// 3.1 添加 permission_scope 字段
	fmt.Println("  1. 添加 permission_scope 字段...")
	_, err = tx.Exec(`
		ALTER TABLE role_permissions 
		ADD COLUMN IF NOT EXISTS permission_scope VARCHAR(10)
	`)
	if err != nil {
		log.Fatalf("Failed to add permission_scope column: %v", err)
	}
	fmt.Println("     ✅ 字段已添加")

	// 3.2 转换数据
	fmt.Println("  2. 转换现有数据...")
	result, err := tx.Exec(`
		UPDATE role_permissions
		SET permission_scope = CASE
			WHEN branch_only = TRUE THEN 'B'
			WHEN assigned_only = TRUE THEN 'S'
			ELSE 'A'
		END
		WHERE permission_scope IS NULL
	`)
	if err != nil {
		log.Fatalf("Failed to update permission_scope: %v", err)
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("     ✅ 已更新 %d 条记录\n", rowsAffected)

	// 3.3 设置约束
	fmt.Println("  3. 设置约束...")
	_, err = tx.Exec(`
		ALTER TABLE role_permissions
		ALTER COLUMN permission_scope SET NOT NULL,
		ADD CONSTRAINT chk_permission_scope CHECK (permission_scope IN ('A', 'S', 'B'))
	`)
	if err != nil {
		log.Fatalf("Failed to set constraints: %v", err)
	}
	fmt.Println("     ✅ 约束已设置")

	// 3.4 删除旧字段
	fmt.Println("  4. 删除旧字段...")
	_, err = tx.Exec(`
		ALTER TABLE role_permissions
		DROP COLUMN IF EXISTS assigned_only,
		DROP COLUMN IF EXISTS branch_only
	`)
	if err != nil {
		log.Fatalf("Failed to drop old columns: %v", err)
	}
	fmt.Println("     ✅ 旧字段已删除")

	// 3.5 验证数据
	fmt.Println("  5. 验证数据完整性...")
	var nullCount, invalidCount int
	err = tx.QueryRow(`
		SELECT 
			COUNT(*) FILTER (WHERE permission_scope IS NULL) as null_count,
			COUNT(*) FILTER (WHERE permission_scope NOT IN ('A', 'S', 'B')) as invalid_count
		FROM role_permissions
	`).Scan(&nullCount, &invalidCount)
	if err != nil {
		log.Fatalf("Failed to validate data: %v", err)
	}

	if nullCount > 0 {
		log.Fatalf("❌ 发现 %d 条记录的 permission_scope 为 NULL", nullCount)
	}
	if invalidCount > 0 {
		log.Fatalf("❌ 发现 %d 条记录的 permission_scope 值无效", invalidCount)
	}
	fmt.Println("     ✅ 数据验证通过")

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Println("")
	fmt.Println("✅ 迁移完成！")

	// 4. 显示迁移后的统计
	fmt.Println("")
	fmt.Println("📊 迁移后的数据统计:")
	scopeStats := getScopeStats(db)
	for scope, count := range scopeStats {
		fmt.Printf("  - permission_scope='%s': %d\n", scope, count)
	}
}

func checkColumnExists(db *sql.DB, tableName, columnName string) bool {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = $1 AND column_name = $2
		)
	`
	err := db.QueryRow(query, tableName, columnName).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func countRecords(db *sql.DB, tableName string) int {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func countRecordsWithCondition(db *sql.DB, tableName, condition string) int {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", tableName, condition)
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func getScopeStats(db *sql.DB) map[string]int {
	stats := make(map[string]int)
	rows, err := db.Query(`
		SELECT permission_scope, COUNT(*) 
		FROM role_permissions 
		GROUP BY permission_scope 
		ORDER BY permission_scope
	`)
	if err != nil {
		return stats
	}
	defer rows.Close()

	for rows.Next() {
		var scope string
		var count int
		if err := rows.Scan(&scope, &count); err == nil {
			stats[scope] = count
		}
	}
	return stats
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

