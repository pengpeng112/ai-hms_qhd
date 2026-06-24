package services

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

type BillingService struct {
	db          *gorm.DB
	hisPriceSvc *HisPriceService
	tenantID    int64
}

func NewBillingService(tenantID int64) *BillingService {
	db := database.GetDB()
	svc := &BillingService{
		db:          db,
		hisPriceSvc: NewHisPriceService(),
		tenantID:    tenantID,
	}
	return svc
}

type BuildDraftRequest struct {
	TreatmentID    int64    `json:"treatmentId"`
	PatientID      *int64   `json:"patientId"`
	PrescriptionID *int64   `json:"prescriptionId"`
	DialysisMode   *string  `json:"dialysisMode"`
	AccessType     *string  `json:"accessType"`
	Shift          *string  `json:"shift"`
	CrrtHours      *float64 `json:"crrtHours"`
}

type ChargeRecordResponse struct {
	*models.ChargeRecord
	Lines []models.ChargeLine `json:"lines,omitempty"`
}

func (s *BillingService) BuildDraft(req BuildDraftRequest, userID string, userName string) (*models.ChargeRecord, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}
	if req.TreatmentID <= 0 {
		return nil, fmt.Errorf("治疗记录ID不能为空")
	}

	var existing models.ChargeRecord
	err := s.db.Where("tenant_id = ? AND treatment_id = ? AND status <> ?",
		s.tenantID, req.TreatmentID, models.ChargeStatusCancelled).
		First(&existing).Error
	if err == nil {
		if err2 := s.loadLines(&existing); err2 != nil {
			return nil, err2
		}
		total := s.recalcTotal(existing.Lines)
		existing.TotalAmount = &total
		return &existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("查询已有清单失败: %w", err)
	}

	now := time.Now()
	rec := &models.ChargeRecord{
		ID:             utils.GenerateID(),
		TenantID:       s.tenantID,
		PatientID:      req.PatientID,
		TreatmentID:    req.TreatmentID,
		PrescriptionID: req.PrescriptionID,
		ChargeDate:     &now,
		Shift:          req.Shift,
		DialysisMode:   req.DialysisMode,
		AccessType:     req.AccessType,
		CrrtHours:      req.CrrtHours,
		Status:         models.ChargeStatusDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if rec.PatientID == nil {
		var pid int64
		err := s.db.Raw(`SELECT "PatientId" FROM "Treatment_Treatment" WHERE "Id" = ? AND "TenantId" = ?`,
			req.TreatmentID, s.tenantID).Scan(&pid).Error
		if err == nil && pid > 0 {
			rec.PatientID = &pid
		}
	}

	rec.RecordedBy = &userID
	rec.RecordedName = &userName

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(rec).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建清单失败: %w", err)
	}

	lines, err := s.buildLines(rec.ID, req)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, l := range lines {
		if err := tx.Create(l).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("创建清单明细失败: %w", err)
		}
	}

	ptrLines := make([]models.ChargeLine, len(lines))
	for i, l := range lines {
		ptrLines[i] = *l
	}
	total := s.recalcTotal(ptrLines)
	rec.TotalAmount = &total
	if err := tx.Model(rec).Update("total_amount", total).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新合计失败: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	rec.Lines = ptrLines
	return rec, nil
}

func (s *BillingService) buildLines(recordID string, req BuildDraftRequest) ([]*models.ChargeLine, error) {
	var lines []*models.ChargeLine
	now := time.Now()
	tenantID := s.tenantID

	catalog, catErr := config.LoadBillingCatalog()

	mode := "hd"
	if req.DialysisMode != nil {
		mode = strings.ToUpper(strings.TrimSpace(*req.DialysisMode))
	}

	accessType := "avf"
	if req.AccessType != nil {
		accessType = strings.ToUpper(strings.TrimSpace(*req.AccessType))
	}

	// A: 治疗费
	treatmentLine := s.buildTreatmentLine(recordID, tenantID, mode, now, catalog, catErr)
	if treatmentLine != nil {
		lines = append(lines, treatmentLine)
	}

	// B: 耗材费
	materialLines, err := s.buildMaterialLines(recordID, tenantID, req.TreatmentID, now, catalog, catErr)
	if err != nil {
		return nil, err
	}
	lines = append(lines, materialLines...)

	// C: 护理费
	if accessType != "" {
		nursingLines := s.buildNursingLines(recordID, tenantID, accessType, now, catalog, catErr)
		lines = append(lines, nursingLines...)
	}

	// D: 注射费
	if req.TreatmentID > 0 {
		hasInjection, err := s.hasIntravenousDrug(req.TreatmentID)
		if err != nil {
			return nil, err
		}
		if hasInjection {
			injectionLine := s.buildInjectionLine(recordID, tenantID, now, catalog, catErr)
			if injectionLine != nil {
				lines = append(lines, injectionLine)
			}
		}
	}

	// E: 药品项
	drugLines, err := s.buildDrugLines(recordID, tenantID, req.TreatmentID, now)
	if err != nil {
		return nil, err
	}
	lines = append(lines, drugLines...)

	return lines, nil
}

