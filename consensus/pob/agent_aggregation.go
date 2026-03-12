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
	"encoding/binary"
	"math"

	"github.com/probechain/go-probe/common"
	"golang.org/x/crypto/sha3"
)

// ---------------------------------------------------------------------------
// Heartbeat Bloom Filter
// ---------------------------------------------------------------------------
// At 1M agents, individual heartbeat messages would overwhelm the network.
// Instead, SmartLight relay nodes aggregate heartbeats from their managed agents
// into a Bloom filter. 1M heartbeats compress to ~128KB (0.1% false positive rate).

const (
	// HeartbeatBloomSize is the Bloom filter size in bytes for heartbeat aggregation.
	// 128KB = 131072 bytes = 1048576 bits → supports ~1M entries at 0.1% FPR with 10 hashes.
	HeartbeatBloomSize = 131072

	// HeartbeatBloomHashes is the number of hash functions for the Bloom filter.
	HeartbeatBloomHashes = 10
)

// HeartbeatBloom is a Bloom filter for aggregating agent heartbeats.
// Relay nodes use this to compress 1M heartbeat proofs into 128KB.
type HeartbeatBloom struct {
	Bits       []byte `json:"bits"`       // Bloom filter bit array
	Count      uint64 `json:"count"`      // Number of heartbeats recorded
	BlockRange [2]uint64 `json:"blockRange"` // [startBlock, endBlock] covered
}

// NewHeartbeatBloom creates a new heartbeat Bloom filter.
func NewHeartbeatBloom(startBlock, endBlock uint64) *HeartbeatBloom {
	return &HeartbeatBloom{
		Bits:       make([]byte, HeartbeatBloomSize),
		Count:      0,
		BlockRange: [2]uint64{startBlock, endBlock},
	}
}

// Add records an agent's heartbeat in the Bloom filter.
func (b *HeartbeatBloom) Add(agentAddr common.Address, blockNumber uint64) {
	for i := uint64(0); i < HeartbeatBloomHashes; i++ {
		pos := b.hashPosition(agentAddr, blockNumber, i)
		byteIdx := pos / 8
		bitIdx := pos % 8
		b.Bits[byteIdx] |= 1 << bitIdx
	}
	b.Count++
}

// Contains checks if an agent's heartbeat might be in the Bloom filter.
// May return false positives but never false negatives.
func (b *HeartbeatBloom) Contains(agentAddr common.Address, blockNumber uint64) bool {
	for i := uint64(0); i < HeartbeatBloomHashes; i++ {
		pos := b.hashPosition(agentAddr, blockNumber, i)
		byteIdx := pos / 8
		bitIdx := pos % 8
		if b.Bits[byteIdx]&(1<<bitIdx) == 0 {
			return false
		}
	}
	return true
}

// hashPosition computes the bit position for the i-th hash function.
func (b *HeartbeatBloom) hashPosition(addr common.Address, blockNum uint64, hashIdx uint64) uint64 {
	h := sha3.NewLegacyKeccak256()
	h.Write(addr[:])
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], blockNum)
	h.Write(buf[:])
	binary.BigEndian.PutUint64(buf[:], hashIdx)
	h.Write(buf[:])
	var hash [32]byte
	h.Sum(hash[:0])
	totalBits := uint64(HeartbeatBloomSize) * 8
	return binary.BigEndian.Uint64(hash[:8]) % totalBits
}

// FalsePositiveRate returns the estimated false positive rate for the current fill level.
func (b *HeartbeatBloom) FalsePositiveRate() float64 {
	totalBits := float64(HeartbeatBloomSize) * 8
	k := float64(HeartbeatBloomHashes)
	n := float64(b.Count)
	return math.Pow(1-math.Exp(-k*n/totalBits), k)
}

// Size returns the Bloom filter size in bytes.
func (b *HeartbeatBloom) Size() int {
	return len(b.Bits)
}

// ---------------------------------------------------------------------------
// Attestation Aggregation
// ---------------------------------------------------------------------------
// SmartLight relays aggregate agent attestations using BLS signature aggregation
// before forwarding to validators. This reduces on-chain data from O(agents) to O(1).

