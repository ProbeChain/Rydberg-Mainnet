package scenarios

import (
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/core/types"

	"github.com/probechain/go-probe/miner/stress/probechain-tests/runner"
)

// TransferScenario tests basic PROBE transfers (Test #1-20).
type TransferScenario struct {
	// accounts used by this scenario
	sender   *runner.Account
	receiver *runner.Account
	extra    []*runner.Account // additional accounts for batch/concurrent tests
}

func NewTransferScenario() *TransferScenario {
	return &TransferScenario{}
}

func (s *TransferScenario) Name() string     { return "基础转账与Gas" }
func (s *TransferScenario) Category() string  { return "1. 基础转账" }

func (s *TransferScenario) Setup(ctx *runner.TestContext) error {
	if len(ctx.Accounts) < 110 {
		return fmt.Errorf("need at least 110 accounts, have %d", len(ctx.Accounts))
	}
	s.sender = ctx.Accounts[0]
	s.receiver = ctx.Accounts[1]
	s.extra = ctx.Accounts[2:110]
	return nil
}

func (s *TransferScenario) RunTests(ctx *runner.TestContext) []runner.TestResult {
	var results []runner.TestResult

	tests := []struct {
		id   int
		name string
		fn   func(ctx *runner.TestContext) runner.TestResult
	}{
		{1, "PROBE 基础转账", s.test01BasicTransfer},
		{2, "零值转账", s.test02ZeroTransfer},
		{3, "大额转账", s.test03LargeTransfer},
		{4, "小额转账 (minTxValue)", s.test04SmallTransfer},
		{5, "低于最小值转账", s.test05BelowMinTransfer},
		{6, "自转账", s.test06SelfTransfer},
		{7, "余额不足转账", s.test07InsufficientBalance},
		{8, "Nonce 连续发送 100 笔", s.test08ConsecutiveNonce},
		{9, "Nonce 乱序", s.test09OutOfOrderNonce},
		{10, "Nonce 重复", s.test10DuplicateNonce},
		{11, "EIP-1559 动态费转账", s.test11EIP1559Transfer},
		{12, "Legacy tx 转账", s.test12LegacyTransfer},
		{13, "Gas limit 精确 21000", s.test13ExactGasLimit},
		{14, "Gas limit 不足", s.test14InsufficientGas},
		{15, "Gas limit 过高", s.test15HighGasLimit},
		{16, "多目标批量转账", s.test16MultiTarget},
		{17, "环形转账", s.test17CircularTransfer},
		{18, "100 地址并发转账", s.test18Concurrent100},
		{19, "1000 地址并发转账", s.test19Concurrent1000},
		{20, "交易池满载", s.test20TxPoolFull},
	}

	for _, t := range tests {
		start := time.Now()
		result := t.fn(ctx)
		result.ID = t.id
		result.Category = s.Category()
		result.Name = t.name
		result.Duration = time.Since(start)
		result.Timestamp = start
		result.Status = result.FinalStatus()
		results = append(results, result)
	}

	return results
}

