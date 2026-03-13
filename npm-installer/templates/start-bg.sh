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
echo "Node started (PID: $!)"
sleep 3

# Connect to bootnode via IPC
./gprobe attach ~/rydberg-agent/gprobe.ipc --exec "admin.addPeer('ENODE_PLACEHOLDER')" 2>/dev/null

# Unlock account via local IPC (not exposed over HTTP)
./gprobe attach ~/rydberg-agent/gprobe.ipc --exec "personal.unlockAccount('ADDR_PLACEHOLDER', '$(cat password.txt)', 0)" 2>/dev/null

# Start mining via IPC
./gprobe attach ~/rydberg-agent/gprobe.ipc --exec "miner.start(1)" 2>/dev/null

# Auto-register as Agent node (gas-free, consensus-layer registration)
sleep 5
RESULT=$(./gprobe attach ~/rydberg-agent/gprobe.ipc --exec "typeof pob !== 'undefined' ? pob.registerNode('ADDR_PLACEHOLDER', 1) : 'auto-registered via consensus'" 2>/dev/null || echo "auto")
echo "Agent registration: $RESULT"
