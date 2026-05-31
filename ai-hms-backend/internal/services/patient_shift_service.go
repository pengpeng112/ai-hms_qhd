package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"gorm.io/gorm"
)

// PatientShiftService 患者排班服务
type PatientShiftService struct {
	db *gorm.DB
}

// NewPatientShiftService 创建患者排班服务
func NewPatientShiftService() *PatientShiftService {
	return &PatientShiftService{
		db: database.GetDB(),
	}
}

// ListRequest 获取患者排班列表请求
type PatientShiftListRequest struct {
	Page      int        `form:"page"`
	PageSize  int        `form:"pageSize"`
	PatientId *int64     `form:"patientId"`
	ShiftId   *int64     `form:"shiftId"`
	WardId    *int64     `form:"wardId"`
	BedId     *int64     `form:"bedId"`
	StartDate *time.Time `form:"startDate" time_format:"2006-01-02"`
	EndDate   *time.Time `form:"endDate" time_format:"2006-01-02"`
	Status    *int       `form:"status"`
	TenantId  int64      `form:"-"`
}

// ListResponse 获取患者排班列表响应
type PatientShiftListResponse struct {
	Items     []models.PatientShift `json:"items"`
	Total     int64                 `json:"total"`
	Page      int                   `json:"page"`
	PageSize  int                   `json:"pageSize"`
	TotalPage int                   `json:"totalPage"`
}

