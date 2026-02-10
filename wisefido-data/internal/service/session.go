package service

import (
	"context"
)

// context keys used to store trusted session values in request context
type ContextKey string

const (
	ContextKeyUserID   ContextKey = "X-User-Id"
	ContextKeyTenantID ContextKey = "X-Tenant-Id"
	ContextKeyUserType ContextKey = "X-User-Type"
	ContextKeyUserRole ContextKey = "X-User-Role"
)

// GetSessionFromContext 从 context 中读取已验证的会话字段。
// 返回 ok==false 表示会话不可用或不完整。
func GetSessionFromContext(ctx context.Context) (userID, tenantID, userType, role string, ok bool) {
	if ctx == nil {
		return "", "", "", "", false
	}
	uid, _ := ctx.Value(ContextKeyUserID).(string)
	tid, _ := ctx.Value(ContextKeyTenantID).(string)
	ut, _ := ctx.Value(ContextKeyUserType).(string)
	role, _ = ctx.Value(ContextKeyUserRole).(string)
	if uid == "" || tid == "" || ut == "" {
		return "", "", "", "", false
	}
	return uid, tid, ut, role, true
}

// MustSession 是 GetSessionFromContext 的别名，用于在 handler 中简洁调用。
func MustSession(ctx context.Context) (userID, tenantID, userType, role string, ok bool) {
	return GetSessionFromContext(ctx)
}
