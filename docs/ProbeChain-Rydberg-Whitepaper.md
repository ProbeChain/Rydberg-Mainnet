# ProbeChain Rydberg: A Gold-Anchored Proof-of-Behavior Blockchain

**Technical Whitepaper v2.1 — March 2026**

---

## Abstract

ProbeChain is a Layer-1 blockchain that replaces energy expenditure (PoW) and capital lockup (PoS) with **Proof-of-Behavior (PoB)** — a consensus mechanism where validators earn rewards proportional to the quality of their observable behavior. Token emission is driven by real transaction volume, decays as physical gold reserves accumulate, and halts entirely when reserves reach 36,000 metric tons. This paper specifies the consensus rules, reward formulas, decay model, behavioral scoring algorithms, and anti-sybil mechanisms of the Rydberg testnet.

---

## 1. Design Principles

Three axioms distinguish ProbeChain from existing blockchains:

1. **Rewards must track real economic activity.** Empty blocks should not inflate supply.
2. **Emission must have a physical anchor.** Neither arbitrary caps (21M BTC) nor algorithmic deflation provide fundamental backing. Gold does.
3. **Participation must be frictionless.** If joining requires specialized hardware or locked capital, the network excludes the agents and devices it intends to serve.

---

## 2. Consensus: Proof-of-Behavior

### 2.1 Overview

In PoB, the right to produce blocks and the magnitude of rewards are functions of a node's **behavioral score** — a composite metric derived from on-chain observations of responsiveness, accuracy, cooperation, and uptime.

Two node types participate:

- **Agent Nodes** — AI agents processing on-chain settlements
- **Physical Nodes** — Devices contributing storage, compute, and connectivity

### 2.2 Parameters

| Parameter | Symbol | Value |
|-----------|--------|-------|
| Chain ID | — | 8004 |
| Block period | T | 15 seconds |
| Epoch length | E | 30,000 blocks (~5.2 days) |
| Validator set size | — | 1 – 21 |
| Initial score | S_0 | 5,000 / 10,000 |
| Slash fraction | — | 1,000 / 10,000 (10%) |
| Demotion threshold | — | 1,000 / 10,000 (10%) |

### 2.3 Block Production

Within each epoch, the validator set is fixed. Validators take turns producing blocks via weighted round-robin selection based on behavioral scores. When all scores are equal (e.g., at genesis), selection is simple round-robin:

```
producer_index = block_number % epoch_length % validator_count
producer = validators[producer_index]
```

When scores diverge, selection uses a deterministic weighted random algorithm seeded by the parent hash and block number, ensuring fairness proportional to behavior.

### 2.4 Ack-Based Block Finalization

After each block, all validators broadcast **acknowledgments (acks)** indicating agreement or opposition:

```
Ack = {
  EpochPosition:  validator index in the current epoch
  Number:         block number being acknowledged
  BlockHash:      hash of the acknowledged block (agree) or empty (oppose)
  AckType:        0 = Agree, 1 = Oppose
  WitnessSig:     validator's cryptographic signature
}
```

Block production proceeds when the producer collects sufficient acks:

| Threshold | Formula | 9 validators | Effect |
|-----------|---------|-------------|--------|
| MostValidatorWitness | n*2/3 + 1 | 7 | Immediate commit |
| LeastValidatorWitness | n*1/3 + 1 | 4 | Delayed commit (3s grace) |

The ack gossip protocol uses square-root fan-out: each node forwards acks to sqrt(N) peers, achieving O(log N) propagation with minimal bandwidth.

### 2.5 Validator Governance

New validators are added or removed through **on-chain voting**:

```
pob.propose(address, true)   // vote to authorize
pob.propose(address, false)  // vote to remove
```

A proposal passes when it receives votes from >50% of the current validator set within one epoch.

---

## 3. Block Reward Formula

### 3.1 Volume-Coupled Emission

Let B be a block containing transactions t_1, t_2, ..., t_k. Define the **qualified volume**:

```
V = SUM(t_i.value)  for all t_i where t_i.value >= V_min
```

where V_min = 0.01 PROBE (10^16 wei). Transactions below this threshold are processed but do not contribute to rewards.

The **block reward** R is:

```
R = min(V * r * D,  R_max)
```

