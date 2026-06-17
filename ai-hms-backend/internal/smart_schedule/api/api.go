// Package api 提供 Gin HTTP 路由:排班生成、周视图、冲突队列、演示数据。
package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/repo"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/sched"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/seed"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/service"
)

// isDevSuperuser:开发模式下,未带 X-Role 的请求放行(便于本地联调)。
// release 模式下强制忽略 DEV_SUPERUSER，避免包初始化早于 gin.SetMode 导致防护失效。
func isDevSuperuser() bool {
	return strings.EqualFold(os.Getenv("DEV_SUPERUSER"), "true") && gin.Mode() != gin.ReleaseMode
}

// Server 持有数据库句柄。
type Server struct {
	DB *gorm.DB
}

// NewServer 构造函数，由调用方注入 *gorm.DB。
func NewServer(db *gorm.DB) *Server {
	return &Server{DB: db}
}

// Register 注册路由到给定的 RouterGroup 下。
func (s *Server) Register(rg *gin.RouterGroup) {
	// 老库守则禁止运行时 DDL：此处只校验排班唯一索引是否已由 DBA 预建，
	// 缺失则告警并在 /schedule/health 标红，绝不自动创建（见 scripts/schedule_unique_indexes.sql）。
	if missing, err := repo.VerifyIndexes(s.DB); err != nil {
		log.Printf("[smart_schedule] 校验排班唯一索引失败: %v", err)
	} else if len(missing) > 0 {
		log.Printf("[smart_schedule] 缺少排班唯一索引 %v —— 请由 DBA 执行 scripts/schedule_unique_indexes.sql 创建；缺失期间并发排班可能产生重复行（应用不会自动建索引）", missing)
	}

	if isDevSuperuser() {
		log.Println("[smart_schedule] ⚠️  DEV_SUPERUSER=true —— 所有角色校验已绕过！请确认仅在本地开发环境使用，生产环境严禁开启。")
	} else if strings.EqualFold(os.Getenv("DEV_SUPERUSER"), "true") && gin.Mode() == gin.ReleaseMode {
		log.Println("[smart_schedule] DEV_SUPERUSER=true 已在 release 模式下被强制忽略。")
	}

	// responseWrapper 必须先于 tenantMiddleware 注册：中间件按注册顺序执行，
	// 只有 wrapper 先安装缓冲 writer，tenantMiddleware/guard 的 401/403 错误响应
	// 才能被拦截并归一为统一错误格式（否则它们绕过 wrapper 直写原始 writer）。
	rg.Use(responseWrapper())
	rg.Use(tenantMiddleware())

	// 只读
	rg.GET("/schedule/week", s.weekView)
	rg.GET("/schedule/board", s.board)
	rg.GET("/schedule/diffs", s.diffs)
	rg.GET("/schedule/quality", s.quality)
	rg.GET("/schedule/health", s.healthCheck)
	rg.GET("/conflicts", s.listConflicts)
	// 生成 / 演示数据(管理类:护士长/主班)
	rg.POST("/schedule/generate", guard(RoleHeadNurse, RoleChargeNurse), s.generate)
	rg.POST("/admin/seed", guard(RoleHeadNurse), s.seedDemo)
	// 确认(① 护士长;②③ 护士长/主班)
	rg.POST("/schedule/confirm-plan", guard(RoleHeadNurse), s.confirmPlan)
	rg.POST("/schedule/confirm-day", guard(RoleHeadNurse, RoleChargeNurse), s.confirmDay)
	// 排班调整(护士长/主班)
	rg.POST("/shifts/:id/cancel", guard(RoleHeadNurse, RoleChargeNurse), s.cancelShift)
	rg.POST("/shifts/:id/absent", guard(RoleHeadNurse, RoleChargeNurse), s.absentShift)
	rg.POST("/shifts/:id/move", guard(RoleHeadNurse, RoleChargeNurse), s.moveShift)
	// 治疗执行(上机/下机,护士/主班/护士长均可)
	rg.POST("/shifts/:id/start", guard(RoleNurse, RoleChargeNurse, RoleHeadNurse), s.startTreatment)
	rg.POST("/shifts/:id/complete", guard(RoleNurse, RoleChargeNurse, RoleHeadNurse), s.completeTreatment)
	// 临时透析 / CRRT(医嘱 + 护士长/主班确认)
	rg.POST("/schedule/temporary", guard(RoleDoctor, RoleHeadNurse, RoleChargeNurse), s.insertTemporary)
	rg.POST("/schedule/crrt", guard(RoleDoctor, RoleHeadNurse, RoleChargeNurse), s.insertCrrt)
	rg.GET("/schedule/crrt", s.listCrrt)
	// 停机(工程师/护士长)、假日(机构)、方案变更(医生/护士长)
	rg.POST("/machines/:id/outage", guard(RoleHeadNurse), s.machineOutage)
	rg.POST("/schedule/holiday", guard(RoleHeadNurse), s.setHoliday)
	rg.POST("/patients/:id/plan-change", guard(RoleDoctor, RoleHeadNurse), s.planChange)
	// 补透(护士长/主班)
	rg.POST("/patients/:id/makeup", guard(RoleHeadNurse, RoleChargeNurse), s.makeup)

	// 医护人力排班·月基线(④ v1)：主任排医生/护士长排护士
	rg.POST("/staff-duty", guard(RoleHeadNurse, RoleDoctor), s.upsertStaffDuty)
	rg.GET("/staff-duty", guard(RoleHeadNurse, RoleDoctor, RoleChargeNurse), s.listStaffDuty)
	rg.DELETE("/staff-duty/:id", guard(RoleHeadNurse, RoleDoctor), s.deleteStaffDuty)
	rg.GET("/duty/resolve", s.resolveDuty)
	// 日覆盖 + 接班（④ v2）
	rg.POST("/staff-duty/override", guard(RoleHeadNurse, RoleChargeNurse, RoleDoctor), s.createOverride)
	rg.GET("/duty/my-duties", s.myDuties)
	rg.POST("/duty/check-in", s.checkIn)
	rg.GET("/duty/check-in/status", s.checkInStatus)

	s.registerAdmin(rg) // 资源/病人/骨架/模板维护 + 冲突处理(P1)
}

