package service

import (
	"context"
	"database/sql"
	"fmt"
	"wisefido-card-manage/internal/config"
	"wisefido-card-manage/internal/repository"

	"owl-common/card"
	"owl-common/database"

	"go.uber.org/zap"
)

// CardService 卡片管理服务
type CardService struct {
	config      *config.Config
	logger      *zap.Logger
	db          *sql.DB
	cardRepo    *repository.CardRepository
	cardCreator *card.CardCreator
}

// NewCardService 创建卡片管理服务
func NewCardService(cfg *config.Config, logger *zap.Logger) (*CardService, error) {
	// 初始化数据库
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 创建 Repository
	cardRepo := repository.NewCardRepository(db, logger)

	// 创建 CardCreator（使用 owl-common/card 包）
	cardCreator := card.NewCardCreator(cardRepo, logger)

	return &CardService{
		config:      cfg,
		logger:      logger,
		db:          db,
		cardRepo:    cardRepo,
		cardCreator: cardCreator,
	}, nil
}

// CreateCardsForUnit 为指定单元创建/更新卡片
func (s *CardService) CreateCardsForUnit(ctx context.Context, tenantID, unitID string) (*card.CardUpdateStats, error) {
	s.logger.Info("Creating cards for unit",
		zap.String("tenant_id", tenantID),
		zap.String("unit_id", unitID),
	)

	stats, err := s.cardCreator.CreateCardsForUnit(tenantID, unitID)
	if err != nil {
		s.logger.Error("Failed to create cards for unit",
			zap.String("tenant_id", tenantID),
			zap.String("unit_id", unitID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create cards for unit: %w", err)
	}

	s.logger.Info("Successfully created cards for unit",
		zap.String("tenant_id", tenantID),
		zap.String("unit_id", unitID),
		zap.Int("existing_count", stats.ExistingCount),
		zap.Int("created_count", stats.CreatedCount),
		zap.Int("updated_count", stats.UpdatedCount),
		zap.Int("deleted_count", stats.DeletedCount),
		zap.Int("unchanged_count", stats.UnchangedCount),
	)

	return stats, nil
}

// CreateAllCards 为所有单元创建/更新卡片（服务启动时或定时任务调用）
func (s *CardService) CreateAllCards(ctx context.Context) error {
	tenantID := s.config.CardManage.TenantID
	if tenantID == "" {
		s.logger.Warn("TENANT_ID is not set, skipping card creation")
		return fmt.Errorf("tenant_id is required, please set TENANT_ID environment variable")
	}

	s.logger.Info("Creating cards for all units on startup",
		zap.String("tenant_id", tenantID),
	)

	// 获取所有 unit ID
	unitIDs, err := s.cardRepo.GetAllUnits(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get all units: %w", err)
	}

	s.logger.Info("Found units for card creation",
		zap.Int("total_units", len(unitIDs)),
	)

	successCount := 0
	errorCount := 0

	for _, unitID := range unitIDs {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, err := s.cardCreator.CreateCardsForUnit(tenantID, unitID)
			if err != nil {
				s.logger.Error("Failed to create cards for unit",
					zap.String("tenant_id", tenantID),
					zap.String("unit_id", unitID),
					zap.Error(err),
				)
				errorCount++
			} else {
				successCount++
			}
		}
	}

	s.logger.Info("Completed creating cards for all units on startup",
		zap.Int("success_count", successCount),
		zap.Int("error_count", errorCount),
		zap.Int("total_units", len(unitIDs)),
	)

	return nil
}

// Close 关闭服务
func (s *CardService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
