package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/schedule_engine"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"gorm.io/gorm"
)

// ScheduleGenerateService 排班生成服务
type ScheduleGenerateService struct {
	db *gorm.DB
}

func NewScheduleGenerateService() *ScheduleGenerateService {
	return &ScheduleGenerateService{db: database.GetDB()}
}

// GenerateScheduleRequest 生成请求
type GenerateScheduleRequest struct {
	TemplateID *int64 `json:"templateId"`
	StartDate  string `json:"startDate" binding:"required"`
	Weeks      int    `json:"weeks" binding:"required"`
	WardID     *int64 `json:"wardId"`
}

// GenerateScheduleResponse 生成响应
type GenerateScheduleResponse struct {
	StartDate      string `json:"startDate"`
	Weeks          int    `json:"weeks"`
	DialysisDays   int    `json:"dialysisDays"`
	Drafts         int    `json:"drafts"`
	Conflicts      int    `json:"conflicts"`
	ParityAssigned int    `json:"parityAssigned"`
}

// Generate 执行排班生成
func (s *ScheduleGenerateService) Generate(tenantID, creatorID int64, req GenerateScheduleRequest) (*GenerateScheduleResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	startDate, err := ParseScheduleDate(req.StartDate)
	if err != nil {
		return nil, errors.New("startDate格式应为YYYY-MM-DD")
	}

	// 1) 获取奇偶周锚点
	anchor := getAnchorMonday(s.db, tenantID)

	// 2) 构建 Board
	board, err := s.buildBoard(tenantID, anchor, startDate, req.Weeks)
	if err != nil {
		return nil, err
	}

	// 3) 构建 SessionItems(从模板项)
	items, err := s.buildSessionItems(tenantID, req.TemplateID, req.WardID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("没有可用的排班骨架数据，请先创建排班模板或在患者管理中配置排班骨架(PatientProfile)")
	}

	// 4) HDF奇偶周分配
	eng := schedule_engine.NewEngine(board)
	eng.SpillHorizonDays = schedule_engine.DefaultSpillHorizon

	parityAssigned := 0
	if parityResult := s.assignHdfParity(tenantID, items); parityResult > 0 {
		parityAssigned = parityResult
	}

	// 5) 展开透析日并运行引擎
	dates := schedule_engine.ExpandDialysisDates(startDate, req.Weeks, func(t time.Time) bool {
		return !board.NonDialysisDays[t.Format("2006-01-02")]
	})

	result := eng.Generate(items, dates)

	// 6) 持久化草稿和冲突
	draftCount, conflictCount, err := s.persistResults(tenantID, creatorID, result)
	if err != nil {
		return nil, err
	}

	return &GenerateScheduleResponse{
		StartDate:      result.StartDate,
		Weeks:          req.Weeks,
		DialysisDays:   result.DialysisDays,
		Drafts:         draftCount,
		Conflicts:      conflictCount,
		ParityAssigned: parityAssigned,
	}, nil
}

