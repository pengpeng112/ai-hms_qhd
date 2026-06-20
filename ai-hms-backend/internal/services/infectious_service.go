package services

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

type InfectiousService struct {
	db       *gorm.DB
	tenantID int64
}

func NewInfectiousService() *InfectiousService {
	return &InfectiousService{db: database.GetDB(), tenantID: LegacyTenantID}
}

type ScreenItem struct {
	Item   string `json:"item"`
	Result string `json:"result"` // negative/positive/indeterminate
}

type ScreenInput struct {
	ScreenDate time.Time    `json:"screenDate"`
	Source     string       `json:"source"`
	Items      []ScreenItem `json:"items"`
	Note       string       `json:"note"`
}

func (s *InfectiousService) Screen(patientID int64, in ScreenInput) (*models.PatientInfectious, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if len(in.Items) == 0 {
		return nil, errors.New("筛查项目不能为空")
	}
	overall := models.InfectiousNegative
	var positives []string
	hasIndeterminate := false
	for _, it := range in.Items {
		switch it.Result {
		case models.InfItemPositive:
			positives = append(positives, it.Item)
		case models.InfItemIndeterminate:
			hasIndeterminate = true
		}
	}
	if len(positives) > 0 {
		overall = models.InfectiousPositive
	} else if hasIndeterminate {
		overall = models.InfectiousPending
	}
	itemsJSON, _ := json.Marshal(in.Items)
	screenDate := in.ScreenDate
	due := screenDate.AddDate(0, 6, 0)
	source := in.Source
	if source == "" {
		source = "manual"
	}
	rec := &models.PatientInfectious{
		ID:              utils.GenerateID(),
		TenantID:        s.tenantID,
		PatientID:       strconv.FormatInt(patientID, 10),
		ScreenDate:      &screenDate,
		Items:           string(itemsJSON),
		Source:          source,
		ResultOverall:   overall,
		PositiveMarkers: strings.Join(positives, ","),
		NextDueDate:     &due,
		ZoneTag:         "normal",
		Note:            in.Note,
	}
	if err := s.db.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

// latest 取该患者最新一条筛查记录（无则 nil,nil）
func (s *InfectiousService) latest(patientID int64) (*models.PatientInfectious, error) {
	var rec models.PatientInfectious
	err := s.db.Where("patient_id = ?", strconv.FormatInt(patientID, 10)).
		Order("screen_date DESC, created_at DESC").First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}
