package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/sched"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/service"
)

// 资源/病人/骨架/模板的录入维护接口(P1 #4/#5)+ 冲突处理(#6),统一挂 /admin。

func (s *Server) registerAdmin(v1 *gin.RouterGroup) {
	a := v1.Group("/admin")
	// 读(不设守卫,与 board 等只读一致)
	a.GET("/wards", s.listWards)
	a.GET("/machines", s.listMachines)
	a.GET("/shifts", s.listShiftDefs)
	a.GET("/patients", s.listPatients)
	a.GET("/profiles", s.listProfiles)
	a.GET("/profiles/:pid", s.getProfile)
	a.GET("/templates", s.listTemplates)
	a.GET("/template-items", s.listTemplateItems)
	a.GET("/incomplete-profiles", s.listIncomplete)
	// 写(机构管理 ≈ 护士长;骨架兼顾医嘱 → 医生/护士长)
	a.POST("/wards", guard(RoleHeadNurse), s.createWard)
	a.PUT("/wards/:id", guard(RoleHeadNurse), s.updateWard)
	a.POST("/machines", guard(RoleHeadNurse), s.createMachine)
	a.PUT("/machines/:id", guard(RoleHeadNurse), s.updateMachine)
	a.POST("/shifts", guard(RoleHeadNurse), s.createShiftDef)
	a.PUT("/shifts/:id", guard(RoleHeadNurse), s.updateShiftDef)
	a.POST("/disable", guard(RoleHeadNurse), s.setDisabled)
	a.POST("/patients", guard(RoleHeadNurse), s.upsertPatient)
	a.POST("/profiles", guard(RoleDoctor, RoleHeadNurse), s.upsertProfile)
	a.POST("/templates/rebuild", guard(RoleHeadNurse), s.rebuildTemplate)
	// 冲突处理(护士长/主班)
	v1.POST("/conflicts/:id/resolve", guard(RoleHeadNurse, RoleChargeNurse), s.resolveConflict)
	// 病人生命周期 / 院感(决策 26/27)
	v1.POST("/patients/:id/place", guard(RoleHeadNurse, RoleChargeNurse), s.placePatient)
	v1.POST("/patients/:id/discharge", guard(RoleDoctor, RoleHeadNurse), s.dischargePatient)
	v1.POST("/patients/:id/infection", guard(RoleDoctor, RoleHeadNurse), s.setInfection)
	v1.POST("/patients/:id/infection-waive", guard(RoleHeadNurse), s.waiveInfection)
}

func idParam(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "无效的 ID"})
		return 0, false
	}
	return id, true
}

func badReq(c *gin.Context, err error) { c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()}) }
func okJSON(c *gin.Context, v interface{}) { c.JSON(http.StatusOK, v) }

// ---- 病区 ----
func (s *Server) listWards(c *gin.Context) {
	ws, err := service.ListWards(s.DB, tenantOf(c))
	if err != nil { badReq(c, err); return }
	okJSON(c, gin.H{"items": ws})
}
func (s *Server) createWard(c *gin.Context) {
	var w model.Ward
	if c.ShouldBindJSON(&w) != nil { c.JSON(400, gin.H{"error": "请求体非法"}); return }
	r, err := service.CreateWard(s.DB, tenantOf(c), userOf(c), &w)
	if err != nil { badReq(c, err); return }
	okJSON(c, r)
}
func (s *Server) updateWard(c *gin.Context) {
	id, ok := idParam(c, "id"); if !ok { return }
	var w model.Ward
	if c.ShouldBindJSON(&w) != nil { c.JSON(400, gin.H{"error": "请求体非法"}); return }
	if err := service.UpdateWard(s.DB, tenantOf(c), id, &w); err != nil { badReq(c, err); return }
	okJSON(c, gin.H{"ok": true})
}

