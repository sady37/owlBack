// Package service — Resident V2 Service (Forward Design)
//
// 业务规则集中点；不依赖 v1 ResidentService。
package service

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/scope"

	"owl-common/card"
	"owl-common/ddns"
	"owl-common/spatial"

	"go.uber.org/zap"
)

type ResidentService struct {
	repo     *repository.PostgresResidentsRepository
	cardRepo *repository.PostgresCardRepository
	pub      *publisher.ConfigPublisher
	ddns     *ddns.Client
	owlDom   string
	logger   *zap.Logger
}

func NewResidentService(repo *repository.PostgresResidentsRepository) *ResidentService {
	return &ResidentService{repo: repo, logger: zap.NewNop(), owlDom: "owl."}
}

func (s *ResidentService) SetLogger(logger *zap.Logger) {
	if logger != nil {
		s.logger = logger
	}
}

// SetCardDeps 注入 card 写路径相关依赖（main.go 装配）。
//
//	cardRepo / pub 任一为 nil 则跳过 card 写入（保持 ResidentService 主路径可独立测试）；
//	ddnsClient 为 nil 时跳过 DNS 注册（仅写 DB card），owl Domain 默认 "owl."。
func (s *ResidentService) SetCardDeps(cardRepo *repository.PostgresCardRepository, pub *publisher.ConfigPublisher, ddnsClient *ddns.Client, owlDomain string) {
	s.cardRepo = cardRepo
	s.pub = pub
	s.ddns = ddnsClient
	if strings.TrimSpace(owlDomain) != "" {
		s.owlDom = owlDomain
		if !strings.HasSuffix(s.owlDom, ".") {
			s.owlDom += "."
		}
	}
}

// ============================================================================
// 业务规则 — 角色权限
// ============================================================================
//
// Admin / Manager   : full CRUD
// Nurse             : 改 resident 所有字段；不能改 admission/discharge (move_in/out_date)
// Resident          : 仅可 GET 自己 hoa
// SystemAdmin       : 平台级，跨 tenant 只读

const (
	roleAdmin       = "Admin"
	roleManager     = "Manager"
	roleNurse       = "Nurse"
	roleCaregiver   = "Caregiver"
	roleResident    = "Resident"
	roleFamily      = "Family"
	roleSystemAdmin = "SystemAdmin"
)

// isFamilyRole — Family role 大小写兼容判断
func isFamilyRole(role string) bool {
	return strings.EqualFold(role, roleFamily)
}

func canCRUDResident(role string) bool {
	switch role {
	case roleAdmin, roleManager, "tenant_admin", "manager":
		return true
	}
	return false
}

func canEditResident(role string) bool {
	if canCRUDResident(role) {
		return true
	}
	return role == roleNurse || role == "nurse"
}

// canAppendNote — Caregiver 仅可写 note 字段（交班簿）；其余 resident 字段不可改。
// 用于 update path 的"note-only"放宽：除了 canEditResident 的角色之外，Caregiver 也通过。
func canAppendNote(role string) bool {
	if canEditResident(role) {
		return true
	}
	return role == roleCaregiver || role == "caregiver"
}

// isNoteOnlyUpdate — Update input 中除 Note 外其它"resident-level"字段必须全 nil。
// Caregiver 提交 PUT 时若含其它字段（哪怕值没变也是恶意 curl 风险），直接拒。
func isNoteOnlyUpdate(in *domain.ResidentUpdateInput) bool {
	if in == nil {
		return false
	}
	if in.Note == nil {
		return false
	}
	return in.Nickname == nil &&
		in.ResidentAccount == nil &&
		in.Status == nil &&
		in.ServiceLevel == nil &&
		in.AdmissionDate == nil &&
		in.DischargeDate == nil &&
		in.FamilyAccess == nil &&
		in.UnitID == nil &&
		in.RoomID == nil &&
		in.BedID == nil &&
		in.CaregiverUserIDs == nil &&
		in.CareTeamIDs == nil &&
		in.FamilyUserIDs == nil &&
		in.PHI == nil &&
		in.Contacts == nil
}

