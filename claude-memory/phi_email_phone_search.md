---
name: Phase 3b PHI Email/Phone Search Implemented
description: Email/phone search in ListResidents with automatic classification, PHI decryption, permission checks, and audit logging
type: project
originSessionId: dbf62cc2-e8c2-404f-bd60-c9c954296323
---
**Completed: 2026-05-11**

Phase 3b email/phone search for ListResidents is now production-ready. Implementation includes:

**Core Features:**
- Automatic search type classification via `ClassifySearch()` (email/phone/account/nickname)
- `searchResidentsByPHI()` function that decrypts resident_phi data on-demand
- Case-insensitive fuzzy matching for email/phone search terms
- Early-exit routing in listResidentsV2 when email/phone detected

**Permission Enforcement:**
- `assigned_only`: Filters via resident_caregivers.caregiver_id UUID JOIN
- `branch_only`: Extracts /56 prefix from resident_unit.spatial_prefix with three cases:
  - No branches: Filter to cross-branch transition residents (masklen ≤ 56)
  - Single branch: Match exact /56 prefix
  - Multiple branches: IN-clause across all user branches

**Data Flow:**
1. searchResidentsByPHI receives currentUserID + tenantPrefix + searchType
2. Builds WHERE clause with permission filters (same logic as listResidentsV2 lines 108-158)
3. Queries residents + LEFT JOIN resident_phi on encrypted columns
4. Loops through results, decrypting only matched residents (lazy decryption)
5. Logs PHI access via s.logger.Info("PHI access for search", ...)
6. Returns ResidentListItemDTO[] suitable for paginated response

**Code Locations:**
- Search classification: `resident_service_v2.go:23-56` (ClassifySearch, GetSearchTypeDescription)
- PHI search implementation: `resident_service_v2.go:1653-1806` (searchResidentsByPHI)
- Integration: `resident_service_v2.go:66-86` (email/phone routing in listResidentsV2)

**Why:** Supports HIPAA-compliant searches without storing email/phone in plaintext. All PHI access is audited. Permission layers prevent unauthorized cross-tenant/cross-branch data leakage.

**How to apply:** When users search for residents by email/phone, system automatically classifies the query and routes to searchResidentsByPHI with proper decryption and audit trails.