// buildBoard 构建引擎 Board
func (s *ScheduleGenerateService) buildBoard(tenantID int64, anchor time.Time, startDate time.Time, weeks int) (*schedule_engine.Board, error) {
	endDate := startDate.AddDate(0, 0, weeks*7)
	board := &schedule_engine.Board{
		TenantID:        tenantID,
		Anchor:          anchor,
		StartDate:       startDate,
		EndDate:         endDate,
		Occupied:        map[int64][]schedule_engine.Occupancy{},
		NonDialysisDays: map[string]bool{},
		Outages:         map[int64]map[string]bool{},
	}

	// 病区
	var wards []models.Ward
	if err := s.db.Where(`"TenantId" = ? AND "IsDisabled" = ?`, tenantID, false).Find(&wards).Error; err != nil {
		return nil, err
	}
	// 病区扩展(ZoneType)
	var wardExts []models.WardExt
	s.db.Where(`"TenantId" = ?`, tenantID).Find(&wardExts)
	zoneMap := map[int64]string{}
	for _, we := range wardExts {
		zoneMap[we.WardId] = we.ZoneType
	}
	for _, w := range wards {
		zone := zoneMap[w.Id]
		if zone == "" {
			zone = "A"
		}
		board.Wards = append(board.Wards, schedule_engine.WardInfo{
			ID:       w.Id,
			Name:     w.Name,
			ZoneType: zone,
			WardID:   w.Id,
		})
	}

	// 床位+机器
	var beds []models.Bed
	s.db.Where(`"TenantId" = ? AND "IsDisabled" = ?`, tenantID, false).Find(&beds)
	var bedExts []models.BedMachineExt
	s.db.Where(`"TenantId" = ?`, tenantID).Find(&bedExts)
	extMap := map[int64]models.BedMachineExt{}
	for _, be := range bedExts {
		extMap[be.BedId] = be
	}
	for _, b := range beds {
		mt := "HD"
		modes := []string{"HD"}
		pos := 0
		wardID := int64(0)
		if b.WardId != nil {
			wardID = *b.WardId
		}
		if ext, ok := extMap[b.Id]; ok {
			mt = ext.MachineType
			modes = normalizeModes(ext.SupportedModes)
			if ext.PositionIndex > 0 {
				pos = ext.PositionIndex
			}
		}
		board.Beds = append(board.Beds, schedule_engine.BedInfo{
			ID:             b.Id,
			Name:           b.Name,
			WardID:         wardID,
			MachineType:    mt,
			SupportedModes: modes,
			PositionIndex:  pos,
			IsDisabled:     false,
		})
	}

	// 班次
	var shifts []models.Shift
	s.db.Where(`"TenantId" = ? AND "IsDisabled" = ?`, tenantID, false).Find(&shifts)
	for _, sh := range shifts {
		board.Shifts = append(board.Shifts, schedule_engine.ShiftInfo{
			ID:        sh.Id,
			Name:      sh.Name,
			StartTime: sh.StartTime,
			EndTime:   sh.EndTime,
			Sort:      sh.Sort,
		})
	}

	// 已有占用
	var occs []models.PatientShift
	s.db.Where(`"TenantId" = ? AND "TreatmentTime" >= ? AND "TreatmentTime" < ? AND "Status" NOT IN ?`,
		tenantID, startDate.Format("2006-01-02")+" 00:00:00", endDate.Format("2006-01-02")+" 00:00:00",
		[]int{
			MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusCancelled),
			MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusUserCancelled),
			MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusTransferred),
		},
	).Find(&occs)
	for _, occ := range occs {
		bedID := int64(0)
		if occ.BedId != nil {
			bedID = *occ.BedId
		}
		wardID := int64(0)
		if occ.WardId != nil {
			wardID = *occ.WardId
		}
		board.Occupied[bedID] = append(board.Occupied[bedID], schedule_engine.Occupancy{
			PatientShiftID: occ.Id,
			PatientID:      int64(occ.PatientId),
			Date:           occ.ScheduleDate,
			ShiftID:        occ.ShiftId,
			BedID:          bedID,
			WardID:         wardID,
			Status:         occ.Status,
		})
	}

	// 日历(非透析日)
	var cals []models.Calendar
	s.db.Where(`"TenantId" = ? AND "CalDate" >= ? AND "CalDate" <= ?`, tenantID, startDate, endDate).Find(&cals)
	for _, cal := range cals {
		if !cal.IsDialysisDay {
			board.NonDialysisDays[cal.CalDate.Format("2006-01-02")] = true
		}
	}

	// 停机
	var outages []models.MachineOutage
	s.db.Where(`"TenantId" = ? AND "StartAt" < ? AND ("EndAt" IS NULL OR "EndAt" > ?)`,
		tenantID, endDate, startDate).Find(&outages)
	for _, o := range outages {
		if board.Outages[o.BedId] == nil {
			board.Outages[o.BedId] = map[string]bool{}
		}
		end := o.StartAt.AddDate(0, 0, 1)
		if o.EndAt != nil {
			end = *o.EndAt
		}
		for d := o.StartAt; !d.After(end); d = d.AddDate(0, 0, 1) {
			board.Outages[o.BedId][d.Format("2006-01-02")] = true
		}
	}

	return board, nil
}

