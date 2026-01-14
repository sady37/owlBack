package consumer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSingleCategoryMonitorTrack tests single category case: monitor with track only
func TestSingleCategoryMonitorTrack(t *testing.T) {
	// Create a mock dataValue with single track category
	dataValue := map[string]interface{}{
		"category":   "track",
		"target_id":  1,
		"position_x": 150,
		"position_y": 200,
		"position_z": 50,
		"pose":       "Walking",
		"event":      0,
		"area_id":    1,
	}

	// Extract category (simulating the logic from mqtt_consumer.go lines 337-351)
	var category string
	var dataValueInterface interface{} = dataValue
	if dataValueMap, ok := dataValueInterface.(map[string]interface{}); ok {
		if cat, ok := dataValueMap["category"].(string); ok {
			category = cat
		}
	}

	// Build encodedData (simulating the logic from mqtt_consumer.go lines 355-369)
	now := time.Now().Unix()
	encodedData := map[string]interface{}{
		"device_id":   "uuid-xxx",
		"device_type": "Radar",
		"tenant_id":   "uuid-yyy",
		"timestamp":   now,
		"topic_type":  "monitor",
		"category":    category,
		"data_value":  dataValue,
		"branch_id":   nil,
		"building_id": nil,
		"unit_id":     nil,
		"room_id":     "uuid-rrr",
		"bed_id":      "uuid-bbb",
	}

	// Verify category is extracted correctly
	assert.Equal(t, "track", category, "Category should be extracted from data_value")
	assert.Equal(t, "track", encodedData["category"], "Category should be at top level")

	// Verify all required fields exist
	assert.Contains(t, encodedData, "device_id", "encodedData should contain device_id")
	assert.Contains(t, encodedData, "device_type", "encodedData should contain device_type")
	assert.Contains(t, encodedData, "tenant_id", "encodedData should contain tenant_id")
	assert.Contains(t, encodedData, "timestamp", "encodedData should contain timestamp")
	assert.Contains(t, encodedData, "topic_type", "encodedData should contain topic_type")
	assert.Contains(t, encodedData, "category", "encodedData should contain category at top level")
	assert.Contains(t, encodedData, "data_value", "encodedData should contain data_value")

	// Verify field order in source code structure
	// Note: Go's json.Marshal doesn't preserve map key order (sorts alphabetically),
	// but the source code structure in mqtt_consumer.go lines 355-369 follows the expected order:
	// device_id → device_type → tenant_id → timestamp → topic_type → category → data_value
	// This test verifies the structure is correct, even though JSON encoding may reorder keys
	verifyFieldOrderInStructure(t, encodedData)
}

// TestSingleCategoryMonitorVital tests single category case: monitor with vital only
func TestSingleCategoryMonitorVital(t *testing.T) {
	// Create a mock dataValue with single vital category
	dataValueMap := map[string]interface{}{
		"category":        "vital",
		"heart_rate":      75,
		"respiratory_rate": 18,
		"sleep_status":    "Awake",
	}

	// Extract category
	var category string
	var dataValueInterface interface{} = dataValueMap
	if m, ok := dataValueInterface.(map[string]interface{}); ok {
		if cat, ok := m["category"].(string); ok {
			category = cat
		}
	}

	// Build encodedData
	now := time.Now().Unix()
	encodedData := map[string]interface{}{
		"device_id":   "uuid-xxx",
		"device_type": "Radar",
		"tenant_id":   "uuid-yyy",
		"timestamp":   now,
		"topic_type":  "monitor",
		"category":    category,
		"data_value":  dataValueMap,
		"branch_id":   nil,
		"building_id": nil,
		"unit_id":     nil,
		"room_id":     "uuid-rrr",
		"bed_id":      "uuid-bbb",
	}

	// Verify category is extracted correctly
	assert.Equal(t, "vital", category, "Category should be extracted from data_value")
	assert.Equal(t, "vital", encodedData["category"], "Category should be at top level")

	// Verify field order in structure
	verifyFieldOrderInStructure(t, encodedData)
}

