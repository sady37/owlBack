package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"wisefido-sleepace/internal/config"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// SleepaceProxy is a transparent HTTP proxy to sleepace-service (Java).
// It auto-injects the token and forwards request/response as-is.
type SleepaceProxy struct {
	client *resty.Client
	appID  string
	secret string
	logger *zap.Logger
}

func NewSleepaceProxy(cfg *config.SleepaceConfig, logger *zap.Logger) *SleepaceProxy {
	client := resty.New().
		SetBaseURL(cfg.HTTPAddress).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	return &SleepaceProxy{
		client: client,
		appID:  cfg.AppID,
		secret: cfg.SecretKey,
		logger: logger,
	}
}

type proxyRequest struct {
	Token map[string]string      `json:"token"`
	Data  map[string]interface{} `json:"data"`
}

// Handler handles  POST /api/v1/proxy/sleepace/*
// The trailing path after /api/v1/proxy is forwarded to sleepace-service.
// Example: POST /api/v1/proxy/sleepace/getalarmnotifyconfig
//
//	→ POST http://sleepace-service:8090/sleepace/getalarmnotifyconfig
func (p *SleepaceProxy) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		targetPath := extractProxyPath(r.URL.Path)
		if targetPath == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing target path"})
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "read body failed"})
			return
		}
		defer r.Body.Close()

		var data map[string]interface{}
		if len(body) > 0 {
			if err := json.Unmarshal(body, &data); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
				return
			}
		}

		req := proxyRequest{
			Token: map[string]string{"appId": p.appID, "secureKey": p.secret},
			Data:  data,
		}

		resp := map[string]interface{}{}
		httpResp, err := p.client.R().
			SetBody(req).
			SetResult(&resp).
			Post(targetPath)

		if err != nil {
			p.logger.Error("proxy request failed", zap.String("path", targetPath), zap.Error(err))
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": "sleepace-service unreachable"})
			return
		}

		logFields := []zap.Field{
			zap.String("path", targetPath),
			zap.Int("upstream_status", httpResp.StatusCode()),
		}
		logFields = append(logFields, proxySleepaceReqFields(data)...)
		logFields = append(logFields,
			zap.String("response_summary", proxyResponseSummary(resp)),
			zap.String("response_preview", proxyResponseLogPreview(resp)),
		)
		p.logger.Debug("proxy", logFields...)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpResp.StatusCode())
		w.Write(httpResp.Body())
	}
}

// extractProxyPath extracts the sleepace-service path from the request URL.
// /api/v1/proxy/sleepace/getalarmnotifyconfig → /sleepace/getalarmnotifyconfig
func extractProxyPath(urlPath string) string {
	const prefix = "/api/v1/proxy"
	idx := strings.Index(urlPath, prefix)
	if idx < 0 {
		return ""
	}
	path := urlPath[idx+len(prefix):]
	if path == "" || path == "/" {
		return ""
	}
	return path
}

// proxySleepaceReqFields 厂家 data 体里的时间区间（get24HourDailyWithMaxReport 等为 Unix 秒）。
func proxySleepaceReqFields(data map[string]interface{}) []zap.Field {
	if len(data) == 0 {
		return nil
	}
	fs := make([]zap.Field, 0, 4)
	if v, ok := data["startTime"]; ok {
		fs = append(fs, zap.Any("req_startTime", v))
	}
	if v, ok := data["endTime"]; ok {
		fs = append(fs, zap.Any("req_endTime", v))
	}
	if v, ok := data["userId"]; ok {
		fs = append(fs, zap.Any("req_userId", v))
	}
	if v, ok := data["deviceId"]; ok {
		fs = append(fs, zap.Any("req_deviceId", v))
	}
	return fs
}

func proxyResponseSummary(resp map[string]interface{}) string {
	if resp == nil {
		return ""
	}
	st, _ := resp["status"]
	msg, _ := resp["msg"]
	n := -1
	if raw, ok := resp["data"]; ok && raw != nil {
		if arr, ok := raw.([]interface{}); ok {
			n = len(arr)
		}
	}
	return fmt.Sprintf("status=%v msg=%v data_items=%d", st, msg, n)
}

// proxyResponseLogPreview 整包 JSON 在首个 "maxReport" 处截断，避免 maxReport/detail 等大字段刷满日志。
func proxyResponseLogPreview(resp map[string]interface{}) string {
	if resp == nil {
		return ""
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return fmt.Sprintf("%v", resp)
	}
	s := string(b)
	if idx := strings.Index(s, `"maxReport"`); idx >= 0 {
		s = s[:idx] + `"...<omitted from maxReport>"`
	}
	const maxLen = 4096
	if len(s) > maxLen {
		return s[:maxLen] + fmt.Sprintf("...(%d bytes total)", len(s))
	}
	return s
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
