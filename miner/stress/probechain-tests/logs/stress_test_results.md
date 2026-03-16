# ProbeChain Rydberg 压力测试结果

执行时间: 2026-03-15
服务器: validator-1 (4C8G, ecs.c6.xlarge, upgraded from 2C4G)
节点: node1 (single node stress test)
链状态: Block ~83000, 9 peers

## 纯 TPS 测试 (16,000 笔交易)

| 轮次 | 交易数 | 提交速率 | 耗时 | 说明 |
|------|--------|---------|------|------|
| R1 | 1,000 | 913 tx/s | 1.1s | 短时爆发 |
| R2 | 5,000 | 647 tx/s | 7.7s | 中等负载,50个块 |
| R3 | 10,000 | 332 tx/s | 30.1s | 持续高负载,83个块 |
| **总计** | **16,000** | — | — | 全部提交成功 |

## 分析
- 峰值提交 TPS: 913 tx/s (1000笔burst)
- 持续提交 TPS: 332 tx/s (10000笔sustained)
- R3结束时 txpool pending=9325, 说明链处理速度 < 提交速度
- 链吞吐约 (10000-9325)/30 ≈ 22 tx/block（受出块间隔限制）
- 出块间隔 ~1s (追赶模式), 正常 15s 时每块可容纳更多交易

## 服务器升级记录
- 原始: ecs.t6-c1m2.large (2C4G) — personal.newAccount 太慢
- 升级: ecs.c6.xlarge (4C8G) — 解决了性能瓶颈
