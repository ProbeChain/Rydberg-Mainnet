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
// PoB V2.1 Reward Model — OZ Gold Standard
// ---------------------------------------------------------------------------
//
// Philosophy:
//   Bitcoin solved central bank money printing (fixed supply, time-based halving).
//   PROBE solves Agent GDP measurement, settlement, and behavior governance
//   (volume-coupled emission, gold-reserve-based decay).
//
// OZ Gold Standard:
//   OZ is Probe Banks' gold-backed token (1 OZ = 1 troy ounce of 99.99% gold).
//   100% physical gold backing. No fractional reserve.
//   OZ totalSupply() on-chain = Probe Banks' actual gold reserves.
//
// Emission Halt Condition:
//   When Probe Banks stores 36,000 metric tons of gold (1,157,425,200 OZ),
//   PROBE emission stops. This parallels how global central banks' ~32,000
//   tons of gold support $150 trillion of carbon-based GDP. Probe Banks'
//   36,000 tons will support the silicon-based Agent GDP (denominated in OZ).
//
// Block Reward Formula:
//   qualifiedVolume = Σ tx.Value() for qualified transactions
//   decay = (1 - goldReserveOZ / targetOZ)^n
//   reward = min(qualifiedVolume × rewardRate × decay, maxBlockReward)
//   Empty block: heartbeatReward × decay
//
// Qualified Transaction Filter:
//   - tx.Value() >= MinTxValueWei (default 0.01 PROBE)
//   - tx.To() != sender (no self-transfers, when sender info available)
//
// Anti-Sybil Economics:
//   Base fee floor at 1 Gwei (cannot drop to zero even with empty blocks).
//   Each tx costs ≥ 21,000 gas × 1 Gwei = 0.000021 PROBE (burned via EIP-1559).
//   At rewardRate = 5 bps (0.05%), break-even tx value = 0.042 PROBE.
//   MinTxValue = 0.01 PROBE provides 4.2× safety margin:
//     Reward for 0.01 PROBE tx: 0.01 × 5/10000 = 0.000005 PROBE
//     Gas cost for 1 tx:        21000 × 1 Gwei  = 0.000021 PROBE
//     Net: -0.000016 PROBE (loss, 4.2× margin) ✓

// CalcQualifiedVolume computes the total value and count of qualified transactions.
// A transaction qualifies if its value meets the minimum threshold.
// When sender information is available (senders map), self-transfers are excluded.
func CalcQualifiedVolume(txs []*types.Transaction, senders map[common.Hash]common.Address, config *params.PoBV2Config) (*big.Int, uint64) {
	if config == nil {
		config = params.DefaultPoBV2Config()
	}
	minValue := new(big.Int).SetUint64(config.MinTxValueWei)
	volume := new(big.Int)
	var count uint64

	for _, tx := range txs {
		// Filter: value must meet minimum threshold
		if tx.Value().Cmp(minValue) < 0 {
			continue
		}
		// Filter: no self-transfers (when sender info available)
		if senders != nil {
			sender, ok := senders[tx.Hash()]
			if ok && tx.To() != nil && sender == *tx.To() {
				continue
			}
		}
		volume.Add(volume, tx.Value())
		count++
	}
	return volume, count
}

// CalcDecayFactor returns the emission decay factor in basis points (0-10000).
// decay = ((targetOZ - goldReserveOZ) / targetOZ)^n × 10000
// Returns 10000 (full emission) when reserves are 0.
// Returns 0 when reserves >= target (emission stops).
func CalcDecayFactor(goldReserveOZ *big.Int, config *params.PoBV2Config) *big.Int {
	if config == nil {
		config = params.DefaultPoBV2Config()
	}
	target, ok := new(big.Int).SetString(config.GoldReserveTargetOZ, 10)
	if !ok || target.Sign() <= 0 {
		return new(big.Int).SetUint64(10000) // Unparseable target → full emission
	}
	if goldReserveOZ == nil || goldReserveOZ.Sign() <= 0 {
		return new(big.Int).SetUint64(10000) // No reserves → full emission
	}
	if goldReserveOZ.Cmp(target) >= 0 {
		return new(big.Int) // Target reached → zero emission
	}

	// Linear component: (target - reserve) × 10000 / target
	remaining := new(big.Int).Sub(target, goldReserveOZ)
	decay := new(big.Int).Mul(remaining, big.NewInt(10000))
	decay.Div(decay, target)

	// Apply exponent for non-linear decay (n > 1)
	if config.DecayExponent > 1 {
		base := new(big.Int).Set(decay)
		for i := uint64(1); i < config.DecayExponent; i++ {
			decay.Mul(decay, base)
			decay.Div(decay, big.NewInt(10000))
		}
	}
	return decay
}

