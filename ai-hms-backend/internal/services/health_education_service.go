package services

import (
	"errors"
	"fmt"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"gorm.io/gorm"
)

type HealthEducationService struct {
	db *gorm.DB
}

func NewHealthEducationService() *HealthEducationService {
	return &HealthEducationService{db: database.GetDB()}
}

type HealthEducationContentDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Sort          string `json:"sort"`
	AttachmentIDs string `json:"attachmentIds"`
	Type          string `json:"type"`
	Classify      string `json:"classify"`
	Note          string `json:"note"`
	IsDisabled    bool   `json:"isDisabled"`
}

type PatientHealthEducationDTO struct {
	ID                   string `json:"id"`
	PatientID            string `json:"patientId"`
	HealthEducationID     string `json:"healthEducationId"`
	HealthEducationName   string `json:"healthEducationName"`
	OperatorID           string `json:"operatorId"`
	OperatorName         string `json:"operatorName"`
	EducationTime         string `json:"educationTime"`
	EducationType        string `json:"educationType"`
	EducationResult      string `json:"educationResult"`
	NurseSign            string `json:"nurseSign"`
	PatientSign          string `json:"patientSign"`
	FinishTime           string `json:"finishTime"`
	Note                 string `json:"note"`
	CreatedAt            string `json:"createdAt"`
}

type CreatePatientHealthEducationRequest struct {
	HealthEducationID string `json:"healthEducationId" binding:"required"`
	OperatorID        string `json:"operatorId"`
	EducationTime     string `json:"educationTime"`
	EducationType     string `json:"educationType"`
	EducationResult   string `json:"educationResult"`
	NurseSign         string `json:"nurseSign"`
	PatientSign       string `json:"patientSign"`
	FinishTime        string `json:"finishTime"`
	Note              string `json:"note"`
}

type rawHealthEducationRow struct {
	ID            int64  `gorm:"column:Id"`
	Name          string `gorm:"column:Name"`
	Description   string `gorm:"column:Description"`
	Sort          string `gorm:"column:Sort"`
	AttachmentIDs string `gorm:"column:AttachmentIds"`
	Type          string `gorm:"column:Type"`
	Classify      string `gorm:"column:Classify"`
}

type rawPatientHealthEducationRow struct {
	ID                 int64  `gorm:"column:Id"`
	PatientID          int64  `gorm:"column:PatientId"`
	HealthEducationID  int64  `gorm:"column:HealthEducationId"`
	OperatorID         int64  `gorm:"column:OperatorId"`
	EducationTime      string `gorm:"column:EducationTime"`
	EducationType      string `gorm:"column:EducationType"`
	EducationResult    string `gorm:"column:EducationResult"`
	NurseSign          string `gorm:"column:NurseSign"`
	PatientSign        string `gorm:"column:PatientSign"`
	FinishTime         string `gorm:"column:FinishTime"`
	Note               string `gorm:"column:Note"`
	CreatedAt          string `gorm:"column:CreateTime"`
	HeName             string `gorm:"column:HeName"`
	OpName             string `gorm:"column:OpName"`
}

func (s *HealthEducationService) ListContents() ([]HealthEducationContentDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var rows []rawHealthEducationRow
	err := s.db.Table(`"Auxiliary_HealthEducation"`).
		Select(`"Id", "Name", "Description", "Sort", "AttachmentIds", "Type", "Classify"`).
		Order(`COALESCE(CAST("Sort" AS NUMERIC), 0) ASC, "Id" ASC`).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("query health education contents failed: %w", err)
	}
	result := make([]HealthEducationContentDTO, 0, len(rows))
	for _, row := range rows {
		result = append(result, HealthEducationContentDTO{
			ID:            fmt.Sprintf("%d", row.ID),
			Name:          row.Name,
			Description:   row.Description,
			Sort:          row.Sort,
			AttachmentIDs: row.AttachmentIDs,
			Type:          row.Type,
			Classify:      row.Classify,
		})
	}
	return result, nil
}

func (s *HealthEducationService) ListPatientEducations(patientIDStr string) ([]PatientHealthEducationDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var rows []rawPatientHealthEducationRow
	err := s.db.Table(`"Auxiliary_PatientHealthEducation" AS pe`).
		Select(`pe."Id", pe."PatientId", pe."HealthEducationId", pe."OperatorId", pe."EducationTime", pe."EducationType", pe."EducationResult", pe."NurseSign", pe."PatientSign", pe."FinishTime", pe."Note", pe."CreateTime", COALESCE(he."Name", '') AS "HeName", COALESCE(op."Name", '') AS "OpName"`).
		Joins(`LEFT JOIN "Auxiliary_HealthEducation" AS he ON he."Id" = pe."HealthEducationId"`).
		Joins(`LEFT JOIN "Organ_Employee" AS op ON op."UserId" = pe."OperatorId"`).
		Where(`pe."PatientId" = ?`, patientIDStr).
		Order(`pe."CreateTime" DESC`).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("query patient health educations failed: %w", err)
	}
	result := make([]PatientHealthEducationDTO, 0, len(rows))
	for _, row := range rows {
		dto := PatientHealthEducationDTO{
			ID:                 fmt.Sprintf("%d", row.ID),
			PatientID:          fmt.Sprintf("%d", row.PatientID),
			HealthEducationID:  fmt.Sprintf("%d", row.HealthEducationID),
			HealthEducationName: row.HeName,
			OperatorID:         fmt.Sprintf("%d", row.OperatorID),
			OperatorName:       row.OpName,
			EducationTime:      row.EducationTime,
			EducationType:      row.EducationType,
			EducationResult:    row.EducationResult,
			NurseSign:          row.NurseSign,
			PatientSign:        row.PatientSign,
			FinishTime:         row.FinishTime,
			Note:               row.Note,
			CreatedAt:          row.CreatedAt,
		}
		result = append(result, dto)
	}
	return result, nil
}

