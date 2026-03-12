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
	"github.com/probechain/go-probe/core/rawdb"
	"github.com/probechain/go-probe/core/state"
	"github.com/probechain/go-probe/core/types"
	"github.com/probechain/go-probe/params"
)

func TestCalcBlockReward_ZeroTx(t *testing.T) {
	cfg := params.DefaultPoBV2Config()
	reward := CalcBlockReward(0, cfg)
	// 0 tx → (0+1) × 1e14 = 1e14
	expected := big.NewInt(1e14)
	if reward.Cmp(expected) != 0 {
		t.Errorf("0 tx reward: got %s, want %s", reward, expected)
	}
}

func TestCalcBlockReward_OneTx(t *testing.T) {
	cfg := params.DefaultPoBV2Config()
	reward := CalcBlockReward(1, cfg)
	// 1 tx → (1+1) × 1e14 = 2e14
	expected := big.NewInt(2e14)
	if reward.Cmp(expected) != 0 {
		t.Errorf("1 tx reward: got %s, want %s", reward, expected)
	}
}

func TestCalcBlockReward_100Tx(t *testing.T) {
	cfg := params.DefaultPoBV2Config()
	reward := CalcBlockReward(100, cfg)
	// 100 tx → (100+1) × 1e14 = 101e14
	expected := new(big.Int).Mul(big.NewInt(101), big.NewInt(1e14))
	if reward.Cmp(expected) != 0 {
		t.Errorf("100 tx reward: got %s, want %s", reward, expected)
	}
}

func TestCalcBlockReward_MaxCap(t *testing.T) {
	cfg := params.DefaultPoBV2Config()
	reward := CalcBlockReward(100000, cfg)
	// 100K tx → capped at 10 PROBE = 1e19
	expected := new(big.Int).Mul(big.NewInt(1e10), big.NewInt(1e9))
	if reward.Cmp(expected) != 0 {
		t.Errorf("100K tx reward: got %s, want %s", reward, expected)
	}
}

func TestCalcBlockReward_OverCap(t *testing.T) {
	cfg := params.DefaultPoBV2Config()
	reward := CalcBlockReward(200000, cfg)
	// Over 100K tx → still capped at 10 PROBE
	expected := new(big.Int).Mul(big.NewInt(1e10), big.NewInt(1e9))
	if reward.Cmp(expected) != 0 {
		t.Errorf("200K tx reward: got %s, want %s", reward, expected)
	}
}

func TestSplitBlockReward(t *testing.T) {
	cfg := params.DefaultPoBV2Config()
	total := big.NewInt(1e18) // 1 PROBE

	producer, agentPool, physicalPool := SplitBlockReward(total, cfg)

	// Agent: 40%, Physical: 30%, Producer: 30% + dust
	expectedAgent := new(big.Int).Div(new(big.Int).Mul(total, big.NewInt(4000)), big.NewInt(10000))
	expectedPhysical := new(big.Int).Div(new(big.Int).Mul(total, big.NewInt(3000)), big.NewInt(10000))

	if agentPool.Cmp(expectedAgent) != 0 {
		t.Errorf("agent pool: got %s, want %s", agentPool, expectedAgent)
	}
	if physicalPool.Cmp(expectedPhysical) != 0 {
		t.Errorf("physical pool: got %s, want %s", physicalPool, expectedPhysical)
	}

	// All parts must sum to total
	sum := new(big.Int).Add(producer, agentPool)
	sum.Add(sum, physicalPool)
	if sum.Cmp(total) != 0 {
		t.Errorf("split sum: got %s, want %s", sum, total)
	}
}

func TestCountRealTransactions(t *testing.T) {
	txs := []*types.Transaction{
		types.NewTransaction(0, common.Address{}, big.NewInt(1e18), 21000, big.NewInt(1), nil),  // value > 0
		types.NewTransaction(1, common.Address{}, big.NewInt(0), 21000, big.NewInt(1), nil),     // value = 0
		types.NewTransaction(2, common.Address{}, big.NewInt(5e17), 21000, big.NewInt(1), nil),  // value > 0
	}

	count := CountRealTransactions(txs)
	if count != 2 {
		t.Errorf("real tx count: got %d, want 2", count)
	}
}

func TestCalcTxVolumeWei(t *testing.T) {
	txs := []*types.Transaction{
		types.NewTransaction(0, common.Address{}, big.NewInt(1e18), 21000, big.NewInt(1), nil),
		types.NewTransaction(1, common.Address{}, big.NewInt(0), 21000, big.NewInt(1), nil),
		types.NewTransaction(2, common.Address{}, big.NewInt(5e17), 21000, big.NewInt(1), nil),
	}

	volume := CalcTxVolumeWei(txs)
	expected := big.NewInt(1.5e18) // 1e18 + 5e17
	if volume.Cmp(expected) != 0 {
		t.Errorf("tx volume: got %s, want %s", volume, expected)
	}
}

