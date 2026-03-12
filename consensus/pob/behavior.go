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
	"github.com/probechain/go-probe/common"
)

const (
	// maxScore is the maximum behavior score (basis points).
	maxScore = uint64(10000)
	// defaultInitialScore is the default starting score for new validators.
	defaultInitialScore = uint64(5000)
)

// BehaviorScore holds the composite and per-dimension scores for a validator.
type BehaviorScore struct {
	Total              uint64 `json:"total"`              // Composite score (0-10000 basis points)
	Liveness           uint64 `json:"liveness"`           // Liveness dimension score
	Correctness        uint64 `json:"correctness"`        // Correctness dimension score
	Cooperation        uint64 `json:"cooperation"`        // Cooperation dimension score
	Consistency        uint64 `json:"consistency"`        // Consistency dimension score
	SignalSovereignty  uint64 `json:"signalSovereignty"`  // Signal sovereignty dimension score
	LastUpdate         uint64 `json:"lastUpdate"`         // Block number of last score update
}

// ValidatorHistory tracks the on-chain actions of a validator for scoring.
type ValidatorHistory struct {
	BlocksProposed   uint64 `json:"blocksProposed"`
	BlocksMissed     uint64 `json:"blocksMissed"`
	InvalidProposals uint64 `json:"invalidProposals"`
	AcksGiven        uint64 `json:"acksGiven"`
	AcksMissed       uint64 `json:"acksMissed"`
	SlashCount       uint64 `json:"slashCount"`
	RydbergVerified  uint64 `json:"rydbergVerified"` // Blocks with verified Rydberg time source
	RadioSyncs       uint64 `json:"radioSyncs"`      // Successful radio-based time syncs
	StellarBlocks    uint64 `json:"stellarBlocks"`    // Blocks produced with AtomicTime present
}

// BehaviorAgent is the AI scoring agent that evaluates validator behavior
// across five dimensions: liveness, correctness, cooperation, consistency,
// and signal sovereignty.
type BehaviorAgent struct {
	// Weights for each dimension: [liveness, correctness, cooperation, consistency, signalSovereignty].
	// Each is expressed as a percentage out of 100 (must sum to 100).
	weights [5]uint64
}

// NewBehaviorAgent creates a new BehaviorAgent with the default dimension weights.
func NewBehaviorAgent() *BehaviorAgent {
	return &BehaviorAgent{
		weights: [5]uint64{25, 25, 18, 17, 15}, // liveness, correctness, cooperation, consistency, signalSovereignty
	}
}

// EvaluateValidator scores a validator based on its history.
// Returns a BehaviorScore with per-dimension and total scores in basis points (0-10000).
func (ba *BehaviorAgent) EvaluateValidator(addr common.Address, history *ValidatorHistory, blockNumber uint64) *BehaviorScore {
	liveness := ba.calcLiveness(history)
	correctness := ba.calcCorrectness(history)
	cooperation := ba.calcCooperation(history)
	consistency := ba.calcConsistency(history)
	signalSovereignty := ba.calcSignalSovereignty(history)

	total := (liveness*ba.weights[0] + correctness*ba.weights[1] +
		cooperation*ba.weights[2] + consistency*ba.weights[3] +
		signalSovereignty*ba.weights[4]) / 100

	if total > maxScore {
		total = maxScore
	}

	return &BehaviorScore{
		Total:             total,
		Liveness:          liveness,
		Correctness:       correctness,
		Cooperation:       cooperation,
		Consistency:       consistency,
		SignalSovereignty: signalSovereignty,
		LastUpdate:        blockNumber,
	}
}

// calcLiveness scores based on block production rate.
// Perfect production = maxScore, deducted for misses.
func (ba *BehaviorAgent) calcLiveness(h *ValidatorHistory) uint64 {
	totalOpportunities := h.BlocksProposed + h.BlocksMissed
	if totalOpportunities == 0 {
		return maxScore // No opportunities yet, assume perfect
	}
	return (h.BlocksProposed * maxScore) / totalOpportunities
}

