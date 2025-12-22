package service

import (
	"testing"
	"wisefido-card-aggregator/internal/repository"

	"github.com/stretchr/testify/mock"
)

// MockCardRepository 是 CardRepository 的 mock 实现
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

// GetAllUnits 获取所有单元ID（用于 Service 层测试）
func (m *MockCardRepository) GetAllUnits(tenantID string) ([]string, error) {
	args := m.Called(tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func TestAggregatorService_Start_Stop(t *testing.T) {
	// Service 层当前设计直接创建数据库连接，难以进行单元测试
	// 需要重构以支持依赖注入才能进行完整的单元测试
	t.Skip("Service 层需要重构以支持依赖注入才能进行单元测试")
}

func TestAggregatorService_CreateAllCards_Success(t *testing.T) {
	// 这个测试需要重构 Service 层以支持依赖注入
	// 当前 Service 层直接创建了 CardRepository，难以 mock
	t.Skip("需要重构 Service 层以支持依赖注入")
}

func TestAggregatorService_CreateAllCards_NoTenantID(t *testing.T) {
	// 由于 NewAggregatorService 需要数据库连接，我们无法直接测试
	// 需要重构以支持依赖注入
	t.Skip("需要重构 Service 层以支持依赖注入")
}

// 由于 Service 层当前的设计（直接创建数据库连接），
// 完整的单元测试需要重构 Service 层以支持依赖注入。
// 建议的改进：
// 1. 将 CardRepository 和 CardCreator 作为依赖注入
// 2. 或者创建接口，允许在测试中注入 mock 实现
