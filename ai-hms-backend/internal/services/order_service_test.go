//go:build cgo

package services

import (
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const testPatientID = "1"

func newTestOrderService(t *testing.T) *OrderService {
	t.Helper()

	// 每个测试使用独立内存数据库，防止并行测试数据污染，且永远不会连接生产库
	db, err := gorm.Open(sqlite.Open("file::memory:?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(&models.Order{}, &models.OrderTemplate{}, &models.OrderTemplateItem{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	return &OrderService{db: db}
}

func strPtr(v string) *string {
	return &v
}

func mustCreateOrder(t *testing.T, db *gorm.DB, order models.Order) models.Order {
	t.Helper()
	if order.ID == "" {
		order.ID = uuid.New().String()
	}
	if order.PatientID == 0 {
		order.PatientID = modeltypes.LegacyID(1)
	}
	if order.Type == "" {
		order.Type = models.OrderTypeLongTerm
	}
	if order.Content == "" {
		order.Content = "默认内容"
	}
	if order.DoctorID == "" {
		order.DoctorID = "D1"
	}
	if order.DoctorName == "" {
		order.DoctorName = "医生A"
	}
	if order.Status == "" {
		order.Status = models.OrderStatusPending
	}
	if order.StartTime.IsZero() {
		order.StartTime = time.Date(2026, 3, 6, 9, 0, 0, 0, time.Local)
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order failed: %v", err)
	}
	return order
}

func mustCreateTemplate(t *testing.T, db *gorm.DB, template models.OrderTemplate, items ...models.OrderTemplateItem) models.OrderTemplate {
	t.Helper()
	if template.ID == "" {
		template.ID = uuid.New().String()
	}
	if template.Name == "" {
		template.Name = "模板A"
	}
	if template.Type == "" {
		template.Type = models.OrderTypeLongTerm
	}
	if template.Category == "" {
		template.Category = models.OrderCategoryMedicine
	}
	if template.Content == "" {
		template.Content = "模板内容"
	}
	if template.Priority == "" {
		template.Priority = models.OrderPriorityNormal
	}
	if err := db.Create(&template).Error; err != nil {
		t.Fatalf("create template failed: %v", err)
	}
	for i := range items {
		if items[i].ID == "" {
			items[i].ID = uuid.New().String()
		}
		items[i].TemplateID = template.ID
		if err := db.Create(&items[i]).Error; err != nil {
			t.Fatalf("create template item failed: %v", err)
		}
	}
	template.Items = items
	return template
}

func TestCreateFromTemplateAppliesOverridesAndRemapsGroupIDPerImport(t *testing.T) {
	svc := newTestOrderService(t)
	groupID := uuid.New().String()
	template := mustCreateTemplate(t, svc.db, models.OrderTemplate{
		Type:     models.OrderTypeLongTerm,
		Category: models.OrderCategoryMedicine,
	},
		models.OrderTemplateItem{
			DrugName:  "肝素钠",
			Dosage:    strPtr("2500"),
			Unit:      strPtr("iu"),
			Route:     strPtr("静脉推注"),
			Frequency: strPtr("每次透析"),
			Timing:    strPtr("透析开始"),
			GroupID:   &groupID,
		},
		models.OrderTemplateItem{
			DrugName:  "低分子肝素",
			Dosage:    strPtr("1"),
			Unit:      strPtr("支"),
			Route:     strPtr("静脉推注"),
			Frequency: strPtr("每次透析"),
			Timing:    strPtr("透析开始"),
			GroupID:   &groupID,
		},
	)

	items := []CreateFromTemplateItemRequest{
		{
			TemplateItemID: template.Items[0].ID,
			Dose:           strPtr("3000"),
			Route:          strPtr("体外循环"),
			Frequency:      strPtr("每周三次"),
			Timing:         strPtr("开机"),
		},
		{
			TemplateItemID: template.Items[1].ID,
		},
	}

	firstBatch, err := svc.CreateFromTemplate(testPatientID, 3, "D1", "医生A", CreateFromTemplateRequest{
		TemplateID: template.ID,
		Items:      items,
	})
	if err != nil {
		t.Fatalf("CreateFromTemplate returned error: %v", err)
	}
	secondBatch, err := svc.CreateFromTemplate(testPatientID, 3, "D1", "医生A", CreateFromTemplateRequest{
		TemplateID: template.ID,
		Items:      items,
	})
	if err != nil {
		t.Fatalf("CreateFromTemplate returned error: %v", err)
	}
	if len(firstBatch) != 2 || len(secondBatch) != 2 {
		t.Fatalf("expected 2 orders per batch, got %d and %d", len(firstBatch), len(secondBatch))
	}
	got := firstBatch[0]
	if got.Dose != "3000" || got.Route != "体外循环" || got.Timing != "开机" {
		t.Fatalf("expected override fields applied, got %+v", got)
	}
	if firstBatch[0].GroupID == nil || firstBatch[1].GroupID == nil {
		t.Fatalf("expected first batch group ids assigned, got %+v %+v", firstBatch[0].GroupID, firstBatch[1].GroupID)
	}
	if secondBatch[0].GroupID == nil || secondBatch[1].GroupID == nil {
		t.Fatalf("expected second batch group ids assigned, got %+v %+v", secondBatch[0].GroupID, secondBatch[1].GroupID)
	}
	if *firstBatch[0].GroupID != *firstBatch[1].GroupID {
		t.Fatalf("expected first batch orders grouped together, got %+v %+v", firstBatch[0].GroupID, firstBatch[1].GroupID)
	}
	if *secondBatch[0].GroupID != *secondBatch[1].GroupID {
		t.Fatalf("expected second batch orders grouped together, got %+v %+v", secondBatch[0].GroupID, secondBatch[1].GroupID)
	}
	if *firstBatch[0].GroupID == groupID || *secondBatch[0].GroupID == groupID {
		t.Fatalf("expected imported orders to use remapped group ids, got first=%q second=%q template=%q", *firstBatch[0].GroupID, *secondBatch[0].GroupID, groupID)
	}
	if *firstBatch[0].GroupID == *secondBatch[0].GroupID {
		t.Fatalf("expected each import batch to get a distinct group id, got %q", *firstBatch[0].GroupID)
	}
}

func TestCreateFromTemplateUsesRequestedOrderTypeInsteadOfTemplateType(t *testing.T) {
	svc := newTestOrderService(t)
	template := mustCreateTemplate(t, svc.db, models.OrderTemplate{
		Type:     models.OrderTypeLongTerm,
		Category: models.OrderCategoryMedicine,
	}, models.OrderTemplateItem{
		DrugName:  "肝素钠",
		Route:     strPtr("静脉推注"),
		Frequency: strPtr("每次透析"),
		Timing:    strPtr("透析开始"),
	})

	orders, err := svc.CreateFromTemplate(testPatientID, 3, "D1", "医生A", CreateFromTemplateRequest{
		TemplateID: template.ID,
		Type:       models.OrderTypeTemporary,
		Items: []CreateFromTemplateItemRequest{
			{
				TemplateItemID: template.Items[0].ID,
				ExecTiming:     strPtr("立即执行"),
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateFromTemplate returned error: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
	got := orders[0]
	if got.Type != models.OrderTypeTemporary {
		t.Fatalf("expected temporary order type, got %s", got.Type)
	}
	if got.ExecTiming != "立即执行" {
		t.Fatalf("expected exec timing preserved, got %+v", got.ExecTiming)
	}
	if got.Frequency != nil || got.Timing != "" {
		t.Fatalf("expected long-term fields cleared for temporary order, got %+v", got)
	}
}

func TestStopAffectsOnlyCurrentImportedBatchWhenTemplateHasGroupedItems(t *testing.T) {
	svc := newTestOrderService(t)
	groupID := uuid.New().String()
	template := mustCreateTemplate(t, svc.db, models.OrderTemplate{
		Type:     models.OrderTypeLongTerm,
		Category: models.OrderCategoryMedicine,
	},
		models.OrderTemplateItem{DrugName: "A", GroupID: &groupID},
		models.OrderTemplateItem{DrugName: "B", GroupID: &groupID},
	)
	req := CreateFromTemplateRequest{
		TemplateID: template.ID,
		Items: []CreateFromTemplateItemRequest{
			{TemplateItemID: template.Items[0].ID},
			{TemplateItemID: template.Items[1].ID},
		},
	}

	firstBatch, err := svc.CreateFromTemplate(testPatientID, 3, "D1", "医生A", req)
	if err != nil {
		t.Fatalf("CreateFromTemplate first batch returned error: %v", err)
	}
	secondBatch, err := svc.CreateFromTemplate(testPatientID, 3, "D1", "医生A", req)
	if err != nil {
		t.Fatalf("CreateFromTemplate second batch returned error: %v", err)
	}

	if _, err := svc.Stop(testPatientID, secondBatch[0].ID, "测试停用", nil); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}

	for _, order := range secondBatch {
		var got models.Order
		if err := svc.db.First(&got, "id = ?", order.ID).Error; err != nil {
			t.Fatalf("reload second batch order failed: %v", err)
		}
		if got.Status != models.OrderStatusStopped {
			t.Fatalf("expected second batch stopped, got %+v", got)
		}
	}
	for _, order := range firstBatch {
		var got models.Order
		if err := svc.db.First(&got, "id = ?", order.ID).Error; err != nil {
			t.Fatalf("reload first batch order failed: %v", err)
		}
		if got.Status != models.OrderStatusPending {
			t.Fatalf("expected first batch unaffected, got %+v", got)
		}
	}
}

func TestGroupAndUngroupPersistGroupID(t *testing.T) {
	svc := newTestOrderService(t)
	a := mustCreateOrder(t, svc.db, models.Order{Name: "A"})
	b := mustCreateOrder(t, svc.db, models.Order{Name: "B"})

	grouped, err := svc.Group(testPatientID, []string{a.ID, b.ID})
	if err != nil {
		t.Fatalf("Group returned error: %v", err)
	}
	if len(grouped) != 2 {
		t.Fatalf("expected 2 grouped orders, got %d", len(grouped))
	}
	if grouped[0].GroupID == nil || grouped[1].GroupID == nil || *grouped[0].GroupID != *grouped[1].GroupID {
		t.Fatalf("expected same group id, got %+v %+v", grouped[0].GroupID, grouped[1].GroupID)
	}

	ungrouped, err := svc.Ungroup(testPatientID, []string{a.ID, b.ID})
	if err != nil {
		t.Fatalf("Ungroup returned error: %v", err)
	}
	for _, order := range ungrouped {
		if order.GroupID != nil {
			t.Fatalf("expected group id cleared, got %+v", order.GroupID)
		}
	}
}

func TestGroupRejectsMixedTypeAndStoppedOrders(t *testing.T) {
	svc := newTestOrderService(t)
	longTerm := mustCreateOrder(t, svc.db, models.Order{Name: "长期A", Type: models.OrderTypeLongTerm})
	temporary := mustCreateOrder(t, svc.db, models.Order{Name: "临时B", Type: models.OrderTypeTemporary})

	if _, err := svc.Group(testPatientID, []string{longTerm.ID, temporary.ID}); err == nil {
		t.Fatalf("expected mixed-type group to fail")
	}

	stopped := mustCreateOrder(t, svc.db, models.Order{Name: "已停用", Status: models.OrderStatusStopped})
	if _, err := svc.Group(testPatientID, []string{longTerm.ID, stopped.ID}); err == nil {
		t.Fatalf("expected stopped-order group to fail")
	}
}

func TestReviseCreatesReplacementForNonLinkedFieldsAndKeepsGroupID(t *testing.T) {
	svc := newTestOrderService(t)
	groupID := uuid.New().String()
	origin := mustCreateOrder(t, svc.db, models.Order{
		Name:     "原医嘱",
		Dose:     "10",
		Content:  "原内容",
		GroupID:  &groupID,
		Category: models.OrderCategoryMedicine,
	})

	stopDate := "2026-03-07"
	revised, err := svc.Revise(testPatientID, origin.ID, 3, "D2", "医生B", OrderReviseRequest{
		Name:     strPtr("新医嘱"),
		Dose:     strPtr("20"),
		StopDate: &stopDate,
	})
	if err != nil {
		t.Fatalf("Revise returned error: %v", err)
	}
	if revised.ID == origin.ID {
		t.Fatalf("expected replacement order, got same id")
	}
	if revised.GroupID == nil || *revised.GroupID != groupID {
		t.Fatalf("expected revised order to inherit group id, got %+v", revised.GroupID)
	}

	var reloaded models.Order
	if err := svc.db.First(&reloaded, "id = ?", origin.ID).Error; err != nil {
		t.Fatalf("reload original order failed: %v", err)
	}
	if reloaded.Status != models.OrderStatusStopped {
		t.Fatalf("expected original order stopped, got %s", reloaded.Status)
	}
	if reloaded.EndTime == nil || reloaded.EndTime.Hour() != 0 || reloaded.EndTime.Minute() != 0 {
		t.Fatalf("expected stop date normalized to start of day, got %+v", reloaded.EndTime)
	}
}

func TestStopNormalizesStopDateAndStopsWholeGroup(t *testing.T) {
	svc := newTestOrderService(t)
	groupID := uuid.New().String()
	a := mustCreateOrder(t, svc.db, models.Order{Name: "A", GroupID: &groupID})
	b := mustCreateOrder(t, svc.db, models.Order{Name: "B", GroupID: &groupID})
	stopDate := "2026-03-09"

	affected, err := svc.Stop(testPatientID, a.ID, "医生停用", &stopDate)
	if err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if len(affected) != 2 {
		t.Fatalf("expected group stop to affect 2 orders, got %d", len(affected))
	}
	for _, order := range affected {
		if order.Status != models.OrderStatusStopped {
			t.Fatalf("expected stopped status, got %+v", order)
		}
		if order.EndTime == nil || order.EndTime.Format("2006-01-02 15:04:05") != "2026-03-09 00:00:00" {
			t.Fatalf("expected normalized stop date, got %+v", order.EndTime)
		}
	}

	var reloaded models.Order
	if err := svc.db.First(&reloaded, "id = ?", b.ID).Error; err != nil {
		t.Fatalf("reload second order failed: %v", err)
	}
	if reloaded.Status != models.OrderStatusStopped {
		t.Fatalf("expected second order stopped, got %s", reloaded.Status)
	}
}

func TestMarkExpiredOrdersStopsLongTermAndTemporary(t *testing.T) {
	svc := newTestOrderService(t)
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local)
	past := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)

	longTerm := mustCreateOrder(t, svc.db, models.Order{
		Name:    "长期",
		Type:    models.OrderTypeLongTerm,
		EndTime: &past,
		Status:  models.OrderStatusPending,
	})
	temporary := mustCreateOrder(t, svc.db, models.Order{
		Name:    "临时",
		Type:    models.OrderTypeTemporary,
		EndTime: &past,
		Status:  models.OrderStatusExecuting,
	})

	if err := markExpiredOrders(svc.db, now); err != nil {
		t.Fatalf("markExpiredOrders returned error: %v", err)
	}

	for _, id := range []string{longTerm.ID, temporary.ID} {
		var order models.Order
		if err := svc.db.First(&order, "id = ?", id).Error; err != nil {
			t.Fatalf("reload order %s failed: %v", id, err)
		}
		if order.Status != models.OrderStatusStopped {
			t.Fatalf("expected order %s stopped by cron, got %s", id, order.Status)
		}
	}
}
