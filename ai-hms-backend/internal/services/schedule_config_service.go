package services

import (
	"errors"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

// ScheduleConfigService 排班配置服务（WardExt / BedMachineExt / PatientProfile / TenantSetting / Calendar）
type ScheduleConfigService struct {
	db *gorm.DB
}

func NewScheduleConfigService() *ScheduleConfigService {
	return &ScheduleConfigService{db: database.GetDB()}
}

func (s *ScheduleConfigService) dbCheck() error {
	if s.db == nil {
		return errors.New("database not available")
	}
	return nil
}

// ===================== 老表存在性校验（不加外键、不 ALTER 老表） =====================

func (s *ScheduleConfigService) wardExists(tenantID, wardID int64) error {
	var count int64
	err := s.db.Model(&models.Ward{}).Where(`"TenantId" = ? AND "Id" = ?`, tenantID, wardID).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("病区不存在")
	}
	return nil
}

func (s *ScheduleConfigService) bedExists(tenantID, bedID int64) error {
	var count int64
	err := s.db.Model(&models.Bed{}).Where(`"TenantId" = ? AND "Id" = ?`, tenantID, bedID).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("床位不存在")
	}
	return nil
}

func (s *ScheduleConfigService) patientExists(tenantID, patientID int64) error {
	var count int64
	err := s.db.Table("Register_PatientInfomation").Where(`"TenantId" = ? AND "Id" = ?`, tenantID, patientID).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("患者不存在")
	}
	return nil
}

func (s *ScheduleConfigService) shiftExists(tenantID, shiftID int64) error {
	var count int64
	err := s.db.Model(&models.Shift{}).Where(`"TenantId" = ? AND "Id" = ?`, tenantID, shiftID).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("班次不存在")
	}
	return nil
}

// ===================== WardExt 病区扩展 =====================

type WardExtRequest struct {
	WardId       int64  `json:"wardId" binding:"required"`
	ZoneType     string `json:"zoneType" binding:"required"`
	ParentWardId *int64 `json:"parentWardId"`
	IsSubZone    *bool  `json:"isSubZone"`
	Note         string `json:"note"`
}

func (s *ScheduleConfigService) UpsertWardExt(tenantID, creatorID int64, req WardExtRequest) (*models.WardExt, error) {
	if err := s.dbCheck(); err != nil {
		return nil, err
	}

	zone := strings.ToUpper(strings.TrimSpace(req.ZoneType))
	if zone != "A" && zone != "B" && zone != "C" {
		return nil, errors.New("ZoneType 必须为 A、B 或 C")
	}

	if err := s.wardExists(tenantID, req.WardId); err != nil {
		return nil, err
	}
	if req.ParentWardId != nil && *req.ParentWardId != 0 {
		if err := s.wardExists(tenantID, *req.ParentWardId); err != nil {
			return nil, err
		}
	}

	var ext models.WardExt
	err := s.db.Where(`"TenantId" = ? AND "WardId" = ?`, tenantID, req.WardId).First(&ext).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ext = models.WardExt{
			TenantId:  tenantID,
			WardId:    req.WardId,
			ZoneType:  zone,
			Note:      req.Note,
			CreatorId: creatorID,
		}
		if req.ParentWardId != nil {
			ext.ParentWardId = req.ParentWardId
		}
		if req.IsSubZone != nil {
			ext.IsSubZone = *req.IsSubZone
		}
		if err := s.db.Create(&ext).Error; err != nil {
			return nil, err
		}
		return &ext, nil
	}
	if err != nil {
		return nil, err
	}

	ext.ZoneType = zone
	if req.ParentWardId != nil {
		ext.ParentWardId = req.ParentWardId
	}
	if req.IsSubZone != nil {
		ext.IsSubZone = *req.IsSubZone
	}
	if req.Note != "" {
		ext.Note = req.Note
	}
	if err := s.db.Save(&ext).Error; err != nil {
		return nil, err
	}
	return &ext, nil
}

func (s *ScheduleConfigService) ListWardExts(tenantID int64) ([]models.WardExt, error) {
	if err := s.dbCheck(); err != nil {
		return nil, err
	}
	var items []models.WardExt
	err := s.db.Where(`"TenantId" = ?`, tenantID).Order(`"Id"`).Find(&items).Error
	return items, err
}

// ===================== BedMachineExt 机器扩展 =====================

