package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PrescriptionService 处方服务
type PrescriptionService struct {
	db *gorm.DB
}

// NewPrescriptionService 创建处方服务
func NewPrescriptionService() *PrescriptionService {
	return &PrescriptionService{
		db: database.GetDB(),
	}
}

// PrescriptionCreateRequest 创建处方请求
type PrescriptionCreateRequest struct {
	PrescriptionDate string                          `json:"prescriptionDate" binding:"required"`
	Duration         int                             `json:"duration"`
	DryWeight        float64                         `json:"dryWeight"`
	ExtraWeight      float64                         `json:"extraWeight"`
	DialysisMode     models.DialysisMode             `json:"dialysisMode"`
	Anticoagulant    models.Anticoagulant            `json:"anticoagulant"`
	Parameters       models.DialysisParameters       `json:"parameters"`
	Materials        models.MaterialList             `json:"materials"`
	OrderItems       models.PrescriptionOrderItemList `json:"orderItems"`
	Notes            string                          `json:"notes"`
}

// PrescriptionUpdateRequest 更新处方请求（全量替换语义）
type PrescriptionUpdateRequest struct {
	Duration      *int                              `json:"duration"`
	DryWeight     *float64                          `json:"dryWeight"`
	ExtraWeight   *float64                          `json:"extraWeight"`
	DialysisMode  *models.DialysisMode              `json:"dialysisMode"`
	Anticoagulant *models.Anticoagulant             `json:"anticoagulant"`
	Parameters    *models.DialysisParameters        `json:"parameters"`
	Materials     *models.MaterialList              `json:"materials"`
	OrderItems    *models.PrescriptionOrderItemList  `json:"orderItems"` // 全量替换
	Notes         *string                           `json:"notes"`
}

// PrescriptionExtractRequest 提取长嘱请求
type PrescriptionExtractRequest struct {
	Date string `json:"date" binding:"required"` // yyyy-MM-dd
}

// getActiveTreatmentPlan 获取患者启用的治疗方案（updated_at DESC 取第一条）
func (s *PrescriptionService) getActiveTreatmentPlan(patientID string) (*models.TreatmentPlan, error) {
	var plan models.TreatmentPlan
	err := s.db.Where("patient_id = ? AND status = ?", patientID, models.TreatmentPlanStatusActive).
		Order("updated_at DESC").
		First(&plan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("请先创建启用的治疗方案")
		}
		return nil, err
	}
	return &plan, nil
}

// List 获取处方列表（按日期倒序）
func (s *PrescriptionService) List(patientID string) ([]models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var prescriptions []models.Prescription
	err := s.db.Where("patient_id = ?", patientID).
		Order("prescription_date DESC").
		Find(&prescriptions).Error
	return prescriptions, err
}

// Get 获取处方详情（患者隔离）
func (s *PrescriptionService) Get(patientID, prescriptionID string) (*models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var p models.Prescription
	err := s.db.First(&p, "id = ? AND patient_id = ?", prescriptionID, patientID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("prescription not found")
		}
		return nil, err
	}
	return &p, nil
}

