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
	"time"

	"go.uber.org/zap"

	"wisefido-sensor/internal/roomengine"
	"wisefido-sensor/internal/zoneengine/wiring"
)

type cellVetoRequest struct {
	DeviceAddr string `json:"device_addr"` // /128 device host text
	X          int    `json:"x"`           // canvas cm
	Y          int    `json:"y"`           // canvas cm
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
			http.Error(w, "bad request (need device_addr + x,y)", http.StatusBadRequest)
			return
		}
		cleared, blocked, ok := engine.VetoCell(req.DeviceAddr, req.X, req.Y, time.Now().UnixMilli())
		if !ok {
			http.Error(w, "device not routed / cell out of grid", http.StatusNotFound)
			return
		}
		logger.Info("cell_veto_applied",
			zap.String("device_addr", req.DeviceAddr), zap.Int("x", req.X), zap.Int("y", req.Y),
			zap.Bool("cleared", cleared), zap.Bool("learn_blocked", blocked))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]bool{"cleared": cleared, "learn_blocked": blocked})
	})

	// 事件触发：wisefido-data handle 完即调，即时处理单条 alarm 反馈（event_id 去重）。
	mux.HandleFunc("/roomengine/feedback/ingest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			EventID string `json:"event_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.EventID == "" {
			http.Error(w, "bad request (need event_id)", http.StatusBadRequest)
			return
		}
		processed, err := engine.IngestFeedback(r.Context(), req.EventID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger.Info("feedback_ingest_applied", zap.String("event_id", req.EventID), zap.Bool("processed", processed))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"processed": processed})
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
