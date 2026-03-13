# ProbeChain Rydberg Testnet — Windows One-Line Installer
# Usage: irm https://raw.githubusercontent.com/ProbeChain/Rydberg-Mainnet/main/scripts/install.ps1 | iex
# Or:    powershell -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/ProbeChain/Rydberg-Mainnet/main/scripts/install.ps1 | iex"
$ErrorActionPreference = "Stop"

# ─── Configuration ────────────────────────────────────────────────
$REPO = "ProbeChain/Rydberg-Mainnet"
$INSTALL_DIR = "$HOME\rydberg-node"
$NETWORKID = 8004
$PORT = 30398
$HTTP_PORT = 8549
$BOOTNODE = "enode://c56b6a7949fa9f6cf6e809863223fa9a444440a8f7fd4776ff5437f4c0db8d5775f7c0d3bfa0e6270242aa3811b776c9ef19d12c47a0f6e76f25b430a99071e9@bore.pub:9208"
# ──────────────────────────────────────────────────────────────────

function Write-Info  { Write-Host "[INFO] $args" -ForegroundColor Cyan }
function Write-Ok    { Write-Host "[OK]   $args" -ForegroundColor Green }
function Write-Warn  { Write-Host "[WARN] $args" -ForegroundColor Yellow }
function Write-Fail  { Write-Host "[FAIL] $args" -ForegroundColor Red; exit 1 }

# ─── Banner ───────────────────────────────────────────────────────
Write-Host ""
Write-Host "  ____            _           ____ _           _        " -ForegroundColor White
Write-Host " |  _ \ _ __ ___ | |__   ___ / ___| |__   __ _(_)_ __  " -ForegroundColor White
Write-Host " | |_) | '__/ _ \| '_ \ / _ \ |   | '_ \ / _`` | | '_ \ " -ForegroundColor White
Write-Host " |  __/| | | (_) | |_) |  __/ |___| | | | (_| | | | | |" -ForegroundColor White
Write-Host " |_|   |_|  \___/|_.__/ \___|\____|_| |_|\__,_|_|_| |_|" -ForegroundColor White
Write-Host ""
Write-Host "  Rydberg Testnet — OZ Gold Standard (PoB V2.1)" -ForegroundColor White
Write-Host "  Chain ID: 8004 | Windows Installer" -ForegroundColor White
Write-Host ""

# ─── Fast path: use npm installer if Node.js is available ────────
if (Get-Command npx -ErrorAction SilentlyContinue) {
    Write-Info "Node.js detected - launching npm installer for best experience..."
    npx rydberg-agent-node
    exit $LASTEXITCODE
}
Write-Info "Node.js not found - using standalone installer (no dependencies needed)."

# ─── Pre-checks ──────────────────────────────────────────────────
$ARCH = if ([System.Environment]::Is64BitOperatingSystem) { "x86_64" } else { "x86" }
if ($ARCH -ne "x86_64") { Write-Fail "Only 64-bit Windows is supported." }

if (Test-Path $INSTALL_DIR) {
    Write-Warn "Directory $INSTALL_DIR already exists."
    $ans = Read-Host "Overwrite and reinstall? [y/N]"
    if ($ans -notmatch "^[Yy]$") { Write-Info "Aborted."; exit 0 }
    Remove-Item -Recurse -Force $INSTALL_DIR
}

New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null
Set-Location $INSTALL_DIR

# ─── Step 1: Get latest release info ─────────────────────────────
Write-Info "Fetching latest release..."
$headers = @{ "User-Agent" = "rydberg-installer/1.0" }
$releaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest" -Headers $headers
$RELEASE_TAG = $releaseInfo.tag_name
Write-Ok "Release: $RELEASE_TAG"

# ─── Step 2: Download gprobe binary ──────────────────────────────
$zipAsset = $releaseInfo.assets | Where-Object { $_.name -match "windows.*x86_64.*\.zip$" } | Select-Object -First 1
if (-not $zipAsset) { Write-Fail "No Windows binary found in release $RELEASE_TAG" }

$sumAsset = $releaseInfo.assets | Where-Object { $_.name -eq "SHA256SUMS" } | Select-Object -First 1

Write-Info "Downloading gprobe binary..."
$zipPath = Join-Path $INSTALL_DIR $zipAsset.name
Invoke-WebRequest -Uri $zipAsset.browser_download_url -OutFile $zipPath -UseBasicParsing
Write-Ok "Binary downloaded: $($zipAsset.name)"

# Verify SHA256 if available
if ($sumAsset) {
    $sumPath = Join-Path $INSTALL_DIR "SHA256SUMS"
    Invoke-WebRequest -Uri $sumAsset.browser_download_url -OutFile $sumPath -UseBasicParsing
    $sums = Get-Content $sumPath
    $expectedLine = $sums | Where-Object { $_ -match $zipAsset.name }
    if ($expectedLine) {
        $expectedHash = ($expectedLine -split "\s+")[0].ToLower()
        $actualHash = (Get-FileHash $zipPath -Algorithm SHA256).Hash.ToLower()
        if ($actualHash -ne $expectedHash) {
            Write-Fail "SHA256 mismatch!`n  expected: $expectedHash`n  actual:   $actualHash"
        }
        Write-Ok "SHA256 verified"
    }
    Remove-Item $sumPath -Force -ErrorAction SilentlyContinue
}

