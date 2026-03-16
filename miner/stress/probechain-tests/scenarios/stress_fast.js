var CB = "0x4be266e7634b5fca22ec130b46ab7df422e02143";
personal.unlockAccount(CB, PW_PLACEHOLDER, 3600);

console.log("╔══════════════════════════════════════════════════════════╗");
console.log("║  ProbeChain 万地址压力测试 — Fast Mode (4C8G)            ║");
console.log("╚══════════════════════════════════════════════════════════╝");

// Use all existing accounts
var accts = personal.listAccounts;
console.log("Existing accounts: " + accts.length);

// Unlock all with known passwords
var unlocked = 0;
var passwords = [PW_PLACEHOLDER,PW_PLACEHOLDER,PW_PLACEHOLDER,'tp1','tp2'];
for (var i = 0; i < accts.length; i++) {
    for (var p = 0; p < passwords.length; p++) {
        try { personal.unlockAccount(accts[i], passwords[p], 3600); unlocked++; break; } catch(e) {}
    }
    // Also try stress passwords
    try { personal.unlockAccount(accts[i], 's' + i, 3600); unlocked++; } catch(e) {}
    try { personal.unlockAccount(accts[i], 'stress' + i, 3600); unlocked++; } catch(e) {}
}
console.log("Unlocked: " + unlocked);

// Create 50 new accounts quickly (4C8G should handle it)
console.log("\n=== Creating 50 new accounts ===");
var t0 = Date.now();
var newAccts = [];
for (var i = 0; i < 50; i++) {
    newAccts.push(personal.newAccount('p' + i));
    if ((i+1) % 10 === 0) console.log("  " + (i+1) + "/50 (" + ((Date.now()-t0)/1000).toFixed(1) + "s)");
}
console.log("Created 50 in " + ((Date.now()-t0)/1000).toFixed(1) + "s");

// Fund new accounts
console.log("\n=== Funding 50 accounts ===");
for (var i = 0; i < 50; i++) {
    probe.sendTransaction({from:CB, to:newAccts[i], value:web3.toWei(5,'probeer'), gas:21000});
}
admin.sleep(20);

// Unlock new accounts
for (var i = 0; i < 50; i++) { personal.unlockAccount(newAccts[i], 'p'+i, 3600); }

// All test accounts
var all = newAccts;
console.log("Test accounts: " + all.length);

// ═══ Round 1: 50 concurrent transfers ═══
console.log("\n=== Round 1: " + all.length + " concurrent transfers ===");
var r1Start = Date.now();
var r1Block = probe.blockNumber;
var r1Sent = 0;
for (var i = 0; i < all.length; i++) {
    try { probe.sendTransaction({from:all[i], to:all[(i+1)%all.length], value:'0x1', gas:21000}); r1Sent++; } catch(e){}
}
var r1Time = (Date.now()-r1Start)/1000;
console.log("Sent: " + r1Sent + " in " + r1Time.toFixed(2) + "s (" + (r1Sent/r1Time).toFixed(0) + " tx/s submit)");
admin.sleep(20);
console.log("Blocks: " + (probe.blockNumber - r1Block));

// ═══ Round 2: Each account sends 20 tx = 1000 total ═══
console.log("\n=== Round 2: " + all.length + " x 20 = " + (all.length*20) + " tx ===");
var r2Start = Date.now();
var r2Block = probe.blockNumber;
var r2Sent = 0;
for (var i = 0; i < all.length; i++) {
    for (var j = 0; j < 20; j++) {
        try { probe.sendTransaction({from:all[i], to:all[(i+j+1)%all.length], value:'0x1', gas:21000}); r2Sent++; } catch(e){}
    }
}
var r2Time = (Date.now()-r2Start)/1000;
console.log("Sent: " + r2Sent + " in " + r2Time.toFixed(1) + "s (" + (r2Sent/r2Time).toFixed(0) + " tx/s)");
admin.sleep(30);
console.log("Blocks: " + (probe.blockNumber - r2Block));
console.log("Txpool: " + JSON.stringify(txpool.status));

// ═══ Round 3: Sustained 2 minutes ═══
console.log("\n=== Round 3: Sustained 2 min ===");
var r3Start = Date.now();
var r3Block = probe.blockNumber;
var r3End = r3Start + 120000;
var r3Sent = 0;
var idx = 0;
while (Date.now() < r3End) {
    try { probe.sendTransaction({from:all[idx%all.length], to:all[(idx+1)%all.length], value:'0x1', gas:21000}); r3Sent++; } catch(e){}
    idx++;
}
var r3Time = (Date.now()-r3Start)/1000;
console.log("Sent: " + r3Sent + " in " + r3Time.toFixed(0) + "s (" + (r3Sent/r3Time).toFixed(1) + " tx/s sustained)");
admin.sleep(30);
var r3Blocks = probe.blockNumber - r3Block;
console.log("Blocks: " + r3Blocks);
console.log("Txpool: " + JSON.stringify(txpool.status));

// ═══ Round 4: Burst — CB sends 5000 tx rapidly ═══
console.log("\n=== Round 4: CB burst 5000 tx ===");
var r4Start = Date.now();
var r4Block = probe.blockNumber;
var r4Sent = 0;
for (var i = 0; i < 5000; i++) {
    try { probe.sendTransaction({from:CB, to:all[i%all.length], value:'0x1', gas:21000}); r4Sent++; } catch(e){}
}
var r4Time = (Date.now()-r4Start)/1000;
console.log("Sent: " + r4Sent + " in " + r4Time.toFixed(1) + "s (" + (r4Sent/r4Time).toFixed(0) + " tx/s)");
admin.sleep(60);
var r4Blocks = probe.blockNumber - r4Block;
console.log("Blocks: " + r4Blocks + ", txpool: " + JSON.stringify(txpool.status));

// ═══ Summary ═══
console.log("\n╔══════════════════════════════════════════════════════════╗");
console.log("║  压力测试结果                                            ║");
console.log("╠══════════════════════════════════════════════════════════╣");
console.log("║  R1: " + r1Sent + " concurrent (" + (r1Sent/r1Time).toFixed(0) + " tx/s submit)");
console.log("║  R2: " + r2Sent + " batch (" + (r2Sent/r2Time).toFixed(0) + " tx/s)");
console.log("║  R3: " + r3Sent + " sustained 2min (" + (r3Sent/r3Time).toFixed(1) + " tx/s)");
console.log("║  R4: " + r4Sent + " burst (" + (r4Sent/r4Time).toFixed(0) + " tx/s)");
console.log("║  Final block: " + probe.blockNumber);
console.log("║  Txpool: " + JSON.stringify(txpool.status));
console.log("╚══════════════════════════════════════════════════════════╝");
