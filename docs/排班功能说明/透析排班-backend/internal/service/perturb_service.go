package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/sdsph/dialysis-scheduling/internal/config"
	"github.com/sdsph/dialysis-scheduling/internal/model"
	"github.com/sdsph/dialysis-scheduling/internal/repo"
	"github.com/sdsph/dialysis-scheduling/internal/sched"
)

// 扰动处理:临时透析插入(决策 12)、设备停机迁移(决策 17)。
// 均遵循"主→备→报警入冲突队列"。

var ErrNoSlot = errors.New("当前/后续班次均无可用机位,已报警入冲突队列")

// raiseConflictDB 写一条冲突/待处理队列记录。
func raiseConflictDB(g *gorm.DB, tenant int64, patientID *int64, date *time.Time, shiftID, wardID *int64, ctype string, severity int16, detail string) {
	c := &model.ConflictQueue{
		BaseModel:    model.BaseModel{TenantId: tenant},
		PatientId:    patientID,
		ScheduleDate: date,
		ShiftId:      shiftID,
		WardId:       wardID,
		ConflictType: ctype,
		Severity:     severity,
		Detail:       detail,
		Status:       0,
	}
	g.Create(c)
}

// patientBookedAt 病人在某(日期+班次)是否已有有效排班。
func patientBookedAt(g *gorm.DB, tenant, patientID int64, date time.Time, shiftID int64) bool {
	var cnt int64
	g.Model(&model.PatientShift{}).Where(
		`"TenantId" = ? AND "PatientId" = ? AND "ScheduleDate" = ? AND "ShiftId" = ? AND "Status" NOT IN ?`,
		tenant, patientID, date, shiftID, []int16{sched.StatusCancelled, sched.StatusAbsent},
	).Count(&cnt)
	return cnt > 0
}

// InsertTemporary 临时透析(决策 12):在指定区,从当前班次起逐班找空机插入。
// 按医嘱模式匹配机型;可借用请假/缺席空出的机位(IsBorrowedSlot);不进模板(SourceType=临时)。
// 全部班次无空位 → 报警入冲突队列并返回 ErrNoSlot。
func InsertTemporary(g *gorm.DB, tenant, patientID, wardID int64, date time.Time, mode string) (*model.PatientShift, error) {
	anchor := config.AnchorMonday(g, tenant)
	d := dayStart(date)
	board, err := repo.LoadBoard(g, tenant, anchor, d, d.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}
	now := time.Now()

	for _, sh := range board.ShiftList() {
		if sh.IsDisabled {
			continue
		}
		if patientBookedAt(g, tenant, patientID, d, sh.Id) {
			continue // 该病人此班已有排班,换下一班
		}
		m := board.FindFreeForMode(wardID, sh.Id, d, mode)
		if m == nil {
			continue
		}
		// 该机位是否曾被取消/缺席空出(借用留痕)
		var borrowedCnt int64
		g.Model(&model.PatientShift{}).Where(
			`"TenantId" = ? AND "MachineId" = ? AND "ScheduleDate" = ? AND "ShiftId" = ? AND "Status" IN ?`,
			tenant, m.Id, d, sh.Id, []int16{sched.StatusCancelled, sched.StatusAbsent},
		).Count(&borrowedCnt)

		shiftID, machineID := sh.Id, m.Id
		rec := &model.PatientShift{
			BaseModel:      model.BaseModel{TenantId: tenant},
			PatientId:      patientID,
			ScheduleDate:   d,
			ShiftId:        &shiftID,
			WardId:         m.WardId,
			MachineId:      &machineID,
			Status:         sched.StatusConfirmed, // 急诊即时排入
			DialysisMode:   mode,
			SourceType:     sched.SourceTemporary, // 不进模板
			RecordForm:     sched.RecordFormRegular,
			IsBorrowedSlot: borrowedCnt > 0,
			Confirm1At:     &now,
		}
		if err := g.Create(rec).Error; err != nil {
			return nil, err
		}
		return rec, nil
	}

	wid := wardID
	raiseConflictDB(g, tenant, &patientID, &d, nil, &wid, sched.ConflictNoMachine, sched.SeverityAlert, "临时透析:当前及后续班次均无空机")
	return nil, ErrNoSlot
}

// OutageResult 停机迁移结果摘要。
type OutageResult struct {
	OutageId  int64 `json:"outageId"`
	Affected  int   `json:"affected"`
	Migrated  int   `json:"migrated"`
	Conflicts int   `json:"conflicts"`
}

