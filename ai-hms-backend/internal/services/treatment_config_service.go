package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// ===== 方案模板服务 =====

// PlanTemplateService 方案模板服务
type PlanTemplateService struct {
	db *gorm.DB
}

// NewPlanTemplateService 创建方案模板服务
func NewPlanTemplateService() *PlanTemplateService {
	return &PlanTemplateService{
		db: database.GetDB(),
	}
}

// PlanTemplateListRequest 获取方案模板列表请求
type PlanTemplateListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	Search    string `form:"search"`
	Mode      string `form:"mode"`      // HD, HDF, HP, HF, HFD
	Category  string `form:"category"`  // 分类
	IsEnabled *bool  `form:"isEnabled"` // 启用状态
}

// PlanTemplateListResponse 获取方案模板列表响应
type PlanTemplateListResponse struct {
	Items     []models.PlanTemplate `json:"items"`
	Total     int64                `json:"total"`
	Page      int                  `json:"page"`
	PageSize  int                  `json:"pageSize"`
	TotalPage int                  `json:"totalPage"`
}

// List 获取方案模板列表
func (s *PlanTemplateService) List(req PlanTemplateListRequest) (*PlanTemplateListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.PlanTemplate{})

	// 筛选条件
	if req.Search != "" {
		query = query.Where("name LIKE ?", "%"+req.Search+"%")
	}
	if req.Mode != "" {
		query = query.Where("mode = ?", req.Mode)
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *req.IsEnabled)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var items []models.PlanTemplate
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Offset(offset).
		Limit(req.PageSize).
		Order("is_default DESC, created_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &PlanTemplateListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// Get 获取方案模板详情
func (s *PlanTemplateService) Get(id string) (*models.PlanTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var template models.PlanTemplate
	err := s.db.First(&template, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan template not found")
		}
		return nil, err
	}

	return &template, nil
}

// PlanTemplateCreateRequest 创建方案模板请求
type PlanTemplateCreateRequest struct {
	Name        string                         `json:"name" binding:"required"`
	Description string                         `json:"description"`
	Mode        string                         `json:"mode" binding:"required,oneof=HD HFD HP HF HDF"`
	Category    string                         `json:"category"`
	IsDefault   bool                           `json:"isDefault"`
	IsEnabled   bool                           `json:"isEnabled"`
	TemplateContent models.PlanTemplateContent `json:"templateContent"`
}

// Create 创建方案模板
func (s *PlanTemplateService) Create(req PlanTemplateCreateRequest) (*models.PlanTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 如果设置为默认模板，先取消其他默认模板
	if req.IsDefault {
		s.db.Model(&models.PlanTemplate{}).
			Where("mode = ? AND is_default = ?", req.Mode, true).
			Update("is_default", false)
	}

	template := models.PlanTemplate{
		ID:              utils.GenerateID(),
		Name:            req.Name,
		Description:     req.Description,
		Mode:            req.Mode,
		Category:        req.Category,
		IsDefault:       req.IsDefault,
		IsEnabled:       req.IsEnabled,
		TemplateContent: req.TemplateContent,
	}

	if err := s.db.Create(&template).Error; err != nil {
		return nil, err
	}

	return &template, nil
}

// PlanTemplateUpdateRequest 更新方案模板请求
type PlanTemplateUpdateRequest struct {
	Name           *string                         `json:"name"`
	Description    *string                         `json:"description"`
	Mode           *string                         `json:"mode"`
	Category       *string                         `json:"category"`
	IsDefault      *bool                           `json:"isDefault"`
	IsEnabled      *bool                           `json:"isEnabled"`
	TemplateContent *models.PlanTemplateContent    `json:"templateContent"`
}

