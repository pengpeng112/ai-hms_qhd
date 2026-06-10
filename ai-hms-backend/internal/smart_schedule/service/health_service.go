package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
)

// HealthSummary 排班数据健康检查结果
type HealthSummary struct {
	CheckedAt string   `json:"checkedAt"`
	TenantId  int64    `json:"tenantId"`

	// 基础资源
	WardCount     int `json:"wardCount"`
	WardDisabled  int `json:"wardDisabled"`
	MachineCount  int `json:"machineCount"`
	MachineDisabled int `json:"machineDisabled"`
	MachineNoWard int `json:"machineNoWard"`
	ShiftCount    int `json:"shiftCount"`
	ShiftNoCode   int `json:"shiftNoCode"`

	// 排班数据
	PatientCount     int `json:"patientCount"`
	ProfileCount     int `json:"profileCount"`
	ProfileNoShift   int `json:"profileNoShift"`
	ProfileNoWard    int `json:"profileNoWard"`
	TemplateCount    int `json:"templateCount"`
	TemplateActive   int `json:"templateActive"`
	TemplateItemCount int `json:"templateItemCount"`

	// 排班记录
	TotalShifts    int `json:"totalShifts"`
	DraftShifts    int `json:"draftShifts"`
	ConfirmedShifts int `json:"confirmedShifts"`
	CancelledShifts int `json:"cancelledShifts"`
	AbsentShifts   int `json:"absentShifts"`
	ShiftsNoMachine int `json:"shiftsNoMachine"`

	// 冲突
	OpenConflicts   int `json:"openConflicts"`
	TotalConflicts  int `json:"totalConflicts"`
	ConflictTypes   map[string]int `json:"conflictTypes"`

	// 异常项
	PatientsWithShiftsButNoProfileCount int `json:"patientsWithShiftsButNoProfileCount"`
	PatientsWithProfileButNoPatientCount int `json:"patientsWithProfileButNoPatientCount"`
	Warnings                             []string `json:"warnings"`
}

