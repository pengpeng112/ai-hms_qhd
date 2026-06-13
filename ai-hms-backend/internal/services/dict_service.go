package services

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DictService 字典服务
type DictService struct {
	db *gorm.DB
}

// NewDictService 创建字典服务
func NewDictService() *DictService {
	return &DictService{
		db: database.GetDB(),
	}
}

// DictTypeListResponse 字典类型列表响应
type DictTypeListResponse struct {
	Items []models.DictType `json:"items"`
	Total int64             `json:"total"`
}

// DictItemListResponse 字典项列表响应
type DictItemListResponse struct {
	Items []models.DictItem `json:"items"`
	Total int64             `json:"total"`
}

type legacyCodeDictionaryRow struct {
	Type       string `gorm:"column:Type"`
	Code       string `gorm:"column:Code"`
	Name       string `gorm:"column:Name"`
	Sort       int    `gorm:"column:Sort"`
	IsDisabled bool   `gorm:"column:IsDisabled"`
}

type legacyCodeDictionaryTypeAgg struct {
	Type string `gorm:"column:Type"`
	Sort int    `gorm:"column:Sort"`
	Cnt  int64  `gorm:"column:Cnt"`
}

type unifiedDictTypeMeta struct {
	Code string
	Name string
}

var legacyTypeToUnifiedCode = map[string]unifiedDictTypeMeta{
	"DialysisMethod":     {Code: models.DictTypeDialysisMode, Name: "透析方式"},
	"HeparinType":        {Code: models.DictTypeAnticoagulant, Name: "抗凝剂类型"},
	"Dialysate":          {Code: models.DictTypeDialysateType, Name: "透析液类型"},
	"DialysateFlow":      {Code: models.DictTypeDialysateFlow, Name: "透析液流量"},
	"GlucoseConOptions":  {Code: models.DictTypeGlucose, Name: "葡萄糖类型"},
	"MaterialType":       {Code: models.DictTypeMaterialCat, Name: "材料分类"},
	"DrugType":           {Code: models.DictTypeDrugCat, Name: "药品分类"},
	"UseMethodType":      {Code: models.DictTypeOrderType, Name: "医嘱类型"},
	"CatalogType":        {Code: models.DictTypeOrderCategory, Name: "医嘱分类"},
	"AccessType":         {Code: models.DictTypeVascularAccess, Name: "血管通路类型"},
	"AccessPosition":     {Code: models.DictTypeVascularSite, Name: "血管通路部位"},
	"VenousType":         {Code: models.DictTypeVeinType, Name: "静脉类型"},
	"ArteryType":         {Code: models.DictTypeArteryType, Name: "动脉类型"},
	"ExpenseType":        {Code: models.DictTypeInsuranceType, Name: "医保类型"},
	"PatientType":        {Code: models.DictTypePatientType, Name: "患者类型"},
	"IDType":             {Code: models.DictTypeIDType, Name: "证件类型"},
	"HospPatientType":    {Code: models.DictTypeVisitCategory, Name: "就诊类别"},
	"ABOType":            {Code: models.DictTypeBloodTypeABO, Name: "ABO血型"},
	"RHType":             {Code: models.DictTypeBloodTypeRH, Name: "Rh血型"},
	"EducationLevel":     {Code: models.DictTypeEducationLevel, Name: "文化程度"},
	"MaritalStatus":      {Code: models.DictTypeMaritalStatus, Name: "婚姻状况"},
	"UseWayType":         {Code: models.DictTypeOrderRoute, Name: "医嘱用法"},
	"FrequencyType":      {Code: models.DictTypeOrderFrequency, Name: "医嘱频次"},
	"UseOpportunityType": {Code: models.DictTypeOrderTiming, Name: "医嘱使用时机"},
	"OutComeType":        {Code: models.DictTypeOutcome, Name: "患者转归"},
	"OutComeReason":      {Code: models.DictTypeOutcome, Name: "患者转归"},
}

var legacyTypeToDisplayName = map[string]string{
	"RHType":                         "RH血型",
	"ABOType":                        "ABO血型",
	"AccessPosition":                 "通路部位",
	"VascularAccessChange_Type":      "手术类型",
	"AccessType":                     "通路类型",
	"DrugType":                       "药品类别",
	"ArteryType":                     "动脉类型",
	"VenousType":                     "静脉类型",
	"BasicUnitOptions":               "药品基本单位",
	"CatheterizeMethodType":          "中心静脉置管方法",
	"Dialysate":                      "透析液",
	"DialysisMethod":                 "治疗方式",
	"DilutionMnt":                    "稀释类别",
	"DiseaseCourseType":              "病程类型",
	"Disinfection10Disinfectant":     "机表消毒液",
	"Disinfection20Disinfectant":     "液路消毒液",
	"DisinfectionType":               "消毒类型",
	"Disinfection10Way":              "机表消毒方式",
	"Disinfection20Way":              "液路消毒方式",
	"DuringSymptomType":              "透中症状",
	"EducationLevel":                 "文化程度",
	"EquipmentInfomationMaintenance": "运维厂家",
	"EquipmentInfomationType":        "设备类型",
	"ExpenseType":                    "费用类别",
	"FamilyType":                     "家属类型",
	"HospPatientType":                "门诊类别",
	"IDType":                         "证件类型",
	"InfectionType":                  "传染病",
	"MaritalStatus":                  "婚姻状况",
	"MaterialType":                   "材料分类",
	"OutComeStatus":                  "转归状态",
	"OutComeType":                    "转归类型",
	"PatientType":                    "患者类型",
	"PressurePointOptions":           "测压部位",
	"RelationshipOptions":            "患者关系",
	"SealType":                       "封管液",
	"InfectionDay":                   "传染病检测周期",
	"SpecificationUnitOptions":       "规格单位",
	"OutComeReason":                  "转归原因",
	"DialysateFlow":                  "透析液流速",
	"FluxType":                       "通量",
	"UseMethodType":                  "医嘱用法",
	"UseWayType":                     "医嘱使用途径",
	"UseOpportunityType":             "医嘱使用时机",
	"ElectronicDocumentType":         "电子文书类型",
	"FrequencyType":                  "透析治疗频次描述",
	"HealthEducationType":            "宣教内容类型",
	"InfectionIntervalDay":           "传染病提醒间隔天数",
	"CatalogType":                    "项目种类",
}

var unifiedCodeToLegacyTypes = map[string][]string{
	models.DictTypeDialysisMode:   {"DialysisMethod"},
	models.DictTypeAnticoagulant:  {"HeparinType"},
	models.DictTypeDialysateType:  {"Dialysate"},
	models.DictTypeDialysateGroup: {"Dialysate"},
	models.DictTypeDialysateFlow:  {"DialysateFlow"},
	models.DictTypeGlucose:        {"GlucoseConOptions"},
	models.DictTypeMaterialCat:    {"MaterialType"},
	models.DictTypeDrugCat:        {"DrugType"},
	models.DictTypeOrderType:      {"UseMethodType"},
	models.DictTypeOrderCategory:  {"CatalogType"},
	models.DictTypeVascularAccess: {"AccessType"},
	models.DictTypeVascularSite:   {"AccessPosition"},
	models.DictTypeVeinType:       {"VenousType"},
	models.DictTypeArteryType:     {"ArteryType"},
	models.DictTypeInsuranceType:  {"ExpenseType"},
	models.DictTypePatientType:    {"PatientType"},
	models.DictTypeIDType:         {"IDType"},
	models.DictTypeVisitCategory:  {"HospPatientType"},
	models.DictTypeBloodTypeABO:   {"ABOType"},
	models.DictTypeBloodTypeRH:    {"RHType"},
	models.DictTypeEducationLevel: {"EducationLevel"},
	models.DictTypeMaritalStatus:  {"MaritalStatus"},
	models.DictTypeOrderRoute:     {"UseWayType"},
	models.DictTypeOrderFrequency: {"FrequencyType"},
	models.DictTypeOrderTiming:    {"UseOpportunityType"},
	models.DictTypeOutcome:        {"OutComeType", "OutComeReason"},
}

