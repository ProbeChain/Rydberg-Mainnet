package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/common/hexutil"
	"github.com/probechain/go-probe/core/types"
	"github.com/probechain/go-probe/crypto"
	"github.com/probechain/go-probe/rlp"
)

// RPCClient is a lightweight JSON-RPC client that preserves raw responses
// for evidence. It does NOT interpret results — callers must verify on-chain state.
type RPCClient struct {
	url    string
	client *http.Client
	id     uint64
}

// RPCRequest is a JSON-RPC 2.0 request.
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      uint64      `json:"id"`
}

// RPCResponse is a JSON-RPC 2.0 response, keeping raw result for evidence.
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      uint64          `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *RPCError       `json:"error,omitempty"`
	// RawBody stores the complete HTTP response for evidence
	RawBody string `json:"-"`
}

// RPCError is a JSON-RPC error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// NewRPCClient creates a client for the given RPC endpoint.
func NewRPCClient(url string) *RPCClient {
	return &RPCClient{
		url: url,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Call makes a JSON-RPC call and returns the raw response.
func (c *RPCClient) Call(ctx context.Context, method string, params ...interface{}) (*RPCResponse, error) {
	id := atomic.AddUint64(&c.id, 1)

	if params == nil {
		params = []interface{}{}
	}

	req := RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var rpcResp RPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("unmarshal response (raw: %s): %w", string(respBody), err)
	}
	rpcResp.RawBody = string(respBody)

	if rpcResp.Error != nil {
		return &rpcResp, rpcResp.Error
	}

	return &rpcResp, nil
}

// GetBalance queries the balance of an address. Returns the raw hex and parsed big.Int.
func (c *RPCClient) GetBalance(ctx context.Context, address string, block string) (*big.Int, string, error) {
	if block == "" {
		block = "latest"
	}
	resp, err := c.Call(ctx, "probe_getBalance", address, block)
	if err != nil {
		return nil, "", err
	}

	var hexBalance string
	if err := json.Unmarshal(resp.Result, &hexBalance); err != nil {
		return nil, resp.RawBody, fmt.Errorf("parse balance: %w", err)
	}

	balance, ok := new(big.Int).SetString(strings.TrimPrefix(hexBalance, "0x"), 16)
	if !ok {
		return nil, resp.RawBody, fmt.Errorf("invalid balance hex: %s", hexBalance)
	}

	return balance, resp.RawBody, nil
}

// GetTransactionCount returns the nonce for an address.
func (c *RPCClient) GetTransactionCount(ctx context.Context, address string, block string) (uint64, error) {
	if block == "" {
		block = "pending"
	}
	resp, err := c.Call(ctx, "probe_getTransactionCount", address, block)
	if err != nil {
		return 0, err
	}

	var hexNonce string
	if err := json.Unmarshal(resp.Result, &hexNonce); err != nil {
		return 0, fmt.Errorf("parse nonce: %w", err)
	}

	nonce, err := hexutil.DecodeUint64(hexNonce)
	if err != nil {
		return 0, fmt.Errorf("decode nonce hex: %w", err)
	}

	return nonce, nil
}

// SendRawTransaction sends a signed transaction and returns the tx hash.
func (c *RPCClient) SendRawTransaction(ctx context.Context, signedTxHex string) (string, string, error) {
	resp, err := c.Call(ctx, "probe_sendRawTransaction", signedTxHex)
	if err != nil {
		if resp != nil {
			return "", resp.RawBody, err
		}
		return "", "", err
	}

	var txHash string
	if err := json.Unmarshal(resp.Result, &txHash); err != nil {
		return "", resp.RawBody, fmt.Errorf("parse tx hash: %w", err)
	}

	return txHash, resp.RawBody, nil
}

// GetTransactionReceipt gets the receipt for a tx hash. Returns nil if pending.
type TxReceipt struct {
	TxHash            string `json:"transactionHash"`
	BlockNumber       string `json:"blockNumber"`
	BlockHash         string `json:"blockHash"`
	Status            string `json:"status"` // "0x1" success, "0x0" failure
	GasUsed           string `json:"gasUsed"`
	ContractAddress   string `json:"contractAddress,omitempty"`
	Logs              []TxLog `json:"logs"`
	TransactionIndex  string `json:"transactionIndex"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
}

type TxLog struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
}

func (c *RPCClient) GetTransactionReceipt(ctx context.Context, txHash string) (*TxReceipt, string, error) {
	resp, err := c.Call(ctx, "probe_getTransactionReceipt", txHash)
	if err != nil {
		return nil, "", err
	}

	// null result means tx is still pending
	if string(resp.Result) == "null" {
		return nil, resp.RawBody, nil
	}

	var receipt TxReceipt
	if err := json.Unmarshal(resp.Result, &receipt); err != nil {
		return nil, resp.RawBody, fmt.Errorf("parse receipt: %w", err)
	}

	return &receipt, resp.RawBody, nil
}

