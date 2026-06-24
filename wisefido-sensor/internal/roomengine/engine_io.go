package roomengine

// engine_io.go — 生产 I/O 焊回层（S0.c）：Xsensor replay 道裁掉了发布/持久化/反馈/daily-reload，
// 生产必需。方法体逐字复原自旧 ws engine.go（git HEAD），字段已加进 engine.go 的 Engine struct。

import (
	"context"
	"database/sql"
	"fmt"
	"net/netip"
	"time"

	"owl-common/observation"
	rediscommon "owl-common/redis"

	"go.uber.org/zap"
)

func (e *Engine) SetDailyLayoutReload(hourLocal int, db *sql.DB) {
	if hourLocal < -1 || hourLocal > 23 {
		hourLocal = -1
	}
	e.dailyReloadHour = hourLocal
	e.dailyReloadDB = db
}

func (e *Engine) recordLastSrcSeq(deviceAddr string, seq uint64) {
	if deviceAddr == "" || seq == 0 {
		return
	}
	e.srcSeqMu.Lock()
	e.lastSrcSeq[deviceAddr] = seq
	e.srcSeqMu.Unlock()
}

func (e *Engine) readLastSrcSeq(deviceAddr string) uint64 {
	if deviceAddr == "" {
		return 0
	}
	e.srcSeqMu.RLock()
	defer e.srcSeqMu.RUnlock()
	return e.lastSrcSeq[deviceAddr]
}

func (e *Engine) nextAgentSeq(ctx context.Context, producer string) int64 {
	if e.redisClient == nil || producer == "" {
		return 0
	}
	key := "wisefido-sensor:seq:" + producer
	v, err := e.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return 0
	}
	return v
}

func (e *Engine) SetAIPublishConfig(mode, source string) {
	if mode == "" {
		mode = "log&publish"
	}
	if source == "" {
		source = "sensor.caregiver01"
	}
	e.mu.Lock()
	e.aiPublishMode = mode
	e.aiSource = source
	e.mu.Unlock()
}

func (e *Engine) publishEnabled() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.aiPublishMode == "log&publish"
}

func (e *Engine) PublishAIEvent(ctx context.Context, p AIPayload, category string, nowMs int64) {
	streamDef := rediscommon.StreamEvent
	if category == "track_verdict" || category == "ghost" || category == CategorySensorDecision {
		streamDef = rediscommon.StreamAITrackVerdict
	}
	e.publishAIMessage(ctx, p, category, "event",
		streamDef.Name, streamDef.MaxLen, streamDef.RetentionSeconds, nowMs)
}

func (e *Engine) PublishAIAlarm(ctx context.Context, p AIPayload, category string, nowMs int64) {
	e.publishAIMessage(ctx, p, category, "alarm",
		rediscommon.StreamAlarm.Name,
		rediscommon.StreamAlarm.MaxLen, rediscommon.StreamAlarm.RetentionSeconds, nowMs)
}