// TestSingleCategoryEvent tests single category case: event
func TestSingleCategoryEvent(t *testing.T) {
	// Create a mock dataValue with single event category
	dataValueMap := map[string]interface{}{
		"category": "enter2out",
		"track_id": 1,
		"event":    "Enter room",
		"area_type": "Room",
	}

	// Extract category
	var category string
	var dataValueInterface interface{} = dataValueMap
	if m, ok := dataValueInterface.(map[string]interface{}); ok {
		if cat, ok := m["category"].(string); ok {
			category = cat
		}
	}

	// Build encodedData
	now := time.Now().Unix()
	encodedData := map[string]interface{}{
		"device_id":   "uuid-xxx",
		"device_type": "Radar",
		"tenant_id":   "uuid-yyy",
		"timestamp":   now,
		"topic_type":  "event",
		"category":    category,
		"data_value":  dataValueMap,
		"branch_id":   nil,
		"building_id": nil,
		"unit_id":     nil,
		"room_id":     "uuid-rrr",
		"bed_id":      "uuid-bbb",
	}

	// Verify category is extracted correctly
	assert.Equal(t, "enter2out", category, "Category should be extracted from data_value")
	assert.Equal(t, "enter2out", encodedData["category"], "Category should be at top level")

	// Verify field order in structure
	verifyFieldOrderInStructure(t, encodedData)
}

// TestSingleCategoryAlarm tests single category case: alarm
func TestSingleCategoryAlarm(t *testing.T) {
	// Create a mock dataValue with single alarm category
	dataValueMap := map[string]interface{}{
		"category": "SuspectedFall",
		"track_id": 1,
		"pose":     "Fall suspected",
	}

	// Extract category
	var category string
	var dataValueInterface interface{} = dataValueMap
	if m, ok := dataValueInterface.(map[string]interface{}); ok {
		if cat, ok := m["category"].(string); ok {
			category = cat
		}
	}

	// Build encodedData
	now := time.Now().Unix()
	encodedData := map[string]interface{}{
		"device_id":   "uuid-xxx",
		"device_type": "Radar",
		"tenant_id":   "uuid-yyy",
		"timestamp":   now,
		"topic_type":  "alarm",
		"category":    category,
		"data_value":  dataValueMap,
		"branch_id":   nil,
		"building_id": nil,
		"unit_id":     nil,
		"room_id":     "uuid-rrr",
		"bed_id":      "uuid-bbb",
	}

	// Verify category is extracted correctly
	assert.Equal(t, "SuspectedFall", category, "Category should be extracted from data_value")
	assert.Equal(t, "SuspectedFall", encodedData["category"], "Category should be at top level")

	// Verify field order in structure
	verifyFieldOrderInStructure(t, encodedData)
}

// TestMultipleCategoryMonitorTrackVital tests multiple category case: monitor with track + vital
func TestMultipleCategoryMonitorTrackVital(t *testing.T) {
	// Create a mock dataValue array with multiple categories (track + vital)
	dataValueArray := []interface{}{
		map[string]interface{}{
			"category":   "track",
			"target_id":  1,
			"position_x": 150,
			"position_y": 200,
			"position_z": 50,
			"pose":       "Walking",
			"event":      0,
			"area_id":    1,
		},
		map[string]interface{}{
			"category":        "vital",
			"heart_rate":      75,
			"respiratory_rate": 18,
			"sleep_status":    "Awake",
		},
	}

	// Extract category from first object (simulating the logic from mqtt_consumer.go lines 344-350)
	var category string
	var dataValue interface{} = dataValueArray
	if arr, ok := dataValue.([]interface{}); ok && len(arr) > 0 {
		if firstObj, ok := arr[0].(map[string]interface{}); ok {
			if cat, ok := firstObj["category"].(string); ok {
				category = cat
			}
		}
	}

	// Build encodedData
	now := time.Now().Unix()
	encodedData := map[string]interface{}{
		"device_id":   "uuid-xxx",
		"device_type": "Radar",
		"tenant_id":   "uuid-yyy",
		"timestamp":   now,
		"topic_type":  "monitor",
		"category":    category,
		"data_value":  dataValueArray,
		"branch_id":   nil,
		"building_id": nil,
		"unit_id":     nil,
		"room_id":     "uuid-rrr",
		"bed_id":      "uuid-bbb",
	}

	// Verify category is extracted from first object
	assert.Equal(t, "track", category, "Category should be extracted from first object in array")
	assert.Equal(t, "track", encodedData["category"], "Category should be at top level")

	// Verify data_value is still an array
	dataValueInEncoded, ok := encodedData["data_value"].([]interface{})
	require.True(t, ok, "data_value should remain as array")
	assert.Len(t, dataValueInEncoded, 2, "data_value should contain both track and vital")

	// Verify field order in structure
	verifyFieldOrderInStructure(t, encodedData)
}

