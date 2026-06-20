package database

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestVerifyRequiredTables(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// 只建应用依赖新表中的 1 张 → 其他表应判为缺失
	if err := db.Exec(`CREATE TABLE sign_record (id varchar(36) primary key)`).Error; err != nil {
		t.Fatalf("create sign_record: %v", err)
	}

	missing := VerifyRequiredTables(db)
	has := map[string]bool{}
	for _, m := range missing {
		has[m] = true
	}
	if has["sign_record"] {
		t.Errorf("sign_record 已建，不应在缺失列表")
	}
	if !has["Schedule_StaffDuty"] || !has["Schedule_StaffDutyOverride"] || !has["exam_reports"] || !has["sync_job_runs"] {
		t.Errorf("未建的新表应判为缺失，得 %v", missing)
	}
	if len(missing) != len(RequiredNewTables)-1 {
		t.Errorf("缺失数量不正确，得 %d (%v)", len(missing), missing)
	}
}

func TestVerifyRequiredTablesNilDB(t *testing.T) {
	if m := VerifyRequiredTables(nil); m != nil {
		t.Errorf("nil db 应返回 nil, 得 %v", m)
	}
}
