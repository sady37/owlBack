package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
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

// authService 实现
type authService struct {
	authRepo     repository.AuthRepository
	tenantsRepo  repository.TenantsRepository
	logger       *zap.Logger
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(authRepo repository.AuthRepository, tenantsRepo repository.TenantsRepository, logger *zap.Logger) AuthService {
	return &authService{
		authRepo:    authRepo,
		tenantsRepo: tenantsRepo,
		logger:      logger,
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
		systemTenantID := "00000000-0000-0000-0000-000000000001"
		if match.TenantID == systemTenantID {
			institutions = append(institutions, Institution{
				ID:          systemTenantID,
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

// SendVerificationCode 发送验证码
func (s *authService) SendVerificationCode(ctx context.Context, req SendVerificationCodeRequest) (*SendVerificationCodeResponse, error) {
	// TODO: 实现发送验证码逻辑
	return nil, fmt.Errorf("database not available")
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

// VerifyCode 验证验证码
func (s *authService) VerifyCode(ctx context.Context, req VerifyCodeRequest) (*VerifyCodeResponse, error) {
	// TODO: 实现验证验证码逻辑
	return nil, fmt.Errorf("database not available")
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
	// TODO: 实现重置密码逻辑
	return nil, fmt.Errorf("database not available")
}

