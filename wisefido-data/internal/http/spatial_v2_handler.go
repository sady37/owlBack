package httpapi

// SpatialV2Handler 提供 owl_v2 IPAM/DDNS REST API。
//
// 设计：
//   - 不接现有 v1 service 层 (BranchService/UnitService 等基于 db1.0 UUID schema)
//   - 直接调 owl-common/ipam.Backend (PG-first，可选 kea audit)
//   - 可选 ddns.Client (注册时同步推 BIND)
//   - 路由前缀 /admin/api/v2/spatial/... 与 v1 完全隔离
//
// 所有 POST 接收 JSON：
//   {parent: "fd00:0:3::/48", attrs: {...}}
//
// 返回：
//   {prefix: "fd00:0:3:100::/56", dns_name: "..."}

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"strings"

	"owl-common/ddns"
	"owl-common/ipam"

	"go.uber.org/zap"
)

// SpatialV2Handler 是 v2 spatial IPAM API 处理器。
// ipam 必填；db 必填（LIST/GET/PUT/DELETE 直 SQL，绕过 ipam allocation 抽象）；
// ddns 为 nil 时跳过 DNS 注册（仍正常分配 prefix）。
type SpatialV2Handler struct {
	ipam   ipam.Backend
	db     *sql.DB
	ddns   *ddns.Client // 可选
	logger *zap.Logger
}

// NewSpatialV2Handler 构造。
func NewSpatialV2Handler(backend ipam.Backend, db *sql.DB, ddnsClient *ddns.Client, logger *zap.Logger) *SpatialV2Handler {
	return &SpatialV2Handler{
		ipam:   backend,
		db:     db,
		ddns:   ddnsClient,
		logger: logger,
	}
}

// =============================================================================
// 路由注册
// =============================================================================

// RegisterSpatialV2Routes 把 v2 spatial 路由注册到 router。
//
// Path 形态：
//   - /spatial/<entity>          POST=alloc, GET=list (tenant scope via query/header)
//   - /spatial/<entity>/{prefix} GET=detail, PUT=update, DELETE=soft-delete
//     prefix 用 `_` 替代 `:`，斜杠用 `_` (URL safe)，例如 fd00_0_3_100___56 = fd00:0:3:100::/56
func (r *Router) RegisterSpatialV2Routes(h *SpatialV2Handler) {
	r.Handle("/admin/api/v2/spatial/branches", h.HandleBranches)
	r.Handle("/admin/api/v2/spatial/branches/", h.HandleBranchByPrefix)
	r.Handle("/admin/api/v2/spatial/sites", h.HandleSites)
	r.Handle("/admin/api/v2/spatial/units", h.HandleUnits)
	r.Handle("/admin/api/v2/spatial/rooms", h.HandleRooms)
	r.Handle("/admin/api/v2/spatial/beds", h.HandleBeds)
	r.Handle("/admin/api/v2/spatial/devices", h.HandleDevices)
	r.Handle("/admin/api/v2/spatial/tenants/", h.HandleLookupTenant)
}

// =============================================================================
// Request/Response DTO
// =============================================================================

type allocBranchReq struct {
	Parent string            `json:"parent"` // tenant /48, e.g. "fd00:0:3::/48"
	Attrs  ipam.BranchAttrs  `json:"attrs"`
}

type allocSiteReq struct {
	Parent   string         `json:"parent"`   // branch /56
	Building uint8          `json:"building"` // 0..14
	Floor    uint8          `json:"floor"`    // 0..14
	Attrs    ipam.SiteAttrs `json:"attrs"`
}

type allocUnitReq struct {
	Parent string         `json:"parent"` // site /64
	Attrs  ipam.UnitAttrs `json:"attrs"`
}

type allocRoomReq struct {
	Parent string         `json:"parent"` // unit /80
	Attrs  ipam.RoomAttrs `json:"attrs"`
}

