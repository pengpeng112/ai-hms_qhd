package services

import (
	"errors"
	"fmt"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"gorm.io/gorm"
)

type BedService struct {
	db *gorm.DB
}

func NewBedService() *BedService {
	return &BedService{db: database.GetDB()}
}

type BedDTO struct {
	ID                 int64              `json:"id"`
	Name               string             `json:"name"`
	WardID             int64              `json:"wardId"`
	WardName           string             `json:"wardName"`
	Sort               int64              `json:"sort"`
	Note               string             `json:"note"`
	FEPId              *int64             `json:"fepId"`
	FEPName            string             `json:"fepName"`
	AcquisiteConnectId *int64             `json:"acquisiteConnectId"`
	Equipments         []BedEquipmentDTO  `json:"equipments"`
	DefaultEquipmentName string           `json:"defaultEquipmentName"`
	EquipmentCount     int64              `json:"equipmentCount"`
	IsDisabled         bool               `json:"isDisabled"`
}

type BedEquipmentDTO struct {
	EquipmentId   int64  `json:"equipmentId"`
	EquipmentName string `json:"equipmentName"`
	IsDefault     bool   `json:"isDefault"`
	Sort          int64  `json:"sort"`
}

type BedCreateRequest struct {
	Name               string             `json:"name" binding:"required"`
	WardID             int64              `json:"wardId" binding:"required"`
	Sort               int64              `json:"sort"`
	Note               string             `json:"note"`
	IsDisabled         bool               `json:"isDisabled"`
	FEPId              *int64             `json:"fepId"`
	AcquisiteConnectId *int64             `json:"acquisiteConnectId"`
	Equipments         []BedEquipmentDTO  `json:"equipments"`
}

type BedUpdateRequest struct {
	Name               *string            `json:"name"`
	WardID             *int64             `json:"wardId"`
	Sort               *int64             `json:"sort"`
	Note               *string            `json:"note"`
	IsDisabled         *bool              `json:"isDisabled"`
	FEPId              *int64             `json:"fepId"`
	AcquisiteConnectId *int64             `json:"acquisiteConnectId"`
	Equipments         []BedEquipmentDTO  `json:"equipments"`
}

type bedRawRow struct {
	ID                 int64  `gorm:"column:Id"`
	Name               string `gorm:"column:Name"`
	Sort               int64  `gorm:"column:Sort"`
	WardID             int64  `gorm:"column:WardId"`
	IsDisabled         bool   `gorm:"column:IsDisabled"`
	Note               string `gorm:"column:Note"`
	FEPId              *int64 `gorm:"column:FEPId"`
	AcquisiteConnectId *int64 `gorm:"column:AcquisiteConnectId"`
	WardName           string `gorm:"column:WardName"`
}

type bedEquipmentRawRow struct {
	ID            int64  `gorm:"column:Id"`
	EquipmentId   int64  `gorm:"column:EquipmentId"`
	BedId         int64  `gorm:"column:BedId"`
	Sort          int64  `gorm:"column:Sort"`
	IsDefault     bool   `gorm:"column:IsDefault"`
	IsDisabled    bool   `gorm:"column:IsDisabled"`
	EquipmentName string `gorm:"column:EquipmentName"`
}

