package services

import (
	"testing"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/models"
)

// 验证内嵌量表规则可加载且分档自洽（不依赖数据库）。
func TestNursingScalesLoadAndScore(t *testing.T) {
	scales, err := config.LoadNursingScales()
	if err != nil {
		t.Fatalf("加载量表规则失败: %v", err)
	}
	byType := map[string]config.NursingScale{}
	for _, s := range scales {
		byType[s.ScaleType] = s
	}

	// Morse：高分=高危
	morse, ok := byType["morse"]
	if !ok || !morse.Enabled {
		t.Fatal("morse 量表缺失或未启用")
	}
	cases := []struct {
		score int
		level string
	}{
		{0, models.NursingRiskLow},
		{24, models.NursingRiskLow},
		{25, models.NursingRiskModerate},
		{44, models.NursingRiskModerate},
		{45, models.NursingRiskHigh},
		{125, models.NursingRiskHigh},
	}
	for _, c := range cases {
		band := morse.ScoreBand(c.score)
		if band == nil {
			t.Fatalf("morse 分数 %d 未命中任何分档", c.score)
		}
		if band.Level != c.level {
			t.Errorf("morse 分数 %d 期望 %s 实得 %s", c.score, c.level, band.Level)
		}
	}

	// Braden：低分=高危（方向相反，分档仍按区间命中）
	braden := byType["braden"]
	if b := braden.ScoreBand(9); b == nil || b.Level != models.NursingRiskHigh {
		t.Errorf("braden 9 分应为 high")
	}
	if b := braden.ScoreBand(20); b == nil || b.Level != models.NursingRiskNone {
		t.Errorf("braden 20 分应为 none")
	}
}

// 验证条目取值合法性校验。
func TestValidOptionValue(t *testing.T) {
	item := config.NursingScaleItem{
		Key: "history",
		Options: []config.NursingScaleOption{
			{Label: "无", Value: 0}, {Label: "有", Value: 25},
		},
	}
	if !validOptionValue(item, 25) {
		t.Error("25 应为合法取值")
	}
	if validOptionValue(item, 17) {
		t.Error("17 应为非法取值")
	}
}
