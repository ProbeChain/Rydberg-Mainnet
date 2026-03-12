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
	"github.com/probechain/go-probe/crypto"
)

func TestNodeTypeString(t *testing.T) {
	if NodeTypeAgent.String() != "Agent" {
		t.Error("Agent type string mismatch")
	}
	if NodeTypePhysical.String() != "Physical" {
		t.Error("Physical type string mismatch")
	}
	if NodeTypeUnknown.String() != "Unknown" {
		t.Error("Unknown type string mismatch")
	}
}

func TestRegistrationChallenge(t *testing.T) {
	blockHash := common.HexToHash("0xaabb")
	applicant := common.HexToAddress("0x1111111111111111111111111111111111111111")

	c1 := RegistrationChallenge(blockHash, 100, applicant)
	c2 := RegistrationChallenge(blockHash, 100, applicant)

	if c1 != c2 {
		t.Error("same inputs should produce same challenge")
	}
	if c1 == (common.Hash{}) {
		t.Error("challenge should not be empty")
	}

	// Different block number → different challenge
	c3 := RegistrationChallenge(blockHash, 101, applicant)
	if c1 == c3 {
		t.Error("different block should produce different challenge")
	}
}

func TestAgentRegistrationProof(t *testing.T) {
	// Generate agent key
	agentKey, _ := crypto.GenerateKey()
	agentAddr := crypto.PubkeyToAddress(agentKey.PublicKey)

	// Generate operator key (must be different)
	operatorKey, _ := crypto.GenerateKey()
	operatorAddr := crypto.PubkeyToAddress(operatorKey.PublicKey)

	agentID := common.HexToHash("0xdeadbeef")
	blockHash := common.HexToHash("0xaabb")
	challenge := RegistrationChallenge(blockHash, 100, agentAddr)

	// Agent signs the registration message
	msg := MakeAgentRegistrationMessage(agentID, operatorAddr, challenge)
	sig, err := crypto.Sign(msg[:], agentKey)
	if err != nil {
		t.Fatalf("sign error: %v", err)
	}

	proof := &AgentRegistrationProof{
		AgentID:         agentID,
		AgentAddress:    agentAddr,
		OperatorAddress: operatorAddr,
		Challenge:       challenge,
		Signature:       sig,
	}

	// Should verify successfully
	recovered, err := proof.Verify(challenge)
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if recovered != agentAddr {
		t.Errorf("recovered address mismatch: got %s, want %s", recovered.Hex(), agentAddr.Hex())
	}
}

func TestAgentRegistration_WrongChallenge(t *testing.T) {
	agentKey, _ := crypto.GenerateKey()
	agentAddr := crypto.PubkeyToAddress(agentKey.PublicKey)
	operatorKey, _ := crypto.GenerateKey()
	operatorAddr := crypto.PubkeyToAddress(operatorKey.PublicKey)

	agentID := common.HexToHash("0xdeadbeef")
	challenge := RegistrationChallenge(common.Hash{}, 100, agentAddr)

	msg := MakeAgentRegistrationMessage(agentID, operatorAddr, challenge)
	sig, _ := crypto.Sign(msg[:], agentKey)

	proof := &AgentRegistrationProof{
		AgentID:         agentID,
		AgentAddress:    agentAddr,
		OperatorAddress: operatorAddr,
		Challenge:       challenge,
		Signature:       sig,
	}

	// Wrong challenge should fail
	wrongChallenge := common.HexToHash("0xwrongwrongwrongwrong")
	_, err := proof.Verify(wrongChallenge)
	if err != ErrChallengeMismatch {
		t.Errorf("expected ErrChallengeMismatch, got %v", err)
	}
}

func TestAgentRegistration_SameKeyRejected(t *testing.T) {
	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)

	proof := &AgentRegistrationProof{
		AgentID:         common.HexToHash("0xdeadbeef"),
		AgentAddress:    addr,
		OperatorAddress: addr, // Same as agent — should be rejected
		Challenge:       common.Hash{},
		Signature:       make([]byte, 65),
	}

	_, err := proof.Verify(common.Hash{})
	if err != ErrOperatorKeyIsAgentKey {
		t.Errorf("expected ErrOperatorKeyIsAgentKey, got %v", err)
	}
}