func ComputeHealth(g *gorm.DB, tenant int64) (*HealthSummary, error) {
	h := &HealthSummary{CheckedAt: time.Now().Format(time.RFC3339), TenantId: tenant, ConflictTypes: map[string]int{}}
	var n int64

	// ---- 基础资源 ----
	g.Model(&model.Ward{}).Where(`"TenantId"=?`, tenant).Count(&n); h.WardCount = int(n)
	g.Model(&model.Ward{}).Where(`"TenantId"=? AND "IsDisabled"=true`, tenant).Count(&n); h.WardDisabled = int(n)

	g.Model(&model.Machine{}).Where(`"TenantId"=?`, tenant).Count(&n); h.MachineCount = int(n)
	g.Model(&model.Machine{}).Where(`"TenantId"=? AND "IsDisabled"=true`, tenant).Count(&n); h.MachineDisabled = int(n)
	g.Model(&model.Machine{}).Where(`"TenantId"=? AND ("WardId"=0 OR "WardId" IS NULL)`, tenant).Count(&n); h.MachineNoWard = int(n)

	g.Model(&model.Shift{}).Where(`"TenantId"=?`, tenant).Count(&n); h.ShiftCount = int(n)
	g.Model(&model.Shift{}).Where(`"TenantId"=? AND ("ShiftCode"='' OR "ShiftCode" IS NULL)`, tenant).Count(&n); h.ShiftNoCode = int(n)

	// ---- 排班数据 ----
	g.Model(&model.Patient{}).Where(`"TenantId"=?`, tenant).Count(&n); h.PatientCount = int(n)

	g.Model(&model.PatientProfile{}).Where(`"TenantId"=?`, tenant).Count(&n); h.ProfileCount = int(n)
	g.Model(&model.PatientProfile{}).Where(`"TenantId"=? AND "PatientStatus" != 20 AND ("ShiftId"=0 OR "ShiftId" IS NULL)`, tenant).Count(&n); h.ProfileNoShift = int(n)
	g.Model(&model.PatientProfile{}).Where(`"TenantId"=? AND "PatientStatus" != 20 AND ("HomeWardId"=0 OR "HomeWardId" IS NULL)`, tenant).Count(&n); h.ProfileNoWard = int(n)

	g.Model(&model.ScheduleTemplate{}).Where(`"TenantId"=?`, tenant).Count(&n); h.TemplateCount = int(n)
	g.Model(&model.ScheduleTemplate{}).Where(`"TenantId"=? AND "IsActive"=true`, tenant).Count(&n); h.TemplateActive = int(n)
	g.Model(&model.ScheduleTemplateItem{}).Where(`"TenantId"=?`, tenant).Count(&n); h.TemplateItemCount = int(n)

	// ---- 排班记录 ----
	g.Model(&model.PatientShift{}).Where(`"TenantId"=?`, tenant).Count(&n); h.TotalShifts = int(n)
	g.Model(&model.PatientShift{}).Where(`"TenantId"=? AND "Status"=10`, tenant).Count(&n); h.DraftShifts = int(n)
	g.Model(&model.PatientShift{}).Where(`"TenantId"=? AND "Status"=20`, tenant).Count(&n); h.ConfirmedShifts = int(n)
	g.Model(&model.PatientShift{}).Where(`"TenantId"=? AND "Status"=70`, tenant).Count(&n); h.CancelledShifts = int(n)
	g.Model(&model.PatientShift{}).Where(`"TenantId"=? AND "Status"=80`, tenant).Count(&n); h.AbsentShifts = int(n)
	g.Model(&model.PatientShift{}).Where(`"TenantId"=? AND "MachineId"=0`, tenant).Count(&n); h.ShiftsNoMachine = int(n)

	// ---- 冲突 ----
	g.Model(&model.ConflictQueue{}).Where(`"TenantId"=?`, tenant).Count(&n); h.TotalConflicts = int(n)
	g.Model(&model.ConflictQueue{}).Where(`"TenantId"=? AND "Status"=0`, tenant).Count(&n); h.OpenConflicts = int(n)
	type cntRow struct{ ConflictType string; N int }
	var crs []cntRow
	g.Model(&model.ConflictQueue{}).Select(`"ConflictType" AS conflict_type, count(*) AS n`).
		Where(`"TenantId"=? AND "Status"=0`, tenant).Group(`"ConflictType"`).Scan(&crs)
	for _, r := range crs { h.ConflictTypes[r.ConflictType] = r.N }

	// ---- 异常项(用SQL JOIN直接计数,避免逐条查询) ----
	if err := g.Raw(`SELECT count(*) FROM (
		SELECT DISTINCT ps."PatientId" FROM "Schedule_PatientShift" ps
		LEFT JOIN "Schedule_PatientProfile" pp ON pp."TenantId" = ps."TenantId" AND pp."PatientId" = ps."PatientId"
		WHERE ps."TenantId" = ? AND pp."Id" IS NULL LIMIT 500
	) t`, tenant).Scan(&n).Error; err != nil {
		return nil, fmt.Errorf("computeHealth patientsWithShiftsButNoProfile: %w", err)
	}
	h.PatientsWithShiftsButNoProfileCount = int(n)

	if err := g.Raw(`SELECT count(*) FROM (
		SELECT pp."PatientId" FROM "Schedule_PatientProfile" pp
		LEFT JOIN "Schedule_Patient" p ON p."TenantId" = pp."TenantId" AND p."Id" = pp."PatientId"
		WHERE pp."TenantId" = ? AND p."Id" IS NULL LIMIT 500
	) t`, tenant).Scan(&n).Error; err != nil {
		return nil, fmt.Errorf("computeHealth patientsWithProfileButNoPatient: %w", err)
	}
	h.PatientsWithProfileButNoPatientCount = int(n)

	// ---- 警告 ----
	if h.MachineCount == 0 { h.Warnings = append(h.Warnings, "无条件床位数据") }
	if h.MachineCount > 0 && h.MachineNoWard > 0 { h.Warnings = append(h.Warnings, fmt.Sprintf(" %d 台机器无归属病区", h.MachineNoWard)) }
	if h.MachineCount > 0 && h.MachineDisabled > 0 { h.Warnings = append(h.Warnings, fmt.Sprintf(" %d 台机器已停用", h.MachineDisabled)) }
	if h.ShiftCount == 0 { h.Warnings = append(h.Warnings, "无人值班次数据") }
	if h.PatientCount == 0 { h.Warnings = append(h.Warnings, "条件病人主档(Generate 将从空病人库生成排班)") }
	if h.ProfileCount == 0 { h.Warnings = append(h.Warnings, "条件排班骨架(Engine reads from template)") }
	if h.ProfileNoShift > 0 { h.Warnings = append(h.Warnings, fmt.Sprintf(" %d 个有效骨架未绑定班次", h.ProfileNoShift)) }
	if h.ProfileNoWard > 0 { h.Warnings = append(h.Warnings, fmt.Sprintf(" %d 个有效骨架未绑定病区", h.ProfileNoWard)) }
	if h.TemplateCount == 0 { h.Warnings = append(h.Warnings, "条件排班模板(Generate 无骨架可排)") }
	if h.TemplateActive == 0 { h.Warnings = append(h.Warnings, "条件生效模板") }
	if h.ShiftNoCode > 0 { h.Warnings = append(h.Warnings, fmt.Sprintf(" %d 个班次未填写班次码", h.ShiftNoCode)) }
	if h.PatientsWithShiftsButNoProfileCount > 0 { h.Warnings = append(h.Warnings, fmt.Sprintf(" %d 个病人有排班记录但无排班骨架(抽样)", h.PatientsWithShiftsButNoProfileCount)) }
	if h.PatientsWithProfileButNoPatientCount > 0 { h.Warnings = append(h.Warnings, fmt.Sprintf(" %d 个病人有骨架但无主档(抽样)", h.PatientsWithProfileButNoPatientCount)) }
	if h.OpenConflicts > 100 { h.Warnings = append(h.Warnings, fmt.Sprintf("待处理冲突 %d 条,建议清理", h.OpenConflicts)) }

	return h, nil
}
