// Package ddns wraps `nsupdate` (BIND9 dynamic update) with TSIG auth.
//
// 使用方式：
//
//	c := ddns.New(ddns.Config{
//	    Server: "127.0.0.1", Port: 5353,
//	    KeyName: "ddns-update", Algorithm: "hmac-sha256",
//	    Secret: os.Getenv("DDNS_TSIG_SECRET"),
//	    OwlDomain: "owl.",  // suffix
//	})
//	c.RegisterDevice(ctx, "fd00:0:3:101:65:101:0:1/128", "john-bed", "tenant3.owl.")
//	→ 推 forward AAAA + reverse PTR 到 BIND
//
// 实现：通过 os/exec 调用 nsupdate 二进制（系统标配），用 TSIG key 文件做鉴权。
// 不依赖 Go DNS 库（github.com/miekg/dns 等），保持依赖最小。
package ddns

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"os/exec"
	"strings"

	"owl-common/spatial"
)

// Config 配置 DDNS client。所有字段除 Port 都 required。
type Config struct {
	Server    string // BIND 主机名 / IP，e.g. "127.0.0.1" or "owl-bind"
	Port      int    // BIND DNS 端口；空(0)默认 53
	KeyName   string // TSIG key name，e.g. "ddns-update"
	Algorithm string // TSIG algorithm，e.g. "hmac-sha256"
	Secret    string // TSIG shared secret base64
	OwlDomain string // owl 域名 suffix，e.g. "owl." (注意尾点)
}

// Client wraps nsupdate. 线程安全 — 每次调用 fork 一次 nsupdate process。
type Client struct {
	cfg Config
}

// New constructs a Client. 返回前不验证连通性（lazy, 调 Register* 时才知道是否能连）。
func New(cfg Config) (*Client, error) {
	if cfg.Server == "" || cfg.KeyName == "" || cfg.Secret == "" || cfg.Algorithm == "" {
		return nil, errors.New("ddns: Server / KeyName / Algorithm / Secret all required")
	}
	if cfg.Port == 0 {
		cfg.Port = 53
	}
	if cfg.OwlDomain == "" {
		cfg.OwlDomain = "owl."
	}
	if !strings.HasSuffix(cfg.OwlDomain, ".") {
		cfg.OwlDomain += "."
	}
	if _, err := exec.LookPath("nsupdate"); err != nil {
		return nil, fmt.Errorf("ddns: nsupdate binary not found in PATH (apt install dnsutils): %w", err)
	}
	return &Client{cfg: cfg}, nil
}

// RegisterDevice 注册一个 device 的 DNS records (forward AAAA + reverse PTR)。
//
//	addr        - device /128 IPv6 spatial address
//	shortName   - DNS 短名 (一段 label)，e.g. "john-bed" / "kitchen-radar"
//	zone        - 该名所属 zone, e.g. "tenant3.owl." (尾点)
//
// 写入：
//   - forward: <shortName>.<zone> AAAA <addr>     in zone "<zone>"
//   - reverse: <addr ip6.arpa>    PTR  <shortName>.<zone>  in zone "0.0.0.0.0.0.d.f.ip6.arpa."
//
// 使用 update add（不是 replace），所以同名重复 add 会累积；幂等性请由调用方
// 保证（先 UnregisterDevice 再 Register，或检查后再加）。TTL 默认 300s。
func (c *Client) RegisterDevice(ctx context.Context, addr netip.Addr, shortName, zone string) error {
	if !spatial.IsOwlAddr(addr) {
		return fmt.Errorf("ddns: addr %s not in owl namespace fd00:0000::/32", addr)
	}
	if shortName == "" || zone == "" {
		return errors.New("ddns: shortName + zone required")
	}
	if !strings.HasSuffix(zone, ".") {
		zone += "."
	}
	fqdn := shortName + "." + zone
	ptrName := spatial.ReverseDNS(addr) + "."
	revZone := "0.0.0.0.0.0.d.f.ip6.arpa."

	cmds := fmt.Sprintf(`server %s %d
zone %s
update add %s 300 AAAA %s
send
zone %s
update add %s 300 PTR %s
send
`, c.cfg.Server, c.cfg.Port,
		zone, fqdn, addr.String(),
		revZone, ptrName, fqdn)

	return c.run(ctx, cmds)
}

// UnregisterDevice 删除 device 的 forward + reverse records。
func (c *Client) UnregisterDevice(ctx context.Context, addr netip.Addr, shortName, zone string) error {
	if !strings.HasSuffix(zone, ".") {
		zone += "."
	}
	fqdn := shortName + "." + zone
	ptrName := spatial.ReverseDNS(addr) + "."
	revZone := "0.0.0.0.0.0.d.f.ip6.arpa."

	cmds := fmt.Sprintf(`server %s %d
zone %s
update delete %s AAAA
send
zone %s
update delete %s PTR
send
`, c.cfg.Server, c.cfg.Port,
		zone, fqdn,
		revZone, ptrName)
	return c.run(ctx, cmds)
}

