package services

import (
	"errors"
	"log"
	"strings"

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

// ListTypes 获取字典类型列表
func (s *DictService) ListTypes() (*DictTypeListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var types []models.DictType
	if err := s.db.Order("sort_order ASC").Find(&types).Error; err != nil {
		return nil, err
	}

	return &DictTypeListResponse{
		Items: types,
		Total: int64(len(types)),
	}, nil
}

// GetTypeByCode 根据代码获取字典类型
func (s *DictService) GetTypeByCode(code string) (*models.DictType, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var dictType models.DictType
	if err := s.db.Where("code = ?", code).First(&dictType).Error; err != nil {
		return nil, err
	}

	return &dictType, nil
}

// GetItemsByTypeCode 获取指定类型的字典项列表
func (s *DictService) GetItemsByTypeCode(typeCode string, isEnabledOnly bool) (*DictItemListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	query := s.db.Where("type_code = ?", typeCode)
	if isEnabledOnly {
		query = query.Where("is_enabled = ?", true)
	}

	var items []models.DictItem
	if err := query.Order("sort_order ASC").Find(&items).Error; err != nil {
		return nil, err
	}

	return &DictItemListResponse{
		Items: items,
		Total: int64(len(items)),
	}, nil
}

// GetItemsByTypeCodeTree 获取指定类型的字典项树形结构（用于级联选择）
func (s *DictService) GetItemsByTypeCodeTree(typeCode string, isEnabledOnly bool) ([]models.DictItem, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	query := s.db.Where("type_code = ?", typeCode)
	if isEnabledOnly {
		query = query.Where("is_enabled = ?", true)
	}

	var items []models.DictItem
	if err := query.Order("sort_order ASC").Find(&items).Error; err != nil {
		return nil, err
	}

	// 构建树形结构
	return buildTree(items, ""), nil
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

// CreateType 创建字典类型
func (s *DictService) CreateType(dictType *models.DictType) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	dictType.ID = uuid.New().String()
	return s.db.Create(dictType).Error
}

// UpdateType 更新字典类型（code 变更时级联更新 dict_items.type_code）
func (s *DictService) UpdateType(id string, updates map[string]interface{}) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	// 未修改 code，直接更新
	newCodeRaw, codeChanged := updates["code"]
	if !codeChanged {
		return s.db.Model(&models.DictType{}).Where("id = ?", id).Updates(updates).Error
	}

	// code 修改时需要级联更新 dict_items.type_code
	var dictType models.DictType
	if err := s.db.Where("id = ?", id).First(&dictType).Error; err != nil {
		return errors.New("字典类型不存在")
	}

	oldCode := dictType.Code
	newCode, ok := newCodeRaw.(string)
	if !ok {
		return errors.New("字典类型代码格式错误")
	}
	newCode = strings.TrimSpace(newCode)
	if newCode == "" {
		return errors.New("字典类型代码不能为空")
	}
	updates["code"] = newCode

	// code 未变化，直接更新其他字段
	if oldCode == newCode {
		return s.db.Model(&models.DictType{}).Where("id = ?", id).Updates(updates).Error
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 先更新字典类型本身（受唯一约束保护）
		if err := tx.Model(&models.DictType{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}

		// 再级联更新该类型下所有字典项的 type_code
		if err := tx.Model(&models.DictItem{}).
			Where("type_code = ?", oldCode).
			Update("type_code", newCode).Error; err != nil {
			return err
		}
		return nil
	})
}