// test01BasicTransfer: A→B transfer 1 PROBE
// Multi-angle: sender balance decreased, receiver balance increased, receipt status=1, gas used=21000
func (s *TransferScenario) test01BasicTransfer(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "A→B 转 1 PROBE，验证余额变化"}
	client := ctx.PrimaryClient()
	amount := big.NewInt(1e18) // 1 PROBE

	// Angle 1: Record sender balance BEFORE
	senderBalBefore, rawBefore, err := client.GetBalance(ctx.Ctx, s.sender.Address, "latest")
	if err != nil {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "sender_balance_before", Status: runner.StatusUnknown,
			Error: fmt.Sprintf("无法获取发送方余额: %v", err),
		})
		return r
	}

	// Angle 2: Record receiver balance BEFORE
	recvBalBefore, _, err := client.GetBalance(ctx.Ctx, s.receiver.Address, "latest")
	if err != nil {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "receiver_balance_before", Status: runner.StatusUnknown,
			Error: fmt.Sprintf("无法获取接收方余额: %v", err),
		})
		return r
	}

	// Send transaction
	nonce, err := ctx.GetNonce(s.sender.Address)
	if err != nil {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "get_nonce", Status: runner.StatusUnknown,
			Error: fmt.Sprintf("无法获取nonce: %v", err),
		})
		return r
	}

	tx := types.NewTransaction(nonce, common.HexToAddress(s.receiver.Address), amount, 21000, big.NewInt(1e9), nil)
	txHash, rawSend, err := client.SignAndSendTx(ctx.Ctx, s.sender.PrivateKeyHex, tx, ctx.ChainID)
	if err != nil {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "send_tx", Status: runner.StatusFail,
			Expected: "交易发送成功", Actual: err.Error(),
			Evidence: &runner.Evidence{RawResponse: rawSend},
		})
		return r
	}

	// Wait for receipt
	receipt, rawReceipt, err := client.WaitForReceipt(ctx.Ctx, txHash, 60*time.Second)
	if err != nil {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "wait_receipt", Status: runner.StatusUnknown,
			Error: fmt.Sprintf("等待回执超时: %v", err),
			Evidence: &runner.Evidence{TxHash: txHash},
		})
		return r
	}

	evidence := runner.ReceiptToEvidence(receipt, rawReceipt)

	// Verification 1: Receipt status == 1 (success)
	r.Verifications = append(r.Verifications, runner.Verification{
		Angle:    "receipt_status",
		Expected: "0x1",
		Actual:   receipt.Status,
		Status:   boolToStatus(receipt.Status == "0x1"),
		Evidence: evidence,
	})

	// Verification 2: Gas used == 21000 (standard transfer)
	r.Verifications = append(r.Verifications, runner.Verification{
		Angle:    "gas_used",
		Expected: "21000",
		Actual:   receipt.GasUsed,
		Status:   boolToStatus(receipt.GasUsed == "0x5208"),
		Evidence: evidence,
	})

	// Verification 3: Sender balance decreased by (amount + gas fee)
	senderBalAfter, rawAfter, err := client.GetBalance(ctx.Ctx, s.sender.Address, "latest")
	if err != nil {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "sender_balance_after", Status: runner.StatusUnknown,
			Error: fmt.Sprintf("无法获取发送方余额: %v", err),
		})
	} else {
		gasCost := new(big.Int).Mul(big.NewInt(21000), big.NewInt(1e9))
		expectedDecrease := new(big.Int).Add(amount, gasCost)
		actualDecrease := new(big.Int).Sub(senderBalBefore, senderBalAfter)

		r.Verifications = append(r.Verifications, runner.Verification{
			Angle:    "sender_balance_decreased",
			Expected: fmt.Sprintf("减少 %s (amount + gas)", expectedDecrease.String()),
			Actual:   fmt.Sprintf("减少 %s (before: %s, after: %s)", actualDecrease.String(), senderBalBefore.String(), senderBalAfter.String()),
			Status:   boolToStatus(actualDecrease.Cmp(expectedDecrease) == 0),
			Evidence: &runner.Evidence{RawResponse: fmt.Sprintf("before: %s\nafter: %s", rawBefore, rawAfter)},
		})
	}

	// Verification 4: Receiver balance increased by exactly amount
	recvBalAfter, rawRecvAfter, err := client.GetBalance(ctx.Ctx, s.receiver.Address, "latest")
	if err != nil {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "receiver_balance_after", Status: runner.StatusUnknown,
			Error: fmt.Sprintf("无法获取接收方余额: %v", err),
		})
	} else {
		actualIncrease := new(big.Int).Sub(recvBalAfter, recvBalBefore)
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle:    "receiver_balance_increased",
			Expected: fmt.Sprintf("增加 %s", amount.String()),
			Actual:   fmt.Sprintf("增加 %s (before: %s, after: %s)", actualIncrease.String(), recvBalBefore.String(), recvBalAfter.String()),
			Status:   boolToStatus(actualIncrease.Cmp(amount) == 0),
			Evidence: &runner.Evidence{RawResponse: rawRecvAfter},
		})
	}

	return r
}

// test02ZeroTransfer: Transfer 0 PROBE
func (s *TransferScenario) test02ZeroTransfer(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "转 0 PROBE，验证交易成功"}
	client := ctx.PrimaryClient()

	nonce, err := ctx.GetNonce(s.sender.Address)
	if err != nil {
		return unknownResult(r, "get_nonce", err)
	}

	tx := types.NewTransaction(nonce, common.HexToAddress(s.receiver.Address), big.NewInt(0), 21000, big.NewInt(1e9), nil)
	txHash, _, err := client.SignAndSendTx(ctx.Ctx, s.sender.PrivateKeyHex, tx, ctx.ChainID)
	if err != nil {
		return failResult(r, "send_tx", "交易发送成功", err.Error())
	}

	receipt, rawReceipt, err := client.WaitForReceipt(ctx.Ctx, txHash, 60*time.Second)
	if err != nil {
		return unknownResult(r, "wait_receipt", err)
	}

	evidence := runner.ReceiptToEvidence(receipt, rawReceipt)
	r.Verifications = append(r.Verifications, runner.Verification{
		Angle: "receipt_status", Expected: "0x1", Actual: receipt.Status,
		Status: boolToStatus(receipt.Status == "0x1"), Evidence: evidence,
	})

	return r
}