// Update 更新方案模板
func (s *PlanTemplateService) Update(id string, req PlanTemplateUpdateRequest) (*models.PlanTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var template models.PlanTemplate
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan template not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Mode != nil {
		updates["mode"] = *req.Mode
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.IsDefault != nil {
		// 如果设置为默认模板，先取消同模式的其他默认模板
		if *req.IsDefault {
			s.db.Model(&models.PlanTemplate{}).
				Where("mode = ? AND is_default = ? AND id != ?", template.Mode, true, id).
				Update("is_default", false)
		}
		updates["is_default"] = *req.IsDefault
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.TemplateContent != nil {
		updates["template_content"] = *req.TemplateContent
	}

	if err := s.db.Model(&template).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的数据
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &template, nil
}

// Delete 删除方案模板
func (s *PlanTemplateService) Delete(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Delete(&models.PlanTemplate{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("plan template not found")
	}

	return nil
}

// ToggleEnabled 切换启用状态
func (s *PlanTemplateService) ToggleEnabled(id string) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	var template models.PlanTemplate
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("plan template not found")
		}
		return false, err
	}

	newStatus := !template.IsEnabled
	if err := s.db.Model(&template).Update("is_enabled", newStatus).Error; err != nil {
		return false, err
	}

	return newStatus, nil
}

// SetDefault 设置默认模板
func (s *PlanTemplateService) SetDefault(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	var template models.PlanTemplate
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("plan template not found")
		}
		return err
	}

	// 取消同模式的其他默认模板
	s.db.Model(&models.PlanTemplate{}).
		Where("mode = ? AND is_default = ?", template.Mode, true).
		Update("is_default", false)

	// 设置当前模板为默认
	if err := s.db.Model(&template).Update("is_default", true).Error; err != nil {
		return err
	}

	return nil
}

// ===== 材料目录服务 =====

// MaterialCatalogService 材料目录服务
type MaterialCatalogService struct {
	db *gorm.DB
}

// NewMaterialCatalogService 创建材料目录服务
func NewMaterialCatalogService() *MaterialCatalogService {
	return &MaterialCatalogService{
		db: database.GetDB(),
	}
}

// MaterialCatalogListRequest 获取材料目录列表请求
type MaterialCatalogListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	Search    string `form:"search"`
	Category  string `form:"category"`
	IsEnabled *bool  `form:"isEnabled"`
}

// MaterialCatalogListResponse 获取材料目录列表响应
type MaterialCatalogListResponse struct {
	Items     []models.MaterialCatalog `json:"items"`
	Total     int64                    `json:"total"`
	Page      int                      `json:"page"`
	PageSize  int                      `json:"pageSize"`
	TotalPage int                      `json:"totalPage"`
}

// List 获取材料目录列表
func (s *MaterialCatalogService) List(req MaterialCatalogListRequest) (*MaterialCatalogListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.MaterialCatalog{})

	// 筛选条件
	if req.Search != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *req.IsEnabled)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var items []models.MaterialCatalog
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Offset(offset).
		Limit(req.PageSize).
		Order("sort_order ASC, category, code, created_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &MaterialCatalogListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// Get 获取材料目录详情
func (s *MaterialCatalogService) Get(id string) (*models.MaterialCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var catalog models.MaterialCatalog
	err := s.db.First(&catalog, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("material catalog not found")
		}
		return nil, err
	}

	return &catalog, nil
}

// MaterialCatalogCreateRequest 创建材料目录请求
type MaterialCatalogCreateRequest struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	ShortName    string `json:"shortName"`
	Mnemonic     string `json:"mnemonic"`
	Category     string `json:"category"`
	Spec         string `json:"spec"`
	StandardType string `json:"standardType"`
	Brand        string `json:"brand"`
	Unit         string `json:"unit"`
	Packaging    string `json:"packaging"`
	Manufacturer string `json:"manufacturer"`
	SortOrder    int    `json:"sortOrder"`
	IsEnabled    bool   `json:"isEnabled"`
	Notes        string `json:"notes"`
}

