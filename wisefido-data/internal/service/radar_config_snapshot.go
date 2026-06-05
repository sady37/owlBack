package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wisefido-qinglan/encode"

	"go.uber.org/zap"
)

// radar.* 配置快照落 spatial_config（source=device_reported），供 sensor 查库做概率计算。
// 挂在 GetOriginalProperties 回读路径而非下发后：install/boundary/area 下发会触发雷达重启，
// 刚下发回读拿不到稳定值；FE 读当前配置时设备已稳定，此刻固件 properties 才是真实 actual。
// 坐标统一 cm（与 radar 协议规范、roomengine canvas 一致），area_type 存固件原始 int（0-5）。
const (
	cfgKeyInstallMod    = "radar.install_mod"
	cfgKeyInstallHeight = "radar.install_height_cm"
	cfgKeyBoundary      = "radar.boundary"
	cfgKeyDeclareArea   = "radar.declare_area"
)

// RadarConfigSnapshot 给 sensor 的 radar 配置（坐标 cm；area_type 固件原始 int 0-5）。
// Source: "db"=命中快照 / "firmware"=冷启动现读固件已落库 / "empty"=离线或未绑定（无轨迹可算，空属正常）。
type RadarConfigSnapshot struct {
	DeviceAddr      string          `json:"device_addr"`
	InstallMod      *int            `json:"install_mod,omitempty"`
	InstallHeightCm *int            `json:"install_height_cm,omitempty"`
	Boundary        json.RawMessage `json:"boundary,omitempty"`
	DeclareArea     json.RawMessage `json:"declare_area,omitempty"`
	Source          string          `json:"source"`
}

type xyPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type declareAreaSnap struct {
	AreaID   int       `json:"area_id"`
	AreaType int       `json:"area_type"`
	Vertices []xyPoint `json:"vertices"`
}

