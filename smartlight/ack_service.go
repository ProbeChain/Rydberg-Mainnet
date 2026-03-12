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
	"github.com/probechain/go-probe/core/types"
	"github.com/probechain/go-probe/log"
)

// AckService monitors new headers, verifies them, signs ACK attestations,
// and batches them for broadcast to full nodes.
type AckService struct {
	config  *Config
	address common.Address
	signFn  SignerFunc
	sender  PeerSender

	mu       sync.Mutex
	pending  []HeaderAck
	lastSend time.Time
	acksGiven uint64
}

// NewAckService creates a new ACK attestation service.
func NewAckService(config *Config, address common.Address, signFn SignerFunc, sender PeerSender) *AckService {
	return &AckService{
		config:   config,
		address:  address,
		signFn:   signFn,
		sender:   sender,
		pending:  make([]HeaderAck, 0, 32),
		lastSend: time.Now(),
	}
}

// ProcessHeader verifies a header and adds an ACK attestation to the pending batch.
func (s *AckService) ProcessHeader(header *types.Header) {
	if header == nil {
		return
	}

	blockNum := header.Number.Uint64()
	blockHash := header.Hash()

	// Light verification: check that header fields are internally consistent
	valid := s.verifyHeader(header)

	// Sign the attestation
	sig, err := s.signAck(blockNum, blockHash, valid)
	if err != nil {
		log.Warn("SmartLight: failed to sign ACK", "block", blockNum, "err", err)
		return
	}

	ack := HeaderAck{
		BlockNumber: blockNum,
		BlockHash:   blockHash,
		Valid:       valid,
		Signature:   sig,
	}

	s.mu.Lock()
	s.pending = append(s.pending, ack)
	s.acksGiven++
	shouldFlush := time.Since(s.lastSend) >= s.config.AckBatchInterval || len(s.pending) >= 32
	s.mu.Unlock()

	if shouldFlush {
		s.Flush()
	}
}

// Flush sends all pending ACKs as a batch.
func (s *AckService) Flush() {
	s.mu.Lock()
	if len(s.pending) == 0 {
		s.mu.Unlock()
		return
	}
	batch := make([]HeaderAck, len(s.pending))
	copy(batch, s.pending)
	s.pending = s.pending[:0]
	s.lastSend = time.Now()
	s.mu.Unlock()

	msg := AckBatchMsg{
		Address: s.address,
		Acks:    batch,
	}
	if err := s.sender.Send(SLPAckBatchMsg, msg); err != nil {
		log.Warn("SmartLight: failed to send ACK batch", "count", len(batch), "err", err)
	} else {
		log.Debug("SmartLight: sent ACK batch", "count", len(batch))
	}
}

// AcksGiven returns the total number of ACKs signed.
func (s *AckService) AcksGiven() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.acksGiven
}

// verifyHeader performs lightweight header verification suitable for a light node.
func (s *AckService) verifyHeader(header *types.Header) bool {
	// Basic sanity checks that don't require full state
	if header.Number == nil || header.Number.Sign() < 0 {
		return false
	}
	if header.GasUsed > header.GasLimit {
		return false
	}
	if header.Difficulty == nil || header.Difficulty.Sign() <= 0 {
		return false
	}
	return true
}

// signAck creates a signature over (blockNumber || blockHash || valid).
func (s *AckService) signAck(blockNum uint64, blockHash common.Hash, valid bool) ([]byte, error) {
	// Build message: 8 bytes blockNum + 32 bytes hash + 1 byte valid
	msg := make([]byte, 41)
	binary.BigEndian.PutUint64(msg[0:8], blockNum)
	copy(msg[8:40], blockHash[:])
	if valid {
		msg[40] = 1
	}
	return s.signFn(msg)
}
