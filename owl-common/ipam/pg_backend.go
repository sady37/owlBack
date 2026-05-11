package ipam

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"net/netip"
	"os"

	"owl-common/spatial"
)

// stderrLogger 是 kea sync 失败时 fallback log 输出；调用方可替换。
var stderrLogger = os.Stderr

// PGBackend 直接在 owl_v2 spatial 表上做 IPAM。
//
// 并发安全：每次 AllocateXxx 在事务内取 PG advisory lock（key = parent prefix 前 8 字节 hash）。
// 同 parent 下并发分配会串行；不同 parent 互不阻塞。
//
// 可选 kea sync：构造时传入 KeaClient 后，每次 Allocate 成功 commit 后会异步调
// kea lease6-add 把分配作为 IA_PD lease 写入 kea database。失败仅 log，不影响主路径。
type PGBackend struct {
	db  *sql.DB
	kea *KeaClient // 可选，nil 跳过 kea sync
}

// NewPGBackend 用现成的 *sql.DB（owl-common/database 创建的）构造。
// 不带 kea sync。
func NewPGBackend(db *sql.DB) *PGBackend {
	return &PGBackend{db: db}
}

// NewPGBackendWithKea 同 NewPGBackend，加 kea lease audit。
// kea 为 nil 时等价 NewPGBackend。
func NewPGBackendWithKea(db *sql.DB, kea *KeaClient) *PGBackend {
	return &PGBackend{db: db, kea: kea}
}

// recordKeaLease 后台记 prefix lease 到 kea；失败不阻塞主路径，仅 log warning。
func (p *PGBackend) recordKeaLease(ctx context.Context, prefix netip.Prefix, comment string) {
	if p.kea == nil {
		return
	}
	duid := EncodeDUIDForPrefix(prefix)
	// iaid 用 prefix bits 长度作传递性识别；同 DUID + 不同 iaid 是不同 lease
	iaid := uint32(prefix.Bits())
	if err := p.kea.RecordPrefixLease(ctx, prefix, duid, iaid, comment, 0); err != nil {
		fmt.Fprintf(stderrLogger, "ipam: kea lease6-add failed for %s: %v\n", prefix, err)
	}
}

func (p *PGBackend) recordKeaAddrLease(ctx context.Context, addr netip.Addr, comment string) {
	if p.kea == nil {
		return
	}
	duid := EncodeDUIDForAddr(addr)
	if err := p.kea.RecordAddressLease(ctx, addr, duid, 128, comment, 0); err != nil {
		fmt.Fprintf(stderrLogger, "ipam: kea lease6-add (address) failed for %s: %v\n", addr, err)
	}
}

// =============================================================================
// AllocateBranch — tenant /48 → branch /56
// =============================================================================

func (p *PGBackend) AllocateBranch(ctx context.Context, tenant netip.Prefix, attrs BranchAttrs) (netip.Prefix, error) {
	if tenant.Bits() != spatial.PrefixLenTenant {
		return netip.Prefix{}, fmt.Errorf("%w: AllocateBranch expects /48, got /%d", ErrInvalidParent, tenant.Bits())
	}
	if attrs.Name == "" {
		return netip.Prefix{}, fmt.Errorf("%w: BranchAttrs.Name required", ErrInvalidAttrs)
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return netip.Prefix{}, err
	}
	defer tx.Rollback()

	if err := acquireLock(ctx, tx, tenant); err != nil {
		return netip.Prefix{}, err
	}

	// branch_slot 1..254：0 = unbound 哨兵（device.device_ipv6 byte 6=0 表示未绑 branch）；
	// 0xFF (255) 保留作 subject namespace（resident HoA / caregiver 等）
	var nextSlot int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(branch_slot), 0) + 1
		FROM branches WHERE branch_id << $1::INET
	`, tenant.String()).Scan(&nextSlot)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("query MAX branch_slot: %w", err)
	}
	if nextSlot < 1 || nextSlot > 254 {
		return netip.Prefix{}, ErrSlotExhausted
	}

	branchPrefix, err := spatial.DeriveBranchPrefix(tenant, uint8(nextSlot))
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("derive branch prefix: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO branches (branch_id, branch_slot, branch_name, timezone, address)
		VALUES ($1::INET, $2, $3, NULLIF($4, ''), NULLIF($5, ''))
	`, branchPrefix.String(), nextSlot, attrs.Name, attrs.Timezone, attrs.Address)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("insert branch: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return netip.Prefix{}, err
	}
	p.recordKeaLease(ctx, branchPrefix, "branch:"+attrs.Name)
	return branchPrefix, nil
}

