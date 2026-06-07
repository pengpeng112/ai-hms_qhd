package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

// ===================== Board 数据结构 =====================

type ScheduleBoard struct {
	StartDate       time.Time                          `json:"startDate"`
	EndDate         time.Time                          `json:"endDate"`
	Wards           []models.Ward                      `json:"wards"`
	WardExts        map[int64]models.WardExt           `json:"wardExts"`
	Beds            []models.Bed                       `json:"beds"`
	BedMachineExts  map[int64]models.BedMachineExt     `json:"bedMachineExts"`
	Shifts          []models.Shift                     `json:"shifts"`
	Profiles        map[int64]models.PatientProfile    `json:"profiles"`
	Patients        map[int64]ScheduleBoardPatient     `json:"patients"`
	Occupancies     map[int64][]ScheduleBoardOccupancy `json:"-"`
	CalendarEntries map[string]ScheduleBoardCalendar   `json:"calendarEntries"`
	Outages         []models.MachineOutage             `json:"outages"`
	Issues          []PrecheckIssue                    `json:"issues"`
}

type ScheduleBoardPatient struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type ScheduleBoardOccupancy struct {
	PatientShiftId int64     `json:"patientShiftId"`
	PatientId      int64     `json:"patientId"`
	Date           time.Time `json:"date"`
	ShiftId        int64     `json:"shiftId"`
	BedId          int64     `json:"bedId"`
	WardId         int64     `json:"wardId"`
	Status         int       `json:"status"`
	RuleStatus     int16     `json:"ruleStatus"`
	DialysisMode   string    `json:"dialysisMode"`
}

type ScheduleBoardCalendar struct {
	Calendar  models.Calendar `json:"calendar"`
	OpenWards []int64         `json:"openWards"`
	OpenBeds  []int64         `json:"openBeds"`
}

type PrecheckIssue struct {
	Type      string `json:"type"`
	Severity  int16  `json:"severity"`
	Detail    string `json:"detail"`
	PatientId *int64 `json:"patientId,omitempty"`
	BedId     *int64 `json:"bedId,omitempty"`
	WardId    *int64 `json:"wardId,omitempty"`
	ShiftId   *int64 `json:"shiftId,omitempty"`
	Date      string `json:"date,omitempty"`
	Suggested string `json:"suggested,omitempty"`
}

const (
	SeverityInfo     int16 = 10
	SeverityWarning  int16 = 20
	SeverityCritical int16 = 30
)

// ===================== Board 加载服务 =====================

type ScheduleBoardService struct {
	db *gorm.DB
}

func NewScheduleBoardService() *ScheduleBoardService {
	return &ScheduleBoardService{db: database.GetDB()}
}

