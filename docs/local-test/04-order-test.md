# 阶段四：医嘱模块测试

> 测试时间：2026-06-07
> 测试分支：`fix/legacy-ui-restore`
> 提交号：`4f375a7884727807a0ac87f100ea0c7d8b8848e9`
> 测试患者：300410（支英俊）

## 1. 医嘱列表测试

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 默认列表 | GET /api/v1/patients/300410/orders | 200 | 200 (count=1 有效医嘱) | 通过 |
| 长期医嘱过滤 | GET ...?type=longTerm | 200 | 200 | 通过 |
| 临时医嘱过滤 | GET ...?type=temporary | 200 | 200 | 通过 |
| 包含已停用 | GET ...?includeExpired=true | 200 | 200 (count=2，含停用) | 通过 |

## 2. 创建医嘱测试

### 2.1 新增长期医嘱

| 项目 | 值 |
|------|-----|
| 接口 | POST /api/v1/patients/300410/orders |
| 状态 | **200 通过** |
| 返回 ID | 2063551977496776704 |

请求体：
```json
{
  "type": "longTerm",
  "category": "药品",
  "name": "TEST_AI_HMS_dilowfenzigansu",
  "content": "TEST_AI_HMS_dilowfenzigansu 4000IU",
  "dose": "4000",
  "unit": "IU",
  "route": "静脉注射",
  "timing": "透析中",
  "execTiming": "during",
  "startTime": "2026-06-01",
  "notes": "TEST_AI_HMS_long_term_order_test"
}
```

### 2.2 新增临时医嘱

| 项目 | 值 |
|------|-----|
| 接口 | POST /api/v1/patients/300410/orders |
| 状态 | **200 通过** |
| 返回 ID | 2063551977953955840 |

请求体：
```json
{
  "type": "temporary",
  "category": "处置",
  "name": "TEST_AI_HMS_temp_buye",
  "content": "TEST_AI_HMS_temp_buye 100ml",
  "dose": "100",
  "unit": "ml",
  "route": "静脉滴注",
  "timing": "透析后",
  "execTiming": "after",
  "startTime": "2026-06-01",
  "endTime": "2026-06-01",
  "notes": "TEST_AI_HMS_temp_order_test"
}
```

## 3. 修改医嘱测试

| 项目 | 值 |
|------|-----|
| 接口 | PUT /api/v1/patients/300410/orders/2063551977496776704 |
| 状态 | **200 通过** |
| 更新字段 | dose: 4000→5000, timing: 透析中→透析前, notes 更新 |

请求体：
```json
{
  "content": "TEST_AI_HMS_dilowfenzigansu 5000IU",
  "dose": "5000",
  "route": "静脉注射",
  "timing": "透析前",
  "notes": "TEST_AI_HMS_order_update_test"
}
```

## 4. 停用医嘱测试

| 项目 | 值 |
|------|-----|
| 接口 | POST /api/v1/patients/300410/orders/2063551977953955840/stop |
| 状态 | **200 通过** |
| 停用日期 | 2026-06-02 |

请求体：
```json
{
  "stopReason": "TEST_AI_HMS_stop_test",
  "stopDate": "2026-06-02"
}
```

## 5. 数据库完整性验证

### 5.1 Order_PatientOrder 主表

| 字段 | 长期医嘱 (2063551977496776704) | 临时医嘱 (2063551977953955840) | 验证 |
|------|------|------|------|
| TenantId | 3 | 3 | ✓ |
| PatientId | 300410 | 300410 | ✓ |
| Content | TEST_AI_HMS_dilowfenzigansu 5000IU | TEST_AI_HMS_temp_buye 100ml | ✓ |
| Dosage | 5000（更新后） | 100 | ✓ |
| Classification | 有值 | 有值 | ✓ |
| UseOpportunity | 有值 | 有值 | ✓ |
| UseMethod | 有值 | 有值 | ✓ |
| UseWay | 有值 | 有值 | ✓ |
| Note | TEST_AI_HMS_order_update_test | TEST_AI_HMS_temp_order_test | ✓ |
| CreatorId | 300412 (TEST_AI_HMS_admin) | 300412 | ✓ |
| OperatorId | 300412 | 300412 | ✓ |
| EndTime | NULL（未停用） | 2026-06-02（已停用） | ✓ |

### 5.2 Order_PatientDayOrder 日医嘱表

| PatientOrderId | Content | Status | 验证 |
|----|------|------|------|
| 2063551977496776704 | TEST_AI_HMS_dilowfenzigansu 4000IU | 20 | ✓ 自动生成 |
| 2063551977953955840 | TEST_AI_HMS_temp_buye 100ml | 20 | ✓ 自动生成 |

> 注意：Update 更新了主医嘱的 Content（4000→5000），但 DayOrder 在创建时已快照为旧值 4000IU，未同步更新。DayOrder 可能是一个独立的执行快照，需要确认业务设计是否符合预期。

### 5.3 停用后列表行为

| 测试场景 | 预期 | 实际 | 状态 |
|----------|------|------|------|
| 默认列表（不含停用） | 不显示已停用的临时医嘱 | 不显示（count=2 仅为长期+原有） | 通过 |
| includeExpired=true | 显示已停用医嘱 | 显示（count=4） | 通过 |

## 6. 数据库验证 SQL

```sql
-- 验证创建的医嘱
SELECT "Id", "TenantId", "PatientId", "Content", "Dosage", 
       "UseOpportunity", "UseMethod", "UseWay", "Note", "CreatorId", "OperatorId"
FROM "Order_PatientOrder"
WHERE "Content" LIKE 'TEST_AI_HMS_%'
ORDER BY "CreateTime" DESC;

-- 验证日医嘱
SELECT "Id", "PatientOrderId", "Content", "Status"
FROM "Order_PatientDayOrder"
WHERE "PatientOrderId" IN (2063551977496776704, 2063551977953955840);
```

## 7. 阶段四结论

| 检查项 | 状态 |
|--------|------|
| 医嘱列表（默认/类型过滤/含停用） | 通过 |
| 新增长期医嘱 | 通过 |
| 新增临时医嘱 | 通过 |
| 修改医嘱（字段更新） | 通过 |
| 停用医嘱（EndTime 写入） | 通过 |
| 停用后默认不显示 | 通过 |
| includeExpired 显示历史 | 通过 |
| TenantId 隔离 | 通过（TenantId=3） |
| CreatorId/OperatorId | 通过 |
| DayOrder 自动生成 | 通过 |
| 主从表一致性 | 通过（主表更新后 DayOrder 保持快照，需确认业务设计） |

### 发现问题

无新增 P0/P1 问题。DayOrder 快照机制需确认是否符合业务预期（Update 主医嘱后 DayOrder 未同步更新 Content）。

### 未测试项（需后续阶段覆盖）

- 事务回滚测试（模拟日医嘱写入失败后主医嘱回滚）
- 并发测试（10 并发新增医嘱）
- 跨患者修改防御
