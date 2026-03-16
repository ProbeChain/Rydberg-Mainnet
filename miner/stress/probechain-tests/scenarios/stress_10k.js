// ProbeChain 万地址压力测试
// 在单节点上生成大量账户并发送并发交易，测量 TPS 和成功率

var CB = "0x4be266e7634b5fca22ec130b46ab7df422e02143";
var PW = PW_PLACEHOLDER;

console.log("╔══════════════════════════════════════════════════════════╗");
console.log("║  ProbeChain Rydberg — 万地址压力测试                     ║");
console.log("╚══════════════════════════════════════════════════════════╝");

personal.unlockAccount(CB, PW, 3600);

// ─── Phase 1: 生成测试账户 ─────────────────────────────────
var BATCH = 100; // 每批账户数 (先从100开始,逐步加到1000)
console.log("\n=== Phase 1: 生成 " + BATCH + " 个测试账户 ===");

var accounts = [];
var startTime = Date.now();
for (var i = 0; i < BATCH; i++) {
    var acct = personal.newAccount("stress" + i);
    accounts.push(acct);
    if ((i + 1) % 20 === 0) console.log("  Created " + (i + 1) + "/" + BATCH);
}
var createTime = (Date.now() - startTime) / 1000;
console.log("  " + accounts.length + " accounts created in " + createTime.toFixed(1) + "s");

// ─── Phase 2: 资金分配 ─────────────────────────────────────
console.log("\n=== Phase 2: 分配资金 ===");
var fundStart = Date.now();
var funded = 0;
for (var i = 0; i < accounts.length; i++) {
    try {
        probe.sendTransaction({from: CB, to: accounts[i], value: web3.toWei(10, "probeer"), gas: 21000});
        funded++;
    } catch(e) {}
    if ((i + 1) % 20 === 0) console.log("  Funded " + (i + 1) + "/" + accounts.length);
}
var fundTime = (Date.now() - fundStart) / 1000;
console.log("  " + funded + " funded in " + fundTime.toFixed(1) + "s");

// Wait for funding txs to confirm
console.log("  Waiting for confirmations...");
admin.sleep(30);

// Verify a sample
var verified = 0;
for (var i = 0; i < Math.min(5, accounts.length); i++) {
    var bal = probe.getBalance(accounts[i]);
    if (bal.gt(0)) verified++;
}
console.log("  Sample verification: " + verified + "/5 accounts have funds");

// ─── Phase 3: 压力测试 Round 1 — 基础转账 ──────────────────
console.log("\n=== Phase 3: Round 1 — " + BATCH + " 并发转账 ===");
var blockBefore = probe.blockNumber;
var txStart = Date.now();
var txSent = 0;
var txFailed = 0;
var txHashes = [];

// Unlock all accounts
for (var i = 0; i < accounts.length; i++) {
    personal.unlockAccount(accounts[i], "stress" + i, 300);
}

// Send transactions: each account sends 1 tx to the next account
for (var i = 0; i < accounts.length; i++) {
    try {
        var to = accounts[(i + 1) % accounts.length];
        var hash = probe.sendTransaction({from: accounts[i], to: to, value: web3.toWei(0.001, "probeer"), gas: 21000});
        txHashes.push(hash);
        txSent++;
    } catch(e) {
        txFailed++;
    }
}
var sendTime = (Date.now() - txStart) / 1000;
console.log("  Sent: " + txSent + " tx in " + sendTime.toFixed(2) + "s (" + (txSent/sendTime).toFixed(1) + " tx/s submit rate)");
console.log("  Failed to send: " + txFailed);

// Wait for confirmations
console.log("  Waiting for block confirmations...");
admin.sleep(45);

var blockAfter = probe.blockNumber;
var blocksProduced = blockAfter - blockBefore;

// Check confirmations
var confirmed = 0;
var receiptFailed = 0;
for (var i = 0; i < txHashes.length; i++) {
    var rc = probe.getTransactionReceipt(txHashes[i]);
    if (rc) {
        if (rc.status == 1 || rc.status === true || rc.status === "0x1") confirmed++;
        else receiptFailed++;
    }
}
console.log("  Confirmed: " + confirmed + "/" + txSent);
console.log("  Receipt failed: " + receiptFailed);
console.log("  Blocks produced: " + blocksProduced + " (from " + blockBefore + " to " + blockAfter + ")");
console.log("  Effective TPS: " + (confirmed / 45).toFixed(1));

