package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// declareAreaClient 取 wisefido-data 的固件活体 declare_area（cmd=read），抽床区
// (area_type∈{2床,5监护床}) 的 area_id 集合。替代 canvas/baseline 几何源——治"下发区域
// vs 固件活体"漂移：area_id 是固件自身槽位号，与上报 track 的 FwAreaID 同源，免 join canvas。
type declareAreaClient struct {
	baseURL string
	http    *http.Client
	logger  *zap.Logger
	cache   map[string][]int // uidHex → 床 area_id（bootstrap 一次性，replay 单轮内不变）
}

func newDeclareAreaClient(baseURL string, logger *zap.Logger) *declareAreaClient {
	return &declareAreaClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 5 * time.Second},
		logger:  logger,
		cache:   map[string][]int{},
	}
}

// bedAreaIDs 返回该 device 固件声明的床区 area_id。失败返回 error（不兜底 canvas，
// 换源后 canvas 床 area_id 路径已删；拉不到 = 真问题，由 caller 暴露）。
func (c *declareAreaClient) bedAreaIDs(ctx context.Context, uidHex string) ([]int, error) {
	if ids, ok := c.cache[uidHex]; ok {
		return ids, nil
	}
	url := fmt.Sprintf("%s/internal/radar/device/%s/declare-area", c.baseURL, uidHex)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("declare-area GET %s: %w", uidHex, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("declare-area %s status %d: %s", uidHex, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var props map[string]interface{}
	if err := json.Unmarshal(body, &props); err != nil {
		return nil, fmt.Errorf("declare-area %s decode: %w", uidHex, err)
	}
	da, _ := props["declare_area"].(string)
	ids := parseBedAreaIDs(da)
	c.cache[uidHex] = ids
	c.logger.Info("declare_area_fetched",
		zap.String("uid", uidHex), zap.Ints("bed_area_ids", ids), zap.String("raw", da))
	return ids, nil
}

// parseBedAreaIDs 解析 GetOriginalProperties 的 cm 全逗号 declare_area
// `{id,type,x1,y1,x2,y2,x3,y3,x4,y4},{...}`，取 area_type∈{2,5} 的 area_id。
// 只读前两段(id,type)，几何顶点丢弃（覆盖判定只需 area_id 身份）。注意：data 侧已 cm，
// 不复用 radar.ParseDeclareAreaFromDM（那会把输入当 DM 再 ×10，单位翻倍）。
func parseBedAreaIDs(s string) []int {
	var out []int
	for _, seg := range strings.Split(s, "},{") {
		seg = strings.Trim(seg, "{}, ")
		if seg == "" {
			continue
		}
		f := strings.SplitN(seg, ",", 3)
		if len(f) < 2 {
			continue
		}
		id, err1 := strconv.Atoi(strings.TrimSpace(f[0]))
		typ, err2 := strconv.Atoi(strings.TrimSpace(f[1]))
		if err1 != nil || err2 != nil {
			continue
		}
		if typ == 2 || typ == 5 {
			out = append(out, id)
		}
	}
	return out
}
