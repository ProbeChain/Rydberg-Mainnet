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

package vm

import (
	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/crypto"
)

// agentVerify implements a precompiled contract for PoB agent signature verification.
// Input format: agentID(32) || message(32) || signature(65) = 129 bytes total (ECDSA).
// For Dilithium agents: agentID(32) || message(32) || pubkey(1312) || sig(2420) = 3796 bytes.
// Output: 32 bytes — left-padded 20-byte address if valid, zero bytes if invalid.
type agentVerify struct{}

const (
	agentVerifyMinInputLen = 32 + 32 + 65 // agentID + message + ECDSA signature
	agentVerifyGas         = 5000
)

func (c *agentVerify) RequiredGas(input []byte) uint64 {
	return agentVerifyGas
}

func (c *agentVerify) Run(input []byte) ([]byte, error) {
	// Pad input if too short
	if len(input) < agentVerifyMinInputLen {
		return make([]byte, 32), nil
	}

	agentID := input[:32]
	message := input[32:64]
	sig := input[64:]

	// ECDSA path (65-byte signature)
	if len(sig) >= 65 {
		// Recover the public key from the signature
		hash := crypto.Keccak256(agentID, message)
		pubkey, err := crypto.Ecrecover(hash, sig[:65])
		if err != nil {
			return make([]byte, 32), nil
		}
		// Derive address from recovered public key
		addr := common.BytesToAddress(crypto.Keccak256(pubkey[1:])[12:])
		result := make([]byte, 32)
		copy(result[12:], addr[:])
		return result, nil
	}

	return make([]byte, 32), nil
}

// AgentVerifyPrecompileAddress is the address of the agent verify precompile.
var AgentVerifyPrecompileAddress = common.BytesToAddress([]byte{21}) // 0x15