// List 获取患者排班列表
func (s *PatientShiftService) List(req PatientShiftListRequest) (*PatientShiftListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.TenantId <= 0 {
		return nil, errors.New("invalid tenant")
	}

	// 默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.PatientShift{}).
		Where("\"TenantId\" = ?", req.TenantId)

	// 筛选条件
	if req.PatientId != nil {
		query = query.Where("\"PatientId\" = ?", *req.PatientId)
	}
	if req.ShiftId != nil {
		query = query.Where("\"ShiftId\" = ?", *req.ShiftId)
	}
	if req.WardId != nil {
		query = query.Where("\"WardId\" = ?", *req.WardId)
	}
	if req.BedId != nil {
		query = query.Where("\"BedId\" = ?", *req.BedId)
	}
	if req.Status != nil {
		query = query.Where("\"Status\" = ?", MapPatientShiftStatusNewToLegacy(*req.Status))
	}
	if req.StartDate != nil {
		query = query.Where("DATE(\"TreatmentTime\") >= DATE(?)", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("DATE(\"TreatmentTime\") <= DATE(?)", *req.EndDate)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var items []models.PatientShift
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Preload("Patient").
		Preload("Shift").
		Preload("Bed").
		Preload("Ward").
		Offset(offset).
		Limit(req.PageSize).
		Order("\"TreatmentTime\" DESC, \"CreateTime\" DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	for i := range items {
		items[i].Status = MapPatientShiftStatusLegacyToNew(items[i].Status)
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &PatientShiftListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// Get 获取患者排班详情
func (s *PatientShiftService) Get(id, tenantId int64) (*models.PatientShift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patientShift models.PatientShift
	err := s.db.
		Preload("Patient").
		Preload("Shift").
		Preload("Bed").
		Preload("Ward").
		Where("\"Id\" = ?", id).
		Where("\"TenantId\" = ?", tenantId).
		First(&patientShift).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient shift not found")
		}
		return nil, err
	}

	patientShift.Status = MapPatientShiftStatusLegacyToNew(patientShift.Status)
	return &patientShift, nil
}

// CreateRequest 创建患者排班请求
type PatientShiftCreateRequest struct {
	PatientId     int64  `json:"patientId" binding:"required"`
	ScheduleDate  string `json:"scheduleDate" binding:"required"`
	ShiftId       int64  `json:"shiftId" binding:"required"`
	BedId         *int64 `json:"bedId"`
	WardId        *int64 `json:"wardId"`
	PatientPlanId *int64 `json:"patientPlanId"`
	ShiftTiming   *int   `json:"shiftTiming"`
	Status        *int   `json:"status"`
	Notes         string `json:"notes"`
}

// Create 创建患者排班
func (s *PatientShiftService) Create(req PatientShiftCreateRequest, tenantId, creatorId int64) (*models.PatientShift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	scheduleDate, err := ParseScheduleDate(req.ScheduleDate)
	if err != nil {
		return nil, err
	}

	status := MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusConfirmed)
	if req.Status != nil {
		status = MapPatientShiftStatusNewToLegacy(*req.Status)
	}

	patientShift := models.PatientShift{
		TenantId:      tenantId,
		PatientId:     modeltypes.LegacyID(req.PatientId),
		ScheduleDate:  scheduleDate,
		ShiftId:       req.ShiftId,
		BedId:         req.BedId,
		WardId:        req.WardId,
		PatientPlanId: req.PatientPlanId,
		ShiftTiming:   req.ShiftTiming,
		Status:        status,
		IsDisabled:    false,
		Notes:         req.Notes,
		CreatorId:     creatorId,
	}

	if err := s.db.Create(&patientShift).Error; err != nil {
		return nil, err
	}
	patientShift.Status = MapPatientShiftStatusLegacyToNew(patientShift.Status)

	return &patientShift, nil
}

// ParseScheduleDate 解析排班日期，支持 YYYY-MM-DD 和 RFC3339
func ParseScheduleDate(v string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("invalid scheduleDate format, expected YYYY-MM-DD or RFC3339")
}

// UpdateRequest 更新患者排班请求
type PatientShiftUpdateRequest struct {
	ShiftId       *int64     `json:"shiftId"`
	BedId         *int64     `json:"bedId"`
	WardId        *int64     `json:"wardId"`
	PatientPlanId *int64     `json:"patientPlanId"`
	ShiftTiming   *int       `json:"shiftTiming"`
	TreatmentTime *time.Time `json:"treatmentTime"`
	Status        *int       `json:"status"`
	Notes         *string    `json:"notes"`
}

// ValidateShiftMutation 统一校验排班变更（create/update/move）
// 返回第一个冲突的 error，无冲突返回 nil
func (s *PatientShiftService) ValidateShiftMutation(patientId int64, bedId, tenantId int64, date time.Time, shiftId int64, excludeId *int64) error {
	hasPatientConflict, err := s.CheckConflict(patientId, tenantId, date, shiftId, excludeId)
	if err != nil {
		return err
	}
	if hasPatientConflict {
		return errors.New("schedule conflict: patient already has shift at the same date and shift")
	}

	if bedId > 0 {
		hasBedConflict, err := s.CheckBedConflict(bedId, tenantId, date, shiftId, excludeId)
		if err != nil {
			return err
		}
		if hasBedConflict {
			return errors.New("schedule conflict: bed already occupied at the same date and shift")
		}
	}

	return nil
}

// Update 更新患者排班
func (s *PatientShiftService) Update(id, tenantId int64, req PatientShiftUpdateRequest) (*models.PatientShift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patientShift models.PatientShift
	if err := s.db.Where("\"Id\" = ?", id).Where("\"TenantId\" = ?", tenantId).First(&patientShift).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient shift not found")
		}
		return nil, err
	}

	// 计算生效的 target 值
	effectivePatientId := int64(patientShift.PatientId)
	effectiveDate := patientShift.ScheduleDate
	effectiveShiftId := patientShift.ShiftId
	effectiveBedId := int64(0)
	if patientShift.BedId != nil {
		effectiveBedId = *patientShift.BedId
	}

	changed := false
	if req.TreatmentTime != nil {
		effectiveDate = *req.TreatmentTime
		changed = true
	}
	if req.ShiftId != nil && *req.ShiftId != patientShift.ShiftId {
		effectiveShiftId = *req.ShiftId
		changed = true
	}
	if req.BedId != nil {
		effectiveBedId = *req.BedId
		changed = true
	}

	// 任何关键字段变化时重算冲突
	if changed {
		excludeId := id
		if err := s.ValidateShiftMutation(effectivePatientId, effectiveBedId, tenantId, effectiveDate, effectiveShiftId, &excludeId); err != nil {
			return nil, err
		}
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.ShiftId != nil {
		updates["ShiftId"] = *req.ShiftId
	}
	if req.BedId != nil {
		updates["BedId"] = *req.BedId
	}
	if req.WardId != nil {
		updates["WardId"] = *req.WardId
	}
	if req.PatientPlanId != nil {
		updates["PatientPlanId"] = *req.PatientPlanId
	}
	if req.ShiftTiming != nil {
		updates["ShiftTiming"] = *req.ShiftTiming
	}
	if req.TreatmentTime != nil {
		updates["TreatmentTime"] = *req.TreatmentTime
	}
	if req.Status != nil {
		updates["Status"] = MapPatientShiftStatusNewToLegacy(*req.Status)
	}

	if err := s.db.Model(&patientShift).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的数据
	if err := s.db.
		Preload("Patient").
		Preload("Shift").
		Preload("Bed").
		Preload("Ward").
		Where("\"Id\" = ?", id).
		Where("\"TenantId\" = ?", tenantId).
		First(&patientShift).Error; err != nil {
		return nil, err
	}

	patientShift.Status = MapPatientShiftStatusLegacyToNew(patientShift.Status)

	return &patientShift, nil
}