// tenantMiddleware 从主系统 JWT context 取租户与角色。
// 由外部 AuthMiddleware 保证 context 已设置 tenant/userId/roles。
func tenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant := middleware.GetTenantID(c)
		if tenant <= 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "error": "缺少有效租户"})
			return
		}
		c.Set("tenant", tenant)
		c.Set("role", mapRole(middleware.GetRoles(c)))
		c.Set("userId", middleware.GetUserID(c))
		c.Next()
	}
}

func mapRole(roles []string) string {
	for _, r := range roles {
		// 管理员角色集合复用 middleware.AdminRoles 单一真值源，避免角色名字面量重复。
		if middleware.IsAdminRole(r) {
			return RoleHeadNurse
		}
		switch r {
		case "DOCTOR", "医生":
			return RoleDoctor
		case "HEAD_NURSE", "护士长":
			return RoleHeadNurse
		case "CHARGE_NURSE", "主班护士":
			return RoleChargeNurse
		case "NURSE", "护士":
			return RoleNurse
		}
	}
	return ""
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
			if isDevSuperuser() {
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

// bufferedWriter 捕获 handler 写入的响应体，延迟到中间件统一包装后写出。
type bufferedWriter struct {
	gin.ResponseWriter
	buf        bytes.Buffer
	statusCode int
}

func (w *bufferedWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *bufferedWriter) WriteHeader(code int) {
	w.statusCode = code
}

func (w *bufferedWriter) WriteHeaderNow() {}

// responseWrapper 将 v2 API 的裸 JSON 响应包装为与主系统 v1 一致的统一格式：
//   - 2xx: { success:true, data, timestamp }
//   - 4xx/5xx: { success:false, error:{ code, message }, timestamp }
//
// 错误归一是关键：v2 handler/中间件历史上直写 {code:403,error:"..."} 或 {error:"..."}，
// 前端 restRequest/getErrorMessage 只识别 error.message，导致排班操作失败时
// 用户只能看到兜底文案。此处集中归一，handler 无需逐个改造。
// 响应体已含 success 字段时原样透传（兼容已统一格式的输出，避免二次包装）。
func responseWrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		bw := &bufferedWriter{ResponseWriter: c.Writer, statusCode: http.StatusOK}
		c.Writer = bw
		c.Next()
		orig := bw.ResponseWriter
		if bw.buf.Len() > 0 {
			var data interface{}
			if err := json.Unmarshal(bw.buf.Bytes(), &data); err == nil {
				m, isObj := data.(map[string]interface{})
				// 已是统一格式则原样透传
				if isObj {
					if _, has := m["success"]; has {
						orig.WriteHeader(bw.statusCode)
						orig.Write(bw.buf.Bytes())
						return
					}
				}
				switch {
				case bw.statusCode >= 200 && bw.statusCode < 300:
					wrapped, _ := json.Marshal(gin.H{
						"success":   true,
						"data":      data,
						"timestamp": time.Now().Format(time.RFC3339),
					})
					writeWrapped(orig, bw.statusCode, wrapped)
					return
				case bw.statusCode >= 400:
					code, msg := normalizeError(m, isObj, bw.statusCode)
					wrapped, _ := json.Marshal(gin.H{
						"success":   false,
						"error":     gin.H{"code": code, "message": msg},
						"timestamp": time.Now().Format(time.RFC3339),
					})
					writeWrapped(orig, bw.statusCode, wrapped)
					return
				}
			}
		}
		orig.WriteHeader(bw.statusCode)
		if bw.buf.Len() > 0 {
			orig.Write(bw.buf.Bytes())
		}
	}
}

