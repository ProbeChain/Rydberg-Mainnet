// ProbeChain Rydberg Testnet — 200 Test Cases + 10K Address Stress Test
//
// This program runs 200 functional tests across 10 categories and then
// executes stress tests with up to 10,000 concurrent addresses.
//
// Design principles (per user feedback):
//   1. Every test result MUST be backed by on-chain evidence (tx hash, block, raw RPC response)
//   2. Tests may report UNKNOWN/SKIP — never fabricate a PASS
//   3. Same function tested from MULTIPLE angles (balance before/after, receipt status, events)
//   4. All assertions compare against LIVE chain state, not in-memory expectations
//
// Usage:
//   go run . -config config/testnet.json -mode tests     # run 200 test cases only
//   go run . -config config/testnet.json -mode stress    # run stress tests only
//   go run . -config config/testnet.json -mode all       # run both
//   go run . -config config/testnet.json -mode preflight # just check node connectivity

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/probechain/go-probe/miner/stress/probechain-tests/accounts"
	"github.com/probechain/go-probe/miner/stress/probechain-tests/runner"
	"github.com/probechain/go-probe/miner/stress/probechain-tests/scenarios"
)

func main() {
	configPath := flag.String("config", "config/testnet.json", "path to testnet config")
	mode := flag.String("mode", "all", "test mode: preflight, tests, stress, all")
	accountsPerNode := flag.Int("accounts", 1000, "accounts per node (total = 9 * this)")
	flag.Parse()

	// Load config
	configData, err := ioutil.ReadFile(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read config: %v\n", err)
		os.Exit(1)
	}

	var config runner.TestnetConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse config: %v\n", err)
		os.Exit(1)
	}

	if *accountsPerNode > 0 {
		config.AccountsPerNode = *accountsPerNode
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n\n  ⚠ Interrupted — generating partial report...")
		cancel()
	}()

	// Create test context
	testCtx := runner.NewTestContext(ctx, &config)

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║  ProbeChain Rydberg 测试网 — 产品级测试框架              ║")
	fmt.Println("║  Chain ID: 8004 | PoB V2.1 | 9 Nodes                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Phase 0: Preflight — check all nodes are reachable
	fmt.Println("── Phase 0: 预检 (Preflight) ──────────────────────────────")
	if !preflight(testCtx) {
		fmt.Println("  ✗ 预检失败 — 无法连接到节点")
		if *mode == "preflight" {
			os.Exit(1)
		}
		fmt.Println("  ⚠ 继续运行但部分测试可能返回 UNKNOWN")
	}

	if *mode == "preflight" {
		return
	}

	// Phase 1: Generate accounts
	fmt.Println("\n── Phase 1: 生成测试账户 ──────────────────────────────────")
	allAccts, groupedAccts := accounts.GenerateAccounts(len(config.Nodes), config.AccountsPerNode)
	testCtx.Accounts = allAccts
	testCtx.NodeAccts = groupedAccts

	// Phase 2: Fund accounts
	fmt.Println("\n── Phase 2: 分配测试资金 ──────────────────────────────────")
	amountPerAccount := new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18)) // 100,000 PROBE each
	hasFunderKeys := false
	for _, k := range config.FunderKeys {
		if k != "" && k != "GENESIS_PRIVATE_KEY_NODE_0" {
			hasFunderKeys = true
			break
		}
	}

	if hasFunderKeys {
		funded, err := accounts.FundAccounts(testCtx, amountPerAccount, 50)
		if err != nil {
			fmt.Printf("  ⚠ Funding had errors: %v\n", err)
		}
		fmt.Printf("  资金分配完成: %d 笔\n", funded)
	} else {
		fmt.Println("  ⊘ 未配置 funder 私钥 — 跳过资金分配")
		fmt.Println("  提示: 在 config/testnet.json 中填入创世账户私钥")
	}

	// Phase 3: Run 200 test cases
	if *mode == "tests" || *mode == "all" {
		fmt.Println("\n── Phase 3: 200 测试用例 ──────────────────────────────────")
		orch := runner.NewOrchestrator()

		// Register all 10 scenarios
		orch.Register(scenarios.NewTransferScenario())
		orch.Register(scenarios.NewERC20Scenario())
		// TODO: Register remaining scenarios as they are implemented
		// orch.Register(scenarios.NewIdentityScenario())
		// orch.Register(scenarios.NewKeyTradingScenario())
		// orch.Register(scenarios.NewProSwapScenario())
		// orch.Register(scenarios.NewFarmingScenario())
		// orch.Register(scenarios.NewPredictionScenario())
		// orch.Register(scenarios.NewSettlementScenario())
		// orch.Register(scenarios.NewWalletScenario())
		// orch.Register(scenarios.NewPoBScenario())

		results := orch.RunAll(testCtx)

		// Phase 4: Stress tests
		var metrics []runner.StressMetrics
		if *mode == "all" {
			fmt.Println("\n── Phase 4: 万地址压力测试 ────────────────────────────────")
			rounds := []runner.StressRound{
				{RoundNum: 1, Scenario: "1. 基础转账", Concurrency: 1000, Duration: 5 * time.Minute},
				{RoundNum: 2, Scenario: "2. ERC20 代币", Concurrency: 2000, Duration: 5 * time.Minute},
				// Additional rounds to be added
			}
			metrics = orch.RunStressAll(testCtx, rounds)
		}

		// Final report
		runner.PrintSummary(results, metrics)

		// Save report to file
		saveReport(results, metrics)
	}

	if *mode == "stress" {
		fmt.Println("\n── 压力测试 (独立模式) ─────────────────────────────────────")
		orch := runner.NewOrchestrator()
		orch.Register(scenarios.NewTransferScenario())

		rounds := []runner.StressRound{
			{RoundNum: 1, Scenario: "1. 基础转账", Concurrency: 1000, Duration: 5 * time.Minute},
			{RoundNum: 2, Scenario: "1. 基础转账", Concurrency: 5000, Duration: 10 * time.Minute},
			{RoundNum: 3, Scenario: "1. 基础转账", Concurrency: 10000, Duration: 30 * time.Minute},
		}
		metrics := orch.RunStressAll(testCtx, rounds)
		runner.PrintSummary(nil, metrics)
	}
}

// preflight checks all 9 nodes are reachable and synced.
func preflight(ctx *runner.TestContext) bool {
	allOk := true
	for i, node := range ctx.Config.Nodes {
		client := ctx.Pool.GetByIndex(i)
		blockNum, err := client.GetBlockNumber(ctx.Ctx)
		if err != nil {
			fmt.Printf("  ✗ %s (%s): 无法连接 — %v\n", node.Name, node.RPCURL, err)
			allOk = false
			continue
		}
		fmt.Printf("  ✓ %s (%s): block #%d\n", node.Name, node.RPCURL, blockNum)
	}
	return allOk
}

// saveReport writes the test results to a JSON file.
func saveReport(results []runner.TestResult, metrics []runner.StressMetrics) {
	report := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"chainId":   8004,
		"tests":     results,
		"stress":    metrics,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("  ⚠ 无法序列化报告: %v\n", err)
		return
	}

	filename := fmt.Sprintf("report_%s.json", time.Now().Format("20060102_150405"))
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("  ⚠ 无法保存报告: %v\n", err)
		return
	}

	fmt.Printf("\n  报告已保存: %s\n", filename)
}