func (e *Engine) publishAIMessage(ctx context.Context, p AIPayload,
	category, topicType, streamName string, maxLen int64, retentionSec int, nowMs int64) {
	// p.DeviceAddr 现为 canonical IPv6 字符串（上游已切；engine 内部 Map 同步切）
	addr, _ := netip.ParseAddr(p.DeviceAddr)
	e.mu.RLock()
	baseType := e.deviceAddrToType[p.DeviceAddr]
	defaultSource := e.aiSource
	mode := e.aiPublishMode
	g := e.grids[p.RoomID]
	e.mu.RUnlock()
	// SubjectEntity 留空：sensor 不做 device→card 反查（非其职责）；
	// cardagg alarm_router 在 SubjectEntity 空时调 metaCache.LookupCardByDeviceAddr LPM 兜底。

	if baseType == "" {
		baseType = "Radar" // 兜底：路由表缺失时默认按 radar 派生
	}
	// device_type 保持源 sensor 类型（不再拼 ".AI<NodeID>" 后缀）。
	// AI 派生身份由 fields["source"] 一等公民字段表达。
	deviceType := baseType

	// alarm/event 流统一走 EventItem 契约（与 qinglan/sleepace publisher 一致）：
	// EventItem 提供生命周期 + first-class 业务字段（TrackID/Pose/HeartRate/RespiratoryRate）；
	// 其余 sensor-specific 字段（position / track_confidence / area_type / source / reason / evidence）
	// 作为 dataValue map 同级平铺补充。
	//
	// EventStatus：默认 "instant"；payload 显式指定（如 RecordRadarAlarm forward Initialization→end）
	// 时覆盖，让下游 cardagg AlarmRouter 按 EndPolicy=AutoResolve 关 alarm。
	eventStatus := p.EventStatus
	if eventStatus == "" {
		eventStatus = "instant"
	}
	item := observation.NewEventItem(nowMs, eventStatus)
	item.TrackID = p.Track.TrackID
	if p.Track.Pose != 0 {
		item.Pose = p.Track.Pose
	}
	if p.Track.HeartRate != 0 {
		item.HeartRate = p.Track.HeartRate
	}
	if p.Track.RespiratoryRate != 0 {
		item.RespiratoryRate = p.Track.RespiratoryRate
	}
	fields, _ := observation.EventItemToDataMap(&item)
	if fields == nil {
		fields = make(map[string]interface{})
	}
	// 显式补 track_id：EventItem.TrackID 是 omitempty，track_id=0（合法雷达 track）会被省，
	// 致下游 cardagg ai_override（按 (device,track_id) 键）收不到 track_id=0 的 track → FE 透明度失效。
	fields[observation.FieldTrackID] = p.Track.TrackID
	// alarm_level 不由 sensor 盖：由 cardagg 从 device_config 单点决定（alarm_router Resolve）。
	// sensor 只在源头 gate is_enabled（发 alarm vs event）+ 时间型阈值，不碰 level。
	// sensor-specific 业务扩展字段平铺
	if p.Track.PositionX != nil {
		fields[observation.FieldPositionX] = *p.Track.PositionX
	}
	if p.Track.PositionY != nil {
		fields[observation.FieldPositionY] = *p.Track.PositionY
	}
	if p.Track.PositionZ != nil {
		fields[observation.FieldPositionZ] = *p.Track.PositionZ
	}
	if p.Track.TrackConfidence != 0 {
		fields[observation.FieldTrackConfidence] = p.Track.TrackConfidence
	}
	if p.Track.LogicID != "" {
		fields[observation.FieldLogicID] = p.Track.LogicID
	}
	if p.Event != "" {
		fields["decision_event"] = p.Event
	}
	// AI 派生 track_verdict 与床状态无关；仅 sleepad_radar_conflict 显式传 BedStatus 才保留。
	if p.Track.BedStatus != observation.BedStatusUnchanged {
		fields[observation.FieldBedStatus] = p.Track.BedStatus
	}
	// area_type engine 自己算（observation.Track 的 AreaType 是字符串，engine 这边类型不同）
	if g != nil {
		px, py := 0, 0
		if p.Track.PositionX != nil {
			px = *p.Track.PositionX
		}
		if p.Track.PositionY != nil {
			py = *p.Track.PositionY
		}
		if cell := g.CellAt(px, py); cell != nil {
			fields["area_type"] = areaTypeProtocolStr(cell.Belief[0].Type)
		}
	}
	// PR5b: Source（一等公民）+ Reason / Evidence（审计元数据）
	// Source 默认 = e.aiSource（来自 cfg.AIPublish.Source，如 "AI.Caregiver01"）。
	// p.Source 非空时尊重 caller override（未来多角色场景，如健康风险模块发 verdict）
	source := p.Source
	if source == "" {
		source = defaultSource
	}
	if source != "" {
		fields["source"] = source
	}
	if p.Reason != "" {
		fields["reason"] = p.Reason
	}
	if len(p.Evidence) > 0 {
		fields["evidence"] = p.Evidence
	}
	if p.IncidentMs > 0 {
		fields["incident_ts_ms"] = p.IncidentMs
	}
	// 北极星 reasoning trace：verdict 必带触发它的 source envelope.seq —
	// 下游审计可一句 grep 把 AI verdict 反向链回 producer 的具体 envelope。
	if srcSeq := e.readLastSrcSeq(p.DeviceAddr); srcSeq != 0 {
		fields["trigger_seq_num"] = srcSeq
	}

	willPublish := e.redisClient != nil && mode == "log&publish"

	// 预先取 producer + seq，让 ai_emit log 带 trace_id（"<producer>.<seqN>"），
	// 跨服务 grep trace_id 即可 join sensor→cardagg→data 整条链。
	producer := source
	if producer == "" {
		producer = defaultSource
	}
	seq := e.nextAgentSeq(ctx, producer)
	traceID := fmt.Sprintf("%s.%d", producer, seq)

	// 任何模式都打 ai_emit 审计日志：sandbox 演示靠这条 log 看 AI 在思考
	e.logger.Info("ai_emit",
		zap.String("trace_id", traceID),
		zap.String("source", source),
		zap.String("mode", mode),
		zap.String("device_type", deviceType),
		zap.String("device_addr", p.DeviceAddr),
		zap.String("device_uid_hex", e.DeviceUIDHex(p.DeviceAddr)),
		zap.String("category", category),
		zap.String("topic_type", topicType),
		zap.String("would_publish_to", streamName),
		zap.Bool("published", willPublish),
		zap.Int("track_id", p.Track.TrackID),
		zap.Int("track_confidence", p.Track.TrackConfidence),
		zap.Int64("ts_ms", nowMs),
	)

	if !willPublish {
		return
	}

	// Producer = sensor agent /128 INET；
	// SubjectEntity 留空（cardagg LPM 反查兜底，见上注释）；
	// DeviceAddr = p.DeviceAddr parse 后 /128。
	msg := rediscommon.IoTStreamMessage{
		Producer:       producer,
		SequenceNumber: uint64(seq),
		SubjectEntity:  "",
		DeviceAddr:     addr,
		DeviceType:     deviceType,
		Timestamp:      nowMs,
		TopicType:      topicType,
		Category:       category,
		DataValue:      []interface{}{fields},
	}
	if _, err := rediscommon.PublishToStream(ctx, e.redisClient, streamName, msg.ToStreamMap(), maxLen, retentionSec); err != nil {
		e.logger.Warn("ai_publish_failed",
			zap.String("source", source),
			zap.String("stream", streamName),
			zap.String("category", category),
			zap.String("device_addr", p.DeviceAddr),
			zap.Error(err),
		)
	}
}

