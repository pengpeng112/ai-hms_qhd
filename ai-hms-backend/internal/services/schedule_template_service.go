package services

import (
	"errors"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var inactivePatientShiftLegacyStatuses = []int{
	MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusCancelled),
	MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusUserCancelled),
	MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusTransferred),
}

// TemplateBusinessError 模板业务校验错误
type TemplateBusinessError struct{ msg string }

func (e *TemplateBusinessError) Error() string { return e.msg }
func newBusinessError(msg string) error        { return &TemplateBusinessError{msg: msg} }

// IsTemplateBusinessError 检查是否为模板业务校验错误
func IsTemplateBusinessError(err error) bool {
	var be *TemplateBusinessError
	return errors.As(err, &be)
}

type ScheduleTemplateService struct {
	db *gorm.DB
}

func NewScheduleTemplateService() *ScheduleTemplateService {
	return &ScheduleTemplateService{db: database.GetDB()}
}

// ===================== 请求/响应结构 =====================

type ScheduleTemplateItemRequest struct {
	PatientId     int64  `json:"patientId" binding:"required"`
	ZoneTag       string `json:"zoneTag" binding:"required"`
	WardId        *int64 `json:"wardId"`
	ShiftId       *int64 `json:"shiftId"`
	FreqPattern   int16  `json:"freqPattern"`
	FixedHdBedId  *int64 `json:"fixedHdBedId"`
	FixedHdfBedId *int64 `json:"fixedHdfBedId"`
	HdfEnabled    bool   `json:"hdfEnabled"`
	HdfWeekday    *int16 `json:"hdfWeekday"`
	HdfWeekParity *int16 `json:"hdfWeekParity"`
}

type ScheduleTemplateSaveRequest struct {
	Id     int64                         `json:"id"`
	Name   string                        `json:"name" binding:"required"`
	Scope  string                        `json:"scope"`
	WardId *int64                        `json:"wardId"`
	Items  []ScheduleTemplateItemRequest `json:"items" binding:"required"`
}

type ScheduleTemplateApplyRequest struct {
	TemplateId int64   `json:"templateId" binding:"required"`
	TargetDate string  `json:"targetDate" binding:"required"`
	WardId     *int64  `json:"wardId"`
	ItemIds    []int64 `json:"itemIds"`
}

type ScheduleTemplateResponse struct {
	Template  models.ScheduleTemplate       `json:"template"`
	Items     []models.ScheduleTemplateItem `json:"items"`
	ItemCount int                           `json:"itemCount"`
}

type ScheduleTemplateApplyResponse struct {
	CreatedShifts    []int64 `json:"createdShifts"`
	CreatedShiftExts []int64 `json:"createdShiftExts"`
	Count            int     `json:"count"`
}

// ===================== ListTemplates =====================

func (s *ScheduleTemplateService) ListTemplates(tenantID int64, wardID *int64) ([]ScheduleTemplateResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var tmpls []models.ScheduleTemplate
	q := s.db.Where(`"TenantId" = ? AND "IsActive" = ?`, tenantID, true)
	if wardID != nil && *wardID != 0 {
		q = q.Where(`("WardId" = ? OR ("WardId" IS NULL AND "Scope" = 'ALL'))`, *wardID)
	}
	if err := q.Order(`"Id" DESC`).Find(&tmpls).Error; err != nil {
		return nil, err
	}

	if len(tmpls) == 0 {
		return []ScheduleTemplateResponse{}, nil
	}

	tmplIDs := make([]int64, len(tmpls))
	for i, t := range tmpls {
		tmplIDs[i] = t.Id
	}

	var items []models.ScheduleTemplateItem
	if err := s.db.Where(`"TenantId" = ? AND "TemplateId" IN ?`, tenantID, tmplIDs).
		Order(`"Id"`).Find(&items).Error; err != nil {
		return nil, err
	}

	itemMap := map[int64][]models.ScheduleTemplateItem{}
	for _, it := range items {
		itemMap[it.TemplateId] = append(itemMap[it.TemplateId], it)
	}

	result := make([]ScheduleTemplateResponse, len(tmpls))
	for i, t := range tmpls {
		its := itemMap[t.Id]
		if its == nil {
			its = []models.ScheduleTemplateItem{}
		}
		result[i] = ScheduleTemplateResponse{
			Template:  t,
			Items:     its,
			ItemCount: len(its),
		}
	}
	return result, nil
}

// ===================== SaveTemplate =====================

