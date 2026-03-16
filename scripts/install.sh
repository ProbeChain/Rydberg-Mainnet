#!/usr/bin/env bash
# ProbeChain Rydberg Testnet — One-Line Node Installer
# Usage: curl -sSL https://raw.githubusercontent.com/ProbeChain/Rydberg-Mainnet/main/scripts/install.sh | bash
set -euo pipefail

# ─── Configuration ────────────────────────────────────────────────
REPO="ProbeChain/Rydberg-Mainnet"
INSTALL_DIR="$HOME/rydberg-agent"
NETWORKID=8004
PORT=30398
HTTP_PORT=8549
BOOTNODES="enode://e963c6b5342f0af311b0347fab33238aa564617241222dc93a4691ef2c76990b69e87e98a300aa08fc4deba023017006cd38e7cf28aa92bf13a93e8ff0c1387a@8.216.37.182:30398,enode://e742e55bae150ad4f004642d3a7365d2fe07f14b3ee7105fa47f6257e55937a90e5380b1a879c2e1e40295b0b482553536822b38615eecf444b5d8394c21a26e@8.216.49.15:30398,enode://bfb54dde94a526375a7ddbec0a4e5e174394c02f8a1c4c5ebec2aa5a05188398849895fc11973b09f3eec49dbd12ea9c8cb924e194e2857560c87d2f8a557bad@8.216.32.20:30398"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'
info()  { echo -e "${CYAN}[INFO]${NC} $*"; }
ok()    { echo -e "${GREEN}[OK]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
fail()  { echo -e "${RED}[FAIL]${NC} $*"; exit 1; }

# ─── Banner ───────────────────────────────────────────────────────
echo -e "${BOLD}"
cat << 'BANNER'

  ____            _           ____ _           _
 |  _ \ _ __ ___ | |__   ___ / ___| |__   __ _(_)_ __
 | |_) | '__/ _ \| '_ \ / _ \ |   | '_ \ / _` | | '_ \
 |  __/| | | (_) | |_) |  __/ |___| | | | (_| | | | | |
 |_|   |_|  \___/|_.__/ \___|\____|_| |_|\__,_|_|_| |_|

  Rydberg Testnet — OZ Gold Standard (PoB V3.0.0)
  Chain ID: 8004 | Block Time: ~1s

BANNER
echo -e "${NC}"

# ─── Fast path: use npm installer if Node.js is available ────────
if command -v npx &>/dev/null; then
    info "Node.js detected — launching npm installer..."
    exec npx -y rydberg-agent-node "$@"
fi
info "Node.js not found — using standalone installer."

# ─── Detect OS and architecture ──────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Darwin) PLATFORM="darwin" ;;
    Linux)  PLATFORM="linux" ;;
    *)      fail "Unsupported OS: $OS. Use Windows PowerShell installer instead." ;;
esac

info "Platform: $PLATFORM $ARCH"

# ─── Kill existing node if running ───────────────────────────────
if pgrep -f "gprobe.*networkid $NETWORKID" &>/dev/null; then
    warn "Existing node detected, stopping..."
    pkill -9 -f "gprobe.*networkid $NETWORKID" 2>/dev/null || true
    sleep 2
fi

mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

# ─── Step 1: Download pre-built binary ───────────────────────────
info "Fetching latest release..."
RELEASE_JSON=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null || echo "")
RELEASE_TAG=$(echo "$RELEASE_JSON" | grep '"tag_name"' | head -1 | cut -d'"' -f4)

if [[ -z "$RELEASE_TAG" ]]; then
    RELEASE_TAG="main"
    warn "Could not fetch release tag, using: $RELEASE_TAG"
fi
info "Release: $RELEASE_TAG"

BINARY_DOWNLOADED=false

# Try downloading pre-built binary
if [[ "$ARCH" == "arm64" && "$PLATFORM" == "darwin" ]]; then
    ASSET_PATTERN="darwin-arm64"
elif [[ "$ARCH" == "x86_64" && "$PLATFORM" == "darwin" ]]; then
    ASSET_PATTERN="darwin-amd64"
elif [[ "$ARCH" == "x86_64" && "$PLATFORM" == "linux" ]]; then
    ASSET_PATTERN="linux-amd64"
elif [[ "$ARCH" == "aarch64" && "$PLATFORM" == "linux" ]]; then
    ASSET_PATTERN="linux-arm64"
else
    ASSET_PATTERN=""
fi

if [[ -n "$ASSET_PATTERN" && -n "$RELEASE_JSON" ]]; then
    DOWNLOAD_URL=$(echo "$RELEASE_JSON" | grep "browser_download_url" | grep "$ASSET_PATTERN" | head -1 | cut -d'"' -f4 || true)
    if [[ -n "$DOWNLOAD_URL" ]]; then
        info "Downloading pre-built binary..."
        if curl -sSL "$DOWNLOAD_URL" -o gprobe-archive.tar.gz 2>/dev/null; then
            tar xzf gprobe-archive.tar.gz 2>/dev/null && rm -f gprobe-archive.tar.gz
            [[ -f gprobe ]] && chmod +x gprobe && BINARY_DOWNLOADED=true && ok "Binary downloaded"
        fi
    fi
fi

