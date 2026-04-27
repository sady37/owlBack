package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"wisefido-data/internal/service"
)

// SleepaceFirmwareHandler 处理 sleepace 固件相关 5 个端点。
//
// 厂家直传（admin 工具 / 不走 UI）：
//
//	POST /sleepace/api/v1/sleepace/firmware/upload         multipart "file" → 直接推厂家
//	POST /sleepace/api/v1/sleepace/firmware/delete         body {"device_type":int, "device_version":string}
//	GET  /sleepace/api/v1/sleepace/firmware/versions       channel 由 wisefido-sleepace 注入
//
// 前端走的本地端点：
//
//	POST /sleepace/api/v1/sleepace/firmware/local-upload   multipart "file" + form "device_type"+"device_version"
//	                                                       → 存 owlBack/ota/<filename> + 追加 update.ini stub
//	                                                       两个 form 都填 → 顺带推一份给厂家
//	POST /sleepace/api/v1/sleepace/firmware/local-delete   body {"filename":"..."} → 按 update.ini 查 (devType, devVer) 调厂家 delete
//	                                                       不动本地 zip、不动 update.ini
type SleepaceFirmwareHandler struct {
	svc    service.DeviceMonitorSettingsService
	logger *zap.Logger
}

func NewSleepaceFirmwareHandler(svc service.DeviceMonitorSettingsService, logger *zap.Logger) *SleepaceFirmwareHandler {
	return &SleepaceFirmwareHandler{svc: svc, logger: logger}
}

// Dispatch 基于 path 后缀分发：upload | delete | versions。
func (h *SleepaceFirmwareHandler) Dispatch(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/internal/sleepace/firmware/")
	path = strings.TrimPrefix(path, "/sleepace/api/v1/sleepace/firmware/")
	action := strings.Trim(path, "/")

	switch action {
	case "upload":
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		h.handleUpload(w, r)
	case "delete":
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		h.handleDelete(w, r)
	case "versions":
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "GET or POST only", http.StatusMethodNotAllowed)
			return
		}
		h.handleVersions(w, r)
	case "local-upload":
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		h.handleLocalUpload(w, r)
	case "local-delete":
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		h.handleLocalDelete(w, r)
	default:
		writeFirmwareJSON(w, http.StatusBadRequest, map[string]any{
			"status": "error",
			"error":  "unknown action; expected upload | delete | versions | local-upload | local-delete",
		})
	}
}

func (h *SleepaceFirmwareHandler) handleUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeFirmwareJSON(w, http.StatusBadRequest, map[string]any{"status": "error", "error": "parse multipart: " + err.Error()})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeFirmwareJSON(w, http.StatusBadRequest, map[string]any{"status": "error", "error": "missing form field 'file'"})
		return
	}
	defer file.Close()

	if err := h.svc.UploadSleepaceFirmware(r.Context(), header.Filename, file); err != nil {
		h.logger.Error("upload firmware failed", zap.String("filename", header.Filename), zap.Error(err))
		writeFirmwareJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "error": err.Error(), "filename": header.Filename})
		return
	}
	writeFirmwareJSON(w, http.StatusOK, map[string]any{"status": "ok", "filename": header.Filename, "size": header.Size})
}

func (h *SleepaceFirmwareHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DeviceType    int    `json:"device_type"`
		DeviceVersion string `json:"device_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeFirmwareJSON(w, http.StatusBadRequest, map[string]any{"status": "error", "error": "invalid json: " + err.Error()})
		return
	}
	if body.DeviceVersion == "" {
		writeFirmwareJSON(w, http.StatusBadRequest, map[string]any{"status": "error", "error": "device_version required"})
		return
	}
	if err := h.svc.DeleteSleepaceFirmware(r.Context(), body.DeviceType, body.DeviceVersion); err != nil {
		writeFirmwareJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "error": err.Error()})
		return
	}
	writeFirmwareJSON(w, http.StatusOK, map[string]any{
		"status":         "ok",
		"device_type":    body.DeviceType,
		"device_version": body.DeviceVersion,
	})
}

func (h *SleepaceFirmwareHandler) handleVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := h.svc.ListSleepaceFirmwareVersions(r.Context())
	if err != nil {
		writeFirmwareJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "error": err.Error()})
		return
	}
	writeFirmwareJSON(w, http.StatusOK, map[string]any{"status": "ok", "data": versions, "count": len(versions)})
}

func (h *SleepaceFirmwareHandler) handleLocalUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeFirmwareJSON(w, http.StatusBadRequest, map[string]any{"status": "error", "error": "parse multipart: " + err.Error()})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeFirmwareJSON(w, http.StatusBadRequest, map[string]any{"status": "error", "error": "missing form field 'file'"})
		return
	}
	defer file.Close()

	deviceType := r.FormValue("device_type")
	deviceVersion := r.FormValue("device_version")

	pushed, err := h.svc.LocalUploadSleepaceFirmware(r.Context(), header.Filename, file, deviceType, deviceVersion)
	if err != nil {
		h.logger.Error("local upload failed",
			zap.String("filename", header.Filename),
			zap.String("device_type", deviceType),
			zap.String("device_version", deviceVersion),
			zap.Error(err))
		writeFirmwareJSON(w, http.StatusInternalServerError, map[string]any{
			"status": "error", "error": err.Error(),
			"filename": header.Filename, "pushed_to_vendor": pushed,
		})
		return
	}
	writeFirmwareJSON(w, http.StatusOK, map[string]any{
		"status":           "ok",
		"filename":         header.Filename,
		"size":             header.Size,
		"device_type":      deviceType,
		"device_version":   deviceVersion,
		"pushed_to_vendor": pushed,
	})
}

func (h *SleepaceFirmwareHandler) handleLocalDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Filename string `json:"filename"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeFirmwareJSON(w, http.StatusBadRequest, map[string]any{"status": "error", "error": "invalid json: " + err.Error()})
		return
	}
	if body.Filename == "" {
		writeFirmwareJSON(w, http.StatusBadRequest, map[string]any{"status": "error", "error": "filename required"})
		return
	}
	if err := h.svc.LocalDeleteSleepaceFirmware(r.Context(), body.Filename); err != nil {
		writeFirmwareJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "error": err.Error(), "filename": body.Filename})
		return
	}
	writeFirmwareJSON(w, http.StatusOK, map[string]any{"status": "ok", "filename": body.Filename})
}

func writeFirmwareJSON(w http.ResponseWriter, status int, body map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

