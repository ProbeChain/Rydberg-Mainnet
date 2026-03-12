// Copyright 2024 The ProbeChain Authors
// This file is part of the ProbeChain.
//
// The ProbeChain is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The ProbeChain is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the ProbeChain. If not, see <http://www.gnu.org/licenses/>.

// Package dashboard provides a lightweight web dashboard for headless PoB agent nodes.
// It serves a status page on localhost:8547/dashboard with real-time agent statistics.
package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/probechain/go-probe/log"
)

// AgentStatus holds the current agent node status for the dashboard.
type AgentStatus struct {
	AgentID       string `json:"agentId"`
	Address       string `json:"address"`
	BlockNumber   uint64 `json:"blockNumber"`
	PeerCount     int    `json:"peerCount"`
	Syncing       bool   `json:"syncing"`
	PowerMode     string `json:"powerMode"`
	Uptime        string `json:"uptime"`

	// Behavior score
	ScoreTotal          uint64 `json:"scoreTotal"`
	ScoreResponsiveness uint64 `json:"scoreResponsiveness"`
	ScoreAccuracy       uint64 `json:"scoreAccuracy"`
	ScoreReliability    uint64 `json:"scoreReliability"`
	ScoreCooperation    uint64 `json:"scoreCooperation"`
	ScoreEconomy        uint64 `json:"scoreEconomy"`
	ScoreSovereignty    uint64 `json:"scoreSovereignty"`

	// Task stats
	TasksDone      uint64 `json:"tasksDone"`
	TasksSucceeded uint64 `json:"tasksSucceeded"`
	TasksActive    int32  `json:"tasksActive"`

	// Reward stats
	TotalRewards string `json:"totalRewards"`
	EpochRewards string `json:"epochRewards"`

	// Network stats
	AgentCount     int    `json:"agentCount"`
	RelayCount     int    `json:"relayCount"`
	ValidatorCount int    `json:"validatorCount"`
}

// StatusProvider is the interface that the dashboard uses to get current status.
type StatusProvider interface {
	GetAgentStatus() *AgentStatus
}

// Dashboard serves the agent web dashboard.
type Dashboard struct {
	provider StatusProvider
	server   *http.Server
	startTime time.Time
	mu       sync.RWMutex
}

// New creates a new agent dashboard.
func New(provider StatusProvider, addr string) *Dashboard {
	d := &Dashboard{
		provider:  provider,
		startTime: time.Now(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/dashboard", d.handleDashboard)
	mux.HandleFunc("/dashboard/api/status", d.handleAPIStatus)

	d.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return d
}

// Start begins serving the dashboard.
func (d *Dashboard) Start() error {
	log.Info("Agent dashboard starting", "addr", d.server.Addr)
	go func() {
		if err := d.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Warn("Agent dashboard failed", "err", err)
		}
	}()
	return nil
}

// Stop shuts down the dashboard server.
func (d *Dashboard) Stop() error {
	return d.server.Close()
}

// handleAPIStatus returns the current agent status as JSON.
func (d *Dashboard) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := d.provider.GetAgentStatus()
	if status == nil {
		http.Error(w, "Status unavailable", http.StatusServiceUnavailable)
		return
	}
	status.Uptime = time.Since(d.startTime).Truncate(time.Second).String()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:*")
	json.NewEncoder(w).Encode(status)
}

