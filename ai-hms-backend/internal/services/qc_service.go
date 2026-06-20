package services

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"gorm.io/gorm"
)

// 从检验结果文本中提取首个数值（兜底解析带单位/箭头/比较符的 LIS 结果）。
var qcNumberRe = regexp.MustCompile(`[-+]?[0-9]*\.?[0-9]+`)

type QCService struct {
	db       *gorm.DB
	concepts []config.IndicatorConcept
}

func NewQCService() *QCService {
	concepts, _ := config.LoadIndicatorConcepts()
	return &QCService{db: database.GetDB(), concepts: concepts}
}

func monthRange(year, month int) (time.Time, time.Time) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 1, 0)
}

// parseFloatPtr 解析检验结果为数值；nil = 非数值（如"阴性"）。
// LIS_ExaminationItem.Result 为 varchar，实务常带单位/箭头/比较符（"12.5↑"、"12.5 g/L"、"<0.01"、">120"），
// 故整体解析失败时回退到"提取首个数值"，避免有效检验值被静默当缺测、拉低质控赋分。
func parseFloatPtr(s string) *float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return &v
	}
	m := qcNumberRe.FindString(s)
	if m == "" {
		return nil
	}
	if v, err := strconv.ParseFloat(m, 64); err == nil {
		return &v
	}
	return nil
}

func (s *QCService) conceptByID(id string) *config.IndicatorConcept {
	for i := range s.concepts {
		if s.concepts[i].ConceptID == id {
			return &s.concepts[i]
		}
	}
	return nil
}

