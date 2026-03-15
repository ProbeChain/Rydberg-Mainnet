#!/usr/bin/env node
'use strict';

const { cmdInstall, cmdStatus, cmdStop, cmdStart, cmdLogs } = require('../lib/installer');

const COMMANDS = {
  status: cmdStatus,
  start:  cmdStart,
  stop:   cmdStop,
  logs:   cmdLogs,
};

async function main() {
  const arg = process.argv[2];

  if (arg === '--help' || arg === '-h') {
    console.log(`
  rydberg-agent-node — ProbeChain Rydberg Testnet Agent Node

  Usage:
    npx rydberg-agent-node           Install and start an Agent node
    npx rydberg-agent-node status    Show node status
    npx rydberg-agent-node start     Start the node
    npx rydberg-agent-node stop      Stop the node
    npx rydberg-agent-node logs      Show recent logs

  More info: https://probechain.org
`);
    return;
  }

  if (arg === '--version' || arg === '-v') {
    const pkg = require('../package.json');
    console.log(`rydberg-agent-node v${pkg.version}`);
    return;
  }

  const handler = COMMANDS[arg] || cmdInstall;

  try {
    await handler();
  } catch (err) {
    console.error(`\x1b[31m[ERROR]\x1b[0m ${err.message}`);
    process.exit(1);
  }
}

main();