// AggregatedAttestation represents a batch of agent attestations aggregated by a relay.
type AggregatedAttestation struct {
	// RelayAddr is the SmartLight relay that performed the aggregation.
	RelayAddr common.Address `json:"relayAddr"`

	// TaskID is the task being attested (empty for general attestations).
	TaskID common.Hash `json:"taskId"`

	// BlockNumber is the block number this attestation covers.
	BlockNumber uint64 `json:"blockNumber"`

	// AgentCount is the number of agents whose attestations are aggregated.
	AgentCount uint64 `json:"agentCount"`

	// ParticipantBloom is a Bloom filter of participating agent addresses.
	// Used for quick membership checks without storing all addresses.
	ParticipantBloom []byte `json:"participantBloom"`

	// AggregateSignature is the BLS aggregated signature from all participating agents.
	// For ECDSA-only agents, this is a compact multi-signature.
	AggregateSignature []byte `json:"aggregateSignature"`

	// ResultHash is the keccak256 hash of the agreed-upon result.
	// Agents that attested to a different result are excluded.
	ResultHash common.Hash `json:"resultHash"`

	// Confidence is the average confidence level (0-10000 basis points).
	Confidence uint64 `json:"confidence"`

	// RelaySignature is the relay's signature over the entire attestation.
	RelaySignature []byte `json:"relaySignature"`
}

// ParticipantBloomSize is the size of the participant Bloom filter in bytes.
// 4KB supports ~1000 agents per relay at 0.1% FPR.
const ParticipantBloomSize = 4096

// NewAggregatedAttestation creates a new aggregated attestation.
func NewAggregatedAttestation(relayAddr common.Address, taskID common.Hash, blockNumber uint64) *AggregatedAttestation {
	return &AggregatedAttestation{
		RelayAddr:        relayAddr,
		TaskID:           taskID,
		BlockNumber:      blockNumber,
		ParticipantBloom: make([]byte, ParticipantBloomSize),
	}
}

// AddParticipant marks an agent as a participant in the aggregated attestation.
func (a *AggregatedAttestation) AddParticipant(agentAddr common.Address) {
	// Use 3 hash functions for the small participant bloom
	for i := uint64(0); i < 3; i++ {
		pos := participantBloomPos(agentAddr, i)
		totalBits := uint64(ParticipantBloomSize) * 8
		pos = pos % totalBits
		byteIdx := pos / 8
		bitIdx := pos % 8
		a.ParticipantBloom[byteIdx] |= 1 << bitIdx
	}
	a.AgentCount++
}

// IsParticipant checks if an agent might be a participant.
func (a *AggregatedAttestation) IsParticipant(agentAddr common.Address) bool {
	for i := uint64(0); i < 3; i++ {
		pos := participantBloomPos(agentAddr, i)
		totalBits := uint64(ParticipantBloomSize) * 8
		pos = pos % totalBits
		byteIdx := pos / 8
		bitIdx := pos % 8
		if a.ParticipantBloom[byteIdx]&(1<<bitIdx) == 0 {
			return false
		}
	}
	return true
}

func participantBloomPos(addr common.Address, hashIdx uint64) uint64 {
	h := sha3.NewLegacyKeccak256()
	h.Write(addr[:])
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], hashIdx)
	h.Write(buf[:])
	var hash [32]byte
	h.Sum(hash[:0])
	return binary.BigEndian.Uint64(hash[:8])
}

// Hash returns the keccak256 hash of the aggregated attestation for signing.
func (a *AggregatedAttestation) Hash() common.Hash {
	h := sha3.NewLegacyKeccak256()
	h.Write(a.RelayAddr[:])
	h.Write(a.TaskID[:])
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], a.BlockNumber)
	h.Write(buf[:])
	binary.BigEndian.PutUint64(buf[:], a.AgentCount)
	h.Write(buf[:])
	h.Write(a.ResultHash[:])
	binary.BigEndian.PutUint64(buf[:], a.Confidence)
	h.Write(buf[:])
	var hash common.Hash
	h.Sum(hash[:0])
	return hash
}