func (s *BillingService) buildTreatmentLine(recordID string, tenantID int64, mode string, now time.Time,
	catalog *config.BillingCatalog, catErr error) *models.ChargeLine {

	line := &models.ChargeLine{
		ID:             utils.GenerateID(),
		TenantID:       tenantID,
		ChargeRecordID: recordID,
		Category:       models.ChargeCatTreatment,
		Source:         models.ChargeSourceAuto,
		Billable:       true,
		CreatedAt:      now,
	}

	// 优先匹配 HIS 价表
	if hisItem := s.resolveByClassAndName("E", mode); hisItem != nil {
		line.HisPriceItemID = &hisItem.ID
		line.HisItemCode = &hisItem.ItemCode
		itemClass := "E"
		line.HisItemClass = &itemClass
		if hisItem.ItemName != nil {
			line.HisItemName = hisItem.ItemName
		}
		line.ItemName = *hisItem.ItemName
		if hisItem.Price != nil {
			line.UnitPrice = hisItem.Price
		}
		if hisItem.Units != nil {
			line.Unit = hisItem.Units
		}
		ps := models.PriceSourceHisPriceList
		line.PriceSource = &ps
		ms := models.MatchStatusMatched
		line.MatchedStatus = &ms
	} else if catErr == nil && catalog != nil {
		if tf, ok := catalog.TreatmentFeeFor(mode); ok {
			line.ItemName = tf.Name
			line.UnitPrice = &tf.Price
			unit := tf.Unit
			line.Unit = &unit
			line.ItemCode = &tf.InsuranceCode
			ps := models.PriceSourceCatalog
			line.PriceSource = &ps
		} else {
			line.ItemName = "治疗费-" + mode
			ps := models.PriceSourceUnknown
			line.PriceSource = &ps
		}
	} else {
		line.ItemName = "治疗费-" + mode
		ps := models.PriceSourceUnknown
		line.PriceSource = &ps
	}

	qty := 1.0
	line.Quantity = &qty

	if line.UnitPrice != nil {
		amt := math.Round(qty**line.UnitPrice*100) / 100
		line.Amount = &amt
	}

	return line
}

