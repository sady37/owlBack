---
name: owl_v2 dbv2 已部署（owlrd 保留作冷备份）
description: owl_v2 数据库（IPv6 spatial 主键 + 41 个 SQL）2026-05-08 部署完成；同 volume 内 owlrd 保留作回退
type: project
originSessionId: 53d81dc2-b99e-4371-95da-247c2cf5c46c
---
owl_v2 是新生产库，承载 IPv6 spatial 主键 + Datagram v1 envelope 的全套 schema。owlrd 旧库在同 postgres volume 内保留作冷备份回退（不删除）。

**Why:** Phase A protocol foundation 完成（IPv6 spatial v2 + envelope）后做完全不向后兼容的 schema 重建；owlrd 留底降低风险，确认 owl_v2 业务跑通后再决定是否 DROP owlrd。

**How to apply:**
- DB 连接：`DB_NAME=owl_v2`（owlBack/.env 已切换）
- 重建 owl_v2：`bash owlBack/scripts/setup_owl_v2.sh --drop`（不影响 owlrd）
- schema 源：`owlRD/dbv2/*.sql`（41 个文件，按文件名数字顺序 init）
- 51 张表 / 4 view / 3 hypertable（audit_log + event_log + monitor_stream）
- BIND9 在 docker port 5353（5353 是因 host port 53 被 systemd-resolved 占用），owl service 配 `BIND_PORT=5353`
- kea-dhcp6 暂时 deferred（Docker Hub 无可用镜像）；当前用 Go SELECT MAX(slot)+1 简易 IPAM allocator
- 切回 owlrd：编辑 .env `DB_NAME=owlrd` + 重启服务

**Schema ordering quirks（再次重建/编辑时注意）：**
- 16_spatial_views 只放无 device 依赖的 view；devices/residents 相关 view 在 39_post_setup_views.sql
- 33_resident_contacts.linked_user_id / 34_resident_caregivers.{caregiver_hoa,care_team_id} 不带 FK（users 在 40_、care_teams 在 44_，FK 由应用层维护）
- partial index predicate 不能用 NOW()（非 IMMUTABLE）；"active" 一律用 `valid_to IS NULL`
- hypertable 表（audit_log）PK 必须 include partitioning column（occurred_at）
