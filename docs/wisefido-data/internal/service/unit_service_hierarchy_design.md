# ListUnitsWithFullHierarchy 接口设计

## 架构说明

**当前架构**：
- Service 层返回 `domain.Unit`, `domain.Room`, `domain.Bed` 等 domain 对象
- Handler 层使用 `unitToJSON()`, `roomToJSON()`, `bedToJSON()` 等函数将 domain 对象转换为 JSON map
- 前端通过 `getUnitsApi()` 获取 units，然后对每个 unit 调用 `getRoomsApi()`，再单独调用 `getDevicesApi()`

**新接口架构**：
- Service 层返回 `ListUnitsWithFullHierarchyResponse`，包含 `UnitWithFullHierarchy[]`
- Handler 层新增 `unitWithFullHierarchyToJSON()` 函数将 Service 返回的结构转换为 JSON
- 前端调用新的 API（`getUnitsWithFullHierarchyApi()`），直接获取完整层级结构

## 1. 输入参数 (Request)

```go
type ListUnitsWithFullHierarchyRequest struct {
    TenantID   string  // 必填
    BranchID   *string // 可选（优先使用，如果提供则忽略 BranchName）
    BranchName *string // 可选（向后兼容，如果 BranchID 未提供则使用此字段）
    BuildingID *string // 可选（优先使用，UUID 类型，如果提供则忽略 Building）
    Building   *string // 可选（向后兼容，如果 BuildingID 未提供则使用此字段）
    Floor      *string // 可选
    UnitType   *string // 可选
    Search     *string // 可选（模糊搜索 unit_name）
    // 注意：不分页，返回所有匹配的 units（因为前端需要完整层级结构）
}
```

## 2. 输出结构 (Response)

```go
type ListUnitsWithFullHierarchyResponse struct {
    Items []*UnitWithFullHierarchy `json:"items"`
    Total int                      `json:"total"`
}

// UnitWithFullHierarchy 包含完整的层级结构
type UnitWithFullHierarchy struct {
    *domain.Unit // Unit 基本信息
    Rooms []*RoomWithBedsAndDevices `json:"rooms"`
}

// RoomWithBedsAndDevices 房间及其床位和设备
type RoomWithBedsAndDevices struct {
    *domain.Room // Room 基本信息
    Beds         []*BedWithDevices `json:"beds"`
    DeviceIDs    []string          `json:"device_ids"`    // 绑定到 room 的 device IDs（用于前端选中并向后端传递）
    DeviceNames  []string          `json:"device_names"` // 绑定到 room 的 device names（用于前端显示）
}

// BedWithDevices 床位及其设备
type BedWithDevices struct {
    *domain.Bed // Bed 基本信息
    DeviceIDs   []string `json:"device_ids"`   // 绑定到 bed 的 device IDs（用于前端选中并向后端传递）
    DeviceNames []string `json:"device_names"`  // 绑定到 bed 的 device names（用于前端显示）
}
```

## 3. SQL 查询策略

### 方案 A：分步查询（推荐，性能更好）

**步骤 1：查询 Units（复用现有的 ListUnits 逻辑）**
```sql
SELECT 
    u.unit_id::text,
    u.tenant_id::text,
    u.branch_id::text,
    COALESCE(b.branch_name, NULL) as branch_name,
    u.unit_name,
    u.building_id::text,
    COALESCE(bd.building_name, NULL) as building_name,
    u.floor,
    u.layout_config::text,
    u.unit_type,
    u.is_public,
    u.is_shared_unit,
    u.timezone
FROM units u
LEFT JOIN branches b ON b.branch_id = u.branch_id
LEFT JOIN buildings bd ON bd.building_id = u.building_id
WHERE u.tenant_id = $1 AND [其他过滤条件]
ORDER BY u.floor, u.unit_name
```

**步骤 2：批量查询 Rooms（使用 IN 子句）**
```sql
SELECT 
    r.room_id::text,
    r.tenant_id::text,
    r.unit_id::text,
    u.unit_name,
    u.floor,
    r.room_name,
    r.layout_config::text
FROM rooms r
LEFT JOIN units u ON u.unit_id = r.unit_id
WHERE r.tenant_id = $1 AND r.unit_id = ANY($2::uuid[])
ORDER BY r.unit_id, r.room_name
```

