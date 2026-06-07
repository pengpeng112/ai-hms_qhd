package services

import (
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
)

type StatisticsService struct{}

type QualityItem struct {
	Month int     `json:"month"`
	Ktv   float64 `json:"ktv"`
	Hb    float64 `json:"hb"`
	Alb   float64 `json:"alb"`
}

type InfectionItem struct {
	Month int `json:"month"`
	HbsAg int `json:"hbsAg"`
	Hcv   int `json:"hcv"`
	Hiv   int `json:"hiv"`
	Tp    int `json:"tp"`
}

type VascularItem struct {
	Month int `json:"month"`
	Avf   int `json:"avf"`
	Avg   int `json:"avg"`
	Tcc   int `json:"tcc"`
}

type WorkloadItem struct {
	UserID     int64  `json:"userId"`
	Name       string `json:"name"`
	Treatments int    `json:"treatments"`
	Punctures  int    `json:"punctures"`
}

func NewStatisticsService() *StatisticsService {
	return &StatisticsService{}
}

func (s *StatisticsService) QualityByYear(tenantId int64, year int) ([]QualityItem, error) {
	items := make([]QualityItem, 12)
	for i := 1; i <= 12; i++ {
		items[i-1] = QualityItem{Month: i}
	}
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0)

	db := database.GetDB()
	if db == nil {
		return items, nil
	}

	type labRow struct {
		Month      int    `gorm:"column:month"`
		ItemName   string `gorm:"column:item_name"`
		ItemCode   string `gorm:"column:item_code"`
		ResultSign string `gorm:"column:result_sign"`
	}
	var rows []labRow
	err := db.Table(`"LIS_ExaminationItem" AS i`).
		Select(`EXTRACT(MONTH FROM COALESCE(e."ResultTime", i."LastModifyTime"))::int AS month,
			i."ItemName" AS item_name, i."ItemCode" AS item_code, i."ResultSign" AS result_sign`).
		Joins(`JOIN "LIS_Examination" AS e ON e."Id" = i."ExaminationId"`).
		Where(`e."TenantId" = ? AND COALESCE(e."ResultTime", i."LastModifyTime") >= ? AND COALESCE(e."ResultTime", i."LastModifyTime") < ?`, tenantId, start, end).
		Find(&rows).Error
	if err != nil {
		return items, nil
	}

	type counter struct{ total, normal int }
	ktvCount := map[int]counter{}
	hbCount := map[int]counter{}
	albCount := map[int]counter{}

	for _, row := range rows {
		if row.Month < 1 || row.Month > 12 {
			continue
		}
		name := strings.ToLower(row.ItemName + " " + row.ItemCode)
		isNormal := strings.ToUpper(strings.TrimSpace(row.ResultSign)) == "N" || row.ResultSign == ""
		switch {
		case strings.Contains(name, "kt/v") || strings.Contains(name, "ktv"):
			c := ktvCount[row.Month]
			c.total++
			if isNormal {
				c.normal++
			}
			ktvCount[row.Month] = c
		case strings.Contains(name, "hb") || strings.Contains(name, "hemoglobin"):
			c := hbCount[row.Month]
			c.total++
			if isNormal {
				c.normal++
			}
			hbCount[row.Month] = c
		case strings.Contains(name, "alb") || strings.Contains(name, "albumin"):
			c := albCount[row.Month]
			c.total++
			if isNormal {
				c.normal++
			}
			albCount[row.Month] = c
		}
	}

	for i := range items {
		month := items[i].Month
		if c := ktvCount[month]; c.total > 0 {
			items[i].Ktv = float64(c.normal) * 100 / float64(c.total)
		}
		if c := hbCount[month]; c.total > 0 {
			items[i].Hb = float64(c.normal) * 100 / float64(c.total)
		}
		if c := albCount[month]; c.total > 0 {
			items[i].Alb = float64(c.normal) * 100 / float64(c.total)
		}
	}
	return items, nil
}

