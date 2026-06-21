package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var disinfTestDBSeq int

func newDisinfTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	disinfTestDBSeq++
	dsn := fmt.Sprintf("file:disinf_%d?mode=memory&cache=shared", disinfTestDBSeq)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(&models.DisinfectionCompliance{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec(`CREATE TABLE ["Auxiliary_EquipmentDisinfection"] (
		"Id" BIGINT PRIMARY KEY, "TenantId" INTEGER, "EquipmentId" INTEGER, "DisinfectUserId" INTEGER,
		"DisinfectWay" TEXT, "Disinfectant" TEXT, "StartTime" DATETIME, "EndTime" DATETIME,
		"TreatmentId" INTEGER, "Type" TEXT, "Status" INTEGER, "CreatorId" INTEGER,
		"CreateTime" DATETIME, "LastModifyTime" DATETIME)`).Error; err != nil {
		t.Fatalf("create base: %v", err)
	}
	return db
}

func TestDisinf_Record_WritesBaseAndCompliance(t *testing.T) {
	db := newDisinfTestDB(t)
	s := &DisinfectionService{db: db, tenantID: 3}
	now := time.Now()
	rec, err := s.Record(DisinfectRecordInput{
		DeviceID: 101, DisinfectType: models.DisinfectTypeTerminal, Disinfectant: "过氧乙酸",
		Concentration: "0.2%", OperatorID: 9, StartTime: now, EndTime: now.Add(30 * time.Minute),
		ResidualCheck: models.DisinfectResultPass, Result: models.DisinfectResultPass, Source: "manual",
	})
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	var baseCnt int64
	db.Table(`"Auxiliary_EquipmentDisinfection"`).Where(`"EquipmentId" = ? AND "Type" = ?`, 101, "terminal").Count(&baseCnt)
	if baseCnt != 1 {
		t.Fatalf("base 应 1 条, got %d", baseCnt)
	}
	var comp models.DisinfectionCompliance
	if err := db.First(&comp, "disinfection_id = ?", rec.DisinfectionID).Error; err != nil {
		t.Fatalf("compliance 未写: %v", err)
	}
	if comp.ResidualCheck != "pass" || comp.DeviceID != 101 {
		t.Fatalf("compliance 字段错: %+v", comp)
	}
	if _, err := s.Record(DisinfectRecordInput{DeviceID: 101, DisinfectType: "bogus", Source: "manual"}); err == nil {
		t.Fatalf("未知类型应拒")
	}
}

func TestDisinf_SaveCompliance_Upsert(t *testing.T) {
	db := newDisinfTestDB(t)
	s := &DisinfectionService{db: db, tenantID: 3}
	now := time.Now()
	rec, _ := s.Record(DisinfectRecordInput{DeviceID: 102, DisinfectType: models.DisinfectTypeHeat, OperatorID: 9, StartTime: now, EndTime: now, Source: "manual"})
	if _, err := s.SaveCompliance(rec.DisinfectionID, ComplianceInput{ResidualCheck: models.DisinfectResultFail, Result: models.DisinfectResultFail}); err != nil {
		t.Fatalf("save compliance: %v", err)
	}
	var comp models.DisinfectionCompliance
	db.First(&comp, "disinfection_id = ?", rec.DisinfectionID)
	if comp.ResidualCheck != "fail" {
		t.Fatalf("应更新为 fail, got %s", comp.ResidualCheck)
	}
	s.SaveCompliance(rec.DisinfectionID, ComplianceInput{ResidualCheck: models.DisinfectResultPass, Result: models.DisinfectResultPass})
	var cnt int64
	db.Model(&models.DisinfectionCompliance{}).Where("disinfection_id = ?", rec.DisinfectionID).Count(&cnt)
	if cnt != 1 {
		t.Fatalf("应仍 1 行(幂等), got %d", cnt)
	}
	if _, err := s.SaveCompliance(999999, ComplianceInput{ResidualCheck: "pass"}); err == nil {
		t.Fatalf("不存在记录应拒")
	}
}

func TestDisinf_MachineStatus(t *testing.T) {
	db := newDisinfTestDB(t)
	s := &DisinfectionService{db: db, tenantID: 3}
	now := time.Now()
	s.Record(DisinfectRecordInput{DeviceID: 201, DisinfectType: models.DisinfectTypeTerminal, OperatorID: 9, StartTime: now, EndTime: now, ResidualCheck: "pass", Result: "pass", Source: "manual"})
	s.Record(DisinfectRecordInput{DeviceID: 201, DisinfectType: models.DisinfectTypeDecalc, OperatorID: 9, StartTime: now, EndTime: now, ResidualCheck: "pass", Result: "pass", Source: "manual"})
	if st := s.MachineStatus(201); st.State != DisinfMachineOK {
		t.Fatalf("机201 应 OK, got %s (%v)", st.State, st.Reasons)
	}
	s.Record(DisinfectRecordInput{DeviceID: 202, DisinfectType: models.DisinfectTypeTerminal, OperatorID: 9, StartTime: now, EndTime: now, ResidualCheck: "fail", Result: "fail", Source: "manual"})
	if st := s.MachineStatus(202); st.State != DisinfMachineBlocked {
		t.Fatalf("机202 残留fail 应 BLOCKED, got %s", st.State)
	}
	s.Record(DisinfectRecordInput{DeviceID: 202, DisinfectType: models.DisinfectTypeTerminal, OperatorID: 9, StartTime: now.Add(time.Hour), EndTime: now.Add(time.Hour), ResidualCheck: "pass", Result: "pass", Source: "manual"})
	s.Record(DisinfectRecordInput{DeviceID: 202, DisinfectType: models.DisinfectTypeDecalc, OperatorID: 9, StartTime: now, EndTime: now, ResidualCheck: "pass", Result: "pass", Source: "manual"})
	if st := s.MachineStatus(202); st.State != DisinfMachineOK {
		t.Fatalf("机202 补合格后应解除 OK, got %s", st.State)
	}
	if st := s.MachineStatus(203); st.State != DisinfMachineWarn {
		t.Fatalf("机203 终末未做 应 WARN, got %s", st.State)
	}
	// 机 204：终末今日做了但除钙是8天前 → WARN(除钙到期)
	s.Record(DisinfectRecordInput{DeviceID: 204, DisinfectType: models.DisinfectTypeTerminal, OperatorID: 9, StartTime: now, EndTime: now, ResidualCheck: "pass", Result: "pass", Source: "manual"})
	s.Record(DisinfectRecordInput{DeviceID: 204, DisinfectType: models.DisinfectTypeDecalc, OperatorID: 9, StartTime: now.AddDate(0, 0, -8), EndTime: now.AddDate(0, 0, -8), ResidualCheck: "pass", Result: "pass", Source: "manual"})
	if st := s.MachineStatus(204); st.State != DisinfMachineWarn || !st.DecalcOverdue {
		t.Fatalf("机204 除钙8天前 应 WARN+DecalcOverdue, got %s overdue=%v", st.State, st.DecalcOverdue)
	}
}

func TestDisinf_AlertsAndStats(t *testing.T) {
	db := newDisinfTestDB(t)
	s := &DisinfectionService{db: db, tenantID: 3}
	now := time.Now()
	s.Record(DisinfectRecordInput{DeviceID: 301, DisinfectType: models.DisinfectTypeTerminal, OperatorID: 9, StartTime: now, EndTime: now, ResidualCheck: "fail", Result: "fail", Source: "manual"})
	a, err := s.Alerts([]int64{301, 302})
	if err != nil {
		t.Fatalf("alerts: %v", err)
	}
	if len(a.Blocked) != 1 || a.Blocked[0].DeviceID != 301 {
		t.Fatalf("停机卡应含 301, got %+v", a.Blocked)
	}
	foundWarn := false
	for _, w := range a.Warn {
		if w.DeviceID == 302 {
			foundWarn = true
		}
	}
	if !foundWarn {
		t.Fatalf("302 应在 Warn, got %+v", a.Warn)
	}
}