func (s *BillingService) buildMaterialLines(recordID string, tenantID int64, treatmentID int64, now time.Time,
	catalog *config.BillingCatalog, catErr error) ([]*models.ChargeLine, error) {

	type mtRow struct {
		ChargeItemID int64
		Name         string
		Num          float64
		Unit         string
	}
	var rows []mtRow
	err := s.db.Raw(`SELECT
			"ChargeItemId" AS charge_item_id,
			COALESCE(
				(SELECT MIN("Name") FROM "Stock_ChargeItem" sci WHERE sci."Id" = "Treatment_MaterialTrace"."ChargeItemId" AND sci."TenantId" = "Treatment_MaterialTrace"."TenantId"),
				''
			) AS name,
			COALESCE("Num", 0) AS num,
			COALESCE(
				(SELECT MIN("Unit") FROM "Stock_ChargeItem" sci WHERE sci."Id" = "Treatment_MaterialTrace"."ChargeItemId" AND sci."TenantId" = "Treatment_MaterialTrace"."TenantId"),
				''
			) AS unit
		FROM "Treatment_MaterialTrace"
		WHERE "TenantId" = ? AND "TreatmentId" = ?`, tenantID, treatmentID).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询治疗耗材失败: %w", err)
	}

	var lines []*models.ChargeLine
	for _, r := range rows {
		line := &models.ChargeLine{
			ID:             utils.GenerateID(),
			TenantID:       tenantID,
			ChargeRecordID: recordID,
			Category:       models.ChargeCatMaterial,
			ItemName:       r.Name,
			Source:         models.ChargeSourceAuto,
			Billable:       true,
			ChargeItemID:   &r.ChargeItemID,
			CreatedAt:      now,
		}
		qty := r.Num
		if qty <= 0 {
			qty = 1
		}
		line.Quantity = &qty
		if r.Unit != "" {
			line.Unit = &r.Unit
		}

		if catErr == nil && catalog != nil {
			if mp, ok := catalog.MaterialFor(r.ChargeItemID); ok {
				if !mp.Billable {
					line.Billable = false
					line.UnitPrice = nil
				} else if mp.UnitPrice != nil {
					line.UnitPrice = mp.UnitPrice
					ps := models.PriceSourceCatalog
					line.PriceSource = &ps
				}
			}
		}

		if line.Billable && (line.UnitPrice == nil || *line.UnitPrice == 0) {
			if hisItem := s.resolveByClassAndName("I", r.Name); hisItem != nil && hisItem.Price != nil {
				line.HisPriceItemID = &hisItem.ID
				line.HisItemCode = &hisItem.ItemCode
				itemClass := "I"
				line.HisItemClass = &itemClass
				if hisItem.ItemName != nil {
					line.HisItemName = hisItem.ItemName
				}
				line.UnitPrice = hisItem.Price
				ps := models.PriceSourceHisPriceList
				line.PriceSource = &ps
				ms := models.MatchStatusMatched
				line.MatchedStatus = &ms
			}
		}

		if line.Billable && line.UnitPrice != nil {
			amt := math.Round(qty**line.UnitPrice*100) / 100
			line.Amount = &amt
		}

		lines = append(lines, line)
	}
	return lines, nil
}

func (s *BillingService) buildNursingLines(recordID string, tenantID int64, accessType string, now time.Time,
	catalog *config.BillingCatalog, catErr error) []*models.ChargeLine {

	if catErr != nil || catalog == nil {
		return nil
	}
	fees := catalog.NursingFeeFor(accessType)
	var lines []*models.ChargeLine
	for _, f := range fees {
		line := &models.ChargeLine{
			ID:             utils.GenerateID(),
			TenantID:       tenantID,
			ChargeRecordID: recordID,
			Category:       models.ChargeCatNursing,
			ItemName:       f.Name,
			Source:         models.ChargeSourceAuto,
			Billable:       true,
			CreatedAt:      now,
		}
		qty := f.Qty
		if qty <= 0 {
			qty = 1
		}
		price := f.Price
		line.Quantity = &qty
		line.UnitPrice = &price
		line.Unit = &f.Unit
		line.ItemCode = &f.InsuranceCode
		ps := models.PriceSourceCatalog
		line.PriceSource = &ps
		amt := math.Round(qty*price*100) / 100
		line.Amount = &amt

		if hisItem := s.resolveByClassAndName("K", f.Name); hisItem != nil {
			line.HisPriceItemID = &hisItem.ID
			line.HisItemCode = &hisItem.ItemCode
			itemClass := "K"
			line.HisItemClass = &itemClass
			if hisItem.ItemName != nil {
				line.HisItemName = hisItem.ItemName
			}
		}
		lines = append(lines, line)
	}
	return lines
}

func (s *BillingService) buildInjectionLine(recordID string, tenantID int64, now time.Time,
	catalog *config.BillingCatalog, catErr error) *models.ChargeLine {

	if catErr != nil || catalog == nil {
		return nil
	}
	fee := catalog.InjectionFee
	line := &models.ChargeLine{
		ID:             utils.GenerateID(),
		TenantID:       tenantID,
		ChargeRecordID: recordID,
		Category:       models.ChargeCatInjection,
		ItemName:       fee.Name,
		Source:         models.ChargeSourceAuto,
		Billable:       true,
		CreatedAt:      now,
	}
	qty := 1.0
	price := fee.Price
	line.Quantity = &qty
	line.UnitPrice = &price
	line.Unit = &fee.Unit
	line.ItemCode = &fee.InsuranceCode
	ps := models.PriceSourceCatalog
	line.PriceSource = &ps
	amt := math.Round(qty*price*100) / 100
	line.Amount = &amt
	return line
}

