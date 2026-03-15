package runner

import (
	"fmt"
	"sync"
	"time"
)

// Orchestrator runs all test scenarios and collects results.
type Orchestrator struct {
	scenarios []Scenario
	results   []TestResult
	mu        sync.Mutex
}

// NewOrchestrator creates an orchestrator.
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{}
}

// Register adds a scenario to be executed.
func (o *Orchestrator) Register(s Scenario) {
	o.scenarios = append(o.scenarios, s)
}

// RunAll executes all 200 test cases sequentially across all scenarios.
// Returns the full result set.
func (o *Orchestrator) RunAll(ctx *TestContext) []TestResult {
	var allResults []TestResult

	for _, s := range o.scenarios {
		fmt.Printf("\n══════════════════════════════════════════════════════════\n")
		fmt.Printf("  场景: %s (%s)\n", s.Name(), s.Category())
		fmt.Printf("══════════════════════════════════════════════════════════\n")

		// Setup
		fmt.Printf("  Setting up...\n")
		if err := s.Setup(ctx); err != nil {
			fmt.Printf("  ✗ Setup failed: %v\n", err)
			// Record a FAIL result for the setup
			allResults = append(allResults, TestResult{
				Category: s.Category(),
				Name:     s.Name() + " — Setup",
				Status:   StatusFail,
				Error:    err.Error(),
			})
			continue
		}
		fmt.Printf("  ✓ Setup complete\n\n")

		// Run tests
		results := s.RunTests(ctx)
		for i := range results {
			// Compute final status from verifications
			results[i].Status = results[i].FinalStatus()
			allResults = append(allResults, results[i])

			// Print each result
			fmt.Printf("  %s\n", results[i].String())

			// Print verification details for non-PASS
			if results[i].Status != StatusPass {
				for _, v := range results[i].Verifications {
					if v.Status != StatusPass {
						fmt.Printf("    角度: %s\n", v.Angle)
						fmt.Printf("    期望: %s\n", v.Expected)
						fmt.Printf("    实际: %s\n", v.Actual)
						if v.Error != "" {
							fmt.Printf("    错误: %s\n", v.Error)
						}
					}
				}
			}
		}
	}

	return allResults
}

// RunStressAll executes stress tests across all StressScenarios.
func (o *Orchestrator) RunStressAll(ctx *TestContext, rounds []StressRound) []StressMetrics {
	var allMetrics []StressMetrics

	for _, round := range rounds {
		fmt.Printf("\n══════════════════════════════════════════════════════════\n")
		fmt.Printf("  压力测试 Round %d: %s (并发: %d, 持续: %s)\n",
			round.RoundNum, round.Scenario, round.Concurrency, round.Duration)
		fmt.Printf("══════════════════════════════════════════════════════════\n")

		// Find the matching scenario
		var stressScenario StressScenario
		for _, s := range o.scenarios {
			if ss, ok := s.(StressScenario); ok && s.Category() == round.Scenario {
				stressScenario = ss
				break
			}
		}

		if stressScenario == nil {
			fmt.Printf("  ⊘ Scenario %s not found or doesn't support stress testing\n", round.Scenario)
			continue
		}

		// Select accounts for this round
		accounts := selectAccounts(ctx.Accounts, round.Concurrency)

		// Execute stress test
		start := time.Now()
		metrics := stressScenario.RunStress(ctx, accounts, round.Duration)
		metrics.Round = round.RoundNum
		metrics.Scenario = round.Scenario
		metrics.Duration = time.Since(start)

		allMetrics = append(allMetrics, metrics)

		// Print metrics
		fmt.Printf("  TPS:     %.1f avg / %.1f peak\n", metrics.AvgTPS, metrics.PeakTPS)
		fmt.Printf("  延迟:    %.1f ms avg / %.1f ms P95 / %.1f ms P99\n",
			metrics.AvgLatencyMs, metrics.P95LatencyMs, metrics.P99LatencyMs)
		fmt.Printf("  成功率:  %.2f%% (%d/%d)\n",
			metrics.SuccessRate*100, metrics.TotalTxConfirmed, metrics.TotalTxSent)
		fmt.Printf("  出块:    %d blocks, %.1f ms avg interval\n",
			metrics.BlocksProduced, metrics.AvgBlockTimeMs)
	}

	return allMetrics
}

