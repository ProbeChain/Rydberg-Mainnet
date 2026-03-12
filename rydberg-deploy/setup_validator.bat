@echo off
REM =============================================================
REM ProbeChain Rydberg Upgrade — Validator Node Setup (Windows)
REM =============================================================
REM Usage: setup_validator.bat
REM Run this on the Windows validator machine.
REM =============================================================

set GENESIS_NODE_IP=192.168.110.142
set GENESIS_ENODE=enode://59a7202485ca9e6067cb2d9f1071ac3c7258c9450648ee6bcfab07ba14533794a9282a7b287fdec904e15e3bb45e51c439eecd4d7904c6ebffb5d676f9e27107@%GENESIS_NODE_IP%:30303
set NETWORK_ID=8004
set DATADIR=data-validator
set PASSWORD_FILE=password.txt

echo ============================================
echo   ProbeChain Rydberg Upgrade
echo   Validator Node Setup (Windows)
echo ============================================
echo.

REM Step 1: Check gprobe binary
if not exist gprobe-windows-amd64.exe (
    echo ERROR: gprobe-windows-amd64.exe not found in current directory!
    pause
    exit /b 1
)

REM Step 2: Create password
set /p VALIDATOR_PASSWORD=Enter a password for your validator account:
echo %VALIDATOR_PASSWORD%> %PASSWORD_FILE%
echo.

REM Step 3: Create validator account
echo === Creating validator account ===
gprobe-windows-amd64.exe --datadir %DATADIR% account new --password %PASSWORD_FILE%

REM Extract account address
for /f "tokens=*" %%a in ('gprobe-windows-amd64.exe --datadir %DATADIR% account list 2^>^&1 ^| findstr "Account #0"') do set ACCOUNT_LINE=%%a
REM Parse the address between { and }
for /f "tokens=2 delims={}" %%b in ("%ACCOUNT_LINE%") do set ACCOUNT=%%b

echo.
echo Your validator address: 0x%ACCOUNT%
echo.

REM Step 4: Initialize with genesis
echo === Initializing chain with PoB genesis ===
gprobe-windows-amd64.exe --datadir %DATADIR% init genesis_pob.json
echo.

REM Step 5: Ports
set PORT=30304
set HTTP_PORT=8546
echo P2P port: %PORT%
echo HTTP port: %HTTP_PORT%
echo.

REM Step 6: Start node
echo === Starting validator node ===
echo Connecting to genesis node at: %GENESIS_NODE_IP%
echo.

start /b gprobe-windows-amd64.exe ^
  --datadir %DATADIR% ^
  --networkid %NETWORK_ID% ^
  --port %PORT% ^
  --http --http.addr 0.0.0.0 --http.port %HTTP_PORT% ^
  --http.api "probe,net,web3,personal,admin,miner,txpool,debug,pob" ^
  --http.corsdomain "*" ^
  --consensus pob ^
  --mine ^
  --miner.probebase "0x%ACCOUNT%" ^
  --unlock "0x%ACCOUNT%" ^
  --password %PASSWORD_FILE% ^
  --allow-insecure-unlock ^
  --bootnodes %GENESIS_ENODE% ^
  --verbosity 3 ^
  > validator.log 2>&1

echo Node starting... waiting 8 seconds...
timeout /t 8 /nobreak >nul

echo.
echo ============================================
echo   Validator Node Started!
echo ============================================
echo   Address:  0x%ACCOUNT%
echo   Port:     %PORT%
echo   HTTP:     %HTTP_PORT%
echo.
echo ============================================
echo   NEXT STEPS (on genesis node 192.168.110.142):
echo ============================================
echo.
echo   1. Vote to add this validator:
echo      curl -X POST http://127.0.0.1:8545 -H "Content-Type: application/json" ^
         -d "{\"jsonrpc\":\"2.0\",\"method\":\"pob_propose\",\"params\":[\"0x%ACCOUNT%\", true],\"id\":1}"
echo.
echo   Log file: validator.log
echo ============================================
echo.
echo NOTE: Make sure Windows Firewall allows port %PORT% (TCP/UDP)
echo       and that both machines are on the same network.
echo.
pause
