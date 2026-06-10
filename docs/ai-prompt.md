你正在维护一个透析排班系统（ai-hms_qhd），包含 Go 后端和 React 前端。

## 项目结构
- `ai-hms-backend/` — Go 后端（Gin + GORM + PostgreSQL）
- `ai-hms-frontend/` — React 前端（Vite + TypeScript + Ant Design 6 + Tailwind 4）
- `docs/排班模块交接文档.md` — 完整交接文档（数据库/API/状态机），**先读这个**
- `docs/README.md` — 文档索引

## 排班核心代码（后端）
```
internal/smart_schedule/
├── api/api.go           # v2 API 路由（/api/v2）
├── model/models.go      # 14 张表的 GORM 模型
├── repo/repo.go         # 数据访问 + 索引
├── sched/constants.go   # 状态机/频率/机型/冲突常量
├── service/             # 业务逻辑层
│   ├── schedule_service.go   # 排班生成
│   ├── weekview.go           # 周看板
│   ├── ops_service.go        # 确认/取消/缺席/移床
│   ├── treatment_service.go  # ★上机/下机+院感闸门
│   ├── crrt_service.go       # CRRT
│   ├── perturb_service.go    # 停机/假日/方案变更
│   └── lifecycle_service.go  # 入科/出组/院感
```

## 排班前端代码
- `src/pages/SmartSchedulePage.tsx` — 排班管理主页面
- `src/services/smartScheduleApi.ts` — v2 API 封装
- `src/layouts/Sidebar.tsx` — 侧边栏（排班入口 `/schedule`）

## 硬规则（不可违反）
1. **禁止 AutoMigrate** — 全局禁用，DDL 需手动 SQL
2. **GORM**: `SingularTable: true, NoLowerCase: true`，表名列名大小写敏感，SQL 必须双引号
3. **老库列名**: 排班日期用 `"TreatmentTime"`（不是 ScheduleDate），机器表用 `"Schedule_Bed"`（不是 Schedule_Machine）
4. **NOT NULL**: `ShiftId/MachineId` 是 `int64 NOT NULL DEFAULT 0`，判空用 `> 0` 不用 `IS NOT NULL`
5. **响应包装**: `responseWrapper()` 自动将裸 JSON 包为 `{success, data, timestamp}`
6. **租户**: 硬编码 `LegacyTenantID = 3`，由 JWT 中间件注入 Context
7. **前端 TypeScript strict**: `strict/noUnusedLocals/noUnusedParameters` 全部开启

## 当前系统状态
- 365 在科患者，360 有排班骨架，30,461 排班记录，27 台机器
- 冲突队列 2,053 条开放（历史数据，非代码 bug）
- 质量评分 0 分（因 usedSlots 含历史重叠数据）
- 登录: `TEST_AI_HMS_admin / Test@123456`
- 后端: `cd ai-hms-backend && go run ./cmd/server`（端口 8080）
- 前端: `cd ai-hms-frontend && npm run dev`（端口 5173）
- 验证: 后端 `go build ./cmd/server && go test ./internal/...`，前端 `npm run build && npm run lint`

## 排班状态机
```
0(待排) → 10(草稿) → 20(已确认) → 50(透析中) → 60(已完成)
                                      ↑                ↑
                              POST /shifts/:id/start   POST /shifts/:id/complete
                              含院感闸门: negative→放行, positive仅C区, unknown需豁免
```

## 上次会话完成
- 补全上机/下机 API（排班模块 treatment_service.go + 主系统同步钩子）
- 修复 8 项 GORM 错误忽略 bug
- 前端补全格子菜单上机/下机按钮 + API 封装
- 数据库一致性 14 项检查全部通过

## 文档导航
先读 `docs/排班模块交接文档.md`，再读 `docs/sql/v2_merge_legacy.sql`，需要时参考 `docs/排班功能说明/透析排班-backend-v1.4/`（v1.4 独立参考后端）。
