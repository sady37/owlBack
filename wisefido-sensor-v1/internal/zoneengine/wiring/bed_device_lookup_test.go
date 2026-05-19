package wiring

import (
	"testing"
)

// Pure helper tests — DB path 走集成测试（不在 unit 范围）。

func TestPreferredDeviceTypeForAlarm(t *testing.T) {
	tests := []struct {
		alarmType string
		want      string
	}{
		{"LeftBed", "Radar"},
		{"InBed", "Radar"},
		{"BedNightAbsence", "Radar"},
		{"Fall", "Radar"},
		{"SuspectedFall", "Radar"},
		{"SittingOnGround", "Radar"},
		{"SuspectedSittingOnGround", "Radar"},
		{"Stay", ""},          // bathroom 占用，不归因
		{"NightAbsence", ""},  // room 维度
		{"Offline", ""},       // 设备健康类不走这个 lookup
		{"", ""},
		{"UnknownType", ""},
	}
	for _, tt := range tests {
		t.Run(tt.alarmType, func(t *testing.T) {
			got := preferredDeviceTypeForAlarm(tt.alarmType)
			if got != tt.want {
				t.Errorf("preferredDeviceTypeForAlarm(%q) = %q, want %q",
					tt.alarmType, got, tt.want)
			}
		})
	}
}

func TestTruncatePrefix(t *testing.T) {
	tests := []struct {
		name    string
		cidr    string
		newMask int
		want    string
	}{
		{
			// /96 keeps groups 1-6 fully; /88 zeros low 8 bits of group 6 (bed slot)
			// group6 0301 → 0300 (keep room=03, zero bed slot)
			name:    "bed /96 → room /88 zeros bed slot",
			cidr:    "fd00:0:3:111:3:301::/96",
			newMask: 88,
			want:    "fd00:0:3:111:3:300::/88",
		},
		{
			// /80 keeps groups 1-5 fully, zeros group 6+; group5=3 stays
			name:    "bed /96 → unit /80 zeros room+bed",
			cidr:    "fd00:0:3:111:3:301::/96",
			newMask: 80,
			want:    "fd00:0:3:111:3::/80",
		},
		{
			name:    "same mask returns empty (no-op)",
			cidr:    "fd00:0:3:111:3:301::/96",
			newMask: 96,
			want:    "",
		},
		{
			name:    "wider mask returns empty (avoid expand)",
			cidr:    "fd00:0:3:111:3::/88",
			newMask: 96,
			want:    "",
		},
		{
			name:    "invalid input",
			cidr:    "not-a-cidr",
			newMask: 88,
			want:    "",
		},
		{
			name:    "empty input",
			cidr:    "",
			newMask: 88,
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncatePrefix(tt.cidr, tt.newMask)
			if got != tt.want {
				t.Errorf("truncatePrefix(%q, %d) = %q, want %q",
					tt.cidr, tt.newMask, got, tt.want)
			}
		})
	}
}

func TestBedDeviceLookup_NilDB(t *testing.T) {
	// nil db 不应 panic；FindPrimaryDevice 直接返回 zero Addr。
	l := NewBedDeviceLookup(nil, nil)

	cases := []struct {
		zoneID, alarmType string
	}{
		{"fd00:0:3:111:3:301::/96", "LeftBed"},
		{"fd00:0:3:111:3::/88", "Fall"},
		{"", "LeftBed"},
		{"fd00:0:3:111:3:301::/96", "Stay"},     // unsupported
		{"fd00:0:3:111:3:301::/96", "UnknownX"}, // unrecognized
	}
	for _, c := range cases {
		addr := l.FindPrimaryDevice(c.zoneID, c.alarmType)
		if addr.IsValid() {
			t.Errorf("nil db, FindPrimaryDevice(%q, %q) = %v, want zero",
				c.zoneID, c.alarmType, addr)
		}
	}
}
