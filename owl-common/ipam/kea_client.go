package ipam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"time"
)

// KeaClient 调 kea-ctrl-agent REST API 注入 lease。
//
// 用途：作 owl IPAM 决策的 audit/mirror 层。
// PGBackend 分配 slot 成功后调 KeaClient.RecordPrefixLease 把这次分配写到 kea
// lease database。kea lease 不参与 owl 业务路径（不影响 alarm/event 流量），
// 但提供：
//   1. 标准 DHCP lease 数据库（合规）
//   2. 未来真接入 DHCPv6 客户端时零迁移成本
//   3. lease6-get-* 命令便于排查"这个 prefix 谁占了"
//
// 不在 prod 配置时 KeaClient 可设 nil；PGBackend 会跳过 kea 调用。
type KeaClient struct {
	baseURL  string
	username string
	password string
	httpc    *http.Client
}

// NewKeaClient 构造，baseURL 形如 "http://localhost:8000"。
func NewKeaClient(baseURL, username, password string) *KeaClient {
	return &KeaClient{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpc:    &http.Client{Timeout: 5 * time.Second},
	}
}

type keaCommand struct {
	Command   string         `json:"command"`
	Service   []string       `json:"service"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type keaResponse struct {
	Result    int            `json:"result"`
	Text      string         `json:"text,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// call 发一条 kea 命令，返回第一条响应（kea 总返回数组）。
func (k *KeaClient) call(ctx context.Context, cmd keaCommand) (*keaResponse, error) {
	body, _ := json.Marshal(cmd)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, k.baseURL+"/", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if k.username != "" {
		req.SetBasicAuth(k.username, k.password)
	}
	resp, err := k.httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kea call: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kea HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var arr []keaResponse
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, fmt.Errorf("kea decode: %w (raw: %s)", err, string(raw))
	}
	if len(arr) == 0 {
		return nil, fmt.Errorf("kea returned empty response array")
	}
	return &arr[0], nil
}

// RecordPrefixLease 注入一个 IA_PD lease（spatial prefix 分配审计）。
//
//   prefix     : 分配的 IPv6 prefix (e.g. fd00:0:3:6300::/56 for branch slot 99)
//   duid       : 业务 owner identifier (建议: tenant_slot+role+slot 编码，便于 lease6-get-by-duid)
//   iaid       : 32-bit identity association id (建议: slot 自身)
//   comment    : free-form 描述，进 lease 的 user-context.comment
//   validLft   : lease 有效期 (秒); 0 用默认 86400 (1天)
//
// kea lease 永不 expire 用 validLft 设大值即可（u32 max ≈ 136 年）。owl 设备
// 几乎不变动，建议设 365 day = 31536000。
func (k *KeaClient) RecordPrefixLease(ctx context.Context, prefix netip.Prefix, duid string, iaid uint32, comment string, validLft uint32) error {
	if validLft == 0 {
		validLft = 31536000 // 365 day
	}
	args := map[string]any{
		"subnet-id":  1, // owl B2B namespace subnet
		"type":       "IA_PD",
		"ip-address": prefix.Masked().Addr().String(),
		"prefix-len": prefix.Bits(),
		"duid":       duid,
		"iaid":       iaid,
		"valid-lft":  validLft,
	}
	if comment != "" {
		args["comment"] = comment
	}

	resp, err := k.call(ctx, keaCommand{
		Command:   "lease6-add",
		Service:   []string{"dhcp6"},
		Arguments: args,
	})
	if err != nil {
		return err
	}
	// result codes: 0=success / 1=error / 2=unsupported / 3=empty / 4=conflict
	// 4=conflict (lease already exists) 当幂等 OK；1+"already exists" 同
	if resp.Result == 4 || (resp.Result == 1 && containsAny(resp.Text, "already exists", "duplicate")) {
		return nil
	}
	if resp.Result != 0 {
		return fmt.Errorf("kea lease6-add: result=%d text=%q", resp.Result, resp.Text)
	}
	return nil
}

// RecordAddressLease 注入一个 IA_NA lease（device /128 单地址分配）。
func (k *KeaClient) RecordAddressLease(ctx context.Context, addr netip.Addr, duid string, iaid uint32, comment string, validLft uint32) error {
	if validLft == 0 {
		validLft = 31536000
	}
	args := map[string]any{
		"subnet-id":  1,
		"ip-address": addr.String(),
		"duid":       duid,
		"iaid":       iaid,
		"valid-lft":  validLft,
	}
	if comment != "" {
		args["comment"] = comment
	}
	resp, err := k.call(ctx, keaCommand{
		Command:   "lease6-add",
		Service:   []string{"dhcp6"},
		Arguments: args,
	})
	if err != nil {
		return err
	}
	if resp.Result == 4 || (resp.Result == 1 && containsAny(resp.Text, "already exists", "duplicate")) {
		return nil
	}
	if resp.Result != 0 {
		return fmt.Errorf("kea lease6-add: result=%d text=%q", resp.Result, resp.Text)
	}
	return nil
}

// VersionGet ping kea，验证 ctrl-agent 可达 + auth ok。
func (k *KeaClient) VersionGet(ctx context.Context) (string, error) {
	resp, err := k.call(ctx, keaCommand{
		Command: "version-get",
		Service: []string{"dhcp6"},
	})
	if err != nil {
		return "", err
	}
	if resp.Result != 0 {
		return "", fmt.Errorf("kea version-get failed: %s", resp.Text)
	}
	return resp.Text, nil
}

// =============================================================================
// DUID 编码 — 把 owl spatial prefix/addr 直接当 DUID identifier
// =============================================================================
//
// kea DUID 是 RFC 3315 定义的 client identifier，本质是 hex 字符串。
// owl 用 DUID-EN (RFC 8415, type=2)：
//   bytes 0-1 : 0x00 0x02 (DUID-EN type)
//   bytes 2-5 : Enterprise Number (IANA PEN; OWL 占位 = ASCII "olbr" = 0x6F6C6272)
//   bytes 6+  : Identifier (owl 直接用 spatial prefix 的字节 4-15)
//
// 这样 DUID 天然唯一（每条分配的 prefix 不同 = DUID 不同）+ 可读（hex dump
// 直接看出是 fd00:0:T:... 哪一段）。

const (
	owlDUIDType   = "00:02"
	owlEnterprise = "6f:6c:62:72" // ASCII "olbr"
)

// EncodeDUIDForPrefix 把 spatial prefix 的字节 4-15 编入 DUID identifier。
//
//	prefix fd00:0:3:6300::/56 → "00:02:6f:6c:62:72:00:03:63:00:00:00:00:00:00:00:00:00"
func EncodeDUIDForPrefix(prefix netip.Prefix) string {
	return encodeDUIDFromBytes(prefix.Masked().Addr().As16())
}

// EncodeDUIDForAddr 同上，从 /128 host address 编 DUID（device 用）。
func EncodeDUIDForAddr(addr netip.Addr) string {
	return encodeDUIDFromBytes(addr.As16())
}

func encodeDUIDFromBytes(b [16]byte) string {
	parts := make([]byte, 0, 64)
	parts = append(parts, []byte(owlDUIDType+":"+owlEnterprise)...)
	for i := 4; i < 16; i++ {
		parts = append(parts, ':')
		parts = append(parts, hex2(b[i])...)
	}
	return string(parts)
}

func hex2(b byte) []byte {
	const hexdigits = "0123456789abcdef"
	return []byte{hexdigits[b>>4], hexdigits[b&0x0f]}
}
