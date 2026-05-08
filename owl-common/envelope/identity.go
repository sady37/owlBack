package envelope

import "github.com/google/uuid"

// ServiceIdentity 表达一个 producer service 的三层身份（per doc §3.5）：
//
//	软件类型+版本 → Source URN
//	进程实例     → NodeID
//	单条消息     → ID (ULID, per Datagram)
//
// 每个 service 启动时一次性确定 Source 与 NodeID；之后所有 emit 共享。
type ServiceIdentity struct {
	// Source URN: scheme://name/version
	//
	// 推荐 scheme 表达信任层级（per doc §8.1 trust hierarchy）：
	//   "device-gateway://qinglan/v2.0.0"          ← firmware adapter
	//   "sensor://owl.engine.lost_fall/v1.0.0"     ← AI 推断引擎
	//   "cognitive://room.engine.fall_verifier/v1.0.0"  ← 认知层
	//   "owlfront://user-action/v1.0.0"            ← 前端用户操作
	Source string

	// NodeID 进程实例 UUID。每次启动新生成，区分版本/重启实例。
	NodeID string

	// Version 软件版本（也已编码进 Source URN 末段；保留作显式查询）。
	Version string
}

// NewServiceIdentity 创建 producer 身份。NodeID 自动 mint。
func NewServiceIdentity(source, version string) ServiceIdentity {
	return ServiceIdentity{
		Source:  source,
		NodeID:  uuid.NewString(),
		Version: version,
	}
}
