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
	"math/big"
	"testing"

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/core/state"
	"github.com/probechain/go-probe/core/types"
	"github.com/probechain/go-probe/params"
	"github.com/probechain/go-probe/probedb/memorydb"
	"github.com/probechain/go-probe/core/rawdb"
)

// TestAgentScoring tests the agent behavior scoring dimensions.
func TestAgentScoring(t *testing.T) {
	agent := NewAgentScoringAgent()
	addr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	// Test with perfect history
	history := &AgentHistory{
		TasksCompleted:   100,
		TasksFailed:      0,
		AttestationsOK:   50,
		AttestationsBad:  0,
		HeartbeatsSent:   200,
		HeartbeatsMissed: 0,
		Uptime:           100000,
		LastActive:       1000,
		StakeAmount:      1e18, // 1 PROBE
		SlashCount:       0,
	}

	score := agent.EvaluateAgent(addr, history, 1000)

	if score.Total == 0 {
		t.Fatal("Expected non-zero total score for perfect history")
	}
	if score.Responsiveness != maxScore {
		t.Errorf("Expected max responsiveness, got %d", score.Responsiveness)
	}
	if score.Accuracy != maxScore {
		t.Errorf("Expected max accuracy, got %d", score.Accuracy)
	}
	if score.LastUpdate != 1000 {
		t.Errorf("Expected LastUpdate 1000, got %d", score.LastUpdate)
	}

	t.Logf("Perfect agent score: Total=%d, Resp=%d, Acc=%d, Rel=%d, Coop=%d, Eco=%d, Sov=%d",
		score.Total, score.Responsiveness, score.Accuracy, score.Reliability,
		score.Cooperation, score.Economy, score.Sovereignty)
}

// TestAgentScoringPoorHistory tests scoring with a poor history.
func TestAgentScoringPoorHistory(t *testing.T) {
	agent := NewAgentScoringAgent()
	addr := common.HexToAddress("0xdead")

	history := &AgentHistory{
		TasksCompleted:   10,
		TasksFailed:      90,
		AttestationsOK:   5,
		AttestationsBad:  45,
		HeartbeatsSent:   20,
		HeartbeatsMissed: 180,
		Uptime:           100,
		LastActive:       500,
		StakeAmount:      1e17, // 0.1 PROBE minimum
		SlashCount:       3,
	}

	score := agent.EvaluateAgent(addr, history, 500)

	if score.Total >= defaultInitialScore {
		t.Errorf("Expected score below default for poor history, got %d", score.Total)
	}
	if score.Responsiveness >= defaultInitialScore {
		t.Errorf("Expected low responsiveness for 90%% missed heartbeats, got %d", score.Responsiveness)
	}
}

// TestAgentScoringEmptyHistory tests default scoring for new agents.
func TestAgentScoringEmptyHistory(t *testing.T) {
	agent := NewAgentScoringAgent()
	addr := common.HexToAddress("0xnew")

	history := &AgentHistory{}
	score := agent.EvaluateAgent(addr, history, 100)

	// New agent with empty history should get high marks on most dimensions
	if score.Responsiveness != maxScore {
		t.Errorf("Expected max responsiveness for empty history, got %d", score.Responsiveness)
	}
	if score.Accuracy != maxScore {
		t.Errorf("Expected max accuracy for empty history, got %d", score.Accuracy)
	}
}

// TestSnapshotRegisterAgent tests agent registration in snapshots.
func TestSnapshotRegisterAgent(t *testing.T) {
	config := &params.PobConfig{Period: 15, Epoch: 30000}
	snap := newSnapshot(config, nil, 0, common.Hash{}, nil)

	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	pubKey := []byte("test-pubkey-data")

	// Register agents
	snap.RegisterAgent(addr1, pubKey, 100)
	snap.RegisterAgent(addr2, nil, 200)

	if !snap.IsAgent(addr1) {
		t.Fatal("addr1 should be a registered agent")
	}
	if !snap.IsAgent(addr2) {
		t.Fatal("addr2 should be a registered agent")
	}
	if snap.AgentCount() != 2 {
		t.Errorf("Expected 2 agents, got %d", snap.AgentCount())
	}

	// Verify pubkey stored
	if stored, ok := snap.AgentPubKeys[addr1]; !ok || len(stored) == 0 {
		t.Fatal("addr1 pubkey not stored")
	}

	// Verify initial score
	if snap.Agents[addr1].Total != defaultInitialScore {
		t.Errorf("Expected initial score %d, got %d", defaultInitialScore, snap.Agents[addr1].Total)
	}

	// Unregister
	snap.UnregisterAgent(addr1)
	if snap.IsAgent(addr1) {
		t.Fatal("addr1 should no longer be registered")
	}
	if snap.AgentCount() != 1 {
		t.Errorf("Expected 1 agent after unregister, got %d", snap.AgentCount())
	}
}

// TestSnapshotCopyAgent tests that agent data is properly deep copied.
func TestSnapshotCopyAgent(t *testing.T) {
	config := &params.PobConfig{Period: 15, Epoch: 30000}
	snap := newSnapshot(config, nil, 0, common.Hash{}, nil)

	addr := common.HexToAddress("0x3333333333333333333333333333333333333333")
	snap.RegisterAgent(addr, []byte("key"), 100)
	snap.AgentHistories[addr].TasksCompleted = 42

	// Copy
	cpy := snap.copy()

	// Modify original
	snap.AgentHistories[addr].TasksCompleted = 999

	// Copy should be unchanged
	if cpy.AgentHistories[addr].TasksCompleted != 42 {
		t.Errorf("Expected copy to have 42 tasks, got %d", cpy.AgentHistories[addr].TasksCompleted)
	}
}

