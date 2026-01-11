package card

// RepositoryInterface defines card repository interface
// This interface abstracts the database operations for card creation and management
type RepositoryInterface interface {
	GetUnitInfo(tenantID, unitID string) (*UnitInfo, error)
	GetActiveBedsByUnit(tenantID, unitID string) ([]ActiveBedInfo, error)
	GetDevicesByBed(tenantID, bedID string) ([]DeviceInfo, error)
	GetUnboundDevicesByUnit(tenantID, unitID string) ([]DeviceInfo, error)
	GetResidentByBed(tenantID, bedID string) (*ResidentInfo, error)
	GetResidentsByUnit(tenantID, unitID string) ([]ResidentInfo, error)
	GetCardsByUnit(tenantID, unitID string) ([]CardWithContent, error)
	DeleteCard(tenantID, cardID string) error
	DeleteCardsByUnit(tenantID, unitID string) error
	CreateCard(
		tenantID, cardType string,
		bedID *string, unitID, cardName, cardAddress string,
		residentID *string,
		devicesJSON, residentsJSON []byte,
	) (string, error)
	GetAllUnits(tenantID string) ([]string, error)
	GetUnitIDByBedID(tenantID, bedID string) (string, error)
	CountCardsByTenant(tenantID string) (int, error)
	UpdateCard(
		tenantID, cardID string,
		cardType string,
		bedID *string, unitID, cardName, cardAddress string,
		residentID *string,
		devicesJSON, residentsJSON []byte,
	) error
}
