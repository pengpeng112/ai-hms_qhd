package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"gorm.io/gorm"
)

type MonthlySummaryService struct{}

func NewMonthlySummaryService() *MonthlySummaryService {
	return &MonthlySummaryService{}
}

type MonthlySummaryDTO struct {
	ID         int64                      `json:"id"`
	PatientID  int64                      `json:"patientId"`
	Year       int                        `json:"year"`
	Month      int                        `json:"month"`
	Content    map[string]interface{}     `json:"content"`
	CreatedAt  time.Time                  `json:"createdAt"`
	UpdatedAt  time.Time                  `json:"updatedAt"`
}

type SaveMonthlySummaryRequest struct {
	Content map[string]interface{} `json:"content" binding:"required"`
}

type legacyMonthlySummaryRow struct {
	ID              int64     `gorm:"column:Id"`
	TenantID        int64     `gorm:"column:TenantId"`
	PatientID       int64     `gorm:"column:PatientId"`
	ContentJsonb    string    `gorm:"column:ContentJsonb"`
	CreateTime      time.Time `gorm:"column:CreateTime"`
	LastModifyTime  time.Time `gorm:"column:LastModifyTime"`
}

func (legacyMonthlySummaryRow) TableName() string { return `"Treatment_TreatmentMonthSummarySheet"` }

func (s *MonthlySummaryService) Get(patientID int64, tenantID int64, year int, month int) (*MonthlySummaryDTO, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not available")
	}

	var row legacyMonthlySummaryRow
	err := db.Table(`"Treatment_TreatmentMonthSummarySheet"`).
		Where(`"PatientId" = ? AND "TenantId" = ? AND "Year" = ? AND "Month" = ?`, patientID, tenantID, year, month).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	content := map[string]interface{}{}
	if row.ContentJsonb != "" {
		json.Unmarshal([]byte(row.ContentJsonb), &content)
	}

	return &MonthlySummaryDTO{
		ID:        row.ID,
		PatientID: row.PatientID,
		Year:      year,
		Month:     month,
		Content:   content,
		CreatedAt: row.CreateTime,
		UpdatedAt: row.LastModifyTime,
	}, nil
}

func (s *MonthlySummaryService) Save(patientID int64, tenantID int64, year int, month int, req SaveMonthlySummaryRequest, creatorID int64) (*MonthlySummaryDTO, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not available")
	}

	contentBytes, err := json.Marshal(req.Content)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %w", err)
	}

	now := time.Now()

	// Upsert: try update first
	var existing legacyMonthlySummaryRow
	findErr := db.Table(`"Treatment_TreatmentMonthSummarySheet"`).
		Where(`"PatientId" = ? AND "TenantId" = ? AND "Year" = ? AND "Month" = ?`, patientID, tenantID, year, month).
		First(&existing).Error

	if findErr == nil {
		// Update
		updates := map[string]interface{}{
			`"ContentJsonb"`:   string(contentBytes),
			`"LastModifyTime"`: now,
		}
		db.Table(`"Treatment_TreatmentMonthSummarySheet"`).
			Where(`"Id" = ?`, existing.ID).
			Updates(updates)

		return &MonthlySummaryDTO{
			ID:        existing.ID,
			PatientID: patientID,
			Year:      year,
			Month:     month,
			Content:   req.Content,
			CreatedAt: existing.CreateTime,
			UpdatedAt: now,
		}, nil
	}

	// Insert
	newID, err := nextLegacyID()
	if err != nil {
		return nil, err
	}

	values := map[string]interface{}{
		`"Id"`:             int64(newID),
		`"TenantId"`:       tenantID,
		`"PatientId"`:      patientID,
		`"Year"`:           year,
		`"Month"`:          month,
		`"ContentJsonb"`:   string(contentBytes),
		`"CreatorId"`:      creatorID,
		`"CreateTime"`:     now,
		`"LastModifyTime"`: now,
	}

	if err := db.Table(`"Treatment_TreatmentMonthSummarySheet"`).Create(values).Error; err != nil {
		return nil, fmt.Errorf("create monthly summary: %w", err)
	}

	return &MonthlySummaryDTO{
		ID:        int64(newID),
		PatientID: patientID,
		Year:      year,
		Month:     month,
		Content:   req.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
