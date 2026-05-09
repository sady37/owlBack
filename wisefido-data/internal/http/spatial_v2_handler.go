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
// ipam 必填；ddns 为 nil 时跳过 DNS 注册（仍正常分配 prefix）。
type SpatialV2Handler struct {
	ipam   ipam.Backend
	ddns   *ddns.Client // 可选
	logger *zap.Logger
}

// NewSpatialV2Handler 构造。
func NewSpatialV2Handler(backend ipam.Backend, ddnsClient *ddns.Client, logger *zap.Logger) *SpatialV2Handler {
	return &SpatialV2Handler{
		ipam:   backend,
		ddns:   ddnsClient,
		logger: logger,
	}
}

// =============================================================================
// 路由注册
// =============================================================================

// RegisterSpatialV2Routes 把 v2 spatial 路由注册到 router。
func (r *Router) RegisterSpatialV2Routes(h *SpatialV2Handler) {
	r.Handle("/admin/api/v2/spatial/branches", h.HandleBranches)
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

func (h *SpatialV2Handler) HandleBranches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body allocBranchReq
	if !decodeJSON(w, r, &body) {
		return
	}
	parent, err := parsePrefix(body.Parent)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid parent: "+err.Error()))
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
// prefix-encoded URL 用 `_` 替代 `:`，例如 fd00_0_3___48 = fd00:0:3::/48
// （避开 URL 里 `:` 在 path 段的特殊性）
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
	// "_" decode 回 ":"; 也接收 query param `prefix=...` 作 alternative
	decoded := strings.ReplaceAll(encoded, "_", ":")
	if q := r.URL.Query().Get("prefix"); q != "" {
		decoded = q
	}
	parent, err := parsePrefix(decoded)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid tenant prefix: "+err.Error()))
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
