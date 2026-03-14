# ProbeChain Rydberg Testnet

A Layer-1 blockchain for the agent economy. **Proof-of-Behavior** consensus. **Gold-anchored** emission. **One line to join.**

```bash
npx rydberg-agent-node
```

That's it. No staking. No hardware. No dependencies. Your node starts earning rewards in under 60 seconds.

---

## Network Info

| Field | Value |
|-------|-------|
| Chain ID | `8004` |
| RPC | `https://proscan.pro/chain/rydberg-rpc` |
| Explorer | [proscan.pro/rydberg](https://proscan.pro/rydberg) |
| Token | PROBE (18 decimals) |
| Block time | ~15 seconds |
| Consensus | Proof-of-Behavior (PoB) |
| Validators | 9 (multi-node, cloud-hosted) |

### Add to MetaMask

| Field | Value |
|-------|-------|
| Network Name | ProbeChain Rydberg Testnet |
| RPC URL | `https://proscan.pro/chain/rydberg-rpc` |
| Chain ID | `8004` |
| Currency Symbol | `PROBE` |
| Block Explorer | `https://proscan.pro/rydberg` |

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
R = min(V * r * D, R_max)
```

| Symbol | Meaning | Value |
|--------|---------|-------|
| V | Qualified transaction volume in block (value >= 0.01 PROBE) | -- |
| r | Reward rate | 0.05% (5 bps) |
| R_max | Maximum reward per block | 10 PROBE |
| D | Gold reserve decay factor | 0 ~ 1 |

Empty blocks receive a heartbeat reward of 10 Gwei * D -- enough to keep the chain alive, not enough to farm.

### Reward Split

Every block reward is divided three ways:

```
Block Producer .... 30%    (the validator who sealed the block)
Agent Pool ........ 40%    (AI agent nodes, by behavior score)
Physical Pool ..... 30%    (device nodes, by behavior score)
```

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

---

## One-Line Install

### macOS / Linux
```bash
curl -sSL https://raw.githubusercontent.com/ProbeChain/Rydberg-Mainnet/main/scripts/install.sh | bash
```

### Windows PowerShell
```powershell
irm https://raw.githubusercontent.com/ProbeChain/Rydberg-Mainnet/main/scripts/install.ps1 | iex
```

### If you have Node.js
```bash
npx rydberg-agent-node
```

The installer downloads the binary, creates a wallet, initializes genesis, connects to bootnodes, and starts mining. No manual steps.

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

---

## Bootnodes

The testnet runs on 3 dedicated cloud nodes (Alibaba Cloud, Tokyo region) hosting 9 validators:

```
enode://e963c6b5342f0af311b0347fab33238aa564617241222dc93a4691ef2c76990b69e87e98a300aa08fc4deba023017006cd38e7cf28aa92bf13a93e8ff0c1387a@8.216.37.182:30398
enode://e742e55bae150ad4f004642d3a7365d2fe07f14b3ee7105fa47f6257e55937a90e5380b1a879c2e1e40295b0b482553536822b38615eecf444b5d8394c21a26e@8.216.49.15:30398
enode://bfb54dde94a526375a7ddbec0a4e5e174394c02f8a1c4c5ebec2aa5a05188398849895fc11973b09f3eec49dbd12ea9c8cb924e194e2857560c87d2f8a557bad@8.216.32.20:30398
```

---

## Architecture

```
cmd/gprobe         CLI entry point
consensus/pob      Proof-of-Behavior consensus engine
core               Blockchain core: chain, state, tx pool, EVM
miner              Block production and ack-based consensus loop
probe              Protocol handlers, sync, peer management
p2p                Devp2p networking, node discovery
crypto             secp256k1, Dilithium (post-quantum)
params             Chain config and genesis parameters
```

## Documentation

- **[Technical Whitepaper](docs/ProbeChain-Rydberg-Whitepaper.md)** -- Consensus formulas, emission model, scoring algorithms
- **[Genesis Config](genesis.json)** -- All on-chain parameters

## Links

| Resource | URL |
|----------|-----|
| Public RPC | `https://proscan.pro/chain/rydberg-rpc` |
| Block Explorer | [proscan.pro/rydberg](https://proscan.pro/rydberg) |
| Website | [probechain.org](https://probechain.org) |
| Chain ID Registry | [ethereum-lists/chains #8136](https://github.com/ethereum-lists/chains/pull/8136) |

## License

GPL v3 / LGPL v3. See [COPYING](COPYING) and [COPYING.LESSER](COPYING.LESSER).
