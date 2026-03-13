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

package pob

import (
	"github.com/probechain/go-probe/common"
)

// ---------------------------------------------------------------------------
// Relay Network (Phase 3.3)
// ---------------------------------------------------------------------------
// SmartLight nodes act as relay nodes in the PoB architecture. Each relay
// manages 100-1000 agents: aggregating heartbeats, forwarding task assignments,
// and batching attestations before sending them to validators.
//
// Relay scoring includes an "agent management quality" dimension that measures
// how well the relay manages its assigned agents.

const (
	// MinAgentsPerRelay is the minimum agents a relay should manage to get full scores.
	MinAgentsPerRelay = 10

	// MaxAgentsPerRelay is the maximum agents a single relay can manage.
	MaxAgentsPerRelay = 1000

	// DefaultAgentsPerRelay is the target number of agents per relay.
	DefaultAgentsPerRelay = 100
)

// RelayInfo tracks a SmartLight node's relay responsibilities.
type RelayInfo struct {
	// ManagedAgents are the agent addresses this relay is responsible for.
	ManagedAgents []common.Address `json:"managedAgents"`

	// AgentHeartbeatsRelayed is the total heartbeats aggregated and forwarded.
	AgentHeartbeatsRelayed uint64 `json:"agentHeartbeatsRelayed"`

	// AgentAttestationsRelayed is the total attestations aggregated and forwarded.
	AgentAttestationsRelayed uint64 `json:"agentAttestationsRelayed"`

	// AgentTasksAssigned is the total tasks dispatched to managed agents.
	AgentTasksAssigned uint64 `json:"agentTasksAssigned"`

	// AgentTasksCompleted is the total tasks completed by managed agents.
	AgentTasksCompleted uint64 `json:"agentTasksCompleted"`

	// InvalidAggregations counts aggregation errors (bad bloom filters, invalid sigs).
	InvalidAggregations uint64 `json:"invalidAggregations"`

	// AgentDrops counts agents that went offline under this relay's management.
	AgentDrops uint64 `json:"agentDrops"`
}

// NewRelayInfo creates a new empty relay info.
func NewRelayInfo() *RelayInfo {
	return &RelayInfo{
		ManagedAgents: make([]common.Address, 0),
	}
}

// AgentCount returns the number of agents managed by this relay.
func (r *RelayInfo) AgentCount() int {
	return len(r.ManagedAgents)
}

// AddAgent assigns an agent to this relay.
func (r *RelayInfo) AddAgent(agent common.Address) {
	for _, a := range r.ManagedAgents {
		if a == agent {
			return // already managed
		}
	}
	r.ManagedAgents = append(r.ManagedAgents, agent)
}

// RemoveAgent removes an agent from this relay.
func (r *RelayInfo) RemoveAgent(agent common.Address) {
	for i, a := range r.ManagedAgents {
		if a == agent {
			r.ManagedAgents = append(r.ManagedAgents[:i], r.ManagedAgents[i+1:]...)
			return
		}
	}
}