// areaTypeProtocolStr cell 区域 → 固件 declare_area 协议字符串（下发映射）。
func areaTypeProtocolStr(t AreaType) string {
	switch t {
	case AreaBed, AreaMonitorBed:
		return "bed" // 固件 Bed(2)，中心在雷达下方自动升 monitor_bed
	case AreaEnter:
		return "door"
	case AreaReflector, AreaInterfer:
		return "interfer" // 固件 masking(3)：反射/运动干扰都屏蔽
	case AreaDeny:
		return "custom" // 固件 custom(1)：家具装饰，雷达不处理
	}
	return "none" // sit/lying/active/unknown 不下发
}

func (e *Engine) hydrateRoom(roomID string, grid *RoomGrid, expectedHash string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	storedHash, payload, found, err := e.persister.Load(ctx, roomID)
	if err != nil {
		e.logger.Warn("snapshot load failed", zap.String("room_id", roomID), zap.Error(err))
		return
	}
	if !found {
		e.logger.Info("no snapshot found, fresh start", zap.String("room_id", roomID))
		return
	}

	snap, err := UnmarshalSnapshot(payload)
	if err != nil {
		e.logger.Warn("snapshot unmarshal failed", zap.String("room_id", roomID), zap.Error(err))
		return
	}
	// layout_hash 仅作观测：不一致=layout 内容改过，但只要 grid extent(W/H/OX/OY)未变，cell index↔物理
	// 稳定 → 仍可按 index 复用 learned cell（SourceHuman 在 DecodeSnapshot 内被跳过，人标恒在上）。
	// extent 变了（layout 编辑触发 ApplyOptimizedExtent 缩放）→ DecodeSnapshot 报 dimension mismatch → 冷启。
	if storedHash != expectedHash {
		e.logger.Info("snapshot layout_hash differs (layout edited); reuse learned cells if grid extent unchanged",
			zap.String("room_id", roomID))
	}
	if err := DecodeSnapshot(snap, grid); err != nil {
		e.logger.Info("snapshot not reusable, fresh start (likely grid resized by layout edit)",
			zap.String("room_id", roomID), zap.Error(err))
		return
	}
	e.logger.Info("snapshot hydrated",
		zap.String("room_id", roomID),
		zap.Int("cells", len(snap.Cells)),
	)
}

func (e *Engine) decayLoop(ctx context.Context) {
	ticker := time.NewTicker(e.decayInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.mu.RLock()
			dtSec := e.decayInterval.Seconds()
			dp := e.decayParams
			for _, grid := range e.grids {
				grid.DecayAll(dtSec, dp)
			}
			e.mu.RUnlock()
			e.logger.Debug("decay all rooms done", zap.Float64("dt_sec", dtSec))
		}
	}
}

