package aggregator

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"owl-common/card"
)

// MockCardRepository is a mock implementation of CardRepository
type MockCardRepository struct {
	mock.Mock
}

func (m *MockCardRepository) GetUnitInfo(tenantID, unitID string) (*card.UnitInfo, error) {
	args := m.Called(tenantID, unitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*card.UnitInfo), args.Error(1)
}

func (m *MockCardRepository) GetActiveBedsByUnit(tenantID, unitID string) ([]card.ActiveBedInfo, error) {
	args := m.Called(tenantID, unitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]card.ActiveBedInfo), args.Error(1)
}

func (m *MockCardRepository) GetDevicesByBed(tenantID, bedID string) ([]card.DeviceInfo, error) {
	args := m.Called(tenantID, bedID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]card.DeviceInfo), args.Error(1)
}

func (m *MockCardRepository) GetUnboundDevicesByUnit(tenantID, unitID string) ([]card.DeviceInfo, error) {
	args := m.Called(tenantID, unitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]card.DeviceInfo), args.Error(1)
}

func (m *MockCardRepository) GetResidentByBed(tenantID, bedID string) (*card.ResidentInfo, error) {
	args := m.Called(tenantID, bedID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*card.ResidentInfo), args.Error(1)
}

func (m *MockCardRepository) GetResidentsByUnit(tenantID, unitID string) ([]card.ResidentInfo, error) {
	args := m.Called(tenantID, unitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]card.ResidentInfo), args.Error(1)
}

func (m *MockCardRepository) GetCardsByUnit(tenantID, unitID string) ([]card.CardWithContent, error) {
	args := m.Called(tenantID, unitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]card.CardWithContent), args.Error(1)
}

func (m *MockCardRepository) DeleteCard(tenantID, cardID string) error {
	args := m.Called(tenantID, cardID)
	return args.Error(0)
}

func (m *MockCardRepository) DeleteCardsByUnit(tenantID, unitID string) error {
	args := m.Called(tenantID, unitID)
	return args.Error(0)
}

func (m *MockCardRepository) CreateCard(
	tenantID, cardType string,
	bedID *string, unitID, cardName, cardAddress string,
	residentID *string,
	devicesJSON, residentsJSON []byte,
) (string, error) {
	args := m.Called(tenantID, cardType, bedID, unitID, cardName, cardAddress,
		residentID, devicesJSON, residentsJSON)
	return args.String(0), args.Error(1)
}

