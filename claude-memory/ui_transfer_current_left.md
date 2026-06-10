---
name: Transfer UI convention — Current 在左
description: a-transfer 双栏分配组件，约定当前已分配的项放左侧，可选池放右侧
type: feedback
originSessionId: 45c27d41-bfc6-45c1-bff5-228bbb490b7e
---
a-transfer / 任何左右双栏分配 UI（成员管理、权限分配、tags 等），约定 **Current（已分配的）放左侧，Available（可选池）放右侧**。

**Why:** 用户习惯——左侧="现状/当前"，右侧="备选"；多个页面统一规则。

**How to apply:**
- `:titles="['Current', 'Available']"`（不是 ant-design 默认的 `['Source', 'Target']`）
- `targetKeys` 装 *available* 的 id 集合（不是 members）
- transfer change 事件 direction 语义随之翻转：`'left'` = 加入（Available → Current）；`'right'` = 移出
- 落地 例子：[CareTeamList.vue Members modal](owlFront/src/views/admin/care-teams/CareTeamList.vue)