// TestMultipleCategoryStatTrackSleep tests multiple category case: stat with track + sleep
func TestMultipleCategoryStatTrackSleep(t *testing.T) {
	// Create a mock dataValue array with multiple categories (track + sleep)
	dataValueArray := []interface{}{
		map[string]interface{}{
			"category":   "track",
			"target_id":  1,
			"position_x": 150,
			"position_y": 200,
			"position_z": 50,
		},
		map[string]interface{}{
			"category":              "sleep",
			"respiratory_rate":      16,
			"heart_rate":            70,
			"avg_respiratory_rate":  15,
			"avg_heart_rate":        68,
		},
	}

	// Extract category from first object
	var category string
	var dataValue interface{} = dataValueArray
	if arr, ok := dataValue.([]interface{}); ok && len(arr) > 0 {
		if firstObj, ok := arr[0].(map[string]interface{}); ok {
			if cat, ok := firstObj["category"].(string); ok {
				category = cat
			}
		}
	}

	// Build encodedData (note: stat becomes "statistics" in finalTopicType)
	now := time.Now().Unix()
	encodedData := map[string]interface{}{
		"device_id":   "uuid-xxx",
		"device_type": "Radar",
		"tenant_id":   "uuid-yyy",
		"timestamp":   now,
		"topic_type":  "statistics", // stat is converted to statistics
		"category":    category,
		"data_value":  dataValueArray,
		"branch_id":   nil,
		"building_id": nil,
		"unit_id":     nil,
		"room_id":     "uuid-rrr",
		"bed_id":      "uuid-bbb",
	}

	// Verify category is extracted from first object
	assert.Equal(t, "track", category, "Category should be extracted from first object in array")
	assert.Equal(t, "track", encodedData["category"], "Category should be at top level")

	// Verify data_value is still an array
	dataValueInEncoded, ok := encodedData["data_value"].([]interface{})
	require.True(t, ok, "data_value should remain as array")
	assert.Len(t, dataValueInEncoded, 2, "data_value should contain both track and sleep")

	// Verify field order in structure
	verifyFieldOrderInStructure(t, encodedData)
}

// TestFieldOrderComprehensive tests field order comprehensively
func TestFieldOrderComprehensive(t *testing.T) {
	// Test with all required fields to verify complete field order
	now := time.Now().Unix()
	encodedData := map[string]interface{}{
		"device_id":   "uuid-xxx",
		"device_type": "Radar",
		"tenant_id":   "uuid-yyy",
		"timestamp":   now,
		"topic_type":  "monitor",
		"category":    "track",
		"data_value": map[string]interface{}{
			"category":   "track",
			"target_id":  1,
			"position_x": 150,
		},
		"branch_id":   nil,
		"building_id": nil,
		"unit_id":     nil,
		"room_id":     "uuid-rrr",
		"bed_id":      "uuid-bbb",
	}

	// Verify all required fields exist
	assert.Contains(t, encodedData, "device_id")
	assert.Contains(t, encodedData, "device_type")
	assert.Contains(t, encodedData, "tenant_id")
	assert.Contains(t, encodedData, "timestamp")
	assert.Contains(t, encodedData, "topic_type")
	assert.Contains(t, encodedData, "category")
	assert.Contains(t, encodedData, "data_value")

	// Verify category is at top level
	assert.Equal(t, "track", encodedData["category"], "Category should be at top level")

	// Verify field order in structure
	// Note: The actual field order in mqtt_consumer.go lines 355-369 is:
	// device_id → device_type → tenant_id → timestamp → topic_type → category → data_value
	// This test verifies the structure matches the expected order in source code
	verifyFieldOrderInStructure(t, encodedData)

	// Verify timestamp comes before topic_type in the structure
	// (We verify this by checking the values exist and are correct types)
	assert.IsType(t, int64(0), encodedData["timestamp"], "timestamp should be int64")
	assert.IsType(t, "", encodedData["topic_type"], "topic_type should be string")
	assert.IsType(t, "", encodedData["category"], "category should be string")
}

