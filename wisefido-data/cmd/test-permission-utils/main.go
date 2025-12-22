package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"wisefido-data/internal/config"
	httpapi "wisefido-data/internal/http"
)

func main() {
	cfg := config.Load()

	dbNames := []string{"wisefido_data", "owlrd"}
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

	ctx := context.Background()

	// Test 1: Get Manager permissions for residents (R)
	fmt.Println("=== Test 1: Manager permissions for residents (R) ===")
	perm, err := httpapi.GetResourcePermission(db, ctx, "Manager", "residents", "R")
	if err != nil {
		log.Fatalf("Failed to get permission: %v", err)
	}
	fmt.Printf("Assigned Only: %v\n", perm.AssignedOnly)
	fmt.Printf("Branch Only: %v\n", perm.BranchOnly)
	fmt.Printf("Expected: AssignedOnly=false, BranchOnly=true\n")
	if perm.AssignedOnly == false && perm.BranchOnly == true {
		fmt.Println("✅ Test 1 PASSED")
	} else {
		fmt.Println("❌ Test 1 FAILED")
	}

	// Test 2: Get Admin permissions for residents (R)
	fmt.Println("=== Test 2: Admin permissions for residents (R) ===")
	perm2, err := httpapi.GetResourcePermission(db, ctx, "Admin", "residents", "R")
	if err != nil {
		log.Fatalf("Failed to get permission: %v", err)
	}
	fmt.Printf("Assigned Only: %v\n", perm2.AssignedOnly)
	fmt.Printf("Branch Only: %v\n", perm2.BranchOnly)
	fmt.Printf("Expected: AssignedOnly=false, BranchOnly=false\n")
	if perm2.AssignedOnly == false && perm2.BranchOnly == false {
		fmt.Println("✅ Test 2 PASSED")
	} else {
		fmt.Println("❌ Test 2 FAILED")
	}

	// Test 3: Get Nurse permissions for residents (R)
	fmt.Println("=== Test 3: Nurse permissions for residents (R) ===")
	perm3, err := httpapi.GetResourcePermission(db, ctx, "Nurse", "residents", "R")
	if err != nil {
		log.Fatalf("Failed to get permission: %v", err)
	}
	fmt.Printf("Assigned Only: %v\n", perm3.AssignedOnly)
	fmt.Printf("Branch Only: %v\n", perm3.BranchOnly)
	fmt.Printf("Expected: AssignedOnly=true, BranchOnly=false\n")
	if perm3.AssignedOnly == true && perm3.BranchOnly == false {
		fmt.Println("✅ Test 3 PASSED")
	} else {
		fmt.Println("❌ Test 3 FAILED")
	}

	// Test 4: Test ApplyBranchFilter with NULL branch_tag
	fmt.Println("=== Test 4: ApplyBranchFilter with NULL branch_tag ===")
	query := "SELECT * FROM residents r LEFT JOIN units u ON u.unit_id = r.unit_id"
	args := []any{}
	var userBranchTag sql.NullString // NULL
	httpapi.ApplyBranchFilter(&query, &args, userBranchTag, "u", true)
	fmt.Printf("Query: %s\n", query)
	fmt.Printf("Args: %v\n", args)
	fmt.Printf("Expected: WHERE (u.branch_tag IS NULL OR u.branch_tag = '-')\n")
	expectedQuery := "SELECT * FROM residents r LEFT JOIN units u ON u.unit_id = r.unit_id WHERE (u.branch_tag IS NULL OR u.branch_tag = '-')"
	if len(args) == 0 && query == expectedQuery {
		fmt.Println("✅ Test 4 PASSED")
	} else {
		fmt.Printf("❌ Test 4 FAILED - Query mismatch\n")
		fmt.Printf("  Expected: %s\n", expectedQuery)
		fmt.Printf("  Got:      %s\n\n", query)
	}

	// Test 5: Test ApplyBranchFilter with branch_tag value
	fmt.Println("=== Test 5: ApplyBranchFilter with branch_tag='BranchA' ===")
	query2 := "SELECT * FROM residents r LEFT JOIN units u ON u.unit_id = r.unit_id"
	args2 := []any{}
	userBranchTag2 := sql.NullString{String: "BranchA", Valid: true}
	httpapi.ApplyBranchFilter(&query2, &args2, userBranchTag2, "u", true)
	fmt.Printf("Query: %s\n", query2)
	fmt.Printf("Args: %v\n", args2)
	fmt.Printf("Expected: WHERE u.branch_tag = $1, Args: [BranchA]\n")
	expectedQuery2 := "SELECT * FROM residents r LEFT JOIN units u ON u.unit_id = r.unit_id WHERE u.branch_tag = $1"
	if len(args2) == 1 && args2[0] == "BranchA" && query2 == expectedQuery2 {
		fmt.Println("✅ Test 5 PASSED")
	} else {
		fmt.Printf("❌ Test 5 FAILED\n")
		fmt.Printf("  Expected Query: %s\n", expectedQuery2)
		fmt.Printf("  Got Query:      %s\n", query2)
		fmt.Printf("  Expected Args:  [BranchA]\n")
		fmt.Printf("  Got Args:       %v\n\n", args2)
	}

	fmt.Println("✅ All tests completed!")
}