// ValidateMaterialCatalogCreateRequest 校验创建材料目录请求
func ValidateMaterialCatalogCreateRequest(req MaterialCatalogCreateRequest) error {
	missing := make([]string, 0, 4)
	if strings.TrimSpace(req.Category) == "" {
		missing = append(missing, "材料分类")
	}
	if strings.TrimSpace(req.Name) == "" {
		missing = append(missing, "材料名称")
	}
	if strings.TrimSpace(req.ShortName) == "" {
		missing = append(missing, "简称")
	}
	if strings.TrimSpace(req.Unit) == "" {
		missing = append(missing, "单位")
	}
	if len(missing) > 0 {
		return fmt.Errorf("缺少必填字段：%s", strings.Join(missing, "、"))
	}
	return nil
}

func ensureMaterialCatalogCode(code string) string {
	trimmed := strings.TrimSpace(code)
	if trimmed != "" {
		return trimmed
	}
	return "AUTO-" + uuid.NewString()
}

func optionalStringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (s *MaterialCatalogService) resolveMaterialCategory(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || s.db == nil {
		return trimmed
	}

	var dictItem models.DictItem
	err := s.db.
		Where("type_code = ? AND is_enabled = ? AND code = ?", models.DictTypeMaterialCat, true, trimmed).
		Order("sort_order ASC").
		First(&dictItem).Error
	if err == nil {
		return strings.TrimSpace(dictItem.Name)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return trimmed
	}

	err = s.db.
		Where("type_code = ? AND is_enabled = ? AND name = ?", models.DictTypeMaterialCat, true, trimmed).
		Order("sort_order ASC").
		First(&dictItem).Error
	if err == nil {
		return strings.TrimSpace(dictItem.Name)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return trimmed
	}

	// 兼容历史停用字典项
	err = s.db.
		Where("type_code = ? AND code = ?", models.DictTypeMaterialCat, trimmed).
		Order("sort_order ASC").
		First(&dictItem).Error
	if err == nil {
		return strings.TrimSpace(dictItem.Name)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return trimmed
	}

	err = s.db.
		Where("type_code = ? AND name = ?", models.DictTypeMaterialCat, trimmed).
		Order("sort_order ASC").
		First(&dictItem).Error
	if err == nil {
		return strings.TrimSpace(dictItem.Name)
	}

	return trimmed
}

func mapMaterialCatalogWriteError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "material_catalogs_code") {
			return errors.New("material code already exists")
		}
		return errors.New("material catalog already exists")
	}

	return err
}

