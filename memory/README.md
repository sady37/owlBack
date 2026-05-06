# Claude Code Memory Snapshots

本目录是 owl 项目的 [Claude Code](https://claude.com/claude-code) 本地 memory 快照，
用于版本化保存 Claude 关于本项目的项目知识 / 设计决策 / 教训记录 / 反馈。

## 关系

- **本机权威 source**：`~/.claude/projects/-home-wisefido-owl/memory/`
  Claude Code 实时读写位置；每次会话从此处加载上下文。
- **本目录 (owlBack/memory/)**：Git 同步快照
  团队 review / 多机同步 / 历史回溯 用。

## 同步

Claude 在本机更新 memory 后，手动同步到本目录 + 提交：

```bash
cp ~/.claude/projects/-home-wisefido-owl/memory/*.md owlBack/memory/
cd owlBack
git add memory/ && git commit -m "memory: <date> snapshot"
git push
```

## 索引

`MEMORY.md` 是 Claude 的索引文件，列出所有 memory 主题及一行说明。
单文件命名规则：`<topic>.md`，类型字段（user / feedback / project / reference）在 frontmatter。

## 不进 build

`memory/` 已在 `.dockerignore` 里隐式排除（因 `*.md` 已 ignored）。
