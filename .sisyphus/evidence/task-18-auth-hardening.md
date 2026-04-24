# T18 Auth Hardening Evidence

## 审查结论

- **内置管理员创建**：`ai-hms-backend/internal/services/auth_service.go`
  - 之前：`BUILTIN_ADMIN_USER/BUILTIN_ADMIN_PASS` 未配置时回退到代码内置 `hms_admin / Hms@Admin2024`。
  - 现在：仅当 `AUTH_EMERGENCY_ENABLED=true` 时才启用；默认关闭。
- **DEFAULT_PASSWORD 回退逻辑**：`ai-hms-backend/internal/services/auth_service.go`
  - 之前：非 release 模式下未配置 `DEFAULT_PASSWORD` 时自动回退到代码内置 `admin@123qwe`。
  - 现在：仅当 `AUTH_EMERGENCY_ENABLED=true` 时才启用；默认关闭。开启后未显式设置 `DEFAULT_PASSWORD` 仍会回退到代码内应急口令，并已在模板/文档中明确记录。
- **JWT tenant_id 要求**：`ai-hms-backend/internal/middleware/auth.go`
  - `claims.TenantID <= 0` 时返回 `403` 和 `缺少租户信息`。

## 代码与文档改动

- `ai-hms-backend/internal/services/auth_service.go`
  - 新增 `resolveEmergencyAuthEnabled`
  - 新增 `resolveBuiltinAdminCredentials`
  - `resolveBackdoorPassword` 改为受 `AUTH_EMERGENCY_ENABLED` 控制
- `ai-hms-backend/internal/services/auth_service_test.go`
  - 覆盖应急开关启用/关闭
  - 覆盖内置管理员凭据默认值与显式值
- `ai-hms-backend/internal/middleware/auth_test.go`
  - 断言缺失 `tenant_id` 时返回 403 且包含 `缺少租户信息`
- `.env.production.template`
  - 明确 `AUTH_EMERGENCY_ENABLED=false`
  - 明确内置管理员与 `DEFAULT_PASSWORD` 仅在应急开关开启时生效
- `ai-hms-backend/.env.example`
  - 增加应急认证环境变量示例
- `docs/environment-contract.md`
  - 增加应急认证说明与 JWT `tenant_id` 安全约束
- `ai-hms-backend/README.md`
  - 移除硬编码 `admin/admin123` 登录示例
  - 说明正常登录应使用真实数据库账号

## 验证记录

### LSP Diagnostics

- `ai-hms-backend/internal/services/auth_service.go`：No diagnostics found
- `ai-hms-backend/internal/services/auth_service_test.go`：No diagnostics found
- `ai-hms-backend/internal/middleware/auth_test.go`：No diagnostics found

### 测试

命令：`go test ./...`

结果：通过

关键输出：

```text
ok   github.com/elliotxin/ai-hms-backend/internal/middleware
ok   github.com/elliotxin/ai-hms-backend/internal/services
ok   github.com/elliotxin/ai-hms-backend/internal/utils
```

### 构建

命令：`go build ./cmd/server`

结果：通过

## 结果对应预期

- 生产默认不启用紧急认证入口
- 普通数据库用户登录流程未改动
- 应急入口已显式文档化
- 缺失 `tenant_id` 的 JWT 会被拒绝