func canEditAdmissionDischarge(role string) bool {
	// Nurse 不能改 admission/discharge（涉及财务）
	return canCRUDResident(role)
}

// ============================================================================
// Read
// ============================================================================

type ListResidentsV2Request struct {
	TenantPrefix    string
	CurrentUserID   string // user_id UUID（Family scope 必填）
	CurrentUserHOA  string
	CurrentUserRole string
	Filter          domain.ResidentListFilter
}

type ListResidentsV2Response struct {
	Items []*domain.Resident `json:"items"`
	Total int                  `json:"total"`
}

func (s *ResidentService) List(ctx context.Context, req ListResidentsV2Request) (*ListResidentsV2Response, error) {
	if req.TenantPrefix == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	sc := scope.MustFromContext(ctx)

	// Resident 只看自己（middleware 解析 hoa；fallback 兼容旧 caller 没注入 ctx 的场景）
	hoa := ""
	if sc != nil && sc.IsResident() {
		hoa = sc.HoA
	} else if req.CurrentUserRole == roleResident {
		hoa = req.CurrentUserHOA
	}
	if hoa != "" {
		one, err := s.repo.GetResident(ctx, req.TenantPrefix, hoa)
		if err != nil {
			return &ListResidentsV2Response{Items: []*domain.Resident{}, Total: 0}, nil
		}
		return &ListResidentsV2Response{Items: []*domain.Resident{&one.Resident}, Total: 1}, nil
	}

	// Family 走 resident_caregivers link
	if (sc != nil && sc.IsFamily()) || isFamilyRole(req.CurrentUserRole) {
		uid := ""
		if sc != nil {
			uid = sc.UserID
		}
		if uid == "" {
			uid = req.CurrentUserID
		}
		if strings.TrimSpace(uid) == "" {
			return nil, fmt.Errorf("permission denied: Family role requires authenticated user_id")
		}
		req.Filter.FamilyUserID = uid
	} else if sc != nil && sc.IsStaffBranchScoped() {
		// Staff (Manager/Nurse/Caregiver) 按 Current Branch 严格过滤
		if sc.HasCurrentBranch() {
			req.Filter.BranchPrefix = sc.CurrentBranchID
		} else {
			// 无 current branch → 看空
			return &ListResidentsV2Response{Items: []*domain.Resident{}, Total: 0}, nil
		}
	}
	// 兜底兼容：ctx 没注入 scope 但调用方传了旧字段 — 继续保留旧行为防回归
	if sc == nil && !isFamilyRole(req.CurrentUserRole) &&
		req.CurrentUserRole != roleAdmin && req.CurrentUserRole != roleSystemAdmin {
		if cb, err := s.repo.GetCurrentBranchID(ctx, req.CurrentUserID); err == nil && cb != "" {
			req.Filter.BranchPrefix = cb
		}
	}

	// Email/Phone 搜索：PHI 加密存，DB LIKE 无法直接命中
	// 策略：清掉 DB 端 search、扩大 PageSize 拉满 scope 内 resident，再逐条 GetResident 解密 PHI+contacts 匹配
	origSearch := strings.TrimSpace(req.Filter.Search)
	postFilter := SearchTypeUnknown
	if origSearch != "" {
		postFilter = ClassifySearch(origSearch)
	}
	if postFilter == SearchTypeEmail || postFilter == SearchTypePhone {
		req.Filter.Search = ""
		if req.Filter.PageSize < 10000 {
			req.Filter.PageSize = 10000
		}
	}

	items, total, err := s.repo.ListResidents(ctx, req.TenantPrefix, req.Filter)
	if err != nil {
		return nil, err
	}

	if postFilter == SearchTypeEmail || postFilter == SearchTypePhone {
		items = s.filterByPHIContact(ctx, req.TenantPrefix, items, origSearch, postFilter)
		total = len(items)
	}
	return &ListResidentsV2Response{Items: items, Total: total}, nil
}

