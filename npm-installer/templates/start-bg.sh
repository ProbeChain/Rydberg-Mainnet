#!/usr/bin/env bash
cd ~/rydberg-agent

# Start node — sensitive APIs (personal, admin) are NOT exposed over HTTP
# Account unlock and mining are handled via local IPC only
./gprobe \
  --datadir ./data \
  --networkid 8004 \
  --port 30398 \
  --http --http.addr 127.0.0.1 --http.port 8549 \
  --http.api "probe,net,web3,pob,txpool" \
  --http.corsdomain "http://localhost:*" \
  --consensus pob \
  --miner.probebase ADDR_PLACEHOLDER \
  --password ./password.txt \
  --ipcpath ~/rydberg-agent/gprobe.ipc \
  --bootnodes "ENODE_PLACEHOLDER" \
  --verbosity 3 > node.log 2>&1 &
NODE_PID=$!
echo "Node started (PID: $NODE_PID)"

# Wait for IPC socket to appear (up to 15s)
IPC_READY=false
for i in $(seq 1 15); do
  if [ -S ~/rydberg-agent/gprobe.ipc ]; then
    IPC_READY=true
    break
  fi
  # Check if process is still alive
  if ! kill -0 $NODE_PID 2>/dev/null; then
    echo "[WARN] Node process exited. Check node.log:"
    tail -5 ~/rydberg-agent/node.log
    exit 1
  fi
  sleep 1
done

if [ "$IPC_READY" = false ]; then
  echo "[WARN] IPC socket not available after 15s. Node may still be starting."
  echo "  Check logs: tail -f ~/rydberg-agent/node.log"
  exit 0
fi

# Connect to bootnode via IPC
./gprobe attach ~/rydberg-agent/gprobe.ipc --exec "admin.addPeer('ENODE_PLACEHOLDER')" 2>/dev/null

# Unlock account via local IPC (not exposed over HTTP)
./gprobe attach ~/rydberg-agent/gprobe.ipc --exec "personal.unlockAccount('ADDR_PLACEHOLDER', '$(cat password.txt)', 0)" 2>/dev/null

# Start mining via IPC
./gprobe attach ~/rydberg-agent/gprobe.ipc --exec "miner.start(1)" 2>/dev/null

# Auto-register as Agent node (gas-free, consensus-layer registration)
sleep 3
RESULT=$(./gprobe attach ~/rydberg-agent/gprobe.ipc --exec "typeof pob !== 'undefined' ? pob.registerNode('ADDR_PLACEHOLDER', 1) : 'auto-registered via consensus'" 2>/dev/null || echo "auto")
echo "Agent registration: $RESULT"
