package httpapi

// SpatialHandler 提供 owl_v2 IPAM/DDNS REST API。
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

// SpatialHandler 是 v2 spatial IPAM API 处理器。
// ipam 必填；db 必填（LIST/GET/PUT/DELETE 直 SQL，绕过 ipam allocation 抽象）；
// ddns 为 nil 时跳过 DNS 注册（仍正常分配 prefix）。
type SpatialHandler struct {
	ipam   ipam.Backend
	db     *sql.DB
	ddns   *ddns.Client // 可选
	logger *zap.Logger
}

// NewSpatialHandler 构造。
func NewSpatialHandler(backend ipam.Backend, db *sql.DB, ddnsClient *ddns.Client, logger *zap.Logger) *SpatialHandler {
	return &SpatialHandler{
		ipam:   backend,
		db:     db,
		ddns:   ddnsClient,
		logger: logger,
	}
}

// =============================================================================
// 路由注册
// =============================================================================

// RegisterSpatialRoutes 把 v2 spatial 路由注册到 router。
//
// Path 形态：
//   - /spatial/<entity>          POST=alloc, GET=list (tenant scope via query/header)
//   - /spatial/<entity>/{prefix} GET=detail, PUT=update, DELETE=soft-delete
//     prefix 用 `_` 替代 `:`，斜杠用 `_` (URL safe)，例如 fd00_0_3_100___56 = fd00:0:3:100::/56
func (r *Router) RegisterSpatialRoutes(h *SpatialHandler) {
	r.Handle("/admin/api/v2/spatial/branches", h.HandleBranches)
	r.Handle("/admin/api/v2/spatial/branches/", h.HandleBranchByPrefix)
	r.Handle("/admin/api/v2/spatial/sites", h.HandleSites)
	r.Handle("/admin/api/v2/spatial/units", h.HandleUnits)
	r.Handle("/admin/api/v2/spatial/units/", h.HandleUnitByPrefix)
	r.Handle("/admin/api/v2/spatial/rooms", h.HandleRooms)
	r.Handle("/admin/api/v2/spatial/rooms/", h.HandleRoomByPrefix)
	r.Handle("/admin/api/v2/spatial/beds", h.HandleBeds)
	r.Handle("/admin/api/v2/spatial/beds/", h.HandleBedByPrefix)
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

func (h *SpatialHandler) HandleBranches(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.allocBranch(w, r)
	case http.MethodGet:
		h.listBranches(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *SpatialHandler) allocBranch(w http.ResponseWriter, r *http.Request) {
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
func (h *SpatialHandler) listBranches(w http.ResponseWriter, r *http.Request) {
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
func (h *SpatialHandler) fetchBranchUnits(ctx context.Context, branchPrefix string) ([]branchUnitItem, error) {
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
func (h *SpatialHandler) fetchBranchResidents(ctx context.Context, branchPrefix string) ([]branchResidentItem, error) {
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
func (h *SpatialHandler) fetchBranchUsers(ctx context.Context, branchPrefix string) ([]branchUserItem, error) {
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
func (h *SpatialHandler) HandleBranchByPrefix(w http.ResponseWriter, r *http.Request) {
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

func (h *SpatialHandler) getBranch(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
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

func (h *SpatialHandler) updateBranch(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
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
func (h *SpatialHandler) deleteBranch(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
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

func (h *SpatialHandler) HandleSites(w http.ResponseWriter, r *http.Request) {
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

// =============================================================================
// Units — LIST/GET/PUT/DELETE + alloc (POST)
// =============================================================================

type unitRow struct {
	Prefix         string             `json:"prefix"` // /80 CIDR
	Slot           int                `json:"slot"`
	Name           string             `json:"name"`
	Timezone       string             `json:"timezone,omitempty"`
	UnitProperty   int                `json:"unit_property"` // 0=Home, 1=Facility
	UnitType       int                `json:"unit_type"`     // 1/2/3
	CreatedAt      string             `json:"created_at"`
	UpdatedAt      string             `json:"updated_at"`
	BranchPrefix   string             `json:"branch_prefix"`             // /56 (branch 段，byte 6)
	BranchName     string             `json:"branch_name,omitempty"`     // 来自 branches join
	BuildingPrefix string             `json:"building_prefix,omitempty"` // /60 (building 段，byte 7 高 4 bit)；跨 floor 稳定
	BuildingName   string             `json:"building_name,omitempty"`   // 来自 sites.site_name
	FloorPrefix    string             `json:"floor_prefix,omitempty"`    // /64 (= site_id, building<<4|floor)；per-floor 唯一
	FloorName      string             `json:"floor_name,omitempty"`      // FE 显示用 "1F"/"2F" 等
	Floor          int                `json:"floor,omitempty"`           // 来自 sites.floor (1..14)
	Rooms          []unitRoomItem     `json:"rooms,omitempty"`
	Beds           []unitBedItem      `json:"beds,omitempty"`
	Residents      []unitResidentItem `json:"residents,omitempty"` // 当前活跃 binding 到该 unit 的 resident(s)
}

type unitRoomItem struct {
	RoomPrefix string `json:"room_prefix"` // /88
	RoomName   string `json:"room_name"`
	RoomType   int    `json:"room_type,omitempty"` // 0=Default, 1=Bathroom, 2=Kitchen
	IsPrimary  bool   `json:"is_primary"`
}

type unitBedItem struct {
	BedPrefix string `json:"bed_prefix"` // /96
	BedName   string `json:"bed_name"`
}

type unitResidentItem struct {
	ResidentHoA string `json:"resident_hoa"`
	Nickname    string `json:"nickname"`
}

type updateUnitReq struct {
	Name         *string `json:"name,omitempty"`
	Timezone     *string `json:"timezone,omitempty"`      // 空字符串 = NULL
	UnitProperty *int    `json:"unit_property,omitempty"` // 0/1
	UnitType     *int    `json:"unit_type,omitempty"`     // 1/2/3
}

func (h *SpatialHandler) HandleUnits(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.allocUnit(w, r)
	case http.MethodGet:
		h.listUnits(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *SpatialHandler) allocUnit(w http.ResponseWriter, r *http.Request) {
	var body allocUnitReq
	if !decodeJSON(w, r, &body) {
		return
	}
	parent, err := parsePrefix(body.Parent)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid parent: "+err.Error()))
		return
	}
	tenant, _ := parent.Addr().Prefix(48)
	if !h.scopeAllows(r, tenant) {
		writeJSON(w, http.StatusOK, Fail("permission denied: cross-tenant allocation requires SystemAdmin"))
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

// listUnits GET /admin/api/v2/spatial/units?branch=<prefix>[&tenant=<prefix>]
// 优先级 branch > tenant > X-Tenant-Id。
func (h *SpatialHandler) listUnits(w http.ResponseWriter, r *http.Request) {
	scopeStr := r.URL.Query().Get("branch")
	if scopeStr == "" {
		scopeStr = r.URL.Query().Get("tenant")
	}
	if scopeStr == "" {
		scopeStr = r.Header.Get("X-Tenant-Id")
	}
	if scopeStr == "" {
		writeJSON(w, http.StatusBadRequest, Fail("scope required (?branch=, ?tenant=, or X-Tenant-Id header)"))
		return
	}
	scope, err := parsePrefix(scopeStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid scope: "+err.Error()))
		return
	}
	// scope check against caller's tenant
	tenant, _ := scope.Addr().Prefix(48)
	if !h.scopeAllows(r, tenant) {
		writeJSON(w, http.StatusOK, Fail("permission denied: cross-tenant list requires SystemAdmin"))
		return
	}
	// JOIN sites + branches 派生 building_name/floor/branch_name。
	// IPv6 byte 7 = site_slot = (building<<4 | floor)，所以：
	//   /60 = branch + building (跨 floor 稳定，FE building 聚合 key)
	//   /64 = branch + building + floor (= site_id)
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT host(u.unit_id)||'/'||masklen(u.unit_id), u.unit_slot, u.unit_name,
		       COALESCE(u.timezone,''),
		       u.unit_property, u.unit_type,
		       to_char(u.created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       to_char(u.updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       host(network(set_masklen(u.unit_id, 56)))||'/56' AS branch_prefix,
		       COALESCE(b.branch_name, '') AS branch_name,
		       host(network(set_masklen(u.unit_id, 60)))||'/60' AS building_prefix,
		       host(network(set_masklen(u.unit_id, 64)))||'/64' AS floor_prefix,
		       COALESCE(s.site_name, '') AS building_name,
		       COALESCE(s.floor, 0) AS floor
		FROM units u
		LEFT JOIN sites s ON u.unit_id << s.site_id
		LEFT JOIN branches b ON u.unit_id << b.branch_id
		WHERE u.unit_id << $1::INET
		ORDER BY u.unit_name ASC
	`, scope.String())
	if err != nil {
		h.logger.Error("v2 spatial: list units failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("list units failed: "+err.Error()))
		return
	}
	defer rows.Close()
	items := make([]unitRow, 0, 8)
	for rows.Next() {
		var u unitRow
		if err := rows.Scan(&u.Prefix, &u.Slot, &u.Name, &u.Timezone,
			&u.UnitProperty, &u.UnitType, &u.CreatedAt, &u.UpdatedAt, &u.BranchPrefix,
			&u.BranchName, &u.BuildingPrefix, &u.FloorPrefix, &u.BuildingName, &u.Floor); err != nil {
			h.logger.Error("v2 spatial: scan unit row failed", zap.Error(err))
			writeJSON(w, http.StatusOK, Fail("scan unit row failed: "+err.Error()))
			return
		}
		if u.Floor > 0 {
			u.FloorName = fmt.Sprintf("%dF", u.Floor)
		}
		items = append(items, u)
	}
	rows.Close()

	// 联表 rooms / beds / residents
	for i := range items {
		if rs, err := h.fetchUnitRooms(r.Context(), items[i].Prefix); err == nil {
			items[i].Rooms = rs
		} else {
			h.logger.Warn("v2 spatial: fetch unit rooms failed", zap.String("prefix", items[i].Prefix), zap.Error(err))
		}
		if bs, err := h.fetchUnitBeds(r.Context(), items[i].Prefix); err == nil {
			items[i].Beds = bs
		} else {
			h.logger.Warn("v2 spatial: fetch unit beds failed", zap.String("prefix", items[i].Prefix), zap.Error(err))
		}
		if res, err := h.fetchUnitResidents(r.Context(), items[i].Prefix); err == nil {
			items[i].Residents = res
		} else {
			h.logger.Warn("v2 spatial: fetch unit residents failed", zap.String("prefix", items[i].Prefix), zap.Error(err))
		}
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": items,
		"total": len(items),
	}))
}

// HandleUnitByPrefix dispatches GET/PUT/DELETE on /admin/api/v2/spatial/units/{prefix-encoded}
func (h *SpatialHandler) HandleUnitByPrefix(w http.ResponseWriter, r *http.Request) {
	const base = "/admin/api/v2/spatial/units/"
	encoded := strings.TrimPrefix(r.URL.Path, base)
	if encoded == "" {
		writeJSON(w, http.StatusBadRequest, Fail("unit prefix required"))
		return
	}
	prefix, err := decodePrefixSegment(encoded, r.URL.Query().Get("prefix"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid unit prefix: "+err.Error()))
		return
	}
	if prefix.Bits() != 80 {
		writeJSON(w, http.StatusBadRequest, Fail("unit prefix must be /80, got /"+fmt.Sprint(prefix.Bits())))
		return
	}
	tenant, _ := prefix.Addr().Prefix(48)
	if !h.scopeAllows(r, tenant) {
		writeJSON(w, http.StatusOK, Fail("permission denied: cross-tenant access requires SystemAdmin"))
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getUnit(w, r, prefix)
	case http.MethodPut:
		h.updateUnit(w, r, prefix)
	case http.MethodDelete:
		h.deleteUnit(w, r, prefix)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *SpatialHandler) getUnit(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
	var u unitRow
	var tz sql.NullString
	err := h.db.QueryRowContext(r.Context(), `
		SELECT host(u.unit_id)||'/'||masklen(u.unit_id), u.unit_slot, u.unit_name,
		       u.timezone, u.unit_property, u.unit_type,
		       to_char(u.created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       to_char(u.updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       host(network(set_masklen(u.unit_id, 56)))||'/56',
		       COALESCE(b.branch_name, ''),
		       host(network(set_masklen(u.unit_id, 60)))||'/60',
		       host(network(set_masklen(u.unit_id, 64)))||'/64',
		       COALESCE(s.site_name, ''),
		       COALESCE(s.floor, 0)
		FROM units u
		LEFT JOIN sites s ON u.unit_id << s.site_id
		LEFT JOIN branches b ON u.unit_id << b.branch_id
		WHERE u.unit_id = $1::INET
	`, prefix.String()).Scan(&u.Prefix, &u.Slot, &u.Name, &tz, &u.UnitProperty, &u.UnitType,
		&u.CreatedAt, &u.UpdatedAt, &u.BranchPrefix,
		&u.BranchName, &u.BuildingPrefix, &u.FloorPrefix, &u.BuildingName, &u.Floor)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusOK, Fail("unit not found: "+prefix.String()))
		return
	}
	if err != nil {
		h.logger.Error("v2 spatial: get unit failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("get unit failed: "+err.Error()))
		return
	}
	u.Timezone = tz.String
	if u.Floor > 0 {
		u.FloorName = fmt.Sprintf("%dF", u.Floor)
	}
	// 附 children
	if rs, err := h.fetchUnitRooms(r.Context(), u.Prefix); err == nil {
		u.Rooms = rs
	}
	if bs, err := h.fetchUnitBeds(r.Context(), u.Prefix); err == nil {
		u.Beds = bs
	}
	if res, err := h.fetchUnitResidents(r.Context(), u.Prefix); err == nil {
		u.Residents = res
	}
	writeJSON(w, http.StatusOK, Ok(u))
}

func (h *SpatialHandler) updateUnit(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
	var body updateUnitReq
	if !decodeJSON(w, r, &body) {
		return
	}
	sets := make([]string, 0, 5)
	args := make([]any, 0, 6)
	args = append(args, prefix.String())
	idx := 2
	if body.Name != nil {
		if *body.Name == "" {
			writeJSON(w, http.StatusBadRequest, Fail("unit_name cannot be empty"))
			return
		}
		sets = append(sets, fmt.Sprintf("unit_name = $%d", idx))
		args = append(args, *body.Name)
		idx++
	}
	if body.Timezone != nil {
		sets = append(sets, fmt.Sprintf("timezone = NULLIF($%d, '')", idx))
		args = append(args, *body.Timezone)
		idx++
	}
	if body.UnitProperty != nil {
		if *body.UnitProperty != 0 && *body.UnitProperty != 1 {
			writeJSON(w, http.StatusBadRequest, Fail("unit_property must be 0 (non-residential) or 1 (residential)"))
			return
		}
		sets = append(sets, fmt.Sprintf("unit_property = $%d", idx))
		args = append(args, *body.UnitProperty)
		idx++
	}
	if body.UnitType != nil {
		if *body.UnitType < 1 || *body.UnitType > 3 {
			writeJSON(w, http.StatusBadRequest, Fail("unit_type must be 1, 2, or 3"))
			return
		}
		sets = append(sets, fmt.Sprintf("unit_type = $%d", idx))
		args = append(args, *body.UnitType)
		idx++
	}
	if len(sets) == 0 {
		writeJSON(w, http.StatusBadRequest, Fail("no fields to update"))
		return
	}
	sets = append(sets, "updated_at = NOW()")
	q := "UPDATE units SET " + strings.Join(sets, ", ") + " WHERE unit_id = $1::INET"
	res, err := h.db.ExecContext(r.Context(), q, args...)
	if err != nil {
		h.logger.Error("v2 spatial: update unit failed", zap.Error(err), zap.String("prefix", prefix.String()))
		writeJSON(w, http.StatusOK, Fail("update unit failed: "+err.Error()))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, http.StatusOK, Fail("unit not found: "+prefix.String()))
		return
	}
	h.logger.Info("v2 spatial: unit updated", zap.String("prefix", prefix.String()), zap.Int("fields", len(sets)-1))
	writeJSON(w, http.StatusOK, Ok(map[string]any{"prefix": prefix.String(), "updated": true}))
}

// deleteUnit — units 表无 status 列；条件硬删（拒绝 if 有 rooms/beds/active resident binding）。
// 与 v1 BranchList "Empty branch: hard delete / Branch with data: cannot delete" 同语义。
func (h *SpatialHandler) deleteUnit(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
	// children check：rooms + beds + active resident_unit binding
	var roomCount, bedCount, residentCount int
	row := h.db.QueryRowContext(r.Context(), `
		SELECT
		  (SELECT COUNT(*) FROM rooms WHERE room_id << $1::INET),
		  (SELECT COUNT(*) FROM beds  WHERE bed_id  << $1::INET),
		  (SELECT COUNT(*) FROM resident_unit WHERE valid_to IS NULL AND spatial_prefix << $1::INET)
	`, prefix.String())
	if err := row.Scan(&roomCount, &bedCount, &residentCount); err != nil {
		h.logger.Error("v2 spatial: unit child-count failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("unit child-count failed: "+err.Error()))
		return
	}
	if roomCount+bedCount+residentCount > 0 {
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf(
			"cannot delete unit %s: has %d room(s), %d bed(s), %d active resident(s); remove children first",
			prefix.String(), roomCount, bedCount, residentCount)))
		return
	}
	res, err := h.db.ExecContext(r.Context(), `DELETE FROM units WHERE unit_id = $1::INET`, prefix.String())
	if err != nil {
		h.logger.Error("v2 spatial: delete unit failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("delete unit failed: "+err.Error()))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, http.StatusOK, Fail("unit not found: "+prefix.String()))
		return
	}
	h.logger.Info("v2 spatial: unit deleted", zap.String("prefix", prefix.String()))
	writeJSON(w, http.StatusOK, Ok(map[string]any{"prefix": prefix.String(), "deleted": true}))
}

func (h *SpatialHandler) fetchUnitRooms(ctx context.Context, unitPrefix string) ([]unitRoomItem, error) {
	rows, err := h.db.QueryContext(ctx, `
		SELECT host(room_id)||'/'||masklen(room_id), room_name, COALESCE(room_type, 0), is_primary
		FROM rooms WHERE room_id << $1::INET
		ORDER BY is_primary DESC, room_slot ASC
	`, unitPrefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]unitRoomItem, 0, 4)
	for rows.Next() {
		var rm unitRoomItem
		if err := rows.Scan(&rm.RoomPrefix, &rm.RoomName, &rm.RoomType, &rm.IsPrimary); err != nil {
			return nil, err
		}
		out = append(out, rm)
	}
	return out, rows.Err()
}

func (h *SpatialHandler) fetchUnitBeds(ctx context.Context, unitPrefix string) ([]unitBedItem, error) {
	rows, err := h.db.QueryContext(ctx, `
		SELECT host(bed_id)||'/'||masklen(bed_id), bed_name
		FROM beds WHERE bed_id << $1::INET
		ORDER BY bed_slot ASC
	`, unitPrefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]unitBedItem, 0, 4)
	for rows.Next() {
		var b unitBedItem
		if err := rows.Scan(&b.BedPrefix, &b.BedName); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (h *SpatialHandler) fetchUnitResidents(ctx context.Context, unitPrefix string) ([]unitResidentItem, error) {
	rows, err := h.db.QueryContext(ctx, `
		SELECT DISTINCT host(r.resident_id)||'/'||masklen(r.resident_id), r.nickname
		FROM residents r
		JOIN resident_unit ru ON ru.resident_id = r.resident_id AND ru.valid_to IS NULL
		WHERE ru.spatial_prefix << $1::INET AND r.status = 'active'
		ORDER BY r.nickname ASC
	`, unitPrefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]unitResidentItem, 0, 4)
	for rows.Next() {
		var rs unitResidentItem
		if err := rows.Scan(&rs.ResidentHoA, &rs.Nickname); err != nil {
			return nil, err
		}
		out = append(out, rs)
	}
	return out, rows.Err()
}

// =============================================================================
// Rooms — LIST/GET/PUT/DELETE + alloc (POST)
// =============================================================================

type roomRow struct {
	Prefix       string             `json:"prefix"` // /88 CIDR
	Slot         int                `json:"slot"`
	Name         string             `json:"name"`
	RoomType     int                `json:"room_type,omitempty"` // 0=Default, 1=Bathroom, 2=Kitchen
	IsPrimary    bool               `json:"is_primary"`
	Description  string             `json:"description,omitempty"`
	CreatedAt    string             `json:"created_at"`
	UpdatedAt    string             `json:"updated_at"`
	UnitPrefix   string             `json:"unit_prefix"`             // derived /80
	UnitName     string             `json:"unit_name,omitempty"`
	BranchPrefix string             `json:"branch_prefix"`           // derived /56
	BranchName   string             `json:"branch_name,omitempty"`
	Beds         []roomBedItem      `json:"beds,omitempty"`
	Residents    []roomResidentItem `json:"residents,omitempty"`     // 当前活跃绑定到该 room 或其子 bed 的 resident
}

type roomBedItem struct {
	BedPrefix string `json:"bed_prefix"` // /96
	BedName   string `json:"bed_name"`
}

type roomResidentItem struct {
	ResidentHoA string `json:"resident_hoa"`
	Nickname    string `json:"nickname"`
}

type updateRoomReq struct {
	Name        *string `json:"name,omitempty"`
	RoomType    *int    `json:"room_type,omitempty"`   // 0=Default, 1=Bathroom, 2=Kitchen
	IsPrimary   *bool   `json:"is_primary,omitempty"`
	Description *string `json:"description,omitempty"` // 空字符串=NULL
}

func (h *SpatialHandler) HandleRooms(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.allocRoom(w, r)
	case http.MethodGet:
		h.listRooms(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *SpatialHandler) allocRoom(w http.ResponseWriter, r *http.Request) {
	var body allocRoomReq
	if !decodeJSON(w, r, &body) {
		return
	}
	parent, err := parsePrefix(body.Parent)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid parent: "+err.Error()))
		return
	}
	tenant, _ := parent.Addr().Prefix(48)
	if !h.scopeAllows(r, tenant) {
		writeJSON(w, http.StatusOK, Fail("permission denied: cross-tenant allocation requires SystemAdmin"))
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

// listRooms GET /admin/api/v2/spatial/rooms?unit=<prefix>|?branch=<prefix>|?tenant=<prefix>
func (h *SpatialHandler) listRooms(w http.ResponseWriter, r *http.Request) {
	scopeStr := r.URL.Query().Get("unit")
	if scopeStr == "" {
		scopeStr = r.URL.Query().Get("branch")
	}
	if scopeStr == "" {
		scopeStr = r.URL.Query().Get("tenant")
	}
	if scopeStr == "" {
		scopeStr = r.Header.Get("X-Tenant-Id")
	}
	if scopeStr == "" {
		writeJSON(w, http.StatusBadRequest, Fail("scope required (?unit=, ?branch=, ?tenant=, or X-Tenant-Id header)"))
		return
	}
	scope, err := parsePrefix(scopeStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid scope: "+err.Error()))
		return
	}
	tenant, _ := scope.Addr().Prefix(48)
	if !h.scopeAllows(r, tenant) {
		writeJSON(w, http.StatusOK, Fail("permission denied: cross-tenant list requires SystemAdmin"))
		return
	}
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT host(rm.room_id)||'/'||masklen(rm.room_id), rm.room_slot, rm.room_name,
		       COALESCE(rm.room_type, 0), rm.is_primary, COALESCE(rm.description,''),
		       to_char(rm.created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       to_char(rm.updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       host(network(set_masklen(rm.room_id, 80)))||'/80' AS unit_prefix,
		       COALESCE(u.unit_name, '') AS unit_name,
		       host(network(set_masklen(rm.room_id, 56)))||'/56' AS branch_prefix,
		       COALESCE(b.branch_name, '') AS branch_name
		FROM rooms rm
		LEFT JOIN units u    ON rm.room_id << u.unit_id
		LEFT JOIN branches b ON rm.room_id << b.branch_id
		WHERE rm.room_id << $1::INET
		ORDER BY rm.room_name ASC
	`, scope.String())
	if err != nil {
		h.logger.Error("v2 spatial: list rooms failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("list rooms failed: "+err.Error()))
		return
	}
	defer rows.Close()
	items := make([]roomRow, 0, 8)
	for rows.Next() {
		var rm roomRow
		if err := rows.Scan(&rm.Prefix, &rm.Slot, &rm.Name, &rm.RoomType, &rm.IsPrimary,
			&rm.Description, &rm.CreatedAt, &rm.UpdatedAt, &rm.UnitPrefix, &rm.UnitName,
			&rm.BranchPrefix, &rm.BranchName); err != nil {
			h.logger.Error("v2 spatial: scan room row failed", zap.Error(err))
			writeJSON(w, http.StatusOK, Fail("scan room row failed: "+err.Error()))
			return
		}
		items = append(items, rm)
	}
	rows.Close()

	for i := range items {
		if bs, err := h.fetchRoomBeds(r.Context(), items[i].Prefix); err == nil {
			items[i].Beds = bs
		}
		if res, err := h.fetchRoomResidents(r.Context(), items[i].Prefix); err == nil {
			items[i].Residents = res
		}
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": items,
		"total": len(items),
	}))
}

func (h *SpatialHandler) HandleRoomByPrefix(w http.ResponseWriter, r *http.Request) {
	const base = "/admin/api/v2/spatial/rooms/"
	encoded := strings.TrimPrefix(r.URL.Path, base)
	if encoded == "" {
		writeJSON(w, http.StatusBadRequest, Fail("room prefix required"))
		return
	}
	prefix, err := decodePrefixSegment(encoded, r.URL.Query().Get("prefix"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid room prefix: "+err.Error()))
		return
	}
	if prefix.Bits() != 88 {
		writeJSON(w, http.StatusBadRequest, Fail("room prefix must be /88, got /"+fmt.Sprint(prefix.Bits())))
		return
	}
	tenant, _ := prefix.Addr().Prefix(48)
	if !h.scopeAllows(r, tenant) {
		writeJSON(w, http.StatusOK, Fail("permission denied: cross-tenant access requires SystemAdmin"))
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getRoom(w, r, prefix)
	case http.MethodPut:
		h.updateRoom(w, r, prefix)
	case http.MethodDelete:
		h.deleteRoom(w, r, prefix)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *SpatialHandler) getRoom(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
	var rm roomRow
	var rt sql.NullInt32
	var desc, unitName, branchName sql.NullString
	err := h.db.QueryRowContext(r.Context(), `
		SELECT host(rm.room_id)||'/'||masklen(rm.room_id), rm.room_slot, rm.room_name,
		       rm.room_type, rm.is_primary, rm.description,
		       to_char(rm.created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       to_char(rm.updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       host(network(set_masklen(rm.room_id, 80)))||'/80',
		       u.unit_name,
		       host(network(set_masklen(rm.room_id, 56)))||'/56',
		       b.branch_name
		FROM rooms rm
		LEFT JOIN units u    ON rm.room_id << u.unit_id
		LEFT JOIN branches b ON rm.room_id << b.branch_id
		WHERE rm.room_id = $1::INET
	`, prefix.String()).Scan(&rm.Prefix, &rm.Slot, &rm.Name, &rt, &rm.IsPrimary, &desc,
		&rm.CreatedAt, &rm.UpdatedAt, &rm.UnitPrefix, &unitName, &rm.BranchPrefix, &branchName)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusOK, Fail("room not found: "+prefix.String()))
		return
	}
	if err != nil {
		h.logger.Error("v2 spatial: get room failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("get room failed: "+err.Error()))
		return
	}
	rm.RoomType = int(rt.Int32)
	rm.Description = desc.String
	rm.UnitName = unitName.String
	rm.BranchName = branchName.String
	if bs, err := h.fetchRoomBeds(r.Context(), rm.Prefix); err == nil {
		rm.Beds = bs
	}
	if res, err := h.fetchRoomResidents(r.Context(), rm.Prefix); err == nil {
		rm.Residents = res
	}
	writeJSON(w, http.StatusOK, Ok(rm))
}

func (h *SpatialHandler) updateRoom(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
	var body updateRoomReq
	if !decodeJSON(w, r, &body) {
		return
	}
	sets := make([]string, 0, 4)
	args := []any{prefix.String()}
	idx := 2
	if body.Name != nil {
		if *body.Name == "" {
			writeJSON(w, http.StatusBadRequest, Fail("room_name cannot be empty"))
			return
		}
		sets = append(sets, fmt.Sprintf("room_name = $%d", idx))
		args = append(args, *body.Name)
		idx++
	}
	if body.RoomType != nil {
		if *body.RoomType < 0 || *body.RoomType > 2 {
			writeJSON(w, http.StatusBadRequest, Fail("room_type must be 0 (Default), 1 (Bathroom), or 2 (Kitchen)"))
			return
		}
		sets = append(sets, fmt.Sprintf("room_type = $%d", idx))
		args = append(args, *body.RoomType)
		idx++
	}
	if body.IsPrimary != nil {
		sets = append(sets, fmt.Sprintf("is_primary = $%d", idx))
		args = append(args, *body.IsPrimary)
		idx++
	}
	if body.Description != nil {
		sets = append(sets, fmt.Sprintf("description = NULLIF($%d, '')", idx))
		args = append(args, *body.Description)
		idx++
	}
	if len(sets) == 0 {
		writeJSON(w, http.StatusBadRequest, Fail("no fields to update"))
		return
	}
	sets = append(sets, "updated_at = NOW()")
	q := "UPDATE rooms SET " + strings.Join(sets, ", ") + " WHERE room_id = $1::INET"
	res, err := h.db.ExecContext(r.Context(), q, args...)
	if err != nil {
		h.logger.Error("v2 spatial: update room failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("update room failed: "+err.Error()))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, http.StatusOK, Fail("room not found: "+prefix.String()))
		return
	}
	h.logger.Info("v2 spatial: room updated", zap.String("prefix", prefix.String()), zap.Int("fields", len(sets)-1))
	writeJSON(w, http.StatusOK, Ok(map[string]any{"prefix": prefix.String(), "updated": true}))
}

// deleteRoom — rooms 表无 status 列；条件硬删（拒绝 if 有 beds 或 active resident binding）。
func (h *SpatialHandler) deleteRoom(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
	var bedCount, residentCount int
	if err := h.db.QueryRowContext(r.Context(), `
		SELECT
		  (SELECT COUNT(*) FROM beds WHERE bed_id << $1::INET),
		  (SELECT COUNT(*) FROM resident_unit WHERE valid_to IS NULL AND spatial_prefix <<= $1::INET)
	`, prefix.String()).Scan(&bedCount, &residentCount); err != nil {
		h.logger.Error("v2 spatial: room child-count failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("room child-count failed: "+err.Error()))
		return
	}
	if bedCount+residentCount > 0 {
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf(
			"cannot delete room %s: has %d bed(s), %d active resident(s); remove children first",
			prefix.String(), bedCount, residentCount)))
		return
	}
	res, err := h.db.ExecContext(r.Context(), `DELETE FROM rooms WHERE room_id = $1::INET`, prefix.String())
	if err != nil {
		h.logger.Error("v2 spatial: delete room failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("delete room failed: "+err.Error()))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, http.StatusOK, Fail("room not found: "+prefix.String()))
		return
	}
	h.logger.Info("v2 spatial: room deleted", zap.String("prefix", prefix.String()))
	writeJSON(w, http.StatusOK, Ok(map[string]any{"prefix": prefix.String(), "deleted": true}))
}

func (h *SpatialHandler) fetchRoomBeds(ctx context.Context, roomPrefix string) ([]roomBedItem, error) {
	rows, err := h.db.QueryContext(ctx, `
		SELECT host(bed_id)||'/'||masklen(bed_id), bed_name
		FROM beds WHERE bed_id << $1::INET
		ORDER BY bed_slot ASC
	`, roomPrefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]roomBedItem, 0, 4)
	for rows.Next() {
		var b roomBedItem
		if err := rows.Scan(&b.BedPrefix, &b.BedName); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// fetchRoomResidents — resident_unit.spatial_prefix 是 /96(bed) 或 /88(room) 或更高，
// 用 <<= room_prefix 抓所有挂到本 room 或其子 bed 的活跃 resident。
func (h *SpatialHandler) fetchRoomResidents(ctx context.Context, roomPrefix string) ([]roomResidentItem, error) {
	rows, err := h.db.QueryContext(ctx, `
		SELECT DISTINCT host(r.resident_id)||'/'||masklen(r.resident_id), r.nickname
		FROM residents r
		JOIN resident_unit ru ON ru.resident_id = r.resident_id AND ru.valid_to IS NULL
		WHERE ru.spatial_prefix <<= $1::INET AND r.status = 'active'
		ORDER BY r.nickname ASC
	`, roomPrefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]roomResidentItem, 0, 4)
	for rows.Next() {
		var rs roomResidentItem
		if err := rows.Scan(&rs.ResidentHoA, &rs.Nickname); err != nil {
			return nil, err
		}
		out = append(out, rs)
	}
	return out, rows.Err()
}

// =============================================================================
// Beds — LIST/GET/PUT/DELETE + alloc (POST)
// =============================================================================

type bedRow struct {
	Prefix       string            `json:"prefix"` // /96 CIDR
	Slot         int               `json:"slot"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
	RoomPrefix   string            `json:"room_prefix"`             // derived /88
	RoomName     string            `json:"room_name,omitempty"`
	UnitPrefix   string            `json:"unit_prefix"`             // derived /80
	UnitName     string            `json:"unit_name,omitempty"`
	BranchPrefix string            `json:"branch_prefix"`           // derived /56
	BranchName   string            `json:"branch_name,omitempty"`
	Resident     *bedResidentItem  `json:"resident,omitempty"`      // 当前活跃绑定的 resident（resident_unit.spatial_prefix <<= bed）
}

type bedResidentItem struct {
	ResidentHoA string `json:"resident_hoa"`
	Nickname    string `json:"nickname"`
}

type updateBedReq struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (h *SpatialHandler) HandleBeds(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.allocBed(w, r)
	case http.MethodGet:
		h.listBeds(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *SpatialHandler) allocBed(w http.ResponseWriter, r *http.Request) {
	var body allocBedReq
	if !decodeJSON(w, r, &body) {
		return
	}
	parent, err := parsePrefix(body.Parent)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid parent: "+err.Error()))
		return
	}
	tenant, _ := parent.Addr().Prefix(48)
	if !h.scopeAllows(r, tenant) {
		writeJSON(w, http.StatusOK, Fail("permission denied: cross-tenant allocation requires SystemAdmin"))
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

// listBeds GET /admin/api/v2/spatial/beds?room=<prefix>|?unit=<prefix>|?branch=<prefix>|?tenant=<prefix>
func (h *SpatialHandler) listBeds(w http.ResponseWriter, r *http.Request) {
	scopeStr := r.URL.Query().Get("room")
	if scopeStr == "" {
		scopeStr = r.URL.Query().Get("unit")
	}
	if scopeStr == "" {
		scopeStr = r.URL.Query().Get("branch")
	}
	if scopeStr == "" {
		scopeStr = r.URL.Query().Get("tenant")
	}
	if scopeStr == "" {
		scopeStr = r.Header.Get("X-Tenant-Id")
	}
	if scopeStr == "" {
		writeJSON(w, http.StatusBadRequest, Fail("scope required (?room=, ?unit=, ?branch=, ?tenant=, or X-Tenant-Id header)"))
		return
	}
	scope, err := parsePrefix(scopeStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid scope: "+err.Error()))
		return
	}
	tenant, _ := scope.Addr().Prefix(48)
	if !h.scopeAllows(r, tenant) {
		writeJSON(w, http.StatusOK, Fail("permission denied: cross-tenant list requires SystemAdmin"))
		return
	}
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT host(bd.bed_id)||'/'||masklen(bd.bed_id), bd.bed_slot, bd.bed_name,
		       COALESCE(bd.description,''),
		       to_char(bd.created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       to_char(bd.updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       host(network(set_masklen(bd.bed_id, 88)))||'/88' AS room_prefix,
		       COALESCE(rm.room_name, '') AS room_name,
		       host(network(set_masklen(bd.bed_id, 80)))||'/80' AS unit_prefix,
		       COALESCE(u.unit_name, '') AS unit_name,
		       host(network(set_masklen(bd.bed_id, 56)))||'/56' AS branch_prefix,
		       COALESCE(b.branch_name, '') AS branch_name
		FROM beds bd
		LEFT JOIN rooms    rm ON bd.bed_id << rm.room_id
		LEFT JOIN units    u  ON bd.bed_id << u.unit_id
		LEFT JOIN branches b  ON bd.bed_id << b.branch_id
		WHERE bd.bed_id << $1::INET
		ORDER BY bd.bed_slot ASC
	`, scope.String())
	if err != nil {
		h.logger.Error("v2 spatial: list beds failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("list beds failed: "+err.Error()))
		return
	}
	defer rows.Close()
	items := make([]bedRow, 0, 8)
	for rows.Next() {
		var bd bedRow
		if err := rows.Scan(&bd.Prefix, &bd.Slot, &bd.Name, &bd.Description,
			&bd.CreatedAt, &bd.UpdatedAt, &bd.RoomPrefix, &bd.RoomName,
			&bd.UnitPrefix, &bd.UnitName, &bd.BranchPrefix, &bd.BranchName); err != nil {
			h.logger.Error("v2 spatial: scan bed row failed", zap.Error(err))
			writeJSON(w, http.StatusOK, Fail("scan bed row failed: "+err.Error()))
			return
		}
		items = append(items, bd)
	}
	rows.Close()

	for i := range items {
		if res, err := h.fetchBedResident(r.Context(), items[i].Prefix); err == nil {
			items[i].Resident = res
		}
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": items,
		"total": len(items),
	}))
}

func (h *SpatialHandler) HandleBedByPrefix(w http.ResponseWriter, r *http.Request) {
	const base = "/admin/api/v2/spatial/beds/"
	encoded := strings.TrimPrefix(r.URL.Path, base)
	if encoded == "" {
		writeJSON(w, http.StatusBadRequest, Fail("bed prefix required"))
		return
	}
	prefix, err := decodePrefixSegment(encoded, r.URL.Query().Get("prefix"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid bed prefix: "+err.Error()))
		return
	}
	if prefix.Bits() != 96 {
		writeJSON(w, http.StatusBadRequest, Fail("bed prefix must be /96, got /"+fmt.Sprint(prefix.Bits())))
		return
	}
	tenant, _ := prefix.Addr().Prefix(48)
	if !h.scopeAllows(r, tenant) {
		writeJSON(w, http.StatusOK, Fail("permission denied: cross-tenant access requires SystemAdmin"))
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getBed(w, r, prefix)
	case http.MethodPut:
		h.updateBed(w, r, prefix)
	case http.MethodDelete:
		h.deleteBed(w, r, prefix)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *SpatialHandler) getBed(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
	var bd bedRow
	var desc, roomName, unitName, branchName sql.NullString
	err := h.db.QueryRowContext(r.Context(), `
		SELECT host(bd.bed_id)||'/'||masklen(bd.bed_id), bd.bed_slot, bd.bed_name,
		       bd.description,
		       to_char(bd.created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       to_char(bd.updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		       host(network(set_masklen(bd.bed_id, 88)))||'/88',
		       rm.room_name,
		       host(network(set_masklen(bd.bed_id, 80)))||'/80',
		       u.unit_name,
		       host(network(set_masklen(bd.bed_id, 56)))||'/56',
		       b.branch_name
		FROM beds bd
		LEFT JOIN rooms    rm ON bd.bed_id << rm.room_id
		LEFT JOIN units    u  ON bd.bed_id << u.unit_id
		LEFT JOIN branches b  ON bd.bed_id << b.branch_id
		WHERE bd.bed_id = $1::INET
	`, prefix.String()).Scan(&bd.Prefix, &bd.Slot, &bd.Name, &desc,
		&bd.CreatedAt, &bd.UpdatedAt, &bd.RoomPrefix, &roomName,
		&bd.UnitPrefix, &unitName, &bd.BranchPrefix, &branchName)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusOK, Fail("bed not found: "+prefix.String()))
		return
	}
	if err != nil {
		h.logger.Error("v2 spatial: get bed failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("get bed failed: "+err.Error()))
		return
	}
	bd.Description = desc.String
	bd.RoomName = roomName.String
	bd.UnitName = unitName.String
	bd.BranchName = branchName.String
	if res, err := h.fetchBedResident(r.Context(), bd.Prefix); err == nil {
		bd.Resident = res
	}
	writeJSON(w, http.StatusOK, Ok(bd))
}

func (h *SpatialHandler) updateBed(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
	var body updateBedReq
	if !decodeJSON(w, r, &body) {
		return
	}
	sets := make([]string, 0, 2)
	args := []any{prefix.String()}
	idx := 2
	if body.Name != nil {
		if *body.Name == "" {
			writeJSON(w, http.StatusBadRequest, Fail("bed_name cannot be empty"))
			return
		}
		sets = append(sets, fmt.Sprintf("bed_name = $%d", idx))
		args = append(args, *body.Name)
		idx++
	}
	if body.Description != nil {
		sets = append(sets, fmt.Sprintf("description = NULLIF($%d, '')", idx))
		args = append(args, *body.Description)
		idx++
	}
	if len(sets) == 0 {
		writeJSON(w, http.StatusBadRequest, Fail("no fields to update"))
		return
	}
	sets = append(sets, "updated_at = NOW()")
	q := "UPDATE beds SET " + strings.Join(sets, ", ") + " WHERE bed_id = $1::INET"
	res, err := h.db.ExecContext(r.Context(), q, args...)
	if err != nil {
		h.logger.Error("v2 spatial: update bed failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("update bed failed: "+err.Error()))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, http.StatusOK, Fail("bed not found: "+prefix.String()))
		return
	}
	h.logger.Info("v2 spatial: bed updated", zap.String("prefix", prefix.String()), zap.Int("fields", len(sets)-1))
	writeJSON(w, http.StatusOK, Ok(map[string]any{"prefix": prefix.String(), "updated": true}))
}

// deleteBed — beds 表无 status 列；条件硬删（拒绝 if 有 active resident binding 在该 bed）。
func (h *SpatialHandler) deleteBed(w http.ResponseWriter, r *http.Request, prefix netip.Prefix) {
	var residentCount int
	if err := h.db.QueryRowContext(r.Context(), `
		SELECT COUNT(*) FROM resident_unit WHERE valid_to IS NULL AND spatial_prefix <<= $1::INET
	`, prefix.String()).Scan(&residentCount); err != nil {
		h.logger.Error("v2 spatial: bed child-count failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("bed child-count failed: "+err.Error()))
		return
	}
	if residentCount > 0 {
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf(
			"cannot delete bed %s: has %d active resident(s); unassign first",
			prefix.String(), residentCount)))
		return
	}
	res, err := h.db.ExecContext(r.Context(), `DELETE FROM beds WHERE bed_id = $1::INET`, prefix.String())
	if err != nil {
		h.logger.Error("v2 spatial: delete bed failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("delete bed failed: "+err.Error()))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, http.StatusOK, Fail("bed not found: "+prefix.String()))
		return
	}
	h.logger.Info("v2 spatial: bed deleted", zap.String("prefix", prefix.String()))
	writeJSON(w, http.StatusOK, Ok(map[string]any{"prefix": prefix.String(), "deleted": true}))
}

func (h *SpatialHandler) fetchBedResident(ctx context.Context, bedPrefix string) (*bedResidentItem, error) {
	var rs bedResidentItem
	err := h.db.QueryRowContext(ctx, `
		SELECT host(r.resident_id)||'/'||masklen(r.resident_id), r.nickname
		FROM residents r
		JOIN resident_unit ru ON ru.resident_id = r.resident_id AND ru.valid_to IS NULL
		WHERE ru.spatial_prefix <<= $1::INET AND r.status = 'active'
		LIMIT 1
	`, bedPrefix).Scan(&rs.ResidentHoA, &rs.Nickname)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // no resident on this bed
	}
	if err != nil {
		return nil, err
	}
	return &rs, nil
}

// HandleDevices 注册 device + 可选 DDNS 推 BIND。
func (h *SpatialHandler) HandleDevices(w http.ResponseWriter, r *http.Request) {
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
func (h *SpatialHandler) HandleLookupTenant(w http.ResponseWriter, r *http.Request) {
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
func (h *SpatialHandler) scopeAllows(r *http.Request, scope netip.Prefix) bool {
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
func (h *SpatialHandler) respondAllocError(w http.ResponseWriter, op string, err error) {
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
