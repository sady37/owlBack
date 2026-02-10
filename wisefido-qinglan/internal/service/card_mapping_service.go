package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	commonRedis "owl-common/redis"
	"wisefido-qinglan/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// CardMappingService 管理 deviceUID → 完整映射信息的查询服务
// 订阅 config:card:stream 中的 config.card 事件
// 维护以下 Redis 结构：
//  1. owl:cache:branch:{branchID} → Hash
//     Field: deviceUID (MQTT 标准字段)
//     Value: CardDeviceInfo 的 JSON 序列化
//  2. owl:lookup:device:{deviceUID} → String
//     Value: branchID（用于第一步快速查询）
//
// 查询流程（两步）：
// Step 1: GET owl:lookup:device:{deviceUID} → 获取 branchID
// Step 2: HGET owl:cache:branch:{branchID} {deviceUID} → 获取完整的 CardDeviceInfo
type CardMappingService struct {
	redisClient *redis.Client
	cardRepo    repository.CardRepository
	logger      *zap.Logger
}

// NewCardMappingService 构造 CardMappingService
func NewCardMappingService(redisClient *redis.Client, cardRepo repository.CardRepository, logger *zap.Logger) *CardMappingService {
	return &CardMappingService{
		redisClient: redisClient,
		cardRepo:    cardRepo,
		logger:      logger,
	}
}

// ClearBranchCache 彻底清理该分支的所有缓存记录
func (s *CardMappingService) ClearBranchCache(ctx context.Context, branchID string) error {
	branchKey := fmt.Sprintf("owl:cache:branch:%s", branchID)

	s.logger.Info("Clearing branch cache",
		zap.String("branch_id", branchID),
		zap.String("branch_key", branchKey))

	// 1. 找出该分支下目前关联的所有设备 UID，用于清理全局 lookup 索引
	oldDeviceUIDs, err := s.redisClient.HKeys(ctx, branchKey).Result()
	if err != nil && err != redis.Nil {
		s.logger.Warn("Failed to fetch old device UIDs during clear", zap.Error(err))
	}

	pipe := s.redisClient.Pipeline()

	// 2. 删除分支 Hash 本体
	pipe.Del(ctx, branchKey)

	// 3. 批量删除全局 Lookup 索引
	for _, uid := range oldDeviceUIDs {
		pipe.Del(ctx, fmt.Sprintf("owl:lookup:device:%s", uid))
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute clear pipeline: %w", err)
	}

	return nil
}

