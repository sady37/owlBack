package spatial

// Platform Agent Identity — IPv6 first-class 身份派生。
//
// 背景：owl 全栈 IPv6-native，但 platform agent（sensor/cardagg/cognitive/action/
// ai-health/iot/data/各 gateway）此前没有 first-class 网络身份，导致：
//   - alarm_events.producer 只能是 VARCHAR 字符串（无法 prefix-match / 跨 agent 聚合）
//   - parent_span / trace_id 没可信源（sensor envelope sequence_number 硬编码 "0"）
//
// 终态：platform agent 各拿一段 IPv6 reserved slot（详见 doc/platform_agent_addressing.md），
// envelope.producer 改 INET，trace 链可追。
//
// 本文件提供：
//
//   - PlatformAgentNamespace — uuid_v5 派生用的固定 namespace（owl 平台级常量）
//   - DerivePlatformUID(agentName, ipv6) — 确定性派生 UID
//   - 一组 agent slot 常量（fd00:0:fff1::/48 等）
//
// 派生规则：uid = uuid_v5(PlatformAgentNamespace, agent_name + ":" + ipv6_canonical)
// agent_name 用小写 service 名（"wisefido-sensor"）；ipv6_canonical = netip.Addr.String()。
//
// **重要**：UID 派生**一次**后写入 .env，启动只读不重算。改 IP 时同步更新 .env，
// 否则历史 trace 链就断了。

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/google/uuid"
)

// PlatformAgentNamespace 固定的 platform agent UID 派生 namespace。
//
// 生成方式（一次性）：
//
//	uuid.NewSHA1(uuid.NameSpaceDNS, []byte("owl.platform"))
//
// 值固化在常量里，便于跨语言/工具复算。
var PlatformAgentNamespace = uuid.MustParse("8d52a4c3-1d28-516a-86df-6e6b1b3f6e6e")

// Platform Agent Slot — 每个 agent 占用的 /48 段。
//
// 与 owl B2B 业务 tenant 段隔离（业务占 fd00:0:1::/48 ~ fd00:0:9::/48，
// reserved 段为后端 fd00:0:fff0::/48 ~ fd00:0:ffff::/48）。
//
// 详见 owlBack/doc/platform_agent_addressing.md §2.2。
const (
	SlotSensor   = "fd00:0:fff1::/48" // wisefido-sensor
	SlotCardagg  = "fd00:0:fff2::/48" // wisefido-cardagg（仅 audit/monitoring，永不当 producer）
	SlotData     = "fd00:0:fff3::/48" // wisefido-data
	SlotIoT      = "fd00:0:fff4::/48" // wisefido-iot
	SlotQinglan  = "fd00:0:fff5::/48" // wisefido-qinglan gateway
	SlotSleepace = "fd00:0:fff6::/48" // wisefido-sleepace gateway
	SlotAIHealth = "fd00:0:fff7::/48" // wisefido-ai-health
	// fd00:0:fff8 ~ fff9 预留 cognitive / action 未来
	// fd00:0:fffa ~ fffe 5 slot 备用
)

// PlatformAgentBoundary 平台 agent 段统一根 /44（fd00:0:fff0::/44 含 fff0~ffff）。
// 用于 sanity check：判定某 INET 是否在 platform agent 段。
const PlatformAgentBoundary = "fd00:0:fff0::/44"

// AgentSlot service 名称 → /48 slot 映射。
//
// 多 node 部署时，agent 内部用 host 段 1, 2, 3... 区分（如 fd00:0:fff1::1/128 = sensor node-1）。
var AgentSlot = map[string]string{
	"wisefido-sensor":    SlotSensor,
	"wisefido-cardagg":   SlotCardagg,
	"wisefido-data":      SlotData,
	"wisefido-iot":       SlotIoT,
	"wisefido-qinglan":   SlotQinglan,
	"wisefido-sleepace":  SlotSleepace,
	"wisefido-ai-health": SlotAIHealth,
}

// DerivePlatformUID 由 agent 名称 + IPv6 派生 UID（uuid_v5）。
//
//	agentName: 小写 service 名（"wisefido-sensor"）
//	ipv6:      agent 的 /128 IPv6 字符串（带 prefix 也接受）
//
// 返回的 UUID 是确定性的——同样输入永远得同样输出。生产中**只在工具命令行一次性派生**，
// 结果写入 .env 的 SENSOR_UID（或对应 agent）字段；启动时只读不重算，避免 IP 改了
// 历史 trace 断链。
//
// 不验证 ipv6 是否在 owl B2B 段；调用方自己保证（或用 IsPlatformAgentAddr 预检）。
func DerivePlatformUID(agentName, ipv6 string) (uuid.UUID, error) {
	agentName = strings.TrimSpace(strings.ToLower(agentName))
	if agentName == "" {
		return uuid.Nil, fmt.Errorf("agentName empty")
	}
	addr, err := parsePlatformAddr(ipv6)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse ipv6: %w", err)
	}
	key := agentName + ":" + addr.String()
	return uuid.NewSHA1(PlatformAgentNamespace, []byte(key)), nil
}

// IsPlatformAgentAddr 判定一个 INET 是否落在 platform agent 保留段（fd00:0:fff0::/44）。
//
// 用途：cardagg / 消费端区分 "agent 派生 alarm" 与 "device 直发 alarm"——
// IsPlatformAgentAddr(producer) == true 表示是 agent；false 且 producer==device_addr 表示设备直发。
func IsPlatformAgentAddr(s string) bool {
	addr, err := parsePlatformAddr(s)
	if err != nil {
		return false
	}
	boundary, err := netip.ParsePrefix(PlatformAgentBoundary)
	if err != nil {
		return false
	}
	return boundary.Contains(addr)
}

// ResolveAgentSlot 通过 agent 名称查 /48 slot CIDR；未注册时返回空字符串。
func ResolveAgentSlot(agentName string) string {
	return AgentSlot[strings.TrimSpace(strings.ToLower(agentName))]
}

// parsePlatformAddr 接受 "fd00:0:fff1::1" 或 "fd00:0:fff1::1/128" 形态；返回 /128 Addr。
func parsePlatformAddr(s string) (netip.Addr, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return netip.Addr{}, fmt.Errorf("empty")
	}
	if idx := strings.IndexByte(s, '/'); idx >= 0 {
		s = s[:idx]
	}
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Addr{}, err
	}
	if !addr.Is6() {
		return netip.Addr{}, fmt.Errorf("not IPv6: %s", s)
	}
	return addr, nil
}
