package card

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"go.uber.org/zap"
)


// CardCreator card creator
type CardCreator struct {
	repo   RepositoryInterface
	logger *zap.Logger
}

// NewCardCreator creates a new card creator
func NewCardCreator(repo RepositoryInterface, logger *zap.Logger) *CardCreator {
	return &CardCreator{
		repo:   repo,
		logger: logger,
	}
}

// CreateCardsForUnit creates cards for the specified unit
// According to card creation rules, handles three scenarios:
// - Scenario A: Unit has only 1 ActiveBed
// - Scenario B: Unit has multiple ActiveBeds (≥2)
// - Scenario C: Unit has no ActiveBed
//
// Optimization: Compares existing cards with expected cards before recreating
// If cards are unchanged, skips update to preserve card_id and avoid unnecessary DB operations
// Returns statistics about the update operation
func (c *CardCreator) CreateCardsForUnit(tenantID, unitID string) (*CardUpdateStats, error) {
	stats := &CardUpdateStats{}
	// 1. Get unit information
	unitInfo, err := c.repo.GetUnitInfo(tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit info: %w", err)
	}

	// 2. Get all ActiveBeds under this unit
	activeBeds, err := c.repo.GetActiveBedsByUnit(tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active beds: %w", err)
	}

	// 3. Calculate expected cards (without actually creating them)
	expectedCards, err := c.calculateExpectedCards(tenantID, unitInfo, activeBeds)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate expected cards: %w", err)
	}

	// 4. Get existing cards from database
	existingCards, err := c.repo.GetCardsByUnit(tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing cards: %w", err)
	}

	stats.ExistingCount = len(existingCards)

	// 5. Compare existing cards with expected cards
	if c.compareCards(existingCards, expectedCards) {
		// Cards are unchanged, skip update to preserve card_id
		stats.UnchangedCount = len(existingCards)
		c.logger.Debug("Cards unchanged, skipping update",
			zap.String("unit_id", unitID),
			zap.Int("card_count", len(expectedCards)),
		)
		return stats, nil
	}

	// 6. Cards have changed, perform incremental update
	// Find cards to delete, create, and update
	toDelete, toCreate, toUpdate := c.diffCards(existingCards, expectedCards)

	stats.DeletedCount = len(toDelete)
	stats.CreatedCount = len(toCreate)
	stats.UpdatedCount = len(toUpdate)
	stats.UnchangedCount = stats.ExistingCount - stats.DeletedCount - stats.UpdatedCount

	c.logger.Info("Cards changed, performing incremental update",
		zap.String("unit_id", unitID),
		zap.Int("existing_count", len(existingCards)),
		zap.Int("expected_count", len(expectedCards)),
		zap.Int("to_delete", len(toDelete)),
		zap.Int("to_create", len(toCreate)),
		zap.Int("to_update", len(toUpdate)),
	)

	// 7. Delete cards that are no longer needed
	for _, card := range toDelete {
		if err := c.repo.DeleteCard(tenantID, card.CardID); err != nil {
			c.logger.Error("Failed to delete card",
				zap.String("card_id", card.CardID),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to delete card %s: %w", card.CardID, err)
		}
		c.logger.Debug("Deleted card",
			zap.String("card_id", card.CardID),
			zap.String("card_type", card.CardType),
		)
	}

	// 8. Update cards that have changed (use UPDATE to preserve card_id)
	// This preserves card_id to avoid frontend cache invalidation
	for _, card := range toUpdate {
		// Find the corresponding expected card
		key := cardKey(card.CardType, card.BedID, card.UnitID)
		var expectedCard *ExpectedCard
		for i := range expectedCards {
			if cardKey(expectedCards[i].CardType, expectedCards[i].BedID, expectedCards[i].UnitID) == key {
				expectedCard = &expectedCards[i]
				break
			}
		}
		if expectedCard == nil {
			c.logger.Error("Failed to find expected card for update",
				zap.String("card_id", card.CardID),
				zap.String("key", key),
			)
			return nil, fmt.Errorf("failed to find expected card for update: %s", card.CardID)
		}

		// Update existing card (preserves card_id)
		if err := c.repo.UpdateCard(
			tenantID,
			card.CardID,
			expectedCard.CardType,
			expectedCard.BedID,
			expectedCard.UnitID,
			expectedCard.CardName,
			expectedCard.CardAddress,
			expectedCard.ResidentID,
			expectedCard.DevicesJSON,
			expectedCard.ResidentsJSON,
		); err != nil {
			c.logger.Error("Failed to update card",
				zap.String("card_id", card.CardID),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to update card %s: %w", card.CardID, err)
		}
		c.logger.Debug("Updated card",
			zap.String("card_id", card.CardID),
			zap.String("card_type", expectedCard.CardType),
		)
	}

	// 9. Create new cards (only cards that don't exist)
	for _, expected := range toCreate {
		cardID, err := c.repo.CreateCard(
			tenantID,
			expected.CardType,
			expected.BedID,
			expected.UnitID,
			expected.CardName,
			expected.CardAddress,
			expected.ResidentID,
			expected.DevicesJSON,
			expected.ResidentsJSON,
		)
		if err != nil {
			c.logger.Error("Failed to create card",
				zap.String("card_type", expected.CardType),
				zap.String("bed_id", cardKeyBedID(expected.BedID)),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to create card: %w", err)
		}
		c.logger.Debug("Created card",
			zap.String("card_id", cardID),
			zap.String("card_type", expected.CardType),
		)
	}

	return stats, nil
}

// createActiveBedCardWithUnboundDevices Scenario A: Create 1 ActiveBed card, bind all devices
func (c *CardCreator) createActiveBedCardWithUnboundDevices(
	tenantID string,
	unitInfo *UnitInfo,
	bed ActiveBedInfo,
) error {
	// 1. Get devices bound to this bed
	bedDevices, err := c.repo.GetDevicesByBed(tenantID, bed.BedID)
	if err != nil {
		return fmt.Errorf("failed to get bed devices: %w", err)
	}

	// 2. Get unbound devices under this unit
	unboundDevices, err := c.repo.GetUnboundDevicesByUnit(tenantID, unitInfo.UnitID)
	if err != nil {
		return fmt.Errorf("failed to get unbound devices: %w", err)
	}

	// 3. Merge all devices (bed devices + unbound devices)
	allDevices := append(bedDevices, unboundDevices...)

	// 4. Calculate card name
	cardName, err := c.calculateActiveBedCardName(tenantID, bed, unitInfo)
	if err != nil {
		return fmt.Errorf("failed to calculate card name: %w", err)
	}

	// 5. Calculate card address
	cardAddress := c.calculateCardAddress(unitInfo)

	// 6. Get resident information
	resident, err := c.repo.GetResidentByBed(tenantID, bed.BedID)
	if err != nil {
		return fmt.Errorf("failed to get resident: %w", err)
	}

	var residentID *string
	var residents []ResidentInfo
	if resident != nil {
		residentID = &resident.ResidentID
		residents = []ResidentInfo{*resident}
	} else {
		// If bed is not bound to resident, get residents under unit
		unitResidents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
		if err != nil {
			return fmt.Errorf("failed to get unit residents: %w", err)
		}
		residents = unitResidents
	}

	// 7. Convert to JSON
	devicesJSON, err := ConvertDevicesToJSON(allDevices)
	if err != nil {
		return fmt.Errorf("failed to convert devices to JSON: %w", err)
	}

	residentsJSON, err := ConvertResidentsToJSON(residents)
	if err != nil {
		return fmt.Errorf("failed to convert residents to JSON: %w", err)
	}

	// 8. Create card
	cardID, err := c.repo.CreateCard(
		tenantID,
		"ActiveBed",
		&bed.BedID,
		unitInfo.UnitID,
		cardName,
		cardAddress,
		residentID,
		devicesJSON,
		residentsJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create card: %w", err)
	}

	c.logger.Info("Created ActiveBed card",
		zap.String("card_id", cardID),
		zap.String("bed_id", bed.BedID),
		zap.String("unit_id", unitInfo.UnitID),
		zap.String("card_name", cardName),
		zap.Int("device_count", len(allDevices)),
	)

	return nil
}

// createMultipleActiveBedCards Scenario B: Create multiple ActiveBed cards + optional UnitCard
func (c *CardCreator) createMultipleActiveBedCards(
	tenantID string,
	unitInfo *UnitInfo,
	beds []ActiveBedInfo,
) error {
	// 1. Create card for each ActiveBed
	for _, bed := range beds {
		// Get devices bound to this bed
		bedDevices, err := c.repo.GetDevicesByBed(tenantID, bed.BedID)
		if err != nil {
			return fmt.Errorf("failed to get bed devices: %w", err)
		}

		// Calculate card name
		cardName, err := c.calculateActiveBedCardName(tenantID, bed, unitInfo)
		if err != nil {
			return fmt.Errorf("failed to calculate card name: %w", err)
		}

		// Calculate card address
		cardAddress := c.calculateCardAddress(unitInfo)

		// Get resident information
		resident, err := c.repo.GetResidentByBed(tenantID, bed.BedID)
		if err != nil {
			return fmt.Errorf("failed to get resident: %w", err)
		}

		var residentID *string
		var residents []ResidentInfo
		if resident != nil {
			residentID = &resident.ResidentID
			residents = []ResidentInfo{*resident}
		} else {
			// If bed is not bound to resident, get residents under unit
			unitResidents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
			if err != nil {
				return fmt.Errorf("failed to get unit residents: %w", err)
			}
			residents = unitResidents
		}

		// Convert to JSON
		devicesJSON, err := ConvertDevicesToJSON(bedDevices)
		if err != nil {
			return fmt.Errorf("failed to convert devices to JSON: %w", err)
		}

		residentsJSON, err := ConvertResidentsToJSON(residents)
		if err != nil {
			return fmt.Errorf("failed to convert residents to JSON: %w", err)
		}

		// Create card
		cardID, err := c.repo.CreateCard(
			tenantID,
			"ActiveBed",
			&bed.BedID,
			unitInfo.UnitID,
			cardName,
			cardAddress,
			residentID,
			devicesJSON,
			residentsJSON,
		)
		if err != nil {
			return fmt.Errorf("failed to create card: %w", err)
		}

		c.logger.Info("Created ActiveBed card",
			zap.String("card_id", cardID),
			zap.String("bed_id", bed.BedID),
			zap.String("unit_id", unitInfo.UnitID),
			zap.String("card_name", cardName),
		)
	}

	// 2. Check if there are unbound devices, if yes create UnitCard
	unboundDevices, err := c.repo.GetUnboundDevicesByUnit(tenantID, unitInfo.UnitID)
	if err != nil {
		return fmt.Errorf("failed to get unbound devices: %w", err)
	}

	if len(unboundDevices) > 0 {
		return c.createUnitCard(tenantID, unitInfo)
	}

	return nil
}

// createUnitCard Scenario C: Create UnitCard (only when there are unbound devices)
func (c *CardCreator) createUnitCard(
	tenantID string,
	unitInfo *UnitInfo,
) error {
	// 1. Get unbound devices
	unboundDevices, err := c.repo.GetUnboundDevicesByUnit(tenantID, unitInfo.UnitID)
	if err != nil {
		return fmt.Errorf("failed to get unbound devices: %w", err)
	}

	// 2. If there are no unbound devices, do not create UnitCard
	if len(unboundDevices) == 0 {
		return nil
	}

	// 3. Calculate card name
	cardName, err := c.calculateUnitCardName(tenantID, unitInfo)
	if err != nil {
		return fmt.Errorf("failed to calculate card name: %w", err)
	}

	// 4. Calculate card address
	cardAddress := c.calculateCardAddress(unitInfo)

	// 5. Get resident information
	residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
	if err != nil {
		return fmt.Errorf("failed to get unit residents: %w", err)
	}

	// 6. Convert to JSON
	devicesJSON, err := ConvertDevicesToJSON(unboundDevices)
	if err != nil {
		return fmt.Errorf("failed to convert devices to JSON: %w", err)
	}

	residentsJSON, err := ConvertResidentsToJSON(residents)
	if err != nil {
		return fmt.Errorf("failed to convert residents to JSON: %w", err)
	}

	// 7. Create card
	cardID, err := c.repo.CreateCard(
		tenantID,
		"Location",
		nil, // UnitCard has no bed_id
		unitInfo.UnitID,
		cardName,
		cardAddress,
		nil, // UnitCard has no resident_id
		devicesJSON,
		residentsJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create card: %w", err)
	}

	c.logger.Info("Created UnitCard",
		zap.String("card_id", cardID),
		zap.String("unit_id", unitInfo.UnitID),
		zap.String("card_name", cardName),
		zap.Int("device_count", len(unboundDevices)),
	)

	return nil
}

// calculateActiveBedCardName calculates ActiveBed card name
// Rules:
// 1. If bed is bound to resident → use resident's nickname
// 2. If bed is not bound to resident:
//   - Facility 型 unit (非 publicSpace):
//   - SharedUnit (多人共享): 床上无人 → "Unoccupied"
//   - 独享型 (非 SharedUnit): unit 内无人 → "Unoccupied"
//   - 其他类型: 按原逻辑处理
func (c *CardCreator) calculateActiveBedCardName(
	tenantID string,
	bed ActiveBedInfo,
	unitInfo *UnitInfo,
) (string, error) {
	// 1. Check if bed is bound to resident
	if bed.ResidentID != nil {
		resident, err := c.repo.GetResidentByBed(tenantID, bed.BedID)
		if err != nil {
			return "", fmt.Errorf("failed to get resident: %w", err)
		}
		if resident != nil {
			return resident.Nickname, nil
		}
	}

	// 2. Bed is not bound to resident
	// Check if this is a Facility type unit (not publicSpace)
	isFacility := (unitInfo.UnitType == "Facility" || unitInfo.UnitType == "Institutional") && !unitInfo.IsPublic

	if isFacility {
		// Facility 型 unit (非 publicSpace)
		if unitInfo.IsSharedUnit {
			// SharedUnit (多人共享): 检查床上是否有人
			// bed.ResidentID == nil 表示床上无人
			return "Unoccupied", nil
		} else {
			// 独享型 (非 SharedUnit): 检查 unit 内是否有人
			residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
			if err != nil {
				return "", fmt.Errorf("failed to get unit residents: %w", err)
			}
			if len(residents) == 0 {
				// unit 内无人
				return "Unoccupied", nil
			}
			// unit 内有人，显示第一个 resident nickname
			return residents[0].Nickname, nil
		}
	}

	// 3. 非 Facility 型，按原逻辑处理
	if unitInfo.IsSharedUnit {
		return "disable monitor", nil
	}

	// 4. Non-multi-person room, get first resident's nickname under unit
	residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
	if err != nil {
		return "", fmt.Errorf("failed to get unit residents: %w", err)
	}

	if len(residents) > 0 {
		return residents[0].Nickname, nil
	}

	// 5. If unit has no residents, return default value
	return "Unknown", nil
}

// calculateUnitCardName calculates UnitCard name
// Priority:
// 1. is_public = TRUE → unit_name
// 2. Facility 型 unit (非 publicSpace):
//   - 独享型 (非 SharedUnit): unit 内无人 → "Unoccupied"
//
// 3. is_shared_unit = TRUE → unit_name
// 4. unit_type = 'HomeCare' and unit has residents bound → first resident's nickname
// 5. is_shared_unit = FALSE and unit has residents bound → first resident's nickname
func (c *CardCreator) calculateUnitCardName(
	tenantID string,
	unitInfo *UnitInfo,
) (string, error) {
	// Priority 1: Public space
	if unitInfo.IsPublic {
		return unitInfo.UnitName, nil
	}

	// Priority 2: Facility 型 unit (非 publicSpace)
	isFacility := unitInfo.UnitType == "Facility" || unitInfo.UnitType == "Institutional"
	if isFacility && !unitInfo.IsSharedUnit {
		// 独享型 (非 SharedUnit): 检查 unit 内是否有人
		residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
		if err != nil {
			return "", fmt.Errorf("failed to get unit residents: %w", err)
		}
		if len(residents) == 0 {
			// unit 内无人
			return "Unoccupied", nil
		}
		// unit 内有人，继续后续逻辑
	}

	// Priority 3: Shared unit (multi-person room)
	if unitInfo.IsSharedUnit {
		return unitInfo.UnitName, nil
	}

	// Priority 4 and 5: Get residents under unit
	residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
	if err != nil {
		return "", fmt.Errorf("failed to get unit residents: %w", err)
	}

	if len(residents) > 0 {
		return residents[0].Nickname, nil
	}

	// If no residents, return unit_name
	return unitInfo.UnitName, nil
}

// calculateCardAddress calculates card address
// Rule: branch_name + "-" + building + "-" + unit_name
// Skip empty values or default value "-"
func (c *CardCreator) calculateCardAddress(unitInfo *UnitInfo) string {
	var parts []string

	// branch_name
	if unitInfo.BranchName != "" && unitInfo.BranchName != "-" {
		parts = append(parts, unitInfo.BranchName)
	}

	// building (skip "-")
	if unitInfo.Building != "" && unitInfo.Building != "-" {
		parts = append(parts, unitInfo.Building)
	}

	// unit_name (required)
	if unitInfo.UnitName != "" {
		parts = append(parts, unitInfo.UnitName)
	}

	return strings.Join(parts, "-")
}

// calculateExpectedCards calculates expected cards for a unit without actually creating them
// This is used for comparison to avoid unnecessary card recreation
func (c *CardCreator) calculateExpectedCards(
	tenantID string,
	unitInfo *UnitInfo,
	activeBeds []ActiveBedInfo,
) ([]ExpectedCard, error) {
	var expectedCards []ExpectedCard
	activeBedCount := len(activeBeds)

	if activeBedCount == 0 {
		// Scenario C: Unit has no ActiveBed
		expected, err := c.calculateExpectedUnitCard(tenantID, unitInfo)
		if err != nil {
			return nil, err
		}
		if expected != nil {
			expectedCards = append(expectedCards, *expected)
		}
	} else if activeBedCount == 1 {
		// Scenario A: Unit has only 1 ActiveBed
		expected, err := c.calculateExpectedActiveBedCard(tenantID, unitInfo, activeBeds[0], true)
		if err != nil {
			return nil, err
		}
		if expected != nil {
			expectedCards = append(expectedCards, *expected)
		}
	} else {
		// Scenario B: Unit has multiple ActiveBeds (≥2)
		for _, bed := range activeBeds {
			expected, err := c.calculateExpectedActiveBedCard(tenantID, unitInfo, bed, false)
			if err != nil {
				return nil, err
			}
			if expected != nil {
				expectedCards = append(expectedCards, *expected)
			}
		}
		// Check if UnitCard should be created (if there are unbound devices)
		expected, err := c.calculateExpectedUnitCard(tenantID, unitInfo)
		if err != nil {
			return nil, err
		}
		if expected != nil {
			expectedCards = append(expectedCards, *expected)
		}
	}

	return expectedCards, nil
}

// calculateExpectedActiveBedCard calculates expected ActiveBed card without creating it
func (c *CardCreator) calculateExpectedActiveBedCard(
	tenantID string,
	unitInfo *UnitInfo,
	bed ActiveBedInfo,
	includeUnboundDevices bool,
) (*ExpectedCard, error) {
	// 1. Get devices bound to this bed
	bedDevices, err := c.repo.GetDevicesByBed(tenantID, bed.BedID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bed devices: %w", err)
	}

	var allDevices []DeviceInfo
	if includeUnboundDevices {
		// 2. Get unbound devices under this unit (Scenario A only)
		unboundDevices, err := c.repo.GetUnboundDevicesByUnit(tenantID, unitInfo.UnitID)
		if err != nil {
			return nil, fmt.Errorf("failed to get unbound devices: %w", err)
		}
		allDevices = append(bedDevices, unboundDevices...)
	} else {
		allDevices = bedDevices
	}

	// 3. Calculate card name
	cardName, err := c.calculateActiveBedCardName(tenantID, bed, unitInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate card name: %w", err)
	}

	// 4. Calculate card address
	cardAddress := c.calculateCardAddress(unitInfo)

	// 5. Get resident information
	resident, err := c.repo.GetResidentByBed(tenantID, bed.BedID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resident: %w", err)
	}

	var residentID *string
	var residents []ResidentInfo
	if resident != nil {
		residentID = &resident.ResidentID
		residents = []ResidentInfo{*resident}
	} else {
		// If bed is not bound to resident, get residents under unit
		unitResidents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
		if err != nil {
			return nil, fmt.Errorf("failed to get unit residents: %w", err)
		}
		residents = unitResidents
	}

	// 6. Convert to JSON
	devicesJSON, err := ConvertDevicesToJSON(allDevices)
	if err != nil {
		return nil, fmt.Errorf("failed to convert devices to JSON: %w", err)
	}

	residentsJSON, err := ConvertResidentsToJSON(residents)
	if err != nil {
		return nil, fmt.Errorf("failed to convert residents to JSON: %w", err)
	}

	return &ExpectedCard{
		CardType:      "ActiveBed",
		BedID:         &bed.BedID,
		UnitID:        unitInfo.UnitID,
		CardName:      cardName,
		CardAddress:   cardAddress,
		ResidentID:    residentID,
		DevicesJSON:   devicesJSON,
		ResidentsJSON: residentsJSON,
	}, nil
}

// calculateExpectedUnitCard calculates expected UnitCard without creating it
func (c *CardCreator) calculateExpectedUnitCard(
	tenantID string,
	unitInfo *UnitInfo,
) (*ExpectedCard, error) {
	// 1. Get unbound devices
	unboundDevices, err := c.repo.GetUnboundDevicesByUnit(tenantID, unitInfo.UnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unbound devices: %w", err)
	}

	// 2. If there are no unbound devices, do not create UnitCard
	if len(unboundDevices) == 0 {
		return nil, nil
	}

	// 3. Calculate card name
	cardName, err := c.calculateUnitCardName(tenantID, unitInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate card name: %w", err)
	}

	// 4. Calculate card address
	cardAddress := c.calculateCardAddress(unitInfo)

	// 5. Get resident information
	residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit residents: %w", err)
	}

	// 6. Convert to JSON
	devicesJSON, err := ConvertDevicesToJSON(unboundDevices)
	if err != nil {
		return nil, fmt.Errorf("failed to convert devices to JSON: %w", err)
	}

	residentsJSON, err := ConvertResidentsToJSON(residents)
	if err != nil {
		return nil, fmt.Errorf("failed to convert residents to JSON: %w", err)
	}

	return &ExpectedCard{
		CardType:      "Location",
		BedID:         nil,
		UnitID:        unitInfo.UnitID,
		CardName:      cardName,
		CardAddress:   cardAddress,
		ResidentID:    nil,
		DevicesJSON:   devicesJSON,
		ResidentsJSON: residentsJSON,
	}, nil
}

// compareCards compares existing cards with expected cards
// Returns true if they are identical, false otherwise
func (c *CardCreator) compareCards(
	existing []CardWithContent,
	expected []ExpectedCard,
) bool {
	// 1. Compare count
	if len(existing) != len(expected) {
		return false
	}

	// 2. Sort both lists for consistent comparison
	existingSorted := make([]CardWithContent, len(existing))
	copy(existingSorted, existing)
	sort.Slice(existingSorted, func(i, j int) bool {
		if existingSorted[i].CardType != existingSorted[j].CardType {
			return existingSorted[i].CardType < existingSorted[j].CardType
		}
		bedIDI := ""
		bedIDJ := ""
		if existingSorted[i].BedID != nil {
			bedIDI = *existingSorted[i].BedID
		}
		if existingSorted[j].BedID != nil {
			bedIDJ = *existingSorted[j].BedID
		}
		return bedIDI < bedIDJ
	})

	expectedSorted := make([]ExpectedCard, len(expected))
	copy(expectedSorted, expected)
	sort.Slice(expectedSorted, func(i, j int) bool {
		if expectedSorted[i].CardType != expectedSorted[j].CardType {
			return expectedSorted[i].CardType < expectedSorted[j].CardType
		}
		bedIDI := ""
		bedIDJ := ""
		if expectedSorted[i].BedID != nil {
			bedIDI = *expectedSorted[i].BedID
		}
		if expectedSorted[j].BedID != nil {
			bedIDJ = *expectedSorted[j].BedID
		}
		return bedIDI < bedIDJ
	})

	// 3. Compare each card
	for i := range existingSorted {
		existingCard := existingSorted[i]
		expectedCard := expectedSorted[i]

		// Compare basic fields
		if existingCard.CardType != expectedCard.CardType {
			return false
		}
		if !compareStringPtr(existingCard.BedID, expectedCard.BedID) {
			return false
		}
		if existingCard.UnitID != expectedCard.UnitID {
			return false
		}
		if existingCard.CardName != expectedCard.CardName {
			return false
		}
		if existingCard.CardAddress != expectedCard.CardAddress {
			return false
		}
		if !compareStringPtr(existingCard.ResidentID, expectedCard.ResidentID) {
			return false
		}

		// Compare JSONB fields (devices and residents)
		if !compareJSONB(existingCard.DevicesJSON, expectedCard.DevicesJSON) {
			return false
		}
		if !compareJSONB(existingCard.ResidentsJSON, expectedCard.ResidentsJSON) {
			return false
		}
	}

	return true
}

// compareStringPtr compares two string pointers (handles nil)
func compareStringPtr(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// compareJSONB compares two JSONB values by parsing and comparing content
// This handles JSON key ordering differences
func compareJSONB(a, b json.RawMessage) bool {
	// Handle empty/nil cases
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) == 0 || len(b) == 0 {
		return false
	}

	// Parse both JSON values
	var jsonA, jsonB interface{}
	if err := json.Unmarshal(a, &jsonA); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &jsonB); err != nil {
		return false
	}

	// Re-marshal both to normalize JSON (handles key ordering)
	normalizedA, err := json.Marshal(jsonA)
	if err != nil {
		return false
	}
	normalizedB, err := json.Marshal(jsonB)
	if err != nil {
		return false
	}

	return string(normalizedA) == string(normalizedB)
}

