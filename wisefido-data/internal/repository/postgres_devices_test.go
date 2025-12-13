package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresDevicesRepo_ListDevices(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresDevicesRepo(db)

	// count
	mock.ExpectQuery(`SELECT COUNT\(\*\)\s+FROM devices d`).
		WithArgs("tenant-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// list
	mock.ExpectQuery(`SELECT\s+d.device_id::text`).
		WithArgs("tenant-1", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"device_id", "tenant_id", "device_store_id", "device_name", "device_model", "device_type",
			"serial_number", "uid", "imei", "comm_mode", "firmware_version", "mcu_model",
			"status", "business_access", "monitoring_enabled", "unit_id", "bound_room_id", "bound_bed_id", "metadata",
		}).AddRow(
			"device-1", "tenant-1", nil, "Radar-01", "WF-RADAR", "Radar",
			"SN", "UID", nil, "WiFi", "1.0.0", nil,
			"offline", "pending", false, "unit-1", nil, nil, nil,
		))

	items, total, err := repo.ListDevices(context.Background(), "tenant-1", map[string]any{
		"page": 1,
		"size": 20,
	})
	if err != nil {
		t.Fatalf("ListDevices err: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 item/total, got len=%d total=%d", len(items), total)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}