**步骤 3：批量查询 Beds（使用 IN 子句）**
```sql
SELECT 
    b.bed_id::text,
    b.tenant_id::text,
    b.room_id::text,
    r.room_name,
    b.bed_name,
    b.mattress_material,
    b.mattress_thickness
FROM beds b
LEFT JOIN rooms r ON r.room_id = b.room_id
WHERE b.tenant_id = $1 AND b.room_id = ANY($2::uuid[])
ORDER BY b.room_id, b.bed_name
```

**步骤 4：批量查询 Device IDs 和 Names（使用聚合函数）**
```sql
-- 查询每个 room 的 device IDs 和 names
SELECT 
    d.bound_room_id::text as room_id,
    ARRAY_AGG(d.device_id::text ORDER BY d.device_name) FILTER (WHERE d.device_id IS NOT NULL) as device_ids,
    ARRAY_AGG(d.device_name ORDER BY d.device_name) FILTER (WHERE d.device_name IS NOT NULL) as device_names
FROM devices d
WHERE d.tenant_id = $1 
  AND d.bound_room_id IS NOT NULL
  AND d.bound_room_id = ANY($2::uuid[])
GROUP BY d.bound_room_id

-- 查询每个 bed 的 device IDs 和 names
SELECT 
    d.bound_bed_id::text as bed_id,
    ARRAY_AGG(d.device_id::text ORDER BY d.device_name) FILTER (WHERE d.device_id IS NOT NULL) as device_ids,
    ARRAY_AGG(d.device_name ORDER BY d.device_name) FILTER (WHERE d.device_name IS NOT NULL) as device_names
FROM devices d
WHERE d.tenant_id = $1 
  AND d.bound_bed_id IS NOT NULL
  AND d.bound_bed_id = ANY($3::uuid[])
GROUP BY d.bound_bed_id
```

### 方案 B：单次查询（使用 JSON 聚合，数据量大时可能较慢）

```sql
SELECT 
    u.unit_id::text,
    u.tenant_id::text,
    u.branch_id::text,
    COALESCE(b.branch_name, NULL) as branch_name,
    u.unit_name,
    u.building_id::text,
    COALESCE(bd.building_name, NULL) as building_name,
    u.floor,
    u.layout_config::text,
    u.unit_type,
    u.is_public,
    u.is_shared_unit,
    u.timezone,
    COALESCE(
        json_agg(
            DISTINCT jsonb_build_object(
                'room_id', r.room_id::text,
                'tenant_id', r.tenant_id::text,
                'unit_id', r.unit_id::text,
                'room_name', r.room_name,
                'layout_config', r.layout_config::text,
                'beds', (
                    SELECT json_agg(
                        jsonb_build_object(
                            'bed_id', b2.bed_id::text,
                            'bed_name', b2.bed_name,
                            'device_names', (
                                SELECT ARRAY_AGG(d2.device_name ORDER BY d2.device_name)
                                FROM devices d2
                                WHERE d2.tenant_id = $1 
                                  AND d2.bound_bed_id = b2.bed_id
                                  AND d2.device_name IS NOT NULL
                            )
                        )
                    )
                    FROM beds b2
                    WHERE b2.tenant_id = $1 AND b2.room_id = r.room_id
                ),
                'device_names', (
                    SELECT ARRAY_AGG(d3.device_name ORDER BY d3.device_name)
                    FROM devices d3
                    WHERE d3.tenant_id = $1 
                      AND d3.bound_room_id = r.room_id
                      AND d3.device_name IS NOT NULL
                )
            )
        ) FILTER (WHERE r.room_id IS NOT NULL),
        '[]'::json
    ) as rooms
FROM units u
LEFT JOIN branches b ON b.branch_id = u.branch_id
LEFT JOIN buildings bd ON bd.building_id = u.building_id
LEFT JOIN rooms r ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
WHERE u.tenant_id = $1 AND [其他过滤条件]
GROUP BY u.unit_id, u.tenant_id, u.branch_id, b.branch_name, u.unit_name, 
         u.building_id, bd.building_name, u.floor, u.layout_config, 
         u.unit_type, u.is_public, u.is_shared_unit, u.timezone
ORDER BY u.floor, u.unit_name
```

