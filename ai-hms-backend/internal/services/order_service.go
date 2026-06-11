package services

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"github.com/jackc/pgx/v5/pgconn"
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

type legacyPatientOrder struct {
	ID             int64               `gorm:"column:Id"`
	TenantID       int64               `gorm:"column:TenantId"`
	PatientID      modeltypes.LegacyID `gorm:"column:PatientId"`
	OrderTPLID     int64               `gorm:"column:OrderTPLId"`
	OrderGroup     int64               `gorm:"column:OrderGroup"`
	Type           int                 `gorm:"column:Type"`
	DrugID         int64               `gorm:"column:DrugId"`
	Classification string              `gorm:"column:Classification"`
	Content        string              `gorm:"column:Content"`
	Dosage         string              `gorm:"column:Dosage"`
	UseOpportunity string              `gorm:"column:UseOpportunity"`
	UseMethod      string              `gorm:"column:UseMethod"`
	UseWay         string              `gorm:"column:UseWay"`
	Note           string              `gorm:"column:Note"`
	OperatorID     int64               `gorm:"column:OperatorId"`
	StartTime      *time.Time          `gorm:"column:StartTime"`
	EndTime        *time.Time          `gorm:"column:EndTime"`
	IsDisabled     bool                `gorm:"column:IsDisabled"`
	CreatorID      int64               `gorm:"column:CreatorId"`
	CreateTime     time.Time           `gorm:"column:CreateTime"`
	LastModifyTime time.Time           `gorm:"column:LastModifyTime"`
	PatientPlanID  int64               `gorm:"column:PatientPlanId"`
	UseNum         float64             `gorm:"column:UseNum"`
	ChargeItemID   int64               `gorm:"column:ChargeItemId"`
	AllDosage      float64             `gorm:"column:AllDosage"`
}

func (legacyPatientOrder) TableName() string { return "Order_PatientOrder" }

type legacyPatientDayOrder struct {
	ID             int64      `gorm:"column:Id"`
	PatientOrderID int64      `gorm:"column:PatientOrderId"`
	Status         int        `gorm:"column:Status"`
	TreatmentTime  *time.Time `gorm:"column:TreatmentTime"`
	LastModifyTime time.Time  `gorm:"column:LastModifyTime"`
	CreateTime     time.Time  `gorm:"column:CreateTime"`
}

func (legacyPatientDayOrder) TableName() string { return "Order_PatientDayOrder" }

type legacyCodeDictionary struct {
	Code       string `gorm:"column:Code"`
	Type       string `gorm:"column:Type"`
	Name       string `gorm:"column:Name"`
	IsDisabled bool   `gorm:"column:IsDisabled"`
}

func isLegacySchemaCompatError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "42P01" || pgErr.Code == "42703"
}

// NewOrderService 创建医嘱服务
func NewOrderService() *OrderService {
	return &OrderService{db: database.GetDB()}
}

// OrderListRequest 医嘱列表请求
type OrderListRequest struct {
	PatientID      string `form:"-"`
	TenantID       int64  `form:"-"`
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

	return s.listLegacyOrders(req)
}

func mapLegacyOrderType(raw int, endTime *time.Time) string {
	switch raw {
	case 20, 2:
		return models.OrderTypeTemporary
	case 10, 1:
		return models.OrderTypeLongTerm
	default:
		if endTime != nil {
			return models.OrderTypeTemporary
		}
		return models.OrderTypeLongTerm
	}
}

func mapLegacyDayOrderStatus(raw int, dict map[int]string) string {
	if name := strings.TrimSpace(dict[raw]); name != "" {
		if strings.Contains(name, "不执行") {
			return models.OrderStatusStopped
		}
		if strings.Contains(name, "确定") {
			return models.OrderStatusPending
		}
		switch {
		case strings.Contains(name, "作废"), strings.Contains(name, "停"), strings.Contains(name, "取消"):
			return models.OrderStatusStopped
		case strings.Contains(name, "执行"):
			return models.OrderStatusExecuted
		case strings.Contains(name, "确认"), strings.Contains(name, "核对"), strings.Contains(name, "提交"), strings.Contains(name, "草稿"):
			return models.OrderStatusPending
		}
	}
	switch {
	case raw <= 10:
		return models.OrderStatusPending
	case raw >= 50:
		return models.OrderStatusStopped
	case raw >= 40:
		return models.OrderStatusExecuted
	case raw >= 20:
		return models.OrderStatusExecuting
	default:
		return models.OrderStatusPending
	}
}

