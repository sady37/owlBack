package httpapi

// TenantStatusCache 在 auth_middleware 中缓存 tenant.status，避免每个请求都查 DB。
//
// D-001b：suspend tenant 后已发的 access token 仍然有效（最长 24h）；middleware 必须每次请求
// 验 tenant.status 才能立刻拒绝该 tenant 用户。如果每次都查 DB，成本太高（admin 操作高频时
// 几十次/秒）。这个 cache 用 5min TTL 在内存层挡住绝大多数查询。
//
// Cache invalidation：admin_tenants_handlers 在改 tenant.status 后调 Invalidate(prefix) 主动失效。
//
// 多实例部署时本 cache 不跨实例同步；单实例 5min 内陈旧 status 是可接受的窗口（攻击场景不存在
// 因为 D-001a 已经把"新 login"路径关掉了，剩下的只是已登录 admin 在 cache 失效前的≤5min 残留）。

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

const tenantStatusCacheTTL = 5 * time.Minute

type tenantStatusEntry struct {
	status string
	expAt  time.Time
}

type TenantStatusCache struct {
	mu      sync.RWMutex
	entries map[string]tenantStatusEntry
	db      *sql.DB
}

// NewTenantStatusCache 创建 cache。db 用于 miss 时回源查 tenants.status。
func NewTenantStatusCache(db *sql.DB) *TenantStatusCache {
	return &TenantStatusCache{
		entries: make(map[string]tenantStatusEntry),
		db:      db,
	}
}

// Get 取 tenant 的 status，命中返回 cached 值；miss 查 DB 并写 cache。
// 对于查不到的 tenant_id 返回 'unknown'，由调用方决定是否拒绝。
func (c *TenantStatusCache) Get(ctx context.Context, tenantPrefix string) string {
	if tenantPrefix == "" {
		return "active" // 无 tenant_id（platform admin）放行
	}
	now := time.Now()
	c.mu.RLock()
	if e, ok := c.entries[tenantPrefix]; ok && now.Before(e.expAt) {
		c.mu.RUnlock()
		return e.status
	}
	c.mu.RUnlock()

	// miss → 回源
	var status string
	err := c.db.QueryRowContext(ctx,
		`SELECT COALESCE(status, 'active') FROM tenants WHERE tenant_id = $1::INET`,
		tenantPrefix,
	).Scan(&status)
	if err != nil {
		// 查不到（被 hard delete 的空 tenant 也走这里）→ unknown，由调用方拒绝
		status = "unknown"
	}
	c.mu.Lock()
	c.entries[tenantPrefix] = tenantStatusEntry{status: status, expAt: now.Add(tenantStatusCacheTTL)}
	c.mu.Unlock()
	return status
}

// Invalidate 显式作废一个 tenant 的 cache 条目。
// admin 改 tenant.status 后必须调用，否则其他实例 / 当前实例 cache 在 5min 内仍返回旧值。
func (c *TenantStatusCache) Invalidate(tenantPrefix string) {
	if tenantPrefix == "" {
		return
	}
	c.mu.Lock()
	delete(c.entries, tenantPrefix)
	c.mu.Unlock()
}

// InvalidateAll 清空所有 cache 条目。批量 reseed / 维护时用。
func (c *TenantStatusCache) InvalidateAll() {
	c.mu.Lock()
	c.entries = make(map[string]tenantStatusEntry)
	c.mu.Unlock()
}
