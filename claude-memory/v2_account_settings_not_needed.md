---
name: v2 Resident account-settings 端点不重建
description: v2 不再需要 /admin/api/v2/residents/:hoa/account-settings；v1 路径残留是 FE Sidebar 的清理项
type: project
originSessionId: 116b375e-502d-4c66-a3e8-44aef3579be5
---
v2 cutover 评估结论（2026-05-11）：v2 **不**为 resident 重建 `account-settings` 端点。

**Why:**
- resident 在 v2 不是 user：resident_id INET（hoa）≠ user_id UUID，没自己的密码
- v1 端点功能在 v2 已经拆解到其他路径：
  - password reset → 走 user_service 的 user-level 端点（resident 自助场景下也是其同名 user 的 password 改）
  - email/phone → PHI 加密字段，PUT `/admin/api/v2/residents/:hoa` body `{ phi: { resident_email, resident_phone } }`
  - save_email/save_phone（接收通知）→ resident_contacts.receive_email/receive_sms，PUT body `{ contacts: [{ receive_email, receive_sms }] }`
  - family_access → ResidentUpdateInput.FamilyAccess，PUT body `{ family_access: true|false }`

**How to apply:**
- 不要建 v2 account-settings 路由。
- FE 后续清理：[owlFront/src/api/account/accountSettings.ts](owlFront/src/api/account/accountSettings.ts) 的 ResidentGet/ResidentUpdate 路径死掉之后会 fail；fallback 应改为走 user 路径（v2 所有登录者都是 user，包括 B2C Family）。这是 FE cleanup 任务，不是 BE cutover。
- 如果后续 Sidebar "修改账户设置" 弹窗里出现 resident-only 字段（如 family_access toggle）也只该出现在 staff 改 resident profile 的常规 PUT 流程，不是登录用户的 "我自己的账户设置"。
