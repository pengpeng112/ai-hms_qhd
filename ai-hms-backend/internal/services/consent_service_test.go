package services

import (
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/models"
)

// 验证同意书模板可加载且核心类型齐全（不依赖数据库）。
func TestConsentTemplatesLoad(t *testing.T) {
	tpls, err := config.LoadConsentTemplates()
	if err != nil {
		t.Fatalf("加载同意书模板失败: %v", err)
	}
	want := []string{"dialysis", "cvc", "avf", "transfusion", "drug", "self_pay"}
	have := map[string]bool{}
	for _, tpl := range tpls {
		have[tpl.ConsentType] = true
	}
	for _, w := range want {
		if !have[w] {
			t.Errorf("缺少同意书类型: %s", w)
		}
	}
}

// 验证过期视图翻转逻辑：已签且 expires_at 已过 → expired。
func TestApplyConsentExpiry(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	future := time.Now().Add(24 * time.Hour)
	rows := []models.ConsentRecord{
		{Status: models.ConsentStatusSigned, ExpiresAt: &past},   // 应翻 expired
		{Status: models.ConsentStatusSigned, ExpiresAt: &future}, // 仍 signed
		{Status: models.ConsentStatusSigned, ExpiresAt: nil},     // 长期有效，仍 signed
		{Status: models.ConsentStatusPending, ExpiresAt: &past},  // pending 不动
	}
	applyConsentExpiry(rows)
	if rows[0].Status != models.ConsentStatusExpired {
		t.Errorf("已过期已签应翻 expired，实得 %s", rows[0].Status)
	}
	if rows[1].Status != models.ConsentStatusSigned {
		t.Errorf("未到期应仍 signed，实得 %s", rows[1].Status)
	}
	if rows[2].Status != models.ConsentStatusSigned {
		t.Errorf("无到期应仍 signed，实得 %s", rows[2].Status)
	}
	if rows[3].Status != models.ConsentStatusPending {
		t.Errorf("pending 不应被改，实得 %s", rows[3].Status)
	}
}
