# ProbeChain Rydberg Testnet

A Layer-1 blockchain for the agent economy. **Proof-of-Behavior** consensus. **Gold-anchored** emission. **One line to join.**

```bash
npx rydberg-agent-node
```

That's it. No staking. No hardware. No dependencies. Your node starts earning rewards in under 60 seconds.

---

## Why ProbeChain

|  | Bitcoin | Ethereum | **ProbeChain** |
|--|---------|----------|----------------|
| Consensus | Energy (PoW) | Capital (PoS) | **Behavior (PoB)** |
| Block time | 10 min | 12 sec | **15 sec** |
| Join cost | ASIC hardware | 32 ETH (~$80K) | **One command** |
| Empty block reward | Full subsidy | Full subsidy | **Near-zero** |
| Supply model | Fixed 21M | Deflationary | **Gold-anchored decay** |
| Emission driver | Time | Time | **Real tx volume** |

## How Rewards Work

Block rewards are proportional to real economic activity, not time:

```
R = min(V × r × D, R_max)
```

| Symbol | Meaning | Value |
|--------|---------|-------|
| V | Qualified transaction volume in block (value >= 0.01 PROBE) | — |
| r | Reward rate | 0.05% (5 bps) |
| R_max | Maximum reward per block | 10 PROBE |
| D | Gold reserve decay factor | 0 ~ 1 |

Empty blocks receive a heartbeat reward of 10 Gwei × D — enough to keep the chain alive, not enough to farm.

### Reward Split

Every block reward is divided three ways:

```
Block Producer .... 30%    (the validator who sealed the block)
Agent Pool ........ 40%    (AI agent nodes, by behavior score)
Physical Pool ..... 30%    (device nodes, by behavior score)
```

Within each pool, your share = your score / total scores.

### Gold Reserve Decay

PROBE emission decays as [Probe Banks](https://probebanks.com) accumulates physical gold reserves:

```
D = ((G_target - G_current) / G_target)^n
```

| Parameter | Value | Meaning |
|-----------|-------|---------|
| G_target | 1,157,425,200 oz | 36,000 metric tons of gold |
| G_current | On-chain oracle | Current Probe Banks reserves |
| n | 1 | Linear decay |

When reserves reach target, D = 0 and emission stops. Transaction fees alone sustain the network.

**Why gold?** The world's central banks hold ~32,000 tons of gold to underpin $150 trillion of carbon-based GDP. Probe Banks' 36,000 tons will underpin the silicon-based Agent GDP.

## Network Parameters

| Parameter | Value |
|-----------|-------|
| Chain ID | 8004 |
| Block time | 15 seconds |
| Gas limit | 30,000,000 |
| Token | PROBE (18 decimals) |
| Min qualified tx | 0.01 PROBE |
| Max block reward | 10 PROBE |
| Reward rate | 5 bps (0.05%) |
| Heartbeat reward | 10 Gwei |
| Epoch | 30,000 blocks (~5.2 days) |
| Validator slots | 1 – 21 |

## One-Line Install

### macOS / Linux (zero dependencies)
```bash
curl -sSL https://raw.githubusercontent.com/ProbeChain/Rydberg-Mainnet/main/scripts/install.sh | bash
```

### Windows PowerShell (zero dependencies)
```powershell
irm https://raw.githubusercontent.com/ProbeChain/Rydberg-Mainnet/main/scripts/install.ps1 | iex
```

### If you have Node.js
```bash
npx rydberg-agent-node
```

The installer downloads the binary, creates a wallet, initializes genesis, connects to the network, and starts mining. No manual steps.

## Node Operations

```bash
# Check status
~/rydberg-agent/gprobe attach ~/rydberg-agent/gprobe.ipc --exec "JSON.stringify({block: probe.blockNumber, peers: net.peerCount})"

# View logs
tail -f ~/rydberg-agent/node.log

# Stop
bash ~/rydberg-agent/stop.sh

# Restart
bash ~/rydberg-agent/start-bg.sh
```

## Behavioral Scoring

### Agent Nodes (AI agents)

| Dimension | Weight | What it measures |
|-----------|--------|-----------------|
| Responsiveness | 20% | Latency to consensus messages |
| Accuracy | 25% | Correctness of tasks and attestations |
| Reliability | 15% | Uptime over time |
| Cooperation | 15% | Willingness to relay and attest for peers |
| Economy | 15% | Stake-weighted efficiency |
| Sovereignty | 10% | Independence from coordinated voting |

### Physical Nodes (devices)

| Dimension | Weight | What it measures |
|-----------|--------|-----------------|
| Storage | 40% | Verified usable storage provided |
| Uptime | 25% | Continuous availability |
| Data Service | 20% | Retrieval speed and throughput |
| Integrity | 15% | Proof-of-storage correctness |

### Slashing

Each slash reduces a node's score by 10%. Nodes below 10% score are demoted from the validator set.

## Architecture

```
cmd/gprobe         CLI entry point
consensus/pob      Proof-of-Behavior consensus engine
core               Blockchain core: chain, state, tx pool, EVM
miner              Block production
probe              Protocol handlers, sync, peer management
p2p                Devp2p networking, node discovery
crypto             secp256k1, Dilithium (post-quantum)
params             Chain config and genesis parameters
```

## Documentation

- **[Technical Whitepaper](docs/ProbeChain-Rydberg-Whitepaper.md)** — Consensus formulas, emission model, scoring algorithms
- **[Genesis Config](genesis.json)** — All on-chain parameters

## Bootnode

```
enode://c56b6a7949fa9f6cf6e809863223fa9a444440a8f7fd4776ff5437f4c0db8d5775f7c0d3bfa0e6270242aa3811b776c9ef19d12c47a0f6e76f25b430a99071e9@bore.pub:9208
```

## License

GPL v3 / LGPL v3. See [COPYING](COPYING) and [COPYING.LESSER](COPYING.LESSER).
