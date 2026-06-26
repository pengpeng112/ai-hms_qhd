package idh

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Scorer IDH 风险评分可插拔接口。
// 本期默认 StubScorer；真模型随后以 HTTPScorer 接 Python 微服务替换，调用方不变。
type Scorer interface {
	Score(ctx context.Context, in RiskInput) RiskResult
}

// StubScorer 占位实现：恒返回不可用。墙上不显 IDH 风险，链路保持完整、不报错。
type StubScorer struct{}

func (StubScorer) Score(_ context.Context, _ RiskInput) RiskResult {
	return RiskResult{Available: false}
}

// 概率 → high/medium/low 的默认切点（经验值，可调，后续可移入配置）。
const (
	idhHighCut   = 0.5
	idhMediumCut = 0.2
)

// LevelFromProbability 把模型概率映射为分级。
func LevelFromProbability(p float64) string {
	switch {
	case p >= idhHighCut:
		return "high"
	case p >= idhMediumCut:
		return "medium"
	default:
		return "low"
	}
}

const defaultTimeout = 5 * time.Second

// HTTPScorer 调用 Python「IDH 预警」FastAPI 微服务（POST /idh/score）。
// 已搭骨架、默认不启用；接入时填 BaseURL 并在调用方装载 30 时点特征窗口。
type HTTPScorer struct {
	baseURL string
	http    *http.Client
}

func NewHTTPScorer(cfg Config) *HTTPScorer {
	to := time.Duration(cfg.TimeoutSec) * time.Second
	if to <= 0 {
		to = defaultTimeout
	}
	return &HTTPScorer{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		http:    &http.Client{Timeout: to},
	}
}

type scoreResponse struct {
	Probability float64 `json:"probability"`
}

// Score 调微服务评分；任何失败/无窗口都降级为不可用，绝不阻断实时监控主流程。
func (h *HTTPScorer) Score(ctx context.Context, in RiskInput) RiskResult {
	if h.baseURL == "" || len(in.Window) == 0 {
		return RiskResult{Available: false}
	}
	body, err := json.Marshal(in)
	if err != nil {
		return RiskResult{Available: false}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.baseURL+"/idh/score", bytes.NewReader(body))
	if err != nil {
		return RiskResult{Available: false}
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.http.Do(req)
	if err != nil {
		return RiskResult{Available: false}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return RiskResult{Available: false}
	}
	var out scoreResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return RiskResult{Available: false}
	}
	return RiskResult{Available: true, Probability: out.Probability, Level: LevelFromProbability(out.Probability)}
}