func conceptMatches(c *config.IndicatorConcept, itemCode, itemName string) bool {
	if c == nil {
		return false
	}
	code := strings.ToUpper(strings.TrimSpace(itemCode))
	for _, h := range c.ItemCodeHints {
		if code == strings.ToUpper(strings.TrimSpace(h)) {
			return true
		}
	}
	name := strings.ToLower(itemName)
	for _, kw := range c.NameKeywords {
		if kw != "" && strings.Contains(name, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func (s *QCService) latestLabValue(patientID int64, conceptID string, start, end time.Time) *float64 {
	c := s.conceptByID(conceptID)
	if c == nil {
		return nil
	}
	var rows []struct {
		ItemCode    string
		ItemName    string
		ResultValue string
		TestedAt    *time.Time
	}
	if err := s.db.Table(`"LIS_ExaminationItem" AS i`).
		Select(`i."ItemCode" AS "ItemCode", i."ItemName" AS "ItemName", i."Result" AS "ResultValue", COALESCE(e."ResultTime", i."LastModifyTime") AS "TestedAt"`).
		Joins(`JOIN "LIS_Examination" AS e ON e."Id" = i."ExaminationId"`).
		Where(`e."PatientId" = ? AND e."TenantId" = ? AND COALESCE(e."ResultTime", i."LastModifyTime") >= ? AND COALESCE(e."ResultTime", i."LastModifyTime") < ?`, patientID, LegacyTenantID, start, end).
		Order(`COALESCE(e."ResultTime", i."LastModifyTime") DESC`).
		Find(&rows).Error; err != nil {
		return nil
	}
	for _, r := range rows {
		if conceptMatches(c, r.ItemCode, r.ItemName) {
			return parseFloatPtr(r.ResultValue)
		}
	}
	return nil
}

func (s *QCService) avgBeforeSigns(patientID int64, start, end time.Time) (sp, dp, hr *float64) {
	var row struct {
		AvgSBP *float64
		AvgDBP *float64
		AvgHR  *float64
	}
	s.db.Table(`"Treatment_BeforeSigns" AS b`).
		Select(`AVG(NULLIF(b."SBP",0)) AS avg_sbp, AVG(NULLIF(b."DBP",0)) AS avg_dbp, AVG(NULLIF(b."HeartRate",0)) AS avg_hr`).
		Joins(`JOIN "Treatment_Treatment" AS t ON t."Id" = b."TreatmentId"`).
		Where(`t."TenantId" = ? AND t."PatientId" = ? AND COALESCE(t."StartTime", t."SignInTime", t."ReceptionTime", t."CreateTime") >= ? AND COALESCE(t."StartTime", t."SignInTime", t."ReceptionTime", t."CreateTime") < ?`,
			LegacyTenantID, patientID, start, end).
		Take(&row)
	return row.AvgSBP, row.AvgDBP, row.AvgHR
}

func (s *QCService) avgTUFPercent(patientID int64, start, end time.Time) *float64 {
	var avgUF *float64
	s.db.Table(`"Treatment_Treatment" AS t`).
		Select(`AVG(NULLIF(t."RealUFQuantity",0))`).
		Where(`t."TenantId" = ? AND t."PatientId" = ? AND COALESCE(t."StartTime", t."SignInTime", t."ReceptionTime", t."CreateTime") >= ? AND COALESCE(t."StartTime", t."SignInTime", t."ReceptionTime", t."CreateTime") < ?`,
			LegacyTenantID, patientID, start, end).
		Take(&avgUF)
	if avgUF == nil || *avgUF == 0 {
		return nil
	}
	var dryWeight *float64
	s.db.Table(`"Plan_PatientPlan"`).
		Select(`"DryWeight"`).
		Where(`"PatientId" = ? AND "TenantId" = ? AND COALESCE("IsDisabled",false)=false`, patientID, LegacyTenantID).
		Order(`"CreateTime" DESC`).
		Limit(1).Take(&dryWeight)
	if dryWeight == nil || *dryWeight == 0 {
		return nil
	}
	pct := *avgUF / *dryWeight * 100
	return &pct
}

func (s *QCService) BuildQCInput(patientID int64, year, month int) QCInput {
	start, end := monthRange(year, month)
	sp, dp, hr := s.avgBeforeSigns(patientID, start, end)
	return QCInput{
		AvgSP: sp, AvgDP: dp, AvgHR: hr,
		TUFPercent: s.avgTUFPercent(patientID, start, end),
		URR:        s.latestLabValue(patientID, "URR", start, end),
		KtV:        s.latestLabValue(patientID, "KTV", start, end),
		Hb:         s.latestLabValue(patientID, "HEMOGLOBIN", start, end),
		Alb:        s.latestLabValue(patientID, "ALBUMIN", start, end),
		Ca:         s.latestLabValue(patientID, "SERUM_CA", start, end),
		P:          s.latestLabValue(patientID, "SERUM_P", start, end),
		PTH:        s.latestLabValue(patientID, "IPTH", start, end),
	}
}

type QCPatientRow struct {
	PatientID   string         `json:"patientId"`
	PatientName string         `json:"patientName"`
	Score       QCPatientScore `json:"score"`
}

type qcPatientRef struct {
	ID   int64  `gorm:"column:Id"`
	Name string `gorm:"column:Name"`
}

func (s *QCService) patientsByDoctor(doctorID int64) ([]qcPatientRef, error) {
	var rows []qcPatientRef
	if err := s.db.Table(`"Register_PatientInfomation"`).
		Select(`"Id", "Name"`).
		Where(`"TenantId" = ? AND "ResponsibilityDrId" = ?`, LegacyTenantID, doctorID).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *QCService) ScoreDoctorDetail(doctorID int64, year, month int) (QCDoctorScore, []QCPatientRow, error) {
	if s.db == nil {
		return QCDoctorScore{}, nil, errors.New("database not available")
	}
	pats, err := s.patientsByDoctor(doctorID)
	if err != nil {
		return QCDoctorScore{}, nil, err
	}
	scores := make([]QCPatientScore, 0, len(pats))
	detail := make([]QCPatientRow, 0, len(pats))
	for _, p := range pats {
		ps := ScorePatient(s.BuildQCInput(p.ID, year, month))
		scores = append(scores, ps)
		detail = append(detail, QCPatientRow{PatientID: strconv.FormatInt(p.ID, 10), PatientName: p.Name, Score: ps})
	}
	d := AggregateDoctor(strconv.FormatInt(doctorID, 10), scores)
	d.DoctorName = s.resolveDoctorName(doctorID)
	return d, detail, nil
}

func (s *QCService) resolveDoctorName(doctorID int64) string {
	if doctorID <= 0 {
		return ""
	}
	var emp struct {
		Name string `gorm:"column:Name"`
	}
	if err := s.db.Table(`"Organ_Employee"`).Select(`"Name"`).
		Where(`"UserId" = ? AND "TenantId" = ?`, doctorID, LegacyTenantID).
		Order(`"Id" ASC`).First(&emp).Error; err == nil && strings.TrimSpace(emp.Name) != "" {
		return strings.TrimSpace(emp.Name)
	}
	if err := s.db.Table(`"Organ_Employee"`).Select(`"Name"`).
		Where(`"Id" = ? AND "TenantId" = ?`, doctorID, LegacyTenantID).
		Order(`"Id" ASC`).First(&emp).Error; err == nil && strings.TrimSpace(emp.Name) != "" {
		return strings.TrimSpace(emp.Name)
	}
	var u struct {
		UserName string `gorm:"column:UserName"`
	}
	if err := s.db.Table(`"Identity_Users"`).Select(`"UserName"`).
		Where(`"Id" = ?`, doctorID).First(&u).Error; err == nil && strings.TrimSpace(u.UserName) != "" {
		return strings.TrimSpace(u.UserName)
	}
	return ""
}

func (s *QCService) ScoreDoctors(year, month int) ([]QCDoctorScore, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var doctorIDs []int64
	if err := s.db.Table(`"Register_PatientInfomation"`).
		Where(`"TenantId" = ? AND "ResponsibilityDrId" IS NOT NULL AND "ResponsibilityDrId" > 0`, LegacyTenantID).
		Distinct().Pluck(`"ResponsibilityDrId"`, &doctorIDs).Error; err != nil {
		return nil, err
	}
	out := make([]QCDoctorScore, 0, len(doctorIDs))
	for _, did := range doctorIDs {
		d, _, err := s.ScoreDoctorDetail(did, year, month)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}
