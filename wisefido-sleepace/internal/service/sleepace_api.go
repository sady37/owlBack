package service

import (
	"encoding/json"
	"errors"
	"time"

	"wisefido-sleepace/internal/config"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type SleepaceToken struct {
	AppId     string `json:"appId"`
	SecureKey string `json:"secureKey"`
}

type SleepaceRequest struct {
	Token *SleepaceToken `json:"token"`
	Data  map[string]any `json:"data"`
}

type SleepaceResponse struct {
	Status int             `json:"status"`
	Msg    string          `json:"msg"`
	Data   json.RawMessage `json:"data"`
}

// SleepaceAPI is the HTTP client for communicating with sleepace-service (Java).
type SleepaceAPI struct {
	client *resty.Client
	token  *SleepaceToken
	cfg    *config.SleepaceConfig
	logger *zap.Logger
}

func NewSleepaceAPI(cfg *config.SleepaceConfig, logger *zap.Logger) *SleepaceAPI {
	client := resty.New().
		SetBaseURL(cfg.HTTPAddress).
		SetTimeout(10*time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(time.Second).
		SetRetryMaxWaitTime(5*time.Second).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	api := &SleepaceAPI{
		client: client,
		token:  &SleepaceToken{AppId: cfg.AppID, SecureKey: cfg.SecretKey},
		cfg:    cfg,
		logger: logger,
	}
	api.setPushType()
	return api
}

func (a *SleepaceAPI) setPushType() {
	req := SleepaceRequest{
		Token: a.token,
		Data:  map[string]any{"pushType": "MQTT", "alarmDataType": "array"},
	}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/system/pushType/set")
	a.logger.Info("setPushType", zap.Int("status", resp.Status), zap.String("msg", resp.Msg))
}

// BindDevice 绑定设备。当前仅绑定在 left（leftRight=0，单人床/养老设备）。
func (a *SleepaceAPI) BindDevice(deviceCode, userID string, timezone int) error {
	req := SleepaceRequest{
		Token: a.token,
		Data: map[string]any{
			"deviceId": deviceCode, "leftRight": 0,
			"userId": userID, "gender": 1, "age": 50, "timezone": timezone,
		},
	}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/bind")
	if resp.Status != 0 {
		return errors.New(resp.Msg)
	}
	return nil
}

func (a *SleepaceAPI) UnbindDevice(deviceCode string) error {
	req := SleepaceRequest{Token: a.token, Data: map[string]any{"deviceId": deviceCode}}
	resp := SleepaceResponse{}
	_, err := a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/unbind")
	if err != nil {
		return err
	}
	if resp.Status != 0 {
		return errors.New(resp.Msg)
	}
	return nil
}

func (a *SleepaceAPI) SetHeartMode(deviceCode string) error {
	req := SleepaceRequest{Token: a.token, Data: map[string]any{"deviceId": deviceCode, "mode": 1}}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/heartModeSet")
	if resp.Status != 0 {
		return errors.New(resp.Msg)
	}
	return nil
}

func (a *SleepaceAPI) SetRealtimeInterval(deviceCode, userID string, interval int) error {
	req := SleepaceRequest{
		Token: a.token,
		Data: map[string]any{
			"deviceId": deviceCode, "userId": userID, "leftRight": 0, "interval": interval,
		},
	}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/device/updateconfig")
	if resp.Status != 0 {
		return errors.New(resp.Msg)
	}
	return nil
}

func (a *SleepaceAPI) SetLeaveSensibility(deviceCode, userID string, mode int) error {
	req := SleepaceRequest{
		Token: a.token,
		Data: map[string]any{
			"deviceId": deviceCode, "userId": userID, "leftRight": 0, "mode": mode,
		},
	}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/device/updateAlgMode")
	if resp.Status != 0 {
		return errors.New(resp.Msg)
	}
	return nil
}

func (a *SleepaceAPI) SetReportUploadType(deviceCode, userID string, uploadType int) error {
	req := SleepaceRequest{
		Token: a.token,
		Data:  map[string]any{"userId": userID, "deviceId": deviceCode, "reportUploadType": uploadType},
	}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/reportUploadType/set")
	if resp.Status != 0 {
		return errors.New(resp.Msg)
	}
	return nil
}

func (a *SleepaceAPI) SetReportUploadTime(deviceCode, userID string, uploadTime int) error {
	req := SleepaceRequest{
		Token: a.token,
		Data:  map[string]any{"userId": userID, "deviceId": deviceCode, "leftRight": 0, "reportUploadTime": uploadTime},
	}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/setReportUploadTime")
	if resp.Status != 0 {
		return errors.New(resp.Msg)
	}
	return nil
}

// DailyMaxReportQuery 对应厂家 get24HourDailyWithMaxReport 的 data 段（键名 userId / startTime / endTime）。
type DailyMaxReportQuery struct {
	// UserID 即厂家 data.userId；取值必须是 devices.device_id（UUID），不是 device_uid、不是 device_store.device_code。
	UserID    string
	StartTime int64
	EndTime   int64
}

// Get24HourDailyWithMaxReport 调用 sleepace-service POST /sleepace/get24HourDailyWithMaxReport。
func (a *SleepaceAPI) Get24HourDailyWithMaxReport(q DailyMaxReportQuery) ([]json.RawMessage, error) {
	if q.UserID == "" {
		return nil, errors.New("DailyMaxReportQuery.UserID required (devices.device_id UUID → manufacturer data.userId)")
	}
	data := map[string]any{
		"userId":    q.UserID,
		"startTime": q.StartTime,
		"endTime":   q.EndTime,
	}
	req := SleepaceRequest{Token: a.token, Data: data}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/get24HourDailyWithMaxReport")
	if resp.Status != 0 {
		return nil, errors.New(resp.Msg)
	}
	var out []json.RawMessage
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (a *SleepaceAPI) GetAlarmConfig(userID, deviceCode string) (json.RawMessage, error) {
	req := SleepaceRequest{Token: a.token, Data: map[string]any{"userId": userID, "deviceId": deviceCode}}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/getalarmnotifyconfig")
	if resp.Status != 0 {
		return nil, errors.New(resp.Msg)
	}
	return resp.Data, nil
}

func (a *SleepaceAPI) UpdateAlarmConfig(data interface{}) error {
	req := struct {
		Token *SleepaceToken `json:"token"`
		Data  interface{}    `json:"data"`
	}{Token: a.token, Data: data}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/updatealarmnotifyconfig")
	if resp.Status != 0 {
		return errors.New(resp.Msg)
	}
	return nil
}

func (a *SleepaceAPI) SetBedParameters(deviceCode, userID string, thickness, material int) error {
	req := SleepaceRequest{
		Token: a.token,
		Data: map[string]any{
			"userId": userID, "deviceId": deviceCode, "leftRight": 0, "thickness": thickness, "material": material,
		},
	}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/updateSetting")
	if resp.Status != 0 {
		return errors.New(resp.Msg)
	}
	return nil
}

func (a *SleepaceAPI) GetLeavingMode(userID, deviceCode string) (int, error) {
	req := SleepaceRequest{
		Token: a.token,
		Data:  map[string]any{"userId": userID, "deviceId": deviceCode, "leftRight": 0},
	}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/device/getAlgMode")
	if resp.Status != 0 {
		return 0, errors.New(resp.Msg)
	}
	d := struct {
		AloMode int `json:"aloMode"`
	}{}
	if err := json.Unmarshal(resp.Data, &d); err != nil {
		return 0, err
	}
	return d.AloMode, nil
}

// SleepaceDeviceInfo holds the mapping returned by sleepace-service.
//
//	PlaintextId = hardware label (e.g. "BM87224601903") = device_store.device_uid
//	DeviceId    = sleepace platform ID (e.g. "1ua3erivl9pv1") = device_store.device_code
type SleepaceDeviceInfo struct {
	DeviceId    string `json:"deviceId"`
	PlaintextId string `json:"plaintextId"`
	DeviceType  int    `json:"deviceType"`
	Version     string `json:"deviceVersion"` // 厂家：设备版本 string
}

// GetDeviceInfoByPlaintextId queries sleepace-service by hardware label (BM87...).
// Returns the sleepace platform deviceId that will appear in MQTT messages.
func (a *SleepaceAPI) GetDeviceInfoByPlaintextId(plaintextId string) (*SleepaceDeviceInfo, error) {
	req := SleepaceRequest{
		Token: a.token,
		Data:  map[string]any{"plaintextId": plaintextId},
	}
	resp := SleepaceResponse{}
	_, err := a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/deviceInfo/plaintextId")
	if err != nil {
		return nil, err
	}
	if resp.Status != 0 {
		return nil, errors.New(resp.Msg)
	}
	var info SleepaceDeviceInfo
	if err := json.Unmarshal(resp.Data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// GetDeviceInfoByDeviceId queries sleepace-service by platform ID (1ua3erivl9pv1).
// Returns the hardware plaintextId (BM87...).
func (a *SleepaceAPI) GetDeviceInfoByDeviceId(deviceId string) (*SleepaceDeviceInfo, error) {
	req := SleepaceRequest{
		Token: a.token,
		Data:  map[string]any{"deviceId": deviceId},
	}
	resp := SleepaceResponse{}
	_, err := a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/deviceInfo/deviceId")
	if err != nil {
		return nil, err
	}
	if resp.Status != 0 {
		return nil, errors.New(resp.Msg)
	}
	var info SleepaceDeviceInfo
	if err := json.Unmarshal(resp.Data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// InitializeDeviceByCode 绑定设备。Sleepace API 仅需：device_id = device_code (wisefido), userid = userID (wisefido device_id UUID)。
// 使用 device_code 调用 bind 并做后续配置，返回 device_code。
func (a *SleepaceAPI) InitializeDeviceByCode(deviceCode, userID string, timezoneSeconds int) (string, error) {
	tz := timezoneSeconds
	if tz == 0 {
		tz = a.cfg.Timezone
	}
	if err := a.BindDevice(deviceCode, userID, tz); err != nil {
		return "", err
	}
	if err := a.SetHeartMode(deviceCode); err != nil {
		return deviceCode, err
	}
	if err := a.SetRealtimeInterval(deviceCode, userID, a.cfg.RealtimeInterval); err != nil {
		return deviceCode, err
	}
	if err := a.SetLeaveSensibility(deviceCode, userID, a.cfg.LeaveSensibility); err != nil {
		return deviceCode, err
	}
	if err := a.SetReportUploadType(deviceCode, userID, a.cfg.ReportUploadType); err != nil {
		return deviceCode, err
	}
	if a.cfg.ReportUploadType == 0 {
		if err := a.SetReportUploadTime(deviceCode, userID, a.cfg.ReportUploadTime); err != nil {
			return deviceCode, err
		}
	}
	return deviceCode, nil
}
