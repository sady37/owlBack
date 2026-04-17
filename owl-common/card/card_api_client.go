package card

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CardAPIClient queries device baseline from wisefido-data HTTP API.
// Drop-in replacement for CardDB in gateway services — no DB dependency.
type CardAPIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewCardAPIClient(baseURL string) *CardAPIClient {
	return &CardAPIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *CardAPIClient) LookupBaseline(ctx context.Context, deviceKey string) (*DeviceBaseline, error) {
	if deviceKey == "" {
		return nil, fmt.Errorf("empty device key")
	}
	u := fmt.Sprintf("%s/internal/device-baseline/%s", c.baseURL, url.PathEscape(deviceKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("device not found: %s", deviceKey)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for device %s", resp.StatusCode, deviceKey)
	}
	var b DeviceBaseline
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &b, nil
}

func (c *CardAPIClient) ListBaselines(ctx context.Context, deviceTypes []string) ([]DeviceBaseline, error) {
	u := fmt.Sprintf("%s/internal/device-baselines?device_type=%s",
		c.baseURL, url.QueryEscape(strings.Join(deviceTypes, ",")))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var baselines []DeviceBaseline
	if err := json.NewDecoder(resp.Body).Decode(&baselines); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return baselines, nil
}

func (c *CardAPIClient) LoadCodeToUIDMap(ctx context.Context, deviceTypes []string) (map[string]string, error) {
	baselines, err := c.ListBaselines(ctx, deviceTypes)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(baselines))
	for _, b := range baselines {
		if b.DeviceCode != "" && b.DeviceUID != "" {
			m[b.DeviceCode] = b.DeviceUID
		}
	}
	return m, nil
}
