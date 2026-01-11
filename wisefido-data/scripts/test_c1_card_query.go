//go:build ignore
// +build ignore

package main

import (
	"context"
	"database/sql"
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

	// c1 的信息（从上一个脚本得到）
	tenantID := "bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c"
	userID := "c3b9bde2-4be4-4671-af81-f79694486932"
	userRole := "Caregiver"
	unitIDs := []string{"d8b82a88-30c3-4c62-a7c5-c4d88d495619"}
	assignedResidentIDs := []string{"b266ec84-c626-41ad-8bdd-3a72b4194fc0"}

	fmt.Printf("=== 测试 getCardIDsForStaff 查询逻辑 ===\n")
	fmt.Printf("Tenant ID: %s\n", tenantID)
	fmt.Printf("User ID: %s\n", userID)
	fmt.Printf("User Role: %s\n", userRole)
	fmt.Printf("Unit IDs: %v\n", unitIDs)
	fmt.Printf("Assigned Resident IDs: %v\n", assignedResidentIDs)

	// 构建查询（模拟 getCardIDsForStaff 的逻辑）
	var query strings.Builder
	var args []any
	argIdx := 1

	// SELECT 子句
	query.WriteString(`
		SELECT DISTINCT c.card_id::text, c.card_name, c.card_type, c.resident_id::text, c.unit_id::text
		FROM cards c
		LEFT JOIN units u ON c.unit_id = u.unit_id
		LEFT JOIN branches br ON u.branch_id = br.branch_id
		WHERE c.tenant_id = $` + fmt.Sprintf("%d", argIdx) + `
	`)
	args = append(args, tenantID)
	argIdx++

	var conditions []string

	// Nurse/Caregiver 分支
	if len(unitIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf(
			`(c.unit_id = ANY($%d::uuid[]) AND u.is_shared_unit = FALSE)`,
			argIdx,
		))
		args = append(args, pq.Array(unitIDs))
		argIdx++
	}

	if len(assignedResidentIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf(
			`(c.card_type = 'ActiveBed' AND c.resident_id = ANY($%d::uuid[]))`,
			argIdx,
		))
		args = append(args, pq.Array(assignedResidentIDs))
		argIdx++
	}

	// 组合条件：OR 连接
	if len(conditions) > 0 {
		query.WriteString(` AND (`)
		query.WriteString(strings.Join(conditions, " OR "))
		query.WriteString(`)`)
	}

	fmt.Printf("\n=== 生成的 SQL 查询 ===\n")
	fmt.Printf("%s\n", query.String())
	fmt.Printf("\n参数: %v\n", args)

	// 执行查询
	rows, err := db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		log.Fatalf("执行查询失败: %v", err)
	}
	defer rows.Close()

	var cardIDs []string
	fmt.Printf("\n=== 查询结果 ===\n")
	for rows.Next() {
		var cardID, cardName, cardType, residentID, unitID sql.NullString
		if err := rows.Scan(&cardID, &cardName, &cardType, &residentID, &unitID); err != nil {
			log.Printf("扫描行失败: %v", err)
			continue
		}
		cardIDs = append(cardIDs, cardID.String)
		fmt.Printf("Card ID: %s, Name: %s, Type: %s", cardID.String, cardName.String, cardType.String)
		if residentID.Valid {
			fmt.Printf(", Resident ID: %s", residentID.String)
		}
		if unitID.Valid {
			fmt.Printf(", Unit ID: %s", unitID.String)
		}
		fmt.Println()
	}

	fmt.Printf("\n总共返回 %d 张卡片\n", len(cardIDs))

	// 检查所有卡片
	var totalCards int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM cards
		WHERE tenant_id = $1
	`, tenantID).Scan(&totalCards)
	if err != nil {
		log.Fatalf("查询卡片总数失败: %v", err)
	}
	fmt.Printf("总卡片数: %d\n", totalCards)

	// 检查每张卡片是否符合条件
	fmt.Printf("\n=== 检查所有卡片是否符合条件 ===\n")
	rows2, err := db.QueryContext(ctx, `
		SELECT c.card_id::text, c.card_name, c.card_type, c.resident_id::text, c.unit_id::text, u.is_shared_unit
		FROM cards c
		LEFT JOIN units u ON c.unit_id = u.unit_id
		WHERE c.tenant_id = $1
	`, tenantID)
	if err != nil {
		log.Fatalf("查询所有卡片失败: %v", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var cardID, cardName, cardType, residentID, unitID sql.NullString
		var isSharedUnit sql.NullBool
		if err := rows2.Scan(&cardID, &cardName, &cardType, &residentID, &unitID, &isSharedUnit); err != nil {
			log.Printf("扫描行失败: %v", err)
			continue
		}

		matches := false
		var matchReason string

		// 条件 1: unit_id 匹配且 is_shared_unit = FALSE
		if unitID.Valid && unitID.String == unitIDs[0] && (!isSharedUnit.Valid || !isSharedUnit.Bool) {
			matches = true
			matchReason = "unit_id 匹配且非共享"
		}

		// 条件 2: card_type = 'ActiveBed' 且 resident_id 匹配
		if cardType.String == "ActiveBed" && residentID.Valid && residentID.String == assignedResidentIDs[0] {
			matches = true
			matchReason = "ActiveBed 且 resident_id 匹配"
		}

		status := "❌"
		if matches {
			status = "✅"
		}

		fmt.Printf("%s Card: %s (%s) - Type: %s", status, cardName.String, cardID.String, cardType.String)
		if residentID.Valid {
			fmt.Printf(", Resident: %s", residentID.String)
		}
		if unitID.Valid {
			fmt.Printf(", Unit: %s", unitID.String)
		}
		if isSharedUnit.Valid {
			fmt.Printf(", IsShared: %v", isSharedUnit.Bool)
		}
		if matches {
			fmt.Printf(" - %s", matchReason)
		}
		fmt.Println()
	}
}

