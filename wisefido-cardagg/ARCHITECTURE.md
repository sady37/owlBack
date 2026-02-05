# wisefido-cardagg Architecture Summary

## Project Status
✅ **Build Successful** - All layered architecture components compiled and integrated

## Architecture Overview

```
main.go (Entry Point)
  ├─ Initialize Redis Client
  ├─ Create CacheRepository (Redis-backed)
  ├─ Create MonitorService (Business Logic)
  ├─ Create MonitorHandler (Consumer)
  └─ Start 3 Stream Consumers (monitor/event/alarm)

Domain Layer (internal/domain/)
  └─ monitor.go
     ├─ MonitorValidationResult
     ├─ MonitorContext
     └─ ValidateMonitorMessage() - checks timestamp (≤300s), card_id (non-empty)

Repository Layer (internal/repository/)
  ├─ repository.go - CacheRepository interface
  └─ realtime_cache.go - Redis implementation
     ├─ GetRealtimeData/SetRealtimeData
     ├─ GetDevicePosture/SetDevicePosture
     ├─ GetVitalSimplified/SetVitalSimplified
     └─ All use JSON marshaling with 10s TTL

Service Layer (internal/service/)
  └─ monitor_service.go - MonitorService
     ├─ ProcessMonitor() - orchestrates full pipeline
     ├─ processTrackData() - extracts postures (0-11), stores DevicePosture
     ├─ processVitalData() - extracts vital signs, stores VitalSimplified
     ├─ parseCategoryCount() - parses "track2.vital1" format
     ├─ poseStringToInt() - maps pose strings to 0-11
     └─ extractIntPtr/extractStringPtr - safe JSON field access

Consumer Layer (internal/consumer/)
  └─ monitor_handler.go - MonitorHandler
     ├─ Handle() - deserializes StreamMessage to IoTStreamMessage
     ├─ Calls ValidateMonitorMessage()
     └─ Delegates to service.ProcessMonitor()
```

## Processing Pipeline

### Monitor Stream Processing
1. **Redis Stream** → Receive IoTStreamMessage (device_id, timestamp, category, data_value[])
2. **Domain Validation** → Check timestamp (now - ts ≤ 300s), card_id non-empty
3. **Track Processing** → Parse category, extract poses, convert to 0-11 ints, store DevicePosture
4. **Vital Processing** → Extract vital fields (HR, RR, sleep_status, stability), store VitalSimplified
5. **Cache Storage** → Redis keys with 10s TTL
   - `vital-focus:card:{cardID}:realtime`
   - `vital-focus:card:{cardID}:posture:{deviceID}`
   - `vital-focus:card:{cardID}:vital:{deviceID}`

### Key Processing Rules
- **Timestamp Validation**: Drop if older than 300 seconds
- **Card ID Validation**: Drop if empty/null
- **Category Format**: Parse "track{N}.vital{M}" (e.g., "track2.vital1" = 2 poses, 1 vital)
- **Posture Mapping**: 
  - 0: Initialization, 1: Walking, 2: Running, 3: Fast Walking
  - 4: Turning, 5: Falling, 6: BedStatic, 7: BedTurn
  - 8: SittingStatic, 9: Bending, 10: BedSitDown, 11: BedSitUp
- **Data Extraction**: Type-safe conversion with nil checks for missing fields

## Cache Key Naming Convention
- **RealtimeData**: `vital-focus:card:{cardID}:realtime`
- **DevicePosture**: `vital-focus:card:{cardID}:posture:{deviceID}`
- **VitalSimplified**: `vital-focus:card:{cardID}:vital:{deviceID}`

## Stream Subscriptions
1. **iot:monitor:stream** → Full processing pipeline (domain → service → cache)
2. **iot:event:stream** → Stub logger (ready for implementation)
3. **iot:alarm:stream** → Stub logger (ready for implementation)

## File Structure
```
wisefido-cardagg/
├─ main.go                                 # Entry point
├─ go.mod                                  # Module definition with owl-common replace
├─ go.sum
├─ internal/
│  ├─ domain/
│  │  └─ monitor.go                        # Domain model & validation
│  ├─ repository/
│  │  ├─ repository.go                     # CacheRepository interface
│  │  └─ realtime_cache.go                 # Redis implementation
│  ├─ service/
│  │  └─ monitor_service.go                # Business logic orchestration
│  └─ consumer/
│     └─ monitor_handler.go                # Message deserialization & routing
```

## Next Steps
1. **Event Stream Implementation** - Create event_service.go and event_handler.go with similar pattern
2. **Alarm Stream Implementation** - Create alarm_service.go and alarm_handler.go
3. **Integration Testing** - Test with actual Redis stream messages
4. **Error Handling Refinement** - Add retry logic and dead-letter handling
5. **Monitoring & Logging** - Add comprehensive logging and metrics

## Build Status
```
✅ go build ./... 
   No errors - all packages compile successfully
```

## Dependencies
- `github.com/go-redis/redis/v8` - Redis client
- `go.uber.org/zap` - Logging
- `owl-common` (local monorepo replace) - Shared types and utilities
