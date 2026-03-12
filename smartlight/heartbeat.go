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

package smartlight

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/log"
)

// HeartbeatService periodically sends signed heartbeat proofs to demonstrate
// the SmartLight node is active and synced.
type HeartbeatService struct {
	config  *Config
	address common.Address
	signFn  SignerFunc
	sender  PeerSender

	mu              sync.RWMutex
	latestBlock     uint64
	latestBlockHash common.Hash
	heartbeatsSent  uint64
	lastHeartbeat   uint64 // block number of last heartbeat
}

// NewHeartbeatService creates a new heartbeat service.
func NewHeartbeatService(config *Config, address common.Address, signFn SignerFunc, sender PeerSender) *HeartbeatService {
	return &HeartbeatService{
		config:  config,
		address: address,
		signFn:  signFn,
		sender:  sender,
	}
}

// Run starts the heartbeat loop. It sends a heartbeat every HeartbeatInterval blocks.
func (h *HeartbeatService) Run(quit <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// Check every 5 seconds if we need to send a heartbeat
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.maybeHeartbeat()
		case <-quit:
			return
		}
	}
}

// UpdateLatestBlock updates the latest known block info.
func (h *HeartbeatService) UpdateLatestBlock(number uint64, hash common.Hash) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.latestBlock = number
	h.latestBlockHash = hash
}

// maybeHeartbeat sends a heartbeat if enough blocks have passed.
func (h *HeartbeatService) maybeHeartbeat() {
	h.mu.RLock()
	blockNum := h.latestBlock
	blockHash := h.latestBlockHash
	lastHB := h.lastHeartbeat
	h.mu.RUnlock()

	if blockNum == 0 {
		return // Not synced yet
	}

	// Send heartbeat every HeartbeatInterval blocks
	if blockNum-lastHB < h.config.HeartbeatInterval {
		return
	}

	timestamp := uint64(time.Now().Unix())

	// Sign: address || blockNumber || blockHash || timestamp
	msg := make([]byte, 20+8+32+8)
	copy(msg[0:20], h.address[:])
	binary.BigEndian.PutUint64(msg[20:28], blockNum)
	copy(msg[28:60], blockHash[:])
	binary.BigEndian.PutUint64(msg[60:68], timestamp)

	sig, err := h.signFn(msg)
	if err != nil {
		log.Warn("SmartLight: heartbeat sign failed", "err", err)
		return
	}

	hbMsg := HeartbeatMsg{
		Address:     h.address,
		BlockNumber: blockNum,
		BlockHash:   blockHash,
		Timestamp:   timestamp,
		Signature:   sig,
	}

	if err := h.sender.Send(SLPHeartbeatMsg, hbMsg); err != nil {
		log.Warn("SmartLight: heartbeat send failed", "err", err)
		return
	}

	h.mu.Lock()
	h.lastHeartbeat = blockNum
	h.heartbeatsSent++
	h.mu.Unlock()

	log.Debug("SmartLight: heartbeat sent", "block", blockNum)
}

// HeartbeatsSent returns the total heartbeats sent.
func (h *HeartbeatService) HeartbeatsSent() uint64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.heartbeatsSent
}