// updateBranchCacheInRedis 负责将分组后的数据批量写入 Redis
func (s *CardMappingService) updateBranchCacheInRedis(ctx context.Context, branchID string, mappingRecords []repository.CardDeviceInfo) error {
	if len(mappingRecords) == 0 {
		return nil
	}

	branchKey := fmt.Sprintf("owl:cache:branch:%s", branchID)

	// Log summary before writing to Redis
	s.logger.Info("Updating branch cache in Redis",
		zap.String("branch_id", branchID),
		zap.Int("records", len(mappingRecords)),
		zap.String("branch_key", branchKey))

	pipe := s.redisClient.Pipeline()

	for _, info := range mappingRecords {
		jsonData, _ := json.Marshal(info)

		// debug log each mapping about to be written
		s.logger.Debug("Preparing to write mapping to Redis",
			zap.String("branch_id", branchID),
			zap.String("device_uid", info.DeviceUID),
			zap.String("card_id", info.CardID))

		// 写入分支 Hash
		pipe.HSet(ctx, branchKey, info.DeviceUID, jsonData)

		// 写入全局 Lookup 索引 (24h TTL)
		lookupKey := fmt.Sprintf("owl:lookup:device:%s", info.DeviceUID)
		pipe.Set(ctx, lookupKey, branchID, 24*time.Hour)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		s.logger.Warn("Failed to execute Redis pipeline for branch cache",
			zap.String("branch_id", branchID),
			zap.Error(err))
	} else {
		s.logger.Info("Branch cache updated in Redis",
			zap.String("branch_id", branchID),
			zap.Int("written", len(mappingRecords)))
	}

	if err != nil {
		return err
	}

	// Verification: read back and ensure mappings and lookup exist and match
	var failed []repository.CardDeviceInfo
	for _, info := range mappingRecords {
		// verify HGET
		jsonValue, gerr := s.redisClient.HGet(ctx, branchKey, info.DeviceUID).Result()
		if gerr != nil && gerr != redis.Nil {
			s.logger.Warn("Verification HGET failed",
				zap.String("branch_id", branchID),
				zap.String("device_uid", info.DeviceUID),
				zap.Error(gerr))
			failed = append(failed, info)
			continue
		}
		if jsonValue == "" {
			s.logger.Warn("Verification: branch hash missing field",
				zap.String("branch_id", branchID),
				zap.String("device_uid", info.DeviceUID))
			failed = append(failed, info)
			continue
		}
		var cdi repository.CardDeviceInfo
		if uerr := json.Unmarshal([]byte(jsonValue), &cdi); uerr != nil {
			s.logger.Warn("Verification: failed to unmarshal stored CardDeviceInfo",
				zap.String("device_uid", info.DeviceUID), zap.Error(uerr))
			failed = append(failed, info)
			continue
		}
		if cdi.CardID != info.CardID || cdi.DeviceUID != info.DeviceUID {
			s.logger.Warn("Verification: stored CardDeviceInfo mismatch",
				zap.String("device_uid", info.DeviceUID),
				zap.String("expected_card_id", info.CardID),
				zap.String("stored_card_id", cdi.CardID))
			failed = append(failed, info)
			continue
		}

		// verify lookup key
		lookupKey := fmt.Sprintf("owl:lookup:device:%s", info.DeviceUID)
		lk, lerr := s.redisClient.Get(ctx, lookupKey).Result()
		if lerr != nil && lerr != redis.Nil {
			s.logger.Warn("Verification GET failed",
				zap.String("lookup_key", lookupKey),
				zap.Error(lerr))
			failed = append(failed, info)
			continue
		}
		if lk != branchID {
			s.logger.Warn("Verification: lookup value mismatch",
				zap.String("lookup_key", lookupKey),
				zap.String("expected_branch", branchID),
				zap.String("stored_branch", lk))
			failed = append(failed, info)
			continue
		}
	}

	if len(failed) > 0 {
		s.logger.Info("Retrying failed mappings after verification",
			zap.String("branch_id", branchID), zap.Int("failed_count", len(failed)))
		pipe2 := s.redisClient.Pipeline()
		for _, info := range failed {
			jsonData, _ := json.Marshal(info)
			pipe2.HSet(ctx, branchKey, info.DeviceUID, jsonData)
			lookupKey := fmt.Sprintf("owl:lookup:device:%s", info.DeviceUID)
			pipe2.Set(ctx, lookupKey, branchID, 24*time.Hour)
		}
		_, err2 := pipe2.Exec(ctx)
		if err2 != nil {
			s.logger.Error("Retry pipeline failed",
				zap.String("branch_id", branchID), zap.Error(err2))
			return err2
		}
		s.logger.Info("Retry succeeded for failed mappings",
			zap.String("branch_id", branchID), zap.Int("recovered", len(failed)))
	}

	return nil
}

// InitializeAllBranchCaches 启动时初始化所有分支的卡片映射缓存
func (s *CardMappingService) InitializeAllBranchCaches(ctx context.Context) error {
	s.logger.Info("Initializing card mapping caches for all branches at startup...")

	// 1. 获取全量映射关系（一次性查出所有分支的所有设备）
	deviceMappings, err := s.cardRepo.GetDeviceCardMappings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get device list: %w", err)
	}

	// 2. 在内存中按 branchID 分组
	// key: branchID, value: 设备信息列表
	branchGroups := make(map[string][]repository.CardDeviceInfo)
	for _, mapping := range deviceMappings {
		info := repository.CardDeviceInfo{
			DeviceUID: mapping.DeviceUID,
			TenantID:  mapping.TenantID,
			BranchID:  mapping.BranchID,
			DeviceID:  mapping.DeviceID,
			CardID:    mapping.CardID,
		}
		branchGroups[mapping.BranchID] = append(branchGroups[mapping.BranchID], info)
	}

	s.logger.Info("Grouping completed", zap.Int("branch_count", len(branchGroups)))

	successCount := 0
	failureCount := 0

	// 3. 遍历分组，执行 清理 + 写入
	for bID, records := range branchGroups {
		// 3.1 清理旧缓存（确保不会留下脏的 lookup 索引）
		if err := s.ClearBranchCache(ctx, bID); err != nil {
			s.logger.Warn("Failed to clear cache for branch", zap.String("branch_id", bID), zap.Error(err))
		}

		// 3.2 写入新缓存
		if err := s.updateBranchCacheInRedis(ctx, bID, records); err != nil {
			s.logger.Error("Failed to update cache for branch", zap.String("branch_id", bID), zap.Error(err))
			failureCount++
		} else {
			successCount++
		}
	}

	s.logger.Info("Initialization completed",
		zap.Int("success_branches", successCount),
		zap.Int("failure_branches", failureCount))

	return nil
}

