//go:build integration
// +build integration

package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"github.com/stretchr/testify/require"
)

// TestUpdateResident_Basic 测试基本的 UpdateResident 功能
func TestUpdateResident_Basic(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建测试用户（Admin）
	userID := "00000000-0000-0000-0000-000000000980"
	createTestUserForCreateResident(t, db, tenantID, userID, "testadmin-update", "password123", "Admin", "")
	createTestPermissionForCreateResident(t, db, "Admin")

	// 创建 service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	logger := getTestLoggerForResident()
	service := NewResidentService(residentsRepo, db, nil, logger)

	ctx := context.Background()

	// 1. 先创建一个住户
	admissionDate := time.Now().Unix()
	createReq := CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		InherentAttributes: &CreateResidentInherentAttributes{
			ResidentAccount: "testresident-update001",
			Nickname:        "Test Resident Update 001",
			Status:          "active",
			AdmissionDate:   &admissionDate,
		},
		UnitRelation: &CreateResidentUnitRelation{
			UnitID: unitID,
		},
	}

	createResp, err := service.CreateResident(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	residentID := createResp.ResidentID

	// 2. 测试更新基本字段
	newNickname := "Updated Nickname"
	updateReq := UpdateResidentRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		InherentAttributes: &UpdateResidentInherentAttributes{
			Nickname: &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  newNickname,
			},
			Note: &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  "Updated note",
			},
		},
	}

	updateResp, err := service.UpdateResident(ctx, updateReq)
	require.NoError(t, err)
	require.NotNil(t, updateResp)
	require.True(t, updateResp.Success)

	// 3. 验证更新成功
	var nickname, note string
	err = db.QueryRow(
		`SELECT nickname, note FROM residents WHERE resident_id = $1`,
		residentID,
	).Scan(&nickname, &note)
	require.NoError(t, err)
	require.Equal(t, newNickname, nickname)
	require.Equal(t, "Updated note", note)

	t.Logf("UpdateResident succeeded, resident_id: %s", residentID)
}

// TestUpdateResident_WithPHI 测试更新 PHI 数据
func TestUpdateResident_WithPHI(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建测试用户（Admin）
	userID := "00000000-0000-0000-0000-000000000981"
	createTestUserForCreateResident(t, db, tenantID, userID, "testadmin-update-phi", "password123", "Admin", "")
	createTestPermissionForCreateResident(t, db, "Admin")

	// 创建 service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	logger := getTestLoggerForResident()
	service := NewResidentService(residentsRepo, db, nil, logger)

	ctx := context.Background()

	// 1. 先创建一个住户（带 PHI）
	admissionDate := time.Now().Unix()
	dob := time.Date(1940, 1, 1, 0, 0, 0, 0, time.UTC)
	dobUnix := dob.Unix()
	createReq := CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		InherentAttributes: &CreateResidentInherentAttributes{
			ResidentAccount: "testresident-update-phi001",
			Nickname:        "Test Resident Update PHI 001",
			Status:          "active",
			AdmissionDate:   &admissionDate,
			PHI: &CreateResidentPHIRequest{
				FirstName:   "John",
				LastName:    "Doe",
				Gender:      "Male",
				DateOfBirth: &dobUnix,
			},
		},
		UnitRelation: &CreateResidentUnitRelation{
			UnitID: unitID,
		},
	}

	createResp, err := service.CreateResident(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	residentID := createResp.ResidentID

	// 2. 更新 PHI 数据
	newDOB := time.Date(1945, 5, 15, 0, 0, 0, 0, time.UTC)
	updateReq := UpdateResidentRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		InherentAttributes: &UpdateResidentInherentAttributes{
			PHI: &UpdateResidentPHIRequest{
				LastName: &domain.UpdateString{
					Action: domain.UpdateActionUpdate,
					Value:  "Smith",
				},
				Gender: &domain.UpdateString{
					Action: domain.UpdateActionUpdate,
					Value:  "Female",
				},
				DateOfBirth: &domain.UpdateTime{
					Action: domain.UpdateActionUpdate,
					Value:  &newDOB,
				},
				WeightLb: &domain.UpdateFloat64{
					Action: domain.UpdateActionUpdate,
					Value:  150.5,
				},
			},
		},
	}

	updateResp, err := service.UpdateResident(ctx, updateReq)
	require.NoError(t, err)
	require.NotNil(t, updateResp)
	require.True(t, updateResp.Success)

	// 3. 验证 PHI 更新成功
	var lastName, gender string
	var dobTime time.Time
	var weightLb sql.NullFloat64
	err = db.QueryRow(
		`SELECT last_name, gender, date_of_birth, weight_lb 
		 FROM resident_phi 
		 WHERE resident_id = $1`,
		residentID,
	).Scan(&lastName, &gender, &dobTime, &weightLb)
	require.NoError(t, err)
	require.Equal(t, "Smith", lastName)
	require.Equal(t, "Female", gender)
	require.Equal(t, newDOB.Year(), dobTime.Year())
	require.Equal(t, newDOB.Month(), dobTime.Month())
	require.Equal(t, newDOB.Day(), dobTime.Day())
	require.True(t, weightLb.Valid)
	require.Equal(t, 150.5, weightLb.Float64)

	t.Logf("UpdateResident with PHI succeeded, resident_id: %s", residentID)
}

