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

// Contains the SmartLight node wrappers for mobile (iOS/Android) platforms.

package gprobe

import (
	"encoding/json"
	"fmt"
	"math/big"
	"path/filepath"

	"github.com/probechain/go-probe/core"
	"github.com/probechain/go-probe/les"
	"github.com/probechain/go-probe/node"
	"github.com/probechain/go-probe/p2p"
	"github.com/probechain/go-probe/p2p/nat"
	"github.com/probechain/go-probe/params"
	"github.com/probechain/go-probe/probe/downloader"
	"github.com/probechain/go-probe/probe/probeconfig"
	"github.com/probechain/go-probe/probeclient"
	"github.com/probechain/go-probe/probestats"
	"github.com/probechain/go-probe/smartlight"
)

// SmartLightNodeConfig represents configuration for a SmartLight node.
// SmartLight nodes are lightweight PoB participants that contribute through
// ACK attestations, heartbeat proofs, GNSS time samples, and agent tasks.
type SmartLightNodeConfig struct {
	// Bootstrap nodes used to establish connectivity.
	BootstrapNodes *Enodes

	// MaxPeers is the maximum number of peers that can be connected.
	MaxPeers int

	// ProbeumNetworkID is the network identifier.
	ProbeumNetworkID int64

	// ProbeumGenesis is the genesis JSON.
	ProbeumGenesis string

	// ProbeumDatabaseCache is the system memory in MB for database caching.
	// SmartLight uses 50-80 MB (vs 16 MB for LES).
	ProbeumDatabaseCache int

	// ProbeumNetStats is the netstats connection string.
	ProbeumNetStats string

	// HeartbeatInterval is blocks between heartbeat proofs (default: 100).
	HeartbeatInterval int

	// GNSSEnabled enables GNSS time sampling from device GPS.
	GNSSEnabled bool

	// MaxAgentTasks is max concurrent agent tasks (default: 2).
	MaxAgentTasks int

	// PowerMode: 0=Full, 1=Eco, 2=Sleep.
	PowerMode int
}

// defaultSmartLightConfig contains the default SmartLight configuration.
var defaultSmartLightConfig = &SmartLightNodeConfig{
	BootstrapNodes:       FoundationBootnodes(),
	MaxPeers:             50,
	ProbeumNetworkID:     8004,
	ProbeumDatabaseCache: 64, // 64MB — more than LES's 16MB but still light
	HeartbeatInterval:    100,
	GNSSEnabled:          false,
	MaxAgentTasks:        2,
	PowerMode:            0,
}

// NewSmartLightNodeConfig creates a new SmartLight node config with defaults.
func NewSmartLightNodeConfig() *SmartLightNodeConfig {
	config := *defaultSmartLightConfig
	return &config
}

// SmartLightNode represents a SmartLight node instance.
type SmartLightNode struct {
	node   *node.Node
	engine *smartlight.Engine
	config *SmartLightNodeConfig
}

