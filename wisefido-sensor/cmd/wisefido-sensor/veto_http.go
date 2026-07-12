package main

// veto_http.go — 轻量 HTTP server，收人否决 Feedback 学习区的通知（[[layout_authority_ai_correction_model]]）。
//
// 通道分工（用户 2026-06-04 定）：vendor 下发走 qinglan RadarAPI（不碰）；veto 是我们自己的语义，
// 走这条独立 HTTP——wisefido-data 在 layout save diff 出被删的 source='Feedback' object 后直调本端点。
// 单一消费者（只有 sensor 持 live engine 能改 cell），故不开 Redis pub/sub，HTTP 直调更简单。

import (
	"context"
	"encoding/json"
	"net/http"
	"net/netip"
	"strconv"
	"time"

	"go.uber.org/zap"

	"owl-common/radarutils"

	"wisefido-sensor/internal/roomengine"
	"wisefido-sensor/internal/zoneengine/wiring"
)

type cellVetoRequest struct {
	DeviceAddr string `json:"device_addr"` // /128 device host text
	X1         int    `json:"x1"`          // canvas cm，fire 点 40×40 足迹左上
	Y1         int    `json:"y1"`
	X2         int    `json:"x2"` // 右下
	Y2         int    `json:"y2"`
	Sticky     bool   `json:"sticky"` // true=handle "Never re-learn" → 额外 MarkLearnBlocked
}

// cellStampRequest data 在 layout 写入（pin / FE resize save）后调，被动刷 grid 完整 rect。
type cellStampRequest struct {
	DeviceAddr string `json:"device_addr"` // /128 device host text
	X1         int    `json:"x1"`          // canvas cm
	Y1         int    `json:"y1"`
	X2         int    `json:"x2"`
	Y2         int    `json:"y2"`
	AreaType   int    `json:"area_type"` // observation.AreaType 0-9
	Conf       int    `json:"conf"`
}

// chairMaxSitRequest data 在 handle(false_alarm + "Sit on Chair") 后调，棘轮抬该椅 anchor maxSit。
type chairMaxSitRequest struct {
	DeviceAddr string  `json:"device_addr"` // /128 device host text
	X          int     `json:"x"`           // fire 点 canvas cm（evidence.fire.x）
	Y          int     `json:"y"`           // evidence.fire.y
	SitDurSec  float64 `json:"sit_dur_sec"` // 实际久坐 still 秒（evidence.fire.stillbox_sec）
}