func (s *HealthEducationService) CreatePatientEducation(patientIDStr string, req CreatePatientHealthEducationRequest) (*PatientHealthEducationDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var operatorID int64
	if req.OperatorID != "" {
		fmt.Sscanf(req.OperatorID, "%d", &operatorID)
	}
	columns := map[string]interface{}{
		`"PatientId"`:         patientIDStr,
		`"HealthEducationId"`: req.HealthEducationID,
		`"OperatorId"`:         operatorID,
		`"EducationTime"`:      req.EducationTime,
		`"EducationType"`:      req.EducationType,
		`"EducationResult"`:    req.EducationResult,
		`"NurseSign"`:          req.NurseSign,
		`"PatientSign"`:        req.PatientSign,
		`"FinishTime"`:         req.FinishTime,
		`"Note"`:               req.Note,
	}
	result := s.db.Table(`"Auxiliary_PatientHealthEducation"`).Create(columns)
	if result.Error != nil {
		return nil, fmt.Errorf("create patient health education failed: %w", result.Error)
	}
	var id int64
	s.db.Raw(`SELECT LASTVAL()`).Scan(&id)
	if id == 0 {
		var maxID int64
		s.db.Table(`"Auxiliary_PatientHealthEducation"`).Select(`MAX("Id")`).Scan(&maxID)
		id = maxID
	}
	items, err := s.ListPatientEducations(patientIDStr)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.ID == fmt.Sprintf("%d", id) {
			return &item, nil
		}
	}
	return &PatientHealthEducationDTO{
		ID:                  fmt.Sprintf("%d", id),
		PatientID:           patientIDStr,
		HealthEducationID:   req.HealthEducationID,
		OperatorID:          req.OperatorID,
		EducationTime:       req.EducationTime,
		EducationType:       req.EducationType,
		EducationResult:     req.EducationResult,
		NurseSign:           req.NurseSign,
		PatientSign:         req.PatientSign,
		FinishTime:          req.FinishTime,
		Note:                req.Note,
	}, nil
}

type CreateHealthEducationContentRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Sort        string `json:"sort"`
	Type        string `json:"type"`
	Classify    string `json:"classify"`
	Note        string `json:"note"`
}

type UpdateHealthEducationContentRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Sort        *string `json:"sort"`
	Type        *string `json:"type"`
	Classify    *string `json:"classify"`
	Note        *string `json:"note"`
	IsDisabled  *bool   `json:"isDisabled"`
}

func (s *HealthEducationService) CreateContent(req CreateHealthEducationContentRequest) (*HealthEducationContentDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	columns := map[string]interface{}{
		`"Name"`:        req.Name,
		`"Description"`: req.Description,
		`"Sort"`:        req.Sort,
		`"Type"`:        req.Type,
		`"Classify"`:    req.Classify,
		`"Note"`:        req.Note,
	}
	result := s.db.Table(`"Auxiliary_HealthEducation"`).Create(columns)
	if result.Error != nil {
		return nil, fmt.Errorf("创建宣教内容失败: %w", result.Error)
	}
	var newID int64
	s.db.Raw(`SELECT LASTVAL()`).Scan(&newID)
	if newID == 0 {
		s.db.Table(`"Auxiliary_HealthEducation"`).Select(`MAX("Id")`).Scan(&newID)
	}
	return &HealthEducationContentDTO{
		ID:          fmt.Sprintf("%d", newID),
		Name:        req.Name,
		Description: req.Description,
		Sort:        req.Sort,
		Type:        req.Type,
		Classify:    req.Classify,
		Note:        req.Note,
	}, nil
}

func (s *HealthEducationService) UpdateContent(id string, req UpdateHealthEducationContentRequest) (*HealthEducationContentDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	columns := map[string]interface{}{}
	if req.Name != nil {
		columns[`"Name"`] = *req.Name
	}
	if req.Description != nil {
		columns[`"Description"`] = *req.Description
	}
	if req.Sort != nil {
		columns[`"Sort"`] = *req.Sort
	}
	if req.Type != nil {
		columns[`"Type"`] = *req.Type
	}
	if req.Classify != nil {
		columns[`"Classify"`] = *req.Classify
	}
	if req.Note != nil {
		columns[`"Note"`] = *req.Note
	}
	if req.IsDisabled != nil {
		columns[`"IsDisabled"`] = *req.IsDisabled
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}
	result := s.db.Table(`"Auxiliary_HealthEducation"`).Where(`"Id" = ?`, id).Updates(columns)
	if result.Error != nil {
		return nil, fmt.Errorf("更新宣教内容失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("content not found")
	}
	contents, err := s.ListContents()
	if err != nil {
		return nil, err
	}
	for _, c := range contents {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("content not found after update")
}

func (s *HealthEducationService) DeleteContent(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	result := s.db.Table(`"Auxiliary_HealthEducation"`).Where(`"Id" = ?`, id).Delete(nil)
	if result.Error != nil {
		return fmt.Errorf("删除宣教内容失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("content not found")
	}
	return nil
}