func (s *ScheduleTemplateService) SaveTemplate(tenantID, creatorID int64, req ScheduleTemplateSaveRequest) (*ScheduleTemplateResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if len(req.Items) == 0 {
		return nil, newBusinessError("模板项不能为空")
	}
	if req.Name == "" {
		return nil, newBusinessError("模板名称不能为空")
	}
	if req.Scope != "" {
		scope := strings.ToUpper(strings.TrimSpace(req.Scope))
		if scope != "ALL" && scope != "A" && scope != "B" && scope != "C" {
			return nil, newBusinessError("Scope 必须为 ALL、A、B 或 C")
		}
		req.Scope = scope
	}

	// 校验模板项 ZoneTag
	for _, it := range req.Items {
		zt := strings.ToUpper(strings.TrimSpace(it.ZoneTag))
		if zt != "A" && zt != "B" && zt != "C" {
			return nil, newBusinessError("模板项 ZoneTag 必须为 A、B 或 C")
		}
		if req.Scope != "" && req.Scope != "ALL" && req.Scope != zt {
			return nil, newBusinessError("模板 Scope 与模板项 ZoneTag 不匹配")
		}
		if it.FreqPattern != 0 && it.FreqPattern != 10 && it.FreqPattern != 20 && it.FreqPattern != 30 && it.FreqPattern != 40 && it.FreqPattern != 90 {
			return nil, newBusinessError("FreqPattern 必须为 10/20/30/40/90 或 0(默认)")
		}
	}

	var tmpl models.ScheduleTemplate
	var result *ScheduleTemplateResponse

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		if req.Id > 0 {
			if err := tx.Where(`"TenantId" = ? AND "Id" = ?`, tenantID, req.Id).First(&tmpl).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return newBusinessError("模板不存在")
				}
				return err
			}
			tmpl.Name = req.Name
			tmpl.Scope = req.Scope
			tmpl.WardId = req.WardId
			tmpl.Version = tmpl.Version + 1
			if err := tx.Save(&tmpl).Error; err != nil {
				return err
			}

			// 删除旧模板项
			if err := tx.Where(`"TenantId" = ? AND "TemplateId" = ?`, tenantID, tmpl.Id).Delete(&models.ScheduleTemplateItem{}).Error; err != nil {
				return err
			}
		} else {
			tmpl = models.ScheduleTemplate{
				TenantId:  tenantID,
				Name:      req.Name,
				Scope:     req.Scope,
				WardId:    req.WardId,
				IsActive:  true,
				Version:   1,
				CreatorId: creatorID,
			}
			if err := tx.Create(&tmpl).Error; err != nil {
				return err
			}
		}

		// 去重：同一 TemplateId + PatientId 只保留一条
		seenPatient := map[int64]bool{}
		items := make([]models.ScheduleTemplateItem, 0, len(req.Items))
		for _, it := range req.Items {
			if seenPatient[it.PatientId] {
				continue
			}
			seenPatient[it.PatientId] = true

			item := models.ScheduleTemplateItem{
				TenantId:        tenantID,
				TemplateId:      tmpl.Id,
				PatientId:       it.PatientId,
				ZoneTag:         strings.ToUpper(strings.TrimSpace(it.ZoneTag)),
				WardId:          it.WardId,
				ShiftId:         it.ShiftId,
				FreqPattern:     it.FreqPattern,
				FixedHdBedId:    it.FixedHdBedId,
				FixedHdfBedId:   it.FixedHdfBedId,
				HdfEnabled:      it.HdfEnabled,
				HdfWeekday:      it.HdfWeekday,
				HdfWeekParity:   it.HdfWeekParity,
				TemplateVersion: tmpl.Version,
				CreatorId:       creatorID,
			}
			if item.FreqPattern == 0 {
				item.FreqPattern = 10
			}
			if item.ZoneTag == "" {
				item.ZoneTag = "A"
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
			items = append(items, item)
		}

		result = &ScheduleTemplateResponse{
			Template:  tmpl,
			Items:     items,
			ItemCount: len(items),
		}
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}
	return result, nil
}

// ===================== ApplyTemplate =====================

