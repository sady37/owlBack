package ipam

import (
	"context"
	"errors"
	"net/netip"
)

// NetBoxBackend 是未来切换到 NetBox 作 IPAM source-of-truth 的占位实现。
//
// 触发点（什么时候真做这个）：
//   - owl 出现 SaaS multi-customer，需要给客户提供 IPAM dashboard / API
//   - 与外部网络团队对接（必须用业界标准 IPAM 协议）
//
// 当前是占位（panic）；保留是为了让业务代码（CreateBranch/Unit/...）能 import
// Backend interface 而非具体实现，未来真接 NetBox 时只换 wire 配置不改业务。
type NetBoxBackend struct {
	APIURL    string
	APIToken  string
	// PG mirror — 即使 NetBox 是 source-of-truth，owl_v2 表仍要本地 mirror（业务字段 / 报警 join 用）
	// PG *sql.DB
}

var ErrNotImplemented = errors.New("ipam: NetBoxBackend not implemented yet — call when SaaS multi-tenant scope arrives")

func (n *NetBoxBackend) AllocateBranch(ctx context.Context, tenant netip.Prefix, attrs BranchAttrs) (netip.Prefix, error) {
	return netip.Prefix{}, ErrNotImplemented
}
func (n *NetBoxBackend) AllocateSite(ctx context.Context, branch netip.Prefix, building, floor uint8, attrs SiteAttrs) (netip.Prefix, error) {
	return netip.Prefix{}, ErrNotImplemented
}
func (n *NetBoxBackend) AllocateUnit(ctx context.Context, site netip.Prefix, attrs UnitAttrs) (netip.Prefix, error) {
	return netip.Prefix{}, ErrNotImplemented
}
func (n *NetBoxBackend) AllocateRoom(ctx context.Context, unit netip.Prefix, attrs RoomAttrs) (netip.Prefix, error) {
	return netip.Prefix{}, ErrNotImplemented
}
func (n *NetBoxBackend) AllocateBed(ctx context.Context, room netip.Prefix, attrs BedAttrs) (netip.Prefix, error) {
	return netip.Prefix{}, ErrNotImplemented
}
func (n *NetBoxBackend) RegisterDevice(ctx context.Context, base netip.Prefix, deviceID, deviceUID string) (netip.Addr, error) {
	return netip.Addr{}, ErrNotImplemented
}
func (n *NetBoxBackend) LookupTenant(ctx context.Context, tenant netip.Prefix) (*TenantInfo, error) {
	return nil, ErrNotImplemented
}
