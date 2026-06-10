package service

import (
	"time"

	"gorm.io/gorm"

	"github.com/sdsph/dialysis-scheduling/internal/model"
	"github.com/sdsph/dialysis-scheduling/internal/sched"
)

// 周视图聚合(对应老系统 schedule_week_service 的"区→机器→班×日"三层结构)。

type ShiftDTO struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Sort int    `json:"sort"`
}

type CellDTO struct {
	Id           int64  `json:"id"` // 排班记录 Id,供前端操作(取消/移床)
	ShiftId      int64  `json:"shiftId"`
	PatientId    int64  `json:"patientId"`
	PatientName  string `json:"patientName"`
	DialysisMode string `json:"dialysisMode"`
	Status       int16  `json:"status"`
	SourceType   int16  `json:"sourceType"`
	Confirms     int    `json:"confirms"` // 已完成确认级别 0~3
}

type MachineDTO struct {
	Id            int64               `json:"id"`
	Code          string              `json:"code"`
	MachineType   string              `json:"machineType"`
	PositionIndex int                 `json:"positionIndex"`
	Cells         map[string]CellDTO  `json:"cells"` // key = "YYYY-MM-DD|shiftId"
}

type WardDTO struct {
	Id       int64        `json:"id"`
	Name     string       `json:"name"`
	ZoneType string       `json:"zoneType"`
	Machines []MachineDTO `json:"machines"`
}

type WeekBoard struct {
	WeekStart string     `json:"weekStart"`
	WeekEnd   string     `json:"weekEnd"`
	Dates     []string   `json:"dates"`
	Shifts    []ShiftDTO `json:"shifts"`
	Wards     []WardDTO  `json:"wards"`
}

func cellKey(date string, shiftID int64) string {
	return date + "|" + itoa(shiftID)
}

// patientNames 加载租户病人 Id→姓名 映射。
func patientNames(g *gorm.DB, tenant int64) map[int64]string {
	var ps []model.Patient
	g.Where(`"TenantId" = ?`, tenant).Find(&ps)
	m := map[int64]string{}
	for _, p := range ps {
		m[p.Id] = p.Name
	}
	return m
}

func itoa(v int64) string {
	// 复用标准库太重,这里就地实现避免再引依赖。
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// BuildWeekBoard 聚合某日所在周(周一~周六,跳过周日)的排班矩阵。
func BuildWeekBoard(g *gorm.DB, tenant int64, date time.Time) (*WeekBoard, error) {
	mon := sched.MondayOf(date)
	sun := mon.AddDate(0, 0, 6)

	var dates []string
	for i := 0; i < 6; i++ { // 周一~周六
		dates = append(dates, mon.AddDate(0, 0, i).Format("2006-01-02"))
	}

	var shifts []model.Shift
	if err := g.Where(`"TenantId" = ? AND "IsDisabled" = false`, tenant).Order(`"Sort"`).Find(&shifts).Error; err != nil {
		return nil, err
	}
	var wards []model.Ward
	if err := g.Where(`"TenantId" = ? AND "IsDisabled" = false`, tenant).Order(`"Sort"`).Find(&wards).Error; err != nil {
		return nil, err
	}
	var machines []model.Machine
	if err := g.Where(`"TenantId" = ? AND "IsDisabled" = false`, tenant).
		Order(`"WardId", "PositionIndex"`).Find(&machines).Error; err != nil {
		return nil, err
	}
	var records []model.PatientShift
	if err := g.Where(
		`"TenantId" = ? AND "ScheduleDate" BETWEEN ? AND ? AND "MachineId" IS NOT NULL AND "Status" NOT IN ?`,
		tenant, mon, sun, []int16{sched.StatusCancelled},
	).Find(&records).Error; err != nil {
		return nil, err
	}
	names := patientNames(g, tenant)

	// machineId -> cellKey -> cell
	occ := map[int64]map[string]CellDTO{}
	for _, r := range records {
		if r.MachineId == nil || r.ShiftId == nil {
			continue
		}
		mid := *r.MachineId
		if occ[mid] == nil {
			occ[mid] = map[string]CellDTO{}
		}
		key := cellKey(r.ScheduleDate.Format("2006-01-02"), *r.ShiftId)
		confirms := 0
		if r.Confirm3At != nil {
			confirms = 3
		} else if r.Confirm2At != nil {
			confirms = 2
		} else if r.Confirm1At != nil || r.Status >= sched.StatusConfirmed {
			confirms = 1
		}
		occ[mid][key] = CellDTO{
			Id: r.Id, ShiftId: *r.ShiftId, PatientId: r.PatientId, PatientName: names[r.PatientId],
			DialysisMode: r.DialysisMode, Status: r.Status, SourceType: r.SourceType,
			Confirms: confirms,
		}
	}

	machinesByWard := map[int64][]MachineDTO{}
	for _, m := range machines {
		machinesByWard[m.WardId] = append(machinesByWard[m.WardId], MachineDTO{
			Id: m.Id, Code: m.Code, MachineType: m.MachineType, PositionIndex: m.PositionIndex,
			Cells: occ[m.Id],
		})
	}

	var wardDTOs []WardDTO
	for _, w := range wards {
		wardDTOs = append(wardDTOs, WardDTO{
			Id: w.Id, Name: w.Name, ZoneType: w.ZoneType, Machines: machinesByWard[w.Id],
		})
	}

	shiftDTOs := make([]ShiftDTO, 0, len(shifts))
	for _, s := range shifts {
		shiftDTOs = append(shiftDTOs, ShiftDTO{Id: s.Id, Name: s.Name, Code: s.ShiftCode, Sort: s.Sort})
	}

	return &WeekBoard{
		WeekStart: mon.Format("2006-01-02"),
		WeekEnd:   sun.Format("2006-01-02"),
		Dates:     dates,
		Shifts:    shiftDTOs,
		Wards:     wardDTOs,
	}, nil
}
