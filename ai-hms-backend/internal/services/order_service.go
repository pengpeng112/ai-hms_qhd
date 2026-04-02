package services

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var activeOrderStatuses = []string{models.OrderStatusPending, models.OrderStatusExecuting}

// OrderServiceError 医嘱域业务错误
type OrderServiceError struct {
	Status  int
	Message string
}

func (e *OrderServiceError) Error() string {
	return e.Message
}

func badOrderRequest(message string) error {
	return &OrderServiceError{Status: http.StatusBadRequest, Message: message}
}

func notFoundOrder(message string) error {
	return &OrderServiceError{Status: http.StatusNotFound, Message: message}
}

func conflictOrder(message string) error {
	return &OrderServiceError{Status: http.StatusConflict, Message: message}
}

// OrderService 医嘱服务
type OrderService struct {
	db *gorm.DB
}

// NewOrderService 创建医嘱服务
func NewOrderService() *OrderService {
	return &OrderService{db: database.GetDB()}
}

// OrderListRequest 医嘱列表请求
type OrderListRequest struct {
	PatientID      string `form:"-"`
	Type           string `form:"type"`
	Statuses       string `form:"statuses"`
	IncludeExpired bool   `form:"includeExpired"`
}

// OrderCreateRequest 创建医嘱请求
type OrderCreateRequest struct {
	Type       string  `json:"type" binding:"required"`
	Category   string  `json:"category"`
	Name       string  `json:"name"`
	Content    string  `json:"content"`
	Dose       string  `json:"dose"`
	Unit       string  `json:"unit"`
	Route      string  `json:"route"`
	Timing     string  `json:"timing"`
	ExecTiming string  `json:"execTiming"`
	DrugID     *uint   `json:"drugId"`
	Spec       string  `json:"spec"`
	GroupID    *string `json:"groupId"`
	Frequency  *string `json:"frequency"`
	Priority   string  `json:"priority"`
	StartTime  *string `json:"startTime"`
	EndTime    *string `json:"endTime"`
	Notes      string  `json:"notes"`
}

// OrderUpdateRequest 基础更新请求
type OrderUpdateRequest struct {
	Category   *string `json:"category"`
	Name       *string `json:"name"`
	Content    *string `json:"content"`
	Dose       *string `json:"dose"`
	Unit       *string `json:"unit"`
	Route      *string `json:"route"`
	Timing     *string `json:"timing"`
	ExecTiming *string `json:"execTiming"`
	DrugID     *uint   `json:"drugId"`
	Spec       *string `json:"spec"`
	GroupID    *string `json:"groupId"`
	Frequency  *string `json:"frequency"`
	Priority   *string `json:"priority"`
	StartTime  *string `json:"startTime"`
	EndTime    *string `json:"endTime"`
	Notes      *string `json:"notes"`
}

// OrderReviseRequest 在用医嘱修订请求
type OrderReviseRequest struct {
	Category   *string `json:"category"`
	Name       *string `json:"name"`
	Content    *string `json:"content"`
	Dose       *string `json:"dose"`
	Unit       *string `json:"unit"`
	Route      *string `json:"route"`
	Timing     *string `json:"timing"`
	ExecTiming *string `json:"execTiming"`
	DrugID     *uint   `json:"drugId"`
	Spec       *string `json:"spec"`
	Frequency  *string `json:"frequency"`
	Priority   *string `json:"priority"`
	StartTime  *string `json:"startTime"`
	StopDate   *string `json:"stopDate"`
	Notes      *string `json:"notes"`
}

// OrderStopRequest 停用医嘱请求
type OrderStopRequest struct {
	StopReason string  `json:"stopReason"`
	StopDate   *string `json:"stopDate"`
}

// OrderGroupRequest 组合医嘱请求
type OrderGroupRequest struct {
	OrderIDs []string `json:"orderIds" binding:"required"`
}

// OrderUngroupRequest 取消组合请求
type OrderUngroupRequest struct {
	OrderIDs []string `json:"orderIds" binding:"required"`
}

