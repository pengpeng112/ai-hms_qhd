package services

import (
	"errors"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

type ScheduleWeekService struct {
	db *gorm.DB
}

func NewScheduleWeekService() *ScheduleWeekService {
	return &ScheduleWeekService{db: database.GetDB()}
}

type WeekWardItem struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Sort             int    `json:"sort"`
	PatientType      string `json:"patientType"`
	InfectionType    string `json:"infectionType"`
	ResponsibleUsers string `json:"responsibleUsers"`
	IsDisabled       bool   `json:"isDisabled"`
	BedCount         int    `json:"bedCount"`
}

type WeekBedItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	WardId   int64  `json:"wardId"`
	WardName string `json:"wardName"`
	Sort     int    `json:"sort"`
	IsDisabled bool `json:"isDisabled"`
}

type WeekShiftItem struct {
	ID                int64  `json:"id"`
	PatientID         int64  `json:"patientId"`
	PatientName       string `json:"patientName"`
	WardID            int64  `json:"wardId"`
	BedID             int64  `json:"bedId"`
	BedName           string `json:"bedName"`
	ShiftID           int64  `json:"shiftId"`
	PatientPlanID     int64  `json:"patientPlanId"`
	DialysisMode      string `json:"dialysisMode"`
	OddWeekFrequency  int    `json:"oddWeekFrequency"`
	EvenWeekFrequency int    `json:"evenWeekFrequency"`
	ShiftTiming       int    `json:"shiftTiming"`
	Status            int    `json:"status"`
	StatusName        string `json:"statusName"`
	TreatmentTime     string `json:"treatmentTime"`
	LastModifyTime    string `json:"lastModifyTime"`
	SourceType        string `json:"sourceType"`
	IsManualAdjusted  bool   `json:"isManualAdjusted"`
}

type PendingPatientItem struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	Spell             string `json:"spell"`
	Gender            string `json:"gender"`
	DialysisMode      string `json:"dialysisMode"`
	PatientPlanID     int64  `json:"patientPlanId"`
	OddWeekFrequency  int    `json:"oddWeekFrequency"`
	EvenWeekFrequency int    `json:"evenWeekFrequency"`
	ExpectedTimes     int    `json:"expectedTimes"`
	ScheduledTimes    int    `json:"scheduledTimes"`
	RemainingTimes    int    `json:"remainingTimes"`
}

type WeekScheduleResponse struct {
	Wards           []WeekWardItem       `json:"wards"`
	Beds            []WeekBedItem        `json:"beds"`
	Shifts          []models.Shift       `json:"shifts"`
	PatientShifts   []WeekShiftItem      `json:"patientShifts"`
	PendingPatients []PendingPatientItem `json:"pendingPatients"`
}

var statusNames = map[int]string{
	0:  "待排",
	10: "草稿",
	20: "已确认",
	30: "用户确认",
	40: "用户取消",
	50: "排班取消",
	60: "转出",
	70: "已取消",
	80: "缺席",
}