// ListTypes 获取字典类型列表
func (s *DictService) ListTypes() (*DictTypeListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	typesByCode := make(map[string]models.DictType)
	legacyTypes, err := s.listLegacyCodeDictionaryTypes()
	if err == nil {
		for _, item := range legacyTypes {
			typesByCode[item.Code] = item
		}
	} else if !isMissingRelationError(err) {
		return nil, err
	}

	if len(typesByCode) == 0 {
		for _, item := range legacyFallbackDictTypes() {
			typesByCode[item.Code] = item
		}
	}

	types := make([]models.DictType, 0, len(typesByCode))
	for _, item := range typesByCode {
		types = append(types, item)
	}
	sort.Slice(types, func(i, j int) bool {
		if types[i].SortOrder == types[j].SortOrder {
			return types[i].Code < types[j].Code
		}
		return types[i].SortOrder < types[j].SortOrder
	})

	return &DictTypeListResponse{
		Items: types,
		Total: int64(len(types)),
	}, nil
}

// GetTypeByCode 根据代码获取字典类型
func (s *DictService) GetTypeByCode(code string) (*models.DictType, error) {
	resp, err := s.ListTypes()
	if err != nil {
		return nil, err
	}
	for _, item := range resp.Items {
		if strings.EqualFold(item.Code, code) {
			dictType := item
			return &dictType, nil
		}
	}
	fallbackType, ok := legacyFallbackDictType(code)
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &fallbackType, nil
}

// GetItemsByTypeCode 获取指定类型的字典项列表
func (s *DictService) GetItemsByTypeCode(typeCode string, isEnabledOnly bool) (*DictItemListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	typeCode = strings.TrimSpace(typeCode)
	items, err := s.listLegacyCodeDictionaryItems(typeCode, isEnabledOnly)
	if err != nil && !isMissingRelationError(err) {
		return nil, err
	}
	if len(items) == 0 {
		items = legacyFallbackDictItems(typeCode, isEnabledOnly)
	}

	return &DictItemListResponse{
		Items: items,
		Total: int64(len(items)),
	}, nil
}

// GetItemsByTypeCodeTree 获取指定类型的字典项树形结构（用于级联选择）
func (s *DictService) GetItemsByTypeCodeTree(typeCode string, isEnabledOnly bool) ([]models.DictItem, error) {
	result, err := s.GetItemsByTypeCode(typeCode, isEnabledOnly)
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	return buildTree(result.Items, ""), nil
}

func isMissingRelationError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return (strings.Contains(message, "relation") && strings.Contains(message, "does not exist")) ||
		strings.Contains(message, "undefined_table") ||
		strings.Contains(message, "undefined_column")
}

