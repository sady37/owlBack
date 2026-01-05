package main

import (
	"context"
	"database/sql"
	"encoding/json"
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

	// 1. 检查 cards 表的数据
	fmt.Println("=== 检查 cards 表数据 ===")
	rows, err := db.QueryContext(ctx, `
		SELECT 
			card_id,
			card_name,
			card_type,
			tenant_id::text,
			bed_id::text,
			unit_id::text,
			resident_id::text,
			residents::text,
			devices::text
		FROM cards
		ORDER BY card_name
		LIMIT 20
	`)
	if err != nil {
		log.Fatalf("Failed to query cards: %v", err)
	}
	defer rows.Close()

	hasIssues := false
	for rows.Next() {
		var cardID, cardName, cardType, tenantID, bedID, unitID, residentID, residentsJSON, devicesJSON sql.NullString
		err := rows.Scan(&cardID, &cardName, &cardType, &tenantID, &bedID, &unitID, &residentID, &residentsJSON, &devicesJSON)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		fmt.Printf("\n--- Card: %s (%s) ---\n", cardName.String, cardID.String)
		fmt.Printf("  Type: %s\n", cardType.String)
		fmt.Printf("  Tenant: %s\n", tenantID.String)
		if bedID.Valid {
			fmt.Printf("  Bed ID: %s\n", bedID.String)
		}
		if unitID.Valid {
			fmt.Printf("  Unit ID: %s\n", unitID.String)
		}
		if residentID.Valid {
			fmt.Printf("  Resident ID: %s\n", residentID.String)
		}

		// 检查 residents JSONB 格式
		if residentsJSON.Valid {
			fmt.Printf("  Residents JSON: %s\n", residentsJSON.String)
			
			// 尝试解析为对象数组
			var residentsArray []map[string]interface{}
			err := json.Unmarshal([]byte(residentsJSON.String), &residentsArray)
			if err != nil {
				// 尝试解析为字符串数组
				var residentsStringArray []string
				err2 := json.Unmarshal([]byte(residentsJSON.String), &residentsStringArray)
				if err2 != nil {
					fmt.Printf("  ❌ ERROR: Residents JSON 格式错误（既不是对象数组也不是字符串数组）\n")
					hasIssues = true
				} else {
					fmt.Printf("  ⚠️  WARNING: Residents JSON 是字符串数组，应该是对象数组\n")
					fmt.Printf("     当前格式: %v\n", residentsStringArray)
					fmt.Printf("     期望格式: [{\"resident_id\": \"%s\"}, ...]\n", residentsStringArray[0])
					hasIssues = true
				}
			} else {
				fmt.Printf("  ✅ Residents JSON 格式正确（对象数组）\n")
				for i, r := range residentsArray {
					if residentID, ok := r["resident_id"].(string); ok {
						fmt.Printf("      [%d] resident_id: %s\n", i, residentID)
					} else {
						fmt.Printf("      [%d] ⚠️  WARNING: 缺少 resident_id 字段\n", i)
						hasIssues = true
					}
				}
			}
		} else {
			fmt.Printf("  ⚠️  WARNING: Residents JSON 为空\n")
		}

		// 检查 devices JSONB 格式
		if devicesJSON.Valid {
			fmt.Printf("  Devices JSON: %s\n", devicesJSON.String)
		} else {
			fmt.Printf("  Devices JSON: []\n")
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	// 2. 检查权限过滤可能失败的原因
	fmt.Println("\n=== 权限过滤诊断 ===")
	
	// 检查 c1 用户（假设是 staff）
	fmt.Println("\n--- 检查 c1 用户（staff）---")
	var c1UserID, c1Role, c1TenantID sql.NullString
	err = db.QueryRowContext(ctx, `
		SELECT user_id::text, role, tenant_id::text
		FROM users
		WHERE user_account = 'c1' OR user_account LIKE 'c1%'
		LIMIT 1
	`).Scan(&c1UserID, &c1Role, &c1TenantID)
	if err == sql.ErrNoRows {
		fmt.Println("  ⚠️  WARNING: 未找到 c1 用户")
	} else if err != nil {
		log.Printf("  ❌ ERROR: 查询 c1 用户失败: %v", err)
	} else {
		fmt.Printf("  User ID: %s\n", c1UserID.String)
		fmt.Printf("  Role: %s\n", c1Role.String)
		fmt.Printf("  Tenant ID: %s\n", c1TenantID.String)
		
		// 检查权限配置
		var permissionScope sql.NullString
		err = db.QueryRowContext(ctx, `
			SELECT permission_scope
			FROM role_permissions
			WHERE tenant_id = '00000000-0000-0000-0000-000000000001'
			  AND role_code = $1
			  AND resource_type = 'cards'
			  AND permission_type = 'R'
		`, c1Role.String).Scan(&permissionScope)
		if err == sql.ErrNoRows {
			fmt.Printf("  ⚠️  WARNING: 未找到权限配置（role=%s），将使用默认值（无限制）\n", c1Role.String)
		} else if err != nil {
			log.Printf("  ❌ ERROR: 查询权限配置失败: %v", err)
		} else {
			fmt.Printf("  Permission Scope: %s\n", permissionScope.String)
			if permissionScope.String == "A" {
				fmt.Printf("  ✅ 权限配置正确（ALL），应该能看到所有卡片\n")
			}
		}
	}

	// 检查 r1 住户
	fmt.Println("\n--- 检查 r1 住户（resident）---")
	var r1ResidentID, r1TenantID sql.NullString
	err = db.QueryRowContext(ctx, `
		SELECT resident_id::text, tenant_id::text
		FROM residents
		WHERE resident_account = 'r1' OR resident_account LIKE 'r1%'
		LIMIT 1
	`).Scan(&r1ResidentID, &r1TenantID)
	if err == sql.ErrNoRows {
		fmt.Println("  ⚠️  WARNING: 未找到 r1 住户")
	} else if err != nil {
		log.Printf("  ❌ ERROR: 查询 r1 住户失败: %v", err)
	} else {
		fmt.Printf("  Resident ID: %s\n", r1ResidentID.String)
		fmt.Printf("  Tenant ID: %s\n", r1TenantID.String)
		
		// 检查该住户的卡片
		var cardCount int
		err = db.QueryRowContext(ctx, `
			SELECT COUNT(*)
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
		`, r1TenantID.String, r1ResidentID.String).Scan(&cardCount)
		if err != nil {
			log.Printf("  ❌ ERROR: 查询卡片失败: %v", err)
		} else {
			fmt.Printf("  ✅ 找到 %d 张卡片（通过权限过滤）\n", cardCount)
			if cardCount == 0 {
				fmt.Printf("  ⚠️  WARNING: 未找到任何卡片，可能的原因：\n")
				fmt.Printf("     1. residents JSONB 格式不正确（应该是对象数组）\n")
				fmt.Printf("     2. resident_id 字段未正确设置\n")
				fmt.Printf("     3. unit_id 字段未正确设置\n")
				hasIssues = true
			}
		}
	}

	if hasIssues {
		fmt.Println("\n❌ 发现问题，请检查上述警告和错误")
	} else {
		fmt.Println("\n✅ 未发现问题")
	}
}