func (e *Engine) beliefScanLoop(ctx context.Context) {
	ticker := time.NewTicker(e.beliefScanInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.scanBeliefAll()
		}
	}
}

func (e *Engine) scanBeliefAll() {
	e.mu.RLock()
	defer e.mu.RUnlock()
	totalCells := 0
	totalLieAnomalies := 0
	lp := e.learnParams
	nowMs := time.Now().UnixMilli()
	for _, grid := range e.grids {
		grid.LearnCellAreas(lp, nowMs)
		totalLieAnomalies += grid.LearnLyingAnomalies(lp)
		for i := range grid.Cells {
			for g := 0; g < 3; g++ {
				grid.Cells[i].UpdateBelief(g, e.paramSets[g])
			}
		}
		totalCells += len(grid.Cells)
	}
	e.logger.Debug("belief scan done",
		zap.Int("total_cells", totalCells),
		zap.Int("lie_anomalies", totalLieAnomalies),
	)
}

func (e *Engine) alarmFeedbackLoop(ctx context.Context) {
	if e.feedbackIngester == nil {
		return
	}
	// 启动延迟 5s，避免与 RegisterRoom / hydrate 同时进 DB
	select {
	case <-ctx.Done():
		return
	case <-time.After(5 * time.Second):
	}
	if n, err := e.feedbackIngester.IngestOnce(ctx); err != nil {
		e.logger.Warn("alarm_feedback ingest at startup", zap.Error(err))
	} else {
		e.logger.Info("alarm_feedback startup ingest", zap.Int("processed", n))
	}
	ticker := time.NewTicker(e.feedbackInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if n, err := e.feedbackIngester.IngestOnce(ctx); err != nil {
				e.logger.Warn("alarm_feedback ingest tick", zap.Error(err))
			} else if n > 0 {
				e.logger.Info("alarm_feedback ingest tick", zap.Int("processed", n))
			}
		}
	}
}

func (e *Engine) snapshotLoop(ctx context.Context) {
	ticker := time.NewTicker(e.snapshotInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.saveAllRooms(ctx)
		}
	}
}

func (e *Engine) saveAllRooms(ctx context.Context) {
	if e.persister == nil {
		return
	}
	e.mu.RLock()
	// 拷贝 (roomID, grid, hash) 三元组到本地切片，释放锁后再做 IO，避免锁内 DB 写
	type roomDump struct {
		id   string
		grid *RoomGrid
		hash string
	}
	dumps := make([]roomDump, 0, len(e.grids))
	for id, g := range e.grids {
		dumps = append(dumps, roomDump{id: id, grid: g, hash: e.layoutHashes[id]})
	}
	e.mu.RUnlock()

	saved, failed := 0, 0
	for _, d := range dumps {
		snap := EncodeSnapshot(d.grid)
		payload, cellCount, err := MarshalSnapshot(snap)
		if err != nil {
			e.logger.Warn("snapshot marshal failed",
				zap.String("room_id", d.id), zap.Error(err))
			failed++
			continue
		}
		if err := e.persister.Save(ctx, d.id, d.hash, cellCount, payload); err != nil {
			e.logger.Warn("snapshot save failed",
				zap.String("room_id", d.id), zap.Error(err))
			failed++
			continue
		}
		saved++
	}
	e.logger.Debug("snapshot batch done",
		zap.Int("saved", saved), zap.Int("failed", failed))
}

func (e *Engine) dailySnapshotLoop(ctx context.Context) {
	for {
		next := nextDailyTriggerHM(time.Now(), e.dailySnapshotHour, e.dailySnapshotMinute)
		wait := time.Until(next)
		e.logger.Info("daily_snapshot_scheduled",
			zap.Time("next", next),
			zap.Duration("wait", wait),
			zap.Int("hour_local", e.dailySnapshotHour),
			zap.Int("minute_local", e.dailySnapshotMinute),
			zap.Int("retain_days", e.historyRetainDays),
		)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		e.saveAllRoomsHistory(ctx, time.Now())
	}
}

