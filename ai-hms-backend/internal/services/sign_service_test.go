package services

import (
	"fmt"
	"testing"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestSignService(t *testing.T) *SignService {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	// 测试库允许 AutoMigrate（生产禁用，由部署阶段按 docs/sql/deploy_new_tables.sql 建表）
	if err := db.AutoMigrate(&models.SignRecord{}); err != nil {
		t.Fatalf("migrate sign_record failed: %v", err)
	}
	return &SignService{db: db}
}

func TestSignServiceSignAndList(t *testing.T) {
	s := newTestSignService(t)

	if _, err := s.Sign(LegacyTenantID, models.SignTargetPrescription, "1001", "9", "张医生"); err != nil {
		t.Fatalf("sign #1 failed: %v", err)
	}
	if _, err := s.Sign(LegacyTenantID, models.SignTargetPrescription, "1001", "9", "张医生"); err != nil {
		t.Fatalf("sign #2 failed: %v", err)
	}
	// 另一对象，不应混入
	if _, err := s.Sign(LegacyTenantID, models.SignTargetPlan, "2002", "8", "李医生"); err != nil {
		t.Fatalf("sign plan failed: %v", err)
	}

	rows, err := s.ListSigns(LegacyTenantID, models.SignTargetPrescription, "1001")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 prescription signs, got %d", len(rows))
	}
	for _, r := range rows {
		if r.TargetType != models.SignTargetPrescription || r.TargetID != "1001" {
			t.Fatalf("unexpected row target: %+v", r)
		}
		if r.SignerID != "9" || r.SignerName != "张医生" {
			t.Fatalf("unexpected signer: %+v", r)
		}
		if r.SignTime.IsZero() {
			t.Fatalf("sign time should be set")
		}
		if r.ID == "" {
			t.Fatalf("id should be generated")
		}
	}

	// 跨对象隔离
	planRows, _ := s.ListSigns(LegacyTenantID, models.SignTargetPlan, "2002")
	if len(planRows) != 1 {
		t.Fatalf("expected 1 plan sign, got %d", len(planRows))
	}
}

func TestSignServiceValidation(t *testing.T) {
	s := newTestSignService(t)

	if _, err := s.Sign(LegacyTenantID, "bogus", "1", "9", ""); err == nil {
		t.Fatalf("expected error for invalid target type")
	}
	if _, err := s.Sign(LegacyTenantID, models.SignTargetPrescription, "", "9", ""); err == nil {
		t.Fatalf("expected error for empty target id")
	}
	if _, err := s.Sign(LegacyTenantID, models.SignTargetPrescription, "1", "", ""); err == nil {
		t.Fatalf("expected error for empty signer id")
	}
}