// calcCorrectness scores based on valid vs invalid proposals.
func (ba *BehaviorAgent) calcCorrectness(h *ValidatorHistory) uint64 {
	totalProposals := h.BlocksProposed + h.InvalidProposals
	if totalProposals == 0 {
		return maxScore
	}
	return (h.BlocksProposed * maxScore) / totalProposals
}

// calcCooperation scores based on acknowledgment participation.
func (ba *BehaviorAgent) calcCooperation(h *ValidatorHistory) uint64 {
	totalAcks := h.AcksGiven + h.AcksMissed
	if totalAcks == 0 {
		return maxScore
	}
	return (h.AcksGiven * maxScore) / totalAcks
}

// calcConsistency scores inversely proportional to slash count.
// No slashes = maxScore. Each slash reduces by 1000 bp.
func (ba *BehaviorAgent) calcConsistency(h *ValidatorHistory) uint64 {
	penalty := h.SlashCount * 1000
	if penalty >= maxScore {
		return 0
	}
	return maxScore - penalty
}

// calcSignalSovereignty scores based on a validator's Stellar-Class capabilities:
// Rydberg-verified blocks, radio-based time syncs, and AtomicTime block production.
// Validators without stellar capabilities receive a neutral baseline score (5000)
// so they are not penalized below the default starting point.
func (ba *BehaviorAgent) calcSignalSovereignty(h *ValidatorHistory) uint64 {
	totalStellarOps := h.RydbergVerified + h.RadioSyncs + h.StellarBlocks
	if totalStellarOps == 0 {
		return defaultInitialScore // Neutral baseline — no penalty for non-stellar nodes
	}

	// Score components:
	// - Rydberg verification: 40% weight (demonstrates atomic receiver)
	// - Radio syncs: 30% weight (demonstrates RF time sync capability)
	// - Stellar blocks: 30% weight (demonstrates AtomicTime block production)
	var score uint64

	// Rydberg component: proportion of stellar blocks that have Rydberg verification
	if h.StellarBlocks > 0 {
		rydbergRatio := (h.RydbergVerified * maxScore) / h.StellarBlocks
		if rydbergRatio > maxScore {
			rydbergRatio = maxScore
		}
		score += rydbergRatio * 40 / 100
	}

	// Radio sync component: capped at maxScore for 100+ syncs
	radioScore := h.RadioSyncs * 100 // Each sync adds 100 bp
	if radioScore > maxScore {
		radioScore = maxScore
	}
	score += radioScore * 30 / 100

	// Stellar block production component
	stellarScore := h.StellarBlocks * 50 // Each stellar block adds 50 bp
	if stellarScore > maxScore {
		stellarScore = maxScore
	}
	score += stellarScore * 30 / 100

	if score > maxScore {
		score = maxScore
	}
	return score
}

// EvaluateValidatorFast returns a cached score during StellarSpeed ticks.
// Full evaluation only happens at epoch boundaries; between ticks, the previous
// score is returned with only the StellarBlocks count updated.
func (ba *BehaviorAgent) EvaluateValidatorFast(addr common.Address, history *ValidatorHistory,
	blockNumber uint64, epochLength uint64, cachedScore *BehaviorScore) *BehaviorScore {

	// At epoch boundaries, do a full evaluation
	if epochLength == 0 || blockNumber%epochLength == 0 || cachedScore == nil {
		return ba.EvaluateValidator(addr, history, blockNumber)
	}

	// Between epochs: return the cached score with an updated timestamp
	return &BehaviorScore{
		Total:             cachedScore.Total,
		Liveness:          cachedScore.Liveness,
		Correctness:       cachedScore.Correctness,
		Cooperation:       cachedScore.Cooperation,
		Consistency:       cachedScore.Consistency,
		SignalSovereignty: cachedScore.SignalSovereignty,
		LastUpdate:        blockNumber,
	}
}

// UpdateScores re-evaluates all validators in the snapshot and returns updated scores.
func (ba *BehaviorAgent) UpdateScores(validators map[common.Address]*BehaviorScore,
	histories map[common.Address]*ValidatorHistory, blockNumber uint64) map[common.Address]*BehaviorScore {

	updated := make(map[common.Address]*BehaviorScore, len(validators))
	for addr := range validators {
		history, ok := histories[addr]
		if !ok {
			history = &ValidatorHistory{}
		}
		updated[addr] = ba.EvaluateValidator(addr, history, blockNumber)
	}
	return updated
}

