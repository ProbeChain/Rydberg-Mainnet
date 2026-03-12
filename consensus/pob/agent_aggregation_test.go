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
	"testing"

	"github.com/probechain/go-probe/common"
)

func TestHeartbeatBloomAddContains(t *testing.T) {
	bloom := NewHeartbeatBloom(100, 200)
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	addr3 := common.HexToAddress("0x3333333333333333333333333333333333333333")

	bloom.Add(addr1, 150)
	bloom.Add(addr2, 150)

	if !bloom.Contains(addr1, 150) {
		t.Error("bloom should contain addr1 at block 150")
	}
	if !bloom.Contains(addr2, 150) {
		t.Error("bloom should contain addr2 at block 150")
	}
	// addr3 was never added — should (almost certainly) not be found
	if bloom.Contains(addr3, 150) {
		t.Log("false positive for addr3 (acceptable but unlikely with only 2 entries)")
	}
	// Same addr, different block — should not match
	if bloom.Contains(addr1, 999) {
		t.Log("false positive for addr1 at block 999")
	}
	if bloom.Count != 2 {
		t.Errorf("count: got %d, want 2", bloom.Count)
	}
}

func TestHeartbeatBloomSize(t *testing.T) {
	bloom := NewHeartbeatBloom(0, 100)
	if bloom.Size() != HeartbeatBloomSize {
		t.Errorf("size: got %d, want %d", bloom.Size(), HeartbeatBloomSize)
	}
}

func TestHeartbeatBloomFPR(t *testing.T) {
	bloom := NewHeartbeatBloom(0, 1000)
	// Empty bloom should have 0 FPR
	if bloom.FalsePositiveRate() != 0 {
		t.Errorf("empty bloom FPR should be 0, got %f", bloom.FalsePositiveRate())
	}

	// Add 1000 entries
	for i := uint64(0); i < 1000; i++ {
		var addr common.Address
		addr[0] = byte(i >> 8)
		addr[1] = byte(i)
		bloom.Add(addr, i)
	}

	fpr := bloom.FalsePositiveRate()
	if fpr > 0.001 {
		t.Errorf("FPR with 1000 entries should be very low, got %f", fpr)
	}
}

func TestHeartbeatBloomBlockRange(t *testing.T) {
	bloom := NewHeartbeatBloom(500, 600)
	if bloom.BlockRange[0] != 500 || bloom.BlockRange[1] != 600 {
		t.Errorf("block range: got [%d, %d], want [500, 600]", bloom.BlockRange[0], bloom.BlockRange[1])
	}
}

func TestAggregatedAttestationParticipants(t *testing.T) {
	relay := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	taskID := common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	att := NewAggregatedAttestation(relay, taskID, 1000)

	agent1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	agent2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	agent3 := common.HexToAddress("0x3333333333333333333333333333333333333333")

	att.AddParticipant(agent1)
	att.AddParticipant(agent2)

	if !att.IsParticipant(agent1) {
		t.Error("agent1 should be a participant")
	}
	if !att.IsParticipant(agent2) {
		t.Error("agent2 should be a participant")
	}
	if att.IsParticipant(agent3) {
		t.Log("false positive for agent3 (acceptable)")
	}
	if att.AgentCount != 2 {
		t.Errorf("agent count: got %d, want 2", att.AgentCount)
	}
}

func TestAggregatedAttestationHash(t *testing.T) {
	relay := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	taskID := common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	att := NewAggregatedAttestation(relay, taskID, 1000)
	att.ResultHash = common.HexToHash("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
	att.Confidence = 9500

	hash1 := att.Hash()
	hash2 := att.Hash()

	if hash1 != hash2 {
		t.Error("hash should be deterministic")
	}
	if hash1 == (common.Hash{}) {
		t.Error("hash should not be empty")
	}

	// Different confidence should produce different hash
	att.Confidence = 5000
	hash3 := att.Hash()
	if hash3 == hash1 {
		t.Error("different confidence should produce different hash")
	}
}

func TestAggregatedAttestationFields(t *testing.T) {
	relay := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	taskID := common.HexToHash("0xbbbb")
	att := NewAggregatedAttestation(relay, taskID, 42)

	if att.RelayAddr != relay {
		t.Error("relay address mismatch")
	}
	if att.TaskID != taskID {
		t.Error("task ID mismatch")
	}
	if att.BlockNumber != 42 {
		t.Error("block number mismatch")
	}
	if len(att.ParticipantBloom) != ParticipantBloomSize {
		t.Errorf("participant bloom size: got %d, want %d", len(att.ParticipantBloom), ParticipantBloomSize)
	}
}
