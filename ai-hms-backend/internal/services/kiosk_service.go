package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KioskPreSignsRequest 自助站透前体征上报请求。
type KioskPreSignsRequest struct {
	TreatmentID   int64     `json:"treatmentId" binding:"required"`
	PatientID     int64     `json:"patientId" binding:"required"`
	MeasuredAt    time.Time `json:"measuredAt" binding:"required"`
	Weight        *float64  `json:"weight,omitempty"`
	SBP           *float64  `json:"sbp,omitempty"`
	DBP           *float64  `json:"dbp,omitempty"`
	BodyTemp      *float64  `json:"bodyTemp,omitempty"`
	HeartRate     *float64  `json:"heartRate,omitempty"`
	Respiration   *float64  `json:"respiration,omitempty"`
	DeviceID      string    `json:"deviceId,omitempty"`
	ClientEventID string    `json:"clientEventId,omitempty"`
}

// KioskCheckInRequest 自助站签到请求。
type KioskCheckInRequest struct {
	TreatmentID int64 `json:"treatmentId" binding:"required"`
	PatientID   int64 `json:"patientId" binding:"required"`
}

// KioskService 自助站接口服务。
type KioskService struct {
	db *gorm.DB
}

func NewKioskService() *KioskService {
	return &KioskService{db: database.GetDB()}
}

func newKioskServiceWithDB(db *gorm.DB) *KioskService {
	return &KioskService{db: db}
}

func (s *KioskService) dbOrErr() (*gorm.DB, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	return s.db, nil
}

// SavePreSigns 保存自助站上传的透前体征。
func (s *KioskService) SavePreSigns(req KioskPreSignsRequest) error {
	db, err := s.dbOrErr()
	if err != nil {
		return err
	}

	now := time.Now()

	// 如果提供了 client_event_id，先检查是否已存在（幂等）。
	if req.ClientEventID != "" {
		var existing int64
		if err := db.Table("kiosk_pre_sign_measurement").
			Where("tenant_id = ? AND client_event_id = ?", LegacyTenantID, req.ClientEventID).
			Count(&existing).Error; err != nil {
			return fmt.Errorf("检查幂等键失败: %w", err)
		}
		if existing > 0 {
			return nil
		}
	}

	// 验证治疗存在且属于该患者。
	var treatmentExists int64
	if err := db.Table(`"Treatment_Treatment"`).
		Where(`"Id" = ? AND "TenantId" = ? AND "PatientId" = ?`, req.TreatmentID, LegacyTenantID, req.PatientID).
		Count(&treatmentExists).Error; err != nil {
		return fmt.Errorf("查询治疗记录失败: %w", err)
	}
	if treatmentExists == 0 {
		return errors.New("治疗记录不存在或不属于该患者")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// 1. 追加明细表。
		id := uuid.New().String()
		detail := map[string]any{
			"id":              id,
			"tenant_id":       LegacyTenantID,
			"treatment_id":    req.TreatmentID,
			"patient_id":      req.PatientID,
			"measured_at":     req.MeasuredAt,
			"weight":          req.Weight,
			"sbp":             req.SBP,
			"dbp":             req.DBP,
			"body_temp":       req.BodyTemp,
			"heart_rate":      req.HeartRate,
			"respiration":     req.Respiration,
			"device_id":       req.DeviceID,
			"source":          "newsystem",
			"client_event_id": req.ClientEventID,
			"raw_payload":     nil,
			"created_at":      now,
		}
		// 去掉空 client_event_id。
		if req.ClientEventID == "" {
			delete(detail, "client_event_id")
		}
		if err := tx.Table("kiosk_pre_sign_measurement").Create(detail).Error; err != nil {
			return fmt.Errorf("追加明细失败: %w", err)
		}

		// 2. 同步最新值到 Treatment_BeforeSigns。
		if err := s.upsertBeforeSigns(tx, req, now); err != nil {
			return err
		}

		return nil
	})
}

func (s *KioskService) upsertBeforeSigns(tx *gorm.DB, req KioskPreSignsRequest, now time.Time) error {
	type existingRow struct {
		ID   int64
		Note *string
	}
	var existing existingRow
	err := tx.Table(`"Treatment_BeforeSigns"`).
		Select(`"Id", "Note"`).
		Where(`"TenantId" = ? AND "TreatmentId" = ?`, LegacyTenantID, req.TreatmentID).
		First(&existing).Error

	newNote := "source=newsystem"
	if err == nil {
		// 已有行：更新，保留原 Note 并追加来源。
		if existing.Note != nil && strings.TrimSpace(*existing.Note) != "" {
			newNote = strings.TrimSpace(*existing.Note) + "; " + newNote
		}
		updates := map[string]any{
			"Weight":         req.Weight,
			"SBP":            req.SBP,
			"DBP":            req.DBP,
			"BodyTemp":       req.BodyTemp,
			"HeartRate":      req.HeartRate,
			"Respiration":    req.Respiration,
			"OperateTime":    req.MeasuredAt,
			"LastModifyTime": now,
			"Note":           newNote,
		}
		if err := tx.Table(`"Treatment_BeforeSigns"`).
			Where(`"Id" = ?`, existing.ID).
			Updates(updates).Error; err != nil {
			return fmt.Errorf("同步旧表失败: %w", err)
		}
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询旧表失败: %w", err)
	}

	// 无行：插入。
	id, err := nextLegacyID()
	if err != nil {
		return fmt.Errorf("生成旧表ID失败: %w", err)
	}
	values := map[string]any{
		"Id":             int64(id),
		"TenantId":       LegacyTenantID,
		"TreatmentId":    req.TreatmentID,
		"Weight":         req.Weight,
		"ExtraWeight":    nil,
		"BodyTemp":       req.BodyTemp,
		"SBP":            req.SBP,
		"DBP":            req.DBP,
		"PressurePoint":  nil,
		"HeartRate":      req.HeartRate,
		"Respiration":    req.Respiration,
		"OperatorId":     nil,
		"OperateTime":    req.MeasuredAt,
		"CreatorId":      nil,
		"CreateTime":     now,
		"LastModifyTime": now,
		"Note":           newNote,
	}
	if err := tx.Table(`"Treatment_BeforeSigns"`).Create(values).Error; err != nil {
		return fmt.Errorf("插入旧表失败: %w", err)
	}
	return nil
}