type allocBedReq struct {
	Parent string        `json:"parent"` // room /88
	Attrs  ipam.BedAttrs `json:"attrs"`
}

type registerDeviceReq struct {
	Base       string `json:"base"`        // bed /96 (also unit /80 / room /88 acceptable for public area device)
	DeviceID   string `json:"device_id"`   // UUID 必须先在 device_factory_meta 已存在
	DeviceUID  string `json:"device_uid"`  // logMAC，末 32bit 派生 /128 host part
	DNSName    string `json:"dns_name"`    // optional 短名 (e.g. "john-bed")，传则同步 DDNS 注册
	DNSZone    string `json:"dns_zone"`    // 同上必填，e.g. "tenant3.owl."
}

type allocResp struct {
	Prefix  string `json:"prefix,omitempty"`
	Address string `json:"address,omitempty"`
	DNSFQDN string `json:"dns_fqdn,omitempty"`
}

// =============================================================================
// Handlers
// =============================================================================

// branchRow 是 GET LIST / GET detail 的统一返回 DTO。
type branchRow struct {
	Prefix    string                `json:"prefix"`               // /56 CIDR
	Slot      int                   `json:"slot"`                 // branch_slot 1..254
	Name      string                `json:"name"`
	Address   string                `json:"address,omitempty"`
	Timezone  string                `json:"timezone,omitempty"`
	Status    string                `json:"status"`               // active / suspended / deleted
	CreatedAt string                `json:"created_at"`
	UpdatedAt string                `json:"updated_at"`
	Units     []branchUnitItem      `json:"units,omitempty"`      // 该 branch 下的 units（prefix-match /56→/80）
	Residents []branchResidentItem  `json:"residents,omitempty"`  // 该 branch 下的当前活跃 residents（via resident_unit）
	Users     []branchUserItem      `json:"users,omitempty"`      // 该 branch 下的 user_roles scope 用户（best-effort）
}

type branchUnitItem struct {
	UnitPrefix string `json:"unit_prefix"` // /80 INET
	UnitName   string `json:"unit_name"`
}

type branchResidentItem struct {
	ResidentHoA string `json:"resident_hoa"` // /128 INET
	Nickname    string `json:"nickname"`
}

type branchUserItem struct {
	UserID      string `json:"user_id"`       // UUID
	UserAccount string `json:"user_account"`
	Nickname    string `json:"nickname"`
}

type updateBranchReq struct {
	Name     *string `json:"name,omitempty"`     // nil = 不改
	Address  *string `json:"address,omitempty"`  // 空字符串 = 清空
	Timezone *string `json:"timezone,omitempty"` // 空字符串 = 清空 (继承 tenant)
	Status   *string `json:"status,omitempty"`   // active / suspended / deleted
}

func (h *SpatialV2Handler) HandleBranches(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.allocBranch(w, r)
	case http.MethodGet:
		h.listBranches(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *SpatialV2Handler) allocBranch(w http.ResponseWriter, r *http.Request) {
	var body allocBranchReq
	if !decodeJSON(w, r, &body) {
		return
	}
	parent, err := parsePrefix(body.Parent)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid parent: "+err.Error()))
		return
	}
	if !h.scopeAllows(r, parent) {
		writeJSON(w, http.StatusOK, Fail("permission denied: cross-tenant allocation requires SystemAdmin"))
		return
	}
	prefix, err := h.ipam.AllocateBranch(r.Context(), parent, body.Attrs)
	if err != nil {
		h.respondAllocError(w, "AllocateBranch", err)
		return
	}
	h.logger.Info("v2 spatial: branch allocated", zap.String("prefix", prefix.String()), zap.String("name", body.Attrs.Name))
	writeJSON(w, http.StatusOK, Ok(allocResp{Prefix: prefix.String()}))
}

