package spatial

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestDerivePlatformUID_Deterministic(t *testing.T) {
	// 同样输入两次必须得同样 UUID（确定性）。
	uid1, err := DerivePlatformUID("wisefido-sensor", "fd00:0:fff1::1")
	if err != nil {
		t.Fatalf("derive 1: %v", err)
	}
	uid2, err := DerivePlatformUID("wisefido-sensor", "fd00:0:fff1::1")
	if err != nil {
		t.Fatalf("derive 2: %v", err)
	}
	if uid1 != uid2 {
		t.Errorf("non-deterministic: %s != %s", uid1, uid2)
	}
	if uid1 == uuid.Nil {
		t.Errorf("nil UUID")
	}
}

func TestDerivePlatformUID_DifferentInputsDifferentUIDs(t *testing.T) {
	cases := []struct {
		name  string
		agent string
		ipv6  string
	}{
		{"sensor n1", "wisefido-sensor", "fd00:0:fff1::1"},
		{"sensor n2", "wisefido-sensor", "fd00:0:fff1::2"},
		{"cardagg n1", "wisefido-cardagg", "fd00:0:fff2::1"},
		{"data n1", "wisefido-data", "fd00:0:fff3::1"},
	}
	seen := map[uuid.UUID]string{}
	for _, c := range cases {
		uid, err := DerivePlatformUID(c.agent, c.ipv6)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if prev, exists := seen[uid]; exists {
			t.Errorf("collision: %s and %s both → %s", c.name, prev, uid)
		}
		seen[uid] = c.name
	}
}

func TestDerivePlatformUID_NormalizesInputs(t *testing.T) {
	// agent name: 大小写 + 空白 等价
	a, _ := DerivePlatformUID("wisefido-sensor", "fd00:0:fff1::1")
	b, _ := DerivePlatformUID("  Wisefido-Sensor  ", "fd00:0:fff1::1")
	if a != b {
		t.Errorf("case/whitespace not normalized: %s != %s", a, b)
	}
	// ipv6: 带 prefix 和不带等价
	c, _ := DerivePlatformUID("wisefido-sensor", "fd00:0:fff1::1")
	d, _ := DerivePlatformUID("wisefido-sensor", "fd00:0:fff1::1/128")
	if c != d {
		t.Errorf("/128 suffix not normalized: %s != %s", c, d)
	}
	// ipv6: 显式零段压缩等价
	e, _ := DerivePlatformUID("wisefido-sensor", "fd00:0:fff1:0:0:0:0:1")
	f, _ := DerivePlatformUID("wisefido-sensor", "fd00:0:fff1::1")
	if e != f {
		t.Errorf("zero-compress not normalized: %s != %s", e, f)
	}
}

func TestDerivePlatformUID_Errors(t *testing.T) {
	tests := []struct {
		name      string
		agent     string
		ipv6      string
		wantErr   bool
		errSubstr string
	}{
		{"empty agent", "", "fd00:0:fff1::1", true, "agentName"},
		{"whitespace agent", "  ", "fd00:0:fff1::1", true, "agentName"},
		{"empty ipv6", "wisefido-sensor", "", true, "empty"},
		{"invalid ipv6", "wisefido-sensor", "not-an-ip", true, "parse"},
		{"ipv4 rejected", "wisefido-sensor", "192.168.1.1", true, "IPv6"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DerivePlatformUID(tt.agent, tt.ipv6)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, want err = %v", err, tt.wantErr)
			}
			if err != nil && tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("err %q missing substr %q", err.Error(), tt.errSubstr)
			}
		})
	}
}

func TestIsPlatformAgentAddr(t *testing.T) {
	cases := []struct {
		name string
		addr string
		want bool
	}{
		{"sensor /128", "fd00:0:fff1::1", true},
		{"sensor with prefix", "fd00:0:fff1::1/128", true},
		{"cardagg /128", "fd00:0:fff2::1", true},
		{"ai-health /128", "fd00:0:fff7::1", true},
		{"boundary fff0", "fd00:0:fff0::1", true},  // 在 /44 内（虽然 fff0 是保留 slot）
		{"boundary ffff", "fd00:0:ffff::1", true},  // 在 /44 内
		{"business tenant 3", "fd00:0:3:111:3:201::abc", false},
		{"business tenant 1", "fd00:0:1::1", false},
		{"outside owl", "fe80::1", false},
		{"empty", "", false},
		{"invalid", "garbage", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := IsPlatformAgentAddr(c.addr)
			if got != c.want {
				t.Errorf("IsPlatformAgentAddr(%q) = %v, want %v", c.addr, got, c.want)
			}
		})
	}
}

func TestResolveAgentSlot(t *testing.T) {
	tests := []struct {
		agent string
		want  string
	}{
		{"wisefido-sensor", SlotSensor},
		{"wisefido-cardagg", SlotCardagg},
		{"  Wisefido-Sensor  ", SlotSensor}, // case + whitespace
		{"unknown-agent", ""},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.agent, func(t *testing.T) {
			got := ResolveAgentSlot(tt.agent)
			if got != tt.want {
				t.Errorf("ResolveAgentSlot(%q) = %q, want %q", tt.agent, got, tt.want)
			}
		})
	}
}

// 固化已部署 agent 的 UID snapshot：若有人改了 PlatformAgentNamespace 或派生算法，
// 测试会直接挂掉，避免 .env 里 pin 的 UID 跟代码不一致导致历史 trace 断链。
// 新增 snapshot：跑一次测试拿 t.Logf 值回填即可。
func TestDerivePlatformUID_Snapshots(t *testing.T) {
	cases := []struct {
		agent  string
		ipv6   string
		expect string
	}{
		{"wisefido-sensor", "fd00:0:fff1::1", "a893ff12-7d72-58dc-a26c-ce724e6007da"},
		{"wisefido-cardagg", "fd00:0:fff2::1", "f95986fd-5af9-55a6-a07d-9929b43d3897"},
	}
	for _, c := range cases {
		got, err := DerivePlatformUID(c.agent, c.ipv6)
		if err != nil {
			t.Fatalf("%s/%s: %v", c.agent, c.ipv6, err)
		}
		if got.String() != c.expect {
			t.Errorf("%s/%s: got %s, want %s — namespace/derivation changed?",
				c.agent, c.ipv6, got, c.expect)
		}
	}
}
