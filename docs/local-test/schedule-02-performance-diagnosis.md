# 排班性能诊断

> 诊断日期: 2026-06-08 | 环境: localhost

## API 耗时

| 接口 | 耗时 | 响应大小 | 问题 |
|------|------|---------|------|
| GET /health | 911ms | 134B | 冷启动 |
| GET /shifts | 20ms | 924B | 正常 |
| GET /schedule/week | 176ms | **412KB** | 待排患者685人全量返回 |
| GET /schedule/week+ward | 123ms | 411KB | 病区过滤几乎无瘦身效果 |
| GET /templates | 13ms | 615B | 正常 |
| GET /conflict-queue | 20ms | 9KB | 正常 |
| GET /board | 280ms | 10KB | 可用 |

## 根因分析

1. **待排患者全量返回**: 685人 × 每人多项字段 = 412KB。应限制返回数量(100)
2. **DATE(TreatmentTime)**: 部分查询仍用 `DATE()` 导致索引失效(已修复部分)
3. **前端渲染**: 685人队列一次性渲染全部卡片
4. **保存后整周刷新**: `loadWeek()` 重新拉取 412KB 数据
5. **缺少组合索引**: `Schedule_PatientShift` 无 `(TenantId, TreatmentTime, ShiftId)` 索引

## 优化建议优先级

| 优先级 | 优化项 | 预期收益 |
|--------|--------|---------|
| P0 | 待排患者限制返回数量(100) | 412KB → ~60KB |
| P0 | 保存后局部更新, 不调 loadWeek() | 消除全局刷新 |
| P1 | 补充组合索引 | SQL 耗时降低 50% |
| P1 | DATE()→范围查询(已做) | 索引可用 |
| P2 | 前端队列虚拟化 | 渲染性能提升 |
