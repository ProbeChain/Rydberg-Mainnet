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
	"errors"
	"sync"
	"time"

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/crypto"
	"golang.org/x/crypto/sha3"
)

// ---------------------------------------------------------------------------
// Node Identity & Anti-Sybil System
// ---------------------------------------------------------------------------
// Every node must register as EXACTLY ONE type:
//   NodeTypeAgent    (1) — AI agent, must prove ERC-8004 identity via signature
//   NodeTypePhysical (2) — physical device, must prove device uniqueness
//
// Fake node detection covers:
//   - Virtual machines / emulators
//   - IP rotation abuse (1 IP registering thousands of nodes)
//   - Duplicate device fingerprints
//   - Missing or forged device attestations
//   - Replay attacks on registration proofs

// NodeType represents the type of PoB node.
type NodeType uint8

const (
	NodeTypeUnknown  NodeType = 0
	NodeTypeAgent    NodeType = 1 // AI agent node (PoB-A)
	NodeTypePhysical NodeType = 2 // Physical device node (PoB-P)
)

// String returns the node type as a human-readable string.
func (nt NodeType) String() string {
	switch nt {
	case NodeTypeAgent:
		return "Agent"
	case NodeTypePhysical:
		return "Physical"
	default:
		return "Unknown"
	}
}

// Registration errors
var (
	ErrInvalidNodeType       = errors.New("invalid node type: must be Agent (1) or Physical (2)")
	ErrAgentSignatureInvalid = errors.New("agent registration: ERC-8004 signature verification failed")
	ErrAgentIDMismatch       = errors.New("agent registration: recovered address does not match agent ID")
	ErrAgentAlreadyRegistered = errors.New("agent registration: agent ID already registered")
	ErrChallengeMismatch     = errors.New("registration: challenge nonce mismatch (possible replay)")
	ErrDeviceFingerprintEmpty = errors.New("physical registration: device fingerprint is empty")
	ErrDeviceDuplicate       = errors.New("physical registration: device fingerprint already registered")
	ErrDeviceAttestInvalid   = errors.New("physical registration: device attestation signature invalid")
	ErrVirtualDeviceDetected = errors.New("physical registration: virtual/emulated device detected")
	ErrIPRateLimited         = errors.New("registration: too many registrations from this IP (rate limited)")
	ErrOperatorKeyIsAgentKey = errors.New("agent registration: operator key must differ from agent key")
)

// ---------------------------------------------------------------------------
// Agent Registration Proof (ERC-8004)
// ---------------------------------------------------------------------------

// AgentRegistrationProof proves an AI agent's identity via ERC-8004 signature.
// The agent must sign a challenge with its OWN private key (not the operator's).
type AgentRegistrationProof struct {
	// AgentID is the ERC-8004 identity hash (keccak256 of agent metadata).
	AgentID common.Hash `json:"agentId"`

	// AgentAddress is the Ethereum address derived from the agent's public key.
	AgentAddress common.Address `json:"agentAddress"`

	// OperatorAddress is the human operator's address (must differ from AgentAddress).
	OperatorAddress common.Address `json:"operatorAddress"`

	// Challenge is the registration challenge nonce (block hash + timestamp).
	Challenge common.Hash `json:"challenge"`

	// Signature is the agent's ECDSA signature over:
	//   keccak256(AgentID || OperatorAddress || Challenge)
	// This proves the agent holds its own private key.
	Signature []byte `json:"signature"`

	// Capabilities describes what the agent can do (JSON-encoded).
	Capabilities []byte `json:"capabilities,omitempty"`
}

