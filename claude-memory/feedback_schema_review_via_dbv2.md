---
name: feedback-schema-review-via-dbv2
description: 表 schema 改动审核流程铁律 — 必须先改 dbv2/CREATE TABLE 文件提审；只看 CREATE 文件做 schema 审核
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 6596aa1c-6420-4198-9805-2c0620c5d97c
---

任何表 schema 改动（加列 / 删列 / 改类型 / 删表 / 改 FK / 加约束）**必须**：

1. **先改** `/home/wisefido/owl/owlRD/dbv2/<NN>_<table>.sql` 里 CREATE TABLE 段（用户看 CREATE 做 schema 审核）
2. **提审 + 等用户同意**
3. **再动** Go 代码 / 业务 SQL / 跑 ALTER

**Why:** 用户审核 schema 改动只看 dbv2 CREATE TABLE 文件 —— 这是 schema 设计的唯一权威源。先跑 ALTER / 改 Go 代码 + 后补 CREATE = drift；用户无法在 PR / 评审节点拦截错误设计。

**How to apply:**
- 出 schema 改动方案时，先 propose CREATE TABLE diff（或直接 Edit dbv2 文件给用户看），不要先跑 `ALTER TABLE` 也不要先改 Go SELECT
- 用户 review CREATE 通过后，按顺序：CREATE diff → ALTER 迁移 SQL → Go 代码改动
- DROP TABLE 同理：先在 dbv2 删/标 retire，再跑 DROP，再清理 code

**违规判定：** 任何"schema 改动跨了 dbv2 CREATE 文件还在那"的状态都算违规。今天 drs 退役里跑了 DROP + 改 Go 但 21/22/23_*.sql 没同步就是这一类。
