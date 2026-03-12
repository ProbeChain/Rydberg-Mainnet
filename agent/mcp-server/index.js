#!/usr/bin/env node
// ProbeChain MCP Server — Model Context Protocol server for PoB agents.
//
// Exposes ProbeChain agent capabilities as MCP tools for AI models
// (Claude, OpenClaw, etc.) to interact with the ProbeChain network.
//
// Usage:
//   node index.js                           # Default: connect to localhost:8547
//   node index.js --rpc http://host:port    # Custom RPC endpoint

"use strict";

const http = require("http");
const net = require("net");

const DEFAULT_RPC = "http://127.0.0.1:8547";
const MCP_VERSION = "2024-11-05";

class ProbeChainMCPServer {
  constructor(rpcUrl) {
    this.rpcUrl = rpcUrl;
    this.rpcId = 1;
  }

  // JSON-RPC call to the local gprobe node
  async rpcCall(method, params = []) {
    const url = new URL(this.rpcUrl);
    const body = JSON.stringify({
      jsonrpc: "2.0",
      id: this.rpcId++,
      method,
      params,
    });

    return new Promise((resolve, reject) => {
      const req = http.request(
        {
          hostname: url.hostname,
          port: url.port,
          path: url.pathname,
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Content-Length": Buffer.byteLength(body),
          },
        },
        (res) => {
          let data = "";
          res.on("data", (chunk) => (data += chunk));
          res.on("end", () => {
            try {
              const result = JSON.parse(data);
              if (result.error) {
                reject(new Error(result.error.message));
              } else {
                resolve(result.result);
              }
            } catch (e) {
              reject(new Error(`Invalid JSON response: ${data.slice(0, 100)}`));
            }
          });
        }
      );
      req.on("error", reject);
      req.write(body);
      req.end();
    });
  }

  // MCP tool definitions
  getToolDefinitions() {
    return [
      {
        name: "probechain_attest",
        description:
          "Submit a cross-agent attestation to the ProbeChain network. " +
          "Attests to a claim with a confidence level, contributing to the agent's PoB behavior score.",
        inputSchema: {
          type: "object",
          properties: {
            claim: {
              type: "string",
              description: "The attestation claim data (e.g., verified computation result)",
            },
            confidence: {
              type: "number",
              minimum: 0,
              maximum: 1,
              description: "Confidence level (0.0-1.0) in the attestation",
            },
            target_agent: {
              type: "string",
              description: "Address of the agent being attested (optional)",
            },
          },
          required: ["claim", "confidence"],
        },
      },
      {
        name: "probechain_task",
        description:
          "Submit or query a task on the ProbeChain agent network. " +
          "Tasks are distributed to agent nodes for execution with verified results.",
        inputSchema: {
          type: "object",
          properties: {
            type: {
              type: "string",
              enum: ["verify", "compute", "query"],
              description: "Task type: verify (check data), compute (run computation), query (read state)",
            },
            data: {
              type: "string",
              description: "Task payload data",
            },
            timeout_ms: {
              type: "number",
              description: "Task timeout in milliseconds (default: 5000)",
            },
          },
          required: ["type", "data"],
        },
      },
      {
        name: "probechain_status",
        description:
          "Get the current status of the local ProbeChain agent node, " +
          "including sync state, behavior score, peer count, and task statistics.",
        inputSchema: {
          type: "object",
          properties: {},
        },
      },
      {
        name: "probechain_score",
        description:
          "Get the behavior score breakdown for the local agent node or a specified address. " +
          "Returns the 6-dimension PoB score: responsiveness, accuracy, reliability, cooperation, economy, sovereignty.",
        inputSchema: {
          type: "object",
          properties: {
            address: {
              type: "string",
              description: "Agent address to query (default: local agent)",
            },
          },
        },
      },
      {
        name: "probechain_agents",
        description:
          "List registered agent nodes on the ProbeChain network with their scores and status.",
        inputSchema: {
          type: "object",
          properties: {
            limit: {
              type: "number",
              description: "Maximum number of agents to return (default: 50)",
            },
          },
        },
      },
    ];
  }

  // Execute an MCP tool call
  async executeTool(name, args) {
    switch (name) {
      case "probechain_attest":
        return this.handleAttest(args);
      case "probechain_task":
        return this.handleTask(args);
      case "probechain_status":
        return this.handleStatus(args);
      case "probechain_score":
        return this.handleScore(args);
      case "probechain_agents":
        return this.handleAgents(args);
      default:
        throw new Error(`Unknown tool: ${name}`);
    }
  }

  async handleAttest(args) {
    const confidenceBps = Math.round(args.confidence * 10000);
    const result = await this.rpcCall("pob_submitAttestation", [
      args.claim,
      confidenceBps,
      args.target_agent || null,
    ]);
    return {
      type: "text",
      text: JSON.stringify({
        success: true,
        attestation_id: result,
        confidence_bps: confidenceBps,
      }),
    };
  }

  async handleTask(args) {
    const timeout = args.timeout_ms || 5000;
    const result = await this.rpcCall("pob_submitAgentTask", [
      args.type,
      args.data,
      timeout,
    ]);
    return {
      type: "text",
      text: JSON.stringify({
        success: true,
        task_id: result.taskId,
        result: result.result,
        gas_used: result.gasUsed,
      }),
    };
  }

  async handleStatus(args) {
    const [blockNumber, peerCount, syncing] = await Promise.all([
      this.rpcCall("probe_blockNumber"),
      this.rpcCall("net_peerCount"),
      this.rpcCall("probe_syncing"),
    ]);

    let agentScore = null;
    try {
      agentScore = await this.rpcCall("pob_getAgentScore", []);
    } catch (_) {}

    return {
      type: "text",
      text: JSON.stringify({
        block_number: parseInt(blockNumber, 16),
        peer_count: parseInt(peerCount, 16),
        syncing: syncing !== false,
        agent_score: agentScore,
        mode: "agent",
      }),
    };
  }

  async handleScore(args) {
    const address = args.address || null;
    const result = await this.rpcCall("pob_getAgentScores", [address]);
    return {
      type: "text",
      text: JSON.stringify(result),
    };
  }

  async handleAgents(args) {
    const limit = args.limit || 50;
    const result = await this.rpcCall("pob_getAgentCount", []);
    const scores = await this.rpcCall("pob_getAgentScores", []);
    return {
      type: "text",
      text: JSON.stringify({
        total_agents: result,
        agents: scores,
      }),
    };
  }

  // Handle MCP protocol messages over stdio
  handleMessage(message) {
    switch (message.method) {
      case "initialize":
        return {
          protocolVersion: MCP_VERSION,
          capabilities: { tools: {} },
          serverInfo: {
            name: "probechain-agent",
            version: "0.1.0",
          },
        };

      case "tools/list":
        return { tools: this.getToolDefinitions() };

      case "tools/call":
        return this.executeTool(
          message.params.name,
          message.params.arguments || {}
        ).then((content) => ({ content: [content] }));

      default:
        throw new Error(`Unknown method: ${message.method}`);
    }
  }

  // Start stdio-based MCP server
  startStdio() {
    let buffer = "";

    process.stdin.setEncoding("utf8");
    process.stdin.on("data", (chunk) => {
      buffer += chunk;
      const lines = buffer.split("\n");
      buffer = lines.pop(); // Keep incomplete line in buffer

      for (const line of lines) {
        if (!line.trim()) continue;
        try {
          const message = JSON.parse(line);
          const resultPromise = this.handleMessage(message);

          if (resultPromise && typeof resultPromise.then === "function") {
            resultPromise
              .then((result) => {
                this.sendResponse(message.id, result);
              })
              .catch((err) => {
                this.sendError(message.id, err.message);
              });
          } else {
            this.sendResponse(message.id, resultPromise);
          }
        } catch (err) {
          process.stderr.write(`MCP parse error: ${err.message}\n`);
        }
      }
    });

    process.stderr.write(
      `ProbeChain MCP Server started (RPC: ${this.rpcUrl})\n`
    );
  }

  sendResponse(id, result) {
    const response = { jsonrpc: "2.0", id, result };
    process.stdout.write(JSON.stringify(response) + "\n");
  }

  sendError(id, message) {
    const response = {
      jsonrpc: "2.0",
      id,
      error: { code: -32000, message },
    };
    process.stdout.write(JSON.stringify(response) + "\n");
  }
}

// Parse CLI args
let rpcUrl = DEFAULT_RPC;
for (let i = 2; i < process.argv.length; i++) {
  if (process.argv[i] === "--rpc" && process.argv[i + 1]) {
    rpcUrl = process.argv[++i];
  }
}

const server = new ProbeChainMCPServer(rpcUrl);
server.startStdio();
