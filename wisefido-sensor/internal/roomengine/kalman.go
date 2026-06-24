package roomengine

import "math"

// KalmanFilter2D 二维匀速运动模型 Kalman 滤波器。
// 状态向量: [x, y, vx, vy]
// 观测向量: [x, y]
//
// 老人步速 0.3-1.0 m/s，dt=1s（1Hz 雷达帧率）。
// 过程噪声 Q 按老人步速标定，观测噪声 R 按雷达精度标定。
type KalmanFilter2D struct {
	// 状态 [x, y, vx, vy]
	X [4]float64

	// 协方差 4×4（对称，用完整矩阵简化实现）
	P [4][4]float64

	// 过程噪声加速度标准差（cm/s²）。老人加速度 ~30cm/s²。
	QAccel float64

	// 观测噪声标准差（cm）。雷达定位精度 ~20cm。
	RStd float64

	// 上一次更新的残差（innovation），用于打分。
	LastInnovX float64
	LastInnovY float64
	// 残差归一化值（Mahalanobis-like），越大越异常。
	LastResidual float64

	// 预测步数（无观测时空转的次数，连续无观测可判消失）
	MissCount int
}

// NewKalmanFilter2D 初始化：位置 (x0,y0)，速度为 0。
func NewKalmanFilter2D(x0, y0 float64) *KalmanFilter2D {
	kf := &KalmanFilter2D{
		QAccel: 30, // cm/s² — 老人正常加速度上限
		RStd:   20, // cm   — 雷达观测噪声
	}
	kf.X = [4]float64{x0, y0, 0, 0}
	// 初始协方差：位置不确定性大，速度更大
	kf.P[0][0] = 400 // 20cm² 位置
	kf.P[1][1] = 400
	kf.P[2][2] = 2500 // 50cm/s² 速度
	kf.P[3][3] = 2500
	return kf
}

// Predict 预测步（dt 秒）。无观测时也可调用（空转）。
// 状态转移: x' = x + vx*dt, y' = y + vy*dt, vx'=vx, vy'=vy
func (kf *KalmanFilter2D) Predict(dt float64) {
	if dt <= 0 {
		dt = 1
	}

	// 状态预测
	kf.X[0] += kf.X[2] * dt
	kf.X[1] += kf.X[3] * dt

	// 协方差预测: P' = F·P·F' + Q
	// F = [[1,0,dt,0],[0,1,0,dt],[0,0,1,0],[0,0,0,1]]
	// 简化计算（利用 F 的稀疏结构）
	var Pnew [4][4]float64

	// F·P
	var FP [4][4]float64
	for j := 0; j < 4; j++ {
		FP[0][j] = kf.P[0][j] + dt*kf.P[2][j]
		FP[1][j] = kf.P[1][j] + dt*kf.P[3][j]
		FP[2][j] = kf.P[2][j]
		FP[3][j] = kf.P[3][j]
	}
	// F·P·F'
	for i := 0; i < 4; i++ {
		Pnew[i][0] = FP[i][0] + dt*FP[i][2]
		Pnew[i][1] = FP[i][1] + dt*FP[i][3]
		Pnew[i][2] = FP[i][2]
		Pnew[i][3] = FP[i][3]
	}

	// 加过程噪声 Q（连续白噪声加速度模型）
	// Q = q² * [[dt⁴/4, 0, dt³/2, 0],
	//           [0, dt⁴/4, 0, dt³/2],
	//           [dt³/2, 0, dt², 0],
	//           [0, dt³/2, 0, dt²]]
	q := kf.QAccel
	dt2 := dt * dt
	dt3 := dt2 * dt
	dt4 := dt3 * dt
	q2 := q * q

	Pnew[0][0] += q2 * dt4 / 4
	Pnew[1][1] += q2 * dt4 / 4
	Pnew[0][2] += q2 * dt3 / 2
	Pnew[2][0] += q2 * dt3 / 2
	Pnew[1][3] += q2 * dt3 / 2
	Pnew[3][1] += q2 * dt3 / 2
	Pnew[2][2] += q2 * dt2
	Pnew[3][3] += q2 * dt2

	kf.P = Pnew
}

// Update 观测更新步。返回残差大小（cm）。
// zx, zy 为观测坐标（房间坐标 cm）。
func (kf *KalmanFilter2D) Update(zx, zy float64) float64 {
	kf.MissCount = 0

	// 残差 (innovation)
	innX := zx - kf.X[0]
	innY := zy - kf.X[1]
	kf.LastInnovX = innX
	kf.LastInnovY = innY

	// S = H·P·H' + R（H = [[1,0,0,0],[0,1,0,0]]）
	// S 是 2×2
	r2 := kf.RStd * kf.RStd
	s00 := kf.P[0][0] + r2
	s01 := kf.P[0][1]
	s10 := kf.P[1][0]
	s11 := kf.P[1][1] + r2

	// S 逆（2×2）
	det := s00*s11 - s01*s10
	if math.Abs(det) < 1e-10 {
		det = 1e-10
	}
	si00 := s11 / det
	si01 := -s01 / det
	si10 := -s10 / det
	si11 := s00 / det

	// K = P·H'·S⁻¹（4×2）
	// P·H' 的列: P[:,0] 和 P[:,1]
	var K [4][2]float64
	for i := 0; i < 4; i++ {
		K[i][0] = kf.P[i][0]*si00 + kf.P[i][1]*si10
		K[i][1] = kf.P[i][0]*si01 + kf.P[i][1]*si11
	}

	// 状态更新: X = X + K·inn
	for i := 0; i < 4; i++ {
		kf.X[i] += K[i][0]*innX + K[i][1]*innY
	}

	// 协方差更新: P = (I - K·H)·P
	var IKH [4][4]float64
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if i == j {
				IKH[i][j] = 1
			}
		}
		// 减去 K·H: K[:,0]*H[0,:] + K[:,1]*H[1,:]
		// H[0,:] = [1,0,0,0], H[1,:] = [0,1,0,0]
		IKH[i][0] -= K[i][0]
		IKH[i][1] -= K[i][1]
	}

	var Pnew [4][4]float64
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				Pnew[i][j] += IKH[i][k] * kf.P[k][j]
			}
		}
	}
	kf.P = Pnew

	// 残差大小
	kf.LastResidual = math.Sqrt(innX*innX + innY*innY)
	return kf.LastResidual
}

// PredictOnly 只预测不更新（track 本帧丢失时调用）。
func (kf *KalmanFilter2D) PredictOnly(dt float64) {
	kf.Predict(dt)
	kf.MissCount++
}

// Velocity 当前估计速度 (vx, vy) cm/s。
func (kf *KalmanFilter2D) Velocity() (vx, vy float64) {
	return kf.X[2], kf.X[3]
}

// Speed 当前估计速率 cm/s。
func (kf *KalmanFilter2D) Speed() float64 {
	return math.Sqrt(kf.X[2]*kf.X[2] + kf.X[3]*kf.X[3])
}

// Position 当前估计位置。
func (kf *KalmanFilter2D) Position() (x, y float64) {
	return kf.X[0], kf.X[1]
}
