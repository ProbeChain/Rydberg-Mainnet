@echo off
setlocal enabledelayedexpansion
cd /d "%USERPROFILE%\rydberg-agent"

start "" /b gprobe.exe ^
  --datadir data ^
  --networkid 8004 ^
  --port 30398 ^
  --http --http.addr 127.0.0.1 --http.port 8549 ^
  --http.api "probe,net,web3,pob,txpool" ^
  --http.corsdomain "http://localhost:*" ^
  --consensus pob ^
  --mine ^
  --miner.probebase ADDR_PLACEHOLDER ^
  --unlock ADDR_PLACEHOLDER ^
  --password password.txt ^
  --allow-insecure-unlock ^
  --ipcpath \\.\pipe\gprobe.ipc ^
  --syncmode full ^
  --bootnodes "BOOTNODES_PLACEHOLDER" ^
  --verbosity 3 > node.log 2>&1

echo Node started in background.

REM Wait for IPC pipe to be ready (up to 15s)
REM Named pipes cannot be checked with "if exist", so try attach in a loop
set IPC_READY=0
for /L %%i in (1,1,15) do (
  if !IPC_READY!==0 (
    gprobe.exe attach \\.\pipe\gprobe.ipc --exec "admin.nodeInfo.protocols.probe.network" >nul 2>&1
    if not errorlevel 1 (
      set IPC_READY=1
    ) else (
      timeout /t 1 /nobreak >nul
    )
  )
)

if %IPC_READY%==0 (
  echo [WARN] IPC not available after 15s. Check node.log
  exit /b 0
)

REM Connect to bootnodes
gprobe.exe attach \\.\pipe\gprobe.ipc --exec "ADDPEER_PLACEHOLDER" >nul 2>&1

REM Register agent
timeout /t 3 /nobreak >nul
gprobe.exe attach \\.\pipe\gprobe.ipc --exec "typeof pob !== 'undefined' ? pob.registerNode('ADDR_PLACEHOLDER', 1) : 'auto'" >nul 2>&1
