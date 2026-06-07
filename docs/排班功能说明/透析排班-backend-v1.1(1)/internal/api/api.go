// Package api 提供 Gin HTTP 路由:排班生成、周视图、冲突队列、演示数据。
package api

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/sdsph/dialysis-scheduling/internal/db"
	"github.com/sdsph/dialysis-scheduling/internal/model"
	"github.com/sdsph/dialysis-scheduling/internal/sched"
	"github.com/sdsph/dialysis-scheduling/internal/seed"
	"github.com/sdsph/dialysis-scheduling/internal/service"
)

// devSuperuser:开发模式下,未带 X-Role 的请求放行(便于本地联调)。
// 生产环境不要设置 DEV_SUPERUSER=true —— 届时未鉴权请求将被拒绝(401)。
var devSuperuser = strings.EqualFold(os.Getenv("DEV_SUPERUSER"), "true")

// Server 持有数据库句柄。
type Server struct {
	DB *gorm.DB
}

// Register 注册路由。
func (s *Server) Register(r *gin.Engine) {
	r.Use(auditMiddleware())            // 请求审计日志
	r.GET("/health", s.health)          // 健康检查(无需鉴权)
	r.StaticFile("/", "web/index.html") // 同源托管前端,避免 CORS
	v1 := r.Group("/api/v1")
	v1.Use(tenantMiddleware())
	{
		// 只读
		v1.GET("/schedule/week", s.weekView)
		v1.GET("/schedule/board", s.board)
		v1.GET("/schedule/diffs", s.diffs)
		v1.GET("/conflicts", s.listConflicts)
		// 生成 / 演示数据(管理类:护士长/主班)
		v1.POST("/schedule/generate", guard(RoleHeadNurse, RoleChargeNurse), s.generate)
		v1.POST("/admin/seed", s.seedDemo)
		// 确认(① 护士长;②③ 护士长/主班)
		v1.POST("/schedule/confirm-plan", guard(RoleHeadNurse), s.confirmPlan)
		v1.POST("/schedule/confirm-day", guard(RoleHeadNurse, RoleChargeNurse), s.confirmDay)
		// 排班调整(护士长/主班)
		v1.POST("/shifts/:id/cancel", guard(RoleHeadNurse, RoleChargeNurse), s.cancelShift)
		v1.POST("/shifts/:id/absent", guard(RoleHeadNurse, RoleChargeNurse), s.absentShift)
		v1.POST("/shifts/:id/move", guard(RoleHeadNurse, RoleChargeNurse), s.moveShift)
		// 临时透析 / CRRT(医嘱 + 护士长/主班确认)
		v1.POST("/schedule/temporary", guard(RoleDoctor, RoleHeadNurse, RoleChargeNurse), s.insertTemporary)
		v1.POST("/schedule/crrt", guard(RoleDoctor, RoleHeadNurse, RoleChargeNurse), s.insertCrrt)
		v1.GET("/schedule/crrt", s.listCrrt)
		// 停机(工程师/护士长)、假日(机构)、方案变更(医生/护士长)
		v1.POST("/machines/:id/outage", guard(RoleHeadNurse), s.machineOutage)
		v1.POST("/schedule/holiday", guard(RoleHeadNurse), s.setHoliday)
		v1.POST("/patients/:id/plan-change", guard(RoleDoctor, RoleHeadNurse), s.planChange)
		// 补透(护士长/主班)
		v1.POST("/patients/:id/makeup", guard(RoleHeadNurse, RoleChargeNurse), s.makeup)
	}
}

// tenantMiddleware 从请求头取租户与角色(后续接老系统鉴权中间件)。
// X-Tenant-Id 缺省 1;X-Role 为角色(生产环境必传,否则受 guard 保护的接口返回 401)。
func tenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant := int64(1)
		if v := c.GetHeader("X-Tenant-Id"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				tenant = n
			}
		}
		c.Set("tenant", tenant)
		c.Set("role", c.GetHeader("X-Role"))
		c.Next()
	}
}

// 角色常量(规范 §11 权限矩阵)。
const (
	RoleDoctor      = "doctor"       // 医生
	RoleHeadNurse   = "head_nurse"   // 护士长
	RoleChargeNurse = "charge_nurse" // 主班护士
	RoleNurse       = "nurse"        // 护士
)

// guard 路由守卫:校验当前角色是否在允许集合内。
// 未带 X-Role 时:开发模式(DEV_SUPERUSER=true)放行,否则返回 401(移除"默认超级用户"放行)。
func guard(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role == "" {
			if devSuperuser {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "error": "未鉴权:请提供 X-Role"})
			return
		}
		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "error": "权限不足:该操作需要 " + joinRoles(roles)})
	}
}

