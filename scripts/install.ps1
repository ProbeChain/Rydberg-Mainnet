# ProbeChain Rydberg Testnet — Windows One-Line Installer
# Usage: irm https://raw.githubusercontent.com/ProbeChain/Rydberg-Mainnet/main/scripts/install.ps1 | iex
$ErrorActionPreference = "Stop"

$REPO = "ProbeChain/Rydberg-Mainnet"
$INSTALL_DIR = "$HOME\rydberg-agent"
$NETWORKID = 8004
$PORT = 30398
$HTTP_PORT = 8549
$BOOTNODES = "enode://e963c6b5342f0af311b0347fab33238aa564617241222dc93a4691ef2c76990b69e87e98a300aa08fc4deba023017006cd38e7cf28aa92bf13a93e8ff0c1387a@8.216.37.182:30398,enode://e742e55bae150ad4f004642d3a7365d2fe07f14b3ee7105fa47f6257e55937a90e5380b1a879c2e1e40295b0b482553536822b38615eecf444b5d8394c21a26e@8.216.49.15:30398,enode://bfb54dde94a526375a7ddbec0a4e5e174394c02f8a1c4c5ebec2aa5a05188398849895fc11973b09f3eec49dbd12ea9c8cb924e194e2857560c87d2f8a557bad@8.216.32.20:30398"

function Write-Info  { Write-Host "[INFO] $args" -ForegroundColor Cyan }
function Write-Ok    { Write-Host "[OK]   $args" -ForegroundColor Green }
function Write-Warn  { Write-Host "[WARN] $args" -ForegroundColor Yellow }
function Write-Fail  { Write-Host "[FAIL] $args" -ForegroundColor Red; exit 1 }

Write-Host ""
Write-Host "  ProbeChain Rydberg Testnet — Agent Node Installer" -ForegroundColor White
Write-Host "  Chain ID: 8004 | PoB V3.0.0 OZ Gold Standard" -ForegroundColor White
Write-Host ""

# Fast path: npm
if (Get-Command npx -ErrorAction SilentlyContinue) {
    Write-Info "Node.js detected — launching npm installer..."
    npx -y rydberg-agent-node
    exit $LASTEXITCODE
}
Write-Info "Node.js not found — using standalone installer."

# Kill existing node
try { taskkill /F /IM gprobe.exe 2>$null } catch {}
Start-Sleep -Seconds 2

# Clean and create install dir
if (Test-Path $INSTALL_DIR) {
    Write-Warn "Existing installation found. Reinstalling..."
    try { Remove-Item "$INSTALL_DIR\password.txt" -Force -ErrorAction SilentlyContinue } catch {}
    Get-ChildItem "$INSTALL_DIR\data\keystore\UTC--*" -ErrorAction SilentlyContinue | Remove-Item -Force
}
New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null
Set-Location $INSTALL_DIR

# Step 1: Download binary
Write-Info "Fetching latest release..."
$headers = @{ "User-Agent" = "rydberg-installer/1.0" }
$releaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest" -Headers $headers
$RELEASE_TAG = $releaseInfo.tag_name
Write-Ok "Release: $RELEASE_TAG"

$zipAsset = $releaseInfo.assets | Where-Object { $_.name -match "windows.*x86_64.*\.zip$" } | Select-Object -First 1
if (-not $zipAsset) { Write-Fail "No Windows binary found in release $RELEASE_TAG" }

Write-Info "Downloading gprobe binary..."
$zipPath = Join-Path $INSTALL_DIR $zipAsset.name
Invoke-WebRequest -Uri $zipAsset.browser_download_url -OutFile $zipPath -UseBasicParsing
Expand-Archive -Path $zipPath -DestinationPath $INSTALL_DIR -Force
Remove-Item $zipPath -Force
Write-Ok "gprobe.exe ready"

# Step 2: Download genesis.json
Write-Info "Downloading genesis.json..."
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/$REPO/$RELEASE_TAG/genesis.json" -OutFile (Join-Path $INSTALL_DIR "genesis.json") -UseBasicParsing
Write-Ok "Genesis downloaded"

