package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UpdateIniSection 表示 update.ini 中一个 file: 段。
//
// 格式（与 Sleepace 用户协商的写法，混合 ini + JSON-ish key）：
//
//	file:mcu_sleepace_BM8701-2_v6.89_20250814.zip
//	"deviceId": 49
//	"deviceVerison":"6.89"
//
// 注：deviceVerison（少 s）是厂家文档原样字段名，故意保留。
// DeviceAddr 字段对应厂家 /sleepace/firmware/delete 的 deviceType（49 = BM8701-2）。
type UpdateIniSection struct {
	File           string
	DeviceAddr       string // 对应厂家 deviceType（数字字符串，如 "49"）
	DeviceVerison  string // 对应厂家 deviceVersion（如 "6.89"）
}

const (
	otaDirEnvKey  = "OTA_DIR"
	// systemd 的 launcher 把 cwd 设为 owlBack，但实测 wisefido-data 的 /proc/PID/cwd 指向
	// owlBack/wisefido-data 子目录（go-build 的中间产物）；与 wisefido-qinglan 保持一致用 "../ota"。
	defaultOtaDir = "../ota"
)

// OtaDir 返回 update.ini 所在目录绝对路径，可用 OTA_DIR 环境变量覆盖。
func OtaDir() string {
	if v := os.Getenv(otaDirEnvKey); v != "" {
		return v
	}
	abs, err := filepath.Abs(defaultOtaDir)
	if err != nil {
		return defaultOtaDir
	}
	return abs
}

// UpdateIniPath 返回 update.ini 完整路径。
func UpdateIniPath() string {
	return filepath.Join(OtaDir(), "update.ini")
}

// LookupUpdateIni 在 update.ini 里按 filename 找一段，返回 deviceAddr/deviceVerison。
// 不区分键大小写；值带不带引号都接受。
// 找不到段返回 (nil, nil)；解析错误返回 error。
func LookupUpdateIni(filename string) (*UpdateIniSection, error) {
	data, err := os.ReadFile(UpdateIniPath())
	if err != nil {
		return nil, fmt.Errorf("read update.ini: %w", err)
	}
	var (
		current *UpdateIniSection
		want    = strings.TrimSpace(filename)
	)
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if strings.HasPrefix(line, "file:") {
			if current != nil {
				return current, nil
			}
			fname := strings.TrimSpace(strings.TrimPrefix(line, "file:"))
			if fname == want {
				current = &UpdateIniSection{File: fname}
			}
			continue
		}
		if current == nil {
			continue
		}
		key, val := splitIniKV(line)
		if key == "" {
			continue
		}
		switch strings.ToLower(key) {
		case "deviceid", "devicetype":
			current.DeviceAddr = val
		case "deviceverison", "deviceversion":
			current.DeviceVerison = val
		case "espsfver":
			// 历史 qinglan 风格段（mcu_sleepace_BM8701-2_v6.89_20250814.zip）的 espsfver 字段
			// 等价于 deviceVerison。仅在 deviceVerison 还没填时回退使用，避免覆盖。
			if current.DeviceVerison == "" {
				current.DeviceVerison = val
			}
		}
	}
	return current, nil
}

// AppendUpdateIniStub 写入一段到 update.ini。
//   - 同名 file: 段不存在 → 末尾追加新段
//   - 已存在但 deviceId/deviceVerison 为空 → 把空字段就地填上（不覆盖已填值）
//   - 已存在且字段已填 → 不动（避免覆盖人手填的值）
func AppendUpdateIniStub(filename, deviceAddr, deviceVerison string) error {
	if filename == "" {
		return fmt.Errorf("filename required")
	}
	existing, _ := LookupUpdateIni(filename)
	if existing == nil {
		return appendNewSection(filename, deviceAddr, deviceVerison)
	}
	// 仅当现有字段为空且新值非空时才填
	needFillID := existing.DeviceAddr == "" && deviceAddr != ""
	needFillVer := existing.DeviceVerison == "" && deviceVerison != ""
	if !needFillID && !needFillVer {
		return nil
	}
	return fillExistingSection(filename, needFillID, deviceAddr, needFillVer, deviceVerison)
}

func appendNewSection(filename, deviceAddr, deviceVerison string) error {
	if err := os.MkdirAll(OtaDir(), 0o755); err != nil {
		return fmt.Errorf("mkdir ota: %w", err)
	}
	f, err := os.OpenFile(UpdateIniPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open update.ini: %w", err)
	}
	defer f.Close()
	verVal := ""
	if deviceVerison != "" {
		verVal = `"` + deviceVerison + `"`
	}
	section := fmt.Sprintf("\nfile:%s\n\"deviceId\": %s\n\"deviceVerison\":%s\n", filename, deviceAddr, verVal)
	if _, err := f.WriteString(section); err != nil {
		return fmt.Errorf("write update.ini: %w", err)
	}
	return nil
}

// fillExistingSection 就地改写指定 file: 段的 deviceId/deviceVerison 行。
// 简单实现：整文件读入 → 行扫描 → 进入目标段后改下一行匹配 key → 整文件回写。
func fillExistingSection(filename string, fillID bool, deviceAddr string, fillVer bool, deviceVerison string) error {
	path := UpdateIniPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read update.ini: %w", err)
	}
	lines := strings.Split(string(data), "\n")
	want := strings.TrimSpace(filename)
	inSection := false
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "file:") {
			fname := strings.TrimSpace(strings.TrimPrefix(line, "file:"))
			inSection = (fname == want)
			continue
		}
		if !inSection {
			continue
		}
		key, val := splitIniKV(line)
		lk := strings.ToLower(key)
		if fillID && val == "" && (lk == "deviceid" || lk == "devicetype") {
			lines[i] = `"deviceId": ` + deviceAddr
		}
		if fillVer && val == "" && (lk == "deviceverison" || lk == "deviceversion") {
			lines[i] = `"deviceVerison":"` + deviceVerison + `"`
		}
	}
	out := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		return fmt.Errorf("write update.ini: %w", err)
	}
	return nil
}

// splitIniKV 解析一行 ini，返回 key, value。
// 接受多种分隔: ":"（带可选 JSON 引号 key）、"\t"、" "（首个空白）。
// 值带的双引号自动剥掉。
func splitIniKV(line string) (string, string) {
	// 优先按 ':' 分（"deviceId": 49 / "deviceVerison":"6.89" 这种）
	if idx := strings.IndexByte(line, ':'); idx > 0 {
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		key = strings.Trim(key, "\"' ")
		val = strings.Trim(val, "\"' ")
		return key, val
	}
	// 否则按首个空白拆（espsfver\t6.89 这种）
	if idx := strings.IndexAny(line, " \t"); idx > 0 {
		return strings.TrimSpace(line[:idx]), strings.Trim(strings.TrimSpace(line[idx:]), "\"' ")
	}
	return "", ""
}
