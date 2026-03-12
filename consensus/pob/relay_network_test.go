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
	"testing"

	"github.com/probechain/go-probe/common"
)

func TestRelayInfoAddRemoveAgent(t *testing.T) {
	relay := NewRelayInfo()
	agent1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	agent2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	relay.AddAgent(agent1)
	relay.AddAgent(agent2)

	if relay.AgentCount() != 2 {
		t.Errorf("agent count: got %d, want 2", relay.AgentCount())
	}
	if !relay.HasAgent(agent1) {
		t.Error("should have agent1")
	}

	// Add duplicate — should not increase count
	relay.AddAgent(agent1)
	if relay.AgentCount() != 2 {
		t.Errorf("duplicate add should not increase count: got %d", relay.AgentCount())
	}

	relay.RemoveAgent(agent1)
	if relay.AgentCount() != 1 {
		t.Errorf("after remove: got %d, want 1", relay.AgentCount())
	}
	if relay.HasAgent(agent1) {
		t.Error("agent1 should be removed")
	}
	if !relay.HasAgent(agent2) {
		t.Error("agent2 should still be present")
	}
}

func TestCalcAgentManagementScoreNoAgents(t *testing.T) {
	score := CalcAgentManagementScore(nil)
	if score != defaultInitialScore {
		t.Errorf("nil relay should return default score: got %d, want %d", score, defaultInitialScore)
	}

	relay := NewRelayInfo()
	score = CalcAgentManagementScore(relay)
	if score != defaultInitialScore {
		t.Errorf("empty relay should return default score: got %d, want %d", score, defaultInitialScore)
	}
}

func TestCalcAgentManagementScorePerfect(t *testing.T) {
	relay := &RelayInfo{
		ManagedAgents:            make([]common.Address, 10),
		AgentHeartbeatsRelayed:   200, // 10 agents × 10 heartbeats = 100 expected, 200 exceeds
		AgentAttestationsRelayed: 100,
		AgentTasksAssigned:       50,
		AgentTasksCompleted:      50, // 100% completion
		InvalidAggregations:      0,
		AgentDrops:               0,
	}

	score := CalcAgentManagementScore(relay)
	if score != maxScore {
		t.Errorf("perfect relay should get max score: got %d, want %d", score, maxScore)
	}
}

func TestCalcAgentManagementScorePoor(t *testing.T) {
	relay := &RelayInfo{
		ManagedAgents:            make([]common.Address, 10),
		AgentHeartbeatsRelayed:   10, // 10 expected 100, only relayed 10
		AgentAttestationsRelayed: 10,
		AgentTasksAssigned:       50,
		AgentTasksCompleted:      5, // Only 10% completion
		InvalidAggregations:      5, // 25% error rate
		AgentDrops:               5, // 33% drop rate
	}

	score := CalcAgentManagementScore(relay)
	if score >= 5000 {
		t.Errorf("poor relay should get low score: got %d", score)
	}
}

func TestAssignRelay(t *testing.T) {
	relays := []common.Address{
		common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		common.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		common.HexToAddress("0xcccccccccccccccccccccccccccccccccccccccc"),
	}

	agent := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaab")
	relay, ok := AssignRelay(agent, relays)
	if !ok {
		t.Fatal("should find a relay")
	}
	// Agent 0xaa...ab is closest to relay 0xaa...aa (XOR distance = 1 in last byte)
	if relay != relays[0] {
		t.Errorf("should be assigned to closest relay 0xaa..aa, got %s", relay.Hex())
	}
}

func TestAssignRelayNoRelays(t *testing.T) {
	agent := common.HexToAddress("0x1111111111111111111111111111111111111111")
	_, ok := AssignRelay(agent, nil)
	if ok {
		t.Error("should return false with no relays")
	}
}

func TestRebalanceRelays(t *testing.T) {
	relays := []common.Address{
		common.HexToAddress("0x1000000000000000000000000000000000000000"),
		common.HexToAddress("0x8000000000000000000000000000000000000000"),
	}
	agents := []common.Address{
		common.HexToAddress("0x1100000000000000000000000000000000000000"),
		common.HexToAddress("0x1200000000000000000000000000000000000000"),
		common.HexToAddress("0x8100000000000000000000000000000000000000"),
		common.HexToAddress("0x8200000000000000000000000000000000000000"),
		common.HexToAddress("0x8300000000000000000000000000000000000000"),
	}

	result := RebalanceRelays(agents, relays, nil)

	totalAssigned := 0
	for _, info := range result {
		totalAssigned += info.AgentCount()
	}
	if totalAssigned != len(agents) {
		t.Errorf("total assigned: got %d, want %d", totalAssigned, len(agents))
	}

	// Agents 0x11, 0x12 should be closer to relay 0x10
	relay1 := result[relays[0]]
	if !relay1.HasAgent(agents[0]) || !relay1.HasAgent(agents[1]) {
		t.Error("agents 0x11, 0x12 should be assigned to relay 0x10")
	}

	// Agents 0x81, 0x82, 0x83 should be closer to relay 0x80
	relay2 := result[relays[1]]
	if !relay2.HasAgent(agents[2]) || !relay2.HasAgent(agents[3]) || !relay2.HasAgent(agents[4]) {
		t.Error("agents 0x81, 0x82, 0x83 should be assigned to relay 0x80")
	}
}

func TestRebalanceRelaysPreservesStats(t *testing.T) {
	relay := common.HexToAddress("0x1000000000000000000000000000000000000000")
	existing := map[common.Address]*RelayInfo{
		relay: {
			ManagedAgents:          []common.Address{},
			AgentHeartbeatsRelayed: 500,
			AgentTasksAssigned:     100,
			AgentTasksCompleted:    90,
		},
	}

	agents := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
	}

	result := RebalanceRelays(agents, []common.Address{relay}, existing)
	info := result[relay]

	if info.AgentHeartbeatsRelayed != 500 {
		t.Errorf("should preserve heartbeats: got %d", info.AgentHeartbeatsRelayed)
	}
	if info.AgentTasksAssigned != 100 {
		t.Errorf("should preserve tasks assigned: got %d", info.AgentTasksAssigned)
	}
}

func TestXorDistance(t *testing.T) {
	a := common.HexToAddress("0xff00000000000000000000000000000000000000")
	b := common.HexToAddress("0x0100000000000000000000000000000000000000")
	dist := xorDistance(a, b)
	if dist[0] != 0xfe {
		t.Errorf("XOR first byte: got 0x%02x, want 0xfe", dist[0])
	}

	// Distance to self should be zero
	self := xorDistance(a, a)
	for i, v := range self {
		if v != 0 {
			t.Errorf("self XOR byte %d: got %d, want 0", i, v)
			break
		}
	}
}
