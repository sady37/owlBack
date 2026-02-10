# 卡片名称和地址 Bug 修复总结

## 问题诊断

### 原始问题
- 卡片名称显示 UUID（如：`bed-uuid - room-uuid`）
- 卡片地址显示 UUID（单元 ID）

### 根本原因
1. **GetActiveBedsByUnit** 查询了 `b.bed_name` 和 `r.room_name`，但 `ActiveBedInfo` 结构体没有这两个字段来存储，导致未被使用
2. **GetUnitInfo** 查询了 `branch_id` 而非 `branch_name`
3. **creator.go buildExpectedCards()** 使用了 `bed.BedID` 和 `bed.RoomID`（UUID）而非实际名称

## 已应用的修复

### 1. 更新 ActiveBedInfo 结构体 ✅
**文件**: `owl-common/card/card_types.go`
```go
type ActiveBedInfo struct {
    BedID             string  // 床 ID
    BedName           string  // ✅ 新增：床名称
    RoomID            string  // 房间 ID
    RoomName          string  // ✅ 新增：房间名称
    UnitID            string  // ✅ 新增：单元 ID
    BoundDeviceCount  int     // ✅ 新增：设备数量
    ResidentID        *string // 住户 ID
}
```

### 2. 修复 GetActiveBedsByUnit SQL ✅
**文件**: `wisefido-data/internal/repository/postgres_card.go`
```sql
-- 现在正确查询并存储名称
SELECT b.bed_id, r.unit_id, COUNT(...), r2.resident_id, 
       b.room_id, COALESCE(b.bed_name, '') as bed_name, 
       COALESCE(r.room_name, '') as room_name  -- ✅ 正确
```

### 3. 修复 GetUnitInfo SQL ✅
**文件**: `wisefido-data/internal/repository/postgres_card.go`
```sql
-- 修复：查询 branch_name 而非 branch_id
LEFT JOIN branches b ON u.branch_id = b.branch_id
SELECT u.unit_id, u.unit_name, 
       COALESCE(b.branch_name, '') as branch_name,  -- ✅ 正确
```

### 4. 修复 UnitInfo 结构体 ✅
**文件**: `owl-common/card/card_types.go`
```go
type UnitInfo struct {
    BranchName string // ✅ 改为 branch_name，而非 branch_id
    // ...
}
```

### 5. 修复 buildExpectedCards 逻辑 ✅
**文件**: `wisefido-data/internal/card/creator.go`

**ActiveBed 卡片**:
```go
// ✅ 现在使用实际名称，而非 UUID
cardName := bed.BedName
if bed.RoomName != "" {
    cardName = bed.RoomName + " - " + bed.BedName
}

// ✅ 构建完整地址：branch_name + "-" + building + "-" + unit_name
if unitInfo != nil {
    parts = append(parts, unitInfo.BranchName, unitInfo.Building, unitInfo.UnitName)
    cardAddr = strings.Join(parts, "-")
}
```

**Location 卡片**:
```go
// ✅ 使用 unit_name 而非 unitID
if unitInfo != nil && unitInfo.UnitName != "" {
    cardName = unitInfo.UnitName
}

// ✅ 相同的完整地址构建逻辑
```

## 验证清单

- [x] ActiveBedRow 有 BedName, RoomName 字段
- [x] GetActiveBedsByUnit 查询并扫描了这两个字段
- [x] GetUnitInfo 查询 branch_name（而非 branch_id）
- [x] UnitInfo 结构体中有 BranchName 字段
- [x] creator.go 使用床名称而非 UUID
- [x] creator.go 使用房间名称而非 UUID
- [x] creator.go 使用单元名称而非 unitID
- [x] location 卡片地址构建逻辑正确

## 预期结果

### Before (错误)
- 卡片名称: `8f7a1e2b-... - 3c4d5e6f-...` ❌
- 卡片地址: `92a3b4c5-...` ❌

### After (正确)
- 卡片名称: `301房 - A床` ✅
- 卡片地址: `院区-2号楼-301房` ✅

## 下一步

1. 编译测试: `go build -o wisefido-data ./cmd/wisefido-data`
2. 启动服务: `./start-data.sh`
3. 触发卡片重新生成
4. 验证 API 返回正确的卡片名称和地址
