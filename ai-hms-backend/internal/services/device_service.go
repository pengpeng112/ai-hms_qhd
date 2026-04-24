package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

type legacyEquipment struct {
	ID               int64      `gorm:"column:Id;primaryKey"`
	TenantId         int64      `gorm:"column:TenantId"`
	Name             string     `gorm:"column:Name"`
	IDNo             string     `gorm:"column:IDNo"`
	SerialNo         string     `gorm:"column:SerialNo"`
	Brand            string     `gorm:"column:Brand"`
	ModelNo          string     `gorm:"column:ModelNo"`
	DialysisMethod   string     `gorm:"column:DialysisMethod"`
	Type             string     `gorm:"column:Type"`
	ManufactureDate  *time.Time `gorm:"column:ManufactureDate"`
	Manufacturer     string     `gorm:"column:Manufacturer"`
	InstallDate      *time.Time `gorm:"column:InstallDate"`
	Maintenance      int64      `gorm:"column:Maintenance"`
	MaintenanceCycle string     `gorm:"column:MaintenanceCycle"`
	Note             string     `gorm:"column:Note"`
	IsDisabled       bool       `gorm:"column:IsDisabled"`
	CreatorId        int64      `gorm:"column:CreatorId"`
	CreateTime       time.Time  `gorm:"column:CreateTime"`
	LastModifyTime   time.Time  `gorm:"column:LastModifyTime"`
	Flux             string     `gorm:"column:Flux"`
}

func (legacyEquipment) TableName() string { return "Auxiliary_EquipmentInfomation" }

type legacyBedEquipmentRel struct {
	ID             int64      `gorm:"column:Id;primaryKey"`
	TenantId       int64      `gorm:"column:TenantId"`
	EquipmentId    int64      `gorm:"column:EquipmentId"`
	Sort           int        `gorm:"column:Sort"`
	BedId          int64      `gorm:"column:BedId"`
	IsDefault      bool       `gorm:"column:IsDefault"`
	IsDisabled     bool       `gorm:"column:IsDisabled"`
	LastModifyTime *time.Time `gorm:"column:LastModifyTime"`
	Type           int        `gorm:"column:Type"`
	ParameterS     string     `gorm:"column:ParameterS"`
}

func (legacyBedEquipmentRel) TableName() string { return "Schedule_BedEquipmentRel" }

type legacyBedEquipmentRelChange struct {
	ID             int64      `gorm:"column:Id;primaryKey"`
	TenantId       int64      `gorm:"column:TenantId"`
	EquipmentId    int64      `gorm:"column:EquipmentId"`
	Sort           int        `gorm:"column:Sort"`
	BedId          int64      `gorm:"column:BedId"`
	IsDefault      bool       `gorm:"column:IsDefault"`
	IsDisabled     bool       `gorm:"column:IsDisabled"`
	Type           int        `gorm:"column:Type"`
	ParameterS     string     `gorm:"column:ParameterS"`
	CreatorId      int64      `gorm:"column:CreatorId"`
	CreateTime     time.Time  `gorm:"column:CreateTime"`
	LastModifyTime *time.Time `gorm:"column:LastModifyTime"`
}

func (legacyBedEquipmentRelChange) TableName() string { return "Schedule_BedEquipmentRelChange" }

type legacyBed struct {
	ID         int64  `gorm:"column:Id;primaryKey"`
	TenantId   int64  `gorm:"column:TenantId"`
	Name       string `gorm:"column:Name"`
	WardId     int64  `gorm:"column:WardId"`
	Sort       int    `gorm:"column:Sort"`
	IsDisabled bool   `gorm:"column:IsDisabled"`
}

func (legacyBed) TableName() string { return "Schedule_Bed" }