// listBranches GET /admin/api/v2/spatial/branches?tenant=fd00:0:3::/48[&include_deleted=true]
// 不传 tenant 时回退到 X-Tenant-Id header；SystemAdmin 可显式跨 tenant 查；其它角色被 gate。
func (h *SpatialV2Handler) listBranches(w http.ResponseWriter, r *http.Request) {
	tenantStr := r.URL.Query().Get("tenant")
	if tenantStr == "" {
		tenantStr = r.Header.Get("X-Tenant-Id")
	}
	if tenantStr == "" {
		writeJSON(w, http.StatusBadRequest, Fail("tenant prefix required (query param `tenant` or X-Tenant-Id header)"))
		return
	}
	tenant, err := parsePrefix(tenantStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid tenant: "+err.Error()))
		return
	}
	if !h.scopeAllows(r, tenant) {
		writeJSON(w, http.StatusOK, Fail("permission denied: cross-tenant list requires SystemAdmin"))
		return
	}
	includeDeleted := strings.EqualFold(r.URL.Query().Get("include_deleted"), "true")
	// 注：这里改 branch_id <<= 同时是为了 prefix 自身也匹配（边界容错）
	q := `
		SELECT host(branch_id)||'/'||masklen(branch_id), branch_slot, branch_name,
		       COALESCE(address,''), COALESCE(timezone,''), status,
		       to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		FROM branches
		WHERE branch_id << $1::INET
	`
	if !includeDeleted {
		q += " AND status <> 'deleted'"
	}
	q += " ORDER BY branch_slot"
	rows, err := h.db.QueryContext(r.Context(), q, tenant.String())
	if err != nil {
		h.logger.Error("v2 spatial: list branches failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("list branches failed: "+err.Error()))
		return
	}
	defer rows.Close()
	items := make([]branchRow, 0, 8)
	for rows.Next() {
		var b branchRow
		if err := rows.Scan(&b.Prefix, &b.Slot, &b.Name, &b.Address, &b.Timezone, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			h.logger.Error("v2 spatial: scan branch row failed", zap.Error(err))
			writeJSON(w, http.StatusOK, Fail("scan branch row failed: "+err.Error()))
			return
		}
		items = append(items, b)
	}
	rows.Close()

	// 用 prefix-match join units / residents / users（v1 等价：每 branch 3 query，与 v1 BranchService 一致）
	for i := range items {
		if items[i].Status == "deleted" {
			continue // 软删 branch 不附 children
		}
		if u, err := h.fetchBranchUnits(r.Context(), items[i].Prefix); err == nil {
			items[i].Units = u
		} else {
			h.logger.Warn("v2 spatial: fetch branch units failed", zap.String("prefix", items[i].Prefix), zap.Error(err))
		}
		if res, err := h.fetchBranchResidents(r.Context(), items[i].Prefix); err == nil {
			items[i].Residents = res
		} else {
			h.logger.Warn("v2 spatial: fetch branch residents failed", zap.String("prefix", items[i].Prefix), zap.Error(err))
		}
		if us, err := h.fetchBranchUsers(r.Context(), items[i].Prefix); err == nil {
			items[i].Users = us
		} else {
			h.logger.Warn("v2 spatial: fetch branch users failed", zap.String("prefix", items[i].Prefix), zap.Error(err))
		}
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": items,
		"total": len(items),
	}))
}

