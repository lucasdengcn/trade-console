# Test

## Instruction

```bash
# 进入交易管理器目录
cd trade-manager

# 运行集成测试
go test -v -run TestIntegration
```

## Manual Test

```bash
# 健康检查
curl http://localhost:8080/health

# 查询持仓
curl -X POST http://localhost:8080/api/v1/positions \
  -H "Content-Type: application/json" \
  -d '{"XyBusinAccount":"88060819"}'

# 下单
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"XyBusinAccount":"88060819","stockCode":"000001","Amount":100}'
```