func TestPhysicalRegistrationProof(t *testing.T) {
	deviceKey, _ := crypto.GenerateKey()
	operatorAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")

	deviceID := MakeDeviceFingerprint(
		[]byte("Intel-i9-13900K"),
		[]byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
		[]byte("DISK-SERIAL-123"),
		[]byte("BOARD-SERIAL-456"),
	)

	challenge := RegistrationChallenge(common.Hash{1}, 50, operatorAddr)
	timestamp := uint64(1700000000)

	msg := MakePhysicalRegistrationMessage(deviceID, operatorAddr, challenge, timestamp)
	sig, _ := crypto.Sign(msg[:], deviceKey)

	proof := &PhysicalRegistrationProof{
		DeviceID:          deviceID,
		DeviceType:        DeviceClassDesktop,
		OperatorAddress:   operatorAddr,
		Challenge:         challenge,
		DeviceAttestation: sig,
		Timestamp:         timestamp,
		HardwareReport: &HardwareReport{
			CPUModel:            "Intel i9-13900K",
			CPUCores:            24,
			RAMBytes:            64 * 1024 * 1024 * 1024,
			DiskBytes:           2 * 1024 * 1024 * 1024 * 1024,
			OSType:              "linux",
			VirtualizationFlags: VirtFlagNone,
		},
	}

	err := proof.Verify(challenge)
	if err != nil {
		t.Fatalf("physical verify error: %v", err)
	}
}

func TestPhysicalRegistration_VirtualDeviceRejected(t *testing.T) {
	deviceKey, _ := crypto.GenerateKey()
	operatorAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	deviceID := MakeDeviceFingerprint([]byte("vm-cpu"), nil, nil, nil)
	challenge := RegistrationChallenge(common.Hash{1}, 50, operatorAddr)

	msg := MakePhysicalRegistrationMessage(deviceID, operatorAddr, challenge, 1700000000)
	sig, _ := crypto.Sign(msg[:], deviceKey)

	proof := &PhysicalRegistrationProof{
		DeviceID:          deviceID,
		OperatorAddress:   operatorAddr,
		Challenge:         challenge,
		DeviceAttestation: sig,
		Timestamp:         1700000000,
		HardwareReport: &HardwareReport{
			CPUCores:            4,
			RAMBytes:            8 * 1024 * 1024 * 1024,
			VirtualizationFlags: VirtFlagVMWare, // VMware detected!
		},
	}

	err := proof.Verify(challenge)
	if err != ErrVirtualDeviceDetected {
		t.Errorf("expected ErrVirtualDeviceDetected, got %v", err)
	}
}

func TestPhysicalRegistration_EmptyFingerprint(t *testing.T) {
	proof := &PhysicalRegistrationProof{
		DeviceID:  DeviceFingerprint{}, // Empty
		Challenge: common.Hash{},
	}
	err := proof.Verify(common.Hash{})
	if err != ErrDeviceFingerprintEmpty {
		t.Errorf("expected ErrDeviceFingerprintEmpty, got %v", err)
	}
}

func TestSybilDetector_IPRateLimit(t *testing.T) {
	sd := NewSybilDetector()

	// Register MaxRegistrationsPerIP agents from same IP
	for i := 0; i < MaxRegistrationsPerIP; i++ {
		agentKey, _ := crypto.GenerateKey()
		agentAddr := crypto.PubkeyToAddress(agentKey.PublicKey)
		operatorKey, _ := crypto.GenerateKey()
		operatorAddr := crypto.PubkeyToAddress(operatorKey.PublicKey)

		agentID := common.BigToHash(common.Big1)
		agentID[0] = byte(i)

		challenge := RegistrationChallenge(common.Hash{}, uint64(i), agentAddr)
		msg := MakeAgentRegistrationMessage(agentID, operatorAddr, challenge)
		sig, _ := crypto.Sign(msg[:], agentKey)

		proof := &AgentRegistrationProof{
			AgentID:         agentID,
			AgentAddress:    agentAddr,
			OperatorAddress: operatorAddr,
			Challenge:       challenge,
			Signature:       sig,
		}

		err := sd.ValidateAgentRegistration(proof, challenge, uint64(i), uint64(i), "192.168.1.1")
		if err != nil {
			t.Fatalf("registration %d failed: %v", i, err)
		}
	}

	// Next registration from same IP should be rate limited
	agentKey, _ := crypto.GenerateKey()
	agentAddr := crypto.PubkeyToAddress(agentKey.PublicKey)
	operatorKey, _ := crypto.GenerateKey()
	operatorAddr := crypto.PubkeyToAddress(operatorKey.PublicKey)

	agentID := common.HexToHash("0xff")
	challenge := RegistrationChallenge(common.Hash{}, 999, agentAddr)
	msg := MakeAgentRegistrationMessage(agentID, operatorAddr, challenge)
	sig, _ := crypto.Sign(msg[:], agentKey)

	proof := &AgentRegistrationProof{
		AgentID:         agentID,
		AgentAddress:    agentAddr,
		OperatorAddress: operatorAddr,
		Challenge:       challenge,
		Signature:       sig,
	}

	err := sd.ValidateAgentRegistration(proof, challenge, 999, 999, "192.168.1.1")
	if err != ErrIPRateLimited {
		t.Errorf("expected ErrIPRateLimited, got %v", err)
	}
}

