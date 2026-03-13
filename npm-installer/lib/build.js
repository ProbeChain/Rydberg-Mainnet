'use strict';

const { execSync, spawnSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const { fetchRelease } = require('./download');

const REPO = 'ProbeChain/Rydberg-Mainnet';

/**
 * Build gprobe from source for non-macOS-arm64 platforms.
 * Clones the repo at the release tag and runs `go build`.
 * Returns the release tag.
 */
async function buildFromSource(installDir) {
  const { tag } = await fetchRelease();
  if (!tag) throw new Error('Could not determine release tag');

  const srcDir = path.join(installDir, 'src');

  // Retry git clone with HTTP/1.1 fallback (HTTP/2 framing errors common in China)
  console.log(`  Cloning ${REPO} @ ${tag}...`);
  let cloned = false;
  for (let attempt = 1; attempt <= 3; attempt++) {
    // Clean up partial clone from previous attempt
    if (fs.existsSync(srcDir)) fs.rmSync(srcDir, { recursive: true, force: true });

    const gitArgs = ['clone', '--branch', tag, '--depth', '1'];
    if (attempt >= 2) {
      gitArgs.unshift('-c', 'http.version=HTTP/1.1');
      console.log(`  Retry ${attempt}/3 (forcing HTTP/1.1)...`);
    }
    gitArgs.push(`https://github.com/${REPO}.git`, srcDir);

    const cloneResult = spawnSync('git', gitArgs, { stdio: 'inherit', timeout: 120000 });
    if (cloneResult.status === 0) {
      cloned = true;
      break;
    }
    if (attempt < 3) {
      const wait = attempt * 3;
      console.log(`  Clone failed, waiting ${wait}s before retry...`);
      spawnSync('sleep', [String(wait)]);
    }
  }
  if (!cloned) {
    throw new Error('git clone failed after 3 attempts. Check network connection to github.com');
  }

  console.log('  Building gprobe (this may take 1-2 minutes)...');
  const buildResult = spawnSync('go', [
    'build', '-o', path.join(installDir, 'gprobe'), './cmd/gprobe',
  ], { cwd: srcDir, stdio: 'inherit' });

  if (buildResult.status !== 0) {
    throw new Error('go build failed');
  }

  // Cleanup source
  fs.rmSync(srcDir, { recursive: true, force: true });

  fs.chmodSync(path.join(installDir, 'gprobe'), 0o755);
  console.log('  Build complete');

  return tag;
}

module.exports = { buildFromSource };