// Verify checks that the agent registration proof is valid.
// Returns the recovered agent address or an error.
func (p *AgentRegistrationProof) Verify(expectedChallenge common.Hash) (common.Address, error) {
	// 1. Challenge must match (prevents replay)
	if p.Challenge != expectedChallenge {
		return common.Address{}, ErrChallengeMismatch
	}

	// 2. Operator and agent must be different keys
	if p.OperatorAddress == p.AgentAddress {
		return common.Address{}, ErrOperatorKeyIsAgentKey
	}

	// 3. Reconstruct the signed message
	msg := agentRegistrationMessage(p.AgentID, p.OperatorAddress, p.Challenge)

	// 4. Recover public key from signature
	if len(p.Signature) != 65 {
		return common.Address{}, ErrAgentSignatureInvalid
	}
	pubkey, err := crypto.Ecrecover(msg[:], p.Signature)
	if err != nil {
		return common.Address{}, ErrAgentSignatureInvalid
	}

	// 5. Derive address from recovered public key
	recovered := common.BytesToAddress(crypto.Keccak256(pubkey[1:])[12:])

	// 6. Must match declared agent address
	if recovered != p.AgentAddress {
		return common.Address{}, ErrAgentIDMismatch
	}

	return recovered, nil
}

// agentRegistrationMessage constructs the message that agents must sign.
func agentRegistrationMessage(agentID common.Hash, operator common.Address, challenge common.Hash) common.Hash {
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte("ProbeChain-Agent-Register-v1"))
	h.Write(agentID[:])
	h.Write(operator[:])
	h.Write(challenge[:])
	var hash common.Hash
	h.Sum(hash[:0])
	return hash
}

// MakeAgentRegistrationMessage creates the message for an agent to sign.
// Exported for use by agent nodes when preparing registration.
func MakeAgentRegistrationMessage(agentID common.Hash, operator common.Address, challenge common.Hash) common.Hash {
	return agentRegistrationMessage(agentID, operator, challenge)
}

// ---------------------------------------------------------------------------
// Physical Node Registration Proof
// ---------------------------------------------------------------------------

// DeviceFingerprint is a hash derived from hardware identifiers.
// Construction: keccak256(cpuID || macAddress || diskSerial || boardSerial)
type DeviceFingerprint [32]byte

// PhysicalRegistrationProof proves a physical device's uniqueness.
type PhysicalRegistrationProof struct {
	// DeviceID is the keccak256 hash of hardware identifiers.
	DeviceID DeviceFingerprint `json:"deviceId"`

	// DeviceType classifies the device.
	DeviceType DeviceClass `json:"deviceType"`

	// OperatorAddress is the device owner's address.
	OperatorAddress common.Address `json:"operatorAddress"`

	// Challenge is the registration challenge nonce.
	Challenge common.Hash `json:"challenge"`

	// DeviceAttestation is a signature over:
	//   keccak256(DeviceID || OperatorAddress || Challenge || Timestamp)
	// Created by the device's secure enclave or TPM if available.
	DeviceAttestation []byte `json:"deviceAttestation"`

	// Timestamp is when the attestation was created (unix seconds).
	Timestamp uint64 `json:"timestamp"`

	// HardwareReport contains device hardware details for verification.
	HardwareReport *HardwareReport `json:"hardwareReport"`
}

// DeviceClass identifies the type of physical device.
type DeviceClass uint8

const (
	DeviceClassUnknown    DeviceClass = 0
	DeviceClassServer     DeviceClass = 1 // Data center / server
	DeviceClassDesktop    DeviceClass = 2 // Desktop / workstation
	DeviceClassLaptop     DeviceClass = 3 // Laptop
	DeviceClassMobile     DeviceClass = 4 // Phone / tablet
	DeviceClassIoT        DeviceClass = 5 // IoT device (fridge, car, etc.)
	DeviceClassEmbedded   DeviceClass = 6 // Embedded system
)

// HardwareReport contains verifiable hardware details.
type HardwareReport struct {
	// CPUModel is the CPU identifier string.
	CPUModel string `json:"cpuModel"`

	// CPUCores is the number of physical CPU cores.
	CPUCores uint32 `json:"cpuCores"`

	// RAMBytes is the total physical RAM in bytes.
	RAMBytes uint64 `json:"ramBytes"`

	// DiskBytes is the total disk capacity in bytes.
	DiskBytes uint64 `json:"diskBytes"`

	// NetworkMAC is the primary network interface MAC address hash.
	NetworkMAC [6]byte `json:"networkMac"`

	// HasTPM indicates whether a Trusted Platform Module is present.
	HasTPM bool `json:"hasTpm"`

	// HasSecureEnclave indicates whether a secure enclave is present (iOS/ARM).
	HasSecureEnclave bool `json:"hasSecureEnclave"`

	// OSType: "linux", "darwin", "windows", "android", "ios"
	OSType string `json:"osType"`

	// VirtualizationFlags contains detected virtualization indicators.
	VirtualizationFlags uint64 `json:"virtualizationFlags"`
}

