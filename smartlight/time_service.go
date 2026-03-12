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
	"sync"
	"time"

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/core/atomic"
	"github.com/probechain/go-probe/log"
)

// TimeService samples GNSS time from the device (iPhone GPS) and submits
// AtomicTimestamp samples to the network for signal sovereignty scoring.
type TimeService struct {
	config  *Config
	address common.Address
	signFn  SignerFunc
	sender  PeerSender
	gnss    GNSSProvider

	mu           sync.RWMutex
	samplesSent  uint64
	latestBlock  uint64
}

// NewTimeService creates a new GNSS time sampling service.
func NewTimeService(config *Config, address common.Address, signFn SignerFunc, sender PeerSender, gnss GNSSProvider) *TimeService {
	return &TimeService{
		config:  config,
		address: address,
		signFn:  signFn,
		sender:  sender,
		gnss:    gnss,
	}
}

// Run starts the periodic GNSS sampling loop.
func (t *TimeService) Run(quit <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(t.config.GNSSSampleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.sample()
		case <-quit:
			return
		}
	}
}

// sample takes a GNSS time sample and submits it.
func (t *TimeService) sample() {
	if !t.gnss.Available() {
		return
	}

	atomicTimeBytes, err := t.gnss.Sample()
	if err != nil {
		log.Debug("SmartLight: GNSS sample failed", "err", err)
		return
	}

	// Validate the sample
	at, err := atomic.DecodeAtomicTimestamp(atomicTimeBytes)
	if err != nil {
		log.Warn("SmartLight: invalid GNSS timestamp", "err", err)
		return
	}

	t.mu.RLock()
	blockNum := t.latestBlock
	t.mu.RUnlock()

	// Sign the time sample
	msg := make([]byte, len(atomicTimeBytes)+8)
	copy(msg, atomicTimeBytes)
	msg[len(atomicTimeBytes)] = byte(blockNum >> 56)
	msg[len(atomicTimeBytes)+1] = byte(blockNum >> 48)
	msg[len(atomicTimeBytes)+2] = byte(blockNum >> 40)
	msg[len(atomicTimeBytes)+3] = byte(blockNum >> 32)
	msg[len(atomicTimeBytes)+4] = byte(blockNum >> 24)
	msg[len(atomicTimeBytes)+5] = byte(blockNum >> 16)
	msg[len(atomicTimeBytes)+6] = byte(blockNum >> 8)
	msg[len(atomicTimeBytes)+7] = byte(blockNum)

	sig, err := t.signFn(msg)
	if err != nil {
		log.Warn("SmartLight: time sample sign failed", "err", err)
		return
	}

	tsMsg := TimeSampleMsg{
		Address:       t.address,
		AtomicTime:    atomicTimeBytes,
		BlockNumber:   blockNum,
		GPSAccuracyNs: at.Uncertainty,
		Signature:     sig,
	}

	if err := t.sender.Send(SLPTimeSampleMsg, tsMsg); err != nil {
		log.Warn("SmartLight: time sample send failed", "err", err)
		return
	}

	t.mu.Lock()
	t.samplesSent++
	t.mu.Unlock()

	log.Debug("SmartLight: GNSS time sample sent", "block", blockNum, "uncertainty", at.Uncertainty)
}

// UpdateLatestBlock updates the latest synced block number.
func (t *TimeService) UpdateLatestBlock(blockNum uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.latestBlock = blockNum
}

// SamplesSent returns the total GNSS time samples sent.
func (t *TimeService) SamplesSent() uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.samplesSent
}