func joinRoles(roles []string) string {
	out := ""
	for i, r := range roles {
		if i > 0 {
			out += "/"
		}
		out += r
	}
	return out
}

func tenantOf(c *gin.Context) int64 {
	if v, ok := c.Get("tenant"); ok {
		return v.(int64)
	}
	return 1
}

// failInternal 记录真实错误到服务端日志,只向客户端返回脱敏的通用提示(避免泄露表结构等)。
func failInternal(c *gin.Context, err error) {
	log.Printf("[ERROR] %s %s tenant=%d role=%q: %v",
		c.Request.Method, c.Request.URL.Path, tenantOf(c), c.GetString("role"), err)
	c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": "服务器内部错误,请稍后重试或联系管理员"})
}

// auditMiddleware 请求审计日志:方法、路径、状态码、耗时、租户、角色、客户端 IP。
func auditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("[AUDIT] %s %s -> %d %v tenant=%v role=%q ip=%s",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start),
			c.Value("tenant"), c.GetString("role"), c.ClientIP())
	}
}

// health GET /health —— 健康检查(含数据库连通性),供容器编排探活。
func (s *Server) health(c *gin.Context) {
	if err := db.Ping(s.DB); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down", "db": "unreachable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "db": "ok", "time": time.Now().Format(time.RFC3339)})
}

type generateReq struct {
	StartDate string `json:"startDate"` // YYYY-MM-DD;空=本周一
	Weeks     int    `json:"weeks"`     // 2 或 4;空/非法=2
}

// generate POST /schedule/generate —— 生成未来 2/4 周草稿。
func (s *Server) generate(c *gin.Context) {
	var req generateReq
	_ = c.ShouldBindJSON(&req)

	weeks := req.Weeks
	if weeks != 2 && weeks != 4 {
		weeks = 2
	}
	start := sched.MondayOf(time.Now())
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "startDate 格式应为 YYYY-MM-DD"})
			return
		}
		start = t
	}

	res, err := service.GenerateSchedule(s.DB, tenantOf(c), start, weeks)
	if err != nil {
		failInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// weekView GET /schedule/week?date=YYYY-MM-DD —— 返回该日所在周(周一~周日)的排班记录。
func (s *Server) weekView(c *gin.Context) {
	day := time.Now()
	if v := c.Query("date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			day = t
		}
	}
	mon := sched.MondayOf(day)
	sun := mon.AddDate(0, 0, 6)

	var shifts []model.PatientShift
	err := s.DB.Where(`"TenantId" = ? AND "ScheduleDate" BETWEEN ? AND ?`, tenantOf(c), mon, sun).
		Order(`"ScheduleDate", "ShiftId", "WardId", "MachineId"`).
		Find(&shifts).Error
	if err != nil {
		failInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"weekStart": mon.Format("2006-01-02"),
		"weekEnd":   sun.Format("2006-01-02"),
		"count":     len(shifts),
		"shifts":    shifts,
	})
}

// board GET /schedule/board?date=YYYY-MM-DD —— 周视图聚合矩阵(区→机器→班×日)。
func (s *Server) board(c *gin.Context) {
	day := time.Now()
	if v := c.Query("date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			day = t
		}
	}
	wb, err := service.BuildWeekBoard(s.DB, tenantOf(c), day)
	if err != nil {
		failInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, wb)
}

func userOf(c *gin.Context) int64 {
	if v := c.GetHeader("X-User-Id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return 1
}

func shiftIdParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的排班ID"})
		return 0, false
	}
	return id, true
}

// 把服务层错误映射到合适的 HTTP 状态码。
func opError(c *gin.Context, err error) {
	switch err {
	case service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case service.ErrLocked, service.ErrOccupied, service.ErrDoubleBook, service.ErrModeMismatch:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		failInternal(c, err)
	}
}

// confirmPlan POST /schedule/confirm-plan —— 第一次确认(护士长),整盘草稿→生效。
func (s *Server) confirmPlan(c *gin.Context) {
	var req struct {
		WeekStart string `json:"weekStart"`
		Weeks     int    `json:"weeks"`
	}
	_ = c.ShouldBindJSON(&req)
	weeks := req.Weeks
	if weeks != 2 && weeks != 4 {
		weeks = 2
	}
	start := sched.MondayOf(time.Now())
	if req.WeekStart != "" {
		if t, err := time.Parse("2006-01-02", req.WeekStart); err == nil {
			start = t
		}
	}
	n, err := service.ConfirmPlan(s.DB, tenantOf(c), userOf(c), start, weeks)
	if err != nil {
		opError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"confirmed": n})
}

