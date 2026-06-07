package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func strDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

type VascularAccessService struct {
	db *gorm.DB
}

type vascularAccessWithImages struct {
	models.VascularAccess
	Images []string
}

func NewVascularAccessService() *VascularAccessService {
	return &VascularAccessService{
		db: database.GetDB(),
	}
}

type VascularAccessResponse struct {
	ID                string   `json:"id"`
	AccessType        string   `json:"accessType"`
	Site              string   `json:"site"`
	Artery            []string `json:"artery"`
	Vein              []string `json:"vein"`
	Side              string   `json:"side"`
	Hospital          string   `json:"hospital"`
	Surgeon           string   `json:"surgeon"`
	SurgeryDate       string   `json:"surgeryDate"`
	FirstUseDate      string   `json:"firstUseDate"`
	AccessNumber      int      `json:"accessNumber"`
	InterventionCount int      `json:"interventionCount"`
	InterventionDate  string   `json:"interventionDate"`
	CatheterMethod    *string  `json:"catheterMethod"`
	CatheterDepth     *string  `json:"catheterDepth"`
	VPuncturePosition []string `json:"vPuncturePosition"`
	APuncturePosition []string `json:"aPuncturePosition"`
	Notes             string   `json:"notes"`
	Images            []string `json:"images"`
	IsDefault         bool     `json:"isDefault"`
	IsDisabled        bool     `json:"isDisabled"`
	CreatedAt         string   `json:"createdAt"`
}

type VascularAccessRequest struct {
	AccessType        string   `json:"accessType" binding:"required"`
	Site              string   `json:"site"`
	Artery            []string `json:"artery"`
	Vein              []string `json:"vein"`
	Side              string   `json:"side"`
	Hospital          string   `json:"hospital"`
	Surgeon           string   `json:"surgeon"`
	SurgeryDate       string   `json:"surgeryDate"`
	FirstUseDate      string   `json:"firstUseDate"`
	AccessNumber      int      `json:"accessNumber"`
	InterventionCount int      `json:"interventionCount"`
	InterventionDate  string   `json:"interventionDate"`
	CatheterMethod    *string  `json:"catheterMethod"`
	CatheterDepth     *string  `json:"catheterDepth"`
	VPuncturePosition []string `json:"vPuncturePosition"`
	APuncturePosition []string `json:"aPuncturePosition"`
	Notes             string   `json:"notes"`
	Images            []string `json:"images"`
	IsDefault         bool     `json:"isDefault"`
	IsDisabled        bool     `json:"isDisabled"`
}

func (s *VascularAccessService) List(patientID string) ([]VascularAccessResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}

	var accesses []models.VascularAccess
	err = s.db.Where(`"PatientId" = ? AND "TenantId" = ?`, legacyPatientID, LegacyTenantID).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "IsDefault"}, Desc: true}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "CreateTime"}, Desc: true}).
		Find(&accesses).Error
	if err != nil {
		return nil, err
	}

	imageMap, err := s.loadImagesByAccessIDs(accesses)
	if err != nil {
		return nil, err
	}

	result := make([]VascularAccessResponse, 0, len(accesses))
	for _, access := range accesses {
		result = append(result, s.buildResponse(vascularAccessWithImages{
			VascularAccess: access,
			Images:         imageMap[access.ID.Int64()],
		}))
	}
	return result, nil
}