## 4. 需要新增的 Repository 方法

```go
// 在 UnitsRepository 接口中添加
ListRoomsByUnitIDs(ctx context.Context, tenantID string, unitIDs []string) ([]*domain.Room, error)
ListBedsByRoomIDs(ctx context.Context, tenantID string, roomIDs []string) ([]*domain.Bed, error)

// 在 DevicesRepository 接口中添加
// DeviceInfo 用于返回设备的 ID 和 Name
type DeviceInfo struct {
    ID   string
    Name string
}

GetDevicesByRoomIDs(ctx context.Context, tenantID string, roomIDs []string) (map[string][]DeviceInfo, error)
GetDevicesByBedIDs(ctx context.Context, tenantID string, bedIDs []string) (map[string][]DeviceInfo, error)
```

## 5. 推荐实现方案

**推荐使用方案 A（分步查询）**，原因：
1. **性能更好**：避免复杂的 JSON 聚合，查询更简单
2. **可维护性高**：可以复用现有的 ListUnits、ListRooms、ListBeds 逻辑
3. **灵活性好**：可以单独优化每个查询
4. **数据量可控**：分步查询可以更好地控制内存使用

**实现步骤**：
1. 调用 `ListUnits` 获取所有 units
2. 批量调用 `ListRoomsByUnitIDs`（传入所有 unit_ids）
3. 批量调用 `ListBedsByRoomIDs`（传入所有 room_ids）
4. 批量查询 device IDs 和 names（使用聚合函数）

## 5. 数据组装逻辑

```go
// 伪代码
func (s *unitService) ListUnitsWithFullHierarchy(ctx context.Context, req ListUnitsWithFullHierarchyRequest) (*ListUnitsWithFullHierarchyResponse, error) {
    // 1. 查询 units
    units, total, err := s.unitsRepo.ListUnits(ctx, req.TenantID, filters, 1, 10000)
    
    // 2. 批量查询所有 rooms
    unitIDs := extractUnitIDs(units)
    allRooms, err := s.unitsRepo.ListRoomsByUnitIDs(ctx, req.TenantID, unitIDs)
    
    // 3. 批量查询所有 beds
    roomIDs := extractRoomIDs(allRooms)
    allBeds, err := s.unitsRepo.ListBedsByRoomIDs(ctx, req.TenantID, roomIDs)
    
    // 4. 批量查询 device IDs 和 names
    roomDevices, err := s.devicesRepo.GetDevicesByRoomIDs(ctx, req.TenantID, roomIDs) // 返回 map[roomID][]DeviceInfo{ID, Name}
    bedDevices, err := s.devicesRepo.GetDevicesByBedIDs(ctx, req.TenantID, extractBedIDs(allBeds)) // 返回 map[bedID][]DeviceInfo{ID, Name}
    
    // 5. 组装数据
    result := make([]*UnitWithFullHierarchy, 0, len(units))
    for _, unit := range units {
        rooms := filterRoomsByUnitID(allRooms, unit.UnitID)
        roomsWithBeds := make([]*RoomWithBedsAndDevices, 0, len(rooms))
        for _, room := range rooms {
            beds := filterBedsByRoomID(allBeds, room.RoomID)
            bedsWithDevices := make([]*BedWithDevices, 0, len(beds))
            for _, bed := range beds {
                bedDeviceList := bedDevices[bed.BedID]
                deviceIDs := make([]string, 0, len(bedDeviceList))
                deviceNames := make([]string, 0, len(bedDeviceList))
                for _, device := range bedDeviceList {
                    deviceIDs = append(deviceIDs, device.ID)
                    deviceNames = append(deviceNames, device.Name)
                }
                bedsWithDevices = append(bedsWithDevices, &BedWithDevices{
                    Bed: bed,
                    DeviceIDs: deviceIDs,
                    DeviceNames: deviceNames,
                })
            }
            roomDeviceList := roomDevices[room.RoomID]
            roomDeviceIDs := make([]string, 0, len(roomDeviceList))
            roomDeviceNames := make([]string, 0, len(roomDeviceList))
            for _, device := range roomDeviceList {
                roomDeviceIDs = append(roomDeviceIDs, device.ID)
                roomDeviceNames = append(roomDeviceNames, device.Name)
            }
            roomsWithBeds = append(roomsWithBeds, &RoomWithBedsAndDevices{
                Room: room,
                Beds: bedsWithDevices,
                DeviceIDs: roomDeviceIDs,
                DeviceNames: roomDeviceNames,
            })
        }
        result = append(result, &UnitWithFullHierarchy{
            Unit: unit,
            Rooms: roomsWithBeds,
        })
    }
    
    return &ListUnitsWithFullHierarchyResponse{
        Items: result,
        Total: total,
    }, nil
}
```

