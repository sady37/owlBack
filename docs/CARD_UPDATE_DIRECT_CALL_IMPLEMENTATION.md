# Card Update Direct Call Implementation

## Overview

This document describes the implementation of direct synchronous card updates in `wisefido-data`, replacing the previous event-driven architecture.

## Architecture Change

### Before (Event-Driven)
```
wisefido-data (发布事件)
    ↓ (Redis Streams)
wisefido-card-aggregator (消费事件)
    ↓ (更新数据库)
PostgreSQL (cards 表)
```

### After (Direct Synchronous Call)
```
wisefido-data (同步调用)
    ↓ (直接调用)
Card Creator (同一进程, owl-common/card)
    ↓ (更新数据库)
PostgreSQL (cards 表)
```

## Implementation Details

### 1. Shared Card Package (`owl-common/card`)

Created a shared package containing:
- **Types** (`types.go`): `ActiveBedInfo`, `UnitInfo`, `DeviceInfo`, `ResidentInfo`, `CardWithContent`, `ExpectedCard`, `CardUpdateStats`
- **Repository Interface** (`repository.go`): `RepositoryInterface` defining all card-related database operations
- **Card Creator** (`creator.go`): `CardCreator` with `CreateCardsForUnit` method
- **Utils** (`utils.go`): `ConvertDevicesToJSON`, `ConvertResidentsToJSON`

### 2. Wisefido-Data Integration

#### Card Repository Implementation
- **File**: `wisefido-data/internal/repository/postgres_card.go`
- **Type**: `PostgresCardRepository` implements `card.RepositoryInterface`
- Uses `wisefido-data`'s database connection

#### Service Layer Integration
- **Device Service** (`device_service.go`):
  - Removed `redisClient` field
  - Added `cardCreator *card.CardCreator` field
  - In `UpdateDevice`: Directly calls `cardCreator.CreateCardsForUnit` when binding changes
  - Removed all event publishing code

- **Unit Service** (`unit_service.go`):
  - Replaced `redisClient` with `cardCreator`
  - In `UpdateUnit`: Directly calls `cardCreator.CreateCardsForUnit` after unit updates
  - Removed event publishing code

- **Resident Service** (`resident_service.go`):
  - Replaced `redisClient` with `cardCreator`
  - In `UpdateResident`: Directly calls `cardCreator.CreateCardsForUnit` after resident updates
  - Removed event publishing code

#### Main Entry Point
- **File**: `wisefido-data/cmd/wisefido-data/main.go`
- Initializes `PostgresCardRepository` and `CardCreator`
- Passes `cardCreator` to `DeviceService`, `UnitService`, and `ResidentService`

### 3. Wisefido-Card-Aggregator Updates

- **Repository** (`card.go`): Updated to use `owl-common/card` types
- **Service** (`aggregator.go`): Updated to use `card.CardCreator` from `owl-common/card`
- **Consumer** (`event_consumer.go`): Updated to use `card.CardCreator` (for event-driven mode, if still needed)
- **Data Aggregator** (`data_aggregator.go`): Updated to use `card.DeviceInfo` and `card.ResidentInfo`

### 4. Database Schema Compliance

- Removed all references to deleted fields:
  - `routing_alarm_tags` (removed from `cards` table)
  - `routing_alarm_user_ids` (removed from `cards` table)
  - `GroupList` and `UserList` from `UnitInfo` (removed from `units` table)

## Benefits

1. **Simpler Architecture**: No event system needed, direct synchronous calls
2. **Real-time Updates**: Cards are updated immediately when data changes
3. **Better Error Handling**: Errors are returned directly, no need to handle event failures
4. **Code Reuse**: Shared `owl-common/card` package used by both services
5. **Type Safety**: All types defined in one place, ensuring consistency

## Migration Notes

- **Event Publishing Removed**: All `PublishToStream` calls have been removed from service layers
- **Redis Client Removed**: `redisClient` field removed from services (except where still needed for other purposes)
- **Card Creator Required**: All services that update cards now require `cardCreator` parameter
- **Backward Compatibility**: `wisefido-card-aggregator` still supports event-driven mode (for polling fallback), but `wisefido-data` no longer publishes events

## Testing

Both services compile successfully:
- ✅ `wisefido-data`: `go build ./cmd/wisefido-data`
- ✅ `wisefido-card-aggregator`: `go build ./cmd/wisefido-card-aggregator`

## Next Steps

1. Test card updates when devices are bound/unbound
2. Test card updates when units are modified
3. Test card updates when residents are bound/unbound
4. Verify card comparison logic prevents unnecessary `card_id` changes