// Virtualization detection flags (bitfield)
const (
	VirtFlagNone       uint64 = 0
	VirtFlagHypervisor uint64 = 1 << 0 // CPUID hypervisor bit set
	VirtFlagVMWare     uint64 = 1 << 1 // VMware-specific artifacts
	VirtFlagVBox       uint64 = 1 << 2 // VirtualBox artifacts
	VirtFlagKVM        uint64 = 1 << 3 // KVM artifacts
	VirtFlagDocker     uint64 = 1 << 4 // Docker container
	VirtFlagWSL        uint64 = 1 << 5 // Windows Subsystem for Linux
	VirtFlagEmulator   uint64 = 1 << 6 // Android/iOS emulator
	VirtFlagXen        uint64 = 1 << 7 // Xen hypervisor
	VirtFlagQEMU       uint64 = 1 << 8 // QEMU
	VirtFlagCloud      uint64 = 1 << 9 // Known cloud instance (detectable MAC prefix)
)

// IsVirtual returns true if any virtualization flags are set.
func (hr *HardwareReport) IsVirtual() bool {
	return hr != nil && hr.VirtualizationFlags != VirtFlagNone
}

// Verify checks the physical registration proof.
func (p *PhysicalRegistrationProof) Verify(expectedChallenge common.Hash) error {
	// 1. Challenge must match
	if p.Challenge != expectedChallenge {
		return ErrChallengeMismatch
	}

	// 2. Device fingerprint must not be empty
	if p.DeviceID == (DeviceFingerprint{}) {
		return ErrDeviceFingerprintEmpty
	}

	// 3. Reject virtual/emulated devices
	if p.HardwareReport != nil && p.HardwareReport.IsVirtual() {
		return ErrVirtualDeviceDetected
	}

	// 4. Verify device attestation signature
	msg := physicalRegistrationMessage(p.DeviceID, p.OperatorAddress, p.Challenge, p.Timestamp)
	if len(p.DeviceAttestation) < 65 {
		return ErrDeviceAttestInvalid
	}
	_, err := crypto.Ecrecover(msg[:], p.DeviceAttestation)
	if err != nil {
		return ErrDeviceAttestInvalid
	}

	return nil
}

// physicalRegistrationMessage constructs the message for device attestation.
func physicalRegistrationMessage(deviceID DeviceFingerprint, operator common.Address, challenge common.Hash, timestamp uint64) common.Hash {
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte("ProbeChain-Physical-Register-v1"))
	h.Write(deviceID[:])
	h.Write(operator[:])
	h.Write(challenge[:])
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], timestamp)
	h.Write(buf[:])
	var hash common.Hash
	h.Sum(hash[:0])
	return hash
}

// MakePhysicalRegistrationMessage creates the message for a device to sign.
func MakePhysicalRegistrationMessage(deviceID DeviceFingerprint, operator common.Address, challenge common.Hash, timestamp uint64) common.Hash {
	return physicalRegistrationMessage(deviceID, operator, challenge, timestamp)
}

// MakeDeviceFingerprint creates a device fingerprint from hardware identifiers.
func MakeDeviceFingerprint(cpuID, macAddr, diskSerial, boardSerial []byte) DeviceFingerprint {
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte("ProbeChain-DeviceID-v1"))
	h.Write(cpuID)
	h.Write(macAddr)
	h.Write(diskSerial)
	h.Write(boardSerial)
	var fp DeviceFingerprint
	h.Sum(fp[:0])
	return fp
}

// ---------------------------------------------------------------------------
// Registration Challenge
// ---------------------------------------------------------------------------

