# PIP-8004: Trustless Agent Consensus Layer

## ProbeChain Improvement Proposal — ERC-8004 融入 PoB 共识层设计方案

---

## 1. 核心理念

ERC-8004 在以太坊上是**应用层合约**（3 个独立 Registry 合约）。
ProbeChain 的做法不同——将 Agent 身份、信誉、验证**下沉到共识层**，由 PoB 引擎原生支持。

```
以太坊 ERC-8004:   Application Layer (Solidity Contracts)
                          ↓ 调用
                   EVM Execution Layer
                          ↓
                   Consensus Layer (不感知 Agent)

ProbeChain PIP-8004:  PoB Consensus Layer (原生感知 Agent)
                          ↓ 共识级别保证
                   Agent Identity + Reputation + Validation
                          ↓ 融合
                   BehaviorScore (5+N 维度)
```

**优势**：
- Agent 信誉由共识层保证，不可被合约层绕过
- 验证者同时担任 Agent 验证者，无需额外信任假设
- 行为评分从「验证者行为」扩展到「Agent 行为」，统一评分体系
- 共识级事件（出块、ACK）天然成为 Agent 活跃度证明

---

## 2. 架构总览

### 2.1 三层实体模型

```
┌─────────────────────────────────────────────┐
│              PoB Consensus Layer             │
│                                             │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐ │
│  │Validator │  │SmartLight│  │  Agent    │ │
│  │(出块者)   │  │(轻节点)   │  │(服务提供者)│ │
│  └────┬─────┘  └────┬─────┘  └─────┬─────┘ │
│       │             │              │        │
│       └─────────────┼──────────────┘        │
│                     ▼                       │
│           ┌─────────────────┐               │
│           │  BehaviorScore  │               │
│           │  统一评分引擎    │               │
│           └─────────────────┘               │
└─────────────────────────────────────────────┘
```

| 实体 | 角色 | 评分维度 | 奖励来源 |
|------|------|---------|---------|
| Validator | 出块、验证、ACK | 5 维 (现有) | 1 PROBE/block |
| SmartLight | 心跳、轻验证、GNSS | 5 维 (现有) | 0.2 PROBE/block pool |
| **Agent** | **提供服务、响应请求、完成任务** | **6 维 (新增)** | **Agent Reward Pool** |

### 2.2 数据流

```
Agent 注册
    │
    ▼
Snapshot.Agents[address] = AgentScore{...}
    │
    ▼
Agent 提供服务 → 用户/Agent 提交 Feedback TX
    │
    ▼
PoB Engine 在 Finalize() 中处理 Feedback
    │
    ▼
BehaviorAgent.EvaluateAgent() 更新 AgentScore
    │
    ▼
Validator 可发起 ValidationRequest
    │
    ▼
验证结果写入 Snapshot.AgentValidations
    │
    ▼
Agent 奖励按 AgentScore 比例分配
```

---

## 3. 数据结构设计

### 3.1 Agent 身份 (Identity Registry → Consensus State)

```go
// consensus/pob/snapshot.go — 新增

// AgentIdentity represents a registered agent in the PoB consensus.
// Unlike ERC-8004 which uses ERC-721, ProbeChain stores agent identity
// directly in the consensus snapshot — no separate NFT contract needed.
type AgentIdentity struct {
    AgentID     uint64         // Auto-incrementing, globally unique
    Owner       common.Address // Agent owner address (who registered it)
    Wallet      common.Address // Agent's operational wallet (for receiving payments)
    URI         string         // Registration file URI (IPFS/HTTPS)
    URIHash     common.Hash    // Keccak256 of registration file content
    RegisterdAt uint64         // Block number when registered
    Active      bool           // Whether agent is currently active
}
```

### 3.2 Agent 行为评分 (Reputation Registry → BehaviorScore 扩展)