// HasAgent checks if an agent is managed by this relay.
func (r *RelayInfo) HasAgent(agent common.Address) bool {
	for _, a := range r.ManagedAgents {
		if a == agent {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Agent Management Quality Scoring
// ---------------------------------------------------------------------------
// This extends SmartLight scoring with a 6th dimension: AgentManagement.
// The new weight distribution for SmartLight relays with agents:
//   liveness 25%, correctness 15%, cooperation 20%, consistency 10%,
//   signalSovereignty 10%, agentManagement 20%.

// CalcAgentManagementScore evaluates how well a relay manages its agents.
// Returns a score in [0, 10000] basis points.
func CalcAgentManagementScore(relay *RelayInfo) uint64 {
	if relay == nil || len(relay.ManagedAgents) == 0 {
		// No agents managed — neutral score (not penalized, not rewarded)
		return defaultInitialScore
	}

	score := uint64(0)

	// 1. Heartbeat relay rate (40% of dimension)
	//    How reliably the relay forwards agent heartbeats.
	heartbeatScore := uint64(maxScore)
	totalExpected := uint64(len(relay.ManagedAgents)) * 10 // ~10 heartbeats per agent per epoch
	if totalExpected > 0 && relay.AgentHeartbeatsRelayed < totalExpected {
		heartbeatScore = relay.AgentHeartbeatsRelayed * maxScore / totalExpected
	}

	// 2. Task completion rate (30% of dimension)
	//    How well agents under this relay complete tasks.
	taskScore := uint64(maxScore)
	if relay.AgentTasksAssigned > 0 {
		taskScore = relay.AgentTasksCompleted * maxScore / relay.AgentTasksAssigned
	}

	// 3. Aggregation accuracy (20% of dimension)
	//    How many aggregation errors the relay has committed.
	aggScore := uint64(maxScore)
	totalAggregations := relay.AgentHeartbeatsRelayed + relay.AgentAttestationsRelayed
	if totalAggregations > 0 && relay.InvalidAggregations > 0 {
		errorRate := relay.InvalidAggregations * maxScore / totalAggregations
		if errorRate > maxScore {
			aggScore = 0
		} else {
			aggScore = maxScore - errorRate
		}
	}

	// 4. Agent retention (10% of dimension)
	//    How many agents dropped under this relay's management.
	retentionScore := uint64(maxScore)
	totalManaged := uint64(len(relay.ManagedAgents)) + relay.AgentDrops
	if totalManaged > 0 && relay.AgentDrops > 0 {
		dropRate := relay.AgentDrops * maxScore / totalManaged
		if dropRate > maxScore {
			retentionScore = 0
		} else {
			retentionScore = maxScore - dropRate
		}
	}

	// Weighted combination: 40% heartbeat + 30% task + 20% aggregation + 10% retention
	score = (heartbeatScore*40 + taskScore*30 + aggScore*20 + retentionScore*10) / 100

	if score > maxScore {
		score = maxScore
	}
	return score
}

// ---------------------------------------------------------------------------
// Relay Assignment
// ---------------------------------------------------------------------------
// Agents are assigned to relays based on address proximity (XOR distance).
// This ensures deterministic, balanced relay assignment without coordination.

// AssignRelay finds the closest relay for an agent based on XOR distance.
// Returns the relay address and whether one was found.
func AssignRelay(agentAddr common.Address, relays []common.Address) (common.Address, bool) {
	if len(relays) == 0 {
		return common.Address{}, false
	}

	bestRelay := relays[0]
	bestDist := xorDistance(agentAddr, relays[0])

	for _, relay := range relays[1:] {
		dist := xorDistance(agentAddr, relay)
		if compareDist(dist, bestDist) < 0 {
			bestDist = dist
			bestRelay = relay
		}
	}
	return bestRelay, true
}

// xorDistance computes the XOR distance between two addresses.
func xorDistance(a, b common.Address) [20]byte {
	var dist [20]byte
	for i := 0; i < 20; i++ {
		dist[i] = a[i] ^ b[i]
	}
	return dist
}

// compareDist compares two XOR distances. Returns <0, 0, >0.
func compareDist(a, b [20]byte) int {
	for i := 0; i < 20; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

// RebalanceRelays assigns agents to relays, ensuring no relay is overloaded.
// Returns a map of relay → RelayInfo with assigned agents.
func RebalanceRelays(
	agents []common.Address,
	relays []common.Address,
	existing map[common.Address]*RelayInfo,
) map[common.Address]*RelayInfo {
	result := make(map[common.Address]*RelayInfo, len(relays))
	for _, relay := range relays {
		if info, ok := existing[relay]; ok {
			// Preserve stats, clear agent list for reassignment
			result[relay] = &RelayInfo{
				ManagedAgents:            make([]common.Address, 0),
				AgentHeartbeatsRelayed:   info.AgentHeartbeatsRelayed,
				AgentAttestationsRelayed: info.AgentAttestationsRelayed,
				AgentTasksAssigned:       info.AgentTasksAssigned,
				AgentTasksCompleted:      info.AgentTasksCompleted,
				InvalidAggregations:      info.InvalidAggregations,
				AgentDrops:               info.AgentDrops,
			}
		} else {
			result[relay] = NewRelayInfo()
		}
	}

	if len(relays) == 0 {
		return result
	}

	// Assign each agent to closest relay, respecting MaxAgentsPerRelay
	for _, agent := range agents {
		assigned := false
		// Sort relays by XOR distance to this agent
		type relayDist struct {
			addr common.Address
			dist [20]byte
		}
		sorted := make([]relayDist, len(relays))
		for i, r := range relays {
			sorted[i] = relayDist{r, xorDistance(agent, r)}
		}
		// Simple insertion sort (relay count is small)
		for i := 1; i < len(sorted); i++ {
			for j := i; j > 0 && compareDist(sorted[j].dist, sorted[j-1].dist) < 0; j-- {
				sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			}
		}

		// Assign to closest relay that isn't full
		for _, rd := range sorted {
			info := result[rd.addr]
			if info.AgentCount() < MaxAgentsPerRelay {
				info.AddAgent(agent)
				assigned = true
				break
			}
		}
		// If all relays are full, assign to closest anyway (overflow)
		if !assigned {
			result[sorted[0].addr].AddAgent(agent)
		}
	}

	return result
}
