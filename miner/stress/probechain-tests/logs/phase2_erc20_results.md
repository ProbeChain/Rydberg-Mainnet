# Phase 2: ERC20 代币测试结果 (20/20 通过)

执行时间: 2026-03-15
Token合约: 0x3a6575bdd09cd185868c55b47ad143b1b1cf66c1
编译: solc 0.8.30 --evm-version london (ProbeChain不支持PUSH0/Shanghai)

| # | 测试 | 结果 | 链上证据 |
|---|------|------|---------|
| 21 | 合约代码存在 | ✓ | codeLen:2700 |
| 22 | totalSupply | ✓ | 1e+24 (1M tokens) |
| 23 | deployer余额 | ✓ | 1e+24 |
| 24 | transfer(1000) | ✓ | tx:0x8de01fd2... 2/2角度 |
| 25 | approve | ✓ | allowance:5e+21 |
| 26 | transferFrom | ✓ | A2: 0→1e+20 |
| 27 | mint | ✓ | 1e+20→1.01e+22 |
| 28 | 非owner mint revert | ✓ | status:0x0 |
| 29 | transfer超额 revert | ✓ | revert |
| 30 | 批量transfer(10笔) | ✓ | 10/10 confirmed |
| 31 | totalSupply一致性 | ✓ | 两次查询相等 |
| 32 | balanceOf批量查询 | ✓ | 20/20 queries |
| 33 | minter查询 | ✓ | = deployer |
| 34 | gas消耗>21000 | ✓ | 32625 |
| 35 | approve+allowance | ✓ | 精确匹配 |
| 36 | approve归零重设 | ✓ | 0→12345 |
| 37 | transferFrom超额 | ✓ | revert |
| 38 | 并发5笔 | ✓ | 5/5 |
| 39 | gas估算 | ✓ | 32637 |
| 40 | supply增长 | ✓ | mint后增加 |

Pass Rate: 20/20 = 100%