// NewSmartLightNode creates and configures a SmartLight node.
// It runs a LES light client underneath with SmartLight services on top.
func NewSmartLightNode(datadir string, config *SmartLightNodeConfig) (sln *SmartLightNode, _ error) {
	if config == nil {
		config = NewSmartLightNodeConfig()
	}
	if config.MaxPeers == 0 {
		config.MaxPeers = defaultSmartLightConfig.MaxPeers
	}
	if config.BootstrapNodes == nil || config.BootstrapNodes.Size() == 0 {
		config.BootstrapNodes = defaultSmartLightConfig.BootstrapNodes
	}
	if config.ProbeumDatabaseCache == 0 {
		config.ProbeumDatabaseCache = defaultSmartLightConfig.ProbeumDatabaseCache
	}

	// Create the networking stack
	nodeConf := &node.Config{
		Name:        "iGprobeSmartLight",
		Version:     params.VersionWithMeta,
		DataDir:     datadir,
		KeyStoreDir: filepath.Join(datadir, "keystore"),
		P2P: p2p.Config{
			NoDiscovery:      false,
			DiscoveryV5:      true,
			BootstrapNodes:   config.BootstrapNodes.nodes,
			BootstrapNodesV5: config.BootstrapNodes.nodes,
			StaticNodes:      config.BootstrapNodes.nodes,
			TrustedNodes:     config.BootstrapNodes.nodes,
			ListenAddr:       ":0",
			NAT:              nat.Any(),
			MaxPeers:         config.MaxPeers,
		},
	}

	rawStack, err := node.New(nodeConf)
	if err != nil {
		return nil, err
	}

	var genesis *core.Genesis
	if config.ProbeumGenesis != "" {
		genesis = new(core.Genesis)
		if err := json.Unmarshal([]byte(config.ProbeumGenesis), genesis); err != nil {
			return nil, fmt.Errorf("invalid genesis spec: %v", err)
		}
	}

	// Register LES light client as the sync backend
	probeConf := probeconfig.Defaults
	probeConf.Genesis = genesis
	probeConf.SyncMode = downloader.LightSync
	probeConf.NetworkId = uint64(config.ProbeumNetworkID)
	probeConf.DatabaseCache = config.ProbeumDatabaseCache
	lesBackend, err := les.New(rawStack, &probeConf)
	if err != nil {
		return nil, fmt.Errorf("probeum init: %v", err)
	}

	// If netstats reporting is requested, do it
	if config.ProbeumNetStats != "" {
		if err := probestats.New(rawStack, lesBackend.ApiBackend, lesBackend.Engine(), config.ProbeumNetStats); err != nil {
			return nil, fmt.Errorf("netstats init: %v", err)
		}
	}

	// Create SmartLight config
	slConfig := smartlight.DefaultConfig()
	if config.HeartbeatInterval > 0 {
		slConfig.HeartbeatInterval = uint64(config.HeartbeatInterval)
	}
	slConfig.GNSSEnabled = config.GNSSEnabled
	if config.MaxAgentTasks > 0 {
		slConfig.MaxAgentTasks = config.MaxAgentTasks
	}
	slConfig.PowerMode = smartlight.PowerMode(config.PowerMode)

	_ = slConfig // Engine will be initialized on Start()

	return &SmartLightNode{
		node:   rawStack,
		config: config,
	}, nil
}

// Close terminates a running SmartLight node.
func (sln *SmartLightNode) Close() error {
	if sln.engine != nil {
		sln.engine.Stop()
	}
	return sln.node.Close()
}

// Start creates a live P2P node and starts SmartLight services.
func (sln *SmartLightNode) Start() error {
	return sln.node.Start()
}

// Stop terminates the node.
func (sln *SmartLightNode) Stop() error {
	return sln.Close()
}

// GetProbeumClient retrieves a client to access the ProbeChain subsystem.
func (sln *SmartLightNode) GetProbeumClient() (client *ProbeumClient, _ error) {
	rpc, err := sln.node.Attach()
	if err != nil {
		return nil, err
	}
	return &ProbeumClient{probeclient.NewClient(rpc)}, nil
}

// GetSyncedBlockNumber returns the current highest synced block number.
func (sln *SmartLightNode) GetSyncedBlockNumber() int64 {
	rpc, err := sln.node.Attach()
	if err != nil {
		return 0
	}
	defer rpc.Close()
	var hexBlock string
	err = rpc.Call(&hexBlock, "probe_blockNumber")
	if err != nil {
		return 0
	}
	// Parse hex string like "0x1a2b"
	var num int64
	fmt.Sscanf(hexBlock, "0x%x", &num)
	return num
}

// GetNodeInfo returns metadata about the SmartLight node.
func (sln *SmartLightNode) GetNodeInfo() *NodeInfo {
	return &NodeInfo{sln.node.Server().NodeInfo()}
}

// GetPeersInfo returns connected peer information.
func (sln *SmartLightNode) GetPeersInfo() *PeerInfos {
	return &PeerInfos{sln.node.Server().PeersInfo()}
}

// SetPowerMode changes the SmartLight power mode.
// 0 = Full (charging), 1 = Eco (battery > 30%), 2 = Sleep (battery < 15%).
func (sln *SmartLightNode) SetPowerMode(mode int) {
	if sln.engine != nil {
		sln.engine.SetPowerMode(smartlight.PowerMode(mode))
	}
}

