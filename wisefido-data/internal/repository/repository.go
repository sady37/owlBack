package repository

import (
	"context"
	"database/sql"
)

type Repository struct {
	DB *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// --- Units ---

type Unit struct {
	UnitID            string         `json:"unit_id"`
	TenantID          string         `json:"tenant_id"`
	BranchTag         string         `json:"branch_tag,omitempty"`
	UnitName          string         `json:"unit_name"`
	Building          string         `json:"building,omitempty"`
	Floor             string         `json:"floor,omitempty"`
	AreaTag           sql.NullString `json:"-"`
	UnitNumber        string         `json:"unit_number"`
	LayoutConfig      sql.NullString `json:"-"`
	UnitType          string         `json:"unit_type"`
	IsPublicSpace     bool           `json:"is_public_space,omitempty"`
	IsMultiPersonRoom bool           `json:"is_multi_person_room,omitempty"`
	Timezone          string         `json:"timezone,omitempty"`
}

func (u Unit) ToJSON() map[string]any {
	m := map[string]any{
		"unit_id":              u.UnitID,
		"tenant_id":            u.TenantID,
		"branch_tag":           u.BranchTag,
		"unit_name":            u.UnitName,
		"building":             u.Building,
		"floor":                u.Floor,
		"unit_number":          u.UnitNumber,
		"unit_type":            u.UnitType,
		"is_public_space":      u.IsPublicSpace,
		"is_multi_person_room": u.IsMultiPersonRoom,
		"timezone":             u.Timezone,
	}
	if u.AreaTag.Valid {
		m["area_tag"] = u.AreaTag.String
	}
	if u.LayoutConfig.Valid {
		// 前端类型是 Record<string,any>，这里直接返回 json string（真实解析可后续加）
		m["layout_config"] = jsonRawOrString(u.LayoutConfig.String)
	}
	return m
}

type Room struct {
	RoomID       string         `json:"room_id"`
	TenantID     sql.NullString `json:"-"`
	UnitID       string         `json:"unit_id"`
	RoomName     string         `json:"room_name"`
	IsDefault    bool           `json:"is_default,omitempty"`
	LayoutConfig sql.NullString `json:"-"`
}

func (r Room) ToJSON() map[string]any {
	m := map[string]any{
		"room_id":    r.RoomID,
		"unit_id":    r.UnitID,
		"room_name":  r.RoomName,
		"is_default": r.IsDefault,
	}
	if r.TenantID.Valid {
		m["tenant_id"] = r.TenantID.String
	}
	if r.LayoutConfig.Valid {
		m["layout_config"] = jsonRawOrString(r.LayoutConfig.String)
	}
	return m
}

type Bed struct {
	BedID             string         `json:"bed_id"`
	TenantID          sql.NullString `json:"-"`
	RoomID            string         `json:"room_id"`
	BedName           string         `json:"bed_name"`
	BedType           sql.NullString `json:"-"`
	MattressMaterial  sql.NullString `json:"-"`
	MattressThickness sql.NullString `json:"-"`
	BoundDeviceCount  sql.NullInt64  `json:"-"`
}

func (b Bed) ToJSON() map[string]any {
	m := map[string]any{
		"bed_id":   b.BedID,
		"room_id":  b.RoomID,
		"bed_name": b.BedName,
	}
	if b.TenantID.Valid {
		m["tenant_id"] = b.TenantID.String
	}
	if b.BedType.Valid {
		m["bed_type"] = b.BedType.String
	}
	if b.MattressMaterial.Valid {
		m["mattress_material"] = b.MattressMaterial.String
	}
	if b.MattressThickness.Valid {
		m["mattress_thickness"] = b.MattressThickness.String
	}
	if b.BoundDeviceCount.Valid {
		m["bound_device_count"] = b.BoundDeviceCount.Int64
	}
	return m
}

type UnitsRepo interface {
	ListBuildings(ctx context.Context, tenantID string, branchTag string) ([]map[string]any, error)
	ListUnits(ctx context.Context, tenantID string, filters map[string]string, page, size int) (items []Unit, total int, err error)
	GetUnit(ctx context.Context, tenantID, unitID string) (*Unit, error)
	CreateUnit(ctx context.Context, tenantID string, payload map[string]any) (*Unit, error)
	UpdateUnit(ctx context.Context, tenantID, unitID string, payload map[string]any) (*Unit, error)
	DeleteUnit(ctx context.Context, tenantID, unitID string) error

	ListRoomsWithBeds(ctx context.Context, unitID string) ([]map[string]any, error)
	CreateRoom(ctx context.Context, unitID string, payload map[string]any) (*Room, error)
	UpdateRoom(ctx context.Context, roomID string, payload map[string]any) (*Room, error)
	DeleteRoom(ctx context.Context, roomID string) error

	ListBeds(ctx context.Context, roomID string) ([]Bed, error)
	CreateBed(ctx context.Context, roomID string, payload map[string]any) (*Bed, error)
	UpdateBed(ctx context.Context, bedID string, payload map[string]any) (*Bed, error)
	DeleteBed(ctx context.Context, bedID string) error
}

// --- Devices ---

type Device struct {
	DeviceID          string         `json:"device_id"`
	TenantID          string         `json:"tenant_id"`
	DeviceStoreID     sql.NullString `json:"-"`
	DeviceName        string         `json:"device_name"`
	DeviceModel       sql.NullString `json:"-"`
	DeviceType        sql.NullString `json:"-"`
	SerialNumber      sql.NullString `json:"-"`
	UID               sql.NullString `json:"-"`
	IMEI              sql.NullString `json:"-"`
	CommMode          sql.NullString `json:"-"`
	FirmwareVersion   sql.NullString `json:"-"`
	MCUModel          sql.NullString `json:"-"`
	Status            string         `json:"status"`
	BusinessAccess    string         `json:"business_access"`
	MonitoringEnabled bool           `json:"monitoring_enabled"`
	UnitID            sql.NullString `json:"-"`
	BoundRoomID       sql.NullString `json:"-"`
	BoundBedID        sql.NullString `json:"-"`
	Metadata          sql.NullString `json:"-"`
}

func (d Device) ToJSON() map[string]any {
	m := map[string]any{
		"device_id":          d.DeviceID,
		"tenant_id":          d.TenantID,
		"device_name":        d.DeviceName,
		"status":             d.Status,
		"business_access":    d.BusinessAccess,
		"monitoring_enabled": d.MonitoringEnabled,
	}
	if d.DeviceStoreID.Valid {
		m["device_store_id"] = d.DeviceStoreID.String
	}
	if d.DeviceModel.Valid {
		m["device_model"] = d.DeviceModel.String
	}
	if d.DeviceType.Valid {
		m["device_type"] = d.DeviceType.String
	}
	if d.SerialNumber.Valid {
		m["serial_number"] = d.SerialNumber.String
	}
	if d.UID.Valid {
		m["uid"] = d.UID.String
	}
	if d.IMEI.Valid {
		m["imei"] = d.IMEI.String
	}
	if d.CommMode.Valid {
		m["comm_mode"] = d.CommMode.String
	}
	if d.FirmwareVersion.Valid {
		m["firmware_version"] = d.FirmwareVersion.String
	}
	if d.MCUModel.Valid {
		m["mcu_model"] = d.MCUModel.String
	}
	if d.UnitID.Valid {
		m["unit_id"] = d.UnitID.String
	}
	if d.BoundRoomID.Valid {
		m["bound_room_id"] = d.BoundRoomID.String
	} else {
		m["bound_room_id"] = nil
	}
	if d.BoundBedID.Valid {
		m["bound_bed_id"] = d.BoundBedID.String
	} else {
		m["bound_bed_id"] = nil
	}
	if d.Metadata.Valid {
		m["metadata"] = jsonRawOrString(d.Metadata.String)
	}
	return m
}

type DevicesRepo interface {
	ListDevices(ctx context.Context, tenantID string, filters map[string]any) (items []Device, total int, err error)
	GetDevice(ctx context.Context, tenantID, deviceID string) (*Device, error)
	UpdateDevice(ctx context.Context, tenantID, deviceID string, payload map[string]any) error
	DisableDevice(ctx context.Context, tenantID, deviceID string) error
}
