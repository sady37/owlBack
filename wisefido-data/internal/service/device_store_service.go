package service

import (
	"context"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// DeviceStoreService 设备库存业务：Sleepace 绑定/解绑、InitialAll 查询回写等。
// Handler 只做请求解析与响应写入，业务逻辑在此层。
type DeviceStoreService struct {
	deviceStoreRepo repository.DeviceStoreRepository
	devicesRepo     repository.DevicesRepository
	unitsRepo       repository.UnitsRepository
	sleepaceGateway *SleepaceGatewayClient
	configPublisher *publisher.ConfigPublisher
	logger          *zap.Logger
}

// NewDeviceStoreService 创建设备库存 Service。sleepaceGateway 可为 nil（未配置时不执行绑定相关逻辑）。
func NewDeviceStoreService(
	deviceStoreRepo repository.DeviceStoreRepository,
	devicesRepo repository.DevicesRepository,
	unitsRepo repository.UnitsRepository,
	sleepaceGateway *SleepaceGatewayClient,
	logger *zap.Logger,
) *DeviceStoreService {
	return &DeviceStoreService{
		deviceStoreRepo: deviceStoreRepo,
		devicesRepo:     devicesRepo,
		unitsRepo:       unitsRepo,
		sleepaceGateway: sleepaceGateway,
		logger:          logger,
	}
}

func (s *DeviceStoreService) SetConfigPublisher(pub *publisher.ConfigPublisher) {
	s.configPublisher = pub
}

// BatchUpdateDeviceStoresNotify 批量更新 device_store 成功后发 config.card。
func (s *DeviceStoreService) BatchUpdateDeviceStoresNotify(ctx context.Context, updates []*domain.DeviceStore) error {
	if err := s.deviceStoreRepo.BatchUpdateDeviceStores(ctx, updates); err != nil {
		return err
	}
	NotifyDeviceStoreBatchAfterUpdate(ctx, s.deviceStoreRepo, s.configPublisher, updates, s.logger)
	return nil
}

// ImportDeviceStoresNotify 导入成功后按插入行发 config.card。
func (s *DeviceStoreService) ImportDeviceStoresNotify(ctx context.Context, items []*domain.DeviceStore) (successCount int, inserted []*domain.DeviceStore, skipped []*domain.DeviceStore, errors []*domain.DeviceStore, err error) {
	successCount, inserted, skipped, errors, err = s.deviceStoreRepo.ImportDeviceStores(ctx, items)
	if err != nil {
		return
	}
	NotifyDeviceStoreFromStores(ctx, s.configPublisher, inserted, "device_store_imported", s.logger)
	return
}

func (s *DeviceStoreService) getTimezoneForDevice(ctx context.Context, tenantID, deviceID string) int {
	if s.devicesRepo == nil || s.unitsRepo == nil || tenantID == "" || deviceID == "" {
		return DefaultTimezoneOffsetSeconds
	}
	dev, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil || dev == nil {
		return DefaultTimezoneOffsetSeconds
	}
	if !dev.UnitID.Valid || dev.UnitID.String == "" {
		return DefaultTimezoneOffsetSeconds
	}
	u, err := s.unitsRepo.GetUnit(ctx, tenantID, dev.UnitID.String)
	if err != nil || u == nil || u.Timezone == "" {
		return DefaultTimezoneOffsetSeconds
	}
	return IANAToOffsetSeconds(u.Timezone)
}

// SyncSleepaceBindResult 同步 Sleepace 绑定结果
type SyncSleepaceBindResult struct {
	Synced  int
	Skipped int
	Failed  int
	Errors  []string
}

// SyncSleepaceBind 先查 bindInfo(userId=device_id)，仅对未绑定的 Sleepad 执行绑定。
func (s *DeviceStoreService) SyncSleepaceBind(ctx context.Context) (*SyncSleepaceBindResult, error) {
	if s.sleepaceGateway == nil {
		return nil, fmt.Errorf("sleepace gateway not configured")
	}
	filters := repository.DeviceStoreFilters{DeviceType: "Sleepad"}
	items, _, err := s.deviceStoreRepo.ListDeviceStores(ctx, filters, 1, 1000, "", "asc")
	if err != nil {
		s.logger.Error("SyncSleepaceBind list failed", zap.Error(err))
		return nil, fmt.Errorf("list device_store: %w", err)
	}
	var synced, skipped, failed int
	var errMsgs []string
	for _, row := range items {
		if row.DeviceUID == "" || row.DeviceID == "" {
			continue
		}
		status, _, bindItems, getErr := s.sleepaceGateway.GetBindInfo(ctx, row.DeviceID)
		if getErr != nil {
			failed++
			msg := getErr.Error()
			if msg == "" {
				msg = "getBindInfo failed"
			}
			errMsgs = append(errMsgs, row.DeviceUID+": getBindInfo "+msg)
			continue
		}
		if status == 0 && len(bindItems) > 0 {
			skipped++
			continue
		}
		tz := s.getTimezoneForDevice(ctx, row.TenantID, row.DeviceID)
		if !row.DeviceCode.Valid || row.DeviceCode.String == "" {
			failed++
			errMsgs = append(errMsgs, row.DeviceUID+": device_code required for bind (from import)")
			continue
		}
		_, initErr := s.sleepaceGateway.InitializeDevice(ctx, row.DeviceCode.String, row.DeviceID, &tz)
		if initErr != nil {
			failed++
			msg := initErr.Error()
			if msg == "" {
				msg = "initialize failed (check wisefido-sleepace / sleepace-service logs)"
			}
			errMsgs = append(errMsgs, row.DeviceUID+": "+msg)
			continue
		}
		synced++
	}
	return &SyncSleepaceBindResult{Synced: synced, Skipped: skipped, Failed: failed, Errors: errMsgs}, nil
}

// InitialAllSleepadResult InitialAll 结果：1) 查询回写 2) 未绑定单独绑定
type InitialAllSleepadResult struct {
	Synced           int
	Failed           int
	NoData           int
	NoDataList       []string
	NoDataBound      int
	NoDataBindFailed int
	Errors           []string
	SuccessDetails   []map[string]any
}

// InitialAllSleepad 1) 查询、回写：查所有 Sleepad 的 bindInfo，有数据则把 deviceVersion 回写 firmware_version；
// 2) 把未绑定单独绑定：对 bindInfo 无数据的设备调用 initialize 绑定。
func (s *DeviceStoreService) InitialAllSleepad(ctx context.Context) (*InitialAllSleepadResult, error) {
	if s.sleepaceGateway == nil {
		return nil, fmt.Errorf("sleepace gateway not configured")
	}
	filters := repository.DeviceStoreFilters{DeviceType: "Sleepad"}
	items, _, err := s.deviceStoreRepo.ListDeviceStores(ctx, filters, 1, 1000, "", "asc")
	if err != nil {
		s.logger.Error("InitialAllSleepad list failed", zap.Error(err))
		return nil, fmt.Errorf("list: %w", err)
	}
	var synced, failed, noData int
	var noDataBound, noDataBindFailed int
	var successDetails []map[string]any
	var errMsgs []string
	var noDataList []string
	for _, row := range items {
		if row.DeviceID == "" {
			continue
		}
		status, msg, bindItems, getErr := s.sleepaceGateway.GetBindInfo(ctx, row.DeviceID)
		if getErr != nil {
			failed++
			errMsgs = append(errMsgs, row.DeviceUID+": "+getErr.Error())
			continue
		}
		if status != 0 {
			failed++
			if msg == "" {
				msg = "bindInfo status non-zero"
			}
			errMsgs = append(errMsgs, row.DeviceUID+": "+msg)
			continue
		}
		if len(bindItems) == 0 {
			noData++
			noDataList = append(noDataList, row.DeviceUID)
			tz := s.getTimezoneForDevice(ctx, row.TenantID, row.DeviceID)
			if !row.DeviceCode.Valid || row.DeviceCode.String == "" {
				noDataBindFailed++
				errMsgs = append(errMsgs, row.DeviceUID+": device_code required for bind (from import)")
				continue
			}
			_, initErr := s.sleepaceGateway.InitializeDevice(ctx, row.DeviceCode.String, row.DeviceID, &tz)
			if initErr != nil {
				noDataBindFailed++
				initMsg := initErr.Error()
				if initMsg == "" {
					initMsg = "initialize failed"
				}
				errMsgs = append(errMsgs, row.DeviceUID+": bind "+initMsg)
			} else {
				noDataBound++
			}
			continue
		}
		first := bindItems[0]
		fwVer := fmt.Sprintf("%.2f", first.DeviceVersion)
		if err := s.deviceStoreRepo.UpdateFirmwareVersion(ctx, row.DeviceID, fwVer); err != nil {
			failed++
			errMsgs = append(errMsgs, row.DeviceUID+": update firmware "+err.Error())
			continue
		}
		if s.configPublisher != nil && row.DeviceID != "" {
			_ = s.configPublisher.PublishCardChangeForDevice(ctx, row.TenantID, row.DeviceID, "device_store_firmware_updated", row.DeviceUID)
		}
		synced++
		successDetails = append(successDetails, map[string]any{
			"device_uid": row.DeviceUID, "device_type": first.DeviceType, "firmware_version": fwVer,
		})
	}
	return &InitialAllSleepadResult{
		Synced:           synced,
		Failed:           failed,
		NoData:           noData,
		NoDataList:       noDataList,
		NoDataBound:      noDataBound,
		NoDataBindFailed: noDataBindFailed,
		Errors:           errMsgs,
		SuccessDetails:   successDetails,
	}, nil
}

// BindSleepadOne 单条 Sleepad 绑定：调用 initialize(device_code, device_id)。Sleepace device_id = wisefido device_code（如 1ua3erivl9pv1），Sleepace deviceName = wisefido device_uid（如 BM87224601903）。当前仅绑定在 left（leftRight=0）。
func (s *DeviceStoreService) BindSleepadOne(ctx context.Context, deviceID string) error {
	if s.sleepaceGateway == nil {
		return fmt.Errorf("sleepace gateway not configured")
	}
	ds, err := s.deviceStoreRepo.GetDeviceStoreByDeviceID(ctx, deviceID)
	if err != nil || ds == nil {
		return fmt.Errorf("device not found")
	}
	if !strings.EqualFold(ds.DeviceType, "Sleepad") {
		return fmt.Errorf("not a Sleepad device")
	}
	if !ds.DeviceCode.Valid || ds.DeviceCode.String == "" {
		return fmt.Errorf("device_code required for bind (from import)")
	}
	tz := s.getTimezoneForDevice(ctx, ds.TenantID, ds.DeviceID)
	_, initErr := s.sleepaceGateway.InitializeDevice(ctx, ds.DeviceCode.String, ds.DeviceID, &tz)
	if initErr != nil {
		return initErr
	}
	if s.configPublisher != nil && ds.DeviceID != "" {
		_ = s.configPublisher.PublishCardChangeForDevice(ctx, ds.TenantID, ds.DeviceID, "sleepad_bind", ds.DeviceUID)
	}
	return nil
}

// UnbindSleepadOne 单条 Sleepad 解绑：仅调用 Sleepace unbind(device_code)。不清空 device_code。
func (s *DeviceStoreService) UnbindSleepadOne(ctx context.Context, deviceID string) error {
	if s.sleepaceGateway == nil {
		return fmt.Errorf("sleepace gateway not configured")
	}
	ds, err := s.deviceStoreRepo.GetDeviceStoreByDeviceID(ctx, deviceID)
	if err != nil || ds == nil {
		return fmt.Errorf("device not found")
	}
	if !strings.EqualFold(ds.DeviceType, "Sleepad") {
		return fmt.Errorf("not a Sleepad device")
	}
	if !ds.DeviceCode.Valid || ds.DeviceCode.String == "" {
		return fmt.Errorf("device has no device_code, already unbound")
	}
	if err := s.sleepaceGateway.UnbindDevice(ctx, ds.DeviceCode.String); err != nil {
		return err
	}
	if s.configPublisher != nil && ds.DeviceID != "" {
		_ = s.configPublisher.PublishCardChangeForDevice(ctx, ds.TenantID, ds.DeviceID, "sleepad_unbind", ds.DeviceUID)
	}
	return nil
}

// DeleteDeviceStoreAndNotify 删除 device_store 行并发送 config.card（device_store_deleted）。
func (s *DeviceStoreService) DeleteDeviceStoreAndNotify(ctx context.Context, deviceUID string) error {
	if deviceUID == "" {
		return fmt.Errorf("device_uid required")
	}
	row, err := s.deviceStoreRepo.GetDeviceStore(ctx, deviceUID)
	if err != nil {
		return err
	}
	if row == nil {
		return fmt.Errorf("device not found")
	}
	if err := s.deviceStoreRepo.DeleteDeviceStore(ctx, deviceUID); err != nil {
		return err
	}
	if s.configPublisher != nil && row.DeviceID != "" {
		_ = s.configPublisher.PublishCardChangeForDevice(ctx, row.TenantID, row.DeviceID, "device_store_deleted", row.DeviceUID)
	}
	return nil
}

// PostImportSleepadBind 导入后对新插入的 Sleepad 调用 Sleepace initialize 绑定。device_code 来自厂家导入，不在此写回。
func (s *DeviceStoreService) PostImportSleepadBind(ctx context.Context, inserted []*domain.DeviceStore) {
	if s.sleepaceGateway == nil || len(inserted) == 0 {
		return
	}
	for _, row := range inserted {
		if row.DeviceUID == "" || row.DeviceID == "" {
			continue
		}
		if !strings.EqualFold(row.DeviceType, "Sleepad") {
			continue
		}
		if !row.DeviceCode.Valid || row.DeviceCode.String == "" {
			continue
		}
		tz := s.getTimezoneForDevice(ctx, row.TenantID, row.DeviceID)
		_, initErr := s.sleepaceGateway.InitializeDevice(ctx, row.DeviceCode.String, row.DeviceID, &tz)
		if initErr != nil {
			s.logger.Warn("Sleepace initialize failed for imported Sleepad",
				zap.String("device_uid", row.DeviceUID),
				zap.String("device_id", row.DeviceID),
				zap.Error(initErr))
		}
	}
}
