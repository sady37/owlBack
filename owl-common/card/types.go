package card

// ActiveBedInfo ActiveBed information
type ActiveBedInfo struct {
	BedID            string
	UnitID           string
	BoundDeviceCount int
	ResidentID       *string
	RoomID           string
}

// UnitInfo Unit information
type UnitInfo struct {
	UnitID       string
	UnitName     string
	BranchName   string
	Building     string
	IsPublic     bool // 对应数据库字段 is_public
	IsSharedUnit bool // 对应数据库字段 is_shared_unit
	UnitType     string
}

// DeviceInfo device information
type DeviceInfo struct {
	DeviceID          string
	DeviceUID         string  // devices.device_uid，与 cards.devices JSON、card-overview 对齐
	DeviceCode        string  // device_store.device_code，与 card-overview 对齐
	DeviceName        string
	DeviceType        string
	DeviceModel       string
	BoundBedID        *string
	BedName           *string // Bed name (if bound to bed)
	BoundRoomID       *string // Room ID where device is bound (if bound to room)
	RoomName          *string // Room name (if bound to room)
	UnitID            string
	MonitoringEnabled bool
}

// ResidentInfo resident information
type ResidentInfo struct {
	ResidentID string
	Nickname   string
	UnitID     *string
	BedID      *string
}

// CardWithContent card with devices and residents JSONB content (for comparison)
type CardWithContent struct {
	CardID        string
	CardType      string
	BedID         *string
	UnitID        string
	CardName      string
	CardAddress   string
	ResidentID    *string
	DevicesJSON   []byte
	ResidentsJSON []byte
}

// ExpectedCard represents an expected card (for comparison, without card_id)
type ExpectedCard struct {
	CardType      string
	BedID         *string
	UnitID        string
	CardName      string
	CardAddress   string
	ResidentID    *string
	DevicesJSON   []byte
	ResidentsJSON []byte
}

// CardUpdateStats statistics for card updates
type CardUpdateStats struct {
	ExistingCount  int // Number of existing cards before update
	DeletedCount   int // Number of cards deleted
	CreatedCount   int // Number of cards created
	UpdatedCount   int // Number of cards updated (deleted + created)
	UnchangedCount int // Number of cards that remained unchanged
}