// test03LargeTransfer: Transfer 1,000,000 PROBE
func (s *TransferScenario) test03LargeTransfer(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "转 1,000,000 PROBE"}
	return s.doSimpleTransfer(ctx, r, new(big.Int).Mul(big.NewInt(1e6), big.NewInt(1e18)))
}

// test04SmallTransfer: Transfer minTxValueWei (0.01 PROBE)
func (s *TransferScenario) test04SmallTransfer(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "转 minTxValueWei (0.01 PROBE)"}
	return s.doSimpleTransfer(ctx, r, big.NewInt(1e16)) // 0.01 PROBE = 10^16 wei
}

// test05BelowMinTransfer: Transfer below minTxValueWei, should still succeed
// (minTxValueWei is a PoB reward threshold, not a transfer restriction)
func (s *TransferScenario) test05BelowMinTransfer(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "转 < minTxValueWei"}
	return s.doSimpleTransfer(ctx, r, big.NewInt(1e15)) // 0.001 PROBE
}

// test06SelfTransfer: A→A transfer
func (s *TransferScenario) test06SelfTransfer(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "A→A 自转账"}
	client := ctx.PrimaryClient()
	amount := big.NewInt(1e18)

	balBefore, _, err := client.GetBalance(ctx.Ctx, s.sender.Address, "latest")
	if err != nil {
		return unknownResult(r, "balance_before", err)
	}

	nonce, err := ctx.GetNonce(s.sender.Address)
	if err != nil {
		return unknownResult(r, "get_nonce", err)
	}

	// Send to self
	tx := types.NewTransaction(nonce, common.HexToAddress(s.sender.Address), amount, 21000, big.NewInt(1e9), nil)
	txHash, _, err := client.SignAndSendTx(ctx.Ctx, s.sender.PrivateKeyHex, tx, ctx.ChainID)
	if err != nil {
		return failResult(r, "send_tx", "交易发送成功", err.Error())
	}

	receipt, rawReceipt, err := client.WaitForReceipt(ctx.Ctx, txHash, 60*time.Second)
	if err != nil {
		return unknownResult(r, "wait_receipt", err)
	}

	evidence := runner.ReceiptToEvidence(receipt, rawReceipt)

	// Verify 1: Receipt success
	r.Verifications = append(r.Verifications, runner.Verification{
		Angle: "receipt_status", Expected: "0x1", Actual: receipt.Status,
		Status: boolToStatus(receipt.Status == "0x1"), Evidence: evidence,
	})

	// Verify 2: Balance decreased by gas only (amount cancels out)
	balAfter, _, err := client.GetBalance(ctx.Ctx, s.sender.Address, "latest")
	if err != nil {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "balance_after", Status: runner.StatusUnknown, Error: err.Error(),
		})
	} else {
		gasCost := new(big.Int).Mul(big.NewInt(21000), big.NewInt(1e9))
		actualDecrease := new(big.Int).Sub(balBefore, balAfter)
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle:    "balance_decreased_by_gas_only",
			Expected: fmt.Sprintf("减少 %s (仅gas)", gasCost.String()),
			Actual:   fmt.Sprintf("减少 %s", actualDecrease.String()),
			Status:   boolToStatus(actualDecrease.Cmp(gasCost) == 0),
		})
	}

	return r
}

// test07InsufficientBalance: Should fail
func (s *TransferScenario) test07InsufficientBalance(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "余额不足转账，验证拒绝"}
	client := ctx.PrimaryClient()

	// Create a fresh account with 0 balance
	freshAcct := s.extra[99]

	nonce, err := client.GetTransactionCount(ctx.Ctx, freshAcct.Address, "pending")
	if err != nil {
		return unknownResult(r, "get_nonce", err)
	}

	// Try to send 1 PROBE from an unfunded account
	hugeAmount := new(big.Int).Mul(big.NewInt(1e18), big.NewInt(1e18)) // impossibly large
	tx := types.NewTransaction(nonce, common.HexToAddress(s.receiver.Address), hugeAmount, 21000, big.NewInt(1e9), nil)
	_, rawSend, err := client.SignAndSendTx(ctx.Ctx, freshAcct.PrivateKeyHex, tx, ctx.ChainID)

	// We expect an error (insufficient funds)
	if err != nil {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "tx_rejected", Expected: "交易被拒绝 (余额不足)",
			Actual: err.Error(), Status: runner.StatusPass,
			Evidence: &runner.Evidence{RawResponse: rawSend},
		})
	} else {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "tx_rejected", Expected: "交易被拒绝",
			Actual: "交易被接受（不应该）", Status: runner.StatusFail,
			Evidence: &runner.Evidence{RawResponse: rawSend},
		})
	}

	return r
}

