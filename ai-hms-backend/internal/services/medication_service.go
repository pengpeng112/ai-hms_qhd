package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

type MedicationService struct {
	db       *gorm.DB
	tenantID int64
}

func NewMedicationService() *MedicationService {
	return &MedicationService{db: database.GetDB(), tenantID: LegacyTenantID}
}

type MaRecordInput struct {
	OrderID          int64  `json:"orderId"`
	TreatmentID      int64  `json:"treatmentId"`
	DrugName         string `json:"drugName"`
	Category         string `json:"category"`
	Dose             string `json:"dose"`
	Route            string `json:"route"`
	Timing           string `json:"timing"`
	AdministeredBy   string `json:"administeredBy"`
	AdministeredName string `json:"administeredName"`
	Note             string `json:"note"`
}

func (s *MedicationService) RecordAdmin(patientID int64, in MaRecordInput) (*models.MedicationAdmin, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if strings.TrimSpace(in.AdministeredBy) == "" {
		return nil, errors.New("执行人不能为空")
	}
	if strings.TrimSpace(in.DrugName) == "" {
		return nil, errors.New("药品名称不能为空")
	}
	if in.OrderID == 0 {
		return nil, errors.New("医嘱ID不能为空")
	}
	ma := &models.MedicationAdmin{
		ID:               utils.GenerateID(),
		TenantID:         s.tenantID,
		PatientID:        patientID,
		OrderID:          in.OrderID,
		TreatmentID:      in.TreatmentID,
		DrugName:         strings.TrimSpace(in.DrugName),
		Category:         in.Category,
		Dose:             in.Dose,
		Route:            in.Route,
		Timing:           in.Timing,
		AdministeredBy:   strings.TrimSpace(in.AdministeredBy),
		AdministeredName: in.AdministeredName,
		AdministeredAt:   time.Now(),
		Status:           models.MAStatusRecorded,
		Note:             in.Note,
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ma).Error; err != nil {
			return err
		}
		ss := &SignService{db: tx}
		_, err := ss.Sign(s.tenantID, models.SignTargetMedicationAdmin, ma.ID, ma.AdministeredBy, ma.AdministeredName)
		return err
	})
	if err != nil {
		return nil, err
	}
	return ma, nil
}

func (s *MedicationService) SecondCheck(id string, checkerID, checkerName string) (*models.MedicationAdmin, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if strings.TrimSpace(checkerID) == "" {
		return nil, errors.New("核对人不能为空")
	}
	var ma models.MedicationAdmin
	if err := s.db.Where("id = ? AND tenant_id = ?", id, s.tenantID).First(&ma).Error; err != nil {
		return nil, errors.New("给药记录不存在")
	}
	if ma.Status == models.MAStatusVerified {
		return nil, errors.New("该记录已核对，不可重复核对")
	}
	if ma.Status != models.MAStatusRecorded {
		return nil, errors.New("仅已记录的给药可核对")
	}
	if ma.AdministeredBy == checkerID {
		return nil, errors.New("核对人不能是执行人本人")
	}
	now := time.Now()
	updates := map[string]any{
		"second_check_by":   checkerID,
		"second_check_name": checkerName,
		"second_check_at":   now,
		"status":            models.MAStatusVerified,
		"updated_at":        now,
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ma).Updates(updates).Error; err != nil {
			return err
		}
		ss := &SignService{db: tx}
		_, err := ss.Sign(s.tenantID, models.SignTargetMedicationAdmin, ma.ID, checkerID, checkerName)
		return err
	}); err != nil {
		return nil, err
	}
	return s.getByID(id)
}

func (s *MedicationService) getByID(id string) (*models.MedicationAdmin, error) {
	var ma models.MedicationAdmin
	if err := s.db.Where("id = ? AND tenant_id = ?", id, s.tenantID).First(&ma).Error; err != nil {
		return nil, errors.New("记录不存在")
	}
	return &ma, nil
}

