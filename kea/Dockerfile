# kea 2.4 for owl_v2 IPAM
# Ubuntu 24.04 LTS 提供 kea 2.4.1（22.04 仅有 2.0.2）
# 24.04 多了 pgsql_cb 配置后端 + DDNS 联动稳定 + lease_cmds 等开源 hooks
FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
        kea-dhcp6-server \
        kea-ctrl-agent \
        kea-dhcp-ddns-server \
        ca-certificates \
        netcat-openbsd \
        dnsutils \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /var/lib/kea /var/run/kea \
    && chmod 0755 /var/lib/kea /var/run/kea

# 547/udp DHCPv6 (owl 设备不发 DHCPv6，监听仅作占位)
# 8000/tcp ctrl-agent REST API
EXPOSE 547/udp 8000/tcp

# entrypoint 由 docker-compose 指定（dhcp6 vs ctrl 走不同 service）
