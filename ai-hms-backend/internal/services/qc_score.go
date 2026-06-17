package services

const qcBaseScorePerPatient = 20

type QCInput struct {
	AvgSP, AvgDP, AvgHR *float64
	CTRPre, CTRPost     *float64
	URR, KtV            *float64
	TUFPercent          *float64
	Hb, Alb, Ca, P, PTH *float64
}

type QCPatientScore struct {
	Base     int            `json:"base"`
	Items    map[string]int `json:"items"`
	OnTarget map[string]bool `json:"onTarget"`
	Quality  int            `json:"quality"`
	Total    int            `json:"total"`
}

var qcItemMax = map[string]int{
	"bloodPressure": 10, "heartRate": 5, "CTR": 10, "dialysisAdequacy": 10,
	"fluidControl": 5, "anemia": 10, "nutrition": 10, "calcium": 5, "phosphorus": 5, "PTH": 5,
}

func scoreBloodPressure(sp, dp *float64) (int, bool) {
	if sp == nil || dp == nil {
		return 0, false
	}
	s, d := *sp, *dp
	switch {
	case s <= 150 && d <= 90:
		return 10, true
	case s <= 160 && d <= 95:
		return 5, false
	case s <= 180 && d <= 100:
		return 0, false
	default:
		return -5, false
	}
}

func scoreHeartRate(hr *float64) (int, bool) {
	if hr == nil {
		return 0, false
	}
	if *hr >= 60 && *hr <= 100 {
		return 5, true
	}
	return 0, false
}

func scoreCTR(pre, post *float64) (int, bool) {
	if pre == nil && post == nil {
		return 0, false
	}
	best, has := -999, false
	consider := func(v *float64, t10, t5, t0 float64) {
		if v == nil {
			return
		}
		has = true
		var sc int
		switch {
		case *v <= t10:
			sc = 10
		case *v <= t5:
			sc = 5
		case *v <= t0:
			sc = 0
		default:
			sc = -5
		}
		if sc > best {
			best = sc
		}
	}
	consider(pre, 52, 55, 58)
	consider(post, 49, 52, 55)
	if !has {
		return 0, false
	}
	return best, best == 10
}

func scoreAdequacy(urr, ktv *float64) (int, bool) {
	if urr == nil && ktv == nil {
		return 0, false
	}
	u, k := -1.0, -1.0
	if urr != nil {
		u = *urr
	}
	if ktv != nil {
		k = *ktv
	}
	switch {
	case u >= 70 || k >= 1.3:
		return 10, true
	case (u >= 60 && u < 70) || (k >= 1.2 && k < 1.3):
		return 5, false
	case (u >= 50 && u < 60) || (k >= 1.1 && k < 1.2):
		return 0, false
	default:
		return -5, false
	}
}

func scoreFluid(tufPct *float64) (int, bool) {
	if tufPct == nil {
		return 0, false
	}
	t := *tufPct
	switch {
	case t < 4:
		return 5, true
	case t < 5:
		return 2, false
	case t <= 6:
		return 0, false
	default:
		return -3, false
	}
}

func scoreAnemia(hb *float64) (int, bool) {
	if hb == nil {
		return 0, false
	}
	h := *hb
	switch {
	case h >= 110 && h <= 130:
		return 10, true
	case (h >= 100 && h < 110) || (h > 130 && h <= 140):
		return 5, false
	case (h >= 90 && h < 100) || (h > 140 && h <= 150):
		return 0, false
	default:
		return -5, false
	}
}

func scoreNutrition(alb *float64) (int, bool) {
	if alb == nil {
		return 0, false
	}
	a := *alb
	switch {
	case a >= 40 && a <= 55:
		return 10, true
	case a >= 35 && a < 40:
		return 5, false
	case a >= 30 && a < 35:
		return 0, false
	default:
		return -5, false
	}
}

func scoreCalcium(ca *float64) (int, bool) {
	if ca == nil {
		return 0, false
	}
	c := *ca
	switch {
	case c >= 2.1 && c <= 2.5:
		return 5, true
	case (c >= 1.8 && c < 2.1) || (c > 2.5 && c <= 2.8):
		return 0, false
	default:
		return -5, false
	}
}

func scorePhosphorus(p *float64) (int, bool) {
	if p == nil {
		return 0, false
	}
	v := *p
	switch {
	case v >= 1.13 && v <= 1.78:
		return 5, true
	case (v >= 0.8 && v < 1.13) || (v > 1.78 && v <= 2.2):
		return 0, false
	default:
		return -5, false
	}
}

func scorePTH(pth *float64) (int, bool) {
	if pth == nil {
		return 0, false
	}
	v := *pth
	switch {
	case v >= 150 && v <= 300:
		return 5, true
	case (v >= 100 && v < 150) || (v > 300 && v <= 600):
		return 2, false
	case (v >= 50 && v < 100) || (v > 600 && v <= 800):
		return 0, false
	default:
		return -3, false
	}
}

func ScorePatient(in QCInput) QCPatientScore {
	items := map[string]int{}
	onTarget := map[string]bool{}
	set := func(key string, sc int, ok bool) {
		items[key] = sc
		onTarget[key] = ok
	}
	sc, ok := scoreBloodPressure(in.AvgSP, in.AvgDP)
	set("bloodPressure", sc, ok)
	sc, ok = scoreHeartRate(in.AvgHR)
	set("heartRate", sc, ok)
	sc, ok = scoreCTR(in.CTRPre, in.CTRPost)
	set("CTR", sc, ok)
	sc, ok = scoreAdequacy(in.URR, in.KtV)
	set("dialysisAdequacy", sc, ok)
	sc, ok = scoreFluid(in.TUFPercent)
	set("fluidControl", sc, ok)
	sc, ok = scoreAnemia(in.Hb)
	set("anemia", sc, ok)
	sc, ok = scoreNutrition(in.Alb)
	set("nutrition", sc, ok)
	sc, ok = scoreCalcium(in.Ca)
	set("calcium", sc, ok)
	sc, ok = scorePhosphorus(in.P)
	set("phosphorus", sc, ok)
	sc, ok = scorePTH(in.PTH)
	set("PTH", sc, ok)

	quality := 0
	for _, v := range items {
		quality += v
	}
	return QCPatientScore{
		Base: qcBaseScorePerPatient, Items: items, OnTarget: onTarget,
		Quality: quality, Total: qcBaseScorePerPatient + quality,
	}
}

type QCDoctorScore struct {
	DoctorID      string             `json:"doctorId"`
	PatientCount  int                `json:"patientCount"`
	QuantityScore int                `json:"quantityScore"`
	QualityScore  int                `json:"qualityScore"`
	TotalScore    int                `json:"totalScore"`
	OnTargetRate  map[string]float64 `json:"onTargetRate"`
}

func AggregateDoctor(doctorID string, patients []QCPatientScore) QCDoctorScore {
	n := len(patients)
	d := QCDoctorScore{DoctorID: doctorID, PatientCount: n, OnTargetRate: map[string]float64{}}
	onTargetCount := map[string]int{}
	for _, p := range patients {
		d.QualityScore += p.Quality
		for key, ok := range p.OnTarget {
			if ok {
				onTargetCount[key]++
			}
		}
	}
	d.QuantityScore = n * qcBaseScorePerPatient
	d.TotalScore = d.QuantityScore + d.QualityScore
	if n > 0 {
		for key := range qcItemMax {
			d.OnTargetRate[key] = float64(onTargetCount[key]) / float64(n)
		}
	}
	return d
}