// IsEmissionActive checks whether PROBE emission should continue.
// Returns false when Probe Banks gold reserves have reached the target.
func IsEmissionActive(goldReserveOZ *big.Int, config *params.PoBV2Config) bool {
	if config == nil {
		config = params.DefaultPoBV2Config()
	}
	target, ok := new(big.Int).SetString(config.GoldReserveTargetOZ, 10)
	if !ok {
		return true // Unparseable → keep emitting
	}
	if goldReserveOZ == nil {
		return true // No data → keep emitting
	}
	return goldReserveOZ.Cmp(target) < 0
}

// CalcBlockReward computes the block reward based on qualified transaction volume
// and current gold reserve decay factor.
func CalcBlockReward(qualifiedVolume *big.Int, goldReserveOZ *big.Int, config *params.PoBV2Config) *big.Int {
	if config == nil {
		config = params.DefaultPoBV2Config()
	}

	decay := CalcDecayFactor(goldReserveOZ, config)
	if decay.Sign() == 0 {
		return new(big.Int) // Emission stopped
	}

	if qualifiedVolume == nil || qualifiedVolume.Sign() == 0 {
		// Empty block: heartbeat reward × decay
		heartbeat := new(big.Int).SetUint64(config.HeartbeatRewardWei)
		heartbeat.Mul(heartbeat, decay)
		heartbeat.Div(heartbeat, big.NewInt(10000))
		return heartbeat
	}

	// reward = qualifiedVolume × rewardRateBps / 10000
	reward := new(big.Int).Mul(qualifiedVolume, new(big.Int).SetUint64(config.RewardRateBps))
	reward.Div(reward, big.NewInt(10000))

	// Apply gold-reserve decay
	reward.Mul(reward, decay)
	reward.Div(reward, big.NewInt(10000))

	// Cap at maximum block reward
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

// accumulateRewardsV2 implements the PoB V2.1 reward distribution.
// Called from PobFinalize when the PoB V2 fork is active.
func accumulateRewardsV2(config *params.ChainConfig, statedb *state.StateDB, header *types.Header, txs []*types.Transaction, snap *Snapshot) {
	v2cfg := config.PoBV2
	if v2cfg == nil {
		v2cfg = params.DefaultPoBV2Config()
	}

	// Check if emission is still active (based on Probe Banks gold reserves)
	goldReserve := snap.GoldReserveOZ
	if goldReserve == nil {
		goldReserve = new(big.Int)
	}
	if !IsEmissionActive(goldReserve, v2cfg) {
		return // Gold reserve target reached, no more emission
	}

	// Calculate qualified transaction volume (MinTxValue filter applied)
	// Phase 1: sender info not available, self-transfer check skipped
	qualifiedVolume, _ := CalcQualifiedVolume(txs, nil, v2cfg)

	// Calculate block reward (volume-coupled × gold-reserve decay)
	totalReward := CalcBlockReward(qualifiedVolume, goldReserve, v2cfg)

	// Split reward: producer, agent pool, physical pool
	producerReward, agentPool, physicalPool := SplitBlockReward(totalReward, v2cfg)

	// 1. Block producer gets their share
	producer := header.ValidatorAddr
	if producerReward.Sign() > 0 {
		statedb.AddBalance(producer, producerReward)
	}

	// 2. Distribute agent pool by behavior score
	agentScores := make(map[common.Address]uint64, len(snap.Agents))
	for addr, score := range snap.Agents {
		agentScores[addr] = score.Total
	}
	remainder := DistributeByScore(statedb, agentPool, agentScores)
	if remainder.Sign() > 0 {
		statedb.AddBalance(producer, remainder)
	}

	// 3. Distribute physical node pool by behavior score
	physicalScores := make(map[common.Address]uint64, len(snap.PhysicalNodes))
	for addr, score := range snap.PhysicalNodes {
		physicalScores[addr] = score.Total
	}
	remainder = DistributeByScore(statedb, physicalPool, physicalScores)
	if remainder.Sign() > 0 {
		statedb.AddBalance(producer, remainder)
	}

	// 4. Track cumulative GDP (for measurement/analytics, does NOT affect emission)
	if snap.CumulativeGDPWei == nil {
		snap.CumulativeGDPWei = new(big.Int)
	}
	snap.CumulativeGDPWei.Add(snap.CumulativeGDPWei, qualifiedVolume)
}
