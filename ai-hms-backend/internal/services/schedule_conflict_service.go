package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

// ScheduleConflictService 冲突队列服务
type ScheduleConflictService struct {
	db *gorm.DB
}

func NewScheduleConflictService() *ScheduleConflictService {
	return &ScheduleConflictService{db: database.GetDB()}
}

func (s *ScheduleConflictService) dbCheck() error {
	if s.db == nil {
		return errors.New("database not available")
	}
	return nil
}

// ListConflicts 获取冲突列表
func (s *ScheduleConflictService) ListConflicts(tenantID int64, page, pageSize int, conflictType string, status *int16) ([]models.ConflictQueue, int64, error) {
	if err := s.dbCheck(); err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	q := s.db.Model(&models.ConflictQueue{}).Where(`"TenantId" = ?`, tenantID)
	if conflictType != "" {
		q = q.Where(`"ConflictType" = ?`, conflictType)
	}
	if status != nil {
		q = q.Where(`"Status" = ?`, *status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.ConflictQueue
	offset := (page - 1) * pageSize
	if err := q.Order(`"CreateTime" DESC`).Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ResolveConflict 解决冲突
func (s *ScheduleConflictService) ResolveConflict(tenantID, conflictID, resolverID int64, note string) error {
	if err := s.dbCheck(); err != nil {
		return err
	}
	now := time.Now()
	return s.db.Model(&models.ConflictQueue{}).
		Where(`"TenantId" = ? AND "Id" = ?`, tenantID, conflictID).
		Updates(map[string]interface{}{
			"Status":    1,
			"ResolvedBy": resolverID,
			"ResolvedAt": now,
			"Detail":    note,
		}).Error
}

// IgnoreConflict 忽略冲突
func (s *ScheduleConflictService) IgnoreConflict(tenantID, conflictID, resolverID int64, note string) error {
	if err := s.dbCheck(); err != nil {
		return err
	}
	now := time.Now()
	return s.db.Model(&models.ConflictQueue{}).
		Where(`"TenantId" = ? AND "Id" = ?`, tenantID, conflictID).
		Updates(map[string]interface{}{
			"Status":    2,
			"ResolvedBy": resolverID,
			"ResolvedAt": now,
			"Detail":    note,
		}).Error
}

// CancelPatientShift 取消排班(请假)
func (s *ScheduleConflictService) CancelPatientShift(tenantID, shiftID int64, reason string) error {
	if err := s.dbCheck(); err != nil {
		return err
	}
	return s.db.Model(&models.PatientShift{}).
		Where(`"TenantId" = ? AND "Id" = ?`, tenantID, shiftID).
		Updates(map[string]interface{}{
			"Status": MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusCancelled),
			"Notes":  "取消原因: " + reason,
		}).Error
}

// MarkAbsent 标记缺席
func (s *ScheduleConflictService) MarkAbsent(tenantID, shiftID int64, reason string) error {
	if err := s.dbCheck(); err != nil {
		return err
	}
	return s.db.Model(&models.PatientShift{}).
		Where(`"TenantId" = ? AND "Id" = ?`, tenantID, shiftID).
		Updates(map[string]interface{}{
			"Status": MapPatientShiftStandardStatusToLegacy(models.StdStatusAbsent),
			"Notes":  "缺席原因: " + reason,
		}).Error
}
