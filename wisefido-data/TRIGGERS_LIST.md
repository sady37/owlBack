# 数据库触发器完整列表

## 触发器列表（共54个）

### users表（4个触发器）
1. `trigger_cleanup_user_from_tags` - AFTER DELETE - `cleanup_user_from_tags()`
2. `trigger_sync_user_tags` - AFTER INSERT/UPDATE - `sync_user_tags_to_catalog()`
3. `trigger_users_lowercase_account` - BEFORE INSERT/UPDATE - `ensure_lowercase_user_account()`

### residents表（7个触发器）
1. `trigger_residents_lowercase_account` - BEFORE INSERT/UPDATE - `ensure_lowercase_resident_account()`
2. `trigger_sync_family_tag` - AFTER INSERT/UPDATE - `sync_family_tag_to_catalog()`
3. `trigger_validate_resident_bed_room` - BEFORE INSERT/UPDATE - `validate_resident_bed_room()`
4. `trigger_validate_resident_location_tenant` - BEFORE INSERT/UPDATE - `validate_resident_location_tenant()`
5. `trigger_validate_resident_room_unit` - BEFORE INSERT/UPDATE - `validate_resident_room_unit()`

### units表（7个触发器）
1. `trigger_cleanup_location_from_tags` - AFTER DELETE - `cleanup_location_from_tags()`
2. `trigger_sync_area_tag` - AFTER INSERT/UPDATE - `sync_area_tag_to_catalog()`
3. `trigger_sync_branch_tag` - AFTER INSERT/UPDATE - `sync_branch_tag_to_catalog()`
4. `trigger_sync_units_grouplist_to_cards` - AFTER INSERT/UPDATE - `sync_units_grouplist_to_cards()`
5. `trigger_validate_unit_alarm_user_tenant` - BEFORE INSERT/UPDATE - `validate_unit_alarm_user_tenant()`

### devices表（10个触发器）
1. `trigger_update_bed_device_count_on_bind` - AFTER INSERT/UPDATE/DELETE - `update_bed_device_count()`
2. `trigger_update_bed_device_count_on_monitoring` - AFTER UPDATE - `update_bed_device_count()`
3. `trigger_validate_device_bed_room` - BEFORE INSERT/UPDATE - `validate_device_bed_room()`
4. `trigger_validate_device_bed_tenant` - BEFORE INSERT/UPDATE - `validate_device_bed_tenant()`
5. `trigger_validate_device_identifier` - BEFORE INSERT/UPDATE - `validate_device_identifier()`
6. `trigger_validate_device_store_tenant` - BEFORE INSERT/UPDATE - `validate_device_store_tenant()`

### cards表（6个触发器）
1. `trigger_cleanup_card_from_tags` - AFTER DELETE - `cleanup_card_from_tags()`
2. `trigger_validate_card_bed_unit` - BEFORE INSERT/UPDATE - `validate_card_bed_unit()`
3. `trigger_validate_card_resident_bed` - BEFORE INSERT/UPDATE - `validate_card_resident_bed()`
4. `trigger_validate_card_tenant` - BEFORE INSERT/UPDATE - `validate_card_tenant()`

### 其他表（20个触发器）
- **beds表**: `trigger_validate_bed_tenant` (BEFORE INSERT/UPDATE)
- **rooms表**: `trigger_validate_room_tenant` (BEFORE INSERT/UPDATE)
- **resident_caregivers表**: `trigger_validate_resident_caregiver_tenant` (BEFORE INSERT/UPDATE)
- **iot_timeseries表**: `trigger_validate_iot_timeseries_location_device`, `trigger_validate_iot_timeseries_location_tenant` (BEFORE INSERT/UPDATE)
- **alarm_events表**: `trigger_update_alarm_events_updated_at` (BEFORE UPDATE)

## 引用tag_objects的函数（需要修复或删除）
1. `drop_object_from_all_tags` - 引用tag_objects
2. `drop_tag` - 引用tag_objects
3. `drop_tag_type` - 引用tag_objects
4. `get_tags_for_object` - 引用tag_objects
5. `get_tags_for_tenant` - 引用tag_objects
6. `sync_area_tag_to_catalog` - 引用tag_objects
7. `sync_branch_tag_to_catalog` - 引用tag_objects
8. `sync_location_tag_to_catalog` - 引用tag_objects
9. `sync_user_tags_to_catalog` - 引用tag_objects
10. `update_tag_objects` - 引用tag_objects

