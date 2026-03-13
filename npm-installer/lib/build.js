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

  console.log(`  Cloning ${REPO} @ ${tag}...`);
  const cloneResult = spawnSync('git', [
    'clone', '--branch', tag, '--depth', '1',
    `https://github.com/${REPO}.git`, srcDir,
  ], { stdio: 'inherit' });

  if (cloneResult.status !== 0) {
    throw new Error('git clone failed');
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
