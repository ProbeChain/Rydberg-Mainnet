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

	"github.com/probechain/go-probe/common"
)

// SmartLight Protocol (SLP) sub-protocol message codes.
// These extend the LES protocol starting at 0x20.
const (
	SLPRegisterMsg     = 0x20 // SmartLight registration request
	SLPRegisterAckMsg  = 0x21 // Registration acknowledgment
	SLPHeartbeatMsg    = 0x22 // Heartbeat proof
	SLPHeartbeatAckMsg = 0x23 // Heartbeat acknowledgment
	SLPAckBatchMsg     = 0x24 // Batch ACK for block headers
	SLPTimeSampleMsg   = 0x25 // GNSS time sample submission
	SLPTaskRequestMsg  = 0x26 // Agent task assignment
	SLPTaskResultMsg   = 0x27 // Agent task result
)

// Agent SLP messages (PoB extension)
const (
	SLPAgentRegisterMsg  = 0x28 // Agent registration — agentID, capabilities, stake proof
	SLPAgentHeartbeatMsg = 0x29 // Agent heartbeat — every 250 blocks
	SLPAgentTaskAssignMsg = 0x2A // Task assignment from relay to agent
	SLPAgentTaskResultMsg = 0x2B // Task result + proof from agent
	SLPAgentAttestMsg    = 0x2C // Cross-agent attestation
)

// SLPProtocolLength is the total number of messages in the SLP sub-protocol.
const SLPProtocolLength = 0x2D // All messages up to 0x2C

// RegisterMsg is sent by a SmartLight node to register with a full node.
type RegisterMsg struct {
	Address    common.Address // SmartLight node address
	PublicKey  []byte         // Dilithium public key (1312 bytes) or ECDSA public key (65 bytes)
	Stake      *big.Int       // Staked PROBE amount
	DeviceInfo DeviceInfo     // Device attestation data
	Signature  []byte         // Registration signature
}

// DeviceInfo contains device attestation information for anti-Sybil.
type DeviceInfo struct {
	DeviceToken []byte  // iOS DeviceCheck token
	GPSLat      float64 // GPS latitude (for >100m dedup)
	GPSLon      float64 // GPS longitude
}

// HeartbeatMsg is a signed heartbeat proof from a SmartLight node.
type HeartbeatMsg struct {
	Address     common.Address // SmartLight node address
	BlockNumber uint64         // Current synced block number
	BlockHash   common.Hash    // Hash of the synced block
	Timestamp   uint64         // Unix timestamp
	Signature   []byte         // Signature over (address || blockNumber || blockHash || timestamp)
}

// AckBatchMsg contains a batch of header attestations from a SmartLight node.
type AckBatchMsg struct {
	Address common.Address // SmartLight node address
	Acks    []HeaderAck    // Batch of header acknowledgments
}

// HeaderAck is a single header attestation.
type HeaderAck struct {
	BlockNumber uint64      // Block number being attested
	BlockHash   common.Hash // Block hash being attested
	Valid       bool        // Whether the header was verified as valid
	Signature   []byte      // Signature over (blockNumber || blockHash || valid)
}

// TimeSampleMsg contains a GNSS time sample from a SmartLight node.
type TimeSampleMsg struct {
	Address       common.Address // SmartLight node address
	AtomicTime    []byte         // Encoded AtomicTimestamp (17 bytes)
	BlockNumber   uint64         // Associated block number
	GPSAccuracyNs uint32         // GPS clock accuracy in nanoseconds
	Signature     []byte         // Signature
}

// TaskRequestMsg assigns a lightweight agent task to a SmartLight node.
type TaskRequestMsg struct {
	TaskID      common.Hash    // Unique task identifier
	AgentAddr   common.Address // Agent requesting the task
	Payload     []byte         // Task payload (max 25MB)
	MaxMemoryMB int            // Maximum memory allowed
	TimeoutMs   uint64         // Timeout in milliseconds
}

// TaskResultMsg returns the result of an agent task.
type TaskResultMsg struct {
	TaskID    common.Hash    // Task identifier
	Address   common.Address // SmartLight node that executed
	Result    []byte         // Task result
	Success   bool           // Whether the task succeeded
	GasUsed   uint64         // Computational gas used
	Signature []byte         // Signature
}

// ---------------------------------------------------------------------------
// Agent SLP Messages (PoB)
// ---------------------------------------------------------------------------

// AgentRegisterMsg is sent by an Agent node to register with the network.
type AgentRegisterMsg struct {
	AgentID      common.Hash    // ERC-8004 style identity hash
	Address      common.Address // Agent node address
	PublicKey    []byte         // Ed25519 or Dilithium public key
	Capabilities []string       // Agent capabilities (e.g., "verify", "compute", "attest")
	StakeProof   []byte         // Proof of 0.1 PROBE stake
	Signature    []byte         // Registration signature
}

// AgentHeartbeatMsg is a periodic heartbeat from an Agent node (every 250 blocks).
type AgentHeartbeatMsg struct {
	AgentID     common.Hash    // Agent identity hash
	Address     common.Address // Agent node address
	BlockNumber uint64         // Current synced block number
	BlockHash   common.Hash    // Hash of the synced block
	TaskLoad    uint8          // Current task load (0-100%)
	Timestamp   uint64         // Unix timestamp
	Signature   []byte         // Heartbeat signature
}

// AgentTaskAssignMsg assigns a task from a relay to an agent.
type AgentTaskAssignMsg struct {
	TaskID      common.Hash    // Unique task identifier
	RelayAddr   common.Address // Relay node that assigned the task
	AgentAddr   common.Address // Target agent address
	TaskType    uint8          // 0=Verify, 1=Compute, 2=Attest
	Payload     []byte         // Task payload
	TimeoutMs   uint64         // Timeout in milliseconds
	RewardWei   uint64         // Expected reward in wei
	Signature   []byte         // Relay signature
}

// AgentTaskResultMsg returns the result of an agent task execution.
type AgentTaskResultMsg struct {
	TaskID    common.Hash    // Task identifier
	AgentAddr common.Address // Agent that executed
	Result    []byte         // Task result
	Success   bool           // Whether the task succeeded
	ProofHash common.Hash    // Hash of the execution proof
	GasUsed   uint64         // Computational gas used
	Signature []byte         // Agent signature
}

// AgentAttestMsg is a cross-agent attestation message.
type AgentAttestMsg struct {
	AttesterAddr common.Address // Agent providing attestation
	TargetAddr   common.Address // Agent being attested
	TaskID       common.Hash    // Task being attested (optional)
	Claim        []byte         // Attestation claim data
	Confidence   uint16         // Confidence level (0-10000 basis points)
	BlockNumber  uint64         // Block number of attestation
	Signature    []byte         // Attester signature
}