func TestIsEmissionActive(t *testing.T) {
	cfg := params.DefaultPoBV2Config()

	// Zero GDP → active
	if !IsEmissionActive(big.NewInt(0), cfg) {
		t.Error("emission should be active at zero GDP")
	}

	// Nil GDP → active
	if !IsEmissionActive(nil, cfg) {
		t.Error("emission should be active with nil GDP")
	}

	// Below target → active
	halfTarget, _ := new(big.Int).SetString("75000000000000000000000000000000", 10)
	if !IsEmissionActive(halfTarget, cfg) {
		t.Error("emission should be active below target")
	}

	// At or above target → inactive
	target, _ := new(big.Int).SetString(cfg.AgentGDPTargetWei, 10)
	if IsEmissionActive(target, cfg) {
		t.Error("emission should stop at target")
	}

	// Above target → inactive
	overTarget := new(big.Int).Add(target, big.NewInt(1))
	if IsEmissionActive(overTarget, cfg) {
		t.Error("emission should stop above target")
	}
}

func TestDistributeByScore(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(db), nil)

	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	pool := big.NewInt(1e18) // 1 PROBE
	scores := map[common.Address]uint64{
		addr1: 8000,
		addr2: 2000,
	}

	remainder := DistributeByScore(stateDB, pool, scores)

	bal1 := stateDB.GetBalance(addr1)
	bal2 := stateDB.GetBalance(addr2)

	// addr1 should get 80%, addr2 should get 20%
	expected1 := new(big.Int).Div(new(big.Int).Mul(pool, big.NewInt(8000)), big.NewInt(10000))
	expected2 := new(big.Int).Div(new(big.Int).Mul(pool, big.NewInt(2000)), big.NewInt(10000))

	if bal1.Cmp(expected1) != 0 {
		t.Errorf("addr1 balance: got %s, want %s", bal1, expected1)
	}
	if bal2.Cmp(expected2) != 0 {
		t.Errorf("addr2 balance: got %s, want %s", bal2, expected2)
	}

	// Remainder should be zero or near-zero
	if remainder.Cmp(big.NewInt(1)) > 0 {
		t.Errorf("remainder too large: %s", remainder)
	}
}

func TestDistributeByScore_Empty(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(db), nil)

	pool := big.NewInt(1e18)
	remainder := DistributeByScore(stateDB, pool, map[common.Address]uint64{})

	if remainder.Cmp(pool) != 0 {
		t.Errorf("empty scores should return full pool as remainder: got %s", remainder)
	}
}

func TestCalcPoBV2Difficulty(t *testing.T) {
	cfg := params.DefaultPoBV2Config()

	// 0 nodes → initial difficulty (1)
	if d := CalcPoBV2Difficulty(0, cfg); d != 1 {
		t.Errorf("0 nodes: got %d, want 1", d)
	}

	// 500 nodes → still 1 (below threshold of 1000)
	if d := CalcPoBV2Difficulty(500, cfg); d != 1 {
		t.Errorf("500 nodes: got %d, want 1", d)
	}

	// 2000 nodes → difficulty 2
	if d := CalcPoBV2Difficulty(2000, cfg); d != 2 {
		t.Errorf("2000 nodes: got %d, want 2", d)
	}

	// 10000 nodes → difficulty 10
	if d := CalcPoBV2Difficulty(10000, cfg); d != 10 {
		t.Errorf("10000 nodes: got %d, want 10", d)
	}
}