func mapLegacyOrderStatus(item legacyPatientOrder, now time.Time, dayStatus *int, dayStatusDict map[int]string) string {
	if dayStatus != nil {
		return mapLegacyDayOrderStatus(*dayStatus, dayStatusDict)
	}
	if item.IsDisabled || (item.EndTime != nil && item.EndTime.Before(now)) {
		return models.OrderStatusStopped
	}
	if item.StartTime != nil && item.StartTime.After(now) {
		return models.OrderStatusPending
	}
	return models.OrderStatusExecuting
}

func (s *OrderService) loadLatestLegacyDayOrderStatus(tenantID int64, orderIDs []int64) (map[int64]int, error) {
	result := make(map[int64]int, len(orderIDs))
	if len(orderIDs) == 0 {
		return result, nil
	}
	var rows []legacyPatientDayOrder
	err := s.db.Table(`"Order_PatientDayOrder"`).
		Select(`"Id", "PatientOrderId", "Status", "TreatmentTime", "LastModifyTime", "CreateTime"`).
		Where(`"TenantId" = ? AND "PatientOrderId" IN ?`, tenantID, orderIDs).
		Order(`"TreatmentTime" DESC NULLS LAST`).
		Order(`"LastModifyTime" DESC`).
		Order(`"Id" DESC`).
		Find(&rows).Error
	if err != nil {
		if isLegacySchemaCompatError(err) {
			return result, nil
		}
		return nil, err
	}
	for _, row := range rows {
		if row.PatientOrderID <= 0 {
			continue
		}
		if _, ok := result[row.PatientOrderID]; ok {
			continue
		}
		result[row.PatientOrderID] = row.Status
	}
	return result, nil
}

func (s *OrderService) loadLegacyDayOrderStatusDict() (map[int]string, error) {
	result := make(map[int]string)
	var rows []legacyCodeDictionary
	err := s.db.Table(`"CodeDictionary_CodeDictionarys"`).
		Select(`"Code", "Type", "Name", "IsDisabled"`).
		Where(`"Type" = ? AND COALESCE("IsDisabled", false) = false`, "PatientDayOrderStatus").
		Find(&rows).Error
	if err != nil {
		if isLegacySchemaCompatError(err) {
			return result, nil
		}
		return nil, err
	}
	for _, row := range rows {
		code, parseErr := strconv.Atoi(strings.TrimSpace(row.Code))
		if parseErr != nil {
			continue
		}
		result[code] = strings.TrimSpace(row.Name)
	}
	return result, nil
}

func buildLegacyOrderGroupID(group int64) *string {
	if group <= 0 {
		return nil
	}
	value := strconv.FormatInt(group, 10)
	return &value
}

