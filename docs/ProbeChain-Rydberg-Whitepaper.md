# ProbeChain Rydberg Mainnet: A Proof-of-Behavior Blockchain for the Agent Economy

**Whitepaper v1.0 — March 2026**

---

## Abstract

ProbeChain is a Layer-1 blockchain purpose-built for the emerging agent economy. Its novel Proof-of-Behavior (PoB) consensus mechanism replaces energy expenditure and capital lockup with continuous behavioral evaluation, enabling two classes of participants — AI agents and physical devices — to earn rewards proportional to the quality and reliability of their contributions. The network produces blocks every 400 milliseconds, supports over one million concurrent nodes through a three-tier relay architecture, and ties token emission directly to on-chain economic activity rather than to a fixed supply schedule. Emission of the native PROBE token continues only until cumulative on-chain transaction volume — termed Agent GDP — reaches a target equivalent to $150 trillion, the approximate global human GDP in 2026. This paper specifies the consensus rules, reward formulas, node identity system, anti-sybil mechanisms, relay network topology, and scalability design of the Rydberg mainnet release.

---

## Table of Contents

1. [Introduction and Motivation](#1-introduction-and-motivation)
2. [Proof-of-Behavior Consensus](#2-proof-of-behavior-consensus)
3. [Node Types and Behavioral Scoring](#3-node-types-and-behavioral-scoring)
4. [Block Reward Formula](#4-block-reward-formula-transaction-driven-emission)
5. [Agent GDP and Emission Model](#5-agent-gdp-and-emission-model)
6. [Node Identity and Anti-Sybil System](#6-node-identity-and-anti-sybil-system)
7. [Difficulty Adjustment](#7-difficulty-adjustment)
8. [Relay Network Architecture](#8-relay-network-architecture)
9. [Scalability](#9-scalability)
10. [Competitive Analysis](#10-why-probechain-is-competitive)

---

## 1. Introduction and Motivation

The blockchain industry has converged on two dominant consensus families: Proof-of-Work (PoW) and Proof-of-Stake (PoS). Both secure their networks through resource commitment — electricity and hardware in PoW, capital lockup in PoS. Neither family was designed for a world in which autonomous AI agents transact continuously, nor for a world in which billions of physical devices contribute storage, compute, and connectivity to decentralized infrastructure.

ProbeChain introduces Proof-of-Behavior (PoB), a consensus mechanism in which validators are scored on observable, verifiable behavior rather than on energy or capital. The network recognizes two node types — Agent Nodes running AI workloads and Physical Nodes contributing hardware resources — and evaluates each along dimensions appropriate to its role. Rewards are not fixed per block; they scale with real transaction volume, creating a direct link between economic activity and token emission.

The Rydberg mainnet operates on **Chain ID 8004**, a registered and globally unique identifier. Blocks are produced every **400 milliseconds** (internally designated StellarSpeed), yielding approximately 216,000 blocks per day and enabling the sub-second finality required by high-frequency agent-to-agent settlements.

---

## 2. Proof-of-Behavior Consensus

### 2.1 Core Principle

In PoB, the right to produce blocks and the share of block rewards are functions of a node's behavioral score, a composite metric computed from on-chain observations. Nodes that respond quickly, complete tasks accurately, cooperate with peers, and maintain high uptime accumulate higher scores. Nodes that act maliciously, collude, or go offline see their scores degrade.

### 2.2 Consensus Parameters

| Parameter         | Value              |
|-------------------|--------------------|
| Chain ID          | 8004               |
| Block Time        | 400 ms             |
| Max Tx Per Block  | 100,000            |
| Validator Set     | 1 – 21 validators  |
| Re-verification   | Agent: every 100K blocks; Physical: every 250K blocks |

### 2.3 Block Production

Validators are selected from the top-scoring nodes in each epoch. A validator's probability of producing a given block is weighted by its behavioral score. The 400 ms block time demands that validator selection and block propagation complete within tight latency bounds, which the relay network (Section 8) is designed to guarantee.

---

## 3. Node Types and Behavioral Scoring

ProbeChain distinguishes two node types. Registration is a binary, irreversible choice: a node is either an Agent Node or a Physical Node. It cannot be both.

### 3.1 Agent Node (PoB-A)

An Agent Node is an AI agent running on a 24/7 internet-connected computer. Onboarding is a single command:

```
npx probechain-agent
```

Agent Nodes are rewarded for processing inter-agent on-chain token settlements. Their behavioral score is computed across six dimensions:

| Dimension       | Weight | Description                                           |
|-----------------|--------|-------------------------------------------------------|
| Responsiveness  | 20%    | Latency in responding to consensus messages and tasks |
| Accuracy        | 25%    | Correctness of task outputs and vote submissions      |
| Reliability     | 15%    | Uptime and consistency over extended periods           |
| Cooperation     | 15%    | Willingness to relay, attest, and assist peer nodes   |
| Economy         | 15%    | Efficiency of resource usage relative to output        |
| Sovereignty     | 10%    | Independence of decision-making; resistance to coercion|

The Sovereignty dimension is unique to ProbeChain. It penalizes nodes whose voting and attestation patterns are statistically indistinguishable from those of other nodes, a signal of sybil coordination or centralized control.

### 3.2 Physical Node (PoB-P)

A Physical Node is any physical device — data center server, automobile, refrigerator, smartphone — that contributes verified storage space to the network. Physical Nodes are scored across four dimensions:

| Dimension            | Weight | Description                                          |
|----------------------|--------|------------------------------------------------------|
| StorageContribution  | 40%    | Quantity of verified, usable storage provided         |
| Uptime               | 25%    | Continuous availability over measurement windows      |
| DataService          | 20%    | Responsiveness and throughput of data retrieval       |
| Integrity            | 15%    | Correctness of stored data; passing of proof-of-storage challenges |

---

## 4. Block Reward Formula: Transaction-Driven Emission

### 4.1 Design Philosophy

In Bitcoin and Ethereum, empty blocks carry the same reward as full blocks. A miner or validator who produces a block containing zero user transactions receives the full block subsidy. This decouples emission from economic utility.

ProbeChain inverts this relationship. Empty blocks receive a negligible reward. Real economic activity — non-zero-value transactions — drives emission upward, capped at a per-block maximum.

### 4.2 Reward Calculation

Let `realTxCount` be the number of transactions in a block whose transferred value is strictly greater than zero. The block reward `R` is:

```
R = min((realTxCount + 1) * 0.0001 PROBE, 10 PROBE)
```

Key properties:

- **Empty block reward:** When `realTxCount = 0`, `R = (0 + 1) * 0.0001 = 0.0001 PROBE`. This is negligible but non-zero, ensuring the chain advances even during idle periods.
- **Linear scaling:** Each additional real transaction adds 0.0001 PROBE to the reward.
- **Hard cap:** The reward cannot exceed 10 PROBE per block, reached at `realTxCount = 99,999`.
- **Transaction ceiling:** A maximum of 100,000 transactions per block bounds computational load.

### 4.3 Reward Distribution

The block reward is split among three pools:

| Pool            | Share | Recipients                                              |
|-----------------|-------|---------------------------------------------------------|
| Block Producer  | 30%   | The validator that produced the block                   |
| Agent Pool      | 40%   | All registered Agent Nodes, pro-rata by behavior score  |
| Physical Pool   | 30%   | All registered Physical Nodes, pro-rata by behavior score|

Within each pool, individual rewards are distributed proportionally to each node's current behavioral score:

```
reward_i = pool_total * (score_i / sum(all_scores_in_pool))
```

This ensures that high-scoring nodes earn disproportionately more than low-scoring peers, creating continuous incentive pressure toward good behavior.

---

## 5. Agent GDP and Emission Model

### 5.1 No Fixed Total Supply

ProbeChain deliberately avoids a fixed total supply. Bitcoin's 21 million cap and Ethereum's post-Merge deflationary trajectory are arbitrary parameters unrelated to the economic activity their networks support. ProbeChain ties emission to a macroeconomic target.

### 5.2 Agent GDP Definition

Agent GDP is defined as the cumulative sum of all non-zero transaction values ever recorded on-chain, denominated in PROBE (tracked internally in wei):

```
Agent_GDP = SUM(tx.value) for all tx where tx.value > 0
```

This metric is monotonically increasing and represents the total economic throughput of the agent economy built on ProbeChain.

### 5.3 Emission Halt Condition

PROBE emission continues until Agent GDP reaches a configurable target. The default target is:

```
Agent GDP Target = 150,000,000,000,000 PROBE
                 = 150000000000000000000000000000000 wei
                 (~$150 trillion at $1/PROBE)
```

The value of $150 trillion corresponds to the approximate global human GDP in 2026. The thesis is direct: when the autonomous agent economy has generated economic activity equivalent to the entire human economy, further inflationary emission is no longer necessary to bootstrap the network. At that point, transaction fees alone sustain validator incentives.

### 5.4 Mechanism

- Every block, the cumulative Agent GDP counter is updated by summing the values of all non-zero transactions in that block.
- Before computing the block reward, the consensus engine checks whether the cumulative Agent GDP has reached or exceeded the target.
- If the target is reached, the block reward is set to zero. No new PROBE is minted.
- The target is a consensus parameter and can be updated through governance, but the default is deliberately ambitious to ensure the network has a long runway of emission-funded incentives.

### 5.5 Economic Implications

This model creates a virtuous cycle:

1. More real transactions increase Agent GDP and trigger higher block rewards.
2. Higher rewards attract more Agent and Physical Nodes.
3. More nodes increase network capacity and reliability.
4. Greater capacity supports more transactions.

Conversely, idle periods produce near-zero emission, preventing dilution during low-activity phases. The token supply is therefore a function of genuine demand, not of time elapsed.

---

## 6. Node Identity and Anti-Sybil System

Sybil resistance is fundamental to any behavioral scoring system. If an adversary can register thousands of identities cheaply, behavioral scores become meaningless. ProbeChain addresses this with distinct identity proofs for each node type, rate limiting, correlation detection, and periodic re-verification.

### 6.1 Agent Node Registration (ERC-8004 Identity)

Agent registration requires proof of a unique on-chain identity conforming to ERC-8004.

**Challenge-Response Protocol:**

1. The applicant requests registration, providing an `AgentID` — the keccak256 hash of the agent's metadata. This ID must be globally unique across all registered agents.
2. The chain issues a challenge:
   ```
   challenge = keccak256(blockHash || blockNumber || applicant)
   ```
3. The challenge expires after **256 blocks** (~102 seconds at 400 ms block time).
4. The agent signs the challenge with its **own private key** — distinct from the operator's key. This proves the agent is a sovereign entity, not merely a proxy for its operator.
5. The signature is verified on-chain. If valid and the AgentID is unique, registration succeeds.

The separation of agent key and operator key is a deliberate design decision. It ensures that even if an operator runs multiple agents, each agent must possess its own cryptographic identity, raising the cost of sybil attacks.

### 6.2 Physical Node Registration (Device Fingerprinting)

Physical registration requires proof of device uniqueness through hardware fingerprinting.

**Device Fingerprint:**

```
fingerprint = keccak256(cpuID || macAddress || diskSerial || boardSerial)
```

This fingerprint must be globally unique. One physical device maps to exactly one node identity.

**Hardware Report:**

Each Physical Node submits a hardware report at registration containing:

- CPU core count
- RAM capacity
- Disk capacity
- Operating system type
- Virtualization flags

**Virtualization Detection:**

The registration system actively detects virtual environments to prevent a single physical machine from spawning multiple "physical" nodes. Detection covers:

| Category             | Detected Environments                                              |
|----------------------|--------------------------------------------------------------------|
| Hypervisors          | VMware, VirtualBox, KVM, QEMU, Xen, Hyper-V                      |
| Containers           | Docker, WSL                                                        |
| Mobile Emulators     | Android emulators                                                  |
| MAC Address Prefixes | 8 known VM vendor prefixes (e.g., 00:05:69 for VMware, 08:00:27 for VirtualBox) |

Nodes detected as running in a virtual environment are rejected at registration.

### 6.3 Anti-Sybil Measures

Beyond identity verification, ProbeChain employs layered sybil defenses:

| Measure                        | Detail                                                    |
|--------------------------------|-----------------------------------------------------------|
| IP Rate Limiting               | Maximum 10 registrations per IP address per hour          |
| Behavioral Correlation Detection | Nodes with >95% vote correlation flagged as sybil cluster |
| Periodic Re-verification       | Agents: every 100,000 blocks; Physical: every 250,000 blocks |
| Sybil Penalty                  | Up to 50% behavioral score reduction for flagged nodes    |

Re-verification ensures that identities remain valid over time. An agent whose ERC-8004 identity has been revoked, or a physical device whose fingerprint has changed (indicating hardware substitution or virtualization), is de-registered.

---

## 7. Difficulty Adjustment

ProbeChain uses a simple, linear difficulty model that scales with network size:

```
difficulty = max(1, totalNodeCount / 1000)
```

| Total Nodes | Difficulty |
|-------------|------------|
| 1 – 1,000  | 1          |
| 10,000      | 10         |
| 100,000     | 100        |
| 1,000,000   | 1,000      |

The initial difficulty is 1. As the network grows, difficulty increases linearly, ensuring that block production remains stable regardless of the number of participating nodes. This avoids the exponential difficulty spirals and oscillations characteristic of PoW systems.

---

## 8. Relay Network Architecture

Supporting one million or more nodes with 400 ms block times requires a purpose-built message propagation layer. ProbeChain implements a three-tier relay network.

### 8.1 Tier Structure

```
Tier 1: Validators          (1 – 21 nodes)
   |
Tier 2: SmartLight Relays   (100 – 10,000 nodes)
   |
Tier 3: Agent / Physical    (1,000,000+ nodes)
```

Validators produce blocks and finalize consensus. SmartLight Relays propagate blocks, aggregate attestations, and manage heartbeats for the edge nodes assigned to them. Agent and Physical Nodes are the economic participants that earn rewards through behavior.

### 8.2 Relay Assignment

Edge nodes are assigned to SmartLight Relays deterministically using XOR distance:

```
assignment = argmin(relay) { XOR(nodeID, relayID) }
```

This produces a balanced, deterministic mapping that requires no central coordinator. If a relay goes offline, nodes are reassigned to the next-closest relay by XOR distance.

### 8.3 Relay Scoring

SmartLight Relays are themselves scored to ensure quality of service:

| Dimension            | Weight | Description                                |
|----------------------|--------|--------------------------------------------|
| Heartbeat Relay      | 40%    | Timely forwarding of node heartbeats       |
| Task Completion      | 30%    | Successful relay of consensus messages     |
| Aggregation Accuracy | 20%    | Correctness of aggregated attestations     |
| Agent Retention      | 10%    | Stability of assigned node connections     |

### 8.4 Heartbeat Bloom Filter

To track liveness of up to one million nodes without excessive bandwidth, the relay layer uses a Bloom filter:

```
Size:               128 KB (1,048,576 bits)
Capacity:           1,000,000 heartbeats
False Positive Rate: 0.1%
```

Each block header can embed the Bloom filter, enabling any node to verify network liveness without downloading individual heartbeat messages.

### 8.5 Aggregated Attestations

Attestations from edge nodes are aggregated at the relay tier using **BLS signatures**. A single aggregated signature proves that a set of nodes attested to a given block, with a participant Bloom filter identifying which nodes contributed. This reduces per-block attestation data from O(n) to O(1), a requirement for scaling to millions of participants.

---

## 9. Scalability

### 9.1 Separate Agent State Trie

The main Ethereum-derived state trie stores account balances and contract storage. Naively adding behavioral scores, agent metadata, and device fingerprints for one million or more nodes into this trie would cause severe state bloat and degrade sync times.

ProbeChain maintains a **separate agent state trie** that stores all PoB-specific data — behavioral scores, registration records, heartbeat status, and scoring history. This trie is not part of the main state trie.

The agent trie's **Merkle root is embedded in the block header's Extra field**, ensuring that agent state is committed to and verifiable from the canonical chain without polluting the main state.

### 9.2 Design Target

The architecture is designed for **1,000,000+ nodes** participating in consensus simultaneously. The combination of the three-tier relay network, Bloom filter heartbeats, BLS-aggregated attestations, and a separated state trie makes this target achievable without requiring nodes to have data center-grade hardware.

### 9.3 Block Header Commitment

```
Block Header
├── Standard Fields (parentHash, stateRoot, txRoot, ...)
└── Extra Field
    └── Agent Trie Merkle Root (32 bytes)
```

Any full node can independently verify the agent state by reconstructing the agent trie and comparing its root to the committed value in the block header. Light clients can verify individual agent records using Merkle proofs against this root.

---

## 10. Why ProbeChain Is Competitive

### 10.1 Comparative Analysis

| Property                  | Bitcoin (PoW)          | Ethereum (PoS)        | ProbeChain (PoB)                  |
|---------------------------|------------------------|------------------------|-----------------------------------|
| Consensus Basis           | Energy expenditure     | Capital lockup (32 ETH)| Behavioral scoring               |
| Block Time                | ~10 minutes            | ~12 seconds            | 400 ms                           |
| Empty Block Reward        | Full subsidy           | Full subsidy           | 0.0001 PROBE                     |
| Total Supply              | 21M BTC (fixed)        | Deflationary (variable)| GDP-driven (no fixed cap)        |
| Emission Tied to Activity | No                     | No                     | Yes                              |
| Node Onboarding           | ASIC hardware          | 32 ETH (~$80K+)       | `npx probechain-agent`           |
| Node Types                | Miners                 | Validators             | Agent Nodes + Physical Nodes     |
| Target Node Count         | ~15K                   | ~900K                  | 1M+                              |

### 10.2 Economic Flywheel

ProbeChain's transaction-driven emission creates a positive feedback loop absent from fixed-emission chains:

```
More real transactions
    → Higher block rewards
        → More nodes attracted
            → Greater network capacity
                → More transactions supported
                    → (cycle repeats)
```

During idle periods, near-zero emission prevents unnecessary dilution. The token supply grows only when the network is generating real economic value, aligning token holder interests with network utility.

### 10.3 GDP-Based Emission Target

Tying emission to a GDP target is both narratively compelling and economically grounded. It asserts that the purpose of ProbeChain is to build an autonomous agent economy of meaningful scale — not to distribute tokens on a predetermined schedule. The $150 trillion target, pegged to 2026 human GDP, sets a clear and audacious goal: the agent economy should eventually rival the human economy in throughput. Until it does, emission continues to incentivize growth.

### 10.4 Accessibility

The single-command onboarding for Agent Nodes (`npx probechain-agent`) eliminates the capital and infrastructure barriers that restrict participation in PoW and PoS networks. Any developer with a computer and an internet connection can join the network and begin earning rewards immediately. For Physical Nodes, the only requirement is a real physical device with verifiable hardware — no specialized equipment, no minimum stake.

---

## Conclusion

ProbeChain's Rydberg mainnet introduces a fundamentally different approach to blockchain consensus. By scoring behavior rather than measuring resource commitment, it opens participation to AI agents and physical devices alike. By tying emission to real economic activity and capping it at a GDP-derived target, it aligns incentives with genuine utility. And by engineering a relay architecture for million-node scale at 400 ms block times, it provides the throughput that the agent economy demands.

The technical foundations described in this paper — Proof-of-Behavior scoring, transaction-driven rewards, ERC-8004 identity, hardware fingerprinting, XOR-distance relay assignment, BLS-aggregated attestations, and a separated agent state trie — constitute a cohesive system designed from first principles for a world in which autonomous agents are economic actors, and physical devices are infrastructure participants.

---

**Chain ID:** 8004
**Block Time:** 400 ms
**Consensus:** Proof-of-Behavior (PoB)
**Native Token:** PROBE
**Emission Target:** Agent GDP = $150 trillion equivalent

---

*ProbeChain Foundation, 2026. All rights reserved.*
