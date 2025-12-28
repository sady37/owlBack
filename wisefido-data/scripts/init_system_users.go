//go:build ignore
// +build ignore

package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接参数
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "owlrd")

	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ 连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ 数据库连接测试失败: %v", err)
	}

	fmt.Println("==========================================")
	fmt.Println("初始化 System Tenant 和系统用户")
	fmt.Println("==========================================")
	fmt.Printf("数据库: %s\n", dbName)
	fmt.Printf("主机: %s:%s\n", dbHost, dbPort)
	fmt.Printf("用户: %s\n", dbUser)
	fmt.Println("")

	// 1. 创建 System tenant
	fmt.Println("1. 创建 System tenant...")
	systemTenantID := "00000000-0000-0000-0000-000000000001"
	_, err = db.Exec(`
		INSERT INTO tenants (tenant_id, tenant_name, tenant_type, domain, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id) DO UPDATE SET
			tenant_name = EXCLUDED.tenant_name,
			tenant_type = EXCLUDED.tenant_type,
			domain = EXCLUDED.domain,
			status = EXCLUDED.status
	`, systemTenantID, "System", "organization", "system.local", "active")
	if err != nil {
		log.Fatalf("❌ 创建 System tenant 失败: %v", err)
	}
	fmt.Println("   ✅ System tenant 创建成功")

	// 2. 创建 sysadmin 用户
	fmt.Println("2. 创建 sysadmin 用户...")
	sysadminAccountHash := sha256Hash(strings.ToLower("sysadmin"))
	sysadminPasswordHash := sha256Hash("ChangeMe123!")
	_, err = db.Exec(`
		INSERT INTO users (
			tenant_id, user_account, user_account_hash, password_hash, nickname, role, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id, user_account) DO UPDATE SET
			user_account_hash = EXCLUDED.user_account_hash,
			password_hash = EXCLUDED.password_hash,
			nickname = EXCLUDED.nickname,
			role = EXCLUDED.role,
			status = EXCLUDED.status
	`, systemTenantID, "sysadmin", sysadminAccountHash, sysadminPasswordHash, "SystemAdmin", "SystemAdmin", "active")
	if err != nil {
		log.Fatalf("❌ 创建 sysadmin 用户失败: %v", err)
	}
	fmt.Println("   ✅ sysadmin 用户创建成功")

	// 3. 创建 sysopr 用户
	fmt.Println("3. 创建 sysopr 用户...")
	sysoprAccountHash := sha256Hash(strings.ToLower("sysopr"))
	sysoprPasswordHash := sha256Hash("ChangeMe123!")
	_, err = db.Exec(`
		INSERT INTO users (
			tenant_id, user_account, user_account_hash, password_hash, nickname, role, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id, user_account) DO UPDATE SET
			user_account_hash = EXCLUDED.user_account_hash,
			password_hash = EXCLUDED.password_hash,
			nickname = EXCLUDED.nickname,
			role = EXCLUDED.role,
			status = EXCLUDED.status
	`, systemTenantID, "sysopr", sysoprAccountHash, sysoprPasswordHash, "SystemOperator", "SystemOperator", "active")
	if err != nil {
		log.Fatalf("❌ 创建 sysopr 用户失败: %v", err)
	}
	fmt.Println("   ✅ sysopr 用户创建成功")

	// 4. 验证创建结果
	fmt.Println("")
	fmt.Println("验证创建结果...")
	rows, err := db.Query(`
		SELECT user_account, nickname, role, status
		FROM users
		WHERE tenant_id = $1
		ORDER BY user_account
	`, systemTenantID)
	if err != nil {
		log.Fatalf("❌ 查询用户失败: %v", err)
	}
	defer rows.Close()

	fmt.Println("")
	fmt.Println("已创建的用户：")
	for rows.Next() {
		var account, nickname, role, status string
		if err := rows.Scan(&account, &nickname, &role, &status); err != nil {
			log.Fatalf("❌ 扫描用户数据失败: %v", err)
		}
		fmt.Printf("  - %s (%s, %s, %s)\n", account, nickname, role, status)
	}

	fmt.Println("")
	fmt.Println("✅ 初始化成功！")
	fmt.Println("")
	fmt.Println("登录信息：")
	fmt.Println("  - Tenant ID: 00000000-0000-0000-0000-000000000001")
	fmt.Println("  - 账号: sysadmin 或 sysopr")
	fmt.Println("  - 密码: ChangeMe123!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func sha256Hash(s string) []byte {
	hash := sha256.Sum256([]byte(s))
	return hash[:]
}