```go
// consensus/pob/behavior.go — 新增

// AgentScore extends BehaviorScore with agent-specific dimensions.
type AgentScore struct {
    Total             uint64 // Composite score (0-10000)
    Responsiveness    uint64 // Dimension 1: 响应速度与可用性
    Accuracy          uint64 // Dimension 2: 服务结果正确性
    Reliability       uint64 // Dimension 3: 长期稳定性
    Cooperation       uint64 // Dimension 4: 与其他 Agent 的协作度
    Economy           uint64 // Dimension 5: 经济行为合理性 (不欺诈)
    Sovereignty       uint64 // Dimension 6: 独立性 (非 Sybil)
    LastUpdate        uint64 // Block number of last update
    FeedbackCount     uint64 // Total feedback received
    ValidationCount   uint64 // Total validations received
}

// AgentHistory tracks agent on-chain actions.
type AgentHistory struct {
    RequestsServed    uint64 // Total service requests handled
    RequestsFailed    uint64 // Failed/timeout requests
    PositiveFeedback  uint64 // Feedback with value > 0
    NegativeFeedback  uint64 // Feedback with value <= 0
    ValidationsPass   uint64 // Validation responses >= 50
    ValidationsFail   uint64 // Validation responses < 50
    PaymentsReceived  uint64 // x402 payments received (count)
    PaymentsDisputed  uint64 // Disputed payments
    SlashCount        uint64 // Times slashed
    TasksCompleted    uint64 // Agent tasks completed for SmartLight
    CollabCount       uint64 // Agent-to-Agent collaborations
}
```

### 3.3 Agent 验证 (Validation Registry → Validator 职责扩展)

```go
// consensus/pob/snapshot.go — 新增

// AgentValidation represents a validator's assessment of an agent.
type AgentValidation struct {
    Validator    common.Address // PoB validator who performed validation
    AgentID      uint64         // Target agent
    Response     uint8          // 0-100 score
    Tag          string         // Category tag
    ResponseHash common.Hash    // Hash of detailed response (off-chain)
    BlockNumber  uint64         // When validated
}
```

### 3.4 Snapshot 扩展

```go
// consensus/pob/snapshot.go — 修改 Snapshot 结构

type Snapshot struct {
    // ... 现有字段保持不变 ...
    Validators  map[common.Address]*BehaviorScore
    Histories   map[common.Address]*ValidatorHistory
    SmartLights map[common.Address]*SmartLightScore
    SLHistories map[common.Address]*SmartLightHistory

    // === PIP-8004 新增 ===
    NextAgentID      uint64                              // Auto-increment counter
    Agents           map[uint64]*AgentIdentity           // AgentID → Identity
    AgentOwners      map[common.Address][]uint64         // Owner → AgentIDs
    AgentScores      map[uint64]*AgentScore              // AgentID → Score
    AgentHistories   map[uint64]*AgentHistory            // AgentID → History
    AgentValidations map[common.Hash]*AgentValidation    // RequestHash → Validation
}
```

---

## 4. 共识层交互设计

### 4.1 Agent 生命周期 (通过特殊交易类型)

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Register    │────→│   Active     │────→│  Demoted     │
│  (质押 10    │     │  (提供服务)   │     │  (评分过低)   │
│   PROBE)     │     │              │     │              │
└──────────────┘     └──────┬───────┘     └──────┬───────┘
                            │                     │
                            ▼                     ▼
                     ┌──────────────┐     ┌──────────────┐
                     │  Validated   │     │  Deregistered│
                     │  (验证者认证) │     │  (退出/清退)  │
                     └──────────────┘     └──────────────┘
```

### 4.2 新交易类型

```go
// core/types/transaction.go — 新增交易类型常量

const (
    // Agent transactions (PIP-8004)
    TxTypeAgentRegister     = 0x70  // 注册 Agent
    TxTypeAgentUpdate       = 0x71  // 更新 Agent URI/Wallet
    TxTypeAgentDeregister   = 0x72  // 注销 Agent
    TxTypeAgentFeedback     = 0x73  // 提交反馈
    TxTypeAgentValidReq     = 0x74  // 请求验证
    TxTypeAgentValidResp    = 0x75  // 验证响应 (仅 Validator 可发)
)
```

### 4.3 Finalize 中的 Agent 状态更新

```go
// consensus/pob/pob.go — PobFinalize() 扩展