// RegisterOutageAndMigrate 登记停机时段并迁移受影响排班(决策 17)。
// 临时停机(≤48h,type=10):为受影响病人就近找替代机位,改本次 MachineId(固定机位不变,修好可归位);
// 长期/报废(type=20):仅报警入队,由人工永久迁移。
func RegisterOutageAndMigrate(g *gorm.DB, tenant, machineID int64, start, end time.Time, outageType int16, reason string) (*OutageResult, error) {
	endPtr := end
	o := &model.MachineOutage{
		BaseModel:  model.BaseModel{TenantId: tenant},
		MachineId:  machineID,
		StartAt:    start,
		EndAt:      &endPtr,
		OutageType: outageType,
		Reason:     reason,
	}
	if err := g.Create(o).Error; err != nil {
		return nil, err
	}

	anchor := config.AnchorMonday(g, tenant)
	ds, de := dayStart(start), dayStart(end)
	// Board 在停机登记后加载,故已冻结该机器(替代时不会再选中它)。
	board, err := repo.LoadBoard(g, tenant, anchor, ds, de.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}

	var affected []*model.PatientShift
	if err := g.Where(
		`"TenantId" = ? AND "MachineId" = ? AND "ScheduleDate" BETWEEN ? AND ? AND "Status" IN ?`,
		tenant, machineID, ds, de, []int16{sched.StatusPending, sched.StatusDraft, sched.StatusConfirmed},
	).Find(&affected).Error; err != nil {
		return nil, err
	}

	res := &OutageResult{OutageId: o.Id, Affected: len(affected)}
	for _, s := range affected {
		if s.ShiftId == nil {
			continue
		}
		patientID := s.PatientId
		date := s.ScheduleDate
		shiftID := *s.ShiftId

		if outageType == sched.OutageLong {
			raiseConflictDB(g, tenant, &patientID, &date, &shiftID, &s.WardId, sched.ConflictMachineOutage, sched.SeverityAlert, "长期停机/报废:需人工永久重定机位")
			res.Conflicts++
			continue
		}

		repl := board.FindFreeForMode(s.WardId, shiftID, date, s.DialysisMode)
		if repl == nil {
			raiseConflictDB(g, tenant, &patientID, &date, &shiftID, &s.WardId, sched.ConflictMachineOutage, sched.SeverityAlert, "停机迁移:本区本班无替代空机")
			res.Conflicts++
			continue
		}
		if err := g.Model(s).Updates(map[string]interface{}{"MachineId": repl.Id, "WardId": repl.WardId}).Error; err != nil {
			return nil, err
		}
		board.MarkOccupied(repl.Id, sched.Cell{WardId: repl.WardId, ShiftId: shiftID, Date: date}, patientID)
		res.Migrated++
	}
	return res, nil
}

// HolidayResult 假日挪班结果摘要。
type HolidayResult struct {
	Cancelled int `json:"cancelled"` // 因停诊取消的排班
	Suggested int `json:"suggested"` // 已给出挪班建议(入冲突队列 HINT)
	NoSlot    int `json:"noSlot"`    // 前后均无空位(报警)
}

// SetHoliday 将某日设为非透析日并处理受影响排班(决策 19):
// 当日排班取消(停诊),并为每人就近(前后 7 天)给"挪班建议"入冲突队列;系统不自动改定。
func SetHoliday(g *gorm.DB, tenant int64, date time.Time, holidayMode int16) (*HolidayResult, error) {
	d := dayStart(date)

	var cal model.Calendar
	err := g.Where(`"TenantId" = ? AND "CalDate" = ?`, tenant, d).First(&cal).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cal = model.Calendar{BaseModel: model.BaseModel{TenantId: tenant}, CalDate: d, IsDialysisDay: false, HolidayMode: holidayMode}
		if err := g.Create(&cal).Error; err != nil {
			return nil, err
		}
	} else if err == nil {
		g.Model(&cal).Updates(map[string]interface{}{"IsDialysisDay": false, "HolidayMode": holidayMode})
	} else {
		return nil, err
	}

	var affected []*model.PatientShift
	if err := g.Where(`"TenantId" = ? AND "ScheduleDate" = ? AND "Status" IN ?`,
		tenant, d, []int16{sched.StatusPending, sched.StatusDraft, sched.StatusConfirmed}).Find(&affected).Error; err != nil {
		return nil, err
	}

	anchor := config.AnchorMonday(g, tenant)
	board, err := repo.LoadBoard(g, tenant, anchor, d.AddDate(0, 0, -7), d.AddDate(0, 0, 8))
	if err != nil {
		return nil, err
	}

	res := &HolidayResult{}
	for _, s := range affected {
		g.Model(s).Updates(map[string]interface{}{"Status": sched.StatusCancelled, "CancelReason": "假日停透"})
		res.Cancelled++
		if s.ShiftId == nil {
			continue
		}
		pid, sd, shid, wid := s.PatientId, s.ScheduleDate, *s.ShiftId, s.WardId
		altDate, altMachine := nearestHolidaySlot(board, s.WardId, *s.ShiftId, s.DialysisMode, d)
		if altMachine != nil {
			detail := "建议挪到 " + altDate.Format("2006-01-02") + " " + altMachine.Code
			raiseConflictDB(g, tenant, &pid, &sd, &shid, &wid, sched.ConflictHolidayReplan, sched.SeverityHint, detail)
			res.Suggested++
		} else {
			raiseConflictDB(g, tenant, &pid, &sd, &shid, &wid, sched.ConflictHolidayReplan, sched.SeverityAlert, "假日挪班:前后 7 天无空位")
			res.NoSlot++
		}
	}
	return res, nil
}

