// Copyright 2024 The ProbeChain Authors
// This file is part of the ProbeChain.

package smartlight

import (
	"math/big"
	"testing"
	"time"

	"github.com/probechain/go-probe/common"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.HeartbeatInterval != 100 {
		t.Errorf("HeartbeatInterval = %d, want 100", cfg.HeartbeatInterval)
	}
	if cfg.AckWeight != 0.3 {
		t.Errorf("AckWeight = %f, want 0.3", cfg.AckWeight)
	}
	if cfg.MaxAgentTasks != 2 {
		t.Errorf("MaxAgentTasks = %d, want 2", cfg.MaxAgentTasks)
	}
	if cfg.MaxTaskMemoryMB != 25 {
		t.Errorf("MaxTaskMemoryMB = %d, want 25", cfg.MaxTaskMemoryMB)
	}
	if cfg.TaskTimeout != 5*time.Second {
		t.Errorf("TaskTimeout = %v, want 5s", cfg.TaskTimeout)
	}
	if cfg.PowerMode != PowerModeFull {
		t.Errorf("PowerMode = %d, want Full(0)", cfg.PowerMode)
	}
	if cfg.MaxRAMMB != 80 {
		t.Errorf("MaxRAMMB = %d, want 80", cfg.MaxRAMMB)
	}

	// Verify stake = 10 PROBE
	expectedStake := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))
	if cfg.StakeRequired.Cmp(expectedStake) != 0 {
		t.Errorf("StakeRequired = %s, want %s", cfg.StakeRequired, expectedStake)
	}

	// Verify reward pool = 0.2 PROBE
	expectedReward := new(big.Int).Mul(big.NewInt(2), big.NewInt(1e17))
	if cfg.RewardPoolPerBlock.Cmp(expectedReward) != 0 {
		t.Errorf("RewardPoolPerBlock = %s, want %s", cfg.RewardPoolPerBlock, expectedReward)
	}
}

func TestPowerModeString(t *testing.T) {
	tests := []struct {
		mode PowerMode
		want string
	}{
		{PowerModeFull, "Full"},
		{PowerModeEco, "Eco"},
		{PowerModeSleep, "Sleep"},
		{PowerMode(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("PowerMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestRewardTracker(t *testing.T) {
	addr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	tracker := NewRewardTracker(addr)

	// Default score should be 5000
	score := tracker.GetScore()
	if score.Total != 5000 {
		t.Errorf("Initial score = %d, want 5000", score.Total)
	}

	// Record some actions
	tracker.RecordHeartbeat()
	tracker.RecordHeartbeat()
	tracker.RecordAck()
	tracker.RecordGNSSSample()
	tracker.RecordTaskResult(true)
	tracker.RecordTaskResult(false)

	// Trigger recalculation
	tracker.OnEpochBoundary(30000)

	score = tracker.GetScore()
	if score.Total == 0 {
		t.Error("Score should not be 0 after positive actions")
	}
	if score.Liveness != 10000 {
		t.Errorf("Liveness = %d, want 10000 (no missed heartbeats)", score.Liveness)
	}
	if score.Cooperation != 10000 {
		t.Errorf("Cooperation = %d, want 10000 (no missed ACKs)", score.Cooperation)
	}
}

func TestRewardTrackerEpoch(t *testing.T) {
	addr := common.HexToAddress("0xdeadbeef")
	tracker := NewRewardTracker(addr)

	stats := tracker.GetStats()
	if stats.CurrentEpoch != 0 {
		t.Errorf("Initial epoch = %d, want 0", stats.CurrentEpoch)
	}

	tracker.OnEpochBoundary(30000)
	stats = tracker.GetStats()
	if stats.CurrentEpoch != 1 {
		t.Errorf("Epoch after 30000 = %d, want 1", stats.CurrentEpoch)
	}

	tracker.OnEpochBoundary(60000)
	stats = tracker.GetStats()
	if stats.CurrentEpoch != 2 {
		t.Errorf("Epoch after 60000 = %d, want 2", stats.CurrentEpoch)
	}
}

func TestSmartLightScoreRecalculation(t *testing.T) {
	addr := common.HexToAddress("0xabcdef")
	tracker := NewRewardTracker(addr)

	// Simulate 50% heartbeat success rate
	tracker.mu.Lock()
	tracker.history.HeartbeatsSent = 50
	tracker.history.HeartbeatsMissed = 50
	tracker.history.AcksGiven = 100
	tracker.history.AcksMissed = 0
	tracker.history.ValidAttestations = 95
	tracker.history.InvalidAttests = 5
	tracker.history.GNSSSamples = 100
	tracker.mu.Unlock()

	tracker.OnEpochBoundary(30000)
	score := tracker.GetScore()

	// Liveness should be 50% → 5000 bp
	if score.Liveness != 5000 {
		t.Errorf("Liveness = %d, want 5000", score.Liveness)
	}

	// Cooperation should be 100% → 10000 bp
	if score.Cooperation != 10000 {
		t.Errorf("Cooperation = %d, want 10000", score.Cooperation)
	}

	// Correctness should be 95% → 9500 bp
	if score.Correctness != 9500 {
		t.Errorf("Correctness = %d, want 9500", score.Correctness)
	}

	// SignalSovereignty: 100 samples × 100 = 10000 bp (capped)
	if score.SignalSovereignty != 10000 {
		t.Errorf("SignalSovereignty = %d, want 10000", score.SignalSovereignty)
	}

	// Total = (5000*30 + 9500*20 + 10000*25 + 10000*10 + 10000*15) / 100
	// = (150000 + 190000 + 250000 + 100000 + 150000) / 100
	// = 840000 / 100 = 8400
	if score.Total != 8400 {
		t.Errorf("Total = %d, want 8400", score.Total)
	}
}