type legacyDeviceRecord struct {
	ID               int64      `gorm:"column:id"`
	TenantId         int64      `gorm:"column:tenant_id"`
	Name             string     `gorm:"column:name"`
	IDNo             string     `gorm:"column:id_no"`
	SerialNo         string     `gorm:"column:serial_no"`
	Brand            string     `gorm:"column:brand"`
	Model            string     `gorm:"column:model"`
	DialysisMethod   string     `gorm:"column:dialysis_method"`
	DeviceType       string     `gorm:"column:device_type"`
	Manufacturer     string     `gorm:"column:manufacturer"`
	BedNumber        string     `gorm:"column:bed_number"`
	BedId            *int64     `gorm:"column:bed_id"`
	WardId           *int64     `gorm:"column:ward_id"`
	WardName         string     `gorm:"column:ward_name"`
	Status           string     `gorm:"column:status"`
	PurchaseDate     *time.Time `gorm:"column:purchase_date"`
	ManufactureDate  *time.Time `gorm:"column:manufacture_date"`
	InstallDate      *time.Time `gorm:"column:install_date"`
	LastMaintained   *time.Time `gorm:"column:last_maintained"`
	Maintenance      *int64     `gorm:"column:maintenance"`
	MaintenanceCycle string     `gorm:"column:maintenance_cycle"`
	Flux             string     `gorm:"column:flux"`
	Notes            string     `gorm:"column:notes"`
	IsDisabled       bool       `gorm:"column:is_disabled"`
	CreatorId        int64      `gorm:"column:creator_id"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

type legacyEquipmentUsageLog struct {
	ID             int64      `gorm:"column:Id;primaryKey"`
	TenantId       int64      `gorm:"column:TenantId"`
	EquipmentId    int64      `gorm:"column:EquipmentId"`
	UseUserId      int64      `gorm:"column:UseUserId"`
	UseStartTime   *time.Time `gorm:"column:UseStartTime"`
	UseDuration    float64    `gorm:"column:UseDuration"`
	Note           string     `gorm:"column:Note"`
	CreatorId      int64      `gorm:"column:CreatorId"`
	CreateTime     time.Time  `gorm:"column:CreateTime"`
	LastModifyTime time.Time  `gorm:"column:LastModifyTime"`
}

func (legacyEquipmentUsageLog) TableName() string { return "Auxiliary_EquipmentUsageLog" }

type legacyEquipmentMaintenance struct {
	ID             int64      `gorm:"column:Id;primaryKey"`
	TenantId       int64      `gorm:"column:TenantId"`
	EquipmentId    int64      `gorm:"column:EquipmentId"`
	Type           string     `gorm:"column:Type"`
	Mode           string     `gorm:"column:Mode"`
	OperatorId     int64      `gorm:"column:OperatorId"`
	OperateTime    *time.Time `gorm:"column:OperateTime"`
	Description    string     `gorm:"column:Description"`
	Note           string     `gorm:"column:Note"`
	CreatorId      int64      `gorm:"column:CreatorId"`
	CreateTime     time.Time  `gorm:"column:CreateTime"`
	LastModifyTime time.Time  `gorm:"column:LastModifyTime"`
}

func (legacyEquipmentMaintenance) TableName() string { return "Auxiliary_EquipmentMaintenance" }

type legacyEquipmentDisinfection struct {
	ID              int64      `gorm:"column:Id;primaryKey"`
	TenantId        int64      `gorm:"column:TenantId"`
	EquipmentId     int64      `gorm:"column:EquipmentId"`
	DisinfectUserId int64      `gorm:"column:DisinfectUserId"`
	DisinfectWay    string     `gorm:"column:DisinfectWay"`
	StartTime       *time.Time `gorm:"column:StartTime"`
	Description     string     `gorm:"column:Description"`
	Note            string     `gorm:"column:Note"`
	CreatorId       int64      `gorm:"column:CreatorId"`
	CreateTime      time.Time  `gorm:"column:CreateTime"`
	LastModifyTime  time.Time  `gorm:"column:LastModifyTime"`
	TreatmentId     int64      `gorm:"column:TreatmentId"`
	Status          int        `gorm:"column:Status"`
	EndTime         *time.Time `gorm:"column:EndTime"`
	Type            string     `gorm:"column:Type"`
	Disinfectant    string     `gorm:"column:Disinfectant"`
}

func (legacyEquipmentDisinfection) TableName() string { return "Auxiliary_EquipmentDisinfection" }

type DeviceService struct {
	db *gorm.DB
}

func NewDeviceService() *DeviceService {
	return &DeviceService{db: database.GetDB()}
}

type DeviceListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	Status    string `form:"status"`
	BedNumber string `form:"bedNumber"`
	WardId    *int64 `form:"wardId"`
	Keyword   string `form:"keyword"`
}

type DeviceListResponse struct {
	Items     []models.Device `json:"items"`
	Total     int64           `json:"total"`
	Page      int             `json:"page"`
	PageSize  int             `json:"pageSize"`
	TotalPage int             `json:"totalPage"`
}

type DeviceCreateRequest struct {
	Name             string     `json:"name" binding:"required"`
	IDNo             string     `json:"idNo"`
	SerialNo         string     `json:"serialNo"`
	Brand            string     `json:"brand"`
	Model            string     `json:"model"`
	ModelNo          string     `json:"modelNo"`
	DialysisMethod   string     `json:"dialysisMethod"`
	DeviceType       string     `json:"deviceType"`
	Type             string     `json:"type"`
	ManufactureDate  *time.Time `json:"manufactureDate"`
	Manufacturer     string     `json:"manufacturer"`
	InstallDate      *time.Time `json:"installDate"`
	Maintenance      *int64     `json:"maintenance"`
	MaintenanceCycle string     `json:"maintenanceCycle"`
	Flux             string     `json:"flux"`
	BedNumber        string     `json:"bedNumber"`
	WardId           *int64     `json:"wardId"`
	Notes            string     `json:"notes"`
	Note             string     `json:"note"`
}

type DeviceUpdateRequest struct {
	Name             *string    `json:"name"`
	IDNo             *string    `json:"idNo"`
	SerialNo         *string    `json:"serialNo"`
	Brand            *string    `json:"brand"`
	Model            *string    `json:"model"`
	ModelNo          *string    `json:"modelNo"`
	DialysisMethod   *string    `json:"dialysisMethod"`
	DeviceType       *string    `json:"deviceType"`
	Type             *string    `json:"type"`
	ManufactureDate  *time.Time `json:"manufactureDate"`
	Manufacturer     *string    `json:"manufacturer"`
	InstallDate      *time.Time `json:"installDate"`
	Maintenance      *int64     `json:"maintenance"`
	MaintenanceCycle *string    `json:"maintenanceCycle"`
	Flux             *string    `json:"flux"`
	BedNumber        *string    `json:"bedNumber"`
	WardId           *int64     `json:"wardId"`
	Status           *string    `json:"status"`
	Notes            *string    `json:"notes"`
	Note             *string    `json:"note"`
}

type DeviceLogListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}

type DeviceUsageLogDTO struct {
	ID             int64      `json:"id"`
	TenantId       int64      `json:"tenantId"`
	EquipmentId    int64      `json:"equipmentId"`
	UseUserId      int64      `json:"useUserId"`
	UseStartTime   *time.Time `json:"useStartTime"`
	UseDuration    float64    `json:"useDuration"`
	Note           string     `json:"note"`
	CreatorId      int64      `json:"creatorId"`
	CreateTime     time.Time  `json:"createTime"`
	LastModifyTime time.Time  `json:"lastModifyTime"`
}

type DeviceMaintenanceRecordDTO struct {
	ID             int64      `json:"id"`
	TenantId       int64      `json:"tenantId"`
	EquipmentId    int64      `json:"equipmentId"`
	Type           string     `json:"type"`
	Mode           string     `json:"mode"`
	OperatorId     int64      `json:"operatorId"`
	OperateTime    *time.Time `json:"operateTime"`
	Description    string     `json:"description"`
	Note           string     `json:"note"`
	CreatorId      int64      `json:"creatorId"`
	CreateTime     time.Time  `json:"createTime"`
	LastModifyTime time.Time  `json:"lastModifyTime"`
}

type DeviceDisinfectionRecordDTO struct {
	ID              int64      `json:"id"`
	TenantId        int64      `json:"tenantId"`
	EquipmentId     int64      `json:"equipmentId"`
	DisinfectUserId int64      `json:"disinfectUserId"`
	DisinfectWay    string     `json:"disinfectWay"`
	StartTime       *time.Time `json:"startTime"`
	EndTime         *time.Time `json:"endTime"`
	Description     string     `json:"description"`
	Note            string     `json:"note"`
	Type            string     `json:"type"`
	Disinfectant    string     `json:"disinfectant"`
	Status          int        `json:"status"`
	TreatmentId     int64      `json:"treatmentId"`
	CreatorId       int64      `json:"creatorId"`
	CreateTime      time.Time  `json:"createTime"`
	LastModifyTime  time.Time  `json:"lastModifyTime"`
}

type DeviceUsageLogListResponse struct {
	Items     []DeviceUsageLogDTO `json:"items"`
	Total     int64               `json:"total"`
	Page      int                 `json:"page"`
	PageSize  int                 `json:"pageSize"`
	TotalPage int                 `json:"totalPage"`
}

type DeviceMaintenanceRecordListResponse struct {
	Items     []DeviceMaintenanceRecordDTO `json:"items"`
	Total     int64                        `json:"total"`
	Page      int                          `json:"page"`
	PageSize  int                          `json:"pageSize"`
	TotalPage int                          `json:"totalPage"`
}

type DeviceDisinfectionRecordListResponse struct {
	Items     []DeviceDisinfectionRecordDTO `json:"items"`
	Total     int64                         `json:"total"`
	Page      int                           `json:"page"`
	PageSize  int                           `json:"pageSize"`
	TotalPage int                           `json:"totalPage"`
}

func normalizeDeviceStatus(status string) string {
	normalized := strings.TrimSpace(strings.ToLower(status))
	switch normalized {
	case "", "error":
		if normalized == "error" {
			return models.DeviceStatusAlarm
		}
		return models.DeviceStatusNormal
	case models.DeviceStatusNormal, models.DeviceStatusWarning, models.DeviceStatusAlarm, models.DeviceStatusOffline, models.DeviceStatusMaintenance:
		return normalized
	default:
		return normalized
	}
}

func normalizeDevicePage(page, pageSize, defaultPageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	return page, pageSize
}

func resolveDeviceModel(model, modelNo string) string {
	if strings.TrimSpace(modelNo) != "" {
		return strings.TrimSpace(modelNo)
	}
	return strings.TrimSpace(model)
}

func resolveDeviceType(deviceType, fallback string) string {
	if strings.TrimSpace(deviceType) != "" {
		return strings.TrimSpace(deviceType)
	}
	return strings.TrimSpace(fallback)
}

func resolveDeviceNotes(notes, note string) string {
	if strings.TrimSpace(note) != "" {
		return strings.TrimSpace(note)
	}
	return strings.TrimSpace(notes)
}

func toDeviceDTO(row legacyDeviceRecord) models.Device {
	return models.Device{
		ID:               strconv.FormatInt(row.ID, 10),
		TenantId:         row.TenantId,
		Name:             row.Name,
		IDNo:             row.IDNo,
		SerialNo:         row.SerialNo,
		Brand:            row.Brand,
		Model:            row.Model,
		DialysisMethod:   row.DialysisMethod,
		DeviceType:       row.DeviceType,
		Manufacturer:     row.Manufacturer,
		BedNumber:        row.BedNumber,
		BedId:            row.BedId,
		WardId:           row.WardId,
		WardName:         row.WardName,
		Status:           normalizeDeviceStatus(row.Status),
		PurchaseDate:     row.InstallDate,
		ManufactureDate:  row.ManufactureDate,
		InstallDate:      row.InstallDate,
		LastMaintained:   row.LastMaintained,
		Maintenance:      row.Maintenance,
		MaintenanceCycle: row.MaintenanceCycle,
		Flux:             row.Flux,
		Notes:            row.Notes,
		IsDisabled:       row.IsDisabled,
		CreatorId:        row.CreatorId,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

func (s *DeviceService) baseLegacyDeviceQuery() *gorm.DB {
	lastMaintSubquery := s.db.Table(`"Auxiliary_EquipmentMaintenance"`).
		Select(`"EquipmentId", MAX("OperateTime") AS last_maintained`).
		Where(`"TenantId" = ?`, legacyTenantID).
		Group(`"EquipmentId"`)

	return s.db.Table(`"Auxiliary_EquipmentInfomation" AS e`).
		Joins(`LEFT JOIN "Schedule_BedEquipmentRel" AS rel ON rel."EquipmentId" = e."Id" AND rel."TenantId" = e."TenantId" AND COALESCE(rel."IsDisabled", false) = false AND rel."IsDefault" = true`).
		Joins(`LEFT JOIN "Schedule_Bed" AS bed ON bed."Id" = rel."BedId" AND bed."TenantId" = e."TenantId"`).
		Joins(`LEFT JOIN "Schedule_Ward" AS ward ON ward."Id" = bed."WardId" AND ward."TenantId" = e."TenantId"`).
		Joins(`LEFT JOIN (?) AS maint ON maint."EquipmentId" = e."Id"`, lastMaintSubquery).
		Where(`e."TenantId" = ? AND COALESCE(e."IsDisabled", false) = false`, legacyTenantID)
}

func (s *DeviceService) legacyDeviceSelect() string {
	return `e."Id" AS id,
e."TenantId" AS tenant_id,
e."Name" AS name,
COALESCE(e."IDNo", '') AS id_no,
COALESCE(e."SerialNo", '') AS serial_no,
COALESCE(e."Brand", '') AS brand,
COALESCE(e."ModelNo", '') AS model,
COALESCE(e."DialysisMethod", '') AS dialysis_method,
COALESCE(e."Type", '') AS device_type,
COALESCE(e."Manufacturer", '') AS manufacturer,
COALESCE(bed."Name", '') AS bed_number,
bed."Id" AS bed_id,
ward."Id" AS ward_id,
COALESCE(ward."Name", '') AS ward_name,
COALESCE(NULLIF(rel."ParameterS", ''), ?) AS status,
e."InstallDate" AS purchase_date,
e."ManufactureDate" AS manufacture_date,
e."InstallDate" AS install_date,
maint.last_maintained AS last_maintained,
e."Maintenance" AS maintenance,
COALESCE(e."MaintenanceCycle", '') AS maintenance_cycle,
COALESCE(e."Flux", '') AS flux,
COALESCE(e."Note", '') AS notes,
COALESCE(e."IsDisabled", false) AS is_disabled,
COALESCE(e."CreatorId", 0) AS creator_id,
COALESCE(e."CreateTime", NOW()) AS created_at,
COALESCE(e."LastModifyTime", e."CreateTime", NOW()) AS updated_at`
}

func (s *DeviceService) List(req DeviceListRequest) (*DeviceListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	req.Page, req.PageSize = normalizeDevicePage(req.Page, req.PageSize, 50)

	query := s.baseLegacyDeviceQuery()
	if status := normalizeDeviceStatus(req.Status); strings.TrimSpace(req.Status) != "" {
		query = query.Where(`COALESCE(NULLIF(rel."ParameterS", ''), ?) = ?`, models.DeviceStatusNormal, status)
	}
	if req.BedNumber != "" {
		query = query.Where(`bed."Name" LIKE ?`, "%"+strings.TrimSpace(req.BedNumber)+"%")
	}
	if req.WardId != nil {
		query = query.Where(`ward."Id" = ?`, *req.WardId)
	}
	if req.Keyword != "" {
		like := "%" + strings.TrimSpace(req.Keyword) + "%"
		query = query.Where(`(e."Name" LIKE ? OR e."IDNo" LIKE ? OR e."SerialNo" LIKE ? OR e."Brand" LIKE ? OR e."ModelNo" LIKE ?)`, like, like, like, like, like)
	}

	var total int64
	if err := query.Distinct(`e."Id"`).Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []legacyDeviceRecord
	offset := (req.Page - 1) * req.PageSize
	if err := query.Select(s.legacyDeviceSelect(), models.DeviceStatusNormal).
		Offset(offset).
		Limit(req.PageSize).
		Order(`ward."Id" ASC NULLS LAST`).
		Order(`bed."Sort" ASC NULLS LAST`).
		Order(`e."Id" ASC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]models.Device, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDeviceDTO(row))
	}
	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}
	return &DeviceListResponse{Items: items, Total: total, Page: req.Page, PageSize: req.PageSize, TotalPage: totalPage}, nil
}

