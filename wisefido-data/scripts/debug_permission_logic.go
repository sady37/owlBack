package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// 从环境变量获取数据库连接字符串
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/owlrd?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 系统租户 ID
	systemTenantID := "00000000-0000-0000-0000-000000000001"

	// 1. 获取 c1 用户信息
	var c1UserID, c1Role, c1TenantID sql.NullString
	err = db.QueryRowContext(ctx, `
		SELECT user_id::text, role, tenant_id::text
		FROM users
		WHERE user_account = 'c1' OR user_account LIKE 'c1%'
		LIMIT 1
	`).Scan(&c1UserID, &c1Role, &c1TenantID)
	if err == sql.ErrNoRows {
		log.Fatalf("未找到 c1 用户")
	} else if err != nil {
		log.Fatalf("查询 c1 用户失败: %v", err)
	}

	fmt.Printf("=== c1 用户信息 ===\n")
	fmt.Printf("User ID: %s\n", c1UserID.String)
	fmt.Printf("Role: %s\n", c1Role.String)
	fmt.Printf("Tenant ID: %s\n", c1TenantID.String)

	// 2. 检查权限配置（模拟 getResourcePermission 的逻辑）
	var permissionScope sql.NullString
	err = db.QueryRowContext(ctx, `
		SELECT permission_scope
		FROM role_permissions
		WHERE tenant_id = $1
		  AND role_code = $2
		  AND resource_type = $3
		  AND permission_type = $4
	`, systemTenantID, c1Role.String, "cards", "R").Scan(&permissionScope)
	if err == sql.ErrNoRows {
		fmt.Printf("\n⚠️  未找到权限配置，将使用默认值（无限制）\n")
		fmt.Printf("   这意味着 AssignedOnly = false, BranchOnly = false\n")
		fmt.Printf("   问题：如果权限配置不存在，应该返回空结果，而不是所有卡片！\n")
		return
	} else if err != nil {
		log.Fatalf("查询权限配置失败: %v", err)
	}

	fmt.Printf("\n权限配置: %s\n", permissionScope.String)
	fmt.Printf("  S = AssignedOnly\n")
	fmt.Printf("  A = All\n")
	fmt.Printf("  B = BranchOnly\n")

	// 3. 解析权限配置（模拟 getResourcePermission 的逻辑）
	var assignedOnly bool
	switch permissionScope.String {
	case "A":
		assignedOnly = false
		fmt.Printf("\n解析结果: AssignedOnly = false (可以看到所有卡片)\n")
	case "S":
		assignedOnly = true
		fmt.Printf("\n解析结果: AssignedOnly = true (只能看到分配的卡片)\n")
	case "B":
		assignedOnly = false
		fmt.Printf("\n解析结果: AssignedOnly = false, BranchOnly = true (只能看到自己院区的卡片)\n")
	default:
		assignedOnly = false
		fmt.Printf("\n解析结果: AssignedOnly = false (默认值，可以看到所有卡片)\n")
	}

	// 4. 检查 ListVitalFocusCards 的逻辑
	fmt.Printf("\n=== ListVitalFocusCards 逻辑检查 ===\n")
	if assignedOnly {
		fmt.Printf("✅ 应该设置 PermissionFilter.AssignedOnly = true\n")
		fmt.Printf("✅ 应该设置 PermissionFilter.UserIDForAssignment = %s\n", c1UserID.String)
	} else {
		fmt.Printf("⚠️  不会设置 PermissionFilter.AssignedOnly\n")
		fmt.Printf("⚠️  这意味着不会应用 AssignedOnly 过滤\n")
		fmt.Printf("⚠️  如果权限配置是 'S'，但这里显示 false，说明权限解析有问题！\n")
	}

	// 5. 如果权限配置是 'S'，但 assignedOnly 是 false，说明问题
	if permissionScope.String == "S" && !assignedOnly {
		fmt.Printf("\n❌ 发现问题：权限配置是 'S'，但 assignedOnly 解析为 false！\n")
		fmt.Printf("   这会导致权限过滤不生效，用户能看到所有卡片\n")
	}
}

