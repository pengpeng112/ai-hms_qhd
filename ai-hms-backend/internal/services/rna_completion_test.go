package services

import "testing"

// 用处方重算结果驱动：会话末（UF 满、已透满、实测 C_d=处方 C_d）完成率应≈100%。
func TestComputeRNaCompletion_FullSession(t *testing.T) {
	vuf := 3.0
	presc := CalculateRNaPrescription(RNaCalculateRequest{
		CPre: 140, DryWeight: 60, HeightCm: 170, AgeYears: 60, IsMale: true,
		VUF: &vuf, Driver: "rna", RNa: 1.0,
	})
	// 实测 C_d 取处方算出的精确 C_d、UF 与时长取满。
	got := computeRNaCompletion(RNaCompletionInput{
		Presc: presc, CPre: 140, UFActual: vuf, MeanCd: presc.CD, ElapsedH: RNaDefaultT,
	})
	if !got.Available {
		t.Fatal("should be available")
	}
	if got.Percent < 99 || got.Percent > 101 {
		t.Fatalf("full session percent = %.2f, want ~100 (mReal=%.1f mTarget=%.1f)", got.Percent, got.MRealized, got.MTarget)
	}
}

// 半程（UF 半、已透半）完成率应在 40–60% 区间（对流随 UF、弥散随时间，均约半）。
func TestComputeRNaCompletion_HalfSession(t *testing.T) {
	vuf := 3.0
	presc := CalculateRNaPrescription(RNaCalculateRequest{
		CPre: 140, DryWeight: 60, HeightCm: 170, AgeYears: 60, IsMale: true,
		VUF: &vuf, Driver: "rna", RNa: 1.0,
	})
	got := computeRNaCompletion(RNaCompletionInput{
		Presc: presc, CPre: 140, UFActual: vuf / 2, MeanCd: presc.CD, ElapsedH: RNaDefaultT / 2,
	})
	if got.Percent < 40 || got.Percent > 60 {
		t.Fatalf("half session percent = %.2f, want ~50", got.Percent)
	}
}

// 目标为 0（无处方）→ 完成率 0，不 panic。
func TestComputeRNaCompletion_ZeroTarget(t *testing.T) {
	got := computeRNaCompletion(RNaCompletionInput{Presc: RNaCalculateResult{MTarget: 0, CPost: 140}, CPre: 140, UFActual: 1, MeanCd: 138, ElapsedH: 2})
	if got.Percent != 0 {
		t.Fatalf("zero target percent = %.2f, want 0", got.Percent)
	}
}