func (s *ScheduleBoardService) LoadBoard(tenantID int64, start, end time.Time) (*ScheduleBoard, error) {
	if tenantID <= 0 {
		return nil, errors.New("invalid tenant")
	}
	if start.IsZero() || end.IsZero() {
		return nil, errors.New("start and end date required")
	}
	if end.Before(start) {
		return nil, errors.New("end date must be after start date")
	}
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	endExclusive := end.AddDate(0, 0, 1)

	board := &ScheduleBoard{
		StartDate:       start,
		EndDate:         end,
		WardExts:        map[int64]models.WardExt{},
		BedMachineExts:  map[int64]models.BedMachineExt{},
		Profiles:        map[int64]models.PatientProfile{},
		Patients:        map[int64]ScheduleBoardPatient{},
		Occupancies:     map[int64][]ScheduleBoardOccupancy{},
		CalendarEntries: map[string]ScheduleBoardCalendar{},
	}

	var errs []error
	pushErr := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	// 1. 批量加载病区
	var wards []models.Ward
	if err := s.db.Where(`"TenantId" = ? AND "IsDisabled" = ?`, tenantID, false).Find(&wards).Error; err != nil {
		errs = append(errs, err)
	} else {
		board.Wards = wards
	}

	// 2. 加载病区扩展
	var wardExts []models.WardExt
	if err := s.db.Where(`"TenantId" = ?`, tenantID).Find(&wardExts).Error; err != nil {
		errs = append(errs, err)
	} else {
		for _, we := range wardExts {
			board.WardExts[we.WardId] = we
		}
	}

	// 3. 批量加载床位
	var beds []models.Bed
	if err := s.db.Where(`"TenantId" = ? AND "IsDisabled" = ?`, tenantID, false).Find(&beds).Error; err != nil {
		errs = append(errs, err)
	} else {
		board.Beds = beds
	}

	// 4. 加载床位机器扩展
	var bedExts []models.BedMachineExt
	if err := s.db.Where(`"TenantId" = ?`, tenantID).Find(&bedExts).Error; err != nil {
		errs = append(errs, err)
	} else {
		for _, be := range bedExts {
			board.BedMachineExts[be.BedId] = be
		}
	}

	// 5. 批量加载班次
	var shifts []models.Shift
	if err := s.db.Where(`"TenantId" = ? AND "IsDisabled" = ?`, tenantID, false).Find(&shifts).Error; err != nil {
		errs = append(errs, err)
	} else {
		board.Shifts = shifts
	}

	// 6. 批量加载患者 Profile
	var profiles []models.PatientProfile
	if err := s.db.Where(`"TenantId" = ?`, tenantID).Find(&profiles).Error; err != nil {
		errs = append(errs, err)
	} else {
		for _, p := range profiles {
			board.Profiles[p.PatientId] = p
		}
	}

	// 7. 批量加载患者基本信息
	var patients []models.Patient
	if err := s.db.Table("Register_PatientInfomation").Where(`"TenantId" = ?`, tenantID).Find(&patients).Error; err != nil {
		errs = append(errs, err)
	} else {
		for _, p := range patients {
			board.Patients[int64(p.ID)] = ScheduleBoardPatient{
				Id:   int64(p.ID),
				Name: p.Name,
			}
		}
	}

	// 8. 批量加载已有排班占用 (Schedule_PatientShift)
	var occupancies []models.PatientShift
	if err := s.db.Where(`"TenantId" = ? AND DATE("TreatmentTime") >= DATE(?) AND DATE("TreatmentTime") <= DATE(?)`,
		tenantID, start, end).Find(&occupancies).Error; err != nil {
		errs = append(errs, err)
	} else {
		psIDs := make([]int64, len(occupancies))
		for i, occ := range occupancies {
			psIDs[i] = occ.Id
		}

		// 9. 批量查询 PatientShiftExt
		var psExts []models.PatientShiftExt
		if len(psIDs) > 0 {
			if err := s.db.Where(`"TenantId" = ? AND "PatientShiftId" IN ?`, tenantID, psIDs).Find(&psExts).Error; err != nil {
				errs = append(errs, err)
			}
		}

		psExtMap := map[int64]models.PatientShiftExt{}
		for _, e := range psExts {
			psExtMap[e.PatientShiftId] = e
		}

		// 构建 bed->ward 反查
		bedWard := map[int64]int64{}
		for _, b := range board.Beds {
			if b.WardId != nil {
				bedWard[b.Id] = *b.WardId
			}
		}

		for _, occ := range occupancies {
			bedID := int64(0)
			if occ.BedId != nil {
				bedID = *occ.BedId
			}
			wardID := int64(0)
			if occ.WardId != nil {
				wardID = *occ.WardId
			}
			// WardId=0 时通过 BedId 反查
			if wardID == 0 && bedID > 0 {
				if w, ok := bedWard[bedID]; ok {
					wardID = w
				}
			}

			mode := "HD"
			ruleStatus := int16(10)
			if ext, ok := psExtMap[occ.Id]; ok {
				mode = ext.DialysisMode
				ruleStatus = ext.RuleStatus
			}

			board.Occupancies[occ.ShiftId] = append(board.Occupancies[occ.ShiftId], ScheduleBoardOccupancy{
				PatientShiftId: occ.Id,
				PatientId:      int64(occ.PatientId),
				Date:           occ.ScheduleDate,
				ShiftId:        occ.ShiftId,
				BedId:          bedID,
				WardId:         wardID,
				Status:         occ.Status,
				RuleStatus:     ruleStatus,
				DialysisMode:   mode,
			})
		}
	}

	// 10. 批量加载日历
	var cals []models.Calendar
	if err := s.db.Where(`"TenantId" = ? AND "CalDate" >= ? AND "CalDate" <= ?`, tenantID, start, end).Find(&cals).Error; err != nil {
		errs = append(errs, err)
	} else {
		calIDs := make([]int64, len(cals))
		for i, c := range cals {
			calIDs[i] = c.Id
		}

		// 加载日历开放病区
		var openWards []models.CalendarOpenWard
		if len(calIDs) > 0 {
			pushErr(s.db.Where(`"TenantId" = ? AND "CalendarId" IN ?`, tenantID, calIDs).Find(&openWards).Error)
		}
		calWards := map[int64][]int64{}
		for _, ow := range openWards {
			calWards[ow.CalendarId] = append(calWards[ow.CalendarId], ow.WardId)
		}

		// 加载日历开放床位
		var openBeds []models.CalendarOpenBed
		if len(calIDs) > 0 {
			pushErr(s.db.Where(`"TenantId" = ? AND "CalendarId" IN ?`, tenantID, calIDs).Find(&openBeds).Error)
		}
		calBeds := map[int64][]int64{}
		for _, ob := range openBeds {
			calBeds[ob.CalendarId] = append(calBeds[ob.CalendarId], ob.BedId)
		}

		for _, c := range cals {
			k := c.CalDate.Format("2006-01-02")
			board.CalendarEntries[k] = ScheduleBoardCalendar{
				Calendar:  c,
				OpenWards: calWards[c.Id],
				OpenBeds:  calBeds[c.Id],
			}
		}
	}

	// 11. 批量加载停机（与 [start, endExclusive) 相交即可）
	var outages []models.MachineOutage
	if err := s.db.Where(`"TenantId" = ? AND "StartAt" < ? AND ("EndAt" IS NULL OR "EndAt" > ?)`,
		tenantID, endExclusive, start).Find(&outages).Error; err != nil {
		errs = append(errs, err)
	} else {
		board.Outages = outages
	}

	if len(errs) > 0 {
		return board, errs[0]
	}

	return board, nil
}

