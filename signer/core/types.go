// Package core implements the core functionality of the external signer (clef).
// This is a minimal stub providing only the SendTxArgs type.
package core

import (
	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/common/hexutil"
	"github.com/probechain/go-probe/core/types"
)

// SendTxArgs represents the arguments to submit a transaction.
type SendTxArgs struct {
	From                 common.MixedcaseAddress  `json:"from"`
	To                   *common.MixedcaseAddress `json:"to"`
	Gas                  hexutil.Uint64           `json:"gas"`
	GasPrice             *hexutil.Big             `json:"gasPrice"`
	MaxFeePerGas         *hexutil.Big             `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *hexutil.Big             `json:"maxPriorityFeePerGas"`
	Value                hexutil.Big              `json:"value"`
	Nonce                hexutil.Uint64           `json:"nonce"`
	Data                 *hexutil.Bytes           `json:"data"`
	Input                *hexutil.Bytes           `json:"input,omitempty"`
	ChainID              *hexutil.Big             `json:"chainId,omitempty"`
	AccessList           *types.AccessList        `json:"accessList,omitempty"`
}
