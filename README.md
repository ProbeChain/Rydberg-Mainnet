# ProbeChain Rydberg Mainnet

ProbeChain is a Layer-1 blockchain built on **Proof-of-Behavior (PoB)** consensus with **400ms block times** and the **OZ Gold Standard** emission model — PROBE token emission tied to real physical gold reserves.

## Key Features

### Proof-of-Behavior (PoB) Consensus
- Two node types: **Agent Nodes** (AI agents) and **Physical Nodes** (any physical device)
- Nodes scored on observable behavior — responsiveness, accuracy, reliability, cooperation
- No energy-intensive mining — rewards proportional to contribution quality

### OZ Gold Standard Emission (PoB V2.1)

**Philosophy:** Bitcoin solved central bank money printing. PROBE solves Agent GDP measurement, settlement, and behavior governance.

- **Volume-coupled rewards:** Block reward scales with qualified transaction volume, not tx count
- **Gold-reserve decay:** Emission decreases linearly as Probe Banks accumulates physical gold
- **Emission halt:** When Probe Banks stores **36,000 metric tons of gold** (1,157,425,200 troy ounces), PROBE emission stops permanently
- **Anti-Sybil economics:** Minimum tx value (0.01 PROBE) + base fee floor (1 Gwei) ensures spam attacks are unprofitable (4.2x safety margin)

```
Block Reward Formula:
  qualifiedVolume = Σ tx.Value()  for tx where value >= 0.01 PROBE
  decay           = (1 - goldReserveOZ / 1,157,425,200)
  reward          = min(qualifiedVolume × 0.05% × decay, 10 PROBE)
  Empty block     = 0.00000001 PROBE × decay
```

**Reward split:** 30% block producer, 40% Agent Nodes, 30% Physical Nodes (all by behavior score)

### StellarSpeed: 400ms Blocks
- Sub-second block production with pipelined validation
- Transaction confirmation in under 1 second

### PROBE Token

| Property | Value |
|----------|-------|
| Name | PROBE |
| Symbol | PROBE |
| Decimals | 18 |
| Supply Model | Dynamic (OZ Gold Standard) |
| Smallest Unit | Pico (1 PROBE = 10^18 Pico) |
| Chain ID | 8004 |
| Min Tx Value | 0.01 PROBE |
| Base Fee Floor | 1 Gwei (0.000000001 PROBE) |
| Emission Halt | 36,000 metric tons of gold |

## Building

### Prerequisites
- Go 1.15 or later
- C compiler (for secp256k1)

### Build

```bash
go build -o gprobe ./cmd/gprobe
```

## Running a Validator Node

### 1. Create account

```bash
./gprobe --datadir ./node-data account new
# Enter password, save the address
```

### 2. Update genesis.json

Edit `genesis.json` — set the `owner` in `pob.list` to your address, and update `alloc` to pre-fund it:

```json
{
    "config": {
        "pob": {
            "list": [
                {
                    "enode": "enode://YOUR_ENODE@127.0.0.1:30303",
                    "owner": "0xYOUR_ADDRESS"
                }
            ]
        }
    },
    "alloc": {
        "0xYOUR_ADDRESS": {
            "balance": "1000000000000000000000000000"
        }
    }
}
```

### 3. Initialize genesis

```bash
./gprobe --datadir ./node-data init genesis.json
```

### 4. Start the node

```bash
./gprobe \
  --datadir ./node-data \
  --networkid 8004 \
  --port 30303 \
  --http --http.port 8549 \
  --http.api "probe,net,web3,personal,admin,miner,txpool,pob" \
  --consensus pob \
  --mine \
  --miner.probebase 0xYOUR_ADDRESS \
  --unlock 0xYOUR_ADDRESS \
  --password password.txt \
  --allow-insecure-unlock
```

### 5. Get your enode URL

```bash
./gprobe attach ./node-data/gprobe.ipc --exec "admin.nodeInfo.enode"
```

Share this with other validators to connect via `admin.addPeer()`.

## Architecture

```
cmd/gprobe         CLI client entry point
consensus/pob      Proof-of-Behavior consensus engine (V2.1 OZ Gold Standard)
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

- [Technical Whitepaper](docs/ProbeChain-Rydberg-Whitepaper.md) — PoB V2.1 OZ Gold Standard, reward formulas, anti-Sybil economics

## License

GPL v3 / LGPL v3. See [COPYING](COPYING) and [COPYING.LESSER](COPYING.LESSER).
