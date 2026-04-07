# Codex 修复计划 —— Demock 二期遗留问题修复

> 审核日期：2026-04-07
> 基于 DEMOCK_CHANGELOG.md 中已完成任务的代码审查结果
> 所有任务均可独立执行，无依赖顺序要求（除非特别说明）

---

## 优先级说明

| 级别 | 标识 | 含义 |
|------|------|------|
| 严重 | 🔴 CRITICAL | 数据隔离漏洞或生产安全问题，必须修复 |
| 中等 | 🟡 MEDIUM | 并发安全或类型安全问题，强烈建议修复 |
| 低   | 🟢 LOW | 代码规范或防御性编程改进 |

---

## FIX-1 🔴 CRITICAL — statistics_service.go：所有统计接口缺少 tenant_id 隔离

### 问题描述

`ai-hms-backend/internal/services/statistics_service.go` 中 4 个统计方法均未接收 `tenantId` 参数，直接对全表查询，**导致多租户环境下跨租户数据泄露**。

涉及方法：
- `QualityByYear(year int)` — 查 `lab_report_items` 表，无 tenant 过滤
- `InfectionByYear(year int)` — 查 `infection_infos` 表，无 tenant 过滤
- `VascularByYear(year int)` — 查 `vascular_accesses` 表，无 tenant 过滤
- `WorkloadByYearMonth(yearMonth string)` — 查 `treatments` 和 `users` 表，无 tenant 过滤

### 修复步骤

**文件：`ai-hms-backend/internal/services/statistics_service.go`**

1. 将 `QualityByYear(year int)` 签名改为 `QualityByYear(tenantId int64, year int)`，在 DB 查询链中增加 `.Where("tenant_id = ?", tenantId)`：

   ```go
   // 修改前（第 56-58 行）：
   if err := database.GetDB().
       Where("tested_at >= ? AND tested_at < ?", start, end).
       Find(&rows).Error; err != nil {

   // 修改后：
   if err := database.GetDB().
       Where("tenant_id = ? AND tested_at >= ? AND tested_at < ?", tenantId, start, end).
       Find(&rows).Error; err != nil {
   ```

2. 将 `InfectionByYear(year int)` 签名改为 `InfectionByYear(tenantId int64, year int)`，在 DB 查询链中增加 tenant 过滤：

   ```go
   // 修改前（第 123-125 行）：
   if err := database.GetDB().
       Where("update_date >= ? AND update_date < ?", start, end).
       Find(&rows).Error; err != nil {

   // 修改后：
   if err := database.GetDB().
       Where("tenant_id = ? AND update_date >= ? AND update_date < ?", tenantId, start, end).
       Find(&rows).Error; err != nil {
   ```

3. 将 `VascularByYear(year int)` 签名改为 `VascularByYear(tenantId int64, year int)`，在 DB 查询链中增加 tenant 过滤：

   ```go
   // 修改前（第 162-165 行）：
   if err := database.GetDB().
       Where("created_at >= ? AND created_at < ?", start, end).
       Find(&rows).Error; err != nil {

   // 修改后：
   if err := database.GetDB().
       Where("tenant_id = ? AND created_at >= ? AND created_at < ?", tenantId, start, end).
       Find(&rows).Error; err != nil {
   ```

4. 将 `WorkloadByYearMonth(yearMonth string)` 签名改为 `WorkloadByYearMonth(tenantId int64, yearMonth string)`，并在两处 DB 查询中增加 tenant 过滤：

   ```go
   // 修改前（第 195-198 行）：
   if err := database.GetDB().
       Where("treatment_date >= ? AND treatment_date < ?", start, end).
       Find(&treatments).Error; err != nil {

   // 修改后：
   if err := database.GetDB().
       Where("tenant_id = ? AND treatment_date >= ? AND treatment_date < ?", tenantId, start, end).
       Find(&treatments).Error; err != nil {

   // 修改前（第 212-213 行）：
   if err := database.GetDB().Find(&users).Error; err == nil {

   // 修改后：
   if err := database.GetDB().Where("tenant_id = ?", tenantId).Find(&users).Error; err == nil {
   ```

### 编译验证
```bash
cd ai-hms_qhd/ai-hms-backend
go build ./...
go vet ./...
```

---

## FIX-2 🔴 CRITICAL — statistics_handler.go：Handler 未传递 tenantId 给 Service

### 问题描述

`ai-hms-backend/internal/api/v1/statistics_handler.go` 中 4 个 Handler 在调用 Service 时未从请求上下文提取 `tenantId`（其他 Handler 如 `clinical_task_handler.go` 均使用 `middleware.GetTenantID(c)`）。

**此任务必须在 FIX-1 完成后执行**（因为 FIX-1 修改了 Service 方法签名）。