// TestUpdateResident_WithContacts 测试更新 Contacts 数据
func TestUpdateResident_WithContacts(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建测试用户（Admin）
	userID := "00000000-0000-0000-0000-000000000982"
	createTestUserForCreateResident(t, db, tenantID, userID, "testadmin-update-contacts", "password123", "Admin", "")
	createTestPermissionForCreateResident(t, db, "Admin")

	// 创建 service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	logger := getTestLoggerForResident()
	service := NewResidentService(residentsRepo, db, nil, logger)

	ctx := context.Background()

	// 1. 先创建一个住户（带联系人）
	admissionDate := time.Now().Unix()
	createReq := CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		InherentAttributes: &CreateResidentInherentAttributes{
			ResidentAccount: "testresident-update-contacts001",
			Nickname:        "Test Resident Update Contacts 001",
			Status:          "active",
			AdmissionDate:   &admissionDate,
			Contacts: []*CreateResidentContactRequest{
			{
				Slot:             "A",
				IsEnabled:        true,
				ContactFirstName: "Jane",
				ContactLastName:  "Doe",
				ContactPhone:     "555-1234",
				ContactEmail:     "jane@example.com",
			},
		},
		},
		UnitRelation: &CreateResidentUnitRelation{
			UnitID: unitID,
		},
	}

	createResp, err := service.CreateResident(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	residentID := createResp.ResidentID

	// 2. 更新联系人数据
	alertWindow := json.RawMessage(`{"start": "08:00", "end": "20:00"}`)
	updateReq := UpdateResidentRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		InherentAttributes: &UpdateResidentInherentAttributes{
			Contacts: []*UpdateResidentContactRequest{
				{
					Slot: "A",
					ContactFirstName: &domain.UpdateString{
						Action: domain.UpdateActionUpdate,
						Value:  "Jane Updated",
					},
					ContactPhone: &domain.UpdateString{
						Action: domain.UpdateActionUpdate,
						Value:  "555-5678",
					},
					ReceiveSMS: &domain.UpdateBool{
						Action: domain.UpdateActionUpdate,
						Value:  true,
					},
					ReceiveEmail: &domain.UpdateBool{
						Action: domain.UpdateActionUpdate,
						Value:  true,
					},
					AlertTimeWindow: &domain.UpdateJSON{
						Action: domain.UpdateActionUpdate,
						Value:  alertWindow,
					},
				},
			},
		},
	}

	updateResp, err := service.UpdateResident(ctx, updateReq)
	require.NoError(t, err)
	require.NotNil(t, updateResp)
	require.True(t, updateResp.Success)

	// 3. 验证联系人更新成功
	var firstName, phone string
	var receiveSMS, receiveEmail bool
	var alertWindowDB sql.NullString
	err = db.QueryRow(
		`SELECT contact_first_name, contact_phone, receive_sms, receive_email, alert_time_window::text
		 FROM resident_contacts 
		 WHERE resident_id = $1 AND slot = 'A'`,
		residentID,
	).Scan(&firstName, &phone, &receiveSMS, &receiveEmail, &alertWindowDB)
	require.NoError(t, err)
	require.Equal(t, "Jane Updated", firstName)
	require.Equal(t, "555-5678", phone)
	require.Equal(t, true, receiveSMS)
	require.Equal(t, true, receiveEmail)
	require.True(t, alertWindowDB.Valid)

	t.Logf("UpdateResident with Contacts succeeded, resident_id: %s", residentID)
}

