# OwlBack 手动验证指南

> **说明**: 如果 Go 命令不在 PATH 中，请按照本指南手动验证

---

## 🔧 环境设置

### 1. 检查 Go 安装

```bash
# 方法 1: 检查 Go 是否在 PATH 中
which go

# 方法 2: 检查常见安装位置
ls -la /usr/local/go/bin/go
ls -la ~/go/bin/go

# 方法 3: 检查 Go 版本（如果找到）
/usr/local/go/bin/go version
```

### 2. 设置 Go 环境变量

如果 Go 已安装但不在 PATH 中，添加到 PATH:

```bash
# 对于 zsh (macOS 默认)
export PATH=$PATH:/usr/local/go/bin
# 或
export PATH=$PATH:~/go/bin

# 添加到 ~/.zshrc 使其永久生效
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
source ~/.zshrc
```

---

## ✅ 验证步骤

### 步骤 1: 代码格式检查

```bash
cd /Users/sady3721/project/owlBack

# 检查代码格式
go fmt ./...

# 如果没有输出，说明格式正确
```

### 步骤 2: 代码规范检查

```bash
# 检查代码规范
go vet ./...

# 查看输出，应该没有错误
```

### 步骤 3: 编译所有服务

```bash
# 编译 wisefido-radar
cd wisefido-radar
go build ./cmd/wisefido-radar
echo "✅ wisefido-radar 编译成功" || echo "❌ wisefido-radar 编译失败"

# 编译 wisefido-sleepace
cd ../wisefido-sleepace
go build ./cmd/wisefido-sleepace
echo "✅ wisefido-sleepace 编译成功" || echo "❌ wisefido-sleepace 编译失败"

# 编译 wisefido-data-transformer
cd ../wisefido-data-transformer
go build ./cmd/wisefido-data-transformer
echo "✅ wisefido-data-transformer 编译成功" || echo "❌ wisefido-data-transformer 编译失败"

# 编译 wisefido-sensor-fusion
cd ../wisefido-sensor-fusion
go build ./cmd/wisefido-sensor-fusion
echo "✅ wisefido-sensor-fusion 编译成功" || echo "❌ wisefido-sensor-fusion 编译失败"
```

### 步骤 4: 依赖验证

```bash
cd /Users/sady3721/project/owlBack

# 验证所有模块的依赖
go mod verify

# 应该输出: all modules verified
```

### 步骤 5: 运行验证脚本

```bash
cd /Users/sady3721/project/owlBack

# 运行验证脚本
chmod +x scripts/verify.sh
./scripts/verify.sh
```

---

## 📊 验证结果记录

### 验证检查清单

- [ ] Go 环境配置正确
- [ ] 代码格式检查通过 (`go fmt`)
- [ ] 代码规范检查通过 (`go vet`)
- [ ] wisefido-radar 编译成功
- [ ] wisefido-sleepace 编译成功
- [ ] wisefido-data-transformer 编译成功
- [ ] wisefido-sensor-fusion 编译成功
- [ ] 依赖验证通过 (`go mod verify`)

### 编译结果

| 服务 | 状态 | 错误信息（如有） |
|------|------|-----------------|
| wisefido-radar | ⬜ | |
| wisefido-sleepace | ⬜ | |
| wisefido-data-transformer | ⬜ | |
| wisefido-sensor-fusion | ⬜ | |

### 发现的问题

1. _________________________________
2. _________________________________
3. _________________________________

---

## 🔍 常见问题

### 问题 1: `go: command not found`

**解决方案**:
```bash
# 找到 Go 安装路径
find /usr/local -name "go" -type f 2>/dev/null
find ~ -name "go" -type f -path "*/bin/go" 2>/dev/null

# 添加到 PATH
export PATH=$PATH:/usr/local/go/bin
```

### 问题 2: 模块依赖错误

**解决方案**:
```bash
# 下载依赖
go mod download

# 整理依赖
go mod tidy
```

### 问题 3: 编译错误

**解决方案**:
1. 检查错误信息
2. 查看相关文档
3. 检查依赖是否正确安装

---

## 📝 验证报告模板

### 验证结果

**验证日期**: _______________

**Go 版本**: _______________

**验证人员**: _______________

#### 编译结果

- [ ] wisefido-radar: ✅ / ❌
- [ ] wisefido-sleepace: ✅ / ❌
- [ ] wisefido-data-transformer: ✅ / ❌
- [ ] wisefido-sensor-fusion: ✅ / ❌

#### 代码检查

- [ ] `go fmt`: ✅ / ❌
- [ ] `go vet`: ✅ / ❌
- [ ] `go mod verify`: ✅ / ❌

#### 总体评估

- [ ] 通过，可以部署
- [ ] 有条件通过，需要修复以下问题:
  1. _______________
  2. _______________
- [ ] 不通过，需要重大修复

---

## 🚀 快速验证命令

```bash
# 一键验证（需要 Go 在 PATH 中）
cd /Users/sady3721/project/owlBack && \
go fmt ./... && \
go vet ./... && \
cd wisefido-radar && go build ./cmd/wisefido-radar && \
cd ../wisefido-sleepace && go build ./cmd/wisefido-sleepace && \
cd ../wisefido-data-transformer && go build ./cmd/wisefido-data-transformer && \
cd ../wisefido-sensor-fusion && go build ./cmd/wisefido-sensor-fusion && \
cd .. && go mod verify && \
echo "✅ 所有验证通过"
```

---

**最后更新**: 2024-12-19

