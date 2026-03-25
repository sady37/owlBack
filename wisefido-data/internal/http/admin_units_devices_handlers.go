package httpapi

import "go.uber.org/zap"

// AdminAPI 历史遗留占位；Unit/Device/DeviceStore 等已迁至独立 Handler，RegisterAdminUnitDeviceRoutes 未注册路径。
type AdminAPI struct {
	Stub *StubHandler
	Log  *zap.Logger
}

func NewAdminAPI(stub *StubHandler, log *zap.Logger) *AdminAPI {
	return &AdminAPI{Stub: stub, Log: log}
}
