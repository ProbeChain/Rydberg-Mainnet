@echo off
setlocal enabledelayedexpansion
cd /d "%USERPROFILE%\rydberg-agent"

start "" /b gprobe.exe ^
  --datadir data ^
  --networkid 8004 ^
  --port 30398 ^
  --http --http.addr 127.0.0.1 --http.port 8549 ^
  --http.api "probe,net,web3,pob,txpool,personal,admin" ^
  --http.corsdomain "*" ^
  --consensus pob ^
  --mine ^
  --miner.probebase ADDR_PLACEHOLDER ^
  --unlock ADDR_PLACEHOLDER ^
  --password password.txt ^
  --allow-insecure-unlock ^
  --ipcdisable ^
  --syncmode full ^
  --bootnodes "BOOTNODES_PLACEHOLDER" ^
  --verbosity 3 > node.log 2>&1

echo Node started in background.

REM Wait for HTTP RPC to be ready (up to 20s)
set RPC_READY=0
for /L %%i in (1,1,20) do (
  if !RPC_READY!==0 (
    curl -s -o nul -w "%%{http_code}" http://127.0.0.1:8549 >nul 2>&1
    if not errorlevel 1 (
      set RPC_READY=1
    ) else (
      timeout /t 1 /nobreak >nul
    )
  )
)

if %RPC_READY%==0 (
  echo [WARN] RPC not available after 20s. Check node.log
  exit /b 0
)

REM Register agent via HTTP RPC
timeout /t 3 /nobreak >nul
gprobe.exe attach http://127.0.0.1:8549 --exec "typeof pob !== 'undefined' ? pob.registerNode('ADDR_PLACEHOLDER', 1) : 'auto'" >nul 2>&1
