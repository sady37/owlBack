package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"wisefido-qinglan/internal/ota"
	"wisefido-qinglan/internal/tcp"

	"github.com/gorilla/mux"
)

// OTAHandler handles OTA-related HTTP API requests
type OTAHandler struct {
	otaManager *ota.Manager
	tcpServer  *tcp.Server
}

// NewOTAHandler creates a new OTA handler
func NewOTAHandler(otaManager *ota.Manager, tcpServer *tcp.Server) *OTAHandler {
	return &OTAHandler{
		otaManager: otaManager,
		tcpServer:  tcpServer,
	}
}

// RegisterRoutes registers OTA routes on the router
func (h *OTAHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/ota/trigger/{uid}", h.TriggerOTA).Methods("POST")
	router.HandleFunc("/api/v1/ota/trigger/batch", h.TriggerBatchOTA).Methods("POST")
	router.HandleFunc("/api/v1/ota/status", h.GetTCPDevices).Methods("GET")
	router.HandleFunc("/api/v1/ota/firmware/{vendor}", h.ListFirmware).Methods("GET")
	router.HandleFunc("/api/v1/ota/firmware/{vendor}", h.UploadFirmware).Methods("POST")
	router.HandleFunc("/api/v1/ota/firmware/{vendor}/{filename}", h.DeleteFirmware).Methods("DELETE")
}

// TriggerOTA triggers OTA for a single device
func (h *OTAHandler) TriggerOTA(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	var req ota.PushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	req.UID = uid

	log.Printf("[OTA-API] trigger OTA: uid=%s esp=%s radar=%s", uid, req.EspFirmware, req.RadarFirmware)

	result := h.otaManager.PushToDevice(req)

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