## 6. Handler 层实现

### 6.1 路由注册

在 `unit_handler.go` 的 `ServeHTTP` 方法中添加新路由：

```go
// Units with Full Hierarchy
case r.URL.Path == "/admin/api/v1/units/with-hierarchy" && r.Method == http.MethodGet:
    h.ListUnitsWithFullHierarchy(w, r)
```

### 6.2 Handler 方法实现

```go
// ListUnitsWithFullHierarchy 查询 Units 及其完整的层级结构（Rooms, Beds, Devices）
func (h *UnitHandler) ListUnitsWithFullHierarchy(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    tenantID, ok := h.tenantIDFromReq(w, r)
    if !ok {
        return
    }

    // 解析查询参数（与 ListUnits 类似）
    branchID := r.URL.Query().Get("branch_id")
    branchName := r.URL.Query().Get("branch_name")
    buildingID := r.URL.Query().Get("building_id")
    buildingName := r.URL.Query().Get("building")
    floor := r.URL.Query().Get("floor")
    unitType := r.URL.Query().Get("unit_type")
    search := r.URL.Query().Get("search")

    req := service.ListUnitsWithFullHierarchyRequest{
        TenantID:   tenantID,
        BranchID:   stringPtrOrNil(branchID),
        BranchName: stringPtrOrNil(branchName),
        BuildingID: stringPtrOrNil(buildingID),
        Building:   stringPtrOrNil(buildingName),
        Floor:      stringPtrOrNil(floor),
        UnitType:   stringPtrOrNil(unitType),
        Search:     stringPtrOrNil(search),
    }

    resp, err := h.unitService.ListUnitsWithFullHierarchy(ctx, req)
    if err != nil {
        h.logger.Error("ListUnitsWithFullHierarchy failed", zap.Error(err))
        writeJSON(w, http.StatusOK, Fail(err.Error()))
        return
    }

    // 转换响应格式
    out := make([]any, 0, len(resp.Items))
    for _, unitWithHierarchy := range resp.Items {
        out = append(out, unitWithFullHierarchyToJSON(unitWithHierarchy))
    }

    writeJSON(w, http.StatusOK, Ok(map[string]any{
        "items": out,
        "total": resp.Total,
    }))
}
```

### 6.3 JSON 转换函数