// ProportionalSlash reduces a validator's score proportionally to the severity.
// severity is in basis points (0-10000); the actual deduction is:
//
//	deduction = currentScore * severity * slashFraction / (10000 * 10000)
//
// Returns the new total score.
func (ba *BehaviorAgent) ProportionalSlash(score *BehaviorScore, severity uint64, slashFraction uint64) uint64 {
	if severity > maxScore {
		severity = maxScore
	}
	deduction := (score.Total * severity * slashFraction) / (maxScore * maxScore)
	if deduction >= score.Total {
		score.Total = 0
	} else {
		score.Total -= deduction
	}
	return score.Total
}

// DefaultBehaviorScore returns a behavior score initialized to the given initial score.
func DefaultBehaviorScore(initialScore uint64, blockNumber uint64) *BehaviorScore {
	return &BehaviorScore{
		Total:             initialScore,
		Liveness:          maxScore,
		Correctness:       maxScore,
		Cooperation:       maxScore,
		Consistency:       maxScore,
		SignalSovereignty: defaultInitialScore, // Neutral baseline for non-stellar nodes
		LastUpdate:        blockNumber,
	}
}

// ---------------------------------------------------------------------------
// SmartLight scoring
// ---------------------------------------------------------------------------

// SmartLightHistory tracks on-chain actions of a SmartLight node for scoring.
type SmartLightHistory struct {
	HeartbeatsSent    uint64 `json:"heartbeatsSent"`    // Total heartbeat proofs sent
	HeartbeatsMissed  uint64 `json:"heartbeatsMissed"`  // Missed heartbeat windows
	AcksGiven         uint64 `json:"acksGiven"`         // ACK attestations sent
	AcksMissed        uint64 `json:"acksMissed"`        // Missed ACK opportunities
	ValidAttestations uint64 `json:"validAttestations"` // Correct header attestations
	InvalidAttests    uint64 `json:"invalidAttests"`    // Incorrect header attestations
	SlashCount        uint64 `json:"slashCount"`        // Slashing events
	GNSSSamples       uint64 `json:"gnssSamples"`       // GNSS time samples submitted
	TasksCompleted    uint64 `json:"tasksCompleted"`    // Agent tasks completed
	TasksFailed       uint64 `json:"tasksFailed"`       // Agent tasks failed
}

// SmartLightScore holds the composite and per-dimension scores for a SmartLight node.
// SmartLight uses different weights than full validators:
// liveness 30%, correctness 20%, cooperation 25%, consistency 10%, signalSovereignty 15%.
type SmartLightScore struct {
	Total             uint64 `json:"total"`             // Composite score (0-10000 basis points)
	Liveness          uint64 `json:"liveness"`          // Heartbeat participation
	Correctness       uint64 `json:"correctness"`       // Header attestation accuracy
	Cooperation       uint64 `json:"cooperation"`       // ACK participation rate
	Consistency       uint64 `json:"consistency"`       // No invalid attestations
	SignalSovereignty uint64 `json:"signalSovereignty"` // GNSS time contribution
	LastUpdate        uint64 `json:"lastUpdate"`        // Block number of last update
}

// SmartLightAgent is the scoring agent for SmartLight nodes.
type SmartLightAgent struct {
	// Weights: [liveness, correctness, cooperation, consistency, signalSovereignty]
	weights [5]uint64
}

// NewSmartLightAgent creates a SmartLight scoring agent with SmartLight-specific weights.
func NewSmartLightAgent() *SmartLightAgent {
	return &SmartLightAgent{
		weights: [5]uint64{30, 20, 25, 10, 15},
	}
}