// ===================== 预检规则 =====================

func modeSupports(supportedModes string, target string) bool {
	modes := normalizeModes(supportedModes)
	for _, m := range modes {
		if m == target {
			return true
		}
	}
	return false
}

func (s *ScheduleBoardService) RunPrecheck(board *ScheduleBoard) []PrecheckIssue {
	var issues []PrecheckIssue

	profiledPatientIDs := map[int64]bool{}
	for pid := range board.Profiles {
		profiledPatientIDs[pid] = true
	}

	bedIDsWithMachine := map[int64]bool{}
	for bid := range board.BedMachineExts {
		bedIDsWithMachine[bid] = true
	}

	// 1. 已排班患者缺 Profile
	for _, dayOccs := range board.Occupancies {
		for _, occ := range dayOccs {
			if !profiledPatientIDs[occ.PatientId] {
				issues = append(issues, PrecheckIssue{
					Type:      "MISSING_PROFILE",
					Severity:  SeverityWarning,
					Detail:    "患者无排班骨架配置",
					PatientId: &occ.PatientId,
					Date:      occ.Date.Format("2006-01-02"),
					ShiftId:   &occ.ShiftId,
				})
			}
		}
	}

	// 2. 床缺机器类型
	for _, bed := range board.Beds {
		if !bedIDsWithMachine[bed.Id] {
			issues = append(issues, PrecheckIssue{
				Type:     "BED_NO_MACHINE_TYPE",
				Severity: SeverityWarning,
				Detail:   "床位未配置机器类型",
				BedId:    &bed.Id,
			})
		}
	}

	// 3. 机器 SupportedModes 与 MachineType 不一致
	for bid, ext := range board.BedMachineExts {
		if _, err := validateMachineTypeAndModes(ext.MachineType, ext.SupportedModes); err != nil {
			issues = append(issues, PrecheckIssue{
				Type:     "MACHINE_MODE_MISMATCH",
				Severity: SeverityCritical,
				Detail:   "机器模式配置不一致: " + err.Error(),
				BedId:    &bid,
			})
		}
	}

	// 4. Profile ZoneTag 不在 A/B/C
	for pid, prof := range board.Profiles {
		if prof.ZoneTag != "A" && prof.ZoneTag != "B" && prof.ZoneTag != "C" {
			issues = append(issues, PrecheckIssue{
				Type:      "INVALID_ZONE_TAG",
				Severity:  SeverityWarning,
				Detail:    "患者分区标签无效: " + prof.ZoneTag,
				PatientId: &pid,
			})
		}
	}

	// 5. Profile HomeWardId 不存在或缺少 WardExt
	wardIDs := map[int64]bool{}
	for _, w := range board.Wards {
		wardIDs[w.Id] = true
	}
	for pid, prof := range board.Profiles {
		if prof.HomeWardId != nil && *prof.HomeWardId != 0 {
			if !wardIDs[*prof.HomeWardId] {
				issues = append(issues, PrecheckIssue{
					Type:      "HOME_WARD_NOT_FOUND",
					Severity:  SeverityWarning,
					Detail:    "HomeWardId 对应的病区不存在",
					PatientId: &pid,
					WardId:    prof.HomeWardId,
				})
			} else if _, ok := board.WardExts[*prof.HomeWardId]; !ok {
				issues = append(issues, PrecheckIssue{
					Type:      "HOME_WARD_NO_EXT",
					Severity:  SeverityWarning,
					Detail:    "HomeWardId 存在但缺少病区扩展配置 (ZoneType)",
					PatientId: &pid,
					WardId:    prof.HomeWardId,
				})
			}
		}
	}

	// 6. Profile FixedHdBedId / FixedHdfBedId 不存在
	for pid, prof := range board.Profiles {
		if prof.FixedHdBedId != nil && *prof.FixedHdBedId != 0 && !bedIDsWithMachine[*prof.FixedHdBedId] {
			issues = append(issues, PrecheckIssue{
				Type:      "FIXED_HD_BED_NOT_FOUND",
				Severity:  SeverityWarning,
				Detail:    "固定 HD 床位未配置或不存在",
				PatientId: &pid,
				BedId:     prof.FixedHdBedId,
			})
		}
		if prof.FixedHdfBedId != nil && *prof.FixedHdfBedId != 0 && !bedIDsWithMachine[*prof.FixedHdfBedId] {
			issues = append(issues, PrecheckIssue{
				Type:      "FIXED_HDF_BED_NOT_FOUND",
				Severity:  SeverityWarning,
				Detail:    "固定 HDF 床位未配置或不存在",
				PatientId: &pid,
				BedId:     prof.FixedHdfBedId,
			})
		}
	}

	// 7. 固定床位不支持对应模式（看 SupportedModes）
	for pid, prof := range board.Profiles {
		if prof.FixedHdBedId != nil && *prof.FixedHdBedId != 0 {
			if ext, ok := board.BedMachineExts[*prof.FixedHdBedId]; ok {
				if !modeSupports(ext.SupportedModes, "HD") {
					issues = append(issues, PrecheckIssue{
						Type:      "FIXED_BED_MODE_MISMATCH",
						Severity:  SeverityWarning,
						Detail:    "固定 HD 床位 SupportedModes 不支持 HD",
						PatientId: &pid,
						BedId:     prof.FixedHdBedId,
					})
				}
			}
		}
		if prof.FixedHdfBedId != nil && *prof.FixedHdfBedId != 0 {
			if ext, ok := board.BedMachineExts[*prof.FixedHdfBedId]; ok {
				if !modeSupports(ext.SupportedModes, "HDF") {
					issues = append(issues, PrecheckIssue{
						Type:      "FIXED_HDF_BED_MODE_MISMATCH",
						Severity:  SeverityWarning,
						Detail:    "固定 HDF 床位 SupportedModes 不支持 HDF",
						PatientId: &pid,
						BedId:     prof.FixedHdfBedId,
					})
				}
			}
		}
	}

	// 8. HDF 患者但没有可用 HDF 机器
	hasHDFMachine := false
	for _, ext := range board.BedMachineExts {
		if !ext.IsDisabled && modeSupports(ext.SupportedModes, "HDF") {
			hasHDFMachine = true
			break
		}
	}
	for pid, prof := range board.Profiles {
		if prof.HdfEnabled {
			if !hasHDFMachine {
				issues = append(issues, PrecheckIssue{
					Type:      "HDF_PATIENT_NO_MACHINE",
					Severity:  SeverityWarning,
					Detail:    "患者需要 HDF 但没有可用 HDF 机器",
					PatientId: &pid,
				})
			}
		}
	}

	// 8b. CRRT 机器配置预检
	hasCRRTMachine := false
	for _, ext := range board.BedMachineExts {
		if !ext.IsDisabled && modeSupports(ext.SupportedModes, "CRRT") {
			hasCRRTMachine = true
			break
		}
	}
	if !hasCRRTMachine {
		issues = append(issues, PrecheckIssue{
			Type:     "CRRT_MACHINE_NOT_CONFIGURED",
			Severity: SeverityWarning,
			Detail:   "没有可用的 CRRT 机器配置",
		})
	}
	// 8c. CRRT 病区配置预检
	if len(board.WardExts) == 0 {
		issues = append(issues, PrecheckIssue{
			Type:     "CRRT_WARD_NOT_CONFIGURED",
			Severity: SeverityWarning,
			Detail:   "没有病区扩展配置，无法确认 CRRT 承载病区",
		})
	}

	// 9. Calendar 非透析日容量计算标记
	for k, cal := range board.CalendarEntries {
		if !cal.Calendar.IsDialysisDay {
			issues = append(issues, PrecheckIssue{
				Type:     "NON_DIALYSIS_DAY",
				Severity: SeverityInfo,
				Detail:   "非透析日: 该日不参与容量计算",
				Date:     k,
			})
		}
	}

	// 10. 停机覆盖的 Bed 按日期+班次维度标记
	for _, outage := range board.Outages {
		issues = append(issues, PrecheckIssue{
			Type:     "MACHINE_OUTAGE",
			Severity: SeverityInfo,
			Detail:   "设备停机: " + outage.Reason,
			BedId:    &outage.BedId,
			Date:     outage.StartAt.Format("2006-01-02"),
			ShiftId:  outage.ShiftId,
		})
	}

	// 11. 排班占用冲突检测（不修复，只报告）
	type slotKey struct {
		date    int64
		shiftId int64
		bedId   int64
	}
	bedSlots := map[slotKey][]int64{}
	patientSlots := map[slotKey][]int64{}

	for _, dayOccs := range board.Occupancies {
		for _, occ := range dayOccs {
			sk := slotKey{
				date:    occ.Date.Unix() / 86400,
				shiftId: occ.ShiftId,
				bedId:   occ.BedId,
			}
			if occ.BedId > 0 {
				bedSlots[sk] = append(bedSlots[sk], occ.PatientShiftId)
			}
			psk := slotKey{
				date:    occ.Date.Unix() / 86400,
				shiftId: occ.ShiftId,
				bedId:   occ.PatientId,
			}
			patientSlots[psk] = append(patientSlots[psk], occ.PatientShiftId)
		}
	}
	for _, occs := range bedSlots {
		if len(occs) > 1 {
			issues = append(issues, PrecheckIssue{
				Type:     "BED_CONFLICT",
				Severity: SeverityCritical,
				Detail:   "同床同日同班存在多条排班占用",
			})
		}
	}
	for _, occs := range patientSlots {
		if len(occs) > 1 {
			issues = append(issues, PrecheckIssue{
				Type:     "PATIENT_CONFLICT",
				Severity: SeverityCritical,
				Detail:   "同患者同日同班存在多条排班占用",
			})
		}
	}

	return issues
}

