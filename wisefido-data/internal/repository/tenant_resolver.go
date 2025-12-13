package repository

import (
	"context"
	"database/sql"
)

type TenantResolver interface {
	TenantIDByUserID(ctx context.Context, userID string) (string, error)
	TenantIDByUnitID(ctx context.Context, unitID string) (string, error)
	TenantIDByDeviceID(ctx context.Context, deviceID string) (string, error)
}

type PostgresTenantResolver struct {
	db *sql.DB
}

func NewPostgresTenantResolver(db *sql.DB) *PostgresTenantResolver {
	return &PostgresTenantResolver{db: db}
}

func (r *PostgresTenantResolver) TenantIDByUserID(ctx context.Context, userID string) (string, error) {
	var tenantID string
	err := r.db.QueryRowContext(ctx, "SELECT tenant_id::text FROM users WHERE user_id = $1", userID).Scan(&tenantID)
	return tenantID, err
}

func (r *PostgresTenantResolver) TenantIDByUnitID(ctx context.Context, unitID string) (string, error) {
	var tenantID string
	err := r.db.QueryRowContext(ctx, "SELECT tenant_id::text FROM units WHERE unit_id = $1", unitID).Scan(&tenantID)
	return tenantID, err
}

func (r *PostgresTenantResolver) TenantIDByDeviceID(ctx context.Context, deviceID string) (string, error) {
	var tenantID string
	err := r.db.QueryRowContext(ctx, "SELECT tenant_id::text FROM devices WHERE device_id = $1", deviceID).Scan(&tenantID)
	return tenantID, err
}




