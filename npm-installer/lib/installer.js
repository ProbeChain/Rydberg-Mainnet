'use strict';

const fs = require('fs');
const path = require('path');
const os = require('os');
const readline = require('readline');
const { Writable } = require('stream');
const { spawnSync, execSync } = require('child_process');

const { detect, checkRequirements, hasCommand } = require('./platform');
const { downloadBinary, downloadGenesis, fetchBootnodes } = require('./download');
const { buildFromSource } = require('./build');

// ─── Platform ────────────────────────────────────────────────────
const isWin = process.platform === 'win32';

// ─── Constants ──────────────────────────────────────────────────
const INSTALL_DIR = path.join(os.homedir(), 'rydberg-agent');
const BINARY_NAME = isWin ? 'gprobe.exe' : 'gprobe';
const GPROBE_BIN = path.join(INSTALL_DIR, BINARY_NAME);
const DATA_DIR = path.join(INSTALL_DIR, 'data');
const IPC_PATH = isWin ? 'http://127.0.0.1:8549' : path.join(INSTALL_DIR, 'gprobe.ipc');
const PASSWORD_FILE = path.join(INSTALL_DIR, 'password.txt');
const LOG_FILE = path.join(INSTALL_DIR, 'node.log');
const START_SCRIPT = path.join(INSTALL_DIR, isWin ? 'start-bg.bat' : 'start-bg.sh');
const TEMPLATE_DIR = path.join(__dirname, '..', 'templates');

// ─── Helpers ────────────────────────────────────────────────────

function log(msg) { console.log(`\x1b[36m[INFO]\x1b[0m ${msg}`); }
function ok(msg)  { console.log(`\x1b[32m[OK]\x1b[0m ${msg}`); }
function warn(msg){ console.log(`\x1b[33m[WARN]\x1b[0m ${msg}`); }
function fail(msg){ console.error(`\x1b[31m[FAIL]\x1b[0m ${msg}`); process.exit(1); }

/**
 * Cross-platform sleep (synchronous).
 */
function sleepSync(seconds) {
  if (isWin) {
    spawnSync(process.env.COMSPEC || 'cmd.exe', ['/c', `timeout /t ${seconds} /nobreak >nul`]);
  } else {
    spawnSync('sleep', [String(seconds)]);
  }
}

/**
 * Read password from stdin with hidden input.
 */
function readPassword(prompt) {
  return new Promise((resolve) => {
    const muted = new Writable({
      write(_chunk, _enc, cb) { cb(); },
    });
    const rl = readline.createInterface({
      input: process.stdin,
      output: muted,
      terminal: true,
    });
    process.stdout.write(prompt);
    rl.question('', (answer) => {
      rl.close();
      process.stdout.write('\n');
      resolve(answer);
    });
  });
}

/**
 * Run gprobe with arguments, return stdout.
 */
function gprobe(...args) {
  const result = spawnSync(GPROBE_BIN, args, {
    cwd: INSTALL_DIR,
    encoding: 'utf8',
    timeout: 30000,
  });
  return (result.stdout || '') + (result.stderr || '');
}

/**
 * Attach to IPC and execute a JS expression.
 */
function ipcExec(expr) {
  const result = spawnSync(GPROBE_BIN, [
    'attach', IPC_PATH, '--exec', expr,
  ], {
    cwd: INSTALL_DIR,
    encoding: 'utf8',
    timeout: 15000,
  });
  return (result.stdout || '').trim();
}

/**
 * Check if the node is running.
 * On Unix: check for IPC socket file.
 * On Windows: try HTTP RPC (IPC named pipes have permission issues).
 */
function isNodeRunning() {
  if (isWin) {
    try {
      const r = ipcExec('admin.nodeInfo.protocols.probe.network');
      return r.includes('8004');
    } catch {
      return false;
    }
  }
  return fs.existsSync(IPC_PATH);
}

/**
 * Get the account address from the keystore directory.
 */