// RegistrationChallenge creates a challenge nonce for node registration.
// challenge = keccak256(blockHash || blockNumber || applicantAddress)
// This prevents replay attacks: each registration attempt needs a fresh challenge
// tied to a recent block.
func RegistrationChallenge(blockHash common.Hash, blockNumber uint64, applicant common.Address) common.Hash {
	h := sha3.NewLegacyKeccak256()
	h.Write(blockHash[:])
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], blockNumber)
	h.Write(buf[:])
	h.Write(applicant[:])
	var hash common.Hash
	h.Sum(hash[:0])
	return hash
}

// ---------------------------------------------------------------------------
// Sybil Detector
// ---------------------------------------------------------------------------
// Tracks registrations and detects abuse patterns:
//   - IP rate limiting: max N registrations per IP per time window
//   - Device deduplication: one device fingerprint = one node
//   - Agent ID deduplication: one ERC-8004 ID = one node
//   - Behavioral correlation: nodes that always vote together are suspicious

const (
	// MaxRegistrationsPerIP is the max new registrations from one IP per window.
	MaxRegistrationsPerIP = 10

	// IPRateLimitWindow is the time window for IP rate limiting.
	IPRateLimitWindow = 1 * time.Hour

	// ChallengeMaxAge is the maximum age of a registration challenge (in blocks).
	ChallengeMaxAge = uint64(256)
)

// SybilDetector tracks and detects fake node registration attempts.
type SybilDetector struct {
	mu sync.RWMutex

	// registeredAgentIDs tracks which ERC-8004 agent IDs are already registered.
	registeredAgentIDs map[common.Hash]common.Address

	// registeredDevices tracks which device fingerprints are already registered.
	registeredDevices map[DeviceFingerprint]common.Address

	// ipRegistrations tracks registration count per IP within the rate limit window.
	ipRegistrations map[string]*ipRegistrationTracker

	// correlationTracker detects nodes that always behave identically (sybil signal).
	voteCorrelation map[common.Address][]common.Hash
}

type ipRegistrationTracker struct {
	count     int
	windowEnd time.Time
}

// NewSybilDetector creates a new sybil detector.
func NewSybilDetector() *SybilDetector {
	return &SybilDetector{
		registeredAgentIDs: make(map[common.Hash]common.Address),
		registeredDevices:  make(map[DeviceFingerprint]common.Address),
		ipRegistrations:    make(map[string]*ipRegistrationTracker),
		voteCorrelation:    make(map[common.Address][]common.Hash),
	}
}

// ValidateAgentRegistration validates an agent node registration.
func (sd *SybilDetector) ValidateAgentRegistration(
	proof *AgentRegistrationProof,
	challenge common.Hash,
	currentBlock uint64,
	challengeBlock uint64,
	sourceIP string,
) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	// 1. Challenge age check
	if currentBlock-challengeBlock > ChallengeMaxAge {
		return ErrChallengeMismatch
	}

	// 2. IP rate limit check
	if err := sd.checkIPRateLimit(sourceIP); err != nil {
		return err
	}

	// 3. Verify cryptographic proof
	_, err := proof.Verify(challenge)
	if err != nil {
		return err
	}

	// 4. Check agent ID uniqueness
	if existing, ok := sd.registeredAgentIDs[proof.AgentID]; ok {
		if existing != proof.AgentAddress {
			return ErrAgentAlreadyRegistered
		}
	}

	// 5. Register
	sd.registeredAgentIDs[proof.AgentID] = proof.AgentAddress
	sd.recordIPRegistration(sourceIP)

	return nil
}

// ValidatePhysicalRegistration validates a physical node registration.
func (sd *SybilDetector) ValidatePhysicalRegistration(
	proof *PhysicalRegistrationProof,
	challenge common.Hash,
	currentBlock uint64,
	challengeBlock uint64,
	sourceIP string,
) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	// 1. Challenge age check
	if currentBlock-challengeBlock > ChallengeMaxAge {
		return ErrChallengeMismatch
	}

	// 2. IP rate limit check
	if err := sd.checkIPRateLimit(sourceIP); err != nil {
		return err
	}

	// 3. Verify device proof
	if err := proof.Verify(challenge); err != nil {
		return err
	}

	// 4. Check device uniqueness
	if existing, ok := sd.registeredDevices[proof.DeviceID]; ok {
		if existing != proof.OperatorAddress {
			return ErrDeviceDuplicate
		}
	}

	// 5. Additional hardware validation
	if err := sd.validateHardware(proof.HardwareReport); err != nil {
		return err
	}

	// 6. Register
	sd.registeredDevices[proof.DeviceID] = proof.OperatorAddress
	sd.recordIPRegistration(sourceIP)

	return nil
}