### 修复步骤

**文件：`ai-hms-backend/internal/api/v1/statistics_handler.go`**

1. 在文件 `import` 块中增加 middleware 包导入：

   ```go
   // 修改前：
   import (
       "fmt"
       "time"

       "github.com/elliotxin/ai-hms-backend/internal/services"
       "github.com/elliotxin/ai-hms-backend/pkg/response"
       "github.com/gin-gonic/gin"
   )

   // 修改后：
   import (
       "fmt"
       "time"

       "github.com/elliotxin/ai-hms-backend/internal/middleware"
       "github.com/elliotxin/ai-hms-backend/internal/services"
       "github.com/elliotxin/ai-hms-backend/pkg/response"
       "github.com/gin-gonic/gin"
   )
   ```

2. 修改 `Quality` Handler（第 20-28 行）：

   ```go
   // 修改前：
   func (h *StatisticsHandler) Quality(c *gin.Context) {
       year := parseYear(c.Query("year"))
       items, err := h.service.QualityByYear(year)

   // 修改后：
   func (h *StatisticsHandler) Quality(c *gin.Context) {
       tenantId := middleware.GetTenantID(c)
       year := parseYear(c.Query("year"))
       items, err := h.service.QualityByYear(tenantId, year)
   ```

3. 修改 `Infection` Handler（第 30-38 行）：

   ```go
   // 修改前：
   func (h *StatisticsHandler) Infection(c *gin.Context) {
       year := parseYear(c.Query("year"))
       items, err := h.service.InfectionByYear(year)

   // 修改后：
   func (h *StatisticsHandler) Infection(c *gin.Context) {
       tenantId := middleware.GetTenantID(c)
       year := parseYear(c.Query("year"))
       items, err := h.service.InfectionByYear(tenantId, year)
   ```

4. 修改 `Vascular` Handler（第 40-48 行）：

   ```go
   // 修改前：
   func (h *StatisticsHandler) Vascular(c *gin.Context) {
       year := parseYear(c.Query("year"))
       items, err := h.service.VascularByYear(year)

   // 修改后：
   func (h *StatisticsHandler) Vascular(c *gin.Context) {
       tenantId := middleware.GetTenantID(c)
       year := parseYear(c.Query("year"))
       items, err := h.service.VascularByYear(tenantId, year)
   ```

5. 修改 `Workload` Handler（第 50-58 行）：

   ```go
   // 修改前：
   func (h *StatisticsHandler) Workload(c *gin.Context) {
       yearMonth := c.DefaultQuery("yearMonth", time.Now().Format("2006-01"))
       items, err := h.service.WorkloadByYearMonth(yearMonth)

   // 修改后：
   func (h *StatisticsHandler) Workload(c *gin.Context) {
       tenantId := middleware.GetTenantID(c)
       yearMonth := c.DefaultQuery("yearMonth", time.Now().Format("2006-01"))
       items, err := h.service.WorkloadByYearMonth(tenantId, yearMonth)
   ```

### 编译验证
```bash
cd ai-hms_qhd/ai-hms-backend
go build ./...
go vet ./...
```

---

## FIX-3 🟡 MEDIUM — Monitoring.tsx：OrderListModal 中硬编码医嘱数据残留

### 问题描述

`ai-hms-frontend/src/pages/Monitoring.tsx` 中 `OrderListModal` 组件（第 560 行起）仍使用 `useState` 硬编码初始化医嘱列表：

- 第 571-596 行：`longOrders` 含 3 条硬编码医嘱，其中时间字段为 `'2025-12-01 08:30'` 和 `'2025-12-01 08:31'`，医生字段为 `'王医生'`
- 第 598-615 行：`tempOrders` 含 2 条硬编码医嘱，医生字段为 `'王医生'`

> 注：内容字段出现乱码（GBK 编码的药品名被错误解读），属于后续接入真实 API 时的衍生问题，本次一并清除。

### 修复步骤

**文件：`ai-hms-frontend/src/pages/Monitoring.tsx`**

将 `OrderListModal` 组件内 `longOrders` 和 `tempOrders` 两个 `useState` 硬编码初始值改为空数组，并增加 TODO 注释：

```tsx
// 修改前（第 571-615 行）：
const [longOrders] = useState<OrderItem[]>([
  {
    id: 1,
    content: '0.9% 姘寲閽犳敞灏勬恫 100ml',
    frequency: '姣忔閫忔瀽',
    doctor: '王医生',
    time: '2025-12-01 08:30',
    status: 'ACTIVE'
  },
  // ... 其余硬编码项
])

const [tempOrders] = useState<OrderItem[]>([
  {
    id: 4,
    content: '50% 钁¤悇绯栨敞灏勬恫 20ml iv',
    // ...
    doctor: '王医生',
    time: currentOrderTime,
    status: 'EXECUTED'
  },
  // ...
])

// 修改后：
// TODO: 从 /api/v1/patients/:id/orders?type=LONG 加载长期医嘱
const [longOrders] = useState<OrderItem[]>([])

// TODO: 从 /api/v1/patients/:id/orders?type=TEMP 加载临时医嘱
const [tempOrders] = useState<OrderItem[]>([])
```

