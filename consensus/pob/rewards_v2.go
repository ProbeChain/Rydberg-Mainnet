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

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/core/state"
	"github.com/probechain/go-probe/core/types"
	"github.com/probechain/go-probe/params"
)

// ---------------------------------------------------------------------------
// PoB V2 Reward Model
// ---------------------------------------------------------------------------
//
// Two node types:
//   Agent Node (PoB-A)    — AI agents, rewarded for inter-agent settlement processing
//   Physical Node (PoB-P) — any physical device, rewarded for storage contribution
//
// Block reward = f(txCount):
//   0 real tx  → 0.0001 PROBE
//   1 real tx  → 0.0002 PROBE
//   N real tx  → (N+1) × 0.0001 PROBE
//   Cap at 100K tx → 10 PROBE
//
// Emission halts when cumulative Agent GDP reaches target (~$150T).

// CalcBlockReward computes the block reward based on the number of real transactions.
// reward = min((txCount + 1) * baseReward, maxReward)
func CalcBlockReward(txCount uint64, config *params.PoBV2Config) *big.Int {
	if config == nil {
		config = params.DefaultPoBV2Config()
	}
	base := new(big.Int).SetUint64(config.BaseRewardWei)
	count := txCount
	if count > config.MaxTxPerBlock {
		count = config.MaxTxPerBlock
	}
	// (count + 1) * baseReward
	reward := new(big.Int).Mul(base, new(big.Int).SetUint64(count+1))
	max := new(big.Int).SetUint64(config.MaxBlockRewardWei)
	if reward.Cmp(max) > 0 {
		reward.Set(max)
	}
	return reward
}

// SplitBlockReward splits the total block reward into producer, agent pool, and physical pool.
// Returns (producerReward, agentPool, physicalPool).
// Any remainder from integer division goes to the producer.
func SplitBlockReward(totalReward *big.Int, config *params.PoBV2Config) (producer, agentPool, physicalPool *big.Int) {
	if config == nil {
		config = params.DefaultPoBV2Config()
	}
	bps := new(big.Int).SetUint64(10000)

	agentPool = new(big.Int).Mul(totalReward, new(big.Int).SetUint64(config.AgentShareBps))
	agentPool.Div(agentPool, bps)

	physicalPool = new(big.Int).Mul(totalReward, new(big.Int).SetUint64(config.PhysicalShareBps))
	physicalPool.Div(physicalPool, bps)

	// Producer gets the remainder (guaranteed >= ProducerShareBps worth, plus rounding dust)
	producer = new(big.Int).Sub(totalReward, agentPool)
	producer.Sub(producer, physicalPool)

	return producer, agentPool, physicalPool
}

// CountRealTransactions counts non-zero-value transactions in a block.
// Zero-value transactions (like contract deployments with 0 value) are excluded
// from the reward calculation to prevent spam gaming.
func CountRealTransactions(txs []*types.Transaction) uint64 {
	var count uint64
	for _, tx := range txs {
		if tx.Value().Sign() > 0 {
			count++
		}
	}
	return count
}

// CalcTxVolumeWei sums the total value (in wei) of all non-zero transactions.
// This contributes to the Agent GDP calculation.
func CalcTxVolumeWei(txs []*types.Transaction) *big.Int {
	total := new(big.Int)
	for _, tx := range txs {
		if tx.Value().Sign() > 0 {
			total.Add(total, tx.Value())
		}
	}
	return total
}

// IsEmissionActive checks whether PROBE emission should continue.
// Returns false when cumulative Agent GDP has reached the target.
func IsEmissionActive(cumulativeGDP *big.Int, config *params.PoBV2Config) bool {
	if config == nil {
		config = params.DefaultPoBV2Config()
	}
	target, ok := new(big.Int).SetString(config.AgentGDPTargetWei, 10)
	if !ok {
		// If target is unparseable, emission continues
		return true
	}
	if cumulativeGDP == nil {
		return true
	}
	return cumulativeGDP.Cmp(target) < 0
}