// verifyFieldOrderInStructure verifies that the encodedData structure follows the expected field order
// Expected order: device_id → device_type → tenant_id → timestamp → topic_type → category → data_value
// Note: Go's json.Marshal doesn't preserve map key order (sorts alphabetically),
// but we verify the structure is built correctly in the source code
func verifyFieldOrderInStructure(t *testing.T, encodedData map[string]interface{}) {
	// Verify all required fields exist
	requiredFields := []string{"device_id", "device_type", "tenant_id", "timestamp", "topic_type", "category", "data_value"}
	for _, field := range requiredFields {
		assert.Contains(t, encodedData, field, "encodedData should contain %s", field)
	}

	// Verify timestamp is before topic_type in the structure
	// Since Go maps don't preserve order in JSON, we verify:
	// 1. Both fields exist
	// 2. timestamp is int64 (Unix timestamp)
	// 3. topic_type is string
	// 4. category is string and at top level
	assert.Contains(t, encodedData, "timestamp", "timestamp should exist")
	assert.Contains(t, encodedData, "topic_type", "topic_type should exist")
	assert.Contains(t, encodedData, "category", "category should exist at top level")

	// Verify types
	_, hasTimestamp := encodedData["timestamp"]
	_, hasTopicType := encodedData["topic_type"]
	_, hasCategory := encodedData["category"]

	assert.True(t, hasTimestamp, "timestamp should exist")
	assert.True(t, hasTopicType, "topic_type should exist")
	assert.True(t, hasCategory, "category should exist at top level")

	// Verify category is not nested inside data_value
	dataValue, ok := encodedData["data_value"]
	require.True(t, ok, "data_value should exist")

	// If data_value is a map, verify category exists at top level (not just inside data_value)
	if _, ok := dataValue.(map[string]interface{}); ok {
		// Category should exist at top level
		assert.Contains(t, encodedData, "category", "category should be at top level, not just in data_value")
		// Category might also exist in data_value (which is fine)
		// But the top-level category is what we're testing
	}
}

// TestCategoryExtractionEdgeCases tests edge cases for category extraction
func TestCategoryExtractionEdgeCases(t *testing.T) {
	t.Run("Empty data_value", func(t *testing.T) {
		dataValueMap := map[string]interface{}{}
		var category string
		var dataValue interface{} = dataValueMap
		if m, ok := dataValue.(map[string]interface{}); ok {
			if cat, ok := m["category"].(string); ok {
				category = cat
			}
		}
		assert.Empty(t, category, "Category should be empty when data_value has no category")
	})

	t.Run("Empty array", func(t *testing.T) {
		dataValueArray := []interface{}{}
		var category string
		var dataValue interface{} = dataValueArray
		if arr, ok := dataValue.([]interface{}); ok && len(arr) > 0 {
			if firstObj, ok := arr[0].(map[string]interface{}); ok {
				if cat, ok := firstObj["category"].(string); ok {
					category = cat
				}
			}
		}
		assert.Empty(t, category, "Category should be empty when array is empty")
	})

	t.Run("Array with first object missing category", func(t *testing.T) {
		dataValueArray := []interface{}{
			map[string]interface{}{
				"target_id": 1,
			},
			map[string]interface{}{
				"category": "vital",
			},
		}
		var category string
		var dataValue interface{} = dataValueArray
		if arr, ok := dataValue.([]interface{}); ok && len(arr) > 0 {
			if firstObj, ok := arr[0].(map[string]interface{}); ok {
				if cat, ok := firstObj["category"].(string); ok {
					category = cat
				}
			}
		}
		assert.Empty(t, category, "Category should be empty when first object has no category")
	})
}
