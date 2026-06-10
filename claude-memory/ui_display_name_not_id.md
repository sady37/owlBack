---
name: 执行用 ID，展示用 name
description: UI 规则 — 任何选择器/列表显示给人看的必须是 human-readable name；提交后端用 ID/CIDR；不能让 IPv6 CIDR 或 UUID 落到屏幕上
type: feedback
originSessionId: 116b375e-502d-4c66-a3e8-44aef3579be5
---
UI 规则：**所有面向最终用户的展示一律 name；ID/CIDR/UUID 只在 API payload / URL params 里走**。

**Why:** v2 schema 主键是 IPv6 INET CIDR（如 `fd00:0:3:100::/56`）或 UUID。直接显示到 UI 上完全不可读，老人/护士看不懂。曾经在 CareTeamList Create modal 的 branch 下拉框里 option 用 children 渲染 name 但 antd-vue a-select 在某些情况下回退显示 v-model 绑定的 value（CIDR），导致下拉框已选项显示成 IPv6 CIDR。

**How to apply:**

1. **a-select 下拉**优先用 `:options` 显式 label 形式，避免 a-select-option children 推断：
   ```vue
   <a-select
     v-model:value="form.branch_id"
     :options="branches.map(b => ({ value: b.branch_id, label: b.branch_name }))"
   />
   ```
   或在 `<a-select-option>` 上显式加 `:label="b.branch_name"` 属性。

2. **a-table 列**用 `dataIndex` 指向 name 字段（`branch_name` / `nickname` / `unit_name`），不要指向 id 字段。BE list 端点要同时返 id + name（id 给操作用，name 给展示）。

3. **a-tag / 详情面板**显示资源关联时，绑定 name；点击跳转用 id。

4. **breadcrumb / 面包屑**用 name；route param 走 id（URL 里短 hash/uuid 比 CIDR 好编码）。

5. **autocomplete / 搜索建议**：option label = `{name} ({account})`，value = id。

6. **多语言**：name 由 i18n key 渲染时，id 不动；翻译失败兜底显示 raw name 不显示 id。

7. **错误消息**要尽量带 name + 简短 id 后缀，但 ID 用 `…` 截短：
   `"care_team xxxxxxxx-… not assigned to branch Denver"` 不要 `"care_team xxxxxxxx-yyyy-… not assigned to branch fd00:0:3:100::/56"`

这条规则一票否决：哪怕快速 hack 也不允许把 raw CIDR/UUID 直接渲染到屏幕。BE 必须同时返 id 和 name 让 FE 拼装。

## 根本规则：BE 返 ID 必须成对返 name（v1.0 约定）

任何 BE response 含 `*_id` / `*_uuid` / `hoa` 字段时，**必须同时返回对应的 `*_name` / display 字段**（同一 JSON object 同级，1:1 配对）。FE 直接用 name 显示，**绝不让 FE 二次反查 options/list 拼名字**。

示例正确格式：
```json
{
  "user_id": "uuid-...",
  "user_account": "demo",
  "nickname": "Demo",
  "branches": [
    {"branch_id": "fd00:0:3:100::/56", "branch_name": "Denver"},
    {"branch_id": "fd00:0:3:200::/56", "branch_name": "Spring"}
  ],
  "care_teams": [
    {"team_id": "uuid-...", "team_name": "Emerge"}
  ]
}
```

**❌ 错误做法**（让 FE 反查 options）：
```json
{
  "branch_ids": ["fd00:0:3:100::/56", "fd00:0:3:200::/56"]
  // FE 得自己 fetchBranches() 后 map 找 branch_name → 必然踩 race
}
```

**根本解决了哪些隐性失败模式**：
- FE 不用 await options ready → 无 race
- value 格式微差不再导致显示 ID（直接用配对的 name）
- 离职 / 已删除 branch 仍能历史显示（snapshot name 已固化）

每次写 BE handler，最后一步检查 response：含 `*_id`？同级是否有 `*_name`？没有就加 JOIN。

## 隐性失败模式（必查）

仅用 `:options` 不够 — **antd a-select 在 v-model value 找不到匹配 option 时，会回退把 value 字符串当 displayValue 直接渲染**。常见触发场景：

1. **options 异步 fetch，未 await load 完**就打开 modal / 设 v-model：
   ```vue
   <a-select v-model:value="form.branch_id" :options="branchOptions" />
   <!-- branchOptions 来自 branches.value.map(...)；branches 是 onMounted 异步 fetch -->
   ```
   modal 打开时 branches 还 []，options 空 → a-select 直接渲染 `form.branch_id`（CIDR）。

2. **value 格式与 options.value 格式微差**（如 `/56` 后缀有/无、IPv6 缩写不一致、case sensitivity）→ 找不到匹配 → fallback raw。

3. **不可改字段用 disabled a-select**：disabled 状态下 antd 不渲染下拉，但仍走 value→option.label 反查，反查失败一样回退。

**强制做法**：

- **不可改字段（Edit 时 branch / tenant 等）禁用 select**，直接 `<a-input :value="entity.branch_name" disabled />` 渲染 name 文本。已有 record.branch_name 时绝不二次反查。
- **可改字段必须 await options ready**：modal 打开前 `await fetchBranches()`，或加 `:loading` 状态禁用 select 直到 options.length > 0。
- **value 格式必须与 options.value 严格 1:1 一致**：BE 同时给 list 端点和详情端点都返同一份 CIDR 表达（带 /56 或不带，约定一种）。
- **静态扫描盲区**：仅看模板是 `<a-select :options=...>` 不够，必须 cross-check options 数据源是否同步可用 + value 来源是否同格式。

测试方法：刷新页面立刻打开 Edit modal（不要等几秒）—— 如果看到 CIDR / UUID 文本，就是踩到 race。