func (p *ProofOfBehavior) PobFinalize(chain, header, state, txs) {
    snap := p.snapshot(...)

    // 1. 现有逻辑: 验证者奖励 + SmartLight 奖励
    accumulateRewards(...)
    accumulateSmartLightRewards(...)

    // 2. 新增: 处理 Agent 交易
    for _, tx := range txs {
        switch tx.Type() {
        case TxTypeAgentRegister:
            snap.RegisterAgent(tx.From(), tx.AgentURI(), tx.AgentWallet())
            // 扣除 10 PROBE 质押
        case TxTypeAgentFeedback:
            snap.ProcessFeedback(tx.AgentID(), tx.From(), tx.Value(), tx.Tags())
            // 更新 AgentScore
        case TxTypeAgentValidResp:
            // 仅允许 Validator 地址发送
            if snap.isValidator(tx.From()) {
                snap.ProcessValidation(tx.RequestHash(), tx.Response(), tx.Tag())
            }
        }
    }

    // 3. 新增: Agent 奖励分配
    accumulateAgentRewards(snap, state, header)

    // 4. 新增: Agent 降级检查
    for agentID, score := range snap.AgentScores {
        if score.Total < p.pobConfig.DemotionThreshold {
            snap.DemoteAgent(agentID)
        }
    }
}
```

---

## 5. Agent 评分引擎

### 5.1 六维评分模型

```go
// consensus/pob/behavior.go — 新增

// AgentBehaviorWeights defines scoring weights for agents.
var AgentBehaviorWeights = [6]uint64{
    20, // Responsiveness (响应性)
    25, // Accuracy (准确性)
    15, // Reliability (可靠性)
    15, // Cooperation (协作性)
    15, // Economy (经济性)
    10, // Sovereignty (独立性)
}

func (ba *BehaviorAgent) EvaluateAgent(history *AgentHistory) *AgentScore {
    score := &AgentScore{}

    // D1: Responsiveness — 处理请求的成功率
    totalReqs := history.RequestsServed + history.RequestsFailed
    if totalReqs > 0 {
        score.Responsiveness = (history.RequestsServed * 10000) / totalReqs
    }

    // D2: Accuracy — 基于反馈和验证
    totalFeedback := history.PositiveFeedback + history.NegativeFeedback
    if totalFeedback > 0 {
        score.Accuracy = (history.PositiveFeedback * 10000) / totalFeedback
    }

    // D3: Reliability — 长期稳定性 (基于无 slash 的连续服务)
    score.Reliability = 10000 - min(history.SlashCount * 1000, 10000)

    // D4: Cooperation — Agent 间协作
    totalTasks := history.TasksCompleted + history.CollabCount
    if totalTasks > 0 {
        score.Cooperation = min(totalTasks * 100, 10000)
    }

    // D5: Economy — 经济行为合理性
    totalPayments := history.PaymentsReceived + history.PaymentsDisputed
    if totalPayments > 0 {
        score.Economy = ((totalPayments - history.PaymentsDisputed) * 10000) / totalPayments
    }

    // D6: Sovereignty — 反 Sybil (由 Validator 验证结果决定)
    totalValid := history.ValidationsPass + history.ValidationsFail
    if totalValid > 0 {
        score.Sovereignty = (history.ValidationsPass * 10000) / totalValid
    }

    // Weighted total
    score.Total = (score.Responsiveness * AgentBehaviorWeights[0] +
        score.Accuracy * AgentBehaviorWeights[1] +
        score.Reliability * AgentBehaviorWeights[2] +
        score.Cooperation * AgentBehaviorWeights[3] +
        score.Economy * AgentBehaviorWeights[4] +
        score.Sovereignty * AgentBehaviorWeights[5]) / 100

    return score
}
```

### 5.2 评分体系对比

```
             PoB Validator        SmartLight           Agent (PIP-8004)
            ─────────────        ──────────           ────────────────
