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
	"math/big"

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/rlp"
	"golang.org/x/crypto/sha3"
)

// AgentTrieState is a compact representation of an agent's state
// stored in a separate Merkle trie (not the main state trie).
// This prevents agent state from bloating the main chain.
type AgentTrieState struct {
	Address     common.Address `json:"address"`
	Score       uint64         `json:"score"`
	StakeWei    *big.Int       `json:"stakeWei"`
	LastActive  uint64         `json:"lastActive"`
	TasksDone   uint64         `json:"tasksDone"`
	TasksOK     uint64         `json:"tasksOK"`
	AttestsOK   uint64         `json:"attestsOK"`
	Heartbeats  uint64         `json:"heartbeats"`
	SlashCount  uint64         `json:"slashCount"`
	Registered  uint64         `json:"registered"` // Block number of registration
}

// Encode serializes an AgentTrieState to RLP bytes.
func (a *AgentTrieState) Encode() ([]byte, error) {
	return rlp.EncodeToBytes(a)
}

// DecodeAgentTrieState deserializes an AgentTrieState from RLP bytes.
func DecodeAgentTrieState(data []byte) (*AgentTrieState, error) {
	var state AgentTrieState
	if err := rlp.DecodeBytes(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// AgentTrieKey returns the trie key for an agent address.
// Key = keccak256(address) — same hashing as Ethereum account trie.
func AgentTrieKey(addr common.Address) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write(addr[:])
	return h.Sum(nil)
}

// AgentTrieRoot computes the Merkle root of all agent states.
// This root is included in the block header Extra field after the agent consensus fork.
// The trie is separate from the main state trie to prevent chain bloat at 1M+ agents.
func AgentTrieRoot(agents map[common.Address]*AgentScore, histories map[common.Address]*AgentHistory) common.Hash {
	if len(agents) == 0 {
		return common.Hash{}
	}

	// Simple Merkle tree: sort keys, hash leaves, combine pairwise
	leaves := make([]common.Hash, 0, len(agents))
	for addr, score := range agents {
		history := histories[addr]
		leaf := agentLeafHash(addr, score, history)
		leaves = append(leaves, leaf)
	}

	// Sort leaves for deterministic ordering
	for i := 0; i < len(leaves); i++ {
		for j := i + 1; j < len(leaves); j++ {
			if leaves[j].Hex() < leaves[i].Hex() {
				leaves[i], leaves[j] = leaves[j], leaves[i]
			}
		}
	}

	return merkleRoot(leaves)
}

// agentLeafHash computes the hash of a single agent's state.
func agentLeafHash(addr common.Address, score *AgentScore, history *AgentHistory) common.Hash {
	h := sha3.NewLegacyKeccak256()
	h.Write(addr[:])

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], score.Total)
	h.Write(buf[:])

	if history != nil {
		binary.BigEndian.PutUint64(buf[:], history.TasksCompleted)
		h.Write(buf[:])
		binary.BigEndian.PutUint64(buf[:], history.AttestationsOK)
		h.Write(buf[:])
		binary.BigEndian.PutUint64(buf[:], history.HeartbeatsSent)
		h.Write(buf[:])
		binary.BigEndian.PutUint64(buf[:], history.LastActive)
		h.Write(buf[:])
	}

	var hash common.Hash
	h.Sum(hash[:0])
	return hash
}

// merkleRoot builds a Merkle root from a sorted list of leaf hashes.
func merkleRoot(leaves []common.Hash) common.Hash {
	if len(leaves) == 0 {
		return common.Hash{}
	}
	if len(leaves) == 1 {
		return leaves[0]
	}

	// Pad to even length
	for len(leaves)%2 != 0 {
		leaves = append(leaves, leaves[len(leaves)-1])
	}

	// Build tree bottom-up
	for len(leaves) > 1 {
		var nextLevel []common.Hash
		for i := 0; i < len(leaves); i += 2 {
			h := sha3.NewLegacyKeccak256()
			h.Write(leaves[i][:])
			h.Write(leaves[i+1][:])
			var combined common.Hash
			h.Sum(combined[:0])
			nextLevel = append(nextLevel, combined)
		}
		leaves = nextLevel
	}

	return leaves[0]
}

// EncodeAgentTrieRoot encodes the agent trie root into the block header Extra field.
// Format: [...existing extra data...] + [32 bytes agent trie root]
// The root is appended after the existing vanity + seal data.
func EncodeAgentTrieRoot(extra []byte, root common.Hash) []byte {
	result := make([]byte, len(extra)+32)
	copy(result, extra)
	copy(result[len(extra):], root[:])
	return result
}

// DecodeAgentTrieRoot extracts the agent trie root from the block header Extra field.
// Returns empty hash if extra data doesn't contain an agent root.
func DecodeAgentTrieRoot(extra []byte, vanityLen, sealLen int) common.Hash {
	agentRootStart := vanityLen + sealLen
	if len(extra) < agentRootStart+32 {
		return common.Hash{}
	}
	// Agent root is after vanity+seal, but before behavior data
	// For simplicity, check if extra has 32 additional bytes
	remaining := extra[vanityLen : len(extra)-sealLen]
	if len(remaining) < 32 {
		return common.Hash{}
	}
	var root common.Hash
	copy(root[:], remaining[len(remaining)-32:])
	return root
}
