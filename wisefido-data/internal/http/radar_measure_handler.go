package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wisefido-data/internal/measure"

	"go.uber.org/zap"
)

// 测量相关 handler：track 反查 + 4 个无状态算法 + detection rectangle 下发。
//
// 路径前缀：/radar-device/api/v1/radar-device/device/{id}/measure/<sub>
//   GET   .../tracks?from=<unix_ms>&to=<unix_ms>     monitor_stream 反查
//   POST  .../fit-polygon                            cells → 凸包多边形 (wall)
//   POST  .../fit-rectangle                          cells → AABB (bed/long_sofa)
//   POST  .../fit-entry-lines                        cells → 线段集合 (enter)
//   POST  .../detection-rectangle                    polygon + FOV → 最大内接矩形并下发 firmware

// GetTracks 反查 monitor_stream 在 [from, to] 区间该雷达的 track 点。
//
// 用途：FE 离散化 measure session 期间 raw track 成 cell 颜色。session 状态由 FE 维护，
// backend 只做按 (device_addr, time_window) 反查。
func (h *RadarHandler) GetTracks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rawUID := extractRadarPathSeg(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/measure/tracks")
	deviceUID, ok := parseRadarDeviceUID(rawUID); if !ok {
		http.Error(w, "Invalid device UID", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	device, err := h.radarInstall.GetDeviceByUID(r.Context(), tenantID, deviceUID)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	if h.stubHandler != nil && h.stubHandler.DB != nil {
		callerUID := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-User-Role")
		if err := VerifyDeviceInScope(h.stubHandler.DB, r.Context(), callerUID, role, device.DeviceAddr); err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
	}
	fromMs, _ := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
	toMs, _ := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
	if fromMs <= 0 || toMs <= 0 || toMs <= fromMs {
		http.Error(w, "Invalid from/to", http.StatusBadRequest)
		return
	}
	if toMs-fromMs > 10*60*1000 {
		http.Error(w, "Time window too large (max 10 min)", http.StatusBadRequest)
		return
	}
	if h.monitorPlaybackRepo == nil {
		writeJSON(w, http.StatusOK, Fail("monitor playback repo not configured"))
		return
	}
	start := time.UnixMilli(fromMs).UTC()
	end := time.UnixMilli(toMs).UTC()
	rows, err := h.monitorPlaybackRepo.GetMonitorRowsByAddr(r.Context(), device.DeviceAddr, start, end, 30000)
	if err != nil {
		h.logger.Error("GetTracks: monitor_stream query", zap.String("uid", deviceUID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	tracks := extractTrackPoints(rows)
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
		"device_uid": deviceUID,
		"from_ms":    fromMs,
		"to_ms":      toMs,
		"count":      len(tracks),
		"tracks":     tracks,
	}))
}

// extractTrackPoints monitor_stream payload 形如 [{position_x, position_y, position_z, remaining_time, track_id, ...}]。
// 一行 payload 可能含多 track；展开为 flat 点列表，附 ts。
func extractTrackPoints(rows []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		ts, _ := row["timestamp"].(int64)
		// data_value 可能是 []interface{} 或单 map；正常应是数组
		dv := row["data_value"]
		arr, ok := dv.([]interface{})
		if !ok {
			continue
		}
		for _, item := range arr {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			point := map[string]interface{}{"timestamp": ts}
			for _, k := range []string{"position_x", "position_y", "position_z", "remaining_time", "track_id", "pose", "area_id", "track_confidence"} {
				if v, ok := m[k]; ok {
					point[k] = v
				}
			}
			out = append(out, point)
		}
	}
	return out
}

type fitCellsRequest struct {
	Cells   []measure.Cell `json:"cells"`
	MinHits int            `json:"min_hits"`
}

// FitPolygon wall intent — 凸包提议。L 形房间用户需手动调整顶点。
func (h *RadarHandler) FitPolygon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req fitCellsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.MinHits <= 0 {
		req.MinHits = measure.WallSatHits / 2
	}
	poly, ok := measure.FitPolygon(req.Cells, req.MinHits)
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
		"polygon":    poly,
		"fitted":     ok,
		"cell_count": len(req.Cells),
	}))
}