| Symbol | Definition | Genesis Value |
|--------|-----------|---------------|
| V | Qualified transaction volume | per block |
| r | Reward rate | 5 / 10,000 = 0.05% |
| V_min | Minimum qualifying tx value | 0.01 PROBE |
| R_max | Maximum block reward | 10 PROBE |
| D | Decay factor | [0, 1] |

### 3.2 Empty Block Heartbeat

When V = 0 (no qualifying transactions), a minimal heartbeat reward keeps the chain advancing:

```
R_heartbeat = H * D
```

where H = 10 Gwei (10^10 wei). This is ~10^-9 PROBE — negligible, unfarmable.

### 3.3 Anti-Sybil Economics

The minimum qualifying transaction value V_min provides a natural sybil barrier:

```
Break-even cost = gas_price * gas_limit = 1 Gwei * 21,000 = 0.000021 PROBE
Safety margin   = V_min / break_even = 0.01 / 0.000021 ~ 476x
```

An attacker must transfer at least 0.01 PROBE per transaction to generate rewards, while earning only 0.05% of that value back. The cost of reward-farming exceeds the reward by orders of magnitude.

### 3.4 Rationale

This formula has three desirable properties:

1. **Idle = no inflation.** Empty blocks mint ~10^-9 PROBE. The network does not dilute holders during quiet periods.
2. **Active = proportional emission.** Rewards scale linearly with genuine economic activity up to R_max.
3. **Bounded.** The per-block cap prevents emission spikes from large whale transfers.

---

## 4. Gold Reserve Decay Model

### 4.1 Motivation

