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

package types

import (
	"math/big"

	"github.com/probechain/go-probe/common"
)

// Agent operation codes for PoB consensus.
const (
	AgentOpRegister  = uint8(0) // Register as an agent node
	AgentOpHeartbeat = uint8(1) // Agent heartbeat proof
	AgentOpTaskResult = uint8(2) // Task execution result
	AgentOpAttest    = uint8(3) // Cross-agent attestation
)

// AgentTx is a transaction type for PoB agent consensus operations.
// It carries agent registration, heartbeat, task result, and attestation data.
type AgentTx struct {
	ChainID    *big.Int
	Nonce      uint64
	GasTipCap  *big.Int        // a.k.a. maxPriorityFeePerGas
	GasFeeCap  *big.Int        // a.k.a. maxFeePerGas
	Gas        uint64
	To         *common.Address `rlp:"nil"` // Agent registry address
	Value      *big.Int
	Data       []byte          // ABI-encoded agent operation
	AgentID    common.Hash     // ERC-8004 style identity hash
	OpCode     uint8           // Agent operation code (Register/Heartbeat/TaskResult/Attest)
	AccessList AccessList

	// Signature fields — supports both ECDSA and Ed25519/Dilithium
	V *big.Int
	R *big.Int
	S *big.Int
}

// copy creates a deep copy of the transaction data.
func (tx *AgentTx) copy() TxData {
	cpy := &AgentTx{
		Nonce:   tx.Nonce,
		To:      tx.To,
		Data:    common.CopyBytes(tx.Data),
		Gas:     tx.Gas,
		AgentID: tx.AgentID,
		OpCode:  tx.OpCode,
		// Deep copy below.
		AccessList: make(AccessList, len(tx.AccessList)),
		Value:      new(big.Int),
		ChainID:    new(big.Int),
		GasTipCap:  new(big.Int),
		GasFeeCap:  new(big.Int),
		V:          new(big.Int),
		R:          new(big.Int),
		S:          new(big.Int),
	}
	copy(cpy.AccessList, tx.AccessList)
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
	}
	if tx.GasTipCap != nil {
		cpy.GasTipCap.Set(tx.GasTipCap)
	}
	if tx.GasFeeCap != nil {
		cpy.GasFeeCap.Set(tx.GasFeeCap)
	}
	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}
	return cpy
}

// accessors for innerTx.
func (tx *AgentTx) txType() byte           { return AgentTxType }
func (tx *AgentTx) chainID() *big.Int      { return tx.ChainID }
func (tx *AgentTx) protected() bool        { return true }
func (tx *AgentTx) accessList() AccessList { return tx.AccessList }
func (tx *AgentTx) data() []byte           { return tx.Data }
func (tx *AgentTx) gas() uint64            { return tx.Gas }
func (tx *AgentTx) gasFeeCap() *big.Int    { return tx.GasFeeCap }
func (tx *AgentTx) gasTipCap() *big.Int    { return tx.GasTipCap }
func (tx *AgentTx) gasPrice() *big.Int     { return tx.GasFeeCap }
func (tx *AgentTx) value() *big.Int        { return tx.Value }
func (tx *AgentTx) nonce() uint64          { return tx.Nonce }
func (tx *AgentTx) to() *common.Address    { return tx.To }

func (tx *AgentTx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

func (tx *AgentTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID, tx.V, tx.R, tx.S = chainID, v, r, s
}
