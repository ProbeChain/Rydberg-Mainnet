#!/usr/bin/env node
// download-gprobe.js — Downloads the pre-built gprobe binary for the current platform.
// Called automatically via npm postinstall.

"use strict";

const fs = require("fs");
const path = require("path");
const os = require("os");
const https = require("https");
const { execSync } = require("child_process");

const VERSION = "0.1.0";
const BASE_URL = "https://releases.probechain.io/gprobe";

function getPlatformBinary() {
  const platform = os.platform();
  const arch = os.arch();

  const platformMap = {
    "darwin-x64": "gprobe-darwin-amd64",
    "darwin-arm64": "gprobe-darwin-arm64",
    "linux-x64": "gprobe-linux-amd64",
    "linux-arm64": "gprobe-linux-arm64",
    "win32-x64": "gprobe-windows-amd64.exe",
  };

  const key = `${platform}-${arch}`;
  const binary = platformMap[key];
  if (!binary) {
    throw new Error(
      `Unsupported platform: ${key}. Supported: ${Object.keys(platformMap).join(", ")}`
    );
  }
  return binary;
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https
      .get(url, (response) => {
        if (response.statusCode === 302 || response.statusCode === 301) {
          // Follow redirect
          downloadFile(response.headers.location, dest)
            .then(resolve)
            .catch(reject);
          return;
        }
        if (response.statusCode !== 200) {
          reject(new Error(`Download failed with status ${response.statusCode}`));
          return;
        }
        response.pipe(file);
        file.on("finish", () => {
          file.close(resolve);
        });
      })
      .on("error", (err) => {
        fs.unlink(dest, () => {});
        reject(err);
      });
  });
}

async function main() {
  const binDir = path.join(__dirname, "..", "bin");
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  const binaryName = getPlatformBinary();
  const destName = os.platform() === "win32" ? "gprobe.exe" : "gprobe";
  const destPath = path.join(binDir, destName);

  // Skip download if binary already exists
  if (fs.existsSync(destPath)) {
    console.log(`gprobe binary already exists at ${destPath}`);
    return;
  }

  // Try to find a locally built gprobe first
  const localBuild = path.join(__dirname, "..", "..", "build", "bin", "gprobe");
  if (fs.existsSync(localBuild)) {
    console.log(`Using locally built gprobe from ${localBuild}`);
    fs.copyFileSync(localBuild, destPath);
    fs.chmodSync(destPath, 0o755);
    return;
  }

  const url = `${BASE_URL}/v${VERSION}/${binaryName}`;
  console.log(`Downloading gprobe v${VERSION} for ${os.platform()}-${os.arch()}...`);
  console.log(`URL: ${url}`);

  try {
    await downloadFile(url, destPath);
    fs.chmodSync(destPath, 0o755);
    console.log(`gprobe downloaded successfully to ${destPath}`);
  } catch (err) {
    console.warn(`Warning: Could not download gprobe binary: ${err.message}`);
    console.warn("You can build it manually: cd go-probe && make gprobe");
    console.warn(
      "Then copy the binary to: " + destPath
    );
  }
}

main().catch((err) => {
  console.error("postinstall error:", err.message);
  // Don't fail the install — user can provide binary manually
  process.exit(0);
});
