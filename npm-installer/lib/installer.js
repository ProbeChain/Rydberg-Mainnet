'use strict';

const fs = require('fs');
const path = require('path');
const os = require('os');
const readline = require('readline');
const { Writable } = require('stream');
const { spawnSync, execSync } = require('child_process');

const { detect, checkRequirements, hasCommand } = require('./platform');
const { downloadBinary, downloadGenesis, fetchBootnode } = require('./download');
const { buildFromSource } = require('./build');

// ─── Platform ────────────────────────────────────────────────────
const isWin = process.platform === 'win32';

// ─── Constants ──────────────────────────────────────────────────
const INSTALL_DIR = path.join(os.homedir(), 'rydberg-agent');
const BINARY_NAME = isWin ? 'gprobe.exe' : 'gprobe';
const GPROBE_BIN = path.join(INSTALL_DIR, BINARY_NAME);
const DATA_DIR = path.join(INSTALL_DIR, 'data');
const IPC_PATH = isWin ? '\\\\.\\pipe\\gprobe.ipc' : path.join(INSTALL_DIR, 'gprobe.ipc');
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
 * On Windows: named pipes can't be detected with fs, so try IPC attach.
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
  console.log('\x1b[1m  Chain ID: 8004 | PoB V2.1 OZ Gold Standard\x1b[0m');
  console.log('');

  // 1. Platform detection
  const info = detect();
  log(`Platform: ${info.label}`);

  const reqs = checkRequirements(info);
  if (!reqs.ok) {
    fail(`Missing requirements: ${reqs.missing.join(', ')}\nInstall them and retry.`);
  }

  // 2. Check if already installed
  if (fs.existsSync(GPROBE_BIN)) {
    if (isNodeRunning()) {
      ok('Rydberg Agent node is already installed and running');
      await cmdStatus();
      return;
    }
    // Node exists but not running — proceed with install
    warn('Existing installation found but not a running Rydberg node. Reinstalling...');
  }

  // 3. Password input
  const pwd = await readPassword('Enter node password (min 6 chars): ');
  if (!pwd || pwd.length < 6) {
    fail('Password must be at least 6 characters');
  }
  const pwd2 = await readPassword('Confirm password: ');
  if (pwd !== pwd2) {
    fail('Passwords do not match');
  }

  // Create install directory
  fs.mkdirSync(INSTALL_DIR, { recursive: true });

  // Save password (restricted permissions)
  if (isWin) {
    fs.writeFileSync(PASSWORD_FILE, pwd);
    try {
      execSync(`icacls "${PASSWORD_FILE}" /inheritance:r /grant:r "%USERNAME%:(R)"`, { stdio: 'ignore' });
    } catch {}
  } else {
    fs.writeFileSync(PASSWORD_FILE, pwd, { mode: 0o600 });
  }
  ok('Password saved');

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

  // 6. Create account
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

  // 8. Fetch bootnode
  log('Fetching bootnode...');
  const enode = await fetchBootnode(releaseTag);
  ok('Bootnode retrieved');

  // 9. Generate start script from template (use hex address for IPC compatibility)
  log('Generating start script...');
  if (isWin) {
    const template = fs.readFileSync(path.join(TEMPLATE_DIR, 'start-bg.bat'), 'utf8');
    const script = template
      .replace(/ADDR_PLACEHOLDER/g, hexAddr)
      .replace(/ENODE_PLACEHOLDER/g, enode);
    fs.writeFileSync(START_SCRIPT, script);
  } else {
    const template = fs.readFileSync(path.join(TEMPLATE_DIR, 'start-bg.sh'), 'utf8');
    const script = template
      .replace(/ADDR_PLACEHOLDER/g, hexAddr)
      .replace(/ENODE_PLACEHOLDER/g, enode);
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
