# Sleepace MySQL 分析记录

日期：2026-03-26

---

## 一、MySQL 连接信息

| 项目 | 值 |
|------|----|
| 类型 | Docker 容器（`sleepace-mysql`，镜像 `mysql:5.7`） |
| Host | `127.0.0.1`（宿主机访问）/ `172.18.0.4`（Docker 内部） |
| Port | `3306` |
| User | `root` |
| Password | `mysql` |
| Docker 网络 | `owlback_default` |

**连接命令：**
```bash
mysql -h 127.0.0.1 -P 3306 -uroot -pmysql
```

---

## 二、数据库列表

```sql
SHOW DATABASES;
```

| 数据库 | 用途 |
|--------|------|
| `z1_manage` | 设备出厂/批次管理 |
| `sleepace_pro_user` | 用户与设备绑定关系、设备状态 |
| `sleepace_tb_system` | 告警配置、告警记录 |
| `sleepace_tb_data` | 睡眠报告数据 |
| `sleepace_pro_version` | 固件版本管理 |

---

## 三、设备出厂管理（`z1_manage.device_random`）

`device_batch` 表为空，实际设备出厂数据存在 **`device_random`** 表。

### 查询所有设备数量及批次
```sql
USE z1_manage;
SELECT batch, COUNT(*) AS qty, MIN(createTime) AS create_time
FROM device_random
GROUP BY batch
ORDER BY batch;
```

**结果（共 59 台，6 个批次）：**

| batch | 数量 | 入库时间 |
|-------|------|----------|
| 1788 | 4 台 | 2024-05-28 |
| 1799 | 51 台 | 2024-07-22 |
| 1814 | 1 台 | 2024-10-22 |
| 1845 | 1 台 | 2024-12-24 |
| 1846 | 1 台 | 2024-12-24 |
| 1855 | 3 台 | 2025-02-15 |

### 查询指定批次设备
```sql
SELECT seqid, deviceId, deviceName, macId, deviceType, status, createTime
FROM device_random
WHERE batch = 1855;
```

**批次 1855 的 3 台设备：**

| seqid | deviceId | deviceName | macId |
|-------|----------|------------|-------|
| 42490774 | `r0nqo00b34vtf` | BM87225200672 | f041c827e116 |
| 42490779 | `ka9hv61zy0qy7` | BM87225200677 | f041c827e11b |
| 42490901 | `zigkw5main1dx` | BM87225200799 | f041c827e195 |

### 查询指定 deviceId 是否存在
```sql
SELECT * FROM z1_manage.device_random WHERE deviceId = '9qxqhv9lr6wzx';
```

**结论：** `deviceId = 9qxqhv9lr6wzx`（BM87224C02300，batch 1846）存在于数据库。

> ⚠️ 注意：查询时必须使用 `-h 127.0.0.1` 通过 TCP 连接，不能省略；同时避免将查询结果通过 `| grep -v Warning` 管道过滤，否则当结果为空时 grep 返回 exit code 1，易误判为无数据。

---

## 四、设备字段说明

| 字段 | 说明 |
|------|------|
| `seqid` | 主键，自增 |
| `deviceId` | 设备唯一 ID（随机字符串） |
| `deviceName` | 设备名，格式 BM8722+年月+序号 |
| `macId` | MAC 地址 |
| `deviceType` | 设备型号（`49` = 睡眠板） |
| `batch` | 出厂批次号 |
| `bleName` | 蓝牙名称 |
| `extContent` | WiFi 配置（SSID,密码,服务器,端口） |
| `status` | `1` = 正常 |

---

## 五、MySQL 连接进程分析

### 检查连接来源
```bash
# MySQL 内部视角
mysql -h 127.0.0.1 -P 3306 -uroot -pmysql -e "
SELECT user, SUBSTRING_INDEX(host,':',1) AS from_host, db, command, COUNT(*) AS conns
FROM information_schema.processlist
WHERE id != CONNECTION_ID()
GROUP BY user, SUBSTRING_INDEX(host,':',1), db, command
ORDER BY from_host, db;"

# 宿主机进程视角
sudo lsof -i TCP:3306 -nP | grep -v LISTEN | grep -v docker-pr
```

### 结论

| 来源 | 进程 | PID | 连接数 | 数据库 |
|------|------|-----|--------|--------|
| 宿主机 (`127.0.0.1`) | Java `sleepace-service` | 920376 | 25 条 | 5 个库各 5 条 |
| Docker 内部 (`172.18.0.1`) | — | — | — | （同上，NAT 后显示） |

**唯一连接 MySQL 的进程：**
```
PID 920376: /usr/bin/java ... com.medica.SleepaceStartServer
路径: /home/wisefido/owl/sleepace/sleepace-service/
启动于: Mar22（持续运行）
子进程: ChildProcessMain P_0~P_5（不持有 MySQL 连接）
```

连接路径：
```
Java进程 (127.0.0.1) → docker-proxy (PID 918922) → sleepace-mysql容器 (172.18.0.4:3306)
```

---

## 六、服务架构说明

```
owlFront (Vue 3)
    ↓
wisefido-data (Go, :8080) ← PostgreSQL (owlrd)
wisefido-sleepace (Go, :8083) ← PostgreSQL (owlrd) + 调用 Java HTTP API
    ↓ HTTP
sleepace-service (Java, :8090) ← MySQL Docker (:3306)
    └─ sleepace_pro_user, sleepace_tb_system, sleepace_tb_data, sleepace_pro_version, z1_manage
```

**Go `wisefido-sleepace` 不连 MySQL**，它通过 HTTP 调用 Java `sleepace-service`，自身数据存 PostgreSQL (`owlrd`)。