func (s *OrderService) listLegacyOrders(req OrderListRequest) ([]models.Order, error) {
	legacyPatientID, err := parseLegacyID(req.PatientID)
	if err != nil {
		return nil, badOrderRequest("患者ID格式错误")
	}

	tenantID := req.TenantID
	if tenantID <= 0 {
		tenantID = LegacyTenantID
	}

	var rows []legacyPatientOrder
	query := s.db.Table(`"Order_PatientOrder"`).
		Where(`"PatientId" = ? AND "TenantId" = ?`, legacyPatientID, tenantID)
	if !req.IncludeExpired {
		query = query.Where(`COALESCE("IsDisabled", false) = false`)
	}
	if err := query.
		Order(`"StartTime" DESC NULLS LAST`).
		Order(`"CreateTime" DESC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	drugIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.DrugID > 0 {
			drugIDs = append(drugIDs, row.DrugID)
		}
	}
	drugCatalogMap := make(map[int64]legacyDrugCatalog)
	if len(drugIDs) > 0 {
		var drugs []legacyDrugCatalog
		if err := s.db.Where(`"Id" IN ?`, drugIDs).Find(&drugs).Error; err != nil {
			return nil, err
		}
		for _, drug := range drugs {
			drugCatalogMap[drug.ID] = drug
		}
	}

	orderIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.ID > 0 {
			orderIDs = append(orderIDs, row.ID)
		}
	}
	dayStatusMap, err := s.loadLatestLegacyDayOrderStatus(tenantID, orderIDs)
	if err != nil {
		return nil, err
	}
	dayStatusDict, err := s.loadLegacyDayOrderStatusDict()
	if err != nil {
		return nil, err
	}

	statusFilter := make(map[string]struct{})
	if strings.TrimSpace(req.Statuses) != "" {
		for _, raw := range strings.Split(req.Statuses, ",") {
			status := strings.TrimSpace(raw)
			if status != "" {
				statusFilter[status] = struct{}{}
			}
		}
	}

	now := time.Now()
	result := make([]models.Order, 0, len(rows))
	for _, row := range rows {
		orderType := mapLegacyOrderType(row.Type, row.EndTime)
		if strings.TrimSpace(req.Type) != "" && orderType != strings.TrimSpace(req.Type) {
			continue
		}

		var dayStatus *int
		if v, ok := dayStatusMap[row.ID]; ok {
			dayStatus = &v
		}
		status := mapLegacyOrderStatus(row, now, dayStatus, dayStatusDict)
		if len(statusFilter) > 0 {
			if _, ok := statusFilter[status]; !ok {
				continue
			}
		}

		drug := drugCatalogMap[row.DrugID]
		operatorName, lookupErr := (&PrescriptionService{db: s.db}).lookupLegacyUserDisplayName(row.OperatorID)
		if lookupErr != nil {
			return nil, lookupErr
		}

		startTime := row.CreateTime
		if row.StartTime != nil && !row.StartTime.IsZero() {
			startTime = *row.StartTime
		}
		var endTime *time.Time
		if row.EndTime != nil && !row.EndTime.IsZero() {
			endTime = row.EndTime
		}

		result = append(result, models.Order{
			ID:         strconv.FormatInt(row.ID, 10),
			TenantID:   row.TenantID,
			PatientID:  row.PatientID,
			Type:       orderType,
			Category:   strings.TrimSpace(row.Classification),
			Name:       firstNonEmptyText(strings.TrimSpace(drug.Name), strings.TrimSpace(row.Content)),
			Content:    strings.TrimSpace(row.Content),
			Dose:       strings.TrimSpace(row.Dosage),
			Unit:       strings.TrimSpace(drug.BasicUnit),
			Route:      firstNonEmptyText(strings.TrimSpace(row.UseWay), strings.TrimSpace(row.UseMethod)),
			Timing:     strings.TrimSpace(row.UseOpportunity),
			ExecTiming: "",
			Spec:       strings.TrimSpace(drug.Specification),
			GroupID:    buildLegacyOrderGroupID(row.OrderGroup),
			DoctorID:   strconv.FormatInt(row.OperatorID, 10),
			DoctorName: operatorName,
			Status:     status,
			StartTime:  startTime,
			EndTime:    endTime,
			Frequency:  nil,
			Priority:   models.OrderPriorityNormal,
			Notes:      strings.TrimSpace(row.Note),
			CreatedAt:  row.CreateTime,
			UpdatedAt:  row.LastModifyTime,
		})
	}

	return result, nil
}

func (s *OrderService) Create(patientID string, tenantID int64, doctorID, doctorName string, req OrderCreateRequest) (*models.Order, error) {
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
	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, badOrderRequest("患者ID格式错误")
	}

	if tenantID <= 0 {
		tenantID = LegacyTenantID
	}
	operatorID, _, resolveErr := (&PrescriptionService{db: s.db}).resolveLegacyUserID(doctorID, doctorName)
	if resolveErr != nil {
		return nil, resolveErr
	}
	legacyType := mapOrderTypeToLegacy(req.Type)
	orderGroup := int64(0)
	if req.GroupID != nil && strings.TrimSpace(*req.GroupID) != "" {
		if parsed, parseErr := strconv.ParseInt(strings.TrimSpace(*req.GroupID), 10, 64); parseErr == nil && parsed > 0 {
			orderGroup = parsed
		}
	}
	patientPlanID, planErr := s.resolveLatestPatientPlanID(int64(legacyPatientID), tenantID)
	if planErr != nil {
		return nil, planErr
	}
	content := resolveContent(req.Name, req.Content)
	now := time.Now()

	var orderIDInt int64
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		orderID, idErr := nextLegacyID()
		if idErr != nil {
			return idErr
		}
		createRow := map[string]any{
			"Id":             orderID,
			"TenantId":       tenantID,
			"PatientId":      legacyPatientID,
			"OrderTPLId":     int64(0),
			"OrderGroup":     orderGroup,
			"Type":           legacyType,
			"DrugId":         int64(0),
			"Classification": strings.TrimSpace(req.Category),
			"Content":        strings.TrimSpace(content),
			"Dosage":         strings.TrimSpace(req.Dose),
			"UseOpportunity": strings.TrimSpace(req.Timing),
			"UseMethod":      strings.TrimSpace(req.Route),
			"UseWay":         strings.TrimSpace(req.Route),
			"Note":           strings.TrimSpace(req.Notes),
			"OperatorId":     operatorID,
			"StartTime":      startAt,
			"EndTime":        endAt,
			"IsDisabled":     false,
			"CreatorId":      operatorID,
			"CreateTime":     now,
			"LastModifyTime": now,
			"PatientPlanId":  patientPlanID,
			"UseNum":         1.0,
			"ChargeItemId":   int64(0),
		}
		if err := tx.Table(`"Order_PatientOrder"`).Create(createRow).Error; err != nil {
			return err
		}

		dayOrderID, dayIDErr := nextLegacyID()
		if dayIDErr != nil {
			return dayIDErr
		}
		dayRow := map[string]any{
			"Id":              dayOrderID,
			"TenantId":        tenantID,
			"PatientId":       legacyPatientID,
			"TreatmentTime":   startAt,
			"PatientOrderId":  orderID,
			"OrderGroup":      orderGroup,
			"Status":          20,
			"CaseStatus":      "",
			"Classification":  strings.TrimSpace(req.Category),
			"DrugId":          int64(0),
			"Content":         strings.TrimSpace(content),
			"Dosage":          strings.TrimSpace(req.Dose),
			"UseOpportunity":  strings.TrimSpace(req.Timing),
			"UseMethod":       strings.TrimSpace(req.Route),
			"UseWay":          strings.TrimSpace(req.Route),
			"Note":            strings.TrimSpace(req.Notes),
			"OperatorId":      operatorID,
			"CreatorId":       operatorID,
			"CreateTime":      now,
			"LastModifyTime":  now,
			"UseNum":          1.0,
			"DealOpportunity": mapDealOpportunityCode(req.ExecTiming, req.Type),
			"ChargeItemId":    int64(0),
		}
		if err := tx.Table(`"Order_PatientDayOrder"`).Create(dayRow).Error; err != nil {
			return err
		}
		orderIDInt = orderID.Int64()
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	created, listErr := s.listLegacyOrders(OrderListRequest{
		PatientID:      patientID,
		TenantID:       tenantID,
		IncludeExpired: true,
	})
	if listErr != nil {
		return nil, listErr
	}
	orderIDText := strconv.FormatInt(orderIDInt, 10)
	for i := range created {
		if created[i].ID == orderIDText {
			return &created[i], nil
		}
	}
	return nil, errors.New("order created but not found")
}

func mapOrderTypeToLegacy(value string) int {
	switch strings.TrimSpace(value) {
	case models.OrderTypeTemporary, "20", "2":
		return 20
	default:
		return 10
	}
}

func mapDealOpportunityCode(execTiming string, orderType string) string {
	text := strings.TrimSpace(execTiming)
	if strings.Contains(text, "立即") || strings.EqualFold(text, "immediate") {
		return "10"
	}
	if strings.Contains(text, "普通") || strings.EqualFold(text, "normal") {
		return "20"
	}
	if strings.TrimSpace(orderType) == models.OrderTypeTemporary {
		return "10"
	}
	return "20"
}

func (s *OrderService) resolveLatestPatientPlanID(patientID, tenantID int64) (int64, error) {
	var row struct {
		ID int64 `gorm:"column:Id"`
	}
	err := s.db.Table(`"Plan_PatientPlan"`).
		Select(`"Id"`).
		Where(`"TenantId" = ? AND "PatientId" = ?`, tenantID, patientID).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Order(`"Id" DESC`).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return row.ID, nil
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

	updates := map[string]interface{}{
		"LastModifyTime": time.Now(),
	}
	if req.Category != nil {
		updates["Classification"] = strings.TrimSpace(*req.Category)
	}
	if req.Content != nil {
		updates["Content"] = strings.TrimSpace(*req.Content)
	}
	if req.Dose != nil {
		updates["Dosage"] = strings.TrimSpace(*req.Dose)
	}
	if req.Route != nil {
		v := strings.TrimSpace(*req.Route)
		updates["UseMethod"] = v
		updates["UseWay"] = v
	}
	if req.Timing != nil {
		updates["UseOpportunity"] = strings.TrimSpace(*req.Timing)
	}
	if req.DrugID != nil {
		updates["DrugId"] = int64(*req.DrugID)
	}
	if req.StartTime != nil {
		startAt, err := parseOptionalOrderTime(req.StartTime)
		if err != nil {
			return nil, badOrderRequest("开始时间格式错误，请使用 YYYY-MM-DD")
		}
		updates["StartTime"] = startAt
	}
	if req.EndTime != nil {
		endAt, err := parseOptionalStopDate(req.EndTime)
		if err != nil {
			return nil, badOrderRequest("停用日期格式错误，请使用 YYYY-MM-DD")
		}
		updates["EndTime"] = endAt
	}
	if req.Notes != nil {
		updates["Note"] = strings.TrimSpace(*req.Notes)
	}
	if req.GroupID != nil {
		if parsed, parseErr := strconv.ParseInt(strings.TrimSpace(*req.GroupID), 10, 64); parseErr == nil {
			updates["OrderGroup"] = parsed
		}
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
		updates["Content"] = resolveContent(nextName, nextContent)
	}

	legacyPID, _ := parseLegacyID(patientID)
	oid, _ := strconv.ParseInt(orderID, 10, 64)
	if err := s.db.Table(`"Order_PatientOrder"`).
		Where(`"Id" = ? AND "PatientId" = ? AND "TenantId" = ?`, oid, legacyPID, order.TenantID).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	return s.getOrder(patientID, orderID)
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

	orders, err := s.listOrdersByIDs(patientID, normalizedIDs)
	if err != nil {
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

	orderGroup := time.Now().UnixNano() / 1000000
	intIDs := orderIDsToInt64(normalizedIDs)
	if err := s.db.Table(`"Order_PatientOrder"`).
		Where(`"Id" IN ?`, intIDs).
		Update("OrderGroup", orderGroup).Error; err != nil {
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

	orders, err := s.listOrdersByIDs(patientID, normalizedIDs)
	if err != nil {
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

	intIDs := orderIDsToInt64(normalizedIDs)
	if err := s.db.Table(`"Order_PatientOrder"`).
		Where(`"Id" IN ?`, intIDs).
		Update("OrderGroup", 0).Error; err != nil {
		return nil, err
	}
	return s.listOrdersByIDs(patientID, normalizedIDs)
}

func orderIDsToInt64(ids []string) []int64 {
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if parsed, err := strconv.ParseInt(id, 10, 64); err == nil {
			result = append(result, parsed)
		}
	}
	return result
}

// Revise 承载医生侧“修改在用医嘱”语义
func (s *OrderService) Revise(patientID, orderID string, tenantID int64, doctorID, doctorName string, req OrderReviseRequest) (*models.Order, error) {
	return nil, conflictOrder("老库医嘱暂不支持修订功能")
}

// Copy 老库医嘱暂不支持复制
func (s *OrderService) Copy(patientID, orderID string, tenantID int64, doctorID, doctorName string) (*models.Order, error) {
	return nil, conflictOrder("老库医嘱暂不支持复制功能")
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
		"EndTime":        *stopAt,
		"LastModifyTime": time.Now(),
	}
	// 未来停用日期仅写 EndTime 和停嘱原因，不立即置 IsDisabled=true。
	// 当天或过去的停用日期才立即停用。
	today := time.Now()
	if stopAt.After(today) {
		updates["IsDisabled"] = false
	} else {
		updates["IsDisabled"] = true
	}
	// 修复：医生填写的停嘱原因此前被静默丢弃。老库无专用停嘱原因列
	//（Order_PatientOrder 仅有 Note varchar(1024)），故以追加形式写入 Note，
	// 不覆盖既有备注；超长时截断原因，保证总长不越列宽。
	if reason := strings.TrimSpace(stopReason); reason != "" {
		const maxReasonLen = 200
		if utf8.RuneCountInString(reason) > maxReasonLen {
			reason = string([]rune(reason)[:maxReasonLen])
		}
		suffix := "\n[停嘱原因 " + time.Now().Format("2006-01-02 15:04") + "] " + reason
		updates["Note"] = gorm.Expr(`LEFT(COALESCE("Note", '') || ?, 1024)`, suffix)
	}

	intIDs := make([]int64, 0, len(orderIDs))
	for _, id := range orderIDs {
		if oid, parseErr := strconv.ParseInt(id, 10, 64); parseErr == nil {
			intIDs = append(intIDs, oid)
		}
	}
	// 补 TenantId 过滤：与全库写操作惯例一致（纵深防御，归属已由 getOrder 校验）。
	if err := s.db.Table(`"Order_PatientOrder"`).
		Where(`"Id" IN ? AND "TenantId" = ?`, intIDs, LegacyTenantID).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.listOrdersByIDs(patientID, orderIDs)
}

// CreateFromTemplate 从模板创建医嘱 — 老库暂不支持
func (s *OrderService) CreateFromTemplate(patientID string, tenantID int64, doctorID, doctorName string, req CreateFromTemplateRequest) ([]models.Order, error) {
	return nil, conflictOrder("老库医嘱暂不支持从模板创建")
}

func (s *OrderService) stopScope(patientID string, order models.Order) ([]models.Order, error) {
	return []models.Order{order}, nil
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

func (s *OrderService) getOrder(patientID, orderID string) (*models.Order, error) {
	legacyPID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, badOrderRequest("患者ID格式错误")
	}
	oid, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return nil, badOrderRequest("医嘱ID格式错误")
	}
	var row legacyPatientOrder
	if err := s.db.Table(`"Order_PatientOrder"`).
		Where(`"Id" = ? AND "PatientId" = ? AND "TenantId" = ?`, oid, legacyPID, LegacyTenantID).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, notFoundOrder("医嘱不存在")
		}
		return nil, err
	}
	order := legacyOrderToDTO(row)
	return &order, nil
}

func legacyOrderToDTO(row legacyPatientOrder) models.Order {
	order := models.Order{
		ID:        strconv.FormatInt(row.ID, 10),
		TenantID:  row.TenantID,
		PatientID: modeltypes.LegacyID(row.PatientID),
		Type:      legacyTypeToString(row.Type),
		Category:  row.Classification,
		Content:   row.Content,
		Dose:      row.Dosage,
		Route:     row.UseMethod,
		Timing:    row.UseOpportunity,
		Status:    legacyDisabledToStatus(row.IsDisabled),
		Priority:  models.OrderPriorityNormal,
		Notes:     row.Note,
		CreatedAt: row.CreateTime,
		UpdatedAt: row.LastModifyTime,
	}
	if row.StartTime != nil {
		order.StartTime = *row.StartTime
	}
	order.EndTime = row.EndTime
	return order
}

func legacyTypeToString(t int) string {
	if t == 20 {
		return models.OrderTypeTemporary
	}
	return models.OrderTypeLongTerm
}

func legacyDisabledToStatus(disabled bool) string {
	if disabled {
		return models.OrderStatusStopped
	}
	return models.OrderStatusPending
}

func (s *OrderService) listOrdersByIDs(patientID string, orderIDs []string) ([]models.Order, error) {
	var rows []legacyPatientOrder
	if err := s.db.Table(`"Order_PatientOrder"`).
		Where(`"Id" IN ?`, orderIDs).
		Order(`"Id" ASC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]models.Order, 0, len(rows))
	for _, row := range rows {
		o := legacyOrderToDTO(row)
		result = append(result, o)
	}
	return result, nil
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
