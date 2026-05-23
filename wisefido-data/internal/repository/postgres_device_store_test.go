package repository

import "testing"

func TestDeriveUIDSuffix(t *testing.T) {
	cases := []struct {
		uid  string
		want string
	}{
		// 标准 12-hex MAC-style UID（radar）
		{"E598A2ACD523", "a2ac:d523"},
		{"9923003AB197", "003a:b197"},
		{"4D8710D5CABB", "10d5:cabb"},
		// Sleepace BM 前缀：M 是非 hex 要剥；B 是合法 hex 保留
		// "BM87224700978" → strip M → "B87224700978" (12 hex) → 末 8 hex = "24700978"
		{"BM87224700978", "2470:0978"},
		// "BM87225200672" → strip M → "B87225200672" → 末 8 hex = "25200672"
		{"BM87225200672", "2520:0672"},
		// 不足 8 hex：左侧补 0
		{"abc", "0000:0abc"},
		{"", "0000:0000"},
		// 大小写归一到小写
		{"ABCDEF12", "abcd:ef12"},
		// 含分隔符（": -"）剥后保留 hex
		{"AA:BB:CC:DD:EE:FF", "ccdd:eeff"},
	}
	for _, c := range cases {
		got := deriveUIDSuffix(c.uid)
		if got != c.want {
			t.Errorf("deriveUIDSuffix(%q) = %q, want %q", c.uid, got, c.want)
		}
	}
}