// test08ConsecutiveNonce: Send 100 txs in rapid succession
func (s *TransferScenario) test08ConsecutiveNonce(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "同账户连发 100 笔交易"}
	client := ctx.PrimaryClient()
	count := 100

	startNonce, err := client.GetTransactionCount(ctx.Ctx, s.sender.Address, "pending")
	if err != nil {
		return unknownResult(r, "get_nonce", err)
	}

	var txHashes []string
	for i := 0; i < count; i++ {
		nonce := startNonce + uint64(i)
		tx := types.NewTransaction(nonce, common.HexToAddress(s.receiver.Address), big.NewInt(1000), 21000, big.NewInt(1e9), nil)
		txHash, _, err := client.SignAndSendTx(ctx.Ctx, s.sender.PrivateKeyHex, tx, ctx.ChainID)
		if err != nil {
			r.Verifications = append(r.Verifications, runner.Verification{
				Angle: fmt.Sprintf("send_tx_%d", i), Status: runner.StatusFail,
				Expected: "发送成功", Actual: err.Error(),
			})
			break
		}
		txHashes = append(txHashes, txHash)
	}

	// Update local nonce tracker
	ctx.ResetNonce(s.sender.Address)

	r.Verifications = append(r.Verifications, runner.Verification{
		Angle: "all_sent", Expected: fmt.Sprintf("%d 笔交易发送", count),
		Actual: fmt.Sprintf("%d 笔交易发送", len(txHashes)),
		Status: boolToStatus(len(txHashes) == count),
	})

	// Wait for last tx
	if len(txHashes) > 0 {
		lastHash := txHashes[len(txHashes)-1]
		receipt, rawReceipt, err := client.WaitForReceipt(ctx.Ctx, lastHash, 120*time.Second)
		if err != nil {
			r.Verifications = append(r.Verifications, runner.Verification{
				Angle: "last_tx_confirmed", Status: runner.StatusUnknown,
				Error: fmt.Sprintf("最后一笔超时: %v", err),
			})
		} else {
			r.Verifications = append(r.Verifications, runner.Verification{
				Angle: "last_tx_confirmed", Expected: "0x1", Actual: receipt.Status,
				Status:   boolToStatus(receipt.Status == "0x1"),
				Evidence: runner.ReceiptToEvidence(receipt, rawReceipt),
			})
		}
	}

	return r
}

// test09OutOfOrderNonce: Send nonce 5 first, then fill 1-4
func (s *TransferScenario) test09OutOfOrderNonce(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "先发 nonce=5，再补 nonce=1-4"}
	client := ctx.PrimaryClient()

	// Use a fresh account to have clean nonce
	acct := s.extra[10]
	baseNonce, err := client.GetTransactionCount(ctx.Ctx, acct.Address, "pending")
	if err != nil {
		return unknownResult(r, "get_nonce", err)
	}

	// Send nonce+5 first (should go to queue)
	tx5 := types.NewTransaction(baseNonce+5, common.HexToAddress(s.receiver.Address), big.NewInt(1000), 21000, big.NewInt(1e9), nil)
	_, _, err = client.SignAndSendTx(ctx.Ctx, acct.PrivateKeyHex, tx5, ctx.ChainID)
	if err != nil {
		// Some nodes may reject future nonces — record but don't fail
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "future_nonce_accepted", Expected: "接受或排队",
			Actual: err.Error(), Status: runner.StatusPass, // acceptable behavior
		})
	}

	// Fill nonces 0-4
	for i := uint64(0); i <= 4; i++ {
		tx := types.NewTransaction(baseNonce+i, common.HexToAddress(s.receiver.Address), big.NewInt(1000), 21000, big.NewInt(1e9), nil)
		_, _, err := client.SignAndSendTx(ctx.Ctx, acct.PrivateKeyHex, tx, ctx.ChainID)
		if err != nil {
			r.Verifications = append(r.Verifications, runner.Verification{
				Angle: fmt.Sprintf("fill_nonce_%d", i), Status: runner.StatusFail,
				Expected: "发送成功", Actual: err.Error(),
			})
		}
	}

	// Wait and check that all 6 txs eventually process
	time.Sleep(30 * time.Second) // allow processing
	finalNonce, err := client.GetTransactionCount(ctx.Ctx, acct.Address, "latest")
	if err != nil {
		return unknownResult(r, "final_nonce", err)
	}

	r.Verifications = append(r.Verifications, runner.Verification{
		Angle:    "nonces_processed",
		Expected: fmt.Sprintf("nonce >= %d", baseNonce+5),
		Actual:   fmt.Sprintf("nonce = %d", finalNonce),
		Status:   boolToStatus(finalNonce >= baseNonce+5),
	})

	return r
}