// fetchBranchUnits — units 在该 branch /56 下（unit_id /80 << branch /56）
func (h *SpatialV2Handler) fetchBranchUnits(ctx context.Context, branchPrefix string) ([]branchUnitItem, error) {
	rows, err := h.db.QueryContext(ctx, `
		SELECT host(unit_id)||'/'||masklen(unit_id), unit_name
		FROM units WHERE unit_id << $1::INET
		ORDER BY unit_name ASC
	`, branchPrefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]branchUnitItem, 0, 4)
	for rows.Next() {
		var u branchUnitItem
		if err := rows.Scan(&u.UnitPrefix, &u.UnitName); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// fetchBranchResidents — residents 当前活跃绑定到该 branch /56 下的 unit（resident_unit + units 联表）
// 注意：residents.resident_id 在 subject namespace 不直接 prefix-match branch；用 resident_unit.spatial_prefix 反查。
func (h *SpatialV2Handler) fetchBranchResidents(ctx context.Context, branchPrefix string) ([]branchResidentItem, error) {
	rows, err := h.db.QueryContext(ctx, `
		SELECT DISTINCT host(r.resident_id)||'/'||masklen(r.resident_id), r.nickname
		FROM residents r
		JOIN resident_unit ru ON ru.resident_id = r.resident_id AND ru.valid_to IS NULL
		WHERE ru.spatial_prefix << $1::INET AND r.status = 'active'
		ORDER BY r.nickname ASC
	`, branchPrefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]branchResidentItem, 0, 4)
	for rows.Next() {
		var r branchResidentItem
		if err := rows.Scan(&r.ResidentHoA, &r.Nickname); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// fetchBranchUsers — users 通过 user_roles.scope 关联该 branch /56；scope NULL = tenant-wide (不算入 branch 列)
func (h *SpatialV2Handler) fetchBranchUsers(ctx context.Context, branchPrefix string) ([]branchUserItem, error) {
	rows, err := h.db.QueryContext(ctx, `
		SELECT DISTINCT u.user_id::text, u.user_account, COALESCE(u.nickname,'')
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.user_id AND ur.valid_to IS NULL
		WHERE ur.scope IS NOT NULL AND ur.scope <<= $1::INET AND u.status = 'active'
		ORDER BY u.user_account ASC
	`, branchPrefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]branchUserItem, 0, 4)
	for rows.Next() {
		var u branchUserItem
		if err := rows.Scan(&u.UserID, &u.UserAccount, &u.Nickname); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// HandleBranchByPrefix dispatches GET/PUT/DELETE on /admin/api/v2/spatial/branches/{prefix-encoded}
func (h *SpatialV2Handler) HandleBranchByPrefix(w http.ResponseWriter, r *http.Request) {
	const base = "/admin/api/v2/spatial/branches/"
	encoded := strings.TrimPrefix(r.URL.Path, base)
	if encoded == "" {
		writeJSON(w, http.StatusBadRequest, Fail("branch prefix required"))
		return
	}
	prefix, err := decodePrefixSegment(encoded, r.URL.Query().Get("prefix"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid branch prefix: "+err.Error()))
		return
	}
	if prefix.Bits() != 56 {
		writeJSON(w, http.StatusBadRequest, Fail("branch prefix must be /56, got /"+fmt.Sprint(prefix.Bits())))
		return
	}
	// Scope gate: tenant 包含 branch
	tenant, _ := prefix.Addr().Prefix(48)
	if !h.scopeAllows(r, tenant) {
		writeJSON(w, http.StatusOK, Fail("permission denied: cross-tenant access requires SystemAdmin"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getBranch(w, r, prefix)
	case http.MethodPut:
		h.updateBranch(w, r, prefix)
	case http.MethodDelete:
		h.deleteBranch(w, r, prefix)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *SpatialV2Handler) getBranch(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
	var b branchRow
	var addr, tz sql.NullString
	err := h.db.QueryRowContext(r.Context(), `
		SELECT host(branch_id)||'/'||masklen(branch_id), branch_slot, branch_name,
		       address, timezone, status,
		       to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		FROM branches WHERE branch_id = $1::INET
	`, prefix.String()).Scan(&b.Prefix, &b.Slot, &b.Name, &addr, &tz, &b.Status, &b.CreatedAt, &b.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusOK, Fail("branch not found: "+prefix.String()))
		return
	}
	if err != nil {
		h.logger.Error("v2 spatial: get branch failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("get branch failed: "+err.Error()))
		return
	}
	b.Address = addr.String
	b.Timezone = tz.String
	writeJSON(w, http.StatusOK, Ok(b))
}

func (h *SpatialV2Handler) updateBranch(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
	var body updateBranchReq
	if !decodeJSON(w, r, &body) {
		return
	}
	sets := make([]string, 0, 4)
	args := make([]any, 0, 5)
	args = append(args, prefix.String()) // $1 = WHERE
	idx := 2
	if body.Name != nil {
		if *body.Name == "" {
			writeJSON(w, http.StatusBadRequest, Fail("branch_name cannot be empty"))
			return
		}
		sets = append(sets, fmt.Sprintf("branch_name = $%d", idx))
		args = append(args, *body.Name)
		idx++
	}
	if body.Address != nil {
		sets = append(sets, fmt.Sprintf("address = NULLIF($%d, '')", idx))
		args = append(args, *body.Address)
		idx++
	}
	if body.Timezone != nil {
		sets = append(sets, fmt.Sprintf("timezone = NULLIF($%d, '')", idx))
		args = append(args, *body.Timezone)
		idx++
	}
	if body.Status != nil {
		switch *body.Status {
		case "active", "suspended", "deleted":
		default:
			writeJSON(w, http.StatusBadRequest, Fail("status must be one of: active / suspended / deleted"))
			return
		}
		sets = append(sets, fmt.Sprintf("status = $%d", idx))
		args = append(args, *body.Status)
		idx++
	}
	if len(sets) == 0 {
		writeJSON(w, http.StatusBadRequest, Fail("no fields to update"))
		return
	}
	sets = append(sets, "updated_at = NOW()")
	q := "UPDATE branches SET " + strings.Join(sets, ", ") + " WHERE branch_id = $1::INET"
	res, err := h.db.ExecContext(r.Context(), q, args...)
	if err != nil {
		h.logger.Error("v2 spatial: update branch failed", zap.Error(err), zap.String("prefix", prefix.String()))
		writeJSON(w, http.StatusOK, Fail("update branch failed: "+err.Error()))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, http.StatusOK, Fail("branch not found: "+prefix.String()))
		return
	}
	h.logger.Info("v2 spatial: branch updated", zap.String("prefix", prefix.String()), zap.Int("fields", len(sets)-1))
	writeJSON(w, http.StatusOK, Ok(map[string]any{"prefix": prefix.String(), "updated": true}))
}

// deleteBranch 软删 — status='deleted'。R-002: HIPAA + 业务恢复，不硬删。
func (h *SpatialV2Handler) deleteBranch(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
	res, err := h.db.ExecContext(r.Context(), `
		UPDATE branches SET status='deleted', updated_at=NOW()
		WHERE branch_id = $1::INET AND status <> 'deleted'
	`, prefix.String())
	if err != nil {
		h.logger.Error("v2 spatial: soft-delete branch failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("soft-delete branch failed: "+err.Error()))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, http.StatusOK, Fail("branch not found or already deleted: "+prefix.String()))
		return
	}
	h.logger.Info("v2 spatial: branch soft-deleted", zap.String("prefix", prefix.String()))
	writeJSON(w, http.StatusOK, Ok(map[string]any{"prefix": prefix.String(), "status": "deleted"}))
}

func (h *SpatialV2Handler) HandleSites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body allocSiteReq
	if !decodeJSON(w, r, &body) {
		return
	}
	parent, err := parsePrefix(body.Parent)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid parent: "+err.Error()))
		return
	}
	prefix, err := h.ipam.AllocateSite(r.Context(), parent, body.Building, body.Floor, body.Attrs)
	if err != nil {
		h.respondAllocError(w, "AllocateSite", err)
		return
	}
	h.logger.Info("v2 spatial: site allocated", zap.String("prefix", prefix.String()))
	writeJSON(w, http.StatusOK, Ok(allocResp{Prefix: prefix.String()}))
}

func (h *SpatialV2Handler) HandleUnits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body allocUnitReq
	if !decodeJSON(w, r, &body) {
		return
	}
	parent, err := parsePrefix(body.Parent)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid parent: "+err.Error()))
		return
	}
	prefix, err := h.ipam.AllocateUnit(r.Context(), parent, body.Attrs)
	if err != nil {
		h.respondAllocError(w, "AllocateUnit", err)
		return
	}
	h.logger.Info("v2 spatial: unit allocated", zap.String("prefix", prefix.String()), zap.String("name", body.Attrs.Name))
	writeJSON(w, http.StatusOK, Ok(allocResp{Prefix: prefix.String()}))
}

func (h *SpatialV2Handler) HandleRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body allocRoomReq
	if !decodeJSON(w, r, &body) {
		return
	}
	parent, err := parsePrefix(body.Parent)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid parent: "+err.Error()))
		return
	}
	prefix, err := h.ipam.AllocateRoom(r.Context(), parent, body.Attrs)
	if err != nil {
		h.respondAllocError(w, "AllocateRoom", err)
		return
	}
	h.logger.Info("v2 spatial: room allocated", zap.String("prefix", prefix.String()))
	writeJSON(w, http.StatusOK, Ok(allocResp{Prefix: prefix.String()}))
}

func (h *SpatialV2Handler) HandleBeds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body allocBedReq
	if !decodeJSON(w, r, &body) {
		return
	}
	parent, err := parsePrefix(body.Parent)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid parent: "+err.Error()))
		return
	}
	prefix, err := h.ipam.AllocateBed(r.Context(), parent, body.Attrs)
	if err != nil {
		h.respondAllocError(w, "AllocateBed", err)
		return
	}
	h.logger.Info("v2 spatial: bed allocated", zap.String("prefix", prefix.String()))
	writeJSON(w, http.StatusOK, Ok(allocResp{Prefix: prefix.String()}))
}

