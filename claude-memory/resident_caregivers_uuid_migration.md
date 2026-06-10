---
name: resident_caregivers UUID migration
description: Schema migration from INET to UUID for caregiver_id/family_id columns
type: project
originSessionId: dbf62cc2-e8c2-404f-bd60-c9c954296323
---
## Schema Change
Migrated `resident_caregivers` table to use UUID instead of INET for caregiver/family references:
- `caregiver_id`: INET → UUID (references users.user_id)
- `family_id`: INET → UUID (references users.user_id)

**Why:** Phase B' design treated `hoa` as placeholder-only (not consumed by business logic). The table incorrectly stored references via `hoa` (INET) instead of `user_id` (UUID), causing save/load failures when users lacked `hoa` values (admin/manager/family types).

**How to apply:** 
1. Deploy migration 34a_resident_caregivers_uuid_migration.sql
2. Update backend queries/inserts to use user_id (already done in resident_service_v2.go)
3. Frontend unchanged (sends user_id UUIDs, backend now stores them directly)

## Code Changes (resident_service_v2.go)
- loadResidentCaregiversV2 (line 337, 378): Changed `JOIN users u ON u.hoa = rc.caregiver_id` to `JOIN users u ON u.user_id = rc.caregiver_id` (both for caregivers and family)
- writeResidentCaregivers (line 819-825): Changed SELECT with hoa lookup to direct VALUES insert with UUID
- writeResidentFamily (line 884-890): Changed SELECT with COALESCE(u.hoa, ...) to direct VALUES insert with UUID

## Migration Path
1. Migration script creates NULL values for existing INET data (no preserved data, schema-fresh)
2. New saves use UUID directly
3. Backward compatible: FK constraints optional per design