type BedMachineExtRequest struct {
	BedId          int64  `json:"bedId" binding:"required"`
	MachineCode    string `json:"machineCode"`
	MachineType    string `json:"machineType" binding:"required"`
	SupportedModes string `json:"supportedModes"`
	PositionIndex  *int   `json:"positionIndex"`
	IsDisabled     *bool  `json:"isDisabled"`
	Note           string `json:"note"`
}

func normalizeModes(raw string) []string {
	parts := strings.Split(raw, ",")
	seen := map[string]bool{}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.ToUpper(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

func validateMachineTypeAndModes(machineType, supportedModes string) (string, error) {
	mt := strings.ToUpper(strings.TrimSpace(machineType))
	if mt != "HD" && mt != "HDF" && mt != "CRRT" {
		return "", errors.New("MachineType 必须为 HD、HDF 或 CRRT")
	}

	inputModes := supportedModes
	if strings.TrimSpace(inputModes) == "" {
		switch mt {
		case "HD":
			inputModes = "HD"
		case "HDF":
			inputModes = "HD,HDF,HF"
		case "CRRT":
			inputModes = "CRRT"
		}
	}

	normalized := normalizeModes(inputModes)
	if len(normalized) == 0 {
		return "", errors.New("SupportedModes 不能为空")
	}

	modeSet := map[string]bool{}
	for _, m := range normalized {
		if m != "HD" && m != "HDF" && m != "HF" && m != "CRRT" {
			return "", errors.New("SupportedModes 只能包含 HD、HDF、HF、CRRT")
		}
		modeSet[m] = true
	}

	hasHD := modeSet["HD"]
	hasHDF := modeSet["HDF"]
	hasHF := modeSet["HF"]
	hasCRRT := modeSet["CRRT"]

	switch mt {
	case "HD":
		if hasHDF || hasHF || hasCRRT {
			return "", errors.New("HD 机只支持 HD 模式")
		}
		if !hasHD {
			return "", errors.New("HD 机必须包含 HD 模式")
		}
	case "HDF":
		if hasCRRT {
			return "", errors.New("HDF 机不支持 CRRT 模式")
		}
		if !hasHD || !hasHDF {
			return "", errors.New("HDF 机必须包含 HD 和 HDF 模式")
		}
	case "CRRT":
		if hasHD || hasHDF || hasHF {
			return "", errors.New("CRRT 机只支持 CRRT 模式")
		}
		if !hasCRRT {
			return "", errors.New("CRRT 机必须包含 CRRT 模式")
		}
	}

	return strings.Join(normalized, ","), nil
}

func (s *ScheduleConfigService) UpsertBedMachineExt(tenantID, creatorID int64, req BedMachineExtRequest) (*models.BedMachineExt, error) {
	if err := s.dbCheck(); err != nil {
		return nil, err
	}

	modes, err := validateMachineTypeAndModes(req.MachineType, req.SupportedModes)
	if err != nil {
		return nil, err
	}

	if err := s.bedExists(tenantID, req.BedId); err != nil {
		return nil, err
	}

	var ext models.BedMachineExt
	derr := s.db.Where(`"TenantId" = ? AND "BedId" = ?`, tenantID, req.BedId).First(&ext).Error
	if errors.Is(derr, gorm.ErrRecordNotFound) {
		ext = models.BedMachineExt{
			TenantId:       tenantID,
			BedId:          req.BedId,
			MachineCode:    req.MachineCode,
			MachineType:    strings.ToUpper(strings.TrimSpace(req.MachineType)),
			SupportedModes: modes,
			CreatorId:      creatorID,
		}
		if req.PositionIndex != nil {
			ext.PositionIndex = *req.PositionIndex
		}
		if req.IsDisabled != nil {
			ext.IsDisabled = *req.IsDisabled
		}
		ext.Note = req.Note
		if err := s.db.Create(&ext).Error; err != nil {
			return nil, err
		}
		return &ext, nil
	}
	if derr != nil {
		return nil, derr
	}

	ext.MachineType = strings.ToUpper(strings.TrimSpace(req.MachineType))
	ext.SupportedModes = modes
	if req.MachineCode != "" {
		ext.MachineCode = req.MachineCode
	}
	if req.PositionIndex != nil {
		ext.PositionIndex = *req.PositionIndex
	}
	if req.IsDisabled != nil {
		ext.IsDisabled = *req.IsDisabled
	}
	if req.Note != "" {
		ext.Note = req.Note
	}
	if err := s.db.Save(&ext).Error; err != nil {
		return nil, err
	}
	return &ext, nil
}

func (s *ScheduleConfigService) ListBedMachineExts(tenantID int64) ([]models.BedMachineExt, error) {
	if err := s.dbCheck(); err != nil {
		return nil, err
	}
	var items []models.BedMachineExt
	err := s.db.Where(`"TenantId" = ?`, tenantID).Order(`"Id"`).Find(&items).Error
	return items, err
}

// ===================== PatientProfile 患者骨架 =====================

type PatientProfileRequest struct {
	PatientId           int64   `json:"patientId" binding:"required"`
	ZoneTag             string  `json:"zoneTag" binding:"required"`
	HomeWardId          *int64  `json:"homeWardId"`
	FreqPattern         *int16  `json:"freqPattern"`
	ShiftId             *int64  `json:"shiftId"`
	DefaultMode         *string `json:"defaultMode"`
	HdfEnabled          *bool   `json:"hdfEnabled"`
	HdfWeekday          *int16  `json:"hdfWeekday"`
	HdfWeekParity       *int16  `json:"hdfWeekParity"`
	FixedHdBedId        *int64  `json:"fixedHdBedId"`
	FixedHdfBedId       *int64  `json:"fixedHdfBedId"`
	IsAdmissionRejected *bool   `json:"isAdmissionRejected"`
	EffectiveFrom       *string `json:"effectiveFrom"`
}

func validateFreqPattern(freq int16) error {
	switch freq {
	case 10, 20, 30, 40, 90:
		return nil
	}
	return errors.New("FreqPattern 必须为 10/20/30/40/90")
}

func (s *ScheduleConfigService) UpsertPatientProfile(tenantID, creatorID int64, req PatientProfileRequest) (*models.PatientProfile, error) {
	if err := s.dbCheck(); err != nil {
		return nil, err
	}

	zone := strings.ToUpper(strings.TrimSpace(req.ZoneTag))
	if zone != "A" && zone != "B" && zone != "C" {
		return nil, errors.New("ZoneTag 必须为 A、B 或 C")
	}

	freq := int16(10)
	if req.FreqPattern != nil {
		freq = *req.FreqPattern
	}
	if err := validateFreqPattern(freq); err != nil {
		return nil, err
	}

	if err := s.patientExists(tenantID, req.PatientId); err != nil {
		return nil, err
	}
	if req.HomeWardId != nil && *req.HomeWardId != 0 {
		if err := s.wardExists(tenantID, *req.HomeWardId); err != nil {
			return nil, err
		}
	}
	if req.ShiftId != nil && *req.ShiftId != 0 {
		if err := s.shiftExists(tenantID, *req.ShiftId); err != nil {
			return nil, err
		}
	}
	if req.FixedHdBedId != nil && *req.FixedHdBedId != 0 {
		if err := s.bedExists(tenantID, *req.FixedHdBedId); err != nil {
			return nil, err
		}
	}
	if req.FixedHdfBedId != nil && *req.FixedHdfBedId != 0 {
		if err := s.bedExists(tenantID, *req.FixedHdfBedId); err != nil {
			return nil, err
		}
	}

	var profile models.PatientProfile
	derr := s.db.Where(`"TenantId" = ? AND "PatientId" = ?`, tenantID, req.PatientId).First(&profile).Error
	if errors.Is(derr, gorm.ErrRecordNotFound) {
		profile = models.PatientProfile{
			TenantId:    tenantID,
			PatientId:   req.PatientId,
			ZoneTag:     zone,
			FreqPattern: freq,
			CreatorId:   creatorID,
		}
		if err := s.patchProfileFields(&profile, req); err != nil {
			return nil, err
		}
		if err := s.db.Create(&profile).Error; err != nil {
			return nil, err
		}
		return &profile, nil
	}
	if derr != nil {
		return nil, derr
	}

	profile.ZoneTag = zone
	if req.FreqPattern != nil {
		profile.FreqPattern = *req.FreqPattern
	}
	if err := s.patchProfileFields(&profile, req); err != nil {
		return nil, err
	}
	if err := s.db.Save(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *ScheduleConfigService) patchProfileFields(p *models.PatientProfile, req PatientProfileRequest) error {
	if req.HomeWardId != nil {
		p.HomeWardId = req.HomeWardId
	}
	if req.ShiftId != nil {
		p.ShiftId = req.ShiftId
	}
	if req.DefaultMode != nil {
		mode := strings.ToUpper(strings.TrimSpace(*req.DefaultMode))
		if mode != "HD" && mode != "HDF" && mode != "HF" && mode != "CRRT" {
			return errors.New("DefaultMode 必须为 HD、HDF、HF 或 CRRT")
		}
		p.DefaultMode = mode
	}
	if req.HdfEnabled != nil {
		p.HdfEnabled = *req.HdfEnabled
	}
	if req.HdfWeekday != nil {
		if *req.HdfWeekday < 1 || *req.HdfWeekday > 6 {
			return errors.New("HdfWeekday 必须为 1-6（周一到周六）")
		}
		p.HdfWeekday = req.HdfWeekday
	}
	if req.HdfWeekParity != nil {
		p.HdfWeekParity = req.HdfWeekParity
	}
	if req.FixedHdBedId != nil {
		p.FixedHdBedId = req.FixedHdBedId
	}
	if req.FixedHdfBedId != nil {
		p.FixedHdfBedId = req.FixedHdfBedId
	}
	if req.IsAdmissionRejected != nil {
		p.IsAdmissionRejected = *req.IsAdmissionRejected
	}
	if req.EffectiveFrom != nil {
		t, err := time.Parse("2006-01-02", *req.EffectiveFrom)
		if err != nil {
			return errors.New("EffectiveFrom 格式应为 YYYY-MM-DD")
		}
		p.EffectiveFrom = &t
	}
	return nil
}

func (s *ScheduleConfigService) ListPatientProfiles(tenantID int64) ([]models.PatientProfile, error) {
	if err := s.dbCheck(); err != nil {
		return nil, err
	}
	var items []models.PatientProfile
	err := s.db.Where(`"TenantId" = ?`, tenantID).Order(`"Id"`).Find(&items).Error
	return items, err
}

// ===================== TenantSetting 排班配置 =====================

type TenantSettingRequest struct {
	SettingKey   string `json:"settingKey" binding:"required"`
	SettingValue string `json:"settingValue" binding:"required"`
	SettingType  string `json:"settingType" binding:"required"`
}

func validateSetting(key, value, typ string) error {
	switch key {
	case "OddEvenWeekAnchorMonday":
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return errors.New("OddEvenWeekAnchorMonday 必须是 YYYY-MM-DD 格式的周一")
		}
		// 校验是否为周一
		t, _ := time.Parse("2006-01-02", value)
		if t.Weekday() != time.Monday {
			return errors.New("OddEvenWeekAnchorMonday 必须是周一")
		}
	case "DraftWeeks":
		if value != "2" && value != "4" {
			return errors.New("DraftWeeks 必须为 2 或 4")
		}
	case "SpillHorizonDays":
		n, err := parseInt(value)
		if err != nil || n <= 0 {
			return errors.New("SpillHorizonDays 必须为正整数")
		}
	case "LowSlotWarnThreshold":
		n, err := parseInt(value)
		if err != nil || n < 0 {
			return errors.New("LowSlotWarnThreshold 必须为非负整数")
		}
	case "LastGenerationVersion":
		n, err := parseInt(value)
		if err != nil || n < 0 {
			return errors.New("LastGenerationVersion 必须为非负整数")
		}
	}
	if typ != "date" && typ != "int" && typ != "string" && typ != "bool" {
		return errors.New("SettingType 必须为 date、int、string 或 bool")
	}
	return nil
}

func parseInt(v string) (int, error) {
	n := 0
	for _, c := range v {
		if c < '0' || c > '9' {
			return 0, errors.New("不是有效整数")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

func (s *ScheduleConfigService) UpsertTenantSetting(tenantID, creatorID int64, req TenantSettingRequest) (*models.TenantSetting, error) {
	if err := s.dbCheck(); err != nil {
		return nil, err
	}

	key := strings.TrimSpace(req.SettingKey)
	if key == "" {
		return nil, errors.New("SettingKey 不能为空")
	}
	if err := validateSetting(key, req.SettingValue, req.SettingType); err != nil {
		return nil, err
	}

	var st models.TenantSetting
	err := s.db.Where(`"TenantId" = ? AND "SettingKey" = ?`, tenantID, key).First(&st).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		st = models.TenantSetting{
			TenantId:     tenantID,
			SettingKey:   key,
			SettingValue: req.SettingValue,
			SettingType:  req.SettingType,
			CreatorId:    creatorID,
		}
		if err := s.db.Create(&st).Error; err != nil {
			return nil, err
		}
		return &st, nil
	}
	if err != nil {
		return nil, err
	}

	st.SettingValue = req.SettingValue
	st.SettingType = req.SettingType
	if err := s.db.Save(&st).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *ScheduleConfigService) ListTenantSettings(tenantID int64) ([]models.TenantSetting, error) {
	if err := s.dbCheck(); err != nil {
		return nil, err
	}
	var items []models.TenantSetting
	err := s.db.Where(`"TenantId" = ?`, tenantID).Order(`"SettingKey"`).Find(&items).Error
	return items, err
}

// ===================== Calendar 日历 =====================

type CalendarRequest struct {
	CalDate       string  `json:"calDate" binding:"required"`
	IsDialysisDay *bool   `json:"isDialysisDay" binding:"required"`
	HolidayMode   *int16  `json:"holidayMode"`
	Note          string  `json:"note"`
	OpenWardIds   []int64 `json:"openWardIds"`
	OpenBedIds    []int64 `json:"openBedIds"`
}

func (s *ScheduleConfigService) UpsertCalendar(tenantID, creatorID int64, req CalendarRequest) (*models.Calendar, []models.CalendarOpenWard, []models.CalendarOpenBed, error) {
	if err := s.dbCheck(); err != nil {
		return nil, nil, nil, err
	}

	calDate, err := time.Parse("2006-01-02", req.CalDate)
	if err != nil {
		return nil, nil, nil, errors.New("calDate 格式应为 YYYY-MM-DD")
	}

	for _, wid := range req.OpenWardIds {
		if wid == 0 {
			continue
		}
		if err := s.wardExists(tenantID, wid); err != nil {
			return nil, nil, nil, err
		}
	}
	for _, bid := range req.OpenBedIds {
		if bid == 0 {
			continue
		}
		if err := s.bedExists(tenantID, bid); err != nil {
			return nil, nil, nil, err
		}
	}

	var cal models.Calendar
	var openWards []models.CalendarOpenWard
	var openBeds []models.CalendarOpenBed

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		derr := tx.Where(`"TenantId" = ? AND "CalDate" = ?`, tenantID, calDate).First(&cal).Error
		if errors.Is(derr, gorm.ErrRecordNotFound) {
			cal = models.Calendar{
				TenantId:      tenantID,
				CalDate:       calDate,
				IsDialysisDay: *req.IsDialysisDay,
				CreatorId:     creatorID,
			}
			if req.HolidayMode != nil {
				cal.HolidayMode = *req.HolidayMode
			}
			cal.Note = req.Note
			if err := tx.Create(&cal).Error; err != nil {
				return err
			}
		} else if derr != nil {
			return derr
		} else {
			cal.IsDialysisDay = *req.IsDialysisDay
			if req.HolidayMode != nil {
				cal.HolidayMode = *req.HolidayMode
			}
			if req.Note != "" {
				cal.Note = req.Note
			}
			if err := tx.Save(&cal).Error; err != nil {
				return err
			}
		}

		if err := tx.Where(`"TenantId" = ? AND "CalendarId" = ?`, tenantID, cal.Id).Delete(&models.CalendarOpenWard{}).Error; err != nil {
			return err
		}
		if err := tx.Where(`"TenantId" = ? AND "CalendarId" = ?`, tenantID, cal.Id).Delete(&models.CalendarOpenBed{}).Error; err != nil {
			return err
		}

		for _, wid := range req.OpenWardIds {
			if wid == 0 {
				continue
			}
			cw := models.CalendarOpenWard{TenantId: tenantID, CalendarId: cal.Id, WardId: wid}
			if err := tx.Create(&cw).Error; err != nil {
				return err
			}
			openWards = append(openWards, cw)
		}
		for _, bid := range req.OpenBedIds {
			if bid == 0 {
				continue
			}
			cb := models.CalendarOpenBed{TenantId: tenantID, CalendarId: cal.Id, BedId: bid}
			if err := tx.Create(&cb).Error; err != nil {
				return err
			}
			openBeds = append(openBeds, cb)
		}

		return nil
	})

	if txErr != nil {
		return nil, nil, nil, txErr
	}

	return &cal, openWards, openBeds, nil
}

func (s *ScheduleConfigService) ListCalendars(tenantID int64) ([]models.Calendar, error) {
	if err := s.dbCheck(); err != nil {
		return nil, err
	}
	var items []models.Calendar
	err := s.db.Where(`"TenantId" = ?`, tenantID).Order(`"CalDate"`).Find(&items).Error
	return items, err
}
