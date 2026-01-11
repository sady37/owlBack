# VitalFocusService 设计文档

## 📋 VitalFocus 数据查询需求分析

### 1. API 端点

| 端点 | 方法 | 功能 | 数据源 |
|------|------|------|--------|
| `/data/api/v1/data/vital-focus/cards` | GET | 获取卡片列表 | Redis 缓存（`vital-focus:card:*:full`） |
| `/data/api/v1/data/vital-focus/card/:id` | GET | 获取卡片详情 | Redis 缓存（支持 card_id 和 resident_id） |
| `/data/api/v1/data/vital-focus/selection` | POST | 保存用户选择 | Redis（`vital-focus:selection:user:{X-User-Id}`） |

### 2. 数据转换需求

**decodeAndNormalizeFullCard** 需要处理：
- 字段类型规范化（device_type: string → number）
- 数据源规范化（heart_source/breath_source: "Sleepace"/"Radar" → "s"/"r"/"-"）
- 住户数据规范化（last_name 必填，从 nickname 填充）
- 错误处理（JSON 解析失败、字段缺失）

### 3. 权限检查需求

- tenant_id 过滤（只返回当前租户的卡片）
- 用户选择保存（需要 user_id）

---

## 🏗️ VitalFocusService 设计

### 接口定义

```go
package service

import (
    "context"
    "wisefido-data/internal/models"
    "wisefido-data/internal/store"
    "go.uber.org/zap"
)

type VitalFocusService struct {
    kv     store.KV
    logger *zap.Logger
}

func NewVitalFocusService(kv store.KV, logger *zap.Logger) *VitalFocusService {
    return &VitalFocusService{
        kv:     kv,
        logger: logger,
    }
}
```

### 方法定义