// GetRadarConfigSnapshot read-through:先查 spatial_config radar.*；库空（冷启动）则解析 uid 现读固件 →
// 落库 → 再返回。离线/未绑定/读不到时返回 Source="empty"（设备离线本就无轨迹可算，空属正常）。
func (s *RadarInstall) GetRadarConfigSnapshot(ctx context.Context, deviceAddr string) (*RadarConfigSnapshot, error) {
	snap, found, err := s.readRadarConfigFromDB(ctx, deviceAddr)
	if err != nil {
		return nil, err
	}
	if found {
		snap.Source = "db"
		return snap, nil
	}

	var uid string
	err = s.db.QueryRowContext(ctx, `SELECT device_uid FROM devices WHERE device_addr=$1::inet`, deviceAddr).Scan(&uid)
	if err == sql.ErrNoRows {
		return &RadarConfigSnapshot{DeviceAddr: deviceAddr, Source: "empty"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("resolve device_uid: %w", err)
	}

	// 全读（nil keys）:qinglan 多 key 过滤有分组丢失 quirk（4 key 请求会丢 install_*），全读才稳。
	props, err := s.GetDeviceProperties(ctx, uid, nil)
	if err != nil {
		s.logger.Info("radar config read-through: firmware unreachable (likely offline)",
			zap.String("device_addr", deviceAddr), zap.String("uid", uid), zap.Error(err))
		return &RadarConfigSnapshot{DeviceAddr: deviceAddr, Source: "empty"}, nil
	}
	s.persistRadarConfigSnapshot(ctx, deviceAddr, props)
	snap, _, err = s.readRadarConfigFromDB(ctx, deviceAddr)
	if err != nil {
		return nil, err
	}
	snap.Source = "firmware"
	return snap, nil
}

func (s *RadarInstall) readRadarConfigFromDB(ctx context.Context, deviceAddr string) (*RadarConfigSnapshot, bool, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT config_key, config_value FROM spatial_config
		 WHERE spatial_prefix = $1::inet AND config_key LIKE 'radar.%'`, deviceAddr)
	if err != nil {
		return nil, false, fmt.Errorf("query spatial_config radar.*: %w", err)
	}
	defer rows.Close()

	snap := &RadarConfigSnapshot{DeviceAddr: deviceAddr}
	n := 0
	for rows.Next() {
		var key string
		var val []byte
		if err := rows.Scan(&key, &val); err != nil {
			return nil, false, fmt.Errorf("scan spatial_config row: %w", err)
		}
		n++
		switch key {
		case cfgKeyInstallMod:
			var i int
			if json.Unmarshal(val, &i) == nil {
				snap.InstallMod = &i
			}
		case cfgKeyInstallHeight:
			var i int
			if json.Unmarshal(val, &i) == nil {
				snap.InstallHeightCm = &i
			}
		case cfgKeyBoundary:
			snap.Boundary = json.RawMessage(val)
		case cfgKeyDeclareArea:
			snap.DeclareArea = json.RawMessage(val)
		}
	}
	return snap, n > 0, rows.Err()
}

// verifyRadarConfigActual 下发后回读固件落库（"保证写入"）。install/boundary/area 下发触发重启，
// 首次回读失败（设备重启中）→ 90s 后重试一次（覆盖重启重连窗）；全失败=设备离线（无轨迹可算，空属正常）。
// 异步执行：用 Background ctx（request ctx 在 HTTP 响应返回时即被 cancel）。
func (s *RadarInstall) verifyRadarConfigActual(deviceAddr, uid string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if s.tryReadbackPersist(ctx, deviceAddr, uid) {
		return
	}
	s.logger.Info("radar config verify: first readback failed, retry in 90s",
		zap.String("device_addr", deviceAddr), zap.String("uid", uid))
	select {
	case <-time.After(90 * time.Second):
	case <-ctx.Done():
		return
	}
	if !s.tryReadbackPersist(ctx, deviceAddr, uid) {
		s.logger.Warn("radar config verify: retry readback failed (device likely offline)",
			zap.String("device_addr", deviceAddr), zap.String("uid", uid))
	}
}

func (s *RadarInstall) tryReadbackPersist(ctx context.Context, deviceAddr, uid string) bool {
	props, err := s.GetDeviceProperties(ctx, uid, nil)
	if err != nil {
		return false
	}
	s.persistRadarConfigSnapshot(ctx, deviceAddr, props)
	return true
}

// persistRadarConfigSnapshot 从一次已读到的固件 properties 落 radar.* 进 spatial_config。
// 仅 UPSERT readback 中真实存在的 key（部分读不覆盖其余）；best-effort，失败只 warn 不影响读返回。
func (s *RadarInstall) persistRadarConfigSnapshot(ctx context.Context, deviceAddr string, props map[string]interface{}) {
	if deviceAddr == "" || len(props) == 0 {
		return
	}

	// 注：qinglan GET 已把设备 dm 归一化为 cm（install_height 200=2m、rectangle -500=5m 实测），
	// 故此处一律按 cm 处理，不再 ×10。key 用 qinglan 规范名 install_model/install_height（非下发侧 radar_install_*）。
	if v, ok := props["install_model"]; ok && v != nil {
		s.upsertRadarConfig(ctx, deviceAddr, cfgKeyInstallMod, toIntForProp(v))
	}
	if v, ok := props["install_height"]; ok && v != nil {
		s.upsertRadarConfig(ctx, deviceAddr, cfgKeyInstallHeight, toIntForProp(v))
	}
	if v, ok := props["rectangle"]; ok && v != nil {
		if pts, ok := parseRectangleCm(encode.ToStr(v)); ok {
			s.upsertRadarConfig(ctx, deviceAddr, cfgKeyBoundary, pts)
		}
	}
	if v, ok := props["declare_area"]; ok && v != nil {
		areas := parseDeclareAreasFromReadback(v)
		s.logger.Info("radar config snapshot: declare_area readback",
			zap.String("device_addr", deviceAddr), zap.Any("raw", v), zap.Int("parsed", len(areas)))
		s.upsertRadarConfig(ctx, deviceAddr, cfgKeyDeclareArea, areas)
	}
}

func (s *RadarInstall) upsertRadarConfig(ctx context.Context, deviceAddr, key string, value interface{}) {
	payload, err := json.Marshal(value)
	if err != nil {
		s.logger.Warn("radar config snapshot: marshal", zap.String("key", key), zap.Error(err))
		return
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO spatial_config (
			spatial_prefix, config_key, config_value,
			source, state_device, state_db, last_synced_at, last_acked_at, updated_at
		) VALUES (
			$1::inet, $2, $3::jsonb,
			'device_reported', 3, 3, now(), now(), now()
		)
		ON CONFLICT (spatial_prefix, config_key) DO UPDATE SET
			config_value   = EXCLUDED.config_value,
			source         = EXCLUDED.source,
			state_device   = EXCLUDED.state_device,
			state_db       = EXCLUDED.state_db,
			last_synced_at = EXCLUDED.last_synced_at,
			last_acked_at  = EXCLUDED.last_acked_at,
			updated_at     = EXCLUDED.updated_at
	`, deviceAddr, key, string(payload))
	if err != nil {
		s.logger.Warn("radar config snapshot: upsert",
			zap.String("device_addr", deviceAddr), zap.String("key", key), zap.Error(err))
	}
}