// GetCardIDByDeviceUID 通过 deviceUID 查询 CardDeviceInfo（先查 lookup，然后查分支 hash）
func (s *CardMappingService) GetCardIDByDeviceUID(ctx context.Context, deviceUID string) (*repository.CardDeviceInfo, error) {
	// Step 1: lookup branch
	lookupKey := fmt.Sprintf("owl:lookup:device:%s", deviceUID)
	branchID, err := s.redisClient.Get(ctx, lookupKey).Result()
	if err != nil && err != redis.Nil {
		s.logger.Warn("Failed to GET lookup key", zap.String("lookup_key", lookupKey), zap.Error(err))
		return nil, fmt.Errorf("failed to query lookup key: %w", err)
	}
	if branchID == "" {
		return nil, fmt.Errorf("device_uid not found in cache: %s", deviceUID)
	}

	// Step 2: HGET branch hash
	cacheKey := fmt.Sprintf("owl:cache:branch:%s", branchID)
	jsonValue, err := s.redisClient.HGet(ctx, cacheKey, deviceUID).Result()
	if err != nil && err != redis.Nil {
		s.logger.Warn("Failed to HGET branch cache", zap.String("branch_key", cacheKey), zap.String("device_uid", deviceUID), zap.Error(err))
		return nil, fmt.Errorf("failed to query branch cache: %w", err)
	}
	if jsonValue == "" {
		return nil, fmt.Errorf("device_uid not found in branch cache: %s", deviceUID)
	}

	var cdi repository.CardDeviceInfo
	if err := json.Unmarshal([]byte(jsonValue), &cdi); err != nil {
		s.logger.Warn("Failed to unmarshal CardDeviceInfo", zap.String("device_uid", deviceUID), zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal CardDeviceInfo: %w", err)
	}
	return &cdi, nil
}

func (s *CardMappingService) HandleCardChanged(ctx context.Context, data map[string]interface{}) error {
	s.logger.Info("HandleCardChanged invoked")
	cardData := &commonRedis.CardChangeData{}
	if err := s.parseCardChangeData(data, cardData); err != nil {
		return err
	}

	// 1. 清理受影响的分支
	if err := s.ClearBranchCache(ctx, cardData.BranchID); err != nil {
		return err
	}

	// 2. 精准查询该分支最新数据
	mappingRecords, err := s.cardRepo.GetDeviceCardMappingsByBranch(ctx, cardData.TenantID, cardData.BranchID)
	if err != nil {
		return err
	}

	// Log DB return for diagnostics
	s.logger.Info("Fetched branch card mappings from DB",
		zap.String("tenant_id", cardData.TenantID),
		zap.String("branch_id", cardData.BranchID),
		zap.Int("mappings_count", len(mappingRecords)))

	// 3. 转换为切片并写入
	var list []repository.CardDeviceInfo
	for _, v := range mappingRecords {
		list = append(list, v)
	}

	return s.updateBranchCacheInRedis(ctx, cardData.BranchID, list)
}

// parseCardChangeData 解析卡片变更数据
func (s *CardMappingService) parseCardChangeData(data map[string]interface{}, cardData *commonRedis.CardChangeData) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal card change data: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, cardData); err != nil {
		return fmt.Errorf("failed to unmarshal card change data: %w", err)
	}

	if cardData.TenantID == "" || cardData.CardID == "" {
		return fmt.Errorf("tenant_id and card_id are required in card change data")
	}

	return nil
}