func TestAccumulateRewardsV2(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(db), nil)

	producer := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	agent1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	agent2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	phys1 := common.HexToAddress("0x3333333333333333333333333333333333333333")

	config := &params.ChainConfig{
		PoBV2: params.DefaultPoBV2Config(),
	}

	snap := &Snapshot{
		Agents: map[common.Address]*AgentScore{
			agent1: {Total: 7000},
			agent2: {Total: 3000},
		},
		PhysicalNodes: map[common.Address]*PhysicalNodeScore{
			phys1: {Total: 5000},
		},
		CumulativeAgentGDP: new(big.Int),
	}

	header := &types.Header{
		Number:   big.NewInt(100),
		Coinbase: producer,
	}

	// 10 real transactions
	txs := make([]*types.Transaction, 10)
	for i := 0; i < 10; i++ {
		txs[i] = types.NewTransaction(uint64(i), common.Address{}, big.NewInt(1e18), 21000, big.NewInt(1), nil)
	}

	accumulateRewardsV2(config, stateDB, header, txs, snap)

	// Total reward = (10+1) × 1e14 = 11e14
	totalReward := big.NewInt(11e14)

	// Verify everyone got something
	producerBal := stateDB.GetBalance(producer)
	agent1Bal := stateDB.GetBalance(agent1)
	agent2Bal := stateDB.GetBalance(agent2)
	phys1Bal := stateDB.GetBalance(phys1)

	if producerBal.Sign() <= 0 {
		t.Error("producer should have a balance")
	}
	if agent1Bal.Sign() <= 0 {
		t.Error("agent1 should have a balance")
	}
	if agent2Bal.Sign() <= 0 {
		t.Error("agent2 should have a balance")
	}
	if phys1Bal.Sign() <= 0 {
		t.Error("phys1 should have a balance")
	}

	// agent1 (score 7000) should get more than agent2 (score 3000)
	if agent1Bal.Cmp(agent2Bal) <= 0 {
		t.Errorf("agent1 (%s) should get more than agent2 (%s)", agent1Bal, agent2Bal)
	}

	// Sum of all balances should equal total reward
	total := new(big.Int).Add(producerBal, agent1Bal)
	total.Add(total, agent2Bal)
	total.Add(total, phys1Bal)
	if total.Cmp(totalReward) != 0 {
		t.Errorf("total distributed: got %s, want %s", total, totalReward)
	}

	// GDP should be updated (10 tx × 1e18 each = 1e19)
	expectedGDP := new(big.Int).Mul(big.NewInt(1e10), big.NewInt(1e9))
	if snap.CumulativeAgentGDP.Cmp(expectedGDP) != 0 {
		t.Errorf("GDP: got %s, want %s", snap.CumulativeAgentGDP, expectedGDP)
	}
}

func TestAccumulateRewardsV2_EmptyBlock(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(db), nil)

	producer := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	config := &params.ChainConfig{
		PoBV2: params.DefaultPoBV2Config(),
	}

	snap := &Snapshot{
		Agents:             map[common.Address]*AgentScore{},
		PhysicalNodes:      map[common.Address]*PhysicalNodeScore{},
		CumulativeAgentGDP: new(big.Int),
	}

	header := &types.Header{
		Number:   big.NewInt(1),
		Coinbase: producer,
	}

	// Empty block — no transactions
	accumulateRewardsV2(config, stateDB, header, nil, snap)

	// Producer gets full reward (0.0001 PROBE = 1e14) since no agents/physical nodes
	producerBal := stateDB.GetBalance(producer)
	expected := big.NewInt(1e14)
	if producerBal.Cmp(expected) != 0 {
		t.Errorf("empty block producer reward: got %s, want %s", producerBal, expected)
	}
}

func TestAccumulateRewardsV2_GDPHalt(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(db), nil)

	producer := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	config := &params.ChainConfig{
		PoBV2: params.DefaultPoBV2Config(),
	}

	// GDP already at target
	target, _ := new(big.Int).SetString(config.PoBV2.AgentGDPTargetWei, 10)
	snap := &Snapshot{
		Agents:             map[common.Address]*AgentScore{},
		PhysicalNodes:      map[common.Address]*PhysicalNodeScore{},
		CumulativeAgentGDP: target,
	}

	header := &types.Header{
		Number:   big.NewInt(1),
		Coinbase: producer,
	}

	txs := []*types.Transaction{
		types.NewTransaction(0, common.Address{}, big.NewInt(1e18), 21000, big.NewInt(1), nil),
	}

	accumulateRewardsV2(config, stateDB, header, txs, snap)

	// Nobody should get any reward
	producerBal := stateDB.GetBalance(producer)
	if producerBal.Sign() != 0 {
		t.Errorf("GDP halt: producer should get 0, got %s", producerBal)
	}
}

func TestPhysicalNodeScoring(t *testing.T) {
	scorer := NewPhysicalNodeScoringAgent()

	history := &PhysicalNodeHistory{
		StorageBytesProvided: 1000000,
		StorageBytesVerified: 1000000, // 100% verified
		UptimeBlocks:         9000,
		DowntimeBlocks:       1000, // 90% uptime
		DataServiced:         1000, // Full data service
		InvalidProofs:        0,
		LastActive:           10000,
	}

	score := scorer.EvaluatePhysicalNode(history, 10000)

	if score.StorageContribution != 10000 {
		t.Errorf("storage: got %d, want 10000", score.StorageContribution)
	}
	if score.Uptime != 9000 {
		t.Errorf("uptime: got %d, want 9000", score.Uptime)
	}
	if score.DataService != 10000 {
		t.Errorf("data service: got %d, want 10000", score.DataService)
	}
	if score.Integrity != 10000 {
		t.Errorf("integrity: got %d, want 10000", score.Integrity)
	}

	// Total = (10000×40 + 9000×25 + 10000×20 + 10000×15) / 100
	// = (400000 + 225000 + 200000 + 150000) / 100 = 9750
	if score.Total != 9750 {
		t.Errorf("total: got %d, want 9750", score.Total)
	}
}
