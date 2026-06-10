package service

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/sched"
)

// isUniqueViolation 判断是否 PostgreSQL 唯一约束冲突(23505)——并发撞唯一索引时触发。
func isUniqueViolation(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23505"
}

// 排班操作:三级确认、取消、缺席、移床/换班。均带编辑保护(规范 §8)。

var (
	ErrNotFound  = errors.New("排班记录不存在")
	ErrLocked    = errors.New("该排班已锁定(历史日期或已开始治疗),不可修改")
	ErrOccupied  = errors.New("目标机位在该日该班已被占用")
	ErrDoubleBook = errors.New("该病人在目标日期+班次已有排班")
	ErrModeMismatch = errors.New("目标机器机型不支持该治疗模式")
)

func dayStart(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// canEdit 编辑保护:已上机/已完成、或排班日早于今天,均锁定。
func canEdit(s *model.PatientShift) error {
	if s.Status == sched.StatusInDialysis || s.Status == sched.StatusCompleted {
		return ErrLocked
	}
	if dayStart(s.ScheduleDate).Before(dayStart(time.Now())) {
		return ErrLocked
	}
	return nil
}

func loadShift(g *gorm.DB, tenant, id int64) (*model.PatientShift, error) {
	var s model.PatientShift
	err := g.Where(`"TenantId" = ? AND "Id" = ?`, tenant, id).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &s, err
}

// ConfirmPlan 第一次确认(护士长):把 [weekStart, +weeks] 内的草稿(10)整体确认为生效(20)。
func ConfirmPlan(g *gorm.DB, tenant, by int64, weekStart time.Time, weeks int) (int64, error) {
	end := weekStart.AddDate(0, 0, weeks*7)
	now := time.Now()
	res := g.Model(&model.PatientShift{}).
		Where(`"TenantId" = ? AND "TreatmentTime" >= ? AND "TreatmentTime" < ? AND "Status" = ?`,
			tenant, weekStart, end, sched.StatusDraft).
		Updates(map[string]interface{}{"Status": sched.StatusConfirmed, "Confirm1At": now, "Confirm1By": by})
	return res.RowsAffected, res.Error
}

// ConfirmDay 第二/三次确认:对某日已确认(20)的排班盖 Confirm2At 或 Confirm3At。
// level=2 次日确认;level=3 当日确认。
func ConfirmDay(g *gorm.DB, tenant, by int64, date time.Time, level int) (int64, error) {
	col := "Confirm2At"
	byCol := "Confirm2By"
	if level == 3 {
		col, byCol = "Confirm3At", "Confirm3By"
	}
	now := time.Now()
	d := dayStart(date)
	res := g.Model(&model.PatientShift{}).
		Where(`"TenantId" = ? AND "TreatmentTime" = ? AND "Status" = ?`, tenant, d, sched.StatusConfirmed).
		Updates(map[string]interface{}{col: now, byCol: by})
	return res.RowsAffected, res.Error
}

// CancelShift 取消(提前请假/计划取消):Status→70,留痕。
func CancelShift(g *gorm.DB, tenant, id int64, reason string) error {
	s, err := loadShift(g, tenant, id)
	if err != nil {
		return err
	}
	if err := canEdit(s); err != nil {
		return err
	}
	return g.Model(s).Updates(map[string]interface{}{
		"Status": sched.StatusCancelled, "CancelReason": reason,
	}).Error
}

// MarkAbsent 当日缺席(爽约):Status→80,留痕,机位释放可借。
func MarkAbsent(g *gorm.DB, tenant, id int64, reason string) error {
	s, err := loadShift(g, tenant, id)
	if err != nil {
		return err
	}
	if err := canEdit(s); err != nil {
		return err
	}
	return g.Model(s).Updates(map[string]interface{}{
		"Status": sched.StatusAbsent, "CancelReason": reason,
	}).Error
}

// MoveShift 移床/换班:把一条排班移到新机器(可同时改日期/班次)。
// 校验:编辑保护 + 目标机型匹配 + 目标机位空闲 + 病人不双排。
func MoveShift(g *gorm.DB, tenant, id int64, newMachineId int64, newDate *time.Time, newShiftId *int64) error {
	s, err := loadShift(g, tenant, id)
	if err != nil {
		return err
	}
	if err := canEdit(s); err != nil {
		return err
	}

	var m model.Machine
	if err := g.Where(`"TenantId" = ? AND "Id" = ?`, tenant, newMachineId).First(&m).Error; err != nil {
		return errors.New("目标机器不存在")
	}
	if !sched.MachineSupports(m.MachineType, s.DialysisMode) {
		return ErrModeMismatch
	}

	date := s.ScheduleDate
	if newDate != nil {
		date = dayStart(*newDate)
	}
	shiftId := s.ShiftId
	if newShiftId != nil {
		shiftId = *newShiftId
	}

	// 目标机位是否被占(排除自身,取消/缺席不占)
	var cnt int64
	if err := g.Model(&model.PatientShift{}).Where(
		`"TenantId" = ? AND "MachineId" = ? AND "TreatmentTime" = ? AND "ShiftId" = ? AND "Id" <> ? AND "Status" NOT IN ?`,
		tenant, newMachineId, date, shiftId, id, []int16{sched.StatusCancelled, sched.StatusAbsent},
	).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return ErrOccupied
	}

	// 病人是否已在目标日期+班次有排班
	var cnt2 int64
	if err := g.Model(&model.PatientShift{}).Where(
		`"TenantId" = ? AND "PatientId" = ? AND "TreatmentTime" = ? AND "ShiftId" = ? AND "Id" <> ? AND "Status" NOT IN ?`,
		tenant, s.PatientId, date, shiftId, id, []int16{sched.StatusCancelled, sched.StatusAbsent},
	).Count(&cnt2).Error; err != nil {
		return err
	}
	if cnt2 > 0 {
		return ErrDoubleBook
	}

	err = g.Model(s).Updates(map[string]interface{}{
		"MachineId": newMachineId, "WardId": m.WardId, "TreatmentTime": date, "ShiftId": shiftId,
	}).Error
	if isUniqueViolation(err) { // 并发:目标位被别人抢先占用
		return ErrOccupied
	}
	return err
}
