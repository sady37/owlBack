---
name: feedback_work_on_main_no_branches
description: owlBack git 工作流铁律 — 直接在 main 上提交，不开 feature 分支
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 2bf7aa0b-7142-40c1-812e-722af1229f2d
---

owlBack（repo root = /home/wisefido/owl/owlBack）**直接在 main 上 commit，不再开 feature 分支**。2026-06-01 用户明确："只留 main"+"不能再搞分支了"，把累积的 belief 工作 fast-forward 进 main 后删掉 feature 分支（本地+远程），其余别人分支（wenhe-study / feat/radar-fall-verifier）不动。

**Why**：用户单人推进、分支只增维护负担与混乱；他要单线 main。

**How to apply**：
- 用户说 commit/push 时，直接在 main 上提交并 push origin main，**不要新建分支**（与默认"on default branch → branch first"相反，此 repo 是已确认例外）。
- 不擅自删别人的远程分支；只清理用户明确指认的。

**坑（2026-06-01 踩过）**：`mqtt/config/mosquitto.conf` 是 docker 服务配置，owner uid 1883、目录 `mqtt/config/` 也 1883:1883，wisefido(1002) 无写权 → `git checkout`/`reset --hard` 会因无法 unlink 报 Permission denied。该文件已被 git 跟踪且常被服务带外改写。**绕法**：内容若已等于目标 commit，用 `git reset --mixed <commit>`（移 ref+index，不碰工作树）再 `git restore .`（恢复其他被 checkout 删掉的文件，mosquitto 因匹配被跳过），避免直接写该文件。
