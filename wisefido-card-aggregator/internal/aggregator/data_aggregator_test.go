package aggregator_test

import (
	"context"
	"encoding/json"
	"testing"

	agg "wisefido-card-aggregator/internal/aggregator"
	"wisefido-card-aggregator/internal/config"
	"wisefido-card-aggregator/internal/models"
	"wisefido-card-aggregator/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDataAggregator_AggregateCard_MergesDBAndCaches(t *testing.T) {
	// mock db
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := zap.NewNop()
	cardRepo := repository.NewCardRepository(db, logger)

	// fake redis kv
	kv := newFakeKVStore()

	cfg := &config.Config{}

	aggregator := agg.NewDataAggregator(cfg, kv, cardRepo, logger)

	tenantID := "tenant-1"
	cardID := "card-1"

	// 1) GetCardByID
	mock.ExpectQuery(`SELECT\s+card_id`).
		WithArgs(cardID, tenantID).
		WillReturnRows(sqlmock.NewRows([]string{
			"card_id", "tenant_id", "card_type", "bed_id", "unit_id", "card_name", "card_address", "resident_id",
			"unhandled_alarm_0", "unhandled_alarm_1", "unhandled_alarm_2", "unhandled_alarm_3", "unhandled_alarm_4",
			"icon_alarm_level", "pop_alarm_emerge",
		}).AddRow(
			cardID, tenantID, "ActiveBed", "bed-1", "unit-1", "BedCard", "Addr", "resident-1",
			0, 0, 0, 0, 1,
			3, 0,
		))

	// 2) GetCardDevices
	devicesJSON := []map[string]any{
		{
			"device_id":    "device-1",
			"device_name":  "Radar01",
			"device_type":  "Radar",
			"device_model": "M1",
			"bed_id":       "bed-1",
			"bed_name":     "BedName",
			"room_id":      nil,
			"room_name":    nil,
			"unit_id":      "unit-1",
		},
	}
	devicesBytes, _ := json.Marshal(devicesJSON)
	mock.ExpectQuery(`SELECT\s+devices`).
		WithArgs(cardID).
		WillReturnRows(sqlmock.NewRows([]string{"devices"}).AddRow(devicesBytes))

	// 3) GetCardResidents
	residentsJSON := []map[string]any{
		{"resident_id": "resident-1", "nickname": "Nick"},
	}
	residentsBytes, _ := json.Marshal(residentsJSON)
	mock.ExpectQuery(`SELECT\s+residents`).
		WithArgs(cardID).
		WillReturnRows(sqlmock.NewRows([]string{"residents"}).AddRow(residentsBytes))

	// 4) fake redis realtime + alarms
	realtime := map[string]any{
		"heart":        70,
		"breath":       18,
		"heart_source": "Sleepace",
		"breath_source": "Sleepace",
		"person_count":  1,
		"timestamp":     1700000000,
	}
	rtBytes, _ := json.Marshal(realtime)
	require.NoError(t, kv.Set(context.Background(), "vital-focus:card:card-1:realtime", string(rtBytes), 0))

	alarms := []map[string]any{
		{
			"event_id":      "alarm-1",
			"event_type":    "Fall",
			"category":      "safety",
			"alarm_level":   "ALERT",
			"alarm_status":  "active",
			"triggered_at":  1700000001,
			"trigger_data":  map[string]any{"confidence": 80},
		},
	}
	alarmsBytes, _ := json.Marshal(alarms)
	require.NoError(t, kv.Set(context.Background(), "vital-focus:card:card-1:alarms", string(alarmsBytes), 0))

	// aggregate
	out, err := aggregator.AggregateCard(context.Background(), tenantID, cardID)
	require.NoError(t, err)
	require.NotNil(t, out)

	require.Equal(t, cardID, out.CardID)
	require.Equal(t, tenantID, out.TenantID)
	require.Equal(t, "ActiveBed", out.CardType)
	require.NotNil(t, out.PrimaryResidentID)
	require.Equal(t, "resident-1", *out.PrimaryResidentID)

	// realtime merged
	require.NotNil(t, out.Heart)
	require.Equal(t, 70, *out.Heart)
	require.NotNil(t, out.Breath)
	require.Equal(t, 18, *out.Breath)
	require.NotNil(t, out.HeartSource)
	require.Equal(t, "s", *out.HeartSource)

	// alarms merged
	require.Len(t, out.Alarms, 1)
	require.Equal(t, "alarm-1", out.Alarms[0].EventID)
	require.Equal(t, "Fall", out.Alarms[0].EventType)

	// counts
	require.Equal(t, 1, out.DeviceCount)
	require.Equal(t, 1, out.ResidentCount)

	// unhandled alarms
	require.NotNil(t, out.UnhandledAlarm4)
	require.Equal(t, 1, *out.UnhandledAlarm4)
	require.NotNil(t, out.TotalUnhandledAlarms)
	require.Equal(t, 1, *out.TotalUnhandledAlarms)

	require.NoError(t, mock.ExpectationsWereMet())
	_ = models.VitalFocusCard{} // keep import alive for future expansions
}


