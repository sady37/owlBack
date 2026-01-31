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
	// 从环境变量获取数据库连接信息
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "owlrd"
	}

	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 测试查询1: 检查 device_store 中符合条件的设备数量
	query1 := `SELECT COUNT(*) FROM device_store WHERE allow_access = TRUE AND device_type = 'Radar'`
	var count1 int
	err = db.QueryRowContext(ctx, query1).Scan(&count1)
	if err != nil {
		log.Fatalf("Failed to query device_store count: %v", err)
	}
	fmt.Printf("1. device_store 中 allow_access=TRUE AND device_type='Radar' 的设备数量: %d\n", count1)

	// 测试查询2: 检查 initialDeviceCheck 使用的完整查询
	query2 := `
		SELECT ds.device_uid, ds.device_id,
			ds.tenant_id::text,
			ds.device_code,
			ds.device_type,
			b.branch_id::text,
			bu.building_id::text,
			u.unit_id::text,
			COALESCE(r.room_id::text, r_bed.room_id::text) as room_id,
			bd.bed_id::text
		FROM device_store ds
		LEFT JOIN devices d ON ds.device_id = d.device_id
		LEFT JOIN rooms r ON d.bound_room_id = r.room_id
		LEFT JOIN beds bd ON d.bound_bed_id = bd.bed_id
		LEFT JOIN rooms r_bed ON bd.room_id = r_bed.room_id
		LEFT JOIN units u ON COALESCE(r.unit_id, r_bed.unit_id) = u.unit_id
		LEFT JOIN buildings bu ON u.building_id = bu.building_id
		LEFT JOIN branches b ON bu.branch_id = b.branch_id
		WHERE ds.allow_access = TRUE AND ds.device_type = 'Radar'
		LIMIT 10
	`

	rows, err := db.QueryContext(ctx, query2)
	if err != nil {
		log.Fatalf("Failed to query devices: %v", err)
	}
	defer rows.Close()

	fmt.Printf("\n2. initialDeviceCheck 查询结果（前10条）:\n")
	fmt.Println("   device_uid | device_id | tenant_id | device_code | device_type | branch_id | building_id | unit_id | room_id | bed_id")
	fmt.Println("   " + "----------------------------------------------------------------------------------------------------------------------------------")

	count2 := 0
	for rows.Next() {
		var deviceUID, deviceID, tenantID, deviceCode, deviceType string
		var branchID, buildingID, unitID, roomID, bedID sql.NullString
		if err := rows.Scan(
			&deviceUID, &deviceID, &tenantID, &deviceCode, &deviceType,
			&branchID, &buildingID, &unitID, &roomID, &bedID,
		); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		count2++
		branchStr := "NULL"
		if branchID.Valid {
			branchStr = branchID.String[:8] + "..."
		}
		buildingStr := "NULL"
		if buildingID.Valid {
			buildingStr = buildingID.String[:8] + "..."
		}
		unitStr := "NULL"
		if unitID.Valid {
			unitStr = unitID.String[:8] + "..."
		}
		roomStr := "NULL"
		if roomID.Valid {
			roomStr = roomID.String[:8] + "..."
		}
		bedStr := "NULL"
		if bedID.Valid {
			bedStr = bedID.String[:8] + "..."
		}
		fmt.Printf("   %s | %s | %s | %s | %s | %s | %s | %s | %s | %s\n",
			deviceUID, deviceID[:8]+"...", tenantID[:8]+"...", deviceCode, deviceType,
			branchStr, buildingStr, unitStr, roomStr, bedStr)
	}
	fmt.Printf("\n   实际查询到的设备数量: %d\n", count2)

	// 测试查询3: 检查 device_store 中所有设备的状态
	query3 := `SELECT device_uid, allow_access, device_type FROM device_store WHERE device_type = 'Radar' LIMIT 10`
	rows3, err := db.QueryContext(ctx, query3)
	if err != nil {
		log.Fatalf("Failed to query device_store: %v", err)
	}
	defer rows3.Close()

	fmt.Printf("\n3. device_store 中所有 Radar 设备（前10条）:\n")
	fmt.Println("   device_uid | allow_access | device_type")
	fmt.Println("   " + "----------------------------------------")
	for rows3.Next() {
		var deviceUID string
		var allowAccess bool
		var deviceType string
		if err := rows3.Scan(&deviceUID, &allowAccess, &deviceType); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		fmt.Printf("   %s | %v | %s\n", deviceUID, allowAccess, deviceType)
	}
}