func (s *DictService) listLegacyCodeDictionaryTypes() ([]models.DictType, error) {
	var rows []legacyCodeDictionaryTypeAgg
	err := s.db.Table(`"CodeDictionary_CodeDictionarys"`).
		Select(`"Type", MIN("Sort") AS "Sort", COUNT(1) AS "Cnt"`).
		Where(`"Type" IS NOT NULL AND TRIM("Type") <> ''`).
		Where(`COALESCE("IsDisabled", false) = false`).
		Group(`"Type"`).
		Order(`MIN("Sort") ASC, "Type" ASC`).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	now := time.Now()
	types := make([]models.DictType, 0, len(rows))
	for idx, row := range rows {
		legacyType := strings.TrimSpace(row.Type)
		if legacyType == "" {
			continue
		}
		code := legacyType
		name := legacyTypeDisplayName(legacyType)
		if unified, ok := legacyTypeToUnifiedCode[legacyType]; ok {
			code = unified.Code
			if unified.Name != "" {
				name = unified.Name
			}
		}
		sortOrder := row.Sort
		if sortOrder == 0 {
			sortOrder = (idx + 1) * 10
		}
		types = append(types, models.DictType{
			ID:          code,
			Code:        code,
			Name:        name,
			Description: legacyType,
			Source:      "legacy",
			SortOrder:   sortOrder,
			IsEnabled:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	dedup := make(map[string]models.DictType, len(types))
	for _, item := range types {
		existing, ok := dedup[item.Code]
		if !ok || item.SortOrder < existing.SortOrder {
			dedup[item.Code] = item
		}
	}
	result := make([]models.DictType, 0, len(dedup))
	for _, item := range dedup {
		result = append(result, item)
	}
	return result, nil
}

func (s *DictService) listLegacyCodeDictionaryItems(typeCode string, isEnabledOnly bool) ([]models.DictItem, error) {
	legacyTypes := unifiedCodeToLegacyTypes[typeCode]
	if len(legacyTypes) == 0 {
		legacyTypes = []string{typeCode}
	}

	if typeCode == models.DictTypeOutcome {
		return s.buildOutcomeDictItems(isEnabledOnly)
	}

	var rows []legacyCodeDictionaryRow
	query := s.db.Table(`"CodeDictionary_CodeDictionarys"`).
		Select(`"Type", "Code", "Name", "Sort", "IsDisabled"`).
		Where(`"Type" IN ?`, legacyTypes)
	if isEnabledOnly {
		query = query.Where(`COALESCE("IsDisabled", false) = false`)
	}
	if err := query.Order(`"Type" ASC, "Sort" ASC, "Code" ASC`).Find(&rows).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	items := make([]models.DictItem, 0, len(rows))
	for _, row := range rows {
		enabled := !row.IsDisabled
		if isEnabledOnly && !enabled {
			continue
		}
		itemCode := strings.TrimSpace(row.Code)
		if itemCode == "" {
			continue
		}
		itemName := strings.TrimSpace(row.Name)
		if itemName == "" {
			itemName = itemCode
		}
		items = append(items, models.DictItem{
			ID:          fmt.Sprintf("%s|%s|%s", typeCode, row.Type, itemCode),
			TypeCode:    typeCode,
			Code:        itemCode,
			Name:        itemName,
			Description: legacyTypeDisplayName(strings.TrimSpace(row.Type)),
			Source:      "legacy",
			SortOrder:   row.Sort,
			IsEnabled:   enabled,
			CreatedAt:   now,
			UpdatedAt:   now,
			ParentCode:  "",
		})
	}
	return items, nil
}

func (s *DictService) buildOutcomeDictItems(isEnabledOnly bool) ([]models.DictItem, error) {
	var rows []legacyCodeDictionaryRow
	query := s.db.Table(`"CodeDictionary_CodeDictionarys"`).
		Select(`"Type", "Code", "Name", "Sort", "IsDisabled"`).
		Where(`"Type" IN ?`, []string{"OutComeType", "OutComeReason"})
	if isEnabledOnly {
		query = query.Where(`COALESCE("IsDisabled", false) = false`)
	}
	if err := query.Order(`"Type" ASC, "Sort" ASC, "Code" ASC`).Find(&rows).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	items := make([]models.DictItem, 0, len(rows))
	parentCodeSet := make(map[string]struct{})

	for _, row := range rows {
		if row.Type != "OutComeType" {
			continue
		}
		code := strings.TrimSpace(row.Code)
		name := strings.TrimSpace(row.Name)
		if code == "" {
			continue
		}
		if name == "" {
			name = code
		}
		parentCodeSet[code] = struct{}{}
		items = append(items, models.DictItem{
			ID:         fmt.Sprintf("%s|%s|%s", models.DictTypeOutcome, row.Type, code),
			TypeCode:   models.DictTypeOutcome,
			Code:       code,
			Name:       name,
			Source:     "legacy",
			SortOrder:  row.Sort,
			IsEnabled:  !row.IsDisabled,
			CreatedAt:  now,
			UpdatedAt:  now,
			ParentCode: "",
		})
	}

	for idx, row := range rows {
		if row.Type != "OutComeReason" {
			continue
		}
		reasonCode := strings.TrimSpace(row.Code)
		rawName := strings.TrimSpace(row.Name)
		parentCode, reasonName := parseLegacyOutcomeReason(rawName)
		if reasonName == "" {
			reasonName = reasonCode
		}
		if reasonCode == "" {
			reasonCode = fmt.Sprintf("OUTCOME_REASON_%d", idx+1)
		}
		if _, ok := parentCodeSet[parentCode]; !ok {
			parentCode = ""
		}
		items = append(items, models.DictItem{
			ID:         fmt.Sprintf("%s|%s|%s", models.DictTypeOutcome, row.Type, reasonCode),
			TypeCode:   models.DictTypeOutcome,
			Code:       reasonCode,
			Name:       reasonName,
			Source:     "legacy",
			SortOrder:  row.Sort,
			IsEnabled:  !row.IsDisabled,
			CreatedAt:  now,
			UpdatedAt:  now,
			ParentCode: parentCode,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].ParentCode == items[j].ParentCode {
			if items[i].SortOrder == items[j].SortOrder {
				return items[i].Code < items[j].Code
			}
			return items[i].SortOrder < items[j].SortOrder
		}
		return items[i].ParentCode < items[j].ParentCode
	})
	return items, nil
}

func parseLegacyOutcomeReason(raw string) (parentCode, reasonName string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	parts := strings.SplitN(raw, "|", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return "", raw
}

func parseLegacyDictItemID(id string) (string, string, error) {
	parts := strings.SplitN(id, "|", 3)
	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid legacy dict item id: %s", id)
	}
	return strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2]), nil
}

func (s *DictService) queryLegacyDictItem(legacyType, code string) (name string, sortOrder int, isDisabled bool, err error) {
	type row struct {
		Name       string `gorm:"column:Name"`
		Sort       int    `gorm:"column:Sort"`
		IsDisabled bool   `gorm:"column:IsDisabled"`
	}
	var r row
	if err := s.db.Table(`"CodeDictionary_CodeDictionarys"`).
		Select(`"Name", "Sort", COALESCE("IsDisabled", false) AS "IsDisabled"`).
		Where(`"Type" = ? AND "Code" = ?`, legacyType, code).
		First(&r).Error; err != nil {
		return "", 0, false, err
	}
	return r.Name, r.Sort, r.IsDisabled, nil
}

func legacyTypeDisplayName(legacyType string) string {
	normalized := strings.TrimSpace(legacyType)
	if normalized == "" {
		return ""
	}
	if name, ok := legacyTypeToDisplayName[normalized]; ok && strings.TrimSpace(name) != "" {
		return name
	}
	return normalized
}

func legacyFallbackDictTypes() []models.DictType {
	return []models.DictType{
		{ID: models.DictTypeDialysisMode, Code: models.DictTypeDialysisMode, Name: "透析方式", SortOrder: 10, IsEnabled: true},
		{ID: models.DictTypeInsuranceType, Code: models.DictTypeInsuranceType, Name: "医保类型", SortOrder: 20, IsEnabled: true},
		{ID: models.DictTypePatientType, Code: models.DictTypePatientType, Name: "患者类型", SortOrder: 30, IsEnabled: true},
	}
}

func legacyFallbackDictType(code string) (models.DictType, bool) {
	for _, item := range legacyFallbackDictTypes() {
		if item.Code == code {
			return item, true
		}
	}
	return models.DictType{}, false
}

func legacyFallbackDictItems(typeCode string, isEnabledOnly bool) []models.DictItem {
	items := map[string][]models.DictItem{
		models.DictTypeDialysisMode: {
			{ID: "DIALYSIS_MODE_HD", TypeCode: models.DictTypeDialysisMode, Code: "HD", Name: "HD", SortOrder: 10, IsEnabled: true},
			{ID: "DIALYSIS_MODE_HDF", TypeCode: models.DictTypeDialysisMode, Code: "HDF", Name: "HDF", SortOrder: 20, IsEnabled: true},
			{ID: "DIALYSIS_MODE_HP", TypeCode: models.DictTypeDialysisMode, Code: "HP", Name: "HP", SortOrder: 30, IsEnabled: true},
			{ID: "DIALYSIS_MODE_HDHP", TypeCode: models.DictTypeDialysisMode, Code: "HD+HP", Name: "HD+HP", SortOrder: 40, IsEnabled: true},
		},
		models.DictTypeInsuranceType: {
			{ID: "INSURANCE_TYPE_1", TypeCode: models.DictTypeInsuranceType, Code: "市职工普通", Name: "市职工普通", SortOrder: 10, IsEnabled: true},
			{ID: "INSURANCE_TYPE_2", TypeCode: models.DictTypeInsuranceType, Code: "异地居民医保", Name: "异地居民医保", SortOrder: 20, IsEnabled: true},
			{ID: "INSURANCE_TYPE_3", TypeCode: models.DictTypeInsuranceType, Code: "城乡居民医保", Name: "城乡居民医保", SortOrder: 30, IsEnabled: true},
			{ID: "INSURANCE_TYPE_4", TypeCode: models.DictTypeInsuranceType, Code: "自费", Name: "自费", SortOrder: 40, IsEnabled: true},
		},
		models.DictTypePatientType: {
			{ID: "PATIENT_TYPE_10", TypeCode: models.DictTypePatientType, Code: "10", Name: "门诊", SortOrder: 10, IsEnabled: true},
			{ID: "PATIENT_TYPE_20", TypeCode: models.DictTypePatientType, Code: "20", Name: "住院", SortOrder: 20, IsEnabled: true},
		},
	}

	result := append([]models.DictItem(nil), items[typeCode]...)
	if !isEnabledOnly {
		return result
	}
	filtered := make([]models.DictItem, 0, len(result))
	for _, item := range result {
		if item.IsEnabled {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// buildTree 递归构建树形结构
func buildTree(items []models.DictItem, parentCode string) []models.DictItem {
	var result []models.DictItem

	for _, item := range items {
		// 检查是否属于当前父级
		var shouldInclude bool
		if parentCode == "" {
			// 顶级项：parent_code 为空
			shouldInclude = (item.ParentCode == "")
		} else {
			// 子项：parent_code 匹配
			shouldInclude = (item.ParentCode == parentCode)
		}

		if shouldInclude {
			// 递归获取子项
			item.Children = buildTree(items, item.Code)
			result = append(result, item)
		}
	}

	return result
}

// CreateType 创建字典类型 — 老库 CodeDictionary_CodeDictionarys 无独立类型表，类型通过条目 Type 列隐式管理
func (s *DictService) CreateType(dictType *models.DictType) error {
	return errors.New("字典类型创建暂不可用：老库 CodeDictionary_CodeDictionarys 无独立类型表，类型由条目 Type 列隐式定义")
}

// UpdateType 更新字典类型 — 老库无独立类型表，类型由条目 Type 列隐式管理
func (s *DictService) UpdateType(id string, updates map[string]interface{}) error {
	return errors.New("字典类型更新暂不可用：老库 CodeDictionary_CodeDictionarys 无独立类型表，请直接修改该类型下所有条目的 Type 列")
}

// DeleteType 删除字典类型 — 老库无独立类型表
func (s *DictService) DeleteType(id string) error {
	return errors.New("字典类型删除暂不可用：老库 CodeDictionary_CodeDictionarys 无独立类型表，请直接禁用/删除该类型下所有条目")
}

// CreateItem 创建字典项 — 写入老库 CodeDictionary_CodeDictionarys，返回可复用的 legacy ID
func (s *DictService) CreateItem(item *models.DictItem) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	if item.TypeCode == "" || item.Code == "" {
		return errors.New("字典类型代码和字典编码不能为空")
	}
	legacyTypeCode := item.TypeCode
	if legacyTypes, ok := unifiedCodeToLegacyTypes[item.TypeCode]; ok && len(legacyTypes) > 0 {
		legacyTypeCode = legacyTypes[0]
	}
	isDisabled := true
	if item.IsEnabled {
		isDisabled = false
	}
	if err := s.legacyDictUpsert(legacyTypeCode, item.Code, item.Name, item.SortOrder, isDisabled); err != nil {
		return err
	}
	item.ID = fmt.Sprintf("%s|%s|%s", legacyTypeCode, legacyTypeCode, item.Code)
	return nil
}

// UpdateItem 更新字典项 — 通过 legacy ID 定位老库 CodeDictionary_CodeDictionarys 条目
func (s *DictService) UpdateItem(id string, updates map[string]interface{}) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	legacyType, code, err := parseLegacyDictItemID(id)
	if err != nil {
		return err
	}

	name, sortOrder, isDisabled, err := s.queryLegacyDictItem(legacyType, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("字典项不存在")
		}
		return err
	}

	if v, ok := updates["name"].(string); ok && v != "" {
		name = v
	} else if v, ok := updates["Name"].(string); ok && v != "" {
		name = v
	}

	if v, ok := updates["sort_order"]; ok {
		switch sv := v.(type) {
		case float64:
			sortOrder = int(sv)
		case int:
			sortOrder = sv
		}
	} else if v, ok := updates["SortOrder"]; ok {
		switch sv := v.(type) {
		case float64:
			sortOrder = int(sv)
		case int:
			sortOrder = sv
		}
	}

	if v, ok := updates["is_enabled"].(bool); ok {
		isDisabled = !v
	} else if v, ok := updates["IsEnabled"].(bool); ok {
		isDisabled = !v
	}

	newCode := code
	if v, ok := updates["code"].(string); ok && v != "" {
		newCode = v
	} else if v, ok := updates["Code"].(string); ok && v != "" {
		newCode = v
	}

	if newCode != code {
		if err := s.legacyDictDisable(legacyType, code); err != nil {
			return err
		}
	}

	return s.legacyDictUpsert(legacyType, newCode, name, sortOrder, isDisabled)
}

// DeleteItem 删除字典项 — 通过 legacy ID 定位，软删除老库 CodeDictionary_CodeDictionarys 条目
func (s *DictService) DeleteItem(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	legacyType, code, err := parseLegacyDictItemID(id)
	if err != nil {
		return err
	}

	return s.legacyDictDisable(legacyType, code)
}

// ToggleItemEnabled 切换字典项启用状态 — 通过 legacy ID 定位，读写老库 CodeDictionary_CodeDictionarys
func (s *DictService) ToggleItemEnabled(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	legacyType, code, err := parseLegacyDictItemID(id)
	if err != nil {
		return err
	}

	name, sortOrder, isDisabled, err := s.queryLegacyDictItem(legacyType, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("字典项不存在")
		}
		return err
	}

	return s.legacyDictUpsert(legacyType, code, name, sortOrder, !isDisabled)
}