# Step 3: Auto-generate password + create account
$pass = -join ((1..32) | ForEach-Object { [char](Get-Random -Min 97 -Max 123) })
$passFile = Join-Path $INSTALL_DIR "password.txt"
$pass | Out-File -Encoding ascii -NoNewline $passFile
Write-Info "Auto-generated node password"

# Clear old keystore
Get-ChildItem "$INSTALL_DIR\data\keystore\UTC--*" -ErrorAction SilentlyContinue | Remove-Item -Force

$acctOutput = & ".\gprobe.exe" --datadir .\data account new --password $passFile 2>&1 | Out-String
$ADDRESS = [regex]::Match($acctOutput, '0x[0-9a-fA-F]{40}').Value
if (-not $ADDRESS) { Write-Fail "Failed to create account.`n$acctOutput" }
Write-Ok "Account: $ADDRESS"

# Step 4: Init genesis
Write-Info "Initializing genesis..."
& ".\gprobe.exe" --datadir .\data init genesis.json 2>&1 | Select-Object -Last 1
Write-Ok "Genesis initialized (Chain ID: $NETWORKID)"

# Step 5: Fetch bootnodes
Write-Info "Fetching bootnodes..."
try {
    Invoke-WebRequest -Uri "https://raw.githubusercontent.com/$REPO/$RELEASE_TAG/bootnodes.txt" -OutFile (Join-Path $INSTALL_DIR "bootnodes.txt") -UseBasicParsing
    $ENODE = (Get-Content (Join-Path $INSTALL_DIR "bootnodes.txt") | Select-Object -First 1).Trim()
} catch { $ENODE = $BOOTNODES }
Write-Ok "Bootnodes ready"

# Step 6: Create start script (no IPC, use HTTP RPC)
$startBat = @"
@echo off
setlocal enabledelayedexpansion
cd /d "$INSTALL_DIR"
start "" /b gprobe.exe ^
  --datadir data ^
  --networkid $NETWORKID ^
  --port $PORT ^
  --http --http.addr 127.0.0.1 --http.port $HTTP_PORT ^
  --http.api "probe,net,web3,pob,txpool,personal,admin" ^
  --http.corsdomain "*" ^
  --consensus pob ^
  --mine ^
  --miner.probebase $ADDRESS ^
  --unlock $ADDRESS ^
  --password password.txt ^
  --allow-insecure-unlock ^
  --ipcdisable ^
  --syncmode full ^
  --bootnodes "$ENODE" ^
  --verbosity 3 > node.log 2>&1
echo Node started.
timeout /t 5 /nobreak >nul
gprobe.exe attach http://127.0.0.1:$HTTP_PORT --exec "typeof pob !== 'undefined' ? pob.registerNode('$ADDRESS', 1) : 'auto'" >nul 2>&1
"@
$startBat | Out-File -Encoding ascii (Join-Path $INSTALL_DIR "start-bg.bat")

$stopBat = @"
@echo off
taskkill /F /IM gprobe.exe 2>nul && echo Node stopped || echo No running node
"@
$stopBat | Out-File -Encoding ascii (Join-Path $INSTALL_DIR "stop.bat")

# Step 7: Start node
Write-Info "Starting node..."
& cmd.exe /c (Join-Path $INSTALL_DIR "start-bg.bat")
Start-Sleep -Seconds 8

$block = try { & ".\gprobe.exe" attach http://127.0.0.1:$HTTP_PORT --exec "probe.blockNumber" 2>$null } catch { "#syncing..." }
$peers = try { & ".\gprobe.exe" attach http://127.0.0.1:$HTTP_PORT --exec "admin.peers.length" 2>$null } catch { "0" }

Write-Host ""
Write-Host "============================================" -ForegroundColor Green
Write-Host "  Rydberg Agent Node Deployed!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""
Write-Host "  Address: $ADDRESS"
Write-Host "  Block:   $block"
Write-Host "  Peers:   $peers"
Write-Host ""
Write-Host "  Logs:    type $INSTALL_DIR\node.log" -ForegroundColor Cyan
Write-Host "  Stop:    $INSTALL_DIR\stop.bat" -ForegroundColor Cyan
Write-Host "  Status:  .\gprobe.exe attach http://127.0.0.1:$HTTP_PORT --exec `"probe.blockNumber`"" -ForegroundColor Cyan
Write-Host ""