func (s *BedService) List(includeDisabled bool) ([]BedDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	query := s.db.Table(`"Schedule_Bed" AS b`).
		Select(`b."Id", b."Name", b."Sort", COALESCE(b."WardId", 0) AS "WardId", b."IsDisabled", COALESCE(b."Note", '') AS "Note", b."FEPId", b."AcquisiteConnectId", COALESCE(w."Name", '') AS "WardName"`).
		Joins(`LEFT JOIN "Schedule_Ward" AS w ON w."Id" = b."WardId"`)

	if !includeDisabled {
		query = query.Where(`b."IsDisabled" = false`)
	}

	var rows []bedRawRow
	if err := query.Order(`b."Sort" ASC, b."Id" ASC`).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询床位列表失败: %w", err)
	}

	bedIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		bedIDs = append(bedIDs, row.ID)
	}

	equipMap := make(map[int64][]bedEquipmentRawRow)
	if len(bedIDs) > 0 {
		var equipRows []bedEquipmentRawRow
		if err := s.db.Table(`"Schedule_BedEquipmentRel" AS r`).
			Select(`r."Id", r."EquipmentId", r."BedId", COALESCE(r."Sort", 0) AS "Sort", COALESCE(r."IsDefault", false) AS "IsDefault", COALESCE(r."IsDisabled", false) AS "IsDisabled", COALESCE(eq."Name", '') AS "EquipmentName"`).
		Joins(`LEFT JOIN "Auxiliary_EquipmentInfomation" AS eq ON eq."Id" = r."EquipmentId"`).
		Where(`r."BedId" IN ? AND r."IsDisabled" = false`, bedIDs).
			Find(&equipRows).Error; err != nil {
			return nil, fmt.Errorf("查询床位设备失败: %w", err)
		}
		for _, eq := range equipRows {
			equipMap[eq.BedId] = append(equipMap[eq.BedId], eq)
		}
	}

	result := make([]BedDTO, 0, len(rows))
	for _, row := range rows {
		var defaultName string
		var equipCount int64
		equipList := make([]BedEquipmentDTO, 0)
		for _, eq := range equipMap[row.ID] {
			equipList = append(equipList, BedEquipmentDTO{
				EquipmentId:   eq.EquipmentId,
				EquipmentName: eq.EquipmentName,
				IsDefault:     eq.IsDefault,
				Sort:          eq.Sort,
			})
			if eq.IsDefault {
				defaultName = eq.EquipmentName
			}
		}
		equipCount = int64(len(equipList))

		dto := BedDTO{
			ID:                   row.ID,
			Name:                 row.Name,
			WardID:               row.WardID,
			WardName:             row.WardName,
			Sort:                 row.Sort,
			Note:                 row.Note,
			FEPId:                row.FEPId,
			AcquisiteConnectId:   row.AcquisiteConnectId,
			Equipments:           equipList,
			DefaultEquipmentName: defaultName,
			EquipmentCount:       equipCount,
			IsDisabled:           row.IsDisabled,
		}
		result = append(result, dto)
	}
	return result, nil
}

func (s *BedService) GetByID(id int64) (*BedDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var row bedRawRow
	if err := s.db.Table(`"Schedule_Bed" AS b`).
		Select(`b."Id", b."Name", b."Sort", COALESCE(b."WardId", 0) AS "WardId", b."IsDisabled", COALESCE(b."Note", '') AS "Note", b."FEPId", b."AcquisiteConnectId", COALESCE(w."Name", '') AS "WardName"`).
		Joins(`LEFT JOIN "Schedule_Ward" AS w ON w."Id" = b."WardId"`).
		Where(`b."Id" = ?`, id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("bed not found")
		}
		return nil, fmt.Errorf("查询床位失败: %w", err)
	}

	var equipRows []bedEquipmentRawRow
	s.db.Table(`"Schedule_BedEquipmentRel" AS r`).
		Select(`r."Id", r."EquipmentId", r."BedId", COALESCE(r."Sort", 0) AS "Sort", COALESCE(r."IsDefault", false) AS "IsDefault", COALESCE(r."IsDisabled", false) AS "IsDisabled", COALESCE(eq."Name", '') AS "EquipmentName"`).
		Joins(`LEFT JOIN "Auxiliary_EquipmentInfomation" AS eq ON eq."Id" = r."EquipmentId"`).
		Where(`r."BedId" = ? AND r."IsDisabled" = false`, id).Find(&equipRows)

	var defaultName string
	equipList := make([]BedEquipmentDTO, 0)
	for _, eq := range equipRows {
		equipList = append(equipList, BedEquipmentDTO{
			EquipmentId:   eq.EquipmentId,
			EquipmentName: eq.EquipmentName,
			IsDefault:     eq.IsDefault,
			Sort:          eq.Sort,
		})
		if eq.IsDefault {
			defaultName = eq.EquipmentName
		}
	}

	return &BedDTO{
		ID:                   row.ID,
		Name:                 row.Name,
		WardID:               row.WardID,
		WardName:             row.WardName,
		Sort:                 row.Sort,
		Note:                 row.Note,
		FEPId:                row.FEPId,
		AcquisiteConnectId:   row.AcquisiteConnectId,
		Equipments:           equipList,
		DefaultEquipmentName: defaultName,
		EquipmentCount:       int64(len(equipList)),
		IsDisabled:           row.IsDisabled,
	}, nil
}

