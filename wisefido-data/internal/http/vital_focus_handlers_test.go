package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"wisefido-data/internal/store"

	"go.uber.org/zap"
)

type fakeKV struct {
	data map[string]string
}

func (f *fakeKV) Get(ctx context.Context, key string) (string, error) {
	v, ok := f.data[key]
	if !ok {
		return "", store.ErrMiss
	}
	return v, nil
}

func (f *fakeKV) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	f.data[key] = value
	return nil
}

func (f *fakeKV) ScanKeys(ctx context.Context, pattern string) ([]string, error) {
	// for tests, return all keys regardless of pattern
	keys := make([]string, 0, len(f.data))
	for k := range f.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func TestGetCards_WrapsResultAndPaginates(t *testing.T) {
	logger := zap.NewNop()
	kv := &fakeKV{data: map[string]string{}}

	// full cache produced by card-aggregator: device_type is string, heart_source is "Sleepace"
	kv.data["vital-focus:card:card-1:full"] = `{
	  "card_id":"card-1","tenant_id":"t1","card_type":"ActiveBed",
	  "card_name":"A","card_address":"Addr",
	  "residents":[{"resident_id":"r1","nickname":"Smith"}],
	  "devices":[{"device_id":"d1","device_name":"Radar01","device_type":"Radar"}],
	  "device_count":1,"resident_count":1,
	  "heart":70,"breath":18,"heart_source":"Sleepace","breath_source":"Radar"
	}`
	kv.data["vital-focus:card:card-2:full"] = `{
	  "card_id":"card-2","tenant_id":"t2","card_type":"Location",
	  "card_name":"B","card_address":"Addr",
	  "residents":[],"devices":[],
	  "device_count":0,"resident_count":0
	}`

	h := NewVitalFocusHandler(kv, logger)

	req := httptest.NewRequest(http.MethodGet, "/data/api/v1/data/vital-focus/cards?page=1&pageSize=1&tenant_id=t1", nil)
	w := httptest.NewRecorder()
	h.GetCards(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `"code":2000`) {
		t.Fatalf("expected wrapper code=2000, got: %s", body)
	}
	// tenant filter + pagination => only card-1
	if !strings.Contains(body, `"card_id":"card-1"`) || strings.Contains(body, `"card_id":"card-2"`) {
		t.Fatalf("expected only card-1, got: %s", body)
	}
	// normalization: device_type => number 2, sources => s/r
	if !strings.Contains(body, `"device_type":2`) {
		t.Fatalf("expected device_type normalized to 2, got: %s", body)
	}
	if !strings.Contains(body, `"heart_source":"s"`) || !strings.Contains(body, `"breath_source":"r"`) {
		t.Fatalf("expected source normalized, got: %s", body)
	}
}

func TestGetCardByIDOrResident_TriesCardThenResident(t *testing.T) {
	logger := zap.NewNop()
	kv := &fakeKV{data: map[string]string{}}

	kv.data["vital-focus:card:card-1:full"] = `{
	  "card_id":"card-1","tenant_id":"t1","card_type":"ActiveBed",
	  "card_name":"A","card_address":"Addr",
	  "primary_resident_id":"resident-1",
	  "residents":[{"resident_id":"resident-1","nickname":"Smith"}],
	  "devices":[],
	  "device_count":0,"resident_count":1
	}`

	h := NewVitalFocusHandler(kv, logger)

	// by card id
	req := httptest.NewRequest(http.MethodGet, "/data/api/v1/data/vital-focus/card/card-1", nil)
	w := httptest.NewRecorder()
	h.GetCardByIDOrResident(w, req, "card-1")
	if !strings.Contains(w.Body.String(), `"card_id":"card-1"`) {
		t.Fatalf("expected card detail, got: %s", w.Body.String())
	}

	// by resident id (should scan)
	w2 := httptest.NewRecorder()
	h.GetCardByIDOrResident(w2, req, "resident-1")
	if !strings.Contains(w2.Body.String(), `"card_id":"card-1"`) {
		t.Fatalf("expected card found by resident, got: %s", w2.Body.String())
	}
}