// buildSessionItems 从模板项或PatientProfile构建待排SessionItem
// 若指定templateID则只用该模板项，否则从所有PatientProfile加载
func (s *ScheduleGenerateService) buildSessionItems(tenantID int64, templateID, wardID *int64) ([]schedule_engine.SessionItem, error) {
	// 指定模板时从模板项加载
	if templateID != nil && *templateID > 0 {
		return s.buildFromTemplate(tenantID, *templateID, wardID)
	}

	// 无模板指定时从PatientProfile加载，若为空则兜底从Plan_PatientPlan读取
	var items []schedule_engine.SessionItem

	var profiles []models.PatientProfile
	s.db.Where(`"TenantId" = ?`, tenantID).Find(&profiles)

	if len(profiles) == 0 {
		// 兜底: 从老表 Plan_PatientPlan 读取患者骨架
		return s.buildFromLegacyPlan(tenantID, wardID)
	}

	for _, p := range profiles {
		if p.FreqPattern == schedule_engine.FreqTemporary {
			continue
		}
		if p.HomeWardId == nil || p.ShiftId == nil {
			continue
		}
		items = append(items, schedule_engine.SessionItem{
			PatientID:     p.PatientId,
			FreqPattern:   p.FreqPattern,
			FixedHdBedID:  p.FixedHdBedId,
			FixedHdfBedID: p.FixedHdfBedId,
			HdfEnabled:    p.HdfEnabled,
			HdfWeekday:    p.HdfWeekday,
			HdfWeekParity: p.HdfWeekParity,
			WardID:        p.HomeWardId,
			ShiftID:       p.ShiftId,
			PatientPlanID: &p.PatientId,
		})
	}
	if len(items) == 0 {
		return nil, errors.New("没有可用的患者排班骨架データ，请先创建排班模板或在患者管理中配置排班骨架")
	}
	return items, nil
}

// buildFromLegacyPlan 从老表 Plan_PatientPlan 兜底构建 SessionItem
func (s *ScheduleGenerateService) buildFromLegacyPlan(tenantID int64, wardID *int64) ([]schedule_engine.SessionItem, error) {
	type planRow struct {
		ID                int64  `gorm:"column:Id"`
		PatientID         int64  `gorm:"column:PatientId"`
		DialysisMethod    string `gorm:"column:DialysisMethod"`
		OddWeekFrequency  int    `gorm:"column:OddWeekFrequency"`
		EvenWeekFrequency int    `gorm:"column:EvenWeekFrequency"`
	}
	var rows []planRow
	q := s.db.Table(`"Plan_PatientPlan"`).
		Select(`"Id", "PatientId", COALESCE("DialysisMethod", 'HD') AS "DialysisMethod",
			COALESCE("OddWeekFrequency", 0) AS "OddWeekFrequency",
			COALESCE("EvenWeekFrequency", 0) AS "EvenWeekFrequency"`).
		Where(`"TenantId" = ? AND "IsDisabled" = false`, tenantID)
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}

	// 从本工具周视图中获取的班次默认取第一个活跃班次=上午班(shift 1)
	var shifts []models.Shift
	s.db.Where(`"TenantId" = ? AND "IsDisabled" = ?`, tenantID, false).Order(`"Sort" ASC`).Limit(1).Find(&shifts)
	defaultShiftID := int64(0)
	if len(shifts) > 0 {
		defaultShiftID = shifts[0].Id
	}

	// 取各患者所在病区
	type bedWard struct {
		BedID  int64 `gorm:"column:BedId"`
		WardID int64 `gorm:"column:WardId"`
	}
	var bw []bedWard
	s.db.Table(`"Schedule_Bed"`).Select(`"Id" AS "BedId", "WardId"`).
		Where(`"TenantId" = ? AND "IsDisabled" = false`, tenantID).Find(&bw)
	wardByBed := map[int64]int64{}
	for _, x := range bw {
		wardByBed[x.BedID] = x.WardID
	}

	// 取各个病区首个床位作为默认
	type wardBed struct{ WardID, BedID int64 }
	var wb []wardBed
	s.db.Table(`"Schedule_Bed"`).Select(`"WardId", MIN("Id") AS "BedId"`).
		Where(`"TenantId" = ? AND "IsDisabled" = false`, tenantID).
		Group(`"WardId"`).Find(&wb)
	defaultBed := map[int64]int64{}
	for _, x := range wb {
		defaultBed[x.WardID] = x.BedID
	}

	var items []schedule_engine.SessionItem
	for _, r := range rows {
		freq := freqFromPlan(r.OddWeekFrequency, r.EvenWeekFrequency)
		if freq == schedule_engine.FreqTemporary {
			continue
		}

		wid := int64(1)
		if wardID != nil && *wardID > 0 {
			wid = *wardID
		}
		planID := r.ID
		items = append(items, schedule_engine.SessionItem{
			PatientID:    r.PatientID,
			FreqPattern:  freq,
			WardID:       &wid,
			ShiftID:      &defaultShiftID,
			PatientPlanID: &planID,
		})
	}
	return items, nil
}