func (e *Engine) saveAllRoomsHistory(ctx context.Context, nowLocal time.Time) {
	if e.historyPersister == nil {
		return
	}
	e.mu.RLock()
	type roomDump struct {
		id   string
		grid *RoomGrid
		hash string
	}
	dumps := make([]roomDump, 0, len(e.grids))
	for id, g := range e.grids {
		dumps = append(dumps, roomDump{id: id, grid: g, hash: e.layoutHashes[id]})
	}
	e.mu.RUnlock()

	saved, failed := 0, 0
	for _, d := range dumps {
		snap := EncodeSnapshot(d.grid)
		payload, cellCount, err := MarshalSnapshot(snap)
		if err != nil {
			e.logger.Warn("daily snapshot marshal failed",
				zap.String("room_id", d.id), zap.Error(err))
			failed++
			continue
		}
		if err := e.historyPersister.SaveDaily(ctx, d.id, d.hash, nowLocal, cellCount, payload, e.historyRetainDays); err != nil {
			e.logger.Warn("daily snapshot save failed",
				zap.String("room_id", d.id), zap.Error(err))
			failed++
			continue
		}
		saved++
	}
	e.logger.Info("daily_snapshot_done",
		zap.Int("saved", saved),
		zap.Int("failed", failed),
		zap.String("snapshot_date", nowLocal.Format("2006-01-02")),
	)
}

func nextDailyTriggerHM(now time.Time, hour, minute int) time.Time {
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !t.After(now) {
		t = t.Add(24 * time.Hour)
	}
	return t
}

func (e *Engine) dailyLayoutReloadLoop(ctx context.Context) {
	for {
		next := nextDailyTrigger(time.Now(), e.dailyReloadHour)
		wait := time.Until(next)
		e.logger.Info("daily_layout_reload_scheduled",
			zap.Time("next", next),
			zap.Duration("wait", wait),
			zap.Int("hour_local", e.dailyReloadHour),
		)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		e.runDailyLayoutReload(ctx)
	}
}

func nextDailyTrigger(now time.Time, hour int) time.Time {
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if !t.After(now) {
		t = t.Add(24 * time.Hour)
	}
	return t
}

func (e *Engine) runDailyLayoutReload(ctx context.Context) {
	if e.dailyReloadDB == nil {
		return
	}
	// v2 schema: layout 在 room_visual_layout 表（PK=spatial_prefix）；
	// rooms 表无 tenant_id/unit_id 列；tenant_id 由 room_id INET prefix /48 派生；
	// unit timezone 通过 unit /80 LPM contains room /88 取。
	// 按 /128 device 加载 canvas、按 /88 room 聚合（与 registerAllRooms 共用 LoadRoomCanvases
	// + BuildRoomConfigFromCanvases，避免双写漂移）。room 元数据单独查。
	canvasesByRoom, err := LoadRoomCanvases(ctx, e.dailyReloadDB, e.logger)
	if err != nil {
		e.logger.Warn("daily_reload load canvases failed", zap.Error(err))
		return
	}
	rows, err := e.dailyReloadDB.QueryContext(ctx, `
		SELECT r.room_id::text,
		       r.room_name,
		       COALESCE(u.timezone, '') AS timezone,
		       host(set_masklen(r.room_id, 48))::text || '/48' AS tenant_pref
		FROM rooms r
		LEFT JOIN units u ON u.unit_id >>= r.room_id`)
	if err != nil {
		e.logger.Warn("daily_reload query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	type pendingReload struct {
		cfg      RoomConfig
		tenantID string
		newHash  string
	}
	var pending []pendingReload
	for rows.Next() {
		var roomID, roomName, tenantID, timezone string
		if err := rows.Scan(&roomID, &roomName, &timezone, &tenantID); err != nil {
			continue
		}
		cfg, hasLayout := BuildRoomConfigFromCanvases(roomID, canvasesByRoom[roomID], e.logger)
		if !hasLayout {
			continue
		}
		cfg.RoomName = roomName
		cfg.Timezone = timezone
		ApplyOptimizedExtent(&cfg)
		newHash := LayoutHash(cfg)
		e.mu.RLock()
		oldHash := e.layoutHashes[roomID]
		e.mu.RUnlock()
		if oldHash == newHash {
			continue // 没变
		}
		pending = append(pending, pendingReload{cfg: cfg, tenantID: tenantID, newHash: newHash})
	}
	for _, pr := range pending {
		e.logger.Info("daily_reload room changed, resetting",
			zap.String("room_id", pr.cfg.RoomID),
			zap.String("new_hash", pr.newHash[:12]),
		)
		// RegisterRoom 替换 TrackManager + grid，hydrateRoom 见 hash 不同会丢弃旧 snapshot
		e.RegisterRoom(pr.cfg)
		e.SetRoomTenant(pr.cfg.RoomID, pr.tenantID)
	}
	if len(pending) > 0 && e.persister != nil {
		// 立刻持久化新状态，覆盖旧 snapshot
		e.saveAllRooms(ctx)
	}
}