// =============================================================================
// AllocateSite — branch /56 → site /64 (slot = building<<4 | floor)
// =============================================================================

func (p *PGBackend) AllocateSite(ctx context.Context, branch netip.Prefix, building, floor uint8, attrs SiteAttrs) (netip.Prefix, error) {
	if branch.Bits() != spatial.PrefixLenBranch {
		return netip.Prefix{}, fmt.Errorf("%w: AllocateSite expects /56, got /%d", ErrInvalidParent, branch.Bits())
	}
	if building > 14 || floor > 14 {
		return netip.Prefix{}, fmt.Errorf("%w: building/floor must be 0..14 (15=0xF reserved)", ErrInvalidAttrs)
	}

	sitePrefix, err := spatial.DeriveSitePrefix(branch, building, floor)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("derive site prefix: %w", err)
	}
	siteSlot := spatial.PackSiteSlot(building, floor)

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return netip.Prefix{}, err
	}
	defer tx.Rollback()

	if err := acquireLock(ctx, tx, branch); err != nil {
		return netip.Prefix{}, err
	}

	// site 是显式指定 (building, floor)，不是 MAX+1 自动分配；冲突直接报
	_, err = tx.ExecContext(ctx, `
		INSERT INTO sites (site_id, site_slot, building, floor, site_name, address)
		VALUES ($1::INET, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''))
	`, sitePrefix.String(), siteSlot, building, floor, attrs.Name, attrs.Address)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("insert site: %w", wrapDuplicate(err))
	}

	if err := tx.Commit(); err != nil {
		return netip.Prefix{}, err
	}
	p.recordKeaLease(ctx, sitePrefix, "site:"+attrs.Name)
	return sitePrefix, nil
}

// =============================================================================
// AllocateUnit — site /64 → unit /80 (auto slot)
// =============================================================================

func (p *PGBackend) AllocateUnit(ctx context.Context, site netip.Prefix, attrs UnitAttrs) (netip.Prefix, error) {
	if site.Bits() != spatial.PrefixLenSite {
		return netip.Prefix{}, fmt.Errorf("%w: AllocateUnit expects /64, got /%d", ErrInvalidParent, site.Bits())
	}
	if attrs.Name == "" {
		return netip.Prefix{}, fmt.Errorf("%w: UnitAttrs.Name required", ErrInvalidAttrs)
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return netip.Prefix{}, err
	}
	defer tx.Rollback()

	if err := acquireLock(ctx, tx, site); err != nil {
		return netip.Prefix{}, err
	}

	var nextSlot int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(unit_slot), 0) + 1
		FROM units WHERE unit_id << $1::INET
	`, site.String()).Scan(&nextSlot)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("query MAX unit_slot: %w", err)
	}
	if nextSlot > 65534 {
		return netip.Prefix{}, ErrSlotExhausted
	}

	unitPrefix, err := spatial.DeriveUnitPrefix(site, uint16(nextSlot))
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("derive unit prefix: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO units (unit_id, unit_slot, unit_name, unit_type,
		                   unit_layout_type, is_public, is_shared_unit, timezone)
		VALUES ($1::INET, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, $7, NULLIF($8, ''))
	`, unitPrefix.String(), nextSlot, attrs.Name, attrs.UnitType,
		attrs.LayoutType, attrs.IsPublic, attrs.IsShared, attrs.Timezone)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("insert unit: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return netip.Prefix{}, err
	}
	p.recordKeaLease(ctx, unitPrefix, "unit:"+attrs.Name)
	return unitPrefix, nil
}

