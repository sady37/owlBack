package aggregator

import (
	"errors"
	"testing"
	"wisefido-card-aggregator/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockCardRepository is a mock implementation of CardRepository
type MockCardRepository struct {
	mock.Mock
}

func (m *MockCardRepository) GetUnitInfo(tenantID, unitID string) (*repository.UnitInfo, error) {
	args := m.Called(tenantID, unitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.UnitInfo), args.Error(1)
}

func (m *MockCardRepository) GetActiveBedsByUnit(tenantID, unitID string) ([]repository.ActiveBedInfo, error) {
	args := m.Called(tenantID, unitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.ActiveBedInfo), args.Error(1)
}

func (m *MockCardRepository) GetDevicesByBed(tenantID, bedID string) ([]repository.DeviceInfo, error) {
	args := m.Called(tenantID, bedID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.DeviceInfo), args.Error(1)
}

func (m *MockCardRepository) GetUnboundDevicesByUnit(tenantID, unitID string) ([]repository.DeviceInfo, error) {
	args := m.Called(tenantID, unitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.DeviceInfo), args.Error(1)
}

func (m *MockCardRepository) GetResidentByBed(tenantID, bedID string) (*repository.ResidentInfo, error) {
	args := m.Called(tenantID, bedID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ResidentInfo), args.Error(1)
}

func (m *MockCardRepository) GetResidentsByUnit(tenantID, unitID string) ([]repository.ResidentInfo, error) {
	args := m.Called(tenantID, unitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.ResidentInfo), args.Error(1)
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

func setupCardCreator() (*CardCreator, *MockCardRepository) {
	mockRepo := new(MockCardRepository)
	logger := zap.NewNop()
	creator := NewCardCreator(mockRepo, logger)
	return creator, mockRepo
}

func TestCreateCardsForUnit_ScenarioA_SingleActiveBed(t *testing.T) {
	creator, mockRepo := setupCardCreator()

	tenantID := "tenant-123"
	unitID := "unit-456"
	bedID := "bed-1"

	// Prepare test data
	unitInfo := &repository.UnitInfo{
		UnitID:            unitID,
		UnitName:          "E203",
		BranchName:        "BranchA",
		Building:          "MainBuilding",
		IsPublicSpace:     false,
		IsMultiPersonRoom: false,
		UnitType:          "Institutional",
		GroupList:         []byte(`["tag1"]`),
		UserList:          []byte(`["user-1"]`),
	}

	activeBeds := []repository.ActiveBedInfo{
		{
			BedID:            bedID,
			UnitID:           unitID,
			BoundDeviceCount: 2,
			ResidentID:       stringPtr("resident-1"),
			RoomID:           "room-1",
		},
	}

	bedName := "BedA"
	bedDevices := []repository.DeviceInfo{
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
	unboundDevices := []repository.DeviceInfo{
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

	resident := &repository.ResidentInfo{
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
	err := creator.CreateCardsForUnit(tenantID, unitID)

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
	unitInfo := &repository.UnitInfo{
		UnitID:            unitID,
		UnitName:          "E203",
		BranchName:        "BranchA",
		Building:          "MainBuilding",
		IsPublicSpace:     false,
		IsMultiPersonRoom: false,
		UnitType:          "Institutional",
		GroupList:         []byte(`[]`),
		UserList:          []byte(`[]`),
	}

	activeBeds := []repository.ActiveBedInfo{
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
	bed1Devices := []repository.DeviceInfo{
		{DeviceID: "device-1", DeviceName: "Radar01", DeviceType: "Radar", BoundBedID: &bedID1, BedName: &bed1Name, BoundRoomID: nil, RoomName: nil, UnitID: unitID, MonitoringEnabled: true},
	}

	bed2Name := "BedB"
	bed2Devices := []repository.DeviceInfo{
		{DeviceID: "device-2", DeviceName: "Radar02", DeviceType: "Radar", BoundBedID: &bedID2, BedName: &bed2Name, BoundRoomID: nil, RoomName: nil, UnitID: unitID, MonitoringEnabled: true},
	}

	roomID := "room-1"
	roomName := "Room1"
	unboundDevices := []repository.DeviceInfo{
		{DeviceID: "device-3", DeviceName: "SleepPad01", DeviceType: "SleepPad", BoundBedID: nil, BedName: nil, BoundRoomID: &roomID, RoomName: &roomName, UnitID: unitID, MonitoringEnabled: true},
	}

	resident1 := &repository.ResidentInfo{ResidentID: "resident-1", Nickname: "Smith", BedID: &bedID1}
	resident2 := &repository.ResidentInfo{ResidentID: "resident-2", Nickname: "Jones", BedID: &bedID2}

	unitResidents := []repository.ResidentInfo{*resident1, *resident2}

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
	err := creator.CreateCardsForUnit(tenantID, unitID)

	// Verify results
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateCardsForUnit_ScenarioC_NoActiveBed(t *testing.T) {
	creator, mockRepo := setupCardCreator()

	tenantID := "tenant-123"
	unitID := "unit-456"

	// Prepare test data
	unitInfo := &repository.UnitInfo{
		UnitID:            unitID,
		UnitName:          "E203",
		BranchName:        "BranchA",
		Building:          "MainBuilding",
		IsPublicSpace:     false,
		IsMultiPersonRoom: false,
		UnitType:          "Institutional",
		GroupList:         []byte(`[]`),
		UserList:          []byte(`[]`),
	}

	roomID := "room-1"
	roomName := "Room1"
	unboundDevices := []repository.DeviceInfo{
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

	unitResidents := []repository.ResidentInfo{
		{ResidentID: "resident-1", Nickname: "Smith", UnitID: &unitID},
	}

	// Setup mock expectations
	mockRepo.On("GetUnitInfo", tenantID, unitID).Return(unitInfo, nil)
	mockRepo.On("GetActiveBedsByUnit", tenantID, unitID).Return([]repository.ActiveBedInfo{}, nil)
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
	err := creator.CreateCardsForUnit(tenantID, unitID)

	// Verify results
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateCardsForUnit_ScenarioC_NoUnboundDevices(t *testing.T) {
	creator, mockRepo := setupCardCreator()

	tenantID := "tenant-123"
	unitID := "unit-456"

	// Prepare test data
	unitInfo := &repository.UnitInfo{
		UnitID:            unitID,
		UnitName:          "E203",
		BranchName:        "BranchA",
		Building:          "MainBuilding",
		IsPublicSpace:     false,
		IsMultiPersonRoom: false,
		UnitType:          "Institutional",
		GroupList:         []byte(`[]`),
		UserList:          []byte(`[]`),
	}

	// Setup mock expectations (no unbound devices, should not create UnitCard)
	mockRepo.On("GetUnitInfo", tenantID, unitID).Return(unitInfo, nil)
	mockRepo.On("GetActiveBedsByUnit", tenantID, unitID).Return([]repository.ActiveBedInfo{}, nil)
	mockRepo.On("DeleteCardsByUnit", tenantID, unitID).Return(nil)
	mockRepo.On("GetUnboundDevicesByUnit", tenantID, unitID).Return([]repository.DeviceInfo{}, nil)

	// Execute test
	err := creator.CreateCardsForUnit(tenantID, unitID)

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
	err := creator.CreateCardsForUnit(tenantID, unitID)

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get unit info")
	mockRepo.AssertExpectations(t)
}

func TestCreateCardsForUnit_Error_GetActiveBedsFailed(t *testing.T) {
	creator, mockRepo := setupCardCreator()

	tenantID := "tenant-123"
	unitID := "unit-456"

	unitInfo := &repository.UnitInfo{
		UnitID:    unitID,
		UnitName:  "E203",
		GroupList: []byte(`[]`),
		UserList:  []byte(`[]`),
	}

	// Setup mock expectations
	mockRepo.On("GetUnitInfo", tenantID, unitID).Return(unitInfo, nil)
	mockRepo.On("GetActiveBedsByUnit", tenantID, unitID).Return(nil, errors.New("database error"))

	// Execute test
	err := creator.CreateCardsForUnit(tenantID, unitID)

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get active beds")
	mockRepo.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}
