# UnitView vs UnitList (Edit Mode) Service 调用对比

## 两个场景的需求

### 1. UnitView（查看模式）
- **需求**：展示所有 units（根据角色所在的 branch 过滤，admin/manager/IT）
- **数据范围**：所有 units（根据 branch_id 过滤）
- **数据结构**：完整的层级结构（Unit → Room → Bed → Device）
- **使用场景**：只读查看，不需要编辑

### 2. UnitList（编辑模式）
- **需求**：编辑单个 unit 的数据
- **数据范围**：当前选中的 unit 的 rooms, beds, devices
- **数据结构**：单个 unit 的 rooms 和 beds（包含 devices）
- **使用场景**：编辑操作（创建/更新/删除 rooms, beds, devices）

## Service 调用对比

### UnitView 使用的 Service

```typescript
// UnitView.vue - fetchAllData()
const result = await getUnitsWithFullHierarchyApi({
  tenant_id: tenantId,
  branch_id: userInfo.branch_id,  // 根据角色所在的 branch 过滤
})
// 返回：所有 units 的完整层级结构
// {
//   items: [
//     {
//       unit_id: "unit-001",
//       unit_name: "E101",
//       rooms: [
//         {
//           room_id: "room-001",
//           room_name: "bedroom",
//           device_ids: [...],
//           device_names: [...],
//           beds: [
//             {
//               bed_id: "bed-001",
//               bed_name: "BedA",
//               device_ids: [...],
//               device_names: [...]
//             }
//           ]
//         }
//       ]
//     },
//     ...
//   ],
//   total: 100
// }
```

**后端 Service**：`ListUnitsWithFullHierarchy`
- 批量查询所有 units（根据 branch_id 过滤）
- 一次性返回完整的层级结构
- 支持分页（但通常不分页，返回所有匹配的 units）

### UnitList（编辑模式）使用的 Service

```typescript
// UnitList.vue (useUnit.ts) - fetchRoomsWithBeds()
const result = await getRoomsApi({ unit_id: unitId })
// 返回：单个 unit 的 rooms 和 beds
// [
//   {
//     room_id: "room-001",
//     room_name: "bedroom",
//     beds: [
//       {
//         bed_id: "bed-001",
//         bed_name: "BedA"
//       }
//     ]
//   }
// ]

// UnitList.vue (useDevice.ts) - fetchDevices()
const result = await getDevicesApi({
  tenant_id: tenantId,
  business_access: 'approved',
  include_bound: true,
})
// 返回：所有 devices（用于绑定到 room/bed）
```

**后端 Service**：
1. `ListRoomsWithBeds`：查询单个 unit 的 rooms 和 beds
2. `ListDevices`：查询所有 devices（用于绑定操作）

## 结论：两个 Service 是否相同？

**答案：不同**

### 原因

1. **数据范围不同**：
   - UnitView：批量查询所有 units（根据 branch_id 过滤）
   - UnitList：单个 unit 的 rooms 和 beds

2. **数据结构不同**：
   - UnitView：需要完整的层级结构（Unit → Room → Bed → Device），包含 device_ids 和 device_names
   - UnitList：只需要 rooms 和 beds，devices 单独查询（用于绑定操作）

3. **使用场景不同**：
   - UnitView：只读查看，一次性加载所有数据
   - UnitList：编辑操作，按需加载单个 unit 的数据

### Service 映射关系

| 前端场景 | 使用的 Service | 说明 |
|---------|---------------|------|
| UnitView | `ListUnitsWithFullHierarchy` | 批量查询所有 units 的完整层级结构 |
| UnitList（编辑单个 unit） | `ListRoomsWithBeds` + `ListDevices` | 单个 unit 的 rooms/beds + 所有 devices |

## 设计建议

### 1. UnitView 使用 `ListUnitsWithFullHierarchy`

```go
// Service 层
func (s *unitService) ListUnitsWithFullHierarchy(ctx context.Context, req ListUnitsWithFullHierarchyRequest) (*ListUnitsWithFullHierarchyResponse, error) {
    // 支持 branch_id 过滤（根据角色所在的 branch）
    // 返回所有 units 的完整层级结构
}
```

**优势**：
- 一次 API 调用获取所有数据
- 减少网络往返次数
- 数据一致性更好

### 2. UnitList（编辑模式）继续使用现有 Service

```go
// Service 层
func (s *unitService) ListRoomsWithBeds(ctx context.Context, req ListRoomsWithBedsRequest) (*ListRoomsWithBedsResponse, error) {
    // 查询单个 unit 的 rooms 和 beds
    // 不包含 devices（devices 单独查询）
}
```

**优势**：
- 按需加载，性能更好
- 编辑操作只需要单个 unit 的数据
- 保持现有逻辑不变

## 总结

- **UnitView**：使用新的 `ListUnitsWithFullHierarchy` Service（批量查询，完整层级结构）
- **UnitList（编辑模式）**：继续使用现有的 `ListRoomsWithBeds` + `ListDevices` Service（单个 unit，按需加载）

两者使用不同的 Service，因为：
1. 数据范围不同（所有 units vs 单个 unit）
2. 数据结构不同（完整层级 vs 部分层级）
3. 使用场景不同（只读查看 vs 编辑操作）