// =============================================================================
// AllocateRoom — unit /80 → room /88 (auto slot)
// =============================================================================

func (p *PGBackend) AllocateRoom(ctx context.Context, unit netip.Prefix, attrs RoomAttrs) (netip.Prefix, error) {
	if unit.Bits() != spatial.PrefixLenUnit {
		return netip.Prefix{}, fmt.Errorf("%w: AllocateRoom expects /80, got /%d", ErrInvalidParent, unit.Bits())
	}
	if attrs.Name == "" {
		return netip.Prefix{}, fmt.Errorf("%w: RoomAttrs.Name required", ErrInvalidAttrs)
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return netip.Prefix{}, err
	}
	defer tx.Rollback()

	if err := acquireLock(ctx, tx, unit); err != nil {
		return netip.Prefix{}, err
	}

	var nextSlot int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(room_slot), 0) + 1
		FROM rooms WHERE room_id << $1::INET
	`, unit.String()).Scan(&nextSlot)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("query MAX room_slot: %w", err)
	}
	if nextSlot > 254 {
		return netip.Prefix{}, ErrSlotExhausted
	}

	roomPrefix, err := spatial.DeriveRoomPrefix(unit, uint8(nextSlot))
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("derive room prefix: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO rooms (room_id, room_slot, room_name, room_type, is_primary)
		VALUES ($1::INET, $2, $3, NULLIF($4, ''), $5)
	`, roomPrefix.String(), nextSlot, attrs.Name, attrs.RoomType, attrs.IsPrimary)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("insert room: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return netip.Prefix{}, err
	}
	p.recordKeaLease(ctx, roomPrefix, "room:"+attrs.Name)
	return roomPrefix, nil
}

// =============================================================================
// AllocateBed — room /88 → bed /96 (auto slot)
// =============================================================================

func (p *PGBackend) AllocateBed(ctx context.Context, room netip.Prefix, attrs BedAttrs) (netip.Prefix, error) {
	if room.Bits() != spatial.PrefixLenRoom {
		return netip.Prefix{}, fmt.Errorf("%w: AllocateBed expects /88, got /%d", ErrInvalidParent, room.Bits())
	}
	if attrs.Name == "" {
		return netip.Prefix{}, fmt.Errorf("%w: BedAttrs.Name required", ErrInvalidAttrs)
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return netip.Prefix{}, err
	}
	defer tx.Rollback()

	if err := acquireLock(ctx, tx, room); err != nil {
		return netip.Prefix{}, err
	}

	var nextSlot int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(bed_slot), 0) + 1
		FROM beds WHERE bed_id << $1::INET
	`, room.String()).Scan(&nextSlot)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("query MAX bed_slot: %w", err)
	}
	if nextSlot > 254 {
		return netip.Prefix{}, ErrSlotExhausted
	}

	bedPrefix, err := spatial.DeriveBedPrefix(room, uint8(nextSlot))
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("derive bed prefix: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO beds (bed_id, bed_slot, bed_name, description)
		VALUES ($1::INET, $2, $3, NULLIF($4, ''))
	`, bedPrefix.String(), nextSlot, attrs.Name, attrs.Description)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("insert bed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return netip.Prefix{}, err
	}
	p.recordKeaLease(ctx, bedPrefix, "bed:"+attrs.Name)
	return bedPrefix, nil
}

// =============================================================================
// RegisterDevice — base prefix (/96 bed 或 /88 room 或 /80 unit) + device_uid → /128
// =============================================================================