// ─── Phase 4: 压力测试 Round 2 — 每账户发10笔 ─────────────
console.log("\n=== Phase 4: Round 2 — " + BATCH + " 账户各发10笔 ===");
var r2Start = Date.now();
var r2Block = probe.blockNumber;
var r2Sent = 0;
var r2Failed = 0;

for (var i = 0; i < accounts.length; i++) {
    personal.unlockAccount(accounts[i], "stress" + i, 300);
    for (var j = 0; j < 10; j++) {
        try {
            var to = accounts[(i + j + 1) % accounts.length];
            probe.sendTransaction({from: accounts[i], to: to, value: "0x1", gas: 21000});
            r2Sent++;
        } catch(e) {
            r2Failed++;
        }
    }
    if ((i + 1) % 20 === 0) console.log("  Progress: " + (i + 1) + "/" + accounts.length + " accounts, " + r2Sent + " tx sent");
}
var r2SendTime = (Date.now() - r2Start) / 1000;
console.log("  Total sent: " + r2Sent + " in " + r2SendTime.toFixed(1) + "s (" + (r2Sent/r2SendTime).toFixed(1) + " tx/s)");
console.log("  Failed: " + r2Failed);

// Wait for processing
console.log("  Waiting for processing...");
admin.sleep(60);

var r2BlockAfter = probe.blockNumber;
console.log("  Blocks: " + (r2BlockAfter - r2Block) + " (from " + r2Block + " to " + r2BlockAfter + ")");

// Check txpool
var pool = txpool.status;
console.log("  Txpool: pending=" + pool.pending + " queued=" + pool.queued);

// ─── Phase 5: 压力测试 Round 3 — 持续发送 2 分钟 ──────────
console.log("\n=== Phase 5: Round 3 — 持续发送 2 分钟 ===");
var r3Start = Date.now();
var r3Sent = 0;
var r3End = r3Start + 120000; // 2 minutes
var r3Block = probe.blockNumber;
var acctIdx = 0;

while (Date.now() < r3End) {
    try {
        var from = accounts[acctIdx % accounts.length];
        var to = accounts[(acctIdx + 1) % accounts.length];
        probe.sendTransaction({from: from, to: to, value: "0x1", gas: 21000});
        r3Sent++;
    } catch(e) {}
    acctIdx++;
    // Re-unlock periodically
    if (acctIdx % accounts.length === 0) {
        for (var k = 0; k < accounts.length; k++) {
            personal.unlockAccount(accounts[k], "stress" + k, 300);
        }
    }
}
var r3Time = (Date.now() - r3Start) / 1000;
console.log("  Sent: " + r3Sent + " tx in " + r3Time.toFixed(1) + "s");
console.log("  Submit rate: " + (r3Sent / r3Time).toFixed(1) + " tx/s");

admin.sleep(30);
var r3BlockAfter = probe.blockNumber;
var r3Pool = txpool.status;
console.log("  Blocks: " + (r3BlockAfter - r3Block));
console.log("  Txpool after: pending=" + r3Pool.pending + " queued=" + r3Pool.queued);

// ─── Summary ───────────────────────────────────────────────
console.log("\n╔══════════════════════════════════════════════════════════╗");
console.log("║  压力测试结果汇总                                       ║");
console.log("╠══════════════════════════════════════════════════════════╣");
console.log("║  账户数: " + accounts.length);
console.log("║  Round 1: " + txSent + " tx, " + confirmed + " confirmed, " + (txSent/sendTime).toFixed(1) + " tx/s submit");
console.log("║  Round 2: " + r2Sent + " tx (" + BATCH + "×10), " + (r2Sent/r2SendTime).toFixed(1) + " tx/s submit");
console.log("║  Round 3: " + r3Sent + " tx in 2min, " + (r3Sent/r3Time).toFixed(1) + " tx/s sustained");
console.log("║  最终区块: " + probe.blockNumber);
console.log("║  Txpool: " + JSON.stringify(txpool.status));
console.log("╚══════════════════════════════════════════════════════════╝");