func (s *BedService) Create(req BedCreateRequest) (*BedDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	columns := map[string]interface{}{
		`"Name"`:               req.Name,
		`"WardId"`:             req.WardID,
		`"Sort"`:               req.Sort,
		`"Note"`:               req.Note,
		`"IsDisabled"`:         req.IsDisabled,
		`"FEPId"`:              req.FEPId,
		`"AcquisiteConnectId"`: req.AcquisiteConnectId,
	}
	result := s.db.Table(`"Schedule_Bed"`).Create(columns)
	if result.Error != nil {
		return nil, fmt.Errorf("创建床位失败: %w", result.Error)
	}
	var newID int64
	s.db.Raw(`SELECT LASTVAL()`).Scan(&newID)
	if newID == 0 {
		s.db.Table(`"Schedule_Bed"`).Select(`MAX("Id")`).Scan(&newID)
	}
	s.syncBedEquipments(newID, req.Equipments)
	return s.GetByID(newID)
}

func (s *BedService) Update(id int64, req BedUpdateRequest) (*BedDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	columns := map[string]interface{}{}
	if req.Name != nil {
		columns[`"Name"`] = *req.Name
	}
	if req.WardID != nil {
		columns[`"WardId"`] = *req.WardID
	}
	if req.Sort != nil {
		columns[`"Sort"`] = *req.Sort
	}
	if req.IsDisabled != nil {
		columns[`"IsDisabled"`] = *req.IsDisabled
	}
	if req.Note != nil {
		columns[`"Note"`] = *req.Note
	}
	if req.FEPId != nil {
		columns[`"FEPId"`] = *req.FEPId
	}
	if req.AcquisiteConnectId != nil {
		columns[`"AcquisiteConnectId"`] = *req.AcquisiteConnectId
	}
	if len(columns) > 0 {
		result := s.db.Table(`"Schedule_Bed"`).Where(`"Id" = ?`, id).Updates(columns)
		if result.Error != nil {
			return nil, fmt.Errorf("更新床位失败: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return nil, fmt.Errorf("bed not found")
		}
	}
	if req.Equipments != nil {
		s.syncBedEquipments(id, req.Equipments)
	}
	return s.GetByID(id)
}

func (s *BedService) Delete(id int64) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	result := s.db.Table(`"Schedule_Bed"`).Where(`"Id" = ?`, id).Update(`"IsDisabled"`, true)
	if result.Error != nil {
		return fmt.Errorf("删除床位失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("bed not found")
	}
	s.db.Table(`"Schedule_BedEquipmentRel"`).Where(`"BedId" = ?`, id).Update(`"IsDisabled"`, true)
	return nil
}

func (s *BedService) syncBedEquipments(bedID int64, equipments []BedEquipmentDTO) {
	s.db.Table(`"Schedule_BedEquipmentRel"`).Where(`"BedId" = ?`, bedID).Delete(nil)
	for _, eq := range equipments {
		columns := map[string]interface{}{
			`"BedId"`:       bedID,
			`"EquipmentId"`: eq.EquipmentId,
			`"Sort"`:        eq.Sort,
			`"IsDefault"`:   eq.IsDefault,
			`"IsDisabled"`:  false,
		}
		s.db.Table(`"Schedule_BedEquipmentRel"`).Create(columns)
	}
}
