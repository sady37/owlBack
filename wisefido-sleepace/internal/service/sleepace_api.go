package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

// ConnectionStatusData 厂家 /sleepace/connectioStatus 返回；connectionStatus 可能为数字或字符串。
type ConnectionStatusData struct {
	ConnectionStatus int
}

func (d *ConnectionStatusData) UnmarshalJSON(b []byte) error {
	var aux struct {
		V interface{} `json:"connectionStatus"`
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	switch x := aux.V.(type) {
	case nil:
		return nil
	case float64:
		d.ConnectionStatus = int(x)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(x))
		if err != nil {
			return err
		}
		d.ConnectionStatus = n
	default:
		return fmt.Errorf("connectionStatus: unexpected JSON type %T", x)
	}
	return nil
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

// PushTypeConfig /sleepace/system/pushType/get 返回 data 结构。
type PushTypeConfig struct {
	Channel    int    `json:"channel"`
	PushType   string `json:"pushType"`   // MQTT / HTTP
	PushURL    string `json:"pushUrl"`
	CreateTime int64  `json:"createTime"` // 毫秒
	UpdateTime int64  `json:"updateTime"`
}

// GetPushType 查询当前数据推送配置。厂家 /sleepace/system/pushType/get 不要 data。
func (a *SleepaceAPI) GetPushType() (*PushTypeConfig, error) {
	req := SleepaceRequest{Token: a.token}
	resp := SleepaceResponse{}
	if _, err := a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/system/pushType/get"); err != nil {
		return nil, err
	}
	if resp.Status != 0 {
		return nil, errors.New(resp.Msg)
	}
	if len(resp.Data) == 0 || string(resp.Data) == "null" {
		return nil, nil
	}
	var out PushTypeConfig
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return &out, nil
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

// GetHeartMode 查询心率/呼吸率计算模式（BM8701-2 专用，固件 ≥v6.30）。
// mode=0：计算不出时保留之前值或加小波动；mode=1：计算不出上报 255。
func (a *SleepaceAPI) GetHeartMode(deviceCode string) (int, error) {
	req := SleepaceRequest{Token: a.token, Data: map[string]any{"deviceId": deviceCode}}
	resp := struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
		Data   struct {
			Mode int `json:"mode"`
		} `json:"data"`
	}{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/heartModeGet")
	if resp.Status != 0 {
		return 0, errors.New(resp.Msg)
	}
	return resp.Data.Mode, nil
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

// GetReportUploadType 查询当前历史数据存储方式。0=24 小时定时上报，1=离床 1h 自动生成/stopMonitor 触发。
// BM8701-2 专用（固件 ≥6.60）。
func (a *SleepaceAPI) GetReportUploadType(deviceCode string) (int, error) {
	req := SleepaceRequest{Token: a.token, Data: map[string]any{"deviceId": deviceCode}}
	resp := struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
		Data   struct {
			ReportUploadType int `json:"reportUploadType"`
		} `json:"data"`
	}{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/reportUploadType/get")
	if resp.Status != 0 {
		return 0, errors.New(resp.Msg)
	}
	return resp.Data.ReportUploadType, nil
}

// StopMonitor 强制触发当前监测结束并生成报告（异步：报告随后通过 analysis MQ 事件到达）。
// 仅在历史数据存储方式为 mode=1（离床 1h 自动生成）时可用；mode=0 下调用厂家会返回错误。
// leftRight：单人设备值为 0；双人设备 0/1（BM8701-2 养老场景固定 0）。
func (a *SleepaceAPI) StopMonitor(deviceCode string, leftRight int) error {
	req := SleepaceRequest{
		Token: a.token,
		Data:  map[string]any{"deviceId": deviceCode, "leftRight": leftRight},
	}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/stopMonitor")
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

// GetReportUploadTime 查询设备当前报告上传整点（1-24），对应厂家 /sleepace/getReportUploadTime。
func (a *SleepaceAPI) GetReportUploadTime(deviceCode, userID string) (int, error) {
	req := SleepaceRequest{
		Token: a.token,
		Data:  map[string]any{"userId": userID, "deviceId": deviceCode, "leftRight": 0},
	}
	resp := struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
		Data   struct {
			ReportUploadTime int `json:"reportUploadTime"`
		} `json:"data"`
	}{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/getReportUploadTime")
	if resp.Status != 0 {
		return 0, errors.New(resp.Msg)
	}
	return resp.Data.ReportUploadTime, nil
}

// DailyMaxReportQuery 对应厂家 get24HourDailyWithMaxReport 的 data 段（键名 userId / startTime / endTime）。
type DailyMaxReportQuery struct {
	// UserID 即厂家 data.userId = device_factory_meta.device_uid (logMAC)。
	UserID    string
	StartTime int64
	EndTime   int64
}

// Get24HourDailyWithMaxReport 调用 sleepace-service POST /sleepace/get24HourDailyWithMaxReport。
func (a *SleepaceAPI) Get24HourDailyWithMaxReport(q DailyMaxReportQuery) ([]json.RawMessage, error) {
	if q.UserID == "" {
		return nil, errors.New("DailyMaxReportQuery.UserID required (= device_uid logMAC → manufacturer data.userId)")
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

// GetConnectionStatus 主动查询设备在线状态。
// 注意：厂家接口路径为 /sleepace/connectioStatus（文档原样拼写）。
// 入参 deviceCode=合作方 deviceId（即 device_store.device_code）。
func (a *SleepaceAPI) GetConnectionStatus(deviceCode string) (int, error) {
	req := SleepaceRequest{
		Token: a.token,
		Data:  map[string]any{"deviceId": deviceCode},
	}
	resp := SleepaceResponse{}
	_, err := a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/connectioStatus")
	if err != nil {
		return 0, err
	}
	if resp.Status != 0 {
		return 0, errors.New(resp.Msg)
	}
	var data ConnectionStatusData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return 0, err
	}
	return data.ConnectionStatus, nil
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

// BedSetting /sleepace/getSetting 返回 data：床垫厚度+材质。
// thickness 1-3（1:5-10cm, 2:11-20cm, 3:21-30cm）；material 1-5（1 海绵 2 弹簧 3 乳胶 4 气垫 5 其他）。
type BedSetting struct {
	Thickness int `json:"thickness"`
	Material  int `json:"material"`
}

// GetBedParameters 查询床垫参数。
func (a *SleepaceAPI) GetBedParameters(deviceCode, userID string) (*BedSetting, error) {
	req := SleepaceRequest{
		Token: a.token,
		Data:  map[string]any{"userId": userID, "deviceId": deviceCode, "leftRight": 0},
	}
	resp := struct {
		Status int        `json:"status"`
		Msg    string     `json:"msg"`
		Data   BedSetting `json:"data"`
	}{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/getSetting")
	if resp.Status != 0 {
		return nil, errors.New(resp.Msg)
	}
	out := resp.Data
	return &out, nil
}

// SetUserMode 单双人模式设置（SDC100 专用）：mode 0=单人 1=双人。
func (a *SleepaceAPI) SetUserMode(deviceCode string, mode int) error {
	req := SleepaceRequest{
		Token: a.token,
		Data:  map[string]any{"deviceId": deviceCode, "mode": mode},
	}
	resp := SleepaceResponse{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/userMode/set")
	if resp.Status != 0 {
		return errors.New(resp.Msg)
	}
	return nil
}

// GetUserMode 单双人模式获取（SDC100 专用）。
func (a *SleepaceAPI) GetUserMode(deviceCode string) (int, error) {
	req := SleepaceRequest{
		Token: a.token,
		Data:  map[string]any{"deviceId": deviceCode},
	}
	resp := struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
		Data   struct {
			Mode int `json:"mode"`
		} `json:"data"`
	}{}
	a.client.R().SetBody(req).SetResult(&resp).Post("/sleepace/userMode/get")
	if resp.Status != 0 {
		return 0, errors.New(resp.Msg)
	}
	return resp.Data.Mode, nil
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
//
// 厂家 deviceInfo/deviceId 实际返回字段名：
//
//	{"devicePlaintextId":"BM87...", "deviceId":"...", "version":"6.89"}
//
// 注意：plaintextId 字段在 by-deviceId 接口里实际是 `devicePlaintextId`；
// version 字段是 `version` 不是 `deviceVersion`（与文档不一致，2026-04-27 实测）。
type SleepaceDeviceInfo struct {
	DeviceId    string `json:"deviceId"`
	PlaintextId string `json:"devicePlaintextId"`
	DeviceType  int    `json:"deviceType"`
	Version     string `json:"version"` // 厂家实际字段名 "version"，不是文档说的 deviceVersion
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

// InitializeDeviceByCode 绑定设备。Sleepace API 仅需：device_id = device_code (wisefido), userid = device_uid (logMAC)。
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