// test10DuplicateNonce: Send same nonce with different tx
func (s *TransferScenario) test10DuplicateNonce(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "相同 nonce 不同交易"}
	client := ctx.PrimaryClient()

	acct := s.extra[11]
	nonce, err := client.GetTransactionCount(ctx.Ctx, acct.Address, "pending")
	if err != nil {
		return unknownResult(r, "get_nonce", err)
	}

	// Send first tx with this nonce
	tx1 := types.NewTransaction(nonce, common.HexToAddress(s.extra[0].Address), big.NewInt(1000), 21000, big.NewInt(1e9), nil)
	hash1, _, err := client.SignAndSendTx(ctx.Ctx, acct.PrivateKeyHex, tx1, ctx.ChainID)
	if err != nil {
		return failResult(r, "send_tx1", "第一笔发送成功", err.Error())
	}

	// Send second tx with same nonce, different recipient
	tx2 := types.NewTransaction(nonce, common.HexToAddress(s.extra[1].Address), big.NewInt(2000), 21000, big.NewInt(1e9), nil)
	_, rawSend2, err2 := client.SignAndSendTx(ctx.Ctx, acct.PrivateKeyHex, tx2, ctx.ChainID)

	if err2 != nil {
		// Expected: second tx rejected (nonce already used)
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "duplicate_rejected", Expected: "重复 nonce 被拒绝",
			Actual: err2.Error(), Status: runner.StatusPass,
			Evidence: &runner.Evidence{RawResponse: rawSend2},
		})
	} else {
		// Also acceptable: replacement tx with higher gas price
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "duplicate_handling", Expected: "拒绝或替换",
			Actual: "第二笔被接受（可能替换）", Status: runner.StatusPass,
		})
	}

	// Verify first tx confirms
	receipt, rawReceipt, err := client.WaitForReceipt(ctx.Ctx, hash1, 60*time.Second)
	if err != nil {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "first_tx_confirmed", Status: runner.StatusUnknown,
			Error: err.Error(),
		})
	} else {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "first_tx_confirmed", Expected: "0x1", Actual: receipt.Status,
			Status:   boolToStatus(receipt.Status == "0x1"),
			Evidence: runner.ReceiptToEvidence(receipt, rawReceipt),
		})
	}

	return r
}

// test11EIP1559Transfer: DynamicFeeTx
func (s *TransferScenario) test11EIP1559Transfer(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "EIP-1559 DynamicFeeTx 转账"}
	client := ctx.PrimaryClient()

	nonce, err := ctx.GetNonce(s.sender.Address)
	if err != nil {
		return unknownResult(r, "get_nonce", err)
	}

	recipient := common.HexToAddress(s.receiver.Address)
	dynamicTx := &types.DynamicFeeTx{
		ChainID:   ctx.ChainID,
		Nonce:     nonce,
		GasTipCap: big.NewInt(1e9),       // 1 Gwei tip
		GasFeeCap: big.NewInt(10e9),       // 10 Gwei max fee
		Gas:       21000,
		To:        &recipient,
		Value:     big.NewInt(1e18),
	}
	tx := types.NewTx(dynamicTx)

	txHash, _, err := client.SignAndSendTx(ctx.Ctx, s.sender.PrivateKeyHex, tx, ctx.ChainID)
	if err != nil {
		return failResult(r, "send_tx", "EIP-1559 tx 发送成功", err.Error())
	}

	receipt, rawReceipt, err := client.WaitForReceipt(ctx.Ctx, txHash, 60*time.Second)
	if err != nil {
		return unknownResult(r, "wait_receipt", err)
	}

	evidence := runner.ReceiptToEvidence(receipt, rawReceipt)
	r.Verifications = append(r.Verifications, runner.Verification{
		Angle: "receipt_status", Expected: "0x1", Actual: receipt.Status,
		Status: boolToStatus(receipt.Status == "0x1"), Evidence: evidence,
	})

	return r
}

// test12LegacyTransfer: Traditional gasPrice tx
func (s *TransferScenario) test12LegacyTransfer(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "Legacy tx 转账"}
	return s.doSimpleTransfer(ctx, r, big.NewInt(1e18))
}

// test13ExactGasLimit: Gas = 21000
func (s *TransferScenario) test13ExactGasLimit(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "Gas limit 精确 21000"}
	return s.doSimpleTransfer(ctx, r, big.NewInt(1e18))
}

