package services

import (
	"errors"
	"fmt"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"gorm.io/gorm"
)

type WardService struct {
	db *gorm.DB
}

func NewWardService() *WardService {
	return &WardService{db: database.GetDB()}
}

var wardPatientTypeReverseMap = map[string]string{
	"10": "长期患者", "20": "临时患者",
}

func wardPatientTypeLabel(code string) string {
	if label, ok := wardPatientTypeReverseMap[code]; ok {
		return label
	}
	return code
}

type WardDTO struct {
	ID                   int64  `json:"id"`
	Name                 string `json:"name"`
	Sort                 int    `json:"sort"`
	PatientType          string `json:"patientType"`
	PatientTypeLabel     string `json:"patientTypeLabel"`
	InfectionType        string `json:"infectionType"`
	IsDisabled           bool   `json:"isDisabled"`
	Note                 string `json:"note"`
	ResponsibleUsers     string `json:"responsibleUsers"`
	ResponsibleUserNames string `json:"responsibleUserNames"`
	BedCount             int64  `json:"bedCount"`
}

type WardCreateRequest struct {
	Name             string `json:"name" binding:"required"`
	Sort             int    `json:"sort"`
	PatientType      string `json:"patientType"`
	InfectionType    string `json:"infectionType"`
	Note             string `json:"note"`
	IsDisabled       bool   `json:"isDisabled"`
	ResponsibleUsers string `json:"responsibleUsers"`
}

type WardUpdateRequest struct {
	Name             *string `json:"name"`
	Sort             *int    `json:"sort"`
	PatientType      *string `json:"patientType"`
	InfectionType    *string `json:"infectionType"`
	Note             *string `json:"note"`
	IsDisabled       *bool   `json:"isDisabled"`
	ResponsibleUsers *string `json:"responsibleUsers"`
}

type wardListRow struct {
	ID               int64  `gorm:"column:Id"`
	Name             string `gorm:"column:Name"`
	Sort             int    `gorm:"column:Sort"`
	PatientType      string `gorm:"column:PatientType"`
	InfectionType    string `gorm:"column:InfectionType"`
	IsDisabled       bool   `gorm:"column:IsDisabled"`
	Note             string `gorm:"column:Note"`
	ResponsibleUsers string `gorm:"column:ResponsibleUsers"`
	BedCount         int64  `gorm:"column:BedCount"`
}

type wardRawRow struct {
	ID               int64  `gorm:"column:Id"`
	Name             string `gorm:"column:Name"`
	Sort             int    `gorm:"column:Sort"`
	PatientType      string `gorm:"column:PatientType"`
	InfectionType    string `gorm:"column:InfectionType"`
	IsDisabled       bool   `gorm:"column:IsDisabled"`
	Note             string `gorm:"column:Note"`
	ResponsibleUsers string `gorm:"column:ResponsibleUsers"`
	BedCount         int64  `gorm:"column:BedCount"`
}

func (s *WardService) List(includeDisabled bool) ([]WardDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	query := s.db.Table(`"Schedule_Ward"`).Select(
		`"Id", "Name", "Sort", "PatientType", "InfectionType", "IsDisabled", "Note", "ResponsibleUsers", ` +
			`COALESCE((SELECT COUNT(*) FROM "Schedule_Bed" WHERE "WardId" = "Schedule_Ward"."Id" AND COALESCE("IsDisabled", false) = false), 0) AS "BedCount"`,
	)
	if !includeDisabled {
		query = query.Where(`"IsDisabled" = false`)
	}

	var rows []wardRawRow
	if err := query.Order(`"Sort" ASC, "Id" ASC`).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询病区列表失败: %w", err)
	}

	result := make([]WardDTO, 0, len(rows))
	for _, row := range rows {
		dto := WardDTO{
			ID:               row.ID,
			Name:             row.Name,
			Sort:             row.Sort,
			PatientType:      row.PatientType,
			PatientTypeLabel: wardPatientTypeLabel(row.PatientType),
			InfectionType:    row.InfectionType,
			IsDisabled:       row.IsDisabled,
			Note:             row.Note,
			ResponsibleUsers: row.ResponsibleUsers,
			BedCount:         row.BedCount,
		}
		result = append(result, dto)
	}
	return result, nil
}

