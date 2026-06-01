package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
)

type ConsumableService struct{}

func NewConsumableService() *ConsumableService {
	return &ConsumableService{}
}

type ConsumableDTO struct {
	ID           int64     `json:"id"`
	TreatmentID  int64     `json:"treatmentId"`
	MaterialID   int64     `json:"materialId"`
	MaterialName string    `json:"materialName"`
	Num          float64   `json:"num"`
	Unit         string    `json:"unit"`
	Batch        string    `json:"batch"`
	SerialNo     string    `json:"serialNo"`
	Note         string    `json:"note"`
	CreatedAt    time.Time `json:"createdAt"`
}

type CreateConsumableRequest struct {
	MaterialID int64   `json:"materialId" binding:"required"`
	Num        float64 `json:"num"`
	Unit       string  `json:"unit"`
	Batch      string  `json:"batch"`
	SerialNo   string  `json:"serialNo"`
	Note       string  `json:"note"`
}

type legacyConsumableRow struct {
	ID           int64     `gorm:"column:Id"`
	TenantID     int64     `gorm:"column:TenantId"`
	TreatmentID  int64     `gorm:"column:TreatmentId"`
	ChargeItemID int64     `gorm:"column:ChargeItemId"`
	Num          float64   `gorm:"column:Num"`
	Unit         string    `gorm:"column:Unit"`
	Batch        string    `gorm:"column:Batch"`
	SerialNo     string    `gorm:"column:SerialNo"`
	Note         string    `gorm:"column:Note"`
	CreateTime   time.Time `gorm:"column:CreateTime"`
}

func (s *ConsumableService) ListByTreatment(treatmentID, tenantID int64) ([]ConsumableDTO, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	var rows []legacyConsumableRow
	if err := db.Table(`"Treatment_MaterialTrace"`).
		Where(`"TreatmentId" = ? AND "TenantId" = ?`, treatmentID, tenantID).
		Order(`"CreateTime" DESC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]ConsumableDTO, 0, len(rows))
	for _, r := range rows {
		result = append(result, ConsumableDTO{
			ID: r.ID, TreatmentID: r.TreatmentID, MaterialID: r.ChargeItemID,
			Num: r.Num, Unit: r.Unit, Batch: r.Batch, SerialNo: r.SerialNo,
			Note: r.Note, CreatedAt: r.CreateTime,
		})
	}
	return result, nil
}

func (s *ConsumableService) Create(treatmentID, tenantID, creatorID int64, req CreateConsumableRequest) (*ConsumableDTO, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	newID, err := nextLegacyID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	values := map[string]interface{}{
		`"Id"`: int64(newID), `"TenantId"`: tenantID,
		`"TreatmentId"`: treatmentID, `"ChargeItemId"`: req.MaterialID,
		`"Num"`: req.Num, `"Unit"`: req.Unit,
		`"Batch"`: req.Batch, `"SerialNo"`: req.SerialNo,
		`"Note"`: req.Note, `"CreatorId"`: creatorID,
		`"CreateTime"`: now, `"LastModifyTime"`: now,
	}
	if err := db.Table(`"Treatment_MaterialTrace"`).Create(values).Error; err != nil {
		return nil, fmt.Errorf("create consumable: %w", err)
	}
	return &ConsumableDTO{
		ID: int64(newID), TreatmentID: treatmentID, MaterialID: req.MaterialID,
		Num: req.Num, Unit: req.Unit, Batch: req.Batch, SerialNo: req.SerialNo,
		Note: req.Note, CreatedAt: now,
	}, nil
}

func (s *ConsumableService) Delete(id, tenantID int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not available")
	}
	result := db.Table(`"Treatment_MaterialTrace"`).Where(`"Id" = ? AND "TenantId" = ?`, id, tenantID).Delete(nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("consumable not found")
	}
	return nil
}
