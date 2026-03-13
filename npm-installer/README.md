# rydberg-agent-node

One-command installer for ProbeChain Rydberg testnet **Agent nodes**.

```bash
npx rydberg-agent-node
```

## What it does

1. Detects your OS and architecture
2. Downloads a pre-built binary (macOS arm64) or builds from source (other platforms)
3. Verifies integrity via SHA256 + optional GPG signature
4. Creates a new account with your password
5. Initializes the Rydberg testnet genesis block (Chain ID 8004)
6. Starts the node and auto-registers as an **Agent node** (PoB NodeType=1)

Agent nodes participate in block sync, serve RPC requests, and earn rewards from the 40% Agent reward pool.

## Requirements

- **Node.js** >= 14.0.0
- **macOS arm64**: No additional requirements (pre-built binary)
- **macOS x86_64 / Linux**: `git` and `go` (1.19+) for source build
- **Windows**: Install WSL2 first (`wsl --install -d Ubuntu`)

## Commands

```bash
npx rydberg-agent-node           # Install and start
npx rydberg-agent-node status    # Show node status (block, peers, balance)
npx rydberg-agent-node start     # Start the node
npx rydberg-agent-node stop      # Stop the node
npx rydberg-agent-node logs      # Show recent logs
```

## Installation directory

All files are installed to `~/rydberg-agent/`:

```
~/rydberg-agent/
├── gprobe            # Blockchain client binary
├── data/             # Chain data and keystore
├── genesis.json      # Genesis configuration
├── password.txt      # Account password (mode 0600)
├── start-bg.sh       # Startup script
├── stop.sh           # Stop script
└── node.log          # Node logs
```

## Network details

| Parameter | Value |
|-----------|-------|
| Chain ID | 8004 |
| Consensus | PoB V2.1 (OZ Gold Standard) |
| Block time | 400ms |
| P2P port | 30398 |
| HTTP RPC | 127.0.0.1:8549 |
| RPC APIs | probe, net, web3, pob, txpool |

## Zero dependencies

This package uses only Node.js built-in modules (`https`, `fs`, `path`, `readline`, `child_process`, `crypto`). No npm dependencies.

## License

GPL-3.0