// ImportData 导入字典数据
type ImportData struct {
	Types []models.DictType `json:"types"`
	Items []models.DictItem `json:"items"`
}

// ImportResult 导入结果
type ImportResult struct {
	TypesCreated int `json:"typesCreated"`
	TypesUpdated int `json:"typesUpdated"`
	ItemsCreated int `json:"itemsCreated"`
	ItemsUpdated int `json:"itemsUpdated"`
}

// ImportDicts 批量导入字典数据 — 暂不可用（dict_types/dict_items 表已弃用）
func (s *DictService) ImportDicts(data *ImportData) (*ImportResult, error) {
	return nil, errors.New("批量导入暂不可用：dict_types/dict_items 表已弃用，请直接写入老库 CodeDictionary_CodeDictionarys")
}

// InitDefaultDicts 初始化默认字典数据
func (s *DictService) InitDefaultDicts() error {
	return errors.New("默认字典初始化暂不可用：dict_types/dict_items 表已弃用，字典数据请直接写入老库 CodeDictionary_CodeDictionarys")
}

// InitClinicalDicts 初始化临床诊疗分类字典数据 — 暂不可用
func (s *DictService) InitClinicalDicts() error {
	if s.db == nil {
		return errors.New("database not available")
	}

	// 初始化字典类型（在事务外，使用 FirstOrCreate 幂等操作）
	dictTypes := []models.DictType{
		{Code: models.DictTypePrimaryDisease, Name: "原发病分类", Description: "原发疾病分类", SortOrder: 18},
		{Code: models.DictTypeComplication, Name: "并发症类型", Description: "透析并发症分类", SortOrder: 19},
		{Code: models.DictTypePathology, Name: "病理诊断分类", Description: "病理诊断分类", SortOrder: 20},
		{Code: models.DictTypeTumor, Name: "肿瘤分类", Description: "肿瘤类型分类", SortOrder: 21},
		{Code: models.DictTypeAllergen, Name: "过敏原类型", Description: "过敏原分类", SortOrder: 22},
	}

	for _, dt := range dictTypes {
		dt.ID = uuid.New().String()
		if err := s.db.FirstOrCreate(&dt, models.DictType{Code: dt.Code}).Error; err != nil {
			return err
		}
	}

	// 使用事务清除旧数据并重建字典项
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 清除旧的临床字典项
		clinicalTypes := []string{
			models.DictTypePrimaryDisease,
			models.DictTypeComplication,
			models.DictTypePathology,
			models.DictTypeTumor,
			models.DictTypeAllergen,
		}
		for _, tc := range clinicalTypes {
			if err := tx.Where("type_code = ?", tc).Delete(&models.DictItem{}).Error; err != nil {
				return err
			}
		}

		// 批量创建字典项
		dictItems := getClinicalDictItems()
		for _, di := range dictItems {
			di.ID = uuid.New().String()
			if err := tx.Create(&di).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// InitOutcomeDicts 初始化转归字典数据（幂等：缺失则创建，已有则校正）
//
// 字典结构（树形）：
// OUTCOME (转归)
// ├── 10: 在科（一级，无 parent_code）
// │   └── IN_DEPT: 在科（二级，parent_code=10）
// └── 20: 转出（一级，无 parent_code）
//
//	├── TRANSFER_OUT: 转外院（二级，parent_code=20）
//	├── TRANSPLANT: 转肾移植（二级，parent_code=20）
//	├── PD_TRANSFER: 转腹透（二级，parent_code=20）
//	├── CURED: 病愈（二级，parent_code=20）
//	├── DEATH: 死亡（二级，parent_code=20）
//	└── QUIT: 退出（二级，parent_code=20）
func (s *DictService) InitOutcomeDicts() error {
	if s.db == nil {
		return errors.New("database not available")
	}

	log.Println("[InitOutcomeDicts] 开始初始化转归字典...")

	// 修复空 ID 的字典记录（历史遗留问题）
	var emptyIdTypes []models.DictType
	s.db.Where("id = '' OR id IS NULL").Find(&emptyIdTypes)
	for _, t := range emptyIdTypes {
		newID := uuid.New().String()
		s.db.Model(&models.DictType{}).Where("code = ? AND (id = '' OR id IS NULL)", t.Code).Update("id", newID)
		log.Printf("[InitOutcomeDicts] 修复空 ID 字典类型: %s → %s", t.Code, newID)
	}
	var emptyIdItems []models.DictItem
	s.db.Where("id = '' OR id IS NULL").Find(&emptyIdItems)
	for _, item := range emptyIdItems {
		newID := uuid.New().String()
		s.db.Model(&models.DictItem{}).Where("type_code = ? AND code = ? AND (id = '' OR id IS NULL)", item.TypeCode, item.Code).Update("id", newID)
		log.Printf("[InitOutcomeDicts] 修复空 ID 字典项: %s/%s → %s", item.TypeCode, item.Code, newID)
	}

	// 步骤 1: 清理旧的 OUTCOME_TYPE / OUTCOME_REASON 数据
	oldTypeCodes := []string{"OUTCOME_TYPE", "OUTCOME_REASON"}
	if result := s.db.Where("type_code IN ?", oldTypeCodes).Delete(&models.DictItem{}); result.Error != nil {
		log.Printf("[InitOutcomeDicts] 清理旧字典项失败: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("[InitOutcomeDicts] 已清理 %d 条旧字典项", result.RowsAffected)
	}
	if result := s.db.Where("code IN ?", oldTypeCodes).Delete(&models.DictType{}); result.Error != nil {
		log.Printf("[InitOutcomeDicts] 清理旧字典类型失败: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("[InitOutcomeDicts] 已清理 %d 个旧字典类型", result.RowsAffected)
	}

	// 步骤 2: 创建/更新 OUTCOME 字典类型
	var existingType models.DictType
	if err := s.db.Where("code = ?", models.DictTypeOutcome).First(&existingType).Error; err != nil {
		// 不存在，创建
		existingType = models.DictType{
			ID:          uuid.New().String(),
			Code:        models.DictTypeOutcome,
			Name:        "患者转归",
			Description: "患者转归分类（一级：在科/转出，二级：具体原因）",
			Icon:        "📋",
			SortOrder:   210,
			IsEnabled:   true,
		}
		if err := s.db.Create(&existingType).Error; err != nil {
			log.Printf("[InitOutcomeDicts] 创建 OUTCOME 字典类型失败: %v", err)
			return err
		}
		log.Println("[InitOutcomeDicts] 已创建 OUTCOME 字典类型")
	} else {
		// 已存在，更新
		s.db.Model(&existingType).Updates(map[string]interface{}{
			"name":        "患者转归",
			"description": "患者转归分类（一级：在科/转出，二级：具体原因）",
			"icon":        "📋",
			"sort_order":  210,
			"is_enabled":  true,
		})
		log.Println("[InitOutcomeDicts] OUTCOME 字典类型已存在，已更新")
	}

	// 步骤 3: 创建/更新字典项（用 map 避免 GORM 忽略零值 ParentCode）
	type itemDef struct {
		Code        string
		Name        string
		Description string
		ParentCode  string
		SortOrder   int
	}
	items := []itemDef{
		// 一级分类
		{Code: "10", Name: "在科", Description: "患者仍在科室治疗", ParentCode: "", SortOrder: 1},
		{Code: "20", Name: "转出", Description: "患者转出科室", ParentCode: "", SortOrder: 2},
		// 二级 - 在科
		{Code: "IN_DEPT", Name: "在科", Description: "患者仍在科室", ParentCode: "10", SortOrder: 1},
		// 二级 - 转出
		{Code: "TRANSFER_OUT", Name: "转外院", Description: "转往外院治疗", ParentCode: "20", SortOrder: 1},
		{Code: "TRANSPLANT", Name: "转肾移植", Description: "转为肾移植治疗", ParentCode: "20", SortOrder: 2},
		{Code: "PD_TRANSFER", Name: "转腹透", Description: "转为腹膜透析", ParentCode: "20", SortOrder: 3},
		{Code: "CURED", Name: "病愈", Description: "患者病愈出院", ParentCode: "20", SortOrder: 4},
		{Code: "DEATH", Name: "死亡", Description: "患者死亡", ParentCode: "20", SortOrder: 5},
		{Code: "QUIT", Name: "退出", Description: "患者退出治疗", ParentCode: "20", SortOrder: 6},
	}

	created, updated := 0, 0
	for _, item := range items {
		var existing models.DictItem
		err := s.db.Where("type_code = ? AND code = ?", models.DictTypeOutcome, item.Code).First(&existing).Error
		if err != nil {
			// 不存在，创建
			newItem := models.DictItem{
				ID:          uuid.New().String(),
				TypeCode:    models.DictTypeOutcome,
				Code:        item.Code,
				Name:        item.Name,
				Description: item.Description,
				ParentCode:  item.ParentCode,
				SortOrder:   item.SortOrder,
				IsEnabled:   true,
			}
			if err := s.db.Create(&newItem).Error; err != nil {
				log.Printf("[InitOutcomeDicts] 创建字典项 %s 失败: %v", item.Code, err)
				return err
			}
			created++
		} else {
			// 已存在，用 map 更新（确保空 parent_code 也能写入）
			s.db.Model(&existing).Updates(map[string]interface{}{
				"name":        item.Name,
				"description": item.Description,
				"parent_code": item.ParentCode,
				"sort_order":  item.SortOrder,
				"is_enabled":  true,
			})
			updated++
		}
	}

	log.Printf("[InitOutcomeDicts] 完成: 新建 %d 条, 更新 %d 条", created, updated)
	return nil
}

// getClinicalDictItems 获取临床诊疗字典项数据

// InitMethodDrugs 初始化方法种子数据（幂等：已存在则跳过）
// 在 drug_catalogs 表中插入非药品方法项目
func (s *DictService) InitMethodDrugs() error {
	if s.db == nil {
		return errors.New("database not available")
	}

	log.Println("[InitMethodDrugs] 开始初始化方法种子数据...")

	// 迁移旧数据：将 category="方法" 更新为 "METHOD"（兼容常量变更前创建的记录）
	if result := s.db.Model(&models.DrugCatalog{}).Where("category = ?", "方法").Update("category", models.DrugCategoryMethod); result.RowsAffected > 0 {
		log.Printf("[InitMethodDrugs] 已迁移 %d 条旧方法数据的 category 字段", result.RowsAffected)
	}

	type methodDef struct {
		Code     string
		Name     string
		BaseUnit string
	}
	methods := []methodDef{
		{Code: "MTD_DRESSING", Name: "换药", BaseUnit: "次"},
		{Code: "MTD_SUCTION", Name: "吸痰", BaseUnit: "次"},
		{Code: "MTD_ECG_MONITOR", Name: "心电监护", BaseUnit: "小时"},
		{Code: "MTD_O2_INHALE", Name: "氧气吸入", BaseUnit: "小时"},
		{Code: "MTD_SPO2_MONITOR", Name: "脉氧饱和度监测", BaseUnit: "小时"},
	}

	created := 0
	for i, m := range methods {
		var existing models.DrugCatalog
		err := s.db.Where("code = ?", m.Code).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在，创建
			newItem := models.DrugCatalog{
				Code:      m.Code,
				Name:      m.Name,
				Category:  models.DrugCategoryMethod,
				BaseUnit:  m.BaseUnit,
				IsEnabled: true,
				SortOrder: i + 1,
			}
			if err := s.db.Create(&newItem).Error; err != nil {
				log.Printf("[InitMethodDrugs] 创建 %s 失败: %v", m.Code, err)
				return err
			}
			created++
		} else if err != nil {
			log.Printf("[InitMethodDrugs] 查询 %s 失败: %v", m.Code, err)
			return err
		}
		// 已存在则跳过
	}

	log.Printf("[InitMethodDrugs] 完成: 新建 %d 条方法数据", created)
	return nil
}

func getClinicalDictItems() []models.DictItem {
	return []models.DictItem{
		// ========== 原发病分类 PRIMARY_DISEASE ==========
		// 1. 原发性肾小球疾病（父级）
		{TypeCode: models.DictTypePrimaryDisease, Code: "1", Name: "原发性肾小球疾病", SortOrder: 1},
		{TypeCode: models.DictTypePrimaryDisease, Code: "1-01", Name: "急性肾炎综合征", ParentCode: "1", SortOrder: 1},
		{TypeCode: models.DictTypePrimaryDisease, Code: "1-02", Name: "急进性肾炎综合征", ParentCode: "1", SortOrder: 2},
		{TypeCode: models.DictTypePrimaryDisease, Code: "1-03", Name: "慢性肾炎综合征", ParentCode: "1", SortOrder: 3},
		{TypeCode: models.DictTypePrimaryDisease, Code: "1-04", Name: "肾病综合征", ParentCode: "1", SortOrder: 4},
		{TypeCode: models.DictTypePrimaryDisease, Code: "1-05", Name: "血尿", ParentCode: "1", SortOrder: 5},
		{TypeCode: models.DictTypePrimaryDisease, Code: "1-06", Name: "孤立性尿蛋白", ParentCode: "1", SortOrder: 6},
		{TypeCode: models.DictTypePrimaryDisease, Code: "1-99", Name: "其他", ParentCode: "1", SortOrder: 7},

		// 2. 继发性肾小球疾病（父级）
		{TypeCode: models.DictTypePrimaryDisease, Code: "2", Name: "继发性肾小球疾病", SortOrder: 2},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-01", Name: "高血压肾损害", ParentCode: "2", SortOrder: 1},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-02", Name: "糖尿病肾病", ParentCode: "2", SortOrder: 2},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-03", Name: "肥胖相关性肾病", ParentCode: "2", SortOrder: 3},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-04", Name: "淀粉样变肾损害", ParentCode: "2", SortOrder: 4},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-05", Name: "多发骨髓瘤肾病", ParentCode: "2", SortOrder: 5},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-06", Name: "狼疮性肾炎", ParentCode: "2", SortOrder: 6},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-07", Name: "系统性血管炎肾损害", ParentCode: "2", SortOrder: 7},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-08", Name: "过敏紫癜性肾炎", ParentCode: "2", SortOrder: 8},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-09", Name: "血栓性微血管病肾损害", ParentCode: "2", SortOrder: 9},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-10", Name: "干燥综合征肾损害", ParentCode: "2", SortOrder: 10},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-11", Name: "硬皮病肾损害", ParentCode: "2", SortOrder: 11},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-12", Name: "类风湿性关节炎和强直性脊柱炎肾损害", ParentCode: "2", SortOrder: 12},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-13", Name: "银屑病肾损害", ParentCode: "2", SortOrder: 13},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-14", Name: "乙型肝炎病毒相关性肾炎", ParentCode: "2", SortOrder: 14},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-15", Name: "丙型肝炎病毒相关性肾炎", ParentCode: "2", SortOrder: 15},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-16", Name: "HIV相关性肾损害", ParentCode: "2", SortOrder: 16},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-17", Name: "流行性出血热肾损害", ParentCode: "2", SortOrder: 17},
		{TypeCode: models.DictTypePrimaryDisease, Code: "2-99", Name: "其他", ParentCode: "2", SortOrder: 18},

		// 3. 遗传性及先天性疾病（父级）
		{TypeCode: models.DictTypePrimaryDisease, Code: "3", Name: "遗传性及先天性疾病", SortOrder: 3},
		{TypeCode: models.DictTypePrimaryDisease, Code: "3-01", Name: "多囊肾病", ParentCode: "3", SortOrder: 1},
		{TypeCode: models.DictTypePrimaryDisease, Code: "3-02", Name: "Alport综合征", ParentCode: "3", SortOrder: 2},
		{TypeCode: models.DictTypePrimaryDisease, Code: "3-03", Name: "薄基底膜肾病", ParentCode: "3", SortOrder: 3},
		{TypeCode: models.DictTypePrimaryDisease, Code: "3-04", Name: "近端肾小管损伤及Fanconi综合征", ParentCode: "3", SortOrder: 4},
		{TypeCode: models.DictTypePrimaryDisease, Code: "3-05", Name: "Bartter综合征", ParentCode: "3", SortOrder: 5},
		{TypeCode: models.DictTypePrimaryDisease, Code: "3-06", Name: "Fabry病", ParentCode: "3", SortOrder: 6},
		{TypeCode: models.DictTypePrimaryDisease, Code: "3-07", Name: "脂蛋白肾病", ParentCode: "3", SortOrder: 7},
		{TypeCode: models.DictTypePrimaryDisease, Code: "3-08", Name: "肾发育不良", ParentCode: "3", SortOrder: 8},
		{TypeCode: models.DictTypePrimaryDisease, Code: "3-99", Name: "其他", ParentCode: "3", SortOrder: 9},

		// 4. 肾小管间质疾病（父级）
		{TypeCode: models.DictTypePrimaryDisease, Code: "4", Name: "肾小管间质疾病", SortOrder: 4},
		{TypeCode: models.DictTypePrimaryDisease, Code: "4-01", Name: "急性肾小管间质性肾炎", ParentCode: "4", SortOrder: 1},
		{TypeCode: models.DictTypePrimaryDisease, Code: "4-02", Name: "慢性肾小管间质性肾炎", ParentCode: "4", SortOrder: 2},
		{TypeCode: models.DictTypePrimaryDisease, Code: "4-03", Name: "急性肾小管坏死", ParentCode: "4", SortOrder: 3},
		{TypeCode: models.DictTypePrimaryDisease, Code: "4-04", Name: "肾小管性酸中毒", ParentCode: "4", SortOrder: 4},
		{TypeCode: models.DictTypePrimaryDisease, Code: "4-05", Name: "慢性肾盂肾炎", ParentCode: "4", SortOrder: 5},
		{TypeCode: models.DictTypePrimaryDisease, Code: "4-06", Name: "反流性肾病", ParentCode: "4", SortOrder: 6},
		{TypeCode: models.DictTypePrimaryDisease, Code: "4-07", Name: "梗阻性肾病", ParentCode: "4", SortOrder: 7},
		{TypeCode: models.DictTypePrimaryDisease, Code: "4-99", Name: "其他", ParentCode: "4", SortOrder: 8},

		// 5. 药物性肾损害（父级）
		{TypeCode: models.DictTypePrimaryDisease, Code: "5", Name: "药物性肾损害", SortOrder: 5},
		{TypeCode: models.DictTypePrimaryDisease, Code: "5-01", Name: "马兜铃酸肾病", ParentCode: "5", SortOrder: 1},
		{TypeCode: models.DictTypePrimaryDisease, Code: "5-02", Name: "造影剂肾病", ParentCode: "5", SortOrder: 2},
		{TypeCode: models.DictTypePrimaryDisease, Code: "5-03", Name: "重金属中毒性肾脏损害", ParentCode: "5", SortOrder: 3},
		{TypeCode: models.DictTypePrimaryDisease, Code: "5-04", Name: "放射性肾病及抗肿瘤药物所致的肾损害", ParentCode: "5", SortOrder: 4},
		{TypeCode: models.DictTypePrimaryDisease, Code: "5-05", Name: "氨基苷类抗生素肾损害", ParentCode: "5", SortOrder: 5},
		{TypeCode: models.DictTypePrimaryDisease, Code: "5-99", Name: "其他", ParentCode: "5", SortOrder: 6},

		// 6. 泌尿系肿瘤（独立项）
		{TypeCode: models.DictTypePrimaryDisease, Code: "6", Name: "泌尿系肿瘤", SortOrder: 6},

		// 7. 泌尿系感染和结石（父级）
		{TypeCode: models.DictTypePrimaryDisease, Code: "7", Name: "泌尿系感染和结石", SortOrder: 7},
		{TypeCode: models.DictTypePrimaryDisease, Code: "7-01", Name: "泌尿系结核", ParentCode: "7", SortOrder: 1},
		{TypeCode: models.DictTypePrimaryDisease, Code: "7-02", Name: "肾结石", ParentCode: "7", SortOrder: 2},
		{TypeCode: models.DictTypePrimaryDisease, Code: "7-99", Name: "其他", ParentCode: "7", SortOrder: 3},

		// 8. 肾脏切除术后（独立项）
		{TypeCode: models.DictTypePrimaryDisease, Code: "8", Name: "肾脏切除术后", SortOrder: 8},

		// 9. 原发病不明确（独立项）
		{TypeCode: models.DictTypePrimaryDisease, Code: "9", Name: "原发病不明确", SortOrder: 9},

		// ========== 并发症类型 COMPLICATION ==========
		// 01. 肾性贫血（独立项）
		{TypeCode: models.DictTypeComplication, Code: "01", Name: "肾性贫血", SortOrder: 1},

		// 02. 骨矿物质代谢紊乱（父级）
		{TypeCode: models.DictTypeComplication, Code: "02", Name: "骨矿物质代谢紊乱", SortOrder: 2},
		{TypeCode: models.DictTypeComplication, Code: "02-01", Name: "高运转骨病（需要骨活检支持）", ParentCode: "02", SortOrder: 1},
		{TypeCode: models.DictTypeComplication, Code: "02-02", Name: "低运转骨病（需要骨活检支持）", ParentCode: "02", SortOrder: 2},
		{TypeCode: models.DictTypeComplication, Code: "02-03", Name: "混合型骨病（需要骨活检支持）", ParentCode: "02", SortOrder: 3},
		{TypeCode: models.DictTypeComplication, Code: "02-04", Name: "转移性钙化", ParentCode: "02", SortOrder: 4},
		{TypeCode: models.DictTypeComplication, Code: "02-05", Name: "骨质疏松", ParentCode: "02", SortOrder: 5},
		{TypeCode: models.DictTypeComplication, Code: "02-06", Name: "继发性甲旁亢", ParentCode: "02", SortOrder: 6},
		{TypeCode: models.DictTypeComplication, Code: "02-99", Name: "其他", ParentCode: "02", SortOrder: 7},

		// 03. 营养不良（独立项）
		{TypeCode: models.DictTypeComplication, Code: "03", Name: "营养不良", SortOrder: 3},

		// 04. 淀粉样变性（父级）
		{TypeCode: models.DictTypeComplication, Code: "04", Name: "淀粉样变性", SortOrder: 4},
		{TypeCode: models.DictTypeComplication, Code: "04-01", Name: "腕管综合征", ParentCode: "04", SortOrder: 1},
		{TypeCode: models.DictTypeComplication, Code: "04-02", Name: "心脏损害", ParentCode: "04", SortOrder: 2},
		{TypeCode: models.DictTypeComplication, Code: "04-03", Name: "骨损害", ParentCode: "04", SortOrder: 3},
		{TypeCode: models.DictTypeComplication, Code: "04-99", Name: "其他", ParentCode: "04", SortOrder: 4},

		// 05. 呼吸系统（父级）
		{TypeCode: models.DictTypeComplication, Code: "05", Name: "呼吸系统", SortOrder: 5},
		{TypeCode: models.DictTypeComplication, Code: "05-01", Name: "肺部感染", ParentCode: "05", SortOrder: 1},
		{TypeCode: models.DictTypeComplication, Code: "05-02", Name: "结核", ParentCode: "05", SortOrder: 2},
		{TypeCode: models.DictTypeComplication, Code: "05-03", Name: "胸膜炎", ParentCode: "05", SortOrder: 3},
		{TypeCode: models.DictTypeComplication, Code: "05-04", Name: "胸腔积液", ParentCode: "05", SortOrder: 4},
		{TypeCode: models.DictTypeComplication, Code: "05-05", Name: "尿毒症肺炎", ParentCode: "05", SortOrder: 5},
		{TypeCode: models.DictTypeComplication, Code: "05-99", Name: "其他", ParentCode: "05", SortOrder: 6},

		// 06. 心血管系统（父级）
		{TypeCode: models.DictTypeComplication, Code: "06", Name: "心血管系统", SortOrder: 6},
		{TypeCode: models.DictTypeComplication, Code: "06-01", Name: "高血压", ParentCode: "06", SortOrder: 1},
		{TypeCode: models.DictTypeComplication, Code: "06-02", Name: "低血压", ParentCode: "06", SortOrder: 2},
		{TypeCode: models.DictTypeComplication, Code: "06-03", Name: "心律失常", ParentCode: "06", SortOrder: 3},
		{TypeCode: models.DictTypeComplication, Code: "06-04", Name: "心功能不全", ParentCode: "06", SortOrder: 4},
		{TypeCode: models.DictTypeComplication, Code: "06-05", Name: "急性左心衰竭", ParentCode: "06", SortOrder: 5},
		{TypeCode: models.DictTypeComplication, Code: "06-06", Name: "缺血性心脏病", ParentCode: "06", SortOrder: 6},
		{TypeCode: models.DictTypeComplication, Code: "06-07", Name: "心包炎", ParentCode: "06", SortOrder: 7},
		{TypeCode: models.DictTypeComplication, Code: "06-08", Name: "心肌病变", ParentCode: "06", SortOrder: 8},
		{TypeCode: models.DictTypeComplication, Code: "06-99", Name: "其他", ParentCode: "06", SortOrder: 9},

		// 07. 神经系统（父级）
		{TypeCode: models.DictTypeComplication, Code: "07", Name: "神经系统", SortOrder: 7},
		{TypeCode: models.DictTypeComplication, Code: "07-01", Name: "脑梗塞", ParentCode: "07", SortOrder: 1},
		{TypeCode: models.DictTypeComplication, Code: "07-02", Name: "脑出血", ParentCode: "07", SortOrder: 2},
		{TypeCode: models.DictTypeComplication, Code: "07-03", Name: "神经性病变", ParentCode: "07", SortOrder: 3},
		{TypeCode: models.DictTypeComplication, Code: "07-04", Name: "尿毒性脑病", ParentCode: "07", SortOrder: 4},
		{TypeCode: models.DictTypeComplication, Code: "07-99", Name: "其他", ParentCode: "07", SortOrder: 5},

		// 08. 消化系统（父级）
		{TypeCode: models.DictTypeComplication, Code: "08", Name: "消化系统", SortOrder: 8},
		{TypeCode: models.DictTypeComplication, Code: "08-01", Name: "肝硬化", ParentCode: "08", SortOrder: 1},
		{TypeCode: models.DictTypeComplication, Code: "08-02", Name: "消化道出血", ParentCode: "08", SortOrder: 2},
		{TypeCode: models.DictTypeComplication, Code: "08-99", Name: "其他", ParentCode: "08", SortOrder: 3},

		// 09. 皮肤瘙痒（独立项）
		{TypeCode: models.DictTypeComplication, Code: "09", Name: "皮肤瘙痒", SortOrder: 9},

		// 10. 不安腿（独立项）
		{TypeCode: models.DictTypeComplication, Code: "10", Name: "不安腿", SortOrder: 10},

		// 99. 其他（独立项）
		{TypeCode: models.DictTypeComplication, Code: "99", Name: "其他", SortOrder: 11},

		// ========== 病理诊断分类 PATHOLOGY ==========
		// 1. 原发性肾小球疾病（父级）
		{TypeCode: models.DictTypePathology, Code: "1", Name: "原发性肾小球疾病", SortOrder: 1},
		{TypeCode: models.DictTypePathology, Code: "1-01", Name: "肾小球轻微病变", ParentCode: "1", SortOrder: 1},
		{TypeCode: models.DictTypePathology, Code: "1-02", Name: "微小病变性肾病", ParentCode: "1", SortOrder: 2},
		{TypeCode: models.DictTypePathology, Code: "1-03", Name: "局灶阶段性肾小球损害", ParentCode: "1", SortOrder: 3},
		{TypeCode: models.DictTypePathology, Code: "1-04", Name: "膜性肾病", ParentCode: "1", SortOrder: 4},
		{TypeCode: models.DictTypePathology, Code: "1-05", Name: "系膜增殖性肾炎", ParentCode: "1", SortOrder: 5},
		{TypeCode: models.DictTypePathology, Code: "1-06", Name: "IgA肾病", ParentCode: "1", SortOrder: 6},
		{TypeCode: models.DictTypePathology, Code: "1-07", Name: "毛细血管内增殖性肾炎", ParentCode: "1", SortOrder: 7},
		{TypeCode: models.DictTypePathology, Code: "1-08", Name: "膜增殖性肾炎", ParentCode: "1", SortOrder: 8},
		{TypeCode: models.DictTypePathology, Code: "1-09", Name: "新月体肾炎", ParentCode: "1", SortOrder: 9},
		{TypeCode: models.DictTypePathology, Code: "1-10", Name: "硬化性肾炎", ParentCode: "1", SortOrder: 10},
		{TypeCode: models.DictTypePathology, Code: "1-99", Name: "其他", ParentCode: "1", SortOrder: 11},

		// 2. 继发性肾小球疾病（父级）
		{TypeCode: models.DictTypePathology, Code: "2", Name: "继发性肾小球疾病", SortOrder: 2},
		{TypeCode: models.DictTypePathology, Code: "2-01", Name: "高血压肾硬化", ParentCode: "2", SortOrder: 1},
		{TypeCode: models.DictTypePathology, Code: "2-02", Name: "糖尿病肾病", ParentCode: "2", SortOrder: 2},
		{TypeCode: models.DictTypePathology, Code: "2-03", Name: "肥胖相关性肾病", ParentCode: "2", SortOrder: 3},
		{TypeCode: models.DictTypePathology, Code: "2-04", Name: "淀粉样变性", ParentCode: "2", SortOrder: 4},
		{TypeCode: models.DictTypePathology, Code: "2-05", Name: "多发骨髓瘤肾病", ParentCode: "2", SortOrder: 5},
		{TypeCode: models.DictTypePathology, Code: "2-06", Name: "冷球蛋白血症性肾炎", ParentCode: "2", SortOrder: 6},
		{TypeCode: models.DictTypePathology, Code: "2-07", Name: "轻链型肾病", ParentCode: "2", SortOrder: 7},
		{TypeCode: models.DictTypePathology, Code: "2-08", Name: "狼疮性肾炎", ParentCode: "2", SortOrder: 8},
		{TypeCode: models.DictTypePathology, Code: "2-09", Name: "过敏紫癜性肾炎", ParentCode: "2", SortOrder: 9},
		{TypeCode: models.DictTypePathology, Code: "2-10", Name: "抗基底膜肾炎（Goodpasture综合征）", ParentCode: "2", SortOrder: 10},
		{TypeCode: models.DictTypePathology, Code: "2-11", Name: "系统性血管炎", ParentCode: "2", SortOrder: 11},
		{TypeCode: models.DictTypePathology, Code: "2-12", Name: "血栓性微血管病", ParentCode: "2", SortOrder: 12},
		{TypeCode: models.DictTypePathology, Code: "2-13", Name: "干燥综合征肾损害", ParentCode: "2", SortOrder: 13},
		{TypeCode: models.DictTypePathology, Code: "2-14", Name: "硬皮病肾损害", ParentCode: "2", SortOrder: 14},
		{TypeCode: models.DictTypePathology, Code: "2-15", Name: "乙型肝炎病毒相关性肾炎", ParentCode: "2", SortOrder: 15},
		{TypeCode: models.DictTypePathology, Code: "2-16", Name: "丙型肝炎病毒相关性肾炎", ParentCode: "2", SortOrder: 16},
		{TypeCode: models.DictTypePathology, Code: "2-17", Name: "HIV相关性肾损害", ParentCode: "2", SortOrder: 17},
		{TypeCode: models.DictTypePathology, Code: "2-18", Name: "流行性出血热肾损害", ParentCode: "2", SortOrder: 18},
		{TypeCode: models.DictTypePathology, Code: "2-99", Name: "其他", ParentCode: "2", SortOrder: 19},

		// 3. 遗传性及先天性肾病（父级）
		{TypeCode: models.DictTypePathology, Code: "3", Name: "遗传性及先天性肾病", SortOrder: 3},
		{TypeCode: models.DictTypePathology, Code: "3-01", Name: "Alport综合征", ParentCode: "3", SortOrder: 1},
		{TypeCode: models.DictTypePathology, Code: "3-02", Name: "薄基底膜肾病", ParentCode: "3", SortOrder: 2},
		{TypeCode: models.DictTypePathology, Code: "3-03", Name: "近端肾小管损伤及Fanconi综合征", ParentCode: "3", SortOrder: 3},
		{TypeCode: models.DictTypePathology, Code: "3-04", Name: "Bartter综合征", ParentCode: "3", SortOrder: 4},
		{TypeCode: models.DictTypePathology, Code: "3-05", Name: "Fabry病", ParentCode: "3", SortOrder: 5},
		{TypeCode: models.DictTypePathology, Code: "3-06", Name: "脂蛋白肾病", ParentCode: "3", SortOrder: 6},
		{TypeCode: models.DictTypePathology, Code: "3-07", Name: "肾发育不良", ParentCode: "3", SortOrder: 7},
		{TypeCode: models.DictTypePathology, Code: "3-99", Name: "其他", ParentCode: "3", SortOrder: 8},

		// 4. 肾小管间质疾病（父级）
		{TypeCode: models.DictTypePathology, Code: "4", Name: "肾小管间质疾病", SortOrder: 4},
		{TypeCode: models.DictTypePathology, Code: "4-01", Name: "急性肾小管间质性肾炎", ParentCode: "4", SortOrder: 1},
		{TypeCode: models.DictTypePathology, Code: "4-02", Name: "慢性肾小管间质性肾炎", ParentCode: "4", SortOrder: 2},
		{TypeCode: models.DictTypePathology, Code: "4-03", Name: "急性肾小管坏死", ParentCode: "4", SortOrder: 3},
		{TypeCode: models.DictTypePathology, Code: "4-04", Name: "马兜铃酸肾病", ParentCode: "4", SortOrder: 4},
		{TypeCode: models.DictTypePathology, Code: "4-99", Name: "其他", ParentCode: "4", SortOrder: 5},

		// ========== 肿瘤分类 TUMOR ==========
		// 全部独立项（无子项）
		{TypeCode: models.DictTypeTumor, Code: "1", Name: "消化系统", SortOrder: 1},
		{TypeCode: models.DictTypeTumor, Code: "2", Name: "呼吸系统", SortOrder: 2},
		{TypeCode: models.DictTypeTumor, Code: "3", Name: "血液系统", SortOrder: 3},
		{TypeCode: models.DictTypeTumor, Code: "4", Name: "泌尿生殖系统", SortOrder: 4},
		{TypeCode: models.DictTypeTumor, Code: "5", Name: "神经系统", SortOrder: 5},
		{TypeCode: models.DictTypeTumor, Code: "6", Name: "骨骼肌肉系统", SortOrder: 6},
		{TypeCode: models.DictTypeTumor, Code: "7", Name: "其他", SortOrder: 7},

		// ========== 过敏原类型 ALLERGEN ==========
		// 1. 透析器材过敏（父级）
		{TypeCode: models.DictTypeAllergen, Code: "1", Name: "透析器材过敏", SortOrder: 1},
		{TypeCode: models.DictTypeAllergen, Code: "1-01", Name: "本次使用膜材料", ParentCode: "1", SortOrder: 1},
		{TypeCode: models.DictTypeAllergen, Code: "1-02", Name: "消毒方式", ParentCode: "1", SortOrder: 2},

		// 2. 药物过敏（父级）
		{TypeCode: models.DictTypeAllergen, Code: "2", Name: "药物过敏", SortOrder: 2},
		{TypeCode: models.DictTypeAllergen, Code: "2-01", Name: "抗生素", ParentCode: "2", SortOrder: 1},
		{TypeCode: models.DictTypeAllergen, Code: "2-02", Name: "静脉铁剂", ParentCode: "2", SortOrder: 2},
		{TypeCode: models.DictTypeAllergen, Code: "2-03", Name: "肝素", ParentCode: "2", SortOrder: 3},
		{TypeCode: models.DictTypeAllergen, Code: "2-99", Name: "其他", ParentCode: "2", SortOrder: 4},

		// 3. 食物过敏（独立项）
		{TypeCode: models.DictTypeAllergen, Code: "3", Name: "食物过敏", SortOrder: 3},

		// 9. 其他过敏（独立项）
		{TypeCode: models.DictTypeAllergen, Code: "9", Name: "其他过敏", SortOrder: 4},
	}
}

