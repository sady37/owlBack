package repository

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *CardRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	logger := zap.NewNop()
	repo := NewCardRepository(db, logger)

	return db, mock, repo
}

func TestGetActiveBedsByUnit_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	tenantID := "tenant-123"
	unitID := "unit-456"

	// Setup expected SQL query
	// 注意：查询已改为动态 JOIN devices 表，检查 monitoring_enabled = TRUE
	rows := sqlmock.NewRows([]string{"bed_id", "unit_id", "bound_device_count", "resident_id", "room_id"}).
		AddRow("bed-1", "unit-456", 2, "resident-1", "room-1").
		AddRow("bed-2", "unit-456", 1, nil, "room-1")

	mock.ExpectQuery(`SELECT DISTINCT`).
		WithArgs(tenantID, unitID).
		WillReturnRows(rows)

	// Execute test
	beds, err := repo.GetActiveBedsByUnit(tenantID, unitID)

	// Verify results
	require.NoError(t, err)
	assert.Len(t, beds, 2)
	assert.Equal(t, "bed-1", beds[0].BedID)
	assert.Equal(t, 2, beds[0].BoundDeviceCount)
	assert.NotNil(t, beds[0].ResidentID)
	assert.Equal(t, "resident-1", *beds[0].ResidentID)

	assert.Equal(t, "bed-2", beds[1].BedID)
	assert.Nil(t, beds[1].ResidentID)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetActiveBedsByUnit_EmptyResult(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	tenantID := "tenant-123"
	unitID := "unit-456"

	// Setup expected SQL query (empty result)
	// 注意：查询已改为动态 JOIN devices 表，检查 monitoring_enabled = TRUE
	rows := sqlmock.NewRows([]string{"bed_id", "unit_id", "bound_device_count", "resident_id", "room_id"})

	mock.ExpectQuery(`SELECT DISTINCT`).
		WithArgs(tenantID, unitID).
		WillReturnRows(rows)

	// Execute test
	beds, err := repo.GetActiveBedsByUnit(tenantID, unitID)

	// Verify results
	require.NoError(t, err)
	assert.Len(t, beds, 0)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUnitInfo_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	tenantID := "tenant-123"
	unitID := "unit-456"

	// Setup expected SQL query
	rows := sqlmock.NewRows([]string{
		"unit_id", "unit_name", "branch_name", "building",
		"is_public_space", "is_multi_person_room", "unit_type",
		"groupList", "userList",
	}).
		AddRow(
			"unit-456", "E203", "BranchA", "MainBuilding",
			false, false, "Institutional",
			`["tag1", "tag2"]`, `["user-id-1", "user-id-2"]`,
		)

	mock.ExpectQuery(`SELECT`).
		WithArgs(tenantID, unitID).
		WillReturnRows(rows)

	// Execute test
	unitInfo, err := repo.GetUnitInfo(tenantID, unitID)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, "unit-456", unitInfo.UnitID)
	assert.Equal(t, "E203", unitInfo.UnitName)
	assert.Equal(t, "BranchA", unitInfo.BranchName)
	assert.Equal(t, "MainBuilding", unitInfo.Building)
	assert.False(t, unitInfo.IsPublicSpace)
	assert.False(t, unitInfo.IsMultiPersonRoom)
	assert.Equal(t, "Institutional", unitInfo.UnitType)
	assert.Contains(t, string(unitInfo.GroupList), "tag1")
	assert.Contains(t, string(unitInfo.UserList), "user-id-1")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUnitInfo_WithNullGroupList(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	tenantID := "tenant-123"
	unitID := "unit-456"

	// Setup expected SQL query (groupList and userList are NULL)
	rows := sqlmock.NewRows([]string{
		"unit_id", "unit_name", "branch_name", "building",
		"is_public_space", "is_multi_person_room", "unit_type",
		"groupList", "userList",
	}).
		AddRow(
			"unit-456", "E203", "BranchA", "MainBuilding",
			false, false, "Institutional",
			nil, nil,
		)

	mock.ExpectQuery(`SELECT`).
		WithArgs(tenantID, unitID).
		WillReturnRows(rows)

	// Execute test
	unitInfo, err := repo.GetUnitInfo(tenantID, unitID)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, "[]", string(unitInfo.GroupList))
	assert.Equal(t, "[]", string(unitInfo.UserList))

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUnitInfo_NotFound(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	tenantID := "tenant-123"
	unitID := "unit-456"

	// Setup expected SQL query (returns sql.ErrNoRows)
	mock.ExpectQuery(`SELECT`).
		WithArgs(tenantID, unitID).
		WillReturnError(sql.ErrNoRows)

	// Execute test
	unitInfo, err := repo.GetUnitInfo(tenantID, unitID)

	// Verify results
	assert.Error(t, err)
	assert.Nil(t, unitInfo)
	assert.Contains(t, err.Error(), "unit not found")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCard_ActiveBed(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	tenantID := "tenant-123"
	cardType := "ActiveBed"
	bedID := "bed-1"
	unitID := "unit-456"
	cardName := "Smith"
	cardAddress := "BranchA-MainBuilding-E203"
	residentID := "resident-1"
	devicesJSON := []byte(`[{"device_id": "device-1"}]`)
	residentsJSON := []byte(`[{"resident_id": "resident-1"}]`)

	// Setup expected SQL insert
	rows := sqlmock.NewRows([]string{"card_id"}).
		AddRow("card-123")

	mock.ExpectQuery(`INSERT INTO cards`).
		WithArgs(
			tenantID, cardType, bedID, unitID, cardName, cardAddress,
			residentID, devicesJSON, residentsJSON,
		).
		WillReturnRows(rows)

	// Execute test
	cardID, err := repo.CreateCard(
		tenantID, cardType, &bedID, unitID, cardName, cardAddress,
		&residentID, devicesJSON, residentsJSON,
	)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, "card-123", cardID)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCard_Location(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	tenantID := "tenant-123"
	cardType := "Location"
	unitID := "unit-456"
	cardName := "201"
	cardAddress := "BranchA-MainBuilding-E203"
	devicesJSON := []byte(`[{"device_id": "device-1"}]`)
	residentsJSON := []byte(`[{"resident_id": "resident-1"}]`)

	// Setup expected SQL insert
	rows := sqlmock.NewRows([]string{"card_id"}).
		AddRow("card-456")

	mock.ExpectQuery(`INSERT INTO cards`).
		WithArgs(
			tenantID, cardType, nil, unitID, cardName, cardAddress,
			nil, devicesJSON, residentsJSON,
		).
		WillReturnRows(rows)

	// Execute test
	cardID, err := repo.CreateCard(
		tenantID, cardType, nil, unitID, cardName, cardAddress,
		nil, devicesJSON, residentsJSON,
	)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, "card-456", cardID)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCardsByUnit_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	tenantID := "tenant-123"
	unitID := "unit-456"

	// Setup expected SQL delete
	mock.ExpectExec(`DELETE FROM cards`).
		WithArgs(tenantID, unitID).
		WillReturnResult(sqlmock.NewResult(0, 3)) // Delete 3 records

	// Execute test
	err := repo.DeleteCardsByUnit(tenantID, unitID)

	// Verify results
	require.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCardsByUnit_NoRowsAffected(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	tenantID := "tenant-123"
	unitID := "unit-456"

	// Setup expected SQL delete (no records deleted)
	mock.ExpectExec(`DELETE FROM cards`).
		WithArgs(tenantID, unitID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Execute test
	err := repo.DeleteCardsByUnit(tenantID, unitID)

	// Verify results (should not error even if no records were deleted)
	require.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