// handleDashboard serves the HTML dashboard page.
func (d *Dashboard) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, dashboardHTML)
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>ProbeChain Agent Dashboard</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, monospace;
         background: #0a0e17; color: #e0e6f0; padding: 24px; }
  h1 { font-size: 1.4em; color: #7eb8ff; margin-bottom: 8px; }
  .subtitle { color: #6b7b8d; font-size: 0.85em; margin-bottom: 24px; }
  .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 16px; }
  .card { background: #131a2b; border: 1px solid #1e2a3f; border-radius: 8px; padding: 16px; }
  .card h2 { font-size: 0.85em; color: #6b7b8d; text-transform: uppercase; letter-spacing: 1px; margin-bottom: 12px; }
  .stat { display: flex; justify-content: space-between; padding: 4px 0; border-bottom: 1px solid #1a2235; }
  .stat:last-child { border-bottom: none; }
  .stat .label { color: #8899aa; }
  .stat .value { color: #e0e6f0; font-weight: 600; }
  .score-bar { height: 6px; background: #1a2235; border-radius: 3px; margin-top: 4px; }
  .score-bar .fill { height: 100%; border-radius: 3px; transition: width 0.3s; }
  .score-bar .fill.high { background: #4ade80; }
  .score-bar .fill.mid { background: #fbbf24; }
  .score-bar .fill.low { background: #f87171; }
  .badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.75em; font-weight: 600; }
  .badge.online { background: #064e3b; color: #4ade80; }
  .badge.syncing { background: #422006; color: #fbbf24; }
  .refresh { color: #6b7b8d; font-size: 0.75em; text-align: right; margin-top: 16px; }
</style>
</head>
<body>
<h1>ProbeChain Agent Node</h1>
<p class="subtitle">PoB Consensus Dashboard</p>
<div class="grid" id="dashboard">
  <div class="card">
    <h2>Node Status</h2>
    <div id="node-status">Loading...</div>
  </div>
  <div class="card">
    <h2>Behavior Score</h2>
    <div id="behavior-score">Loading...</div>
  </div>
  <div class="card">
    <h2>Task Statistics</h2>
    <div id="task-stats">Loading...</div>
  </div>
  <div class="card">
    <h2>Network</h2>
    <div id="network-stats">Loading...</div>
  </div>
</div>
<p class="refresh" id="refresh-info"></p>
<script>
function scoreClass(v) { return v >= 7000 ? 'high' : v >= 4000 ? 'mid' : 'low'; }
function scorePct(v) { return (v / 100).toFixed(1) + '%'; }
function scoreBar(label, value) {
  return '<div class="stat"><span class="label">' + label + '</span><span class="value">' + scorePct(value) + '</span></div>' +
    '<div class="score-bar"><div class="fill ' + scoreClass(value) + '" style="width:' + (value/100) + '%"></div></div>';
}
function stat(label, value) {
  return '<div class="stat"><span class="label">' + label + '</span><span class="value">' + value + '</span></div>';
}
function refresh() {
  fetch('/dashboard/api/status').then(r => r.json()).then(s => {
    var syncBadge = s.syncing ? '<span class="badge syncing">Syncing</span>' : '<span class="badge online">Online</span>';
    document.getElementById('node-status').innerHTML =
      stat('Status', syncBadge) +
      stat('Block', '#' + (s.blockNumber || 0).toLocaleString()) +
      stat('Peers', s.peerCount) +
      stat('Mode', s.powerMode || 'Agent') +
      stat('Uptime', s.uptime || '-') +
      stat('Address', (s.address || '-').slice(0,10) + '...');
    document.getElementById('behavior-score').innerHTML =
      stat('Total', scorePct(s.scoreTotal || 0)) +
      scoreBar('Responsiveness', s.scoreResponsiveness || 0) +
      scoreBar('Accuracy', s.scoreAccuracy || 0) +
      scoreBar('Reliability', s.scoreReliability || 0) +
      scoreBar('Cooperation', s.scoreCooperation || 0) +
      scoreBar('Economy', s.scoreEconomy || 0) +
      scoreBar('Sovereignty', s.scoreSovereignty || 0);
    document.getElementById('task-stats').innerHTML =
      stat('Completed', s.tasksDone || 0) +
      stat('Succeeded', s.tasksSucceeded || 0) +
      stat('Active', s.tasksActive || 0) +
      stat('Success Rate', s.tasksDone > 0 ? ((s.tasksSucceeded / s.tasksDone * 100).toFixed(1) + '%') : '-');
    document.getElementById('network-stats').innerHTML =
      stat('Agents', (s.agentCount || 0).toLocaleString()) +
      stat('Relays', (s.relayCount || 0).toLocaleString()) +
      stat('Validators', s.validatorCount || 0) +
      stat('Total Rewards', s.totalRewards || '0') +
      stat('Epoch Rewards', s.epochRewards || '0');
    document.getElementById('refresh-info').textContent = 'Last updated: ' + new Date().toLocaleTimeString();
  }).catch(function(e) {
    document.getElementById('refresh-info').textContent = 'Error: ' + e.message;
  });
}
refresh();
setInterval(refresh, 3000);
</script>
</body>
</html>`