```go
// unitWithFullHierarchyToJSON 转换 UnitWithFullHierarchy 为 JSON
func unitWithFullHierarchyToJSON(unitWithHierarchy *service.UnitWithFullHierarchy) map[string]any {
    // 先转换 Unit 基本信息
    m := unitToJSON(unitWithHierarchy.Unit)
    
    // 添加 Rooms 数组
    rooms := make([]any, 0, len(unitWithHierarchy.Rooms))
    for _, roomWithBeds := range unitWithHierarchy.Rooms {
        rooms = append(rooms, roomWithBedsAndDevicesToJSON(roomWithBeds))
    }
    m["rooms"] = rooms
    
    return m
}

// roomWithBedsAndDevicesToJSON 转换 RoomWithBedsAndDevices 为 JSON
func roomWithBedsAndDevicesToJSON(roomWithBeds *service.RoomWithBedsAndDevices) map[string]any {
    // 先转换 Room 基本信息
    m := roomToJSON(roomWithBeds.Room)
    
    // 添加 Beds 数组
    beds := make([]any, 0, len(roomWithBeds.Beds))
    for _, bedWithDevices := range roomWithBeds.Beds {
        beds = append(beds, bedWithDevicesToJSON(bedWithDevices))
    }
    m["beds"] = beds
    
    // 添加 Device IDs 和 Names
    m["device_ids"] = roomWithBeds.DeviceIDs
    m["device_names"] = roomWithBeds.DeviceNames
    
    return m
}

// bedWithDevicesToJSON 转换 BedWithDevices 为 JSON
func bedWithDevicesToJSON(bedWithDevices *service.BedWithDevices) map[string]any {
    // 先转换 Bed 基本信息
    m := bedToJSON(bedWithDevices.Bed)
    
    // 添加 Device IDs 和 Names
    m["device_ids"] = bedWithDevices.DeviceIDs
    m["device_names"] = bedWithDevices.DeviceNames
    
    return m
}
```

## 7. 前端使用方式

### 7.1 API 定义（unit.ts）

```typescript
// 在 Api enum 中添加
GetUnitsWithFullHierarchy = '/admin/api/v1/units/with-hierarchy',

// 新增 API 函数
export async function getUnitsWithFullHierarchyApi(
  params: GetUnitsParams,
  mode: ErrorMessageMode = 'none'
): Promise<GetUnitsWithFullHierarchyResult> {
  if (useMock) {
    // Mock 实现
    const { unit } = await import('@test/index')
    return unit.mock.mockGetUnitsWithFullHierarchy(params)
  }

  return defHttp.get<GetUnitsWithFullHierarchyResult>(
    {
      url: Api.GetUnitsWithFullHierarchy,
      params,
    },
    { errorMessageMode: mode }
  )
}
```

### 7.2 前端类型定义（unitModel.ts）

```typescript
// 扩展 Unit 类型
export interface UnitWithFullHierarchy extends Unit {
  rooms: RoomWithBedsAndDevices[]
}

export interface RoomWithBedsAndDevices extends Room {
  beds: BedWithDevices[]
  device_ids: string[]
  device_names: string[]
}

export interface BedWithDevices extends Bed {
  device_ids: string[]
  device_names: string[]
}

export interface GetUnitsWithFullHierarchyResult {
  items: UnitWithFullHierarchy[]
  total: number
}
```

### 7.3 UnitView.vue 使用方式

```typescript
// 替换当前的 fetchAllData 方法
const fetchAllData = async () => {
  loading.value = true
  try {
    const userInfo = userStore.getUserInfo
    const tenantId = userInfo?.tenant_id

    if (!tenantId) {
      message.error('No tenant ID available')
      return
    }

    // 一次调用获取所有数据
    const result = await getUnitsWithFullHierarchyApi({
      tenant_id: tenantId,
    })
    
    const allUnits = result.items || []
    
    // 直接使用返回的数据，无需额外查询
    // 注意：仍然需要获取 Buildings（用于分组显示）
    const allBuildings = await getBuildingsApi()
    
    // Transform data to table format（简化版，因为数据已经完整）
    rawTableData.value = transformToTableData(allBuildings, allUnits)
    
    updateTableData()
  } catch (error: any) {
    message.error('Failed to fetch data: ' + (error.message || 'Unknown error'))
    tableData.value = []
  } finally {
    loading.value = false
  }
}
```

### 7.4 优势

1. **减少 API 调用**：从 N+2 次（1 次 units + N 次 rooms + 1 次 devices）减少到 2 次（1 次 units-with-hierarchy + 1 次 buildings）
2. **数据一致性**：所有数据来自同一次查询，保证一致性
3. **性能提升**：减少网络往返次数，数据库层完成 JOIN 操作
4. **代码简化**：前端无需复杂的 Promise.all 和数据处理逻辑