// WaitForReceipt polls for a transaction receipt with timeout.
func (c *RPCClient) WaitForReceipt(ctx context.Context, txHash string, timeout time.Duration) (*TxReceipt, string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		receipt, raw, err := c.GetTransactionReceipt(ctx, txHash)
		if err != nil {
			return nil, raw, err
		}
		if receipt != nil {
			return receipt, raw, nil
		}
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return nil, "", fmt.Errorf("timeout waiting for receipt of %s after %s", txHash, timeout)
}

// GetBlockNumber returns the current block number.
func (c *RPCClient) GetBlockNumber(ctx context.Context) (uint64, error) {
	resp, err := c.Call(ctx, "probe_blockNumber")
	if err != nil {
		return 0, err
	}

	var hexBlock string
	if err := json.Unmarshal(resp.Result, &hexBlock); err != nil {
		return 0, fmt.Errorf("parse block number: %w", err)
	}

	return hexutil.DecodeUint64(hexBlock)
}

// CallContract executes a read-only contract call (probe_call).
func (c *RPCClient) CallContract(ctx context.Context, to string, data []byte, block string) ([]byte, string, error) {
	if block == "" {
		block = "latest"
	}

	callObj := map[string]string{
		"to":   to,
		"data": hexutil.Encode(data),
	}

	resp, err := c.Call(ctx, "probe_call", callObj, block)
	if err != nil {
		if resp != nil {
			return nil, resp.RawBody, err
		}
		return nil, "", err
	}

	var hexResult string
	if err := json.Unmarshal(resp.Result, &hexResult); err != nil {
		return nil, resp.RawBody, fmt.Errorf("parse call result: %w", err)
	}

	result, err := hexutil.Decode(hexResult)
	if err != nil {
		return nil, resp.RawBody, fmt.Errorf("decode call result: %w", err)
	}

	return result, resp.RawBody, nil
}

// EstimateGas estimates gas for a transaction.
func (c *RPCClient) EstimateGas(ctx context.Context, from, to string, data []byte, value *big.Int) (uint64, error) {
	callObj := map[string]string{
		"from": from,
		"to":   to,
	}
	if data != nil {
		callObj["data"] = hexutil.Encode(data)
	}
	if value != nil && value.Sign() > 0 {
		callObj["value"] = hexutil.EncodeBig(value)
	}

	resp, err := c.Call(ctx, "probe_estimateGas", callObj)
	if err != nil {
		return 0, err
	}

	var hexGas string
	if err := json.Unmarshal(resp.Result, &hexGas); err != nil {
		return 0, fmt.Errorf("parse gas estimate: %w", err)
	}

	return hexutil.DecodeUint64(hexGas)
}

// SignAndSendTx signs a transaction with the given private key and sends it.
// Returns txHash, raw RPC response, and error.
func (c *RPCClient) SignAndSendTx(ctx context.Context, privateKeyHex string, tx *types.Transaction, chainID *big.Int) (string, string, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return "", "", fmt.Errorf("parse private key: %w", err)
	}

	signer := types.NewEIP155Signer(chainID)
	signedTx, err := types.SignTx(tx, signer, key)
	if err != nil {
		return "", "", fmt.Errorf("sign tx: %w", err)
	}

	var buf bytes.Buffer
	if err := rlp.Encode(&buf, signedTx); err != nil {
		return "", "", fmt.Errorf("rlp encode: %w", err)
	}

	rawHex := hexutil.Encode(buf.Bytes())
	return c.SendRawTransaction(ctx, rawHex)
}

// ReceiptToEvidence converts a TxReceipt to Evidence for test reporting.
func ReceiptToEvidence(receipt *TxReceipt, rawResponse string) *Evidence {
	if receipt == nil {
		return &Evidence{RawResponse: rawResponse}
	}

	blockNum, _ := hexutil.DecodeUint64(receipt.BlockNumber)
	gasUsed, _ := hexutil.DecodeUint64(receipt.GasUsed)
	status, _ := hexutil.DecodeUint64(receipt.Status)

	ev := &Evidence{
		TxHash:      receipt.TxHash,
		BlockNumber: blockNum,
		GasUsed:     gasUsed,
		Status:      status,
		RawResponse: rawResponse,
	}

	for _, log := range receipt.Logs {
		ev.Logs = append(ev.Logs, LogEvidence{
			Address: log.Address,
			Topics:  log.Topics,
			Data:    log.Data,
		})
	}

	return ev
}

// RPCPool manages multiple RPC clients for load distribution.
type RPCPool struct {
	clients []*RPCClient
	mu      sync.Mutex
	idx     uint64
}

// NewRPCPool creates a pool from multiple endpoints.
func NewRPCPool(urls []string) *RPCPool {
	pool := &RPCPool{}
	for _, url := range urls {
		pool.clients = append(pool.clients, NewRPCClient(url))
	}
	return pool
}

// Get returns a client using round-robin.
func (p *RPCPool) Get() *RPCClient {
	idx := atomic.AddUint64(&p.idx, 1)
	return p.clients[idx%uint64(len(p.clients))]
}

// GetByIndex returns the client for a specific node.
func (p *RPCPool) GetByIndex(i int) *RPCClient {
	return p.clients[i%len(p.clients)]
}

// AddressFromPrivateKey derives the address from a hex private key.
func AddressFromPrivateKey(privateKeyHex string) (common.Address, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(key.PublicKey), nil
}
