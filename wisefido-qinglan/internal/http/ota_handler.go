package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"wisefido-qinglan/internal/ota"
	"wisefido-qinglan/internal/tcp"

	"github.com/gorilla/mux"
)

// DeviceCommander is the interface for sending MQTT commands to devices
type DeviceCommander interface {
	PublishReboot(ctx context.Context, uid string) error
	SetDeviceProperties(ctx context.Context, uid string, properties map[string]interface{}) error
}

// MQTTOTAPusher pushes OTA via MQTT
type MQTTOTAPusher interface {
	PublishOTA(ctx context.Context, uid string, data map[string]interface{}) error
}

// OTAHandler handles OTA-related HTTP API requests
type OTAHandler struct {
	otaManager  *ota.Manager
	tcpServer   *tcp.Server
	commander   DeviceCommander
	mqttOTA     MQTTOTAPusher
}

// NewOTAHandler creates a new OTA handler
func NewOTAHandler(otaManager *ota.Manager, tcpServer *tcp.Server) *OTAHandler {
	return &OTAHandler{
		otaManager: otaManager,
		tcpServer:  tcpServer,
	}
}

// SetCommander injects the MQTT publisher for device commands
func (h *OTAHandler) SetCommander(c DeviceCommander) {
	h.commander = c
}

// SetMQTTOTA injects the MQTT OTA publisher
func (h *OTAHandler) SetMQTTOTA(p MQTTOTAPusher) {
	h.mqttOTA = p
}

// RegisterRoutes registers OTA routes on the router
func (h *OTAHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/ota/trigger/{uid}", h.TriggerOTA).Methods("POST")
	router.HandleFunc("/api/v1/ota/trigger/batch", h.TriggerBatchOTA).Methods("POST")
	router.HandleFunc("/api/v1/ota/status", h.GetTCPDevices).Methods("GET")
	router.HandleFunc("/api/v1/ota/firmware", h.ListFirmwareRoot).Methods("GET")
	router.HandleFunc("/api/v1/ota/firmware", h.UploadFirmwareRoot).Methods("POST")
	router.HandleFunc("/api/v1/ota/firmware/{filename}", h.DeleteFirmwareRoot).Methods("DELETE")
	router.HandleFunc("/api/v1/ota/firmware/{vendor}", h.ListFirmware).Methods("GET")
	router.HandleFunc("/api/v1/ota/firmware/{vendor}", h.UploadFirmware).Methods("POST")
	router.HandleFunc("/api/v1/ota/firmware/{vendor}/{filename}", h.DeleteFirmware).Methods("DELETE")
	// Device command endpoints
	router.HandleFunc("/api/v1/device/restart", h.BatchRestart).Methods("POST")
	router.HandleFunc("/api/v1/device/set-wifi", h.BatchSetWiFi).Methods("POST")
	router.HandleFunc("/api/v1/device/set-iotserver", h.BatchSetIoTServer).Methods("POST")
}