// DistributeByScore distributes a reward pool proportionally to nodes by their behavior score.
// Returns any undistributed remainder.
func DistributeByScore(statedb *state.StateDB, pool *big.Int, scores map[common.Address]uint64) *big.Int {
	if pool.Sign() <= 0 || len(scores) == 0 {
		return new(big.Int).Set(pool)
	}

	var totalScore uint64
	for _, s := range scores {
		totalScore += s
	}
	if totalScore == 0 {
		return new(big.Int).Set(pool)
	}

	distributed := new(big.Int)
	for addr, score := range scores {
		reward := new(big.Int).Mul(pool, new(big.Int).SetUint64(score))
		reward.Div(reward, new(big.Int).SetUint64(totalScore))
		if reward.Sign() > 0 {
			statedb.AddBalance(addr, reward)
			distributed.Add(distributed, reward)
		}
	}
	return new(big.Int).Sub(pool, distributed)
}

// CalcPoBV2Difficulty calculates PoB V2 difficulty based on total node count.
// difficulty = max(1, totalNodes / nodesPerDifficultyUp)
func CalcPoBV2Difficulty(totalNodes uint64, config *params.PoBV2Config) uint64 {
	if config == nil {
		config = params.DefaultPoBV2Config()
	}
	if config.NodesPerDifficultyUp == 0 || totalNodes <= config.NodesPerDifficultyUp {
		return config.InitialDifficulty
	}
	diff := totalNodes / config.NodesPerDifficultyUp
	if diff < config.InitialDifficulty {
		return config.InitialDifficulty
	}
	return diff
}

// accumulateRewardsV2 implements the PoB V2 reward distribution.
// Called from PobFinalize when the PoB V2 fork is active.
func accumulateRewardsV2(config *params.ChainConfig, statedb *state.StateDB, header *types.Header, txs []*types.Transaction, snap *Snapshot) {
	v2cfg := config.PoBV2
	if v2cfg == nil {
		v2cfg = params.DefaultPoBV2Config()
	}

	// Check if emission is still active
	gdp := snap.CumulativeAgentGDP
	if gdp == nil {
		gdp = new(big.Int)
	}
	if !IsEmissionActive(gdp, v2cfg) {
		return // GDP target reached, no more emission
	}

	// Count real (non-zero-value) transactions
	realTxCount := CountRealTransactions(txs)

	// Calculate block reward based on tx count
	totalReward := CalcBlockReward(realTxCount, v2cfg)

	// Split reward: producer, agent pool, physical pool
	producerReward, agentPool, physicalPool := SplitBlockReward(totalReward, v2cfg)

	// 1. Block producer gets their share
	if producerReward.Sign() > 0 {
		statedb.AddBalance(header.Coinbase, producerReward)
	}

	// 2. Distribute agent pool by score
	agentScores := make(map[common.Address]uint64, len(snap.Agents))
	for addr, score := range snap.Agents {
		agentScores[addr] = score.Total
	}
	remainder := DistributeByScore(statedb, agentPool, agentScores)
	// Give agent remainder to producer
	if remainder.Sign() > 0 {
		statedb.AddBalance(header.Coinbase, remainder)
	}

	// 3. Distribute physical node pool by score
	physicalScores := make(map[common.Address]uint64, len(snap.PhysicalNodes))
	for addr, score := range snap.PhysicalNodes {
		physicalScores[addr] = score.Total
	}
	remainder = DistributeByScore(statedb, physicalPool, physicalScores)
	// Give physical remainder to producer
	if remainder.Sign() > 0 {
		statedb.AddBalance(header.Coinbase, remainder)
	}

	// 4. Update cumulative Agent GDP
	txVolume := CalcTxVolumeWei(txs)
	if snap.CumulativeAgentGDP == nil {
		snap.CumulativeAgentGDP = new(big.Int)
	}
	snap.CumulativeAgentGDP.Add(snap.CumulativeAgentGDP, txVolume)
}