func (s *DeviceService) Get(id string) (*models.Device, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	deviceID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, errors.New("device not found")
	}
	return s.getLegacyDeviceByID(deviceID)
}

func (s *DeviceService) Create(req DeviceCreateRequest, tenantId, creatorId int64) (*models.Device, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if tenantId == 0 {
		tenantId = legacyTenantID
	}

	var createdID int64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		equipmentID, err := nextLegacyNumericID(tx, `"Auxiliary_EquipmentInfomation"`)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		maintenance := int64(0)
		if req.Maintenance != nil {
			maintenance = *req.Maintenance
		}
		equipment := legacyEquipment{
			ID:               equipmentID,
			TenantId:         tenantId,
			Name:             strings.TrimSpace(req.Name),
			IDNo:             strings.TrimSpace(req.IDNo),
			SerialNo:         strings.TrimSpace(req.SerialNo),
			Brand:            strings.TrimSpace(req.Brand),
			ModelNo:          resolveDeviceModel(req.Model, req.ModelNo),
			DialysisMethod:   strings.TrimSpace(req.DialysisMethod),
			Type:             resolveDeviceType(req.DeviceType, req.Type),
			ManufactureDate:  req.ManufactureDate,
			Manufacturer:     strings.TrimSpace(req.Manufacturer),
			InstallDate:      req.InstallDate,
			Maintenance:      maintenance,
			MaintenanceCycle: strings.TrimSpace(req.MaintenanceCycle),
			Note:             resolveDeviceNotes(req.Notes, req.Note),
			IsDisabled:       false,
			CreatorId:        creatorId,
			CreateTime:       now,
			LastModifyTime:   now,
			Flux:             strings.TrimSpace(req.Flux),
		}
		if err := tx.Create(&equipment).Error; err != nil {
			return err
		}

		bedID, err := s.resolveLegacyBedID(tx, req.BedNumber, req.WardId)
		if err != nil {
			return err
		}
		if bedID != nil {
			if err := s.upsertLegacyBedRelation(tx, tenantId, equipmentID, *bedID, models.DeviceStatusNormal); err != nil {
				return err
			}
		}
		createdID = equipmentID
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.getLegacyDeviceByID(createdID)
}