// TriggerOTA triggers OTA for a single device (TCP first, MQTT fallback)
func (h *OTAHandler) TriggerOTA(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	var req ota.PushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	req.UID = uid

	// 版本缺省则从 update.ini 按文件名查（esp / radar 独立）
	if req.EspFirmware != "" && (req.EspVersion == "" || strings.Contains(req.EspVersion, ".bin")) {
		if verInfo := ota.ParseUpdateINI(h.otaManager.FirmwareDir, req.EspFirmware); verInfo != nil {
			req.EspVersion = verInfo.EspVer
		}
	}
	if req.RadarFirmware != "" && (req.RadarVersion == "" || strings.Contains(req.RadarVersion, ".bin")) {
		if verInfo := ota.ParseUpdateINI(h.otaManager.FirmwareDir, req.RadarFirmware); verInfo != nil {
			req.RadarVersion = verInfo.RadarVer
		}
	}

	log.Printf("[OTA-API] trigger OTA: uid=%s esp=%s(v%s) radar=%s(v%s)", uid, req.EspFirmware, req.EspVersion, req.RadarFirmware, req.RadarVersion)

	// Try TCP push first
	result := h.otaManager.PushToDevice(req)
	if result.Success {
		log.Printf("[OTA-API] TCP push OK: uid=%s", uid)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// TCP failed, try MQTT push
	if h.mqttOTA != nil {
		log.Printf("[OTA-API] TCP failed (%s), trying MQTT: uid=%s", result.Message, uid)
		// 按协议同时支持 esp (主控) 与 radar (雷达) 两组字段，互相独立
		data := map[string]interface{}{}
		if req.EspFirmware != "" {
			info, err := h.otaManager.GetFirmwareInfo(req.EspFirmware)
			if err == nil {
				data["espfileUrl"] = fmt.Sprintf("%s/%s", h.otaManager.FirmwareURL, req.EspFirmware)
				data["espfilesha256"] = info.SHA256
				data["espfilesize"] = info.Size
				data["espver"] = req.EspVersion
			}
		}
		if req.EspFileURL != "" {
			data["espfileUrl"] = req.EspFileURL
			data["espfilesha256"] = req.EspSHA256
			data["espfilesize"] = req.EspFileSize
			data["espver"] = req.EspVersion
		}
		if req.RadarFirmware != "" {
			info, err := h.otaManager.GetFirmwareInfo(req.RadarFirmware)
			if err == nil {
				data["radarfileUrl"] = fmt.Sprintf("%s/%s", h.otaManager.FirmwareURL, req.RadarFirmware)
				data["radarfilesha256"] = info.SHA256
				data["radarfilesize"] = info.Size
				data["radarver"] = req.RadarVersion
			}
		}
		if len(data) > 0 {
			if err := h.mqttOTA.PublishOTA(r.Context(), uid, data); err != nil {
				result.Message = fmt.Sprintf("TCP: %s, MQTT: %s", result.Message, err.Error())
			} else {
				log.Printf("[OTA-API] MQTT push OK: uid=%s", uid)
				result.Success = true
				result.Message = "OTA pushed via MQTT"
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if !result.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(result)
}

// TriggerBatchOTA triggers OTA for multiple devices
func (h *OTAHandler) TriggerBatchOTA(w http.ResponseWriter, r *http.Request) {
	var reqs []ota.PushRequest
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("[OTA-API] batch trigger OTA: %d devices", len(reqs))

	var results []ota.PushResult
	for _, req := range reqs {
		result := h.otaManager.PushToDevice(req)
		results = append(results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GetTCPDevices returns the list of TCP-connected devices
func (h *OTAHandler) GetTCPDevices(w http.ResponseWriter, r *http.Request) {
	sessions := h.tcpServer.Sessions.GetAllSessions()

	type deviceInfo struct {
		UID       string `json:"uid"`
		Type      string `json:"type"`
		SfVer     string `json:"sf_ver"`
		HwVer     string `json:"hw_ver"`
		ConnectAt string `json:"connect_at"`
		LastHeart string `json:"last_heart"`
		RemoteIP  string `json:"remote_ip"`
	}

	var devices []deviceInfo
	for _, sess := range sessions {
		devices = append(devices, deviceInfo{
			UID:       sess.UID,
			Type:      sess.Type,
			SfVer:     sess.SfVer,
			HwVer:     sess.HwVer,
			ConnectAt: sess.ConnectAt.Format("2006-01-02 15:04:05"),
			LastHeart: sess.LastHeart.Format("2006-01-02 15:04:05"),
			RemoteIP:  sess.Conn.RemoteAddr().String(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"online_count": h.tcpServer.Sessions.OnlineCount(),
		"devices":      devices,
	})
}

// ListFirmware lists firmware files for a vendor
func (h *OTAHandler) ListFirmware(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vendor := vars["vendor"]

	files, err := h.otaManager.ListFirmwareFiles(vendor)
	if err != nil {
		http.Error(w, fmt.Sprintf("list firmware failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// UploadFirmware uploads a firmware file for a vendor
func (h *OTAHandler) UploadFirmware(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vendor := vars["vendor"]

	// Parse multipart form (max 100MB)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, fmt.Sprintf("parse form failed: %v", err), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("firmware")
	if err != nil {
		http.Error(w, fmt.Sprintf("get file failed: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create vendor directory if needed
	vendorDir := filepath.Join(h.otaManager.FirmwareDir, vendor)
	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		http.Error(w, fmt.Sprintf("create dir failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Save file
	destPath := filepath.Join(vendorDir, header.Filename)
	dst, err := os.Create(destPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("create file failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, fmt.Sprintf("save file failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[OTA-API] firmware uploaded: vendor=%s file=%s size=%d", vendor, header.Filename, header.Size)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "firmware uploaded",
		"vendor":   vendor,
		"filename": header.Filename,
		"size":     header.Size,
	})
}

// DeleteFirmware deletes a firmware file
func (h *OTAHandler) DeleteFirmware(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vendor := vars["vendor"]
	filename := vars["filename"]

	fwPath := filepath.Join(h.otaManager.FirmwareDir, vendor, filename)
	if err := os.Remove(fwPath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "firmware file not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("delete failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[OTA-API] firmware deleted: vendor=%s file=%s", vendor, filename)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "firmware deleted",
		"vendor":   vendor,
		"filename": filename,
	})
}

// --- Root firmware endpoints (flat ota/ directory, no vendor) ---

func (h *OTAHandler) ListFirmwareRoot(w http.ResponseWriter, r *http.Request) {
	files, err := h.otaManager.ListFirmwareFiles("")
	if err != nil {
		http.Error(w, fmt.Sprintf("list firmware failed: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (h *OTAHandler) UploadFirmwareRoot(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, fmt.Sprintf("parse form failed: %v", err), http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("firmware")
	if err != nil {
		http.Error(w, fmt.Sprintf("get file failed: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	destPath := filepath.Join(h.otaManager.FirmwareDir, header.Filename)
	dst, err := os.Create(destPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("create file failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, fmt.Sprintf("save file failed: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("[OTA-API] firmware uploaded: file=%s size=%d", header.Filename, header.Size)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "uploaded", "filename": header.Filename, "size": header.Size})
}

func (h *OTAHandler) DeleteFirmwareRoot(w http.ResponseWriter, r *http.Request) {
	filename := mux.Vars(r)["filename"]
	fwPath := filepath.Join(h.otaManager.FirmwareDir, filename)
	if err := os.Remove(fwPath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("delete failed: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("[OTA-API] firmware deleted: file=%s", filename)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "deleted", "filename": filename})
}

// --- Device command endpoints ---

type batchUIDRequest struct {
	UIDs []string `json:"uids"`
}

type batchResult struct {
	UID     string `json:"uid"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// BatchRestart restarts multiple devices via MQTT (dev:0)
func (h *OTAHandler) BatchRestart(w http.ResponseWriter, r *http.Request) {
	if h.commander == nil {
		http.Error(w, "MQTT publisher not configured", http.StatusServiceUnavailable)
		return
	}
	var req batchUIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}
	log.Printf("[Device-API] batch restart: %d devices", len(req.UIDs))

	results := make([]batchResult, 0, len(req.UIDs))
	for _, uid := range req.UIDs {
		if err := h.commander.PublishReboot(r.Context(), uid); err != nil {
			results = append(results, batchResult{UID: uid, Success: false, Message: err.Error()})
		} else {
			results = append(results, batchResult{UID: uid, Success: true, Message: "restart sent"})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// BatchSetWiFi sets WiFi config for multiple devices via MQTT
func (h *OTAHandler) BatchSetWiFi(w http.ResponseWriter, r *http.Request) {
	if h.commander == nil {
		http.Error(w, "MQTT publisher not configured", http.StatusServiceUnavailable)
		return
	}
	var req struct {
		UIDs []string `json:"uids"`
		SSID string   `json:"ssid"`
		Pass string   `json:"pass"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}
	if req.SSID == "" {
		http.Error(w, "ssid is required", http.StatusBadRequest)
		return
	}
	log.Printf("[Device-API] batch set-wifi: %d devices ssid=%s", len(req.UIDs), req.SSID)

	props := map[string]interface{}{
		"wifi_ssid": req.SSID,
		"wifi_pass": req.Pass,
	}
	results := make([]batchResult, 0, len(req.UIDs))
	for _, uid := range req.UIDs {
		if err := h.commander.SetDeviceProperties(r.Context(), uid, props); err != nil {
			results = append(results, batchResult{UID: uid, Success: false, Message: err.Error()})
		} else {
			results = append(results, batchResult{UID: uid, Success: true, Message: "wifi config sent"})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// BatchSetIoTServer sets IoT server address for multiple devices via MQTT
func (h *OTAHandler) BatchSetIoTServer(w http.ResponseWriter, r *http.Request) {
	if h.commander == nil {
		http.Error(w, "MQTT publisher not configured", http.StatusServiceUnavailable)
		return
	}
	var req struct {
		UIDs   []string `json:"uids"`
		Server string   `json:"server"`
		Port   int      `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}
	if req.Server == "" || req.Port == 0 {
		http.Error(w, "server and port are required", http.StatusBadRequest)
		return
	}
	log.Printf("[Device-API] batch set-iotserver: %d devices server=%s:%d", len(req.UIDs), req.Server, req.Port)

	props := map[string]interface{}{
		"ip_port": fmt.Sprintf("%s:%d", req.Server, req.Port),
	}
	results := make([]batchResult, 0, len(req.UIDs))
	for _, uid := range req.UIDs {
		if err := h.commander.SetDeviceProperties(r.Context(), uid, props); err != nil {
			results = append(results, batchResult{UID: uid, Success: false, Message: err.Error()})
		} else {
			results = append(results, batchResult{UID: uid, Success: true, Message: "iotserver config sent"})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
