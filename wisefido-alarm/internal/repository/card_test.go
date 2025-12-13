package repository

import (
	"database/sql"
	"encoding/json"
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

func TestGetCardByID_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	tenantID := "tenant-123"
	cardID := "card-456"

	rows := sqlmock.NewRows([]string{
		"card_id", "tenant_id", "card_type", "bed_id", "unit_id", "card_name", "room_id",
	}).AddRow(
		cardID, tenantID, "ActiveBed", "bed-789", "unit-101", "Bed 1", "room-202",
	)

	mock.ExpectQuery(`SELECT`).
		WithArgs(cardID, tenantID).
		WillReturnRows(rows)

	card, err := repo.GetCardByID(tenantID, cardID)

	require.NoError(t, err)
	assert.NotNil(t, card)
	assert.Equal(t, cardID, card.CardID)
	assert.Equal(t, tenantID, card.TenantID)
	assert.Equal(t, "ActiveBed", card.CardType)
	assert.NotNil(t, card.BedID)
	assert.Equal(t, "bed-789", *card.BedID)
	assert.Equal(t, "unit-101", card.UnitID)
	assert.Equal(t, "Bed 1", card.CardName)
	assert.NotNil(t, card.RoomID)
	assert.Equal(t, "room-202", *card.RoomID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCardByID_NotFound(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	tenantID := "tenant-123"
	cardID := "card-456"

	mock.ExpectQuery(`SELECT`).
		WithArgs(cardID, tenantID).
		WillReturnError(sql.ErrNoRows)

	card, err := repo.GetCardByID(tenantID, cardID)

	assert.Error(t, err)
	assert.Nil(t, card)
	assert.Contains(t, err.Error(), "card not found")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCardDevices_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	cardID := "card-456"

	devicesJSON := []map[string]interface{}{
		{
			"device_id":    "device-1",
			"device_name":  "Radar01",
			"device_type":  "Radar",
			"device_model": "Model-A",
			"bed_id":       "bed-789",
			"bed_name":     "Bed 1",
			"room_id":      nil,
			"room_name":    nil,
			"unit_id":      "unit-101",
		},
		{
			"device_id":    "device-2",
			"device_name":  "Sleepace01",
			"device_type":  "Sleepace",
			"device_model": "Model-B",
			"bed_id":       "bed-789",
			"bed_name":     "Bed 1",
			"room_id":      nil,
			"room_name":    nil,
			"unit_id":      "unit-101",
		},
	}

	devicesJSONBytes, err := json.Marshal(devicesJSON)
	require.NoError(t, err)

	// PostgreSQL JSONB 字段在 sqlmock 中需要作为 []byte 返回
	rows := sqlmock.NewRows([]string{"devices"}).AddRow(devicesJSONBytes)

	mock.ExpectQuery(`SELECT devices`).
		WithArgs(cardID).
		WillReturnRows(rows)

	devices, err := repo.GetCardDevices(cardID)

	require.NoError(t, err)
	assert.Len(t, devices, 2)
	assert.Equal(t, "device-1", devices[0].DeviceID)
	assert.Equal(t, "Radar01", devices[0].DeviceName)
	assert.Equal(t, "Radar", devices[0].DeviceType)
	assert.NotNil(t, devices[0].BedID)
	assert.Equal(t, "bed-789", *devices[0].BedID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllCards_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	tenantID := "tenant-123"

	rows := sqlmock.NewRows([]string{
		"card_id", "tenant_id", "card_type", "bed_id", "unit_id", "card_name", "room_id",
	}).
		AddRow("card-1", tenantID, "ActiveBed", "bed-1", "unit-1", "Bed 1", "room-1").
		AddRow("card-2", tenantID, "Location", nil, "unit-1", "Unit 1", "room-2")

	mock.ExpectQuery(`SELECT`).
		WithArgs(tenantID).
		WillReturnRows(rows)

	cards, err := repo.GetAllCards(tenantID)

	require.NoError(t, err)
	assert.Len(t, cards, 2)
	assert.Equal(t, "card-1", cards[0].CardID)
	assert.Equal(t, "ActiveBed", cards[0].CardType)
	assert.Equal(t, "card-2", cards[1].CardID)
	assert.Equal(t, "Location", cards[1].CardType)
	assert.Nil(t, cards[1].BedID)

	require.NoError(t, mock.ExpectationsWereMet())
}
