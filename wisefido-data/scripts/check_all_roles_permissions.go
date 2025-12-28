//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"

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

	systemTenantID := "00000000-0000-0000-0000-000000000001"

	// 获取所有角色列表
	roles := []string{
		"SystemAdmin",
		"SystemOperator",
		"Admin",
		"Manager",
		"IT",
		"Nurse",
		"Caregiver",
		"Resident",
		"Individual",
	}

	fmt.Fprintf(os.Stdout, "=== 检查所有角色的权限记录 ===\n\n")

	// 检查每个角色的权限
	for _, roleCode := range roles {
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*)
			FROM role_permissions
			WHERE tenant_id = $1
			  AND role_code = $2
			  AND resource_type != 'service_levels'
		`, systemTenantID, roleCode).Scan(&count)
		if err != nil {
			log.Fatalf("Failed to query role %s: %v", roleCode, err)
		}

		status := "✅"
		if count == 0 {
			status = "❌"
		}

		fmt.Fprintf(os.Stdout, "%s %s: %d 条权限记录\n", status, roleCode, count)

		// 如果权限数量为0，列出应该有的资源类型
		if count == 0 {
			fmt.Fprintf(os.Stdout, "   ⚠️  该角色没有权限记录，需要从 03_role_permissions.sql 插入\n")
		}
	}

	fmt.Fprintf(os.Stdout, "\n=== 详细权限统计 ===\n\n")

	// 按角色分组统计
	type RoleStat struct {
		RoleCode      string
		TotalCount    int
		ResourceTypes map[string]int
	}

	stats := make(map[string]*RoleStat)

	for _, roleCode := range roles {
		rows, err := db.Query(`
			SELECT resource_type, COUNT(*) as perm_count
			FROM role_permissions
			WHERE tenant_id = $1
			  AND role_code = $2
			  AND resource_type != 'service_levels'
			GROUP BY resource_type
			ORDER BY resource_type
		`, systemTenantID, roleCode)
		if err != nil {
			log.Fatalf("Failed to query role %s: %v", roleCode, err)
		}

		stat := &RoleStat{
			RoleCode:      roleCode,
			ResourceTypes: make(map[string]int),
		}

		for rows.Next() {
			var resourceType string
			var permCount int
			if err := rows.Scan(&resourceType, &permCount); err != nil {
				log.Fatalf("Failed to scan: %v", err)
			}
			stat.ResourceTypes[resourceType] = permCount
			stat.TotalCount += permCount
		}
		rows.Close()

		stats[roleCode] = stat
	}

	// 按角色代码排序输出
	sort.Strings(roles)
	for _, roleCode := range roles {
		stat := stats[roleCode]
		if stat == nil || stat.TotalCount == 0 {
			continue
		}

		fmt.Fprintf(os.Stdout, "%s (%d 条权限):\n", roleCode, stat.TotalCount)
		
		// 按资源类型排序
		resourceTypes := make([]string, 0, len(stat.ResourceTypes))
		for rt := range stat.ResourceTypes {
			resourceTypes = append(resourceTypes, rt)
		}
		sort.Strings(resourceTypes)

		for _, rt := range resourceTypes {
			count := stat.ResourceTypes[rt]
			fmt.Fprintf(os.Stdout, "  - %s: %d 条\n", rt, count)
		}
		fmt.Fprintf(os.Stdout, "\n")
	}

	// 检查是否有缺失的权限
	fmt.Fprintf(os.Stdout, "=== 检查缺失的权限 ===\n\n")

	// SystemAdmin 应该有的资源类型
	systemAdminResources := []string{"tenants", "roles", "role_permissions", "users", "branches", "device_store"}
	systemAdminStat := stats["SystemAdmin"]
	if systemAdminStat != nil {
		for _, rt := range systemAdminResources {
			if systemAdminStat.ResourceTypes[rt] == 0 {
				fmt.Fprintf(os.Stdout, "❌ SystemAdmin 缺少 %s 权限\n", rt)
			}
		}
	}

	// SystemOperator 应该有的资源类型
	systemOperatorResources := []string{"tenants", "branches", "device_store", "alarm_cloud"}
	systemOperatorStat := stats["SystemOperator"]
	if systemOperatorStat != nil {
		for _, rt := range systemOperatorResources {
			if systemOperatorStat.ResourceTypes[rt] == 0 {
				fmt.Fprintf(os.Stdout, "❌ SystemOperator 缺少 %s 权限\n", rt)
			}
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