func (s *MedicationService) List(treatmentID, patientID, orderID *int64) ([]models.MedicationAdmin, error) {
	q := s.db.Where("tenant_id = ?", s.tenantID)
	if treatmentID != nil && *treatmentID > 0 {
		q = q.Where("treatment_id = ?", *treatmentID)
	}
	if patientID != nil && *patientID > 0 {
		q = q.Where("patient_id = ?", *patientID)
	}
	if orderID != nil && *orderID > 0 {
		q = q.Where("order_id = ?", *orderID)
	}
	var rows []models.MedicationAdmin
	if err := q.Order("administered_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

type MedSuggestion struct {
	Indicator string `json:"indicator"`
	Label     string `json:"label"`
	Value     *float64 `json:"value"`
	Unit      string `json:"unit"`
	Status    string `json:"status"` // low/high/normal/no_data
	Drug      string `json:"drug"`
	DrugLabel string `json:"drugLabel"`
	Direction string `json:"direction"` // increase/decrease
	Advice    string `json:"advice"`
}

type medSuggestionRule struct {
	Indicator     string  `json:"indicator"`
	Label         string  `json:"label"`
	Unit          string  `json:"unit"`
	Low           float64 `json:"low"`
	High          float64 `json:"high"`
	LowDirection  string  `json:"low_direction"`
	HighDirection string  `json:"high_direction"`
	RelatedDrug   string  `json:"related_drug"`
	DrugLabel     string  `json:"drug_label"`
	AdviceLow     string  `json:"advice_low"`
	AdviceHigh    string  `json:"advice_high"`
}

type medDefaultDose struct {
	Drug         string `json:"drug"`
	Name         string `json:"name"`
	Route        string `json:"route"`
	DefaultDose  string `json:"defaultDose"`
	Frequency    string `json:"frequency"`
	Note         string `json:"note"`
	Enabled      bool   `json:"enabled"`
}

func (s *MedicationService) Suggestions(patientID int64) ([]MedSuggestion, error) {
	var rules []medSuggestionRule
	if err := json.Unmarshal(config.MedicationSuggestionRules, &rules); err != nil {
		return nil, fmt.Errorf("规则解析失败: %w", err)
	}
	labValues := s.fetchLatestLabValues(patientID)
	results := make([]MedSuggestion, 0, len(rules))
	for _, r := range rules {
		sug := MedSuggestion{
			Indicator: r.Indicator, Label: r.Label, Unit: r.Unit,
			Drug: r.RelatedDrug, DrugLabel: r.DrugLabel, Status: "no_data",
		}
		if v, ok := labValues[r.Indicator]; ok {
			sug.Value = v
			switch {
			case *v < r.Low:
				sug.Status = "low"
				sug.Direction = r.LowDirection
				sug.Advice = r.AdviceLow
			case *v > r.High:
				sug.Status = "high"
				sug.Direction = r.HighDirection
				sug.Advice = r.AdviceHigh
			default:
				sug.Status = "normal"
			}
		}
		results = append(results, sug)
	}
	return results, nil
}

func (s *MedicationService) fetchLatestLabValues(patientID int64) map[string]*float64 {
	if s.db == nil {
		return nil
	}
	query := s.db.Table(`"LIS_ExaminationItem" ei`).
		Select("ei.\"Name\", ei.\"Result\", ei.\"Code\"").
		Joins(`JOIN "LIS_Examination" e ON e."Id" = ei."MasterId"`).
		Where(`e."PatientId" = ? AND e."TenantId" = ? AND e."IsCancel" = false`, patientID, s.tenantID).
		Order(`e."ReportTime" DESC`).
		Limit(200)
	rows, err := query.Rows()
	if err != nil {
		log.Printf("[medication] fetchLatestLabValues query err: %v", err)
		return nil
	}
	defer rows.Close()
	type rawItem struct{ Name, Result, Code string }
	latest := map[string]*float64{}
	for rows.Next() {
		var it rawItem
		if err := s.db.ScanRows(rows, &it); err != nil {
			continue
		}
		name := strings.ToUpper(strings.TrimSpace(it.Name))
		code := strings.ToUpper(strings.TrimSpace(it.Code))
		key := ""
		switch {
		case strings.Contains(name, "血红蛋白") || strings.Contains(code, "HGB"):
			key = "Hb"
		case strings.Contains(name, "铁蛋白") || strings.Contains(code, "FER"):
			key = "Ferritin"
		case strings.Contains(name, "转铁蛋白饱和") || strings.Contains(code, "TSAT"):
			key = "TSAT"
		case strings.Contains(name, "磷") && (strings.Contains(name, "P") || len(name) <= 4) || code == "P":
			key = "P"
		case strings.Contains(name, "钙") || code == "CA":
			key = "Ca"
		case strings.Contains(name, "甲状旁腺") || strings.Contains(name, "PTH") || strings.Contains(code, "PTH")  || strings.Contains(code, "IPTH"):
			key = "iPTH"
		}
		if key == "" {
			continue
		}
		if _, exists := latest[key]; exists {
			continue
		}
		v := parseFloatPtr(it.Result)
		if v != nil {
			latest[key] = v
		}
	}
	return latest
}

type MedDefaultDose struct {
	Drug        string `json:"drug"`
	Name        string `json:"name"`
	Route       string `json:"route"`
	DefaultDose string `json:"defaultDose"`
	Frequency   string `json:"frequency"`
	Note        string `json:"note"`
}

func (s *MedicationService) DefaultDoses() ([]MedDefaultDose, error) {
	var raw []medDefaultDose
	if err := json.Unmarshal(config.MedicationDefaultDose, &raw); err != nil {
		return nil, fmt.Errorf("默认剂量配置解析失败: %w", err)
	}
	result := make([]MedDefaultDose, 0, len(raw))
	for _, d := range raw {
		if !d.Enabled {
			continue
		}
		result = append(result, MedDefaultDose{
			Drug: d.Drug, Name: d.Name, Route: d.Route,
			DefaultDose: d.DefaultDose, Frequency: d.Frequency, Note: d.Note,
		})
	}
	return result, nil
}