// run executes nsupdate with TSIG and pipes commands to stdin.
func (c *Client) run(ctx context.Context, commands string) error {
	tsigArg := fmt.Sprintf("%s:%s:%s", c.cfg.Algorithm, c.cfg.KeyName, c.cfg.Secret)
	cmd := exec.CommandContext(ctx, "nsupdate", "-y", tsigArg)
	cmd.Stdin = strings.NewReader(commands)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ddns: nsupdate failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	if len(out) > 0 && strings.Contains(string(out), "update failed") {
		return fmt.Errorf("ddns: nsupdate returned error: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// ZoneForTenant 计算 tenant 的正向 zone 名。
//
//	tenant slot 3 → "tenant3.owl."
func ZoneForTenant(slot uint16, owlDomain string) string {
	if !strings.HasSuffix(owlDomain, ".") {
		owlDomain += "."
	}
	return fmt.Sprintf("tenant%d.%s", slot, owlDomain)
}

// RegisterCardName 为 card 注册永久 DNS 名（AAAA + PTR）。
//
// 与 RegisterDevice 区别：
//   - card 用 spatial_prefix（如 /96 bed prefix），但 DNS AAAA 只能指向 host address
//     → 解析方式：取 prefix 的 network address（host 部分 = 0）作为 AAAA target
//   - card_name 是 bed-stable 永久名（不含 PHI），如 "u42-r03-b01.tenant1.owl"
//
// 参数：
//
//	prefix     - card 的 spatial_prefix（netip.Prefix），mask 必须 < 128
//	shortName  - DNS 短名（一段 label），如 "u42-r03-b01"
//	zone       - tenant zone，如 "tenant1.owl."
//
// 写入：
//   - forward: <shortName>.<zone> AAAA <network(prefix)>     in zone <zone>
//   - reverse: <network(prefix) ip6.arpa> PTR <shortName>.<zone>  in zone "0.0.0.0.0.0.d.f.ip6.arpa."
//
// Idempotency：caller 应在 INSERT card 前先 UnregisterCardName 兜底；或假设单次 INSERT only。
func (c *Client) RegisterCardName(ctx context.Context, prefix netip.Prefix, shortName, zone string) error {
	if !prefix.IsValid() || prefix.Bits() < 0 || prefix.Bits() > 128 {
		return fmt.Errorf("ddns: invalid prefix %s", prefix)
	}
	if !spatial.IsOwlAddr(prefix.Addr()) {
		return fmt.Errorf("ddns: prefix %s not in owl namespace fd00:0000::/32", prefix)
	}
	if shortName == "" || zone == "" {
		return errors.New("ddns: shortName + zone required")
	}
	if !strings.HasSuffix(zone, ".") {
		zone += "."
	}
	netAddr := prefix.Masked().Addr()
	fqdn := shortName + "." + zone
	ptrName := spatial.ReverseDNS(netAddr) + "."
	revZone := "0.0.0.0.0.0.d.f.ip6.arpa."

	cmds := fmt.Sprintf(`server %s %d
zone %s
update add %s 300 AAAA %s
send
zone %s
update add %s 300 PTR %s
send
`, c.cfg.Server, c.cfg.Port,
		zone, fqdn, netAddr.String(),
		revZone, ptrName, fqdn)

	return c.run(ctx, cmds)
}

// UnregisterCardName 删除 card 的 forward + reverse records。
func (c *Client) UnregisterCardName(ctx context.Context, prefix netip.Prefix, shortName, zone string) error {
	if !strings.HasSuffix(zone, ".") {
		zone += "."
	}
	netAddr := prefix.Masked().Addr()
	fqdn := shortName + "." + zone
	ptrName := spatial.ReverseDNS(netAddr) + "."
	revZone := "0.0.0.0.0.0.d.f.ip6.arpa."

	cmds := fmt.Sprintf(`server %s %d
zone %s
update delete %s AAAA
send
zone %s
update delete %s PTR
send
`, c.cfg.Server, c.cfg.Port,
		zone, fqdn,
		revZone, ptrName)
	return c.run(ctx, cmds)
}

// CardShortName 生成 card 的永久 DNS 短名（bed-stable，无 PHI）。
//
// 命名规则：
//
//	/48 tenant     → "t"               （整个 tenant zone 公共名）
//	/56 branch     → "br<branch>"      （branch=01..FF）
//	/64 site       → "site<bld><flr>"
//	/80 unit       → "u<unit>"
//	/88 room       → "u<unit>-r<room>"
//	/96 active_bed → "u<unit>-r<room>-b<bed>"  ⭐ 最常见
//	/128 device    → caller 用 RegisterDevice 走 device shortName
//
// 例: SlotsOf(fd00:0:1:112:42:301::) → tenant=1, branch=1, site=18, unit=42, room=3, bed=1
//
//	prefix /96 → "u42-r03-b01"
func CardShortName(prefix netip.Prefix) (string, error) {
	if !prefix.IsValid() {
		return "", fmt.Errorf("ddns: invalid prefix")
	}
	tenant, branch, _, unit, room, bed, _, err := spatial.SlotsOf(prefix.Addr())
	if err != nil {
		return "", err
	}
	switch prefix.Bits() {
	case 48:
		_ = tenant
		return "t", nil
	case 56:
		return fmt.Sprintf("br%02x", branch), nil
	case 64:
		// site = bld<<4 | floor (4+4 split)
		_, _ = branch, unit
		bld, flr := spatial.UnpackSiteSlot(uint8(prefix.Addr().As16()[7]))
		return fmt.Sprintf("site%xb%xf", bld, flr), nil
	case 80:
		return fmt.Sprintf("u%04x", unit), nil
	case 88:
		return fmt.Sprintf("u%04x-r%02x", unit, room), nil
	case 96:
		return fmt.Sprintf("u%04x-r%02x-b%02x", unit, room, bed), nil
	case 128:
		return "", fmt.Errorf("ddns: /128 device prefix should use RegisterDevice")
	}
	return "", fmt.Errorf("ddns: unsupported masklen /%d", prefix.Bits())
}