// Create 创建材料目录
func (s *MaterialCatalogService) Create(req MaterialCatalogCreateRequest) (*models.MaterialCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	code := ensureMaterialCatalogCode(req.Code)

	// 检查代码是否已存在
	var count int64
	if err := s.db.Model(&models.MaterialCatalog{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("material code already exists")
	}

	catalog := models.MaterialCatalog{
		Code:         code,
		Name:         strings.TrimSpace(req.Name),
		ShortName:    optionalStringPtr(req.ShortName),
		Mnemonic:     optionalStringPtr(req.Mnemonic),
		Category:     s.resolveMaterialCategory(req.Category),
		Spec:         strings.TrimSpace(req.Spec),
		StandardType: optionalStringPtr(req.StandardType),
		Brand:        strings.TrimSpace(req.Brand),
		Unit:         strings.TrimSpace(req.Unit),
		Packaging:    optionalStringPtr(req.Packaging),
		Manufacturer: optionalStringPtr(req.Manufacturer),
		SortOrder:    req.SortOrder,
		IsEnabled:    req.IsEnabled,
		Notes:        strings.TrimSpace(req.Notes),
	}

	if err := s.db.Create(&catalog).Error; err != nil {
		return nil, mapMaterialCatalogWriteError(err)
	}

	return &catalog, nil
}

// MaterialCatalogUpdateRequest 更新材料目录请求
type MaterialCatalogUpdateRequest struct {
	Code         *string `json:"code"`
	Name         *string `json:"name"`
	ShortName    *string `json:"shortName"`
	Mnemonic     *string `json:"mnemonic"`
	Category     *string `json:"category"`
	Spec         *string `json:"spec"`
	StandardType *string `json:"standardType"`
	Brand        *string `json:"brand"`
	Unit         *string `json:"unit"`
	Packaging    *string `json:"packaging"`
	Manufacturer *string `json:"manufacturer"`
	SortOrder    *int    `json:"sortOrder"`
	IsEnabled    *bool   `json:"isEnabled"`
	Notes        *string `json:"notes"`
}

// Update 更新材料目录
func (s *MaterialCatalogService) Update(id uint, req MaterialCatalogUpdateRequest) (*models.MaterialCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var catalog models.MaterialCatalog
	if err := s.db.First(&catalog, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("material catalog not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Code != nil {
		code := strings.TrimSpace(*req.Code)
		if code == "" {
			return nil, errors.New("material code cannot be empty")
		}

		var duplicate int64
		if err := s.db.Model(&models.MaterialCatalog{}).
			Where("code = ? AND id <> ?", code, id).
			Count(&duplicate).Error; err != nil {
			return nil, err
		}
		if duplicate > 0 {
			return nil, errors.New("material code already exists")
		}
		updates["code"] = code
	}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.ShortName != nil {
		updates["short_name"] = strings.TrimSpace(*req.ShortName)
	}
	if req.Mnemonic != nil {
		updates["mnemonic"] = strings.TrimSpace(*req.Mnemonic)
	}
	if req.Category != nil {
		updates["category"] = s.resolveMaterialCategory(*req.Category)
	}
	if req.Spec != nil {
		updates["spec"] = strings.TrimSpace(*req.Spec)
	}
	if req.StandardType != nil {
		updates["standard_type"] = strings.TrimSpace(*req.StandardType)
	}
	if req.Brand != nil {
		updates["brand"] = strings.TrimSpace(*req.Brand)
	}
	if req.Unit != nil {
		updates["unit"] = strings.TrimSpace(*req.Unit)
	}
	if req.Packaging != nil {
		updates["packaging"] = strings.TrimSpace(*req.Packaging)
	}
	if req.Manufacturer != nil {
		updates["manufacturer"] = strings.TrimSpace(*req.Manufacturer)
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.Notes != nil {
		updates["notes"] = strings.TrimSpace(*req.Notes)
	}

	if err := s.db.Model(&catalog).Updates(updates).Error; err != nil {
		return nil, mapMaterialCatalogWriteError(err)
	}

	// 重新获取更新后的数据
	if err := s.db.First(&catalog, id).Error; err != nil {
		return nil, err
	}

	return &catalog, nil
}

// Delete 删除材料目录
func (s *MaterialCatalogService) Delete(id uint) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Delete(&models.MaterialCatalog{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("material catalog not found")
	}

	return nil
}

// ToggleEnabled 切换启用状态
func (s *MaterialCatalogService) ToggleEnabled(id uint) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	var catalog models.MaterialCatalog
	if err := s.db.First(&catalog, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("material catalog not found")
		}
		return false, err
	}

	newStatus := !catalog.IsEnabled
	if err := s.db.Model(&catalog).Update("is_enabled", newStatus).Error; err != nil {
		return false, err
	}

	return newStatus, nil
}

// GetCategories 获取材料分类列表
func (s *MaterialCatalogService) GetCategories() ([]string, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var categories []string
	if err := s.db.Model(&models.MaterialCatalog{}).
		Distinct("category").
		Pluck("category", &categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// ===== 药品目录服务 =====

// DrugCatalogService 药品目录服务
type DrugCatalogService struct {
	db *gorm.DB
}

// NewDrugCatalogService 创建药品目录服务
func NewDrugCatalogService() *DrugCatalogService {
	return &DrugCatalogService{
		db: database.GetDB(),
	}
}

// DrugCatalogListRequest 获取药品目录列表请求
type DrugCatalogListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	Search    string `form:"search"`
	Category  string `form:"category"`
	IsEnabled *bool  `form:"isEnabled"`
}

// DrugCatalogListResponse 获取药品目录列表响应
type DrugCatalogListResponse struct {
	Items     []models.DrugCatalog `json:"items"`
	Total     int64                `json:"total"`
	Page      int                  `json:"page"`
	PageSize  int                  `json:"pageSize"`
	TotalPage int                  `json:"totalPage"`
}

// List 获取药品目录列表
func (s *DrugCatalogService) List(req DrugCatalogListRequest) (*DrugCatalogListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.DrugCatalog{})

	// 筛选条件
	if req.Search != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR generic_name LIKE ?",
			"%"+req.Search+"%", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *req.IsEnabled)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var items []models.DrugCatalog
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Offset(offset).
		Limit(req.PageSize).
		Order("category, code, created_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &DrugCatalogListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// Get 获取药品目录详情
func (s *DrugCatalogService) Get(id string) (*models.DrugCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var catalog models.DrugCatalog
	err := s.db.First(&catalog, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("drug catalog not found")
		}
		return nil, err
	}

	return &catalog, nil
}

// DrugCatalogCreateRequest 创建药品目录请求
type DrugCatalogCreateRequest struct {
	Code         string `json:"code"`
	Name         string `json:"name" binding:"required"`
	ShortName    string `json:"shortName"`
	Mnemonic     string `json:"mnemonic"`
	GenericName  string `json:"genericName"`
	Category     string `json:"category" binding:"required"`
	Spec         string `json:"spec"`
	Concentration string `json:"concentration"`
	SpecUnit     string `json:"specUnit"`
	MinUnitDose  string `json:"minUnitDose"`
	BaseUnit     string `json:"baseUnit"`
	Brand        string `json:"brand"`
	Packaging    string `json:"packaging"`
	Manufacturer string `json:"manufacturer"`
	StandardType string `json:"standardType"`
	Timing       string `json:"timing"`
	Tips         string `json:"tips"`
	SortOrder    int    `json:"sortOrder"`
	IsEnabled    bool   `json:"isEnabled"`
	Note         string `json:"note"`
}

func ensureDrugCatalogCode(code string) string {
	trimmed := strings.TrimSpace(code)
	if trimmed != "" {
		return trimmed
	}
	return "AUTO-" + uuid.NewString()
}

// Create 创建药品目录
func (s *DrugCatalogService) Create(req DrugCatalogCreateRequest) (*models.DrugCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	code := ensureDrugCatalogCode(req.Code)

	// 检查代码是否已存在
	var count int64
	if err := s.db.Model(&models.DrugCatalog{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("drug code already exists")
	}

	catalog := models.DrugCatalog{
		Code:         code,
		Name:         strings.TrimSpace(req.Name),
		ShortName:    optionalStringPtr(req.ShortName),
		Mnemonic:     optionalStringPtr(req.Mnemonic),
		GenericName:  strings.TrimSpace(req.GenericName),
		Category:     strings.TrimSpace(req.Category),
		Spec:         strings.TrimSpace(req.Spec),
		Concentration: optionalStringPtr(req.Concentration),
		SpecUnit:     optionalStringPtr(req.SpecUnit),
		MinUnitDose:  optionalStringPtr(req.MinUnitDose),
		BaseUnit:     strings.TrimSpace(req.BaseUnit),
		Brand:        optionalStringPtr(req.Brand),
		Packaging:    optionalStringPtr(req.Packaging),
		Manufacturer: strings.TrimSpace(req.Manufacturer),
		StandardType: optionalStringPtr(req.StandardType),
		Timing:       optionalStringPtr(req.Timing),
		Tips:         optionalStringPtr(req.Tips),
		SortOrder:    req.SortOrder,
		IsEnabled:    req.IsEnabled,
		Note:         strings.TrimSpace(req.Note),
	}

	if err := s.db.Create(&catalog).Error; err != nil {
		return nil, err
	}

	return &catalog, nil
}

// DrugCatalogUpdateRequest 更新药品目录请求
type DrugCatalogUpdateRequest struct {
	Name         *string `json:"name"`
	ShortName    *string `json:"shortName"`
	Mnemonic     *string `json:"mnemonic"`
	GenericName  *string `json:"genericName"`
	Category     *string `json:"category"`
	Spec         *string `json:"spec"`
	Concentration *string `json:"concentration"`
	SpecUnit     *string `json:"specUnit"`
	MinUnitDose  *string `json:"minUnitDose"`
	BaseUnit     *string `json:"baseUnit"`
	Brand        *string `json:"brand"`
	Packaging    *string `json:"packaging"`
	Manufacturer *string `json:"manufacturer"`
	StandardType *string `json:"standardType"`
	Timing       *string `json:"timing"`
	Tips         *string `json:"tips"`
	SortOrder    *int    `json:"sortOrder"`
	IsEnabled    *bool   `json:"isEnabled"`
	Note         *string `json:"note"`
}

// Update 更新药品目录
func (s *DrugCatalogService) Update(id uint, req DrugCatalogUpdateRequest) (*models.DrugCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var catalog models.DrugCatalog
	if err := s.db.First(&catalog, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("drug catalog not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.ShortName != nil {
		updates["short_name"] = strings.TrimSpace(*req.ShortName)
	}
	if req.Mnemonic != nil {
		updates["mnemonic"] = strings.TrimSpace(*req.Mnemonic)
	}
	if req.GenericName != nil {
		updates["generic_name"] = strings.TrimSpace(*req.GenericName)
	}
	if req.Category != nil {
		updates["category"] = strings.TrimSpace(*req.Category)
	}
	if req.Spec != nil {
		updates["spec"] = strings.TrimSpace(*req.Spec)
	}
	if req.Concentration != nil {
		updates["concentration"] = strings.TrimSpace(*req.Concentration)
	}
	if req.SpecUnit != nil {
		updates["spec_unit"] = strings.TrimSpace(*req.SpecUnit)
	}
	if req.MinUnitDose != nil {
		updates["min_unit_dose"] = strings.TrimSpace(*req.MinUnitDose)
	}
	if req.BaseUnit != nil {
		updates["unit"] = strings.TrimSpace(*req.BaseUnit)
	}
	if req.Brand != nil {
		updates["brand"] = strings.TrimSpace(*req.Brand)
	}
	if req.Packaging != nil {
		updates["packaging"] = strings.TrimSpace(*req.Packaging)
	}
	if req.Manufacturer != nil {
		updates["manufacturer"] = strings.TrimSpace(*req.Manufacturer)
	}
	if req.StandardType != nil {
		updates["standard_type"] = strings.TrimSpace(*req.StandardType)
	}
	if req.Timing != nil {
		updates["timing"] = strings.TrimSpace(*req.Timing)
	}
	if req.Tips != nil {
		updates["tips"] = strings.TrimSpace(*req.Tips)
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.Note != nil {
		updates["notes"] = strings.TrimSpace(*req.Note)
	}

	if err := s.db.Model(&catalog).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的数据
	if err := s.db.First(&catalog, id).Error; err != nil {
		return nil, err
	}

	return &catalog, nil
}

// Delete 删除药品目录
func (s *DrugCatalogService) Delete(id uint) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Delete(&models.DrugCatalog{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("drug catalog not found")
	}

	return nil
}

// ToggleEnabled 切换启用状态
func (s *DrugCatalogService) ToggleEnabled(id uint) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	var catalog models.DrugCatalog
	if err := s.db.First(&catalog, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("drug catalog not found")
		}
		return false, err
	}

	newStatus := !catalog.IsEnabled
	if err := s.db.Model(&catalog).Update("is_enabled", newStatus).Error; err != nil {
		return false, err
	}

	return newStatus, nil
}

// GetCategories 获取药品分类列表
func (s *DrugCatalogService) GetCategories() ([]string, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var categories []string
	if err := s.db.Model(&models.DrugCatalog{}).
		Distinct("category").
		Pluck("category", &categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// ===== 医嘱模板服务 =====

// OrderTemplateService 医嘱模板服务
type OrderTemplateService struct {
	db *gorm.DB
}

// NewOrderTemplateService 创建医嘱模板服务
func NewOrderTemplateService() *OrderTemplateService {
	return &OrderTemplateService{
		db: database.GetDB(),
	}
}

// OrderTemplateListRequest 获取医嘱模板列表请求
type OrderTemplateListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	Search    string `form:"search"`
	Type      string `form:"type"`
	Category  string `form:"category"`
	IsEnabled *bool  `form:"isEnabled"`
}

// OrderTemplateListResponse 获取医嘱模板列表响应
type OrderTemplateListResponse struct {
	Items     []models.OrderTemplate `json:"items"`
	Total     int64                  `json:"total"`
	Page      int                    `json:"page"`
	PageSize  int                    `json:"pageSize"`
	TotalPage int                    `json:"totalPage"`
}

// List 获取医嘱模板列表
func (s *OrderTemplateService) List(req OrderTemplateListRequest) (*OrderTemplateListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.OrderTemplate{})

	// 筛选条件
	if req.Search != "" {
		query = query.Where("name LIKE ? OR content LIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *req.IsEnabled)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var items []models.OrderTemplate
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Offset(offset).
		Limit(req.PageSize).
		Order("is_default DESC, created_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &OrderTemplateListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// Get 获取医嘱模板详情（含 items）
func (s *OrderTemplateService) Get(id string) (*models.OrderTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var template models.OrderTemplate
	err := s.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).First(&template, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order template not found")
		}
		return nil, err
	}

	return &template, nil
}

// OrderTemplateItemRequest 医嘱模板条目请求
type OrderTemplateItemRequest struct {
	DrugID      *uint   `json:"drugId"`
	DrugName    string  `json:"drugName" binding:"required"`
	Spec        *string `json:"spec"`
	MinUnitDose *string `json:"minUnitDose"`
	Dosage      *string `json:"dosage"`
	Unit        *string `json:"unit"`
	Route       *string `json:"route"`
	Frequency   *string `json:"frequency"`
	Timing      *string `json:"timing"`
	GroupID     *string `json:"groupId"`
	SortOrder   int     `json:"sortOrder"`
}

// OrderTemplateCreateRequest 创建医嘱模板请求
type OrderTemplateCreateRequest struct {
	Name      string                     `json:"name" binding:"required"`
	Type      string                     `json:"type" binding:"required,oneof=长期 临时"`
	Category  string                     `json:"category" binding:"required"`
	Content   string                     `json:"content" binding:"required"`
	Frequency *string                    `json:"frequency"`
	Priority  string                     `json:"priority"`
	IsDefault bool                       `json:"isDefault"`
	IsEnabled bool                       `json:"isEnabled"`
	Items     []OrderTemplateItemRequest `json:"items"`
}

// Create 创建医嘱模板
func (s *OrderTemplateService) Create(req OrderTemplateCreateRequest) (*models.OrderTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 如果设置为默认模板，先取消其他默认模板
	if req.IsDefault {
		s.db.Model(&models.OrderTemplate{}).
			Where("category = ? AND is_default = ?", req.Category, true).
			Update("is_default", false)
	}

	// 设置默认优先级
	if req.Priority == "" {
		req.Priority = models.OrderPriorityNormal
	}

	template := models.OrderTemplate{
		ID:        utils.GenerateID(),
		Name:      req.Name,
		Type:      req.Type,
		Category:  req.Category,
		Content:   req.Content,
		Frequency: req.Frequency,
		Priority:  req.Priority,
		IsDefault: req.IsDefault,
		IsEnabled: req.IsEnabled,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&template).Error; err != nil {
			return err
		}
		// 批量创建 items
		for _, itemReq := range req.Items {
			item := models.OrderTemplateItem{
				TemplateID:  template.ID,
				DrugID:      itemReq.DrugID,
				DrugName:    itemReq.DrugName,
				Spec:        itemReq.Spec,
				MinUnitDose: itemReq.MinUnitDose,
				Dosage:      itemReq.Dosage,
				Unit:        itemReq.Unit,
				Route:       itemReq.Route,
				Frequency:   itemReq.Frequency,
				Timing:      itemReq.Timing,
				GroupID:     itemReq.GroupID,
				SortOrder:   itemReq.SortOrder,
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 重新加载含 items
	return s.Get(template.ID)
}

// OrderTemplateUpdateRequest 更新医嘱模板请求
type OrderTemplateUpdateRequest struct {
	Name      *string                     `json:"name"`
	Type      *string                     `json:"type"`
	Category  *string                     `json:"category"`
	Content   *string                     `json:"content"`
	Frequency *string                     `json:"frequency"`
	Priority  *string                     `json:"priority"`
	IsDefault *bool                       `json:"isDefault"`
	IsEnabled *bool                       `json:"isEnabled"`
	Items     *[]OrderTemplateItemRequest `json:"items"`
}

// Update 更新医嘱模板
func (s *OrderTemplateService) Update(id string, req OrderTemplateUpdateRequest) (*models.OrderTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var template models.OrderTemplate
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order template not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Frequency != nil {
		updates["frequency"] = *req.Frequency
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.IsDefault != nil {
		// 如果设置为默认模板，先取消同分类的其他默认模板
		if *req.IsDefault {
			s.db.Model(&models.OrderTemplate{}).
				Where("category = ? AND is_default = ? AND id != ?", template.Category, true, id).
				Update("is_default", false)
		}
		updates["is_default"] = *req.IsDefault
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 更新模板字段
		if len(updates) > 0 {
			if err := tx.Model(&template).Updates(updates).Error; err != nil {
				return err
			}
		}
		// 如果传了 items，全量替换
		if req.Items != nil {
			// 删除旧 items
			if err := tx.Where("template_id = ?", id).Delete(&models.OrderTemplateItem{}).Error; err != nil {
				return err
			}
			// 创建新 items
			for _, itemReq := range *req.Items {
				item := models.OrderTemplateItem{
					TemplateID:  id,
					DrugID:      itemReq.DrugID,
					DrugName:    itemReq.DrugName,
					Spec:        itemReq.Spec,
					MinUnitDose: itemReq.MinUnitDose,
					Dosage:      itemReq.Dosage,
					Unit:        itemReq.Unit,
					Route:       itemReq.Route,
					Frequency:   itemReq.Frequency,
					Timing:      itemReq.Timing,
					GroupID:     itemReq.GroupID,
					SortOrder:   itemReq.SortOrder,
				}
				if err := tx.Create(&item).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 重新加载含 items
	return s.Get(id)
}

// Delete 删除医嘱模板（级联删除 items）
func (s *OrderTemplateService) Delete(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 先删除关联的 items
		if err := tx.Where("template_id = ?", id).Delete(&models.OrderTemplateItem{}).Error; err != nil {
			return err
		}
		// 再删除模板
		result := tx.Delete(&models.OrderTemplate{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("order template not found")
		}
		return nil
	})
	return err
}

// ToggleEnabled 切换启用状态
func (s *OrderTemplateService) ToggleEnabled(id string) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	var template models.OrderTemplate
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("order template not found")
		}
		return false, err
	}

	newStatus := !template.IsEnabled
	if err := s.db.Model(&template).Update("is_enabled", newStatus).Error; err != nil {
		return false, err
	}

	return newStatus, nil
}

// SetDefault 设置默认模板
func (s *OrderTemplateService) SetDefault(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	var template models.OrderTemplate
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("order template not found")
		}
		return err
	}

	// 取消同分类的其他默认模板
	s.db.Model(&models.OrderTemplate{}).
		Where("category = ? AND is_default = ?", template.Category, true).
		Update("is_default", false)

	// 设置当前模板为默认
	if err := s.db.Model(&template).Update("is_default", true).Error; err != nil {
		return err
	}

	return nil
}