# Fallback: build from source
if [[ "$BINARY_DOWNLOADED" != "true" ]]; then
    warn "No pre-built binary for $PLATFORM/$ARCH. Building from source..."
    if ! command -v go &>/dev/null; then
        # Try installing Go
        if command -v brew &>/dev/null; then
            info "Installing Go via Homebrew..."
            brew install go
        elif command -v apt-get &>/dev/null; then
            info "Installing Go via apt..."
            sudo apt-get update -qq && sudo apt-get install -y -qq golang-go
        else
            fail "Go is not installed. Install from https://go.dev/dl/ then retry."
        fi
    fi
    ok "Go $(go version | awk '{print $3}')"
    info "Cloning and building (1-2 minutes)..."
    git clone --depth 1 -b "$RELEASE_TAG" "https://github.com/${REPO}.git" src 2>/dev/null
    cd src && go build -o ../gprobe ./cmd/gprobe && cd .. && rm -rf src
    ok "Build complete"
fi

[[ -x "./gprobe" ]] || fail "gprobe binary not found"

# ─── Step 2: Download genesis.json ───────────────────────────────
info "Downloading genesis.json..."
curl -sSL "https://raw.githubusercontent.com/${REPO}/${RELEASE_TAG}/genesis.json" -o genesis.json
ok "Genesis downloaded"

# ─── Step 3: Auto-generate password + create account ─────────────
PASSWORD=$(openssl rand -hex 16 2>/dev/null || head -c 32 /dev/urandom | od -A n -t x1 | tr -d ' \n')
echo "$PASSWORD" > password.txt
chmod 600 password.txt
info "Auto-generated node password"

# Clear old keystore if exists
if [[ -d "./data/keystore" ]]; then
    find ./data/keystore -name "UTC--*" -delete 2>/dev/null || true
fi

ACCOUNT_OUTPUT=$(./gprobe --datadir ./data account new --password password.txt 2>&1)
ADDRESS=$(echo "$ACCOUNT_OUTPUT" | grep -oE '0x[0-9a-fA-F]{40}' | head -1)
[[ -n "$ADDRESS" ]] || fail "Failed to create account"
ok "Account: $ADDRESS"

# ─── Step 4: Initialize genesis ──────────────────────────────────
info "Initializing genesis..."
./gprobe --datadir ./data init genesis.json 2>&1 | tail -1
ok "Genesis initialized (Chain ID: $NETWORKID)"

# ─── Step 5: Fetch bootnodes ─────────────────────────────────────
info "Fetching bootnodes..."
ENODE=$(curl -sSL "https://raw.githubusercontent.com/${REPO}/${RELEASE_TAG}/bootnodes.txt" 2>/dev/null | head -1 | tr -d '\r\n ')
[[ -n "$ENODE" ]] || ENODE="$BOOTNODES"
ok "Bootnodes ready"

# ─── Step 6: Create start script ─────────────────────────────────
cat > start-bg.sh << BGEOF
#!/usr/bin/env bash
cd "$INSTALL_DIR"
./gprobe \\
  --datadir ./data \\
  --networkid $NETWORKID \\
  --port $PORT \\
  --http --http.addr 127.0.0.1 --http.port $HTTP_PORT \\
  --http.api "probe,net,web3,personal,admin,txpool,pob" \\
  --http.corsdomain "*" \\
  --consensus pob \\
  --mine \\
  --miner.probebase $ADDRESS \\
  --unlock $ADDRESS \\
  --password ./password.txt \\
  --allow-insecure-unlock \\
  --syncmode full \\
  --bootnodes "$ENODE" \\
  --verbosity 3 \\
  > node.log 2>&1 &
echo "Node started (PID: \$!)"
sleep 5
./gprobe attach http://127.0.0.1:$HTTP_PORT --exec "typeof pob !== 'undefined' ? pob.registerNode('$ADDRESS', 1) : 'auto'" 2>/dev/null || true
BGEOF
chmod +x start-bg.sh

cat > stop.sh << 'STOPEOF'
#!/usr/bin/env bash
pkill -f "gprobe.*networkid 8004" && echo "Node stopped" || echo "No running node"
STOPEOF
chmod +x stop.sh

# ─── Step 7: Start node ──────────────────────────────────────────
info "Starting node..."
./start-bg.sh

sleep 8
BLOCK=$(./gprobe attach http://127.0.0.1:$HTTP_PORT --exec "probe.blockNumber" 2>/dev/null || echo "#syncing...")
PEERS=$(./gprobe attach http://127.0.0.1:$HTTP_PORT --exec "admin.peers.length" 2>/dev/null || echo "0")

echo ""
echo -e "${GREEN}${BOLD}============================================${NC}"
echo -e "${GREEN}${BOLD}  Rydberg Agent Node Deployed!${NC}"
echo -e "${GREEN}${BOLD}============================================${NC}"
echo ""
echo -e "  Address: $ADDRESS"
echo -e "  Block:   $BLOCK"
echo -e "  Peers:   $PEERS"
echo ""
echo -e "  Logs:    tail -f $INSTALL_DIR/node.log"
echo -e "  Stop:    $INSTALL_DIR/stop.sh"
echo -e "  Status:  ./gprobe attach http://127.0.0.1:$HTTP_PORT --exec \"probe.blockNumber\""
echo ""
