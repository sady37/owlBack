package roomengine

import "math"

// CellType 格子类型（来自 layout 配置，启动时加载）
type CellType uint8

const (
	CellOpen     CellType = 0 // 空地（可行走）
	CellDoor     CellType = 1 // 门 / 入口
	CellBed      CellType = 2 // 床
	CellChair    CellType = 3 // 椅子 / 沙发
	CellWall     CellType = 4 // 墙 / 不可达
	CellFOVEdge  CellType = 5 // 雷达 FOV 边缘（合法出生点）
	CellToilet   CellType = 6 // 马桶
	CellShower   CellType = 7 // 淋浴区
)

// Cell 网格中的一个像素（10cm × 10cm）。
// 静态属性从 layout 配置加载，动态属性从观测中自动生长。
type Cell struct {
	// ---- 静态属性（layout 配置）----
	Type     CellType // 格子类型
	Walkable bool     // 人能否到达此位置

	// ---- 动态属性（观测累计，带衰减）----
	RealDecay    float64 // 被确认 real 的 track 经过此格子的衰减累计次数
	GhostDecay   float64 // 被判定 ghost 的 track 在此格子出现的衰减累计次数
	FlowX        float64 // 经过此格子的 track 平均运动方向 X 分量
	FlowY        float64 // 经过此格子的 track 平均运动方向 Y 分量
	StillCount   float64 // track 在此格子静止的衰减累计时长（秒）
	PoseMismatch float64 // 历史上 pose 与运动学矛盾的衰减累计次数（金属干扰区高）
	LastUpdateMs int64   // 最后更新时间（毫秒）
}

// GhostRatio 返回此格子的 ghost 概率 [0,1]。real+ghost 均为 0 时返回 0。
func (c *Cell) GhostRatio() float64 {
	total := c.RealDecay + c.GhostDecay
	if total < 0.01 {
		return 0
	}
	return c.GhostDecay / total
}

// IsEntry 是否为合法入口（门或 FOV 边缘），新 track 从此出生不扣分。
func (c *Cell) IsEntry() bool {
	return c.Type == CellDoor || c.Type == CellFOVEdge
}

// IsRestZone 是否为合理静止区（床、椅子、马桶），track 在此不动不扣分。
func (c *Cell) IsRestZone() bool {
	return c.Type == CellBed || c.Type == CellChair || c.Type == CellToilet
}

// StillTimeoutSec 静止超时阈值（秒），超过则报警。
// 空地站立 8min，封闭空间（马桶/淋浴）15min，床上由 Sleepad 管不限。
func (c *Cell) StillTimeoutSec() int {
	switch c.Type {
	case CellToilet, CellShower:
		return 15 * 60 // 15 分钟
	case CellBed:
		return 0 // 不限，走 Sleepad 逻辑
	case CellChair:
		return 0 // 椅子暂不限
	default:
		return 8 * 60 // 8 分钟
	}
}

// Decay 时间衰减：所有动态属性按半衰期衰减。
// dtSec 为距上次衰减的秒数，halfLifeSec 为半衰期（推荐 86400 即 24 小时）。
func (c *Cell) Decay(dtSec, halfLifeSec float64) {
	if dtSec <= 0 || halfLifeSec <= 0 {
		return
	}
	factor := math.Pow(0.5, dtSec/halfLifeSec)
	c.RealDecay *= factor
	c.GhostDecay *= factor
	c.FlowX *= factor
	c.FlowY *= factor
	c.StillCount *= factor
	c.PoseMismatch *= factor
}
