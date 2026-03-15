package runner

import (
	"context"
	"math/big"
	"sync"
)

// TestContext provides shared state for all test scenarios.
type TestContext struct {
	Ctx       context.Context
	Config    *TestnetConfig
	Pool      *RPCPool
	ChainID   *big.Int
	Accounts  []*Account        // all 10000 test accounts
	NodeAccts [][]*Account       // accounts grouped by node (9 groups)
	Contracts *ContractAddresses // deployed contract addresses

	// Mutex for nonce management
	nonceMu sync.Mutex
	nonces  map[string]uint64
}

// NewTestContext creates a TestContext from config.
func NewTestContext(ctx context.Context, config *TestnetConfig) *TestContext {
	urls := make([]string, len(config.Nodes))
	for i, n := range config.Nodes {
		urls[i] = n.RPCURL
	}

	return &TestContext{
		Ctx:       ctx,
		Config:    config,
		Pool:      NewRPCPool(urls),
		ChainID:   big.NewInt(config.ChainID),
		Contracts: &config.Contracts,
		nonces:    make(map[string]uint64),
	}
}

// GetNonce returns the next nonce for an address, tracking locally.
// On first call, fetches from chain. Subsequent calls increment locally.
func (tc *TestContext) GetNonce(address string) (uint64, error) {
	tc.nonceMu.Lock()
	defer tc.nonceMu.Unlock()

	if nonce, ok := tc.nonces[address]; ok {
		tc.nonces[address] = nonce + 1
		return nonce, nil
	}

	// First call — fetch from chain
	client := tc.Pool.Get()
	nonce, err := client.GetTransactionCount(tc.Ctx, address, "pending")
	if err != nil {
		return 0, err
	}

	tc.nonces[address] = nonce + 1
	return nonce, nil
}

// ResetNonce forces a re-fetch of nonce from chain.
func (tc *TestContext) ResetNonce(address string) {
	tc.nonceMu.Lock()
	defer tc.nonceMu.Unlock()
	delete(tc.nonces, address)
}

// PrimaryClient returns the RPC client for node 0 (used for deployments).
func (tc *TestContext) PrimaryClient() *RPCClient {
	return tc.Pool.GetByIndex(0)
}

// ClientForNode returns the RPC client for a specific node.
func (tc *TestContext) ClientForNode(nodeIdx int) *RPCClient {
	return tc.Pool.GetByIndex(nodeIdx)
}