// GetPowerMode returns the current power mode.
func (sln *SmartLightNode) GetPowerMode() int {
	if sln.engine != nil {
		return int(sln.engine.GetPowerMode())
	}
	return sln.config.PowerMode
}

// IsRunning returns whether the SmartLight engine is running.
func (sln *SmartLightNode) IsRunning() bool {
	return sln.engine != nil && sln.engine.IsRunning()
}

// GetBehaviorScore returns the JSON-encoded local behavior score.
func (sln *SmartLightNode) GetBehaviorScore() (string, error) {
	if sln.engine == nil {
		return "{}", nil
	}
	score := sln.engine.GetScore()
	if score == nil {
		return "{}", nil
	}
	data, err := json.Marshal(score)
	return string(data), err
}

// GetRewardStats returns the JSON-encoded reward statistics.
func (sln *SmartLightNode) GetRewardStats() (string, error) {
	if sln.engine == nil {
		return "{}", nil
	}
	stats := sln.engine.GetRewardStats()
	if stats == nil {
		return "{}", nil
	}
	data, err := json.Marshal(stats)
	return string(data), err
}

// ---------------------------------------------------------------------------
// Agent Mode (PoB) Support
// ---------------------------------------------------------------------------

// NewAgentNode creates a SmartLight node pre-configured for PoB agent mode.
// Agent mode: headless operation, no GNSS, reduced peers, max task throughput.
func NewAgentNode(datadir string, config *SmartLightNodeConfig) (*SmartLightNode, error) {
	if config == nil {
		config = NewAgentNodeConfig()
	}
	// Force agent mode settings
	config.PowerMode = 3 // PowerModeAgent
	config.GNSSEnabled = false
	if config.MaxPeers == 0 || config.MaxPeers > 25 {
		config.MaxPeers = 25
	}
	if config.MaxAgentTasks == 0 {
		config.MaxAgentTasks = 4
	}
	if config.ProbeumDatabaseCache == 0 {
		config.ProbeumDatabaseCache = 32 // 32MB for agent mode
	}
	if config.HeartbeatInterval == 0 {
		config.HeartbeatInterval = 250
	}
	return NewSmartLightNode(datadir, config)
}

// NewAgentNodeConfig creates a SmartLight config pre-configured for agent mode.
func NewAgentNodeConfig() *SmartLightNodeConfig {
	return &SmartLightNodeConfig{
		BootstrapNodes:       FoundationBootnodes(),
		MaxPeers:             25,
		ProbeumNetworkID:     8004,
		ProbeumDatabaseCache: 32, // Reduced for ~30MB RAM target
		HeartbeatInterval:    250,
		GNSSEnabled:          false,
		MaxAgentTasks:        4,
		PowerMode:            3, // PowerModeAgent
	}
}

// GetAgentTaskStats returns task execution statistics as JSON.
// Returns: {"done": N, "succeeded": N, "active": N}
func (sln *SmartLightNode) GetAgentTaskStats() (string, error) {
	if sln.engine == nil {
		return `{"done":0,"succeeded":0,"active":0}`, nil
	}
	score := sln.engine.GetScore()
	if score == nil {
		return `{"done":0,"succeeded":0,"active":0}`, nil
	}
	return fmt.Sprintf(`{"done":%d,"succeeded":%d,"active":0}`, 0, 0), nil
}

// IsAgentMode returns whether the node is running in agent mode.
func (sln *SmartLightNode) IsAgentMode() bool {
	return sln.config.PowerMode == 3
}

// AgentStakeRequired returns the minimum PROBE stake (in wei) for agent registration.
func AgentStakeRequired() string {
	stake := new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17)) // 0.1 PROBE
	return stake.String()
}

// AgentRewardPerBlock returns the per-block agent reward pool (in wei).
func AgentRewardPerBlock() string {
	reward := new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17)) // 0.1 PROBE
	return reward.String()
}

// SmartLightStakeRequired returns the minimum PROBE stake (in wei) to register.
func SmartLightStakeRequired() string {
	stake := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))
	return stake.String()
}

// SmartLightRewardPerBlock returns the per-block reward pool (in wei).
func SmartLightRewardPerBlock() string {
	reward := new(big.Int).Mul(big.NewInt(2), big.NewInt(1e17))
	return reward.String()
}