// TestAgentTotalScore tests score summation.
func TestAgentTotalScore(t *testing.T) {
	config := &params.PobConfig{Period: 15, Epoch: 30000}
	snap := newSnapshot(config, nil, 0, common.Hash{}, nil)

	snap.RegisterAgent(common.HexToAddress("0x4444444444444444444444444444444444444444"), nil, 100)
	snap.RegisterAgent(common.HexToAddress("0x5555555555555555555555555555555555555555"), nil, 100)

	// Both start at defaultInitialScore (5000)
	expected := defaultInitialScore * 2
	if snap.AgentTotalScore() != expected {
		t.Errorf("Expected total score %d, got %d", expected, snap.AgentTotalScore())
	}
}

// TestAccumulateAgentRewards tests epoch-based agent reward distribution.
func TestAccumulateAgentRewards(t *testing.T) {
	config := &params.PobConfig{Period: 15, Epoch: 100} // Short epoch for testing
	snap := newSnapshot(config, nil, 99, common.Hash{}, nil)

	// Use 5 agents so top 80% = 4 agents qualify
	addr1 := common.HexToAddress("0xaaaa000000000000000000000000000000000001")
	addr2 := common.HexToAddress("0xaaaa000000000000000000000000000000000002")
	addr3 := common.HexToAddress("0xaaaa000000000000000000000000000000000003")
	addr4 := common.HexToAddress("0xaaaa000000000000000000000000000000000004")
	addr5 := common.HexToAddress("0xaaaa000000000000000000000000000000000005")
	snap.RegisterAgent(addr1, nil, 0)
	snap.RegisterAgent(addr2, nil, 0)
	snap.RegisterAgent(addr3, nil, 0)
	snap.RegisterAgent(addr4, nil, 0)
	snap.RegisterAgent(addr5, nil, 0)

	// Assign different scores
	snap.Agents[addr1].Total = 8000
	snap.Agents[addr2].Total = 6000
	snap.Agents[addr3].Total = 4000
	snap.Agents[addr4].Total = 2000
	snap.Agents[addr5].Total = 1000 // This one should be excluded (bottom 20%)

	// Create state
	db := rawdb.NewDatabase(memorydb.New())
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(db), nil)

	// At epoch boundary (block 100)
	header := &types.Header{Number: big.NewInt(100)}
	accumulateAgentRewards(stateDB, snap, header)

	bal1 := stateDB.GetBalance(addr1)
	bal2 := stateDB.GetBalance(addr2)
	bal5 := stateDB.GetBalance(addr5)

	if bal1.Sign() <= 0 {
		t.Fatal("addr1 should have received rewards")
	}
	// addr1 has higher score than addr2, so should get proportionally more
	if bal1.Cmp(bal2) <= 0 {
		t.Errorf("addr1 (score 8000) should get more than addr2 (score 6000), got %s vs %s", bal1, bal2)
	}
	// addr5 should be excluded (bottom 20%)
	if bal5.Sign() != 0 {
		t.Errorf("addr5 should be excluded from rewards (bottom 20%%), got %s", bal5)
	}

	t.Logf("Agent rewards: addr1=%s, addr2=%s, addr5=%s", bal1, bal2, bal5)
}

// TestAccumulateAgentRewardsNonEpoch tests that rewards are NOT distributed outside epoch boundaries.
func TestAccumulateAgentRewardsNonEpoch(t *testing.T) {
	config := &params.PobConfig{Period: 15, Epoch: 100}
	snap := newSnapshot(config, nil, 49, common.Hash{}, nil)

	addr := common.HexToAddress("0xbbbb000000000000000000000000000000000001")
	snap.RegisterAgent(addr, nil, 0)

	db := rawdb.NewDatabase(memorydb.New())
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(db), nil)

	// Not at epoch boundary (block 50)
	header := &types.Header{Number: big.NewInt(50)}
	accumulateAgentRewards(stateDB, snap, header)

	bal := stateDB.GetBalance(addr)
	if bal.Sign() != 0 {
		t.Errorf("Expected no rewards outside epoch boundary, got %s", bal)
	}
}

// TestUpdateAgentScores tests batch score re-evaluation.
func TestUpdateAgentScores(t *testing.T) {
	agent := NewAgentScoringAgent()

	a1 := common.HexToAddress("0xcccc000000000000000000000000000000000001")
	a2 := common.HexToAddress("0xcccc000000000000000000000000000000000002")
	nodes := map[common.Address]*AgentScore{
		a1: DefaultAgentScore(0),
		a2: DefaultAgentScore(0),
	}
	histories := map[common.Address]*AgentHistory{
		a1: {TasksCompleted: 100, HeartbeatsSent: 200},
		a2: {TasksFailed: 50, HeartbeatsMissed: 100},
	}

	updated := agent.UpdateAgentScores(nodes, histories, 1000)

	if len(updated) != 2 {
		t.Fatalf("Expected 2 updated scores, got %d", len(updated))
	}
	if updated[a1].LastUpdate != 1000 {
		t.Error("Expected LastUpdate to be 1000")
	}
}
