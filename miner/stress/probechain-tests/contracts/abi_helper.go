package contracts

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/crypto"
)

// ABI encoding helpers — we manually encode because we want full control
// and transparency over what gets sent on-chain. No magic, no hidden state.

// Keccak256 of function signature, first 4 bytes = selector
func FuncSelector(sig string) []byte {
	hash := crypto.Keccak256([]byte(sig))
	return hash[:4]
}

// EncodeFunctionCall encodes a function call with the given arguments.
// Each arg must be a pre-encoded 32-byte word.
func EncodeFunctionCall(sig string, args ...[]byte) []byte {
	selector := FuncSelector(sig)
	data := make([]byte, 4+len(args)*32)
	copy(data[:4], selector)
	for i, arg := range args {
		// Right-pad or left-pad to 32 bytes depending on type
		copy(data[4+i*32:4+(i+1)*32], padTo32(arg))
	}
	return data
}

// EncodeAddress pads an address to 32 bytes (left-padded with zeros).
func EncodeAddress(addr common.Address) []byte {
	var buf [32]byte
	copy(buf[12:], addr.Bytes())
	return buf[:]
}

// EncodeUint256 encodes a big.Int as a 32-byte word.
func EncodeUint256(val *big.Int) []byte {
	var buf [32]byte
	if val == nil {
		return buf[:]
	}
	b := val.Bytes()
	copy(buf[32-len(b):], b)
	return buf[:]
}

// EncodeUint8 encodes a uint8 as a 32-byte word.
func EncodeUint8(val uint8) []byte {
	var buf [32]byte
	buf[31] = val
	return buf[:]
}

// EncodeBool encodes a boolean as a 32-byte word.
func EncodeBool(val bool) []byte {
	var buf [32]byte
	if val {
		buf[31] = 1
	}
	return buf[:]
}

// EncodeString encodes a dynamic string for ABI.
// Returns (offset_word, encoded_data).
func EncodeString(s string) ([]byte, []byte) {
	strBytes := []byte(s)
	length := len(strBytes)

	// Pad string to 32-byte boundary
	padded := length
	if padded%32 != 0 {
		padded += 32 - (padded % 32)
	}

	data := make([]byte, 32+padded) // 32 for length + padded content
	copy(data[:32], EncodeUint256(big.NewInt(int64(length))))
	copy(data[32:], strBytes)

	return data, nil
}

// EncodeFunctionCallWithString encodes a call with one string parameter.
// For functions like register(string agentURI).
func EncodeFunctionCallWithString(sig string, s string) []byte {
	selector := FuncSelector(sig)

	// Dynamic type: offset (32) + length (32) + padded data
	strBytes := []byte(s)
	length := len(strBytes)
	padded := length
	if padded%32 != 0 {
		padded += 32 - (padded % 32)
	}

	// Total: 4 (selector) + 32 (offset) + 32 (length) + padded
	data := make([]byte, 4+32+32+padded)
	copy(data[:4], selector)
	// Offset to string data (always 0x20 = 32 for single string param)
	copy(data[4:36], EncodeUint256(big.NewInt(32)))
	// String length
	copy(data[36:68], EncodeUint256(big.NewInt(int64(length))))
	// String content
	copy(data[68:], strBytes)

	return data
}

// EncodeFunctionCallWithStringAndArgs encodes selector + fixed args + trailing string.
func EncodeFunctionCallWithStringAndArgs(sig string, fixedArgs [][]byte, s string) []byte {
	selector := FuncSelector(sig)

	strBytes := []byte(s)
	length := len(strBytes)
	padded := length
	if padded%32 != 0 {
		padded += 32 - (padded % 32)
	}

	// Offset to string data: after all fixed args + offset word itself
	offset := int64((len(fixedArgs) + 1) * 32)

	totalSize := 4 + (len(fixedArgs)+1)*32 + 32 + padded
	data := make([]byte, totalSize)
	copy(data[:4], selector)

	pos := 4
	for _, arg := range fixedArgs {
		copy(data[pos:pos+32], padTo32(arg))
		pos += 32
	}
	// Offset to string
	copy(data[pos:pos+32], EncodeUint256(big.NewInt(offset)))
	pos += 32
	// String length
	copy(data[pos:pos+32], EncodeUint256(big.NewInt(int64(length))))
	pos += 32
	// String content
	copy(data[pos:], strBytes)

	return data
}

// DecodeUint256 reads a big.Int from a 32-byte ABI word.
func DecodeUint256(data []byte) *big.Int {
	if len(data) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(data):], data)
		data = padded
	}
	return new(big.Int).SetBytes(data[:32])
}

// DecodeAddress reads an address from a 32-byte ABI word.
func DecodeAddress(data []byte) common.Address {
	if len(data) < 32 {
		return common.Address{}
	}
	return common.BytesToAddress(data[12:32])
}

// DecodeBool reads a boolean from a 32-byte ABI word.
func DecodeBool(data []byte) bool {
	if len(data) < 32 {
		return false
	}
	return data[31] != 0
}

// padTo32 left-pads data to 32 bytes.
func padTo32(data []byte) []byte {
	if len(data) >= 32 {
		return data[:32]
	}
	padded := make([]byte, 32)
	copy(padded[32-len(data):], data)
	return padded
}

// HexToBytes decodes a hex string (with or without 0x prefix) to bytes.
func HexToBytes(s string) ([]byte, error) {
	s = strings.TrimPrefix(s, "0x")
	return hex.DecodeString(s)
}

// BytesToHex encodes bytes to a 0x-prefixed hex string.
func BytesToHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

// Event topic hashes for known events
var (
	// Transfer(address,address,uint256)
	TopicTransfer = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	// Approval(address,address,uint256)
	TopicApproval = crypto.Keccak256Hash([]byte("Approval(address,address,uint256)"))
	// Trade(address,address,bool,uint256,uint256,uint256)
	TopicTrade = crypto.Keccak256Hash([]byte("Trade(address,address,bool,uint256,uint256,uint256)"))
	// Registered(uint256,string,address)
	TopicRegistered = crypto.Keccak256Hash([]byte("Registered(uint256,string,address)"))
	// PairCreated(address,address,address,uint256)
	TopicPairCreated = crypto.Keccak256Hash([]byte("PairCreated(address,address,address,uint256)"))
	// Swap(address,uint256,uint256,uint256,uint256,address)
	TopicSwap = crypto.Keccak256Hash([]byte("Swap(address,uint256,uint256,uint256,uint256,address)"))
	// Deposit(address,uint256)
	TopicDeposit = crypto.Keccak256Hash([]byte("Deposit(address,uint256)"))
	// Withdrawal(address,uint256)
	TopicWithdrawal = crypto.Keccak256Hash([]byte("Withdrawal(address,uint256)"))
)

// Commonly used function selectors
func init() {
	// Pre-validate selectors
	_ = FuncSelector("transfer(address,uint256)")
	_ = FuncSelector("approve(address,uint256)")
	_ = FuncSelector("balanceOf(address)")
	_ = FuncSelector("totalSupply()")
}

// FormatProbe formats wei to PROBE string for display.
func FormatProbe(wei *big.Int) string {
	if wei == nil {
		return "0 PROBE"
	}
	probe := new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetInt64(1e18))
	return fmt.Sprintf("%.6f PROBE", probe)
}