// CreateFromTemplateItemRequest 模板条目覆写请求
type CreateFromTemplateItemRequest struct {
	TemplateItemID string  `json:"templateItemId" binding:"required"`
	Name           *string `json:"name"`
	Content        *string `json:"content"`
	Dose           *string `json:"dose"`
	Unit           *string `json:"unit"`
	Route          *string `json:"route"`
	Frequency      *string `json:"frequency"`
	Timing         *string `json:"timing"`
	ExecTiming     *string `json:"execTiming"`
	Spec           *string `json:"spec"`
}

// CreateFromTemplateRequest 从模板创建医嘱请求
type CreateFromTemplateRequest struct {
	TemplateID string                          `json:"templateId" binding:"required"`
	Type       string                          `json:"type"`
	Items      []CreateFromTemplateItemRequest `json:"items" binding:"required"`
}

func (s *OrderService) List(req OrderListRequest) ([]models.Order, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	query := s.db.Where("patient_id = ?", req.PatientID)
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	if req.Statuses != "" {
		validStatuses := map[string]bool{
			models.OrderStatusPending:   true,
			models.OrderStatusExecuting: true,
			models.OrderStatusExecuted:  true,
			models.OrderStatusStopped:   true,
		}
		raw := strings.Split(req.Statuses, ",")
		statusList := make([]string, 0, len(raw))
		for _, candidate := range raw {
			candidate = strings.TrimSpace(candidate)
			if candidate != "" && validStatuses[candidate] {
				statusList = append(statusList, candidate)
			}
		}
		if len(statusList) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where("status IN ?", statusList)
		}
	} else if !req.IncludeExpired {
		query = query.Where(
			"NOT (end_time IS NOT NULL AND end_time < ? AND status IN ?)",
			time.Now(),
			activeOrderStatuses,
		)
	}

	var orders []models.Order
	if err := query.Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) Create(patientID, doctorID, doctorName string, req OrderCreateRequest) (*models.Order, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if err := validateOrderContent(req.Name, req.Content); err != nil {
		return nil, err
	}

	startAt, err := parseOptionalOrderTime(req.StartTime)
	if err != nil {
		return nil, badOrderRequest("开始时间格式错误，请使用 YYYY-MM-DD")
	}
	endAt, err := parseOptionalStopDate(req.EndTime)
	if err != nil {
		return nil, badOrderRequest("停用日期格式错误，请使用 YYYY-MM-DD")
	}

	order := models.Order{
		ID:         uuid.New().String(),
		PatientID:  patientID,
		Type:       req.Type,
		Category:   req.Category,
		Name:       strings.TrimSpace(req.Name),
		Content:    resolveContent(req.Name, req.Content),
		Dose:       strings.TrimSpace(req.Dose),
		Unit:       strings.TrimSpace(req.Unit),
		Route:      strings.TrimSpace(req.Route),
		Timing:     strings.TrimSpace(req.Timing),
		ExecTiming: strings.TrimSpace(req.ExecTiming),
		DrugID:     req.DrugID,
		Spec:       strings.TrimSpace(req.Spec),
		GroupID:    req.GroupID,
		DoctorID:   doctorID,
		DoctorName: doctorName,
		Status:     models.OrderStatusPending,
		StartTime:  startAt,
		EndTime:    endAt,
		Frequency:  trimStringPtr(req.Frequency),
		Priority:   defaultOrderPriority(req.Priority),
		Notes:      strings.TrimSpace(req.Notes),
	}

	sanitizeOrderByType(&order)

	if err := s.db.Create(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// Update 保留基础更新能力，不承载医生侧“停旧建新”语义
func (s *OrderService) Update(patientID, orderID string, req OrderUpdateRequest) (*models.Order, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	order, err := s.getOrder(patientID, orderID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Category != nil {
		updates["category"] = strings.TrimSpace(*req.Category)
	}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Content != nil {
		updates["content"] = strings.TrimSpace(*req.Content)
	}
	if req.Dose != nil {
		updates["dose"] = strings.TrimSpace(*req.Dose)
	}
	if req.Unit != nil {
		updates["unit"] = strings.TrimSpace(*req.Unit)
	}
	if req.Route != nil {
		updates["route"] = strings.TrimSpace(*req.Route)
	}
	if req.Timing != nil {
		updates["timing"] = strings.TrimSpace(*req.Timing)
	}
	if req.ExecTiming != nil {
		updates["exec_timing"] = strings.TrimSpace(*req.ExecTiming)
	}
	if req.DrugID != nil {
		updates["drug_id"] = *req.DrugID
	}
	if req.Spec != nil {
		updates["spec"] = strings.TrimSpace(*req.Spec)
	}
	if req.GroupID != nil {
		updates["group_id"] = *req.GroupID
	}
	if req.Frequency != nil {
		updates["frequency"] = strings.TrimSpace(*req.Frequency)
	}
	if req.Priority != nil {
		updates["priority"] = defaultOrderPriority(*req.Priority)
	}
	if req.StartTime != nil {
		startAt, err := parseOptionalOrderTime(req.StartTime)
		if err != nil {
			return nil, badOrderRequest("开始时间格式错误，请使用 YYYY-MM-DD")
		}
		updates["start_time"] = startAt
	}
	if req.EndTime != nil {
		endAt, err := parseOptionalStopDate(req.EndTime)
		if err != nil {
			return nil, badOrderRequest("停用日期格式错误，请使用 YYYY-MM-DD")
		}
		updates["end_time"] = endAt
	}
	if req.Notes != nil {
		updates["notes"] = strings.TrimSpace(*req.Notes)
	}

	if req.Name != nil || req.Content != nil {
		nextName := order.Name
		nextContent := order.Content
		if req.Name != nil {
			nextName = strings.TrimSpace(*req.Name)
		}
		if req.Content != nil {
			nextContent = strings.TrimSpace(*req.Content)
		}
		if err := validateOrderContent(nextName, nextContent); err != nil {
			return nil, err
		}
		updates["content"] = resolveContent(nextName, nextContent)
	}

	if len(updates) > 0 {
		if err := s.db.Model(&models.Order{}).Where("id = ? AND patient_id = ?", orderID, patientID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	var refreshed models.Order
	if err := s.db.First(&refreshed, "id = ? AND patient_id = ?", orderID, patientID).Error; err != nil {
		return nil, err
	}
	sanitizeOrderByType(&refreshed)
	return &refreshed, nil
}

// Group 将多条在用同类型医嘱编组
func (s *OrderService) Group(patientID string, orderIDs []string) ([]models.Order, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	normalizedIDs := dedupeNonEmpty(orderIDs)
	if len(normalizedIDs) < 2 {
		return nil, badOrderRequest("至少选择两条在用医嘱才能组合")
	}

	var orders []models.Order
	if err := s.db.Where("patient_id = ? AND id IN ?", patientID, normalizedIDs).Order("created_at ASC").Find(&orders).Error; err != nil {
		return nil, err
	}
	if len(orders) != len(normalizedIDs) {
		return nil, notFoundOrder("部分医嘱不存在")
	}

	firstType := orders[0].Type
	for _, order := range orders {
		if !isActiveOrderStatus(order.Status) {
			return nil, conflictOrder("只能组合在用医嘱")
		}
		if order.Type != firstType {
			return nil, badOrderRequest("只能组合同类型医嘱")
		}
	}

	groupID := uuid.New().String()
	if err := s.db.Model(&models.Order{}).
		Where("patient_id = ? AND id IN ?", patientID, normalizedIDs).
		Update("group_id", groupID).Error; err != nil {
		return nil, err
	}
	return s.listOrdersByIDs(patientID, normalizedIDs)
}

// Ungroup 取消组合
func (s *OrderService) Ungroup(patientID string, orderIDs []string) ([]models.Order, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	normalizedIDs := dedupeNonEmpty(orderIDs)
	if len(normalizedIDs) == 0 {
		return nil, badOrderRequest("请至少选择一条医嘱")
	}

	var orders []models.Order
	if err := s.db.Where("patient_id = ? AND id IN ?", patientID, normalizedIDs).Find(&orders).Error; err != nil {
		return nil, err
	}
	if len(orders) != len(normalizedIDs) {
		return nil, notFoundOrder("部分医嘱不存在")
	}
	for _, order := range orders {
		if !isActiveOrderStatus(order.Status) {
			return nil, conflictOrder("只能取消在用医嘱的组合")
		}
	}

	if err := s.db.Model(&models.Order{}).
		Where("patient_id = ? AND id IN ?", patientID, normalizedIDs).
		Update("group_id", nil).Error; err != nil {
		return nil, err
	}
	return s.listOrdersByIDs(patientID, normalizedIDs)
}

// Revise 承载医生侧“修改在用医嘱”语义
func (s *OrderService) Revise(patientID, orderID, doctorID, doctorName string, req OrderReviseRequest) (*models.Order, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	order, err := s.getOrder(patientID, orderID)
	if err != nil {
		return nil, err
	}
	if !isActiveOrderStatus(order.Status) {
		return nil, conflictOrder("已停用医嘱请使用复制功能")
	}

	hasNonLinked := hasNonLinkedChanges(req)
	hasLinked := hasLinkedChanges(order.Type, req)
	if !hasNonLinked && !hasLinked {
		return order, nil
	}

	if hasNonLinked {
		return s.reviseWithReplacement(patientID, order, doctorID, doctorName, req)
	}
	return s.reviseLinkedOnly(patientID, order, req)
}

// Copy 复制已停用医嘱
func (s *OrderService) Copy(patientID, orderID, doctorID, doctorName string) (*models.Order, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	order, err := s.getOrder(patientID, orderID)
	if err != nil {
		return nil, err
	}
	if isActiveOrderStatus(order.Status) {
		return nil, conflictOrder("在用医嘱不需要复制")
	}

	copied := *order
	copied.ID = uuid.New().String()
	copied.DoctorID = doctorID
	copied.DoctorName = doctorName
	copied.Status = models.OrderStatusPending
	copied.StartTime = time.Now()
	copied.EndTime = nil
	copied.ExecutedAt = nil
	copied.ExecutedBy = nil
	copied.StopReason = nil
	copied.GroupID = nil
	copied.CreatedAt = time.Time{}
	copied.UpdatedAt = time.Time{}

	if err := s.db.Create(&copied).Error; err != nil {
		return nil, err
	}
	return &copied, nil
}

// Stop 停用医嘱，若属于组合则整组停用
func (s *OrderService) Stop(patientID, orderID string, stopReason string, stopDate *string) ([]models.Order, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	order, err := s.getOrder(patientID, orderID)
	if err != nil {
		return nil, err
	}
	stopAt, err := parseStopDateWithDefault(stopDate, time.Now())
	if err != nil {
		return nil, badOrderRequest("停用日期格式错误，请使用 YYYY-MM-DD")
	}

	scope, err := s.stopScope(patientID, *order)
	if err != nil {
		return nil, err
	}
	if len(scope) == 0 {
		return []models.Order{*order}, nil
	}

	orderIDs := make([]string, 0, len(scope))
	for _, item := range scope {
		orderIDs = append(orderIDs, item.ID)
	}

	updates := map[string]interface{}{
		"status":      models.OrderStatusStopped,
		"end_time":    *stopAt,
		"stop_reason": stopReason,
	}

	if err := s.db.Model(&models.Order{}).
		Where("patient_id = ? AND id IN ?", patientID, orderIDs).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.listOrdersByIDs(patientID, orderIDs)
}

// CreateFromTemplate 从模板创建医嘱，支持逐条覆写
func (s *OrderService) CreateFromTemplate(patientID, doctorID, doctorName string, req CreateFromTemplateRequest) ([]models.Order, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if len(req.Items) == 0 {
		return nil, badOrderRequest("未选择任何模板条目")
	}

	var template models.OrderTemplate
	if err := s.db.Preload("Items").First(&template, "id = ?", req.TemplateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, notFoundOrder("医嘱模板不存在")
		}
		return nil, err
	}
	if len(template.Items) == 0 {
		return nil, badOrderRequest("模板中没有医嘱条目")
	}

	orderType := strings.TrimSpace(req.Type)
	if orderType == "" {
		orderType = template.Type
	}
	switch orderType {
	case models.OrderTypeLongTerm, models.OrderTypeTemporary:
	default:
		return nil, badOrderRequest("医嘱类型无效")
	}

	itemMap := make(map[string]models.OrderTemplateItem, len(template.Items))
	for _, item := range template.Items {
		itemMap[item.ID] = item
	}

	seen := make(map[string]bool, len(req.Items))
	groupIDMap := make(map[string]string)
	now := time.Now()
	orders := make([]models.Order, 0, len(req.Items))

	for _, itemReq := range req.Items {
		itemID := strings.TrimSpace(itemReq.TemplateItemID)
		if itemID == "" {
			return nil, badOrderRequest("模板条目不能为空")
		}
		if seen[itemID] {
			continue
		}
		seen[itemID] = true

		item, ok := itemMap[itemID]
		if !ok {
			return nil, badOrderRequest("模板条目不存在")
		}

		name := firstNonEmptyString(itemReq.Name, &item.DrugName)
		content := firstNonEmptyString(itemReq.Content, &item.DrugName)
		if err := validateOrderContent(name, content); err != nil {
			return nil, err
		}

			order := models.Order{
				ID:         uuid.New().String(),
				PatientID:  patientID,
				Type:       orderType,
				Category:   template.Category,
				Name:       name,
				Content:    resolveContent(name, content),
			Dose:       firstNonEmptyString(itemReq.Dose, item.Dosage),
			Unit:       firstNonEmptyString(itemReq.Unit, item.Unit),
			Route:      firstNonEmptyString(itemReq.Route, item.Route),
			Timing:     firstNonEmptyString(itemReq.Timing, item.Timing),
			ExecTiming: trimStringPtrValue(itemReq.ExecTiming),
			DrugID:     item.DrugID,
			Spec:       firstNonEmptyString(itemReq.Spec, item.Spec),
			DoctorID:   doctorID,
			DoctorName: doctorName,
			Status:     models.OrderStatusPending,
			StartTime:  now,
			Priority:   template.Priority,
		}
		if item.GroupID != nil && strings.TrimSpace(*item.GroupID) != "" {
			oldID := *item.GroupID
			if _, exists := groupIDMap[oldID]; !exists {
				groupIDMap[oldID] = uuid.New().String()
			}
			newID := groupIDMap[oldID]
			order.GroupID = &newID
		}

			if orderType == models.OrderTypeLongTerm {
				frequency := firstNonEmptyString(itemReq.Frequency, item.Frequency)
				order.Frequency = trimStringPtr(&frequency)
			}

		sanitizeOrderByType(&order)
		orders = append(orders, order)
	}

	if len(orders) == 0 {
		return nil, badOrderRequest("未选择任何模板条目")
	}
	if err := s.db.Create(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) reviseLinkedOnly(patientID string, order *models.Order, req OrderReviseRequest) (*models.Order, error) {
	updates, err := buildLinkedUpdates(order.Type, req)
	if err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return order, nil
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	target := tx.Model(&models.Order{}).Where("patient_id = ? AND id = ?", patientID, order.ID)
	if order.GroupID != nil && strings.TrimSpace(*order.GroupID) != "" {
		target = tx.Model(&models.Order{}).
			Where("patient_id = ? AND group_id = ? AND status IN ?", patientID, *order.GroupID, activeOrderStatuses)
	}
	if err := target.Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var refreshed models.Order
	if err := tx.First(&refreshed, "patient_id = ? AND id = ?", patientID, order.ID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &refreshed, nil
}

func (s *OrderService) reviseWithReplacement(patientID string, order *models.Order, doctorID, doctorName string, req OrderReviseRequest) (*models.Order, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	oldStopAt := startOfDay(time.Now())
	revisionStopReason := "修订停用"
	if err := tx.Model(&models.Order{}).
		Where("patient_id = ? AND id = ?", patientID, order.ID).
		Updates(map[string]interface{}{
			"status":      models.OrderStatusStopped,
			"end_time":    oldStopAt,
			"stop_reason": revisionStopReason,
		}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	replacement := *order
	replacement.ID = uuid.New().String()
	replacement.DoctorID = doctorID
	replacement.DoctorName = doctorName
	replacement.Status = models.OrderStatusPending
	replacement.StartTime = time.Now()
	replacement.EndTime = nil
	replacement.ExecutedAt = nil
	replacement.ExecutedBy = nil
	replacement.StopReason = nil
	replacement.CreatedAt = time.Time{}
	replacement.UpdatedAt = time.Time{}

	if err := applyReviseRequest(&replacement, req); err != nil {
		tx.Rollback()
		return nil, err
	}
	sanitizeOrderByType(&replacement)

	if err := validateOrderContent(replacement.Name, replacement.Content); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Create(&replacement).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if order.GroupID != nil && strings.TrimSpace(*order.GroupID) != "" && hasLinkedChanges(order.Type, req) {
		updates, err := buildLinkedUpdates(order.Type, req)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if len(updates) > 0 {
			if err := tx.Model(&models.Order{}).
				Where("patient_id = ? AND group_id = ? AND status IN ? AND id <> ?", patientID, *order.GroupID, activeOrderStatuses, replacement.ID).
				Updates(updates).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &replacement, nil
}

func applyReviseRequest(order *models.Order, req OrderReviseRequest) error {
	if req.Category != nil {
		order.Category = strings.TrimSpace(*req.Category)
	}
	if req.Name != nil {
		order.Name = strings.TrimSpace(*req.Name)
	}
	if req.Content != nil {
		order.Content = strings.TrimSpace(*req.Content)
	}
	if req.Dose != nil {
		order.Dose = strings.TrimSpace(*req.Dose)
	}
	if req.Unit != nil {
		order.Unit = strings.TrimSpace(*req.Unit)
	}
	if req.Route != nil {
		order.Route = strings.TrimSpace(*req.Route)
	}
	if req.Timing != nil {
		order.Timing = strings.TrimSpace(*req.Timing)
	}
	if req.ExecTiming != nil {
		order.ExecTiming = strings.TrimSpace(*req.ExecTiming)
	}
	if req.DrugID != nil {
		order.DrugID = req.DrugID
	}
	if req.Spec != nil {
		order.Spec = strings.TrimSpace(*req.Spec)
	}
	if req.Frequency != nil {
		order.Frequency = trimStringPtr(req.Frequency)
	}
	if req.Priority != nil {
		order.Priority = defaultOrderPriority(*req.Priority)
	}
	if req.StartTime != nil {
		startAt, err := parseOptionalOrderTime(req.StartTime)
		if err != nil {
			return badOrderRequest("开始时间格式错误，请使用 YYYY-MM-DD")
		}
		order.StartTime = startAt
	}
	if req.StopDate != nil {
		stopAt, err := parseOptionalStopDate(req.StopDate)
		if err != nil {
			return badOrderRequest("停用日期格式错误，请使用 YYYY-MM-DD")
		}
		order.EndTime = stopAt
	}
	if req.Notes != nil {
		order.Notes = strings.TrimSpace(*req.Notes)
	}
	order.Content = resolveContent(order.Name, order.Content)
	return nil
}

func buildLinkedUpdates(orderType string, req OrderReviseRequest) (map[string]interface{}, error) {
	updates := map[string]interface{}{}
	if req.Route != nil {
		updates["route"] = strings.TrimSpace(*req.Route)
	}
	if req.StopDate != nil {
		stopAt, err := parseOptionalStopDate(req.StopDate)
		if err != nil {
			return nil, badOrderRequest("停用日期格式错误，请使用 YYYY-MM-DD")
		}
		updates["end_time"] = stopAt
	}
	if orderType == models.OrderTypeLongTerm {
		if req.Frequency != nil {
			updates["frequency"] = strings.TrimSpace(*req.Frequency)
		}
		if req.Timing != nil {
			updates["timing"] = strings.TrimSpace(*req.Timing)
		}
	} else {
		if req.ExecTiming != nil {
			updates["exec_timing"] = strings.TrimSpace(*req.ExecTiming)
		}
	}
	return updates, nil
}

func hasLinkedChanges(orderType string, req OrderReviseRequest) bool {
	if req.Route != nil || req.StopDate != nil {
		return true
	}
	if orderType == models.OrderTypeLongTerm {
		return req.Frequency != nil || req.Timing != nil
	}
	return req.ExecTiming != nil
}

func hasNonLinkedChanges(req OrderReviseRequest) bool {
	return req.Category != nil ||
		req.Name != nil ||
		req.Content != nil ||
		req.Dose != nil ||
		req.Unit != nil ||
		req.DrugID != nil ||
		req.Spec != nil ||
		req.Priority != nil ||
		req.StartTime != nil ||
		req.Notes != nil
}

func (s *OrderService) stopScope(patientID string, order models.Order) ([]models.Order, error) {
	if order.GroupID == nil || strings.TrimSpace(*order.GroupID) == "" {
		if !isActiveOrderStatus(order.Status) {
			return nil, nil
		}
		return []models.Order{order}, nil
	}

	var orders []models.Order
	if err := s.db.Where(
		"patient_id = ? AND group_id = ? AND status IN ?",
		patientID,
		*order.GroupID,
		activeOrderStatuses,
	).Order("created_at ASC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) getOrder(patientID, orderID string) (*models.Order, error) {
	var order models.Order
	if err := s.db.First(&order, "id = ? AND patient_id = ?", orderID, patientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, notFoundOrder("医嘱不存在")
		}
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) listOrdersByIDs(patientID string, orderIDs []string) ([]models.Order, error) {
	var orders []models.Order
	if err := s.db.Where("patient_id = ? AND id IN ?", patientID, orderIDs).Order("created_at ASC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func validateOrderContent(name, content string) error {
	if strings.TrimSpace(name) == "" && strings.TrimSpace(content) == "" {
		return badOrderRequest("医嘱项目和医嘱内容不能同时为空")
	}
	return nil
}

func resolveContent(name, content string) string {
	if strings.TrimSpace(content) != "" {
		return strings.TrimSpace(content)
	}
	return strings.TrimSpace(name)
}

func defaultOrderPriority(priority string) string {
	priority = strings.TrimSpace(priority)
	if priority == "" {
		return models.OrderPriorityNormal
	}
	return priority
}

func sanitizeOrderByType(order *models.Order) {
	if order.Type == models.OrderTypeTemporary {
		order.Timing = ""
		order.Frequency = nil
		if strings.TrimSpace(order.ExecTiming) == "" {
			order.ExecTiming = "立即执行"
		}
		return
	}
	order.ExecTiming = ""
}

func isActiveOrderStatus(status string) bool {
	return status == models.OrderStatusPending || status == models.OrderStatusExecuting
}

func parseOptionalOrderTime(input *string) (time.Time, error) {
	if input == nil || strings.TrimSpace(*input) == "" {
		return time.Now(), nil
	}
	parsed, err := parseFlexibleTime(*input)
	if err != nil {
		return time.Time{}, err
	}
	return *parsed, nil
}

func parseOptionalStopDate(input *string) (*time.Time, error) {
	if input == nil || strings.TrimSpace(*input) == "" {
		return nil, nil
	}
	parsed, err := parseFlexibleTime(*input)
	if err != nil {
		return nil, err
	}
	start := startOfDay(*parsed)
	return &start, nil
}

func parseStopDateWithDefault(input *string, fallback time.Time) (*time.Time, error) {
	if input == nil || strings.TrimSpace(*input) == "" {
		start := startOfDay(fallback)
		return &start, nil
	}
	return parseOptionalStopDate(input)
}

func parseFlexibleTime(input string) (*time.Time, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("unsupported time format: %s", input)
}

func startOfDay(t time.Time) time.Time {
	loc := t.Location()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func trimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func trimStringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func firstNonEmptyString(primary *string, fallback *string) string {
	if primary != nil && strings.TrimSpace(*primary) != "" {
		return strings.TrimSpace(*primary)
	}
	if fallback != nil {
		return strings.TrimSpace(*fallback)
	}
	return ""
}

func dedupeNonEmpty(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
