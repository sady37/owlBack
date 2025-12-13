package repository

import (
	"context"
	"database/sql"
	"encoding/json"
)

type PostgresTenantsRepo struct {
	db *sql.DB
}

func NewPostgresTenantsRepo(db *sql.DB) *PostgresTenantsRepo {
	return &PostgresTenantsRepo{db: db}
}

func (r *PostgresTenantsRepo) ListTenants(ctx context.Context, status string, page, size int) ([]Tenant, int, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	offset := (page - 1) * size

	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM tenants WHERE ($1 = '' OR status = $1)`,
		status,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT tenant_id::text, tenant_name, COALESCE(domain,''), COALESCE(email,''), COALESCE(phone,''), COALESCE(status,'active'), COALESCE(metadata,'{}'::jsonb)
		 FROM tenants
		 WHERE ($1 = '' OR status = $1)
		 ORDER BY tenant_name
		 LIMIT $2 OFFSET $3`,
		status, size, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []Tenant{}
	for rows.Next() {
		var t Tenant
		var meta json.RawMessage
		if err := rows.Scan(&t.TenantID, &t.TenantName, &t.Domain, &t.Email, &t.Phone, &t.Status, &meta); err != nil {
			return nil, 0, err
		}
		t.Metadata = meta
		items = append(items, t)
	}
	return items, total, rows.Err()
}

func (r *PostgresTenantsRepo) CreateTenant(ctx context.Context, payload map[string]any) (*Tenant, error) {
	name, _ := payload["tenant_name"].(string)
	domain, _ := payload["domain"].(string)
	email, _ := payload["email"].(string)
	phone, _ := payload["phone"].(string)
	status, _ := payload["status"].(string)
	if status == "" {
		status = "active"
	}

	var meta any
	if v, ok := payload["metadata"]; ok {
		meta = v
	}
	metaBytes, _ := json.Marshal(meta)
	if len(metaBytes) == 0 {
		metaBytes = []byte(`{}`)
	}

	var t Tenant
	var metaOut json.RawMessage
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO tenants (tenant_name, domain, email, phone, status, metadata)
		 VALUES ($1, NULLIF($2,''), NULLIF($3,''), NULLIF($4,''), $5, $6::jsonb)
		 RETURNING tenant_id::text, tenant_name, COALESCE(domain,''), COALESCE(email,''), COALESCE(phone,''), COALESCE(status,'active'), COALESCE(metadata,'{}'::jsonb)`,
		name, domain, email, phone, status, string(metaBytes),
	).Scan(&t.TenantID, &t.TenantName, &t.Domain, &t.Email, &t.Phone, &t.Status, &metaOut)
	if err != nil {
		return nil, err
	}
	t.Metadata = metaOut
	return &t, nil
}

func (r *PostgresTenantsRepo) UpdateTenant(ctx context.Context, tenantID string, payload map[string]any) (*Tenant, error) {
	name, _ := payload["tenant_name"].(string)
	domain, _ := payload["domain"].(string)
	email, _ := payload["email"].(string)
	phone, _ := payload["phone"].(string)
	status, _ := payload["status"].(string)

	metaBytes := []byte{}
	if v, ok := payload["metadata"]; ok {
		metaBytes, _ = json.Marshal(v)
	}
	metaStr := ""
	if len(metaBytes) > 0 {
		metaStr = string(metaBytes)
	}

	var t Tenant
	var metaOut json.RawMessage
	err := r.db.QueryRowContext(ctx,
		`UPDATE tenants SET
		   tenant_name = COALESCE(NULLIF($2,''), tenant_name),
		   domain      = COALESCE(NULLIF($3,''), domain),
		   email       = COALESCE(NULLIF($4,''), email),
		   phone       = COALESCE(NULLIF($5,''), phone),
		   status      = COALESCE(NULLIF($6,''), status),
		   metadata    = CASE WHEN $7 = '' THEN metadata ELSE $7::jsonb END
		 WHERE tenant_id = $1::uuid
		 RETURNING tenant_id::text, tenant_name, COALESCE(domain,''), COALESCE(email,''), COALESCE(phone,''), COALESCE(status,'active'), COALESCE(metadata,'{}'::jsonb)`,
		tenantID, name, domain, email, phone, status, metaStr,
	).Scan(&t.TenantID, &t.TenantName, &t.Domain, &t.Email, &t.Phone, &t.Status, &metaOut)
	if err != nil {
		return nil, err
	}
	t.Metadata = metaOut
	return &t, nil
}

func (r *PostgresTenantsRepo) SetTenantStatus(ctx context.Context, tenantID string, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tenants SET status = $2 WHERE tenant_id = $1::uuid`,
		tenantID, status,
	)
	return err
}