// filterByPHIContact 加载每个 resident 的 PHI + contacts（解密），按 email/phone 子串匹配过滤。
// 命中条件：resident.PHI.{Email,Phone} 或 任一 contact.{ContactEmail,ContactPhone} 子串命中。
// Email 比较忽略大小写；Phone 比较仅看数字字符（忽略 +/-/空格）。
func (s *ResidentService) filterByPHIContact(
	ctx context.Context, tenantPrefix string,
	items []*domain.Resident, needle string, st SearchType,
) []*domain.Resident {
	needleLow := strings.ToLower(strings.TrimSpace(needle))
	needleDigits := digitsOnly(needle)
	out := make([]*domain.Resident, 0, len(items))
	for _, it := range items {
		d, err := s.repo.GetResident(ctx, tenantPrefix, it.ResidentID)
		if err != nil || d == nil {
			continue
		}
		hit := false
		switch st {
		case SearchTypeEmail:
			if d.PHI != nil && d.PHI.ResidentEmail != nil &&
				strings.Contains(strings.ToLower(*d.PHI.ResidentEmail), needleLow) {
				hit = true
			}
			if !hit {
				for _, c := range d.Contacts {
					if c.ContactEmail != nil &&
						strings.Contains(strings.ToLower(*c.ContactEmail), needleLow) {
						hit = true
						break
					}
				}
			}
		case SearchTypePhone:
			if needleDigits == "" {
				continue
			}
			if d.PHI != nil && d.PHI.ResidentPhone != nil &&
				strings.Contains(digitsOnly(*d.PHI.ResidentPhone), needleDigits) {
				hit = true
			}
			if !hit {
				for _, c := range d.Contacts {
					if c.ContactPhone != nil &&
						strings.Contains(digitsOnly(*c.ContactPhone), needleDigits) {
						hit = true
						break
					}
				}
			}
		}
		if hit {
			out = append(out, it)
		}
	}
	return out
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

type GetResidentRequest struct {
	TenantPrefix    string
	HoA             string
	CurrentUserID   string // user_id UUID（Family scope 必填）
	CurrentUserHOA  string
	CurrentUserRole string
}

func (s *ResidentService) Get(ctx context.Context, req GetResidentRequest) (*domain.ResidentDetail, error) {
	if req.TenantPrefix == "" || req.HoA == "" {
		return nil, fmt.Errorf("tenant_id and hoa required")
	}
	// 优先用 scope.FromContext 统一校验；ctx 没注入时 fallback 旧字段
	if sc := scope.MustFromContext(ctx); sc != nil {
		if err := sc.VerifyResident(ctx, s.repo.DB(), req.HoA); err != nil {
			return nil, err
		}
	} else {
		// fallback：兼容旧 caller / 测试场景
		if req.CurrentUserRole == roleResident && req.CurrentUserHOA != req.HoA {
			return nil, fmt.Errorf("permission denied: can only view own profile")
		}
		if isFamilyRole(req.CurrentUserRole) {
			if strings.TrimSpace(req.CurrentUserID) == "" {
				return nil, fmt.Errorf("permission denied: Family role requires authenticated user_id")
			}
			linked, err := s.repo.IsResidentLinkedToFamily(ctx, req.HoA, req.CurrentUserID)
			if err != nil {
				return nil, err
			}
			if !linked {
				return nil, fmt.Errorf("permission denied: not linked as family to this resident")
			}
		}
	}
	return s.repo.GetResident(ctx, req.TenantPrefix, req.HoA)
}

// ============================================================================
// Write
// ============================================================================

type CreateResidentRequest struct {
	TenantPrefix    string
	ActorUserID     string // 操作者 user_id UUID（audit_log 用）
	CurrentUserRole string
	Input           *domain.ResidentCreateInput
}

func (s *ResidentService) Create(ctx context.Context, req CreateResidentRequest) (string, error) {
	if !canCRUDResident(req.CurrentUserRole) {
		return "", fmt.Errorf("permission denied: only Admin/Manager can create residents")
	}
	if req.Input == nil {
		return "", fmt.Errorf("input required")
	}
	hoa, err := s.repo.CreateResident(ctx, req.TenantPrefix, req.Input, req.ActorUserID, req.CurrentUserRole)
	if err != nil {
		return hoa, err
	}
	// Phase F — admission：物化 active_bed card + DDNS register
	s.syncCardForResident(ctx, req.TenantPrefix, hoa, "" /* oldPrefix=none */)
	return hoa, nil
}

type UpdateResidentRequest struct {
	TenantPrefix    string
	HoA             string
	ActorUserID     string // 操作者 user_id UUID（audit_log 用）
	CurrentUserRole string
	Input           *domain.ResidentUpdateInput
}

func (s *ResidentService) Update(ctx context.Context, req UpdateResidentRequest) error {
	s.logger.Info("[ResidentService.Update ENTRY]", zap.String("hoa", req.HoA), zap.String("tenant_prefix", req.TenantPrefix), zap.String("role", req.CurrentUserRole))
	// Family 不能改 resident profile（PHI 写入权限不下放给家属）
	if isFamilyRole(req.CurrentUserRole) {
		return fmt.Errorf("permission denied: Family role cannot edit resident profile")
	}
	if req.Input == nil {
		s.logger.Error("[ResidentService.Update FAILED] input nil")
		return fmt.Errorf("input required")
	}
	// 权限分流：
	//   Admin / Manager / Nurse → 走 canEditResident（完整 update；Nurse 受 admission/discharge gate 约束）
	//   Caregiver               → 走 canAppendNote 但必须 isNoteOnlyUpdate（仅 note 字段，其余字段必须 nil）
	if !canEditResident(req.CurrentUserRole) {
		if canAppendNote(req.CurrentUserRole) && isNoteOnlyUpdate(req.Input) {
			s.logger.Info("[ResidentService.Update] Caregiver note-only update permitted",
				zap.String("role", req.CurrentUserRole))
		} else {
			s.logger.Error("[ResidentService.Update FAILED] permission denied",
				zap.String("role", req.CurrentUserRole),
				zap.Bool("note_only_check", isNoteOnlyUpdate(req.Input)))
			return fmt.Errorf("permission denied: role %q cannot edit residents (caregiver can only append note)", req.CurrentUserRole)
		}
	}
	// Nurse 不能改 admission/discharge
	if !canEditAdmissionDischarge(req.CurrentUserRole) {
		if req.Input.AdmissionDate != nil || req.Input.DischargeDate != nil {
			s.logger.Error("[ResidentService.Update FAILED] Nurse cannot modify dates")
			return fmt.Errorf("permission denied: Nurse cannot modify admission/discharge dates (financial)")
		}
	}
	// 抓 oldPrefix → 让 syncCard 检测转床/出院（admission/transfer/discharge 三态）
	oldPrefix, _ := s.repo.GetActiveSpatialPrefix(ctx, req.HoA)
	err := s.repo.UpdateResident(ctx, req.TenantPrefix, req.HoA, req.Input, req.ActorUserID, req.CurrentUserRole)
	if err != nil {
		s.logger.Error("[ResidentService.Update FAILED]", zap.String("hoa", req.HoA), zap.Error(err))
		return err
	}
	s.syncCardForResident(ctx, req.TenantPrefix, req.HoA, oldPrefix)
	s.logger.Info("[ResidentService.Update SUCCESS]", zap.String("hoa", req.HoA))
	return nil
}

type DeleteResidentRequest struct {
	TenantPrefix    string
	HoA             string
	CurrentUserRole string
	Hard            bool // true → 硬删 (Clear)，需 CheckClearable
}

func (s *ResidentService) Delete(ctx context.Context, req DeleteResidentRequest) error {
	if !canCRUDResident(req.CurrentUserRole) {
		return fmt.Errorf("permission denied: only Admin/Manager can delete residents")
	}
	oldPrefix, _ := s.repo.GetActiveSpatialPrefix(ctx, req.HoA)
	var err error
	if req.Hard {
		err = s.repo.HardDelete(ctx, req.HoA)
	} else {
		err = s.repo.SoftDelete(ctx, req.HoA)
	}
	if err != nil {
		return err
	}
	// 删除/软删 = 出院：清空当前 card 的 resident_id；card 本身保留
	s.syncCardForResident(ctx, req.TenantPrefix, req.HoA, oldPrefix)
	return nil
}

func (s *ResidentService) CheckClearable(ctx context.Context, hoa string) (*domain.ResidentClearCheckResult, error) {
	return s.repo.CheckClearable(ctx, hoa)
}

// ============================================================================
// 转院（discharge + admission）— 业务规则：跨 tenant 不 in-place 改 hoa
// ============================================================================
//
// 流程：
//   step 1: discharge from A — 当前 tenant set status='discharged' + move_out_date=NOW
//   step 2: admission to B — 在 B tenant Create 新 resident（新 hoa），可携带 PHI / contacts / caregivers
// 这是 2 次 API 调用（FE 引导 user），后端不提供 atomic transfer。
// 当前 V2 service 暂不暴露 transfer endpoint；FE 用 Update + Create 两步组合。

func (s *ResidentService) Discharge(ctx context.Context, req UpdateResidentRequest) error {
	if !canCRUDResident(req.CurrentUserRole) {
		return fmt.Errorf("permission denied")
	}
	oldPrefix, _ := s.repo.GetActiveSpatialPrefix(ctx, req.HoA)
	now := nowDateString()
	in := &domain.ResidentUpdateInput{
		Status:      strPtr("discharged"),
		DischargeDate: &now,
	}
	if err := s.repo.UpdateResident(ctx, req.TenantPrefix, req.HoA, in, req.ActorUserID, req.CurrentUserRole); err != nil {
		return err
	}
	s.syncCardForResident(ctx, req.TenantPrefix, req.HoA, oldPrefix)
	return nil
}

// ============================================================================
// Phase F — cards 物化（admission/discharge/transfer）
// ============================================================================
//
// 触发：Create / Update / Discharge / Delete 完成后调用 syncCardForResident。
//
// 业务规则（[doc/cards_v2_migration_checklist.md] § 一.6 / § 一.4）：
//
//	state 转移                 cards 表动作                       DDNS                    publish
//	---------------------------------------------------------------------------------------------
//	old=""  + new=prefix       INSERT (UPSERT) active_bed card    register AAAA + PTR    card.resident_changed (admission)
//	old=A   + new=A            no-op                              -                       -
//	old=A   + new=B (B!=A)     UPDATE A.resident_id=NULL +         register B (new)        2× (discharge A, admission B)
//	                            INSERT/UPSERT B
//	old=A   + new=""           UPDATE A.resident_id=NULL          -                       card.resident_changed (discharge)
//
// 注意：card 自身的删除条件 = "spatial_prefix 下所有 device 移除"，与 resident 解耦
// 因此 discharge 不删 card（仅清 resident_id 指针），space-card 保留供日后再入住直接复用。
func (s *ResidentService) syncCardForResident(ctx context.Context, tenantPrefix, hoa, oldPrefix string) {
	if s.cardRepo == nil || s.pub == nil {
		return // deps 未装配 → 跳过（保持 service 主路径可独立测试）
	}
	newPrefix, _ := s.repo.GetActiveSpatialPrefix(ctx, hoa)
	if newPrefix == oldPrefix {
		return // no spatial change
	}

	// 1) 旧 prefix → 清空 resident_id（保留 card 本身）
	if oldPrefix != "" {
		if cardID, err := s.cardRepo.GetResidentCardIDByPrefix(ctx, oldPrefix); err == nil && cardID != "" {
			empty := ""
			if err := s.cardRepo.UpdateCard(cardID, nil, &empty, nil); err != nil {
				s.logger.Warn("syncCard: clear old card resident_id failed",
					zap.String("card_id", cardID), zap.String("old_prefix", oldPrefix), zap.Error(err))
			} else {
				op := "discharge"
				if newPrefix != "" {
					op = "transfer"
				}
				_ = s.pub.PublishCardResidentChanged(ctx, tenantPrefix, cardID, op, hoa, "", oldPrefix)
				s.logger.Info("syncCard: cleared resident_id from old card",
					zap.String("card_id", cardID), zap.String("old_prefix", oldPrefix),
					zap.String("op", op))
			}
		}
	}

	// 2) 新 prefix → 物化 active_bed card + DDNS register
	if newPrefix != "" {
		s.materializeActiveBedCard(ctx, tenantPrefix, newPrefix, hoa, oldPrefix)
	}
}

// materializeActiveBedCard upserts an active_bed card for the given spatial_prefix
// 并尝试 DDNS register（best-effort，失败仅 log warn 不阻塞业务）。
func (s *ResidentService) materializeActiveBedCard(ctx context.Context, tenantPrefix, prefixStr, hoa, oldPrefix string) {
	prefix, err := netip.ParsePrefix(prefixStr)
	if err != nil {
		s.logger.Warn("syncCard: invalid spatial_prefix", zap.String("prefix", prefixStr), zap.Error(err))
		return
	}
	shortName, err := ddns.CardShortName(prefix)
	if err != nil {
		s.logger.Warn("syncCard: cardShortName derive failed",
			zap.String("prefix", prefixStr), zap.Error(err))
		// 仍然尝试 DB 写入，DNS 跳过
	}

	// card_name 用 resident.nickname；查不到 fallback 用 dns_short_name
	cardName := s.lookupResidentNickname(ctx, tenantPrefix, hoa)
	if cardName == "" {
		cardName = shortName
	}

	// card_type 由 prefix masklen 决定（/96=active_bed, /88=room, /80=unit, ...）
	cardType := card.CardTypeForMasklen(prefix.Bits())
	if cardType == "" {
		s.logger.Warn("syncCard: unsupported masklen — skipping card materialize",
			zap.String("prefix", prefixStr), zap.Int("bits", prefix.Bits()))
		return
	}

	cardID, err := s.cardRepo.CreateCard(prefixStr, cardType, cardName, shortName, &hoa)
	if err != nil {
		s.logger.Error("syncCard: CreateCard failed",
			zap.String("prefix", prefixStr), zap.String("hoa", hoa), zap.Error(err))
		return
	}
	op := "admission"
	prevHoA := ""
	if oldPrefix != "" {
		op = "transfer"
		prevHoA = hoa // 同一个 resident 转床，prev/new HoA 相同；体现 prefix 变化
	}
	_ = s.pub.PublishCardResidentChanged(ctx, tenantPrefix, cardID, op, prevHoA, hoa, prefixStr)

	// DDNS register（best-effort）
	if s.ddns != nil && shortName != "" {
		// tenant_slot 取自 prefix 第 6 byte（spatial.SlotsOf）
		tenantSlot, _, _, _, _, _, _, derr := spatial.SlotsOf(prefix.Addr())
		if derr != nil {
			s.logger.Warn("syncCard: DDNS skipped — slot derive failed",
				zap.String("prefix", prefixStr), zap.Error(derr))
			return
		}
		zone := ddns.ZoneForTenant(tenantSlot, s.owlDom)
		if err := s.ddns.RegisterCardName(ctx, prefix, shortName, zone); err != nil {
			s.logger.Warn("syncCard: DDNS RegisterCardName failed",
				zap.String("prefix", prefixStr), zap.String("zone", zone), zap.Error(err))
		} else {
			s.logger.Info("syncCard: DDNS registered",
				zap.String("fqdn", shortName+"."+zone), zap.String("prefix", prefixStr))
		}
	}
}

// lookupResidentNickname 尽量取 resident.nickname 作为 card_name；查不到返回 ""
func (s *ResidentService) lookupResidentNickname(ctx context.Context, tenantPrefix, hoa string) string {
	d, err := s.repo.GetResident(ctx, tenantPrefix, hoa)
	if err != nil || d == nil {
		return ""
	}
	return strings.TrimSpace(d.Nickname)
}

// ============================================================================
// helpers
// ============================================================================

func strPtr(s string) *string { return &s }
func nowDateString() string {
	// 不引入 time package 在文件里仅为了 Format — 用 PG NOW() 比较安全。
	// 这里返空字符串让 repo 用 NULL；如要精确，改 input 生成 time.Now().Format("2006-01-02") 即可。
	return strings.TrimSpace("") // 占位；调用方应自填
}
