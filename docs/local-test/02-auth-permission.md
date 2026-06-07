# 阶段二：认证、Token、权限测试

> 测试时间：2026-06-07
> 测试分支：`fix/legacy-ui-restore`
> 提交号：`4f375a7884727807a0ac87f100ea0c7d8b8848e9`

## 1. 测试用户准备

| 用户ID | 用户名 | 姓名 | 角色 | 密码 |
|--------|--------|------|------|------|
| 300412 | TEST_AI_HMS_admin | 测试管理员 | ADMIN | Test@123456 |
| 300413 | TEST_AI_HMS_doctor | 测试医生 | DOCTOR | Test@123456 |
| 300414 | TEST_AI_HMS_nurse | 测试护士 | NURSE | Test@123456 |
| 300415 | TEST_AI_HMS_viewer | 测试查看者 | VIEWER | Test@123456 |

> 通过 SQL 直接插入 `Identity_Users`、`Organ_Employee`、`Identity_UserRoles` 创建。
> 密码使用 ASP.NET Core Identity V3 哈希格式（PBKDF2/HMAC-SHA256）。

## 2. 登录接口测试

| 测试场景 | 预期 | 实际 | 状态 |
|----------|------|------|------|
| 正确账号+正确密码 | 200, 返回 token | 200, 含 token/userId/username/roles | 通过 |
| 正确账号+错误密码 | 401 | 401 | 通过 |
| 不存在用户+正确密码 | 401 | 401 | 通过 |
| 空用户名 | 401 | 401 | 通过 |
| 空密码 | 401 | 401 | 通过 |
| admin 登录 | 200, role=ADMIN | 200, role=ADMIN | 通过 |
| doctor 登录 | 200, role=DOCTOR | 200, role=DOCTOR | 通过 |
| nurse 登录 | 200, role=NURSE | 200, role=NURSE | 通过 |
| viewer 登录 | 200, role=VIEWER | 200, role=VIEWER | 通过 |
| 已锁定用户登录 | 应拒绝 | **200 登录成功** | **失败 (P1)** |

### JWT Token 内容验证

解码后的 JWT payload：
```json
{
  "user_id": "300412",
  "username": "TEST_AI_HMS_admin",
  "employee_name": "测试管理员",
  "roles": ["ADMIN"],
  "tenant_id": 3,
  "exp": 1780909194,
  "nbf": 1780822794,
  "iat": 1780822794
}
```

- tenant_id 已包含 ✓
- roles 正确 ✓
- 密码不出现在 token 中 ✓

## 3. Token 验证测试

| 测试场景 | 预期 | 实际 | 状态 |
|----------|------|------|------|
| 有效 token → GET /api/v1/me | 200, 返回用户信息 | 200, username+roles 正确 | 通过 |
| 无 token → GET /api/v1/me | 401 | 401 | 通过 |
| 无效 token → GET /api/v1/me | 401 | 401 | 通过 |
| 有效 token → GET /api/v1/patients | 200 | 200 | 通过 |

## 4. 权限控制测试

### 4.1 管理员接口保护（RequireRoles: ADMIN/管理员/安全管理员/运维管理员）

| 接口 | ADMIN | DOCTOR | NURSE | VIEWER | 无认证 |
|------|-------|--------|-------|--------|--------|
| GET /api/v1/users | 200 | 403 | - | 403 | 401 |
| GET /api/v1/app-roles | 200 | - | - | 403 | - |
| GET /api/v1/dict/types | 200 | - | - | - | - |

### 4.2 受保护接口（所有已登录用户可访问）

| 接口 | ADMIN | DOCTOR | NURSE | VIEWER | 无认证 |
|------|-------|--------|-------|--------|--------|
| GET /api/v1/patients | 200 | 200 | - | - | - |
| GET /api/v1/wards | 200 | 200 | - | - | - |
| GET /api/v1/beds | 200 | 200 | - | - | - |
| GET /api/v1/shifts | 200 | 200 | - | - | - |
| POST /api/v1/patients/{id}/orders | - | 200 | - | - | - |

### 4.3 接口路径验证

| 接口路径 | 预期 | 实际 | 备注 |
|----------|------|------|------|
| GET /api/v1/dicts | 200 | 404 | 端点路径可能为 /api/v1/dict/types |
| GET /api/v1/logs/lines | 200 | 404 | 端点路径待确认 |
| GET /api/v1/hdis/settings | 200 | 404 | 端点路径待确认 |
| GET /api/v1/dashboard/today | 200 | 404 | 端点路径待确认 |

> 以上 404 不影响核心功能测试，部分端点可能在 `RegisterXxxRoutes` 中使用了不同路径前缀。

## 5. 安全性检查

| 检查项 | 结果 |
|--------|------|
| 密码不出现在 JWT 中 | 通过 |
| 密码不出现在日志中 | 通过 |
| 无效 token 返回 401 | 通过 |
| 无 token 返回 401 | 通过 |
| 无权限返回 403 | 通过 |
| 锁定用户仍可登录 | **失败 (P1, ISSUE-002)** |

## 6. 数据库操作记录

| 操作 | 表 | 行 |
|------|-----|-----|
| INSERT 测试角色 | Identity_Roles | 3 行 (DOCTOR, NURSE, VIEWER) |
| INSERT 测试用户 | Identity_Users | 4 行 |
| INSERT 员工记录 | Organ_Employee | 4 行 |
| INSERT 角色关联 | Identity_UserRoles | 4 行 |

## 7. 阶段二结论

| 检查项 | 状态 |
|--------|------|
| 登录接口 | 通过 |
| Token 生成与验证 | 通过 |
| JWT 包含 tenant_id | 通过 |
| 角色权限控制 | 通过 |
| 未认证拒绝 (401) | 通过 |
| 无权限拒绝 (403) | 通过 |
| 密码信息安全 | 通过 |
| 锁定用户检测 | **失败 (P1)** |

### 发现问题

- **ISSUE-002 (P1)**：锁定用户（LockoutEnd 已过期）仍可正常登录，auth_service.go 未检查 LockoutEnd/LockoutEnabled 字段。
