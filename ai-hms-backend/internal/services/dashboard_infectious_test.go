package services

import (
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestDashboard_InfectiousCounts(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:dash_inf?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db.AutoMigrate(&models.PatientInfectious{}, &models.SignRecord{})
	inf := &InfectiousService{db: db, tenantID: LegacyTenantID}
	inf.Screen(6001, ScreenInput{ScreenDate: time.Now(), Source: "manual", Items: []ScreenItem{{Item: "HBsAg", Result: models.InfItemPositive}}})
	ds := &DashboardService{db: db}
	pc, dc, err := ds.InfectiousAlertCounts()
	if err != nil {
		t.Fatalf("counts: %v", err)
	}
	if pc != 1 || dc < 0 {
		t.Fatalf("阳性未处置应 1, got positive=%d due=%d", pc, dc)
	}
}