function getAddress() {
  const keystoreDir = path.join(DATA_DIR, 'keystore');
  if (!fs.existsSync(keystoreDir)) return null;
  const files = fs.readdirSync(keystoreDir);
  if (files.length === 0) return null;
  const match = files[0].match(/([0-9a-f]{40})/);
  return match ? '0x' + match[1] : null;
}

/**
 * Kill gprobe processes (cross-platform).
 */
function killGprobeProcesses() {
  if (isWin) {
    try {
      const out = execSync(
        'wmic process where "CommandLine like \'%gprobe%networkid 8004%\'" get ProcessId',
        { encoding: 'utf8', stdio: ['pipe', 'pipe', 'ignore'] }
      );
      const pids = out.match(/\d+/g);
      if (pids) {
        pids.forEach(pid => {
          try { execSync(`taskkill /PID ${pid} /F`, { stdio: 'ignore' }); } catch {}
        });
        return pids;
      }
    } catch {}
    return null;
  } else {
    try {
      const out = execSync('pgrep -f "gprobe.*networkid 8004"', { encoding: 'utf8' }).trim();
      if (out) {
        const pids = out.split('\n');
        pids.forEach(pid => {
          process.kill(parseInt(pid, 10), 'SIGTERM');
        });
        return pids;
      }
    } catch {}
    return null;
  }
}

/**
 * Kill any process using port 30398 (cross-platform).
 */
function killPortProcesses() {
  if (isWin) {
    try {
      const out = execSync('netstat -ano | findstr :30398 | findstr LISTENING', {
        encoding: 'utf8', stdio: ['pipe', 'pipe', 'ignore'],
      });
      const pids = new Set();
      out.split('\n').forEach(line => {
        const m = line.trim().match(/\s(\d+)\s*$/);
        if (m) pids.add(m[1]);
      });
      pids.forEach(pid => {
        try { execSync(`taskkill /PID ${pid} /F`, { stdio: 'ignore' }); } catch {}
      });
    } catch {}
  } else {
    try {
      execSync('pkill -9 -f "gprobe.*networkid 8004" 2>/dev/null; lsof -ti :30398 | xargs kill -9 2>/dev/null; true', { stdio: 'ignore' });
    } catch {}
  }
}

/**
 * Read the last N lines of a file (cross-platform replacement for `tail`).
 */
function readLastLines(filePath, n) {
  const content = fs.readFileSync(filePath, 'utf8');
  const lines = content.split('\n');
  return lines.slice(-n).join('\n');
}

// ─── Sub-commands ───────────────────────────────────────────────

async function cmdStatus() {
  if (!fs.existsSync(GPROBE_BIN)) {
    fail('Rydberg Agent node is not installed. Run: npx rydberg-agent-node');
  }
  const addr = getAddress();
  const block = ipcExec('probe.blockNumber');
  const peers = ipcExec('admin.peers.length');
  if (!block || block.includes('error') || block.includes('Fatal')) {
    warn('Node is not running');
    console.log(`  Address: ${addr || 'unknown'}`);
    console.log(`  Install: ${INSTALL_DIR}`);
    return;
  }
  console.log(`  Block:   #${block}`);
  console.log(`  Peers:   ${peers}`);
  console.log(`  Address: ${addr}`);
  if (addr) {
    let balance = ipcExec(`web3.fromWei(probe.getBalance('${addr}'), 'probeer')`);
    if (balance.includes('Error') || balance.includes('error')) balance = '0';
    console.log(`  Balance: ${balance} PROBE`);
    let agentStatus = ipcExec(`typeof pob !== 'undefined' ? pob.getNodeRegistrationStatus('${addr}') : 'auto'`);
    if (agentStatus.includes('Error') || agentStatus.includes('error') || agentStatus.includes('ReferenceError')) {
      agentStatus = 'auto (via consensus)';
    }
    console.log(`  Agent:   ${agentStatus}`);
  }
}

async function cmdStop() {
  const pids = killGprobeProcesses();
  if (pids) {
    ok(`Node stopped (PID: ${pids.join(', ')})`);
  } else {
    warn('No running Rydberg node found');
  }
}