### 编译验证
```bash
cd ai-hms_qhd/ai-hms-frontend
npx.cmd tsc --noEmit
```

---

## FIX-4 🟡 MEDIUM — permission_service.go：ensureDefaultsInitialized 缺少并发保护

### 问题描述

`ai-hms-backend/internal/services/permission_service.go` 中 `ensureDefaultsInitialized()` 方法每次请求到达时都会触发 `InitDefaultPermissions()`，没有"只执行一次"的保护机制。在高并发启动时（如前端同时发起多个 API 请求），会有**多个 goroutine 同时进入初始化事务**，虽然 SQL 层的唯一索引会防止数据重复，但会导致不必要的数据库压力，以及在 SQLite（测试场景）下可能触发锁竞争。

### 修复步骤

**文件：`ai-hms-backend/internal/services/permission_service.go`**

1. 在文件顶部 `import` 块中增加 `"sync"` 包导入：

   ```go
   // 修改前：
   import (
       "errors"
       "strings"

       "github.com/elliotxin/ai-hms-backend/internal/database"
       "github.com/elliotxin/ai-hms-backend/internal/models"
       "gorm.io/gorm"
   )

   // 修改后：
   import (
       "errors"
       "strings"
       "sync"

       "github.com/elliotxin/ai-hms-backend/internal/database"
       "github.com/elliotxin/ai-hms-backend/internal/models"
       "gorm.io/gorm"
   )
   ```

2. 在 `PermissionService` 结构体下方（第 16 行之后）声明包级 `sync.Once` 变量：

   ```go
   // 在 type PermissionService struct{ ... } 之后、NewPermissionService 之前插入：
   var defaultPermissionsOnce sync.Once
   ```

3. 将 `ensureDefaultsInitialized()` 方法改为使用 `sync.Once`：

   ```go
   // 修改前（第 103-107 行）：
   func (s *PermissionService) ensureDefaultsInitialized() error {
       // Always run default initialization in idempotent mode so newly added
       // permissions can be backfilled to existing environments.
       return s.InitDefaultPermissions()
   }

   // 修改后：
   func (s *PermissionService) ensureDefaultsInitialized() error {
       var initErr error
       defaultPermissionsOnce.Do(func() {
           initErr = s.InitDefaultPermissions()
       })
       return initErr
   }
   ```

> **注意**：`sync.Once` 是包级变量，进程生命周期内只执行一次。如需在新版本部署后补充新权限，运维流程为：重启服务（`Once` 会在新进程中重新执行）。这与当前每次请求都执行相比是更安全的做法。

### 编译验证
```bash
cd ai-hms_qhd/ai-hms-backend
go build ./...
go vet ./...
```

---

## FIX-5 🟡 MEDIUM — permission_service.go：NURSE_SCHEDULER 角色使用硬编码字符串

### 问题描述

`ai-hms-backend/internal/services/permission_service.go` 第 169 行使用了字符串字面量 `"NURSE_SCHEDULER"` 作为角色键，而其他角色均通过 `models.RoleXxx` 常量引用。`user.go` 中的 `const` 块缺少该常量定义。这使得角色字符串分散在代码中，重构时存在遗漏风险。

当前 `user.go` 角色常量（第 29-37 行）：
```go
const (
    RoleAdmin            = "ADMIN"
    RoleDoctorChief      = "DOCTOR_CHIEF"
    RoleDoctorSupervisor = "DOCTOR_SUPERVISOR"
    RoleDoctorDuty       = "DOCTOR_DUTY"
    RoleNurseHead        = "NURSE_HEAD"
    RoleNurseManager     = "NURSE_MANAGER"
    RoleNurseResponsible = "NURSE_RESPONSIBLE"
    RoleEngineer         = "ENGINEER"
)
```

### 修复步骤

**文件 1：`ai-hms-backend/internal/models/user.go`**

在角色常量 `const` 块末尾（`RoleEngineer` 之后）增加一行：

```go
// 修改前（第 37 行）：
    RoleEngineer        = "ENGINEER"         // 工程师
)

// 修改后：
    RoleEngineer         = "ENGINEER"          // 工程师
    RoleNurseScheduler   = "NURSE_SCHEDULER"   // 排班护士
)
```