// startVetoHTTPServer 在 addr 起 mux，POST /roomengine/cell/veto → engine.VetoCell。
// 进程退出（ctx done）时优雅关闭。
func startVetoHTTPServer(ctx context.Context, addr string, engine *roomengine.Engine, matrix *wiring.MatrixCache, logger *zap.Logger) {
	mux := http.NewServeMux()

	// GET /matrix?unit=<prefix> → dump MM 关系方阵（debug 验证 covers/samebed）。
	mux.HandleFunc("/matrix", func(w http.ResponseWriter, r *http.Request) {
		unit, err := netip.ParsePrefix(r.URL.Query().Get("unit"))
		if err != nil {
			http.Error(w, "bad unit prefix", http.StatusBadRequest)
			return
		}
		d := matrix.Dump(unit)
		if d == nil {
			http.Error(w, "unit not in matrix", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(d)
	})

	mux.HandleFunc("/roomengine/cell/veto", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req cellVetoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DeviceAddr == "" {
			http.Error(w, "bad request (need device_addr + rect)", http.StatusBadRequest)
			return
		}
		rect := radarutils.Rect{X1: req.X1, Y1: req.Y1, X2: req.X2, Y2: req.Y2}
		cleared, blocked, ok := engine.VetoRect(req.DeviceAddr, rect, req.Sticky)
		if !ok {
			http.Error(w, "device not routed / grid not built", http.StatusNotFound)
			return
		}
		logger.Info("cell_veto_applied",
			zap.String("device_addr", req.DeviceAddr),
			zap.Int("x1", req.X1), zap.Int("y1", req.Y1), zap.Int("x2", req.X2), zap.Int("y2", req.Y2),
			zap.Bool("sticky", req.Sticky), zap.Int("cleared", cleared), zap.Int("learn_blocked", blocked))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]int{"cleared": cleared, "learn_blocked": blocked})
	})

	// data 在 layout 写入后调，被动刷 grid 完整 rect（pin 初始正方形 / FE resize 后完整矩形）。零业务判断。
	mux.HandleFunc("/roomengine/cell/stamp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req cellStampRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DeviceAddr == "" {
			http.Error(w, "bad request (need device_addr)", http.StatusBadRequest)
			return
		}
		// layout 已写库；声明区不再烙 belief，直接重建该房 grid，QueryAreaType 即见新 AreaZones。
		ok := engine.ReloadRoomByDevice(r.Context(), req.DeviceAddr)
		if !ok {
			http.Error(w, "device not routed / room grid not built", http.StatusNotFound)
			return
		}
		logger.Info("cell_stamp_reload", zap.String("device_addr", req.DeviceAddr))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]bool{"stamped": true})
	})

	// data 删 Feedback pin 对象时调，强清该 rect 回 Unknown（随后 data 重刷剩余 layout 盖回叠加区）。零业务判断。
	mux.HandleFunc("/roomengine/cell/clear", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req cellStampRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DeviceAddr == "" {
			http.Error(w, "bad request (need device_addr)", http.StatusBadRequest)
			return
		}
		// 删对象后 layout 已写库；重建该房 grid，剩余声明区经 AreaZones 自然生效。
		ok := engine.ReloadRoomByDevice(r.Context(), req.DeviceAddr)
		if !ok {
			http.Error(w, "device not routed / room grid not built", http.StatusNotFound)
			return
		}
		logger.Info("cell_clear_reload", zap.String("device_addr", req.DeviceAddr))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]bool{"cleared": true})
	})

	// data 在 handle(false_alarm) 后调：棘轮抬椅/浴室 maxSit（门控在 data；本端零业务判断，浴室房无椅也成功）。
	// x,y=fire 点 canvas cm；sit_dur_sec=evidence.fire.stillbox_sec。
	mux.HandleFunc("/roomengine/chair/maxsit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req chairMaxSitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DeviceAddr == "" || req.SitDurSec <= 0 {
			http.Error(w, "bad request (need device_addr + x,y + sit_dur_sec)", http.StatusBadRequest)
			return
		}
		ok := engine.RatchetChairMaxSitByDevice(req.DeviceAddr, req.X, req.Y, req.SitDurSec, time.Now().UnixMilli())
		if !ok {
			http.Error(w, "device not routed / not bathroom and point not in any chair pin", http.StatusNotFound)
			return
		}
		logger.Info("chair_maxsit_ratcheted",
			zap.String("device_addr", req.DeviceAddr), zap.Int("x", req.X), zap.Int("y", req.Y),
			zap.Float64("sit_dur_sec", req.SitDurSec))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]bool{"ratcheted": true})
	})

	// 只读诊断：GET /roomengine/cell/at?device_addr=&x=&y= → 该 canvas 点当前 cell 的 AreaType。
	mux.HandleFunc("/roomengine/cell/at", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		addr := q.Get("device_addr")
		x, _ := strconv.Atoi(q.Get("x"))
		y, _ := strconv.Atoi(q.Get("y"))
		if addr == "" {
			http.Error(w, "bad request (need device_addr + x,y)", http.StatusBadRequest)
			return
		}
		area, name, conf, ok := engine.CellAreaAt(addr, x, y)
		if !ok {
			http.Error(w, "device not routed / point out of grid", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"x": x, "y": y, "area_type": area, "area_name": name, "conf": conf,
		})
	})

	mux.HandleFunc("/roomengine/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()
	go func() {
		logger.Info("sensor veto HTTP server listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Warn("sensor veto HTTP server stopped", zap.Error(err))
		}
	}()
}