// EvaluateSmartLight scores a SmartLight node based on its history.
func (sla *SmartLightAgent) EvaluateSmartLight(addr common.Address, history *SmartLightHistory, blockNumber uint64) *SmartLightScore {
	liveness := sla.calcSLLiveness(history)
	correctness := sla.calcSLCorrectness(history)
	cooperation := sla.calcSLCooperation(history)
	consistency := sla.calcSLConsistency(history)
	signalSov := sla.calcSLSignalSovereignty(history)

	total := (liveness*sla.weights[0] + correctness*sla.weights[1] +
		cooperation*sla.weights[2] + consistency*sla.weights[3] +
		signalSov*sla.weights[4]) / 100

	if total > maxScore {
		total = maxScore
	}

	return &SmartLightScore{
		Total:             total,
		Liveness:          liveness,
		Correctness:       correctness,
		Cooperation:       cooperation,
		Consistency:       consistency,
		SignalSovereignty: signalSov,
		LastUpdate:        blockNumber,
	}
}

// calcSLLiveness scores heartbeat participation.
func (sla *SmartLightAgent) calcSLLiveness(h *SmartLightHistory) uint64 {
	total := h.HeartbeatsSent + h.HeartbeatsMissed
	if total == 0 {
		return maxScore
	}
	return (h.HeartbeatsSent * maxScore) / total
}

// calcSLCorrectness scores header attestation accuracy.
func (sla *SmartLightAgent) calcSLCorrectness(h *SmartLightHistory) uint64 {
	total := h.ValidAttestations + h.InvalidAttests
	if total == 0 {
		return maxScore
	}
	return (h.ValidAttestations * maxScore) / total
}

// calcSLCooperation scores ACK participation.
func (sla *SmartLightAgent) calcSLCooperation(h *SmartLightHistory) uint64 {
	total := h.AcksGiven + h.AcksMissed
	if total == 0 {
		return maxScore
	}
	return (h.AcksGiven * maxScore) / total
}

// calcSLConsistency scores inversely proportional to slash count.
func (sla *SmartLightAgent) calcSLConsistency(h *SmartLightHistory) uint64 {
	penalty := h.SlashCount * 1000
	if penalty >= maxScore {
		return 0
	}
	return maxScore - penalty
}

// calcSLSignalSovereignty scores GNSS time sample contribution.
func (sla *SmartLightAgent) calcSLSignalSovereignty(h *SmartLightHistory) uint64 {
	if h.GNSSSamples == 0 {
		return defaultInitialScore // Neutral baseline
	}
	score := h.GNSSSamples * 100 // Each sample adds 100 bp
	if score > maxScore {
		score = maxScore
	}
	return score
}

// UpdateSmartLightScores re-evaluates all SmartLight nodes in a snapshot.
func (sla *SmartLightAgent) UpdateSmartLightScores(nodes map[common.Address]*SmartLightScore,
	histories map[common.Address]*SmartLightHistory, blockNumber uint64) map[common.Address]*SmartLightScore {

	updated := make(map[common.Address]*SmartLightScore, len(nodes))
	for addr := range nodes {
		history, ok := histories[addr]
		if !ok {
			history = &SmartLightHistory{}
		}
		updated[addr] = sla.EvaluateSmartLight(addr, history, blockNumber)
	}
	return updated
}

// DefaultSmartLightScore returns a SmartLight score initialized to defaults.
func DefaultSmartLightScore(blockNumber uint64) *SmartLightScore {
	return &SmartLightScore{
		Total:             defaultInitialScore,
		Liveness:          maxScore,
		Correctness:       maxScore,
		Cooperation:       maxScore,
		Consistency:       maxScore,
		SignalSovereignty: defaultInitialScore,
		LastUpdate:        blockNumber,
	}
}

// ---------------------------------------------------------------------------
// Agent (PoB) scoring
// ---------------------------------------------------------------------------

// AgentHistory tracks on-chain actions of an Agent node for scoring.
type AgentHistory struct {
	TasksCompleted  uint64 `json:"tasksCompleted"`  // Agent tasks successfully completed
	TasksFailed     uint64 `json:"tasksFailed"`     // Agent tasks failed
	AttestationsOK  uint64 `json:"attestationsOK"`  // Correct cross-agent attestations
	AttestationsBad uint64 `json:"attestationsBad"` // Incorrect attestations
	HeartbeatsSent  uint64 `json:"heartbeatsSent"`  // Heartbeats sent
	HeartbeatsMissed uint64 `json:"heartbeatsMissed"` // Missed heartbeat windows
	Uptime          uint64 `json:"uptime"`          // Blocks since registration
	LastActive      uint64 `json:"lastActive"`      // Block number of last activity
	StakeAmount     uint64 `json:"stakeAmount"`     // Staked amount in wei (stored as uint64 for scoring)
	SlashCount      uint64 `json:"slashCount"`      // Slashing events
}