// DeleteType 删除字典类型（级联删除关联的字典项）
func (s *DictService) DeleteType(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	// 先获取字典类型的 code
	var dictType models.DictType
	if err := s.db.Where("id = ?", id).First(&dictType).Error; err != nil {
		return errors.New("字典类型不存在")
	}

	// 使用事务确保数据一致性
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 先删除该类型下的所有字典项
		if err := tx.Where("type_code = ?", dictType.Code).Delete(&models.DictItem{}).Error; err != nil {
			return err
		}

		// 再删除字典类型
		if err := tx.Delete(&models.DictType{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

// CreateItem 创建字典项
func (s *DictService) CreateItem(item *models.DictItem) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	// 验证字典类型是否存在
	var dictType models.DictType
	if err := s.db.Where("code = ?", item.TypeCode).First(&dictType).Error; err != nil {
		return errors.New("字典类型不存在")
	}

	item.ID = uuid.New().String()
	return s.db.Create(item).Error
}

// UpdateItem 更新字典项（code 变更时级联更新后代 code/parent_code）
func (s *DictService) UpdateItem(id string, updates map[string]interface{}) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	// 检查是否修改了 code 字段
	newCode, codeChanged := updates["code"]
	if !codeChanged {
		// 无 code 变更，直接更新
		return s.db.Model(&models.DictItem{}).Where("id = ?", id).Updates(updates).Error
	}

	// code 被修改，需要级联更新子项
	var item models.DictItem
	if err := s.db.Where("id = ?", id).First(&item).Error; err != nil {
		return errors.New("字典项不存在")
	}

	oldCode := item.Code
	newCodeStr, ok := newCode.(string)
	if !ok {
		return errors.New("字典编码格式错误")
	}
	newCodeStr = strings.TrimSpace(newCodeStr)
	if newCodeStr == "" {
		return errors.New("字典编码不能为空")
	}
	updates["code"] = newCodeStr

	// code 未变化，直接更新
	if oldCode == newCodeStr {
		return s.db.Model(&models.DictItem{}).Where("id = ?", id).Updates(updates).Error
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 先取出全部后代，再按前缀改写 code/parent_code，避免只级联一层
		descendants, err := collectDescendantItems(tx, item.TypeCode, oldCode)
		if err != nil {
			return err
		}

		for _, descendant := range descendants {
			nextCode := replaceCodePrefix(descendant.Code, oldCode, newCodeStr)
			nextParentCode := replaceCodePrefix(descendant.ParentCode, oldCode, newCodeStr)
			descendantUpdates := map[string]interface{}{}
			if nextCode != descendant.Code {
				descendantUpdates["code"] = nextCode
			}
			if nextParentCode != descendant.ParentCode {
				descendantUpdates["parent_code"] = nextParentCode
			}
			if len(descendantUpdates) > 0 {
				if err := tx.Model(&models.DictItem{}).
					Where("id = ?", descendant.ID).
					Updates(descendantUpdates).Error; err != nil {
					return err
				}
			}
		}

		// 最后更新当前项，避免和后代更新顺序互相影响
		if err := tx.Model(&models.DictItem{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
}

// DeleteItem 删除字典项（递归级联删除所有后代子项）
func (s *DictService) DeleteItem(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	// 先查出该字典项，获取 code 和 type_code
	var item models.DictItem
	if err := s.db.Where("id = ?", id).First(&item).Error; err != nil {
		return errors.New("字典项不存在")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 收集所有后代的 code
		descendantCodes, err := collectDescendantCodes(tx, item.TypeCode, item.Code)
		if err != nil {
			return err
		}
		// 删除所有后代子项
		if len(descendantCodes) > 0 {
			if err := tx.Where("type_code = ? AND code IN ?", item.TypeCode, descendantCodes).Delete(&models.DictItem{}).Error; err != nil {
				return err
			}
		}
		// 删除该项本身
		if err := tx.Delete(&models.DictItem{}, "id = ?", id).Error; err != nil {
			return err
		}
		return nil
	})
}

// collectDescendantCodes 收集所有后代的 code（单次查询 + 内存 BFS）
func collectDescendantCodes(tx *gorm.DB, typeCode string, parentCode string) ([]string, error) {
	// 一次性加载该类型下所有项的 code 和 parent_code
	type codePair struct {
		Code       string
		ParentCode string
	}
	var pairs []codePair
	if err := tx.Model(&models.DictItem{}).
		Where("type_code = ?", typeCode).
		Select("code, parent_code").
		Scan(&pairs).Error; err != nil {
		return nil, err
	}

	// 构建 parentCode -> childCodes 映射
	childMap := make(map[string][]string)
	for _, p := range pairs {
		if p.ParentCode != "" {
			childMap[p.ParentCode] = append(childMap[p.ParentCode], p.Code)
		}
	}

	// BFS 收集所有后代（visited 防循环）
	var result []string
	visited := make(map[string]bool)
	queue := childMap[parentCode]
	for len(queue) > 0 {
		code := queue[0]
		queue = queue[1:]
		if visited[code] {
			continue
		}
		visited[code] = true
		result = append(result, code)
		queue = append(queue, childMap[code]...)
	}
	return result, nil
}

// collectDescendantItems 收集所有后代字典项（单次查询 + 内存 BFS）
func collectDescendantItems(tx *gorm.DB, typeCode string, parentCode string) ([]models.DictItem, error) {
	// 一次性加载该类型下所有项
	var allItems []models.DictItem
	if err := tx.Where("type_code = ?", typeCode).Find(&allItems).Error; err != nil {
		return nil, err
	}

	// 构建 parentCode -> items 映射
	childMap := make(map[string][]models.DictItem)
	for _, item := range allItems {
		if item.ParentCode != "" {
			childMap[item.ParentCode] = append(childMap[item.ParentCode], item)
		}
	}

	// BFS 收集所有后代（visited 防循环）
	var result []models.DictItem
	visited := make(map[string]bool)
	queue := childMap[parentCode]
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		if visited[item.Code] {
			continue
		}
		visited[item.Code] = true
		result = append(result, item)
		queue = append(queue, childMap[item.Code]...)
	}
	return result, nil
}

// replaceCodePrefix 替换 code 的父前缀（仅匹配完整 code 或 code- 子级）
func replaceCodePrefix(value string, oldCode string, newCode string) string {
	if value == oldCode {
		return newCode
	}
	prefix := oldCode + "-"
	if strings.HasPrefix(value, prefix) {
		return newCode + strings.TrimPrefix(value, oldCode)
	}
	return value
}

// ToggleItemEnabled 切换字典项启用状态
func (s *DictService) ToggleItemEnabled(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	var item models.DictItem
	if err := s.db.Where("id = ?", id).First(&item).Error; err != nil {
		return err
	}

	return s.db.Model(&item).Update("is_enabled", !item.IsEnabled).Error
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

// ImportDicts 批量导入字典数据
func (s *DictService) ImportDicts(data *ImportData) (*ImportResult, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	result := &ImportResult{}

	// 使用事务确保数据一致性
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 导入字典类型
		for _, dt := range data.Types {
			var existing models.DictType
			err := tx.Where("code = ?", dt.Code).First(&existing).Error

			switch err {
			case gorm.ErrRecordNotFound:
				// 创建新类型
				dt.ID = uuid.New().String()
				if err := tx.Create(&dt).Error; err != nil {
					return err
				}
				result.TypesCreated++
			case nil:
				// 更新现有类型
				updates := map[string]interface{}{
					"name":        dt.Name,
					"description": dt.Description,
					"icon":        dt.Icon,
					"sort_order":  dt.SortOrder,
					"is_enabled":  dt.IsEnabled,
				}
				if err := tx.Model(&existing).Updates(updates).Error; err != nil {
					return err
				}
				result.TypesUpdated++
			default:
				return err
			}
		}

		// 导入字典项
		for _, di := range data.Items {
			var existing models.DictItem
			err := tx.Where("type_code = ? AND code = ?", di.TypeCode, di.Code).First(&existing).Error

			switch err {
			case gorm.ErrRecordNotFound:
				// 创建新项
				di.ID = uuid.New().String()
				if err := tx.Create(&di).Error; err != nil {
					return err
				}
				result.ItemsCreated++
			case nil:
				// 更新现有项
				updates := map[string]interface{}{
					"name":        di.Name,
					"description": di.Description,
					"sort_order":  di.SortOrder,
					"is_enabled":  di.IsEnabled,
					"extra":       di.Extra,
					"parent_code": di.ParentCode,
				}
				if err := tx.Model(&existing).Updates(updates).Error; err != nil {
					return err
				}
				result.ItemsUpdated++
			default:
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// InitDefaultDicts 初始化默认字典数据
func (s *DictService) InitDefaultDicts() error {
	if s.db == nil {
		return errors.New("database not available")
	}

	// 初始化字典类型
	dictTypes := []models.DictType{
		{Code: models.DictTypeDialysisMode, Name: "透析方式", Description: "血液透析治疗方式", SortOrder: 1},
		{Code: models.DictTypeAnticoagulant, Name: "抗凝剂类型", Description: "透析抗凝剂类型", SortOrder: 2},
		{Code: models.DictTypeDialysateType, Name: "透析液类型", Description: "透析液配方类型", SortOrder: 3},
		{Code: models.DictTypeDialysateGroup, Name: "透析液组", Description: "透析液离子组分", SortOrder: 4},
		{Code: models.DictTypeDialysateFlow, Name: "透析液流量", Description: "透析液流速", SortOrder: 5},
		{Code: models.DictTypeGlucose, Name: "葡萄糖类型", Description: "透析液葡萄糖含量", SortOrder: 6},
		{Code: models.DictTypeMaterialCat, Name: "材料分类", Description: "透析耗材分类", SortOrder: 7},
		{Code: models.DictTypeDrugCat, Name: "药品分类", Description: "透析药品分类", SortOrder: 8},
		{Code: models.DictTypeOrderType, Name: "医嘱类型", Description: "医嘱类型", SortOrder: 9},
		{Code: models.DictTypeOrderCategory, Name: "医嘱分类", Description: "医嘱分类", SortOrder: 10},
		{Code: models.DictTypeVascularAccess, Name: "血管通路类型", Description: "血管通路类型", SortOrder: 11},
		{Code: models.DictTypeVascularSite, Name: "血管通路部位", Description: "血管通路部位", SortOrder: 12},
		{Code: models.DictTypeVeinType, Name: "静脉类型", Description: "静脉类型", SortOrder: 13},
		{Code: models.DictTypeArteryType, Name: "动脉类型", Description: "动脉类型", SortOrder: 14},
		{Code: models.DictTypeDoctor, Name: "医生列表", Description: "透析中心医生", SortOrder: 15},
		{Code: models.DictTypeHospital, Name: "手术医院", Description: "血管通路手术医院", SortOrder: 16},
		{Code: models.DictTypeSurgeryType, Name: "手术类型", Description: "血管通路干预手术类型", SortOrder: 17},
		{Code: models.DictTypeOrderRoute, Name: "用法", Description: "医嘱用法/给药途径", SortOrder: 23},
		{Code: models.DictTypeOrderFrequency, Name: "频次", Description: "医嘱执行频次", SortOrder: 24},
		{Code: models.DictTypeOrderTiming, Name: "使用时机", Description: "透析中用药时机", SortOrder: 25},
	}

	for _, dt := range dictTypes {
		// 幂等操作：不存在则创建，存在则更新
		var existing models.DictType
		s.db.Where(models.DictType{Code: dt.Code}).
			Assign(models.DictType{
				Name:        dt.Name,
				Description: dt.Description,
				SortOrder:   dt.SortOrder,
			}).
			FirstOrCreate(&existing)
	}

	// 初始化字典项
	dictItems := []models.DictItem{
		// 透析方式
		{TypeCode: models.DictTypeDialysisMode, Code: "HD", Name: "血液透析", Description: "Hemodialysis", SortOrder: 1},
		{TypeCode: models.DictTypeDialysisMode, Code: "HDF", Name: "血液透析滤过", Description: "Hemodiafiltration", SortOrder: 2},
		{TypeCode: models.DictTypeDialysisMode, Code: "HF", Name: "血液过滤", Description: "Hemofiltration", SortOrder: 3},
		{TypeCode: models.DictTypeDialysisMode, Code: "HP", Name: "血液灌流", Description: "Hemoperfusion", SortOrder: 4},
		{TypeCode: models.DictTypeDialysisMode, Code: "CRRT", Name: "连续肾脏替代治疗", Description: "Continuous Renal Replacement Therapy", SortOrder: 5},

		// 抗凝剂类型
		{TypeCode: models.DictTypeAnticoagulant, Code: "HEPARIN", Name: "肝素钠", Description: "普通肝素", SortOrder: 1},
		{TypeCode: models.DictTypeAnticoagulant, Code: "NADROPARIN", Name: "那屈肝素钙", Description: "低分子肝素", SortOrder: 2},
		{TypeCode: models.DictTypeAnticoagulant, Code: "LMWH", Name: "低分子肝素", Description: "Low Molecular Weight Heparin", SortOrder: 3},
		{TypeCode: models.DictTypeAnticoagulant, Code: "NO_HEPARIN", Name: "无肝素", Description: "无抗凝剂透析", SortOrder: 4},
		{TypeCode: models.DictTypeAnticoagulant, Code: "CITRATE", Name: "枸橼酸钠", Description: "局部枸橼酸抗凝", SortOrder: 5},

		// 透析液类型
		{TypeCode: models.DictTypeDialysateType, Code: "A_B", Name: "A液+B液", Description: "碳酸氢盐透析液", SortOrder: 1},
		{TypeCode: models.DictTypeDialysateType, Code: "POWDER", Name: "干粉A+干粉B", Description: "干粉装透析液", SortOrder: 2},
		{TypeCode: models.DictTypeDialysateType, Code: "BAG", Name: "成品袋装", Description: "袋装透析液", SortOrder: 3},

		// 透析液组
		{TypeCode: models.DictTypeDialysateGroup, Code: "A16_K2_CA1.25", Name: "A16-K2/Ca1.25", Description: "钠140 钾2.0 钙1.25", SortOrder: 1},
		{TypeCode: models.DictTypeDialysateGroup, Code: "A18_K3_CA1.5", Name: "A18-K3/Ca1.5", Description: "钠140 钾3.0 钙1.5", SortOrder: 2},
		{TypeCode: models.DictTypeDialysateGroup, Code: "A20_K1.5_CA1.0", Name: "A20-K1.5/Ca1.0", Description: "钠140 钾1.5 钙1.0", SortOrder: 3},
		{TypeCode: models.DictTypeDialysateGroup, Code: "STANDARD", Name: "标准组", Description: "标准配方", SortOrder: 4},

		// 透析液流量
		{TypeCode: models.DictTypeDialysateFlow, Code: "300", Name: "300 ml/min", Description: "", SortOrder: 1},
		{TypeCode: models.DictTypeDialysateFlow, Code: "500", Name: "500 ml/min", Description: "", SortOrder: 2},
		{TypeCode: models.DictTypeDialysateFlow, Code: "700", Name: "700 ml/min", Description: "", SortOrder: 3},
		{TypeCode: models.DictTypeDialysateFlow, Code: "800", Name: "800 ml/min", Description: "", SortOrder: 4},

		// 葡萄糖类型
		{TypeCode: models.DictTypeGlucose, Code: "NO_GLUCOSE", Name: "无糖", Description: "0 mmol/L", SortOrder: 1},
		{TypeCode: models.DictTypeGlucose, Code: "GLUCOSE_1.1", Name: "含糖 (1.1)", Description: "5.5 mmol/L", SortOrder: 2},
		{TypeCode: models.DictTypeGlucose, Code: "GLUCOSE_2.0", Name: "含糖 (2.0)", Description: "11.1 mmol/L", SortOrder: 3},
		{TypeCode: models.DictTypeGlucose, Code: "CUSTOM", Name: "自定义", Description: "自定义葡萄糖浓度", SortOrder: 4},

		// 材料分类
		{TypeCode: models.DictTypeMaterialCat, Code: "DIALYZER", Name: "透析器", Description: "", SortOrder: 1},
		{TypeCode: models.DictTypeMaterialCat, Code: "BLOODLINE", Name: "血液管路", Description: "", SortOrder: 2},
		{TypeCode: models.DictTypeMaterialCat, Code: "NEEDLE", Name: "穿刺针", Description: "", SortOrder: 3},
		{TypeCode: models.DictTypeMaterialCat, Code: "CARE_KIT", Name: "护理包", Description: "", SortOrder: 4},
		{TypeCode: models.DictTypeMaterialCat, Code: "DISINFECTANT", Name: "消毒液", Description: "", SortOrder: 5},
		{TypeCode: models.DictTypeMaterialCat, Code: "OTHER", Name: "其他耗材", Description: "", SortOrder: 6},

		// 药品分类
		{TypeCode: models.DictTypeDrugCat, Code: "ANTICOAGULANT", Name: "抗凝药", Description: "", SortOrder: 1},
		{TypeCode: models.DictTypeDrugCat, Code: "EPO", Name: "促红素", Description: "红细胞生成刺激剂", SortOrder: 2},
		{TypeCode: models.DictTypeDrugCat, Code: "VITAMIN", Name: "维生素类", Description: "", SortOrder: 3},
		{TypeCode: models.DictTypeDrugCat, Code: "ELECTROLYTE", Name: "电解质调节", Description: "", SortOrder: 4},
		{TypeCode: models.DictTypeDrugCat, Code: "ANTIHYPERTENSIVE", Name: "降压药", Description: "", SortOrder: 5},
		{TypeCode: models.DictTypeDrugCat, Code: "EMERGENCY", Name: "紧急用药", Description: "", SortOrder: 6},
		{TypeCode: models.DictTypeDrugCat, Code: "OTHER_DRUG", Name: "其他药品", Description: "", SortOrder: 7},

		// 医嘱类型
		{TypeCode: models.DictTypeOrderType, Code: "LONG_TERM", Name: "长期", Description: "长期医嘱", SortOrder: 1},
		{TypeCode: models.DictTypeOrderType, Code: "TEMPORARY", Name: "临时", Description: "临时医嘱", SortOrder: 2},

		// 医嘱分类
		{TypeCode: models.DictTypeOrderCategory, Code: "BASIC", Name: "基础类", Description: "", SortOrder: 1},
		{TypeCode: models.DictTypeOrderCategory, Code: "MEDICATION", Name: "用药类", Description: "", SortOrder: 2},
		{TypeCode: models.DictTypeOrderCategory, Code: "EMERGENCY", Name: "应急类", Description: "", SortOrder: 3},
		{TypeCode: models.DictTypeOrderCategory, Code: "EXAMINATION", Name: "检查类", Description: "", SortOrder: 4},

		// 血管通路类型
		{TypeCode: models.DictTypeVascularAccess, Code: "AVF", Name: "内瘘 - AVF", Description: "自体动静脉内瘘", SortOrder: 1},
		{TypeCode: models.DictTypeVascularAccess, Code: "AVG", Name: "内瘘 - AVG", Description: "移植物动静脉内瘘", SortOrder: 2},
		{TypeCode: models.DictTypeVascularAccess, Code: "TCC", Name: "导管 - TCC", Description: "带隧道和涤纶套的透析导管", SortOrder: 3},
		{TypeCode: models.DictTypeVascularAccess, Code: "NCC", Name: "导管 - NCC", Description: "无隧道和涤纶套的透析导管", SortOrder: 4},

		// 血管通路部位
		{TypeCode: models.DictTypeVascularSite, Code: "FOREARM", Name: "前臂", Description: "", SortOrder: 1},
		{TypeCode: models.DictTypeVascularSite, Code: "FOREARM_MID", Name: "前臂中段", Description: "", SortOrder: 2},
		{TypeCode: models.DictTypeVascularSite, Code: "WRIST", Name: "腕部", Description: "", SortOrder: 3},
		{TypeCode: models.DictTypeVascularSite, Code: "ELBOW", Name: "肘部", Description: "", SortOrder: 4},
		{TypeCode: models.DictTypeVascularSite, Code: "UPPER_ARM", Name: "上臂", Description: "", SortOrder: 5},
		{TypeCode: models.DictTypeVascularSite, Code: "NECK", Name: "颈部", Description: "", SortOrder: 6},
		{TypeCode: models.DictTypeVascularSite, Code: "SUBCLAVIAN", Name: "锁骨下", Description: "", SortOrder: 7},
		{TypeCode: models.DictTypeVascularSite, Code: "GROIN", Name: "腹股沟", Description: "", SortOrder: 8},
		{TypeCode: models.DictTypeVascularSite, Code: "LOWER_LIMB", Name: "下肢", Description: "", SortOrder: 9},

		// 静脉类型
		{TypeCode: models.DictTypeVeinType, Code: "CEPHALIC", Name: "头静脉", Description: "", SortOrder: 1},
		{TypeCode: models.DictTypeVeinType, Code: "BASILIC", Name: "贵要静脉", Description: "", SortOrder: 2},
		{TypeCode: models.DictTypeVeinType, Code: "MEDIAN_CUBITAL", Name: "肘正中静脉", Description: "", SortOrder: 3},
		{TypeCode: models.DictTypeVeinType, Code: "JUGULAR", Name: "颈内静脉", Description: "", SortOrder: 4},
		{TypeCode: models.DictTypeVeinType, Code: "SUBCLAVIAN", Name: "锁骨下静脉", Description: "", SortOrder: 5},
		{TypeCode: models.DictTypeVeinType, Code: "FEMORAL", Name: "股静脉", Description: "", SortOrder: 6},
		{TypeCode: models.DictTypeVeinType, Code: "MEDIAN_ANTEBRACHIAL", Name: "前臂正中静脉", Description: "", SortOrder: 7},

		// 动脉类型
		{TypeCode: models.DictTypeArteryType, Code: "RADIAL", Name: "桡动脉", Description: "", SortOrder: 1},
		{TypeCode: models.DictTypeArteryType, Code: "BRACHIAL", Name: "肱动脉", Description: "", SortOrder: 2},
		{TypeCode: models.DictTypeArteryType, Code: "ULNAR", Name: "尺动脉", Description: "", SortOrder: 3},
		{TypeCode: models.DictTypeArteryType, Code: "FEMORAL", Name: "股动脉", Description: "", SortOrder: 4},

		// 医生列表
		{TypeCode: models.DictTypeDoctor, Code: "DR_WANG", Name: "王医生", Description: "主任医师", SortOrder: 1},
		{TypeCode: models.DictTypeDoctor, Code: "DR_LI", Name: "李医生", Description: "主治医师", SortOrder: 2},
		{TypeCode: models.DictTypeDoctor, Code: "DR_ZHANG", Name: "张医生", Description: "副主任医师", SortOrder: 3},
		{TypeCode: models.DictTypeDoctor, Code: "DR_CHEN", Name: "陈医生", Description: "住院医师", SortOrder: 4},
		{TypeCode: models.DictTypeDoctor, Code: "DR_LIU", Name: "刘医生", Description: "主治医师", SortOrder: 5},
		{TypeCode: models.DictTypeDoctor, Code: "DR_ZHAO", Name: "赵医生", Description: "住院医师", SortOrder: 6},

		// 手术医院
		{TypeCode: models.DictTypeHospital, Code: "H_LOCAL", Name: "本院", Description: "本透析中心", SortOrder: 1},
		{TypeCode: models.DictTypeHospital, Code: "H_PEK_UNION", Name: "北京协和医院", Description: "血管外科", SortOrder: 2},
		{TypeCode: models.DictTypeHospital, Code: "H_PEK_301", Name: "解放军总医院（301医院）", Description: "血管外科", SortOrder: 3},
		{TypeCode: models.DictTypeHospital, Code: "H_SHANGHAI_HUA_SHAN", Name: "复旦大学附属华山医院", Description: "血管外科", SortOrder: 4},
		{TypeCode: models.DictTypeHospital, Code: "H_GUANG_ZHONG_FIRST", Name: "中山大学附属第一医院", Description: "血管外科", SortOrder: 5},
		{TypeCode: models.DictTypeHospital, Code: "H_ZHE_DA_FIRST", Name: "浙江大学附属第一医院", Description: "血管外科", SortOrder: 6},
		{TypeCode: models.DictTypeHospital, Code: "H_SICHUAN_PROVINCIAL", Name: "四川省人民医院", Description: "血管外科", SortOrder: 7},
		{TypeCode: models.DictTypeHospital, Code: "H_XIJING", Name: "西京医院", Description: "血管外科", SortOrder: 8},
		{TypeCode: models.DictTypeHospital, Code: "H_OTHER", Name: "其他医院", Description: "其他医疗机构", SortOrder: 9},

		// 手术类型（血管通路干预）
		{TypeCode: models.DictTypeSurgeryType, Code: "PTA", Name: "PTA（经皮腔内血管成形术）", Description: "Percutaneous Transluminal Angioplasty", SortOrder: 1},
		{TypeCode: models.DictTypeSurgeryType, Code: "THROMBECTOMY", Name: "血栓清除术", Description: "Fogarty导管取栓", SortOrder: 2},
		{TypeCode: models.DictTypeSurgeryType, Code: "REVISION", Name: "修复术", Description: "内瘘修复手术", SortOrder: 3},
		{TypeCode: models.DictTypeSurgeryType, Code: "LIGATION", Name: "结扎术", Description: "血管结扎手术", SortOrder: 4},
		{TypeCode: models.DictTypeSurgeryType, Code: "CATHETER_EXCHANGE", Name: "导管更换", Description: "透析导管更换", SortOrder: 5},
		{TypeCode: models.DictTypeSurgeryType, Code: "CATHETER_REPOSITIONING", Name: "导管重新定位", Description: "导管位置调整", SortOrder: 6},
		{TypeCode: models.DictTypeSurgeryType, Code: "STENT", Name: "支架植入", Description: "血管支架植入术", SortOrder: 7},
		{TypeCode: models.DictTypeSurgeryType, Code: "BALLOON", Name: "球囊扩张", Description: "球囊扩张术", SortOrder: 8},
		{TypeCode: models.DictTypeSurgeryType, Code: "OTHER", Name: "其他", Description: "其他手术类型", SortOrder: 9},

		// 药品分类 - 方法
		{TypeCode: models.DictTypeDrugCat, Code: "METHOD", Name: "方法", Description: "非药品方法项目", SortOrder: 8, IsEnabled: true},

		// 用法（给药途径）
		{TypeCode: models.DictTypeOrderRoute, Code: "IV_PUSH", Name: "静脉推注", SortOrder: 1, IsEnabled: true},
		{TypeCode: models.DictTypeOrderRoute, Code: "IV_DRIP", Name: "静脉滴注", SortOrder: 2, IsEnabled: true},
		{TypeCode: models.DictTypeOrderRoute, Code: "SC", Name: "皮下注射", SortOrder: 3, IsEnabled: true},
		{TypeCode: models.DictTypeOrderRoute, Code: "IM", Name: "肌肉注射", SortOrder: 4, IsEnabled: true},
		{TypeCode: models.DictTypeOrderRoute, Code: "PO", Name: "口服", SortOrder: 5, IsEnabled: true},
		{TypeCode: models.DictTypeOrderRoute, Code: "IV_PUMP", Name: "静脉泵入", SortOrder: 6, IsEnabled: true},
		{TypeCode: models.DictTypeOrderRoute, Code: "SUBLINGUAL_DISSOLVE", Name: "含化", SortOrder: 7, IsEnabled: true},
		{TypeCode: models.DictTypeOrderRoute, Code: "TUBE_LOCK", Name: "封管", SortOrder: 8, IsEnabled: true},
		{TypeCode: models.DictTypeOrderRoute, Code: "SUBLINGUAL", Name: "舌下给药", SortOrder: 9, IsEnabled: true},
		{TypeCode: models.DictTypeOrderRoute, Code: "RECTAL", Name: "直肠给药", SortOrder: 10, IsEnabled: true},
		{TypeCode: models.DictTypeOrderRoute, Code: "MUCOSAL", Name: "皮肤黏膜给药", SortOrder: 11, IsEnabled: true},
		{TypeCode: models.DictTypeOrderRoute, Code: "INHALATION", Name: "呼吸道给药（雾化吸入）", SortOrder: 12, IsEnabled: true},
		{TypeCode: models.DictTypeOrderRoute, Code: "TOPICAL", Name: "经皮给药（外用）", SortOrder: 13, IsEnabled: true},

		// 频次
		{TypeCode: models.DictTypeOrderFrequency, Code: "STAT", Name: "临时一次", SortOrder: 1, IsEnabled: true},
		{TypeCode: models.DictTypeOrderFrequency, Code: "QW", Name: "每周一次", SortOrder: 2, IsEnabled: true},
		{TypeCode: models.DictTypeOrderFrequency, Code: "BIW", Name: "每周两次", SortOrder: 3, IsEnabled: true},
		{TypeCode: models.DictTypeOrderFrequency, Code: "TIW", Name: "每周三次", SortOrder: 4, IsEnabled: true},
		{TypeCode: models.DictTypeOrderFrequency, Code: "Q2W", Name: "两周一次", SortOrder: 5, IsEnabled: true},
		{TypeCode: models.DictTypeOrderFrequency, Code: "QM", Name: "一月一次", SortOrder: 6, IsEnabled: true},

		// 使用时机
		{TypeCode: models.DictTypeOrderTiming, Code: "ALL", Name: "首+中+末", SortOrder: 1, IsEnabled: true},
		{TypeCode: models.DictTypeOrderTiming, Code: "MID_END", Name: "中+末", SortOrder: 2, IsEnabled: true},
		{TypeCode: models.DictTypeOrderTiming, Code: "START_MID", Name: "首+中", SortOrder: 3, IsEnabled: true},
		{TypeCode: models.DictTypeOrderTiming, Code: "START_END", Name: "首+末", SortOrder: 4, IsEnabled: true},
		{TypeCode: models.DictTypeOrderTiming, Code: "START", Name: "首", SortOrder: 5, IsEnabled: true},
		{TypeCode: models.DictTypeOrderTiming, Code: "MID", Name: "中", SortOrder: 6, IsEnabled: true},
		{TypeCode: models.DictTypeOrderTiming, Code: "END", Name: "末", SortOrder: 7, IsEnabled: true},
	}

	for _, di := range dictItems {
		// 幂等操作：不存在则创建，存在则更新
		var existing models.DictItem
		s.db.Where(models.DictItem{TypeCode: di.TypeCode, Code: di.Code}).
			Assign(models.DictItem{
				Name:        di.Name,
				Description: di.Description,
				SortOrder:   di.SortOrder,
				IsEnabled:   di.IsEnabled,
				Extra:       di.Extra,
				ParentCode:  di.ParentCode,
			}).
			FirstOrCreate(&existing)
	}

	return nil
}

// InitClinicalDicts 初始化临床诊疗分类字典数据
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
//     ├── TRANSFER_OUT: 转外院（二级，parent_code=20）
//     ├── TRANSPLANT: 转肾移植（二级，parent_code=20）
//     ├── PD_TRANSFER: 转腹透（二级，parent_code=20）
//     ├── CURED: 病愈（二级，parent_code=20）
//     ├── DEATH: 死亡（二级，parent_code=20）
//     └── QUIT: 退出（二级，parent_code=20）
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
