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

// feedbackCenters 解析 canvas，返回 source='Feedback' object 的 id → 中心 (x,y) cm。
func feedbackCenters(canvas []byte) map[string][2]int {
	out := map[string][2]int{}
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
		var sx, sy float64
		for _, p := range v {
			sx += p.X
			sy += p.Y
		}
		n := float64(len(v))
		out[o.ID] = [2]int{int(sx / n), int(sy / n)}
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
	oldFb := feedbackCenters(oldCanvas)
	if len(oldFb) == 0 {
		return
	}
	newFb := feedbackCenters(newCanvas)
	deviceAddr := strings.SplitN(spatialPrefix, "/", 2)[0]
	for id, center := range oldFb {
		if _, kept := newFb[id]; kept {
			continue
		}
		s.notifySensorVeto(ctx, deviceAddr, center[0], center[1], id)
	}
}

// notifySensorVeto POST sensor /roomengine/cell/veto {device_addr, x, y}。best-effort，失败仅 warn。
func (s *RadarInstall) notifySensorVeto(ctx context.Context, deviceAddr string, x, y int, objID string) {
	body, _ := json.Marshal(map[string]interface{}{"device_addr": deviceAddr, "x": x, "y": y})
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
	s.logger.Info("feedback_veto_notified_sensor",
		zap.String("device_addr", deviceAddr), zap.String("object_id", objID),
		zap.Int("x", x), zap.Int("y", y), zap.Int("status", resp.StatusCode))
}