# Extract
Write-Info "Extracting..."
Expand-Archive -Path $zipPath -DestinationPath $INSTALL_DIR -Force
Remove-Item $zipPath -Force
Write-Ok "gprobe.exe ready"

# ─── Step 3: Download genesis.json ───────────────────────────────
Write-Info "Downloading genesis.json..."
$genesisUrl = "https://api.github.com/repos/$REPO/contents/genesis.json?ref=$RELEASE_TAG"
$genesisHeaders = @{ "User-Agent" = "rydberg-installer/1.0"; "Accept" = "application/vnd.github.raw" }
Invoke-WebRequest -Uri $genesisUrl -OutFile (Join-Path $INSTALL_DIR "genesis.json") -Headers $genesisHeaders -UseBasicParsing
Write-Ok "genesis.json downloaded"

# ─── Step 4: Download bootnodes.txt ──────────────────────────────
Write-Info "Downloading bootnodes.txt..."
$bootUrl = "https://api.github.com/repos/$REPO/contents/bootnodes.txt?ref=$RELEASE_TAG"
Invoke-WebRequest -Uri $bootUrl -OutFile (Join-Path $INSTALL_DIR "bootnodes.txt") -Headers $genesisHeaders -UseBasicParsing
$ENODE = (Get-Content (Join-Path $INSTALL_DIR "bootnodes.txt") | Select-Object -First 1).Trim()
if (-not $ENODE.StartsWith("enode://")) { $ENODE = $BOOTNODE }
Write-Ok "Bootnode: $($ENODE.Substring(0,40))..."

# ─── Step 5: Create account ──────────────────────────────────────
Write-Info "Creating validator account..."
Write-Host ""
Write-Host "Set a password for your validator account." -ForegroundColor Yellow
Write-Host "Remember this password - you'll need it to start the node." -ForegroundColor Yellow
Write-Host ""

$secPass = Read-Host "Enter password (min 6 chars)" -AsSecureString
$secPass2 = Read-Host "Confirm password" -AsSecureString

$pass = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto(
    [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secPass))
$pass2 = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto(
    [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secPass2))

if ($pass -ne $pass2) { Write-Fail "Passwords do not match." }
if ($pass.Length -lt 6) { Write-Fail "Password must be at least 6 characters." }

$passFile = Join-Path $INSTALL_DIR "password.txt"
$pass | Out-File -Encoding ascii -NoNewline $passFile

$acctOutput = & ".\gprobe.exe" --datadir .\data account new --password $passFile 2>&1 | Out-String
$ADDRESS = [regex]::Match($acctOutput, '0x[0-9a-fA-F]{40}').Value

if (-not $ADDRESS) { Write-Fail "Failed to create account.`n$acctOutput" }
Write-Ok "Account created: $ADDRESS"

# ─── Step 6: Initialize genesis ──────────────────────────────────
Write-Info "Initializing genesis block..."
& ".\gprobe.exe" --datadir .\data init genesis.json 2>&1 | Select-Object -Last 1
Write-Ok "Genesis initialized (Chain ID: $NETWORKID)"

# ─── Step 7: Create start scripts ────────────────────────────────
$startScript = @"
@echo off
cd /d "$INSTALL_DIR"
echo Starting ProbeChain Rydberg node...
echo Address: $ADDRESS
echo HTTP RPC: http://127.0.0.1:$HTTP_PORT
echo Press Ctrl+C to stop.
echo.
gprobe.exe ^
  --datadir .\data ^
  --networkid $NETWORKID ^
  --port $PORT ^
  --http --http.addr 127.0.0.1 --http.port $HTTP_PORT ^
  --http.api "probe,net,web3,personal,admin,miner,txpool,pob" ^
  --http.corsdomain "*" ^
  --consensus pob ^
  --mine ^
  --miner.probebase $ADDRESS ^
  --unlock $ADDRESS ^
  --password password.txt ^
  --allow-insecure-unlock ^
  --syncmode full ^
  --bootnodes "$ENODE" ^
  --verbosity 3
"@
$startScript | Out-File -Encoding ascii (Join-Path $INSTALL_DIR "start.bat")

$stopScript = @"
@echo off
for /f "tokens=2" %%a in ('tasklist /fi "imagename eq gprobe.exe" /nh') do (
    taskkill /PID %%a /F
    echo Node stopped (PID: %%a)
    goto :done
)
echo No running node found.
:done
"@
$stopScript | Out-File -Encoding ascii (Join-Path $INSTALL_DIR "stop.bat")

# ─── Done ─────────────────────────────────────────────────────────
Write-Host ""
Write-Host "============================================" -ForegroundColor Green
Write-Host "  ProbeChain Rydberg Node Installed!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""
Write-Host "  Install directory:  $INSTALL_DIR"
Write-Host "  Validator address:  $ADDRESS"
Write-Host "  Chain ID:           $NETWORKID"
Write-Host "  Bootnode:           $($ENODE.Substring(0,40))..."
Write-Host "  HTTP RPC:           http://127.0.0.1:$HTTP_PORT"
Write-Host ""
Write-Host "  Start node:  cd $INSTALL_DIR; .\start.bat" -ForegroundColor Cyan
Write-Host "  Stop node:   cd $INSTALL_DIR; .\stop.bat" -ForegroundColor Cyan
Write-Host ""
Write-Host "Run .\start.bat to launch your node now!" -ForegroundColor Green
Write-Host ""