func (s *VascularAccessService) Create(patientID string, req *VascularAccessRequest, creatorID int64) (*VascularAccessResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}

	var count int64
	if err := s.db.Model(&models.Patient{}).
		Where(`"Id" = ? AND "TenantId" = ?`, legacyPatientID, LegacyTenantID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("patient not found")
	}

	accessID, err := nextLegacyID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	access := models.VascularAccess{
		ID:                accessID,
		TenantID:          LegacyTenantID,
		PatientID:         legacyPatientID,
		AccessType:        req.AccessType,
		Site:              req.Site,
		Artery:            joinStringList(req.Artery),
		Vein:              joinStringList(req.Vein),
		Side:              req.Side,
		Hospital:          req.Hospital,
		Surgeon:           req.Surgeon,
		AccessNumber:      int64(req.AccessNumber),
		InterventionCount: int64(req.InterventionCount),
		CatheterMethod:    strDeref(req.CatheterMethod),
		CatheterDepth:     parseFloat(strDeref(req.CatheterDepth)),
		VPuncturePosition: joinStringList(req.VPuncturePosition),
		APuncturePosition: joinStringList(req.APuncturePosition),
		Notes:             req.Notes,
		IsDefault:         req.IsDefault,
		IsDisabled:        req.IsDisabled,
		CreatorID:         creatorID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	access.SurgeryDate = parseDateStringPtr(req.SurgeryDate)
	access.FirstUseDate = parseDateStringPtr(req.FirstUseDate)
	access.InterventionDate = parseDateStringPtr(req.InterventionDate)

	var imagePayload []string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if req.IsDefault {
			if err := tx.Model(&models.VascularAccess{}).
				Where(`"PatientId" = ? AND "TenantId" = ? AND "IsDefault" = ?`, legacyPatientID, LegacyTenantID, true).
				Update("IsDefault", false).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&access).Error; err != nil {
			return err
		}

		images, pictureID, err := s.replaceAccessImages(tx, access.ID, creatorID, req.Images)
		if err != nil {
			return err
		}
		imagePayload = images
		access.PictureID = pictureID

		return tx.Model(&models.VascularAccess{}).
			Where(`"Id" = ? AND "TenantId" = ?`, access.ID, LegacyTenantID).
			Updates(map[string]any{
				"LastModifyTime": now,
			}).Error
	})
	if err != nil {
		return nil, err
	}

	resp := s.buildResponse(vascularAccessWithImages{VascularAccess: access, Images: imagePayload})
	return &resp, nil
}

func (s *VascularAccessService) Update(patientID, accessID string, req *VascularAccessRequest, creatorID int64) (*VascularAccessResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}
	legacyAccessID, err := parseLegacyID(accessID)
	if err != nil {
		return nil, errors.New("invalid vascular access id")
	}

	var access models.VascularAccess
	err = s.db.Where(`"Id" = ? AND "PatientId" = ? AND "TenantId" = ?`, legacyAccessID, legacyPatientID, LegacyTenantID).
		First(&access).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("vascular access not found")
	}
	if err != nil {
		return nil, err
	}

	access.AccessType = req.AccessType
	access.Site = req.Site
	access.Artery = joinStringList(req.Artery)
	access.Vein = joinStringList(req.Vein)
	access.Side = req.Side
	access.Hospital = req.Hospital
	access.Surgeon = req.Surgeon
	access.AccessNumber = int64(req.AccessNumber)
	access.InterventionCount = int64(req.InterventionCount)
	access.CatheterMethod = strDeref(req.CatheterMethod)
	access.CatheterDepth = parseFloat(strDeref(req.CatheterDepth))
	access.VPuncturePosition = joinStringList(req.VPuncturePosition)
	access.APuncturePosition = joinStringList(req.APuncturePosition)
	access.Notes = req.Notes
	access.IsDefault = req.IsDefault
	access.IsDisabled = req.IsDisabled
	access.SurgeryDate = parseDateStringPtr(req.SurgeryDate)
	access.FirstUseDate = parseDateStringPtr(req.FirstUseDate)
	access.InterventionDate = parseDateStringPtr(req.InterventionDate)
	access.UpdatedAt = time.Now()
	if access.CreatorID == 0 && creatorID > 0 {
		access.CreatorID = creatorID
	}

	var imagePayload []string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if req.IsDefault {
			if err := tx.Model(&models.VascularAccess{}).
				Where(`"PatientId" = ? AND "TenantId" = ? AND "Id" <> ? AND "IsDefault" = ?`, legacyPatientID, LegacyTenantID, legacyAccessID, true).
				Update("IsDefault", false).Error; err != nil {
				return err
			}
		}

		images, pictureID, err := s.replaceAccessImages(tx, access.ID, creatorID, req.Images)
		if err != nil {
			return err
		}
		imagePayload = images
		access.PictureID = pictureID

		return tx.Model(&models.VascularAccess{}).
			Where(`"Id" = ? AND "PatientId" = ? AND "TenantId" = ?`, legacyAccessID, legacyPatientID, LegacyTenantID).
			Updates(map[string]any{
				"AccessType":        access.AccessType,
				"AccessPosition":    access.Site,
				"Artery":            access.Artery,
				"Venous":            access.Vein,
				"LeftAndRight":      access.Side,
				"OperationHospital": access.Hospital,
				"OperationDr":       access.Surgeon,
				"OperationTime":     access.SurgeryDate,
				"FirstUseTime":      access.FirstUseDate,
				"AccessCount":       access.AccessNumber,
				"InterveneCount":    access.InterventionCount,
				"InterveneTime":     access.InterventionDate,
				"CatheterizeMethod": access.CatheterMethod,
				"CatheterDepth":     access.CatheterDepth,
				"VSidePointCount":   access.VPuncturePosition,
				"ASidePointCount":   access.APuncturePosition,
				"Note":              access.Notes,
				"IsDefault":         access.IsDefault,
				"IsDisabled":        access.IsDisabled,
				"LastModifyTime":    access.UpdatedAt,
			}).Error
	})
	if err != nil {
		return nil, err
	}

	resp := s.buildResponse(vascularAccessWithImages{VascularAccess: access, Images: imagePayload})
	return &resp, nil
}

