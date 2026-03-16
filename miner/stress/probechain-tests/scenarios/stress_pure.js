var CB = "0x4be266e7634b5fca22ec130b46ab7df422e02143";
var A1 = "0x3dff08a96aac3c923fd50f9d878ec73e8fa5f9a1";
personal.unlockAccount(CB, PW_PLACEHOLDER, 3600);

console.log("=== Pure TPS Test (no account creation) ===");
console.log("Block: " + probe.blockNumber);

// Round 1: 1000 tx burst
console.log("\n--- R1: 1000 tx burst ---");
var t1 = Date.now();
var s1 = 0;
for (var i = 0; i < 1000; i++) { try { probe.sendTransaction({from:CB, to:A1, value:'0x1', gas:21000}); s1++; } catch(e){} }
var d1 = (Date.now()-t1)/1000;
console.log("Sent:" + s1 + " Time:" + d1.toFixed(1) + "s Rate:" + (s1/d1).toFixed(0) + "tx/s");
admin.sleep(30);
console.log("Pool:" + JSON.stringify(txpool.status));

// Round 2: 5000 tx burst
console.log("\n--- R2: 5000 tx burst ---");
var b2 = probe.blockNumber;
var t2 = Date.now();
var s2 = 0;
for (var i = 0; i < 5000; i++) { try { probe.sendTransaction({from:CB, to:A1, value:'0x1', gas:21000}); s2++; } catch(e){} }
var d2 = (Date.now()-t2)/1000;
console.log("Sent:" + s2 + " Time:" + d2.toFixed(1) + "s Rate:" + (s2/d2).toFixed(0) + "tx/s");
admin.sleep(45);
console.log("Blocks:" + (probe.blockNumber-b2) + " Pool:" + JSON.stringify(txpool.status));

// Round 3: 10000 tx burst
console.log("\n--- R3: 10000 tx burst ---");
var b3 = probe.blockNumber;
var t3 = Date.now();
var s3 = 0;
for (var i = 0; i < 10000; i++) { try { probe.sendTransaction({from:CB, to:A1, value:'0x1', gas:21000}); s3++; } catch(e){} }
var d3 = (Date.now()-t3)/1000;
console.log("Sent:" + s3 + " Time:" + d3.toFixed(1) + "s Rate:" + (s3/d3).toFixed(0) + "tx/s");
admin.sleep(60);
console.log("Blocks:" + (probe.blockNumber-b3) + " Pool:" + JSON.stringify(txpool.status));

// Summary
console.log("\n=== RESULTS ===");
console.log("R1: " + s1 + "tx " + (s1/d1).toFixed(0) + "tx/s");
console.log("R2: " + s2 + "tx " + (s2/d2).toFixed(0) + "tx/s");
console.log("R3: " + s3 + "tx " + (s3/d3).toFixed(0) + "tx/s");
console.log("Total: " + (s1+s2+s3) + " tx sent");
console.log("Block: " + probe.blockNumber + " Pool:" + JSON.stringify(txpool.status));