func (s *BillingService) buildDrugLines(recordID string, tenantID int64, treatmentID int64, now time.Time) ([]*models.ChargeLine, error) {
	type drugRow struct {
		DrugName string
		Dose     string
		Route    string
	}
	var rows []drugRow
	err := s.db.Table("medication_admin").
		Select("drug_name, dose, route").
		Where("tenant_id = ? AND treatment_id = ?", tenantID, treatmentID).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询治疗药品失败: %w", err)
	}

	var lines []*models.ChargeLine
	for _, r := range rows {
		line := &models.ChargeLine{
			ID:             utils.GenerateID(),
			TenantID:       tenantID,
			ChargeRecordID: recordID,
			Category:       models.ChargeCatDrug,
			ItemName:       r.DrugName,
			Source:         models.ChargeSourceAuto,
			Billable:       true,
			CreatedAt:      now,
		}
		qty := 1.0
		line.Quantity = &qty
		if r.Dose != "" {
			line.Spec = &r.Dose
		}
		if r.Route != "" {
			note := r.Route + "（HIS 查价）"
			line.Note = &note
		}
		ms := models.MatchStatusUnmatched
		line.MatchedStatus = &ms

		matched := s.resolveByClassAndName("A", r.DrugName)
		if matched == nil {
			matched = s.resolveByClassAndName("B", r.DrugName)
		}
		if matched != nil {
			line.HisPriceItemID = &matched.ID
			line.HisItemCode = &matched.ItemCode
			if matched.ItemName != nil {
				line.HisItemName = matched.ItemName
			}
			if matched.ItemClass != nil {
				line.HisItemClass = matched.ItemClass
			}
			ps := models.PriceSourceHisPriceList
			line.PriceSource = &ps
			ms := models.MatchStatusMatched
			line.MatchedStatus = &ms
			if matched.Price != nil {
				line.UnitPrice = matched.Price
				amt := math.Round(qty**matched.Price*100) / 100
				line.Amount = &amt
			}
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func (s *BillingService) resolveByClassAndName(itemClass, name string) *models.HisPriceItem {
	if s.hisPriceSvc == nil || name == "" {
		return nil
	}
	items, err := s.hisPriceSvc.MatchByName(name, &itemClass, true)
	if err != nil || len(items) != 1 {
		return nil
	}
	return &items[0]
}

func (s *BillingService) hasIntravenousDrug(treatmentID int64) (bool, error) {
	var count int64
	err := s.db.Table("medication_admin").
		Where("tenant_id = ? AND treatment_id = ? AND (route LIKE ? OR route LIKE ? OR route = ?)",
			s.tenantID, treatmentID, "%静脉%", "%注射%", "iv").
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("查询静脉注射给药失败: %w", err)
	}
	return count > 0, nil
}

// ---- CRUD ----

func (s *BillingService) GetCharge(id string) (*models.ChargeRecord, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}
	var rec models.ChargeRecord
	err := s.db.Where("id = ? AND tenant_id = ?", id, s.tenantID).First(&rec).Error
	if err != nil {
		return nil, err
	}
	if err := s.loadLines(&rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *BillingService) ListCharges(params ListChargesRequest) (*ListChargesResponse, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 50
	}

	query := s.db.Model(&models.ChargeRecord{}).Where("tenant_id = ?", s.tenantID)
	if params.PatientID != nil && *params.PatientID > 0 {
		query = query.Where("patient_id = ?", *params.PatientID)
	}
	if params.Status != nil && *params.Status != "" {
		query = query.Where("status = ?", *params.Status)
	}
	if params.DateFrom != nil {
		query = query.Where("charge_date >= ?", *params.DateFrom)
	}
	if params.DateTo != nil {
		query = query.Where("charge_date <= ?", *params.DateTo)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []models.ChargeRecord
	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPage++
	}
	return &ListChargesResponse{
		Items:     items,
		Total:     total,
		Page:      params.Page,
		PageSize:  params.PageSize,
		TotalPage: totalPage,
	}, nil
}

type ListChargesRequest struct {
	PatientID *int64 `form:"patientId"`
	Status    *string `form:"status"`
	DateFrom  *string `form:"dateFrom"`
	DateTo    *string `form:"dateTo"`
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
}