func (s *VascularAccessService) Delete(patientID, accessID string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return errors.New("invalid patient id")
	}
	legacyAccessID, err := parseLegacyID(accessID)
	if err != nil {
		return errors.New("invalid vascular access id")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(`"VascularAccessId" = ? AND "TenantId" = ?`, legacyAccessID, LegacyTenantID).
			Delete(&models.VascularAccessImage{}).Error; err != nil {
			return err
		}
		result := tx.Where(`"Id" = ? AND "PatientId" = ? AND "TenantId" = ?`, legacyAccessID, legacyPatientID, LegacyTenantID).
			Delete(&models.VascularAccess{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("vascular access not found")
		}
		return nil
	})
}

func parseDateStringPtr(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return &t
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t
	}
	return nil
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		result = append(result, value)
	}
	return result
}

func joinStringList(values []string) string {
	return strings.Join(compactStrings(values), ",")
}

func splitToSlice(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}

	if strings.HasPrefix(raw, "[") {
		var values []string
		if err := json.Unmarshal([]byte(raw), &values); err == nil {
			return compactStrings(values)
		}
	}

	return compactStrings(strings.Split(raw, ","))
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *VascularAccessService) buildResponse(access vascularAccessWithImages) VascularAccessResponse {
	depthStr := ""
	if access.CatheterDepth != 0 {
		depthStr = strconv.FormatFloat(access.CatheterDepth, 'f', -1, 64)
	}

	resp := VascularAccessResponse{
		ID:                legacyIDString(access.ID),
		AccessType:        access.AccessType,
		Site:              access.Site,
		Artery:            splitToSlice(access.Artery),
		Vein:              splitToSlice(access.Vein),
		Side:              access.Side,
		Hospital:          access.Hospital,
		Surgeon:           access.Surgeon,
		AccessNumber:      int(access.AccessNumber),
		InterventionCount: int(access.InterventionCount),
		CatheterMethod:    strPtr(access.CatheterMethod),
		CatheterDepth:     strPtr(depthStr),
		VPuncturePosition: splitToSlice(access.VPuncturePosition),
		APuncturePosition: splitToSlice(access.APuncturePosition),
		Notes:             access.Notes,
		Images:            append([]string{}, access.Images...),
		IsDefault:         access.IsDefault,
		IsDisabled:        access.IsDisabled,
		CreatedAt:         access.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if access.SurgeryDate != nil {
		resp.SurgeryDate = access.SurgeryDate.Format("2006-01-02")
	}
	if access.FirstUseDate != nil {
		resp.FirstUseDate = access.FirstUseDate.Format("2006-01-02")
	}
	if access.InterventionDate != nil {
		resp.InterventionDate = access.InterventionDate.Format("2006-01-02")
	}

	return resp
}

func (s *VascularAccessService) loadImagesByAccessIDs(accesses []models.VascularAccess) (map[int64][]string, error) {
	result := make(map[int64][]string, len(accesses))
	if len(accesses) == 0 {
		return result, nil
	}

	ids := make([]modeltypes.LegacyID, 0, len(accesses))
	for _, access := range accesses {
		ids = append(ids, access.ID)
	}

	var images []models.VascularAccessImage
	if err := s.db.Where(`"TenantId" = ? AND "VascularAccessId" IN ?`, LegacyTenantID, ids).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "Sort"}, Desc: false}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "CreateTime"}, Desc: false}).
		Find(&images).Error; err != nil {
		return nil, err
	}

	for _, image := range images {
		payload := strings.TrimSpace(image.ImageBase64String)
		if payload == "" {
			continue
		}
		result[image.VascularAccessID.Int64()] = append(result[image.VascularAccessID.Int64()], payload)
	}

	return result, nil
}