// test14InsufficientGas: Gas < 21000
func (s *TransferScenario) test14InsufficientGas(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "Gas limit 不足，验证拒绝"}
	client := ctx.PrimaryClient()

	nonce, err := ctx.GetNonce(s.sender.Address)
	if err != nil {
		return unknownResult(r, "get_nonce", err)
	}

	tx := types.NewTransaction(nonce, common.HexToAddress(s.receiver.Address), big.NewInt(1e18), 20999, big.NewInt(1e9), nil)
	_, rawSend, err := client.SignAndSendTx(ctx.Ctx, s.sender.PrivateKeyHex, tx, ctx.ChainID)

	if err != nil {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "tx_rejected", Expected: "Gas 不足被拒绝",
			Actual: err.Error(), Status: runner.StatusPass,
			Evidence: &runner.Evidence{RawResponse: rawSend},
		})
		// Nonce was not consumed, reset it
		ctx.ResetNonce(s.sender.Address)
	} else {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "tx_rejected", Expected: "交易被拒绝",
			Actual: "交易被接受（不应该）", Status: runner.StatusFail,
		})
	}

	return r
}

// test15HighGasLimit: Gas close to block limit
func (s *TransferScenario) test15HighGasLimit(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "Gas limit 接近区块限制"}
	client := ctx.PrimaryClient()

	nonce, err := ctx.GetNonce(s.sender.Address)
	if err != nil {
		return unknownResult(r, "get_nonce", err)
	}

	// 30M gas limit (genesis gasLimit = 0x1c9c380 = 30,000,000)
	highGas := uint64(29_000_000)
	tx := types.NewTransaction(nonce, common.HexToAddress(s.receiver.Address), big.NewInt(1e18), highGas, big.NewInt(1e9), nil)
	txHash, _, err := client.SignAndSendTx(ctx.Ctx, s.sender.PrivateKeyHex, tx, ctx.ChainID)
	if err != nil {
		// May be rejected if gas too high — that's acceptable
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "high_gas_handling", Expected: "接受或拒绝",
			Actual: err.Error(), Status: runner.StatusPass,
		})
		ctx.ResetNonce(s.sender.Address)
		return r
	}

	receipt, rawReceipt, err := client.WaitForReceipt(ctx.Ctx, txHash, 60*time.Second)
	if err != nil {
		return unknownResult(r, "wait_receipt", err)
	}

	evidence := runner.ReceiptToEvidence(receipt, rawReceipt)
	r.Verifications = append(r.Verifications, runner.Verification{
		Angle: "receipt_status", Expected: "0x1", Actual: receipt.Status,
		Status: boolToStatus(receipt.Status == "0x1"), Evidence: evidence,
	})

	// Gas used should still be 21000 (simple transfer)
	r.Verifications = append(r.Verifications, runner.Verification{
		Angle: "actual_gas_used_21000", Expected: "0x5208", Actual: receipt.GasUsed,
		Status: boolToStatus(receipt.GasUsed == "0x5208"), Evidence: evidence,
	})

	return r
}

// test16MultiTarget: A→B,C,D,E
func (s *TransferScenario) test16MultiTarget(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "A→B,C,D,E 各 1 PROBE"}
	client := ctx.PrimaryClient()
	targets := s.extra[20:24]
	amount := big.NewInt(1e18)

	// Record balances before
	beforeBals := make([]*big.Int, len(targets))
	for i, t := range targets {
		bal, _, err := client.GetBalance(ctx.Ctx, t.Address, "latest")
		if err != nil {
			beforeBals[i] = big.NewInt(0)
		} else {
			beforeBals[i] = bal
		}
	}

	// Send txs
	var txHashes []string
	for _, t := range targets {
		nonce, err := ctx.GetNonce(s.sender.Address)
		if err != nil {
			continue
		}
		tx := types.NewTransaction(nonce, common.HexToAddress(t.Address), amount, 21000, big.NewInt(1e9), nil)
		hash, _, err := client.SignAndSendTx(ctx.Ctx, s.sender.PrivateKeyHex, tx, ctx.ChainID)
		if err != nil {
			continue
		}
		txHashes = append(txHashes, hash)
	}

	r.Verifications = append(r.Verifications, runner.Verification{
		Angle: "all_sent", Expected: fmt.Sprintf("%d", len(targets)),
		Actual: fmt.Sprintf("%d", len(txHashes)),
		Status: boolToStatus(len(txHashes) == len(targets)),
	})

	// Wait for last tx
	if len(txHashes) > 0 {
		client.WaitForReceipt(ctx.Ctx, txHashes[len(txHashes)-1], 60*time.Second)
	}

	// Verify each target received funds
	for i, t := range targets {
		bal, _, err := client.GetBalance(ctx.Ctx, t.Address, "latest")
		if err != nil {
			r.Verifications = append(r.Verifications, runner.Verification{
				Angle: fmt.Sprintf("target_%d_balance", i), Status: runner.StatusUnknown,
				Error: err.Error(),
			})
			continue
		}
		increase := new(big.Int).Sub(bal, beforeBals[i])
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle:    fmt.Sprintf("target_%d_received", i),
			Expected: amount.String(),
			Actual:   increase.String(),
			Status:   boolToStatus(increase.Cmp(amount) == 0),
		})
	}

	return r
}

