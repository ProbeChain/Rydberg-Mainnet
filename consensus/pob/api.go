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
	"fmt"

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/consensus"
	"github.com/probechain/go-probe/core/types"
	"github.com/probechain/go-probe/rpc"
)

// API is a user facing RPC API to allow controlling the validator and voting
// mechanisms of the proof-of-behavior scheme.
type API struct {
	chain consensus.ChainHeaderReader
	pob   *ProofOfBehavior
}

// GetSnapshot retrieves the state snapshot at a given block.
func (api *API) GetSnapshot(number *rpc.BlockNumber) (*Snapshot, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return nil, errUnknownBlock
	}
	return api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
}

// GetSnapshotAtHash retrieves the state snapshot at a given block.
func (api *API) GetSnapshotAtHash(hash common.Hash) (*Snapshot, error) {
	header := api.chain.GetHeaderByHash(hash)
	if header == nil {
		return nil, errUnknownBlock
	}
	return api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
}

// GetBehaviorScores retrieves the behavior scores for all validators at the specified block.
func (api *API) GetBehaviorScores(number *rpc.BlockNumber) (map[common.Address]*BehaviorScore, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return nil, errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	return snap.Validators, nil
}

// GetValidators retrieves the list of authorized validators at the specified block.
func (api *API) GetValidators(number *rpc.BlockNumber) ([]common.Address, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return nil, errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	return snap.validators(), nil
}

// GetValidatorsAtHash retrieves the list of authorized validators at the specified block.
func (api *API) GetValidatorsAtHash(hash common.Hash) ([]common.Address, error) {
	header := api.chain.GetHeaderByHash(hash)
	if header == nil {
		return nil, errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	return snap.validators(), nil
}

// Proposals returns the current proposals the node tries to uphold and vote on.
func (api *API) Proposals() map[common.Address]bool {
	api.pob.lock.RLock()
	defer api.pob.lock.RUnlock()

	proposals := make(map[common.Address]bool)
	for address, auth := range api.pob.proposals {
		proposals[address] = auth
	}
	return proposals
}

// Propose injects a new authorization proposal that the validator will attempt to push through.
func (api *API) Propose(address common.Address, auth bool) {
	api.pob.lock.Lock()
	defer api.pob.lock.Unlock()

	api.pob.proposals[address] = auth
}

// Discard drops a currently running proposal.
func (api *API) Discard(address common.Address) {
	api.pob.lock.Lock()
	defer api.pob.lock.Unlock()

	delete(api.pob.proposals, address)
}

// GetSmartLightScores retrieves the behavior scores for all SmartLight nodes at the specified block.
func (api *API) GetSmartLightScores(number *rpc.BlockNumber) (map[common.Address]*SmartLightScore, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return nil, errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	return snap.SmartLights, nil
}

// GetSmartLightCount returns the number of registered SmartLight nodes.
func (api *API) GetSmartLightCount(number *rpc.BlockNumber) (int, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return 0, errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return 0, err
	}
	return snap.SmartLightCount(), nil
}

// GetAgentScores retrieves the behavior scores for all Agent nodes at the specified block.
func (api *API) GetAgentScores(number *rpc.BlockNumber) (map[common.Address]*AgentScore, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return nil, errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	return snap.Agents, nil
}

// GetAgentCount returns the number of registered Agent nodes.
func (api *API) GetAgentCount(number *rpc.BlockNumber) (int, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return 0, errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return 0, err
	}
	return snap.AgentCount(), nil
}

// GetAgentInfo returns detailed info about a specific agent node.
func (api *API) GetAgentInfo(addr common.Address, number *rpc.BlockNumber) (map[string]interface{}, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return nil, errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	score, ok := snap.Agents[addr]
	if !ok {
		return nil, fmt.Errorf("agent not found: %s", addr.Hex())
	}
	history := snap.AgentHistories[addr]
	if history == nil {
		history = &AgentHistory{}
	}

	return map[string]interface{}{
		"address":        addr,
		"score":          score,
		"history":        history,
		"registered":     true,
		"hasPubKey":      len(snap.AgentPubKeys[addr]) > 0,
	}, nil
}

