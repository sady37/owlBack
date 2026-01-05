# Card Update Architecture Options

## Current Architecture (Event-Driven + Polling)

```
wisefido-data (发布事件)
    ↓ (Redis Streams)
wisefido-card-aggregator (消费事件 + 轮询)
    ↓ (更新数据库)
PostgreSQL (cards 表)
```

**问题**：
- 两个独立的服务，需要维护事件系统
- 并行运行事件驱动和轮询模式不够优雅
- 如果事件丢失，需要等待轮询（最多60分钟延迟）

## Option 1: Direct Synchronous Call (Recommended)

### Architecture
```
wisefido-data (同步调用)
    ↓ (直接调用)
Card Creator (同一进程)
    ↓ (更新数据库)
PostgreSQL (cards 表)
```

### Implementation Approach

#### Approach A: Extract to Shared Package
1. 将 `card_creator` 逻辑提取到 `owl-common` 包
2. `wisefido-data` 和 `wisefido-card-aggregator` 都使用共享包
3. `wisefido-data` 在数据变化时直接调用 `CreateCardsForUnit`

**优点**：
- ✅ 代码复用，避免重复
- ✅ 同步更新，实时性强
- ✅ 不需要事件系统，架构简单
- ✅ 错误处理直接，可以立即返回错误

**缺点**：
- ⚠️ 需要重构代码结构
- ⚠️ `wisefido-data` 的 HTTP 请求会等待卡片更新完成（但通常很快，<100ms）

#### Approach B: Duplicate Logic in wisefido-data
1. 在 `wisefido-data` 中复制卡片更新逻辑
2. 数据变化时直接调用

**优点**：
- ✅ 不需要重构现有代码
- ✅ 实现快速

**缺点**：
- ❌ 代码重复，维护成本高
- ❌ 两个地方需要同步更新逻辑

### Recommended: Approach A (Extract to Shared Package)

```go
// owl-common/card/card_creator.go
package card

type CardCreator struct {
    repo   CardRepositoryInterface
    logger *zap.Logger
}

func (c *CardCreator) CreateCardsForUnit(tenantID, unitID string) (*CardUpdateStats, error) {
    // 现有逻辑
}
```

```go
// wisefido-data/internal/service/device_service.go
import (
    "owl-common/card"
)

func (s *deviceService) UpdateDevice(ctx context.Context, req *UpdateDeviceRequest) (*UpdateDeviceResponse, error) {
    // ... 更新设备 ...
    
    // 如果绑定关系变化，直接更新卡片
    if req.UpdateBoundRoomID || req.UpdateBoundBedID {
        if newDevice.UnitID.Valid {
            _, err := s.cardCreator.CreateCardsForUnit(req.TenantID, newDevice.UnitID.String)
            if err != nil {
                s.logger.Warn("Failed to update cards after device binding change",
                    zap.Error(err),
                    zap.String("unit_id", newDevice.UnitID.String),
                )
                // 不返回错误，只记录警告（卡片更新失败不影响设备更新）
            }
        }
    }
    
    return response, nil
}
```

## Option 2: Keep Event-Driven, Remove Polling

### Architecture
```
wisefido-data (发布事件)
    ↓ (Redis Streams)
wisefido-card-aggregator (只消费事件，不轮询)
    ↓ (更新数据库)
PostgreSQL (cards 表)
```

**优点**：
- ✅ 解耦，异步处理
- ✅ 不阻塞 HTTP 请求

**缺点**：
- ❌ 如果事件丢失，卡片不会更新
- ❌ 需要维护事件系统
- ❌ 需要确保事件可靠性

## Option 3: Keep Polling Only, Remove Event-Driven

### Architecture
```
wisefido-data (不发布事件)
wisefido-card-aggregator (只轮询，每60分钟)
    ↓ (更新数据库)
PostgreSQL (cards 表)
```

**优点**：
- ✅ 简单，不需要事件系统
- ✅ 可靠，不会丢失更新

**缺点**：
- ❌ 延迟高（最多60分钟）
- ❌ 即使没有变化也会执行检查

## Comparison

| 方案 | 实时性 | 复杂度 | 可靠性 | 优雅性 |
|------|--------|--------|--------|--------|
| Option 1 (Direct Call) | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| Option 2 (Event Only) | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| Option 3 (Polling Only) | ⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| Current (Event + Polling) | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ |

## Recommendation

**推荐 Option 1 (Direct Synchronous Call)**，理由：

1. **最优雅**：架构简单，不需要事件系统
2. **实时性强**：数据变化立即更新卡片
3. **可靠性高**：同步调用，不会丢失
4. **错误处理直接**：可以立即知道更新是否成功
5. **性能可接受**：卡片更新通常很快（<100ms），不会显著影响 HTTP 响应时间

### Implementation Steps

1. 将 `card_creator` 逻辑提取到 `owl-common/card` 包
2. 将 `card_repository` 接口也提取到共享包
3. 在 `wisefido-data` 中初始化 `CardCreator`
4. 在 `device_service`, `resident_service`, `unit_service` 中直接调用 `CreateCardsForUnit`
5. 可选：保留 `wisefido-card-aggregator` 作为独立的定时任务（用于全量重建或修复）

