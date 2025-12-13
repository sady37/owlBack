package aggregator_test

import (
	"context"
	"encoding/json"
	"testing"

	agg "wisefido-card-aggregator/internal/aggregator"
	"wisefido-card-aggregator/internal/config"
	"wisefido-card-aggregator/internal/models"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCacheManager_UpdateFullCardCache_WritesJSON(t *testing.T) {
	kv := newFakeKVStore()
	cfg := &config.Config{}
	logger := zap.NewNop()

	cm := agg.NewCacheManager(cfg, kv, logger)

	cardID := "card-1"
	v := &models.VitalFocusCard{
		CardID:    cardID,
		TenantID:  "tenant-1",
		CardType:  "Location",
		CardName:  "UnitCard",
		CardAddress: "Addr",
		DeviceCount: 1,
		ResidentCount: 0,
	}

	err := cm.UpdateFullCardCache(context.Background(), cardID, v)
	require.NoError(t, err)

	raw, err := kv.Get(context.Background(), "vital-focus:card:card-1:full")
	require.NoError(t, err)

	var decoded models.VitalFocusCard
	require.NoError(t, json.Unmarshal([]byte(raw), &decoded))
	require.Equal(t, "card-1", decoded.CardID)
	require.Equal(t, "UnitCard", decoded.CardName)
}