func (s *DeviceService) Update(id string, req DeviceUpdateRequest) (*models.Device, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	deviceID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, errors.New("device not found")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		var equipment legacyEquipment
		if err := tx.Where(`"Id" = ? AND "TenantId" = ? AND COALESCE("IsDisabled", false) = false`, deviceID, legacyTenantID).First(&equipment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("device not found")
			}
			return err
		}

		updates := map[string]interface{}{
			`LastModifyTime`: time.Now().UTC(),
		}
		if req.Name != nil {
			updates[`Name`] = strings.TrimSpace(*req.Name)
		}
		if req.IDNo != nil {
			updates[`IDNo`] = strings.TrimSpace(*req.IDNo)
		}
		if req.SerialNo != nil {
			updates[`SerialNo`] = strings.TrimSpace(*req.SerialNo)
		}
		if req.Brand != nil {
			updates[`Brand`] = strings.TrimSpace(*req.Brand)
		}
		if req.ModelNo != nil {
			updates[`ModelNo`] = strings.TrimSpace(*req.ModelNo)
		} else if req.Model != nil {
			updates[`ModelNo`] = strings.TrimSpace(*req.Model)
		}
		if req.DialysisMethod != nil {
			updates[`DialysisMethod`] = strings.TrimSpace(*req.DialysisMethod)
		}
		if req.DeviceType != nil {
			updates[`Type`] = strings.TrimSpace(*req.DeviceType)
		} else if req.Type != nil {
			updates[`Type`] = strings.TrimSpace(*req.Type)
		}
		if req.ManufactureDate != nil {
			updates[`ManufactureDate`] = req.ManufactureDate
		}
		if req.Manufacturer != nil {
			updates[`Manufacturer`] = strings.TrimSpace(*req.Manufacturer)
		}
		if req.InstallDate != nil {
			updates[`InstallDate`] = req.InstallDate
		}
		if req.Maintenance != nil {
			updates[`Maintenance`] = *req.Maintenance
		}
		if req.MaintenanceCycle != nil {
			updates[`MaintenanceCycle`] = strings.TrimSpace(*req.MaintenanceCycle)
		}
		if req.Flux != nil {
			updates[`Flux`] = strings.TrimSpace(*req.Flux)
		}
		if req.Note != nil {
			updates[`Note`] = strings.TrimSpace(*req.Note)
		} else if req.Notes != nil {
			updates[`Note`] = strings.TrimSpace(*req.Notes)
		}
		if err := tx.Model(&equipment).Updates(updates).Error; err != nil {
			return err
		}

		if req.BedNumber != nil || req.WardId != nil || req.Status != nil {
			currentBedNumber, currentWardID, err := s.currentLegacyBedInfo(tx, deviceID)
			if err != nil {
				return err
			}
			targetBedNumber := currentBedNumber
			if req.BedNumber != nil {
				targetBedNumber = strings.TrimSpace(*req.BedNumber)
			}
			targetWardID := currentWardID
			if req.WardId != nil {
				wardID := *req.WardId
				targetWardID = &wardID
			}

			status := models.DeviceStatusNormal
			if req.Status != nil {
				status = normalizeDeviceStatus(*req.Status)
			}
			if targetBedNumber == "" {
				return s.disableLegacyBedRelation(tx, deviceID)
			}
			bedID, err := s.resolveLegacyBedID(tx, targetBedNumber, targetWardID)
			if err != nil {
				return err
			}
			if bedID == nil {
				return s.disableLegacyBedRelation(tx, deviceID)
			}
			return s.upsertLegacyBedRelation(tx, equipment.TenantId, deviceID, *bedID, status)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.getLegacyDeviceByID(deviceID)
}

