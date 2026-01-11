# 雷达订阅功能快速测试指南

## 🚀 快速开始

### 1. 启动服务

```bash
cd /home/wisefido/owl-project/owlBack/wisefido-radar
./start-radar.sh
```

### 2. 验证订阅管理器启动

查看日志，确认看到以下信息：

```json
{"level":"info","msg":"Subscription manager started","renewal_interval_minutes":50,"subscription_duration":3600,"auto_subscribe":true}
```

### 3. 验证自动订阅

当设备首次连接时，应该看到：

```json
{"level":"info","msg":"Device auto-created from device_store on MQTT connection","device_id":"xxx","uid":"xxx"}

{"level":"info","msg":"Auto-subscribed on device first connection","uid":"xxx"}
```

### 4. 检查 Redis 订阅状态

```bash
# 查看所有订阅
redis-cli -a TeLunSu-36kr --scan --pattern "radar:subscription:*"

# 查看特定设备的订阅信息
redis-cli -a TeLunSu-36kr GET radar:subscription:{uid}
```

### 5. 验证自动续订

等待 50 分钟（或修改配置缩短测试时间），应该看到：

```json
{"level":"info","msg":"Renewing subscription","uid":"xxx","expires_at":"2026-01-09T01:00:00Z"}

{"level":"info","msg":"Successfully renewed subscription","uid":"xxx"}

{"level":"info","msg":"Renewed subscriptions","count":5}
```

## ⚡ 快速测试（缩短续订间隔）

如果想快速测试自动续订功能，可以临时修改配置：

```bash
# 设置续订间隔为 1 分钟（仅用于测试）
export RADAR_SUBSCRIPTION_RENEWAL_INTERVAL=1
export RADAR_SUBSCRIPTION_DURATION=120  # 订阅时长 2 分钟
export RADAR_SUBSCRIPTION_RENEWAL_ADVANCE=1  # 提前 1 分钟续订

# 重启服务
./start-radar.sh
```

这样可以在 2 分钟内看到续订效果。

## 📋 测试检查清单

- [ ] 服务启动成功
- [ ] 订阅管理器启动日志出现
- [ ] Redis 连接正常
- [ ] 设备首次连接时自动订阅
- [ ] Redis 中有订阅记录
- [ ] 自动续订功能正常（等待或缩短配置）

## 🔍 故障排查

### 问题1: 订阅管理器未启动

**检查**:
```bash
# 查看日志
tail -f /tmp/owl_radar_startup.log | grep -i subscription
```

**可能原因**:
- 订阅管理器创建失败
- 配置项未正确加载

### 问题2: 自动订阅未触发

**检查**:
```bash
# 检查设备是否首次连接
redis-cli -a TeLunSu-36kr EXISTS radar:subscription:{uid}

# 检查配置
echo $RADAR_SUBSCRIPTION_AUTO
```

**可能原因**:
- `RADAR_SUBSCRIPTION_AUTO=false`
- 设备已存在订阅记录
- MQTT 消息未正确接收

### 问题3: 自动续订未触发

**检查**:
```bash
# 检查订阅记录
redis-cli -a TeLunSu-36kr GET radar:subscription:{uid}

# 检查配置
echo $RADAR_SUBSCRIPTION_RENEWAL_INTERVAL
```

**可能原因**:
- 续订间隔配置过长
- 订阅已过期
- 续订检查失败

## 📝 测试脚本

已创建的测试脚本：

1. **`test-subscription.sh`** - 基础功能测试
2. **`test-startup.sh`** - 服务启动测试
3. **`test-subscription-detailed.sh`** - 详细代码检查

运行测试：
```bash
./test-subscription.sh
./test-startup.sh
./test-subscription-detailed.sh
```

## ✅ 测试完成标准

- ✅ 服务正常启动
- ✅ 订阅管理器正常启动
- ✅ 设备首次连接时自动订阅
- ✅ Redis 中有订阅记录
- ✅ 订阅自动续订（50分钟后或缩短配置后）

**所有测试通过后，功能即可投入使用！** 🎉
