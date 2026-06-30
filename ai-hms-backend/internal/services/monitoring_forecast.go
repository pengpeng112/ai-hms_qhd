package services

import "time"

// 趋势② 轻量预测：近窗最小二乘线性拟合 + 数据驱动钳位（可叠加临床上下限），
// 把实线序列向「计划下机 plannedEnd」投影为 kind=predicted 的虚线点。
// 诚实边界：线性短时外推，对非线性/突变（如即将 IDH）预测力有限——事件预判归 AI③。

type forecastPoint struct {
	T time.Time
	V float64
}

type forecastOpts struct {
	Step        time.Duration // 投影步长；默认 30min
	LookbackN   int           // 近窗最多取几点；默认 6
	LookbackDur time.Duration // 近窗时间跨度；默认 90min
	ClampLo     *float64      // 可选临床下限（nil=不限）
	ClampHi     *float64      // 可选临床上限
}

const (
	defaultForecastStep        = 30 * time.Minute
	defaultForecastLookbackN   = 6
	defaultForecastLookbackDur = 90 * time.Minute
	forecastSpanFloor          = 1e-6 // 零跨度地板，避免钳位卡死
)

// linearFit 近窗最小二乘拟合结果：V = a + b*(分钟，相对窗口首点 t0)；带数据驱动钳位边界。
type linearFit struct {
	a, b  float64
	t0    time.Time
	envLo float64
	envHi float64
}

// fitRecentLinear 取近窗（最近 LookbackDur 内 且 最多 LookbackN 点，至少 2 点）做最小二乘。
// 点数 <2 返回 ok=false。
func fitRecentLinear(actual []forecastPoint, opts forecastOpts) (linearFit, bool) {
	n := len(actual)
	if n < 2 {
		return linearFit{}, false
	}
	lookbackN := opts.LookbackN
	if lookbackN <= 0 {
		lookbackN = defaultForecastLookbackN
	}
	lookbackDur := opts.LookbackDur
	if lookbackDur <= 0 {
		lookbackDur = defaultForecastLookbackDur
	}
	last := actual[n-1]
	startIdx := n - 1
	for i := n - 1; i >= 0; i-- {
		if last.T.Sub(actual[i].T) > lookbackDur {
			break
		}
		if n-i > lookbackN {
			break
		}
		startIdx = i
	}
	if n-startIdx < 2 {
		startIdx = n - 2
	}
	win := actual[startIdx:]

	t0 := win[0].T
	var sx, sy, sxx, sxy float64
	wmin, wmax := win[0].V, win[0].V
	for _, p := range win {
		x := p.T.Sub(t0).Minutes()
		sx += x
		sy += p.V
		sxx += x * x
		sxy += x * p.V
		if p.V < wmin {
			wmin = p.V
		}
		if p.V > wmax {
			wmax = p.V
		}
	}
	m := float64(len(win))
	denom := m*sxx - sx*sx
	a, b := win[len(win)-1].V, 0.0
	if denom != 0 {
		b = (m*sxy - sx*sy) / denom
		a = (sy - b*sx) / m
	}
	span := wmax - wmin
	if span < forecastSpanFloor {
		span = forecastSpanFloor
	}
	return linearFit{a: a, b: b, t0: t0, envLo: wmin - 0.5*span, envHi: wmax + 0.5*span}, true
}

// eval 在时刻 t 求拟合值，先临床钳位（若给）再数据驱动钳位。
func (f linearFit) eval(t time.Time, opts forecastOpts) float64 {
	v := f.a + f.b*t.Sub(f.t0).Minutes()
	lo, hi := f.envLo, f.envHi
	if opts.ClampLo != nil && *opts.ClampLo > lo {
		lo = *opts.ClampLo
	}
	if opts.ClampHi != nil && *opts.ClampHi < hi {
		hi = *opts.ClampHi
	}
	if v < lo {
		v = lo
	}
	if v > hi {
		v = hi
	}
	return v
}

// forecastSeries 近窗最小二乘投影到 plannedEnd；首点桥接末 actual。
// n<2 / plannedEnd 不晚于末点 / 仅桥接点 → nil。
func forecastSeries(actual []forecastPoint, plannedEnd time.Time, opts forecastOpts) []forecastPoint {
	fit, ok := fitRecentLinear(actual, opts)
	if !ok {
		return nil
	}
	last := actual[len(actual)-1]
	if !plannedEnd.After(last.T) {
		return nil
	}
	step := opts.Step
	if step <= 0 {
		step = defaultForecastStep
	}
	out := []forecastPoint{{T: last.T, V: last.V}}
	for tcur := last.T.Add(step); !tcur.After(plannedEnd); tcur = tcur.Add(step) {
		out = append(out, forecastPoint{T: tcur, V: fit.eval(tcur, opts)})
	}
	if len(out) <= 1 {
		return nil
	}
	return out
}
