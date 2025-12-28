package domain

// ResidentPHIUpdate 住户PHI更新模型（对应 resident_phi 表）
// 用于区分"不更新"、"更新"、"删除"三种状态
// 注意：主键字段（phi_id, tenant_id, resident_id）不在更新模型中
// 注意：此表使用 UPSERT 语义（UNIQUE(tenant_id, resident_id)）
type ResidentPHIUpdate struct {
	// Basic PHI（真实姓名，PII，加密存储）
	FirstName     *UpdateString // VARCHAR(100), nullable
	LastName      *UpdateString // VARCHAR(100), nullable
	Gender        *UpdateString // VARCHAR(10), nullable
	DateOfBirth   *UpdateTime   // DATE, nullable
	ResidentPhone *UpdateString // VARCHAR(25), nullable
	ResidentEmail *UpdateString // VARCHAR(255), nullable

	// Biometric PHI（身高/体重）
	WeightLb *UpdateFloat64 // DECIMAL(5,2), nullable
	HeightFt *UpdateFloat64 // DECIMAL(5,2), nullable
	HeightIn *UpdateFloat64 // DECIMAL(5,2), nullable

	// Functional Mobility（功能性活动能力）
	MobilityLevel *UpdateInt // INTEGER, nullable

	// Functional Health（功能性健康状态）
	TremorStatus  *UpdateString // VARCHAR(20), nullable
	MobilityAid   *UpdateString // VARCHAR(20), nullable
	ADLAssistance *UpdateString // VARCHAR(20), nullable
	CommStatus    *UpdateString // VARCHAR(20), nullable

	// Chronic Conditions / Medical History（老年常见慢病与病史）
	HasHypertension   *UpdateBool   // BOOLEAN, nullable
	HasHyperlipaemia  *UpdateBool   // BOOLEAN, nullable
	HasHyperglycaemia *UpdateBool   // BOOLEAN, nullable
	HasStrokeHistory  *UpdateBool   // BOOLEAN, nullable
	HasParalysis      *UpdateBool   // BOOLEAN, nullable
	HasAlzheimer      *UpdateBool   // BOOLEAN, nullable
	MedicalHistory    *UpdateString // TEXT, nullable

	// 家庭地址信息（PHI，仅用于Home场景）
	HomeAddressStreet     *UpdateString // VARCHAR(255), nullable
	HomeAddressCity       *UpdateString // VARCHAR(100), nullable
	HomeAddressState      *UpdateString // VARCHAR(50), nullable
	HomeAddressPostalCode *UpdateString // VARCHAR(20), nullable
	PlusCode              *UpdateString // VARCHAR(32), nullable
}