**文件 2：`ai-hms-backend/internal/services/permission_service.go`**

将第 169 行的硬编码字符串替换为常量引用：

```go
// 修改前（第 169 行）：
        "NURSE_SCHEDULER": {

// 修改后：
        models.RoleNurseScheduler: {
```

### 编译验证
```bash
cd ai-hms_qhd/ai-hms-backend
go build ./...
go vet ./...
```

---

## FIX-6 🟢 LOW — statistics_handler.go：parseYear() 缺少年份上界校验

### 问题描述

`ai-hms-backend/internal/api/v1/statistics_handler.go` 中 `parseYear()` 函数（第 60-69 行）当前只校验 `year <= 0`，未校验超大年份（如 `year=9999`）。查询跨越99年的时间范围不会导致错误，但会引发全表扫描性能问题，且在某些 PostgreSQL 日期函数中可能溢出。

建议将合法年份限制在 `[2000, currentYear+1]` 范围内。

### 修复步骤

**文件：`ai-hms-backend/internal/api/v1/statistics_handler.go`**

```go
// 修改前（第 60-69 行）：
func parseYear(value string) int {
    if value == "" {
        return time.Now().Year()
    }
    var year int
    if _, err := fmt.Sscanf(value, "%d", &year); err != nil || year <= 0 {
        return time.Now().Year()
    }
    return year
}

// 修改后：
func parseYear(value string) int {
    now := time.Now().Year()
    if value == "" {
        return now
    }
    var year int
    if _, err := fmt.Sscanf(value, "%d", &year); err != nil || year < 2000 || year > now+1 {
        return now
    }
    return year
}
```

### 编译验证
```bash
cd ai-hms_qhd/ai-hms-backend
go build ./...
go vet ./...
```

---

## 全量验证清单（所有 FIX 完成后执行）

```bash
# 后端
cd ai-hms_qhd/ai-hms-backend
go build ./...
go vet ./...

# 前端
cd ai-hms_qhd/ai-hms-frontend
npx.cmd tsc --noEmit

# Mock 残留扫描
cd ai-hms_qhd/ai-hms-frontend
rg -n "2025-12-01|王医生" src/pages/Monitoring.tsx
# 期望：0 命中

# 硬编码角色字符串残留扫描
cd ai-hms_qhd/ai-hms-backend
rg -n '"NURSE_SCHEDULER"' internal/
# 期望：0 命中（已用常量替换）
```

---

## 修复完成后更新 DEMOCK_CHANGELOG.md

执行完以上修复后，在 `DEMOCK_CHANGELOG.md` 末尾追加如下内容：

```markdown
## [FIX-1+2] statistics 接口 tenant_id 隔离修复
- 执行日期：{执行日期}
- 修改文件：`ai-hms-backend/internal/services/statistics_service.go`、`ai-hms-backend/internal/api/v1/statistics_handler.go`
- 变更：4 个统计方法（QualityByYear/InfectionByYear/VascularByYear/WorkloadByYearMonth）增加 tenantId 参数；所有 DB 查询加入 `tenant_id = ?` 过滤；Handler 层从 `middleware.GetTenantID(c)` 提取并传入
- 编译验证：`go build ./...` 通过

## [FIX-3] Monitoring.tsx OrderListModal 硬编码医嘱清除
- 执行日期：{执行日期}
- 修改文件：`ai-hms-frontend/src/pages/Monitoring.tsx`
- 变更：删除 longOrders/tempOrders 硬编码初始值（含 2025-12-01 日期、王医生等），改为空数组加 TODO 注释
- 编译验证：`npx.cmd tsc --noEmit` 通过

## [FIX-4] permission_service.go 初始化并发安全修复
- 执行日期：{执行日期}
- 修改文件：`ai-hms-backend/internal/services/permission_service.go`
- 变更：引入包级 sync.Once，ensureDefaultsInitialized() 仅在进程生命周期内执行一次
- 编译验证：`go build ./...` 通过

## [FIX-5] NURSE_SCHEDULER 角色常量化
- 执行日期：{执行日期}
- 修改文件：`ai-hms-backend/internal/models/user.go`、`ai-hms-backend/internal/services/permission_service.go`
- 变更：user.go 新增 RoleNurseScheduler = "NURSE_SCHEDULER"；permission_service.go 第 169 行改用常量引用
- 编译验证：`go build ./...` 通过

## [FIX-6] parseYear 年份上界校验
- 执行日期：{执行日期}
- 修改文件：`ai-hms-backend/internal/api/v1/statistics_handler.go`
- 变更：parseYear() 合法范围限定为 [2000, currentYear+1]，超出范围回退当前年份
- 编译验证：`go build ./...` 通过
```
