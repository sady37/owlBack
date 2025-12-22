package aggregator

import (
	"fmt"
	"strings"
	"wisefido-card-aggregator/internal/repository"

	"go.uber.org/zap"
)

// CardRepositoryInterface defines card repository interface (for test mocking)
type CardRepositoryInterface interface {
	GetUnitInfo(tenantID, unitID string) (*repository.UnitInfo, error)
	GetActiveBedsByUnit(tenantID, unitID string) ([]repository.ActiveBedInfo, error)
	GetDevicesByBed(tenantID, bedID string) ([]repository.DeviceInfo, error)
	GetUnboundDevicesByUnit(tenantID, unitID string) ([]repository.DeviceInfo, error)
	GetResidentByBed(tenantID, bedID string) (*repository.ResidentInfo, error)
	GetResidentsByUnit(tenantID, unitID string) ([]repository.ResidentInfo, error)
	DeleteCardsByUnit(tenantID, unitID string) error
	CreateCard(
		tenantID, cardType string,
		bedID *string, unitID, cardName, cardAddress string,
		residentID *string,
		devicesJSON, residentsJSON []byte,
	) (string, error)
	GetAllUnits(tenantID string) ([]string, error)
	GetUnitIDByBedID(tenantID, bedID string) (string, error)
}

// CardCreator card creator
type CardCreator struct {
	repo   CardRepositoryInterface
	logger *zap.Logger
}

// NewCardCreator creates a new card creator
func NewCardCreator(repo CardRepositoryInterface, logger *zap.Logger) *CardCreator {
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
func (c *CardCreator) CreateCardsForUnit(tenantID, unitID string) error {
	// 1. Get unit information
	unitInfo, err := c.repo.GetUnitInfo(tenantID, unitID)
	if err != nil {
		return fmt.Errorf("failed to get unit info: %w", err)
	}

	// 2. Get all ActiveBeds under this unit
	activeBeds, err := c.repo.GetActiveBedsByUnit(tenantID, unitID)
	if err != nil {
		return fmt.Errorf("failed to get active beds: %w", err)
	}

	// 3. Delete all old cards under this unit (recreate)
	if err := c.repo.DeleteCardsByUnit(tenantID, unitID); err != nil {
		return fmt.Errorf("failed to delete old cards: %w", err)
	}

	// 4. Determine scenario based on ActiveBed count
	activeBedCount := len(activeBeds)

	if activeBedCount == 0 {
		// Scenario C: Unit has no ActiveBed
		return c.createUnitCard(tenantID, unitInfo)
	} else if activeBedCount == 1 {
		// Scenario A: Unit has only 1 ActiveBed
		return c.createActiveBedCardWithUnboundDevices(tenantID, unitInfo, activeBeds[0])
	} else {
		// Scenario B: Unit has multiple ActiveBeds (≥2)
		return c.createMultipleActiveBedCards(tenantID, unitInfo, activeBeds)
	}
}

// createActiveBedCardWithUnboundDevices Scenario A: Create 1 ActiveBed card, bind all devices
func (c *CardCreator) createActiveBedCardWithUnboundDevices(
	tenantID string,
	unitInfo *repository.UnitInfo,
	bed repository.ActiveBedInfo,
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
	var residents []repository.ResidentInfo
	if resident != nil {
		residentID = &resident.ResidentID
		residents = []repository.ResidentInfo{*resident}
	} else {
		// If bed is not bound to resident, get residents under unit
		unitResidents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
		if err != nil {
			return fmt.Errorf("failed to get unit residents: %w", err)
		}
		residents = unitResidents
	}

	// 7. Convert to JSON
	devicesJSON, err := repository.ConvertDevicesToJSON(allDevices)
	if err != nil {
		return fmt.Errorf("failed to convert devices to JSON: %w", err)
	}

	residentsJSON, err := repository.ConvertResidentsToJSON(residents)
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
	unitInfo *repository.UnitInfo,
	beds []repository.ActiveBedInfo,
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
		var residents []repository.ResidentInfo
		if resident != nil {
			residentID = &resident.ResidentID
			residents = []repository.ResidentInfo{*resident}
		} else {
			// If bed is not bound to resident, get residents under unit
			unitResidents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
			if err != nil {
				return fmt.Errorf("failed to get unit residents: %w", err)
			}
			residents = unitResidents
		}

		// Convert to JSON
		devicesJSON, err := repository.ConvertDevicesToJSON(bedDevices)
		if err != nil {
			return fmt.Errorf("failed to convert devices to JSON: %w", err)
		}

		residentsJSON, err := repository.ConvertResidentsToJSON(residents)
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
	unitInfo *repository.UnitInfo,
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
	devicesJSON, err := repository.ConvertDevicesToJSON(unboundDevices)
	if err != nil {
		return fmt.Errorf("failed to convert devices to JSON: %w", err)
	}

	residentsJSON, err := repository.ConvertResidentsToJSON(residents)
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
//   - Non-multi-person room → show first resident's nickname under unit
//   - Multi-person room → show 'disable monitor'
func (c *CardCreator) calculateActiveBedCardName(
	tenantID string,
	bed repository.ActiveBedInfo,
	unitInfo *repository.UnitInfo,
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

	// 2. Bed is not bound to resident, decide based on unit's is_multi_person_room
	if unitInfo.IsMultiPersonRoom {
		return "disable monitor", nil
	}

	// 3. Non-multi-person room, get first resident's nickname under unit
	residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
	if err != nil {
		return "", fmt.Errorf("failed to get unit residents: %w", err)
	}

	if len(residents) > 0 {
		return residents[0].Nickname, nil
	}

	// 4. If unit has no residents, return default value
	return "Unknown", nil
}

// calculateUnitCardName calculates UnitCard name
// Priority:
// 1. is_public_space = TRUE → unit_name
// 2. is_multi_person_room = TRUE → unit_name
// 3. unit_type = 'HomeCare' and unit has residents bound → first resident's nickname
// 4. is_multi_person_room = FALSE and unit has residents bound → first resident's nickname
func (c *CardCreator) calculateUnitCardName(
	tenantID string,
	unitInfo *repository.UnitInfo,
) (string, error) {
	// Priority 1: Public space
	if unitInfo.IsPublicSpace {
		return unitInfo.UnitName, nil
	}

	// Priority 2: Multi-person room
	if unitInfo.IsMultiPersonRoom {
		return unitInfo.UnitName, nil
	}

	// Priority 3 and 4: Get residents under unit
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
func (c *CardCreator) calculateCardAddress(unitInfo *repository.UnitInfo) string {
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
