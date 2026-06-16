package belief

import "math"

// realness.go — realness 2 态前向滤波(DBN-Zone-Run §G；device-room-zone.md §9)：S^r∈{Real,Mirror}。
// 纯转移矩阵跳变过程 + co-existence 耦合涌现，零 gate(无离散状态机/无硬阈/无 if-then 判定)。
//
//	起 [Real=1,Mirror=0]；逐帧两态跳变：mEv 开 R→M 闸 / rEv 开 M→R 闸（率→概率 1-e^{-rate·dt}）。
//	latch(§9.5)   = mEv=0 时 R→M=0 → Real 吸收恒 1.0（孤轨/无 mirror 证据/cd2b 真人摔静止 → 永不离 Real）。
//	track==2(§G六)= mEv ∝ Coexist（孤轨 Coexist=0 → mEv=0 → 永 Real 涌现，同 [[dbn_cutover_state]]
//	                「孤立 ρ=0→P(Ghost)=0」结构涌现，非 len==2 硬 gate）。
//	墙外(§9.2)    = WallMargin 几何自判别(只镜像在墙外)。同步(§9.3②) = ρ 双 track 对称 → 只归后到者 LaterBorn
//	                破对称(§9.1 先到=real 锚)。近门(§9.3①)/自主移动 = rEv real 率(D 软斜坡，连续无阈)。
//	单 tick(§9.4) = 率 ×dt 积分 → 小 dt 瞬态跳变概率≈0，自然忽略。
//
// 输出 PReal=bR 连续后验(→ FE track confidence ×100；内部 N_r/pFallReal 软阈 0.5)，非二元 ghost/real 判定。
//
// 权重 = 形态锚(铁律 [[fall_data_is_artificial_test]]：非权威值，留 oracle)。
const (
	rcRealBase    = 0.02  // M→R baseline 恢复率/s（misjudged mirror 慢回 real 安全阀）
	rcDoorScaleCm = 120.0 // §9.3① D 软斜坡尺度(近门→real；老人走得慢、出生在门边 D 小)
	rcWDoor       = 1.2   // 近门 real 率权
	rcWAuto       = 1.5   // 自主独立移动 real 率权
	rcWWall       = 1.5   // 墙外反射 mirror 率权
	rcWSync       = 1.2   // 同步移动 mirror 率权
)

// RealnessObs 一帧 realness 软证据(census 由 logic_ID 身份 + 出生位 + 墙/门几何 + 同步度 + 配对译入)。
type RealnessObs struct {
	BirthDoorD float64 // §9.3① 出生地→最近门 cm；<0=无 enter 区(跳过近门率)
	Displaced  bool    // 相对出生位自主位移(走动/跌倒)
	CoexistRho float64 // §9.3② 与配对 track 同步移动强度[0,1]
	LaterBorn  bool    // §9.1 成对中后到者(ghost 候选)：破同步对称——同步 ρ 双 track 共享，只归后到者
	WallMargin float64 // §9.2 墙外反射裕度[0,1](穿墙交点距/尺度；0=墙内非反射；自判别不依赖 LaterBorn)
	Coexist    float64 // co-existence 配对存在[0,1](孤轨=0 → mEv=0 → track==2 涌现)
	DtMs       int64   // 距上帧 ms(率→概率时间积分；单 tick dt 小→跳变概率≈0→§9.4 自然忽略)
}

// RealnessTrack realness 后验(2 态前向滤波)。bR+bM=1。
type RealnessTrack struct{ bR, bM float64 }

// NewRealnessTrack 起纯 Real（PReal=1；无 mirror 证据则恒 1 = 默认 real + latch）。
func NewRealnessTrack() *RealnessTrack { return &RealnessTrack{bR: 1, bM: 0} }

// Update 一帧两态跳变：mEv 开 R→M（mirror 证据，co-existence 耦合）/ rEv 开 M→R（real 证据 + baseline）。
func (r *RealnessTrack) Update(o RealnessObs) {
	dt := float64(o.DtMs) / 1000.0
	if dt <= 0 {
		return // 出生帧/同帧无时间推进
	}
	// mirror 跳变率：墙外几何自判别 + 同步 ρ(只归后到者破对称)，× co-existence(孤轨=0 → 永 Real)。
	sync := 0.0
	if o.LaterBorn {
		sync = rcWSync * o.CoexistRho
	}
	mEv := o.Coexist * (rcWWall*o.WallMargin + sync)
	// real 跳变率：baseline 恢复 + 近门(D 软斜坡) + 自主独立移动(ρ 低)。
	rEv := rcRealBase
	if o.BirthDoorD >= 0 {
		if near := (rcDoorScaleCm - o.BirthDoorD) / rcDoorScaleCm; near > 0 {
			rEv += rcWDoor * near
		}
	}
	if o.Displaced {
		rEv += rcWAuto * (1 - o.CoexistRho)
	}
	// 转移（跳变过程，保归一）：Real 吸收当 mEv=0（latch）。
	pRM := 1 - math.Exp(-mEv*dt)
	pMR := 1 - math.Exp(-rEv*dt)
	r.bR, r.bM = r.bR*(1-pRM)+r.bM*pMR, r.bM*(1-pMR)+r.bR*pRM
}

// PReal 真人后验(= pFallReal 调制 SFallen 发射 + N_r 计数 + FE confidence ×100)。
func (r *RealnessTrack) PReal() float64 { return r.bR }

// PMirror 镜像后验(有真人源 ghost；forensic + FE 观测)。
func (r *RealnessTrack) PMirror() float64 { return r.bM }
