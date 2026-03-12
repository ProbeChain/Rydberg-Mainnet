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
	"time"
)

// PowerMode defines the node's operating power mode.
type PowerMode uint8

const (
	PowerModeFull  PowerMode = 0 // Charging: ACK + GNSS + Agent + Heartbeat
	PowerModeEco   PowerMode = 1 // Battery > 30%: ACK + Heartbeat only
	PowerModeSleep PowerMode = 2 // Battery < 15%: Sync only
	PowerModeAgent PowerMode = 3 // Headless agent: No GNSS, no heartbeat display, max task throughput
)

// Config holds the SmartLight node configuration.
type Config struct {
	// HeartbeatInterval is the number of blocks between heartbeat proofs.
	HeartbeatInterval uint64 `json:"heartbeatInterval"`

	// AckBatchInterval is how often to flush ACK batches.
	AckBatchInterval time.Duration `json:"ackBatchInterval"`

	// AckWeight is the relative weight of SmartLight ACKs vs full-node ACKs.
	// Expressed as a fraction (e.g., 0.3 means 30% weight).
	AckWeight float64 `json:"ackWeight"`

	// MaxAgentTasks is the maximum number of concurrent agent tasks.
	MaxAgentTasks int `json:"maxAgentTasks"`

	// MaxTaskMemoryMB is the maximum memory per agent task in megabytes.
	MaxTaskMemoryMB int `json:"maxTaskMemoryMB"`

	// TaskTimeout is the maximum duration for an agent task.
	TaskTimeout time.Duration `json:"taskTimeout"`

	// GNSSEnabled controls whether GNSS time sampling is active.
	GNSSEnabled bool `json:"gnssEnabled"`

	// GNSSSampleInterval is how often to sample GNSS time.
	GNSSSampleInterval time.Duration `json:"gnssSampleInterval"`

	// StakeRequired is the minimum PROBE stake to register as SmartLight.
	StakeRequired *big.Int `json:"stakeRequired"`

	// RewardPoolPerBlock is the PROBE reward allocated to SmartLight pool per block.
	RewardPoolPerBlock *big.Int `json:"rewardPoolPerBlock"`

	// PowerMode is the current operating power mode.
	PowerMode PowerMode `json:"powerMode"`

	// MaxRAMMB is the target maximum RAM usage in megabytes.
	MaxRAMMB int `json:"maxRamMB"`
}

// DefaultConfig returns the default SmartLight configuration.
func DefaultConfig() *Config {
	return &Config{
		HeartbeatInterval:  100,
		AckBatchInterval:   7 * time.Second,
		AckWeight:          0.3,
		MaxAgentTasks:      2,
		MaxTaskMemoryMB:    25,
		TaskTimeout:        5 * time.Second,
		GNSSEnabled:        false,
		GNSSSampleInterval: 30 * time.Second,
		StakeRequired:      new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18)), // 10 PROBE
		RewardPoolPerBlock: new(big.Int).Mul(big.NewInt(2), big.NewInt(1e17)),  // 0.2 PROBE
		PowerMode:          PowerModeFull,
		MaxRAMMB:           80,
	}
}

// DefaultAgentConfig returns the default configuration for agent mode.
// Agent mode is optimized for headless operation: no GNSS, reduced peer count,
// minimal logging, maximum task throughput, and lower RAM usage.
func DefaultAgentConfig() *Config {
	return &Config{
		HeartbeatInterval:  250,
		AckBatchInterval:   10 * time.Second,
		AckWeight:          0.2,
		MaxAgentTasks:      4,
		MaxTaskMemoryMB:    25,
		TaskTimeout:        5 * time.Second,
		GNSSEnabled:        false,
		GNSSSampleInterval: 0, // Disabled
		StakeRequired:      new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17)), // 0.1 PROBE
		RewardPoolPerBlock: new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17)), // 0.1 PROBE
		PowerMode:          PowerModeAgent,
		MaxRAMMB:           30,
	}
}

// String returns the power mode name.
func (m PowerMode) String() string {
	switch m {
	case PowerModeFull:
		return "Full"
	case PowerModeEco:
		return "Eco"
	case PowerModeSleep:
		return "Sleep"
	case PowerModeAgent:
		return "Agent"
	default:
		return "Unknown"
	}
}
