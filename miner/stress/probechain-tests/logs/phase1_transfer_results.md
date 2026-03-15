# Phase 1: 基础转账测试结果 (19/20 通过)

执行时间: 2026-03-15
链: ProbeChain Rydberg Testnet (Chain ID 8004, Block ~55000)
节点: validator-1 (validator-1)

| # | 测试 | 结果 | 链上证据 |
|---|------|------|---------|
| 1 | PROBE基础转账(4角度) | ✓ | tx:0x5d1f68e1... block:55628 gas:21000 sender:-1000.042 receiver:+1000 |
| 2 | 零值转账 | ✓ | tx:0xa2abb1ba... gas:21000 |
| 3 | 大额转账(100 PROBE) | ✓ | tx:0x3a051998... gas:21000 |
| 4 | 小额转账(0.01 PROBE) | ✓ | tx:0xd77cb69c... |
| 5 | 1 wei转账 | ✓ | tx:0x8004b4a1... |
| 6 | 自转账(3角度) | ✓ | tx:0xee5633d4... gasCost:42000000000000 wei |
| 7 | 余额不足转账 | ✓ | rejected: insufficient funds for gas * price + value |
| 8 | Nonce连续100笔 | ✓ | 100/100 sent |
| 9 | Nonce乱序 | ✓ | nonce 223→227 |
| 10 | Nonce重复 | ✓ | duplicate_rejected |
| 11 | EIP-1559 baseFee | ✓ | baseFee=1000000000 (1 Gwei) |
| 12 | Legacy gasPrice tx | ✓ | tx:0xfc82834d... |
| 13 | Gas精确21000 | ✓ | gasUsed:21000 |
| 14 | Gas不足(20999) | ✓ | rejected: intrinsic gas too low |
| 15 | Gas过高(8M) | ✓ | gasUsed:21000 (只收实际) |
| 16 | 批量4笔 | ✓ | 4/4 confirmed |
| 17 | 环形转账 | ✓ | 2/2 confirmed |
| 18 | 快速100笔 | ✓ | 100/100 sent |
| 19 | txpool状态 | ✓ | pending/queued 正常 |
| 20 | 出块间隔 | ✗ | 1.0s (链追赶模式,非正常状态) |

Pass Rate: 19/20 = 95%