// parseRectangleCm 解析 qinglan 已归一化 cm 的检测矩形 "{x1,y1;x2,y2;x3,y3;x4,y4}"（4 顶点）。
func parseRectangleCm(s string) ([]xyPoint, bool) {
	segs := strings.Split(trimBraces(s), ";")
	if len(segs) != 4 {
		return nil, false
	}
	pts := make([]xyPoint, 0, 4)
	for _, seg := range segs {
		xy := strings.Split(strings.TrimSpace(seg), ",")
		if len(xy) != 2 {
			return nil, false
		}
		x, e1 := strconv.Atoi(strings.TrimSpace(xy[0]))
		y, e2 := strconv.Atoi(strings.TrimSpace(xy[1]))
		if e1 != nil || e2 != nil {
			return nil, false
		}
		pts = append(pts, xyPoint{X: x, Y: y})
	}
	return pts, true
}

// parseDeclareAreaCm 解析单块 "id,type,x1,y1,x2,y2,x3,y3,x4,y4"（cm，qinglan 已归一化）。
func parseDeclareAreaCm(block string) (declareAreaSnap, bool) {
	parts := strings.Split(trimBraces(block), ",")
	if len(parts) != 10 {
		return declareAreaSnap{}, false
	}
	var n [10]int
	for i := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(parts[i]))
		if err != nil {
			return declareAreaSnap{}, false
		}
		n[i] = v
	}
	return declareAreaSnap{
		AreaID:   n[0],
		AreaType: n[1],
		Vertices: []xyPoint{{X: n[2], Y: n[3]}, {X: n[4], Y: n[5]}, {X: n[6], Y: n[7]}, {X: n[8], Y: n[9]}},
	}, true
}

// parseDeclareAreasFromReadback 容忍多种回读形态：[]interface{} 多条 / 单串含多个 {..} 块 / 单条。
// 实测 qinglan 单串返回 "{a1},{a2},{a3}"（cm），splitBraceBlocks 按 {..} 拆。
func parseDeclareAreasFromReadback(v interface{}) []declareAreaSnap {
	out := []declareAreaSnap{}
	appendOne := func(s string) {
		if d, ok := parseDeclareAreaCm(s); ok {
			out = append(out, d)
		}
	}
	switch x := v.(type) {
	case []interface{}:
		for _, e := range x {
			appendOne(encode.ToStr(e))
		}
	case string:
		for _, blk := range splitBraceBlocks(x) {
			appendOne(blk)
		}
	default:
		appendOne(encode.ToStr(v))
	}
	return out
}

// trimBraces 去掉首尾 { } 及残留分隔符。
func trimBraces(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "{")
	for len(s) > 0 && (s[len(s)-1] == '}' || s[len(s)-1] == ',' || s[len(s)-1] == ';') {
		s = s[:len(s)-1]
	}
	return s
}

// splitBraceBlocks 把含多个 {..} 的串拆成单块；无花括号则整串作一块。
func splitBraceBlocks(s string) []string {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "{") {
		if s == "" {
			return nil
		}
		return []string{s}
	}
	var blocks []string
	for {
		i := strings.Index(s, "{")
		if i < 0 {
			break
		}
		j := strings.Index(s[i:], "}")
		if j < 0 {
			break
		}
		blocks = append(blocks, s[i:i+j+1])
		s = s[i+j+1:]
	}
	return blocks
}
