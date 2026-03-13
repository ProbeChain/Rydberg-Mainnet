#!/usr/bin/env bash
# ProbeChain Rydberg Testnet — One-Line Validator Installer
# Usage: curl -sSL https://raw.githubusercontent.com/ProbeChain/Rydberg-Mainnet/main/scripts/install.sh | bash
#   or:  curl -sSL https://github.com/ProbeChain/Rydberg-Mainnet/raw/main/scripts/install.sh | bash
set -euo pipefail

# ─── Configuration ────────────────────────────────────────────────
REPO="ProbeChain/Rydberg-Mainnet"
BRANCH="main"
INSTALL_DIR="$HOME/rydberg-node"
NETWORKID=8004
PORT=30398
HTTP_PORT=8549
GENESIS_NODE_IP="bore.pub"
GENESIS_NODE_PORT=9208
GENESIS_NODE_ENODE="enode://c56b6a7949fa9f6cf6e809863223fa9a444440a8f7fd4776ff5437f4c0db8d5775f7c0d3bfa0e6270242aa3811b776c9ef19d12c47a0f6e76f25b430a99071e9@${GENESIS_NODE_IP}:${GENESIS_NODE_PORT}"
# ──────────────────────────────────────────────────────────────────

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()  { echo -e "${CYAN}[INFO]${NC} $*"; }
ok()    { echo -e "${GREEN}[OK]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
fail()  { echo -e "${RED}[FAIL]${NC} $*"; exit 1; }

# ─── GitHub API helper (avoids raw.githubusercontent.com DNS issues) ──
# Fetches a file from the repo via GitHub Contents API with raw accept header.
github_get_file() {
    local filepath="$1"
    local ref="${2:-$BRANCH}"
    curl -sSL \
        -H "Accept: application/vnd.github.raw" \
        -H "User-Agent: rydberg-installer/1.0" \
        "https://api.github.com/repos/${REPO}/contents/${filepath}?ref=${ref}"
}

# ─── Banner ───────────────────────────────────────────────────────
echo -e "${BOLD}"
cat << 'BANNER'

  ____            _           ____ _           _
 |  _ \ _ __ ___ | |__   ___ / ___| |__   __ _(_)_ __
 | |_) | '__/ _ \| '_ \ / _ \ |   | '_ \ / _` | | '_ \
 |  __/| | | (_) | |_) |  __/ |___| | | | (_| | | | | |
 |_|   |_|  \___/|_.__/ \___|\____|_| |_|\__,_|_|_| |_|

  Rydberg Testnet — OZ Gold Standard (PoB V2.1)
  Chain ID: 8004 | Block Time: 400ms

BANNER
echo -e "${NC}"

# ─── Fast path: use npm installer if Node.js is available ────────
if command -v npx &>/dev/null; then
    info "Node.js detected — launching npm installer for best experience..."
    exec npx rydberg-agent-node "$@"
    # exec replaces this process; below is unreachable
fi
info "Node.js not found — using standalone installer (no dependencies needed)."

# ─── Pre-checks ──────────────────────────────────────────────────
[[ "$(uname)" == "Darwin" ]] || fail "This installer is for macOS only."

ARCH="$(uname -m)"
[[ "$ARCH" == "arm64" || "$ARCH" == "x86_64" ]] || fail "Unsupported architecture: $ARCH"

if [[ -d "$INSTALL_DIR" ]]; then
    warn "Directory $INSTALL_DIR already exists."
    read -rp "Overwrite and reinstall? [y/N] " ans
    [[ "$ans" =~ ^[Yy]$ ]] || { info "Aborted."; exit 0; }
    rm -rf "$INSTALL_DIR"
fi

mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

# ─── Step 1: Get gprobe binary ───────────────────────────────────
info "Detecting environment..."

# Fetch latest release tag
RELEASE_TAG=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | head -1 | cut -d'"' -f4)

if [[ -z "$RELEASE_TAG" ]]; then
    warn "Could not fetch release tag, using branch: $BRANCH"
    RELEASE_TAG="$BRANCH"
fi

info "Release: $RELEASE_TAG"

