// Package ipam abstracts spatial prefix allocation in the owl_v2 IPv6 hierarchy.
//
// 设计原则：把 IPAM 当作可替换的 backend，业务层（CreateBranch / CreateUnit / ...）
// 调 Backend.AllocateXxx，不知道底层是 PG / NetBox / kea。
//
// 默认实现 = PGBackend（直接在 owl_v2 spatial 表上做 advisory lock + slot 分配）。
//
// 可替换实现（占位，未来真要切再实现）：
//   - NetBoxBackend：调 NetBox REST API 拿 prefix，再写本地 owl_v2 表
//   - KeaBackend：通过 kea ctrl-agent lease6-add 注入 lease（commercial subnet_cmds 时也走这里）
package ipam

import (
	"context"
	"net/netip"
)

// Backend abstracts spatial prefix/address allocation.
//
// 6 个 Allocate* 方法对应 6 个 spatial 层级；每个方法都做 advisory lock + slot 分配 + INSERT
// 业务表，外加 derive 出标准 IPv6 prefix（owl-common/spatial 派生函数）。
//
// RegisterDevice 是特例：device /128 完全 stateless 派生 (= bed_prefix + hash(device_uid))，
// 不需要 slot 分配，仅做 INSERT。
type Backend interface {
	AllocateBranch(ctx context.Context, tenant netip.Prefix, attrs BranchAttrs) (netip.Prefix, error)
	AllocateSite(ctx context.Context, branch netip.Prefix, building, floor uint8, attrs SiteAttrs) (netip.Prefix, error)
	AllocateUnit(ctx context.Context, site netip.Prefix, attrs UnitAttrs) (netip.Prefix, error)
	AllocateRoom(ctx context.Context, unit netip.Prefix, attrs RoomAttrs) (netip.Prefix, error)
	AllocateBed(ctx context.Context, room netip.Prefix, attrs BedAttrs) (netip.Prefix, error)

	// RegisterDevice 派生 device /128 spatial_addr (= base + hash(device_uid))
	// 并写入 owl_v2.devices 表。base 通常是 bed /96，少数公共区设备用 unit /80 或 room /88。
	// device_id 必须先在 device_factory_meta 中已存在（device 出厂导入时入库）。
	RegisterDevice(ctx context.Context, base netip.Prefix, deviceID, deviceUID string) (netip.Addr, error)

	// LookupTenant 按 prefix 查 tenant；找不到返回 ErrNotFound。
	LookupTenant(ctx context.Context, tenant netip.Prefix) (*TenantInfo, error)
}