Dim 1:      Liveness (25%)       Liveness (30%)       Responsiveness (20%)
Dim 2:      Correctness (25%)    Correctness (20%)    Accuracy (25%)
Dim 3:      Cooperation (18%)    Cooperation (25%)    Reliability (15%)
Dim 4:      Consistency (17%)    Consistency (10%)    Cooperation (15%)
Dim 5:      Signal (15%)         Signal/GNSS (15%)    Economy (15%)
Dim 6:      —                    —                    Sovereignty (10%)
            ─────────────        ──────────           ────────────────
Total:      0-10000              0-10000              0-10000
Slash at:   <1000                <1000                <1000
```

---

## 6. 奖励经济模型

### 6.1 每区块奖励分配

```
每个 Block (400ms) 的总奖励:
┌─────────────────────────────────────────────┐
│ Validator Reward:     1.0  PROBE            │
│ PoW Miner Reward:     3.0  PROBE            │
│ SmartLight Pool:      0.2  PROBE            │
│ Agent Pool (新增):    0.3  PROBE            │ ← PIP-8004 新增
│ Uncle Reward:         0.5  PROBE (if uncle) │
├─────────────────────────────────────────────┤
│ 合计 (无 uncle):      4.5  PROBE/block      │
│ 每天 (~216,000 块):   ~972,000 PROBE/day    │
│ Agent Pool 每天:      ~64,800 PROBE/day     │
└─────────────────────────────────────────────┘
```

### 6.2 Agent 奖励分配算法

```go
func accumulateAgentRewards(snap *Snapshot, state *state.StateDB, header *types.Header) {
    agentPoolReward := big.NewInt(3e17) // 0.3 PROBE per block

    // Calculate total weighted score
    var totalScore uint64
    for _, score := range snap.AgentScores {
        if score.Total >= snap.config.DemotionThreshold {
            totalScore += score.Total
        }
    }
    if totalScore == 0 {
        return // No active agents, pool burns or carries over
    }

    // Proportional distribution by score
    for agentID, score := range snap.AgentScores {
        if score.Total < snap.config.DemotionThreshold {
            continue
        }
        identity := snap.Agents[agentID]
        share := new(big.Int).Mul(agentPoolReward, big.NewInt(int64(score.Total)))
        share.Div(share, big.NewInt(int64(totalScore)))
        state.AddBalance(identity.Wallet, share)
    }
}
```

---

## 7. RPC API 扩展

### 7.1 新增 API (namespace: "pob")

```go
// consensus/pob/api.go — 新增

// Agent Identity
GetAgent(agentID uint64) (*AgentIdentity, error)
GetAgentsByOwner(owner common.Address) ([]uint64, error)
GetAgentCount() (uint64, error)

// Agent Scores
GetAgentScore(agentID uint64) (*AgentScore, error)
GetAgentScores() (map[uint64]*AgentScore, error)
GetAgentHistory(agentID uint64) (*AgentHistory, error)

// Agent Validation
GetAgentValidations(agentID uint64) ([]*AgentValidation, error)

// Agent Discovery (链上搜索)
SearchAgents(tag string, minScore uint64) ([]uint64, error)
```

### 7.2 与 ERC-8004 的兼容层

为了让符合 ERC-8004 标准的外部工具也能与 ProbeChain 交互，
部署一个薄合约层将共识层数据映射为 ERC-8004 接口：

```solidity
// contracts/PIP8004Bridge.sol
// 只读桥接合约 — 从共识层 precompile 读取 Agent 数据