// ===================== 容量/余位计算 =====================

type BoardCapacitySlot struct {
	Date      string `json:"date"`
	ShiftId   int64  `json:"shiftId"`
	WardId    int64  `json:"wardId"`
	ZoneType  string `json:"zoneType"`
	Capacity  int    `json:"capacity"`
	Occupied  int    `json:"occupied"`
	Available int    `json:"available"`
}

func (s *ScheduleBoardService) ComputeCapacity(board *ScheduleBoard) []BoardCapacitySlot {
	dates := []time.Time{}
	for d := board.StartDate; !d.After(board.EndDate); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
	}

	// 构建病区-床位归属
	bedWard := map[int64]int64{}
	for _, b := range board.Beds {
		if b.WardId != nil {
			bedWard[b.Id] = *b.WardId
		}
	}

	// 构建病区-分区映射
	wardZone := map[int64]string{}
	for wid, we := range board.WardExts {
		wardZone[wid] = we.ZoneType
	}

	// 构建日历日期对应的开放病区/床位
	calOpenWards := map[string]map[int64]bool{}
	calOpenBeds := map[string]map[int64]bool{}
	skipDates := map[string]bool{}
	for k, cal := range board.CalendarEntries {
		if !cal.Calendar.IsDialysisDay {
			skipDates[k] = true
		}
		if len(cal.OpenWards) > 0 {
			ow := map[int64]bool{}
			for _, wid := range cal.OpenWards {
				ow[wid] = true
			}
			calOpenWards[k] = ow
		}
		if len(cal.OpenBeds) > 0 {
			ob := map[int64]bool{}
			for _, bid := range cal.OpenBeds {
				ob[bid] = true
			}
			calOpenBeds[k] = ob
		}
	}

	// 停机覆盖的 Bed (日期+班次维度) —— [StartAt, EndAt) 日期粒度，裁剪到 board 范围
	outageBeds := map[string]map[int64]bool{}
	boardEndClip := board.EndDate.AddDate(0, 0, 1)
	for _, outage := range board.Outages {
		outStart := outage.StartAt
		if outStart.Before(board.StartDate) {
			outStart = board.StartDate
		}
		outEnd := outage.EndAt
		if outEnd == nil {
			outEnd = &boardEndClip
		} else if outEnd.After(boardEndClip) {
			outEnd = &boardEndClip
		}
		if !outStart.Before(*outEnd) {
			continue
		}

		startDay := time.Date(outStart.Year(), outStart.Month(), outStart.Day(), 0, 0, 0, 0, outStart.Location())
		endDay := time.Date(outEnd.Year(), outEnd.Month(), outEnd.Day()+1, 0, 0, 0, 0, outEnd.Location())
		if outEnd.Hour() == 0 && outEnd.Minute() == 0 && outEnd.Second() == 0 && outEnd.Nanosecond() == 0 {
			endDay = time.Date(outEnd.Year(), outEnd.Month(), outEnd.Day(), 0, 0, 0, 0, outEnd.Location())
		}
		for d := startDay; d.Before(endDay); d = d.AddDate(0, 0, 1) {
			shiftI64 := int64(0)
			if outage.ShiftId != nil {
				shiftI64 = *outage.ShiftId
			}
			dk := d.Format("2006-01-02") + fmt.Sprintf("_%d", shiftI64)
			if outageBeds[dk] == nil {
				outageBeds[dk] = map[int64]bool{}
			}
			outageBeds[dk][outage.BedId] = true
		}
	}

	// 已有占用：按 date+shiftId+wardId 统计
	type occKey struct {
		date    string
		shiftId int64
		wardId  int64
	}
	occByWard := map[occKey]int{}
	for _, dayOccs := range board.Occupancies {
		for _, occ := range dayOccs {
			dk := occ.Date.Format("2006-01-02")
			wid := occ.WardId
			if wid == 0 && occ.BedId > 0 {
				if w, ok := bedWard[occ.BedId]; ok {
					wid = w
				}
			}
			oKey := occKey{date: dk, shiftId: occ.ShiftId, wardId: wid}
			occByWard[oKey]++
		}
	}

	slots := []BoardCapacitySlot{}
	for _, shift := range board.Shifts {
		for _, date := range dates {
			dk := date.Format("2006-01-02")
			if skipDates[dk] {
				continue
			}

			// 按病区分组容量
			wardCapacity := map[int64]int{}
			for _, bed := range board.Beds {
				bid := bed.Id
				if bed.IsDisabled {
					continue
				}

				// 扩展层停用
				if ext, ok := board.BedMachineExts[bid]; ok && ext.IsDisabled {
					continue
				}
				// 无机器类型 → 不计入容量
				if _, ok := board.BedMachineExts[bid]; !ok {
					continue
				}

				// 日历开放床位过滤
				if ob, hasOB := calOpenBeds[dk]; hasOB && !ob[bid] {
					continue
				}

				// 停机过滤
				keyShift := dk + fmt.Sprintf("_%d", shift.Id)
				if ob2, hasOB2 := outageBeds[keyShift]; hasOB2 && ob2[bid] {
					continue
				}
				if obAll, hasAll := outageBeds[dk+"_0"]; hasAll && obAll[bid] {
					continue
				}

				wid := int64(0)
				if bw, ok := bedWard[bid]; ok {
					wid = bw
				}

				// 日历开放病区过滤
				if ow, hasOW := calOpenWards[dk]; hasOW && !ow[wid] {
					continue
				}

				wardCapacity[wid]++
			}

			for wid, cap := range wardCapacity {
				oKey := occKey{date: dk, shiftId: shift.Id, wardId: wid}
				occ := occByWard[oKey]

				zone := ""
				if z, ok := wardZone[wid]; ok {
					zone = z
				}
				slots = append(slots, BoardCapacitySlot{
					Date:      dk,
					ShiftId:   shift.Id,
					WardId:    wid,
					ZoneType:  zone,
					Capacity:  cap,
					Occupied:  occ,
					Available: cap - occ,
				})
			}
		}
	}

	return slots
}

func (s *ScheduleBoardService) LoadBoardWithPrecheck(tenantID int64, start, end time.Time) (*ScheduleBoard, []PrecheckIssue, []BoardCapacitySlot, int64, error) {
	t0 := time.Now()

	board, err := s.LoadBoard(tenantID, start, end)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	issues := s.RunPrecheck(board)
	board.Issues = issues

	slots := s.ComputeCapacity(board)

	elapsed := time.Since(t0).Milliseconds()
	log.Printf("[BOARD] loaded board for tenant %d, range %s~%s: %d wards, %d beds, %d occupancies, %d issues, %d ms",
		tenantID, start.Format("2006-01-02"), end.Format("2006-01-02"),
		len(board.Wards), len(board.Beds), countOccupancies(board), len(issues), elapsed)

	return board, issues, slots, elapsed, nil
}

func countOccupancies(board *ScheduleBoard) int {
	c := 0
	for _, occs := range board.Occupancies {
		c += len(occs)
	}
	return c
}
