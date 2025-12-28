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
	fmt.Println("删除 Family 角色")
	fmt.Println("==========================================")
	fmt.Printf("数据库: %s\n", dbName)
	fmt.Printf("主机: %s:%s\n", dbHost, dbPort)
	fmt.Printf("用户: %s\n", dbUser)
	fmt.Println("")

	systemTenantID := "00000000-0000-0000-0000-000000000001"

	// 1. 删除 Family 角色的所有权限记录
	fmt.Println("1. 删除 Family 角色的权限记录...")
	result, err := db.Exec(`
		DELETE FROM role_permissions 
		WHERE tenant_id = $1 
		  AND role_code = 'Family'
	`, systemTenantID)
	if err != nil {
		log.Fatalf("❌ 删除 Family 权限记录失败: %v", err)
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("   ✅ 删除了 %d 条权限记录\n", rowsAffected)

	// 2. 删除 Family 角色
	fmt.Println("2. 删除 Family 角色...")
	result, err = db.Exec(`
		DELETE FROM roles
		WHERE tenant_id = $1
		  AND role_code = 'Family'
	`, systemTenantID)
	if err != nil {
		log.Fatalf("❌ 删除 Family 角色失败: %v", err)
	}
	rowsAffected, _ = result.RowsAffected()
	fmt.Printf("   ✅ 删除了 %d 个角色\n", rowsAffected)

	// 3. 验证删除结果
	fmt.Println("")
	fmt.Println("验证删除结果...")
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM roles
		WHERE tenant_id = $1
		  AND role_code = 'Family'
	`, systemTenantID).Scan(&count)
	if err != nil {
		log.Fatalf("❌ 验证删除结果失败: %v", err)
	}

	if count == 0 {
		fmt.Println("✅ Family 角色已成功删除")
	} else {
		fmt.Printf("⚠️  警告: 仍有 %d 个 Family 角色存在\n", count)
	}

	fmt.Println("")
	fmt.Println("✅ 删除操作完成！")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