func (s *DeviceService) Delete(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	deviceID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return errors.New("device not found")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&legacyEquipment{}).
			Where(`"Id" = ? AND "TenantId" = ? AND COALESCE("IsDisabled", false) = false`, deviceID, legacyTenantID).
			Updates(map[string]interface{}{
				`IsDisabled`:     true,
				`LastModifyTime`: time.Now().UTC(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("device not found")
		}
		return s.disableLegacyBedRelation(tx, deviceID)
	})
}

func (s *DeviceService) UpdateStatus(id, status string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	status = normalizeDeviceStatus(status)
	validStatuses := map[string]bool{
		models.DeviceStatusNormal:      true,
		models.DeviceStatusWarning:     true,
		models.DeviceStatusAlarm:       true,
		models.DeviceStatusOffline:     true,
		models.DeviceStatusMaintenance: true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	deviceID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return errors.New("device not found")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var equipment legacyEquipment
		if err := tx.Where(`"Id" = ? AND "TenantId" = ? AND COALESCE("IsDisabled", false) = false`, deviceID, legacyTenantID).First(&equipment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("device not found")
			}
			return err
		}
		return s.updateLegacyDeviceStatus(tx, deviceID, status)
	})
}

func (s *DeviceService) ListUsageLogs(id string, req DeviceLogListRequest) (*DeviceUsageLogListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	equipmentID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, errors.New("device not found")
	}
	if _, err := s.getLegacyDeviceByID(equipmentID); err != nil {
		return nil, err
	}
	req.Page, req.PageSize = normalizeDevicePage(req.Page, req.PageSize, 20)
	query := s.db.Model(&legacyEquipmentUsageLog{}).Where(`"TenantId" = ? AND "EquipmentId" = ?`, legacyTenantID, equipmentID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []legacyEquipmentUsageLog
	if err := query.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Order(`"UseStartTime" DESC NULLS LAST`).Order(`"Id" DESC`).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]DeviceUsageLogDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, DeviceUsageLogDTO{
			ID:             row.ID,
			TenantId:       row.TenantId,
			EquipmentId:    row.EquipmentId,
			UseUserId:      row.UseUserId,
			UseStartTime:   row.UseStartTime,
			UseDuration:    row.UseDuration,
			Note:           strings.TrimSpace(row.Note),
			CreatorId:      row.CreatorId,
			CreateTime:     row.CreateTime,
			LastModifyTime: row.LastModifyTime,
		})
	}
	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}
	return &DeviceUsageLogListResponse{Items: items, Total: total, Page: req.Page, PageSize: req.PageSize, TotalPage: totalPage}, nil
}