// nearestHolidaySlot 在 ±1..7 天内,找同区同班、透析日、机型匹配的最近空位(用于挪班建议)。
func nearestHolidaySlot(board *sched.Board, wardID, shiftID int64, mode string, d time.Time) (time.Time, *model.Machine) {
	for off := 1; off <= 7; off++ {
		for _, sign := range []int{1, -1} {
			cand := dayStart(d).AddDate(0, 0, off*sign)
			if !board.IsDialysisDay(cand) {
				continue
			}
			if m := board.FindFreeForMode(wardID, shiftID, cand, mode); m != nil {
				return cand, m
			}
		}
	}
	return time.Time{}, nil
}

// PlanChangeResult 方案变更结果摘要。
type PlanChangeResult struct {
	Replanned int    `json:"replanned"` // 未确认→已取消待重排
	Locked    int    `json:"locked"`    // 已二/三次确认→报警人工
	Hint      string `json:"hint"`
}

// ApplyPlanChange 方案变更带生效日(决策 14):
// 更新病人骨架(Profile+模板项);生效日后未二/三次确认的排班取消待重排,已确认的报警人工。
func ApplyPlanChange(g *gorm.DB, tenant, patientID int64, changeType, newValue string, effectiveDate time.Time) (*PlanChangeResult, error) {
	d := dayStart(effectiveDate)

	var prof model.PatientProfile
	if err := g.Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, patientID).First(&prof).Error; err != nil {
		return nil, errors.New("病人排班骨架不存在")
	}

	old, err := applyProfileChange(g, tenant, &prof, changeType, newValue)
	if err != nil {
		return nil, err
	}
	g.Create(&model.PlanChange{
		BaseModel: model.BaseModel{TenantId: tenant}, PatientId: patientID,
		ChangeType: changeType, OldValue: old, NewValue: newValue, EffectiveDate: d,
	})

	var future []*model.PatientShift
	if err := g.Where(`"TenantId" = ? AND "PatientId" = ? AND "ScheduleDate" >= ? AND "Status" IN ?`,
		tenant, patientID, d, []int16{sched.StatusPending, sched.StatusDraft, sched.StatusConfirmed}).Find(&future).Error; err != nil {
		return nil, err
	}

	res := &PlanChangeResult{}
	for _, s := range future {
		if s.Confirm2At != nil || s.Confirm3At != nil {
			pid, sd, wid := s.PatientId, s.ScheduleDate, s.WardId
			raiseConflictDB(g, tenant, &pid, &sd, s.ShiftId, &wid, sched.ConflictPlanChange, sched.SeverityAlert, "已确认排班遇方案变更,需人工处理")
			res.Locked++
		} else {
			g.Model(s).Updates(map[string]interface{}{"Status": sched.StatusCancelled, "CancelReason": "方案变更重排"})
			res.Replanned++
		}
	}
	res.Hint = "已更新骨架并清理未确认排班,请重新生成以按新方案补排"
	return res, nil
}

// applyProfileChange 按变更类型更新 Profile 与模板项,返回旧值字符串。
func applyProfileChange(g *gorm.DB, tenant int64, prof *model.PatientProfile, changeType, newValue string) (string, error) {
	var old string
	profUpd := map[string]interface{}{}
	itemUpd := map[string]interface{}{}
	switch changeType {
	case "FREQ":
		old = itoa(int64(prof.FreqPattern))
		n, e := parseInt16(newValue)
		if e != nil {
			return "", errors.New("FREQ 需为频率模式数字(10/20/30/40/90)")
		}
		profUpd["FreqPattern"] = n
		itemUpd["FreqPattern"] = n
	case "SHIFT":
		if prof.ShiftId != nil {
			old = itoa(*prof.ShiftId)
		}
		n, e := parseInt16(newValue) // shiftId 复用解析
		if e != nil {
			return "", errors.New("SHIFT 需为班次ID")
		}
		sid := int64(n)
		profUpd["ShiftId"] = sid
		itemUpd["ShiftId"] = sid
	case "ZONE":
		old = prof.ZoneTag
		profUpd["ZoneTag"] = newValue
		itemUpd["ZoneTag"] = newValue
	case "MODE":
		old = prof.DefaultMode
		profUpd["DefaultMode"] = newValue
	case "HDF":
		old = boolStr(prof.HdfEnabled)
		on := newValue == "true" || newValue == "on" || newValue == "1"
		profUpd["HdfEnabled"] = on
		itemUpd["HdfEnabled"] = on
	default:
		return "", errors.New("不支持的 changeType(FREQ/SHIFT/ZONE/MODE/HDF)")
	}

	if err := g.Model(&model.PatientProfile{}).Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, prof.PatientId).
		Updates(profUpd).Error; err != nil {
		return "", err
	}
	if len(itemUpd) > 0 {
		g.Model(&model.ScheduleTemplateItem{}).Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, prof.PatientId).Updates(itemUpd)
	}
	return old, nil
}

func parseInt16(s string) (int16, error) {
	n := 0
	if s == "" {
		return 0, errors.New("空值")
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, errors.New("非数字")
		}
		n = n*10 + int(ch-'0')
	}
	return int16(n), nil
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
