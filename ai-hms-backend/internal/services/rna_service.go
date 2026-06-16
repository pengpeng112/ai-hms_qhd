package services

import (
	"fmt"
	"math"
)

// RNa 处方计算服务（钠清除比模型 v2）
//
// 移植自专利设计稿 model_v2.py / 透析处方助手_v1.html，与前端 RNaPrescriptionPanel 保持一致。
// 物理已按 α=0.92(对流占比) / W=0.93(血浆水分率) 修正——可移动钠 m=(α/W)·C≈0.989·C，
// 略低于血清钠（旧"透析液钠 = 血清钠 + 6"为反临床错误，已弃）。
//
// 核心公式（per-session 闭式）：
//   搬水 Na_target = V_UF·C_pre/W
//   本次清钠 M_target = RNa·Na_target = (C_pre·V_UF + δ·TBW)/W,  δ = C_pre − C_post
//   对流 conv = (α/W)·C̄·V_UF ;  弥散 diff = M_target − conv
//   C_d = (α/W)·C̄·(1 + Q/D) − M_target/(D·T),  C̄=(C_pre+C_post)/2, Q=V_UF/T
//   联动：RNa→δ=(RNa−1)·C_pre·V_UF/TBW→C_post ; C_post→δ=C_pre−C_post→RNa=1+δ·TBW/(C_pre·V_UF)

const (
	RNaW            = 0.93  // 血浆水分率
	RNaDefaultAlpha = 0.92  // Donnan 系数 = 对流占比；弥散占比 = 1−α
	RNaDefaultD     = 12.0  // 钠弥散度 (L/h)
	RNaDefaultT     = 4.0   // 透析时长 (h)
	RNaNaFloor      = 135.5 // 透后血清钠安全地板 (mmol/L)
	RNaCdMax        = 148.0 // 透析机透析液钠上限 (mmol/L)
	RNaDeltaCap     = 3.0   // 单次最大钠降幅 (mmol/L)
)

// RNaCalculateRequest 处方计算输入
type RNaCalculateRequest struct {
	CPre      float64  `json:"cPre"`      // 透前血清钠 (mmol/L)
	PreWeight float64  `json:"preWeight"` // 透前体重 (kg)
	DryWeight float64  `json:"dryWeight"` // 干体重 (kg)
	HeightCm  float64  `json:"heightCm"`  // 身高 (cm)
	AgeYears  float64  `json:"ageYears"`  // 年龄 (岁)
	IsMale    bool     `json:"isMale"`    // 性别
	VUF       *float64 `json:"vuf"`       // 超滤量 (L)；传入则覆盖 preWeight−dryWeight

	// 钠目标：二选一驱动（driver=rna 用 RNa；driver=cpost 用 CPost）
	Driver string  `json:"driver"` // "rna" | "cpost"
	RNa    float64 `json:"rNa"`    // 钠清除比（driver=rna）
	CPost  float64 `json:"cPost"`  // 目标透后血清钠（driver=cpost）

	// 高级参数（<=0 时取默认）
	Alpha float64 `json:"alpha"`
	D     float64 `json:"d"`
	T     float64 `json:"t"`
}

// RNaCalculateResult 处方计算结果
type RNaCalculateResult struct {
	TBWDry float64 `json:"tbwDry"` // Watson 估全身水 (L)
	VUF    float64 `json:"vuf"`    // 超滤量 (L)

	RNa   float64 `json:"rNa"`   // 钠清除比（结果）
	CPost float64 `json:"cPost"` // 透后血清钠（结果）
	Delta float64 `json:"delta"` // 降钠幅度 = C_pre − C_post

	NaTarget float64 `json:"naTarget"` // 搬水 = V_UF·C_pre/W (mmol)
	Deload   float64 `json:"deload"`   // 脱载 = M_target − 搬水 (mmol)
	MTarget  float64 `json:"mTarget"`  // 本次清钠 (mmol)
	Conv     float64 `json:"conv"`     // 对流 (mmol)
	Diff     float64 `json:"diff"`     // 弥散 (mmol)

	CD               float64 `json:"cd"`               // 透析液钠精确值 (mmol/L)
	CDRounded        float64 `json:"cdRounded"`        // 取近 0.5 实用值 (mmol/L)
	SafetyStatus     string  `json:"safetyStatus"`     // ok / floor_limited / machine_limit
	SafetyStatusDesc string  `json:"safetyStatusDesc"` // 中文说明
}

