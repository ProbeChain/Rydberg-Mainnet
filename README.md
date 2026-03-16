# ProbeChain Rydberg Testnet

A Layer-1 blockchain for the agent economy. **Proof-of-Behavior** consensus. **Gold-anchored** emission. **One line to join.**

```bash
npx -y rydberg-agent-node
```

That's it. No staking. No hardware. No dependencies. Your node starts earning rewards in under 60 seconds.

---

## Network Info

| Field | Value |
|-------|-------|
| Chain ID | `8004` |
| RPC | `https://proscan.pro/chain/rydberg-rpc` |
| Explorer | [proscan.pro](https://proscan.pro) |
| Token | PROBE (18 decimals) |
| Block time | ~15 seconds |
| Consensus | Proof-of-Behavior V2.1 (PoB) |
| EVM | London-compatible (EIP-1559 enabled) |
| Client | Gprobe v2.0.0 (go-ethereum fork) |

### Add to MetaMask / Wallet

| Field | Value |
|-------|-------|
| Network Name | ProbeChain Rydberg Testnet |
| RPC URL | `https://proscan.pro/chain/rydberg-rpc` |
| Chain ID | `8004` |
| Currency Symbol | `PROBE` |
| Block Explorer | `https://proscan.pro` |

---

## Public Testnet Status

The Rydberg testnet launched for community public testing on **March 16, 2026**. Current validation results:

| Metric | Result |
|--------|--------|
| Functional tests | **196/200 passed (98%)** |
| Stress test transactions | **252,926 tx (zero crashes)** |
| Peak TPS | **1,560 tx/s** (3-node concurrent) |
| Sustained TPS | **433 tx/s** |
| Smart contracts deployed | 7 (ERC20, NFT, DEX, Prediction Market, etc.) |
| Node crashes during stress | **0** |

Full test report: [`miner/stress/probechain-tests/logs/FULL_TEST_REPORT.md`](miner/stress/probechain-tests/logs/FULL_TEST_REPORT.md)

---

## Ecosystem

100+ applications are testing on the Rydberg testnet:

| Product | Domain | Type |
|---------|--------|------|
| Pro.Gold | [pro.gold](https://pro.gold) | Gold exchange |
| AIoAI | [aioai.ai](https://aioai.ai) | AI Agent social platform |
| oz.money | [oz.money](https://oz.money) | Financial services |
| ProSwap | [proswap.pro](https://proswap.pro) | DEX (Uniswap V2 fork) |
| Ashares | [ashares.ai](https://ashares.ai) | AI assets |
| ProPay | — | Multi-chain wallet |
| Probe.Builders | [probe.builders](https://probe.builders) | Developer platform |
| ProScan | [proscan.pro](https://proscan.pro) | Block explorer |
| Probebanks | [probebanks.com](https://probebanks.com) | Banking services |
| UniClaw | [uniclaw.cloud](https://uniclaw.cloud) | A2A Agent collaboration |
| tidal.financial | [tidal.financial](https://tidal.financial) | Financial services |
| And more... | | |

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
| V | Qualified transaction volume in block (value ≥ 0.01 PROBE) | — |
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

### Gold Reserve Decay

PROBE emission decays as [Probe Banks](https://probebanks.com) accumulates physical gold reserves:

```
D = ((G_target − G_current) / G_target)^n
```

| Parameter | Value | Meaning |
|-----------|-------|---------|
| G_target | 1,157,425,200 oz | 36,000 metric tons of gold |
| G_current | On-chain oracle | Current Probe Banks reserves |
| n | 1 | Linear decay |

When reserves reach target, D = 0 and emission stops. Transaction fees alone sustain the network.

---

## Quick Start

### Option 1: NPX (Recommended)

```bash
npx -y rydberg-agent-node
```

### Option 2: Shell Script

```bash
# macOS / Linux
curl -sSL https://raw.githubusercontent.com/ProbeChain/Rydberg-Mainnet/main/scripts/install.sh | bash

# Windows PowerShell
irm https://raw.githubusercontent.com/ProbeChain/Rydberg-Mainnet/main/scripts/install.ps1 | iex
```

### Option 3: Build from Source

```bash
git clone https://github.com/ProbeChain/Rydberg-Mainnet.git
cd Rydberg-Mainnet
go build -o gprobe ./cmd/gprobe

# Initialize
./gprobe --datadir ./data init genesis.json

# Create account
./gprobe --datadir ./data account new

# Start node
./gprobe --datadir ./data \
  --networkid 8004 \
  --port 30398 \
  --http --http.addr 0.0.0.0 --http.port 8549 \
  --http.api probe,net,web3,pob,txpool \
  --consensus pob \
  --syncmode full \
  --bootnodes "enode://e963c6b5342f0af311b0347fab33238aa564617241222dc93a4691ef2c76990b69e87e98a300aa08fc4deba023017006cd38e7cf28aa92bf13a93e8ff0c1387a@8.216.37.182:30398"
```

## Node Operations

```bash
# Check status
~/rydberg-agent/gprobe attach ~/rydberg-agent/gprobe.ipc \
  --exec "JSON.stringify({block: probe.blockNumber, peers: net.peerCount})"

# Check balance
~/rydberg-agent/gprobe attach ~/rydberg-agent/gprobe.ipc \
  --exec "web3.fromWei(probe.getBalance(probe.coinbase), 'probeer')"

# View logs
tail -f ~/rydberg-agent/node.log

# Stop / Restart
kill $(pgrep -f "gprobe.*8004")
bash ~/rydberg-agent/start-bg.sh
```

---

## Architecture

```
cmd/gprobe              CLI entry point — urfave/cli
consensus/pob           Proof-of-Behavior V2.1 consensus engine
  └── pob.go            PoB block sealing, ACK-based finality
core/                   Blockchain core
  ├── blockchain.go     Chain management
  ├── tx_pool.go        Transaction mempool
  ├── state_processor   Block execution
  └── state_transition  EVM state changes
miner/                  Block production
  ├── worker.go         ACK monitor + consensus loop
  └── stress/           Test suite (200 tests + stress scripts)
probe/                  Protocol handlers, sync, peer management
p2p/                    Devp2p networking, node discovery (v4/v5)
common/
  ├── bech32/           Bech32 address encoding (pro1...)
  └── types.go          Address: both 0x-hex and pro1-bech32
internal/jsre/deps/
  └── web3.js           Extended with bech32 address support
crypto/                 secp256k1, BLS12-381, Dilithium (post-quantum ready)
params/                 Chain config and genesis parameters
npm-installer/          `npx -y rydberg-agent-node` one-line installer
```

## Smart Contract Support

ProbeChain is fully EVM-compatible with the following considerations:

| Feature | Status |
|---------|--------|
| Solidity 0.4.x — 0.8.x | ✅ Supported |
| EIP-1559 (Dynamic fees) | ✅ Enabled |
| ERC-20 tokens | ✅ Tested |
| ERC-721 NFTs | ✅ Tested (ERC-8004 Agent Identity) |
| Uniswap V2 AMM | ✅ Tested (MiniSwap) |
| PUSH0 opcode (Shanghai) | ❌ Not supported — use `--evm-version london` |

**Important**: When compiling Solidity ≥ 0.8.20, use:
```bash
solc --optimize --evm-version london YourContract.sol
```

## Address Format

ProbeChain supports **two address formats** interchangeably:

| Format | Example | Usage |
|--------|---------|-------|
| Bech32 | `pro1f03xdemrfd0u5ghvzv95d2ma7s3wqg2rql9jrg` | Display, user-facing |
| Hex | `0x4be266e7634b5fca22ec130b46ab7df422e02143` | Contract calls, raw tx |

Both formats work in all APIs: `sendTransaction`, `getBalance`, `contract.call`, etc.

---

## Bootnodes

```
enode://e963c6b5342f0af311b0347fab33238aa564617241222dc93a4691ef2c76990b69e87e98a300aa08fc4deba023017006cd38e7cf28aa92bf13a93e8ff0c1387a@8.216.37.182:30398
enode://e742e55bae150ad4f004642d3a7365d2fe07f14b3ee7105fa47f6257e55937a90e5380b1a879c2e1e40295b0b482553536822b38615eecf444b5d8394c21a26e@8.216.49.15:30398
enode://bfb54dde94a526375a7ddbec0a4e5e174394c02f8a1c4c5ebec2aa5a05188398849895fc11973b09f3eec49dbd12ea9c8cb924e194e2857560c87d2f8a557bad@8.216.32.20:30398
```

## Documentation

- **[Technical Whitepaper](docs/ProbeChain-Rydberg-Whitepaper.md)** — Consensus formulas, emission model, scoring algorithms
- **[Genesis Config](genesis.json)** — All on-chain parameters
- **[Test Report](miner/stress/probechain-tests/logs/FULL_TEST_REPORT.md)** — 200 functional tests + 252K tx stress test results
- **[Smart Contracts](miner/stress/probechain-tests/contracts/solidity/)** — Tested contract examples (ERC20, NFT, DEX, Prediction Market)

## Links

| Resource | URL |
|----------|-----|
| Website | [probechain.org](https://probechain.org) |
| Public RPC | `https://proscan.pro/chain/rydberg-rpc` |
| Block Explorer | [proscan.pro](https://proscan.pro) |
| AI Agent Platform | [aioai.ai](https://aioai.ai) |
| Gold Exchange | [pro.gold](https://pro.gold) |
| NPM Package | [npmjs.com/package/rydberg-agent-node](https://www.npmjs.com/package/rydberg-agent-node) |

## Contributing

1. Fork this repo
2. Create a feature branch: `git checkout -b feat/your-feature`
3. Commit changes: `git commit -m "feat: add your feature"`
4. Push: `git push origin feat/your-feature`
5. Open a Pull Request

Please ensure `go build ./cmd/gprobe` passes before submitting.

## License

GPL v3 / LGPL v3. See [COPYING](COPYING) and [COPYING.LESSER](COPYING.LESSER).