func (s *VascularAccessService) replaceAccessImages(tx *gorm.DB, accessID modeltypes.LegacyID, creatorID int64, images []string) ([]string, *modeltypes.LegacyID, error) {
	cleaned := compactStrings(images)
	if err := tx.Where(`"VascularAccessId" = ? AND "TenantId" = ?`, accessID, LegacyTenantID).
		Delete(&models.VascularAccessImage{}).Error; err != nil {
		return nil, nil, err
	}
	if len(cleaned) == 0 {
		return []string{}, nil, nil
	}

	now := time.Now()
	rows := make([]models.VascularAccessImage, 0, len(cleaned))
	var pictureID *modeltypes.LegacyID
	for idx, payload := range cleaned {
		imageID, err := nextLegacyID()
		if err != nil {
			return nil, nil, err
		}
		if pictureID == nil {
			pictureID = new(modeltypes.LegacyID)
			*pictureID = imageID
		}
		rows = append(rows, models.VascularAccessImage{
			ID:                imageID,
			TenantID:          LegacyTenantID,
			VascularAccessID:  accessID,
			ImageName:         fmt.Sprintf("vascular-access-%s-%d", legacyIDString(accessID), idx+1),
			ImageBase64String: payload,
			Sort:              idx + 1,
			CreatorID:         creatorID,
			CreatedAt:         now,
			UpdatedAt:         now,
		})
	}

	if err := tx.Create(&rows).Error; err != nil {
		return nil, nil, err
	}

	return cleaned, pictureID, nil
}

type VascularAccessInterventionResponse struct {
	ID                 string `json:"id"`
	VascularAccessID   string `json:"vascularAccessId"`
	PatientID          string `json:"patientId"`
	AccessType         string `json:"accessType"`
	AvgBloodFlow       int    `json:"avgBloodFlow"`
	UsageDays          int    `json:"usageDays"`
	SurgeryType        string `json:"surgeryType"`
	InterventionReason string `json:"interventionReason"`
	Doctor             string `json:"doctor"`
	InterventionDate   string `json:"interventionDate"`
	Description        string `json:"description"`
	CreatedAt          string `json:"createdAt"`
}

type VascularAccessInterventionRequest struct {
	VascularAccessID   string `json:"vascularAccessId" binding:"required"`
	AccessType         string `json:"accessType"`
	AvgBloodFlow       int    `json:"avgBloodFlow"`
	UsageDays          int    `json:"usageDays"`
	SurgeryType        string `json:"surgeryType" binding:"required"`
	InterventionReason string `json:"interventionReason" binding:"required"`
	Doctor             string `json:"doctor"`
	InterventionDate   string `json:"interventionDate" binding:"required"`
	Description        string `json:"description"`
}

