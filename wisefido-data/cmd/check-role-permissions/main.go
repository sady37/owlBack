package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"wisefido-data/internal/config"

	_ "github.com/lib/pq"
)

func main() {
	var roleCode = flag.String("role", "", "Filter by role_code (e.g., 'Manager', 'Nurse')")
	var resourceType = flag.String("resource", "", "Filter by resource_type (e.g., 'residents', 'users')")
	var dbName = flag.String("db", "", "Database name (default: try 'wisefido_data' then 'owlrd')")
	flag.Parse()

	cfg := config.Load()

	dbNames := []string{"wisefido_data", "owlrd"}
	if *dbName != "" {
		dbNames = []string{*dbName}
	}

	var db *sql.DB
	var connectedDB string
	var err error

	for _, name := range dbNames {
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, name, cfg.Database.SSLMode)
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			continue
		}
		if err = db.Ping(); err != nil {
			db.Close()
			continue
		}
		connectedDB = name
		break
	}

	if db == nil || err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}
	defer db.Close()

	fmt.Printf("Connected to database: %s\n\n", connectedDB)

	// 首先查询总记录数
	var totalCount int
	err = db.QueryRow("SELECT COUNT(*) FROM role_permissions").Scan(&totalCount)
	if err != nil {
		log.Fatalf("Count query error: %v", err)
	}
	fmt.Printf("role_permissions 表中的记录数: %d\n\n", totalCount)

	if totalCount == 0 {
		fmt.Println("⚠️  role_permissions 表中没有记录！")
		return
	}

	// Build query
	query := `
		SELECT 
			rp.permission_id::text,
			COALESCE(rp.tenant_id::text, 'System') as tenant_id,
			rp.role_code,
			rp.resource_type,
			rp.permission_type,
			rp.assigned_only,
			rp.branch_only,
			COALESCE(SPLIT_PART(r.description, E'\n', 1), 'Unknown') as role_name,
			COALESCE(r.is_active, true) as role_active
		FROM role_permissions rp
		LEFT JOIN roles r ON r.role_code = rp.role_code
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if *roleCode != "" {
		query += fmt.Sprintf(" AND rp.role_code = $%d", argIdx)
		args = append(args, *roleCode)
		argIdx++
	}

	if *resourceType != "" {
		query += fmt.Sprintf(" AND rp.resource_type = $%d", argIdx)
		args = append(args, *resourceType)
		argIdx++
	}

	query += " ORDER BY rp.role_code, rp.resource_type, rp.permission_type"

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Fatalf("Query error: %v", err)
	}
	defer rows.Close()

	fmt.Println("Role Permissions Table:")
	fmt.Println("Permission ID | Tenant ID | Role Code | Role Name | Resource Type | Permission Type | Assigned Only | Branch Only | Role Active")
	fmt.Println("--------------|-----------|-----------|-----------|---------------|----------------|---------------|-------------|-------------")

	count := 0
	for rows.Next() {
		count++
		var permissionID, tenantID, roleCode, resourceType, permissionType, roleName sql.NullString
		var assignedOnly, branchOnly, roleActive sql.NullBool

		if err := rows.Scan(&permissionID, &tenantID, &roleCode, &resourceType, &permissionType, &assignedOnly, &branchOnly, &roleName, &roleActive); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		pid := getString(permissionID)
		if len(pid) > 12 {
			pid = pid[:12] + "..."
		}
		tid := getString(tenantID)
		if len(tid) > 12 {
			tid = tid[:12] + "..."
		}

		assignedOnlyStr := "FALSE"
		if assignedOnly.Valid && assignedOnly.Bool {
			assignedOnlyStr = "TRUE"
		}
		branchOnlyStr := "FALSE"
		if branchOnly.Valid && branchOnly.Bool {
			branchOnlyStr = "TRUE"
		}
		roleActiveStr := "TRUE"
		if roleActive.Valid && !roleActive.Bool {
			roleActiveStr = "FALSE"
		}

		fmt.Printf("%-12s | %-9s | %-9s | %-9s | %-13s | %-14s | %-13s | %-11s | %-11s\n",
			pid, tid, getString(roleCode), getString(roleName), getString(resourceType), getString(permissionType), assignedOnlyStr, branchOnlyStr, roleActiveStr)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Rows error: %v", err)
	}

	fmt.Printf("\nTotal permissions: %d\n", count)

	// Summary by role
	fmt.Println("\n=== Summary by Role ===")
	summaryQuery := `
		SELECT 
			rp.role_code,
			COALESCE(SPLIT_PART(r.description, E'\n', 1), 'Unknown') as role_name,
			COUNT(*) as permission_count,
			COALESCE(r.is_active, true) as role_active
		FROM role_permissions rp
		LEFT JOIN roles r ON r.role_code = rp.role_code
		GROUP BY rp.role_code, SPLIT_PART(r.description, E'\n', 1), r.is_active
		ORDER BY rp.role_code
	`
	summaryRows, err := db.Query(summaryQuery)
	if err != nil {
		log.Printf("Summary query error: %v", err)
		return
	}
	defer summaryRows.Close()

	fmt.Println("Role Code | Role Name | Permission Count | Role Active")
	fmt.Println("----------|-----------|------------------|-------------")
	for summaryRows.Next() {
		var roleCode, roleName sql.NullString
		var permissionCount int
		var roleActive sql.NullBool
		if err := summaryRows.Scan(&roleCode, &roleName, &permissionCount, &roleActive); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		roleActiveStr := "TRUE"
		if roleActive.Valid && !roleActive.Bool {
			roleActiveStr = "FALSE"
		}
		fmt.Printf("%-9s | %-9s | %-16d | %-11s\n", getString(roleCode), getString(roleName), permissionCount, roleActiveStr)
	}
}

func getString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "NULL"
}