func (s *ScheduleTemplateService) ApplyTemplate(tenantID, creatorID int64, req ScheduleTemplateApplyRequest) (*ScheduleTemplateApplyResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	targetDate, err := ParseScheduleDate(req.TargetDate)
	if err != nil {
		return nil, newBusinessError("targetDate 格式应为 YYYY-MM-DD")
	}

	// 加载模板
	var tmpl models.ScheduleTemplate
	if err := s.db.Where(`"TenantId" = ? AND "Id" = ? AND "IsActive" = ?`, tenantID, req.TemplateId, true).
		First(&tmpl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, newBusinessError("模板不存在或已禁用")
		}
		return nil, err
	}

	// 加载模板项
	var tmplItems []models.ScheduleTemplateItem
	q := s.db.Where(`"TenantId" = ? AND "TemplateId" = ?`, tenantID, req.TemplateId)
	if len(req.ItemIds) > 0 {
		q = q.Where(`"Id" IN ?`, req.ItemIds)
	}
	// 若传入 WardId，过滤模板项：
	// 显式 WardId 匹配的始终允许；WardId IS NULL 的仅在 Scope=ALL 时允许
	if req.WardId != nil && *req.WardId > 0 {
		if tmpl.Scope == "ALL" {
			q = q.Where(`("WardId" = ? OR "WardId" IS NULL)`, *req.WardId)
		} else {
			q = q.Where(`"WardId" = ?`, *req.WardId)
		}
	}
	if err := q.Find(&tmplItems).Error; err != nil {
		return nil, err
	}
	if len(tmplItems) == 0 {
		return nil, newBusinessError("没有可用的模板项")
	}

	type createResult struct {
		shiftId    int64
		shiftExtId int64
	}
	var createdShifts []int64
	var createdShiftExts []int64

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range tmplItems {
			shiftId := item.ShiftId
			if shiftId == nil {
				return newBusinessError("模板项缺少班次配置")
			}

			// 确定目标 Bed
			bedID := item.FixedHdBedId
			if item.HdfEnabled && item.FixedHdfBedId != nil {
				bedID = item.FixedHdfBedId
			}

			// 校验患者同日同班冲突
			hasConflict, err := checkConflictTx(tx, item.PatientId, tenantID, targetDate, *shiftId, nil)
			if err != nil {
				return err
			}
			if hasConflict {
				return newBusinessError("患者同日同班已有排班")
			}

			// 校验同床同日同班冲突
			if bedID != nil && *bedID > 0 {
				hasBedConflict, err := checkBedConflictTx(tx, *bedID, tenantID, targetDate, *shiftId, nil)
				if err != nil {
					return err
				}
				if hasBedConflict {
					return newBusinessError("同床同日同班已有排班")
				}
			}

			// 确定 DialysisMode
			mode := "HD"
			if item.HdfEnabled {
				mode = "HDF"
			}

			// 校验机器支持模式
			if bedID != nil && *bedID > 0 {
				var bedExt models.BedMachineExt
				err := tx.Where(`"TenantId" = ? AND "BedId" = ?`, tenantID, *bedID).First(&bedExt).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return newBusinessError("目标床位缺少机器扩展配置")
				}
				if err != nil {
					return err
				}
				if !modeSupports(bedExt.SupportedModes, mode) {
					return newBusinessError("目标床位不支持" + mode + "模式")
				}
			}

			// 写入 Schedule_PatientShift (Status=10 草稿)
			shiftWardId := item.WardId
			if shiftWardId == nil && req.WardId != nil {
				shiftWardId = req.WardId
			}
			ps := models.PatientShift{
				TenantId:      tenantID,
				PatientId:     modeltypes.LegacyID(item.PatientId),
				ScheduleDate:  targetDate,
				ShiftId:       *shiftId,
				BedId:         bedID,
				WardId:        shiftWardId,
				PatientPlanId: &[]int64{0}[0],
				ShiftTiming:   &[]int{20}[0],
				Status:        10,
				CreatorId:     creatorID,
			}
			if err := tx.Create(&ps).Error; err != nil {
				if isPatientShiftUniqueViolation(err) {
					return ErrPatientShiftDuplicate
				}
				return err
			}

			// 写入 Schedule_PatientShiftExt
			ext := models.PatientShiftExt{
				TenantId:             tenantID,
				PatientShiftId:       ps.Id,
				DialysisMode:         mode,
				SourceType:           10,
				RecordForm:           10,
				RuleStatus:           10,
				SourceTemplateItemId: &item.Id,
				CreatorId:            creatorID,
			}
			tmplVer := item.TemplateVersion
			ext.SourceTemplateVersion = &tmplVer
			if err := tx.Create(&ext).Error; err != nil {
				return err
			}

			createdShifts = append(createdShifts, ps.Id)
			createdShiftExts = append(createdShiftExts, ext.Id)
		}
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	return &ScheduleTemplateApplyResponse{
		CreatedShifts:    createdShifts,
		CreatedShiftExts: createdShiftExts,
		Count:            len(createdShifts),
	}, nil
}

func checkConflictTx(db *gorm.DB, patientID, tenantID int64, date time.Time, shiftID int64, excludeID *int64) (bool, error) {
	var count int64
	q := db.Model(&models.PatientShift{}).
		Where(`"TenantId" = ? AND "PatientId" = ? AND DATE("TreatmentTime") = DATE(?) AND "ShiftId" = ?`,
			tenantID, patientID, date, shiftID).
		Where(`"Status" NOT IN ?`, inactivePatientShiftLegacyStatuses)
	if excludeID != nil {
		q = q.Where(`"Id" <> ?`, *excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func checkBedConflictTx(db *gorm.DB, bedID, tenantID int64, date time.Time, shiftID int64, excludeID *int64) (bool, error) {
	var count int64
	q := db.Model(&models.PatientShift{}).
		Where(`"TenantId" = ? AND "BedId" = ? AND DATE("TreatmentTime") = DATE(?) AND "ShiftId" = ?`,
			tenantID, bedID, date, shiftID).
		Where(`"Status" NOT IN ?`, inactivePatientShiftLegacyStatuses)
	if excludeID != nil {
		q = q.Where(`"Id" <> ?`, *excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func isPatientShiftUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