// Delete 删除患者排班
func (s *PatientShiftService) Delete(id, tenantId int64) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Model(&models.PatientShift{}).
		Where("\"Id\" = ?", id).
		Where("\"TenantId\" = ?", tenantId).
		Update("Status", MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusCancelled))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("patient shift not found")
	}

	return nil
}

// Swap 互换两个排班的床位/日期/班次（事务）
func (s *PatientShiftService) Swap(sourceID, targetID, tenantId int64) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var src, tgt models.PatientShift
		if err := tx.Where("\"Id\" = ? AND \"TenantId\" = ?", sourceID, tenantId).First(&src).Error; err != nil {
			return errors.New("source shift not found")
		}
		if err := tx.Where("\"Id\" = ? AND \"TenantId\" = ?", targetID, tenantId).First(&tgt).Error; err != nil {
			return errors.New("target shift not found")
		}

		srcUpdates := map[string]interface{}{
			"WardId":        tgt.WardId,
			"BedId":         tgt.BedId,
			"ShiftId":       tgt.ShiftId,
			"TreatmentTime": tgt.ScheduleDate,
		}
		tgtUpdates := map[string]interface{}{
			"WardId":        src.WardId,
			"BedId":         src.BedId,
			"ShiftId":       src.ShiftId,
			"TreatmentTime": src.ScheduleDate,
		}

		if err := tx.Model(&models.PatientShift{}).
			Where("\"Id\" = ? AND \"TenantId\" = ?", sourceID, tenantId).
			Updates(srcUpdates).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.PatientShift{}).
			Where("\"Id\" = ? AND \"TenantId\" = ?", targetID, tenantId).
			Updates(tgtUpdates).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetByPatientAndDate 根据患者ID和日期获取排班
func (s *PatientShiftService) GetByPatientAndDate(patientId, tenantId int64, date time.Time) (*models.PatientShift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patientShift models.PatientShift
	err := s.db.
		Where("\"TenantId\" = ?", tenantId).
		Where("\"PatientId\" = ? AND DATE(\"TreatmentTime\") = DATE(?)", patientId, date).
		Preload("Shift").
		Preload("Bed").
		Preload("Ward").
		First(&patientShift).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有排班记录
		}
		return nil, err
	}

	patientShift.Status = MapPatientShiftStatusLegacyToNew(patientShift.Status)
	return &patientShift, nil
}