func (m *MockCardRepository) GetAllUnits(tenantID string) ([]string, error) {
	args := m.Called(tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCardRepository) GetUnitIDByBedID(tenantID, bedID string) (string, error) {
	args := m.Called(tenantID, bedID)
	return args.String(0), args.Error(1)
}

func (m *MockCardRepository) CountCardsByTenant(tenantID string) (int, error) {
	args := m.Called(tenantID)
	return args.Int(0), args.Error(1)
}

func (m *MockCardRepository) UpdateCard(
	tenantID, cardID string,
	cardType string,
	bedID *string, unitID, cardName, cardAddress string,
	residentID *string,
	devicesJSON, residentsJSON []byte,
) error {
	args := m.Called(tenantID, cardID, cardType, bedID, unitID, cardName, cardAddress,
		residentID, devicesJSON, residentsJSON)
	return args.Error(0)
}

func setupCardCreator() (*card.CardCreator, *MockCardRepository) {
	mockRepo := new(MockCardRepository)
	logger := zap.NewNop()
	creator := card.NewCardCreator(mockRepo, logger)
	return creator, mockRepo
}

func TestCreateCardsForUnit_ScenarioA_SingleActiveBed(t *testing.T) {
	creator, mockRepo := setupCardCreator()

	tenantID := "tenant-123"
	unitID := "unit-456"
	bedID := "bed-1"

	// Prepare test data
	unitInfo := &card.UnitInfo{
		UnitID:       unitID,
		UnitName:     "E203",
		BranchName:   "BranchA",
		Building:     "MainBuilding",
		IsPublic:     false,
		IsSharedUnit: false,
		UnitType:     "Institutional",
	}

	activeBeds := []card.ActiveBedInfo{
		{
			BedID:            bedID,
			UnitID:           unitID,
			BoundDeviceCount: 2,
			ResidentID:       stringPtr("resident-1"),
			RoomID:           "room-1",
		},
	}

	bedName := "BedA"
	bedDevices := []card.DeviceInfo{
		{
			DeviceID:          "device-1",
			DeviceName:        "Radar01",
			DeviceType:        "Radar",
			DeviceModel:       "Model-A",
			BoundBedID:        &bedID,
			BedName:           &bedName,
			BoundRoomID:       nil,
			RoomName:          nil,
			UnitID:            unitID,
			MonitoringEnabled: true,
		},
	}

	roomID := "room-1"
	roomName := "Room1"
	unboundDevices := []card.DeviceInfo{
		{
			DeviceID:          "device-2",
			DeviceName:        "SleepPad01",
			DeviceType:        "SleepPad",
			DeviceModel:       "Model-B",
			BoundBedID:        nil,
			BedName:           nil,
			BoundRoomID:       &roomID,
			RoomName:          &roomName,
			UnitID:            unitID,
			MonitoringEnabled: true,
		},
	}

	resident := &card.ResidentInfo{
		ResidentID: "resident-1",
		Nickname:   "Smith",
		UnitID:     &unitID,
		BedID:      &bedID,
	}

	// Setup mock expectations
	mockRepo.On("GetUnitInfo", tenantID, unitID).Return(unitInfo, nil)
	mockRepo.On("GetActiveBedsByUnit", tenantID, unitID).Return(activeBeds, nil)
	mockRepo.On("DeleteCardsByUnit", tenantID, unitID).Return(nil)
	mockRepo.On("GetDevicesByBed", tenantID, bedID).Return(bedDevices, nil)
	mockRepo.On("GetUnboundDevicesByUnit", tenantID, unitID).Return(unboundDevices, nil)
	mockRepo.On("GetResidentByBed", tenantID, bedID).Return(resident, nil)
	mockRepo.On("CreateCard",
		tenantID, "ActiveBed", &bedID, unitID,
		mock.AnythingOfType("string"), // cardName
		mock.AnythingOfType("string"), // cardAddress
		&resident.ResidentID,
		mock.AnythingOfType("[]uint8"), // devicesJSON
		mock.AnythingOfType("[]uint8"), // residentsJSON
	).Return("card-123", nil)

	// Execute test
	_, err := creator.CreateCardsForUnit(tenantID, unitID)

	// Verify results
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateCardsForUnit_ScenarioB_MultipleActiveBeds(t *testing.T) {
	creator, mockRepo := setupCardCreator()

	tenantID := "tenant-123"
	unitID := "unit-456"
	bedID1 := "bed-1"
	bedID2 := "bed-2"

	// Prepare test data
	unitInfo := &card.UnitInfo{
		UnitID:       unitID,
		UnitName:     "E203",
		BranchName:   "BranchA",
		Building:     "MainBuilding",
		IsPublic:     false,
		IsSharedUnit: false,
		UnitType:     "Institutional",
	}

	activeBeds := []card.ActiveBedInfo{
		{
			BedID:            bedID1,
			UnitID:           unitID,
			BoundDeviceCount: 1,
			ResidentID:       stringPtr("resident-1"),
			RoomID:           "room-1",
		},
		{
			BedID:            bedID2,
			UnitID:           unitID,
			BoundDeviceCount: 1,
			ResidentID:       stringPtr("resident-2"),
			RoomID:           "room-1",
		},
	}

	bed1Name := "BedA"
	bed1Devices := []card.DeviceInfo{
		{DeviceID: "device-1", DeviceName: "Radar01", DeviceType: "Radar", BoundBedID: &bedID1, BedName: &bed1Name, BoundRoomID: nil, RoomName: nil, UnitID: unitID, MonitoringEnabled: true},
	}

	bed2Name := "BedB"
	bed2Devices := []card.DeviceInfo{
		{DeviceID: "device-2", DeviceName: "Radar02", DeviceType: "Radar", BoundBedID: &bedID2, BedName: &bed2Name, BoundRoomID: nil, RoomName: nil, UnitID: unitID, MonitoringEnabled: true},
	}

	roomID := "room-1"
	roomName := "Room1"
	unboundDevices := []card.DeviceInfo{
		{DeviceID: "device-3", DeviceName: "SleepPad01", DeviceType: "SleepPad", BoundBedID: nil, BedName: nil, BoundRoomID: &roomID, RoomName: &roomName, UnitID: unitID, MonitoringEnabled: true},
	}

	resident1 := &card.ResidentInfo{ResidentID: "resident-1", Nickname: "Smith", BedID: &bedID1}
	resident2 := &card.ResidentInfo{ResidentID: "resident-2", Nickname: "Jones", BedID: &bedID2}

	unitResidents := []card.ResidentInfo{*resident1, *resident2}

	// Setup mock expectations
	mockRepo.On("GetUnitInfo", tenantID, unitID).Return(unitInfo, nil)
	mockRepo.On("GetActiveBedsByUnit", tenantID, unitID).Return(activeBeds, nil)
	mockRepo.On("DeleteCardsByUnit", tenantID, unitID).Return(nil)

	// Create ActiveBed card for each bed
	mockRepo.On("GetDevicesByBed", tenantID, bedID1).Return(bed1Devices, nil)
	mockRepo.On("GetResidentByBed", tenantID, bedID1).Return(resident1, nil)
	mockRepo.On("CreateCard",
		tenantID, "ActiveBed", &bedID1, unitID,
		mock.AnythingOfType("string"), mock.AnythingOfType("string"),
		&resident1.ResidentID,
		mock.AnythingOfType("[]uint8"), mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"),
	).Return("card-1", nil)

	mockRepo.On("GetDevicesByBed", tenantID, bedID2).Return(bed2Devices, nil)
	mockRepo.On("GetResidentByBed", tenantID, bedID2).Return(resident2, nil)
	mockRepo.On("CreateCard",
		tenantID, "ActiveBed", &bedID2, unitID,
		mock.AnythingOfType("string"), mock.AnythingOfType("string"),
		&resident2.ResidentID,
		mock.AnythingOfType("[]uint8"), mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"),
	).Return("card-2", nil)

	// Create UnitCard (because there are unbound devices)
	mockRepo.On("GetUnboundDevicesByUnit", tenantID, unitID).Return(unboundDevices, nil)
	mockRepo.On("GetResidentsByUnit", tenantID, unitID).Return(unitResidents, nil)
	mockRepo.On("CreateCard",
		tenantID, "Location", mock.Anything, unitID,
		mock.AnythingOfType("string"), mock.AnythingOfType("string"),
		mock.Anything,
		mock.AnythingOfType("[]uint8"), mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"),
	).Return("card-3", nil)

	// Execute test
	_, err := creator.CreateCardsForUnit(tenantID, unitID)

	// Verify results
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateCardsForUnit_ScenarioC_NoActiveBed(t *testing.T) {
	creator, mockRepo := setupCardCreator()

	tenantID := "tenant-123"
	unitID := "unit-456"

	// Prepare test data
	unitInfo := &card.UnitInfo{
		UnitID:       unitID,
		UnitName:     "E203",
		BranchName:   "BranchA",
		Building:     "MainBuilding",
		IsPublic:     false,
		IsSharedUnit: false,
		UnitType:     "Institutional",
	}

	roomID := "room-1"
	roomName := "Room1"
	unboundDevices := []card.DeviceInfo{
		{
			DeviceID:          "device-1",
			DeviceName:        "Radar01",
			DeviceType:        "Radar",
			BoundBedID:        nil,
			BedName:           nil,
			BoundRoomID:       &roomID,
			RoomName:          &roomName,
			UnitID:            unitID,
			MonitoringEnabled: true,
		},
	}

	unitResidents := []card.ResidentInfo{
		{ResidentID: "resident-1", Nickname: "Smith", UnitID: &unitID},
	}

	// Setup mock expectations
	mockRepo.On("GetUnitInfo", tenantID, unitID).Return(unitInfo, nil)
	mockRepo.On("GetActiveBedsByUnit", tenantID, unitID).Return([]card.ActiveBedInfo{}, nil)
	mockRepo.On("DeleteCardsByUnit", tenantID, unitID).Return(nil)
	mockRepo.On("GetUnboundDevicesByUnit", tenantID, unitID).Return(unboundDevices, nil)
	mockRepo.On("GetResidentsByUnit", tenantID, unitID).Return(unitResidents, nil)
	mockRepo.On("CreateCard",
		tenantID, "Location", mock.Anything, unitID,
		mock.AnythingOfType("string"), mock.AnythingOfType("string"),
		mock.Anything,
		mock.AnythingOfType("[]uint8"), mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"),
	).Return("card-123", nil)

	// Execute test
	_, err := creator.CreateCardsForUnit(tenantID, unitID)

	// Verify results
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateCardsForUnit_ScenarioC_NoUnboundDevices(t *testing.T) {
	creator, mockRepo := setupCardCreator()

	tenantID := "tenant-123"
	unitID := "unit-456"

	// Prepare test data
	unitInfo := &card.UnitInfo{
		UnitID:       unitID,
		UnitName:     "E203",
		BranchName:   "BranchA",
		Building:     "MainBuilding",
		IsPublic:     false,
		IsSharedUnit: false,
		UnitType:     "Institutional",
	}

	// Setup mock expectations (no unbound devices, should not create UnitCard)
	mockRepo.On("GetUnitInfo", tenantID, unitID).Return(unitInfo, nil)
	mockRepo.On("GetActiveBedsByUnit", tenantID, unitID).Return([]card.ActiveBedInfo{}, nil)
	mockRepo.On("DeleteCardsByUnit", tenantID, unitID).Return(nil)
	mockRepo.On("GetUnboundDevicesByUnit", tenantID, unitID).Return([]card.DeviceInfo{}, nil)

	// Execute test
	_, err := creator.CreateCardsForUnit(tenantID, unitID)

	// Verify results (should not create any cards, should not error)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	// Verify CreateCard was not called
	mockRepo.AssertNotCalled(t, "CreateCard")
}

func TestCreateCardsForUnit_Error_GetUnitInfoFailed(t *testing.T) {
	creator, mockRepo := setupCardCreator()

	tenantID := "tenant-123"
	unitID := "unit-456"

	// Setup mock expectations (GetUnitInfo fails)
	mockRepo.On("GetUnitInfo", tenantID, unitID).Return(nil, errors.New("database error"))

	// Execute test
	_, err := creator.CreateCardsForUnit(tenantID, unitID)

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get unit info")
	mockRepo.AssertExpectations(t)
}

func TestCreateCardsForUnit_Error_GetActiveBedsFailed(t *testing.T) {
	creator, mockRepo := setupCardCreator()

	tenantID := "tenant-123"
	unitID := "unit-456"

	unitInfo := &card.UnitInfo{
		UnitID:       unitID,
		UnitName:     "E203",
		BranchName:   "BranchA",
		Building:     "MainBuilding",
		IsPublic:     false,
		IsSharedUnit: false,
		UnitType:     "Institutional",
	}

	// Setup mock expectations
	mockRepo.On("GetUnitInfo", tenantID, unitID).Return(unitInfo, nil)
	mockRepo.On("GetActiveBedsByUnit", tenantID, unitID).Return(nil, errors.New("database error"))

	// Execute test
	_, err := creator.CreateCardsForUnit(tenantID, unitID)

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get active beds")
	mockRepo.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}