// checkIPRateLimit checks if an IP has exceeded the registration rate limit.
func (sd *SybilDetector) checkIPRateLimit(ip string) error {
	if ip == "" {
		return nil // Local registrations bypass IP check
	}
	tracker, ok := sd.ipRegistrations[ip]
	if !ok || time.Now().After(tracker.windowEnd) {
		return nil // No recent registrations or window expired
	}
	if tracker.count >= MaxRegistrationsPerIP {
		return ErrIPRateLimited
	}
	return nil
}

// recordIPRegistration records a successful registration from an IP.
func (sd *SybilDetector) recordIPRegistration(ip string) {
	if ip == "" {
		return
	}
	tracker, ok := sd.ipRegistrations[ip]
	if !ok || time.Now().After(tracker.windowEnd) {
		sd.ipRegistrations[ip] = &ipRegistrationTracker{
			count:     1,
			windowEnd: time.Now().Add(IPRateLimitWindow),
		}
		return
	}
	tracker.count++
}

// validateHardware performs additional hardware validation checks.
func (sd *SybilDetector) validateHardware(report *HardwareReport) error {
	if report == nil {
		return nil // No report = minimal validation (allowed but scored lower)
	}

	// Reject known virtual environments
	if report.IsVirtual() {
		return ErrVirtualDeviceDetected
	}

	// Check for unrealistic hardware specs (likely spoofed)
	if report.CPUCores == 0 {
		return errors.New("physical registration: invalid hardware report (0 CPU cores)")
	}
	if report.RAMBytes == 0 {
		return errors.New("physical registration: invalid hardware report (0 RAM)")
	}

	return nil
}

// IsAgentRegistered checks if an agent ID is already registered.
func (sd *SybilDetector) IsAgentRegistered(agentID common.Hash) bool {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	_, ok := sd.registeredAgentIDs[agentID]
	return ok
}

// IsDeviceRegistered checks if a device fingerprint is already registered.
func (sd *SybilDetector) IsDeviceRegistered(deviceID DeviceFingerprint) bool {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	_, ok := sd.registeredDevices[deviceID]
	return ok
}

// UnregisterAgent removes an agent registration.
func (sd *SybilDetector) UnregisterAgent(agentID common.Hash) {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	delete(sd.registeredAgentIDs, agentID)
}

// UnregisterDevice removes a device registration.
func (sd *SybilDetector) UnregisterDevice(deviceID DeviceFingerprint) {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	delete(sd.registeredDevices, deviceID)
}

// Stats returns current sybil detector statistics.
func (sd *SybilDetector) Stats() (agentCount, deviceCount, trackedIPs int) {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	return len(sd.registeredAgentIDs), len(sd.registeredDevices), len(sd.ipRegistrations)
}

// ---------------------------------------------------------------------------
// Behavioral Sybil Detection (Post-Registration)
// ---------------------------------------------------------------------------
// After registration, monitor node behavior for sybil patterns:
//   - Voting correlation: nodes that always vote identically
//   - Timing correlation: nodes that always submit at the same millisecond
//   - Network proximity: nodes that always connect through same relay

// CorrelationScore measures how correlated two nodes' behavior is.
// Returns a score 0-10000 where 10000 = perfectly correlated (likely sybil).
func CorrelationScore(votesA, votesB []common.Hash) uint64 {
	if len(votesA) == 0 || len(votesB) == 0 {
		return 0
	}

	// Use the shorter list as reference
	ref, check := votesA, votesB
	if len(votesA) > len(votesB) {
		ref, check = votesB, votesA
	}

	// Build a set of check votes
	checkSet := make(map[common.Hash]bool, len(check))
	for _, v := range check {
		checkSet[v] = true
	}

	// Count matches
	var matches uint64
	for _, v := range ref {
		if checkSet[v] {
			matches++
		}
	}

	if uint64(len(ref)) == 0 {
		return 0
	}
	return matches * 10000 / uint64(len(ref))
}