// CheckConflict 检查排班冲突（患者+日期+班次维度）
func (s *PatientShiftService) CheckConflict(patientId, tenantId int64, date time.Time, shiftId int64, excludeId *int64) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	query := s.db.Model(&models.PatientShift{}).
		Where("\"TenantId\" = ?", tenantId).
		Where("\"PatientId\" = ? AND DATE(\"TreatmentTime\") = DATE(?) AND \"ShiftId\" = ?", patientId, date, shiftId).
		Where("\"Status\" NOT IN (?)", []int{
			MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusCancelled),
			MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusUserCancelled),
			MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusTransferred),
		})

	if excludeId != nil {
		query = query.Where("\"Id\" != ?", *excludeId)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// CheckBedConflict 检查床位冲突（床位+日期+班次维度）
func (s *PatientShiftService) CheckBedConflict(bedId, tenantId int64, date time.Time, shiftId int64, excludeId *int64) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	query := s.db.Model(&models.PatientShift{}).
		Where("\"TenantId\" = ?", tenantId).
		Where("\"BedId\" = ? AND DATE(\"TreatmentTime\") = DATE(?) AND \"ShiftId\" = ?", bedId, date, shiftId).
		Where("\"Status\" NOT IN (?)", []int{
			MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusCancelled),
			MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusUserCancelled),
			MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusTransferred),
		})

	if excludeId != nil {
		query = query.Where("\"Id\" != ?", *excludeId)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ─── P0-7 老系统四类校验 ───

// isHistoryDate 判断是否为历史日期（< 今天）
func isHistoryDate(d time.Time) bool {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return d.Before(today)
}

// ValidateShiftEdit 编辑/修改/换床/互换前校验（历史保护 + 已过班次保护 + 治疗中保护）
func (s *PatientShiftService) ValidateShiftEdit(shift *models.PatientShift, tenantId int64) error {
	if shift == nil {
		return errors.New("shift not found")
	}

	// 1) 历史数据保护
	if isHistoryDate(shift.ScheduleDate) {
		return errors.New("历史排班不可修改或取消")
	}

	// 2) 已过班次保护（仅当天排班）
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if shift.ScheduleDate.Year() == today.Year() && shift.ScheduleDate.Month() == today.Month() && shift.ScheduleDate.Day() == today.Day() {
		currentTimeStr := now.Format("15:04")
		var currentShift models.Shift
		err := s.db.Where("\"TenantId\" = ?", tenantId).
			Where("\"IsDisabled\" = false").
			Where("\"StartTime\"::time <= ?::time AND \"EndTime\"::time > ?::time", currentTimeStr, currentTimeStr).
			Order("\"Sort\" ASC").
			First(&currentShift).Error
		if err == nil {
			var targetShift models.Shift
			if err2 := s.db.Where("\"Id\" = ? AND \"TenantId\" = ?", shift.ShiftId, tenantId).First(&targetShift).Error; err2 == nil {
				if targetShift.Sort < currentShift.Sort {
					return errors.New("该排班所属班次已过，不可修改或取消")
				}
			}
		}
	}

	// 3) 治疗中保护
	var treatIDs []struct {
		ID int64 `gorm:"column:Id"`
	}
	if err := s.db.Table(`"Treatment_Treatment"`).
		Select(`"Id"`).
		Where(`"ScheduleId" = ? AND "TenantId" = ?`, shift.Id, tenantId).
		Find(&treatIDs).Error; err != nil {
		return err
	}
	for _, tr := range treatIDs {
		var actionCount int64
		s.db.Table(`"Treatment_Action"`).
			Where(`"TreatmentId" = ? AND "Code"::int >= 20`, tr.ID).
			Count(&actionCount)
		if actionCount > 0 {
			return errors.New("该排班已有治疗记录且已完成透前评估，不可修改或取消")
		}
	}

	return nil
}
