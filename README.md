# ProbeChain

ProbeChain is a high-performance blockchain platform built on **Proof-of-Behavior (PoB)** consensus with **400ms block times**, a native **PROBE token**, and the **PROBE Language** — an agent-first smart contract programming language.

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
- Reduced ACK quorum for fast finality
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

### PROBE Language
An agent-first programming language designed for AI agents and smart contracts:

- **Linear type system** — Move-inspired resource safety (assets can't be duplicated or lost)
- **Register-based VM** — 256 registers, 64-bit words, 71 opcodes
- **Native PQC crypto** — Falcon-512, ML-DSA (Dilithium), SLH-DSA (SPHINCS+) verification opcodes
- **Agent primitives** — First-class `agent`, `spawn`, `send`, `recv` constructs
- **BPE-aligned syntax** — ASCII-only tokens optimized for LLM code generation (~70 tokens/task)
- **Bytecode verification** — Safety holds even for buggy compiler output

## Building

### Prerequisites
- Go 1.15 or later
- C compiler (for secp256k1)

### Build the blockchain client

```bash
make gprobe
```

### Build the PROBE Language compiler

```bash
make probec
```

### Build everything

```bash
make all
```

### Run tests

```bash
make test
```

## Running

### Start a node

```bash
./build/bin/gprobe
```

### Compile PROBE Language source

```bash
# Tokenize a .probe file
./build/bin/probec -emit tokens example.probe

# Compile to bytecode (coming soon)
./build/bin/probec -o output.pbc example.probe
```

## Architecture

```
cmd/gprobe           CLI client entry point
cmd/probec           PROBE Language compiler CLI
consensus/pob        Proof-of-Behavior consensus engine
core                 Blockchain core: chain, state, tx pool
probe                Protocol handlers, sync, peer management
p2p                  Devp2p networking, node discovery
miner                Block production
probe-lang           PROBE Language subsystem
  lang/token           Lexical token definitions
  lang/lexer           ASCII-only BPE-aligned tokenizer
  lang/ast             Abstract syntax tree (22 expr, 10 stmt, 9 decl types)
  lang/parser          Recursive descent + Pratt expression parser
  lang/types           Type system with linear type checker
  lang/ir              SSA-form intermediate representation
  lang/codegen         Bytecode generation + Move-inspired verifier
  lang/vm              Register-based virtual machine
  stdlib               Standard library (agent, chain, crypto, math)
  spec/grammar.ebnf    Formal grammar specification
```

## PROBE Language Example

```
agent Echo {
    state {
        count: u64,
    }

    msg handle(data: bytes) -> bytes {
        self.count += 1;
        data
    }
}

resource Token {
    balance: u64,
}

fn transfer(from: &mut Token, to: &mut Token, amount: u64) {
    require(from.balance >= amount, "insufficient balance");
    from.balance -= amount;
    to.balance += amount;
}
```

## VM Opcodes

The PROBE VM provides 71 opcodes across 10 categories:

| Category | Examples |
|----------|----------|
| Arithmetic | `add`, `sub`, `mul`, `div`, `mod`, `neg` |
| Bitwise | `and`, `or`, `xor`, `not`, `shl`, `shr` |
| Comparison | `eq`, `neq`, `lt`, `lte`, `gt`, `gte` |
| Control Flow | `jump`, `jump_if`, `call`, `return`, `halt` |
| Memory | `alloc`, `free`, `load_mem`, `store_mem` |
| Agent | `spawn`, `send`, `recv`, `self` |
| Blockchain | `balance`, `transfer`, `emit`, `block_num` |
| Crypto (PQC) | `sha3`, `shake256`, `falcon512_verify`, `ml_dsa_verify`, `slh_dsa_verify` |
| Resources | `resource_new`, `resource_drop`, `resource_check` |
| Array | `array_new`, `array_get`, `array_set`, `array_len` |

## Development

### Run a single package's tests

```bash
go test ./core/...
go test ./probe-lang/lang/vm/...
go test ./consensus/pob/...
```

### Linting

```bash
make lint
```

Configured linters: `deadcode`, `goconst`, `goimports`, `gosimple`, `govet`, `ineffassign`, `misspell`, `unconvert`, `varcheck`.

### Code generation tools

```bash
make devtools
```

## Documentation

- [Technical Whitepaper](docs/ProbeChain-Rydberg-Whitepaper.md) — Full specification of PoB consensus, reward formulas, Agent GDP model, and anti-sybil mechanisms
- [Documentation](https://doc.probechain.org/) — ProbeChain technical documentation
- [Block Explorer](https://scan.probechain.org/home) — ProbeChain block explorer
- [GitHub](https://github.com/ProbeChain/Rydberg-Mainnet) — Source code

## License

GPL v3 / LGPL v3. See [COPYING](COPYING) and [COPYING.LESSER](COPYING.LESSER).