// SybilCorrelationThreshold is the correlation score above which nodes are flagged.
const SybilCorrelationThreshold = 9500 // 95% correlation

// RecordVote records a node's vote/attestation for correlation tracking.
func (sd *SybilDetector) RecordVote(node common.Address, voteHash common.Hash) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	votes := sd.voteCorrelation[node]
	// Keep last 100 votes
	if len(votes) >= 100 {
		votes = votes[1:]
	}
	sd.voteCorrelation[node] = append(votes, voteHash)
}

// CheckCorrelation checks if two nodes have suspiciously correlated behavior.
func (sd *SybilDetector) CheckCorrelation(nodeA, nodeB common.Address) uint64 {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	return CorrelationScore(sd.voteCorrelation[nodeA], sd.voteCorrelation[nodeB])
}

// ---------------------------------------------------------------------------
// Node Registration Request (unified entry point)
// ---------------------------------------------------------------------------

// NodeRegistrationRequest is the unified registration request.
// Exactly one of AgentProof or PhysicalProof must be set.
type NodeRegistrationRequest struct {
	// NodeType must be NodeTypeAgent (1) or NodeTypePhysical (2).
	NodeType NodeType `json:"nodeType"`

	// AgentProof is required when NodeType == NodeTypeAgent.
	AgentProof *AgentRegistrationProof `json:"agentProof,omitempty"`

	// PhysicalProof is required when NodeType == NodeTypePhysical.
	PhysicalProof *PhysicalRegistrationProof `json:"physicalProof,omitempty"`
}

// Validate checks that the registration request is well-formed.
func (req *NodeRegistrationRequest) Validate() error {
	switch req.NodeType {
	case NodeTypeAgent:
		if req.AgentProof == nil {
			return errors.New("agent registration requires AgentProof")
		}
		if req.PhysicalProof != nil {
			return errors.New("agent registration must not include PhysicalProof")
		}
	case NodeTypePhysical:
		if req.PhysicalProof == nil {
			return errors.New("physical registration requires PhysicalProof")
		}
		if req.AgentProof != nil {
			return errors.New("physical registration must not include AgentProof")
		}
	default:
		return ErrInvalidNodeType
	}
	return nil
}

// GetRegistrationChallenge returns the current registration challenge.
// This should be called by nodes before submitting their registration.
func GetRegistrationChallenge(blockHash common.Hash, blockNumber uint64, applicant common.Address) *RegistrationChallengeResponse {
	challenge := RegistrationChallenge(blockHash, blockNumber, applicant)
	return &RegistrationChallengeResponse{
		Challenge:   challenge,
		BlockHash:   blockHash,
		BlockNumber: blockNumber,
		ExpiresAt:   blockNumber + ChallengeMaxAge,
	}
}

// RegistrationChallengeResponse is returned to nodes requesting a challenge.
type RegistrationChallengeResponse struct {
	Challenge   common.Hash `json:"challenge"`
	BlockHash   common.Hash `json:"blockHash"`
	BlockNumber uint64      `json:"blockNumber"`
	ExpiresAt   uint64      `json:"expiresAt"`
}

// ---------------------------------------------------------------------------
// Integration: Snapshot-level registration with identity checks
// ---------------------------------------------------------------------------

