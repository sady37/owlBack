package ota

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"wisefido-qinglan/internal/tcp"
	pb "wisefido-qinglan/proto/gen"

	"google.golang.org/protobuf/proto"
)

// Manager OTA push manager
type Manager struct {
	TCPServer   *tcp.Server
	FirmwareDir string // firmware storage directory, e.g. ./ota
	FirmwareURL string // HTTP firmware download base URL
}

// NewManager creates OTA manager
func NewManager(tcpServer *tcp.Server, firmwareDir, firmwareURL string) *Manager {
	return &Manager{
		TCPServer:   tcpServer,
		FirmwareDir: firmwareDir,
		FirmwareURL: firmwareURL,
	}
}

// PushRequest OTA push request
type PushRequest struct {
	UID           string
	EspFirmware   string // ESP firmware filename (relative to FirmwareDir)
	EspVersion    string
	EspSHA256     string // optional, auto-calculated if empty
	RadarFirmware string
	RadarVersion  string
	RadarSHA256   string
}

// PushResult OTA push result
type PushResult struct {
	UID     string `json:"uid"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// PushToDevice pushes OTA to device via TCP
func (m *Manager) PushToDevice(req PushRequest) PushResult {
	otaReq := &pb.OTAReq{}

	if req.EspFirmware != "" {
		info, err := m.getFirmwareInfo(req.EspFirmware)
		if err != nil {
			return PushResult{UID: req.UID, Success: false, Message: fmt.Sprintf("ESP firmware error: %v", err)}
		}
		sha := req.EspSHA256
		if sha == "" {
			sha = info.SHA256
		}
		otaReq.Espsfver = req.EspVersion
		otaReq.ESPFileUrl = fmt.Sprintf("%s/%s", m.FirmwareURL, req.EspFirmware)
		otaReq.ESPFileSize = uint32(info.Size)
		otaReq.ESPFileSHA256 = sha
	}

	if req.RadarFirmware != "" {
		info, err := m.getFirmwareInfo(req.RadarFirmware)
		if err != nil {
			return PushResult{UID: req.UID, Success: false, Message: fmt.Sprintf("Radar firmware error: %v", err)}
		}
		sha := req.RadarSHA256
		if sha == "" {
			sha = info.SHA256
		}
		otaReq.Radarsfver = req.RadarVersion
		otaReq.RadarFileUrl = fmt.Sprintf("%s/%s", m.FirmwareURL, req.RadarFirmware)
		otaReq.RadarFileSize = uint32(info.Size)
		otaReq.RadarFileSHA256 = sha
	}

	if otaReq.ESPFileUrl == "" && otaReq.RadarFileUrl == "" {
		return PushResult{UID: req.UID, Success: false, Message: "must provide ESP or Radar firmware"}
	}

	data, err := proto.Marshal(otaReq)
	if err != nil {
		return PushResult{UID: req.UID, Success: false, Message: fmt.Sprintf("marshal failed: %v", err)}
	}

	frame := &tcp.Frame{Type: tcp.TypeOTAPush, Data: data}
	if err := m.TCPServer.PushOTA(req.UID, frame); err != nil {
		return PushResult{UID: req.UID, Success: false, Message: err.Error()}
	}

	if otaReq.ESPFileUrl != "" {
		log.Printf("[OTA] ESP push: UID=%s ver=%s url=%s size=%d sha256=%s", req.UID, otaReq.Espsfver, otaReq.ESPFileUrl, otaReq.ESPFileSize, otaReq.ESPFileSHA256)
	}
	if otaReq.RadarFileUrl != "" {
		log.Printf("[OTA] Radar push: UID=%s ver=%s url=%s size=%d sha256=%s", req.UID, otaReq.Radarsfver, otaReq.RadarFileUrl, otaReq.RadarFileSize, otaReq.RadarFileSHA256)
	}
	log.Printf("[OTA] OTA push sent: UID=%s ESP=%s Radar=%s", req.UID, req.EspFirmware, req.RadarFirmware)
	return PushResult{UID: req.UID, Success: true, Message: "OTA push sent"}
}

type firmwareInfo struct {
	Size   int64
	SHA256 string
}

func (m *Manager) getFirmwareInfo(filename string) (*firmwareInfo, error) {
	filePath := filepath.Join(m.FirmwareDir, filename)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open firmware %s: %w", filePath, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("SHA256 calculation failed: %w", err)
	}

	return &firmwareInfo{
		Size:   stat.Size(),
		SHA256: fmt.Sprintf("%x", hasher.Sum(nil)),
	}, nil
}

// ListFirmwareFiles lists firmware files in vendor directory
func (m *Manager) ListFirmwareFiles(vendor string) ([]FirmwareFile, error) {
	dir := filepath.Join(m.FirmwareDir, vendor)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []FirmwareFile{}, nil
		}
		return nil, err
	}

	var files []FirmwareFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, _ := entry.Info()
		if info != nil {
			files = append(files, FirmwareFile{Name: entry.Name(), Size: info.Size()})
		}
	}
	return files, nil
}

// FirmwareFile firmware file descriptor
type FirmwareFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}