contract PIP8004Bridge is IERC8004Identity, IERC8004Reputation {

    // 通过 precompiled contract 读取共识层状态
    address constant POB_PRECOMPILE = 0x0000000000000000000000000000000000000801;

    function register(string agentURI) external returns (uint256) {
        // 发送 TxTypeAgentRegister 交易
    }

    function giveFeedback(uint256 agentId, int128 value, ...) external {
        // 发送 TxTypeAgentFeedback 交易
    }

    // ERC-721 compatible — Agent 可被当作 NFT 交易
    function ownerOf(uint256 tokenId) external view returns (address) {
        return POB_PRECOMPILE.getAgentOwner(tokenId);
    }
}
```

---

## 8. 与 x402 的集成点

### 8.1 Agent 注册文件中的 x402 支持

```json
{
  "type": "https://probechain.org/pip-8004/registration-v1",
  "name": "WeatherOracle",
  "services": [
    {
      "name": "A2A",
      "endpoint": "https://weather.agent.example/a2a",
      "x402": {
        "enabled": true,
        "pricePerRequest": "100000000000000",
        "token": "PROBE",
        "chainId": 8004
      }
    }
  ],
  "supportedTrust": ["pob-behavior-score"]
}
```

### 8.2 x402 支付自动更新 AgentHistory

```
Agent 收到 x402 支付 → PaymentsReceived++
支付被争议          → PaymentsDisputed++, Economy 维度下降
无争议完成          → PositiveFeedback++, Accuracy 维度上升
```

---

## 9. 实现路线图

### Phase 1: 数据结构 (当前可做)
- [ ] `AgentIdentity`, `AgentScore`, `AgentHistory` 结构定义
- [ ] Snapshot 扩展，添加 Agent 相关 map
- [ ] `newSnapshot()` 初始化 Agent 字段
- [ ] Snapshot 的 RLP 编解码更新

### Phase 2: 交易处理
- [ ] 新交易类型 `0x70-0x75` 定义
- [ ] `snapshot.apply()` 处理 Agent 交易
- [ ] `PobFinalize()` 中 Agent 状态更新
- [ ] Agent 注册质押机制 (10 PROBE)

### Phase 3: 评分引擎
- [ ] `BehaviorAgent.EvaluateAgent()` 六维评分
- [ ] Agent 降级/Slash 逻辑
- [ ] `accumulateAgentRewards()` 奖励分配

### Phase 4: RPC + 工具
- [ ] Agent RPC API 实现
- [ ] PIP8004Bridge 合约
- [ ] CLI 命令: `gprobe agent register/query/feedback`

### Phase 5: 生态集成
- [ ] x402 支付网关
- [ ] ProbeSmartLight iOS 中的 Agent 市场 UI
- [ ] Agent 发现与排名页面

---

## 10. 与纯合约实现的对比

| 维度 | ERC-8004 (以太坊合约) | PIP-8004 (ProbeChain 共识层) |
|------|---------------------|----------------------------|
| 信誉可信度 | 合约级 (可被绕过) | 共识级 (不可绕过) |
| 评分更新 | 需要 gas | 共识引擎自动更新 |
| Sybil 抵抗 | 弱 (任何人可注册) | 强 (Validator 验证 + 质押) |
| 与出块耦合 | 无关 | Agent 活跃度直接影响评分 |
| 跨链兼容 | 原生 | 需要桥接合约 |
| Gas 开销 | 高 (每次反馈上链) | 低 (共识层批量处理) |
| 去中心化 | 完全去中心化 | 依赖 PoB 验证者集 |
| 灵活性 | 高 (合约可升级) | 需要硬分叉升级 |

---

## 11. 创世配置扩展

```json
{
    "config": {
        "pob": {
            "period": 0,
            "tickIntervalMs": 400,
            "epoch": 30000,
            "initialScore": 5000,
            "slashFraction": 1000,
            "demotionThreshold": 1000,
            "agentStakeRequired": "10000000000000000000",
            "agentRewardPerBlock": "300000000000000000",
            "agentScoreWeights": [20, 25, 15, 15, 15, 10],
            "list": [...]
        }
    }
}
```