// confirmDay POST /schedule/confirm-day —— 第二/三次确认。
func (s *Server) confirmDay(c *gin.Context) {
	var req struct {
		Date  string `json:"date"`
		Level int    `json:"level"`
	}
	_ = c.ShouldBindJSON(&req)
	t, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date 格式应为 YYYY-MM-DD"})
		return
	}
	if req.Level != 2 && req.Level != 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "level 须为 2 或 3"})
		return
	}
	n, e := service.ConfirmDay(s.DB, tenantOf(c), userOf(c), t, req.Level)
	if e != nil {
		opError(c, e)
		return
	}
	c.JSON(http.StatusOK, gin.H{"confirmed": n})
}

// cancelShift POST /shifts/:id/cancel —— 取消(提前请假)。
func (s *Server) cancelShift(c *gin.Context) {
	id, ok := shiftIdParam(c)
	if !ok {
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := service.CancelShift(s.DB, tenantOf(c), id, req.Reason); err != nil {
		opError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// absentShift POST /shifts/:id/absent —— 当日缺席。
func (s *Server) absentShift(c *gin.Context) {
	id, ok := shiftIdParam(c)
	if !ok {
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := service.MarkAbsent(s.DB, tenantOf(c), id, req.Reason); err != nil {
		opError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// moveShift POST /shifts/:id/move —— 移床/换班。
func (s *Server) moveShift(c *gin.Context) {
	id, ok := shiftIdParam(c)
	if !ok {
		return
	}
	var req struct {
		MachineId int64  `json:"machineId"`
		Date      string `json:"date"`
		ShiftId   *int64 `json:"shiftId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.MachineId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "需提供 machineId"})
		return
	}
	var datePtr *time.Time
	if req.Date != "" {
		if t, err := time.Parse("2006-01-02", req.Date); err == nil {
			datePtr = &t
		}
	}
	if err := service.MoveShift(s.DB, tenantOf(c), id, req.MachineId, datePtr, req.ShiftId); err != nil {
		opError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// insertTemporary POST /schedule/temporary —— 临时透析插入(急诊加台)。
func (s *Server) insertTemporary(c *gin.Context) {
	var req struct {
		PatientId int64  `json:"patientId"`
		WardId    int64  `json:"wardId"`
		Date      string `json:"date"`
		Mode      string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.PatientId == 0 || req.WardId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "需提供 patientId 与 wardId"})
		return
	}
	date := time.Now()
	if req.Date != "" {
		if t, err := time.Parse("2006-01-02", req.Date); err == nil {
			date = t
		}
	}
	mode := req.Mode
	if mode == "" {
		mode = sched.ModeHD
	}
	rec, err := service.InsertTemporary(s.DB, tenantOf(c), req.PatientId, req.WardId, date, mode)
	if err == service.ErrNoSlot {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		failInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "shift": rec})
}

// machineOutage POST /machines/:id/outage —— 登记停机并迁移受影响排班。
func (s *Server) machineOutage(c *gin.Context) {
	mid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的机器ID"})
		return
	}
	var req struct {
		StartDate string `json:"startDate"`
		EndDate   string `json:"endDate"`
		Type      int16  `json:"type"` // 10临时 20长期/报废
		Reason    string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	start, e1 := time.Parse("2006-01-02", req.StartDate)
	end, e2 := time.Parse("2006-01-02", req.EndDate)
	if e1 != nil || e2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate/endDate 格式应为 YYYY-MM-DD"})
		return
	}
	if req.Type != sched.OutageTemp && req.Type != sched.OutageLong {
		req.Type = sched.OutageTemp
	}
	// 结束日含当天:推到当日 23:59:59
	end = end.Add(24*time.Hour - time.Second)
	res, err := service.RegisterOutageAndMigrate(s.DB, tenantOf(c), mid, start, end, req.Type, req.Reason)
	if err != nil {
		failInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// setHoliday POST /schedule/holiday —— 设为非透析日并处理受影响排班。
func (s *Server) setHoliday(c *gin.Context) {
	var req struct {
		Date string `json:"date"`
		Mode int16  `json:"mode"` // 10全院停 / 20假日值班
	}
	_ = c.ShouldBindJSON(&req)
	d, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date 格式应为 YYYY-MM-DD"})
		return
	}
	mode := req.Mode
	if mode == 0 {
		mode = 10
	}
	res, e := service.SetHoliday(s.DB, tenantOf(c), d, mode)
	if e != nil {
		failInternal(c, e)
		return
	}
	c.JSON(http.StatusOK, res)
}

// planChange POST /patients/:id/plan-change —— 方案变更带生效日。
func (s *Server) planChange(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的病人ID"})
		return
	}
	var req struct {
		ChangeType    string `json:"changeType"`
		NewValue      string `json:"newValue"`
		EffectiveDate string `json:"effectiveDate"`
	}
	_ = c.ShouldBindJSON(&req)
	eff := time.Now().AddDate(0, 0, 1) // 默认次日
	if req.EffectiveDate != "" {
		if t, e := time.Parse("2006-01-02", req.EffectiveDate); e == nil {
			eff = t
		}
	}
	res, e := service.ApplyPlanChange(s.DB, tenantOf(c), pid, req.ChangeType, req.NewValue, eff)
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// diffs GET /schedule/diffs?date=YYYY-MM-DD&weeks=2 —— 应排/已排差异检测。
func (s *Server) diffs(c *gin.Context) {
	day := time.Now()
	if v := c.Query("date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			day = t
		}
	}
	weeks := 2
	if v := c.Query("weeks"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && (n == 2 || n == 4) {
			weeks = n
		}
	}
	res, err := service.ComputeDiffs(s.DB, tenantOf(c), sched.MondayOf(day), weeks)
	if err != nil {
		failInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// insertCrrt POST /schedule/crrt —— 安排 CRRT(C 区,机器+起止时间)。
func (s *Server) insertCrrt(c *gin.Context) {
	var req struct {
		PatientId int64  `json:"patientId"`
		WardId    int64  `json:"wardId"`
		MachineId int64  `json:"machineId"` // 0=自动选
		StartAt   string `json:"startAt"`   // RFC3339 或 YYYY-MM-DD HH:MM
		EndAt     string `json:"endAt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.PatientId == 0 || req.WardId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "需提供 patientId 与 wardId"})
		return
	}
	start, ok := parseTime(req.StartAt)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startAt 格式应为 YYYY-MM-DD HH:MM 或 RFC3339"})
		return
	}
	var endPtr *time.Time
	if req.EndAt != "" {
		if e, ok2 := parseTime(req.EndAt); ok2 {
			endPtr = &e
		}
	}
	rec, err := service.InsertCrrt(s.DB, tenantOf(c), req.PatientId, req.WardId, req.MachineId, start, endPtr)
	if err == service.ErrNoSlot || err == service.ErrOccupied {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "shift": rec})
}

// listCrrt GET /schedule/crrt?date= —— 列出与某日交叠的 CRRT 占用。
func (s *Server) listCrrt(c *gin.Context) {
	day := time.Now()
	if v := c.Query("date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			day = t
		}
	}
	list, err := service.ListCrrt(s.DB, tenantOf(c), day)
	if err != nil {
		failInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": len(list), "items": list})
}

// makeup POST /patients/:id/makeup —— 为病人补排应排未排的次(决策 13)。
func (s *Server) makeup(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的病人ID"})
		return
	}
	var req struct {
		WeekStart string `json:"weekStart"`
		Weeks     int    `json:"weeks"`
	}
	_ = c.ShouldBindJSON(&req)
	weeks := req.Weeks
	if weeks != 2 && weeks != 4 {
		weeks = 2
	}
	start := sched.MondayOf(time.Now())
	if req.WeekStart != "" {
		if t, e := time.Parse("2006-01-02", req.WeekStart); e == nil {
			start = t
		}
	}
	res, e := service.MakeupPatient(s.DB, tenantOf(c), pid, start, weeks)
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// parseTime 解析 RFC3339 或 "YYYY-MM-DD HH:MM"。无时区的按本地时区解释(避免被当成 UTC 偏移)。
func parseTime(s string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04", s, time.Local); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return t, true
	}
	return time.Time{}, false
}

// listConflicts GET /conflicts?status=0 —— 列出冲突/待处理队列。
func (s *Server) listConflicts(c *gin.Context) {
	q := s.DB.Where(`"TenantId" = ?`, tenantOf(c))
	if v := c.Query("status"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q = q.Where(`"Status" = ?`, n)
		}
	}
	var list []model.ConflictQueue
	if err := q.Order(`"Severity" DESC, "Id"`).Find(&list).Error; err != nil {
		failInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": len(list), "conflicts": list})
}

// seedDemo POST /admin/seed —— 写入演示数据(空库时)。
func (s *Server) seedDemo(c *gin.Context) {
	n, err := seed.Demo(s.DB, tenantOf(c))
	if err != nil {
		failInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"seeded": n})
}
