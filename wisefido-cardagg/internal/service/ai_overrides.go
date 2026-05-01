// ai_overrides.go
//
// PR6 — AI 派生 track verdict 的 cardagg 端缓存。
//
// wisefido-ai 通过 iot:event:stream + category=track_verdict 发出"事后裁决"
// （目前主要是 ghost 判定，写入 track_confidence=20）。本缓存接收并按需合并到
// monitor 流的 track 字段，让前端 UI 用 AI 调整后的 confidence 渲染。
//
// 关键设计：
//
//  1. **alarm 路径不受影响**——本缓存仅作用于 UI 合并，绝不参与 alarm 触发判定。
//     用户原则："宁可误报不可漏报"，AI 的 ghost 裁决不能压制 firmware/AI 的 fall alarm。
//
//  2. **release / sandbox 双模式**：
//     - sandbox（默认）：仅记 log，Apply 不修改任何 track 字段——演示"AI 在思考"
//     - release：Apply 把 confidence 覆写到 monitor 流 track，前端按低饱和度渲染
//
//  3. **4 个清理触发**（按事件优先级）：
//     a. tid=88 帧（firmware no-target heartbeat）→ ClearDevice
//     b. EnterRoom 事件 → ClearDevice（新人进，track_id 可能复用，旧 verdict 必须作废）
//     c. ExitRoom 事件 → ClearDevice
//     d. TTL 60s 兜底（GC 定时跑）→ 防上面三个信号丢失/延迟
//
//  4. **device_uid + track_id 复合 key**：track_id 是 firmware 0-15 的循环槽位，
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
	Source     string // 决策路径，如 "ai_ghost_penalty"
	Reason     string // 简短文本（可空）
	UpdatedMs  int64  // 入池/最近刷新时间（TTL 用）
}

// AIOverrideMode publish/合并行为模式。
type AIOverrideMode string

const (
	AIOverrideModeSandbox AIOverrideMode = "sandbox" // 默认：仅 log，不动 track 字段
	AIOverrideModeRelease AIOverrideMode = "release" // 合并到 track_confidence
)

// IsValidAIOverrideMode 校验模式合法性。
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
	byDevice map[string]map[int]AIVerdict // device_uid → track_id → verdict
	mode     AIOverrideMode
	ttlMs    int64
	logger   *zap.Logger

	// 计数器（监控用）
	setCount     int64
	clearCount   int64
	applyHits    int64
	applyMisses int64
	gcRemoved    int64
}

// NewAIOverrideCache 构造缓存。mode 非 "release"/"sandbox" 一律退回 sandbox。
// ttlSec ≤ 0 时取默认 60s。
func NewAIOverrideCache(mode string, ttlSec int, logger *zap.Logger) *AIOverrideCache {
	m := AIOverrideMode(mode)
	if !IsValidAIOverrideMode(mode) {
		m = AIOverrideModeSandbox
	}
	if ttlSec <= 0 {
		ttlSec = 60
	}
	return &AIOverrideCache{
		byDevice: make(map[string]map[int]AIVerdict),
		mode:     m,
		ttlMs:    int64(ttlSec) * 1000,
		logger:   logger,
	}
}

// Set 写入裁决。track_verdict 事件到达时调用。
func (c *AIOverrideCache) Set(deviceUID string, trackID int, v AIVerdict) {
	if deviceUID == "" {
		return
	}
	if v.UpdatedMs == 0 {
		v.UpdatedMs = time.Now().UnixMilli()
	}
	c.mu.Lock()
	m, ok := c.byDevice[deviceUID]
	if !ok {
		m = make(map[int]AIVerdict)
		c.byDevice[deviceUID] = m
	}
	m[trackID] = v
	c.setCount++
	c.mu.Unlock()
	if c.logger != nil {
		c.logger.Info("ai_verdict_cached",
			zap.String("mode", string(c.mode)),
			zap.String("device_uid", deviceUID),
			zap.Int("track_id", trackID),
			zap.Int("confidence", v.Confidence),
			zap.String("source", v.Source),
		)
	}
}

// Apply 合并 verdict 到 track fields。
//
// release 模式：覆写 fields["track_confidence"] + 写 fields["ai_source"]（前端可见）。
// sandbox 模式：仅 log，fields 不动。
//
// 调用方：monitor consumer 在 buffer.Write 前调用——这样合并后的 track fields 会
// 直接进 cardagg 的 realtime snapshot，前端拿到的 track_confidence 就是 AI 调整后的值。
func (c *AIOverrideCache) Apply(deviceUID string, trackID int, fields map[string]interface{}) {
	if deviceUID == "" || fields == nil {
		return
	}
	c.mu.RLock()
	m, ok := c.byDevice[deviceUID]
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

	origConf := readIntField(fields, "track_confidence")
	if mode == AIOverrideModeRelease {
		fields["track_confidence"] = v.Confidence
		fields["ai_source"] = v.Source
	}
	c.bumpHit()
	if c.logger != nil {
		c.logger.Info("ai_verdict_applied",
			zap.String("mode", string(mode)),
			zap.String("device_uid", deviceUID),
			zap.Int("track_id", trackID),
			zap.Int("orig_confidence", origConf),
			zap.Int("ai_confidence", v.Confidence),
			zap.String("source", v.Source),
			zap.Bool("merged", mode == AIOverrideModeRelease),
		)
	}
}

// ClearDevice 清空某 device 的所有 verdicts。tid=88 / Enter/ExitRoom / device offline 调用。
func (c *AIOverrideCache) ClearDevice(deviceUID string) {
	if deviceUID == "" {
		return
	}
	c.mu.Lock()
	m, ok := c.byDevice[deviceUID]
	if !ok {
		c.mu.Unlock()
		return
	}
	cleared := len(m)
	delete(c.byDevice, deviceUID)
	c.clearCount++
	c.mu.Unlock()
	if c.logger != nil && cleared > 0 {
		c.logger.Info("ai_verdict_cleared",
			zap.String("device_uid", deviceUID),
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
	for uid, m := range c.byDevice {
		for tid, v := range m {
			if v.UpdatedMs < cutoff {
				delete(m, tid)
				removed++
			}
		}
		if len(m) == 0 {
			delete(c.byDevice, uid)
		}
	}
	c.gcRemoved += int64(removed)
	if removed > 0 && c.logger != nil {
		c.logger.Debug("ai_verdict_gc",
			zap.Int("removed", removed),
			zap.Int64("ttl_ms", c.ttlMs),
		)
	}
	return removed
}

// RunGCLoop 每 interval 跑一次 GC，直到 ctx.Done()。供 main 启动 goroutine。
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

// Mode 返回当前模式。
func (c *AIOverrideCache) Mode() AIOverrideMode {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mode
}

// SetMode 运行时切换（演示场景动态切换 sandbox↔release）。
func (c *AIOverrideCache) SetMode(mode string) {
	if !IsValidAIOverrideMode(mode) {
		return
	}
	c.mu.Lock()
	c.mode = AIOverrideMode(mode)
	c.mu.Unlock()
}

// Stats 返回当前缓存大小 + 计数器（监控用）。
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

// readIntField 从 jsonb-decoded map 取 int（接受 float64/int/int64/json.Number）。
func readIntField(m map[string]interface{}, key string) int {
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
