// ai_overrides.go — cardagg 端 AI 派生 track verdict 缓存（S5a port from v1）。
//
// 数据流：
//   wisefido-sensor RoomEngine track_manager.emitGhostVerdict
//        → engine.PublishAIEvent category="track_verdict"
//        → ai:track:verdict:stream（专用流，30s TTL；CLAUDE.md 规则 #2.1 不入库）
//        → cardagg AIVerdictHandler.Set(deviceAddr, trackID, AIVerdict)
//        → MonitorHandler.Handle 在 buffer.Write 前 Apply 合并到 track fields
//        → realtime snapshot 带 AI 调整后的 confidence
//
// 关键设计（与 v1 一致）：
//
//  1. **alarm 路径不受影响** —— 本缓存仅作用于 UI 合并，绝不参与 alarm 触发判定。
//     "宁可误报不可漏报"：AI 的 ghost 裁决不能压制 firmware/AI 的 fall alarm。
//
//  2. **release / sandbox 双模式**：
//     · sandbox（默认）：仅 log，Apply 不修改任何 track 字段 —— 演示"AI 在思考"
//     · release：Apply 把 confidence 覆写到 monitor 流 track，前端按低饱和度渲染
//
//  3. **4 个清理触发**（按事件优先级）：
//     a. tid=88 帧（firmware no-target heartbeat）→ ClearDevice（MonitorHandler）
//     b. EnterRoom 事件 → ClearDevice（EventHandler）
//     c. ExitRoom 事件 → ClearDevice（EventHandler）
//     d. TTL 60s 兜底（GC 定时跑）→ 防上面三个信号丢失/延迟
//
//  4. **device_addr + track_id 复合 key**：track_id 是 firmware 0-15 循环槽位，
//     不同 device 间不可共享 verdict。

package service

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// AIVerdict 单条 AI 裁决（一个 track 维度）。
type AIVerdict struct {
	Confidence int    // 0-100；AI 写的连续值，下游按阈值派生类别（如 ≤30 → ghost UI）
	Source     string // 决策路径，如 "AI.Caregiver01"
	Reason     string // ghost_post_real / ghost_penalty / ghost_low_score 等
	UpdatedMs  int64  // 入池/最近刷新时间（TTL 用）
}

// AIOverrideMode publish/合并行为模式。
type AIOverrideMode string

const (
	AIOverrideModeSandbox AIOverrideMode = "sandbox" // 默认：仅 log，不动 track 字段
	AIOverrideModeRelease AIOverrideMode = "release" // 合并到 track_confidence
)

func IsValidAIOverrideMode(s string) bool {
	switch AIOverrideMode(s) {
	case AIOverrideModeSandbox, AIOverrideModeRelease:
		return true
	}
	return false
}

// AIOverrideCache cardagg 端 AI 裁决缓存。线程安全。
type AIOverrideCache struct {
	mu       sync.RWMutex
	byDevice map[string]map[int]AIVerdict // device_addr → track_id → verdict
	mode     AIOverrideMode
	ttlMs    int64
	logger   *zap.Logger

	setCount    int64
	clearCount  int64
	applyHits   int64
	applyMisses int64
	gcRemoved   int64
}