// ---- 机器 ----
func (s *Server) listMachines(c *gin.Context) {
	var wardID int64
	if v := c.Query("wardId"); v != "" { wardID, _ = strconv.ParseInt(v, 10, 64) }
	ms, err := service.ListMachines(s.DB, tenantOf(c), wardID)
	if err != nil { badReq(c, err); return }
	okJSON(c, gin.H{"items": ms})
}
func (s *Server) createMachine(c *gin.Context) {
	var m model.Machine
	if c.ShouldBindJSON(&m) != nil { c.JSON(400, gin.H{"error": "请求体非法"}); return }
	r, err := service.CreateMachine(s.DB, tenantOf(c), userOf(c), &m)
	if err != nil { badReq(c, err); return }
	okJSON(c, r)
}
func (s *Server) updateMachine(c *gin.Context) {
	id, ok := idParam(c, "id"); if !ok { return }
	var m model.Machine
	if c.ShouldBindJSON(&m) != nil { c.JSON(400, gin.H{"error": "请求体非法"}); return }
	if err := service.UpdateMachine(s.DB, tenantOf(c), id, &m); err != nil { badReq(c, err); return }
	okJSON(c, gin.H{"ok": true})
}

// ---- 班次定义 ----
func (s *Server) listShiftDefs(c *gin.Context) {
	ss, err := service.ListShifts(s.DB, tenantOf(c))
	if err != nil { badReq(c, err); return }
	okJSON(c, gin.H{"items": ss})
}
func (s *Server) createShiftDef(c *gin.Context) {
	var sh model.Shift
	if c.ShouldBindJSON(&sh) != nil { c.JSON(400, gin.H{"error": "请求体非法"}); return }
	r, err := service.CreateShift(s.DB, tenantOf(c), userOf(c), &sh)
	if err != nil { badReq(c, err); return }
	okJSON(c, r)
}
func (s *Server) updateShiftDef(c *gin.Context) {
	id, ok := idParam(c, "id"); if !ok { return }
	var sh model.Shift
	if c.ShouldBindJSON(&sh) != nil { c.JSON(400, gin.H{"error": "请求体非法"}); return }
	if err := service.UpdateShift(s.DB, tenantOf(c), id, &sh); err != nil { badReq(c, err); return }
	okJSON(c, gin.H{"ok": true})
}

// ---- 启停 ----
func (s *Server) setDisabled(c *gin.Context) {
	var req struct {
		Type     string `json:"type"`
		Id       int64  `json:"id"`
		Disabled bool   `json:"disabled"`
	}
	if c.ShouldBindJSON(&req) != nil || req.Id == 0 { c.JSON(400, gin.H{"error": "需提供 type 与 id"}); return }
	if err := service.SetDisabled(s.DB, tenantOf(c), req.Type, req.Id, req.Disabled); err != nil { badReq(c, err); return }
	okJSON(c, gin.H{"ok": true})
}

// ---- 病人主档 ----
func (s *Server) listPatients(c *gin.Context) {
	ps, err := service.ListPatients(s.DB, tenantOf(c))
	if err != nil { badReq(c, err); return }
	okJSON(c, gin.H{"items": ps})
}
func (s *Server) upsertPatient(c *gin.Context) {
	var p model.Patient
	if c.ShouldBindJSON(&p) != nil { c.JSON(400, gin.H{"error": "请求体非法"}); return }
	r, err := service.UpsertPatient(s.DB, tenantOf(c), &p)
	if err != nil { badReq(c, err); return }
	okJSON(c, r)
}

// ---- 排班骨架 ----
func (s *Server) listProfiles(c *gin.Context) {
	ps, err := service.ListProfiles(s.DB, tenantOf(c))
	if err != nil { badReq(c, err); return }
	okJSON(c, gin.H{"items": ps})
}
func (s *Server) getProfile(c *gin.Context) {
	pid, ok := idParam(c, "pid"); if !ok { return }
	p, err := service.GetProfile(s.DB, tenantOf(c), pid)
	if err == service.ErrNotFound { c.JSON(http.StatusNotFound, gin.H{"error": "骨架不存在"}); return }
	if err != nil { badReq(c, err); return }
	okJSON(c, p)
}
func (s *Server) upsertProfile(c *gin.Context) {
	var p model.PatientProfile
	if c.ShouldBindJSON(&p) != nil { c.JSON(400, gin.H{"error": "请求体非法"}); return }
	r, err := service.UpsertProfile(s.DB, tenantOf(c), &p)
	if err != nil { badReq(c, err); return }
	okJSON(c, r)
}

