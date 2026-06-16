package v1

import (
	"sort"
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// IndicatorMappingHandler 临床指标编码核对接口
// 用真实 LIS/HDIS 字典自动对齐"概念 ↔ 真实码"，供信息科上线后一次性核对定版 itemCodeHints/indexCodeHints。
type IndicatorMappingHandler struct{}

func NewIndicatorMappingHandler() *IndicatorMappingHandler { return &IndicatorMappingHandler{} }

// RegisterIndicatorMappingRoutes 注册路由（在 main.go 用 protected 组调用）
func RegisterIndicatorMappingRoutes(r *gin.RouterGroup) {
	h := NewIndicatorMappingHandler()
	g := r.Group("/admin/indicator-mapping")
	g.GET("/reconcile", h.Reconcile)
}

// NameCode 一条真实字典项（按 名称+码 去重后的计数）
type NameCode struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	Count          int    `json:"count"`
	AgreesWithHint bool   `json:"agreesWithHint"` // 真实码是否已在候选 hints 中
}

// ConceptReconcile 单个概念的核对结果
type ConceptReconcile struct {
	ConceptID      string     `json:"conceptId"`
	ConceptNameZh  string     `json:"conceptNameZh"`
	Loinc          []string   `json:"loinc"`
	ItemCodeHints  []string   `json:"itemCodeHints"`
	IndexCodeHints []string   `json:"indexCodeHints"`
	MatchedLIS     []NameCode `json:"matchedLis"`  // lab_report_items 按名称命中的真实项
	MatchedHDIS    []NameCode `json:"matchedHdis"` // patient_key_indicators 按名称命中的真实项
	// resolved=命中且码与候选一致 / review=命中但码不在候选(需裁决) / no_data=暂无真实数据
	Status string `json:"status"`
}

// ReconcileResponse 核对总响应
type ReconcileResponse struct {
	Concepts      []ConceptReconcile `json:"concepts"`
	UnmatchedLIS  []NameCode         `json:"unmatchedLis"`  // 未匹配到任何概念的真实 LIS 项
	UnmatchedHDIS []NameCode         `json:"unmatchedHdis"` // 未匹配到任何概念的真实 HDIS 项
	Summary       map[string]int     `json:"summary"`
}

// Reconcile 核对：用真实字典对齐概念，输出建议表（GET /api/v1/admin/indicator-mapping/reconcile）
func (h *IndicatorMappingHandler) Reconcile(c *gin.Context) {
	concepts, err := config.LoadIndicatorConcepts()
	if err != nil {
		response.InternalError(c, "加载指标映射失败: "+err.Error())
		return
	}

	db := database.GetDB()
	type rawNC struct {
		Name string `gorm:"column:name"`
		Code string `gorm:"column:code"`
		Cnt  int    `gorm:"column:cnt"`
	}
	var lisRows, hdisRows []rawNC
	// 新库为 snake_case 列；查不到（表空/不可达）时返回空，不报错
	db.Raw(`SELECT item_name AS name, COALESCE(item_code,'') AS code, COUNT(*) AS cnt
	        FROM lab_report_items WHERE item_name <> ''
	        GROUP BY item_name, item_code`).Scan(&lisRows)
	db.Raw(`SELECT index_name AS name, COALESCE(index_code,'') AS code, COUNT(*) AS cnt
	        FROM patient_key_indicators WHERE index_name <> ''
	        GROUP BY index_name, index_code`).Scan(&hdisRows)

	matchConcept := func(name string, kws []string) bool {
		n := strings.ToUpper(name)
		for _, kw := range kws {
			if kw == "" {
				continue
			}
			if strings.Contains(n, strings.ToUpper(kw)) {
				return true
			}
		}
		return false
	}
	inHints := func(code string, hints []string) bool {
		u := strings.ToUpper(strings.TrimSpace(code))
		if u == "" {
			return false
		}
		for _, h := range hints {
			if strings.ToUpper(h) == u {
				return true
			}
		}
		return false
	}

	lisMatched := make([]bool, len(lisRows))
	hdisMatched := make([]bool, len(hdisRows))

	out := make([]ConceptReconcile, 0, len(concepts))
	for _, cpt := range concepts {
		cr := ConceptReconcile{
			ConceptID: cpt.ConceptID, ConceptNameZh: cpt.ConceptNameZh,
			Loinc: cpt.Loinc, ItemCodeHints: cpt.ItemCodeHints, IndexCodeHints: cpt.IndexCodeHints,
		}
		for i, r := range lisRows {
			if matchConcept(r.Name, cpt.NameKeywords) {
				lisMatched[i] = true
				cr.MatchedLIS = append(cr.MatchedLIS, NameCode{
					Name: r.Name, Code: r.Code, Count: r.Cnt, AgreesWithHint: inHints(r.Code, cpt.ItemCodeHints),
				})
			}
		}
		for i, r := range hdisRows {
			if matchConcept(r.Name, cpt.NameKeywords) {
				hdisMatched[i] = true
				cr.MatchedHDIS = append(cr.MatchedHDIS, NameCode{
					Name: r.Name, Code: r.Code, Count: r.Cnt, AgreesWithHint: inHints(r.Code, cpt.IndexCodeHints),
				})
			}
		}
		switch {
		case len(cr.MatchedLIS)+len(cr.MatchedHDIS) == 0:
			cr.Status = "no_data"
		default:
			allAgree := true
			for _, m := range cr.MatchedLIS {
				if !m.AgreesWithHint {
					allAgree = false
				}
			}
			for _, m := range cr.MatchedHDIS {
				if !m.AgreesWithHint {
					allAgree = false
				}
			}
			if allAgree {
				cr.Status = "resolved"
			} else {
				cr.Status = "review"
			}
		}
		out = append(out, cr)
	}

	var unLis, unHdis []NameCode
	for i, r := range lisRows {
		if !lisMatched[i] {
			unLis = append(unLis, NameCode{Name: r.Name, Code: r.Code, Count: r.Cnt})
		}
	}
	for i, r := range hdisRows {
		if !hdisMatched[i] {
			unHdis = append(unHdis, NameCode{Name: r.Name, Code: r.Code, Count: r.Cnt})
		}
	}
	sort.Slice(unLis, func(i, j int) bool { return unLis[i].Count > unLis[j].Count })
	sort.Slice(unHdis, func(i, j int) bool { return unHdis[i].Count > unHdis[j].Count })

	summary := map[string]int{"lisDistinct": len(lisRows), "hdisDistinct": len(hdisRows)}
	for _, cr := range out {
		summary[cr.Status]++
	}

	response.Success(c, ReconcileResponse{
		Concepts: out, UnmatchedLIS: unLis, UnmatchedHDIS: unHdis, Summary: summary,
	})
}
