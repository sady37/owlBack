package service

import (
	"context"
	"sync"
	"time"

	"owl-common/card"
	"owl-common/observation"

	"go.uber.org/zap"
)

// StateService sensor 侧最小 state writer。
// PR1 阶段只承载 ReadCardStatus + UpdateTargetPose + UpdateTargetLastActive 三个 A 组方法；
// B/C 组迁移再补 Bed/Room/Bathroom/Derive 各路径。
type StateService struct {
	writer *card.Writer
	reader *card.Reader
	logger *zap.Logger

	preparedMu    sync.Mutex
	preparedCards map[string]struct{}
}

func NewStateService(writer *card.Writer, reader *card.Reader, logger *zap.Logger) *StateService {
	return &StateService{
		writer:        writer,
		reader:        reader,
		logger:        logger,
		preparedCards: make(map[string]struct{}),
	}
}

// ReadCardStatus 读取当前卡片状态（Targets/BedState 等），供 event 层防静电等逻辑使用。
func (s *StateService) ReadCardStatus(ctx context.Context, cardID string) (*card.CardStatus, error) {
	if s.reader == nil {
		return nil, nil
	}
	return s.reader.ReadCardStatus(ctx, cardID)
}

// UpdateTargetPose 按事件更新 CardStatus.Target（单 Target）的 Pose、PoseSince 等，读当前后合并写回。
func (s *StateService) UpdateTargetPose(ctx context.Context, cardID string, trackID, pose int, poseSinceMs int64) error {
	if s.writer == nil || s.reader == nil || cardID == "" || trackID <= 0 || trackID == 88 {
		return nil
	}
	curr, err := s.reader.ReadCardStatus(ctx, cardID)
	if err != nil || curr == nil {
		return nil
	}
	ts := curr.Target
	if ts == nil {
		ts = &card.TargetState{TrackID: trackID, UpdatedAt: poseSinceMs}
	}
	ts.UpdatedAt = time.Now().UnixMilli()
	return PublishCardStatus(ctx, s.writer, cardID, PublishFields{Target: ts})
}

// UpdateTargetLastActive 当 moveSec >= moveSecThreshold 或 danceMin >= danceMinThreshold 时更新 Target.LastActiveTs。
func (s *StateService) UpdateTargetLastActive(ctx context.Context, cardID string, tsMs int64, moveSec, moveSecThreshold int, danceMin, danceMinThreshold int) error {
	if s.writer == nil || s.reader == nil || cardID == "" {
		return nil
	}
	if moveSecThreshold <= 0 && danceMinThreshold <= 0 {
		return nil
	}
	if moveSec < moveSecThreshold && (danceMinThreshold <= 0 || danceMin < danceMinThreshold) {
		return nil
	}
	curr, err := s.reader.ReadCardStatus(ctx, cardID)
	if err != nil || curr == nil {
		return nil
	}
	target := curr.Target
	if target == nil {
		target = &card.TargetState{TrackID: observation.TrackUnknownPerson, UpdatedAt: time.Now().UnixMilli()}
	} else {
		t := *target
		target = &t
	}
	target.LastActiveTs = tsMs
	target.UpdatedAt = time.Now().UnixMilli()
	return PublishCardStatus(ctx, s.writer, cardID, PublishFields{Target: target})
}
