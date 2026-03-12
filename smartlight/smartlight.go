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

// Package smartlight implements the SmartLight engine for ProbeChain.
// SmartLight upgrades LES light clients from "read-only sync" to "behavior contributors"
// that participate in PoB consensus through ACK attestations, heartbeat proofs,
// GNSS time samples, and lightweight agent tasks.
package smartlight

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/core/types"
	"github.com/probechain/go-probe/log"
)

var (
	errNotRunning   = errors.New("smartlight: engine not running")
	errAlreadyRunning = errors.New("smartlight: engine already running")
	errNotRegistered  = errors.New("smartlight: node not registered")
)

// HeaderSource provides new block headers to the SmartLight engine.
type HeaderSource interface {
	// SubscribeNewHead subscribes to new block header events.
	SubscribeNewHead(ch chan<- *types.Header) error
	// CurrentHeader returns the current head header.
	CurrentHeader() *types.Header
}

// SignerFunc signs a message with the SmartLight node's key.
type SignerFunc func(data []byte) ([]byte, error)

// GNSSProvider provides GNSS time samples from the device.
type GNSSProvider interface {
	// Sample returns the current GNSS time as an encoded AtomicTimestamp.
	Sample() ([]byte, error)
	// Available returns whether GNSS is currently available.
	Available() bool
}

// PeerSender sends SLP messages to connected full nodes.
type PeerSender interface {
	// Send sends an SLP message to connected peers.
	Send(msgCode uint64, data interface{}) error
}

// Engine is the SmartLight orchestrator that coordinates all services.
type Engine struct {
	config *Config

	address  common.Address
	signFn   SignerFunc
	sender   PeerSender

	ackService   *AckService
	heartbeat    *HeartbeatService
	timeService  *TimeService
	taskRunner   *TaskRunner
	rewardTracker *RewardTracker

	headerCh chan *types.Header

	running   int32 // atomic
	quit      chan struct{}
	wg        sync.WaitGroup

	mu sync.RWMutex
}

// New creates a new SmartLight engine.
func New(config *Config, address common.Address, signFn SignerFunc, sender PeerSender) *Engine {
	if config == nil {
		config = DefaultConfig()
	}
	return &Engine{
		config:   config,
		address:  address,
		signFn:   signFn,
		sender:   sender,
		headerCh: make(chan *types.Header, 64),
	}
}

// Start initializes and starts all SmartLight services.
func (e *Engine) Start(headerSource HeaderSource, gnss GNSSProvider) error {
	if !atomic.CompareAndSwapInt32(&e.running, 0, 1) {
		return errAlreadyRunning
	}
	e.quit = make(chan struct{})

	// Initialize sub-services
	e.ackService = NewAckService(e.config, e.address, e.signFn, e.sender)
	e.heartbeat = NewHeartbeatService(e.config, e.address, e.signFn, e.sender)
	e.taskRunner = NewTaskRunner(e.config, e.address, e.signFn, e.sender)
	e.rewardTracker = NewRewardTracker(e.address)

	if gnss != nil && e.config.GNSSEnabled {
		e.timeService = NewTimeService(e.config, e.address, e.signFn, e.sender, gnss)
	}

	// Subscribe to new headers
	if err := headerSource.SubscribeNewHead(e.headerCh); err != nil {
		atomic.StoreInt32(&e.running, 0)
		return err
	}

	// Start the main event loop
	e.wg.Add(1)
	go e.loop()

	// Start the heartbeat service
	e.wg.Add(1)
	go e.heartbeat.Run(e.quit, &e.wg)

	// Start the time service if available
	if e.timeService != nil {
		e.wg.Add(1)
		go e.timeService.Run(e.quit, &e.wg)
	}

	log.Info("SmartLight engine started",
		"address", e.address.Hex(),
		"mode", e.config.PowerMode.String(),
		"gnss", e.config.GNSSEnabled,
	)
	return nil
}

// Stop gracefully shuts down all SmartLight services.
func (e *Engine) Stop() error {
	if !atomic.CompareAndSwapInt32(&e.running, 1, 0) {
		return errNotRunning
	}
	close(e.quit)
	e.wg.Wait()
	e.taskRunner.Stop()
	log.Info("SmartLight engine stopped")
	return nil
}

// loop is the main event loop that processes new headers.
func (e *Engine) loop() {
	defer e.wg.Done()
	for {
		select {
		case header := <-e.headerCh:
			e.handleNewHeader(header)
		case <-e.quit:
			return
		}
	}
}

// handleNewHeader processes a new block header.
func (e *Engine) handleNewHeader(header *types.Header) {
	if header == nil {
		return
	}

	blockNum := header.Number.Uint64()

	// In Sleep mode, only sync — no ACKs or other activity
	if e.config.PowerMode == PowerModeSleep {
		return
	}

	// In Agent mode, prioritize task execution over ACK/heartbeat display
	if e.config.PowerMode == PowerModeAgent {
		// Still send ACKs and heartbeats, but skip GNSS and display
		e.ackService.ProcessHeader(header)
		e.heartbeat.UpdateLatestBlock(blockNum, header.Hash())
		if blockNum%30000 == 0 {
			e.rewardTracker.OnEpochBoundary(blockNum)
		}
		return
	}

	// Send ACK attestation for the header
	e.ackService.ProcessHeader(header)

	// Update heartbeat with latest block
	e.heartbeat.UpdateLatestBlock(blockNum, header.Hash())

	// Track rewards at epoch boundaries
	if blockNum%30000 == 0 {
		e.rewardTracker.OnEpochBoundary(blockNum)
	}
}

// SetPowerMode changes the operating power mode.
func (e *Engine) SetPowerMode(mode PowerMode) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.PowerMode = mode
	log.Info("SmartLight power mode changed", "mode", mode.String())
}

// GetPowerMode returns the current power mode.
func (e *Engine) GetPowerMode() PowerMode {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config.PowerMode
}

// GetScore returns the local behavior score.
func (e *Engine) GetScore() *SmartLightScore {
	if e.rewardTracker == nil {
		return nil
	}
	return e.rewardTracker.GetScore()
}

// GetRewardStats returns the accumulated reward statistics.
func (e *Engine) GetRewardStats() *RewardStats {
	if e.rewardTracker == nil {
		return nil
	}
	return e.rewardTracker.GetStats()
}

// Address returns the SmartLight node's address.
func (e *Engine) Address() common.Address {
	return e.address
}

// IsRunning returns whether the engine is currently running.
func (e *Engine) IsRunning() bool {
	return atomic.LoadInt32(&e.running) == 1
}

// HandleTaskRequest processes an incoming agent task assignment.
func (e *Engine) HandleTaskRequest(task *TaskRequestMsg) {
	if e.config.PowerMode != PowerModeFull && e.config.PowerMode != PowerModeAgent {
		return // Only execute tasks in Full or Agent mode
	}
	e.taskRunner.Submit(task)
}
