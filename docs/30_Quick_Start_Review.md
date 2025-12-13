# 快速开始代码审查

> **目的**: 快速启动 ChatGPT 和 Claude 的迭代审查流程

---

## 🚀 5 步快速开始

### 步骤 1: 准备代码（5分钟）

```bash
cd /Users/sady3721/project/owlBack

# 1. 格式化代码
/usr/local/go/bin/go fmt ./...

# 2. 编译验证
/usr/local/go/bin/go build ./wisefido-sensor-fusion/cmd/wisefido-sensor-fusion

# 3. 选择关键文件
# 推荐: wisefido-sensor-fusion/internal/fusion/sensor_fusion.go
```

---

### 步骤 2: ChatGPT 第1轮审查（10分钟）

#### 2.1 复制代码

复制以下文件内容：
- `wisefido-sensor-fusion/internal/fusion/sensor_fusion.go`

#### 2.2 使用提示词

```
请审查以下 Go 代码，重点关注：

1. 代码质量和最佳实践
2. 潜在的错误和 bug
3. 性能问题（特别是 N+1 查询）
4. 并发安全问题
5. 错误处理是否完善

请提供：
1. 发现的问题列表（按严重性排序）
2. 具体的改进建议
3. 代码质量评分（1-10 分）

[粘贴代码]
```

#### 2.3 记录反馈

保存到: `docs/reviews/chatgpt_round1_sensor_fusion.md`

---

### 步骤 3: 提交给 Claude（现在）

#### 3.1 提交内容

**请将以下内容提交给 Claude**:

1. **ChatGPT 的反馈**（从步骤 2 获得）
2. **相关代码文件**（`sensor_fusion.go`）
3. **设计文档**（如有）

#### 3.2 提交提示词

```
请分析以下 ChatGPT 的代码审查反馈，并作出回应：

## ChatGPT 审查反馈

[粘贴 ChatGPT 的反馈]

## 相关代码

[粘贴 sensor_fusion.go 代码]

请：
1. 分析每个问题的合理性
2. 解释设计决策（如果不同意 ChatGPT 的观点）
3. 提供修复方案（如果同意）
4. 评估 ChatGPT 审查的质量
5. 准备改进后的代码供第2轮审查
```

---

### 步骤 4: Claude 回应（自动）

Claude 将：
1. ✅ 分析 ChatGPT 的反馈
2. ✅ 解释设计决策
3. ✅ 提供修复方案
4. ✅ 改进代码
5. ✅ 记录回应

**回应将保存到**: `docs/reviews/claude_response_round1_sensor_fusion.md`

---

### 步骤 5: ChatGPT 第2轮审查（10分钟）

#### 5.1 准备材料

提交给 ChatGPT：
1. **改进后的代码**（Claude 修复后）
2. **Claude 的回应**（从步骤 4）
3. **第1轮反馈**（从步骤 2）

#### 5.2 使用第2轮提示词

```
请进行第2轮代码审查：

## 第1轮反馈

[粘贴 ChatGPT 第1轮反馈]

## Claude 的回应

[粘贴 Claude 的回应]

## 改进后的代码

[粘贴改进后的代码]

请评估：
1. 第1轮发现的问题是否已修复？
2. Claude 的回应是否合理？
3. 改进后的代码是否解决了问题？
4. 是否有新的问题？
5. 最终代码质量评分（1-10 分）
```

#### 5.3 记录第2轮反馈

保存到: `docs/reviews/chatgpt_round2_sensor_fusion.md`

---

## 📊 审查进度跟踪

### 当前状态

| 步骤 | 状态 | 完成时间 |
|------|------|---------|
| 准备代码 | ✅ 完成 | 2024-12-19 |
| ChatGPT 第1轮 | ⬜ 待开始 | |
| Claude 回应 | ⬜ 待开始 | |
| 代码修复 | ⬜ 待开始 | |
| ChatGPT 第2轮 | ⬜ 待开始 | |

---

## 🎯 审查目标

### 第1轮目标
- 发现主要问题
- 初步评分 ≥ 6.0/10

### 第2轮目标
- 问题修复验证
- 最终评分 ≥ 8.0/10

---

## 📝 模板文件

- [ChatGPT 反馈模板](../27_ChatGPT_Feedback_Template.md)
- [Claude 回应模板](../28_Claude_Response_Template.md)
- [审查工作流程](../29_Review_Workflow.md)

---

## ✅ 检查清单

### 提交给 Claude 前

- [ ] 代码格式正确
- [ ] 编译通过
- [ ] 有详细注释
- [ ] ChatGPT 反馈已记录
- [ ] 相关代码已准备

### Claude 回应后

- [ ] 所有问题已分析
- [ ] 修复方案已提供
- [ ] 代码已改进
- [ ] 回应已记录

### 提交给 ChatGPT 第2轮前

- [ ] 改进后的代码已准备
- [ ] Claude 回应已记录
- [ ] 修复的问题列表已整理

---

**最后更新**: 2024-12-19