func TestSybilDetector_DuplicateDevice(t *testing.T) {
	sd := NewSybilDetector()

	deviceKey, _ := crypto.GenerateKey()
	deviceID := MakeDeviceFingerprint([]byte("unique-cpu"), []byte{1, 2, 3, 4, 5, 6}, nil, nil)

	op1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	challenge1 := RegistrationChallenge(common.Hash{}, 100, op1)
	msg1 := MakePhysicalRegistrationMessage(deviceID, op1, challenge1, 1700000000)
	sig1, _ := crypto.Sign(msg1[:], deviceKey)

	proof1 := &PhysicalRegistrationProof{
		DeviceID: deviceID, OperatorAddress: op1,
		Challenge: challenge1, DeviceAttestation: sig1, Timestamp: 1700000000,
		HardwareReport: &HardwareReport{CPUCores: 4, RAMBytes: 8e9},
	}

	err := sd.ValidatePhysicalRegistration(proof1, challenge1, 100, 100, "10.0.0.1")
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	// Same device, different operator → rejected
	op2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	challenge2 := RegistrationChallenge(common.Hash{}, 101, op2)
	msg2 := MakePhysicalRegistrationMessage(deviceID, op2, challenge2, 1700000001)
	sig2, _ := crypto.Sign(msg2[:], deviceKey)

	proof2 := &PhysicalRegistrationProof{
		DeviceID: deviceID, OperatorAddress: op2,
		Challenge: challenge2, DeviceAttestation: sig2, Timestamp: 1700000001,
		HardwareReport: &HardwareReport{CPUCores: 4, RAMBytes: 8e9},
	}

	err = sd.ValidatePhysicalRegistration(proof2, challenge2, 101, 101, "10.0.0.2")
	if err != ErrDeviceDuplicate {
		t.Errorf("expected ErrDeviceDuplicate, got %v", err)
	}
}

func TestSybilDetector_DuplicateAgentID(t *testing.T) {
	sd := NewSybilDetector()
	agentID := common.HexToHash("0xdeadbeef")

	// First registration
	key1, _ := crypto.GenerateKey()
	addr1 := crypto.PubkeyToAddress(key1.PublicKey)
	opKey1, _ := crypto.GenerateKey()
	opAddr1 := crypto.PubkeyToAddress(opKey1.PublicKey)

	challenge1 := RegistrationChallenge(common.Hash{}, 100, addr1)
	msg1 := MakeAgentRegistrationMessage(agentID, opAddr1, challenge1)
	sig1, _ := crypto.Sign(msg1[:], key1)

	proof1 := &AgentRegistrationProof{
		AgentID: agentID, AgentAddress: addr1, OperatorAddress: opAddr1,
		Challenge: challenge1, Signature: sig1,
	}
	err := sd.ValidateAgentRegistration(proof1, challenge1, 100, 100, "")
	if err != nil {
		t.Fatalf("first agent registration failed: %v", err)
	}

	// Same AgentID, different agent address → rejected
	key2, _ := crypto.GenerateKey()
	addr2 := crypto.PubkeyToAddress(key2.PublicKey)
	opKey2, _ := crypto.GenerateKey()
	opAddr2 := crypto.PubkeyToAddress(opKey2.PublicKey)

	challenge2 := RegistrationChallenge(common.Hash{}, 101, addr2)
	msg2 := MakeAgentRegistrationMessage(agentID, opAddr2, challenge2)
	sig2, _ := crypto.Sign(msg2[:], key2)

	proof2 := &AgentRegistrationProof{
		AgentID: agentID, AgentAddress: addr2, OperatorAddress: opAddr2,
		Challenge: challenge2, Signature: sig2,
	}
	err = sd.ValidateAgentRegistration(proof2, challenge2, 101, 101, "")
	if err != ErrAgentAlreadyRegistered {
		t.Errorf("expected ErrAgentAlreadyRegistered, got %v", err)
	}
}

