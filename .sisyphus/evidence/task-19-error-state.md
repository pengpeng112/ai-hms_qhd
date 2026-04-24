# T19 Error-state separation

## 已修改
- `ai-hms-frontend/src/services/restClient.ts`
  - 401/403 统一触发登录跳转
  - 新增请求错误分类与治疗加载错误文案
- `ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx`
  - 区分 404 / 500 / 网络异常 / 认证错误
  - 404 显示“暂无治疗记录”并提供创建按钮
  - 500/网络异常显示重试提示
  - 阻止真实失败后自动创建治疗记录
- `ai-hms-backend/internal/api/v1/treatment_handler.go`
  - 患者当日治疗不存在时返回 404

## 验证
- LSP diagnostics：已通过（变更文件无错误）
- 后端：`go test ./...` 通过
- 前端：`npm run build` 通过
