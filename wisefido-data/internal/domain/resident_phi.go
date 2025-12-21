package domain

import (
	"time"
)

// ResidentPHI 住户PHI领域模型（对应 resident_phi 表）
// 存放可选的个人健康信息，物理上与 residents 分离
// DB层面加密存储，不存储明文
type ResidentPHI struct {
	// 主键
	PhiID string `db:"phi_id"` // UUID, PRIMARY KEY

	// 租户和住户
	TenantID   string `db:"tenant_id"`   // UUID, NOT NULL
	ResidentID string `db:"resident_id"`  // UUID, NOT NULL, UNIQUE(tenant_id, resident_id)

	// Basic PHI（真实姓名，PII，加密存储）
	FirstName string `db:"first_name"` // VARCHAR(100), nullable
	LastName  string `db:"last_name"`  // VARCHAR(100), nullable
	Gender    string `db:"gender"`     // VARCHAR(10), nullable（Male/Female/Other/Unknown）
	DateOfBirth *time.Time `db:"date_of_birth"` // DATE, nullable
	ResidentPhone string `db:"resident_phone"`   // VARCHAR(25), nullable
	ResidentEmail string `db:"resident_email"`  // VARCHAR(255), nullable

	// Biometric PHI（身高/体重）
	WeightLb *float64 `db:"weight_lb"` // DECIMAL(5,2), nullable
	HeightFt *float64 `db:"height_ft"`  // DECIMAL(5,2), nullable
	HeightIn *float64 `db:"height_in"`  // DECIMAL(5,2), nullable

	// Functional Mobility（功能性活动能力）
	MobilityLevel *int `db:"mobility_level"` // INTEGER, nullable（0: 无行动能力 ~ 5: 完全独立）

	// Functional Health（功能性健康状态）
	TremorStatus  string `db:"tremor_status"`  // VARCHAR(20), nullable（None/Mild/Severe）
	MobilityAid   string `db:"mobility_aid"`    // VARCHAR(20), nullable（Cane/Wheelchair/None）
	ADLAssistance string `db:"adl_assistance"`  // VARCHAR(20), nullable（Independent/NeedsHelp）
	CommStatus   string `db:"comm_status"`     // VARCHAR(20), nullable（Normal/SpeechDifficulty）

	// Chronic Conditions / Medical History（老年常见慢病与病史）
	HasHypertension bool `db:"has_hypertension"` // BOOLEAN, nullable
	HasHyperlipaemia bool `db:"has_hyperlipaemia"` // BOOLEAN, nullable
	HasHyperglycaemia bool `db:"has_hyperglycaemia"` // BOOLEAN, nullable
	HasStrokeHistory bool `db:"has_stroke_history"`  // BOOLEAN, nullable
	HasParalysis     bool `db:"has_paralysis"`      // BOOLEAN, nullable
	HasAlzheimer     bool `db:"has_alzheimer"`       // BOOLEAN, nullable
	MedicalHistory   string `db:"medical_history"`  // TEXT, nullable

	// 外部HIS（医院信息系统）同步字段
	HISResidentName          string     `db:"his_resident_name"`           // VARCHAR(100), nullable
	HISResidentAdmissionDate *time.Time `db:"his_resident_admission_date"` // DATE, nullable
	HISResidentDischargeDate *time.Time `db:"his_resident_discharge_date"` // DATE, nullable
	HISResidentMetadata      string     `db:"his_resident_metadata"`      // JSONB, nullable（存储为JSON字符串）

	// 家庭地址信息（PHI，仅用于Home场景）
	HomeAddressStreet   string `db:"home_address_street"`   // VARCHAR(255), nullable
	HomeAddressCity     string `db:"home_address_city"`      // VARCHAR(100), nullable
	HomeAddressState    string `db:"home_address_state"`     // VARCHAR(50), nullable
	HomeAddressPostalCode string `db:"home_address_postal_code"` // VARCHAR(20), nullable
	PlusCode            string `db:"plus_code"`             // VARCHAR(32), nullable（Google Plus Code）
}

