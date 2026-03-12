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

package smartlight

import (
	"math/big"
	"sync"

	"github.com/probechain/go-probe/common"
)

const (
	slMaxScore     = uint64(10000) // Maximum behavior score (basis points)
	slDefaultScore = uint64(5000)  // Default starting score
)

// SmartLightScore holds the composite and per-dimension scores for a SmartLight node.
// Uses different weights than full validators: liveness 30%, correctness 20%,
// cooperation 25%, consistency 10%, signal sovereignty 15%.
type SmartLightScore struct {
	Total             uint64 `json:"total"`             // Composite score (0-10000 basis points)
	Liveness          uint64 `json:"liveness"`          // Heartbeat participation
	Correctness       uint64 `json:"correctness"`       // Header attestation accuracy
	Cooperation       uint64 `json:"cooperation"`       // ACK participation rate
	Consistency       uint64 `json:"consistency"`       // No invalid attestations
	SignalSovereignty uint64 `json:"signalSovereignty"` // GNSS time contribution
	LastUpdate        uint64 `json:"lastUpdate"`        // Block number of last score update
}

// SmartLightHistory tracks actions for a SmartLight node's behavior scoring.
type SmartLightHistory struct {
	HeartbeatsSent    uint64 `json:"heartbeatsSent"`    // Total heartbeats sent
	HeartbeatsMissed  uint64 `json:"heartbeatsMissed"`  // Missed heartbeat windows
	AcksGiven         uint64 `json:"acksGiven"`         // Total ACK attestations sent
	AcksMissed        uint64 `json:"acksMissed"`        // Missed ACK opportunities
	ValidAttestations uint64 `json:"validAttestations"` // Correct header attestations
	InvalidAttests    uint64 `json:"invalidAttests"`    // Incorrect header attestations
	SlashCount        uint64 `json:"slashCount"`        // Slashing events
	GNSSSamples       uint64 `json:"gnssSamples"`       // GNSS time samples submitted
	TasksCompleted    uint64 `json:"tasksCompleted"`    // Agent tasks completed
	TasksFailed       uint64 `json:"tasksFailed"`       // Agent tasks failed
}

// RewardStats tracks reward accumulation for a SmartLight node.
type RewardStats struct {
	TotalRewards    *big.Int `json:"totalRewards"`    // Total PROBE earned
	CurrentEpoch    uint64   `json:"currentEpoch"`    // Current epoch number
	EpochRewards    *big.Int `json:"epochRewards"`    // Rewards in current epoch
	LastRewardBlock uint64   `json:"lastRewardBlock"` // Block of last reward
}

// RewardTracker tracks the local SmartLight node's behavior score and rewards.
type RewardTracker struct {
	address common.Address

	mu      sync.RWMutex
	score   *SmartLightScore
	history *SmartLightHistory
	stats   *RewardStats
}

// NewRewardTracker creates a new reward tracker for the given address.
func NewRewardTracker(address common.Address) *RewardTracker {
	return &RewardTracker{
		address: address,
		score: &SmartLightScore{
			Total:             slDefaultScore,
			Liveness:          slMaxScore,
			Correctness:       slMaxScore,
			Cooperation:       slMaxScore,
			Consistency:       slMaxScore,
			SignalSovereignty: slDefaultScore,
		},
		history: &SmartLightHistory{},
		stats: &RewardStats{
			TotalRewards: new(big.Int),
			EpochRewards: new(big.Int),
		},
	}
}

// GetScore returns a copy of the current behavior score.
func (r *RewardTracker) GetScore() *SmartLightScore {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s := *r.score
	return &s
}

// GetStats returns a copy of the current reward stats.
func (r *RewardTracker) GetStats() *RewardStats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return &RewardStats{
		TotalRewards:    new(big.Int).Set(r.stats.TotalRewards),
		CurrentEpoch:    r.stats.CurrentEpoch,
		EpochRewards:    new(big.Int).Set(r.stats.EpochRewards),
		LastRewardBlock: r.stats.LastRewardBlock,
	}
}

// OnEpochBoundary processes epoch boundary events, recalculating scores.
func (r *RewardTracker) OnEpochBoundary(blockNumber uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.recalculate(blockNumber)
	r.stats.CurrentEpoch = blockNumber / 30000
	r.stats.EpochRewards = new(big.Int)
}

// RecordHeartbeat records a successful heartbeat.
func (r *RewardTracker) RecordHeartbeat() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.history.HeartbeatsSent++
}

// RecordAck records a successful ACK.
func (r *RewardTracker) RecordAck() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.history.AcksGiven++
}

// RecordGNSSSample records a GNSS time sample submission.
func (r *RewardTracker) RecordGNSSSample() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.history.GNSSSamples++
}

// RecordTaskResult records an agent task completion.
func (r *RewardTracker) RecordTaskResult(success bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if success {
		r.history.TasksCompleted++
	} else {
		r.history.TasksFailed++
	}
}

// recalculate updates the behavior score based on history.
// SmartLight weights: liveness 30%, correctness 20%, cooperation 25%,
// consistency 10%, signal sovereignty 15%.
func (r *RewardTracker) recalculate(blockNumber uint64) {
	h := r.history

	// Liveness: heartbeat participation rate (30%)
	totalHB := h.HeartbeatsSent + h.HeartbeatsMissed
	var liveness uint64
	if totalHB == 0 {
		liveness = slMaxScore
	} else {
		liveness = (h.HeartbeatsSent * slMaxScore) / totalHB
	}

	// Correctness: attestation accuracy (20%)
	totalAttests := h.ValidAttestations + h.InvalidAttests
	var correctness uint64
	if totalAttests == 0 {
		correctness = slMaxScore
	} else {
		correctness = (h.ValidAttestations * slMaxScore) / totalAttests
	}

	// Cooperation: ACK participation rate (25%)
	totalAcks := h.AcksGiven + h.AcksMissed
	var cooperation uint64
	if totalAcks == 0 {
		cooperation = slMaxScore
	} else {
		cooperation = (h.AcksGiven * slMaxScore) / totalAcks
	}

	// Consistency: inverse of slash count (10%)
	var consistency uint64
	penalty := h.SlashCount * 1000
	if penalty >= slMaxScore {
		consistency = 0
	} else {
		consistency = slMaxScore - penalty
	}

	// Signal Sovereignty: GNSS sample contribution (15%)
	var signalSov uint64
	if h.GNSSSamples == 0 {
		signalSov = slDefaultScore // Neutral baseline
	} else {
		signalSov = h.GNSSSamples * 100
		if signalSov > slMaxScore {
			signalSov = slMaxScore
		}
	}

	total := (liveness*30 + correctness*20 + cooperation*25 +
		consistency*10 + signalSov*15) / 100
	if total > slMaxScore {
		total = slMaxScore
	}

	r.score = &SmartLightScore{
		Total:             total,
		Liveness:          liveness,
		Correctness:       correctness,
		Cooperation:       cooperation,
		Consistency:       consistency,
		SignalSovereignty: signalSov,
		LastUpdate:        blockNumber,
	}
}
