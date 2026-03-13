# ProbeChain Rydberg Mainnet: OZ Gold Standard for the Agent Economy

**Whitepaper v2.1 — March 2026**

---

## Abstract

ProbeChain is a Layer-1 blockchain purpose-built for the emerging agent economy. Its Proof-of-Behavior (PoB) consensus mechanism evaluates nodes on observable behavior rather than energy expenditure or capital lockup, enabling two classes of participants — AI agents and physical devices — to earn rewards proportional to the quality of their contributions.

**PoB V2.1** introduces the **OZ Gold Standard**: PROBE token emission is coupled to on-chain transaction volume and governed by a gold-reserve decay function tied to Probe Banks' physical gold reserves. When Probe Banks accumulates 36,000 metric tons of gold (1,157,425,200 troy ounces of OZ), PROBE emission halts permanently. This parallels how global central banks' ~32,000 tons of gold support $150 trillion in human GDP — Probe Banks' 36,000 tons will support the silicon-based Agent GDP.

This paper specifies the consensus rules, reward formulas, anti-Sybil economics, node identity system, and network architecture.

---

## Table of Contents

1. [Introduction and Motivation](#1-introduction-and-motivation)
2. [Proof-of-Behavior Consensus](#2-proof-of-behavior-consensus)
3. [Node Types and Behavioral Scoring](#3-node-types-and-behavioral-scoring)
4. [OZ Gold Standard: Block Reward Formula](#4-oz-gold-standard-block-reward-formula)
5. [Anti-Sybil Economics](#5-anti-sybil-economics)
6. [Gold-Reserve Decay and Emission Halt](#6-gold-reserve-decay-and-emission-halt)
7. [Node Identity and Anti-Sybil System](#7-node-identity-and-anti-sybil-system)
8. [Difficulty Adjustment](#8-difficulty-adjustment)
9. [Relay Network Architecture](#9-relay-network-architecture)
10. [Scalability](#10-scalability)
11. [Competitive Analysis](#11-why-probechain-is-competitive)

---

## 1. Introduction and Motivation

The blockchain industry has converged on two dominant consensus families: Proof-of-Work (PoW) and Proof-of-Stake (PoS). Both secure their networks through resource commitment — electricity and hardware in PoW, capital lockup in PoS. Neither was designed for a world in which autonomous AI agents transact continuously, nor for a world in which billions of physical devices contribute to decentralized infrastructure.

**Bitcoin solved central bank money printing** with a fixed supply and time-based halving. **ProbeChain solves Agent GDP measurement, settlement, and behavior governance** with volume-coupled emission and gold-reserve-based decay.

ProbeChain introduces Proof-of-Behavior (PoB), a consensus mechanism in which validators are scored on observable, verifiable behavior. The network recognizes two node types — Agent Nodes running AI workloads and Physical Nodes contributing hardware resources — and evaluates each along dimensions appropriate to its role.

The Rydberg mainnet operates on **Chain ID 8004**. Blocks are produced every **400 milliseconds** (StellarSpeed), yielding approximately 216,000 blocks per day and enabling the sub-second finality required by high-frequency agent-to-agent settlements.

---

## 2. Proof-of-Behavior Consensus

### 2.1 Core Principle

In PoB, the right to produce blocks and the share of block rewards are functions of a node's behavioral score. Nodes that respond quickly, complete tasks accurately, cooperate with peers, and maintain high uptime accumulate higher scores. Nodes that act maliciously, collude, or go offline see their scores degrade.

### 2.2 Consensus Parameters

| Parameter         | Value              |
|-------------------|--------------------|
| Chain ID          | 8004               |
| Block Time        | 400 ms             |
| Gas Limit         | 30,000,000         |
| Base Fee Floor    | 1 Gwei             |
| Validator Set     | 1 – 21 validators  |
| Epoch             | 30,000 blocks      |

### 2.3 Block Production

Validators are selected from the top-scoring nodes in each epoch. A validator's probability of producing a given block is weighted by its behavioral score. The 400 ms block time demands tight latency bounds, which the relay network (Section 9) guarantees.

---

## 3. Node Types and Behavioral Scoring

ProbeChain distinguishes two node types. Registration is a binary, irreversible choice.

### 3.1 Agent Node (PoB-A)

An Agent Node is an AI agent running on a 24/7 internet-connected computer. Agent Nodes are rewarded for processing inter-agent on-chain token settlements. Their behavioral score is computed across six dimensions:

| Dimension       | Weight | Description                                           |
|-----------------|--------|-------------------------------------------------------|
| Responsiveness  | 20%    | Latency in responding to consensus messages and tasks |
| Accuracy        | 25%    | Correctness of task outputs and vote submissions      |
| Reliability     | 15%    | Uptime and consistency over extended periods           |
| Cooperation     | 15%    | Willingness to relay, attest, and assist peer nodes   |
| Economy         | 15%    | Efficiency of resource usage relative to output        |
| Sovereignty     | 10%    | Independence of decision-making; resistance to coercion|

### 3.2 Physical Node (PoB-P)

A Physical Node is any physical device that contributes verified storage space to the network. Physical Nodes are scored across four dimensions:

| Dimension            | Weight | Description                                          |
|----------------------|--------|------------------------------------------------------|
| StorageContribution  | 40%    | Quantity of verified, usable storage provided         |
| Uptime               | 25%    | Continuous availability over measurement windows      |
| DataService          | 20%    | Responsiveness and throughput of data retrieval       |
| Integrity            | 15%    | Correctness of stored data; passing proof-of-storage challenges |

---

## 4. OZ Gold Standard: Block Reward Formula

### 4.1 Design Philosophy

In Bitcoin and Ethereum, empty blocks carry the same reward as full blocks, decoupling emission from economic utility. ProbeChain inverts this: rewards are proportional to real economic activity (transaction volume), and emission decreases as physical gold reserves grow.

**OZ** is Probe Banks' gold-backed stablecoin:
- 1 OZ = 1 troy ounce of 99.99% pure gold
- 100% physical gold backing. No fractional reserve.
- `OZ.totalSupply()` on-chain = Probe Banks' actual gold reserves in troy ounces

### 4.2 Qualified Transaction Volume

Not all transactions contribute to the block reward. A transaction qualifies only if:

1. **Value >= MinTxValue** (0.01 PROBE = 10^16 wei)
2. **Not a self-transfer** (sender != recipient, when sender info is available)

```
qualifiedVolume = Σ tx.Value()  for all qualifying transactions in the block
```

The MinTxValue filter prevents micro-transaction spam from inflating rewards.

### 4.3 Block Reward Calculation

```
decay  = (1 - goldReserveOZ / 1,157,425,200)
reward = min(qualifiedVolume × rewardRate × decay, maxBlockReward)
```

Where:
- `rewardRate` = 5 basis points (0.05%)
- `maxBlockReward` = 10 PROBE (10^19 wei)
- `decay` ∈ [0, 1] — linear decay based on Probe Banks' gold reserves

**Empty blocks** receive a heartbeat reward:
```
heartbeatReward = 0.00000001 PROBE × decay = 10^10 wei × decay
```

### 4.4 Genesis Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| `rewardRateBps` | 5 | 0.05% of qualified volume |
| `minTxValueWei` | 10^16 | 0.01 PROBE minimum |
| `heartbeatRewardWei` | 10^10 | 0.00000001 PROBE |
| `maxBlockRewardWei` | 10^19 | 10 PROBE cap |
| `decayExponent` | 1 | Linear decay |
| `goldReserveTargetOZ` | 1,157,425,200 | 36,000 metric tons |

### 4.5 Reward Distribution

The block reward is split among three pools:

| Pool            | Share | Recipients                                              |
|-----------------|-------|---------------------------------------------------------|
| Block Producer  | 30%   | The validator that produced the block                   |
| Agent Pool      | 40%   | All registered Agent Nodes, pro-rata by behavior score  |
| Physical Pool   | 30%   | All registered Physical Nodes, pro-rata by behavior score|

Any remainder from integer division goes to the block producer.

---

## 5. Anti-Sybil Economics

### 5.1 The Problem

If an attacker can spam the network with zero-value or micro-value transactions to inflate qualified volume, they earn rewards that exceed their gas costs — a profitable Sybil attack.

### 5.2 The Solution: EIP-1559 Base Fee Floor

ProbeChain modifies the EIP-1559 base fee calculation to enforce a **floor of 1 Gwei** (0.000000001 PROBE). Without this floor, the base fee drops to zero within ~40 seconds of empty blocks (12.5% decrease per block at 400ms), enabling zero-cost spam.

```
baseFee = max(calculated_baseFee, 1 Gwei)
```

### 5.3 Cost-Reward Analysis

For a single transaction at the minimum qualified value (0.01 PROBE):

```
Reward:   0.01 PROBE × 5/10000 = 0.000005 PROBE
Gas cost: 21,000 gas × 1 Gwei  = 0.000021 PROBE (burned via EIP-1559)

Net:      -0.000016 PROBE per transaction (guaranteed loss)
Safety margin: 4.2×
```

**An attacker loses 4.2 times more in gas than they earn in rewards.** This makes Sybil attacks economically irrational at any scale.

### 5.4 Break-Even Analysis

The break-even transaction value is:

```
breakEven = txGas × gasPrice / rewardRate
          = 21,000 × 1 Gwei / (5/10000)
          = 0.042 PROBE
```

Transactions below 0.042 PROBE generate less reward than their gas cost. The MinTxValue of 0.01 PROBE sits well below this threshold, providing the 4.2× safety margin.

---

## 6. Gold-Reserve Decay and Emission Halt

### 6.1 The OZ Gold Standard

Unlike Bitcoin's arbitrary 21 million cap, ProbeChain's emission is governed by a physically verifiable metric: **Probe Banks' gold reserves**.

- **OZ** is Probe Banks' gold-backed stablecoin (1 OZ = 1 troy ounce of 99.99% gold)
- Probe Banks maintains 100% physical gold backing — no fractional reserve
- The on-chain `totalSupply()` of OZ equals Probe Banks' actual gold reserves in troy ounces
- Gold is stored across multiple jurisdictions (Mongolia, Switzerland, Singapore, UAE)

### 6.2 Decay Function

As Probe Banks accumulates more gold, the emission rate decreases linearly:

```
decay = (targetOZ - goldReserveOZ) / targetOZ
```

| Gold Reserve | Decay Factor | Emission Rate |
|-------------|--------------|---------------|
| 0 OZ | 100% | Full emission |
| 115,742,520 OZ (3,600 tons) | 90% | 90% of base |
| 578,712,600 OZ (18,000 tons) | 50% | Half emission |
| 1,041,682,680 OZ (32,400 tons) | 10% | Near-zero |
| 1,157,425,200 OZ (36,000 tons) | 0% | **Emission stops** |

### 6.3 Why 36,000 Metric Tons?

Global central banks collectively hold ~32,000 metric tons of gold, which underlies the ~$150 trillion carbon-based (human) GDP. By targeting 36,000 metric tons — exceeding central bank reserves — Probe Banks creates a gold reserve sufficient to underpin the silicon-based Agent GDP that ProbeChain is designed to measure and settle.

### 6.4 Smooth Transition

Unlike step-function halving (Bitcoin), the linear decay ensures a smooth, predictable reduction in emission. There are no sudden cliff events. As gold reserves grow, emission gently approaches zero, allowing the market to adjust gradually.

### 6.5 Post-Emission Economics

When emission halts, validators are sustained by transaction fees alone (EIP-1559 base fees + priority tips). The 1 Gwei base fee floor ensures validators always have minimum compensation, even during low-activity periods.

---

## 7. Node Identity and Anti-Sybil System

### 7.1 Agent Node Registration (ERC-8004)

Agent registration requires proof of a unique on-chain identity conforming to ERC-8004.

1. The applicant provides an `AgentID` (keccak256 of metadata), which must be globally unique.
2. The chain issues a challenge: `keccak256(blockHash || blockNumber || applicant)`.
3. The challenge expires after 256 blocks (~102 seconds).
4. The agent signs the challenge with its own private key (distinct from the operator's key).
5. If valid and unique, registration succeeds.

### 7.2 Physical Node Registration (Device Fingerprinting)

```
fingerprint = keccak256(cpuID || macAddress || diskSerial || boardSerial)
```

Fingerprints must be globally unique. Virtualization detection covers VMware, VirtualBox, KVM, QEMU, Xen, Hyper-V, Docker, and WSL.

### 7.3 Layered Defenses

| Measure | Detail |
|---------|--------|
| MinTxValue filter | 0.01 PROBE minimum for reward qualification |
| Base fee floor | 1 Gwei — guaranteed gas cost |
| Self-transfer ban | Sender != recipient for reward qualification |
| IP rate limiting | Max 10 registrations per IP per hour |
| Behavioral correlation | >95% vote correlation flags sybil clusters |
| Periodic re-verification | Agents: every 100K blocks; Physical: every 250K blocks |

---

## 8. Difficulty Adjustment

```
difficulty = max(1, totalNodeCount / 1000)
```

| Total Nodes | Difficulty |
|-------------|------------|
| 1 – 1,000  | 1          |
| 10,000      | 10         |
| 100,000     | 100        |
| 1,000,000   | 1,000      |

---

## 9. Relay Network Architecture

### 9.1 Three-Tier Structure

```
Tier 1: Validators          (1 – 21 nodes)
   |
Tier 2: SmartLight Relays   (100 – 10,000 nodes)
   |
Tier 3: Agent / Physical    (1,000,000+ nodes)
```

### 9.2 Relay Assignment

Edge nodes are assigned to SmartLight Relays via XOR distance:

```
assignment = argmin(relay) { XOR(nodeID, relayID) }
```

### 9.3 Heartbeat Bloom Filter

```
Size:               128 KB (1,048,576 bits)
Capacity:           1,000,000 heartbeats
False Positive Rate: 0.1%
```

### 9.4 BLS Aggregated Attestations

Attestations from edge nodes are aggregated at the relay tier using BLS signatures, reducing per-block attestation data from O(n) to O(1).

---

## 10. Scalability

### 10.1 Separate Agent State Trie

PoB-specific data (behavioral scores, registration records, heartbeats) is stored in a separate agent state trie. Its Merkle root is embedded in the block header's Extra field, keeping the main state trie clean.

### 10.2 Design Target

1,000,000+ nodes participating simultaneously, enabled by the three-tier relay, Bloom filter heartbeats, BLS aggregation, and separated state trie.

---

## 11. Why ProbeChain Is Competitive

| Property | Bitcoin (PoW) | Ethereum (PoS) | ProbeChain (PoB) |
|----------|--------------|-----------------|-------------------|
| Consensus Basis | Energy | Capital (32 ETH) | Behavioral scoring |
| Block Time | ~10 min | ~12 sec | 400 ms |
| Empty Block Reward | Full subsidy | Full subsidy | 0.00000001 PROBE |
| Supply Model | 21M fixed | Deflationary | OZ Gold Standard |
| Emission Tied to Activity | No | No | Yes (volume-coupled) |
| Emission Decay | Time-based halving | Burn rate | Gold-reserve linear decay |
| Anti-Sybil | PoW cost | 32 ETH stake | Gas floor + MinTxValue (4.2x margin) |
| Node Types | Miners | Validators | Agent + Physical |
| Target Nodes | ~15K | ~900K | 1M+ |

### Economic Flywheel

```
More real transactions
    → Higher block rewards
        → More nodes attracted
            → Greater network capacity
                → PROBE price appreciation
                    → Less PROBE needed for same OZ-denominated GDP
                        → Less emission pressure
                            → Greater scarcity
                                → (virtuous cycle)
```

---

## Conclusion

ProbeChain's PoB V2.1 with the OZ Gold Standard represents a fundamentally different approach to blockchain economics. By tying emission to both real transaction volume and physical gold reserves, it creates a system where:

1. **Token emission reflects genuine economic activity**, not elapsed time
2. **Emission decay is physically verifiable** — Probe Banks' gold reserves are auditable
3. **Anti-Sybil protection is mathematically guaranteed** — attacking costs 4.2x more than it rewards
4. **The transition to fee-only is smooth** — linear decay, no cliff events

The design philosophy is clear: Bitcoin solved central bank money printing with a fixed supply. ProbeChain solves Agent GDP measurement, settlement, and behavior governance with volume-coupled emission and gold-reserve-based decay.

---

**Chain ID:** 8004
**Block Time:** 400 ms
**Consensus:** Proof-of-Behavior V2.1
**Native Token:** PROBE
**Emission Model:** OZ Gold Standard
**Emission Halt:** 36,000 metric tons of gold (1,157,425,200 OZ)

---

*ProbeChain Foundation, 2026. All rights reserved.*