type ListChargesResponse struct {
	Items     []models.ChargeRecord `json:"items"`
	Total     int64                 `json:"total"`
	Page      int                   `json:"page"`
	PageSize  int                   `json:"pageSize"`
	TotalPage int                   `json:"totalPage"`
}

func (s *BillingService) AddLine(chargeID string, line *models.ChargeLine) (*models.ChargeLine, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}
	var rec models.ChargeRecord
	if err := s.db.First(&rec, "id = ? AND tenant_id = ?", chargeID, s.tenantID).Error; err != nil {
		return nil, fmt.Errorf("清单不存在: %w", err)
	}
	if rec.Status == models.ChargeStatusChecked {
		return nil, fmt.Errorf("已核对的清单不允许编辑")
	}
	if rec.Status == models.ChargeStatusCancelled {
		return nil, fmt.Errorf("已取消的清单不允许编辑")
	}
	line.ID = utils.GenerateID()
	line.TenantID = s.tenantID
	line.ChargeRecordID = chargeID
	if line.Source == "" {
		line.Source = models.ChargeSourceManual
	}
	if line.Quantity == nil {
		qty := 1.0
		line.Quantity = &qty
	}

	if line.Billable && line.UnitPrice != nil {
		amt := math.Round(*line.Quantity**line.UnitPrice*100) / 100
		line.Amount = &amt
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if err := tx.Create(line).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	allLines, err := s.getLinesWithDB(tx, chargeID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	total := s.recalcTotal(allLines)
	if err := tx.Model(&rec).Update("total_amount", total).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return line, nil
}

func (s *BillingService) UpdateLine(lineID string, patch map[string]interface{}) (*models.ChargeLine, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}
	var line models.ChargeLine
	if err := s.db.First(&line, "id = ? AND tenant_id = ?", lineID, s.tenantID).Error; err != nil {
		return nil, err
	}

	var rec models.ChargeRecord
	if err := s.db.First(&rec, "id = ? AND tenant_id = ?", line.ChargeRecordID, s.tenantID).Error; err != nil {
		return nil, err
	}
	if rec.Status == models.ChargeStatusChecked || rec.Status == models.ChargeStatusCancelled {
		return nil, fmt.Errorf("清单状态不允许编辑明细")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if err := tx.Model(&line).Updates(patch).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if qtyVal, ok := patch["quantity"]; ok {
		line.Quantity = toFloatPtr(qtyVal)
	}
	if priceVal, ok := patch["unit_price"]; ok {
		line.UnitPrice = toFloatPtr(priceVal)
	}
	if line.Billable && line.Quantity != nil && line.UnitPrice != nil {
		amt := math.Round(*line.Quantity**line.UnitPrice*100) / 100
		line.Amount = &amt
		if err := tx.Model(&line).Update("amount", line.Amount).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	allLines, err := s.getLinesWithDB(tx, line.ChargeRecordID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	total := s.recalcTotal(allLines)
	if err := tx.Model(&rec).Update("total_amount", total).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if err := s.db.First(&line, "id = ?", lineID).Error; err != nil {
		return nil, err
	}
	return &line, nil
}

func toFloatPtr(v interface{}) *float64 {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case float64:
		return &val
	case *float64:
		return val
	case int:
		f := float64(val)
		return &f
	case float32:
		f := float64(val)
		return &f
	default:
		return nil
	}
}

func (s *BillingService) DeleteLine(lineID string) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}
	var line models.ChargeLine
	if err := s.db.First(&line, "id = ? AND tenant_id = ?", lineID, s.tenantID).Error; err != nil {
		return err
	}
	var rec models.ChargeRecord
	if err := s.db.First(&rec, "id = ? AND tenant_id = ?", line.ChargeRecordID, s.tenantID).Error; err != nil {
		return err
	}
	if rec.Status == models.ChargeStatusChecked || rec.Status == models.ChargeStatusCancelled {
		return fmt.Errorf("清单状态不允许删除明细")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Delete(&line).Error; err != nil {
		tx.Rollback()
		return err
	}

	allLines, err := s.getLinesWithDB(tx, rec.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	total := s.recalcTotal(allLines)
	if err := tx.Model(&rec).Update("total_amount", total).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// ---- 状态机 ----

func (s *BillingService) Confirm(id string, userID string, userName string) (*models.ChargeRecord, error) {
	return s.setStatus(id, models.ChargeStatusDraft, models.ChargeStatusConfirmed, func(rec *models.ChargeRecord) {
		rec.RecordedBy = &userID
		rec.RecordedName = &userName
	})
}

func (s *BillingService) Check(id string, userID string, userName string) (*models.ChargeRecord, error) {
	now := time.Now()
	return s.setStatus(id, models.ChargeStatusConfirmed, models.ChargeStatusChecked, func(rec *models.ChargeRecord) {
		rec.CheckedBy = &userID
		rec.CheckedName = &userName
		rec.CheckedAt = &now
	})
}

func (s *BillingService) MarkExported(id string) (*models.ChargeRecord, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}
	var rec models.ChargeRecord
	if err := s.db.First(&rec, "id = ? AND tenant_id = ?", id, s.tenantID).Error; err != nil {
		return nil, fmt.Errorf("清单不存在: %w", err)
	}
	now := time.Now()
	rec.ExportedAt = &now
	if err := s.db.Model(&rec).Updates(map[string]interface{}{
		"exported_at": now,
	}).Error; err != nil {
		return nil, fmt.Errorf("记录导出时间失败: %w", err)
	}
	return &rec, nil
}

func (s *BillingService) Cancel(id string, reason string) (*models.ChargeRecord, error) {
	return s.setStatus(id, "", models.ChargeStatusCancelled, func(rec *models.ChargeRecord) {
		rec.Note = &reason
	})
}

var validTransitions = map[string][]string{
	models.ChargeStatusDraft:     {models.ChargeStatusConfirmed, models.ChargeStatusCancelled},
	models.ChargeStatusConfirmed: {models.ChargeStatusChecked, models.ChargeStatusCancelled},
	models.ChargeStatusChecked:   {models.ChargeStatusCancelled},
}

func (s *BillingService) setStatus(id, expectedFrom, target string, patch func(*models.ChargeRecord)) (*models.ChargeRecord, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}
	var rec models.ChargeRecord
	if err := s.db.First(&rec, "id = ? AND tenant_id = ?", id, s.tenantID).Error; err != nil {
		return nil, fmt.Errorf("清单不存在: %w", err)
	}

	if expectedFrom != "" && rec.Status != expectedFrom {
		allowed, ok := validTransitions[rec.Status]
		if !ok {
			return nil, fmt.Errorf("当前状态 %s 不允许操作", rec.Status)
		}
		valid := false
		for _, a := range allowed {
			if a == target {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("不允许从 %s 转换到 %s", rec.Status, target)
		}
	}

	if patch != nil {
		patch(&rec)
	}
	rec.Status = target

	updates := map[string]interface{}{
		"status":      target,
		"updated_at":  time.Now(),
	}
	if rec.RecordedBy != nil {
		updates["recorded_by"] = *rec.RecordedBy
		updates["recorded_name"] = *rec.RecordedName
	}
	if rec.CheckedBy != nil {
		updates["checked_by"] = *rec.CheckedBy
		updates["checked_name"] = *rec.CheckedName
		updates["checked_at"] = rec.CheckedAt
	}
	if rec.Note != nil {
		updates["note"] = *rec.Note
	}

	if err := s.db.Model(&rec).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新状态失败: %w", err)
	}
	rec.Status = target
	return &rec, nil
}

func (s *BillingService) loadLines(rec *models.ChargeRecord) error {
	var lines []models.ChargeLine
	if err := s.db.Where("charge_record_id = ?", rec.ID).Order("created_at").Find(&lines).Error; err != nil {
		return err
	}
	rec.Lines = lines
	return nil
}

func (s *BillingService) getLines(recordID string) ([]models.ChargeLine, error) {
	return s.getLinesWithDB(s.db, recordID)
}

func (s *BillingService) getLinesWithDB(db *gorm.DB, recordID string) ([]models.ChargeLine, error) {
	var lines []models.ChargeLine
	err := db.Where("charge_record_id = ?", recordID).Find(&lines).Error
	return lines, err
}

func (s *BillingService) recalcTotal(lines []models.ChargeLine) float64 {
	var total float64
	for _, l := range lines {
		if l.Billable && l.Amount != nil {
			total += *l.Amount
		}
	}
	return math.Round(total*100) / 100
}
