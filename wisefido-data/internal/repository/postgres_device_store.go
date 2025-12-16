package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type PostgresDeviceStoreRepo struct {
	db *sql.DB
}

func NewPostgresDeviceStoreRepo(db *sql.DB) *PostgresDeviceStoreRepo {
	return &PostgresDeviceStoreRepo{db: db}
}

func (r *PostgresDeviceStoreRepo) ListDeviceStores(ctx context.Context, filters map[string]any) ([]DeviceStore, int, error) {
	where := []string{}
	args := []any{}
	argN := 1

	// Search filter (search by serial_number, uid, imei)
	if search, ok := filters["search"].(string); ok && search != "" {
		where = append(where, fmt.Sprintf("(ds.serial_number ILIKE $%d OR ds.uid ILIKE $%d OR ds.imei ILIKE $%d)", argN, argN, argN))
		args = append(args, "%"+search+"%")
		argN++
	}

	// Tenant filter
	if tenantID, ok := filters["tenant_id"].(string); ok && tenantID != "" {
		where = append(where, fmt.Sprintf("ds.tenant_id = $%d", argN))
		args = append(args, tenantID)
		argN++
	}

	// Device type filter
	if deviceType, ok := filters["device_type"].(string); ok && deviceType != "" {
		where = append(where, fmt.Sprintf("ds.device_type = $%d", argN))
		args = append(args, deviceType)
		argN++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// Count total
	queryCount := `
		SELECT COUNT(*)
		FROM device_store ds
		LEFT JOIN tenants t ON ds.tenant_id = t.tenant_id
		` + whereClause

	var total int
	if err := r.db.QueryRowContext(ctx, queryCount, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Pagination
	page, _ := filters["page"].(int)
	size, _ := filters["size"].(int)
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 100 // Default larger size for device store
	}
	offset := (page - 1) * size

	argsList := append(args, size, offset)
	limitPos := argN
	offsetPos := argN + 1

	// Query data
	query := `
		SELECT
			ds.device_store_id::text,
			ds.device_type,
			ds.device_model,
			ds.serial_number,
			ds.uid,
			ds.imei,
			ds.comm_mode,
			ds.mcu_model,
			ds.firmware_version,
			ds.ota_target_firmware_version,
			ds.ota_target_mcu_model,
			ds.tenant_id::text,
			COALESCE(t.tenant_name, '') as tenant_name,
			ds.allow_access,
			ds.import_date,
			ds.allocate_time
		FROM device_store ds
		LEFT JOIN tenants t ON ds.tenant_id = t.tenant_id
		` + whereClause + `
		ORDER BY ds.import_date DESC, ds.device_type, ds.serial_number
		LIMIT $` + fmt.Sprintf("%d", limitPos) + ` OFFSET $` + fmt.Sprintf("%d", offsetPos)

	rows, err := r.db.QueryContext(ctx, query, argsList...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []DeviceStore{}
	for rows.Next() {
		var d DeviceStore
		if err := rows.Scan(
			&d.DeviceStoreID,
			&d.DeviceType,
			&d.DeviceModel,
			&d.SerialNumber,
			&d.UID,
			&d.IMEI,
			&d.CommMode,
			&d.MCUModel,
			&d.FirmwareVersion,
			&d.OTATargetFirmwareVersion,
			&d.OTATargetMCUModel,
			&d.TenantID,
			&d.TenantName,
			&d.AllowAccess,
			&d.ImportDate,
			&d.AllocateTime,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, d)
	}
	return out, total, rows.Err()
}

func (r *PostgresDeviceStoreRepo) BatchUpdateDeviceStores(ctx context.Context, updates []map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, update := range updates {
		deviceStoreID, ok := update["device_store_id"].(string)
		if !ok || deviceStoreID == "" {
			continue
		}

		setParts := []string{}
		args := []any{}
		argN := 1

		// tenant_id
		if tenantID, ok := update["tenant_id"].(string); ok {
			if tenantID == "" || tenantID == "null" {
				setParts = append(setParts, fmt.Sprintf("tenant_id = $%d", argN))
				args = append(args, "00000000-0000-0000-0000-000000000000") // Default unallocated tenant ID
			} else {
				setParts = append(setParts, fmt.Sprintf("tenant_id = $%d", argN))
				args = append(args, tenantID)
			}
			argN++
		}

		// ota_target_firmware_version
		if val, ok := update["ota_target_firmware_version"]; ok {
			if val == nil || val == "" || val == "null" {
				setParts = append(setParts, fmt.Sprintf("ota_target_firmware_version = $%d", argN))
				args = append(args, nil)
			} else if s, ok := val.(string); ok {
				setParts = append(setParts, fmt.Sprintf("ota_target_firmware_version = $%d", argN))
				args = append(args, s)
			}
			argN++
		}

		// ota_target_mcu_model
		if val, ok := update["ota_target_mcu_model"]; ok {
			if val == nil || val == "" || val == "null" {
				setParts = append(setParts, fmt.Sprintf("ota_target_mcu_model = $%d", argN))
				args = append(args, nil)
			} else if s, ok := val.(string); ok {
				setParts = append(setParts, fmt.Sprintf("ota_target_mcu_model = $%d", argN))
				args = append(args, s)
			}
			argN++
		}

		// allow_access
		if val, ok := update["allow_access"].(bool); ok {
			setParts = append(setParts, fmt.Sprintf("allow_access = $%d", argN))
			args = append(args, val)
			argN++
		}

		// Update allocate_time when tenant_id is set
		if _, hasTenantID := update["tenant_id"]; hasTenantID {
			if tenantID, ok := update["tenant_id"].(string); ok && tenantID != "" && tenantID != "00000000-0000-0000-0000-000000000000" {
				setParts = append(setParts, fmt.Sprintf("allocate_time = CASE WHEN allocate_time IS NULL THEN CURRENT_TIMESTAMP ELSE allocate_time END"))
			}
		}

		if len(setParts) == 0 {
			continue
		}

		query := fmt.Sprintf(`
			UPDATE device_store
			SET %s
			WHERE device_store_id = $%d
		`, strings.Join(setParts, ", "), argN)
		args = append(args, deviceStoreID)

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ImportDeviceStores imports device stores from data (for batch import)
func (r *PostgresDeviceStoreRepo) ImportDeviceStores(ctx context.Context, items []map[string]any) (int, []map[string]any, []map[string]any, error) {
	if len(items) == 0 {
		return 0, nil, nil, nil
	}

	var successCount int
	var errors []map[string]any
	var skipped []map[string]any

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	defer tx.Rollback()

	for rowIdx, item := range items {
		row := rowIdx + 1

		// Convert tenant_name to tenant_id if provided
		if tenantName, ok := item["tenant_name"].(string); ok && tenantName != "" {
			var tenantID string
			err := tx.QueryRowContext(ctx, "SELECT tenant_id::text FROM tenants WHERE tenant_name = $1 LIMIT 1", tenantName).Scan(&tenantID)
			if err != nil {
				if err == sql.ErrNoRows {
					errors = append(errors, map[string]any{
						"row":         row,
						"tenant_name": tenantName,
						"error":       fmt.Sprintf("Tenant name '%s' not found", tenantName),
					})
					continue
				}
				errors = append(errors, map[string]any{
					"row":   row,
					"error": fmt.Sprintf("Failed to lookup tenant: %v", err),
				})
				continue
			}
			item["tenant_id"] = tenantID
			// Remove tenant_name from item since we've converted it to tenant_id
			delete(item, "tenant_name")
		}

		// Validate required fields
		deviceType, _ := item["device_type"].(string)
		if deviceType == "" {
			errors = append(errors, map[string]any{
				"row":   row,
				"error": "Device type is required",
			})
			continue
		}

		serialNumber, _ := item["serial_number"].(string)
		uid, _ := item["uid"].(string)
		if serialNumber == "" && uid == "" {
			errors = append(errors, map[string]any{
				"row":           row,
				"serial_number": serialNumber,
				"uid":           uid,
				"error":         "Serial number or UID is required",
			})
			continue
		}

		// Check if device already exists
		var existingID string
		checkQuery := `
			SELECT device_store_id::text
			FROM device_store
			WHERE (serial_number = $1 AND serial_number IS NOT NULL)
			   OR (uid = $2 AND uid IS NOT NULL)
			LIMIT 1
		`
		err := tx.QueryRowContext(ctx, checkQuery, serialNumber, uid).Scan(&existingID)
		if err == nil {
			// Device already exists, skip
			skipped = append(skipped, map[string]any{
				"row":           row,
				"serial_number": serialNumber,
				"uid":           uid,
				"reason":        "Device already exists",
			})
			continue
		} else if err != sql.ErrNoRows {
			errors = append(errors, map[string]any{
				"row":   row,
				"error": fmt.Sprintf("Failed to check existing device: %v", err),
			})
			continue
		}

		// Insert new device
		insertQuery := `
			INSERT INTO device_store (
				device_type, device_model, serial_number, uid, imei,
				comm_mode, mcu_model, firmware_version,
				tenant_id, allow_access
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		// Get allow_access value (can be bool or string "Yes"/"No")
		allowAccess := false
		if val, ok := item["allow_access"].(bool); ok {
			allowAccess = val
		} else if val, ok := item["allow_access"].(string); ok {
			allowAccess = (val == "Yes" || val == "yes" || val == "TRUE" || val == "true" || val == "1")
		}

		// Get tenant_id value, use default if not provided in Excel
		tenantID := getNullableString(item, "tenant_id")
		if tenantID == nil {
			// If tenant_id is not provided in Excel, use database default
			// Database default is '00000000-0000-0000-0000-000000000000' (unallocated)
			tenantID = "00000000-0000-0000-0000-000000000000"
		}

		args := []any{
			deviceType,
			getNullableString(item, "device_model"),
			getNullableString(item, "serial_number"),
			getNullableString(item, "uid"),
			getNullableString(item, "imei"),
			getNullableString(item, "comm_mode"),
			getNullableString(item, "mcu_model"),
			getNullableString(item, "firmware_version"),
			tenantID,
			allowAccess,
		}

		if _, err := tx.ExecContext(ctx, insertQuery, args...); err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // unique_violation
				skipped = append(skipped, map[string]any{
					"row":           row,
					"serial_number": serialNumber,
					"uid":           uid,
					"reason":        "Device already exists (duplicate)",
				})
			} else {
				errors = append(errors, map[string]any{
					"row":           row,
					"serial_number": serialNumber,
					"uid":           uid,
					"error":         fmt.Sprintf("Failed to insert: %v", err),
				})
			}
			continue
		}

		successCount++
	}

	if err := tx.Commit(); err != nil {
		return 0, nil, nil, err
	}

	return successCount, errors, skipped, nil
}

// getNullableString returns sql.NullString from map
func getNullableString(item map[string]any, key string) interface{} {
	if val, ok := item[key].(string); ok && val != "" {
		return val
	}
	return nil
}
