package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/sched"
)

// 资源与病人的录入/维护(P1 #4)。病区 / 机器 / 班次 / 病人主档 / 排班骨架的增改查停。
// 校验集中在此,非法输入返回友好错误。

func validZone(z string) bool { return z == sched.ZoneA || z == sched.ZoneB || z == sched.ZoneC }
func validMachineType(t string) bool {
	return t == sched.MachineHD || t == sched.MachineHDF || t == sched.MachineCRRT
}
func validFreq(f int16) bool {
	switch f {
	case sched.FreqMonWedFri, sched.FreqTueThuSat, sched.FreqTwoPerWk, sched.FreqOnePerWk, sched.FreqTemporary:
		return true
	}
	return false
}

// ---------- 病区 Ward ----------

func ListWards(g *gorm.DB, tenant int64) ([]model.Ward, error) {
	var ws []model.Ward
	err := g.Where(`"TenantId" = ?`, tenant).Order(`"Sort", "Id"`).Find(&ws).Error
	return ws, err
}

func CreateWard(g *gorm.DB, tenant, creator int64, w *model.Ward) (*model.Ward, error) {
	if w.Name == "" || !validZone(w.ZoneType) {
		return nil, errors.New("病区名必填,ZoneType 须为 A/B/C")
	}
	w.TenantId, w.CreatorId, w.Id = tenant, creator, 0
	if err := g.Create(w).Error; err != nil {
		return nil, err
	}
	return w, nil
}

func UpdateWard(g *gorm.DB, tenant, id int64, w *model.Ward) error {
	if w.Name == "" || !validZone(w.ZoneType) {
		return errors.New("病区名必填,ZoneType 须为 A/B/C")
	}
	return g.Model(&model.Ward{}).Where(`"TenantId" = ? AND "Id" = ?`, tenant, id).
		Updates(map[string]interface{}{
			"Name": w.Name, "ZoneType": w.ZoneType, "ParentWardId": w.ParentWardId,
			"IsSubZone": w.IsSubZone, "Sort": w.Sort, "Note": w.Note,
		}).Error
}

// ---------- 机器 Machine ----------

func ListMachines(g *gorm.DB, tenant int64, wardID int64) ([]model.Machine, error) {
	q := g.Where(`"TenantId" = ?`, tenant)
	if wardID > 0 {
		q = q.Where(`"WardId" = ?`, wardID)
	}
	var ms []model.Machine
	err := q.Order(`"WardId", "PositionIndex"`).Find(&ms).Error
	return ms, err
}

func CreateMachine(g *gorm.DB, tenant, creator int64, m *model.Machine) (*model.Machine, error) {
	if m.Code == "" || !validMachineType(m.MachineType) || m.WardId == 0 {
		return nil, errors.New("机器编号/所属病区必填,机型须为 HD/HDF/CRRT")
	}
	m.TenantId, m.CreatorId, m.Id = tenant, creator, 0
	if err := g.Create(m).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, errors.New("机器编号已存在")
		}
		return nil, err
	}
	return m, nil
}

func UpdateMachine(g *gorm.DB, tenant, id int64, m *model.Machine) error {
	if m.Code == "" || !validMachineType(m.MachineType) {
		return errors.New("机器编号必填,机型须为 HD/HDF/CRRT")
	}
	err := g.Model(&model.Machine{}).Where(`"TenantId" = ? AND "Id" = ?`, tenant, id).
		Updates(map[string]interface{}{
			"WardId": m.WardId, "Code": m.Code, "Name": m.Name,
			"MachineType": m.MachineType, "PositionIndex": m.PositionIndex, "Sort": m.Sort, "Note": m.Note,
		}).Error
	if isUniqueViolation(err) {
		return errors.New("机器编号已存在")
	}
	return err
}

// ---------- 班次 Shift ----------

func ListShifts(g *gorm.DB, tenant int64) ([]model.Shift, error) {
	var ss []model.Shift
	err := g.Where(`"TenantId" = ?`, tenant).Order(`"Sort", "Id"`).Find(&ss).Error
	return ss, err
}

func CreateShift(g *gorm.DB, tenant, creator int64, s *model.Shift) (*model.Shift, error) {
	if s.Name == "" || s.ShiftCode == "" {
		return nil, errors.New("班次名与班次码必填")
	}
	s.TenantId, s.CreatorId, s.Id = tenant, creator, 0
	if err := g.Create(s).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, errors.New("班次码已存在")
		}
		return nil, err
	}
	return s, nil
}

func UpdateShift(g *gorm.DB, tenant, id int64, s *model.Shift) error {
	if s.Name == "" || s.ShiftCode == "" {
		return errors.New("班次名与班次码必填")
	}
	return g.Model(&model.Shift{}).Where(`"TenantId" = ? AND "Id" = ?`, tenant, id).
		Updates(map[string]interface{}{
			"Name": s.Name, "ShiftCode": s.ShiftCode, "StartTime": s.StartTime,
			"EndTime": s.EndTime, "Sort": s.Sort,
		}).Error
}