func TestNodeRegistrationRequest_Validate(t *testing.T) {
	// Invalid type
	req := &NodeRegistrationRequest{NodeType: NodeTypeUnknown}
	if err := req.Validate(); err != ErrInvalidNodeType {
		t.Errorf("expected ErrInvalidNodeType, got %v", err)
	}

	// Agent without proof
	req = &NodeRegistrationRequest{NodeType: NodeTypeAgent}
	if err := req.Validate(); err == nil {
		t.Error("agent without proof should fail")
	}

	// Physical without proof
	req = &NodeRegistrationRequest{NodeType: NodeTypePhysical}
	if err := req.Validate(); err == nil {
		t.Error("physical without proof should fail")
	}

	// Agent with both proofs
	req = &NodeRegistrationRequest{
		NodeType:      NodeTypeAgent,
		AgentProof:    &AgentRegistrationProof{},
		PhysicalProof: &PhysicalRegistrationProof{},
	}
	if err := req.Validate(); err == nil {
		t.Error("agent with physical proof should fail")
	}
}

func TestCorrelationScore(t *testing.T) {
	a := []common.Hash{
		common.HexToHash("0x01"),
		common.HexToHash("0x02"),
		common.HexToHash("0x03"),
	}
	b := []common.Hash{
		common.HexToHash("0x01"),
		common.HexToHash("0x02"),
		common.HexToHash("0x03"),
	}

	score := CorrelationScore(a, b)
	if score != 10000 {
		t.Errorf("identical votes should score 10000, got %d", score)
	}

	c := []common.Hash{
		common.HexToHash("0x04"),
		common.HexToHash("0x05"),
		common.HexToHash("0x06"),
	}
	score2 := CorrelationScore(a, c)
	if score2 != 0 {
		t.Errorf("no overlap should score 0, got %d", score2)
	}
}

func TestIsKnownVMMAC(t *testing.T) {
	// VMware MAC
	if !IsKnownVMMAC([6]byte{0x00, 0x50, 0x56, 0xAA, 0xBB, 0xCC}) {
		t.Error("should detect VMware MAC")
	}
	// VirtualBox MAC
	if !IsKnownVMMAC([6]byte{0x08, 0x00, 0x27, 0x11, 0x22, 0x33}) {
		t.Error("should detect VirtualBox MAC")
	}
	// Real MAC
	if IsKnownVMMAC([6]byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}) {
		t.Error("should not flag real MAC")
	}
}

func TestMakeDeviceFingerprint(t *testing.T) {
	fp1 := MakeDeviceFingerprint([]byte("cpu1"), []byte{1, 2, 3}, []byte("disk1"), []byte("board1"))
	fp2 := MakeDeviceFingerprint([]byte("cpu1"), []byte{1, 2, 3}, []byte("disk1"), []byte("board1"))
	fp3 := MakeDeviceFingerprint([]byte("cpu2"), []byte{1, 2, 3}, []byte("disk1"), []byte("board1"))

	if fp1 != fp2 {
		t.Error("same hardware should produce same fingerprint")
	}
	if fp1 == fp3 {
		t.Error("different hardware should produce different fingerprint")
	}
	if fp1 == (DeviceFingerprint{}) {
		t.Error("fingerprint should not be empty")
	}
}

func TestSybilDetector_Stats(t *testing.T) {
	sd := NewSybilDetector()
	agents, devices, ips := sd.Stats()
	if agents != 0 || devices != 0 || ips != 0 {
		t.Error("new detector should have zero stats")
	}
}

func TestNeedsReverification(t *testing.T) {
	agent := &NodeIdentity{NodeType: NodeTypeAgent, LastVerified: 0}
	if !agent.NeedsReverification(100001) {
		t.Error("agent should need reverification after 100K blocks")
	}
	if agent.NeedsReverification(50000) {
		t.Error("agent should not need reverification within 100K blocks")
	}

	physical := &NodeIdentity{NodeType: NodeTypePhysical, LastVerified: 0}
	if !physical.NeedsReverification(250001) {
		t.Error("physical should need reverification after 250K blocks")
	}
}
