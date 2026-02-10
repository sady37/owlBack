package service

import (
	"fmt"
	"strings"

	commoncard "owl-common/card"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// CardCreateService 卡片创建服务：根据 unit 的 beds/devices/residents 计算期望卡片并与现有卡片 diff，调用 repo 增删改（写 DB）
type CardCreateService struct {
	repo   repository.RepositoryInterface
	logger *zap.Logger
}

// NewCardCreateService 创建卡片创建服务
func NewCardCreateService(repo repository.RepositoryInterface, logger *zap.Logger) *CardCreateService {
	return &CardCreateService{repo: repo, logger: logger}
}

// CreateCardsForUnit 为指定 unit 同步卡片：期望卡片与现有卡片 diff，CreateCard/UpdateCard/DeleteCard
func (c *CardCreateService) CreateCardsForUnit(tenantID, unitID string) (*commoncard.CardUpdateStats, error) {
	unitInfo, err := c.repo.GetUnitInfo(tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("get unit info: %w", err)
	}
	timezone := "UTC"
	if unitInfo != nil && unitInfo.Timezone != "" {
		timezone = unitInfo.Timezone
	}

	expected, err := c.buildExpectedCards(tenantID, unitID, timezone)
	if err != nil {
		return nil, fmt.Errorf("build expected cards: %w", err)
	}

	existing, err := c.repo.GetCardsByUnit(tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("get cards by unit: %w", err)
	}

	stats := &commoncard.CardUpdateStats{ExistingCount: len(existing)}

	expectedByKey := make(map[string]commoncard.ExpectedCard)
	for _, e := range expected {
		key := cardKey(e.CardType, e.BedID, e.UnitID)
		expectedByKey[key] = e
	}
	existingByKey := make(map[string]commoncard.CardWithContent)
	for _, ex := range existing {
		key := cardKey(ex.CardType, ex.BedID, ex.UnitID)
		existingByKey[key] = ex
	}

	for key, ex := range existingByKey {
		if _, ok := expectedByKey[key]; !ok {
			if err := c.repo.DeleteCard(tenantID, ex.CardID); err != nil {
				return nil, fmt.Errorf("delete card %s: %w", ex.CardID, err)
			}
			stats.DeletedCount++
		}
	}

	for key, exp := range expectedByKey {
		ex, has := existingByKey[key]
		if !has {
			_, err := c.repo.CreateCard(tenantID, exp.CardType, exp.BedID, exp.UnitID, exp.CardName, exp.CardAddress, exp.Timezone, exp.ResidentID, exp.Devices, exp.Residents)
			if err != nil {
				return nil, fmt.Errorf("create card: %w", err)
			}
			stats.CreatedCount++
			continue
		}
		if !contentEqual(exp, ex) {
			if err := c.repo.UpdateCard(tenantID, ex.CardID, exp.CardType, exp.BedID, exp.UnitID, exp.CardName, exp.CardAddress, exp.Timezone, exp.ResidentID, exp.Devices, exp.Residents); err != nil {
				return nil, fmt.Errorf("update card %s: %w", ex.CardID, err)
			}
			stats.UpdatedCount++
		} else {
			stats.UnchangedCount++
		}
	}

	return stats, nil
}

func cardKey(cardType string, bedID *string, unitID string) string {
	if cardType == "ActiveBedCard" && bedID != nil {
		return "ActiveBedCard:" + *bedID
	}
	return "UnitCard:" + unitID
}

func contentEqual(e commoncard.ExpectedCard, c commoncard.CardWithContent) bool {
	if e.CardType != c.CardType || e.UnitID != c.UnitID || e.CardName != c.CardName || e.CardAddress != c.CardAddress || e.Timezone != c.Timezone {
		return false
	}
	if (e.BedID == nil) != (c.BedID == nil) {
		return false
	}
	if e.BedID != nil && c.BedID != nil && *e.BedID != *c.BedID {
		return false
	}
	if (e.ResidentID == nil) != (c.ResidentID == nil) {
		return false
	}
	if e.ResidentID != nil && c.ResidentID != nil && *e.ResidentID != *c.ResidentID {
		return false
	}

	// 比较设备
	if len(e.Devices) != len(c.Devices) {
		return false
	}
	for i, ed := range e.Devices {
		cd := c.Devices[i]
		if ed.DeviceID != cd.DeviceID || ed.DeviceName != cd.DeviceName || ed.DeviceType != cd.DeviceType {
			return false
		}
		if (ed.BoundBedID == nil) != (cd.BoundBedID == nil) {
			return false
		}
		if ed.BoundBedID != nil && cd.BoundBedID != nil && *ed.BoundBedID != *cd.BoundBedID {
			return false
		}
	}

	// 比较住户
	if len(e.Residents) != len(c.Residents) {
		return false
	}
	for i, er := range e.Residents {
		cr := c.Residents[i]
		if er.ResidentID != cr.ResidentID || er.Nickname != cr.Nickname {
			return false
		}
	}

	return true
}

func (c *CardCreateService) buildExpectedCards(tenantID, unitID, timezone string) ([]commoncard.ExpectedCard, error) {
	var expected []commoncard.ExpectedCard

	// 1. Get unit info
	unitInfo, err := c.repo.GetUnitInfo(tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("get unit info: %w", err)
	}
	if unitInfo == nil {
		return nil, fmt.Errorf("unit not found: %s", unitID)
	}

	// 2. Get ActiveBeds
	activeBeds, err := c.repo.GetActiveBedsByUnit(tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("get active beds: %w", err)
	}

	activeBedCount := len(activeBeds)

	// 3. Get unbound devices
	unboundDevices, err := c.repo.GetUnboundDevicesByUnit(tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("get unbound devices: %w", err)
	}

	// 4. Process by scenario
	if activeBedCount == 0 {
		// === Scenario C: No ActiveBed ===
		// Only create UnitCard if there are unbound devices
		if len(unboundDevices) > 0 {
			unitCardExp, err := c.calculateExpectedUnitCard(tenantID, unitInfo, unboundDevices, timezone)
			if err != nil {
				return nil, err
			}
			expected = append(expected, *unitCardExp)
		}
	} else if activeBedCount == 1 {
		// === Scenario A: Only 1 ActiveBed ===
		// Create 1 ActiveBed card that includes ALL devices (bound + unbound)
		bed := activeBeds[0]
		bedDevices, err := c.repo.GetDevicesByBed(tenantID, bed.BedID)
		if err != nil {
			return nil, fmt.Errorf("get devices by bed: %w", err)
		}

		// Merge bound + unbound devices
		allDevices := append(bedDevices, unboundDevices...)

		cardName, err := c.calculateActiveBedCardName(tenantID, bed, unitInfo)
		if err != nil {
			return nil, fmt.Errorf("calculate bed card name: %w", err)
		}

		cardAddr := c.calculateCardAddress(unitInfo)

		resident, _ := c.repo.GetResidentByBed(tenantID, bed.BedID)
		var residentID *string
		var residents []commoncard.ResidentInfo
		if resident != nil {
			residentID = &resident.ResidentID
			residents = []commoncard.ResidentInfo{*resident}
		} else {
			unitResidents, err := c.repo.GetResidentsByUnit(tenantID, unitID)
			if err == nil && len(unitResidents) > 0 {
				residents = unitResidents
				// Do NOT assign ResidentID unless actually bound to bed
				residentID = nil
			}
		}

		expected = append(expected, commoncard.ExpectedCard{
			CardType:    "ActiveBedCard",
			BedID:       &bed.BedID,
			UnitID:      unitID,
			CardName:    cardName,
			CardAddress: cardAddr,
			Timezone:    timezone,
			ResidentID:  residentID,
			Devices:     allDevices,
			Residents:   residents,
		})
	} else {
		// === Scenario B: Multiple ActiveBeds (≥2) ===
		// Create N ActiveBed cards (each with only bound devices)
		for _, bed := range activeBeds {
			bedDevices, err := c.repo.GetDevicesByBed(tenantID, bed.BedID)
			if err != nil {
				return nil, fmt.Errorf("get devices by bed: %w", err)
			}

			cardName, err := c.calculateActiveBedCardName(tenantID, bed, unitInfo)
			if err != nil {
				return nil, fmt.Errorf("calculate bed card name: %w", err)
			}

			cardAddr := c.calculateCardAddress(unitInfo)

			resident, _ := c.repo.GetResidentByBed(tenantID, bed.BedID)
			var residentID *string
			var residents []commoncard.ResidentInfo
			if resident != nil {
				// Bed has specific resident
				residentID = &resident.ResidentID
				residents = []commoncard.ResidentInfo{*resident}
			} else {
				// Bed not bound to specific resident
				unitResidents, err := c.repo.GetResidentsByUnit(tenantID, unitID)
				if err == nil && len(unitResidents) > 0 {
					residents = unitResidents
					residentID = nil
				}
			}

			expected = append(expected, commoncard.ExpectedCard{
				CardType:    "ActiveBedCard",
				BedID:       &bed.BedID,
				UnitID:      unitID,
				CardName:    cardName,
				CardAddress: cardAddr,
				Timezone:    timezone,
				ResidentID:  residentID,
				Devices:     bedDevices,
				Residents:   residents,
			})
		}

		// Create UnitCard only if there are unbound devices
		if len(unboundDevices) > 0 {
			unitCardExp, err := c.calculateExpectedUnitCard(tenantID, unitInfo, unboundDevices, timezone)
			if err != nil {
				return nil, err
			}
			expected = append(expected, *unitCardExp)
		}
	}

	return expected, nil
}


// calculateActiveBedCardName calculates ActiveBed card name according to rules
// Rule Table:
// UnitType  | Public | Shared | bed.Resident | unit.Residents | CardName
// Facility  | ✗     | ✓      | ✓            | -              | resident.Nickname
// Facility  | ✗     | ✓      | ✗            | -              | BedName - Unoccupied
// Facility  | ✗     | ✗      | ✓            | -              | resident.Nickname
// Facility  | ✗     | ✗      | ✗            | ✓              | residents[0].Nickname
// Facility  | ✗     | ✗      | ✗            | ✗              | UnitName - Unoccupied
// Facility  | ✓     | -      | -            | -              | UnitName - BedName
// HomeCare  | -     | -      | ✓            | -              | resident.Nickname - BedName
// HomeCare  | -     | -      | ✗            | ✓              | residents[0].Nickname - BedName
func (c *CardCreateService) calculateActiveBedCardName(
    tenantID string,
    bed commoncard.ActiveBedRow,
    unitInfo *commoncard.UnitInfo,
) (string, error) {
	isFacility := unitInfo.UnitType == "Facility" || unitInfo.UnitType == "Institutional"

	// --- 1. Facility (机构养老) 模式 ---
	if isFacility {
		// 优先级 1: 公共区域
		if unitInfo.IsPublic {
			return unitInfo.UnitName + " - " + bed.BedName, nil
		}

		// 优先级 2: 共享/多人房 (IsSharedUnit = true)
		if unitInfo.IsSharedUnit {
			if bed.ResidentID != nil && *bed.ResidentID != "" {
				resident, err := c.repo.GetResidentByBed(tenantID, bed.BedID)
				if err == nil && resident != nil {
					return resident.Nickname, nil
				}
			}
			// 多人房床位无住户
			return bed.BedName + " - Unoccupied", nil
		}

		// 优先级 3: 专属房 (IsSharedUnit = false)
		// 先看床位有没有住户
		if bed.ResidentID != nil && *bed.ResidentID != "" {
			resident, err := c.repo.GetResidentByBed(tenantID, bed.BedID)
			if err == nil && resident != nil {
				return resident.Nickname, nil
			}
		}
		// 床位没住户，看单元有没有住户（套房逻辑）
		residents, _ := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
		if len(residents) > 0 {
			return residents[0].Nickname, nil
		}
		// 全空
		return unitInfo.UnitName + " - Unoccupied", nil
	}

	// --- 2. HomeCare (居家养老) 模式 ---
	// 忽略 IsSharedUnit 属性，强制要求包含 Nickname 和 BedName
	var residentName string
	// 优先获取床位绑定的住户
	if bed.ResidentID != nil && *bed.ResidentID != "" {
		resident, err := c.repo.GetResidentByBed(tenantID, bed.BedID)
		if err == nil && resident != nil {
			residentName = resident.Nickname
		}
	}
	// 如果床位没绑，取单元第一个住户
	if residentName == "" {
		residents, _ := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
		if len(residents) > 0 {
			residentName = residents[0].Nickname
		}
	}

	if residentName != "" {
		return residentName + " - " + bed.BedName, nil
	}
	// 极端情况保底
	return unitInfo.UnitName + " - " + bed.BedName, nil
}

// calculateUnitCardName calculates UnitCard name according to rules
// Rule Table:
// UnitType  | Public | Shared | Unit.Residents | CardName
// Facility  | ✗      | ✓      | -              | UnitName
// Facility  | ✗      | ✗      | ✓              | resident.Nickname - UnitName
// Facility  | ✗      | ✗      | ✗              | UnitName - Unoccupied
// Facility  | ✓      | -      | -              | UnitName
// HomeCare  | -      | -      | -              | resident.Nickname - UnitName
func (c *CardCreateService) calculateUnitCardName(
    tenantID string,
    unitInfo *commoncard.UnitInfo,
) (string, error) {
	isFacility := unitInfo.UnitType == "Facility" || unitInfo.UnitType == "Institutional"

	// 1. Facility + (Public 或 Shared) → 仅显示单元名
	if isFacility && (unitInfo.IsPublic || unitInfo.IsSharedUnit) {
		return unitInfo.UnitName, nil
	}

	// 2. HomeCare 或 Facility Exclusive → 尝试获取住户昵称拼接
	residents, _ := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
	if len(residents) > 0 {
		return residents[0].Nickname + " - " + unitInfo.UnitName, nil
	}

	// 3. 无住户时的保底处理
	if isFacility {
		// 机构专属房空置显示 Unoccupied
		return unitInfo.UnitName + " - Unoccupied", nil
	}
	// 居家模式全空保底
	return unitInfo.UnitName, nil
}

// calculateCardAddress calculates card address: BranchName + "-" + building + "-" + unit_name
// Skips empty values or default "-"
func (c *CardCreateService) calculateCardAddress(unitInfo *commoncard.UnitInfo) string {
	var parts []string

	if unitInfo.BranchName != "" && unitInfo.BranchName != "-" {
		parts = append(parts, unitInfo.BranchName)
	}
	if unitInfo.Building != "" && unitInfo.Building != "-" {
		parts = append(parts, unitInfo.Building)
	}
	if unitInfo.UnitName != "" && unitInfo.UnitName != "-" {
		parts = append(parts, unitInfo.UnitName)
	}

	if len(parts) == 0 {
		return unitInfo.UnitID // Fallback to unitID
	}

	return strings.Join(parts, "-")
}

// calculateExpectedUnitCard calculates expected UnitCard
func (c *CardCreateService) calculateExpectedUnitCard(
	tenantID string,
	unitInfo *commoncard.UnitInfo,
	devices []commoncard.DeviceInfo,
	timezone string,
) (*commoncard.ExpectedCard, error) {
	cardName, err := c.calculateUnitCardName(tenantID, unitInfo)
	if err != nil {
		return nil, fmt.Errorf("calculate unit card name: %w", err)
	}

	cardAddr := c.calculateCardAddress(unitInfo)

	residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
	if err != nil {
		residents = []commoncard.ResidentInfo{}
	}

	return &commoncard.ExpectedCard{
		CardType:    "UnitCard",
		BedID:       nil,
		UnitID:      unitInfo.UnitID,
		CardName:    cardName,
		CardAddress: cardAddr,
		Timezone:    timezone,
		ResidentID:  nil,
		Devices:     devices,
		Residents:   residents,
	}, nil
}
