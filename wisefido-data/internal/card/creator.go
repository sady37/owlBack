package card

import (
	"bytes"
	"fmt"

	commoncard "owl-common/card"

	"go.uber.org/zap"
)

// CardCreator 根据 unit 的 beds/devices/residents 计算期望卡片并与现有卡片 diff，调用 repo 增删改（写 DB）
type CardCreator struct {
	repo   RepositoryInterface
	logger *zap.Logger
}

// NewCardCreator 创建卡片同步器
func NewCardCreator(repo RepositoryInterface, logger *zap.Logger) *CardCreator {
	return &CardCreator{repo: repo, logger: logger}
}

// CreateCardsForUnit 为指定 unit 同步卡片：期望卡片与现有卡片 diff，CreateCard/UpdateCard/DeleteCard
func (c *CardCreator) CreateCardsForUnit(tenantID, unitID string) (*commoncard.CardUpdateStats, error) {
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
			_, err := c.repo.CreateCard(tenantID, exp.CardType, exp.BedID, exp.UnitID, exp.CardName, exp.CardAddress, exp.Timezone, exp.ResidentID, exp.DevicesJSON, exp.ResidentsJSON)
			if err != nil {
				return nil, fmt.Errorf("create card: %w", err)
			}
			stats.CreatedCount++
			continue
		}
		if !contentEqual(exp, ex) {
			if err := c.repo.UpdateCard(tenantID, ex.CardID, exp.CardType, exp.BedID, exp.UnitID, exp.CardName, exp.CardAddress, exp.Timezone, exp.ResidentID, exp.DevicesJSON, exp.ResidentsJSON); err != nil {
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
	if cardType == "ActiveBed" && bedID != nil {
		return "ActiveBed:" + *bedID
	}
	return "Location:" + unitID
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
	if !bytes.Equal(e.DevicesJSON, c.DevicesJSON) || !bytes.Equal(e.ResidentsJSON, c.ResidentsJSON) {
		return false
	}
	return true
}

func (c *CardCreator) buildExpectedCards(tenantID, unitID, timezone string) ([]commoncard.ExpectedCard, error) {
	var expected []commoncard.ExpectedCard

	beds, err := c.repo.GetActiveBedsByUnit(tenantID, unitID)
	if err != nil {
		return nil, err
	}
	for _, bed := range beds {
		devices, err := c.repo.GetDevicesByBed(tenantID, bed.BedID)
		if err != nil {
			return nil, fmt.Errorf("get devices by bed %s: %w", bed.BedID, err)
		}
		resident, _ := c.repo.GetResidentByBed(tenantID, bed.BedID)
		devicesJSON, err := commoncard.ConvertDevicesToJSON(devices)
		if err != nil {
			return nil, err
		}
		residentsJSON, err := commoncard.ConvertResidentsToJSON(residentToSlice(resident))
		if err != nil {
			return nil, err
		}
		cardName := bed.BedID
		if bed.RoomID != "" {
			cardName = bed.RoomID + " - " + bed.BedID
		}
		addr := unitID
		var residentID *string
		if resident != nil {
			residentID = &resident.ResidentID
		}
		expected = append(expected, commoncard.ExpectedCard{
			CardType:      "ActiveBed",
			BedID:         &bed.BedID,
			UnitID:        unitID,
			CardName:      cardName,
			CardAddress:   addr,
			Timezone:      timezone,
			ResidentID:    residentID,
			DevicesJSON:   devicesJSON,
			ResidentsJSON: residentsJSON,
		})
	}

	unbound, err := c.repo.GetUnboundDevicesByUnit(tenantID, unitID)
	if err != nil {
		return nil, err
	}
	if len(unbound) > 0 {
		residents, err := c.repo.GetResidentsByUnit(tenantID, unitID)
		if err != nil {
			return nil, err
		}
		devicesJSON, err := commoncard.ConvertDevicesToJSON(unbound)
		if err != nil {
			return nil, err
		}
		residentsJSON, err := commoncard.ConvertResidentsToJSON(residents)
		if err != nil {
			return nil, err
		}
		unitInfo, _ := c.repo.GetUnitInfo(tenantID, unitID)
		cardName := unitID
		if unitInfo != nil && unitInfo.UnitName != "" {
			cardName = unitInfo.UnitName
		}
		expected = append(expected, commoncard.ExpectedCard{
			CardType:      "Location",
			BedID:         nil,
			UnitID:        unitID,
			CardName:      cardName,
			CardAddress:   unitID,
			Timezone:      timezone,
			ResidentID:    nil,
			DevicesJSON:   devicesJSON,
			ResidentsJSON: residentsJSON,
		})
	}

	return expected, nil
}

func residentToSlice(r *commoncard.ResidentInfo) []commoncard.ResidentInfo {
	if r == nil {
		return nil
	}
	return []commoncard.ResidentInfo{*r}
}