func (s *ScheduleWeekService) GetWeek(startDate, endDate string, tenantID int64, wardID *int64) (*WeekScheduleResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var resp WeekScheduleResponse

	// 1) 病区 DTO
	type wardRow struct {
		ID               int64  `gorm:"column:Id"`
		Name             string `gorm:"column:Name"`
		Sort             int    `gorm:"column:Sort"`
		PatientType      string `gorm:"column:PatientType"`
		InfectionType    string `gorm:"column:InfectionType"`
		ResponsibleUsers string `gorm:"column:ResponsibleUsers"`
	}
	var wards []wardRow
	wq := s.db.Table(`"Schedule_Ward"`).
		Select(`"Id", "Name", "Sort", COALESCE("PatientType", '') AS "PatientType",
			COALESCE("InfectionType", '') AS "InfectionType",
			COALESCE("ResponsibleUsers", '') AS "ResponsibleUsers"`).
		Where(`"TenantId" = ? AND "IsDisabled" = false`, tenantID).
		Order(`"Sort" ASC, "Id" ASC`)
	if err := wq.Find(&wards).Error; err != nil {
		return nil, err
	}

	// 2) 床位 DTO（含 wardName）
	type bedRow struct {
		ID        int64  `gorm:"column:Id"`
		Name      string `gorm:"column:Name"`
		WardID    int64  `gorm:"column:WardId"`
		WardName  string `gorm:"column:WardName"`
		Sort      int    `gorm:"column:Sort"`
	}
	var bedRows []bedRow
	bq := s.db.Table(`"Schedule_Bed" b`).
		Select(`b."Id", b."Name", COALESCE(b."WardId", 0) AS "WardId",
			COALESCE(w."Name", '') AS "WardName", COALESCE(b."Sort", 0) AS "Sort"`).
		Joins(`LEFT JOIN "Schedule_Ward" w ON w."Id" = b."WardId" AND w."TenantId" = b."TenantId"`).
		Where(`b."TenantId" = ? AND b."IsDisabled" = false`, tenantID)
	if wardID != nil {
		bq = bq.Where(`b."WardId" = ?`, *wardID)
	}
	if err := bq.Order(`b."Sort" ASC, b."Id" ASC`).Find(&bedRows).Error; err != nil {
		return nil, err
	}

	// 病区转 DTO + 统计 bedCount
	wardBedCount := map[int64]int{}
	for _, br := range bedRows {
		wardBedCount[br.WardID]++
	}
	for _, w := range wards {
		resp.Wards = append(resp.Wards, WeekWardItem{
			ID:               w.ID,
			Name:             w.Name,
			Sort:             w.Sort,
			PatientType:      w.PatientType,
			InfectionType:    w.InfectionType,
			ResponsibleUsers: w.ResponsibleUsers,
			IsDisabled:       false,
			BedCount:         wardBedCount[w.ID],
		})
	}
	for _, br := range bedRows {
		resp.Beds = append(resp.Beds, WeekBedItem{
			ID:        br.ID,
			Name:      br.Name,
			WardId:    br.WardID,
			WardName:  br.WardName,
			Sort:      br.Sort,
		})
	}

	// 3) 班次
	type shiftRow struct {
		ID        int64     `gorm:"column:Id"`
		Name      string    `gorm:"column:Name"`
		Sort      int       `gorm:"column:Sort"`
		StartTime string    `gorm:"column:StartTime"`
		EndTime   string    `gorm:"column:EndTime"`
	}
	var shiftRows []shiftRow
	if err := s.db.Table(`"Schedule_Shift"`).
		Select(`"Id", "Name", COALESCE("Sort", 0) AS "Sort",
			COALESCE("StartTime"::text, '') AS "StartTime",
			COALESCE("EndTime"::text, '') AS "EndTime"`).
		Where(`"TenantId" = ? AND "IsDisabled" = false`, tenantID).
		Order(`"Sort" ASC, "Id" ASC`).Find(&shiftRows).Error; err != nil {
		return nil, err
	}
	for _, sr := range shiftRows {
		resp.Shifts = append(resp.Shifts, models.Shift{
			Id:        sr.ID,
			Name:      sr.Name,
			Sort:      sr.Sort,
			StartTime: sr.StartTime,
			EndTime:   sr.EndTime,
		})
	}

	// 4) 患者排班
	type row struct {
		ID                int64     `gorm:"column:Id"`
		PatientID         int64     `gorm:"column:PatientId"`
		PatientName       string    `gorm:"column:PatientName"`
		WardID            int64     `gorm:"column:WardId"`
		BedID             int64     `gorm:"column:BedId"`
		BedName           string    `gorm:"column:BedName"`
		ShiftID           int64     `gorm:"column:ShiftId"`
		PatientPlanID     int64     `gorm:"column:PatientPlanId"`
		DialysisMode      string    `gorm:"column:DialysisMode"`
		OddWeekFrequency  int       `gorm:"column:OddWeekFrequency"`
		EvenWeekFrequency int       `gorm:"column:EvenWeekFrequency"`
		ShiftTiming       int       `gorm:"column:ShiftTiming"`
		Status            int       `gorm:"column:Status"`
		TreatmentTime     time.Time `gorm:"column:TreatmentTime"`
		LastModifyTime    time.Time `gorm:"column:LastModifyTime"`
		CreateTime        time.Time `gorm:"column:CreateTime"`
	}

	var rows []row
	shiftQuery := s.db.
		Table(`"Schedule_PatientShift" ps`).
		Select(`ps."Id", ps."PatientId", p."Name" AS "PatientName",
			ps."WardId", ps."BedId", b."Name" AS "BedName",
			ps."ShiftId", COALESCE(ps."PatientPlanId", 0) AS "PatientPlanId",
			COALESCE(pl."DialysisMethod", 'HD') AS "DialysisMode",
			COALESCE(pl."OddWeekFrequency", 0) AS "OddWeekFrequency",
			COALESCE(pl."EvenWeekFrequency", 0) AS "EvenWeekFrequency",
			COALESCE(ps."ShiftTiming", 20) AS "ShiftTiming",
			ps."Status", ps."TreatmentTime", ps."LastModifyTime", ps."CreateTime"`).
		Joins(`INNER JOIN "Register_PatientInfomation" p ON p."Id" = ps."PatientId" AND p."TenantId" = ps."TenantId"`).
		Joins(`LEFT JOIN "Schedule_Bed" b ON b."Id" = ps."BedId" AND b."TenantId" = ps."TenantId"`).
		Joins(`LEFT JOIN "Plan_PatientPlan" pl ON pl."Id" = ps."PatientPlanId" AND pl."TenantId" = ps."TenantId"`).
		Where(`ps."TenantId" = ? AND ps."Status" NOT IN (?) AND ps."TreatmentTime" >= ? AND ps."TreatmentTime" <= ?`,
			tenantID,
			[]int{
				MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusCancelled),
				MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusUserCancelled),
				MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusTransferred),
			},
			startDate+" 00:00:00",
			endDate+" 23:59:59",
		)

	if wardID != nil {
		shiftQuery = shiftQuery.Where(`ps."WardId" = ?`, *wardID)
	}

	if err := shiftQuery.Order(`ps."TreatmentTime" ASC, ps."Id" ASC`).Find(&rows).Error; err != nil {
		return nil, err
	}

	// 统计每个患者本周已排次数（仅长期排班）
	scheduledCount := map[int64]int{}
	for _, r := range rows {
		if r.ShiftTiming == 20 {
			scheduledCount[r.PatientID]++
		}
	}

	for _, r := range rows {
		sourceType := "manual"
		switch {
		case r.ShiftTiming == 10:
			sourceType = "temporary"
		case r.ShiftTiming == 20 && r.PatientPlanID > 0:
			sourceType = "contract"
		}
		isAdjusted := false
		if !r.CreateTime.IsZero() && r.LastModifyTime.Sub(r.CreateTime) > time.Minute {
			isAdjusted = true
		}

		resp.PatientShifts = append(resp.PatientShifts, WeekShiftItem{
			ID:                r.ID,
			PatientID:         r.PatientID,
			PatientName:       r.PatientName,
			WardID:            r.WardID,
			BedID:             r.BedID,
			BedName:           r.BedName,
			ShiftID:           r.ShiftID,
			PatientPlanID:     r.PatientPlanID,
			DialysisMode:      r.DialysisMode,
			OddWeekFrequency:  r.OddWeekFrequency,
			EvenWeekFrequency: r.EvenWeekFrequency,
			ShiftTiming:       r.ShiftTiming,
			Status:            MapPatientShiftStatusLegacyToNew(r.Status),
			StatusName:        statusNames[r.Status],
			TreatmentTime:     r.TreatmentTime.Format("2006-01-02T15:04:05Z07:00"),
			LastModifyTime:    r.LastModifyTime.Format("2006-01-02T15:04:05Z07:00"),
			SourceType:        sourceType,
			IsManualAdjusted:  isAdjusted,
		})
	}

	// 5) 待排班队列（频次差集）
	// 使用科室配置的奇偶周锚定周一；若无配置则 fallback 到 ISO 周号
	startTime, dateErr := time.Parse("2006-01-02", startDate)
	if dateErr != nil {
		return nil, errors.New("invalid startDate format, expected YYYY-MM-DD")
	}
	isOddWeek := isOddDialysisWeek(s.db, tenantID, startTime)

	type pendingRow struct {
		ID                int64  `gorm:"column:Id"`
		Name              string `gorm:"column:Name"`
		Spell             string `gorm:"column:Spell"`
		Gender            string `gorm:"column:Gender"`
		DialysisMode      string `gorm:"column:DialysisMode"`
		PatientPlanID     int64  `gorm:"column:PatientPlanId"`
		OddWeekFrequency  int    `gorm:"column:OddWeekFrequency"`
		EvenWeekFrequency int    `gorm:"column:EvenWeekFrequency"`
	}
	var pendingList []pendingRow
	pendingQuery := s.db.
		Table(`"Plan_PatientPlan" pl`).
		Select(`p."Id", p."Name", COALESCE(p."Spell", '') AS "Spell",
			COALESCE(p."Gender", '') AS "Gender",
			COALESCE(pl."DialysisMethod", 'HD') AS "DialysisMode",
			COALESCE(pl."Id", 0) AS "PatientPlanId",
			COALESCE(pl."OddWeekFrequency", 0) AS "OddWeekFrequency",
			COALESCE(pl."EvenWeekFrequency", 0) AS "EvenWeekFrequency"`).
		Joins(`INNER JOIN "Register_PatientInfomation" p ON p."Id" = pl."PatientId" AND p."TenantId" = pl."TenantId"`).
		Where(`pl."TenantId" = ? AND pl."IsDisabled" = false`, tenantID).
		Order(`p."Id" ASC`)

	if err := pendingQuery.Find(&pendingList).Error; err != nil {
		return nil, err
	}

	// 去重：同一患者可能有多条Plan_PatientPlan记录，只保留首条（频率最高者优先）
	seenPatients := map[int64]bool{}
	for _, pr := range pendingList {
		if seenPatients[pr.ID] {
			continue
		}
		expected := pr.EvenWeekFrequency
		if isOddWeek {
			expected = pr.OddWeekFrequency
		}
		if expected <= 0 {
			continue
		}
		scheduled := scheduledCount[pr.ID]
		remaining := expected - scheduled
		if remaining <= 0 {
			continue
		}

		seenPatients[pr.ID] = true
		resp.PendingPatients = append(resp.PendingPatients, PendingPatientItem{
			ID:                pr.ID,
			Name:              pr.Name,
			Spell:             pr.Spell,
			Gender:            pr.Gender,
			DialysisMode:      pr.DialysisMode,
			PatientPlanID:     pr.PatientPlanID,
			OddWeekFrequency:  pr.OddWeekFrequency,
			EvenWeekFrequency: pr.EvenWeekFrequency,
			ExpectedTimes:     expected,
			ScheduledTimes:    scheduled,
			RemainingTimes:    remaining,
		})
	}

	return &resp, nil
}

// isOddDialysisWeek 判断指定日期属于科室定义的奇数透析周。
// 优先使用 Schedule_TenantSetting 中配置的 OddEvenWeekAnchorMonday；
// 若无配置则 fallback 到 ISO 周号判断奇偶周。
func isOddDialysisWeek(db *gorm.DB, tenantID int64, date time.Time) bool {
	var anchorStr string
	row := db.Table(`"Schedule_TenantSetting"`).
		Select(`"SettingValue"`).
		Where(`"TenantId" = ? AND "SettingKey" = ?`, tenantID, "OddEvenWeekAnchorMonday").
		Row()
	if row != nil {
		_ = row.Scan(&anchorStr)
	}
	anchorStr = strings.TrimSpace(anchorStr)
	if anchorStr != "" {
		anchor, err := time.Parse("2006-01-02", anchorStr)
		if err == nil && !anchor.IsZero() {
			days := int(date.Sub(anchor).Hours() / 24)
			if days >= 0 {
				weekNum := days / 7
				return weekNum%2 == 0
			}
		}
	}

	// fallback: ISO week
	_, isoWeek := date.ISOWeek()
	return isoWeek%2 == 1
}
