//go:build ignore
// +build ignore

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

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

	// ========== 步骤 1: 模拟 validateUserSecurity ==========
	fmt.Printf("=== 步骤 1: validateUserSecurity ===\n")
	c1UserID := "c3b9bde2-4be4-4671-af81-f79694486932"
	
	var actualTenantID, actualUserRole, userStatus string
	err = db.QueryRowContext(ctx,
		`SELECT tenant_id::text, role, COALESCE(status, 'active')
		 FROM users
		 WHERE user_id = $1`,
		c1UserID,
	).Scan(&actualTenantID, &actualUserRole, &userStatus)
	if err != nil {
		log.Fatalf("查询用户失败: %v", err)
	}

	fmt.Printf("Validated Tenant ID: %s\n", actualTenantID)
	fmt.Printf("Validated User Role: %s\n", actualUserRole)
	fmt.Printf("User Status: %s\n", userStatus)

	// 查询 branch_ids
	rows, err := db.QueryContext(ctx,
		`SELECT branch_id::text
		 FROM user_branches
		 WHERE tenant_id = $1 AND user_id = $2`,
		actualTenantID, c1UserID,
	)
	if err != nil {
		log.Fatalf("查询 user_branches 失败: %v", err)
	}
	defer rows.Close()

	var validatedBranchIDs []string
	for rows.Next() {
		var branchID string
		if err := rows.Scan(&branchID); err != nil {
			log.Fatalf("扫描 branch_id 失败: %v", err)
		}
		validatedBranchIDs = append(validatedBranchIDs, branchID)
	}
	fmt.Printf("Validated Branch IDs: %v\n", validatedBranchIDs)

	// ========== 步骤 2: 模拟 getResourcePermission ==========
	fmt.Printf("\n=== 步骤 2: getResourcePermission ===\n")
	systemTenantID := "00000000-0000-0000-0000-000000000001"
	var permissionScope string
	err = db.QueryRowContext(ctx,
		`SELECT permission_scope
		 FROM role_permissions
		 WHERE tenant_id = $1
		   AND role_code = $2
		   AND resource_type = 'cards'
		   AND permission_type = 'R'`,
		systemTenantID, actualUserRole,
	).Scan(&permissionScope)
	if err != nil {
		log.Fatalf("查询权限配置失败: %v", err)
	}

	fmt.Printf("Permission Scope: %s (S=AssignedOnly, A=All, B=BranchOnly)\n", permissionScope)
	
	assignedOnly := permissionScope == "S"
	branchOnly := permissionScope == "B"
	fmt.Printf("AssignedOnly: %v, BranchOnly: %v\n", assignedOnly, branchOnly)

	// ========== 步骤 3: 模拟 getAssignedResidentIDs ==========
	fmt.Printf("\n=== 步骤 3: getAssignedResidentIDs ===\n")
	var assignedResidentIDs []string
	
	if assignedOnly {
		// 获取 user_tags
		var userTagsJSON json.RawMessage
		err = db.QueryRowContext(ctx,
			`SELECT COALESCE(user_tags, '[]'::jsonb)
			 FROM users
			 WHERE tenant_id = $1 AND user_id = $2`,
			actualTenantID, c1UserID,
		).Scan(&userTagsJSON)
		if err != nil {
			log.Fatalf("查询 user_tags 失败: %v", err)
		}

		var userTagsArray []string
		if len(userTagsJSON) > 0 {
			if err := json.Unmarshal(userTagsJSON, &userTagsArray); err != nil {
				fmt.Printf("⚠️  解析 user_tags 失败: %v，使用空数组\n", err)
				userTagsArray = []string{}
			}
		}
		fmt.Printf("User Tags: %v\n", userTagsArray)

		// 查询 resident_caregivers
		userIDJSON, _ := json.Marshal([]string{c1UserID})
		query := `
			SELECT DISTINCT resident_id::text
			FROM resident_caregivers
			WHERE tenant_id = $1
			  AND (
				user_list @> $2::jsonb
				OR
				(
					$3::text[] IS NOT NULL
					AND array_length($3::text[], 1) > 0
					AND group_list IS NOT NULL
					AND group_list ?| $3::text[]
				)
			  )
		`

		rows2, err := db.QueryContext(ctx, query, actualTenantID, userIDJSON, pq.Array(userTagsArray))
		if err != nil {
			log.Fatalf("查询 assigned residents 失败: %v", err)
		}
		defer rows2.Close()

		for rows2.Next() {
			var residentID string
			if err := rows2.Scan(&residentID); err != nil {
				log.Fatalf("扫描 resident_id 失败: %v", err)
			}
			assignedResidentIDs = append(assignedResidentIDs, residentID)
		}
	}

	fmt.Printf("Assigned Resident IDs: %v (count: %d)\n", assignedResidentIDs, len(assignedResidentIDs))

	// ========== 步骤 4: 模拟 getUnitsByResidentIDs ==========
	fmt.Printf("\n=== 步骤 4: getUnitsByResidentIDs ===\n")
	var unitIDs []string
	if len(assignedResidentIDs) > 0 {
		query := `
			SELECT DISTINCT unit_id::text
			FROM residents
			WHERE tenant_id = $1
			  AND resident_id = ANY($2::uuid[])
			  AND unit_id IS NOT NULL
		`

		rows3, err := db.QueryContext(ctx, query, actualTenantID, pq.Array(assignedResidentIDs))
		if err != nil {
			log.Fatalf("查询 units 失败: %v", err)
		}
		defer rows3.Close()

		for rows3.Next() {
			var unitID string
			if err := rows3.Scan(&unitID); err != nil {
				log.Fatalf("扫描 unit_id 失败: %v", err)
			}
			unitIDs = append(unitIDs, unitID)
		}
	}

	fmt.Printf("Unit IDs: %v (count: %d)\n", unitIDs, len(unitIDs))

	// ========== 步骤 5: 模拟 getCardIDsForStaff ==========
	fmt.Printf("\n=== 步骤 5: getCardIDsForStaff ===\n")
	fmt.Printf("User Role: %s\n", actualUserRole)
	fmt.Printf("Unit IDs: %v\n", unitIDs)
	fmt.Printf("Assigned Resident IDs: %v\n", assignedResidentIDs)
	fmt.Printf("Branch IDs: %v\n", validatedBranchIDs)

	var query strings.Builder
	var args []any
	argIdx := 1

	// SELECT 子句
	query.WriteString(`
		SELECT DISTINCT c.card_id::text, c.card_name, c.card_type, c.resident_id::text, c.unit_id::text, u.is_shared_unit
		FROM cards c
		LEFT JOIN units u ON c.unit_id = u.unit_id
		LEFT JOIN branches br ON u.branch_id = br.branch_id
		WHERE c.tenant_id = $` + fmt.Sprintf("%d", argIdx) + `
	`)
	args = append(args, actualTenantID)
	argIdx++

	var conditions []string

	// 根据角色判断
	switch actualUserRole {
	case "Admin":
		fmt.Printf("→ 进入 Admin 分支（All 权限）\n")
		// conditions 保持为空

	case "Manager":
		fmt.Printf("→ 进入 Manager 分支（BranchOnly 权限）\n")
		if len(validatedBranchIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf(`u.branch_id = ANY($%d::uuid[])`, argIdx))
			args = append(args, pq.Array(validatedBranchIDs))
			argIdx++
		} else {
			fmt.Printf("⚠️  Manager 没有绑定院区，返回空列表\n")
			fmt.Printf("\n=== 最终结果 ===\n")
			fmt.Printf("Card IDs: [] (0 张卡片)\n")
			return
		}

	case "Nurse", "Caregiver":
		fmt.Printf("→ 进入 Nurse/Caregiver 分支（AssignedOnly 权限）\n")
		if len(unitIDs) > 0 {
			condition := fmt.Sprintf(
				`(c.unit_id = ANY($%d::uuid[]) AND u.is_shared_unit = FALSE)`,
				argIdx,
			)
			conditions = append(conditions, condition)
			args = append(args, pq.Array(unitIDs))
			argIdx++
			fmt.Printf("  添加条件 1: unit_id 匹配且非共享 (unitIDs: %v)\n", unitIDs)
		}

		if len(assignedResidentIDs) > 0 {
			condition := fmt.Sprintf(
				`(c.card_type = 'ActiveBed' AND c.resident_id = ANY($%d::uuid[]))`,
				argIdx,
			)
			conditions = append(conditions, condition)
			args = append(args, pq.Array(assignedResidentIDs))
			argIdx++
			fmt.Printf("  添加条件 2: ActiveBed 且 resident_id 匹配 (assignedResidentIDs: %v)\n", assignedResidentIDs)
		}

		if len(conditions) == 0 {
			fmt.Printf("⚠️  没有分配条件，返回空列表\n")
			fmt.Printf("\n=== 最终结果 ===\n")
			fmt.Printf("Card IDs: [] (0 张卡片)\n")
			return
		}

		fmt.Printf("  总共 %d 个条件，使用 OR 连接\n", len(conditions))

	default:
		fmt.Printf("⚠️  未知角色: %s，返回空列表\n", actualUserRole)
		fmt.Printf("\n=== 最终结果 ===\n")
		fmt.Printf("Card IDs: [] (0 张卡片)\n")
		return
	}

	// 组合条件
	if len(conditions) > 0 {
		query.WriteString(` AND (`)
		query.WriteString(strings.Join(conditions, " OR "))
		query.WriteString(`)`)
	}

	fmt.Printf("\n生成的 SQL 查询:\n%s\n", query.String())
	fmt.Printf("参数: %v\n", args)

	// 执行查询
	rows4, err := db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		log.Fatalf("执行查询失败: %v", err)
	}
	defer rows4.Close()

	var cardIDs []string
	fmt.Printf("\n查询结果:\n")
	for rows4.Next() {
		var cardID, cardName, cardType, residentID, unitID sql.NullString
		var isSharedUnit sql.NullBool
		if err := rows4.Scan(&cardID, &cardName, &cardType, &residentID, &unitID, &isSharedUnit); err != nil {
			log.Printf("扫描行失败: %v", err)
			continue
		}
		cardIDs = append(cardIDs, cardID.String)
		fmt.Printf("  [%d] Card ID: %s, Name: %s, Type: %s", len(cardIDs), cardID.String, cardName.String, cardType.String)
		if residentID.Valid {
			fmt.Printf(", Resident: %s", residentID.String)
		}
		if unitID.Valid {
			fmt.Printf(", Unit: %s", unitID.String)
		}
		if isSharedUnit.Valid {
			fmt.Printf(", IsShared: %v", isSharedUnit.Bool)
		}
		fmt.Println()
	}

	// ========== 最终结果 ==========
	fmt.Printf("\n=== 最终结果 ===\n")
	fmt.Printf("Card IDs: %v (共 %d 张卡片)\n", cardIDs, len(cardIDs))

	// 检查所有卡片
	var totalCards int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM cards
		WHERE tenant_id = $1
	`, actualTenantID).Scan(&totalCards)
	if err != nil {
		log.Fatalf("查询卡片总数失败: %v", err)
	}
	fmt.Printf("总卡片数: %d\n", totalCards)

	if len(cardIDs) == totalCards {
		fmt.Printf("\n❌ 问题确认：返回了所有卡片！\n")
		fmt.Printf("   应该只返回 %d 张卡片，但实际返回了 %d 张卡片\n", len(cardIDs), totalCards)
	} else if len(cardIDs) < totalCards {
		fmt.Printf("\n✅ 权限过滤生效：返回了 %d 张卡片（总共 %d 张）\n", len(cardIDs), totalCards)
	}
}

