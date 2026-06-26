package services

import "math"

// 实时钠清除比（RNa）完成情况。
//
// 在 RNa 处方模型（rna_service.go）之上，用实时进度算"本次已清钠 / 目标清钠"：
//   对流 conv(t) = r·C̄·UF_actual(t)              —— UF_actual 实测（DMLog 累计超滤）
//   弥散 diff(t) = ∫₀ᵗ D·(r·C̄ − C_d(τ)) dτ        —— C_d 实测（电导率×9.9）沿时间积分
//                = D·t_elapsed·(r·C̄ − mean_Cd)    —— C_d 分段恒定时等于「均值×时长」
//   完成率 = (conv + diff) / M_target
// 其中 r=α/W、C̄=(C_pre+C_post)/2、M_target、D、T 由 CalculateRNaPrescription 按存输入重算。
//
// 弥散通量 D·(r·C̄−C_d) 由模型反解：会话总弥散 diff=D·T·(r·C̄−C_d)，与
// C_d=r·C̄·(1+Q/D)−M_target/(D·T) 自洽（diff=M_target−conv）。

// RNaCompletion 实时 RNa 完成情况（随 live-data 刷新）。输入不全时 Available=false。
type RNaCompletion struct {
	Available bool    `json:"available"`
	Percent   float64 `json:"percent"`   // 完成率 %
	TargetRNa float64 `json:"targetRNa"` // 处方钠清除比
	MTarget   float64 `json:"mTarget"`   // 目标清钠 (mmol)
	MRealized float64 `json:"mRealized"` // 已清钠 (mmol)
	CPre      float64 `json:"cPre"`      // 透前血清钠 (mmol/L)
	CPreAt    string  `json:"cPreAt"`    // C_pre 化验时间（可能非当日）
}

// RNaCompletionInput 实时完成率计算输入。
type RNaCompletionInput struct {
	Presc    RNaCalculateResult // CalculateRNaPrescription 按存输入重算的结果
	CPre     float64            // 透前血清钠 (mmol/L)
	Alpha    float64            // 对流占比（<=0 取默认）
	D        float64            // 钠弥散度 L/h（<=0 取默认）
	UFActual float64            // 实时累计超滤 (L)
	MeanCd   float64            // 整场实测透析液钠均值 (mmol/L)，= 电导率均值×9.9
	ElapsedH float64            // 已透时长 (h)
}

// computeRNaCompletion 纯函数：由处方重算结果 + 实时进度算 RNa 完成率。
func computeRNaCompletion(in RNaCompletionInput) RNaCompletion {
	alpha := in.Alpha
	if alpha <= 0 {
		alpha = RNaDefaultAlpha
	}
	d := in.D
	if d <= 0 {
		d = RNaDefaultD
	}
	r := alpha / RNaW
	cbar := (in.CPre + in.Presc.CPost) / 2

	conv := r * cbar * in.UFActual                 // 对流（实测 UF）
	diff := d * in.ElapsedH * (r*cbar - in.MeanCd) // 弥散（实测 C_d 积分；C_d>r·C̄ 时为负=净加钠）
	mReal := conv + diff

	pct := 0.0
	if in.Presc.MTarget != 0 {
		pct = mReal / in.Presc.MTarget * 100
	}
	if pct < 0 {
		pct = 0
	}
	return RNaCompletion{
		Available: true,
		Percent:   math.Round(pct*10) / 10,
		TargetRNa: in.Presc.RNa,
		MTarget:   in.Presc.MTarget,
		MRealized: math.Round(mReal*10) / 10,
		CPre:      in.CPre,
	}
}