func (s *WardService) GetByID(id int64) (*WardDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var row wardRawRow
	if err := s.db.Table(`"Schedule_Ward"`).Select(
		`"Id", "Name", "Sort", "PatientType", "InfectionType", "IsDisabled", "Note", "ResponsibleUsers", `+
			`COALESCE((SELECT COUNT(*) FROM "Schedule_Bed" WHERE "WardId" = "Schedule_Ward"."Id" AND COALESCE("IsDisabled", false) = false), 0) AS "BedCount"`,
	).Where(`"Id" = ?`, id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ward not found")
		}
		return nil, fmt.Errorf("查询病区失败: %w", err)
	}
	return &WardDTO{
		ID:               row.ID,
		Name:             row.Name,
		Sort:             row.Sort,
		PatientType:      row.PatientType,
		PatientTypeLabel: wardPatientTypeLabel(row.PatientType),
		InfectionType:    row.InfectionType,
		IsDisabled:       row.IsDisabled,
		Note:             row.Note,
		ResponsibleUsers: row.ResponsibleUsers,
		BedCount:         row.BedCount,
	}, nil
}

var wardPatientTypeMap = map[string]string{
	"长期患者": "10", "临时患者": "20",
	"长期": "10", "临时": "20",
	"10": "10", "20": "20",
}

func normalizeWardPatientType(v string) string {
	if mapped, ok := wardPatientTypeMap[v]; ok {
		return mapped
	}
	return v
}

func (s *WardService) Create(req WardCreateRequest) (*WardDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	columns := map[string]interface{}{
		`"Name"`:             req.Name,
		`"Sort"`:             req.Sort,
		`"PatientType"`:      normalizeWardPatientType(req.PatientType),
		`"InfectionType"`:    req.InfectionType,
		`"Note"`:             req.Note,
		`"IsDisabled"`:       req.IsDisabled,
		`"ResponsibleUsers"`: req.ResponsibleUsers,
	}
	result := s.db.Table(`"Schedule_Ward"`).Create(columns)
	if result.Error != nil {
		return nil, fmt.Errorf("创建病区失败: %w", result.Error)
	}
	var newID int64
	s.db.Raw(`SELECT LASTVAL()`).Scan(&newID)
	if newID == 0 {
		s.db.Table(`"Schedule_Ward"`).Select(`MAX("Id")`).Scan(&newID)
	}
	return s.GetByID(newID)
}

func (s *WardService) Update(id int64, req WardUpdateRequest) (*WardDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	columns := map[string]interface{}{}
	if req.Name != nil {
		columns[`"Name"`] = *req.Name
	}
	if req.Sort != nil {
		columns[`"Sort"`] = *req.Sort
	}
	if req.PatientType != nil {
		columns[`"PatientType"`] = normalizeWardPatientType(*req.PatientType)
	}
	if req.InfectionType != nil {
		columns[`"InfectionType"`] = *req.InfectionType
	}
	if req.IsDisabled != nil {
		columns[`"IsDisabled"`] = *req.IsDisabled
	}
	if req.Note != nil {
		columns[`"Note"`] = *req.Note
	}
	if req.ResponsibleUsers != nil {
		columns[`"ResponsibleUsers"`] = *req.ResponsibleUsers
	}
	if len(columns) == 0 {
		return s.GetByID(id)
	}
	result := s.db.Table(`"Schedule_Ward"`).Where(`"Id" = ?`, id).Updates(columns)
	if result.Error != nil {
		return nil, fmt.Errorf("更新病区失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("ward not found")
	}
	return s.GetByID(id)
}

func (s *WardService) Delete(id int64) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	result := s.db.Table(`"Schedule_Ward"`).Where(`"Id" = ?`, id).Update(`"IsDisabled"`, true)
	if result.Error != nil {
		return fmt.Errorf("删除病区失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ward not found")
	}
	return nil
}
