'use strict';

const os = require('os');
const { execSync } = require('child_process');

const isWin = process.platform === 'win32';

/**
 * Detect platform and architecture.
 * Returns { os, arch, prebuilt, label }
 */
function detect() {
  const platform = process.platform;   // 'darwin', 'linux', 'win32'
  const arch = process.arch;           // 'arm64', 'x64', 'ia32'

  // Normalise arch to uname-style names
  const archMap = { x64: 'x86_64', arm64: 'arm64', ia32: 'i386' };
  const normArch = archMap[arch] || arch;

  // Normalise OS
  const osMap = { darwin: 'Darwin', linux: 'Linux', win32: 'Windows' };
  const normOS = osMap[platform] || platform;

  // Pre-built binary available for macOS arm64 and Windows x64
  const prebuilt = (platform === 'darwin' && arch === 'arm64')
                || (platform === 'win32' && arch === 'x64');

  // Friendly label
  const label = `${normOS} ${normArch}`;

  return { os: normOS, platform, arch: normArch, prebuilt, label };
}

/**
 * Check whether a command exists on the system.
 */
function hasCommand(cmd) {
  try {
    const check = isWin ? `where ${cmd}` : `command -v ${cmd}`;
    execSync(check, { stdio: 'ignore' });
    return true;
  } catch {
    return false;
  }
}

/**
 * Validate that required tools are available for the current platform.
 * Returns { ok: true } or { ok: false, missing: [...] }
 */
function checkRequirements(info) {
  if (info.prebuilt) {
    // Pre-built path only needs basic tools (all ship with macOS / Windows)
    return { ok: true };
  }

  // Source build path
  const required = ['git', 'go'];
  const missing = required.filter(c => !hasCommand(c));
  if (missing.length > 0) {
    return { ok: false, missing };
  }
  return { ok: true };
}

module.exports = { detect, hasCommand, checkRequirements };
