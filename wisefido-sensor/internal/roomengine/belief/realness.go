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
// 输出 PReal=bR 连续后验(→ FE track confidence ×100；内部 N_r 软阈 0.5)，非二元 ghost/real 判定。
//
// 权重 = 形态锚(铁律 [[fall_data_is_artificial_test]]：非权威值，留 oracle)。
const (
	rcRealBase    = 0.02  // M→R baseline 恢复率/s（misjudged mirror 慢回 real 安全阀）
	rcRealFloor   = 0.5   // §9.3① 远门出生先验下限 bR₀(far-born 起点)；不为 0 = FN-safe(光凭出生地永不判死成 ghost)
	rcDoorScaleCm = 120.0 // §9.3① door_d 出生先验尺度：door_d→0 近门 bR₀→1；door_d≥此 bR₀→floor
	rcWAuto       = 1.5   // 自主独立移动 real 率权
	rcWWall       = 1.5   // 墙外反射 mirror 率权（× Coexist：ghost 仅 track==2）
	rcWSync       = 1.2   // 同步移动 mirror 率权（× Coexist：需配对）
)

// RealnessObs 一帧 realness 软证据(census 由 墙/门几何 + 同步度 + 配对译入)。door_d 已移出(出生先验,见 NewRealnessTrack)。
type RealnessObs struct {
	Displaced  bool    // 相对出生位自主位移(走动/跌倒)→ real(镜像不自主动)
	CoexistRho float64 // §9.3② 与配对 track 同步移动强度[0,1]
	LaterBorn  bool    // §9.1 成对中后到者(ghost 候选)：破同步对称——同步 ρ 双 track 共享，只归后到者
	WallMargin float64 // §9.2 墙外反射裕度[0,1](穿墙交点距/尺度；0=墙内非反射/非 track==2)
	Coexist    float64 // co-existence 配对存在[0,1](孤轨=0 → mEv=0 → ghost 仅 track==2)
	DtMs       int64   // 距上帧 ms(率→概率时间积分；单 tick dt 小→跳变概率≈0→§9.4 自然忽略)
}

// RealnessTrack realness 后验(2 态前向滤波)。bR+bM=1。
type RealnessTrack struct{ bR, bM float64 }

// NewRealnessTrack 起 realness，出生先验 bR₀ 由 door_d 设(§9.3① 出生一次性事实,非 per-frame 率)：
//
//	近门(door_d→0)→ bR₀=1(确实进门)；远门(door_d≥rcDoorScaleCm,常伴无 EnterRoom)→ bR₀→rcRealFloor
//	(可疑,可能凭空出现/反射)；door_d<0(无 enter 区不可测)→ bR₀=1 不 penalize。floor 不为 0 = FN-safe。
//	confidence(=PReal)不压 fire(eligible 恒真),故远门降 confidence 零 FN 风险(最坏 N_r 少算=偏激进)。
func NewRealnessTrack(birthDoorD float64) *RealnessTrack {
	bR := 1.0
	if birthDoorD >= 0 {
		near := (rcDoorScaleCm - birthDoorD) / rcDoorScaleCm
		if near < 0 {
			near = 0
		} else if near > 1 {
			near = 1
		}
		bR = rcRealFloor + (1-rcRealFloor)*near
	}
	return &RealnessTrack{bR: bR, bM: 1 - bR}
}

// Update 一帧两态跳变：mEv 开 R→M（mirror 证据）/ rEv 开 M→R（real 证据 + baseline 恢复）。
func (r *RealnessTrack) Update(o RealnessObs) {
	dt := float64(o.DtMs) / 1000.0
	if dt <= 0 {
		return // 出生帧/同帧无时间推进
	}
	// mirror 跳变率：墙外几何 + 同步 ρ(只归后到者破对称)，× co-existence(孤轨=0 → mEv=0)。
	// **ghost 计算仅 track==2 + 出生 5s 窗**(成本)：census 在窗内且 coexist>0 才算 reflectSep,孤轨/窗后不算;
	// 此处 ×Coexist 与之一致(孤轨 wall_margin 本就 0)。door_d 不在此(出生先验,NewRealnessTrack 一次性)。
	sync := 0.0
	if o.LaterBorn {
		sync = rcWSync * o.CoexistRho
	}
	mEv := o.Coexist * (rcWWall*o.WallMargin + sync)
	// real 跳变率：baseline 恢复 + 自主独立移动(ρ 低)。door_d 已移出 → 出生先验(NewRealnessTrack)。
	rEv := rcRealBase
	if o.Displaced {
		rEv += rcWAuto * (1 - o.CoexistRho)
	}
	// 转移（跳变过程，保归一）：Real 吸收当 mEv=0（latch）。
	pRM := 1 - math.Exp(-mEv*dt)
	pMR := 1 - math.Exp(-rEv*dt)
	r.bR, r.bM = r.bR*(1-pRM)+r.bM*pMR, r.bM*(1-pMR)+r.bR*pRM
}

// PReal 真人后验(→ N_r 计数 + FE confidence ×100)。
func (r *RealnessTrack) PReal() float64 { return r.bR }

// PMirror 镜像后验(有真人源 ghost；forensic + FE 观测)。
func (r *RealnessTrack) PMirror() float64 { return r.bM }

// ForceReal 孤轨永 Real override（§G六「1 track 永发」）：Coexist==0 即强制 Real，永不让孤轨卡 Mirror。
//
//	出生窗后判定冻结(省算力)，但此 override 实时常开——配对源离开、track 变孤轨时立即纠回 Real，
//	防冻结成 Mirror 的轨变孤轨后 N_r 算成 0(有真人却数 0 人)。
func (r *RealnessTrack) ForceReal() { r.bR, r.bM = 1, 0 }