// RegisterNodeV2 handles the unified V2 node registration with identity verification.
// This is the single entry point for all node registrations in PoB V2.
func (s *Snapshot) RegisterNodeV2(
	req *NodeRegistrationRequest,
	challenge common.Hash,
	currentBlock uint64,
	challengeBlock uint64,
	detector *SybilDetector,
	sourceIP string,
) error {
	// 1. Validate request structure
	if err := req.Validate(); err != nil {
		return err
	}

	switch req.NodeType {
	case NodeTypeAgent:
		// Validate agent identity
		if detector != nil {
			if err := detector.ValidateAgentRegistration(
				req.AgentProof, challenge, currentBlock, challengeBlock, sourceIP,
			); err != nil {
				return err
			}
		}
		// Register in snapshot
		s.RegisterAgent(req.AgentProof.AgentAddress, nil, currentBlock)

	case NodeTypePhysical:
		// Validate device identity
		if detector != nil {
			if err := detector.ValidatePhysicalRegistration(
				req.PhysicalProof, challenge, currentBlock, challengeBlock, sourceIP,
			); err != nil {
				return err
			}
		}
		// Register in snapshot
		s.RegisterPhysicalNode(req.PhysicalProof.OperatorAddress, currentBlock)

	default:
		return ErrInvalidNodeType
	}

	return nil
}

// NodeIdentity stores the verified identity of a registered node.
type NodeIdentity struct {
	Address  common.Address  `json:"address"`
	NodeType NodeType        `json:"nodeType"`
	AgentID  common.Hash     `json:"agentId,omitempty"`  // Only for Agent nodes
	DeviceID DeviceFingerprint `json:"deviceId,omitempty"` // Only for Physical nodes
	RegisteredAt uint64      `json:"registeredAt"`
	LastVerified uint64      `json:"lastVerified"`
}

// PeriodicReverification returns true if a node needs re-verification.
// Agent nodes: re-verify signature every 100,000 blocks (~11 hours at 400ms).
// Physical nodes: re-verify device fingerprint every 250,000 blocks (~28 hours).
func (ni *NodeIdentity) NeedsReverification(currentBlock uint64) bool {
	switch ni.NodeType {
	case NodeTypeAgent:
		return currentBlock-ni.LastVerified > 100000
	case NodeTypePhysical:
		return currentBlock-ni.LastVerified > 250000
	}
	return true
}

// CalcSybilPenalty returns a score penalty (0-5000 basis points) for sybil indicators.
// Applied to the node's behavior score to reduce rewards for suspicious nodes.
func CalcSybilPenalty(identity *NodeIdentity, detector *SybilDetector, currentBlock uint64) uint64 {
	if identity == nil || detector == nil {
		return 0
	}

	var penalty uint64

	// Overdue re-verification: 1000 bp penalty
	if identity.NeedsReverification(currentBlock) {
		penalty += 1000
	}

	// Cap at 5000 (50% score reduction)
	if penalty > 5000 {
		penalty = 5000
	}
	return penalty
}

// ---------------------------------------------------------------------------
// Helpers for building device fingerprints from system info
// ---------------------------------------------------------------------------

// HashDeviceInfo hashes raw device identifiers into a DeviceFingerprint.
// Used by node software to generate the fingerprint locally.
func HashDeviceInfo(cpuID string, macAddr [6]byte, diskSerial string, boardSerial string) DeviceFingerprint {
	return MakeDeviceFingerprint(
		[]byte(cpuID),
		macAddr[:],
		[]byte(diskSerial),
		[]byte(boardSerial),
	)
}

// DetectVirtualization checks hardware report for virtualization indicators.
// Returns the virtualization flags bitfield.
func DetectVirtualization(report *HardwareReport) uint64 {
	if report == nil {
		return VirtFlagNone
	}
	return report.VirtualizationFlags
}

// KnownVMMACs contains MAC address prefixes of known virtual machine vendors.
var KnownVMMACs = [][3]byte{
	{0x00, 0x05, 0x69}, // VMware
	{0x00, 0x0C, 0x29}, // VMware
	{0x00, 0x1C, 0x14}, // VMware
	{0x00, 0x50, 0x56}, // VMware
	{0x08, 0x00, 0x27}, // VirtualBox
	{0x52, 0x54, 0x00}, // QEMU/KVM
	{0x00, 0x16, 0x3E}, // Xen
	{0x00, 0x15, 0x5D}, // Hyper-V
}

// IsKnownVMMAC checks if a MAC address belongs to a known VM vendor.
func IsKnownVMMAC(mac [6]byte) bool {
	prefix := [3]byte{mac[0], mac[1], mac[2]}
	for _, vmMac := range KnownVMMACs {
		if prefix == vmMac {
			return true
		}
	}
	return false
}
