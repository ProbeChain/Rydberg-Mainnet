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
)

func TestAgentTrieKey(t *testing.T) {
	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	key1 := AgentTrieKey(addr)
	key2 := AgentTrieKey(addr)

	if len(key1) != 32 {
		t.Errorf("key length: got %d, want 32", len(key1))
	}
	// Deterministic
	for i := range key1 {
		if key1[i] != key2[i] {
			t.Error("AgentTrieKey should be deterministic")
			break
		}
	}

	// Different addresses produce different keys
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	key3 := AgentTrieKey(addr2)
	same := true
	for i := range key1 {
		if key1[i] != key3[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("different addresses should produce different keys")
	}
}

func TestAgentTrieStateEncodeDecode(t *testing.T) {
	state := &AgentTrieState{
		Address:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
		Score:      8500,
		StakeWei:   big.NewInt(1e17),
		LastActive: 1000,
		TasksDone:  50,
		TasksOK:    48,
		AttestsOK:  200,
		Heartbeats: 100,
		SlashCount: 1,
		Registered: 500,
	}

	data, err := state.Encode()
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("encoded data should not be empty")
	}

	decoded, err := DecodeAgentTrieState(data)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if decoded.Address != state.Address {
		t.Error("address mismatch")
	}
	if decoded.Score != state.Score {
		t.Error("score mismatch")
	}
	if decoded.StakeWei.Cmp(state.StakeWei) != 0 {
		t.Error("stake mismatch")
	}
	if decoded.LastActive != state.LastActive {
		t.Error("lastActive mismatch")
	}
	if decoded.TasksDone != state.TasksDone {
		t.Error("tasksDone mismatch")
	}
	if decoded.TasksOK != state.TasksOK {
		t.Error("tasksOK mismatch")
	}
	if decoded.AttestsOK != state.AttestsOK {
		t.Error("attestsOK mismatch")
	}
	if decoded.Heartbeats != state.Heartbeats {
		t.Error("heartbeats mismatch")
	}
	if decoded.SlashCount != state.SlashCount {
		t.Error("slashCount mismatch")
	}
	if decoded.Registered != state.Registered {
		t.Error("registered mismatch")
	}
}

func TestAgentTrieRootEmpty(t *testing.T) {
	root := AgentTrieRoot(nil, nil)
	if root != (common.Hash{}) {
		t.Error("empty agent set should produce empty root")
	}

	root2 := AgentTrieRoot(map[common.Address]*AgentScore{}, nil)
	if root2 != (common.Hash{}) {
		t.Error("empty agent map should produce empty root")
	}
}

func TestAgentTrieRootDeterministic(t *testing.T) {
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	agents := map[common.Address]*AgentScore{
		addr1: {Total: 8000},
		addr2: {Total: 6000},
	}
	histories := map[common.Address]*AgentHistory{
		addr1: {TasksCompleted: 50, AttestationsOK: 200, HeartbeatsSent: 100, LastActive: 1000},
		addr2: {TasksCompleted: 30, AttestationsOK: 150, HeartbeatsSent: 80, LastActive: 900},
	}

	root1 := AgentTrieRoot(agents, histories)
	root2 := AgentTrieRoot(agents, histories)

	if root1 != root2 {
		t.Error("AgentTrieRoot should be deterministic")
	}
	if root1 == (common.Hash{}) {
		t.Error("root should not be empty for non-empty agent set")
	}
}

func TestAgentTrieRootDifferentStates(t *testing.T) {
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")

	agents1 := map[common.Address]*AgentScore{addr1: {Total: 8000}}
	agents2 := map[common.Address]*AgentScore{addr1: {Total: 5000}}

	root1 := AgentTrieRoot(agents1, nil)
	root2 := AgentTrieRoot(agents2, nil)

	if root1 == root2 {
		t.Error("different scores should produce different roots")
	}
}

func TestMerkleRootSingleLeaf(t *testing.T) {
	leaf := common.HexToHash("0xaaaa")
	root := merkleRoot([]common.Hash{leaf})
	if root != leaf {
		t.Error("single leaf merkle root should equal the leaf")
	}
}

func TestMerkleRootPadding(t *testing.T) {
	// 3 leaves should be padded to 4
	leaves := []common.Hash{
		common.HexToHash("0xaaaa"),
		common.HexToHash("0xbbbb"),
		common.HexToHash("0xcccc"),
	}
	root := merkleRoot(leaves)
	if root == (common.Hash{}) {
		t.Error("merkle root should not be empty")
	}
}

func TestEncodeDecodeAgentTrieRoot(t *testing.T) {
	extra := []byte("vanity-data-here")
	root := common.HexToHash("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")

	encoded := EncodeAgentTrieRoot(extra, root)
	if len(encoded) != len(extra)+32 {
		t.Errorf("encoded length: got %d, want %d", len(encoded), len(extra)+32)
	}

	// Decode: vanityLen = len(extra), sealLen = 0
	decoded := DecodeAgentTrieRoot(encoded, len(extra), 0)
	if decoded != root {
		t.Errorf("decoded root mismatch: got %s, want %s", decoded.Hex(), root.Hex())
	}
}

func TestDecodeAgentTrieRootTooShort(t *testing.T) {
	extra := []byte("short")
	decoded := DecodeAgentTrieRoot(extra, 3, 2)
	if decoded != (common.Hash{}) {
		t.Error("too-short extra should return empty hash")
	}
}