async function cmdStart() {
  if (!fs.existsSync(START_SCRIPT)) {
    const scriptName = isWin ? 'start-bg.bat' : 'start-bg.sh';
    fail(`${scriptName} not found. Run: npx rydberg-agent-node  to install first`);
  }
  log('Starting Rydberg Agent node...');
  let result;
  if (isWin) {
    result = spawnSync(process.env.COMSPEC || 'cmd.exe', ['/c', START_SCRIPT], {
      cwd: INSTALL_DIR,
      stdio: 'inherit',
    });
  } else {
    result = spawnSync('bash', [START_SCRIPT], {
      cwd: INSTALL_DIR,
      stdio: 'inherit',
    });
  }
  if (result.status === 0) {
    ok('Node started');
  } else {
    fail('Failed to start node');
  }
}

async function cmdLogs() {
  if (!fs.existsSync(LOG_FILE)) {
    fail('No log file found. Is the node installed?');
  }
  if (isWin || !hasCommand('tail')) {
    console.log(readLastLines(LOG_FILE, 30));
  } else {
    spawnSync('tail', ['-30', LOG_FILE], { stdio: 'inherit' });
  }
}

// ─── Main Install Flow ─────────────────────────────────────────

async function cmdInstall() {
  // Banner
  console.log('');
  console.log('\x1b[1m  ProbeChain Rydberg Testnet — Agent Node Installer\x1b[0m');
  console.log('\x1b[1m  Chain ID: 8004 | PoB V3.0.0 OZ Gold Standard | 12 Validators\x1b[0m');
  console.log('');

  // 1. Platform detection
  const info = detect();
  log(`Platform: ${info.label}`);

  const reqs = checkRequirements(info);
  if (!reqs.ok) {
    fail(`Missing requirements: ${reqs.missing.join(', ')}\nInstall them and retry.`);
  }

  // 2. Check if already installed — auto-upgrade if old version detected
  let reusePassword = false;
  if (fs.existsSync(GPROBE_BIN)) {
    // Check if genesis is outdated by comparing with latest from GitHub
    let needsUpgrade = false;
    const localGenesis = path.join(INSTALL_DIR, 'genesis.json');
    if (fs.existsSync(localGenesis)) {
      try {
        const local = JSON.parse(fs.readFileSync(localGenesis, 'utf8'));
        const localTs = local.timestamp || '';
        const localValidators = (local.config && local.config.pob && local.config.pob.list) || [];
        // If timestamp or validator count differs from current release, upgrade needed
        // We check by downloading fresh genesis later; for now detect stale chain (0 peers after running)
        if (isNodeRunning()) {
          const peers = ipcExec('admin.peers.length');
          const block = ipcExec('probe.blockNumber');
          const peerNum = parseInt(peers, 10) || 0;
          const blockNum = parseInt(block, 10) || 0;
          if (peerNum === 0 && blockNum > 0) {
            needsUpgrade = true;
            warn(`Old chain detected (block #${blockNum}, 0 peers). Upgrading to latest version...`);
          } else if (peerNum > 0) {
            ok('Rydberg Agent node is already installed and running (latest)');
            await cmdStatus();
            return;
          } else {
            needsUpgrade = true;
            warn('Node running but no peers. Upgrading to latest chain...');
          }
        } else {
          needsUpgrade = true;
          warn('Existing installation found but node not running. Reinstalling...');
        }
      } catch {
        needsUpgrade = true;
      }
    } else {
      needsUpgrade = true;
    }

    if (needsUpgrade) {
      // Kill old processes
      log('Stopping old node...');
      if (isWin) {
        try { execSync('taskkill /F /IM gprobe.exe 2>nul', { stdio: 'ignore' }); } catch {}
        try { execSync('cmd /c "for /f \\"tokens=5\\" %a in (\'netstat -aon ^| findstr :30398 ^| findstr LISTENING\') do taskkill /F /PID %a"', { stdio: 'ignore' }); } catch {}
        try { execSync('cmd /c "for /f \\"tokens=5\\" %a in (\'netstat -aon ^| findstr :8549 ^| findstr LISTENING\') do taskkill /F /PID %a"', { stdio: 'ignore' }); } catch {}
        try { execSync('wmic process where "CommandLine like \'%gprobe%\'" call terminate 2>nul', { stdio: 'ignore' }); } catch {}
        sleepSync(3);
      } else {
        try { execSync('pkill -9 -f "gprobe.*networkid 8004" 2>/dev/null', { stdio: 'ignore' }); } catch {}
        sleepSync(2);
      }
      ok('Old node stopped');

      // Handle old password.txt — ask user
      if (fs.existsSync(PASSWORD_FILE)) {
        warn('Found old password file: ' + PASSWORD_FILE);
        const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
        const answer = await new Promise(resolve => {
          rl.question('\x1b[33m  Delete old password and generate new one? [Y/n]: \x1b[0m', resolve);
        });
        rl.close();
        if (answer.toLowerCase() === 'n') {
          reusePassword = true;
          ok('Keeping old password (will reuse for new account)');
        } else {
          try { fs.unlinkSync(PASSWORD_FILE); } catch {}
          if (isWin) {
            try { execSync(`del /f /q "${PASSWORD_FILE}" 2>nul`, { stdio: 'ignore' }); } catch {}
            try { if (fs.existsSync(PASSWORD_FILE)) fs.renameSync(PASSWORD_FILE, PASSWORD_FILE + '.old'); } catch {}
          }
          ok('Old password deleted');
        }
      }

      // Clean old chaindata (keep keystore dir removal for step 6)
      const chaindata = path.join(DATA_DIR, 'gprobe', 'chaindata');
      const lightdata = path.join(DATA_DIR, 'gprobe', 'lightchaindata');
      const nodes = path.join(DATA_DIR, 'gprobe', 'nodes');
      for (const d of [chaindata, lightdata, nodes]) {
        if (fs.existsSync(d)) {
          fs.rmSync(d, { recursive: true, force: true });
        }
      }
      ok('Old chain data cleaned');
    }
  }

  // 3. Auto-generate password (unless reusing old one)
  let pwd;
  if (reusePassword && fs.existsSync(PASSWORD_FILE)) {
    pwd = fs.readFileSync(PASSWORD_FILE, 'utf8').trim();
    log('Reusing existing password');
  } else {
    pwd = require('crypto').randomBytes(16).toString('hex');
    log('Auto-generated node password');
  }

  // Create install directory
  fs.mkdirSync(INSTALL_DIR, { recursive: true });

  // Save password (restricted permissions, retry if file locked)
  for (let attempt = 0; attempt < 3; attempt++) {
    try {
      if (isWin) {
        fs.writeFileSync(PASSWORD_FILE, pwd);
        try { execSync(`icacls "${PASSWORD_FILE}" /inheritance:r /grant:r "%USERNAME%:(R)"`, { stdio: 'ignore' }); } catch {}
      } else {
        fs.writeFileSync(PASSWORD_FILE, pwd, { mode: 0o600 });
      }
      ok('Password saved');
      break;
    } catch (e) {
      if (attempt < 2) {
        warn(`Password file locked, retrying... (${e.code || e.message})`);
        if (isWin) { try { execSync('taskkill /F /IM gprobe.exe 2>nul', { stdio: 'ignore' }); } catch {} }
        sleepSync(3);
      } else {
        fail(`Cannot write password file: ${e.message}. Stop any running gprobe process and try again.`);
      }
    }
  }

  // 4. Download or build gprobe (fallback to source build if download fails)
  let releaseTag;
  if (info.prebuilt) {
    try {
      const platformLabel = isWin ? 'Windows x64' : 'macOS arm64';
      log(`Downloading pre-built binary (${platformLabel})...`);
      releaseTag = await downloadBinary(INSTALL_DIR);
    } catch (err) {
      warn(`Binary download failed: ${err.message}`);
      log('Falling back to building from source...');
      releaseTag = await buildFromSource(INSTALL_DIR);
    }
  } else {
    log('Building from source...');
    releaseTag = await buildFromSource(INSTALL_DIR);
  }
  ok(`gprobe ready (${releaseTag})`);

  // 5. Download genesis.json
  log('Downloading genesis.json...');
  await downloadGenesis(INSTALL_DIR, releaseTag);
  ok('Genesis config downloaded');

  // 6. Create account (clear old keystore if reinstalling — password mismatch)
  const keystoreDir = path.join(DATA_DIR, 'keystore');
  if (fs.existsSync(keystoreDir)) {
    const oldKeys = fs.readdirSync(keystoreDir).filter(f => f.startsWith('UTC-'));
    if (oldKeys.length > 0) {
      warn('Removing old keystore (new password generated)');
      for (const f of oldKeys) { try { fs.unlinkSync(path.join(keystoreDir, f)); } catch {} }
    }
  }
  log('Creating account...');
  const accountOutput = gprobe('--datadir', './data', 'account', 'new', '--password', 'password.txt');
  // Match pro1... (bech32) or 0x... (hex) address format
  const bech32Match = accountOutput.match(/pro1[a-z0-9]{38,}/);
  const hexMatch = accountOutput.match(/0x[0-9a-fA-F]{40}/);

  // Always extract hex address from keystore (needed for IPC/JS console)
  let hexAddr = hexMatch ? hexMatch[0] : null;
  if (!hexAddr) {
    const keystoreDir = path.join(DATA_DIR, 'keystore');
    if (fs.existsSync(keystoreDir)) {
      const files = fs.readdirSync(keystoreDir);
      const kMatch = files.length > 0 && files[0].match(/([0-9a-f]{40})/);
      if (kMatch) hexAddr = '0x' + kMatch[1];
    }
  }
  if (!hexAddr) {
    fail(`Failed to create account. Output:\n${accountOutput}`);
  }

  // Display address (prefer bech32 pro1 format if available)
  const displayAddr = bech32Match ? bech32Match[0] : hexAddr;
  ok(`Account created: ${displayAddr}`);
  if (bech32Match && hexAddr) {
    log(`Hex address:  ${hexAddr}`);
  }

  // 7. Initialize genesis
  log('Initializing genesis block...');
  const initOutput = gprobe('--datadir', './data', 'init', 'genesis.json');
  ok('Genesis initialized (Chain ID: 8004)');

  // 8. Fetch bootnodes
  log('Fetching bootnodes...');
  const enodes = await fetchBootnodes(releaseTag);
  ok(`${enodes.length} bootnode(s) retrieved`);

  const bootnodesCsv = enodes.join(',');
  const addPeerCmds = enodes.map(e => `admin.addPeer('${e}')`).join('; ');

  // 9. Generate start script from template (use hex address for IPC compatibility)
  log('Generating start script...');
  if (isWin) {
    const template = fs.readFileSync(path.join(TEMPLATE_DIR, 'start-bg.bat'), 'utf8');
    const script = template
      .replace(/ADDR_PLACEHOLDER/g, hexAddr)
      .replace(/BOOTNODES_PLACEHOLDER/g, bootnodesCsv)
      .replace(/ADDPEER_PLACEHOLDER/g, addPeerCmds);
    fs.writeFileSync(START_SCRIPT, script);
  } else {
    const template = fs.readFileSync(path.join(TEMPLATE_DIR, 'start-bg.sh'), 'utf8');
    const script = template
      .replace(/ADDR_PLACEHOLDER/g, hexAddr)
      .replace(/BOOTNODES_PLACEHOLDER/g, bootnodesCsv)
      .replace(/ADDPEER_PLACEHOLDER/g, addPeerCmds);
    fs.writeFileSync(START_SCRIPT, script, { mode: 0o755 });
  }
  ok(`${isWin ? 'start-bg.bat' : 'start-bg.sh'} generated`);

  // Also generate stop script for convenience
  if (isWin) {
    const stopBat = `@echo off
for /f "tokens=2 delims==" %%p in ('wmic process where "CommandLine like '%%gprobe%%networkid 8004%%'" get ProcessId /format:list 2^>nul ^| findstr ProcessId') do (
  taskkill /PID %%p /F >nul 2>&1
  echo Node stopped (PID: %%p^)
)
`;
    fs.writeFileSync(path.join(INSTALL_DIR, 'stop.bat'), stopBat);
  } else {
    const stopScript = `#!/usr/bin/env bash
PID=$(pgrep -f "gprobe.*networkid 8004" || true)
if [ -n "$PID" ]; then
    kill "$PID"
    echo "Node stopped (PID: $PID)"
else
    echo "No running node found."
fi
`;
    fs.writeFileSync(path.join(INSTALL_DIR, 'stop.sh'), stopScript, { mode: 0o755 });
  }

  // 10. Kill any process using port 30398 before starting
  killGprobeProcesses();
  killPortProcesses();

  // Wait for port to be released
  for (let i = 0; i < 10; i++) {
    if (isWin) {
      try {
        execSync('netstat -ano | findstr :30398 | findstr LISTENING', {
          encoding: 'utf8', stdio: ['pipe', 'pipe', 'ignore'],
        });
        sleepSync(1);
      } catch { break; /* port is free */ }
    } else {
      try {
        execSync('lsof -i :30398', { stdio: 'ignore' });
        sleepSync(1);
      } catch { break; /* port is free */ }
    }
  }

  log('Starting Rydberg Agent node...');
  if (isWin) {
    spawnSync(process.env.COMSPEC || 'cmd.exe', ['/c', START_SCRIPT], {
      cwd: INSTALL_DIR,
      stdio: 'inherit',
    });
  } else {
    spawnSync('bash', [START_SCRIPT], {
      cwd: INSTALL_DIR,
      stdio: 'inherit',
    });
  }

  // 11. Verify — check if IPC is available before querying
  log('Verifying node...');
  await new Promise(r => setTimeout(r, 3000));

  const ipcAvailable = isNodeRunning();
  let block = '', peers = '', balance = '0', agentStatus = 'auto (via consensus)';

  if (ipcAvailable) {
    block = ipcExec('probe.blockNumber');
    peers = ipcExec('admin.peers.length');
    balance = ipcExec(`web3.fromWei(probe.getBalance('${hexAddr}'), 'probeer')`);
    if (balance.includes('Error') || balance.includes('Fatal')) balance = '0';
    agentStatus = ipcExec(`typeof pob !== 'undefined' ? pob.getNodeRegistrationStatus('${hexAddr}') : 'auto'`);
    if (agentStatus.includes('Error') || agentStatus.includes('Fatal') || agentStatus.includes('ReferenceError')) {
      agentStatus = 'auto (via consensus)';
    }
  } else {
    const logHint = isWin ? `type ${LOG_FILE}` : `tail -f ${LOG_FILE}`;
    warn('IPC not available — node may have exited. Check: ' + logHint);
  }

  console.log('');
  console.log('\x1b[32m\x1b[1m============================================\x1b[0m');
  console.log('\x1b[32m\x1b[1m  Rydberg Agent Node Deployed!\x1b[0m');
  console.log('\x1b[32m\x1b[1m============================================\x1b[0m');
  console.log('');
  console.log(`  Address: ${displayAddr}`);
  if (displayAddr !== hexAddr) {
    console.log(`  Hex:     ${hexAddr}`);
  }
  console.log(`  Type:    Agent (PoB NodeType=1)`);
  console.log(`  Block:   #${block || 'syncing...'}`);
  console.log(`  Peers:   ${peers || '0'}`);
  console.log(`  Balance: ${balance || '0'} PROBE`);
  console.log(`  Agent:   ${agentStatus}`);
  console.log('');
  console.log('  Agent nodes receive 40% of block rewards,');
  console.log('  distributed by behavior score (initial: 5000).');
  console.log('');
  if (isWin) {
    console.log(`  Logs:    type ${LOG_FILE}`);
  } else {
    console.log(`  Logs:    tail -f ${LOG_FILE}`);
  }
  console.log(`  Stop:    npx rydberg-agent-node stop`);
  console.log(`  Status:  npx rydberg-agent-node status`);
  console.log('');
}

module.exports = { cmdInstall, cmdStatus, cmdStop, cmdStart, cmdLogs };
