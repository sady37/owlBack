package service

// sensor_veto_client.go — Feedback 学习区否决通道（[[layout_authority_ai_correction_model]]）。
//
// 人在 RadarCanvas 删掉 source='Feedback' object = 否决该处自动学习。SaveRoomLayout diff 出被删者，
// 直调 sensor live engine 的 HTTP 端点（单一消费者，不走 Redis）。sensor 收到后对该 cell 调
// ClearNonHumanLearnedZone + MarkLearnBlocked。与 vendor 下发（qinglan RadarAPI）完全分离。

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// sensorHTTPBaseURL sensor live engine HTTP 端点（veto + feedback ingest 共用；同机默认；env SENSOR_HTTP_URL 覆盖）。
func sensorHTTPBaseURL() string {
	if u := os.Getenv("SENSOR_HTTP_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "http://127.0.0.1:8087"
}

type vetoCanvasObject struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Geometry struct {
		Data struct {
			Vertices []struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"vertices"`
		} `json:"data"`
	} `json:"geometry"`
}

// feedbackRects 解析 canvas，返回 source='Feedback' object 的 id → 轴对齐包围盒 [x1,y1,x2,y2] cm。
func feedbackRects(canvas []byte) map[string][4]int {
	out := map[string][4]int{}
	if len(canvas) == 0 {
		return out
	}
	var doc struct {
		Objects []vetoCanvasObject `json:"objects"`
	}
	if err := json.Unmarshal(canvas, &doc); err != nil {
		return out
	}
	for _, o := range doc.Objects {
		if o.Source != "Feedback" || o.ID == "" {
			continue
		}
		v := o.Geometry.Data.Vertices
		if len(v) < 4 {
			continue
		}
		minX, minY := v[0].X, v[0].Y
		maxX, maxY := v[0].X, v[0].Y
		for _, p := range v[1:] {
			minX, minY = math.Min(minX, p.X), math.Min(minY, p.Y)
			maxX, maxY = math.Max(maxX, p.X), math.Max(maxY, p.Y)
		}
		out[o.ID] = [4]int{int(math.Round(minX)), int(math.Round(minY)), int(math.Round(maxX)), int(math.Round(maxY))}
	}
	return out
}

// detectAndNotifyFeedbackVetoes diff 旧/新 canvas：旧有、新无的 Feedback object = 被删 = 否决。
// 仅 /128 device-scope 行有 Feedback object（sensor 只往 /128 append）。device_addr = spatial_prefix host。
func (s *RadarInstall) detectAndNotifyFeedbackVetoes(ctx context.Context, spatialPrefix string, oldCanvas, newCanvas []byte) {
	if len(oldCanvas) == 0 {
		return // 新建行，无旧 Feedback
	}
	if masklen, err := masklenOfCIDR(spatialPrefix); err != nil || masklen != 128 {
		return
	}
	oldFb := feedbackRects(oldCanvas)
	if len(oldFb) == 0 {
		return
	}
	newFb := feedbackRects(newCanvas)
	deviceAddr := strings.SplitN(spatialPrefix, "/", 2)[0]
	for id, rect := range oldFb {
		if _, kept := newFb[id]; kept {
			continue
		}
		// 删 Feedback pin 对象 → 强清该 rect 回 Unknown；SaveRoomLayout 随后 applyDeclare→stampCanvasCells
		// 按剩余 layout 重刷，叠加区（如删 reflector 露出底下 chair）自动盖回，未覆盖区留 Unknown。
		s.notifySensorClear(ctx, deviceAddr, rect[0], rect[1], rect[2], rect[3], id)
	}
}

// notifySensorVeto POST sensor /roomengine/cell/veto {device_addr, x1,y1,x2,y2, sticky}。best-effort，失败仅 warn。
// sticky=false（Clear / 删 layout Feedback 对象）仅清 rect；sticky=true（handle "Never re-learn"）额外 MarkLearnBlocked。
func (s *RadarInstall) notifySensorVeto(ctx context.Context, deviceAddr string, x1, y1, x2, y2 int, sticky bool, objID string) {
	body, _ := json.Marshal(map[string]interface{}{"device_addr": deviceAddr, "x1": x1, "y1": y1, "x2": x2, "y2": y2, "sticky": sticky})
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, sensorHTTPBaseURL()+"/roomengine/cell/veto", bytes.NewReader(body))
	if err != nil {
		s.logger.Warn("notifySensorVeto: build request", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.logger.Warn("notifySensorVeto: call sensor failed (cell mark stays until next reload)",
			zap.String("device_addr", deviceAddr), zap.String("object_id", objID), zap.Error(err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		s.logger.Warn("notifySensorVeto: sensor rejected (route/version mismatch? cell not cleared)",
			zap.String("device_addr", deviceAddr), zap.String("object_id", objID), zap.Int("status", resp.StatusCode))
		return
	}
	s.logger.Info("feedback_veto_notified_sensor",
		zap.String("device_addr", deviceAddr), zap.String("object_id", objID), zap.Bool("sticky", sticky),
		zap.Int("x1", x1), zap.Int("y1", y1), zap.Int("x2", x2), zap.Int("y2", y2), zap.Int("status", resp.StatusCode))
}

// notifySensorClear POST sensor /roomengine/cell/clear：删 Feedback pin 对象时强清该 rect 回 Unknown。
// best-effort；非 2xx → Warn（半部署/旧路由/grid 未建可见）。叠加区盖回由随后 stampCanvasCells 负责。
func (s *RadarInstall) notifySensorClear(ctx context.Context, deviceAddr string, x1, y1, x2, y2 int, objID string) {
	body, _ := json.Marshal(map[string]interface{}{"device_addr": deviceAddr, "x1": x1, "y1": y1, "x2": x2, "y2": y2})
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, sensorHTTPBaseURL()+"/roomengine/cell/clear", bytes.NewReader(body))
	if err != nil {
		s.logger.Warn("notifySensorClear: build request", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.logger.Warn("notifySensorClear: call sensor failed (cell clear waits for reload)",
			zap.String("device_addr", deviceAddr), zap.String("object_id", objID), zap.Error(err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		s.logger.Warn("notifySensorClear: sensor rejected (route/version mismatch? or grid not built — cell not cleared)",
			zap.String("device_addr", deviceAddr), zap.String("object_id", objID), zap.Int("status", resp.StatusCode))
		return
	}
	s.logger.Info("feedback_clear_notified_sensor",
		zap.String("device_addr", deviceAddr), zap.String("object_id", objID),
		zap.Int("x1", x1), zap.Int("y1", y1), zap.Int("x2", x2), zap.Int("y2", y2), zap.Int("status", resp.StatusCode))
}

// notifySensorStamp POST sensor /roomengine/cell/stamp：layout 写入后即时刷 grid 完整 rect（pin / FE resize）。
// best-effort，失败仅 warn（cell 等下次 config:card reload 落地）。
func (s *RadarInstall) notifySensorStamp(ctx context.Context, deviceAddr string, x1, y1, x2, y2, areaType, conf int) {
	body, _ := json.Marshal(map[string]interface{}{
		"device_addr": deviceAddr, "x1": x1, "y1": y1, "x2": x2, "y2": y2,
		"area_type": areaType, "conf": conf,
	})
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, sensorHTTPBaseURL()+"/roomengine/cell/stamp", bytes.NewReader(body))
	if err != nil {
		s.logger.Warn("notifySensorStamp: build request", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.logger.Warn("notifySensorStamp: call sensor failed (cell stamp waits for reload)",
			zap.String("device_addr", deviceAddr), zap.Error(err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		s.logger.Warn("notifySensorStamp: sensor rejected (route/version mismatch? or room grid not built — cell not stamped)",
			zap.String("device_addr", deviceAddr), zap.Int("area_type", areaType), zap.Int("status", resp.StatusCode))
		return
	}
	s.logger.Info("cell_stamp_notified_sensor",
		zap.String("device_addr", deviceAddr), zap.Int("area_type", areaType), zap.Int("status", resp.StatusCode))
}