// HandleDevices 注册 device + 可选 DDNS 推 BIND。
func (h *SpatialV2Handler) HandleDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body registerDeviceReq
	if !decodeJSON(w, r, &body) {
		return
	}
	base, err := parsePrefix(body.Base)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid base: "+err.Error()))
		return
	}
	if body.DeviceID == "" || body.DeviceUID == "" {
		writeJSON(w, http.StatusBadRequest, Fail("device_id and device_uid required"))
		return
	}
	addr, err := h.ipam.RegisterDevice(r.Context(), base, body.DeviceID, body.DeviceUID)
	if err != nil {
		h.respondAllocError(w, "RegisterDevice", err)
		return
	}

	resp := allocResp{Address: addr.String()}

	// 可选 DDNS 注册（dns_name + dns_zone 同时给才推）
	if body.DNSName != "" && body.DNSZone != "" {
		if h.ddns == nil {
			h.logger.Warn("v2 spatial: dns_name given but ddns client not configured", zap.String("dns_name", body.DNSName))
		} else {
			if err := h.ddns.RegisterDevice(r.Context(), addr, body.DNSName, body.DNSZone); err != nil {
				// DDNS 失败不回滚 device 注册，但报告给客户端让它重试
				h.logger.Error("v2 spatial: DDNS register failed", zap.Error(err), zap.String("addr", addr.String()))
				writeJSON(w, http.StatusOK, Fail(fmt.Sprintf(
					"device %s registered but DDNS failed: %v (retry DDNS independently)",
					addr, err)))
				return
			}
			zone := body.DNSZone
			if !strings.HasSuffix(zone, ".") {
				zone += "."
			}
			resp.DNSFQDN = body.DNSName + "." + zone
		}
	}

	h.logger.Info("v2 spatial: device registered",
		zap.String("addr", addr.String()),
		zap.String("device_id", body.DeviceID),
		zap.String("dns_fqdn", resp.DNSFQDN))
	writeJSON(w, http.StatusOK, Ok(resp))
}