// TestUpdateResident_WithCaregivers 测试更新 Caregivers 数据
func TestUpdateResident_WithCaregivers(t *testing.T) {
	db := setupTestDBForResident(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID := createTestTenantAndUnitForResident(t, db)
	defer cleanupTestDataForResident(t, db, tenantID)

	// 创建测试用户（Admin）
	userID := "00000000-0000-0000-0000-000000000983"
	createTestUserForCreateResident(t, db, tenantID, userID, "testadmin-update-caregivers", "password123", "Admin", "")
	createTestPermissionForCreateResident(t, db, "Admin")

	// 创建测试护理人员用户
	caregiverID1 := "00000000-0000-0000-0000-000000000984"
	caregiverID2 := "00000000-0000-0000-0000-000000000985"
	createTestUserForCreateResident(t, db, tenantID, caregiverID1, "caregiver1", "password123", "Caregiver", "")
	createTestUserForCreateResident(t, db, tenantID, caregiverID2, "caregiver2", "password123", "Caregiver", "")

	// 创建 service
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	logger := getTestLoggerForResident()
	service := NewResidentService(residentsRepo, db, nil, logger)

	ctx := context.Background()

	// 1. 先创建一个住户
	admissionDate := time.Now().Unix()
	createReq := CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		InherentAttributes: &CreateResidentInherentAttributes{
			ResidentAccount: "testresident-update-caregivers001",
			Nickname:        "Test Resident Update Caregivers 001",
			Status:          "active",
			AdmissionDate:   &admissionDate,
		},
		UnitRelation: &CreateResidentUnitRelation{
			UnitID: unitID,
		},
	}

	createResp, err := service.CreateResident(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	residentID := createResp.ResidentID

	// 2. 更新护理人员关联
	userListJSON := json.RawMessage(`["` + caregiverID1 + `", "` + caregiverID2 + `"]`)
	groupListJSON := json.RawMessage(`["group1", "group2"]`)
	updateReq := UpdateResidentRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   userID,
		CurrentUserRole: "Admin",
		CaregiverRelation: &UpdateResidentCaregiverRelation{
			UserList: &domain.UpdateJSON{
				Action: domain.UpdateActionUpdate,
				Value:  userListJSON,
			},
			GroupList: &domain.UpdateJSON{
				Action: domain.UpdateActionUpdate,
				Value:  groupListJSON,
			},
		},
	}

	updateResp, err := service.UpdateResident(ctx, updateReq)
	require.NoError(t, err)
	require.NotNil(t, updateResp)
	require.True(t, updateResp.Success)

	// 3. 验证护理人员关联更新成功
	var userListDB, groupListDB sql.NullString
	err = db.QueryRow(
		`SELECT user_list::text, group_list::text
		 FROM resident_caregivers 
		 WHERE resident_id = $1`,
		residentID,
	).Scan(&userListDB, &groupListDB)
	require.NoError(t, err)
	require.True(t, userListDB.Valid)
	require.True(t, groupListDB.Valid)
	require.Contains(t, userListDB.String, caregiverID1)
	require.Contains(t, userListDB.String, caregiverID2)
	require.Contains(t, groupListDB.String, "group1")
	require.Contains(t, groupListDB.String, "group2")

	t.Logf("UpdateResident with Caregivers succeeded, resident_id: %s", residentID)
}