func (s *DeviceService) ListMaintenanceRecords(id string, req DeviceLogListRequest) (*DeviceMaintenanceRecordListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	equipmentID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, errors.New("device not found")
	}
	if _, err := s.getLegacyDeviceByID(equipmentID); err != nil {
		return nil, err
	}
	req.Page, req.PageSize = normalizeDevicePage(req.Page, req.PageSize, 20)
	query := s.db.Model(&legacyEquipmentMaintenance{}).Where(`"TenantId" = ? AND "EquipmentId" = ?`, legacyTenantID, equipmentID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []legacyEquipmentMaintenance
	if err := query.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Order(`"OperateTime" DESC NULLS LAST`).Order(`"Id" DESC`).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]DeviceMaintenanceRecordDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, DeviceMaintenanceRecordDTO{
			ID:             row.ID,
			TenantId:       row.TenantId,
			EquipmentId:    row.EquipmentId,
			Type:           strings.TrimSpace(row.Type),
			Mode:           strings.TrimSpace(row.Mode),
			OperatorId:     row.OperatorId,
			OperateTime:    row.OperateTime,
			Description:    strings.TrimSpace(row.Description),
			Note:           strings.TrimSpace(row.Note),
			CreatorId:      row.CreatorId,
			CreateTime:     row.CreateTime,
			LastModifyTime: row.LastModifyTime,
		})
	}
	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}
	return &DeviceMaintenanceRecordListResponse{Items: items, Total: total, Page: req.Page, PageSize: req.PageSize, TotalPage: totalPage}, nil
}