// diffCards finds differences between existing and expected cards
// Returns: cards to delete, cards to create, cards to update
// A card is considered "to update" if it exists in both but content differs
func (c *CardCreator) diffCards(
	existing []CardWithContent,
	expected []ExpectedCard,
) (toDelete []CardWithContent, toCreate []ExpectedCard, toUpdate []CardWithContent) {
	// Create maps for efficient lookup
	existingMap := make(map[string]CardWithContent)
	expectedMap := make(map[string]*ExpectedCard)

	// Index existing cards by key (card_type + bed_id)
	for _, card := range existing {
		key := cardKey(card.CardType, card.BedID, card.UnitID)
		existingMap[key] = card
	}

	// Index expected cards by key
	for i := range expected {
		key := cardKey(expected[i].CardType, expected[i].BedID, expected[i].UnitID)
		expectedMap[key] = &expected[i]
	}

	// Find cards to delete (exist in existing but not in expected)
	for _, card := range existing {
		key := cardKey(card.CardType, card.BedID, card.UnitID)
		if _, exists := expectedMap[key]; !exists {
			toDelete = append(toDelete, card)
		}
	}

	// Find cards to create or update
	for key, expectedCard := range expectedMap {
		if existingCard, exists := existingMap[key]; exists {
			// Card exists in both, check if content differs
			if c.cardContentDiffers(existingCard, *expectedCard) {
				// Content differs, need to update (preserve card_id)
				toUpdate = append(toUpdate, existingCard)
				// Note: We don't add to toCreate because we'll use UpdateCard instead
			}
			// If content is the same, do nothing (preserve card_id)
		} else {
			// Card doesn't exist, need to create
			toCreate = append(toCreate, *expectedCard)
		}
	}

	return toDelete, toCreate, toUpdate
}