// CalculateRNaPrescription 执行 RNa 钠清除比处方计算（纯函数，无副作用）
func CalculateRNaPrescription(req RNaCalculateRequest) RNaCalculateResult {
	alpha := req.Alpha
	if alpha <= 0 {
		alpha = RNaDefaultAlpha
	}
	d := req.D
	if d <= 0 {
		d = RNaDefaultD
	}
	t := req.T
	if t <= 0 {
		t = RNaDefaultT
	}
	r := alpha / RNaW

	tbw := watsonTBW(req.DryWeight, req.HeightCm, req.AgeYears, req.IsMale)

	vuf := req.PreWeight - req.DryWeight
	if req.VUF != nil {
		vuf = *req.VUF
	}
	if vuf < 0 {
		vuf = 0
	}

	// 钠目标联动
	var rNa, cPost, delta float64
	if req.Driver == "cpost" {
		cPost = req.CPost
		delta = req.CPre - cPost
		if vuf > 0 {
			rNa = 1 + (delta*tbw)/(req.CPre*vuf)
		} else {
			rNa = 1
		}
	} else {
		rNa = req.RNa
		if rNa == 0 {
			rNa = 1.0
		}
		if vuf > 0 {
			delta = ((rNa - 1) * req.CPre * vuf) / tbw
		}
		cPost = req.CPre - delta
	}

	naTarget := (vuf * req.CPre) / RNaW // 搬水
	mTarget := rNa * naTarget           // 本次清钠
	deload := mTarget - naTarget        // 脱载
	cbar := (req.CPre + cPost) / 2
	q := vuf / t
	cd := r*cbar*(1+q/d) - mTarget/(d*t)
	conv := r * cbar * vuf // 对流
	diff := mTarget - conv // 弥散

	// 安全阀
	safetyStatus := "ok"
	safetyDesc := "安全"
	if cPost < RNaNaFloor {
		safetyStatus = "floor_limited"
		safetyDesc = fmt.Sprintf("透后血钠 %.1f 低于地板 %.1f，会致低钠，建议调高 RNa/透后或减脱载", cPost, RNaNaFloor)
	}
	if cd > RNaCdMax {
		safetyStatus = "machine_limit"
		safetyDesc = fmt.Sprintf("透析液钠 %.1f 超机器上限 %.0f，建议延长时间或降 RNa", cd, RNaCdMax)
	}

	return RNaCalculateResult{
		TBWDry:           tbw,
		VUF:              vuf,
		RNa:              rNa,
		CPost:            cPost,
		Delta:            delta,
		NaTarget:         naTarget,
		Deload:           deload,
		MTarget:          mTarget,
		Conv:             conv,
		Diff:             diff,
		CD:               cd,
		CDRounded:        math.Round(cd*2) / 2,
		SafetyStatus:     safetyStatus,
		SafetyStatusDesc: safetyDesc,
	}
}

// watsonTBW Watson 全身水公式（干体重下，最低 10L）
// 男: 2.447 − 0.09156×年龄 + 0.1074×身高(cm) + 0.3362×干体重(kg)
// 女: −2.097 + 0.1069×身高(cm) + 0.2466×干体重(kg)
func watsonTBW(dryWeight, heightCm, age float64, isMale bool) float64 {
	var t float64
	if isMale {
		t = 2.447 - 0.09156*age + 0.1074*heightCm + 0.3362*dryWeight
	} else {
		t = -2.097 + 0.1069*heightCm + 0.2466*dryWeight
	}
	return math.Max(10, t)
}