// AgentScore holds the composite and per-dimension scores for an Agent node.
// Agent scoring uses 6 dimensions:
// Responsiveness 20%, Accuracy 25%, Reliability 15%, Cooperation 15%, Economy 15%, Sovereignty 10%.
type AgentScore struct {
	Total          uint64 `json:"total"`          // Composite score (0-10000 basis points)
	Responsiveness uint64 `json:"responsiveness"` // Heartbeat and task response speed
	Accuracy       uint64 `json:"accuracy"`       // Task and attestation correctness
	Reliability    uint64 `json:"reliability"`     // Uptime and consistency
	Cooperation    uint64 `json:"cooperation"`     // Cross-agent attestation participation
	Economy        uint64 `json:"economy"`         // Stake-weighted contribution
	Sovereignty    uint64 `json:"sovereignty"`     // Independent behavior (anti-Sybil)
	LastUpdate     uint64 `json:"lastUpdate"`      // Block number of last update
}

// AgentScoringAgent is the scoring agent for PoB Agent nodes.
type AgentScoringAgent struct {
	// Weights: [responsiveness, accuracy, reliability, cooperation, economy, sovereignty]
	weights [6]uint64
}

// NewAgentScoringAgent creates an Agent scoring agent with PoB-specific weights.
func NewAgentScoringAgent() *AgentScoringAgent {
	return &AgentScoringAgent{
		weights: [6]uint64{20, 25, 15, 15, 15, 10},
	}
}

// EvaluateAgent scores an Agent node based on its history.
func (asa *AgentScoringAgent) EvaluateAgent(addr common.Address, history *AgentHistory, blockNumber uint64) *AgentScore {
	responsiveness := asa.calcAgentResponsiveness(history)
	accuracy := asa.calcAgentAccuracy(history)
	reliability := asa.calcAgentReliability(history, blockNumber)
	cooperation := asa.calcAgentCooperation(history)
	economy := asa.calcAgentEconomy(history)
	sovereignty := asa.calcAgentSovereignty(history)

	total := (responsiveness*asa.weights[0] + accuracy*asa.weights[1] +
		reliability*asa.weights[2] + cooperation*asa.weights[3] +
		economy*asa.weights[4] + sovereignty*asa.weights[5]) / 100

	if total > maxScore {
		total = maxScore
	}

	return &AgentScore{
		Total:          total,
		Responsiveness: responsiveness,
		Accuracy:       accuracy,
		Reliability:    reliability,
		Cooperation:    cooperation,
		Economy:        economy,
		Sovereignty:    sovereignty,
		LastUpdate:     blockNumber,
	}
}

// calcAgentResponsiveness scores heartbeat participation.
func (asa *AgentScoringAgent) calcAgentResponsiveness(h *AgentHistory) uint64 {
	total := h.HeartbeatsSent + h.HeartbeatsMissed
	if total == 0 {
		return maxScore
	}
	return (h.HeartbeatsSent * maxScore) / total
}

// calcAgentAccuracy scores task and attestation correctness.
func (asa *AgentScoringAgent) calcAgentAccuracy(h *AgentHistory) uint64 {
	totalTasks := h.TasksCompleted + h.TasksFailed
	totalAttests := h.AttestationsOK + h.AttestationsBad

	if totalTasks == 0 && totalAttests == 0 {
		return maxScore
	}

	var taskScore, attestScore uint64
	if totalTasks > 0 {
		taskScore = (h.TasksCompleted * maxScore) / totalTasks
	} else {
		taskScore = maxScore
	}
	if totalAttests > 0 {
		attestScore = (h.AttestationsOK * maxScore) / totalAttests
	} else {
		attestScore = maxScore
	}

	// 60% task accuracy, 40% attestation accuracy
	return (taskScore*60 + attestScore*40) / 100
}