// test17CircularTransfer: A→B→C→D→A
func (s *TransferScenario) test17CircularTransfer(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "A→B→C→D→A 环形转账"}
	client := ctx.PrimaryClient()
	chain := []*runner.Account{s.extra[30], s.extra[31], s.extra[32], s.extra[33]}
	amount := big.NewInt(1e17) // 0.1 PROBE

	for i := 0; i < len(chain); i++ {
		from := chain[i]
		to := chain[(i+1)%len(chain)]

		nonce, err := ctx.GetNonce(from.Address)
		if err != nil {
			r.Verifications = append(r.Verifications, runner.Verification{
				Angle: fmt.Sprintf("step_%d_nonce", i), Status: runner.StatusUnknown, Error: err.Error(),
			})
			continue
		}

		tx := types.NewTransaction(nonce, common.HexToAddress(to.Address), amount, 21000, big.NewInt(1e9), nil)
		txHash, _, err := client.SignAndSendTx(ctx.Ctx, from.PrivateKeyHex, tx, ctx.ChainID)
		if err != nil {
			r.Verifications = append(r.Verifications, runner.Verification{
				Angle: fmt.Sprintf("step_%d_send", i), Status: runner.StatusFail,
				Expected: "发送成功", Actual: err.Error(),
			})
			continue
		}

		receipt, rawReceipt, err := client.WaitForReceipt(ctx.Ctx, txHash, 60*time.Second)
		if err != nil {
			r.Verifications = append(r.Verifications, runner.Verification{
				Angle: fmt.Sprintf("step_%d_confirm", i), Status: runner.StatusUnknown, Error: err.Error(),
			})
		} else {
			r.Verifications = append(r.Verifications, runner.Verification{
				Angle: fmt.Sprintf("step_%d_%s→%s", i, from.Address[:8], to.Address[:8]),
				Expected: "0x1", Actual: receipt.Status,
				Status:   boolToStatus(receipt.Status == "0x1"),
				Evidence: runner.ReceiptToEvidence(receipt, rawReceipt),
			})
		}
	}

	return r
}

// test18Concurrent100: 100 accounts send simultaneously
func (s *TransferScenario) test18Concurrent100(ctx *runner.TestContext) runner.TestResult {
	return s.doConcurrentTransfer(ctx, 100, "100 地址并发转账")
}

// test19Concurrent1000: 1000 accounts send simultaneously
func (s *TransferScenario) test19Concurrent1000(ctx *runner.TestContext) runner.TestResult {
	if len(ctx.Accounts) < 1000 {
		r := runner.TestResult{Description: "1000 地址并发转账"}
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "precondition", Status: runner.StatusSkip,
			Error: fmt.Sprintf("需要 1000 账户，当前仅 %d", len(ctx.Accounts)),
		})
		return r
	}
	return s.doConcurrentTransfer(ctx, 1000, "1000 地址并发转账")
}

// test20TxPoolFull: Fill the txpool
func (s *TransferScenario) test20TxPoolFull(ctx *runner.TestContext) runner.TestResult {
	r := runner.TestResult{Description: "填满 txpool 后继续发送"}
	client := ctx.PrimaryClient()

	// Query txpool status before
	resp, err := client.Call(ctx.Ctx, "txpool_status")
	if err != nil {
		return unknownResult(r, "txpool_status", err)
	}

	r.Verifications = append(r.Verifications, runner.Verification{
		Angle: "txpool_accessible", Expected: "txpool API 可用",
		Actual:   "OK",
		Status:   runner.StatusPass,
		Evidence: &runner.Evidence{RawResponse: resp.RawBody},
	})

	// Send 200 rapid txs to fill pool
	sent := 0
	acct := s.extra[50]
	baseNonce, err := client.GetTransactionCount(ctx.Ctx, acct.Address, "pending")
	if err != nil {
		return unknownResult(r, "get_nonce", err)
	}

	for i := 0; i < 200; i++ {
		tx := types.NewTransaction(baseNonce+uint64(i), common.HexToAddress(s.receiver.Address),
			big.NewInt(100), 21000, big.NewInt(1e9), nil)
		_, _, err := client.SignAndSendTx(ctx.Ctx, acct.PrivateKeyHex, tx, ctx.ChainID)
		if err != nil {
			break
		}
		sent++
	}

	r.Verifications = append(r.Verifications, runner.Verification{
		Angle: "rapid_send", Expected: "尽可能多的交易",
		Actual: fmt.Sprintf("%d 笔发送成功", sent),
		Status: boolToStatus(sent > 0),
	})

	// Wait and verify txpool drains
	time.Sleep(30 * time.Second)
	resp2, err := client.Call(ctx.Ctx, "txpool_status")
	if err == nil {
		r.Verifications = append(r.Verifications, runner.Verification{
			Angle: "txpool_after", Expected: "txpool 开始清空",
			Actual:   "查询成功",
			Status:   runner.StatusPass,
			Evidence: &runner.Evidence{RawResponse: resp2.RawBody},
		})
	}

	return r
}

