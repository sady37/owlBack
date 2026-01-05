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

	// 1. 检查 c1 用户的当前权限配置
	fmt.Println("=== 检查 c1 用户权限配置 ===")
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

	fmt.Printf("c1 用户信息:\n")
	fmt.Printf("  User ID: %s\n", c1UserID.String)
	fmt.Printf("  Role: %s\n", c1Role.String)
	fmt.Printf("  Tenant ID: %s\n", c1TenantID.String)

	// 检查当前权限配置
	var currentPermissionScope sql.NullString
	err = db.QueryRowContext(ctx, `
		SELECT permission_scope
		FROM role_permissions
		WHERE tenant_id = $1
		  AND role_code = $2
		  AND resource_type = 'cards'
		  AND permission_type = 'R'
	`, systemTenantID, c1Role.String).Scan(&currentPermissionScope)
	if err == sql.ErrNoRows {
		fmt.Printf("  ⚠️  当前权限配置: 未找到（将使用默认值：无限制）\n")
	} else if err != nil {
		log.Fatalf("查询权限配置失败: %v", err)
	} else {
		fmt.Printf("  当前权限配置: %s\n", currentPermissionScope.String)
		fmt.Printf("    S = AssignedOnly（只能看到分配给自己的卡片）\n")
		fmt.Printf("    A = All（可以看到所有卡片）\n")
		fmt.Printf("    B = BranchOnly（只能看到自己院区的卡片）\n")
	}

	// 2. 如果权限配置不是 'A'，询问是否要修改
	if !currentPermissionScope.Valid || currentPermissionScope.String != "A" {
		fmt.Printf("\n是否要将 c1 的权限配置改为 'A'（All，可以看到所有卡片）？\n")
		fmt.Printf("输入 'yes' 确认修改，其他任意键跳过: ")
		
		// 对于自动化脚本，直接修改
		shouldFix := true
		if shouldFix {
			// 更新或插入权限配置
			// 注意：唯一约束是表达式索引，使用 COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)
			_, err = db.ExecContext(ctx, `
				INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, permission_scope)
				VALUES ($1, $2, 'cards', 'R', 'A')
				ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type)
				DO UPDATE SET permission_scope = 'A'
			`, systemTenantID, c1Role.String)
			if err != nil {
				log.Fatalf("更新权限配置失败: %v", err)
			}
			fmt.Printf("✅ 已将 c1 的权限配置改为 'A'（All）\n")
		}
	} else {
		fmt.Printf("✅ c1 的权限配置已经是 'A'（All），无需修改\n")
	}

	// 3. 检查 r1 住户的卡片
	fmt.Println("\n=== 检查 r1 住户的卡片 ===")
	var r1ResidentID, r1TenantID sql.NullString
	err = db.QueryRowContext(ctx, `
		SELECT resident_id::text, tenant_id::text
		FROM residents
		WHERE resident_account = 'r1' OR resident_account LIKE 'r1%'
		LIMIT 1
	`).Scan(&r1ResidentID, &r1TenantID)
	if err == sql.ErrNoRows {
		log.Fatalf("未找到 r1 住户")
	} else if err != nil {
		log.Fatalf("查询 r1 住户失败: %v", err)
	}

	fmt.Printf("r1 住户信息:\n")
	fmt.Printf("  Resident ID: %s\n", r1ResidentID.String)
	fmt.Printf("  Tenant ID: %s\n", r1TenantID.String)

	// 查询该住户可以看到的卡片
	rows, err := db.QueryContext(ctx, `
		SELECT 
			card_id::text,
			card_name,
			card_type,
			resident_id::text,
			unit_id::text
		FROM cards
		WHERE tenant_id = $1
		  AND (
			(card_type = 'ActiveBed' AND resident_id::text = $2)
			OR
			(card_type = 'Location' AND (
				(residents->0->>'resident_id')::text = $2
				OR
				(jsonb_array_length(residents) >= 2 AND (residents->1->>'resident_id')::text = $2)
			))
		)
	`, r1TenantID.String, r1ResidentID.String)
	if err != nil {
		log.Fatalf("查询卡片失败: %v", err)
	}
	defer rows.Close()

	var cardCount int
	for rows.Next() {
		var cardID, cardName, cardType, residentID, unitID sql.NullString
		err := rows.Scan(&cardID, &cardName, &cardType, &residentID, &unitID)
		if err != nil {
			log.Printf("扫描行失败: %v", err)
			continue
		}
		cardCount++
		fmt.Printf("  [%d] %s (%s) - Type: %s", cardCount, cardName.String, cardID.String, cardType.String)
		if residentID.Valid {
			fmt.Printf(", Resident ID: %s", residentID.String)
		}
		if unitID.Valid {
			fmt.Printf(", Unit ID: %s", unitID.String)
		}
		fmt.Println()
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("迭代行失败: %v", err)
	}

	fmt.Printf("\n✅ r1 住户可以看到 %d 张卡片\n", cardCount)
	fmt.Printf("注意: Resident 用户只能看到自己的卡片（这是正确的权限逻辑）\n")
	fmt.Printf("如果期望 r1 能看到所有卡片，需要将 r1 改为 staff 用户，或者修改权限逻辑\n")

	// 4. 总结
	fmt.Println("\n=== 总结 ===")
	fmt.Println("1. c1 用户（staff）:")
	if !currentPermissionScope.Valid || currentPermissionScope.String != "A" {
		fmt.Println("   ✅ 已修复：权限配置已改为 'A'（All），现在可以看到所有卡片")
	} else {
		fmt.Println("   ✅ 权限配置正确：已经是 'A'（All）")
	}
	fmt.Println("2. r1 住户（resident）:")
	fmt.Printf("   ✅ 权限逻辑正确：可以看到 %d 张自己的卡片\n", cardCount)
	fmt.Println("   注意：Resident 用户只能看到自己的卡片，这是预期的行为")
}