func writeWrapped(w gin.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(body)
}

// normalizeError 从历史错误体（{code:403,error:"..."} / {error:"..."} / {message:"..."}）
// 提取 code 与 message；缺失时回退 HTTP 状态码与标准状态文案。
func normalizeError(m map[string]interface{}, isObj bool, status int) (string, string) {
	code := strconv.Itoa(status)
	msg := http.StatusText(status)
	if !isObj {
		return code, msg
	}
	if v, ok := m["error"].(string); ok && strings.TrimSpace(v) != "" {
		msg = v
	} else if v, ok := m["message"].(string); ok && strings.TrimSpace(v) != "" {
		msg = v
	}
	switch v := m["code"].(type) {
	case float64: // JSON 数字反序列化为 float64
		code = strconv.Itoa(int(v))
	case string:
		if strings.TrimSpace(v) != "" {
			code = v
		}
	}
	return code, msg
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
	// mon.AddDate(0,0,7) = 下周一 00:00, BETWEEN 含周日全天
	sun := mon.AddDate(0, 0, 7)

	var shifts []model.PatientShift
	err := s.DB.Where(`"TenantId" = ? AND "TreatmentTime" >= ? AND "TreatmentTime" < ?`, tenantOf(c), mon, sun).
		Order(`"TreatmentTime", "ShiftId", "WardId", "MachineId"`).
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
	if v := c.GetString("userId"); v != "" {
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
	case service.ErrLocked, service.ErrOccupied, service.ErrDoubleBook, service.ErrModeMismatch,
		service.ErrNotConfirmed, service.ErrNotToday, service.ErrNotInDialysis,
		service.ErrInfectionUnconfirmed, service.ErrInfectionPositive:
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
		Date        string `json:"date"`
		Mode        int16  `json:"mode"`        // 10全院停 / 20假日值班
		OpenWardIds string `json:"openWardIds"` // 值班模式下开放的病区ID,逗号分隔(空=全开)
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
	res, e := service.SetHoliday(s.DB, tenantOf(c), d, mode, req.OpenWardIds)
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

// startTreatment POST /shifts/:id/start —— 上机(已确认→透析中,含院感闸门+当日校验)。
func (s *Server) startTreatment(c *gin.Context) {
	id, ok := shiftIdParam(c)
	if !ok {
		return
	}
	if err := service.StartTreatment(s.DB, tenantOf(c), userOf(c), id); err != nil {
		opError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已上机"})
}

// completeTreatment POST /shifts/:id/complete —— 下机(透析中→已完成)。
func (s *Server) completeTreatment(c *gin.Context) {
	id, ok := shiftIdParam(c)
	if !ok {
		return
	}
	if err := service.CompleteTreatment(s.DB, tenantOf(c), userOf(c), id); err != nil {
		opError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已下机"})
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

// quality GET /schedule/quality?date=&weeks= —— 排班质量评分(达标率/利用率/稳定率/综合分)。
func (s *Server) quality(c *gin.Context) {
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
	res, err := service.ComputeQuality(s.DB, tenantOf(c), sched.MondayOf(day), weeks)
	if err != nil {
		failInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// listConflicts GET /conflicts?status=0 —— 列出冲突/待处理队列。
func (s *Server) listConflicts(c *gin.Context) {
	q := s.DB.Where(`"TenantId" = ?`, tenantOf(c))
	if v := c.Query("status"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q = q.Where(`"Status" = ?`, n)
		}
	}
	var total int64
	q.Model(&model.ConflictQueue{}).Count(&total)
	limit, offset := pageParams(c, 100)
	var list []model.ConflictQueue
	if err := q.Order(`"Severity" DESC, "Id"`).Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		failInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "count": len(list), "limit": limit, "offset": offset, "conflicts": list})
}

// pageParams 解析分页参数 limit/offset(limit 上限 500,默认 def)。
func pageParams(c *gin.Context, def int) (int, int) {
	limit, offset := def, 0
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 && v <= 500 {
		limit = v
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil && v >= 0 {
		offset = v
	}
	return limit, offset
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

// healthCheck GET /schedule/health —— 数据健康检查。
func (s *Server) healthCheck(c *gin.Context) {
	h, err := service.ComputeHealth(s.DB, tenantOf(c))
	if err != nil {
		failInternal(c, err)
		return
	}
	// 实时校验排班唯一索引是否缺失（只读），缺失则并入健康检查告警，便于运维发现需 DBA 建索引。
	if missing, verr := repo.VerifyIndexes(s.DB); verr == nil {
		for _, idx := range missing {
			h.Warnings = append(h.Warnings,
				"缺少唯一索引 "+idx+"（需 DBA 执行 scripts/schedule_unique_indexes.sql），并发排班存在重复风险")
		}
	}
	c.JSON(http.StatusOK, h)
}

// upsertStaffDuty POST /staff-duty —— 建/改一条月基线排班。
func (s *Server) upsertStaffDuty(c *gin.Context) {
	var req struct {
		StaffId   int64  `json:"staffId"`
		StaffName string `json:"staffName"`
		DutyRole  string `json:"dutyRole"`
		WardId    int64  `json:"wardId"`
		DutyDate  string `json:"dutyDate"`
		Shift     string `json:"shift"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	day, err := time.Parse("2006-01-02", req.DutyDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dutyDate 格式应为 yyyy-MM-dd"})
		return
	}
	res, e := service.UpsertStaffDuty(s.DB, tenantOf(c), service.StaffDutyInput{
		StaffId: req.StaffId, StaffName: req.StaffName, DutyRole: req.DutyRole,
		WardId: req.WardId, DutyDate: day, Shift: req.Shift,
	}, userOf(c))
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// listStaffDuty GET /staff-duty?wardId=&month=YYYY-MM —— 查某室某月排班。
func (s *Server) listStaffDuty(c *gin.Context) {
	wardId, _ := strconv.ParseInt(c.Query("wardId"), 10, 64)
	if wardId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wardId 必填"})
		return
	}
	monthStart, err := time.Parse("2006-01", c.Query("month"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month 格式应为 yyyy-MM"})
		return
	}
	monthEnd := monthStart.AddDate(0, 1, 0)
	rows, e := service.ListStaffDuty(s.DB, tenantOf(c), wardId, monthStart, monthEnd)
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// deleteStaffDuty DELETE /staff-duty/:id
func (s *Server) deleteStaffDuty(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	if e := service.DeleteStaffDuty(s.DB, tenantOf(c), id); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// resolveDuty GET /duty/resolve?wardId=&date=YYYY-MM-DD&dutyRole= —— 解析当班人。
func (s *Server) resolveDuty(c *gin.Context) {
	wardId, _ := strconv.ParseInt(c.Query("wardId"), 10, 64)
	day, err := time.Parse("2006-01-02", c.Query("date"))
	if wardId <= 0 || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wardId 必填、date 格式 yyyy-MM-dd"})
		return
	}
	res, e := service.ResolveDuty(s.DB, tenantOf(c), wardId, day, c.Query("dutyRole"))
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// createOverride POST /staff-duty/override —— 当日覆盖，不动月基线。
func (s *Server) createOverride(c *gin.Context) {
	var req struct {
		DutyDate        string `json:"dutyDate"`
		WardId          int64  `json:"wardId"`
		DutyRole        string `json:"dutyRole"`
		OriginalStaffId int64  `json:"originalStaffId"`
		ActualStaffId   int64  `json:"actualStaffId"`
		ActualStaffName string `json:"actualStaffName"`
		Reason          string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	day, err := time.Parse("2006-01-02", req.DutyDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dutyDate 格式应为 yyyy-MM-dd"})
		return
	}
	res, e := service.CreateOverride(s.DB, tenantOf(c), service.OverrideInput{
		DutyDate: day, WardId: req.WardId, DutyRole: req.DutyRole,
		OriginalStaffId: req.OriginalStaffId, ActualStaffId: req.ActualStaffId,
		ActualStaffName: req.ActualStaffName, Reason: req.Reason,
	}, userOf(c))
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// myDuties GET /duty/my-duties?date= —— 今日我被排/被顶的(室,角色)列表。
func (s *Server) myDuties(c *gin.Context) {
	day, err := time.Parse("2006-01-02", c.Query("date"))
	if err != nil {
		day = time.Now()
	}
	res, e := service.ResolveMyDuties(s.DB, tenantOf(c), userOf(c), day)
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// checkIn POST /duty/check-in —— 接班激活。
func (s *Server) checkIn(c *gin.Context) {
	var req struct {
		WardId       int64  `json:"wardId"`
		ShiftId      int64  `json:"shiftId"`
		OperatorType int64  `json:"operatorType"`
		Type         int64  `json:"type"`
		Note         string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	res, e := service.CheckIn(s.DB, tenantOf(c), userOf(c), req.WardId, req.ShiftId, req.OperatorType, req.Type, req.Note)
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// checkInStatus GET /duty/check-in/status?date= —— 今日是否已接班。
func (s *Server) checkInStatus(c *gin.Context) {
	day, err := time.Parse("2006-01-02", c.Query("date"))
	if err != nil {
		day = time.Now()
	}
	ok, e := service.IsCheckedIn(s.DB, tenantOf(c), userOf(c), day)
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"checkedIn": ok})
}