// FitRectangle bed/long_sofa intent — AABB。
func (h *RadarHandler) FitRectangle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req fitCellsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.MinHits <= 0 {
		req.MinHits = measure.BedSatHits / 2
	}
	rect, ok := measure.FitRectangle(req.Cells, req.MinHits)
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
		"rectangle":  rect,
		"fitted":     ok,
		"cell_count": len(req.Cells),
	}))
}

// FitEntryLines enter intent — 红色 cell 连通簇 → 线段集合。
func (h *RadarHandler) FitEntryLines(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req fitCellsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.MinHits <= 0 {
		req.MinHits = measure.EntrySatHits / 2
	}
	lines := measure.FitEntryLines(req.Cells, req.MinHits)
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
		"lines":      lines,
		"cell_count": len(req.Cells),
	}))
}

type detectionRectRequest struct {
	Polygon measure.Polygon `json:"polygon"`
	FOV     measure.Polygon `json:"fov"`
	Apply   bool            `json:"apply"`
}

// ComputeDetectionRectangle polygon ∩ FOV → 最大内接矩形；apply=true 时下发 firmware。
// FE 在 user 保存 wall polygon 时调用：先 apply=false 预览，确认后 apply=true 下发。
func (h *RadarHandler) ComputeDetectionRectangle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rawUID := extractRadarPathSeg(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/measure/detection-rectangle")
	deviceUID, ok := parseRadarDeviceUID(rawUID); if !ok {
		http.Error(w, "Invalid device UID", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	device, err := h.radarInstall.GetDeviceByUID(r.Context(), tenantID, deviceUID)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	if h.stubHandler != nil && h.stubHandler.DB != nil {
		callerUID := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-User-Role")
		if err := VerifyDeviceInScope(h.stubHandler.DB, r.Context(), callerUID, role, device.DeviceAddr); err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
	}
	var req detectionRectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	rect, ok2 := measure.MaxInscribedRect(req.Polygon, req.FOV)
	if !ok2 {
		writeJSON(w, http.StatusOK, Fail("cannot compute inscribed rectangle from polygon"))
		return
	}
	rectStr := formatRectangleCM(rect)
	resp := map[string]interface{}{
		"device_uid":     deviceUID,
		"rectangle":      rect,
		"rectangle_cm":   rectStr,
		"applied":        false,
	}
	if req.Apply {
		deviceCode, err := h.radarInstall.UpdateConfig(r.Context(), deviceUID, map[string]interface{}{
			"rectangle": rectStr,
		})
		if err != nil {
			h.logger.Error("ComputeDetectionRectangle: UpdateConfig failed",
				zap.String("uid", deviceUID), zap.Int("device_code", deviceCode), zap.Error(err))
			writeJSON(w, http.StatusOK, Result[any]{Code: ResultError, Type: "error", Message: err.Error(), Result: map[string]interface{}{"device_code": deviceCode, "rectangle_cm": rectStr}})
			return
		}
		resp["applied"] = true
		resp["device_code"] = deviceCode
		h.logger.Info("ComputeDetectionRectangle: rectangle applied",
			zap.String("uid", deviceUID), zap.String("rectangle_cm", rectStr), zap.Int("device_code", deviceCode))
	}
	writeJSON(w, http.StatusOK, Ok(resp))
}

// formatRectangleCM cm 轴对齐矩形 → firmware 规范 cm 字符串 {x1,y1;x2,y2;x3,y3;x4,y4}。
// 4 corner 顺序：TL, TR, BL, BR（与 firmware 自测回灌 rectangle 格式一致）。
// downstream wisefido-qinglan EncodeV1ConfigToDeviceProps 会自动 cm→dm。
func formatRectangleCM(r measure.Rectangle) string {
	tl := fmt.Sprintf("%d,%d", r.X, r.Y)
	tr := fmt.Sprintf("%d,%d", r.X+r.Width, r.Y)
	bl := fmt.Sprintf("%d,%d", r.X, r.Y+r.Height)
	br := fmt.Sprintf("%d,%d", r.X+r.Width, r.Y+r.Height)
	return "{" + strings.Join([]string{tl, tr, bl, br}, ";") + "}"
}