// HandleLookupTenant GET /admin/api/v2/spatial/tenants/{prefix-encoded}
// 约定：客户端 URL-encode prefix (`encodeURIComponent`)；net/http 解码 path 后即原 CIDR。
func (h *SpatialV2Handler) HandleLookupTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	const prefixURL = "/admin/api/v2/spatial/tenants/"
	encoded := strings.TrimPrefix(r.URL.Path, prefixURL)
	if encoded == "" {
		writeJSON(w, http.StatusBadRequest, Fail("tenant prefix required"))
		return
	}
	parent, err := decodePrefixSegment(encoded, r.URL.Query().Get("prefix"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid tenant prefix: "+err.Error()))
		return
	}
	if !h.scopeAllows(r, parent) {
		writeJSON(w, http.StatusOK, Fail("permission denied: cross-tenant lookup requires SystemAdmin"))
		return
	}
	t, err := h.ipam.LookupTenant(r.Context(), parent)
	if err != nil {
		if errors.Is(err, ipam.ErrNotFound) {
			writeJSON(w, http.StatusOK, Fail("tenant not found: "+parent.String()))
			return
		}
		writeJSON(w, http.StatusOK, Fail("lookup failed: "+err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"prefix":   t.Prefix.String(),
		"slot":     t.Slot,
		"name":     t.Name,
		"timezone": t.Timezone,
		"contact": map[string]string{
			"name":  t.ContactName,
			"email": t.ContactEmail,
			"phone": t.ContactPhone,
		},
	}))
}