// NewAIOverrideCache 构造。mode 非 "release"/"sandbox" 一律退回 sandbox。ttlSec ≤ 0 默认 60s。
func NewAIOverrideCache(mode string, ttlSec int, logger *zap.Logger) *AIOverrideCache {
	m := AIOverrideMode(mode)
	if !IsValidAIOverrideMode(mode) {
		m = AIOverrideModeSandbox
	}
	if ttlSec <= 0 {
		ttlSec = 60
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AIOverrideCache{
		byDevice: make(map[string]map[int]AIVerdict),
		mode:     m,
		ttlMs:    int64(ttlSec) * 1000,
		logger:   logger,
	}
}

// Set 写入裁决；track_verdict 事件到达时调。
func (c *AIOverrideCache) Set(deviceAddr string, trackID int, v AIVerdict) {
	if deviceAddr == "" {
		return
	}
	if v.UpdatedMs == 0 {
		v.UpdatedMs = time.Now().UnixMilli()
	}
	c.mu.Lock()
	m, ok := c.byDevice[deviceAddr]
	if !ok {
		m = make(map[int]AIVerdict)
		c.byDevice[deviceAddr] = m
	}
	m[trackID] = v
	c.setCount++
	c.mu.Unlock()
	c.logger.Info("ai.score.recv",
		zap.String("mode", string(c.mode)),
		zap.String("device_addr", deviceAddr),
		zap.Int("track_id", trackID),
		zap.Int("confidence", v.Confidence),
		zap.String("source", v.Source),
		zap.String("reason", v.Reason),
	)
}

// Apply 合并 verdict 到 track fields（monitor consumer 在 buffer.Write 前调）。
//
// release 模式：覆写 fields["track_confidence"] + 写 fields["ai_source"]。
// sandbox 模式：仅 log，fields 不动。
func (c *AIOverrideCache) Apply(deviceAddr string, trackID int, fields map[string]interface{}) {
	if deviceAddr == "" || fields == nil {
		return
	}
	c.mu.RLock()
	m, ok := c.byDevice[deviceAddr]
	if !ok {
		c.mu.RUnlock()
		c.bumpMiss()
		return
	}
	v, ok := m[trackID]
	mode := c.mode
	c.mu.RUnlock()
	if !ok {
		c.bumpMiss()
		return
	}

	origConf := readIntFieldAI(fields, "track_confidence")
	if mode == AIOverrideModeRelease {
		fields["track_confidence"] = v.Confidence
		fields["ai_source"] = v.Source
	}
	c.bumpHit()
	changed := mode == AIOverrideModeRelease && origConf != v.Confidence
	logFields := []zap.Field{
		zap.String("mode", string(mode)),
		zap.String("device_addr", deviceAddr),
		zap.Int("track_id", trackID),
		zap.Int("orig_confidence", origConf),
		zap.Int("ai_confidence", v.Confidence),
		zap.String("source", v.Source),
		zap.Bool("merged", mode == AIOverrideModeRelease),
	}
	if changed {
		c.logger.Info("ai.score.apply.changed", logFields...)
	} else {
		c.logger.Debug("ai.score.apply.noop", logFields...)
	}
}

// ClearDevice 清空某 device 全部 verdicts（tid=88 / Enter/ExitRoom / device offline 调）。
func (c *AIOverrideCache) ClearDevice(deviceAddr string) {
	if deviceAddr == "" {
		return
	}
	c.mu.Lock()
	m, ok := c.byDevice[deviceAddr]
	if !ok {
		c.mu.Unlock()
		return
	}
	cleared := len(m)
	delete(c.byDevice, deviceAddr)
	c.clearCount++
	c.mu.Unlock()
	if cleared > 0 {
		c.logger.Info("ai.score.cleared",
			zap.String("device_addr", deviceAddr),
			zap.Int("verdicts_dropped", cleared),
		)
	}
}

// GC 清理过期条目（TTL 兜底），返回删除条数。
func (c *AIOverrideCache) GC(nowMs int64) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	removed := 0
	cutoff := nowMs - c.ttlMs
	for addr, m := range c.byDevice {
		for tid, v := range m {
			if v.UpdatedMs < cutoff {
				delete(m, tid)
				removed++
			}
		}
		if len(m) == 0 {
			delete(c.byDevice, addr)
		}
	}
	c.gcRemoved += int64(removed)
	if removed > 0 {
		c.logger.Debug("ai_verdict_gc",
			zap.Int("removed", removed),
			zap.Int64("ttl_ms", c.ttlMs),
		)
	}
	return removed
}

// RunGCLoop 每 interval 跑一次 GC，直到 stop close。
func (c *AIOverrideCache) RunGCLoop(stop <-chan struct{}, interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case now := <-t.C:
			c.GC(now.UnixMilli())
		}
	}
}

func (c *AIOverrideCache) Mode() AIOverrideMode {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mode
}

func (c *AIOverrideCache) SetMode(mode string) {
	if !IsValidAIOverrideMode(mode) {
		return
	}
	c.mu.Lock()
	c.mode = AIOverrideMode(mode)
	c.mu.Unlock()
}

type AIOverrideStats struct {
	Devices     int
	Tracks      int
	SetCount    int64
	ClearCount  int64
	ApplyHits   int64
	ApplyMisses int64
	GCRemoved   int64
}

func (c *AIOverrideCache) Stats() AIOverrideStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	tracks := 0
	for _, m := range c.byDevice {
		tracks += len(m)
	}
	return AIOverrideStats{
		Devices:     len(c.byDevice),
		Tracks:      tracks,
		SetCount:    c.setCount,
		ClearCount:  c.clearCount,
		ApplyHits:   c.applyHits,
		ApplyMisses: c.applyMisses,
		GCRemoved:   c.gcRemoved,
	}
}

func (c *AIOverrideCache) bumpHit() {
	c.mu.Lock()
	c.applyHits++
	c.mu.Unlock()
}

func (c *AIOverrideCache) bumpMiss() {
	c.mu.Lock()
	c.applyMisses++
	c.mu.Unlock()
}

func readIntFieldAI(m map[string]interface{}, key string) int {
	if m == nil {
		return 0
	}
	switch x := m[key].(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	}
	return 0
}