// ─── 老库字典写操作（CodeDictionary_CodeDictionarys） ───

func (s *DictService) legacyDictUpsert(typeCode, code, name string, sortOrder int, isDisabled bool) error {
	if s.db == nil || typeCode == "" || code == "" {
		return nil
	}
	var count int64
	s.db.Table(`"CodeDictionary_CodeDictionarys"`).
		Where(`"Type" = ? AND "Code" = ?`, typeCode, code).
		Count(&count)
	if count > 0 {
		return s.db.Table(`"CodeDictionary_CodeDictionarys"`).
			Where(`"Type" = ? AND "Code" = ?`, typeCode, code).
			Updates(map[string]interface{}{
				"Name":       name,
				"Sort":       sortOrder,
				"IsDisabled": isDisabled,
			}).Error
	}
	return s.db.Table(`"CodeDictionary_CodeDictionarys"`).Create(map[string]interface{}{
		"Type":       typeCode,
		"Code":       code,
		"Name":       name,
		"Sort":       sortOrder,
		"IsDisabled": isDisabled,
		"OrganId":    0,
		"Builtin":    false,
	}).Error
}

func (s *DictService) legacyDictDisable(typeCode, code string) error {
	if s.db == nil || typeCode == "" || code == "" {
		return nil
	}
	return s.db.Table(`"CodeDictionary_CodeDictionarys"`).
		Where(`"Type" = ? AND "Code" = ?`, typeCode, code).
		Update("IsDisabled", true).Error
}
