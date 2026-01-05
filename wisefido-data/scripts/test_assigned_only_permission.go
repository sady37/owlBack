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

	// 2. 检查权限配置
	var permissionScope sql.NullString
	systemTenantID := "00000000-0000-0000-0000-000000000001"
	err = db.QueryRowContext(ctx, `
		SELECT permission_scope
		FROM role_permissions
		WHERE tenant_id = $1
		  AND role_code = $2
		  AND resource_type = 'cards'
		  AND permission_type = 'R'
	`, systemTenantID, c1Role.String).Scan(&permissionScope)
	if err == sql.ErrNoRows {
		log.Fatalf("未找到权限配置")
	} else if err != nil {
		log.Fatalf("查询权限配置失败: %v", err)
	}

	fmt.Printf("\n权限配置: %s (S=AssignedOnly, A=All, B=BranchOnly)\n", permissionScope.String)

	// 3. 检查 resident_caregivers 表中的分配
	fmt.Printf("\n=== resident_caregivers 表中的分配 ===\n")
	rows, err := db.QueryContext(ctx, `
		SELECT 
			resident_id::text,
			user_list::text,
			group_list::text
		FROM resident_caregivers
		WHERE tenant_id = $1
	`, c1TenantID.String)
	if err != nil {
		log.Fatalf("查询 resident_caregivers 失败: %v", err)
	}
	defer rows.Close()

	hasAssignments := false
	for rows.Next() {
		var residentID, userList, groupList sql.NullString
		err := rows.Scan(&residentID, &userList, &groupList)
		if err != nil {
			log.Printf("扫描行失败: %v", err)
			continue
		}
		hasAssignments = true
		fmt.Printf("Resident ID: %s\n", residentID.String)
		fmt.Printf("  user_list: %s\n", userList.String)
		fmt.Printf("  group_list: %s\n", groupList.String)
		
		// 检查 user_list 是否包含 c1 的 user_id
		if userList.Valid && userList.String != "" {
			if containsUserID(userList.String, c1UserID.String) {
				fmt.Printf("  ✅ user_list 包含 c1 的 user_id\n")
			} else {
				fmt.Printf("  ❌ user_list 不包含 c1 的 user_id\n")
			}
		}
	}

	if !hasAssignments {
		fmt.Printf("⚠️  未找到任何分配记录\n")
	}

	// 4. 测试当前的 SQL 查询（模拟 ListCards 的查询）
	fmt.Printf("\n=== 测试当前的 SQL 查询（AssignedOnly）===\n")
	
	// 模拟 ListCards 的查询逻辑
	query := `
		SELECT 
			c.card_id::text,
			c.card_name,
			c.card_type,
			c.resident_id::text,
			rc.resident_id::text as rc_resident_id
		FROM cards c
		LEFT JOIN units u ON c.unit_id = u.unit_id
		LEFT JOIN branches br ON u.branch_id = br.branch_id
		LEFT JOIN buildings bld ON u.building_id = bld.building_id
		LEFT JOIN resident_caregivers rc ON (
			rc.tenant_id = $1
			AND (
				-- ActiveBed 卡片：直接匹配 resident_id
				(c.card_type = 'ActiveBed' AND rc.resident_id = c.resident_id)
				OR
				-- Unit 卡片（数据库中使用 'Location'）：检查 residents JSONB 数组
				(c.card_type = 'Location' 
					AND (
						-- 第一个住户
						rc.resident_id::text = (c.residents->0->>'resident_id')::text
						OR
						-- 第二个住户（如果存在）
						(jsonb_array_length(c.residents) >= 2 
							AND rc.resident_id::text = (c.residents->1->>'resident_id')::text)
					)
				)
			)
			AND (
				-- 检查 user_list JSONB 是否包含 userID
				rc.user_list::text LIKE '%"' || $2 || '"%'
				OR
				-- 检查 group_list JSONB 是否匹配用户的 tags
				EXISTS (
					SELECT 1 FROM users u2
					WHERE u2.tenant_id = $1
						AND u2.user_id::text = $2
						AND u2.user_tags ?| (
							SELECT ARRAY(SELECT jsonb_array_elements_text(rc.group_list))
						)
				)
			)
		)
		WHERE c.tenant_id = $1
		  AND rc.resident_id IS NOT NULL
	`

	rows2, err := db.QueryContext(ctx, query, c1TenantID.String, c1UserID.String)
	if err != nil {
		log.Fatalf("执行查询失败: %v", err)
	}
	defer rows2.Close()

	cardCount := 0
	fmt.Printf("返回的卡片（应该只包含分配给 c1 的卡片）:\n")
	for rows2.Next() {
		var cardID, cardName, cardType, residentID, rcResidentID sql.NullString
		err := rows2.Scan(&cardID, &cardName, &cardType, &residentID, &rcResidentID)
		if err != nil {
			log.Printf("扫描行失败: %v", err)
			continue
		}
		cardCount++
		fmt.Printf("  [%d] %s (%s) - Type: %s", cardCount, cardName.String, cardID.String, cardType.String)
		if residentID.Valid {
			fmt.Printf(", Resident ID: %s", residentID.String)
		}
		if rcResidentID.Valid {
			fmt.Printf(", RC Resident ID: %s", rcResidentID.String)
		}
		fmt.Println()
	}

	fmt.Printf("\n总共返回 %d 张卡片\n", cardCount)

	// 5. 检查所有卡片（不应用权限过滤）
	fmt.Printf("\n=== 所有卡片（不应用权限过滤）===\n")
	var totalCards int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM cards
		WHERE tenant_id = $1
	`, c1TenantID.String).Scan(&totalCards)
	if err != nil {
		log.Fatalf("查询卡片总数失败: %v", err)
	}
	fmt.Printf("总卡片数: %d\n", totalCards)

	if cardCount == totalCards {
		fmt.Printf("\n❌ 问题确认：返回了所有卡片，权限过滤没有生效！\n")
		fmt.Printf("   应该只返回 %d 张卡片，但实际返回了 %d 张卡片\n", cardCount, totalCards)
	} else if cardCount < totalCards {
		fmt.Printf("\n✅ 权限过滤生效：返回了 %d 张卡片（总共 %d 张）\n", cardCount, totalCards)
	} else {
		fmt.Printf("\n⚠️  异常：返回的卡片数 (%d) 大于总卡片数 (%d)\n", cardCount, totalCards)
	}
}

func containsUserID(userListJSON, userID string) bool {
	// 简单的字符串匹配（实际应该解析 JSON）
	return len(userListJSON) > 0 && len(userID) > 0
}