func freqFromPlan(odd, even int) int16 {
	switch odd {
	case 6:
		return schedule_engine.FreqMonWedFri
	case 5:
		return schedule_engine.FreqTwoPerWk
	case 3:
		return schedule_engine.FreqTwoPerWk
	case 1:
		return schedule_engine.FreqOnePerWk
	}
	if even == 6 {
		return schedule_engine.FreqTueThuSat
	}
	if even == 3 || odd == 3 {
		return schedule_engine.FreqTwoPerWk
	}
	return schedule_engine.FreqTemporary
}

func (s *ScheduleGenerateService) buildFromTemplate(tenantID, templateID int64, wardID *int64) ([]schedule_engine.SessionItem, error) {
	var tmpl models.ScheduleTemplate
	if err := s.db.Where(`"TenantId" = ? AND "Id" = ? AND "IsActive" = ?`, tenantID, templateID, true).
		First(&tmpl).Error; err != nil {
		return nil, errors.New("模板不存在或已禁用")
	}

	var tmplItems []models.ScheduleTemplateItem
	q := s.db.Where(`"TenantId" = ? AND "TemplateId" = ?`, tenantID, templateID)
	if wardID != nil && *wardID > 0 {
		q = q.Where(`"WardId" = ?`, *wardID)
	}
	if err := q.Find(&tmplItems).Error; err != nil {
		return nil, err
	}

	var items []schedule_engine.SessionItem
	for _, it := range tmplItems {
		if it.FreqPattern == schedule_engine.FreqTemporary {
			continue
		}
		if it.WardId == nil || it.ShiftId == nil {
			continue
		}
		tid := it.Id
		pi := int64(0)
		items = append(items, schedule_engine.SessionItem{
			PatientID:      it.PatientId,
			FreqPattern:    it.FreqPattern,
			FixedHdBedID:   it.FixedHdBedId,
			FixedHdfBedID:  it.FixedHdfBedId,
			HdfEnabled:     it.HdfEnabled,
			HdfWeekday:     it.HdfWeekday,
			HdfWeekParity:  it.HdfWeekParity,
			TemplateItemID: &tid,
			WardID:         it.WardId,
			ShiftID:        it.ShiftId,
			PatientPlanID:  &pi,
		})
	}
	return items, nil
}

// assignHdfParity HDF奇偶周分配并写回
func (s *ScheduleGenerateService) assignHdfParity(tenantID int64, items []schedule_engine.SessionItem) int {
	// 从PatientProfile重新加载以获取当前值
	profiles := s.loadProfiles(tenantID)
	assignments := schedule_engine.AssignHdfWeekParity(profiles)
	for _, a := range assignments {
		// 写回 PatientProfile
		s.db.Model(&models.PatientProfile{}).
			Where(`"TenantId" = ? AND "PatientId" = ?`, tenantID, a.PatientID).
			Update("HdfWeekParity", a.Parity)
		// 写回模板项
		s.db.Model(&models.ScheduleTemplateItem{}).
			Where(`"TenantId" = ? AND "PatientId" = ?`, tenantID, a.PatientID).
			Update("HdfWeekParity", a.Parity)
	}
	return len(assignments)
}