func (s *DeviceService) ListDisinfectionRecords(id string, req DeviceLogListRequest) (*DeviceDisinfectionRecordListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	equipmentID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, errors.New("device not found")
	}
	if _, err := s.getLegacyDeviceByID(equipmentID); err != nil {
		return nil, err
	}
	req.Page, req.PageSize = normalizeDevicePage(req.Page, req.PageSize, 20)
	query := s.db.Model(&legacyEquipmentDisinfection{}).Where(`"TenantId" = ? AND "EquipmentId" = ?`, legacyTenantID, equipmentID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []legacyEquipmentDisinfection
	if err := query.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Order(`"StartTime" DESC NULLS LAST`).Order(`"Id" DESC`).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]DeviceDisinfectionRecordDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, DeviceDisinfectionRecordDTO{
			ID:              row.ID,
			TenantId:        row.TenantId,
			EquipmentId:     row.EquipmentId,
			DisinfectUserId: row.DisinfectUserId,
			DisinfectWay:    strings.TrimSpace(row.DisinfectWay),
			StartTime:       row.StartTime,
			EndTime:         row.EndTime,
			Description:     strings.TrimSpace(row.Description),
			Note:            strings.TrimSpace(row.Note),
			Type:            strings.TrimSpace(row.Type),
			Disinfectant:    strings.TrimSpace(row.Disinfectant),
			Status:          row.Status,
			TreatmentId:     row.TreatmentId,
			CreatorId:       row.CreatorId,
			CreateTime:      row.CreateTime,
			LastModifyTime:  row.LastModifyTime,
		})
	}
	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}
	return &DeviceDisinfectionRecordListResponse{Items: items, Total: total, Page: req.Page, PageSize: req.PageSize, TotalPage: totalPage}, nil
}

func nextLegacyNumericID(tx *gorm.DB, table string) (int64, error) {
	var nextID int64
	if err := tx.Raw(`SELECT COALESCE(MAX("Id"), 0) + 1 FROM ` + table).Scan(&nextID).Error; err != nil {
		return 0, err
	}
	return nextID, nil
}

func (s *DeviceService) resolveLegacyBedID(tx *gorm.DB, bedNumber string, wardID *int64) (*int64, error) {
	bedNumber = strings.TrimSpace(bedNumber)
	if bedNumber == "" {
		return nil, nil
	}
	query := tx.Model(&legacyBed{}).Where(`"TenantId" = ? AND "Name" = ? AND COALESCE("IsDisabled", false) = false`, legacyTenantID, bedNumber)
	if wardID != nil {
		query = query.Where(`"WardId" = ?`, *wardID)
	}
	var bed legacyBed
	if err := query.Order(`"Sort" ASC`).First(&bed).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("bed not found: %s", bedNumber)
		}
		return nil, err
	}
	return &bed.ID, nil
}

func (s *DeviceService) getLegacyDeviceByID(deviceID int64) (*models.Device, error) {
	var rows []legacyDeviceRecord
	if err := s.baseLegacyDeviceQuery().
		Where(`e."Id" = ?`, deviceID).
		Select(s.legacyDeviceSelect(), models.DeviceStatusNormal).
		Limit(1).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("device not found")
	}
	device := toDeviceDTO(rows[0])
	return &device, nil
}