func (s *StatisticsService) InfectionByYear(tenantId int64, year int) ([]InfectionItem, error) {
	items := make([]InfectionItem, 12)
	for i := 1; i <= 12; i++ {
		items[i-1] = InfectionItem{Month: i}
	}
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0)

	db := database.GetDB()
	if db == nil {
		return items, nil
	}

	type infRow struct {
		LastModifyTime time.Time `gorm:"column:LastModifyTime"`
		InfectionDesc  string    `gorm:"column:InfectionDesc"`
	}
	var rows []infRow
	if err := db.Table(`"Register_Infection"`).
		Select(`"LastModifyTime"`, `"InfectionDesc"`).
		Where(`"TenantId" = ? AND "LastModifyTime" >= ? AND "LastModifyTime" < ?`, tenantId, start, end).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		month := int(row.LastModifyTime.Month())
		idx := month - 1
		if idx < 0 || idx >= len(items) {
			continue
		}
		lower := strings.ToLower(row.InfectionDesc)
		if strings.Contains(lower, "hbs") || strings.Contains(lower, "乙肝") {
			items[idx].HbsAg++
		}
		if strings.Contains(lower, "hcv") || strings.Contains(lower, "丙肝") {
			items[idx].Hcv++
		}
		if strings.Contains(lower, "hiv") || strings.Contains(lower, "艾滋") {
			items[idx].Hiv++
		}
		if strings.Contains(lower, "tp") || strings.Contains(lower, "梅毒") || strings.Contains(lower, "rpr") {
			items[idx].Tp++
		}
	}
	return items, nil
}

func (s *StatisticsService) VascularByYear(tenantId int64, year int) ([]VascularItem, error) {
	items := make([]VascularItem, 12)
	for i := 1; i <= 12; i++ {
		items[i-1] = VascularItem{Month: i}
	}
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0)

	db := database.GetDB()
	if db == nil {
		return items, nil
	}

	type vascRow struct {
		CreateTime time.Time `gorm:"column:CreateTime"`
		AccessType string    `gorm:"column:AccessType"`
	}
	var rows []vascRow
	if err := db.Table(`"Register_VascularAccess"`).
		Select(`"CreateTime"`, `"AccessType"`).
		Where(`"TenantId" = ? AND "CreateTime" >= ? AND "CreateTime" < ?`, tenantId, start, end).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		month := int(row.CreateTime.Month())
		idx := month - 1
		if idx < 0 || idx >= len(items) {
			continue
		}
		accessType := strings.ToLower(row.AccessType)
		switch {
		case strings.Contains(accessType, "avf") || strings.Contains(accessType, "内瘘"):
			items[idx].Avf++
		case strings.Contains(accessType, "avg"):
			items[idx].Avg++
		case strings.Contains(accessType, "tcc") || strings.Contains(accessType, "导管") || strings.Contains(accessType, "ncc"):
			items[idx].Tcc++
		}
	}
	return items, nil
}

func (s *StatisticsService) WorkloadByYearMonth(tenantId int64, yearMonth string) ([]WorkloadItem, error) {
	start, err := time.ParseInLocation("2006-01", yearMonth, time.Local)
	if err != nil {
		return []WorkloadItem{}, nil
	}
	end := start.AddDate(0, 1, 0)

	db := database.GetDB()
	if db == nil {
		return []WorkloadItem{}, nil
	}

	type wlRow struct {
		CreatorID int64  `gorm:"column:CreatorId"`
		RealName  string `gorm:"column:RealName"`
		Count     int    `gorm:"column:cnt"`
	}
	var rows []wlRow
	if err := db.Table(`"Treatment_Treatment" AS t`).
		Select(`t."CreatorId", COALESCE(e."Name", '') AS "RealName", COUNT(*) AS cnt`).
		Joins(`LEFT JOIN "Organ_Employee" AS e ON e."Id" = t."CreatorId"`).
		Where(`t."TenantId" = ? AND t."StartTime" >= ? AND t."StartTime" < ?`, tenantId, start, end).
		Group(`t."CreatorId", e."Name"`).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]WorkloadItem, 0, len(rows))
	for _, r := range rows {
		result = append(result, WorkloadItem{
			UserID:     r.CreatorID,
			Name:       r.RealName,
			Treatments: r.Count,
			Punctures:  r.Count,
		})
	}
	return result, nil
}
