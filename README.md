# ProbeChain Rydberg Mainnet

ProbeChain is a Layer-1 blockchain built on **Proof-of-Behavior (PoB)** consensus with **400ms block times** and a dynamic **PROBE token** emission model tied to on-chain economic activity.

## Key Features

### Proof-of-Behavior (PoB) Consensus
- Two node types: **Agent Nodes** (AI agents) and **Physical Nodes** (any physical device)
- Nodes are scored on observable behavior — responsiveness, accuracy, reliability, cooperation
- No energy-intensive mining — rewards proportional to contribution quality
- Anti-sybil system with ERC-8004 agent identity, device fingerprinting, and VM detection

### Transaction-Driven Block Rewards
- Block reward scales with real transaction volume: `reward = min((txCount + 1) * 0.0001 PROBE, 10 PROBE)`
- Reward split: 30% block producer, 40% Agent Nodes, 30% Physical Nodes
- **No fixed supply** — emission continues until on-chain Agent GDP reaches $150 trillion (2026 global human GDP equivalent)

### StellarSpeed: 400ms Blocks
- Sub-second block production with pipelined validation
- Transaction confirmation in under 1 second

### PROBE Token
| Property | Value |
|----------|-------|
| Name | PROBE |
| Symbol | PROBE |
| Decimals | 18 |
| Supply Model | Dynamic (Agent GDP-driven emission) |
| Smallest Unit | Pico (1 PROBE = 10^18 Pico) |
| Chain ID | 8004 |

## Building

### Prerequisites
- Go 1.15 or later
- C compiler (for secp256k1)

### Build

```bash
go build -o gprobe ./cmd/gprobe
```

## Running

### Initialize genesis

```bash
./gprobe --datadir ./data init genesis.json
```

### Start a node

```bash
./gprobe --datadir ./data --networkid 8004 --http --http.api "probe,net,web3,personal,admin,miner,txpool,pob" --consensus pob --mine --miner.probebase <YOUR_ADDRESS> --unlock <YOUR_ADDRESS> --password password.txt --allow-insecure-unlock
```

## Architecture

```
cmd/gprobe         CLI client entry point
cmd/utils          Shared CLI flags and configuration
consensus/pob      Proof-of-Behavior consensus engine
core               Blockchain core: chain, state, tx pool, VM
probe              Protocol handlers, sync, peer management
p2p                Devp2p networking, node discovery
crypto             Cryptography (secp256k1, Dilithium PQC)
miner              Block production (StellarSpeed 400ms)
params             Chain configuration and genesis parameters
accounts           Account management and keystore
node               Node lifecycle and service management
rpc                JSON-RPC server
trie               Merkle Patricia Trie
probedb            Database abstraction (LevelDB)
```

## Documentation

- [Technical Whitepaper](docs/ProbeChain-Rydberg-Whitepaper.md) — PoB consensus, reward formulas, Agent GDP model, anti-sybil mechanisms
- [GitHub](https://github.com/ProbeChain/Rydberg-Mainnet)

## License

GPL v3 / LGPL v3. See [COPYING](COPYING) and [COPYING.LESSER](COPYING.LESSER).