get_binary_from_release() {
    info "Downloading pre-built gprobe binary..."
    local RELEASE_URL
    RELEASE_URL=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep "browser_download_url.*darwin.*${ARCH}" \
        | head -1 \
        | cut -d '"' -f 4 2>/dev/null) || true

    if [[ -n "$RELEASE_URL" ]]; then
        curl -sSL "$RELEASE_URL" | tar xz
        chmod +x gprobe
        ok "Binary downloaded from GitHub Release."
        return 0
    fi
    return 1
}

build_from_source() {
    info "Building gprobe from source..."
    if ! command -v go &>/dev/null; then
        if command -v brew &>/dev/null; then
            info "Installing Go via Homebrew..."
            brew install go
        else
            fail "Go is not installed and Homebrew not found.\nInstall Go: https://go.dev/dl/ or install Homebrew first."
        fi
    fi
    ok "Go $(go version | awk '{print $3}') found."

    info "Cloning repository @ ${RELEASE_TAG}..."
    git clone --depth 1 -b "$RELEASE_TAG" "https://github.com/${REPO}.git" src
    cd src
    info "Compiling gprobe (this may take 1-2 minutes)..."
    go build -o ../gprobe ./cmd/gprobe
    cd ..
    rm -rf src
    ok "Build complete."
}

# Try pre-built binary first, fall back to source build
if ! get_binary_from_release; then
    warn "No pre-built binary found for ${ARCH}. Building from source..."
    build_from_source
fi

[[ -x "$INSTALL_DIR/gprobe" ]] || fail "gprobe binary not found after installation."
ok "gprobe ready: $INSTALL_DIR/gprobe"

# ─── Step 2: Download genesis.json (via GitHub API) ──────────────
info "Downloading genesis.json..."
github_get_file "genesis.json" "$RELEASE_TAG" > genesis.json
ok "Genesis config downloaded."

# ─── Step 3: Create account ──────────────────────────────────────
info "Creating validator account..."
echo ""
echo -e "${YELLOW}Set a password for your validator account.${NC}"
echo -e "${YELLOW}Remember this password — you'll need it to start the node.${NC}"
echo ""

read -rsp "Enter password: " PASSWORD
echo ""
read -rsp "Confirm password: " PASSWORD2
echo ""