// calcAgentReliability scores uptime and slash-free consistency.
func (asa *AgentScoringAgent) calcAgentReliability(h *AgentHistory, blockNumber uint64) uint64 {
	penalty := h.SlashCount * 1000
	if penalty >= maxScore {
		return 0
	}
	base := maxScore - penalty

	// Bonus for long uptime: min(uptime/10000 * 1000, 1000)
	uptimeBonus := h.Uptime / 10
	if uptimeBonus > 1000 {
		uptimeBonus = 1000
	}

	score := base + uptimeBonus
	if score > maxScore {
		score = maxScore
	}
	return score
}

// calcAgentCooperation scores cross-agent attestation participation.
func (asa *AgentScoringAgent) calcAgentCooperation(h *AgentHistory) uint64 {
	total := h.AttestationsOK + h.AttestationsBad
	if total == 0 {
		return defaultInitialScore // Neutral baseline
	}
	return (h.AttestationsOK * maxScore) / total
}

// calcAgentEconomy scores stake-weighted contribution.
func (asa *AgentScoringAgent) calcAgentEconomy(h *AgentHistory) uint64 {
	// Minimum stake = 0.1 PROBE = 1e17 wei; reward agents with higher stake
	// but cap at 10 PROBE equivalent for max score
	if h.StakeAmount == 0 {
		return 0
	}
	// Score: min(stakeAmount / 1e17 * 1000, maxScore)
	score := (h.StakeAmount / 1e14) // Normalize to basis points range
	if score > maxScore {
		score = maxScore
	}
	return score
}

// calcAgentSovereignty scores independent behavior (anti-Sybil).
// Agents without significant task history receive neutral baseline.
func (asa *AgentScoringAgent) calcAgentSovereignty(h *AgentHistory) uint64 {
	totalOps := h.TasksCompleted + h.AttestationsOK
	if totalOps == 0 {
		return defaultInitialScore
	}
	// Higher task completion diversity indicates more sovereignty
	score := totalOps * 50
	if score > maxScore {
		score = maxScore
	}
	return score
}

// UpdateAgentScores re-evaluates all Agent nodes in a snapshot.
func (asa *AgentScoringAgent) UpdateAgentScores(nodes map[common.Address]*AgentScore,
	histories map[common.Address]*AgentHistory, blockNumber uint64) map[common.Address]*AgentScore {

	updated := make(map[common.Address]*AgentScore, len(nodes))
	for addr := range nodes {
		history, ok := histories[addr]
		if !ok {
			history = &AgentHistory{}
		}
		updated[addr] = asa.EvaluateAgent(addr, history, blockNumber)
	}
	return updated
}

// DefaultAgentScore returns an Agent score initialized to defaults.
func DefaultAgentScore(blockNumber uint64) *AgentScore {
	return &AgentScore{
		Total:          defaultInitialScore,
		Responsiveness: maxScore,
		Accuracy:       maxScore,
		Reliability:    maxScore,
		Cooperation:    defaultInitialScore,
		Economy:        defaultInitialScore,
		Sovereignty:    defaultInitialScore,
		LastUpdate:     blockNumber,
	}
}

// ---------------------------------------------------------------------------
// Physical Node (PoB-P) scoring
// ---------------------------------------------------------------------------
// Physical nodes are any device with persistent power and connectivity:
// data centers, cars, fridges, phones. Rewarded for storage contribution.

// PhysicalNodeHistory tracks a physical node's contributions.
type PhysicalNodeHistory struct {
	StorageBytesProvided uint64 `json:"storageBytesProvided"` // Total storage contributed (bytes)
	StorageBytesVerified uint64 `json:"storageBytesVerified"` // Storage verified by challenges
	UptimeBlocks         uint64 `json:"uptimeBlocks"`         // Blocks online
	DowntimeBlocks       uint64 `json:"downtimeBlocks"`       // Blocks offline
	DataServiced         uint64 `json:"dataServiced"`          // Data retrieval requests served
	InvalidProofs        uint64 `json:"invalidProofs"`         // Failed storage proofs
	LastActive           uint64 `json:"lastActive"`            // Last active block
	Registered           uint64 `json:"registered"`            // Registration block
}