func (s *VascularAccessService) ListInterventions(patientID, vascularAccessID string) ([]VascularAccessInterventionResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}

	query := s.db.Where(`"PatientId" = ? AND "TenantId" = ?`, legacyPatientID, LegacyTenantID)
	if vascularAccessID != "" {
		legacyVascularAccessID, parseErr := parseLegacyID(vascularAccessID)
		if parseErr != nil {
			return nil, errors.New("invalid vascular access id")
		}
		query = query.Where(`"VascularAccessId" = ?`, legacyVascularAccessID)
	}

	var interventions []models.VascularAccessIntervention
	err = query.Order(clause.OrderByColumn{Column: clause.Column{Name: "ChangeTime"}, Desc: true}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "CreateTime"}, Desc: true}).
		Find(&interventions).Error
	if err != nil {
		return nil, err
	}

	result := make([]VascularAccessInterventionResponse, 0, len(interventions))
	for _, intervention := range interventions {
		result = append(result, s.buildInterventionResponse(intervention))
	}
	return result, nil
}

func (s *VascularAccessService) CreateIntervention(patientID string, req *VascularAccessInterventionRequest) (*VascularAccessInterventionResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}
	legacyVascularAccessID, err := parseLegacyID(req.VascularAccessID)
	if err != nil {
		return nil, errors.New("invalid vascular access id")
	}

	var access models.VascularAccess
	err = s.db.Where(`"Id" = ? AND "PatientId" = ? AND "TenantId" = ?`, legacyVascularAccessID, legacyPatientID, LegacyTenantID).
		First(&access).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("vascular access not found")
	}
	if err != nil {
		return nil, err
	}

	interventionDate, err := time.Parse("2006-01-02", req.InterventionDate)
	if err != nil {
		return nil, errors.New("invalid intervention date format")
	}

	interventionID, err := nextLegacyID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	intervention := models.VascularAccessIntervention{
		ID:                 interventionID,
		TenantID:           LegacyTenantID,
		VascularAccessID:   legacyVascularAccessID,
		PatientID:          legacyPatientID,
		AccessType:         req.AccessType,
		AvgBloodFlow:       float64(req.AvgBloodFlow),
		UsageDays:          req.UsageDays,
		SurgeryType:        req.SurgeryType,
		InterventionReason: req.InterventionReason,
		Doctor:             req.Doctor,
		InterventionDate:   interventionDate,
		Description:        req.Description,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if intervention.AccessType == "" {
		intervention.AccessType = access.AccessType
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&intervention).Error; err != nil {
			return err
		}
		return tx.Model(&models.VascularAccess{}).
			Where(`"Id" = ? AND "PatientId" = ? AND "TenantId" = ?`, legacyVascularAccessID, legacyPatientID, LegacyTenantID).
			Updates(map[string]any{
				"InterveneCount": access.InterventionCount + 1,
				"InterveneTime":  interventionDate,
				"LastModifyTime": now,
			}).Error
	})
	if err != nil {
		return nil, err
	}

	access.InterventionCount++
	access.InterventionDate = &interventionDate

	resp := s.buildInterventionResponse(intervention)
	return &resp, nil
}

func (s *VascularAccessService) DeleteIntervention(patientID, interventionID string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return errors.New("invalid patient id")
	}
	legacyInterventionID, err := parseLegacyID(interventionID)
	if err != nil {
		return errors.New("invalid intervention id")
	}

	result := s.db.Where(`"Id" = ? AND "PatientId" = ? AND "TenantId" = ?`, legacyInterventionID, legacyPatientID, LegacyTenantID).
		Delete(&models.VascularAccessIntervention{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("intervention not found")
	}
	return nil
}

func (s *VascularAccessService) buildInterventionResponse(iv models.VascularAccessIntervention) VascularAccessInterventionResponse {
	return VascularAccessInterventionResponse{
		ID:                 legacyIDString(iv.ID),
		VascularAccessID:   legacyIDString(iv.VascularAccessID),
		PatientID:          legacyIDString(iv.PatientID),
		AccessType:         iv.AccessType,
		AvgBloodFlow:       int(iv.AvgBloodFlow),
		UsageDays:          iv.UsageDays,
		SurgeryType:        iv.SurgeryType,
		InterventionReason: iv.InterventionReason,
		Doctor:             iv.Doctor,
		InterventionDate:   iv.InterventionDate.Format("2006-01-02"),
		Description:        iv.Description,
		CreatedAt:          iv.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
