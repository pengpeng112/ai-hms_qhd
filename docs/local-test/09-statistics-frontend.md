# 阶段九：统计、看板、前端 E2E 测试

> 测试时间：2026-06-07

## 1. 统计看板

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 仪表盘统计 | GET /api/v1/dashboard/stats | 200 | 500 | **失败（P2）** |
| 质量统计 | GET /api/v1/statistics/quality | 200 | 200 | 通过 |
| 感染统计 | GET /api/v1/statistics/infection | 200 | 200 | 通过 |
| 工作负荷统计 | GET /api/v1/statistics/workload | 200 | 200 | 通过 |
| 实时监控 | GET /api/v1/monitoring/live-data | 200 | 200 | 通过 |
| 临床任务 | GET /api/v1/clinical-tasks | 200 | 200 | 通过 |

### Dashboard Stats 错误

```
"failed to encode args[2]: unable to encode 60 into text format for text (OID 25): cannot find encode plan"
```

初步原因：Dashboard 查询中参数类型不匹配（uint/int 传入 text 列）。

---

# 阶段十：安全、并发、性能测试

## 1. 安全测试

| 测试场景 | 输入 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| SQL 注入 | `?name=' OR 1=1 --` | 400 或不返回敏感数据 | 安全处理 | 通过 |
| XSS | `?name=<script>alert(1)</script>` | 安全处理 | 安全处理 | 通过 |
| 路径遍历 | `/patients/../../../etc/passwd` | 404/400 | 404 | 通过 |
| 无 Token | 无 Authorization | 401 | 401（阶段二已验证） | 通过 |
| 无权限 | doctor→admin 接口 | 403 | 403（阶段二已验证） | 通过 |
| 密码不泄露 | 日志检查 | 无密码 | 通过（阶段二已验证） | 通过 |

## 2. 并发测试

因工具限制未执行自动化并发测试（10 并发新增医嘱、5 并发停用同一医嘱等）。建议后续使用 Go 测试套件或 JMeter 完成。

## 3. 性能测试

未执行数据量级性能测试。当前测试库患者 365、医嘱 2347、排班 27094，API 响应均在秒级内。

### 已知性能风险

- `DATE("TreatmentTime")` 影响索引使用（test plan 建议）
- 前端可能存在一次性全量加载
- 部分慢查询需后续优化
