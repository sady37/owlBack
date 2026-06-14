package testkit

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Manifest 是 doc/cases/legos/manifest.json 的 Go 表示。
type Manifest struct {
	Version     string         `json:"manifest_version"`
	Description string         `json:"description"`
	Cases       []ManifestCase `json:"cases"`
}

// ManifestCase 是 manifest 中单个乐高块的描述（索引不复制数据，约束#7）。
type ManifestCase struct {
	ID            string          `json:"id"`
	Class         string          `json:"class"`
	Labels        []string        `json:"labels"`
	DeviceUIDs    []string        `json:"device_uids"`
	DeviceAddrs   []string        `json:"device_addrs"`
	SourceFixture string          `json:"source_fixture"`
	Windows       ManifestWindows `json:"windows"`
	Groundtruth   string          `json:"groundtruth"`
	Causality     string          `json:"causality"`
	Notes         string          `json:"notes"`
}

// ManifestWindows 包含 primary 窗口和可选的 sub 窗口列表。
type ManifestWindows struct {
	Primary ManifestWindow   `json:"primary"`
	Sub     []ManifestWindow `json:"sub,omitempty"`
}

// ManifestWindow 描述单个时间窗口及其数据文件。
type ManifestWindow struct {
	StartMs   int64    `json:"start_ms"`
	EndMs     int64    `json:"end_ms"`
	DurationS int      `json:"duration_s,omitempty"`
	Files     []string `json:"files"`
	Fixture   string   `json:"fixture,omitempty"`
	Role      string   `json:"role,omitempty"`
}

// LoadManifest 读取并解析 manifest.json。
func LoadManifest(path string) (*Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// ResolveCase 加载一个 manifest case 的全量 LegoV2Record。
// casesDir 是 doc/cases/ 的根路径（manifest.json 所在的 legos/ 目录的上层）。
// 返回 primary 窗口的记录和 sub 窗口的记录（各自已按 ts 排序）。
func ResolveCase(c ManifestCase, casesDir string) (primary []LegoV2Record, sub []LegoV2Record, err error) {
	fixtureDir := filepath.Join(casesDir, c.SourceFixture)
	primary, err = LoadWindow(fixtureDir)
	if err != nil {
		return nil, nil, err
	}
	for _, w := range c.Windows.Sub {
		if w.Fixture == "" {
			continue
		}
		subDir := filepath.Join(casesDir, w.Fixture)
		recs, err := LoadWindow(subDir)
		if err != nil {
			return nil, nil, err
		}
		sub = append(sub, recs...)
	}
	return primary, sub, nil
}
