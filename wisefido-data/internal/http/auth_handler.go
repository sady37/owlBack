package httpapi

import (
	"net/http"
	"strings"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// AuthHandler 认证授权 Handler
type AuthHandler struct {
	authService service.AuthService
	logger      *zap.Logger
}

// NewAuthHandler 创建认证授权 Handler
func NewAuthHandler(authService service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路由分发
	switch r.URL.Path {
	case "/auth/api/v1/login":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h.Login(w, r)
	case "/auth/api/v1/institutions/search":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h.SearchInstitutions(w, r)
	case "/auth/api/v1/forgot-password/send-code":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h.SendVerificationCode(w, r)
	case "/auth/api/v1/forgot-password/verify-code":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h.VerifyCode(w, r)
	case "/auth/api/v1/forgot-password/reset":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h.ResetPassword(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// Login 用户登录
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析（支持多种格式）
	// 对齐 owlFront LoginResult（authModel.ts）
	// loginApi 会把 tenant_id/userType 等放在 JSON body（axios beforeRequestHook 会把 params 作为 data）
	var reqBody map[string]any
	_ = readBodyJSON(r, 1<<20, &reqBody)

	// Some clients may wrap params in {params:{...}}
	if p, ok := reqBody["params"].(map[string]any); ok && p != nil {
		if _, ok2 := reqBody["tenant_id"]; !ok2 {
			reqBody["tenant_id"] = p["tenant_id"]
		}
		if _, ok2 := reqBody["userType"]; !ok2 {
			reqBody["userType"] = p["userType"]
		}
		if _, ok2 := reqBody["accountHash"]; !ok2 {
			reqBody["accountHash"] = p["accountHash"]
		}
		if _, ok2 := reqBody["passwordHash"]; !ok2 {
			reqBody["passwordHash"] = p["passwordHash"]
		}
	}

	// 参数优先级：Body > Query
	tenantID, _ := reqBody["tenant_id"].(string)
	if tenantID == "" {
		tenantID = r.URL.Query().Get("tenant_id")
	}

	userType, _ := reqBody["userType"].(string)
	if userType == "" {
		userType = r.URL.Query().Get("userType")
	}
	if userType == "" {
		userType = "staff"
	}

	accountHash, _ := reqBody["accountHash"].(string)
	if accountHash == "" {
		accountHash = r.URL.Query().Get("accountHash")
	}

	passwordHash, _ := reqBody["passwordHash"].(string)
	if passwordHash == "" {
		passwordHash = r.URL.Query().Get("passwordHash")
	}

	// 2. 调用 Service
	req := service.LoginRequest{
		TenantID:     tenantID,
		UserType:     userType,
		AccountHash:  accountHash,
		PasswordHash: passwordHash,
		IPAddress:    getClientIP(r),
		UserAgent:    r.UserAgent(),
	}

	resp, err := h.authService.Login(ctx, req)
	if err != nil {
		// Service 层已经记录了详细的日志，这里只记录错误
		h.logger.Error("Login failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 构建响应（与旧 Handler 格式一致）
	result := map[string]any{
		"accessToken":  resp.AccessToken,
		"refreshToken": resp.RefreshToken,
		"userId":       resp.UserID,
		"user_account": resp.UserAccount,
		"userType":     resp.UserType,
		"role":         resp.Role,
		"nickName":     resp.NickName,
		"tenant_id":    resp.TenantID,
		"tenant_name":  resp.TenantName,
		"domain":       resp.Domain,
		"homePath":     resp.HomePath,
	}

	// Add branchTag if available
	if resp.BranchTag != nil && *resp.BranchTag != "" {
		result["branchTag"] = *resp.BranchTag
	}

	writeJSON(w, http.StatusOK, Ok(result))
}

// SearchInstitutions 搜索机构
func (h *AuthHandler) SearchInstitutions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析（从 Query 参数获取）
	accountHash := strings.TrimSpace(r.URL.Query().Get("accountHash"))
	passwordHash := strings.TrimSpace(r.URL.Query().Get("passwordHash"))
	userType := strings.TrimSpace(r.URL.Query().Get("userType"))
	if userType == "" {
		userType = "staff"
	}

	// 2. 调用 Service
	req := service.SearchInstitutionsRequest{
		AccountHash:  accountHash,
		PasswordHash: passwordHash,
		UserType:     userType,
	}

	resp, err := h.authService.SearchInstitutions(ctx, req)
	if err != nil {
		h.logger.Error("SearchInstitutions failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 构建响应（与旧 Handler 格式一致）
	items := make([]any, 0, len(resp.Institutions))
	for _, inst := range resp.Institutions {
		item := map[string]any{
			"id":          inst.ID,
			"name":        inst.Name,
			"accountType": inst.AccountType,
		}
		if inst.Domain != "" {
			item["domain"] = inst.Domain
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, Ok(items))
}

// SendVerificationCode 发送验证码
func (h *AuthHandler) SendVerificationCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	account, _ := payload["account"].(string)
	userType, _ := payload["userType"].(string)
	tenantID, _ := payload["tenant_id"].(string)
	tenantName, _ := payload["tenant_name"].(string)

	// 2. 调用 Service
	req := service.SendVerificationCodeRequest{
		Account:    account,
		UserType:   userType,
		TenantID:   tenantID,
		TenantName: tenantName,
	}

	resp, err := h.authService.SendVerificationCode(ctx, req)
	if err != nil {
		h.logger.Error("SendVerificationCode failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(resp))
}

// VerifyCode 验证验证码
func (h *AuthHandler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	account, _ := payload["account"].(string)
	code, _ := payload["code"].(string)
	userType, _ := payload["userType"].(string)
	tenantID, _ := payload["tenant_id"].(string)
	tenantName, _ := payload["tenant_name"].(string)

	// 2. 调用 Service
	req := service.VerifyCodeRequest{
		Account:    account,
		Code:       code,
		UserType:   userType,
		TenantID:   tenantID,
		TenantName: tenantName,
	}

	resp, err := h.authService.VerifyCode(ctx, req)
	if err != nil {
		h.logger.Error("VerifyCode failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(resp))
}

// ResetPassword 重置密码
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	token, _ := payload["token"].(string)
	newPassword, _ := payload["newPassword"].(string)
	userType, _ := payload["userType"].(string)

	// 2. 调用 Service
	req := service.ResetPasswordRequest{
		Token:       token,
		NewPassword: newPassword,
		UserType:    userType,
	}

	resp, err := h.authService.ResetPassword(ctx, req)
	if err != nil {
		h.logger.Error("ResetPassword failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(resp))
}

