#!/usr/bin/env node
// probechain-agent — One-command PoB agent node for ProbeChain.
//
// Usage:
//   npx probechain-agent                    # Start with defaults
//   npx probechain-agent --testnet          # Join testnet
//   npx probechain-agent --datadir ./data   # Custom data directory
//   npx probechain-agent --rpc              # Enable local RPC (localhost:8547)

"use strict";

const { spawn, execSync } = require("child_process");
const path = require("path");
const fs = require("fs");
const os = require("os");
const crypto = require("crypto");

const AGENT_VERSION = "0.1.0";
const DEFAULT_DATADIR = path.join(os.homedir(), ".probechain-agent");
const DEFAULT_RPC_PORT = 8547;
const DEFAULT_CACHE = 32; // MB — minimal for agent mode

function findGprobe() {
  // 1. Check in our own bin directory
  const localBin = path.join(__dirname, os.platform() === "win32" ? "gprobe.exe" : "gprobe");
  if (fs.existsSync(localBin)) return localBin;

  // 2. Check in PATH
  try {
    const which = os.platform() === "win32" ? "where" : "which";
    const result = execSync(`${which} gprobe`, { encoding: "utf8" }).trim();
    if (result) return result.split("\n")[0];
  } catch (_) {}

  // 3. Check go-probe build directory
  const repoBin = path.join(__dirname, "..", "..", "build", "bin", "gprobe");
  if (fs.existsSync(repoBin)) return repoBin;

  return null;
}

function parseArgs(args) {
  const opts = {
    datadir: DEFAULT_DATADIR,
    testnet: false,
    rpc: false,
    rpcPort: DEFAULT_RPC_PORT,
    cache: DEFAULT_CACHE,
    maxpeers: 25,
    verbosity: 2, // Minimal logging for agent mode
    bootnodes: null,
  };

  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case "--datadir":
        opts.datadir = args[++i];
        break;
      case "--testnet":
        opts.testnet = true;
        break;
      case "--rpc":
        opts.rpc = true;
        break;
      case "--rpc-port":
        opts.rpcPort = parseInt(args[++i], 10);
        break;
      case "--cache":
        opts.cache = parseInt(args[++i], 10);
        break;
      case "--maxpeers":
        opts.maxpeers = parseInt(args[++i], 10);
        break;
      case "--verbosity":
        opts.verbosity = parseInt(args[++i], 10);
        break;
      case "--bootnodes":
        opts.bootnodes = args[++i];
        break;
      case "--help":
      case "-h":
        printHelp();
        process.exit(0);
      case "--version":
      case "-v":
        console.log(`probechain-agent v${AGENT_VERSION}`);
        process.exit(0);
      default:
        if (args[i].startsWith("-")) {
          console.error(`Unknown option: ${args[i]}`);
          process.exit(1);
        }
    }
  }
  return opts;
}

function printHelp() {
  console.log(`
probechain-agent v${AGENT_VERSION} — PoB Agent Node for ProbeChain

Usage: npx probechain-agent [options]

Options:
  --datadir <path>     Data directory (default: ~/.probechain-agent)
  --testnet            Join testnet instead of mainnet
  --rpc                Enable JSON-RPC server on localhost
  --rpc-port <port>    RPC server port (default: ${DEFAULT_RPC_PORT})
  --cache <mb>         Database cache size in MB (default: ${DEFAULT_CACHE})
  --maxpeers <n>       Maximum peer connections (default: 25)
  --verbosity <0-5>    Log verbosity (default: 2)
  --bootnodes <enodes> Comma-separated bootstrap nodes
  --help, -h           Show this help
  --version, -v        Show version

Agent mode automatically:
  - Runs LES light sync (~30MB RAM)
  - Enables PoB agent consensus participation
  - Generates Ed25519 keypair on first run
  - Disables mining/validator logic
  - Optimizes for maximum task throughput
`);
}

function ensureDatadir(datadir) {
  if (!fs.existsSync(datadir)) {
    fs.mkdirSync(datadir, { recursive: true });
    console.log(`Created data directory: ${datadir}`);
  }

  // Generate agent identity on first run
  const idFile = path.join(datadir, "agent-id.json");
  if (!fs.existsSync(idFile)) {
    const agentID = crypto.randomBytes(32).toString("hex");
    const identity = {
      agentID: `0x${agentID}`,
      createdAt: new Date().toISOString(),
      version: AGENT_VERSION,
    };
    fs.writeFileSync(idFile, JSON.stringify(identity, null, 2));
    console.log(`Generated agent identity: 0x${agentID.slice(0, 16)}...`);
  }
}

function buildGprobeArgs(opts) {
  const args = [
    "--agentmode",
    "--syncmode", "light",
    "--cache", String(opts.cache),
    "--maxpeers", String(opts.maxpeers),
    "--datadir", opts.datadir,
    "--verbosity", String(opts.verbosity),
  ];

  if (opts.rpc) {
    args.push(
      "--http",
      "--http.addr", "127.0.0.1",
      "--http.port", String(opts.rpcPort),
      "--http.api", "probe,net,web3,pob",
      "--http.corsdomain", "http://localhost:*"
    );
  }

  if (opts.testnet) {
    args.push("--networkid", "142858"); // Testnet network ID
  }

  if (opts.bootnodes) {
    args.push("--bootnodes", opts.bootnodes);
  }

  return args;
}

function main() {
  const opts = parseArgs(process.argv.slice(2));
  const gprobePath = findGprobe();

  if (!gprobePath) {
    console.error("Error: gprobe binary not found.");
    console.error("");
    console.error("Install options:");
    console.error("  1. Build from source: cd go-probe && make gprobe");
    console.error("  2. Re-run: npm install probechain-agent (to retry download)");
    console.error(
      "  3. Place gprobe binary in: " + path.join(__dirname, "gprobe")
    );
    process.exit(1);
  }

  ensureDatadir(opts.datadir);

  const args = buildGprobeArgs(opts);

  console.log("");
  console.log("  ProbeChain Agent Node v" + AGENT_VERSION);
  console.log("  ──────────────────────────────────");
  console.log(`  Mode:     PoB Agent (light sync)`);
  console.log(`  Data:     ${opts.datadir}`);
  console.log(`  Cache:    ${opts.cache} MB`);
  console.log(`  Peers:    ${opts.maxpeers}`);
  console.log(`  Network:  ${opts.testnet ? "Testnet" : "Mainnet"}`);
  if (opts.rpc) {
    console.log(`  RPC:      http://127.0.0.1:${opts.rpcPort}`);
  }
  console.log("");

  const child = spawn(gprobePath, args, {
    stdio: "inherit",
    env: { ...process.env, GPROBE_AGENT_MODE: "1" },
  });

  child.on("error", (err) => {
    console.error(`Failed to start gprobe: ${err.message}`);
    process.exit(1);
  });

  child.on("exit", (code, signal) => {
    if (signal) {
      console.log(`\nAgent node terminated by signal: ${signal}`);
    } else if (code !== 0) {
      console.error(`Agent node exited with code: ${code}`);
    } else {
      console.log("\nAgent node stopped.");
    }
    process.exit(code || 0);
  });

  // Handle Ctrl+C gracefully
  process.on("SIGINT", () => {
    console.log("\nShutting down agent node...");
    child.kill("SIGINT");
  });
  process.on("SIGTERM", () => {
    child.kill("SIGTERM");
  });
}

main();
