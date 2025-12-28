//go:build ignore
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

	fmt.Println("🚀 Inserting missing role permissions...")
	fmt.Println("")

	// SystemAdmin permissions
	fmt.Println("📝 Inserting SystemAdmin permissions...")
	systemAdminSQL := `
		INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only) VALUES
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'tenants', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'tenants', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'tenants', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'tenants', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'roles', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'roles', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'roles', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'roles', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'role_permissions', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'role_permissions', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'role_permissions', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'role_permissions', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'device_store', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'device_store', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'device_store', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'device_store', 'D', FALSE)
		ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type) 
		DO UPDATE SET assigned_only = EXCLUDED.assigned_only
	`
	result, err := db.Exec(systemAdminSQL)
	if err != nil {
		log.Printf("⚠️  SystemAdmin insert error (may already exist): %v", err)
	} else {
		count, _ := result.RowsAffected()
		fmt.Printf("   ✅ SystemAdmin: %d permissions\n", count)
	}

	// SystemOperator permissions
	fmt.Println("📝 Inserting SystemOperator permissions...")
	systemOperatorSQL := `
		INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only) VALUES
		('00000000-0000-0000-0000-000000000001', 'SystemOperator', 'tenants', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemOperator', 'tenants', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemOperator', 'tenants', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemOperator', 'tenants', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemOperator', 'device_store', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemOperator', 'device_store', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemOperator', 'device_store', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemOperator', 'device_store', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemOperator', 'alarm_cloud', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemOperator', 'alarm_cloud', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemOperator', 'alarm_cloud', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemOperator', 'alarm_cloud', 'D', FALSE)
		ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type) 
		DO UPDATE SET assigned_only = EXCLUDED.assigned_only
	`
	result, err = db.Exec(systemOperatorSQL)
	if err != nil {
		log.Printf("⚠️  SystemOperator insert error (may already exist): %v", err)
	} else {
		count, _ := result.RowsAffected()
		fmt.Printf("   ✅ SystemOperator: %d permissions\n", count)
	}

	// Manager permissions (with branch_only)
	fmt.Println("📝 Inserting Manager permissions...")
	managerSQL := `
		INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only, branch_only) VALUES
		('00000000-0000-0000-0000-000000000001', 'Manager', 'roles', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'roles', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'users', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'users', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'users', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'users', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'units', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'units', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'units', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'units', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'rooms', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'rooms', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'rooms', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'rooms', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'beds', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'beds', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'beds', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'beds', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'residents', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'residents', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'residents', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'residents', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'resident_phi', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'resident_phi', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'resident_phi', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'resident_phi', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'resident_contacts', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'resident_contacts', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'resident_contacts', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'resident_contacts', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'resident_caregivers', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'resident_caregivers', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'resident_caregivers', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'resident_caregivers', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'devices', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'devices', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'devices', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'devices', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'device_store', 'R', TRUE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'iot_timeseries', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'alarm_events', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'alarm_events', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'alarm_events', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'alarm_events', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'alarm_device', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'alarm_device', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'alarm_device', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'alarm_device', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'alarm_cloud', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'alarm_cloud', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'alarm_cloud', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'alarm_cloud', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'service_levels', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'service_levels', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'service_levels', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'service_levels', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'cards', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'rounds', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'rounds', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'rounds', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'rounds', 'D', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'round_details', 'R', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'round_details', 'C', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'round_details', 'U', FALSE, TRUE),
		('00000000-0000-0000-0000-000000000001', 'Manager', 'round_details', 'D', FALSE, TRUE)
		ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type) 
		DO UPDATE SET assigned_only = EXCLUDED.assigned_only, branch_only = EXCLUDED.branch_only
	`
	result, err = db.Exec(managerSQL)
	if err != nil {
		log.Printf("⚠️  Manager insert error (may already exist): %v", err)
	} else {
		count, _ := result.RowsAffected()
		fmt.Printf("   ✅ Manager: %d permissions\n", count)
	}

	// IT permissions
	fmt.Println("📝 Inserting IT permissions...")
	itSQL := `
		INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only) VALUES
		('00000000-0000-0000-0000-000000000001', 'IT', 'roles', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'users', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'users', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'users', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'users', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'units', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'units', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'units', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'units', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'rooms', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'rooms', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'rooms', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'rooms', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'beds', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'beds', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'beds', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'beds', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'residents', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'resident_caregivers', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'devices', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'devices', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'devices', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'devices', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'device_store', 'R', TRUE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'alarm_events', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'alarm_device', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'alarm_device', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'alarm_device', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'alarm_device', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'config_versions', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'config_versions', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'config_versions', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'IT', 'config_versions', 'D', FALSE)
		ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type) 
		DO UPDATE SET assigned_only = EXCLUDED.assigned_only
	`
	result, err = db.Exec(itSQL)
	if err != nil {
		log.Printf("⚠️  IT insert error (may already exist): %v", err)
	} else {
		count, _ := result.RowsAffected()
		fmt.Printf("   ✅ IT: %d permissions\n", count)
	}

	fmt.Println("")
	fmt.Println("✅ All missing role permissions have been inserted!")
	fmt.Println("")

	// Verify
	rows, err := db.Query(`
		SELECT role_code, COUNT(*) as count
		FROM role_permissions
		WHERE tenant_id = $1
		GROUP BY role_code
		ORDER BY role_code
	`, systemTenantID)
	if err != nil {
		log.Fatalf("Failed to verify: %v", err)
	}
	defer rows.Close()

	fmt.Println("📊 Final permissions count by role:")
	fmt.Println("Role Code          | Count")
	fmt.Println("----------------------------")
	totalCount := 0
	for rows.Next() {
		var roleCode string
		var count int
		if err := rows.Scan(&roleCode, &count); err != nil {
			log.Fatalf("Failed to scan: %v", err)
		}
		fmt.Printf("%-18s | %d\n", roleCode, count)
		totalCount += count
	}
	fmt.Printf("\nTotal: %d permissions\n", totalCount)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
