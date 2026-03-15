# Phase 4: Key Trading 社交代币测试结果 (25/25 通过)

执行时间: 2026-03-15
合约: pro154e9gj0xnvrl3zyxkqyaae08pmuveaf8ffzag4 (KeyTrading)
编译: solcjs 0.6.12 --optimize
Gas: 719501

| # | 测试 | 结果 | 关键证据 |
|---|------|------|---------|
| 61 | 部署验证 | ✓ | code存在 |
| 62 | 首次self-buy | ✓ | supply=1, price曲线正确 |
| 63 | 非self首次购买revert | ✓ | status=0 (正确revert) |
| 64 | 购买1个share | ✓ | A1 balance=1 |
| 65 | 购买5个share | ✓ | supply增加 |
| 66 | 卖出1个share | ✓ | supply减少 |
| 67 | 卖出最后一个revert | ✓ | status=0 |
| 68 | 超额卖出revert | ✓ | status=0 |
| 69 | 价格曲线 | ✓ | price2 > price1 |
| 70 | 费用计算5% | ✓ | 250bps + 250bps |
| 71 | getBuyPrice | ✓ | 非负 |
| 72 | getSellPrice | ✓ | 非负 |
| 73 | getBuyPriceAfterFee | ✓ | > base price |
| 74 | getSellPriceAfterFee | ✓ | < base price |
| 75 | protocolFeeDestination | ✓ | 地址正确 |
| 76 | sharesSupply | ✓ | > 0 |
| 77 | sharesBalance | ✓ | >= 0 |
| 78 | 快速买卖 | ✓ | supply恢复 |
| 79 | 价格随supply上涨 | ✓ | 递增 |
| 80 | 多subject并行 | ✓ | CB和A2都有supply |
| 81 | 超额付款退款 | ✓ | 实际花费 < 付款 |
| 82 | 付款不足revert | ✓ | status=0 |
| 83 | Trade事件 | ✓ | logs.length > 0 |
| 84 | 修改费率 | ✓ | 250→300→250 |
| 85 | setFeeDestination | ✓ | 修改成功 |

Pass Rate: 25/25 = 100%