func (p *PGBackend) RegisterDevice(ctx context.Context, base netip.Prefix, deviceID, deviceUID string) (netip.Addr, error) {
	if base.Bits() < spatial.PrefixLenUnit || base.Bits() > spatial.PrefixLenBed {
		return netip.Addr{}, fmt.Errorf("%w: device base must be /80~/96, got /%d", ErrInvalidParent, base.Bits())
	}
	if deviceID == "" || deviceUID == "" {
		return netip.Addr{}, fmt.Errorf("%w: device_id and device_uid required", ErrInvalidAttrs)
	}

	addr, err := spatial.DeriveDeviceAddr(base, deviceUID)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("derive device addr: %w", err)
	}

	// device 是 stateless 派生，理论上无并发冲突；用 INSERT ... ON CONFLICT DO NOTHING + 检查行数
	// PG INET cast 对 IPv6 host address 默认 /128 (无需显式 suffix)
	res, err := p.db.ExecContext(ctx, `
		INSERT INTO devices (device_ipv6, device_id, monitoring_enabled)
		VALUES ($1::INET, $2, TRUE)
		ON CONFLICT (device_ipv6) DO NOTHING
	`, addr.String(), deviceID)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("insert device: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		// 已存在 — 校验绑定的 device_id 一致
		var existing string
		err := p.db.QueryRowContext(ctx, `SELECT device_id::text FROM devices WHERE device_ipv6 = $1::INET`, addr.String()).Scan(&existing)
		if err != nil {
			return netip.Addr{}, fmt.Errorf("read existing device: %w", err)
		}
		if existing != deviceID {
			return netip.Addr{}, fmt.Errorf("%w: addr %s already bound to device %s", ErrAlreadyExists, addr, existing)
		}
		// 同 device_id 重复 register 是幂等 OK
	}
	p.recordKeaAddrLease(ctx, addr, "device:"+deviceID)
	return addr, nil
}

// =============================================================================
// LookupTenant
// =============================================================================

func (p *PGBackend) LookupTenant(ctx context.Context, tenant netip.Prefix) (*TenantInfo, error) {
	if tenant.Bits() != spatial.PrefixLenTenant {
		return nil, fmt.Errorf("%w: LookupTenant expects /48, got /%d", ErrInvalidParent, tenant.Bits())
	}
	t := &TenantInfo{Prefix: tenant}
	var contactName, contactEmail, contactPhone sql.NullString
	err := p.db.QueryRowContext(ctx, `
		SELECT tenant_slot, tenant_name, timezone, contact_name, contact_email, contact_phone
		FROM tenants WHERE tenant_id = $1::INET
	`, tenant.String()).Scan(&t.Slot, &t.Name, &t.Timezone, &contactName, &contactEmail, &contactPhone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	t.ContactName = contactName.String
	t.ContactEmail = contactEmail.String
	t.ContactPhone = contactPhone.String
	return t, nil
}

// =============================================================================
// Internal helpers
// =============================================================================

// acquireLock 取 PG advisory lock，key = parent prefix 前 8 字节 hash。
// 同 parent 并发分配 slot 时串行执行；commit/rollback 后自动释放。
func acquireLock(ctx context.Context, tx *sql.Tx, parent netip.Prefix) error {
	b := parent.Masked().Addr().As16()
	lockKey := int64(binary.BigEndian.Uint64(b[:8]))
	_, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", lockKey)
	if err != nil {
		return fmt.Errorf("acquire advisory lock for %s: %w", parent, err)
	}
	return nil
}

// wrapDuplicate 把 PG 23505 unique violation 转成 ErrAlreadyExists。
// （site 显式 building/floor 时冲突需要这个区分，而非 generic insert error）
func wrapDuplicate(err error) error {
	if err == nil {
		return nil
	}
	// pq error string 模式匹配；不引入 lib/pq 强类型依赖
	msg := err.Error()
	if containsAny(msg, "duplicate key value", "violates unique constraint", "23505") {
		return fmt.Errorf("%w: %v", ErrAlreadyExists, err)
	}
	return err
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) > 0 && len(s) >= len(sub) {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