func (s *ScheduleGenerateService) loadProfiles(tenantID int64) []schedule_engine.PatientProfile {
	var profs []models.PatientProfile
	s.db.Where(`"TenantId" = ?`, tenantID).Find(&profs)
	result := make([]schedule_engine.PatientProfile, len(profs))
	for i, p := range profs {
		result[i] = schedule_engine.PatientProfile{
			PatientID:     p.PatientId,
			ZoneTag:       p.ZoneTag,
			HomeWardID:    p.HomeWardId,
			ShiftID:       p.ShiftId,
			FreqPattern:   p.FreqPattern,
			DefaultMode:   p.DefaultMode,
			HdfEnabled:    p.HdfEnabled,
			HdfWeekday:    p.HdfWeekday,
			HdfWeekParity: p.HdfWeekParity,
			FixedHdBedID:  p.FixedHdBedId,
			FixedHdfBedID: p.FixedHdfBedId,
		}
	}
	return result
}

// persistResults 持久化草稿和冲突
func (s *ScheduleGenerateService) persistResults(tenantID, creatorID int64, result schedule_engine.GenerateResult) (drafts, conflicts int, err error) {
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		// 保存草稿
		for _, dr := range result.Drafts {
			bedID := dr.BedID
			wardID := dr.WardID
			planID := int64(0)
			if dr.PatientPlanID != nil {
				planID = *dr.PatientPlanID
			}
			ps := models.PatientShift{
				TenantId:      tenantID,
				PatientId:     modeltypes.LegacyID(dr.PatientID),
				ScheduleDate:  dr.Date,
				ShiftId:       dr.ShiftID,
				BedId:         &bedID,
				WardId:        &wardID,
				PatientPlanId: &planID,
				ShiftTiming:   &dr.ShiftTiming,
				Status:        dr.Status,
				CreatorId:     creatorID,
			}
			if err := tx.Create(&ps).Error; err != nil {
				if isPatientShiftUniqueViolation(err) {
					continue // 并发或重复跳过
				}
				return err
			}
			// 写入PatientShiftExt
			ext := models.PatientShiftExt{
				TenantId:             tenantID,
				PatientShiftId:       ps.Id,
				DialysisMode:         dr.DialysisMode,
				SourceType:           dr.SourceType,
				RecordForm:           dr.RecordForm,
				RuleStatus:           10,
				SourceTemplateItemId: dr.TemplateItemID,
				CreatorId:            creatorID,
			}
			if err := tx.Create(&ext).Error; err != nil {
				return err
			}
			drafts++
		}

		// 保存冲突
		for _, c := range result.Conflicts {
			patientID := c.PatientID
			wid := c.WardID
			sid := c.ShiftID
			conflict := models.ConflictQueue{
				TenantId:                tenantID,
				PatientId:               &patientID,
				ScheduleDate:            &c.Date,
				ShiftId:                 &sid,
				WardId:                  &wid,
				ConflictType:            c.ConflictType,
				Severity:                c.Severity,
				Detail:                  c.Detail,
				SuggestedDate:           c.SuggestedDate,
				SuggestedShiftId:        c.SuggestedShiftID,
				SuggestedBedId:          c.SuggestedBedID,
				SuggestedPatientShiftId: c.SuggestedPatientShiftID,
				Status:                  0,
				CreatorId:               creatorID,
			}
			if err := tx.Create(&conflict).Error; err != nil {
				return err
			}
			conflicts++
		}
		return nil
	})

	if txErr != nil {
		return 0, 0, txErr
	}
	return drafts, conflicts, nil
}

func getAnchorMonday(db *gorm.DB, tenantID int64) time.Time {
	var val string
	row := db.Table(`"Schedule_TenantSetting"`).
		Select(`"SettingValue"`).
		Where(`"TenantId" = ? AND "SettingKey" = ?`, tenantID, "OddEvenWeekAnchorMonday").
		Row()
	if row != nil {
		_ = row.Scan(&val)
	}
	return schedule_engine.AnchorFromString(val)
}