// ---------- 通用启停 ----------

// SetDisabled 启用/停用资源(table 限定为 Ward/Machine/Shift 表名)。
func SetDisabled(g *gorm.DB, tenant int64, table string, id int64, disabled bool) error {
	switch table {
	case "ward", "machine", "shift":
	default:
		return errors.New("不支持的资源类型")
	}
	tbl := map[string]string{"ward": "Schedule_v2_Ward", "machine": "Schedule_v2_Machine", "shift": "Schedule_v2_Shift"}[table]
	return g.Table(`"`+tbl+`"`).Where(`"TenantId" = ? AND "Id" = ?`, tenant, id).
		Update("IsDisabled", disabled).Error
}

// ---------- 病人主档 Patient ----------

func ListPatients(g *gorm.DB, tenant int64) ([]model.Patient, error) {
	var ps []model.Patient
	err := g.Where(`"TenantId" = ?`, tenant).Order(`"Id"`).Find(&ps).Error
	return ps, err
}

func UpsertPatient(g *gorm.DB, tenant int64, p *model.Patient) (*model.Patient, error) {
	if p.Id == 0 || p.Name == "" {
		return nil, errors.New("病人ID与姓名必填")
	}
	p.TenantId = tenant
	var exist model.Patient
	err := g.Where(`"TenantId" = ? AND "Id" = ?`, tenant, p.Id).First(&exist).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if e := g.Create(p).Error; e != nil {
			return nil, e
		}
		return p, nil
	}
	if err != nil {
		return nil, err
	}
	return p, g.Model(&model.Patient{}).Where(`"TenantId" = ? AND "Id" = ?`, tenant, p.Id).
		Updates(map[string]interface{}{"Name": p.Name, "Gender": p.Gender}).Error
}

// ---------- 排班骨架 PatientProfile ----------

func ListProfiles(g *gorm.DB, tenant int64) ([]model.PatientProfile, error) {
	var ps []model.PatientProfile
	err := g.Where(`"TenantId" = ?`, tenant).Order(`"PatientId"`).Find(&ps).Error
	return ps, err
}

func GetProfile(g *gorm.DB, tenant, patientID int64) (*model.PatientProfile, error) {
	var p model.PatientProfile
	err := g.Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, patientID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

// validateProfile 校验骨架:分区、频率、HDF 日须落在频率透析日内。
func validateProfile(p *model.PatientProfile) error {
	if p.PatientId == 0 {
		return errors.New("PatientId 必填")
	}
	if !validZone(p.ZoneTag) {
		return errors.New("ZoneTag 须为 A/B/C")
	}
	if !validFreq(p.FreqPattern) {
		return errors.New("FreqPattern 非法(10/20/30/40/90)")
	}
	// 决策 23:每周次数(医嘱)与星期组合(护士)须一致
	if p.WeeklyCount > 0 && p.FreqPattern != sched.FreqTemporary {
		if int16(len(sched.FreqWeekdays(p.FreqPattern))) != p.WeeklyCount {
			return errors.New("星期组合的天数与每周次数不一致")
		}
	}
	if p.HdfEnabled {
		if p.HdfWeekday == nil {
			return errors.New("启用 HDF 时必须指定 HdfWeekday(1=周一..6=周六)")
		}
		ok := false
		for _, wd := range sched.FreqWeekdays(p.FreqPattern) {
			// time.Weekday: 周日=0..周六=6;ISO HdfWeekday: 周一=1..周日=7
			iso := int16(wd)
			if wd == 0 {
				iso = 7
			}
			if iso == *p.HdfWeekday {
				ok = true
			}
		}
		if !ok {
			return errors.New("HdfWeekday 必须是该频率的透析日之一(决策 3)")
		}
	}
	return nil
}

// UpsertProfile 新建或更新病人排班骨架(1:1 病人)。
func UpsertProfile(g *gorm.DB, tenant int64, p *model.PatientProfile) (*model.PatientProfile, error) {
	if err := validateProfile(p); err != nil {
		return nil, err
	}
	p.TenantId = tenant
	var exist model.PatientProfile
	err := g.Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, p.PatientId).First(&exist).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		p.Id = 0
		if e := g.Create(p).Error; e != nil {
			return nil, e
		}
		return p, nil
	}
	if err != nil {
		return nil, err
	}
	return p, g.Model(&model.PatientProfile{}).Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, p.PatientId).
		Updates(map[string]interface{}{
			"ZoneTag": p.ZoneTag, "HomeWardId": p.HomeWardId, "WeeklyCount": p.WeeklyCount, "FreqPattern": p.FreqPattern,
			"ShiftId": p.ShiftId, "DefaultMode": p.DefaultMode, "HdfEnabled": p.HdfEnabled,
			"HdfWeekday": p.HdfWeekday, "IsAdmissionRejected": p.IsAdmissionRejected,
		}).Error
}