func (s *DeviceService) upsertLegacyBedRelation(tx *gorm.DB, tenantID, equipmentID, bedID int64, status string) error {
	var relation legacyBedEquipmentRel
	now := time.Now().UTC()
	err := tx.Where(`"EquipmentId" = ? AND "TenantId" = ? AND "IsDefault" = true`, equipmentID, tenantID).First(&relation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		relationID, nextErr := nextLegacyNumericID(tx, `"Schedule_BedEquipmentRel"`)
		if nextErr != nil {
			return nextErr
		}
		relation = legacyBedEquipmentRel{
			ID:             relationID,
			TenantId:       tenantID,
			EquipmentId:    equipmentID,
			Sort:           10,
			BedId:          bedID,
			IsDefault:      true,
			IsDisabled:     false,
			LastModifyTime: &now,
			Type:           1,
			ParameterS:     normalizeDeviceStatus(status),
		}
		if err := tx.Create(&relation).Error; err != nil {
			return err
		}
		return s.appendLegacyBedRelationChange(tx, relation, 0)
	}
	if err != nil {
		return err
	}
	if err := tx.Model(&relation).Updates(map[string]interface{}{
		`BedId`:          bedID,
		`IsDefault`:      true,
		`IsDisabled`:     false,
		`LastModifyTime`: now,
		`ParameterS`:     normalizeDeviceStatus(status),
	}).Error; err != nil {
		return err
	}
	relation.BedId = bedID
	relation.IsDefault = true
	relation.IsDisabled = false
	relation.ParameterS = normalizeDeviceStatus(status)
	relation.LastModifyTime = &now
	return s.appendLegacyBedRelationChange(tx, relation, 0)
}

func (s *DeviceService) disableLegacyBedRelation(tx *gorm.DB, equipmentID int64) error {
	now := time.Now().UTC()
	var relation legacyBedEquipmentRel
	if err := tx.Where(`"EquipmentId" = ? AND "TenantId" = ? AND "IsDefault" = true`, equipmentID, legacyTenantID).First(&relation).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := tx.Model(&legacyBedEquipmentRel{}).
		Where(`"EquipmentId" = ? AND "TenantId" = ?`, equipmentID, legacyTenantID).
		Updates(map[string]interface{}{`IsDisabled`: true, `LastModifyTime`: now}).Error; err != nil {
		return err
	}
	if relation.ID > 0 {
		relation.IsDisabled = true
		relation.LastModifyTime = &now
		return s.appendLegacyBedRelationChange(tx, relation, 0)
	}
	return nil
}

func (s *DeviceService) currentLegacyBedInfo(tx *gorm.DB, equipmentID int64) (string, *int64, error) {
	var row struct {
		BedName string `gorm:"column:bed_name"`
		WardID  *int64 `gorm:"column:ward_id"`
	}
	err := tx.Table(`"Schedule_BedEquipmentRel" AS rel`).
		Joins(`JOIN "Schedule_Bed" AS bed ON bed."Id" = rel."BedId"`).
		Where(`rel."EquipmentId" = ? AND rel."TenantId" = ? AND COALESCE(rel."IsDisabled", false) = false AND rel."IsDefault" = true`, equipmentID, legacyTenantID).
		Select(`bed."Name" AS bed_name, bed."WardId" AS ward_id`).
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return "", nil, err
	}
	return row.BedName, row.WardID, nil
}

func (s *DeviceService) updateLegacyDeviceStatus(tx *gorm.DB, equipmentID int64, status string) error {
	var relation legacyBedEquipmentRel
	err := tx.Where(`"EquipmentId" = ? AND "TenantId" = ? AND "IsDefault" = true`, equipmentID, legacyTenantID).First(&relation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return tx.Model(&relation).Updates(map[string]interface{}{
		`ParameterS`:     normalizeDeviceStatus(status),
		`LastModifyTime`: time.Now().UTC(),
		`IsDisabled`:     false,
	}).Error
}

func (s *DeviceService) appendLegacyBedRelationChange(tx *gorm.DB, rel legacyBedEquipmentRel, creatorID int64) error {
	changeID, err := nextLegacyNumericID(tx, `"Schedule_BedEquipmentRelChange"`)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	change := legacyBedEquipmentRelChange{
		ID:             changeID,
		TenantId:       rel.TenantId,
		EquipmentId:    rel.EquipmentId,
		Sort:           rel.Sort,
		BedId:          rel.BedId,
		IsDefault:      rel.IsDefault,
		IsDisabled:     rel.IsDisabled,
		Type:           rel.Type,
		ParameterS:     rel.ParameterS,
		CreatorId:      creatorID,
		CreateTime:     now,
		LastModifyTime: rel.LastModifyTime,
	}
	return tx.Create(&change).Error
}
