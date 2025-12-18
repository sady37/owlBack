package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	commoncfg "owl-common/config"
	"owl-common/database"
)

func main() {
	// 从环境变量获取数据库连接信息
	cfg := &commoncfg.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:      parseInt(getEnv("DB_PORT", "5432"), 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Database: getEnv("DB_NAME", "owlrd"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// 连接数据库
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// 1. 检查 residents 表中的 email_hash
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("1. Residents 表中的 email_hash 和 phone_hash")
	fmt.Println("=" + string(make([]byte, 80)))
	query1 := `
		SELECT 
			r.resident_id,
			r.nickname,
			r.resident_account,
			CASE WHEN r.email_hash IS NULL THEN 'NULL' ELSE 'HAS_EMAIL_HASH' END as email_hash_status,
			CASE WHEN r.phone_hash IS NULL THEN 'NULL' ELSE 'HAS_PHONE_HASH' END as phone_hash_status,
			CASE WHEN r.password_hash IS NULL THEN 'NULL' ELSE 'HAS_PASSWORD_HASH' END as password_hash_status
		FROM residents r
		WHERE LOWER(r.nickname) IN ('done', 'smith')
		   OR LOWER(r.resident_account) IN ('done', 'smith', 'r1', 'r2')
		ORDER BY r.nickname;
	`

	rows1, err := db.Query(query1)
	if err != nil {
		log.Fatalf("Failed to query residents: %v", err)
	}
	defer rows1.Close()

	fmt.Printf("%-40s %-20s %-20s %-20s %-20s %-20s\n",
		"resident_id", "nickname", "resident_account", "email_hash", "phone_hash", "password_hash")
	fmt.Println("-" + string(make([]byte, 80)))

	for rows1.Next() {
		var residentID, nickname, residentAccount, emailHashStatus, phoneHashStatus, passwordHashStatus sql.NullString
		err := rows1.Scan(&residentID, &nickname, &residentAccount, &emailHashStatus, &phoneHashStatus, &passwordHashStatus)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		fmt.Printf("%-40s %-20s %-20s %-20s %-20s %-20s\n",
			getString(residentID), getString(nickname), getString(residentAccount),
			getString(emailHashStatus), getString(phoneHashStatus), getString(passwordHashStatus))
	}

	// 2. 检查 resident_phi 表中的记录
	fmt.Println("\n" + "=" + string(make([]byte, 80)))
	fmt.Println("2. Resident_phi 表中的记录（检查是否有数据）")
	fmt.Println("=" + string(make([]byte, 80)))
	query2 := `
		SELECT 
			rp.phi_id,
			rp.resident_id,
			r.nickname,
			rp.first_name,
			rp.last_name,
			rp.gender,
			rp.date_of_birth,
			CASE WHEN rp.resident_phone IS NULL THEN 'NULL' ELSE rp.resident_phone END as resident_phone,
			CASE WHEN rp.resident_email IS NULL THEN 'NULL' ELSE rp.resident_email END as resident_email,
			rp.weight_lb,
			rp.height_ft,
			rp.height_in,
			rp.mobility_level,
			rp.has_hypertension,
			rp.has_hyperlipaemia,
			rp.has_hyperglycaemia
		FROM resident_phi rp
		LEFT JOIN residents r ON rp.resident_id = r.resident_id
		WHERE LOWER(r.nickname) IN ('done', 'smith')
		   OR LOWER(r.resident_account) IN ('done', 'smith', 'r1', 'r2')
		ORDER BY r.nickname;
	`

	rows2, err := db.Query(query2)
	if err != nil {
		log.Fatalf("Failed to query resident_phi: %v", err)
	}
	defer rows2.Close()

	fmt.Printf("%-40s %-40s %-20s %-15s %-15s %-10s %-15s %-30s %-30s %-10s %-10s %-10s %-10s %-10s %-10s %-10s\n",
		"phi_id", "resident_id", "nickname", "first_name", "last_name", "gender", "date_of_birth",
		"resident_phone", "resident_email", "weight_lb", "height_ft", "height_in", "mobility_level",
		"has_hypertension", "has_hyperlipaemia", "has_hyperglycaemia")
	fmt.Println("-" + string(make([]byte, 80)))

	var phiCount int
	for rows2.Next() {
		var phiID, residentID, nickname, firstName, lastName, gender, residentPhone, residentEmail sql.NullString
		var dateOfBirth sql.NullTime
		var weightLb, heightFt, heightIn sql.NullFloat64
		var mobilityLevel sql.NullInt64
		var hasHypertension, hasHyperlipaemia, hasHyperglycaemia sql.NullBool

		err := rows2.Scan(&phiID, &residentID, &nickname, &firstName, &lastName, &gender, &dateOfBirth,
			&residentPhone, &residentEmail, &weightLb, &heightFt, &heightIn, &mobilityLevel,
			&hasHypertension, &hasHyperlipaemia, &hasHyperglycaemia)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		fmt.Printf("%-40s %-40s %-20s %-15s %-15s %-10s %-15s %-30s %-30s %-10s %-10s %-10s %-10s %-10s %-10s %-10s\n",
			getString(phiID), getString(residentID), getString(nickname),
			getString(firstName), getString(lastName), getString(gender), getTime(dateOfBirth),
			getString(residentPhone), getString(residentEmail),
			getFloat(weightLb), getFloat(heightFt), getFloat(heightIn), getInt(mobilityLevel),
			getBool(hasHypertension), getBool(hasHyperlipaemia), getBool(hasHyperglycaemia))
		phiCount++
	}

	if phiCount == 0 {
		fmt.Println("⚠️  未找到 resident_phi 记录！")
	} else {
		fmt.Printf("\n共找到 %d 条 resident_phi 记录\n", phiCount)
	}

	// 3. 检查是否有 resident_phi 记录但数据为空
	fmt.Println("\n" + "=" + string(make([]byte, 80)))
	fmt.Println("3. 检查 resident_phi 记录是否存在但数据为空")
	fmt.Println("=" + string(make([]byte, 80)))
	query3 := `
		SELECT 
			r.resident_id,
			r.nickname,
			CASE WHEN rp.phi_id IS NULL THEN 'NO_PHI_RECORD' ELSE 'HAS_PHI_RECORD' END as phi_status
		FROM residents r
		LEFT JOIN resident_phi rp ON r.resident_id = rp.resident_id
		WHERE LOWER(r.nickname) IN ('done', 'smith')
		   OR LOWER(r.resident_account) IN ('done', 'smith', 'r1', 'r2')
		ORDER BY r.nickname;
	`

	rows3, err := db.Query(query3)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer rows3.Close()

	fmt.Printf("%-40s %-20s %-20s\n", "resident_id", "nickname", "phi_status")
	fmt.Println("-" + string(make([]byte, 80)))

	for rows3.Next() {
		var residentID, nickname, phiStatus sql.NullString
		err := rows3.Scan(&residentID, &nickname, &phiStatus)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		fmt.Printf("%-40s %-20s %-20s\n", getString(residentID), getString(nickname), getString(phiStatus))
	}
}

func getString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return "NULL"
}

func getTime(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format("2006-01-02")
	}
	return "NULL"
}

func getFloat(f sql.NullFloat64) string {
	if f.Valid {
		return fmt.Sprintf("%.2f", f.Float64)
	}
	return "NULL"
}

func getInt(i sql.NullInt64) string {
	if i.Valid {
		return fmt.Sprintf("%d", i.Int64)
	}
	return "NULL"
}

func getBool(b sql.NullBool) string {
	if b.Valid {
		return fmt.Sprintf("%v", b.Bool)
	}
	return "NULL"
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseInt(s string, def int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return i
}

