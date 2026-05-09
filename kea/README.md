# owl_v2 kea + BIND infrastructure

## 重要：双库共存策略

owlrd 旧库**不动**（保留回退），owl_v2 新库**单独创建**。两库同住一个 postgres
容器，容量上忽略不计。

切换 owl 服务到新库由 `.env` 的 `DB_NAME` 决定（Phase B 接入时改 `DB_NAME=owl_v2`）。

## 备份位置

切换前的 v1 配置已存档：
```
/home/wisefido/owl/backup1.0/owlBack-config-v1/
├── docker-compose.yml._v1   (kea+BIND 加入前)
├── .env._v1                 (DB_NAME=owlrd 时刻)
└── env.example._v1
```

## 环境变量约定（.env）

```
# owl_v2
DB_NAME_V2=owl_v2

# kea REST API
KEA_API_HOST / PORT / USER / PASSWORD

# DDNS TSIG（kea-ddns + bind + owl service nsupdate 共用）
DDNS_TSIG_NAME / ALGORITHM / SECRET

# BIND DNS
BIND_HOST / PORT
```

## bring-up 顺序

```bash
# 1. owl 业务侧（含 owlrd 旧库；首次 init 即跑 dbv1）
docker compose up -d postgresql redis mqtt sleepace-mysql

# 2. 创建 owl_v2 新库（不影响 owlrd）
bash scripts/setup_owl_v2.sh

# 3. 启 IPAM + DNS
docker compose up -d kea-dhcp6 kea-ctrl bind

# 4. 验证
curl http://localhost:8000/  # kea-ctrl-agent 应回 "found 200"
dig @localhost tenant1.owl AAAA  # BIND 应解析

# 5. 切换 owl 服务到 owl_v2（Phase B）
#    编辑 .env: DB_NAME=owl_v2
#    重启 wisefido-* services
```

## 回退方案

```bash
# 改回 owlrd
sed -i 's/^DB_NAME=owl_v2/DB_NAME=owlrd/' .env
# 重启 services；owlrd 数据未动，立即可用
```

owl IPv6 寻址体系的运行时支撑：
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

## bring-up 顺序（首次启动）

```bash
# 1. 启 postgres + redis + mqtt + sleepace-mysql（业务侧）
docker compose up -d postgresql redis mqtt sleepace-mysql

# 2. 等 owl_v2 schema 完全初始化（首次约 30s）
docker compose logs -f postgresql | grep "ready to accept connections"

# 3. 启 IPAM + DNS（独立 stack，依赖 owl_v2 已就绪）
docker compose up -d kea-dhcp6 kea-ctrl bind

# 4. 验证
curl http://localhost:8000/  # kea-ctrl-agent 应回 "found 200"
dig @localhost tenant1.owl AAAA  # BIND 应解析

# 5. seed initial tenant（写 owl_v2 + 触发 DDNS 推 BIND）
psql -h localhost -U postgres -d owl_v2 -f ../owlRD/dbv2/90_seed_initial_tenant.sql
```

## 配置文件清单

```
kea/
  dhcp6/
    kea-dhcp6.conf       DHCPv6 daemon 配置（subnet 定义 + DDNS hooks）
  ctrl/
    kea-ctrl-agent.conf  REST API 配置（port 8000）
  ddns/
    kea-ddns.conf        DDNS 推送 BIND（共享 TSIG 密钥）

bind/
  named.conf             BIND 主配置
  zones/
    tenant1.owl.zone     tenant 1 正向 zone（AAAA records）
    d.0.0.f.ip6.arpa.zone  反向 zone (fd00::/8 ULA 范围)
  keys/
    ddns-update.key      DDNS TSIG 密钥（与 kea-ddns 共享）
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

## 注意

- **首次启动时 BIND 会因 zone 文件缺失报错**：seed 脚本生成 zone 后 reload BIND（`docker exec owl-bind rndc reload`）
- **DDNS 共享密钥**：开发环境用明文，**production 必须 rotate 真密钥**
- **kea config 是 hjson**（带注释的 JSON）：编辑前用 `kea-shell` 验证语法
- **bind ip6.arpa zone 长度**：fd00::/8 范围反向 zone 名是 `0.0.0.0.f.d.ip6.arpa`（按 RFC 3596）

## 待实施

当前是骨架。Phase B 完整对接 owl service 时需要：
1. owl-data 加 kea-ctrl-agent client（HTTP wrapper）
2. CreateTenant / Branch / Unit / Room / Bed handler 调用 kea REST
3. kea-ddns 集成 BIND 自动维护 zone
4. seed scripts 90_-93_ 写完
