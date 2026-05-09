# owl_v2 kea + BIND infrastructure

owl_v2 IPv6 寻址体系运行时支撑：
- **kea-dhcp6** + **kea-ctrl-agent**：IPAM，prefix delegation REST API
- **BIND9**：DNS 短名解析 + ip6.arpa 反向

## 角色定位

owl 设备**不直接**走 DHCPv6 协议（设备是 IPv4 + MQTT，owl 地址是逻辑标识）。kea 在 owl 体系里是 **IPAM allocator**：

```
owl Go service (wisefido-data)
     ↓ POST /command (kea REST API)
kea-ctrl-agent (port 8000)
     ↓ unix socket
kea-dhcp6
     ↓ DHCPv6-PD 协议算法
返回分配的 prefix
     ↓
owl 写入 owl_v2.tenants/branches/units/...
     ↓ DDNS update
BIND9 zone files
```

## 数据库策略：双库共存，**不动旧 volume**

postgres 单实例同时托管两库：
- `owlrd`（旧）— 保留作物理冗余备份；不连任何服务
- `owl_v2`（新）— owl 服务的实际工作库

两库互不影响。owl 服务通过 `.env DB_NAME=owl_v2` 只连新库。

### 为什么不清 volume

逻辑 dump 已完整：`/home/wisefido/owl/backup1.0/db-rebuild-20260508-080819/`
- `full.sql` (2.5 GB) — pg_dump 完整 owlrd
- 关键表单独 dump（device_store / units / rooms / layouts_by_name.json）

加上 volume 内 owlrd 保留，**3 层备份冗余**（逻辑 dump + 工程 tarball + 物理 volume）。

### docker-compose `POSTGRES_DB=owl_v2` 的语义

- 首次 init（空 volume）：postgres 创建 owl_v2 默认库 + 跑 initdb.d 全套 dbv2 schema
- 已存在 volume：env 不生效，老 owlrd 保留；用 `setup_owl_v2.sh` 在同实例内补建 owl_v2

两条路径都到达同一终点：postgres 内有 owl_v2 + dbv2 schema。

### 真要回退（紧急情况）

```bash
# 选项 A：直接连回 owlrd
sed -i 's/^DB_NAME=owl_v2/DB_NAME=owlrd/' .env
sudo systemctl restart owlback.*

# 选项 B：从 full.sql 重建 owlrd（如 volume 也丢了）
docker exec -i owl-postgresql psql -U postgres -c "CREATE DATABASE owlrd_restore;"
cat /home/wisefido/owl/backup1.0/db-rebuild-*/full.sql | \
  docker exec -i owl-postgresql psql -U postgres -d owlrd_restore
```

## 备份位置

```
/home/wisefido/owl/backup1.0/
├── db-rebuild-20260508-080819/        full owlrd dump (2.5 GB)
└── owlBack-config-v1/
    ├── docker-compose.yml._v1          (kea+BIND 加入前)
    ├── .env._v1                        (DB_NAME=owlrd 时刻)
    └── env.example._v1
```

## 环境变量约定（.env）

```
# DB
DB_NAME=owl_v2
DB_HOST / PORT / USER / PASSWORD

# kea REST API
KEA_API_HOST / PORT / USER / PASSWORD

# DDNS TSIG（kea-ddns + bind + owl service nsupdate 共用）
DDNS_TSIG_NAME / ALGORITHM / SECRET

# BIND DNS
BIND_HOST / PORT
```

## bring-up 顺序

### 路径 A：volume 已有 owlrd（当前）— 非破坏性

```bash
cd /home/wisefido/owl/owlBack

# 1. 启全部基础设施（owlrd 旧数据保留在 volume）
docker compose up -d

# 2. 等 postgres ready
docker compose logs -f postgresql | grep "ready to accept connections"

# 3. 在同实例内创建 owl_v2 + 跑 dbv2 schema
bash scripts/setup_owl_v2.sh

# 4. 验证（应见两库 + dbv2 表）
psql -h localhost -U postgres -lA | grep -E "owlrd|owl_v2"
psql -h localhost -U postgres -d owl_v2 -c "\dt" | head -20

# 5. 验证 IPAM + DNS
curl http://localhost:8000/                         # kea-ctrl-agent
dig @localhost tenant1.owl AAAA                     # BIND

# 6. 启 wisefido-* services（连 owl_v2 via .env DB_NAME=owl_v2）
sudo systemctl start owlback.cardagg owlback.data owlback.iot owlback.qinglan owlback.sensor owlback.sleepace
```

### 路径 B：全新 volume（清空后重启）— 自动 init

```bash
docker volume rm owlBack_postgres_data
docker compose up -d         # 自动 init owl_v2 with dbv2 schema (initdb.d)
# 跳过 setup_owl_v2.sh；其他步骤同路径 A
```

## 回退到 owlrd（紧急情况）

```bash
docker compose down
# 恢复旧 volume
docker volume create owlBack_postgres_data
docker run --rm -v owlBack_postgres_data:/restore postgres:15 \
  pg_restore -d postgres /backup1.0/db-rebuild-20260508-080819/full.sql

# 改 docker-compose.yml: POSTGRES_DB=owlrd, initdb.d=../owlRD/dbv1
# 改 .env: DB_NAME=owlrd
docker compose up -d
```

## owl Go service 调用 kea REST 范例

```go
// 创建新 tenant：分配 /48 prefix
type Cmd struct {
    Command string                 `json:"command"`
    Service []string               `json:"service"`
    Args    map[string]interface{} `json:"arguments"`
}

cmd := Cmd{
    Command: "subnet6-add",
    Service: []string{"dhcp6"},
    Args: map[string]interface{}{
        "subnet6": []map[string]interface{}{
            {"subnet": "fd00:0:0001::/48", "id": 1},
        },
    },
}
resp, _ := http.Post("http://localhost:8000/", "application/json",
                     bytes.NewReader(jsonMarshal(cmd)))
```

## 配置文件清单

```
kea/
  dhcp6/
    kea-dhcp6.conf       DHCPv6 daemon 配置（subnet 定义 + DDNS hooks）
  ctrl/
    kea-ctrl-agent.conf  REST API 配置（port 8000）

bind/
  named.conf             BIND 主配置
  zones/
    tenant1.owl.zone     tenant 1 正向 zone（AAAA records）
    d.0.0.f.ip6.arpa.zone  反向 zone (fd00::/8 ULA 范围)
  keys/
    ddns-update.key      DDNS TSIG 密钥（与 kea-ddns 共享）

scripts/
  setup_owl_v2.sh        手动 setup 脚本（默认首次 docker compose up
                         自动跑 init.d，无需手动；保留作灾备 / 重跑用）
```

## 待 Phase B

当前是基础设施骨架。Phase B 完整对接 owl service 时需要：
1. owl-data 加 kea-ctrl-agent client（HTTP wrapper）
2. CreateTenant / Branch / Unit / Room / Bed handler 调用 kea REST
3. kea-ddns 集成 BIND 自动维护 zone
4. seed scripts 90_-93_ 写完
