#!/bin/bash
# =============================================================
# ProbeChain Rydberg Upgrade — Validator Node Setup Script
# =============================================================
# Usage: ./setup_validator.sh
# Run this on each new validator machine.
# =============================================================
set -e

GENESIS_NODE_IP="192.168.110.142"
GENESIS_ENODE="enode://59a7202485ca9e6067cb2d9f1071ac3c7258c9450648ee6bcfab07ba14533794a9282a7b287fdec904e15e3bb45e51c439eecd4d7904c6ebffb5d676f9e27107@${GENESIS_NODE_IP}:30303"
NETWORK_ID=8004
DATADIR="./data-validator"
PASSWORD_FILE="./password.txt"

echo "============================================"
echo "  ProbeChain Rydberg Upgrade"
echo "  Validator Node Setup"
echo "============================================"
echo ""

# Step 1: Check gprobe binary
if [ ! -f ./gprobe ]; then
    echo "ERROR: gprobe binary not found in current directory!"
    exit 1
fi
chmod +x ./gprobe

# Step 2: Create password
echo "Enter a password for your validator account:"
read -s VALIDATOR_PASSWORD
echo "$VALIDATOR_PASSWORD" > "$PASSWORD_FILE"
echo ""

# Step 3: Create validator account
echo "=== Creating validator account ==="
./gprobe --datadir "$DATADIR" account new --password "$PASSWORD_FILE" 2>&1
ACCOUNT=$(./gprobe --datadir "$DATADIR" account list 2>&1 | grep "Account #0" | sed 's/.*{\(.*\)}.*/\1/')
echo ""
echo "Your validator address: 0x${ACCOUNT}"
echo ""

# Step 4: Initialize with genesis
echo "=== Initializing chain with PoB genesis ==="
./gprobe --datadir "$DATADIR" init genesis_pob.json 2>&1
echo ""

# Step 5: Determine port (avoid conflict if on same machine)
PORT=30304
HTTP_PORT=8546
echo "P2P port: $PORT"
echo "HTTP port: $HTTP_PORT"
echo ""

# Step 6: Start node
echo "=== Starting validator node ==="
echo "Connecting to genesis node at: $GENESIS_NODE_IP"
echo ""

nohup ./gprobe \
  --datadir "$DATADIR" \
  --networkid $NETWORK_ID \
  --port $PORT \
  --http --http.addr "0.0.0.0" --http.port $HTTP_PORT \
  --http.api "probe,net,web3,personal,admin,miner,txpool,debug,pob" \
  --http.corsdomain "*" \
  --consensus pob \
  --mine \
  --miner.probebase "0x${ACCOUNT}" \
  --unlock "0x${ACCOUNT}" \
  --password "$PASSWORD_FILE" \
  --allow-insecure-unlock \
  --bootnodes "$GENESIS_ENODE" \
  --verbosity 3 \
  > validator.log 2>&1 &

echo "Node PID: $!"
sleep 5

echo ""
echo "=== Checking node status ==="
# Get this node's enode
ENODE=$(curl -s -X POST http://127.0.0.1:$HTTP_PORT \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['enode'])" 2>/dev/null || echo "FAILED")

BLOCK=$(curl -s -X POST http://127.0.0.1:$HTTP_PORT \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"probe_blockNumber","params":[],"id":1}' 2>/dev/null || echo "FAILED")

PEERS=$(curl -s -X POST http://127.0.0.1:$HTTP_PORT \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' 2>/dev/null || echo "FAILED")

echo ""
echo "============================================"
echo "  Validator Node Started!"
echo "============================================"
echo "  Address:  0x${ACCOUNT}"
echo "  Enode:    $ENODE"
echo "  Block:    $BLOCK"
echo "  Peers:    $PEERS"
echo ""
echo "============================================"
echo "  NEXT STEPS (on genesis node):"
echo "============================================"
echo ""
echo "  1. Add this node as peer (if not auto-connected):"
echo "     curl -X POST http://127.0.0.1:8545 -H 'Content-Type: application/json' \\"
echo "       -d '{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"ENODE_URL\"],\"id\":1}'"
echo ""
echo "  2. Vote to add this validator (on genesis node):"
echo "     curl -X POST http://127.0.0.1:8545 -H 'Content-Type: application/json' \\"
echo "       -d '{\"jsonrpc\":\"2.0\",\"method\":\"pob_propose\",\"params\":[\"0x${ACCOUNT}\", true],\"id\":1}'"
echo ""
echo "  Log file: validator.log"
echo "============================================"