// PhysicalNodeScore holds behavior scores for a physical node.
// Dimensions: StorageContribution 40%, Uptime 25%, DataService 20%, Integrity 15%.
type PhysicalNodeScore struct {
	Total               uint64 `json:"total"`
	StorageContribution uint64 `json:"storageContribution"` // How much storage provided
	Uptime              uint64 `json:"uptime"`              // Online reliability
	DataService         uint64 `json:"dataService"`          // Data serving quality
	Integrity           uint64 `json:"integrity"`            // No invalid proofs
	LastUpdate          uint64 `json:"lastUpdate"`
}

// PhysicalNodeScoringAgent evaluates physical node behavior.
type PhysicalNodeScoringAgent struct {
	// Weights: [storageContribution, uptime, dataService, integrity]
	weights [4]uint64
}

// NewPhysicalNodeScoringAgent creates a physical node scoring agent.
func NewPhysicalNodeScoringAgent() *PhysicalNodeScoringAgent {
	return &PhysicalNodeScoringAgent{
		weights: [4]uint64{40, 25, 20, 15},
	}
}

// EvaluatePhysicalNode scores a physical node based on its history.
func (psa *PhysicalNodeScoringAgent) EvaluatePhysicalNode(history *PhysicalNodeHistory, blockNumber uint64) *PhysicalNodeScore {
	if history == nil {
		history = &PhysicalNodeHistory{}
	}
	storage := psa.calcStorage(history)
	uptime := psa.calcUptime(history)
	dataService := psa.calcDataService(history)
	integrity := psa.calcIntegrity(history)

	total := (storage*psa.weights[0] + uptime*psa.weights[1] +
		dataService*psa.weights[2] + integrity*psa.weights[3]) / 100

	if total > maxScore {
		total = maxScore
	}
	return &PhysicalNodeScore{
		Total:               total,
		StorageContribution: storage,
		Uptime:              uptime,
		DataService:         dataService,
		Integrity:           integrity,
		LastUpdate:          blockNumber,
	}
}

func (psa *PhysicalNodeScoringAgent) calcStorage(h *PhysicalNodeHistory) uint64 {
	if h.StorageBytesProvided == 0 {
		return defaultInitialScore
	}
	if h.StorageBytesVerified >= h.StorageBytesProvided {
		return maxScore
	}
	return h.StorageBytesVerified * maxScore / h.StorageBytesProvided
}

func (psa *PhysicalNodeScoringAgent) calcUptime(h *PhysicalNodeHistory) uint64 {
	total := h.UptimeBlocks + h.DowntimeBlocks
	if total == 0 {
		return defaultInitialScore
	}
	return h.UptimeBlocks * maxScore / total
}

func (psa *PhysicalNodeScoringAgent) calcDataService(h *PhysicalNodeHistory) uint64 {
	if h.DataServiced == 0 {
		return defaultInitialScore
	}
	// More data served = higher score, capped at maxScore
	// 1000 requests = full score
	score := h.DataServiced * maxScore / 1000
	if score > maxScore {
		score = maxScore
	}
	return score
}

func (psa *PhysicalNodeScoringAgent) calcIntegrity(h *PhysicalNodeHistory) uint64 {
	totalProofs := h.StorageBytesVerified + h.InvalidProofs
	if totalProofs == 0 {
		return maxScore
	}
	if h.InvalidProofs == 0 {
		return maxScore
	}
	errorRate := h.InvalidProofs * maxScore / totalProofs
	if errorRate >= maxScore {
		return 0
	}
	return maxScore - errorRate
}

// DefaultPhysicalNodeScore returns a physical node score initialized to defaults.
func DefaultPhysicalNodeScore(blockNumber uint64) *PhysicalNodeScore {
	return &PhysicalNodeScore{
		Total:               defaultInitialScore,
		StorageContribution: defaultInitialScore,
		Uptime:              maxScore,
		DataService:         defaultInitialScore,
		Integrity:           maxScore,
		LastUpdate:          blockNumber,
	}
}
