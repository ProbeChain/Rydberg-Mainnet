'use strict';

const https = require('https');
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const { execSync } = require('child_process');
const { hasCommand } = require('./platform');

const isWin = process.platform === 'win32';
const BINARY_NAME = isWin ? 'gprobe.exe' : 'gprobe';

const REPO = 'ProbeChain/Rydberg-Mainnet';
const GITHUB_API = `https://api.github.com/repos/${REPO}/releases/latest`;
const GITHUB_CONTENTS = `https://api.github.com/repos/${REPO}/contents`;

// GitHub proxy mirrors for China network acceleration
// User can override with GITHUB_PROXY env var
const GITHUB_PROXIES = [
  '', // direct first
  'https://ghproxy.net/',
  'https://gh-proxy.com/',
];

function getProxies() {
  const envProxy = process.env.GITHUB_PROXY;
  if (envProxy) return ['', envProxy.replace(/\/$/, '') + '/'];
  return GITHUB_PROXIES;
}

/**
 * HTTPS GET with redirect following (GitHub releases 302 to S3).
 * Returns a Promise<Buffer>.
 */
function httpsGet(url, opts = {}) {
  return new Promise((resolve, reject) => {
    const options = {
      headers: { 'User-Agent': 'rydberg-agent-node/3.1.0' },
      timeout: 120000,
      ...opts,
    };
    const req = https.get(url, options, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        return resolve(httpsGet(res.headers.location, opts));
      }
      if (res.statusCode !== 200) {
        return reject(new Error(`HTTP ${res.statusCode} for ${url}`));
      }
      const chunks = [];
      res.on('data', (c) => chunks.push(c));
      res.on('end', () => resolve(Buffer.concat(chunks)));
      res.on('error', reject);
    });
    req.on('error', reject);
    req.on('timeout', () => { req.destroy(); reject(new Error(`Timeout for ${url}`)); });
  });
}

/**
 * HTTPS GET with automatic retry on failure.
 */
async function httpsGetWithRetry(url, opts = {}, retries = 3) {
  for (let i = 0; i < retries; i++) {
    try {
      return await httpsGet(url, opts);
    } catch (err) {
      if (i < retries - 1) {
        const wait = (i + 1) * 3000;
        process.stdout.write(`\n  Retry ${i + 1}/${retries - 1} in ${wait / 1000}s (${err.message})...`);
        await new Promise(r => setTimeout(r, wait));
      } else {
        throw err;
      }
    }
  }
}

/**
 * Download a file with proxy fallback for China network.
 * Tries direct first, then GitHub proxy mirrors.
 */
async function downloadFileWithProxy(url, dest, label) {
  const proxies = getProxies();
  for (let i = 0; i < proxies.length; i++) {
    const proxyUrl = proxies[i] ? proxies[i] + url : url;
    const via = proxies[i] ? ` via ${proxies[i].replace('https://', '').replace('/', '')}` : '';
    try {
      process.stdout.write(`  Downloading ${label}${via}...`);
      const buf = await httpsGetWithRetry(proxyUrl, {}, 2);
      fs.writeFileSync(dest, buf);
      process.stdout.write(` done (${(buf.length / 1024 / 1024).toFixed(1)} MB)\n`);
      return dest;
    } catch (err) {
      process.stdout.write(` failed\n`);
      if (i < proxies.length - 1) {
        console.log(`  Trying mirror...`);
      } else {
        throw err;
      }
    }
  }
}

/**
 * Fetch the latest release metadata from GitHub.
 * Returns { tag, assets: [{name, url}] }
 */
async function fetchRelease() {
  const data = await httpsGetWithRetry(GITHUB_API);
  const json = JSON.parse(data.toString());
  const tag = json.tag_name;
  if (!tag) throw new Error('Could not determine latest release tag');

  const assets = (json.assets || []).map(a => ({
    name: a.name,
    url: a.browser_download_url,
  }));

  return { tag, assets };
}

/**
 * Verify SHA256 checksum of a file against SHA256SUMS content.
 */
function verifySha256(filePath, sha256sumsPath) {
  const fileName = path.basename(filePath);
  const sums = fs.readFileSync(sha256sumsPath, 'utf8');
  const line = sums.split('\n').find(l => l.includes(fileName));
  if (!line) throw new Error(`${fileName} not found in SHA256SUMS`);

  const expectedHash = line.trim().split(/\s+/)[0].toLowerCase();
  const fileData = fs.readFileSync(filePath);
  const actualHash = crypto.createHash('sha256').update(fileData).digest('hex');

  if (actualHash !== expectedHash) {
    throw new Error(
      `SHA256 mismatch for ${fileName}\n  expected: ${expectedHash}\n  actual:   ${actualHash}`
    );
  }
  console.log(`  SHA256 verified: ${fileName}`);
}

/**
 * Verify GPG signature (optional — skip if gpg not installed).
 */
