package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"owl-common/config"
	"owl-common/database"
)

func main() {
	// Load database config from environment or use defaults
	dbCfg := &config.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     parseInt(getEnv("DB_PORT", "5432"), 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Database: getEnv("DB_NAME", "owlrd"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db, err := database.NewPostgresDB(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	nickname := "done"
	if len(os.Args) > 1 {
		nickname = os.Args[1]
	}

	fmt.Printf("\n=== Checking Resident: %s ===\n\n", nickname)

	// 1. Get resident profile
	var residentID, residentAccount, status, serviceLevel, familyTag, note, unitID, roomID, bedID sql.NullString
	var admissionDate, dischargeDate sql.NullTime
	var canViewStatus sql.NullBool
	var phoneHash, emailHash []byte

	err = db.QueryRow(`
		SELECT 
			r.resident_id::text,
			r.resident_account,
			r.status,
			r.service_level,
			r.admission_date,
			r.discharge_date,
			r.family_tag,
			r.can_view_status,
			r.note,
			r.unit_id::text,
			r.room_id::text,
			r.bed_id::text,
			r.phone_hash,
			r.email_hash
		FROM residents r
		WHERE LOWER(r.nickname) = LOWER($1)
		LIMIT 1
	`, nickname).Scan(
		&residentID, &residentAccount, &status, &serviceLevel,
		&admissionDate, &dischargeDate, &familyTag, &canViewStatus,
		&note, &unitID, &roomID, &bedID, &phoneHash, &emailHash,
	)

	if err == sql.ErrNoRows {
		fmt.Printf("❌ Resident '%s' not found\n", nickname)
		return
	}
	if err != nil {
		log.Fatalf("Failed to query resident: %v", err)
	}

	fmt.Println("📋 PROFILE (residents table):")
	fmt.Printf("  resident_id: %s\n", residentID.String)
	fmt.Printf("  resident_account: %s\n", residentAccount.String)
	fmt.Printf("  nickname: %s\n", nickname)
	fmt.Printf("  status: %s\n", status.String)
	fmt.Printf("  service_level: %s\n", serviceLevel.String)
	if admissionDate.Valid {
		fmt.Printf("  admission_date: %s\n", admissionDate.Time.Format("2006-01-02"))
	}
	if dischargeDate.Valid {
		fmt.Printf("  discharge_date: %s\n", dischargeDate.Time.Format("2006-01-02"))
	}
	fmt.Printf("  family_tag: %s\n", familyTag.String)
	fmt.Printf("  can_view_status: %v\n", canViewStatus.Bool)
	fmt.Printf("  note: %s\n", note.String)
	fmt.Printf("  unit_id: %s\n", unitID.String)
	fmt.Printf("  room_id: %s\n", roomID.String)
	fmt.Printf("  bed_id: %s\n", bedID.String)
	if phoneHash != nil {
		fmt.Printf("  phone_hash: %s\n", hex.EncodeToString(phoneHash))
	} else {
		fmt.Printf("  phone_hash: NULL\n")
	}
	if emailHash != nil {
		fmt.Printf("  email_hash: %s\n", hex.EncodeToString(emailHash))
	} else {
		fmt.Printf("  email_hash: NULL\n")
	}

	// 2. Get PHI data
	fmt.Println("\n🏥 PHI (resident_phi table):")
	rows, err := db.Query(`
		SELECT 
			rp.phi_id::text,
			rp.first_name,
			rp.last_name,
			rp.gender,
			rp.date_of_birth,
			rp.resident_phone,
			rp.resident_email,
			rp.weight_lb,
			rp.height_ft,
			rp.height_in,
			rp.mobility_level,
			rp.tremor_status,
			rp.mobility_aid,
			rp.adl_assistance,
			rp.comm_status,
			rp.has_hypertension,
			rp.has_hyperlipaemia,
			rp.has_hyperglycaemia,
			rp.has_stroke_history,
			rp.has_paralysis,
			rp.has_alzheimer,
			rp.medical_history,
			rp.HIS_resident_name,
			rp.HIS_resident_admission_date,
			rp.HIS_resident_discharge_date,
			rp.home_address_street,
			rp.home_address_city,
			rp.home_address_state,
			rp.home_address_postal_code,
			rp.plus_code
		FROM residents r
		JOIN resident_phi rp ON r.resident_id = rp.resident_id
		WHERE LOWER(r.nickname) = LOWER($1)
	`, nickname)

	if err != nil {
		log.Fatalf("Failed to query PHI: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		fmt.Println("  ❌ No PHI record found")
	} else {
		var phiID, firstName, lastName, gender, residentPhone, residentEmail, tremorStatus, mobilityAid, adlAssistance, commStatus, medicalHistory, hisResidentName, homeAddressStreet, homeAddressCity, homeAddressState, homeAddressPostalCode, plusCode sql.NullString
		var dateOfBirth, hisAdmissionDate, hisDischargeDate sql.NullTime
		var weightLb, heightFt, heightIn sql.NullFloat64
		var mobilityLevel sql.NullInt64
		var hasHypertension, hasHyperlipaemia, hasHyperglycaemia, hasStrokeHistory, hasParalysis, hasAlzheimer sql.NullBool

		err = rows.Scan(
			&phiID, &firstName, &lastName, &gender, &dateOfBirth,
			&residentPhone, &residentEmail, &weightLb, &heightFt, &heightIn,
			&mobilityLevel, &tremorStatus, &mobilityAid, &adlAssistance, &commStatus,
			&hasHypertension, &hasHyperlipaemia, &hasHyperglycaemia, &hasStrokeHistory, &hasParalysis, &hasAlzheimer,
			&medicalHistory, &hisResidentName, &hisAdmissionDate, &hisDischargeDate,
			&homeAddressStreet, &homeAddressCity, &homeAddressState, &homeAddressPostalCode, &plusCode,
		)
		if err != nil {
			log.Fatalf("Failed to scan PHI: %v", err)
		}

		fmt.Printf("  phi_id: %s\n", phiID.String)
		fmt.Printf("  first_name: %s\n", firstName.String)
		fmt.Printf("  last_name: %s\n", lastName.String)
		fmt.Printf("  gender: %s\n", gender.String)
		if dateOfBirth.Valid {
			fmt.Printf("  date_of_birth: %s\n", dateOfBirth.Time.Format("2006-01-02"))
		} else {
			fmt.Printf("  date_of_birth: NULL\n")
		}
		fmt.Printf("  resident_phone: %s\n", residentPhone.String)
		fmt.Printf("  resident_email: %s\n", residentEmail.String)
		if weightLb.Valid {
			fmt.Printf("  weight_lb: %.2f\n", weightLb.Float64)
		} else {
			fmt.Printf("  weight_lb: NULL\n")
		}
		if heightFt.Valid {
			fmt.Printf("  height_ft: %.2f\n", heightFt.Float64)
		} else {
			fmt.Printf("  height_ft: NULL\n")
		}
		if heightIn.Valid {
			fmt.Printf("  height_in: %.2f\n", heightIn.Float64)
		} else {
			fmt.Printf("  height_in: NULL\n")
		}
		if mobilityLevel.Valid {
			fmt.Printf("  mobility_level: %d\n", mobilityLevel.Int64)
		} else {
			fmt.Printf("  mobility_level: NULL\n")
		}
		fmt.Printf("  tremor_status: %s\n", tremorStatus.String)
		fmt.Printf("  mobility_aid: %s\n", mobilityAid.String)
		fmt.Printf("  adl_assistance: %s\n", adlAssistance.String)
		fmt.Printf("  comm_status: %s\n", commStatus.String)
		if hasHypertension.Valid {
			fmt.Printf("  has_hypertension: %v\n", hasHypertension.Bool)
		} else {
			fmt.Printf("  has_hypertension: NULL\n")
		}
		if hasHyperlipaemia.Valid {
			fmt.Printf("  has_hyperlipaemia: %v\n", hasHyperlipaemia.Bool)
		} else {
			fmt.Printf("  has_hyperlipaemia: NULL\n")
		}
		if hasHyperglycaemia.Valid {
			fmt.Printf("  has_hyperglycaemia: %v\n", hasHyperglycaemia.Bool)
		} else {
			fmt.Printf("  has_hyperglycaemia: NULL\n")
		}
		if hasStrokeHistory.Valid {
			fmt.Printf("  has_stroke_history: %v\n", hasStrokeHistory.Bool)
		} else {
			fmt.Printf("  has_stroke_history: NULL\n")
		}
		if hasParalysis.Valid {
			fmt.Printf("  has_paralysis: %v\n", hasParalysis.Bool)
		} else {
			fmt.Printf("  has_paralysis: NULL\n")
		}
		if hasAlzheimer.Valid {
			fmt.Printf("  has_alzheimer: %v\n", hasAlzheimer.Bool)
		} else {
			fmt.Printf("  has_alzheimer: NULL\n")
		}
		fmt.Printf("  medical_history: %s\n", medicalHistory.String)
		fmt.Printf("  HIS_resident_name: %s\n", hisResidentName.String)
		if hisAdmissionDate.Valid {
			fmt.Printf("  HIS_resident_admission_date: %s\n", hisAdmissionDate.Time.Format("2006-01-02"))
		} else {
			fmt.Printf("  HIS_resident_admission_date: NULL\n")
		}
		if hisDischargeDate.Valid {
			fmt.Printf("  HIS_resident_discharge_date: %s\n", hisDischargeDate.Time.Format("2006-01-02"))
		} else {
			fmt.Printf("  HIS_resident_discharge_date: NULL\n")
		}
		fmt.Printf("  home_address_street: %s\n", homeAddressStreet.String)
		fmt.Printf("  home_address_city: %s\n", homeAddressCity.String)
		fmt.Printf("  home_address_state: %s\n", homeAddressState.String)
		fmt.Printf("  home_address_postal_code: %s\n", homeAddressPostalCode.String)
		fmt.Printf("  plus_code: %s\n", plusCode.String)
	}

	// 3. Get contacts
	fmt.Println("\n👥 CONTACTS (resident_contacts table):")
	contactRows, err := db.Query(`
		SELECT 
			rc.contact_id::text,
			rc.slot,
			rc.contact_family_tag,
			rc.is_enabled,
			rc.relationship,
			rc.is_emergency_contact,
			rc.alert_time_window,
			rc.contact_first_name,
			rc.contact_last_name,
			rc.contact_phone,
			rc.contact_email,
			rc.receive_sms,
			rc.receive_email,
			rc.phone_hash,
			rc.email_hash,
			CASE WHEN rc.password_hash IS NOT NULL THEN true ELSE false END as has_password
		FROM residents r
		JOIN resident_contacts rc ON r.resident_id = rc.resident_id
		WHERE LOWER(r.nickname) = LOWER($1)
		ORDER BY rc.slot
	`, nickname)

	if err != nil {
		log.Fatalf("Failed to query contacts: %v", err)
	}
	defer contactRows.Close()

	contactCount := 0
	for contactRows.Next() {
		contactCount++
		var contactID, slot, contactFamilyTag, relationship, contactFirstName, contactLastName, contactPhone, contactEmail sql.NullString
		var isEnabled, isEmergencyContact, receiveSms, receiveEmail, hasPassword sql.NullBool
		var alertTimeWindow sql.NullString
		var phoneHash, emailHash []byte

		err = contactRows.Scan(
			&contactID, &slot, &contactFamilyTag, &isEnabled, &relationship,
			&isEmergencyContact, &alertTimeWindow,
			&contactFirstName, &contactLastName, &contactPhone, &contactEmail,
			&receiveSms, &receiveEmail, &phoneHash, &emailHash, &hasPassword,
		)
		if err != nil {
			log.Fatalf("Failed to scan contact: %v", err)
		}

		fmt.Printf("\n  Contact #%d (Slot: %s):\n", contactCount, slot.String)
		fmt.Printf("    contact_id: %s\n", contactID.String)
		fmt.Printf("    slot: %s\n", slot.String)
		fmt.Printf("    contact_family_tag: %s\n", contactFamilyTag.String)
		fmt.Printf("    is_enabled: %v\n", isEnabled.Bool)
		fmt.Printf("    relationship: %s\n", relationship.String)
		fmt.Printf("    is_emergency_contact: %v\n", isEmergencyContact.Bool)
		fmt.Printf("    alert_time_window: %s\n", alertTimeWindow.String)
		fmt.Printf("    contact_first_name: %s\n", contactFirstName.String)
		fmt.Printf("    contact_last_name: %s\n", contactLastName.String)
		fmt.Printf("    contact_phone: %s\n", contactPhone.String)
		fmt.Printf("    contact_email: %s\n", contactEmail.String)
		fmt.Printf("    receive_sms: %v\n", receiveSms.Bool)
		fmt.Printf("    receive_email: %v\n", receiveEmail.Bool)
		if phoneHash != nil {
			fmt.Printf("    phone_hash: %s\n", hex.EncodeToString(phoneHash))
		} else {
			fmt.Printf("    phone_hash: NULL\n")
		}
		if emailHash != nil {
			fmt.Printf("    email_hash: %s\n", hex.EncodeToString(emailHash))
		} else {
			fmt.Printf("    email_hash: NULL\n")
		}
		fmt.Printf("    has_password: %v\n", hasPassword.Bool)
	}

	if contactCount == 0 {
		fmt.Println("  ❌ No contacts found")
	} else {
		fmt.Printf("\n  Total contacts: %d\n", contactCount)
	}

	fmt.Println("\n=== Done ===\n")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseInt(s string, def int) int {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return def
	}
	return i
}
