# 提交给 ChatGPT - 立即操作指南

> **状态**: ✅ 所有材料已准备完成，可以立即提交给 ChatGPT

---

## 🚀 立即操作步骤

### 步骤 1: 打开完整提示词文件

**文件路径**:
```
/Users/sady3721/project/owlBack/docs/reviews/CHATGPT_ROUND1_PROMPT_COMPLETE.md
```

**或者使用命令**:
```bash
cd /Users/sady3721/project/owlBack
cat docs/reviews/CHATGPT_ROUND1_PROMPT_COMPLETE.md
```

---

### 步骤 2: 复制提示词内容

**复制范围**:
- **开始**: 从 "请审查以下 Go 代码..." 开始
- **结束**: 到 "请详细审查并提供反馈。" 结束

**提示**: 文件中的提示词已经包含所有代码，无需额外替换。

---

### 步骤 3: 打开 ChatGPT

访问 ChatGPT 网站或应用，开始新的对话。

---

### 步骤 4: 粘贴并提交

1. 将复制的完整提示词粘贴到 ChatGPT 输入框
2. 点击发送
3. 等待 ChatGPT 审查反馈

---

### 步骤 5: 记录反馈

**反馈文件位置**:
```
/Users/sady3721/project/owlBack/docs/reviews/chatgpt_round1_sensor_fusion.md
```

**记录内容**:
- ChatGPT 发现的问题列表
- 代码质量评分
- 改进建议
- 总体评价

---

## 📋 提示词文件内容预览

提示词包含：
- ✅ 7 个审查维度（代码质量、错误、性能、并发、错误处理、最佳实践、安全性）
- ✅ 完整的 `sensor_fusion.go` 代码（260 行）
- ✅ 完整的 `card.go` 相关代码
- ✅ 完整的 `sleepace.go` 相关代码
- ✅ 设计背景说明（融合规则、已知问题）
- ✅ 输出要求（问题列表、评分、建议）

---

## ✅ 检查清单

提交前确认：
- [x] 提示词文件已准备（`CHATGPT_ROUND1_PROMPT_COMPLETE.md`）
- [x] 代码已包含在提示词中
- [x] 审查重点已明确
- [x] 反馈记录模板已准备（`chatgpt_round1_sensor_fusion.md`）

---

## 📝 预期反馈格式

ChatGPT 应该提供：
1. **问题列表**（按严重性排序）
   - 问题描述
   - 位置（文件:行号）
   - 严重性（高/中/低）
   - 修复建议

2. **代码质量评分**（1-10 分）
   - 总体评分
   - 分项评分

3. **改进建议**
   - 高优先级建议
   - 中优先级建议

4. **总体评价**
   - 优点
   - 需要改进的地方

---

## 🔄 下一步

### 收到 ChatGPT 反馈后

1. **记录反馈**: 保存到 `chatgpt_round1_sensor_fusion.md`
2. **提交给 Claude**: 将反馈提交给我（Claude）进行分析和回应
3. **修复代码**: 根据反馈修复问题
4. **第2轮审查**: 将改进后的代码提交给 ChatGPT 进行第2轮审查

---

## 📞 需要帮助？

如果遇到问题：
- 提示词文件找不到：检查 `docs/reviews/CHATGPT_ROUND1_PROMPT_COMPLETE.md`
- 代码不完整：文件已包含完整代码，无需额外操作
- 反馈格式不对：参考 `docs/reviews/27_ChatGPT_Feedback_Template.md`

---

**准备完成时间**: 2024-12-19  
**状态**: ✅ 可以立即提交给 ChatGPT

