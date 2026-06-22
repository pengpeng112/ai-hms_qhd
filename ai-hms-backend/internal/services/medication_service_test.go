package services

import (
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newMaTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(&models.MedicationAdmin{}, &models.SignRecord{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestMed_RecordAdmin(t *testing.T) {
	db := newMaTestDB(t)
	s := &MedicationService{db: db, tenantID: 3}

	ma, err := s.RecordAdmin(1001, MaRecordInput{
		OrderID: 5001, DrugName: "促红素", Dose: "3000 IU", Route: "iv",
		AdministeredBy: "10", AdministeredName: "张护士",
	})
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if ma.Status != models.MAStatusRecorded || ma.DrugName != "促红素" {
		t.Fatalf("bad record: %+v", ma)
	}

	// empty executor
	if _, err := s.RecordAdmin(1001, MaRecordInput{OrderID: 5001, DrugName: "促红素"}); err == nil {
		t.Fatal("空执行人应拒")
	}
	// empty drug name
	if _, err := s.RecordAdmin(1001, MaRecordInput{OrderID: 5001, AdministeredBy: "10"}); err == nil {
		t.Fatal("空药品名应拒")
	}
	// zero order id
	if _, err := s.RecordAdmin(1001, MaRecordInput{AdministeredBy: "10", DrugName: "铁剂"}); err == nil {
		t.Fatal("空orderID应拒")
	}

	// sign_record 留痕
	var sr models.SignRecord
	if err := db.Where("target_type = ? AND target_id = ?", models.SignTargetMedicationAdmin, ma.ID).First(&sr).Error; err != nil {
		t.Fatal("应写入sign_record留痕")
	}
}

func TestMed_SecondCheck(t *testing.T) {
	db := newMaTestDB(t)
	s := &MedicationService{db: db, tenantID: 3}

	ma, _ := s.RecordAdmin(1002, MaRecordInput{
		OrderID: 5002, DrugName: "蔗糖铁", Dose: "100 mg", Route: "ivgtt",
		AdministeredBy: "10", AdministeredName: "张护士",
	})

	// 自核拒绝
	if _, err := s.SecondCheck(ma.ID, "10", "张护士"); err == nil {
		t.Fatal("自核应拒")
	}

	// 双核成功
	v, err := s.SecondCheck(ma.ID, "20", "李护士")
	if err != nil {
		t.Fatalf("双核失败: %v", err)
	}
	if v.Status != models.MAStatusVerified || v.SecondCheckBy != "20" {
		t.Fatalf("bad verify: %+v", v)
	}

	// 重复核拒
	if _, err := s.SecondCheck(ma.ID, "30", "王护士"); err == nil {
		t.Fatal("重复核应拒")
	}

	// sign_record 双签留痕
	var count int64
	db.Model(&models.SignRecord{}).Where("target_type = ? AND target_id = ?", models.SignTargetMedicationAdmin, ma.ID).Count(&count)
	if count != 2 {
		t.Fatalf("sign_record应有2条（执行人+核对人），实际%d", count)
	}
}

func TestMed_List(t *testing.T) {
	db := newMaTestDB(t)
	s := &MedicationService{db: db, tenantID: 3}

	s.RecordAdmin(2001, MaRecordInput{OrderID: 6001, TreatmentID: 101, DrugName: "促红素", AdministeredBy: "10"})
	s.RecordAdmin(2001, MaRecordInput{OrderID: 6002, TreatmentID: 101, DrugName: "蔗糖铁", AdministeredBy: "10"})
	s.RecordAdmin(2002, MaRecordInput{OrderID: 6003, TreatmentID: 102, DrugName: "碳酸钙", AdministeredBy: "10"})

	tid := int64(101)
	rows, err := s.List(&tid, nil, nil)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("treatment 101 应有2条记录，实际%d", len(rows))
	}

	pid := int64(2002)
	rows2, err := s.List(nil, &pid, nil)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows2) != 1 {
		t.Fatalf("patient 2002 应有1条，实际%d", len(rows2))
	}
}

func TestMed_DefaultDoses(t *testing.T) {
	s := &MedicationService{db: nil, tenantID: 3}
	doses, err := s.DefaultDoses()
	if err != nil {
		t.Fatalf("doses: %v", err)
	}
	if len(doses) == 0 {
		t.Fatal("应有至少1条enabled默认剂量")
	}
	for _, d := range doses {
		if d.Drug == "antihypertensive" {
			t.Fatal("降压药enabled=false不应出现")
		}
	}
}

func TestMed_Suggestions(t *testing.T) {
	s := &MedicationService{db: nil, tenantID: 3}
	sugs, err := s.Suggestions(1001)
	if err != nil {
		t.Fatalf("suggestions: %v", err)
	}
	if len(sugs) == 0 {
		t.Fatal("应有至少1条建议")
	}
	for _, sug := range sugs {
		if sug.Indicator == "" || sug.Label == "" {
			t.Fatalf("bad suggestion: %+v", sug)
		}
	}
}

func TestMed_RecordAdmin_SignFailureRollback(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:marollback?mode=memory"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(&models.MedicationAdmin{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	s := &MedicationService{db: db, tenantID: 3}

	_, err = s.RecordAdmin(3001, MaRecordInput{
		OrderID: 7001, DrugName: "促红素", AdministeredBy: "10",
	})
	if err == nil {
		t.Fatal("sign_record表不存在应返回错误")
	}

	var count int64
	db.Model(&models.MedicationAdmin{}).Count(&count)
	if count != 0 {
		t.Fatalf("签名失败应回滚，给药记录不应残留，actual=%d", count)
	}
}

func TestMed_SecondCheck_SignFailureRollback(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:mcrollback?mode=memory"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(&models.MedicationAdmin{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	ma := &models.MedicationAdmin{
		ID: "ma-test-1", TenantID: 3, PatientID: 3002,
		OrderID: 7002, DrugName: "蔗糖铁",
		AdministeredBy: "10", AdministeredName: "张护士",
		AdministeredAt: time.Now(), Status: models.MAStatusRecorded,
	}
	if err := db.Create(ma).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	s := &MedicationService{db: db, tenantID: 3}
	_, err = s.SecondCheck("ma-test-1", "20", "李护士")
	if err == nil {
		t.Fatal("sign_record表不存在应返回错误")
	}

	var check models.MedicationAdmin
	db.First(&check, "id = ?", "ma-test-1")
	if check.Status != models.MAStatusRecorded {
		t.Fatalf("签名失败应回滚，状态不应变为verified，actual=%s", check.Status)
	}
}