// Create 创建处方
func (s *PrescriptionService) Create(patientID, doctorID, doctorName string, req PrescriptionCreateRequest) (*models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 自动填充 TreatmentPlanID
	plan, err := s.getActiveTreatmentPlan(patientID)
	if err != nil {
		return nil, err
	}

	date, err := time.Parse("2006-01-02", req.PrescriptionDate)
	if err != nil {
		return nil, errors.New("日期格式错误，应为 yyyy-MM-dd")
	}

	p := models.Prescription{
		ID:               uuid.New().String(),
		PatientID:        patientID,
		TreatmentPlanID:  plan.ID,
		PrescriptionDate: date,
		DoctorID:         doctorID,
		DoctorName:       doctorName,
		Status:           models.PrescriptionStatusPending,
		Duration:         req.Duration,
		DryWeight:        req.DryWeight,
		ExtraWeight:      req.ExtraWeight,
		DialysisMode:     req.DialysisMode,
		Anticoagulant:    req.Anticoagulant,
		Parameters:       req.Parameters,
		Materials:        req.Materials,
		OrderItems:       req.OrderItems,
		Notes:            req.Notes,
	}

	if err := s.db.Create(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// Update 更新处方（仅允许待执行状态，OrderItems 全量替换）
func (s *PrescriptionService) Update(patientID, prescriptionID string, req PrescriptionUpdateRequest) (*models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var p models.Prescription
	if err := s.db.First(&p, "id = ? AND patient_id = ?", prescriptionID, patientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("prescription not found")
		}
		return nil, err
	}

	if p.Status != models.PrescriptionStatusPending {
		return nil, errors.New("仅待执行状态的处方可编辑")
	}

	updates := make(map[string]interface{})
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.DryWeight != nil {
		updates["dry_weight"] = *req.DryWeight
	}
	if req.ExtraWeight != nil {
		updates["extra_weight"] = *req.ExtraWeight
	}
	if req.DialysisMode != nil {
		updates["dialysis_mode"] = *req.DialysisMode
	}
	if req.Anticoagulant != nil {
		updates["anticoagulant"] = *req.Anticoagulant
	}
	if req.Parameters != nil {
		updates["parameters"] = *req.Parameters
	}
	if req.Materials != nil {
		updates["materials"] = *req.Materials
	}
	if req.OrderItems != nil {
		updates["order_items"] = *req.OrderItems
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if len(updates) > 0 {
		if err := s.db.Model(&p).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	if err := s.db.First(&p, "id = ?", prescriptionID).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// Execute 标记处方已执行（幂等）
func (s *PrescriptionService) Execute(patientID, prescriptionID, executedBy string) (*models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var p models.Prescription
	if err := s.db.First(&p, "id = ? AND patient_id = ?", prescriptionID, patientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("prescription not found")
		}
		return nil, err
	}

	// 幂等：已执行不报错
	if p.Status == models.PrescriptionStatusExecuted {
		return &p, nil
	}

	// 只有 待执行/执行中 可以标记执行
	if p.Status == models.PrescriptionStatusCancelled {
		return nil, errors.New("已取消的处方不能执行")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      models.PrescriptionStatusExecuted,
		"executed_at":  now,
		"executed_by":  executedBy,
	}

	if err := s.db.Model(&p).Updates(updates).Error; err != nil {
		return nil, err
	}

	p.Status = models.PrescriptionStatusExecuted
	p.ExecutedAt = &now
	p.ExecutedBy = &executedBy
	return &p, nil
}

// Cancel 取消处方（幂等）
func (s *PrescriptionService) Cancel(patientID, prescriptionID string) (*models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var p models.Prescription
	if err := s.db.First(&p, "id = ? AND patient_id = ?", prescriptionID, patientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("prescription not found")
		}
		return nil, err
	}

	// 幂等：已取消不报错
	if p.Status == models.PrescriptionStatusCancelled {
		return &p, nil
	}

	if p.Status == models.PrescriptionStatusExecuted {
		return nil, errors.New("已执行的处方不能取消")
	}

	if err := s.db.Model(&p).Update("status", models.PrescriptionStatusCancelled).Error; err != nil {
		return nil, err
	}

	p.Status = models.PrescriptionStatusCancelled
	return &p, nil
}

// ExtractFromLongTermOrders 从在用长期医嘱提取处方
func (s *PrescriptionService) ExtractFromLongTermOrders(patientID, doctorID, doctorName, dateStr string) (*models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("日期格式错误，应为 yyyy-MM-dd")
	}

	// 获取启用的治疗方案
	plan, err := s.getActiveTreatmentPlan(patientID)
	if err != nil {
		return nil, err
	}

	// 查询在用长期医嘱
	var orders []models.Order
	if err := s.db.Where("patient_id = ? AND type = ? AND status IN ?",
		patientID,
		models.OrderTypeLongTerm,
		[]string{models.OrderStatusPending, models.OrderStatusExecuting},
	).Order("created_at ASC").Find(&orders).Error; err != nil {
		return nil, err
	}

	// 生成 OrderItems 快照
	var orderItems models.PrescriptionOrderItemList
	for _, o := range orders {
		orderItems = append(orderItems, models.PrescriptionOrderItem{
			OrderID:   o.ID,
			Name:      o.Name,
			Category:  o.Category,
			Dose:      o.Dose,
			Unit:      o.Unit,
			Frequency: func() string { if o.Frequency != nil { return *o.Frequency }; return "" }(),
			Route:     o.Route,
			Spec:      o.Spec,
		})
	}

	// 创建处方，复制治疗方案参数
	p := models.Prescription{
		ID:               uuid.New().String(),
		PatientID:        patientID,
		TreatmentPlanID:  plan.ID,
		PrescriptionDate: date,
		DoctorID:         doctorID,
		DoctorName:       doctorName,
		Status:           models.PrescriptionStatusPending,
		Duration:         plan.Duration,
		DryWeight:        plan.DryWeight,
		ExtraWeight:      plan.ExtraWeight,
		DialysisMode:     plan.DialysisMode,
		Anticoagulant:    plan.Anticoagulant,
		Parameters:       plan.DialysisParameters,
		Materials:        plan.Materials,
		OrderItems:       orderItems,
		Notes:            "",
	}

	if err := s.db.Create(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
