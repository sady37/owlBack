package domain

import (
	"database/sql"
	"encoding/json"
)

// Device 设备领域模型（对应 devices 表）
// 基于实际DB表结构：devices表的所有字段
type Device struct {
	// 主键和租户
	DeviceID   string `db:"device_id"`
	DeviceIPv6 string `db:"device_ipv6"` // canonical IPv6 form (host(d.device_ipv6))；用于 cardagg redis device:status key
	TenantID   string `db:"tenant_id"` // NOT NULL

	// 关联 device_store
	DeviceUID string `db:"device_uid"` // NOT NULL, REFERENCES device_store(device_uid)

	// 物理属性（从 device_store 表获取，只读）
	DeviceType  sql.NullString `db:"device_type"`  // from device_store.device_type
	DeviceModel sql.NullString `db:"device_model"` // from device_store.device_model
	DeviceCode  sql.NullString `db:"device_code"`  // from device_store.device_code
	MAC         sql.NullString `db:"mac"`          // from device_store.mac
	IMEI        sql.NullString `db:"imei"`         // from device_store.imei
	CommMode    sql.NullString `db:"comm_mode"`    // from device_store.comm_mode
	MCUModel    sql.NullString `db:"mcu_model"`    // from device_store.mcu_model
	FirmwareVersion sql.NullString `db:"firmware_version"` // from device_store.firmware_version

	// OTA 字段（从 device_store 获取，只读）
	OTATargetFW  sql.NullString `db:"ota_target_firmware_version"`
	OTATargetMCU sql.NullString `db:"ota_target_mcu_model"`
	OTAPermit    sql.NullString `db:"ota_permit"`
	OTAWay       sql.NullString `db:"ota_way"`
	OTASchedule  sql.NullString `db:"ota_schedule"`
	OTAStatus    sql.NullString `db:"ota_status"`
	OTAProgress  sql.NullInt32  `db:"ota_progress"`
	OTATenantApproved bool     `db:"ota_tenant_approved"`

	// 标识/资产
	DeviceName string `db:"device_name"` // NOT NULL

	// 位置绑定（互斥）
	BoundRoomID sql.NullString `db:"bound_room_id"` // nullable
	BoundBedID  sql.NullString `db:"bound_bed_id"`  // nullable
	
	// 位置信息（通过 JOIN 查询得到，用于返回给前端）
	// 如果绑定到 room：room_id = bound_room_id, unit_id 通过 rooms.unit_id 查询
	// 如果绑定到 bed：room_id 通过 beds.room_id 查询, unit_id 通过 rooms.unit_id 查询
	RoomID sql.NullString `db:"room_id"`   // 计算字段：COALESCE(bound_room_id, bed_room_id)
	UnitID sql.NullString `db:"unit_id"`   // 计算字段：通过 room_id 或 bed_id 查询得到
	CardID sql.NullString `db:"-"`         // 列表/详情：cards.devices 含该 device_id 时的 card_id
	UnitName sql.NullString `db:"-"`       // 列表：units.unit_name（与 UnitID 对应）
	DNSShortName sql.NullString `db:"-"`   // 列表：cards.dns_short_name（与 CardID 对应，FE 用于短码展示）
	BranchID sql.NullString `db:"-"`       // 列表：byte 6 派生 /56；byte 6==0 表示 tenant 池 (NULL)
	BranchName sql.NullString `db:"-"`     // 列表：branches.branch_name

	// 状态/维护
	Status            string `db:"status"`              // NOT NULL, default 'offline' (数据库状态：Enabled/Disabled/Error，用于软删除)
	OnlineStatus      string `db:"-"`                   // 实时在线状态（online/offline/unsubscribed），从 Redis 读取，不存储到数据库
	Access            bool   `db:"access"`              // platform_admin 审批位
	MonitoringEnabled bool   `db:"monitoring_enabled"`  // tenant 业务开关

	// 元数据
	Metadata sql.NullString `db:"metadata"` // nullable, JSONB
}

// ToJSON 转换为JSON格式（用于HTTP响应）
func (d *Device) ToJSON() map[string]any {
	m := map[string]any{
		"device_id":          d.DeviceID,
		"tenant_id":          d.TenantID,
		"device_name":        d.DeviceName,
		"status":             d.Status,        // 数据库状态（Enabled/Disabled/Error）
		"online_status":      d.OnlineStatus,  // 实时在线状态（online/offline/unsubscribed）
		"access":             d.Access,
		"monitoring_enabled": d.MonitoringEnabled,
	}
	m["device_uid"] = d.DeviceUID
	// 物理属性（从 device_store 获取）
	if d.DeviceType.Valid {
		m["device_type"] = d.DeviceType.String
	}
	if d.DeviceModel.Valid {
		m["device_model"] = d.DeviceModel.String
	}
	if d.DeviceCode.Valid {
		m["device_code"] = d.DeviceCode.String
	}
	if d.MAC.Valid {
		m["mac"] = d.MAC.String
	}
	if d.IMEI.Valid {
		m["imei"] = d.IMEI.String
	}
	if d.CommMode.Valid {
		m["comm_mode"] = d.CommMode.String
	}
	if d.MCUModel.Valid {
		m["mcu_model"] = d.MCUModel.String
	}
	if d.FirmwareVersion.Valid {
		m["firmware_version"] = d.FirmwareVersion.String
	}
	if d.OTATargetFW.Valid {
		m["ota_target_firmware_version"] = d.OTATargetFW.String
	}
	if d.OTATargetMCU.Valid {
		m["ota_target_mcu_model"] = d.OTATargetMCU.String
	}
	if d.OTAPermit.Valid {
		m["ota_permit"] = d.OTAPermit.String
	}
	if d.OTAWay.Valid {
		m["ota_way"] = d.OTAWay.String
	}
	if d.OTASchedule.Valid {
		m["ota_schedule"] = d.OTASchedule.String
	}
	if d.OTAStatus.Valid {
		m["ota_status"] = d.OTAStatus.String
	} else {
		m["ota_status"] = "idle"
	}
	if d.OTAProgress.Valid {
		m["ota_progress"] = d.OTAProgress.Int32
	}
	m["ota_tenant_approved"] = d.OTATenantApproved
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
	// 位置信息（用于前端显示，通过 JOIN 查询得到）
	if d.RoomID.Valid {
		m["room_id"] = d.RoomID.String
	} else {
		m["room_id"] = nil
	}
	if d.UnitID.Valid {
		m["unit_id"] = d.UnitID.String
	} else {
		m["unit_id"] = nil
	}
	if d.CardID.Valid {
		m["card_id"] = d.CardID.String
	} else {
		m["card_id"] = nil
	}
	if d.UnitName.Valid {
		m["unit_name"] = d.UnitName.String
	} else {
		m["unit_name"] = nil
	}
	if d.DNSShortName.Valid && d.DNSShortName.String != "" {
		m["dns_short_name"] = d.DNSShortName.String
	}
	if d.BranchID.Valid {
		m["branch_id"] = d.BranchID.String
	} else {
		m["branch_id"] = nil
	}
	if d.BranchName.Valid {
		m["branch_name"] = d.BranchName.String
	} else {
		m["branch_name"] = nil
	}
	if d.Metadata.Valid {
		// 尝试解析JSON，如果失败则返回字符串
		var jsonData any
		if err := json.Unmarshal([]byte(d.Metadata.String), &jsonData); err == nil {
			m["metadata"] = jsonData
		} else {
			m["metadata"] = d.Metadata.String
		}
	}
	return m
}