// GetRelayInfo returns relay management info for a SmartLight node.
func (api *API) GetRelayInfo(addr common.Address, number *rpc.BlockNumber) (map[string]interface{}, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return nil, errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	if !snap.IsSmartLight(addr) {
		return nil, fmt.Errorf("not a SmartLight node: %s", addr.Hex())
	}
	relayInfo := snap.RelayInfos[addr]
	if relayInfo == nil {
		relayInfo = NewRelayInfo()
	}
	mgmtScore := CalcAgentManagementScore(relayInfo)

	return map[string]interface{}{
		"address":            addr,
		"managedAgentCount":  relayInfo.AgentCount(),
		"heartbeatsRelayed":  relayInfo.AgentHeartbeatsRelayed,
		"attestationsRelayed": relayInfo.AgentAttestationsRelayed,
		"tasksAssigned":      relayInfo.AgentTasksAssigned,
		"tasksCompleted":     relayInfo.AgentTasksCompleted,
		"invalidAggregations": relayInfo.InvalidAggregations,
		"agentDrops":         relayInfo.AgentDrops,
		"managementScore":    mgmtScore,
	}, nil
}

// GetRelayCount returns the number of active relay nodes.
func (api *API) GetRelayCount(number *rpc.BlockNumber) (int, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return 0, errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return 0, err
	}
	return snap.RelayCount(), nil
}

// RegisterNode queues a node registration (gas-free, encoded in block header).
// nodeType: 1=Agent, 2=Physical
func (api *API) RegisterNode(addr common.Address, nodeType uint64) (string, error) {
	if nodeType != uint64(NodeTypeAgent) && nodeType != uint64(NodeTypePhysical) {
		return "", fmt.Errorf("invalid nodeType %d: must be 1 (Agent) or 2 (Physical)", nodeType)
	}

	// Check if already registered
	header := api.chain.CurrentHeader()
	if header == nil {
		return "", errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return "", err
	}
	if nodeType == uint64(NodeTypeAgent) && snap.IsAgent(addr) {
		return "already registered as Agent", nil
	}
	if nodeType == uint64(NodeTypePhysical) && snap.IsPhysicalNode(addr) {
		return "already registered as Physical", nil
	}

	api.pob.AddPendingRegistration(addr, uint8(nodeType))
	typeName := NodeType(nodeType).String()
	return fmt.Sprintf("queued %s registration for %s (will be included in next block)", typeName, addr.Hex()), nil
}

// GetNodeRegistrationStatus returns the registration status of a node address.
func (api *API) GetNodeRegistrationStatus(addr common.Address) (map[string]interface{}, error) {
	header := api.chain.CurrentHeader()
	if header == nil {
		return nil, errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"address":    addr,
		"isAgent":    snap.IsAgent(addr),
		"isPhysical": snap.IsPhysicalNode(addr),
	}
	if score, ok := snap.Agents[addr]; ok {
		result["agentScore"] = score.Total
	}
	if score, ok := snap.PhysicalNodes[addr]; ok {
		result["physicalScore"] = score.Total
	}
	return result, nil
}

// GetPhysicalNodeScores retrieves scores for all Physical nodes at the specified block.
func (api *API) GetPhysicalNodeScores(number *rpc.BlockNumber) (map[common.Address]*PhysicalNodeScore, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return nil, errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	return snap.PhysicalNodes, nil
}

// GetPhysicalNodeCount returns the number of registered Physical nodes.
func (api *API) GetPhysicalNodeCount(number *rpc.BlockNumber) (int, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return 0, errUnknownBlock
	}
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return 0, err
	}
	return snap.PhysicalNodeCount(), nil
}

type status struct {
	InturnPercent float64                `json:"inturnPercent"`
	SigningStatus map[common.Address]int `json:"sealerActivity"`
	NumBlocks     uint64                 `json:"numBlocks"`
}

// Status returns the status of the last N blocks.
func (api *API) Status() (*status, error) {
	var (
		numBlocks = uint64(64)
		header    = api.chain.CurrentHeader()
		diff      = uint64(0)
		optimals  = 0
	)
	snap, err := api.pob.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	var (
		validators = snap.validators()
		end        = header.Number.Uint64()
		start      = end - numBlocks
	)
	if numBlocks > end {
		start = 1
		numBlocks = end - start
	}
	signStatus := make(map[common.Address]int)
	for _, v := range validators {
		signStatus[v] = 0
	}
	for n := start; n < end; n++ {
		h := api.chain.GetHeaderByNumber(n)
		if h == nil {
			return nil, fmt.Errorf("missing block %d", n)
		}
		if h.Difficulty.Cmp(diffInTurn) == 0 {
			optimals++
		}
		diff += h.Difficulty.Uint64()
		sealer, err := api.pob.Author(h)
		if err != nil {
			return nil, err
		}
		signStatus[sealer]++
	}
	return &status{
		InturnPercent: float64(100*optimals) / float64(numBlocks),
		SigningStatus: signStatus,
		NumBlocks:     numBlocks,
	}, nil
}
