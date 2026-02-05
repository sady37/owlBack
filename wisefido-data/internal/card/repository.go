package card

import commoncard "owl-common/card"

// RepositoryInterface 卡片仓储接口（建卡/同步用），由 PostgresCardRepository 实现
type RepositoryInterface interface {
	GetUnitInfo(tenantID, unitID string) (*commoncard.UnitInfo, error)
	GetActiveBedsByUnit(tenantID, unitID string) ([]commoncard.ActiveBedRow, error)
	GetDevicesByBed(tenantID, bedID string) ([]commoncard.DeviceInfo, error)
	GetUnboundDevicesByUnit(tenantID, unitID string) ([]commoncard.DeviceInfo, error)
	GetResidentByBed(tenantID, bedID string) (*commoncard.ResidentInfo, error)
	GetResidentsByUnit(tenantID, unitID string) ([]commoncard.ResidentInfo, error)
	GetCardsByUnit(tenantID, unitID string) ([]commoncard.CardWithContent, error)
	DeleteCard(tenantID, cardID string) error
	DeleteCardsByUnit(tenantID, unitID string) error
	CreateCard(
		tenantID, cardType string,
		bedID *string, unitID, cardName, cardAddress, timezone string,
		residentID *string,
		devicesJSON, residentsJSON []byte,
	) (string, error)
	GetAllUnits(tenantID string) ([]string, error)
	GetUnitIDByBedID(tenantID, bedID string) (string, error)
	CountCardsByTenant(tenantID string) (int, error)
	UpdateCard(
		tenantID, cardID string,
		cardType string,
		bedID *string, unitID, cardName, cardAddress, timezone string,
		residentID *string,
		devicesJSON, residentsJSON []byte,
	) error
}

// CardUpdateStats 同步统计（重导出 owl-common/card）
type CardUpdateStats = commoncard.CardUpdateStats