// KioskLookupQuery 自助站患者查询参数。
type KioskLookupQuery struct {
	DialysisNo string `form:"dialysisNo"`
	IDNo       string `form:"idNo"`
	PatientID  *int64 `form:"patientId"`
}

// KioskLookupResult 自助站患者查询结果。
type KioskLookupResult struct {
	PatientID        int64  `json:"patientId"`
	PatientName      string `json:"patientName"`
	Gender           string `json:"gender"`
	BirthDate        string `json:"birthDate"`
	DialysisNo       string `json:"dialysisNo"`
	TodayTreatmentID *int64 `json:"todayTreatmentId,omitempty"`
}

// LookupPatient 自助站侧按透析号/身份证/患者ID查患者及当天治疗。
func (s *KioskService) LookupPatient(q KioskLookupQuery) (*KioskLookupResult, error) {
	db, err := s.dbOrErr()
	if err != nil {
		return nil, err
	}

	type patientRow struct {
		ID         int64      `gorm:"column:Id"`
		Name       string     `gorm:"column:Name"`
		Gender     string     `gorm:"column:Gender"`
		BirthDate  *time.Time `gorm:"column:BirthDate"`
		DialysisNo string     `gorm:"column:DialysisNo"`
		HospitalNo string     `gorm:"column:HospitalNo"`
	}
	var pat patientRow
	query := db.Table(`"Register_PatientInfomation"`).
		Select(`"Id","Name","Gender","BirthDate","DialysisNo","HospitalNo"`).
		Where(`"TenantId" = ?`, LegacyTenantID)

	switch {
	case q.PatientID != nil:
		query = query.Where(`"Id" = ?`, *q.PatientID)
	case q.DialysisNo != "":
		query = query.Where(`"DialysisNo" = ?`, q.DialysisNo)
	case q.IDNo != "":
		query = query.Where(`"IDNo" = ?`, q.IDNo)
	default:
		return nil, errors.New("查询参数不能为空: 请提供 dialysisNo, idNo 或 patientId")
	}

	if err := query.First(&pat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("未找到患者")
		}
		return nil, fmt.Errorf("查询患者失败: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	type treatmentRow struct {
		ID int64 `gorm:"column:Id"`
	}
	var trt treatmentRow
	err = db.Table(`"Treatment_Treatment"`).
		Select(`"Id"`).
		Where(`"PatientId" = ? AND "TenantId" = ? AND DATE(COALESCE("StartTime","SignInTime","ReceptionTime","CreateTime")) = ?`,
			pat.ID, LegacyTenantID, today).
		Order(`COALESCE("StartTime","SignInTime","ReceptionTime","CreateTime") DESC`).
		Limit(1).
		First(&trt).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询当天治疗失败: %w", err)
	}

	result := &KioskLookupResult{
		PatientID:   pat.ID,
		PatientName: pat.Name,
		Gender:      pat.Gender,
		DialysisNo:  pat.DialysisNo,
	}
	if err == nil {
		result.TodayTreatmentID = &trt.ID
	}
	if pat.BirthDate != nil {
		result.BirthDate = pat.BirthDate.Format("2006-01-02")
	}
	return result, nil
}

// CheckIn 自助站签到。
func (s *KioskService) CheckIn(req KioskCheckInRequest) error {
	db, err := s.dbOrErr()
	if err != nil {
		return err
	}

	now := time.Now()

	type treatmentInfo struct {
		Status     string
		SignInTime *time.Time
	}
	var info treatmentInfo
	err = db.Table(`"Treatment_Treatment"`).
		Select(`"Status", "SignInTime"`).
		Where(`"Id" = ? AND "TenantId" = ? AND "PatientId" = ?`, req.TreatmentID, LegacyTenantID, req.PatientID).
		First(&info).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("治疗记录不存在或不属于该患者")
		}
		return fmt.Errorf("查询治疗记录失败: %w", err)
	}

	status := strings.TrimSpace(info.Status)

	// 中断或已结束，不允许签到。
	if status == "50" || status == "60" {
		return errors.New("治疗已中断或已结束，无法签到")
	}

	// 已有 SignInTime 则视为已签到，幂等返回成功。
	if info.SignInTime != nil {
		return nil
	}

	// 补写 SignInTime，不改 Status。
	updates := map[string]any{
		"SignInTime":     now,
		"LastModifyTime": now,
	}
	if err := db.Table(`"Treatment_Treatment"`).
		Where(`"Id" = ? AND "TenantId" = ?`, req.TreatmentID, LegacyTenantID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("签到更新失败: %w", err)
	}
	return nil
}

// Health 健康检查（内存中检查数据库连通性）。
func (s *KioskService) Health() map[string]any {
	db, err := s.dbOrErr()
	if err != nil {
		return map[string]any{"ok": false, "db": err.Error()}
	}
	sqlDB, err2 := db.DB()
	if err2 != nil {
		return map[string]any{"ok": false, "db": "error", "error": err2.Error()}
	}
	if err := sqlDB.Ping(); err != nil {
		return map[string]any{"ok": false, "db": "error", "error": err.Error()}
	}
	return map[string]any{"ok": true, "db": "ok"}
}
