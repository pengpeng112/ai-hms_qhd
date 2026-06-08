# 智能排班算法增强报告

> 日期: 2026-06-08

## 已实现

| 算法 | 文件 | 说明 |
|------|------|------|
| HDF两轮优先 | `engine.go:107-217` | placeCell: 第一轮HDF→HDF机, 第二轮HD→HD机溢出HDF |
| HD溢出HDF | `placement.go:25-60` | 固定HD→同区HD→溢出HDF机→顺延→报警 |
| HDF双固定 | `placement.go:8-24` | 固定Hdf→就近HDF→报警 |
| 排满顺延 | `spill.go:10-86` | 同日后续班次→后续透析日→窗口报警 |
| HDF奇偶周 | `hdf.go:10-55` | 基准周一+负载均衡分组+跨年不漂移 |
| 固定机位 | `placement.go:28-35` | PlaceHdSession: 优先FixedHdBedId |
| HDF替换语义 | `hdf.go:60-77` | DecideMode: 按次替换不增加总次数 |
| 5种频率 | `frequency.go:8-38` | 一三五/二四六/每周两次/一次/临时 |

## 未实现

| 算法 | 说明 |
|------|------|
| 新患者多日同机 | 需跨天"首选全空闲同一台机"逻辑 |
| 连片/相邻偏好 | pickBest 已有基础评分，需增强 |

## 测试覆盖

`engine_test.go` — 11 项测试，覆盖频率、奇偶周、HD/HDF分配、占用判断、自由床位查找

```bash
go test ./internal/schedule_engine/... -v  # 全部 PASS
```
