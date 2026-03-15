package runner

import (
	"fmt"
	"math/big"
	"time"
)

// TestStatus represents the outcome of a test — PASS, FAIL, or UNKNOWN.
// UNKNOWN is used when the test cannot determine the result (e.g. RPC unreachable).
type TestStatus string

const (
	StatusPass    TestStatus = "PASS"
	StatusFail    TestStatus = "FAIL"
	StatusUnknown TestStatus = "UNKNOWN"
	StatusSkip    TestStatus = "SKIP"
)

// Evidence holds raw on-chain data proving the test result.
// Every assertion MUST be backed by Evidence from the chain.
type Evidence struct {
	TxHash      string `json:"txHash,omitempty"`
	BlockNumber uint64 `json:"blockNumber,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	Value       string `json:"value,omitempty"`
	GasUsed     uint64 `json:"gasUsed,omitempty"`
	Status      uint64 `json:"receiptStatus,omitempty"` // 1=success, 0=revert
	RawResponse string `json:"rawResponse,omitempty"`   // raw RPC JSON for manual review
	Logs        []LogEvidence `json:"logs,omitempty"`
}

// LogEvidence stores decoded event log data.
type LogEvidence struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
}

// Verification is a single check within a test.
// Each test has multiple verifications (multi-angle).
type Verification struct {
	Angle    string     `json:"angle"`    // e.g. "sender_balance_decreased"
	Expected string     `json:"expected"` // what we expected
	Actual   string     `json:"actual"`   // what we got from chain
	Status   TestStatus `json:"status"`
	Evidence *Evidence  `json:"evidence,omitempty"`
	Error    string     `json:"error,omitempty"`
}

// TestResult is the complete result of a single test case.
type TestResult struct {
	ID            int            `json:"id"`
	Category      string         `json:"category"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Status        TestStatus     `json:"status"`
	Verifications []Verification `json:"verifications"`
	Duration      time.Duration  `json:"duration"`
	Timestamp     time.Time      `json:"timestamp"`
	Error         string         `json:"error,omitempty"`
}

// FinalStatus computes the overall status from all verifications.
// If ANY verification is FAIL → FAIL.
// If ANY verification is UNKNOWN and none FAIL → UNKNOWN.
// Only if ALL verifications PASS → PASS.
func (r *TestResult) FinalStatus() TestStatus {
	if len(r.Verifications) == 0 {
		return StatusUnknown
	}
	hasUnknown := false
	for _, v := range r.Verifications {
		if v.Status == StatusFail {
			return StatusFail
		}
		if v.Status == StatusUnknown {
			hasUnknown = true
		}
	}
	if hasUnknown {
		return StatusUnknown
	}
	return StatusPass
}

// StressMetrics holds metrics from a stress test round.
type StressMetrics struct {
	Round           int           `json:"round"`
	Scenario        string        `json:"scenario"`
	Duration        time.Duration `json:"duration"`
	TotalTxSent     uint64        `json:"totalTxSent"`
	TotalTxConfirmed uint64       `json:"totalTxConfirmed"`
	TotalTxFailed   uint64        `json:"totalTxFailed"`
	TotalTxPending  uint64        `json:"totalTxPending"`
	AvgTPS          float64       `json:"avgTPS"`
	PeakTPS         float64       `json:"peakTPS"`
	AvgLatencyMs    float64       `json:"avgLatencyMs"`
	P50LatencyMs    float64       `json:"p50LatencyMs"`
	P95LatencyMs    float64       `json:"p95LatencyMs"`
	P99LatencyMs    float64       `json:"p99LatencyMs"`
	SuccessRate     float64       `json:"successRate"`
	AvgGasUsed      uint64        `json:"avgGasUsed"`
	BlocksProduced  uint64        `json:"blocksProduced"`
	AvgBlockTimeMs  float64       `json:"avgBlockTimeMs"`
}

// Account holds a test account's key material.
type Account struct {
	PrivateKeyHex string
	Address       string
	Nonce         uint64
	Balance       *big.Int
	NodeIndex     int // which node (0-8) this account belongs to
}

// NodeConfig holds RPC endpoint info for one node.
type NodeConfig struct {
	Index   int    `json:"index"`
	Name    string `json:"name"`
	RPCURL  string `json:"rpcURL"`
	Owner   string `json:"owner"`
	Enode   string `json:"enode"`
}

// TestnetConfig holds the full testnet configuration.
type TestnetConfig struct {
	ChainID        int64        `json:"chainId"`
	Nodes          []NodeConfig `json:"nodes"`
	FunderKeys     []string     `json:"funderKeys"` // private keys of genesis accounts
	AccountsPerNode int         `json:"accountsPerNode"`
	Contracts      ContractAddresses `json:"contracts"`
}

// ContractAddresses stores deployed contract addresses.
type ContractAddresses struct {
	WPROBE            string   `json:"wprobe"`
	ERC20Tokens       []string `json:"erc20Tokens"`
	IdentityRegistry  string   `json:"identityRegistry"`
	KeyTrading        string   `json:"keyTrading"`
	ProSwapFactory    string   `json:"proswapFactory"`
	ProSwapRouter     string   `json:"proswapRouter"`
	MasterChef        string   `json:"masterChef"`
	PredictionMarket  string   `json:"predictionMarket"`
	ExchangeSettlement string  `json:"exchangeSettlement"`
}

// Scenario is the interface each test scenario must implement.
type Scenario interface {
	Name() string
	Category() string
	Setup(ctx *TestContext) error
	RunTests(ctx *TestContext) []TestResult
}

// StressScenario extends Scenario with stress testing capability.
type StressScenario interface {
	Scenario
	RunStress(ctx *TestContext, accounts []*Account, duration time.Duration) StressMetrics
}

// String formats TestResult for display.
func (r *TestResult) String() string {
	symbol := "?"
	switch r.Status {
	case StatusPass:
		symbol = "✓"
	case StatusFail:
		symbol = "✗"
	case StatusUnknown:
		symbol = "?"
	case StatusSkip:
		symbol = "⊘"
	}
	return fmt.Sprintf("[%s] #%d %s — %s (%s) [%d verifications]",
		symbol, r.ID, r.Category, r.Name, r.Duration.Round(time.Millisecond), len(r.Verifications))
}