[[ "$PASSWORD" == "$PASSWORD2" ]] || fail "Passwords do not match."
[[ ${#PASSWORD} -ge 6 ]] || fail "Password must be at least 6 characters."

echo "$PASSWORD" > password.txt
chmod 600 password.txt

# Create account and capture address
ACCOUNT_OUTPUT=$(./gprobe --datadir ./data account new --password password.txt 2>&1)
ADDRESS=$(echo "$ACCOUNT_OUTPUT" | grep -oE '0x[0-9a-fA-F]{40}' | head -1)

if [[ -z "$ADDRESS" ]]; then
    fail "Failed to create account. Output:\n$ACCOUNT_OUTPUT"
fi
ok "Account created: $ADDRESS"

# ─── Step 4: Initialize genesis ──────────────────────────────────
info "Initializing genesis block..."
./gprobe --datadir ./data init genesis.json 2>&1 | tail -1
ok "Genesis initialized (Chain ID: $NETWORKID)."

# ─── Step 5: Fetch bootnode (via GitHub API) ─────────────────────
info "Fetching bootnode..."
ENODE=$(github_get_file "bootnodes.txt" "$RELEASE_TAG" | head -1 | tr -d '\r\n ')
if [[ -z "$ENODE" || ! "$ENODE" =~ ^enode:// ]]; then
    warn "Could not fetch bootnode from API, using default."
    ENODE="$GENESIS_NODE_ENODE"
fi
ok "Bootnode: ${ENODE:0:40}..."

# ─── Step 6: Create start scripts ────────────────────────────────
cat > start.sh << STARTEOF
#!/usr/bin/env bash
cd "$INSTALL_DIR"
echo "Starting ProbeChain Rydberg validator..."
echo "Address: $ADDRESS"
echo "HTTP RPC: http://127.0.0.1:$HTTP_PORT"
echo "Press Ctrl+C to stop."
echo ""
./gprobe \\
  --datadir ./data \\
  --networkid $NETWORKID \\
  --port $PORT \\
  --http --http.addr 127.0.0.1 --http.port $HTTP_PORT \\
  --http.api "probe,net,web3,personal,admin,miner,txpool,pob" \\
  --http.corsdomain "*" \\
  --consensus pob \\
  --mine \\
  --miner.probebase $ADDRESS \\
  --unlock $ADDRESS \\
  --password ./password.txt \\
  --allow-insecure-unlock \\
  --bootnodes "$ENODE" \\
  --verbosity 3
STARTEOF
chmod +x start.sh

# Background start script with auto-connect
cat > start-bg.sh << BGEOF
#!/usr/bin/env bash
cd "$INSTALL_DIR"
./gprobe \\
  --datadir ./data \\
  --networkid $NETWORKID \\
  --port $PORT \\
  --http --http.addr 127.0.0.1 --http.port $HTTP_PORT \\
  --http.api "probe,net,web3,personal,admin,miner,txpool,pob" \\
  --http.corsdomain "*" \\
  --consensus pob \\
  --mine \\
  --miner.probebase $ADDRESS \\
  --unlock $ADDRESS \\
  --password ./password.txt \\
  --allow-insecure-unlock \\
  --ipcpath /tmp/gprobe-rydberg.ipc \\
  --bootnodes "$ENODE" \\
  --verbosity 3 \\
  > node.log 2>&1 &
NODE_PID=\$!
echo ""
echo "============================================"
echo "  ProbeChain Rydberg Node Started!"
echo "============================================"
echo ""
echo "  PID:     \$NODE_PID"
echo "  Address: $ADDRESS"
echo "  Logs:    tail -f $INSTALL_DIR/node.log"
echo "  Console: $INSTALL_DIR/gprobe attach /tmp/gprobe-rydberg.ipc"
echo ""
echo "  Check block:   ./gprobe attach /tmp/gprobe-rydberg.ipc --exec \"probe.blockNumber\""
echo "  Check balance: ./gprobe attach /tmp/gprobe-rydberg.ipc --exec \"web3.fromWei(probe.getBalance('$ADDRESS'), 'probeer')\""
echo "  Check peers:   ./gprobe attach /tmp/gprobe-rydberg.ipc --exec \"admin.peers.length\""
echo "  Stop node:     kill \$NODE_PID"
echo ""
# Auto-connect to bootnode after startup
sleep 3
./gprobe attach /tmp/gprobe-rydberg.ipc --exec "admin.addPeer('$ENODE')" 2>/dev/null && echo "Auto-connected to bootnode" || echo "Note: Will auto-discover via bootnodes."
echo ""
BGEOF
chmod +x start-bg.sh

# Stop script
cat > stop.sh << STOPEOF
#!/usr/bin/env bash
PID=\$(pgrep -f "gprobe.*networkid $NETWORKID" || true)
if [[ -n "\$PID" ]]; then
    kill "\$PID"
    echo "Node stopped (PID: \$PID)"
else
    echo "No running node found."
fi
STOPEOF
chmod +x stop.sh

# ─── Done ─────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}${BOLD}============================================${NC}"
echo -e "${GREEN}${BOLD}  ProbeChain Rydberg Node Installed!${NC}"
echo -e "${GREEN}${BOLD}============================================${NC}"
echo ""
echo -e "  ${BOLD}Install directory:${NC}  $INSTALL_DIR"
echo -e "  ${BOLD}Validator address:${NC}  $ADDRESS"
echo -e "  ${BOLD}Chain ID:${NC}           $NETWORKID"
echo -e "  ${BOLD}Bootnode:${NC}           ${ENODE:0:40}..."
echo -e "  ${BOLD}HTTP RPC:${NC}           http://127.0.0.1:$HTTP_PORT"
echo ""
echo -e "  ${CYAN}Start node:${NC}    cd $INSTALL_DIR && ./start-bg.sh"
echo -e "  ${CYAN}Stop node:${NC}     cd $INSTALL_DIR && ./stop.sh"
echo -e "  ${CYAN}View logs:${NC}     tail -f $INSTALL_DIR/node.log"
echo ""
echo -e "${GREEN}${BOLD}Run ./start-bg.sh to launch your node now!${NC}"
echo ""
