package service

import (
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// CardCreateService — v2 stub。
//
// v1 范式：CreateCardsForUnit 按 unit 算 expected cards → diff → batch CRUD。
// v2 范式：事件驱动 per-prefix reconcile —— 不再做批量计算。
//
// 真实落地的 v2 路径：
//   - resident 入住 / 出院 / 转床 → resident_service 写 resident_unit episode + 直接
//     INSERT/UPDATE/DELETE cards 行 + DDNS register/unregister + emit CloudEvents
//   - device bind / unbind → device service 在 IP 分配后由 ipam 触发同样的卡 reconcile
//
// 本文件保留以兼容 CardSyncService 内部依赖 + main.go 启动期 reconcile job 的 stub
// 调用；返回零 stats。真正的 v2 实现见 [doc/cards_v2_migration_checklist.md § Phase F]。

// CardUpdateStats v2 兼容签名（card-common 已删原类型）。零字段表示 no-op。
type CardUpdateStats struct {
	ExistingCount  int
	DeletedCount   int
	CreatedCount   int
	UpdatedCount   int
	UnchangedCount int
}

// CardCreateService stub
type CardCreateService struct {
	repo   repository.RepositoryInterface
	logger *zap.Logger
}

// NewCardCreateService 创建卡片创建服务
func NewCardCreateService(repo repository.RepositoryInterface, logger *zap.Logger) *CardCreateService {
	return &CardCreateService{repo: repo, logger: logger}
}

// CreateCardsForUnit v2 no-op。返回零 stats 让 caller 不报错。
//
// Phase F 待做：resident lifecycle hook 入住 → 直接 INSERT card 行 + DDNS register。
func (c *CardCreateService) CreateCardsForUnit(tenantID, unitID string) (*CardUpdateStats, error) {
	c.logger.Debug("CardCreateService.CreateCardsForUnit skipped (v2 event-driven; pending Phase F)",
		zap.String("tenant_id", tenantID), zap.String("unit_id", unitID))
	return &CardUpdateStats{}, nil
}
