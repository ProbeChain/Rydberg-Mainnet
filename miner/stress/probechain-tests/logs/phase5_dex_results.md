# Phase 5: ProSwap DEX 测试结果 (29/30 通过)

执行时间: 2026-03-15
MiniSwap: pro1d3ugls6ld8ddce3lgqfc8weu6u98s2fg3jvnvd
TokenA: 0x3a6575bdd09cd185868c55b47ad143b1b1cf66c1 (SimpleToken)
TokenB: pro1jasausgpme4x30k7nl3uumtnyl72svncu2qxkj
测试方式: 两个ERC20代币组成交易对，AMM自动做市

| # | 测试 | 结果 | 说明 |
|---|------|------|------|
| 86 | TokenB部署 | ✓ | gas:373174 |
| 87 | TokenA余额 | ✓ | |
| 88 | TokenB余额 | ✓ | |
| 89 | MiniSwap合约 | ✓ | |
| 90 | 创建交易对 | ✓ | |
| 91 | 重复创建revert | ✓ | |
| 92 | 相同地址revert | ✓ | |
| 93 | 添加流动性 | ✓ | r0,r1>0 |
| 94 | LP余额 | ✓ | |
| 95 | 准备金查询 | ✓ | |
| 96 | A→B swap | ✓ | TokenB余额增加 |
| 97 | B→A swap | ✓ | TokenA余额增加 |
| 98 | 滑点保护 | ✓ | revert |
| 99 | getAmountOut | ✓ | |
| 100 | K值>0 | ✓ | |
| 101 | swap后准备金变化 | ✓ | |
| 102 | 移除流动性 | ✓ | LP减少 |
| 103 | 连续5次swap | ✓ | 5/5 |
| 104 | pairCount | ✓ | |
| 105 | getPairId | ✓ | |
| 106 | swap事件 | ✓ | logs>0 |
| 107 | TokenA supply | ✓ | |
| 108 | TokenB supply | ✓ | |
| 109 | TokenA transfer | ✓ | |
| 110 | 再次addLiquidity | ✗ | approve额度不足(测试问题,非链问题) |
| 111 | 价格影响 | ✓ | 大单rate < 小单rate |
| 112 | swap gas | ✓ | >21000 |
| 113 | addLiquidity gas | ✓ | >21000 |
| 114 | 批量10次swap | ✓ | 10/10 |
| 115 | 全流程验证 | ✓ | pair+lp+reserves都正常 |

Pass Rate: 29/30 = 96.7%
注: #110失败因approve额度用完,不是链或合约问题
