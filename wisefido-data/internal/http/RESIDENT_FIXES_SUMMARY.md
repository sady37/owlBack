# Resident 相关问题修复总结

## 问题 1: Service Level 下拉框是空的 ✅

### 原因
- `AdminServiceLevels` 函数只是返回空数组，没有从数据库读取数据

### 修复
- 实现了从 `service_levels` 表读取数据的逻辑
- 返回所有 service levels，包括 level_code, description, color, color_hex, priority
- 按 priority 和 level_code 排序

### 文件
- `admin_other_handlers.go` - AdminServiceLevels 函数

## 问题 2: Save 后 resident 列表是空 ✅

### 原因
- `AdminResidents` 的 GET 和 POST 方法都是 stub 实现，返回空数组或 stub 数据

### 修复
- **GET /admin/api/v1/residents**: 实现了从数据库读取 residents 列表
  - 支持 search 参数（搜索 nickname 或 unit_name）
  - 支持 status 和 service_level 过滤
  - JOIN units, rooms, beds 表获取完整信息
- **POST /admin/api/v1/residents**: 实现了创建 resident
  - 自动生成 resident_account（如果未提供）
  - 生成 account_hash 和 password_hash
  - 支持创建 PHI 记录（如果提供了 first_name/last_name）
- **GET /admin/api/v1/residents/:id**: 实现了获取 resident 详情
  - 支持 include_phi 和 include_contacts 参数
  - 返回完整的 resident 信息
- **PUT /admin/api/v1/residents/:id**: 实现了更新 resident
  - 支持更新所有字段（nickname, status, service_level, admission_date, etc.）
- **DELETE /admin/api/v1/residents/:id**: 实现了软删除（标记为 discharged）

### 文件
- `admin_residents_handlers.go` - AdminResidents 函数

## 问题 3: Resident Profile 页面卡顿 ⚠️

### 可能原因
1. **并行加载多个数据源**：
   - `onMounted` 时并行加载：fetchServiceLevels(), fetchUnits(), fetchCaregivers(), fetchCaregiverTags()
   - 如果某个 API 响应慢，会阻塞整个页面渲染

2. **Deep Watch 监听**：
   - `watch(() => localResidentData.value, ...)` 使用 `{ deep: true }`
   - 可能导致频繁的更新和重新渲染

3. **Computed 属性复杂度**：
   - `sortedAndFilteredUnits` computed 属性在每次 units 变化时都会重新计算和排序

### 建议优化
1. **延迟加载非关键数据**：
   - Service Levels 可以在需要时再加载（当用户点击下拉框时）
   - Caregivers 和 Caregiver Tags 可以在需要时再加载

2. **使用防抖/节流**：
   - 对 watch 回调使用防抖，减少更新频率

3. **优化 computed 属性**：
   - 使用 memoization 缓存计算结果

### 文件
- `ResidentProfileContent.vue` - 需要优化性能

## 测试建议

1. **Service Level 下拉框**：
   - 创建 resident 时，Service Level 下拉框应该显示所有可用的 service levels
   - 验证数据来自数据库（不是 mock）

2. **Resident 列表**：
   - 创建 resident 后，列表应该显示新创建的 resident
   - 验证搜索和过滤功能正常

3. **Resident Profile 页面**：
   - 检查页面加载时间
   - 检查是否有不必要的 API 调用
   - 检查是否有内存泄漏