```go
// GetCards 获取卡片列表
func (s *VitalFocusService) GetCards(
    ctx context.Context,
    tenantID string,
    page, size int,
) (*models.GetVitalFocusCardsModel, error) {
    // 1. 从 Redis 扫描 full cache
    keys, err := s.kv.ScanKeys(ctx, "vital-focus:card:*:full")
    if err != nil {
        // 联调友好：返回空列表
        s.logger.Warn("ScanKeys failed, returning empty cards list", zap.Error(err))
        return &models.GetVitalFocusCardsModel{
            Items: []models.VitalFocusCard{},
            Pagination: models.BackendPagination{
                Size:  size,
                Page:  page,
                Count: 0,
            },
        }, nil
    }
    
    // 2. 读取并规范化所有卡片
    all := make([]models.VitalFocusCard, 0, len(keys))
    for _, key := range keys {
        raw, err := s.kv.Get(ctx, key)
        if err != nil {
            continue
        }
        card, ok := s.normalizeCard(raw)
        if !ok {
            continue
        }
        // 3. tenant_id 过滤
        if tenantID != "" && card.TenantID != tenantID {
            continue
        }
        all = append(all, card)
    }
    
    // 4. 排序和分页
    s.sortCardsByID(all)
    total := len(all)
    start := (page - 1) * size
    if start > total {
        start = total
    }
    end := start + size
    if end > total {
        end = total
    }
    
    return &models.GetVitalFocusCardsModel{
        Items: all[start:end],
        Pagination: models.BackendPagination{
            Size:      size,
            Page:      page,
            Count:     total,
        },
    }, nil
}

// GetCardByIDOrResident 获取卡片详情（支持 card_id 和 resident_id）
func (s *VitalFocusService) GetCardByIDOrResident(
    ctx context.Context,
    id string,
) (*models.VitalFocusCardInfo, error) {
    // 1. 先当作 card_id 直接读取
    if card, ok := s.getCardByCardID(ctx, id); ok {
        return s.toCardInfo(card), nil
    }
    
    // 2. 再按 resident_id 查找（扫描 full cache）
    keys, err := s.kv.ScanKeys(ctx, "vital-focus:card:*:full")
    if err != nil {
        return nil, fmt.Errorf("failed to scan cards: %w", err)
    }
    
    for _, key := range keys {
        raw, err := s.kv.Get(ctx, key)
        if err != nil {
            continue
        }
        card, ok := s.normalizeCard(raw)
        if !ok {
            continue
        }
        // 检查 primary_resident_id
        if card.PrimaryResidentID != nil && *card.PrimaryResidentID == id {
            return s.toCardInfo(card), nil
        }
        // 检查 residents 列表
        for _, r := range card.Residents {
            if r.ResidentID == id {
                return s.toCardInfo(card), nil
            }
        }
    }
    
    return nil, fmt.Errorf("card not found")
}

// SaveSelection 保存用户选择
func (s *VitalFocusService) SaveSelection(
    ctx context.Context,
    userID string,
    selectedCardIDs []string,
) error {
    if userID == "" {
        userID = "anonymous"
    }
    
    key := "vital-focus:selection:user:" + userID
    data := map[string]any{
        "selected_card_ids": selectedCardIDs,
    }
    
    raw, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal selection: %w", err)
    }
    
    // 保存 7 天
    return s.kv.Set(ctx, key, string(raw), 7*24*time.Hour)
}

// normalizeCard 规范化卡片数据
func (s *VitalFocusService) normalizeCard(raw string) (models.VitalFocusCard, bool) {
    // 1. 解析 JSON
    var m map[string]any
    if err := json.Unmarshal([]byte(raw), &m); err != nil {
        return models.VitalFocusCard{}, false
    }
    
    // 2. 转换为模型
    b, err := json.Marshal(m)
    if err != nil {
        return models.VitalFocusCard{}, false
    }
    var card models.VitalFocusCard
    if err := json.Unmarshal(b, &card); err != nil {
        return models.VitalFocusCard{}, false
    }
    
    // 3. 规范化 residents（last_name 必填）
    for i := range card.Residents {
        if card.Residents[i].LastName == "" {
            if card.Residents[i].Nickname != "" {
                card.Residents[i].LastName = card.Residents[i].Nickname
            } else {
                card.Residents[i].LastName = "-"
            }
        }
    }
    
    // 4. 规范化 devices（device_type: string → number）
    for i := range card.Devices {
        switch v := card.Devices[i].DeviceType.(type) {
        case string:
            card.Devices[i].DeviceType = s.deviceTypeToNumber(v)
        case float64:
            card.Devices[i].DeviceType = int(v)
        }
    }
    
    // 5. 规范化数据源（heart_source/breath_source）
    if card.HeartSource != "" {
        card.HeartSource = s.normalizeSource(card.HeartSource)
    }
    if card.BreathSource != "" {
        card.BreathSource = s.normalizeSource(card.BreathSource)
    }
    
    return card, true
}

// deviceTypeToNumber 设备类型转换为数字
func (s *VitalFocusService) deviceTypeToNumber(s string) int {
    switch s {
    case "Sleepace", "SleepPad", "Sleepad", "SleepAd":
        return 1
    case "Radar":
        return 2
    default:
        return 0
    }
}

// normalizeSource 规范化数据源
func (s *VitalFocusService) normalizeSource(s string) string {
    switch s {
    case "s", "r", "-":
        return s
    case "Sleepace", "SleepPad":
        return "s"
    case "Radar":
        return "r"
    default:
        return "-"
    }
}
```

---

## 📋 总结

### VitalFocusService 职责

1. **权限检查**：tenant_id 过滤
2. **数据转换**：复杂的数据规范化（字段类型转换、数据源规范化、住户数据规范化）
3. **错误处理**：Redis 不可用时的友好处理
4. **业务编排**：排序、分页

### 为什么需要 Service

- **数据转换复杂**：需要处理多种数据格式不一致的情况
- **错误处理**：需要友好的错误处理（Redis 不可用时返回空列表）
- **业务逻辑**：排序、分页、tenant_id 过滤

---

## 🚀 实现优先级

**Phase 3: 中优先级**（复杂度中）
- ✅ **VitalFocusService** - 数据规范化转换、错误处理