// =============================================================================
// Helpers
// =============================================================================

func parsePrefix(s string) (netip.Prefix, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return netip.Prefix{}, errors.New("empty")
	}
	return netip.ParsePrefix(s)
}

// decodePrefixSegment 把 URL path 段解成 prefix。
// 约定：客户端用 `encodeURIComponent` URL-编码 prefix，path 经 net/http 解码后即原 CIDR。
// 兼容：可选 ?prefix= 查询参数覆盖（用于调试 curl 等场景）。
func decodePrefixSegment(pathSeg, queryPrefix string) (netip.Prefix, error) {
	src := queryPrefix
	if src == "" {
		src = pathSeg
	}
	return parsePrefix(src)
}

// scopeAllows checks whether the current caller may operate within `scope`.
// SystemAdmin / SystemOperator pass; others must X-Tenant-Id ⊇ scope.
func (h *SpatialV2Handler) scopeAllows(r *http.Request, scope netip.Prefix) bool {
	role := r.Header.Get("X-User-Role")
	if strings.EqualFold(role, "SystemAdmin") || strings.EqualFold(role, "SystemOperator") {
		return true
	}
	own := r.Header.Get("X-Tenant-Id")
	if own == "" {
		return false
	}
	ownPrefix, err := parsePrefix(own)
	if err != nil {
		return false
	}
	// caller's tenant must contain (or equal) the requested scope
	return ownPrefix.Contains(scope.Addr())
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid JSON body: "+err.Error()))
		return false
	}
	return true
}

// respondAllocError 把 ipam 错误映射到 HTTP 响应。
// 全部用 200+Fail (与现有 handler 风格一致)。
func (h *SpatialV2Handler) respondAllocError(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, ipam.ErrInvalidParent):
		writeJSON(w, http.StatusBadRequest, Fail(op+": "+err.Error()))
	case errors.Is(err, ipam.ErrInvalidAttrs):
		writeJSON(w, http.StatusBadRequest, Fail(op+": "+err.Error()))
	case errors.Is(err, ipam.ErrSlotExhausted):
		writeJSON(w, http.StatusOK, Fail(op+": slot range exhausted at this level"))
	case errors.Is(err, ipam.ErrAlreadyExists):
		writeJSON(w, http.StatusOK, Fail(op+": "+err.Error()))
	case errors.Is(err, ipam.ErrNotFound):
		writeJSON(w, http.StatusOK, Fail(op+": "+err.Error()))
	default:
		h.logger.Error("v2 spatial: "+op+" failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(op+" failed: "+err.Error()))
	}
}
