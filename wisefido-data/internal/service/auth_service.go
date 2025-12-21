package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// AuthService 认证授权服务接口
type AuthService interface {
	// 登录功能
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	
	// 搜索机构功能
	SearchInstitutions(ctx context.Context, req SearchInstitutionsRequest) (*SearchInstitutionsResponse, error)
	
	// 密码重置功能（待实现）
	SendVerificationCode(ctx context.Context, req SendVerificationCodeRequest) (*SendVerificationCodeResponse, error)
	VerifyCode(ctx context.Context, req VerifyCodeRequest) (*VerifyCodeResponse, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) (*ResetPasswordResponse, error)
}

// verificationCodeStore 验证码存储（内存实现，后续可以改为 Redis）
type verificationCodeStore struct {
	mu    sync.RWMutex
	codes map[string]verificationCodeData // key: account:userType:tenantName -> data
}

type verificationCodeData struct {
	Code      string
	ExpiresAt time.Time
}

type resetTokenStore struct {
	mu    sync.RWMutex
	tokens map[string]resetTokenData // key: token -> data
}

type resetTokenData struct {
	Account    string
	UserType   string
	TenantID   string
	TenantName string
	ExpiresAt  time.Time
}

// authService 实现
type authService struct {
	authRepo     repository.AuthRepository
	tenantsRepo  repository.TenantsRepository
	db           *sql.DB // 用于验证码和重置密码功能（需要直接查询数据库）
	logger       *zap.Logger
	codeStore    *verificationCodeStore
	tokenStore   *resetTokenStore
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(authRepo repository.AuthRepository, tenantsRepo repository.TenantsRepository, db *sql.DB, logger *zap.Logger) AuthService {
	return &authService{
		authRepo:    authRepo,
		tenantsRepo: tenantsRepo,
		db:          db,
		logger:      logger,
		codeStore: &verificationCodeStore{
			codes: make(map[string]verificationCodeData),
		},
		tokenStore: &resetTokenStore{
			tokens: make(map[string]resetTokenData),
		},
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	TenantID     string // 可选，如果为空则自动解析
	UserType     string // "staff" | "resident"，默认为 "staff"
	AccountHash  string // SHA256(account) 的 hex 编码，必填
	PasswordHash string // SHA256(password) 的 hex 编码，必填
	IPAddress    string // 客户端 IP（用于日志）
	UserAgent    string // 客户端 User-Agent（用于日志）
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string  `json:"accessToken"`  // 访问令牌（占位符）
	RefreshToken string  `json:"refreshToken"` // 刷新令牌（占位符）
	UserID       string  `json:"userId"`       // 用户 ID
	UserAccount  string  `json:"user_account"` // 用户账号
	UserType     string  `json:"userType"`     // 用户类型
	Role         string  `json:"role"`         // 角色
	NickName     string  `json:"nickName"`     // 昵称
	TenantID     string  `json:"tenant_id"`    // 租户 ID
	TenantName   string  `json:"tenant_name"`  // 租户名称
	Domain       string  `json:"domain"`        // 域名
	HomePath     string  `json:"homePath"`      // 首页路径
	BranchTag    *string `json:"branchTag,omitempty"` // 分支标签（可选）
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// 1. 参数验证和规范化
	req.AccountHash = strings.TrimSpace(req.AccountHash)
	req.PasswordHash = strings.TrimSpace(req.PasswordHash)
	if req.AccountHash == "" {
		s.logger.Warn("User login failed: missing credentials",
			zap.String("ip_address", req.IPAddress),
			zap.String("user_agent", req.UserAgent),
			zap.String("reason", "missing_credentials"),
			zap.String("missing_field", "accountHash"),
		)
		return nil, fmt.Errorf("missing credentials")
	}
	if req.PasswordHash == "" {
		s.logger.Warn("User login failed: missing credentials",
			zap.String("ip_address", req.IPAddress),
			zap.String("user_agent", req.UserAgent),
			zap.String("reason", "missing_credentials"),
			zap.String("missing_field", "passwordHash"),
		)
		return nil, fmt.Errorf("missing credentials")
	}

	normalizedUserType := strings.ToLower(strings.TrimSpace(req.UserType))
	if normalizedUserType == "" {
		normalizedUserType = "staff"
	}

	// 2. Hash 解码和验证
	accountHashBytes, err := hex.DecodeString(req.AccountHash)
	if err != nil || len(accountHashBytes) == 0 {
		s.logger.Warn("User login failed: invalid account hash format",
			zap.String("ip_address", req.IPAddress),
			zap.String("user_agent", req.UserAgent),
			zap.String("reason", "invalid_account_hash"),
			zap.Error(err),
		)
		return nil, fmt.Errorf("invalid credentials")
	}

	passwordHashBytes, err := hex.DecodeString(req.PasswordHash)
	if err != nil || len(passwordHashBytes) == 0 {
		s.logger.Warn("User login failed: invalid password hash format",
			zap.String("ip_address", req.IPAddress),
			zap.String("user_agent", req.UserAgent),
			zap.String("reason", "invalid_password_hash"),
			zap.Error(err),
		)
		return nil, fmt.Errorf("invalid credentials")
	}

	// 3. Tenant ID 自动解析（如果为空）
	tenantID := req.TenantID
	if tenantID == "" {
		var matches []repository.TenantLoginMatch
		var err error
		if normalizedUserType == "resident" {
			matches, err = s.authRepo.SearchTenantsForResidentLogin(ctx, accountHashBytes, passwordHashBytes)
		} else {
			matches, err = s.authRepo.SearchTenantsForUserLogin(ctx, accountHashBytes, passwordHashBytes)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to resolve tenant: %w", err)
		}

		if len(matches) == 0 {
			s.logger.Warn("User login failed: invalid credentials",
				zap.String("ip_address", req.IPAddress),
				zap.String("user_agent", req.UserAgent),
				zap.String("user_type", normalizedUserType),
				zap.String("reason", "invalid_credentials"),
				zap.String("note", "no matching tenant found"),
			)
			return nil, fmt.Errorf("invalid credentials")
		}

		if len(matches) > 1 {
			// IMPORTANT: keep message aligned with owlFront expectations.
			return nil, fmt.Errorf("Multiple institutions found, please select one")
		}

		tenantID = matches[0].TenantID
	}

	// 4. 用户验证和登录
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	var userID, userAccount, nickName, role, tenantName, domain, branchTag string

	switch normalizedUserType {
	case "resident":
		// Step 1: Try resident_contacts table first
		contactInfo, err := s.authRepo.GetResidentContactForLogin(ctx, tenantID, accountHashBytes, passwordHashBytes)
		if err == nil {
			// Family contact login succeeded
			if !contactInfo.IsEnabled {
				s.logger.Warn("User login failed: account not active",
					zap.String("user_id", contactInfo.ContactID),
					zap.String("tenant_id", tenantID),
					zap.String("user_type", normalizedUserType),
					zap.String("reason", "account_not_active"),
					zap.String("note", "family contact not enabled"),
					zap.String("ip_address", req.IPAddress),
				)
				return nil, fmt.Errorf("user is not active")
			}

			userID = contactInfo.ContactID
			userAccount = contactInfo.ContactID // For family contacts, expose a stable identifier as user_account
			if strings.TrimSpace(contactInfo.ContactFirstName+" "+contactInfo.ContactLastName) != "" {
				nickName = strings.TrimSpace(contactInfo.ContactFirstName + " " + contactInfo.ContactLastName)
			} else {
				nickName = contactInfo.Role
			}
			role = contactInfo.Role
			tenantName = contactInfo.TenantName
			domain = contactInfo.Domain
			branchTag = contactInfo.BranchTag
		} else {
			// Step 2: Try resident login
			residentInfo, err := s.authRepo.GetResidentForLogin(ctx, tenantID, accountHashBytes, passwordHashBytes)
			if err != nil {
				s.logger.Warn("User login failed: invalid credentials",
					zap.String("tenant_id", tenantID),
					zap.String("user_type", normalizedUserType),
					zap.String("ip_address", req.IPAddress),
					zap.String("user_agent", req.UserAgent),
					zap.String("reason", "invalid_credentials"),
					zap.String("note", "resident login failed"),
				)
				return nil, fmt.Errorf("invalid credentials")
			}

			if residentInfo.Status != "active" {
				s.logger.Warn("User login failed: account not active",
					zap.String("user_id", residentInfo.ResidentID),
					zap.String("tenant_id", tenantID),
					zap.String("user_type", normalizedUserType),
					zap.String("status", residentInfo.Status),
					zap.String("ip_address", req.IPAddress),
					zap.String("reason", "account_not_active"),
				)
				return nil, fmt.Errorf("user is not active")
			}

			userID = residentInfo.ResidentID
			userAccount = residentInfo.ResidentAccount
			nickName = residentInfo.Nickname
			role = residentInfo.Role
			tenantName = residentInfo.TenantName
			domain = residentInfo.Domain
			branchTag = residentInfo.BranchTag
		}
	default: // staff
		userInfo, err := s.authRepo.GetUserForLogin(ctx, tenantID, accountHashBytes, passwordHashBytes)
		if err != nil {
			s.logger.Warn("User login failed: invalid credentials",
				zap.String("tenant_id", tenantID),
				zap.String("user_type", normalizedUserType),
				zap.String("ip_address", req.IPAddress),
				zap.String("user_agent", req.UserAgent),
				zap.String("reason", "invalid_credentials"),
				zap.String("note", "staff login failed"),
			)
			return nil, fmt.Errorf("invalid credentials")
		}

		if userInfo.Status != "active" {
			s.logger.Warn("User login failed: account not active",
				zap.String("user_id", userInfo.UserID),
				zap.String("user_account", userInfo.UserAccount),
				zap.String("tenant_id", tenantID),
				zap.String("user_type", normalizedUserType),
				zap.String("status", userInfo.Status),
				zap.String("ip_address", req.IPAddress),
				zap.String("reason", "account_not_active"),
			)
			return nil, fmt.Errorf("user is not active")
		}

		userID = userInfo.UserID
		userAccount = userInfo.UserAccount
		nickName = userInfo.Nickname
		role = userInfo.Role
		tenantName = userInfo.TenantName
		domain = userInfo.Domain
		branchTag = userInfo.BranchTag
	}

	// 5. 登录后处理
	if nickName == "" {
		// Prefer nickname; fall back to role/userAccount for display
		if role != "" {
			nickName = role
		} else {
			nickName = userAccount
		}
	}

	// Update last_login_at for staff users
	if normalizedUserType == "staff" {
		if err := s.authRepo.UpdateUserLastLogin(ctx, userID); err != nil {
			// Log error but don't fail login
			s.logger.Warn("Failed to update last_login_at",
				zap.String("user_id", userID),
				zap.Error(err),
			)
		}
	}

	// Log successful login
	s.logger.Info("User login successful",
		zap.String("user_id", userID),
		zap.String("user_account", userAccount),
		zap.String("user_type", normalizedUserType),
		zap.String("tenant_id", tenantID),
		zap.String("tenant_name", tenantName),
		zap.String("role", role),
		zap.String("ip_address", req.IPAddress),
		zap.String("user_agent", req.UserAgent),
		zap.Time("login_time", time.Now()),
	)

	// 6. 构建响应
	resp := &LoginResponse{
		AccessToken:  "stub-access-token",
		RefreshToken: "stub-refresh-token",
		UserID:       userID,
		UserAccount:  userAccount,
		UserType:     normalizedUserType,
		Role:         role,
		NickName:     nickName,
		TenantID:     tenantID,
		TenantName:   tenantName,
		Domain:       domain,
		HomePath:     "/monitoring/overview",
	}

	if branchTag != "" {
		resp.BranchTag = &branchTag
	}

	return resp, nil
}

// SearchInstitutionsRequest 搜索机构请求
type SearchInstitutionsRequest struct {
	AccountHash  string // SHA256(account) 的 hex 编码，必填
	PasswordHash string // SHA256(password) 的 hex 编码，必填
	UserType     string // "staff" | "resident"，默认为 "staff"
}

// Institution 机构信息
type Institution struct {
	ID          string `json:"id"`          // 机构 ID
	Name        string `json:"name"`       // 机构名称
	Domain      string `json:"domain,omitempty"` // 机构域名（可选）
	AccountType string `json:"accountType"` // 账号类型（email/phone/account）
}

// SearchInstitutionsResponse 搜索机构响应
type SearchInstitutionsResponse struct {
	Institutions []Institution `json:"institutions"`
}

// SearchInstitutions 搜索机构
func (s *authService) SearchInstitutions(ctx context.Context, req SearchInstitutionsRequest) (*SearchInstitutionsResponse, error) {
	// 1. 参数验证和规范化
	req.AccountHash = strings.TrimSpace(req.AccountHash)
	req.PasswordHash = strings.TrimSpace(req.PasswordHash)
	if req.AccountHash == "" || req.PasswordHash == "" {
		return &SearchInstitutionsResponse{Institutions: []Institution{}}, nil
	}

	normalizedUserType := strings.ToLower(strings.TrimSpace(req.UserType))
	if normalizedUserType == "" {
		normalizedUserType = "staff"
	}

	// 2. Hash 解码和验证
	accountHashBytes, err := hex.DecodeString(req.AccountHash)
	if err != nil || len(accountHashBytes) == 0 {
		return &SearchInstitutionsResponse{Institutions: []Institution{}}, nil
	}

	passwordHashBytes, err := hex.DecodeString(req.PasswordHash)
	if err != nil || len(passwordHashBytes) == 0 {
		return &SearchInstitutionsResponse{Institutions: []Institution{}}, nil
	}

	// 3. 查询匹配的机构
	var matches []repository.TenantLoginMatch
	if normalizedUserType == "resident" {
		matches, err = s.authRepo.SearchTenantsForResidentLogin(ctx, accountHashBytes, passwordHashBytes)
	} else {
		matches, err = s.authRepo.SearchTenantsForUserLogin(ctx, accountHashBytes, passwordHashBytes)
	}
	if err != nil {
		return &SearchInstitutionsResponse{Institutions: []Institution{}}, nil
	}

	if len(matches) == 0 {
		return &SearchInstitutionsResponse{Institutions: []Institution{}}, nil
	}

	// 4. 机构信息补充
	institutions := make([]Institution, 0, len(matches))
	for _, match := range matches {
		tenant, err := s.tenantsRepo.GetTenant(ctx, match.TenantID)
		if err != nil {
			// If tenant not found, still return tenant_id with accountType
			institutions = append(institutions, Institution{
				ID:          match.TenantID,
				AccountType: match.AccountType,
			})
			continue
		}

		// Special handling for System tenant
		if match.TenantID == SystemTenantID {
			institutions = append(institutions, Institution{
				ID:          SystemTenantID,
				Name:        "System",
				AccountType: match.AccountType,
			})
			continue
		}

		if tenant.Status == "deleted" {
			continue
		}

		inst := Institution{
			ID:          match.TenantID,
			Name:        tenant.TenantName,
			AccountType: match.AccountType,
		}
		if tenant.Domain != "" {
			inst.Domain = tenant.Domain
		}
		institutions = append(institutions, inst)
	}

	return &SearchInstitutionsResponse{Institutions: institutions}, nil
}

// SendVerificationCodeRequest 发送验证码请求
type SendVerificationCodeRequest struct {
	Account    string // 账号（email/phone/userAccount）
	UserType   string // "staff" | "resident"
	TenantID   string // 租户 ID（可选）
	TenantName string // 租户名称（可选）
}

// SendVerificationCodeResponse 发送验证码响应
type SendVerificationCodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// sha256HexAuth 计算字符串的 SHA256 hash（hex 编码）（用于 auth_service，避免与 user_service.go 中的 sha256Hex 冲突）
func sha256HexAuth(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// generateVerificationCode 生成6位数字验证码
func generateVerificationCode() (string, error) {
	// 生成 100000-999999 之间的随机数
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate verification code: %w", err)
	}
	code := 100000 + int(n.Int64())
	return fmt.Sprintf("%06d", code), nil
}

// createCodeKey 创建验证码存储键
func createCodeKey(account, userType, tenantName string) string {
	// 使用小写和规范化，确保键的一致性
	normalizedAccount := strings.ToLower(strings.TrimSpace(account))
	normalizedUserType := strings.ToLower(strings.TrimSpace(userType))
	normalizedTenantName := strings.ToLower(strings.TrimSpace(tenantName))
	return fmt.Sprintf("%s:%s:%s", normalizedAccount, normalizedUserType, normalizedTenantName)
}

// SendVerificationCode 发送验证码
func (s *authService) SendVerificationCode(ctx context.Context, req SendVerificationCodeRequest) (*SendVerificationCodeResponse, error) {
	// 参数验证
	if strings.TrimSpace(req.Account) == "" {
		return nil, fmt.Errorf("account is required")
	}
	if strings.TrimSpace(req.UserType) == "" {
		return nil, fmt.Errorf("user_type is required")
	}
	if strings.TrimSpace(req.TenantName) == "" {
		return nil, fmt.Errorf("tenant_name is required")
	}

	// 1. 查找用户（验证账号是否存在）
	normalizedUserType := strings.ToLower(strings.TrimSpace(req.UserType))
	var found bool

	if s.db != nil {
		// 从数据库查找用户
		accountHash, err := hex.DecodeString(sha256HexAuth(strings.ToLower(strings.TrimSpace(req.Account))))
		if err == nil {
			if normalizedUserType == "staff" {
				// 查找 users 表
				var userIDFromDB sql.NullString
				var tenantIDFromDB sql.NullString
				err := s.db.QueryRowContext(ctx,
					`SELECT user_id::text, tenant_id::text FROM users 
					 WHERE user_account_hash = $1 OR email_hash = $1 OR phone_hash = $1
					 LIMIT 1`,
					accountHash,
				).Scan(&userIDFromDB, &tenantIDFromDB)
				if err == nil && userIDFromDB.Valid {
					found = true
				}
			} else if normalizedUserType == "resident" {
				// 查找 residents 表或 resident_contacts 表
				var residentIDFromDB sql.NullString
				var tenantIDFromDB sql.NullString
				err := s.db.QueryRowContext(ctx,
					`SELECT resident_id::text, tenant_id::text FROM residents 
					 WHERE resident_account_hash = $1 OR email_hash = $1 OR phone_hash = $1
					 LIMIT 1`,
					accountHash,
				).Scan(&residentIDFromDB, &tenantIDFromDB)
				if err == nil && residentIDFromDB.Valid {
					found = true
				} else {
					// 尝试 resident_contacts 表
					var contactIDFromDB sql.NullString
					err := s.db.QueryRowContext(ctx,
						`SELECT contact_id::text FROM resident_contacts 
						 WHERE email_hash = $1 OR phone_hash = $1
						 LIMIT 1`,
						accountHash,
					).Scan(&contactIDFromDB)
					if err == nil && contactIDFromDB.Valid {
						found = true
					}
				}
			}
		}
	}

	if !found {
		// 为了安全，即使账号不存在也返回成功（防止账号枚举攻击）
		s.logger.Warn("SendVerificationCode: account not found",
			zap.String("account", req.Account),
			zap.String("user_type", req.UserType),
			zap.String("tenant_name", req.TenantName),
		)
		return &SendVerificationCodeResponse{
			Success: true,
			Message: "If the account exists, a verification code has been sent",
		}, nil
	}

	// 2. 生成验证码
	code, err := generateVerificationCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification code: %w", err)
	}

	// 3. 存储验证码（5分钟有效期）
	key := createCodeKey(req.Account, req.UserType, req.TenantName)
	s.codeStore.mu.Lock()
	s.codeStore.codes[key] = verificationCodeData{
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	s.codeStore.mu.Unlock()

	// 4. 清理过期验证码（后台清理）
	go s.cleanupExpiredCodes()

	// 5. 发送验证码（TODO: 实际发送邮件或短信）
	// 当前实现：只记录日志，不实际发送
	s.logger.Info("Verification code generated",
		zap.String("account", req.Account),
		zap.String("user_type", req.UserType),
		zap.String("tenant_name", req.TenantName),
		zap.String("code", code), // 开发环境可以记录，生产环境应该移除
	)

	// TODO: 实际发送验证码到用户邮箱或手机
	// - 如果账号是邮箱，发送邮件
	// - 如果账号是手机，发送短信
	// - 需要集成邮件服务或短信服务

	return &SendVerificationCodeResponse{
		Success: true,
		Message: "Verification code has been sent",
	}, nil
}

// cleanupExpiredCodes 清理过期的验证码
func (s *authService) cleanupExpiredCodes() {
	s.codeStore.mu.Lock()
	defer s.codeStore.mu.Unlock()

	now := time.Now()
	for key, data := range s.codeStore.codes {
		if now.After(data.ExpiresAt) {
			delete(s.codeStore.codes, key)
		}
	}
}

// cleanupExpiredTokens 清理过期的重置令牌
func (s *authService) cleanupExpiredTokens() {
	s.tokenStore.mu.Lock()
	defer s.tokenStore.mu.Unlock()

	now := time.Now()
	for token, data := range s.tokenStore.tokens {
		if now.After(data.ExpiresAt) {
			delete(s.tokenStore.tokens, token)
		}
	}
}


// VerifyCodeRequest 验证验证码请求
type VerifyCodeRequest struct {
	Account    string // 账号
	Code       string // 验证码
	UserType   string // "staff" | "resident"
	TenantID   string // 租户 ID（可选）
	TenantName string // 租户名称（必填）
}

// VerifyCodeResponse 验证验证码响应
type VerifyCodeResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"` // 验证令牌（用于重置密码）
	Message string `json:"message,omitempty"`
}

// generateResetToken 生成重置密码令牌
func generateResetToken() (string, error) {
	// 生成 32 字节的随机令牌
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate reset token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// VerifyCode 验证验证码
func (s *authService) VerifyCode(ctx context.Context, req VerifyCodeRequest) (*VerifyCodeResponse, error) {
	// 参数验证
	if strings.TrimSpace(req.Account) == "" {
		return nil, fmt.Errorf("account is required")
	}
	if strings.TrimSpace(req.Code) == "" {
		return nil, fmt.Errorf("code is required")
	}
	if strings.TrimSpace(req.UserType) == "" {
		return nil, fmt.Errorf("user_type is required")
	}
	if strings.TrimSpace(req.TenantName) == "" {
		return nil, fmt.Errorf("tenant_name is required")
	}

	// 1. 查找验证码
	key := createCodeKey(req.Account, req.UserType, req.TenantName)
	s.codeStore.mu.RLock()
	data, exists := s.codeStore.codes[key]
	s.codeStore.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("verification code not found or expired")
	}

	// 2. 检查是否过期
	if time.Now().After(data.ExpiresAt) {
		// 清理过期验证码
		s.codeStore.mu.Lock()
		delete(s.codeStore.codes, key)
		s.codeStore.mu.Unlock()
		return nil, fmt.Errorf("verification code expired")
	}

	// 3. 验证验证码
	if data.Code != req.Code {
		return nil, fmt.Errorf("invalid verification code")
	}

	// 4. 验证码正确，生成重置令牌
	token, err := generateResetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate reset token: %w", err)
	}

	// 5. 查找租户信息
	var tenantID string
	if req.TenantID != "" {
		tenantID = req.TenantID
	} else if s.tenantsRepo != nil {
		// 根据 tenant_name 查找 tenant_id
		tenants, _, err := s.tenantsRepo.ListTenants(ctx, repository.TenantFilters{
			Search: req.TenantName,
		}, 1, 10)
		if err == nil && len(tenants) > 0 {
			// 精确匹配 tenant_name（不区分大小写）
			for _, t := range tenants {
				if strings.EqualFold(t.TenantName, req.TenantName) {
					tenantID = t.TenantID
					break
				}
			}
		}
	}

	// 6. 存储重置令牌（10分钟有效期）
	s.tokenStore.mu.Lock()
	s.tokenStore.tokens[token] = resetTokenData{
		Account:    req.Account,
		UserType:   req.UserType,
		TenantID:   tenantID,
		TenantName: req.TenantName,
		ExpiresAt:  time.Now().Add(10 * time.Minute),
	}
	s.tokenStore.mu.Unlock()

	// 7. 清理过期令牌（后台清理）
	go s.cleanupExpiredTokens()

	// 8. 删除已使用的验证码
	s.codeStore.mu.Lock()
	delete(s.codeStore.codes, key)
	s.codeStore.mu.Unlock()

	return &VerifyCodeResponse{
		Success: true,
		Token:   token,
		Message: "Verification code verified successfully",
	}, nil
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Token       string // 验证令牌（从 VerifyCode 获取）
	NewPassword string // 新密码（明文，后端会进行 hash）
	UserType    string // "staff" | "resident"
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ResetPassword 重置密码
func (s *authService) ResetPassword(ctx context.Context, req ResetPasswordRequest) (*ResetPasswordResponse, error) {
	// 参数验证
	if strings.TrimSpace(req.Token) == "" {
		return nil, fmt.Errorf("token is required")
	}
	if strings.TrimSpace(req.NewPassword) == "" {
		return nil, fmt.Errorf("new_password is required")
	}
	if strings.TrimSpace(req.UserType) == "" {
		return nil, fmt.Errorf("user_type is required")
	}

	if s.db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	// 1. 验证重置令牌
	s.tokenStore.mu.RLock()
	tokenData, exists := s.tokenStore.tokens[req.Token]
	s.tokenStore.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("invalid or expired reset token")
	}

	// 2. 检查令牌是否过期
	if time.Now().After(tokenData.ExpiresAt) {
		// 清理过期令牌
		s.tokenStore.mu.Lock()
		delete(s.tokenStore.tokens, req.Token)
		s.tokenStore.mu.Unlock()
		return nil, fmt.Errorf("reset token expired")
	}

	// 3. 验证 user_type 是否匹配
	if strings.ToLower(strings.TrimSpace(req.UserType)) != strings.ToLower(strings.TrimSpace(tokenData.UserType)) {
		return nil, fmt.Errorf("user_type mismatch")
	}

	// 4. 查找用户并更新密码
	normalizedUserType := strings.ToLower(strings.TrimSpace(req.UserType))
	accountHash, err := hex.DecodeString(sha256Hex(strings.ToLower(strings.TrimSpace(tokenData.Account))))
	if err != nil {
		return nil, fmt.Errorf("failed to hash account: %w", err)
	}

	// 计算新密码的 hash
	passwordHashHex := sha256HexAuth(req.NewPassword)
	passwordHash, err := hex.DecodeString(passwordHashHex)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var rowsAffected int64
	if normalizedUserType == "staff" {
		// 更新 users 表的 password_hash
		result, err := s.db.ExecContext(ctx,
			`UPDATE users 
			 SET password_hash = $1
			 WHERE (user_account_hash = $2 OR email_hash = $2 OR phone_hash = $2)
			   AND tenant_id = $3`,
			passwordHash, accountHash, tokenData.TenantID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reset password: %w", err)
		}
		rowsAffected, _ = result.RowsAffected()
	} else if normalizedUserType == "resident" {
		// 先尝试更新 residents 表
		result, err := s.db.ExecContext(ctx,
			`UPDATE residents 
			 SET password_hash = $1
			 WHERE (resident_account_hash = $2 OR email_hash = $2 OR phone_hash = $2)
			   AND tenant_id = $3`,
			passwordHash, accountHash, tokenData.TenantID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reset password: %w", err)
		}
		rowsAffected, _ = result.RowsAffected()

		// 如果没有更新 residents 表，尝试更新 resident_contacts 表
		if rowsAffected == 0 {
			result, err := s.db.ExecContext(ctx,
				`UPDATE resident_contacts 
				 SET password_hash = $1
				 WHERE (email_hash = $2 OR phone_hash = $2)
				   AND tenant_id = $3`,
				passwordHash, accountHash, tokenData.TenantID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to reset password: %w", err)
			}
			rowsAffected, _ = result.RowsAffected()
		}
	} else {
		return nil, fmt.Errorf("unsupported user_type: %s", req.UserType)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("account not found")
	}

	// 5. 删除已使用的重置令牌
	s.tokenStore.mu.Lock()
	delete(s.tokenStore.tokens, req.Token)
	s.tokenStore.mu.Unlock()

	s.logger.Info("Password reset successful",
		zap.String("account", tokenData.Account),
		zap.String("user_type", tokenData.UserType),
		zap.String("tenant_id", tokenData.TenantID),
	)

	return &ResetPasswordResponse{
		Success: true,
		Message: "Password has been reset successfully",
	}, nil
}