// ---- 模板 ----
func (s *Server) listTemplates(c *gin.Context) {
	ts, err := service.ListTemplates(s.DB, tenantOf(c))
	if err != nil { badReq(c, err); return }
	okJSON(c, gin.H{"items": ts})
}
func (s *Server) listTemplateItems(c *gin.Context) {
	tid, _ := strconv.ParseInt(c.Query("templateId"), 10, 64)
	its, err := service.ListTemplateItems(s.DB, tenantOf(c), tid)
	if err != nil { badReq(c, err); return }
	okJSON(c, gin.H{"items": its})
}
func (s *Server) rebuildTemplate(c *gin.Context) {
	var req struct{ Name string `json:"name"` }
	_ = c.ShouldBindJSON(&req)
	r, err := service.RebuildTemplateFromProfiles(s.DB, tenantOf(c), req.Name)
	if err != nil { badReq(c, err); return }
	okJSON(c, r)
}

// ---- 资料待补 ----
func (s *Server) listIncomplete(c *gin.Context) {
	items, err := service.ListIncompleteProfiles(s.DB, tenantOf(c))
	if err != nil { badReq(c, err); return }
	okJSON(c, gin.H{"count": len(items), "items": items})
}

// ---- 病人生命周期 / 院感 ----
func (s *Server) placePatient(c *gin.Context) {
	id, ok := idParam(c, "id"); if !ok { return }
	var req struct {
		WeekStart string `json:"weekStart"`
		Weeks     int    `json:"weeks"`
	}
	_ = c.ShouldBindJSON(&req)
	weeks := req.Weeks
	if weeks != 2 && weeks != 4 { weeks = 2 }
	start := schedMonday(req.WeekStart)
	placed, conflicts, err := service.PlaceNewPatientService(s.DB, tenantOf(c), id, start, weeks)
	if err == service.ErrNotFound { c.JSON(404, gin.H{"error": "病人骨架不存在"}); return }
	if err != nil { badReq(c, err); return }
	okJSON(c, gin.H{"placed": placed, "conflicts": conflicts})
}

func (s *Server) dischargePatient(c *gin.Context) {
	id, ok := idParam(c, "id"); if !ok { return }
	var req struct{ Reason string `json:"reason"` }
	_ = c.ShouldBindJSON(&req)
	if err := service.DischargePatient(s.DB, tenantOf(c), userOf(c), id, req.Reason); err != nil {
		if err == service.ErrNotFound { c.JSON(404, gin.H{"error": "病人骨架不存在"}); return }
		badReq(c, err); return
	}
	okJSON(c, gin.H{"ok": true})
}

func (s *Server) setInfection(c *gin.Context) {
	id, ok := idParam(c, "id"); if !ok { return }
	var req struct{ Status string `json:"status"` }
	_ = c.ShouldBindJSON(&req)
	if err := service.SetInfectionStatus(s.DB, tenantOf(c), id, req.Status); err != nil {
		if err == service.ErrNotFound { c.JSON(404, gin.H{"error": "病人不存在"}); return }
		badReq(c, err); return
	}
	okJSON(c, gin.H{"ok": true})
}

func (s *Server) waiveInfection(c *gin.Context) {
	id, ok := idParam(c, "id"); if !ok { return }
	if err := service.WaiveInfection(s.DB, tenantOf(c), userOf(c), id); err != nil {
		if err == service.ErrNotFound { c.JSON(404, gin.H{"error": "病人不存在"}); return }
		badReq(c, err); return
	}
	okJSON(c, gin.H{"ok": true})
}

// schedMonday 解析 weekStart(空=本周一)。
func schedMonday(v string) time.Time {
	if v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			return sched.MondayOf(t)
		}
	}
	return sched.MondayOf(time.Now())
}

// ---- 冲突处理 ----
func (s *Server) resolveConflict(c *gin.Context) {
	id, ok := idParam(c, "id"); if !ok { return }
	var req struct{ Action string `json:"action"` } // accept / ignore
	_ = c.ShouldBindJSON(&req)
	if err := service.ResolveConflict(s.DB, tenantOf(c), userOf(c), id, req.Action == "accept"); err != nil {
		if err == service.ErrNotFound { c.JSON(http.StatusNotFound, gin.H{"error": "冲突项不存在"}); return }
		badReq(c, err); return
	}
	okJSON(c, gin.H{"ok": true})
}
