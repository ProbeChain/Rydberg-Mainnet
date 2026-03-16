var nodeInfo = admin.nodeInfo;
var myAddr = probe.coinbase;
console.log("=== Multi-Node Stress Test ===");
console.log("Node: " + nodeInfo.name);
console.log("Block: " + probe.blockNumber + " Peers: " + admin.peers.length);

var hexAddrs = [
    '0x4be266e7634b5fca22ec130b46ab7df422e02143',
    '0x9adbf008aed2057d48dd3d4eb720ee2833660c1f',
    '0x4f6ff433914d4fc9ce921fdce7fa1dca1390ef86',
    '0x4cfca7772d2105c2ce914dc482abe1225e6ebc39',
    '0xa259702af0a009e59d0762a335575c72fc897002',
    '0x5eff7908fd6e9f9bc2ac3faed55251f0127d752f',
    '0x28ac9d8c23b7499407dff85bf50651fd6e62352d',
    '0xaf2e662a612e864d4ec272811ff1f519994c18d1',
    '0x123993ae3cda61266b8e924e2f37270e61f199a2'
];
var unlocked = false;
for (var h = 0; h < hexAddrs.length && !unlocked; h++) {
    try { personal.unlockAccount(hexAddrs[h], PW_PLACEHOLDER, 3600); myAddr = hexAddrs[h]; unlocked = true; } catch(e) {}
    try { personal.unlockAccount(hexAddrs[h], '', 3600); myAddr = hexAddrs[h]; unlocked = true; } catch(e) {}
}
if (!unlocked) { console.log("UNLOCK FAILED"); }
console.log("Using: " + myAddr);

var target = hexAddrs[0];
if (myAddr === hexAddrs[0]) target = hexAddrs[1];

console.log("\n--- R1: 5000 tx burst ---");
var b1=probe.blockNumber,t1=Date.now(),s1=0;
for(var i=0;i<5000;i++){try{probe.sendTransaction({from:myAddr,to:target,value:'0x1',gas:21000});s1++}catch(e){}}
var d1=(Date.now()-t1)/1000;
console.log("Sent:"+s1+" Time:"+d1.toFixed(1)+"s Rate:"+(s1/d1).toFixed(0)+"tx/s");
admin.sleep(30);
console.log("Blocks:"+(probe.blockNumber-b1)+" Pool:"+JSON.stringify(txpool.status));

console.log("\n--- R2: 10000 tx burst ---");
var b2=probe.blockNumber,t2=Date.now(),s2=0;
for(var i=0;i<10000;i++){try{probe.sendTransaction({from:myAddr,to:target,value:'0x1',gas:21000});s2++}catch(e){}}
var d2=(Date.now()-t2)/1000;
console.log("Sent:"+s2+" Time:"+d2.toFixed(1)+"s Rate:"+(s2/d2).toFixed(0)+"tx/s");
admin.sleep(45);
console.log("Blocks:"+(probe.blockNumber-b2)+" Pool:"+JSON.stringify(txpool.status));

console.log("\n--- R3: Sustained 3 min ---");
var b3=probe.blockNumber,t3=Date.now(),s3=0,end3=t3+180000;
while(Date.now()<end3){try{probe.sendTransaction({from:myAddr,to:target,value:'0x1',gas:21000});s3++}catch(e){}}
var d3=(Date.now()-t3)/1000;
console.log("Sent:"+s3+" Time:"+d3.toFixed(0)+"s Rate:"+(s3/d3).toFixed(1)+"tx/s");
admin.sleep(30);
console.log("Blocks:"+(probe.blockNumber-b3)+" Pool:"+JSON.stringify(txpool.status));

var total=s1+s2+s3;
console.log("\n=== NODE RESULTS ===");
console.log("Addr:"+myAddr);
console.log("R1:"+s1+"tx "+(s1/d1).toFixed(0)+"tx/s | R2:"+s2+"tx "+(s2/d2).toFixed(0)+"tx/s | R3:"+s3+"tx "+(s3/d3).toFixed(1)+"tx/s");
console.log("Total:"+total+" Block:"+probe.blockNumber+" Pool:"+JSON.stringify(txpool.status));
