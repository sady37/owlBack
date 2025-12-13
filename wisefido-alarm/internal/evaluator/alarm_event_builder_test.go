package evaluator

import (
	"encoding/json"
	"testing"

	"wisefido-alarm/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlarmEventBuilder_BuildAlarmEvent(t *testing.T) {
	builder := NewAlarmEventBuilder("tenant-123", "device-456")

	triggerData := BuildTriggerData(
		"Fall",
		"Radar",
		intPtr(72),
		intPtr(18),
		stringPtr("Standing"),
		stringPtr("Standing position"),
		stringPtr("161898004"),
		stringPtr("Fall"),
		intPtr(85),
		intPtr(60),
	)

	metadata := map[string]interface{}{
		"trigger_source": "cloud",
		"card_id":        "card-789",
	}

	event, err := builder.BuildAlarmEvent(
		"Fall",
		"safety",
		"ALERT",
		triggerData,
		metadata,
	)

	require.NoError(t, err)
	assert.NotNil(t, event)
	assert.NotEmpty(t, event.EventID)
	assert.Equal(t, "tenant-123", event.TenantID)
	assert.Equal(t, "device-456", event.DeviceID)
	assert.Equal(t, "Fall", event.EventType)
	assert.Equal(t, "safety", event.Category)
	assert.Equal(t, "ALERT", event.AlarmLevel)
	assert.Equal(t, "active", event.AlarmStatus)

	// 验证 trigger_data 序列化
	var triggerDataParsed models.TriggerData
	err = json.Unmarshal([]byte(event.TriggerData), &triggerDataParsed)
	require.NoError(t, err)
	assert.Equal(t, "Fall", triggerDataParsed.EventType)
	assert.Equal(t, "Radar", triggerDataParsed.Source)
	assert.Equal(t, intPtr(72), triggerDataParsed.HeartRate)

	// 验证 metadata 序列化
	var metadataParsed map[string]interface{}
	err = json.Unmarshal([]byte(event.Metadata), &metadataParsed)
	require.NoError(t, err)
	assert.Equal(t, "cloud", metadataParsed["trigger_source"])
	assert.Equal(t, "card-789", metadataParsed["card_id"])

	// 验证 notified_users 默认为空数组
	assert.Equal(t, "[]", event.NotifiedUsers)
}

func TestBuildTriggerData(t *testing.T) {
	triggerData := BuildTriggerData(
		"Fall",
		"Radar",
		intPtr(72),
		intPtr(18),
		stringPtr("Standing"),
		stringPtr("Standing position"),
		stringPtr("161898004"),
		stringPtr("Fall"),
		intPtr(85),
		intPtr(60),
	)

	assert.NotNil(t, triggerData)
	assert.Equal(t, "Fall", triggerData.EventType)
	assert.Equal(t, "Radar", triggerData.Source)
	assert.Equal(t, intPtr(72), triggerData.HeartRate)
	assert.Equal(t, intPtr(18), triggerData.RespiratoryRate)
	assert.Equal(t, stringPtr("Standing"), triggerData.Posture)
	assert.Equal(t, stringPtr("Standing position"), triggerData.PostureDisplay)
	assert.Equal(t, stringPtr("161898004"), triggerData.SNOMEDCode)
	assert.Equal(t, stringPtr("Fall"), triggerData.SNOMEDDisplay)
	assert.Equal(t, intPtr(85), triggerData.Confidence)
	assert.Equal(t, intPtr(60), triggerData.DurationSec)
}

func TestBuildTriggerData_WithNilValues(t *testing.T) {
	triggerData := BuildTriggerData(
		"Fall",
		"Radar",
		nil, // heart_rate
		nil, // respiratory_rate
		nil, // posture
		nil, // posture_display
		nil, // snomed_code
		nil, // snomed_display
		nil, // confidence
		nil, // duration_sec
	)

	assert.NotNil(t, triggerData)
	assert.Equal(t, "Fall", triggerData.EventType)
	assert.Equal(t, "Radar", triggerData.Source)
	assert.Nil(t, triggerData.HeartRate)
	assert.Nil(t, triggerData.RespiratoryRate)
}

// 辅助函数
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