// StressRound defines parameters for one stress test round.
type StressRound struct {
	RoundNum    int
	Scenario    string
	Concurrency int
	Duration    time.Duration
}

// selectAccounts picks N accounts from the pool.
func selectAccounts(all []*Account, n int) []*Account {
	if n >= len(all) {
		return all
	}
	return all[:n]
}

// PrintSummary prints the final test report.
func PrintSummary(results []TestResult, metrics []StressMetrics) {
	fmt.Printf("\n")
	fmt.Printf("╔══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║           ProbeChain Rydberg 测试报告                    ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════╝\n")

	// Count by status
	counts := map[TestStatus]int{}
	categories := map[string]map[TestStatus]int{}
	for _, r := range results {
		counts[r.Status]++
		if _, ok := categories[r.Category]; !ok {
			categories[r.Category] = map[TestStatus]int{}
		}
		categories[r.Category][r.Status]++
	}

	total := len(results)
	pass := counts[StatusPass]
	fail := counts[StatusFail]
	unknown := counts[StatusUnknown]
	skip := counts[StatusSkip]

	fmt.Printf("\n  RESULTS: %d/%d passed, %d failed, %d unknown, %d skipped\n",
		pass, total, fail, unknown, skip)
	if total > 0 {
		fmt.Printf("  Pass Rate: %.1f%%\n", float64(pass)/float64(total)*100)
	}

	// Per-category breakdown
	fmt.Printf("\n  %-30s %8s %8s %8s %8s\n", "类别", "测试数", "通过", "失败", "未知")
	fmt.Printf("  %s\n", "────────────────────────────────────────────────────────")
	for cat, statuses := range categories {
		catTotal := statuses[StatusPass] + statuses[StatusFail] + statuses[StatusUnknown] + statuses[StatusSkip]
		fmt.Printf("  %-30s %8d %8d %8d %8d\n",
			cat, catTotal, statuses[StatusPass], statuses[StatusFail], statuses[StatusUnknown])
	}

	// Stress test summary
	if len(metrics) > 0 {
		fmt.Printf("\n  压力测试汇总:\n")
		fmt.Printf("  %-20s %10s %10s %10s %10s\n", "场景", "Avg TPS", "Peak TPS", "成功率", "P99延迟")
		for _, m := range metrics {
			fmt.Printf("  %-20s %10.1f %10.1f %9.1f%% %8.0fms\n",
				m.Scenario, m.AvgTPS, m.PeakTPS, m.SuccessRate*100, m.P99LatencyMs)
		}
	}

	// Print failed tests
	if fail > 0 {
		fmt.Printf("\n  失败测试详情:\n")
		for _, r := range results {
			if r.Status == StatusFail {
				fmt.Printf("  ✗ #%d %s: %s\n", r.ID, r.Name, r.Error)
				for _, v := range r.Verifications {
					if v.Status == StatusFail {
						fmt.Printf("    [%s] 期望: %s, 实际: %s\n", v.Angle, v.Expected, v.Actual)
						if v.Evidence != nil && v.Evidence.TxHash != "" {
							fmt.Printf("    tx: %s\n", v.Evidence.TxHash)
						}
					}
				}
			}
		}
	}

	// Print unknown tests
	if unknown > 0 {
		fmt.Printf("\n  未确定测试:\n")
		for _, r := range results {
			if r.Status == StatusUnknown {
				fmt.Printf("  ? #%d %s: %s\n", r.ID, r.Name, r.Error)
			}
		}
	}
}
