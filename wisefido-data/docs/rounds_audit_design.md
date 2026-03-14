# Automatic Rounds 与告警审计设计

## 1. 结论：用 rounds 表 + 导出

- **rounds 表**：单独保留，作为巡房 SOP 与审计的主数据源；支持按时间、人员、租户查询和导出。
- **导出**：以 DB 为准，按需导出 CSV/Excel 供内部审计；可选后续支持 Word/PDF 单轮报告。

## 2. Rounds 表（扩展现有）

现有表：`rounds`（round_id, tenant_id, round_type, unit_id, executor_id, round_time, notes, status）、`round_details`（按住户/床位的明细）。

为支持 Overview 的 Automatic Rounds（手动勾选类别后完成一轮）：

- **扩展字段**（在 `rounds` 上）：
  - `started_at` TIMESTAMPTZ NULL — 本轮开始时间（可选，未开则用 round_time 近似）。
  - `items_checked` JSONB NULL — 本轮勾选项快照，例如：
    ```json
    [
      { "key": "unhandled", "checked": true, "count": 2, "checked_at": "2025-03-13T10:00:00Z" },
      { "key": "leftbed", "checked": true, "count": 1, "checked_at": "2025-03-13T10:01:00Z" }
    ]
    ```
- **约定**：
  - `round_type = 'manual'` 表示 Overview 发起的“手动巡查一轮”。
  - `round_time` 作为完成时间（round 创建或完成时写入）。
  - `executor_id` 为执行人（护士）user_id；审计时关联 users 表取姓名。

API：

- `POST /data/api/v1/data/vital-focus/rounds` — 创建一轮（body: started_at?, items_checked）；返回 round_id。前端在“Complete Rounds”时调用。
- `GET /data/api/v1/data/vital-focus/rounds` — 列表，支持 tenant_id、executor_id、start_time、end_time、page、size；用于审计列表页。
- `GET /data/api/v1/data/vital-focus/rounds/export?from=&to=&format=csv` — 导出为 CSV（或 format=excel）；审计用。

## 3. 告警审计（与 rounds 分离）

- **目标**：对“当天有 alarm”保留快照，处理前后可截图或结构化状态，便于事后审计。
- **建议**：
  - **告警发生**：现有 alarm 事件/推送可落库时，存一条 alarm 快照（时间、卡片、级别、文案、当时状态摘要）。
  - **告警处理**：护士点“处理”时，记录一条 handle 记录，包含：处理人、处理时间、处理前状态（或截图 URL/二进制）、处理后状态（同上）；与 alarm 事件关联。
- **存储**：单独表（如 `alarm_handles` 或 `alarm_audit`），不塞进 rounds；rounds 只负责“巡房一轮”的谁/何时/勾了哪些项。
- **导出**：告警/处理记录可单独按日或按卡片导出，供合规审计；与 rounds 导出可并列提供。

## 4. 导出格式建议

- **优先**：从 DB 查 rounds 列表，服务端生成 **CSV** 或 **Excel**，按日期范围/人员筛选后下载；实现简单、易做审计表格。
- **可选**：单轮或按日生成 **Word/PDF** 报告（含表格+可选截图），用于归档或打印；可后续迭代。

## 5. 前端流程（Overview）

- 打开 Automatic Rounds → 勾选各项 → 点击 Complete Rounds：
  - 调用 `POST .../rounds`，传 `started_at`（可选，可用弹窗打开时间）、`items_checked`（当前勾选项 + 每项 count/checked_at）。
  - 成功后关闭弹窗、清空勾选；可 toaster 提示“本轮已记录，可用于审计”。
- 审计侧：单独页面或管理端“Rounds 记录”列表，筛选后“导出 CSV/Excel”。