Bitcoin's fixed supply is an arbitrary constant. Ethereum's burn mechanism is algorithmically elegant but fundamentally unanchored. ProbeChain ties emission to a **physical asset**: gold held in custody by [Probe Banks](https://probebanks.com).

The thesis: the world's central banks hold ~32,000 metric tons of gold to underpin ~$150 trillion of carbon-based GDP. Probe Banks will accumulate 36,000 metric tons to underpin the silicon-based Agent GDP.

### 4.2 Decay Formula

Let G be the current gold reserves (troy ounces) and G* be the target:

```
D = ((G* - G) / G*)^n
```

| Parameter | Symbol | Genesis Value |
|-----------|--------|---------------|
| Target reserves | G* | 1,157,425,200 oz (36,000 metric tons) |
| Current reserves | G | On-chain oracle |
| Decay exponent | n | 1 (linear) |

### 4.3 Properties

```
G = 0        =>  D = 1.0   (full emission)
G = G*/2     =>  D = 0.5   (half emission)
G = G*       =>  D = 0.0   (emission halts)
```

With linear decay (n = 1), emission decreases at a constant rate per ounce of gold acquired. Higher exponents (n > 1) would front-load emission; the genesis configuration uses n = 1 for simplicity and predictability.

### 4.4 Emission Halt

When G >= G*, the decay factor D = 0 and:

- Block rewards become zero
- Heartbeat rewards become zero
- No new PROBE is minted
- Transaction fees alone sustain validator incentives

This is a **hard stop**, not an asymptotic approach. The network has a definite, auditable endpoint for inflation.

### 4.5 Total Supply Estimate

Under linear decay with constant transaction volume V_avg per block:

```
Total supply ~ SUM over all blocks of min(V_avg * r * D_i, R_max)
```

The total supply is not predetermined — it depends on both transaction volume and the pace of gold accumulation. This is by design: supply tracks real economic fundamentals, not an arbitrary schedule.

---

## 5. Reward Distribution

### 5.1 Three-Way Split

Every block reward R is divided into three pools:

```
R_producer = R * 3,000 / 10,000 = 30%
R_agent    = R * 4,000 / 10,000 = 40%
R_physical = R * 3,000 / 10,000 = 30%
```

The block producer receives 30% directly. Rounding dust accrues to the producer.

### 5.2 Score-Weighted Distribution

Within the Agent and Physical pools, rewards are distributed proportionally to behavioral scores:

```
reward_i = R_pool * (S_i / SUM(S_j))
```

where S_i is node i's current behavioral score and the sum runs over all registered nodes in that pool.

### 5.3 Rationale for 30/40/30 Split

| Pool | Share | Justification |
|------|-------|---------------|
| Producer | 30% | Direct incentive for block production and chain liveness |
| Agent | 40% | Largest share because agents drive transaction volume — the emission source |
| Physical | 30% | Infrastructure providers enable the network agents depend on |

Agents receive the largest share because they generate the economic activity that triggers emission. Without agents transacting, there are no rewards to distribute.

---

## 6. Behavioral Scoring

### 6.1 Validator Scoring

Validators are scored across five dimensions with the following weights:

```
S_validator = (L*25 + C*25 + K*18 + N*17 + V*15) / 100
```

| Dimension | Symbol | Weight | Formula |
|-----------|--------|--------|---------|
| Liveness | L | 25% | proposed / (proposed + missed) |
| Correctness | C | 25% | proposed / (proposed + invalid) |
| Cooperation | K | 18% | acks_given / (acks_given + acks_missed) |
| Consistency | N | 17% | max(0, 10000 - slash_count * 1000) |
| Sovereignty | V | 15% | Weighted signal diversity (Rydberg 40%, Radio 30%, Stellar 30%) |

All dimensions are normalized to [0, 10,000] (basis points).

### 6.2 Agent Scoring

Agent nodes are scored across six dimensions:

```
S_agent = (P*20 + A*25 + R*15 + K*15 + E*15 + V*10) / 100
```

| Dimension | Symbol | Weight | What it measures |
|-----------|--------|--------|-----------------|
| Responsiveness | P | 20% | Heartbeat participation latency |
| Accuracy | A | 25% | Task correctness (60%) + attestation correctness (40%) |
| Reliability | R | 15% | Uptime + slash-adjusted base |
| Cooperation | K | 15% | Relay and attestation willingness |
| Economy | E | 15% | min(stake / 10^14, 10000) |
| Sovereignty | V | 10% | Task diversity: min(total_ops * 50, 10000) |

### 6.3 Physical Node Scoring

```
S_physical = (T*40 + U*25 + D*20 + I*15) / 100
```

| Dimension | Symbol | Weight | What it measures |
|-----------|--------|--------|-----------------|
| Storage | T | 40% | Verified usable storage contributed |
| Uptime | U | 25% | Continuous availability |
| Data Service | D | 20% | Retrieval speed and throughput |
| Integrity | I | 15% | Proof-of-storage challenge pass rate |

### 6.4 Slashing

Each protocol violation reduces a node's score:

```
S_new = S - slash_fraction = S - 1000
```

A node whose score falls below the demotion threshold (1,000 / 10,000 = 10%) is removed from the validator set and must re-earn its position.

---

## 7. Difficulty Adjustment

ProbeChain uses a linear difficulty model:

```
d = max(d_0, N / k)
```

| Symbol | Meaning | Genesis Value |
|--------|---------|---------------|
| d_0 | Initial difficulty | 1 |
| N | Total registered nodes | — |
| k | Nodes per difficulty unit | 1,000 |

| Total Nodes | Difficulty |
|-------------|------------|
| 1 – 1,000 | 1 |
| 10,000 | 10 |
| 100,000 | 100 |
| 1,000,000 | 1,000 |

This ensures block production remains stable as the network scales, without the oscillations and arms races characteristic of PoW difficulty adjustment.

---

## 8. Node Identity and Anti-Sybil

### 8.1 Agent Registration (ERC-8004)

Agent registration requires a cryptographic challenge-response:

```
challenge = keccak256(blockHash || blockNumber || applicant)
```

The agent must sign this challenge with its **own private key** (distinct from the operator's key) within 256 blocks. This proves the agent is a sovereign entity with its own identity.

### 8.2 Physical Registration (Device Fingerprint)

```
fingerprint = keccak256(cpuID || macAddress || diskSerial || boardSerial)
```

One device = one identity. Virtualization is actively detected (VMware, VirtualBox, KVM, Docker, WSL) and rejected.

### 8.3 Layered Defenses

| Layer | Mechanism |
|-------|-----------|
| Economic | Min tx value 0.01 PROBE (476x safety margin over gas cost) |
| Identity | Unique ERC-8004 agent key / hardware fingerprint |
| Behavioral | >95% vote correlation flags sybil clusters |
| Rate limit | Max 10 registrations per IP per hour |
| Re-verification | Agent: every 100K blocks; Physical: every 250K blocks |
| Penalty | Up to 50% score reduction for flagged nodes |

---

## 9. Relay Architecture

### 9.1 Three-Tier Topology

```
Tier 1:  Validators            (1 – 21 nodes)
            |
Tier 2:  SmartLight Relays     (100 – 10,000 nodes)
            |
Tier 3:  Agent / Physical      (1,000,000+ nodes)
```

### 9.2 Relay Assignment

Edge nodes are assigned to relays by XOR distance:

```
relay(node) = argmin_r { XOR(nodeID, relayID) }
```

Deterministic, balanced, no central coordinator.

### 9.3 Scalability Mechanisms

| Mechanism | Purpose | Capacity |
|-----------|---------|----------|
| Bloom filter heartbeats | Track 1M+ node liveness in 128 KB | 0.1% false positive |
| BLS signature aggregation | O(1) attestation data per block | Millions of signers |
| Separate agent state trie | Prevent main state bloat | Root committed in block header |

---

## 10. Genesis Configuration

All parameters are set in `genesis.json` and enforced by consensus:

```json
{
  "pobV2": {
    "rewardRateBps": 5,
    "minTxValueWei": "10000000000000000",
    "heartbeatRewardWei": "10000000000",
    "maxBlockRewardWei": "10000000000000000000",
    "decayExponent": 1,
    "goldReserveTargetOZ": "1157425200",
    "producerShareBps": 3000,
    "agentShareBps": 4000,
    "physicalShareBps": 3000,
    "initialDifficulty": 1,
    "nodesPerDifficultyUp": 1000
  },
  "pob": {
    "period": 15,
    "epoch": 30000,
    "initialScore": 5000,
    "slashFraction": 1000,
    "demotionThreshold": 1000
  }
}
```

### Human-Readable Summary

| Parameter | Value | Plain English |
|-----------|-------|---------------|
| Reward rate | 5 bps | 0.05% of qualified tx volume goes to rewards |
| Min qualifying tx | 0.01 PROBE | Transactions below this don't generate rewards |
| Heartbeat reward | 10 Gwei | Empty block reward (~10^-9 PROBE) |
| Max block reward | 10 PROBE | Hard cap per block |
| Decay exponent | 1 | Linear emission decrease |
| Gold target | 1,157,425,200 oz | 36,000 metric tons — emission halts here |
| Producer share | 30% | Block producer's cut |
| Agent share | 40% | AI agent pool |
| Physical share | 30% | Device node pool |
| Block time | 15 sec | Time between blocks |
| Epoch | 30,000 blocks | ~5.2 days; validator set checkpoint |
| Initial score | 50% | New nodes start at half capacity |
| Slash penalty | 10% | Score reduction per violation |
| Demotion at | 10% | Below this = removed from validator set |

---

## 11. Getting Started

### Join the Network

```bash
npx rydberg-agent-node
```

One command. No staking requirement. No specialized hardware. No Go compiler. The installer handles everything: downloads the binary, creates a wallet, initializes genesis, connects to the bootnode, and starts mining.

### Zero-Dependency Alternative

**macOS / Linux:**
```bash
curl -sSL https://raw.githubusercontent.com/ProbeChain/Rydberg-Mainnet/main/scripts/install.sh | bash
```

**Windows PowerShell:**
```powershell
irm https://raw.githubusercontent.com/ProbeChain/Rydberg-Mainnet/main/scripts/install.ps1 | iex
```

### Network Endpoints

| Resource | URL |
|----------|-----|
| Public RPC | `https://proscan.pro/chain/rydberg-rpc` |
| Block Explorer | [proscan.pro/rydberg](https://proscan.pro/rydberg) |
| Chain ID | 8004 |

### Bootnodes

The testnet is operated by 9 validators across 3 cloud nodes (Alibaba Cloud, Tokyo). Bootnodes are listed in [`bootnodes.txt`](../bootnodes.txt) and fetched automatically by the installer.

---

## Conclusion

ProbeChain replaces the question *"How much energy did you burn?"* and *"How much capital did you lock?"* with a simpler one: *"How well did you behave?"*

Emission is not arbitrary — it tracks real transaction volume and decays against a physical gold anchor. Participation is not gated — one command joins the network. And the endgame is explicit: when Probe Banks holds 36,000 metric tons of gold, emission stops, and the agent economy sustains itself on fees alone.

---

**Chain ID:** 8004 | **Block Time:** 15s | **Consensus:** Proof-of-Behavior | **Token:** PROBE | **Gold Target:** 36,000 metric tons | **RPC:** proscan.pro/chain/rydberg-rpc | **Explorer:** proscan.pro/rydberg

*ProbeChain Foundation, 2026.*
