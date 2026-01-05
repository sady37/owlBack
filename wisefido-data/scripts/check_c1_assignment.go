package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/lib/pq"
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
	var c1UserTagsJSON sql.NullString
	err = db.QueryRowContext(ctx, `
		SELECT user_id::text, role, tenant_id::text, COALESCE(user_tags::text, '[]')
		FROM users
		WHERE user_account = 'c1' OR user_account LIKE 'c1%'
		LIMIT 1
	`).Scan(&c1UserID, &c1Role, &c1TenantID, &c1UserTagsJSON)
	if err == sql.ErrNoRows {
		log.Fatalf("未找到 c1 用户")
	} else if err != nil {
		log.Fatalf("查询 c1 用户失败: %v", err)
	}

	fmt.Printf("=== c1 用户信息 ===\n")
	fmt.Printf("User ID: %s\n", c1UserID.String)
	fmt.Printf("Role: %s\n", c1Role.String)
	fmt.Printf("Tenant ID: %s\n", c1TenantID.String)
	fmt.Printf("User Tags: %s\n", c1UserTagsJSON.String)

	// 2. 解析 user_tags
	var userTagsArray []string
	if c1UserTagsJSON.Valid && c1UserTagsJSON.String != "" && c1UserTagsJSON.String != "[]" {
		if err := json.Unmarshal([]byte(c1UserTagsJSON.String), &userTagsArray); err != nil {
			fmt.Printf("⚠️  解析 user_tags 失败: %v\n", err)
			userTagsArray = []string{}
		}
	}
	fmt.Printf("Parsed User Tags: %v\n", userTagsArray)

	// 3. 模拟 getAssignedResidentIDs 查询
	fmt.Printf("\n=== 模拟 getAssignedResidentIDs 查询 ===\n")
	
	// 构建 userID 的 JSONB 值（用于 @> 操作符）
	userIDJSON, _ := json.Marshal([]string{c1UserID.String})
	fmt.Printf("UserID JSON for @> operator: %s\n", string(userIDJSON))

	query := `
		SELECT DISTINCT resident_id::text
		FROM resident_caregivers
		WHERE tenant_id = $1
		  AND (
			-- 条件 A: user_list JSONB 数组包含 userID
			user_list @> $2::jsonb
			OR
			-- 条件 B: group_list JSONB 数组与 user_tags 有交集（如果 user_tags 不为空）
			(
				$3::text[] IS NOT NULL
				AND array_length($3::text[], 1) > 0
				AND group_list IS NOT NULL
				AND group_list ?| $3::text[]
			)
		  )
	`

	rows, err := db.QueryContext(ctx, query, c1TenantID.String, userIDJSON, pq.Array(userTagsArray))
	if err != nil {
		log.Fatalf("执行查询失败: %v", err)
	}
	defer rows.Close()

	var assignedResidentIDs []string
	for rows.Next() {
		var residentID string
		if err := rows.Scan(&residentID); err != nil {
			log.Printf("扫描行失败: %v", err)
			continue
		}
		assignedResidentIDs = append(assignedResidentIDs, residentID)
	}

	fmt.Printf("assignedResidentIDs: %v (count: %d)\n", assignedResidentIDs, len(assignedResidentIDs))

	// 4. 模拟 getUnitsByResidentIDs 查询
	fmt.Printf("\n=== 模拟 getUnitsByResidentIDs 查询 ===\n")
	var unitIDs []string
	if len(assignedResidentIDs) > 0 {
		query2 := `
			SELECT DISTINCT unit_id::text
			FROM residents
			WHERE tenant_id = $1
			  AND resident_id = ANY($2::uuid[])
			  AND unit_id IS NOT NULL
		`

		rows2, err := db.QueryContext(ctx, query2, c1TenantID.String, pq.Array(assignedResidentIDs))
		if err != nil {
			log.Fatalf("执行查询失败: %v", err)
		}
		defer rows2.Close()

		for rows2.Next() {
			var unitID string
			if err := rows2.Scan(&unitID); err != nil {
				log.Printf("扫描行失败: %v", err)
				continue
			}
			unitIDs = append(unitIDs, unitID)
		}
	}

	fmt.Printf("unitIDs: %v (count: %d)\n", unitIDs, len(unitIDs))

	// 5. 检查 resident_caregivers 表中的所有记录
	fmt.Printf("\n=== resident_caregivers 表中的所有记录 ===\n")
	rows3, err := db.QueryContext(ctx, `
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
	defer rows3.Close()

	hasRecords := false
	for rows3.Next() {
		hasRecords = true
		var residentID, userList, groupList sql.NullString
		err := rows3.Scan(&residentID, &userList, &groupList)
		if err != nil {
			log.Printf("扫描行失败: %v", err)
			continue
		}
		fmt.Printf("Resident ID: %s\n", residentID.String)
		fmt.Printf("  user_list: %s\n", userList.String)
		fmt.Printf("  group_list: %s\n", groupList.String)
		
		// 检查 user_list 是否包含 c1 的 user_id
		if userList.Valid && userList.String != "" {
			var userListArray []string
			if err := json.Unmarshal([]byte(userList.String), &userListArray); err == nil {
				contains := false
				for _, uid := range userListArray {
					if uid == c1UserID.String {
						contains = true
						break
					}
				}
				if contains {
					fmt.Printf("  ✅ user_list 包含 c1 的 user_id\n")
				} else {
					fmt.Printf("  ❌ user_list 不包含 c1 的 user_id\n")
				}
			}
		}
	}

	if !hasRecords {
		fmt.Printf("⚠️  resident_caregivers 表中没有任何记录\n")
	}

	// 6. 总结
	fmt.Printf("\n=== 总结 ===\n")
	if len(assignedResidentIDs) == 0 && len(unitIDs) == 0 {
		fmt.Printf("✅ assignedResidentIDs 和 unitIDs 都为空\n")
		fmt.Printf("   按照当前逻辑，应该返回空列表（0 张卡片）\n")
		fmt.Printf("   如果仍然返回 4 张卡片，说明查询逻辑有问题\n")
	} else {
		fmt.Printf("⚠️  assignedResidentIDs 或 unitIDs 不为空\n")
		fmt.Printf("   assignedResidentIDs count: %d\n", len(assignedResidentIDs))
		fmt.Printf("   unitIDs count: %d\n", len(unitIDs))
		fmt.Printf("   应该根据这些 ID 过滤卡片\n")
	}
}