function verifyGpg(sha256sumsPath, sigPath, pubkeyPath) {
  if (!hasCommand('gpg')) {
    console.log('  GPG not available — skipping signature verification');
    return;
  }
  if (!fs.existsSync(sigPath) || !fs.existsSync(pubkeyPath)) {
    console.log('  GPG signature files not found — skipping');
    return;
  }
  try {
    const devNull = isWin ? 'NUL' : '/dev/null';
    execSync(`gpg --import "${pubkeyPath}" 2>${devNull}`, { stdio: 'ignore' });
    execSync(`gpg --verify "${sigPath}" "${sha256sumsPath}" 2>${devNull}`, { stdio: 'ignore' });
    console.log('  GPG signature verified (ProbeChain <dev@probechain.org>)');
  } catch {
    throw new Error('GPG signature verification failed');
  }
}

/**
 * Download pre-built binary for macOS arm64 or Windows x64.
 * Performs SHA256 + optional GPG verification.
 * Uses proxy mirrors as fallback for China network.
 */
async function downloadBinary(installDir) {
  const { tag, assets } = await fetchRelease();

  // Match platform-specific archive asset
  let archiveAsset;
  if (isWin) {
    archiveAsset = assets.find(a => /windows.*x86_64.*\.zip$/i.test(a.name));
    if (!archiveAsset) throw new Error('No windows-x86_64 binary found in release');
  } else {
    archiveAsset = assets.find(a => /darwin.*arm64.*tar\.gz$/.test(a.name));
    if (!archiveAsset) throw new Error('No darwin-arm64 binary found in release');
  }

  const sumAsset = assets.find(a => a.name === 'SHA256SUMS');
  const sigAsset = assets.find(a => a.name === 'SHA256SUMS.asc');
  const keyAsset = assets.find(a => /gpg.*public/.test(a.name));

  if (!sumAsset) throw new Error('No SHA256SUMS found in release. Cannot verify binary integrity.');

  const archivePath = path.join(installDir, archiveAsset.name);
  const sumPath = path.join(installDir, 'SHA256SUMS');

  await downloadFileWithProxy(archiveAsset.url, archivePath, 'gprobe binary');
  await downloadFileWithProxy(sumAsset.url, sumPath, 'SHA256SUMS');

  // GPG verification (optional)
  let sigPath = null, keyPath = null;
  if (sigAsset) {
    sigPath = path.join(installDir, 'SHA256SUMS.asc');
    await downloadFileWithProxy(sigAsset.url, sigPath, 'SHA256SUMS.asc');
  }
  if (keyAsset) {
    keyPath = path.join(installDir, 'probechain-gpg-public.asc');
    await downloadFileWithProxy(keyAsset.url, keyPath, 'GPG public key');
  }
  if (sigPath && keyPath) {
    verifyGpg(sumPath, sigPath, keyPath);
  }

  // SHA256 verification (mandatory)
  verifySha256(archivePath, sumPath);

  // Extract
  if (isWin) {
    // Use PowerShell Expand-Archive for .zip files
    execSync(
      `powershell -NoProfile -Command "Expand-Archive -Path '${archivePath}' -DestinationPath '${installDir}' -Force"`,
      { cwd: installDir }
    );
  } else {
    execSync(`tar xzf "${archivePath}"`, { cwd: installDir });
    fs.chmodSync(path.join(installDir, 'gprobe'), 0o755);
  }
  console.log('  Binary extracted and ready');

  // Cleanup
  [archivePath, sumPath, sigPath, keyPath].forEach(f => {
    if (f && fs.existsSync(f)) fs.unlinkSync(f);
  });

  return tag;
}

/**
 * Fetch a file from the repo via GitHub Contents API (avoids raw.githubusercontent.com
 * which is DNS-blocked in some regions).
 * Uses Accept: application/vnd.github.raw to get raw content directly.
 */
async function fetchRepoFile(filePath, tag) {
  const url = `${GITHUB_CONTENTS}/${filePath}?ref=${tag}`;
  const data = await httpsGetWithRetry(url, {
    headers: {
      'User-Agent': 'rydberg-agent-node/3.1.0',
      'Accept': 'application/vnd.github.raw',
    },
  });
  return data;
}

/**
 * Download genesis.json pinned to a specific release tag.
 */
async function downloadGenesis(installDir, tag) {
  process.stdout.write('  Downloading genesis.json...');
  const data = await fetchRepoFile('genesis.json', tag);
  const dest = path.join(installDir, 'genesis.json');
  fs.writeFileSync(dest, data);
  process.stdout.write(' done\n');
  return dest;
}

/**
 * Fetch bootnode enode URL from the release tag.
 */
async function fetchBootnode(tag) {
  const data = await fetchRepoFile('bootnodes.txt', tag);
  const enode = data.toString().trim().split('\n')[0].trim();
  if (!enode.startsWith('enode://')) {
    throw new Error('Invalid bootnode format in bootnodes.txt');
  }
  return enode;
}

module.exports = {
  httpsGet,
  fetchRelease,
  downloadBinary,
  downloadGenesis,
  fetchBootnode,
};
