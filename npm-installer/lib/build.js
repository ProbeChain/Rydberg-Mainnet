'use strict';

const { execSync, spawnSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const { fetchRelease } = require('./download');

const isWin = process.platform === 'win32';
const BINARY_NAME = isWin ? 'gprobe.exe' : 'gprobe';
const REPO = 'ProbeChain/Rydberg-Mainnet';

/**
 * Cross-platform synchronous sleep.
 */
function sleepSync(seconds) {
  if (isWin) {
    spawnSync(process.env.COMSPEC || 'cmd.exe', ['/c', `timeout /t ${seconds} /nobreak >nul`]);
  } else {
    spawnSync('sleep', [String(seconds)]);
  }
}

// Git clone URLs to try (direct + mirrors for China)
function getCloneUrls() {
  const envProxy = process.env.GITHUB_PROXY;
  if (envProxy) {
    const proxy = envProxy.replace(/\/$/, '');
    return [
      `https://github.com/${REPO}.git`,
      `${proxy}/https://github.com/${REPO}.git`,
    ];
  }
  return [
    `https://github.com/${REPO}.git`,
    `https://ghproxy.net/https://github.com/${REPO}.git`,
  ];
}

/**
 * Build gprobe from source for platforms without a pre-built binary.
 * Clones the repo at the release tag and runs `go build`.
 * Tries direct GitHub then proxy mirrors with HTTP/1.1 fallback.
 * Returns the release tag.
 */
async function buildFromSource(installDir) {
  const { tag } = await fetchRelease();
  if (!tag) throw new Error('Could not determine release tag');

  const srcDir = path.join(installDir, 'src');
  const urls = getCloneUrls();

  console.log(`  Cloning ${REPO} @ ${tag}...`);
  let cloned = false;

  for (let attempt = 1; attempt <= 4; attempt++) {
    if (fs.existsSync(srcDir)) fs.rmSync(srcDir, { recursive: true, force: true });

    // Cycle through URLs: attempt 1-2 use direct, 3-4 use mirror
    const urlIdx = attempt <= 2 ? 0 : Math.min(1, urls.length - 1);
    const url = urls[urlIdx];
    const useHttp11 = attempt % 2 === 0; // even attempts force HTTP/1.1

    const gitArgs = [];
    if (useHttp11) gitArgs.push('-c', 'http.version=HTTP/1.1');
    gitArgs.push('clone', '--branch', tag, '--depth', '1', url, srcDir);

    const label = url.includes('github.com/' + REPO)
      ? (useHttp11 ? 'HTTP/1.1' : 'direct')
      : url.replace(/https?:\/\//, '').split('/')[0];
    if (attempt > 1) console.log(`  Retry ${attempt}/4 (${label})...`);

    const cloneResult = spawnSync('git', gitArgs, { stdio: 'inherit', timeout: 180000 });
    if (cloneResult.status === 0) {
      cloned = true;
      break;
    }
    if (attempt < 4) {
      const wait = attempt * 2;
      console.log(`  Clone failed, waiting ${wait}s...`);
      sleepSync(wait);
    }
  }
  if (!cloned) {
    throw new Error('git clone failed after 4 attempts. Set GITHUB_PROXY=https://your-proxy to use a custom mirror.');
  }

  const outputPath = path.join(installDir, BINARY_NAME);
  console.log('  Building gprobe (this may take 1-2 minutes)...');
  const buildResult = spawnSync('go', [
    'build', '-o', outputPath, './cmd/gprobe',
  ], { cwd: srcDir, stdio: 'inherit' });

  if (buildResult.status !== 0) {
    throw new Error('go build failed');
  }

  // Cleanup source
  fs.rmSync(srcDir, { recursive: true, force: true });

  if (!isWin) {
    fs.chmodSync(outputPath, 0o755);
  }
  console.log('  Build complete');

  return tag;
}

module.exports = { buildFromSource };