// cardKey generates a unique key for a card based on card_type, bed_id, and unit_id
// For ActiveBed cards: "ActiveBed:<bed_id>"
// For Location cards: "Location:<unit_id>"
func cardKey(cardType string, bedID *string, unitID string) string {
	if cardType == "ActiveBed" && bedID != nil {
		return fmt.Sprintf("ActiveBed:%s", *bedID)
	}
	return fmt.Sprintf("Location:%s", unitID)
}

// cardKeyBedID returns bed_id as string for logging (handles nil)
func cardKeyBedID(bedID *string) string {
	if bedID == nil {
		return "<nil>"
	}
	return *bedID
}

// cardContentDiffers checks if card content differs (excluding card_id)
func (c *CardCreator) cardContentDiffers(
	existing CardWithContent,
	expected ExpectedCard,
) bool {
	// Compare basic fields
	if existing.CardType != expected.CardType {
		return true
	}
	if !compareStringPtr(existing.BedID, expected.BedID) {
		return true
	}
	if existing.UnitID != expected.UnitID {
		return true
	}
	if existing.CardName != expected.CardName {
		return true
	}
	if existing.CardAddress != expected.CardAddress {
		return true
	}
	if !compareStringPtr(existing.ResidentID, expected.ResidentID) {
		return true
	}

	// Compare JSONB fields (devices and residents)
	if !compareJSONB(existing.DevicesJSON, expected.DevicesJSON) {
		return true
	}
	if !compareJSONB(existing.ResidentsJSON, expected.ResidentsJSON) {
		return true
	}

	return false
}
