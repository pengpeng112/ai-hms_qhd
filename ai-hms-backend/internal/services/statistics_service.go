package services

import (
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
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
	if db == nil || !db.Migrator().HasTable(&models.LabReportItem{}) {
		return items, nil
	}

	var rows []models.LabReportItem
	query := db.Model(&models.LabReportItem{})
	if db.Migrator().HasColumn(&models.LabReportItem{}, "tenant_id") {
		query = query.Where("tenant_id = ?", tenantId)
	}
	if !db.Migrator().HasColumn(&models.LabReportItem{}, "tested_at") {
		return items, nil
	}
	if err := query.Where("tested_at >= ? AND tested_at < ?", start, end).Find(&rows).Error; err != nil {
		return nil, err
	}

	type counter struct{ total, normal int }
	ktvCount := map[int]counter{}
	hbCount := map[int]counter{}
	albCount := map[int]counter{}

	for _, row := range rows {
		if row.TestedAt == nil {
			continue
		}
		month := int(row.TestedAt.Month())
		name := strings.ToLower(row.ItemName + " " + row.ItemCode)
		isNormal := strings.ToUpper(strings.TrimSpace(row.AbnormalFlag)) == "N" || row.AbnormalFlag == ""
		switch {
		case strings.Contains(name, "kt/v") || strings.Contains(name, "ktv"):
			c := ktvCount[month]
			c.total++
			if isNormal {
				c.normal++
			}
			ktvCount[month] = c
		case strings.Contains(name, "hb") || strings.Contains(name, "hemoglobin"):
			c := hbCount[month]
			c.total++
			if isNormal {
				c.normal++
			}
			hbCount[month] = c
		case strings.Contains(name, "alb") || strings.Contains(name, "albumin"):
			c := albCount[month]
			c.total++
			if isNormal {
				c.normal++
			}
			albCount[month] = c
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
	if db == nil || !db.Migrator().HasTable(&models.InfectionInfo{}) {
		return items, nil
	}

	var rows []models.InfectionInfo
	query := db.Model(&models.InfectionInfo{})
	if db.Migrator().HasColumn(&models.InfectionInfo{}, "tenant_id") {
		query = query.Where("tenant_id = ?", tenantId)
	}
	if !db.Migrator().HasColumn(&models.InfectionInfo{}, "update_date") {
		return items, nil
	}
	if err := query.Where("update_date >= ? AND update_date < ?", start, end).Find(&rows).Error; err != nil {
		return nil, err
	}
	isPositive := func(v string) bool {
		val := strings.ToLower(strings.TrimSpace(v))
		return val == "positive" || val == "阳性"
	}
	for _, row := range rows {
		month := int(row.UpdateDate.Month())
		idx := month - 1
		if idx < 0 || idx >= len(items) {
			continue
		}
		if isPositive(row.HbsAg) {
			items[idx].HbsAg++
		}
		if isPositive(row.HcvAb) {
			items[idx].Hcv++
		}
		if isPositive(row.HivAb) {
			items[idx].Hiv++
		}
		if isPositive(row.TpaB) {
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
	if db == nil || !db.Migrator().HasTable(&models.VascularAccess{}) {
		return items, nil
	}

	var rows []models.VascularAccess
	query := db.Model(&models.VascularAccess{})
	if db.Migrator().HasColumn(&models.VascularAccess{}, "tenant_id") {
		query = query.Where("tenant_id = ?", tenantId)
	}
	dateColumn := ""
	switch {
	case db.Migrator().HasColumn(&models.VascularAccess{}, "created_at"):
		dateColumn = "created_at"
	case db.Migrator().HasColumn(&models.VascularAccess{}, "create_time"):
		dateColumn = "create_time"
	default:
		return items, nil
	}
	if err := query.Where(dateColumn+" >= ? AND "+dateColumn+" < ?", start, end).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		month := int(row.CreatedAt.Month())
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
		case strings.Contains(accessType, "tcc") || strings.Contains(accessType, "导管"):
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
	if db == nil || !db.Migrator().HasTable(&models.Treatment{}) {
		return []WorkloadItem{}, nil
	}

	var treatments []models.Treatment
	treatmentQuery := db.Model(&models.Treatment{})
	if db.Migrator().HasColumn(&models.Treatment{}, "tenant_id") {
		treatmentQuery = treatmentQuery.Where("tenant_id = ?", tenantId)
	}
	if err := treatmentQuery.Where("treatment_date >= ? AND treatment_date < ?", start, end).Find(&treatments).Error; err != nil {
		return nil, err
	}

	workloads := map[int64]*WorkloadItem{}
	for _, treatment := range treatments {
		item, ok := workloads[treatment.CreatorId]
		if !ok {
			item = &WorkloadItem{UserID: treatment.CreatorId, Name: "", Treatments: 0, Punctures: 0}
			workloads[treatment.CreatorId] = item
		}
		item.Treatments++
		item.Punctures++
	}

	var users []models.User
	userQuery := db.Model(&models.User{})
	if db.Migrator().HasColumn(&models.User{}, "tenant_id") {
		userQuery = userQuery.Where("tenant_id = ?", tenantId)
	}
	if err := userQuery.Find(&users).Error; err == nil {
		nameMap := map[int64]string{}
		for _, user := range users {
			id, parseErr := strconv.ParseInt(user.ID, 10, 64)
			if parseErr == nil {
				nameMap[id] = user.RealName
			}
		}
		for id, item := range workloads {
			if name, ok := nameMap[id]; ok {
				item.Name = name
			}
		}
	}

	result := make([]WorkloadItem, 0, len(workloads))
	for _, item := range workloads {
		result = append(result, *item)
	}
	return result, nil
}