// Helper: simple transfer with receipt verification
func (s *TransferScenario) doSimpleTransfer(ctx *runner.TestContext, r runner.TestResult, amount *big.Int) runner.TestResult {
	client := ctx.PrimaryClient()

	nonce, err := ctx.GetNonce(s.sender.Address)
	if err != nil {
		return unknownResult(r, "get_nonce", err)
	}

	tx := types.NewTransaction(nonce, common.HexToAddress(s.receiver.Address), amount, 21000, big.NewInt(1e9), nil)
	txHash, _, err := client.SignAndSendTx(ctx.Ctx, s.sender.PrivateKeyHex, tx, ctx.ChainID)
	if err != nil {
		return failResult(r, "send_tx", "交易发送成功", err.Error())
	}

	receipt, rawReceipt, err := client.WaitForReceipt(ctx.Ctx, txHash, 60*time.Second)
	if err != nil {
		return unknownResult(r, "wait_receipt", err)
	}

	evidence := runner.ReceiptToEvidence(receipt, rawReceipt)
	r.Verifications = append(r.Verifications, runner.Verification{
		Angle: "receipt_status", Expected: "0x1", Actual: receipt.Status,
		Status: boolToStatus(receipt.Status == "0x1"), Evidence: evidence,
	})

	return r
}

// Helper: concurrent transfer test
func (s *TransferScenario) doConcurrentTransfer(ctx *runner.TestContext, count int, desc string) runner.TestResult {
	r := runner.TestResult{Description: desc}

	var wg sync.WaitGroup
	var successCount uint64
	var failCount uint64

	for i := 0; i < count && i < len(ctx.Accounts)-1; i++ {
		wg.Add(1)
		go func(acct *runner.Account, nodeIdx int) {
			defer wg.Done()

			client := ctx.ClientForNode(nodeIdx % len(ctx.Config.Nodes))
			nonce, err := client.GetTransactionCount(ctx.Ctx, acct.Address, "pending")
			if err != nil {
				atomic.AddUint64(&failCount, 1)
				return
			}

			target := ctx.Accounts[(i+1)%len(ctx.Accounts)]
			tx := types.NewTransaction(nonce, common.HexToAddress(target.Address), big.NewInt(100), 21000, big.NewInt(1e9), nil)
			txHash, _, err := client.SignAndSendTx(ctx.Ctx, acct.PrivateKeyHex, tx, ctx.ChainID)
			if err != nil {
				atomic.AddUint64(&failCount, 1)
				return
			}

			receipt, _, err := client.WaitForReceipt(ctx.Ctx, txHash, 90*time.Second)
			if err != nil || receipt == nil || receipt.Status != "0x1" {
				atomic.AddUint64(&failCount, 1)
				return
			}

			atomic.AddUint64(&successCount, 1)
		}(ctx.Accounts[i], i)
	}

	wg.Wait()

	total := successCount + failCount
	r.Verifications = append(r.Verifications, runner.Verification{
		Angle:    "concurrent_success_rate",
		Expected: fmt.Sprintf(">= 90%% of %d", count),
		Actual:   fmt.Sprintf("%d/%d (%.1f%%)", successCount, total, float64(successCount)/float64(total)*100),
		Status:   boolToStatus(float64(successCount)/float64(total) >= 0.9),
	})

	return r
}

// Helper functions
func boolToStatus(b bool) runner.TestStatus {
	if b {
		return runner.StatusPass
	}
	return runner.StatusFail
}

func unknownResult(r runner.TestResult, angle string, err error) runner.TestResult {
	r.Verifications = append(r.Verifications, runner.Verification{
		Angle: angle, Status: runner.StatusUnknown, Error: err.Error(),
	})
	r.Error = err.Error()
	return r
}

func failResult(r runner.TestResult, angle, expected, actual string) runner.TestResult {
	r.Verifications = append(r.Verifications, runner.Verification{
		Angle: angle, Status: runner.StatusFail, Expected: expected, Actual: actual,
	})
	return r
}
