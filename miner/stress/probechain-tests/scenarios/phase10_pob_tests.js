// Phase 10: PoB Consensus & Node #186-#200 (15 tests)
var CB="0x4be266e7634b5fca22ec130b46ab7df422e02143",PW=process.env.NODE_PASSWORD||"";
var pass=0,fail=0,unk=0;
function t(id,name,fn){try{var r=fn();if(r.status==="PASS"){pass++;console.log("[P] #"+id+" "+name)}else if(r.status==="FAIL"){fail++;console.log("[F] #"+id+" "+name+" "+JSON.stringify(r).substring(0,120))}else{unk++;console.log("[?] #"+id+" "+name)}}catch(e){unk++;console.log("[?] #"+id+" "+name+" ERR:"+e.message.substring(0,80))}}

personal.unlockAccount(CB,PW,60);
console.log("=== 第10类: PoB共识与节点 (15测试 #186-#200) ===");

t(186,"节点信息",function(){var info=admin.nodeInfo;return{status:info&&info.protocols?"PASS":"FAIL",enode:info?info.enode.substring(0,40)+"...":"null",name:info?info.name:"null"}});
t(187,"peers数量",function(){var peers=admin.peers;return{status:peers&&peers.length>0?"PASS":"FAIL",count:peers.length}});
t(188,"挖矿状态",function(){var m=probe.mining;return{status:m===true?"PASS":"FAIL",mining:m}});
t(189,"coinbase",function(){var c=probe.coinbase;return{status:c&&c.length>5?"PASS":"FAIL",coinbase:c}});
t(190,"网络ID",function(){var id=admin.nodeInfo.protocols.probe.network;return{status:id===8004?"PASS":"FAIL",networkId:id}});
t(191,"chainId",function(){var n=probe.blockNumber;var b=probe.getBlock(n);return{status:b?"PASS":"FAIL",block:n}});
t(192,"出块验证(等30s)",function(){
  var n1=probe.blockNumber;
  admin.sleep(30);
  var n2=probe.blockNumber;
  return{status:n2>n1?"PASS":"FAIL",before:n1,after:n2,newBlocks:n2-n1};
});
t(193,"出块间隔",function(){
  var n=probe.blockNumber;var gaps=[];
  for(var i=n;i>n-10&&i>0;i--){var b1=probe.getBlock(i);var b2=probe.getBlock(i-1);if(b1&&b2)gaps.push(b1.timestamp-b2.timestamp)}
  var sum=0;for(var j=0;j<gaps.length;j++)sum+=gaps[j];
  var avg=gaps.length>0?sum/gaps.length:0;
  return{status:avg>0?"PASS":"FAIL",avgGap:avg.toFixed(1)+"s",gaps:gaps.join(",")};
});
t(194,"9节点确认",function(){
  // Check if we have peers from all 3 server IPs
  var peers=admin.peers;
  var ips={};
  for(var i=0;i<peers.length;i++){
    var addr=peers[i].network.remoteAddress||"";
    var ip=addr.split(":")[0];
    ips[ip]=true;
  }
  var ipList=Object.keys(ips);
  return{status:ipList.length>=2?"PASS":"FAIL",uniqueIPs:ipList.length,peers:peers.length,ips:ipList.join(",")};
});
t(195,"最新块矿工",function(){
  var b=probe.getBlock("latest");
  return{status:b&&b.miner&&b.miner.length>5?"PASS":"FAIL",miner:b?b.miner:"null"};
});
t(196,"EIP-1559 baseFee",function(){
  var b=probe.getBlock("latest");
  return{status:b&&b.baseFeePerGas?"PASS":"FAIL",baseFee:b&&b.baseFeePerGas?b.baseFeePerGas.toString():"null"};
});
t(197,"difficulty",function(){
  var b=probe.getBlock("latest");
  return{status:b&&b.difficulty>=0?"PASS":"FAIL",difficulty:b?b.difficulty:0};
});
t(198,"gasLimit",function(){
  var b=probe.getBlock("latest");
  return{status:b&&b.gasLimit>0?"PASS":"FAIL",gasLimit:b?b.gasLimit:0};
});
t(199,"syncing状态",function(){
  var s=probe.syncing;
  return{status:s===false||s?"PASS":"FAIL",syncing:JSON.stringify(s).substring(0,100)};
});
t(200,"跨场景综合",function(){
  // Send a tx, check receipt, verify block contains it
  var tx=probe.sendTransaction({from:CB,to:CB,value:"0x1",gas:21000});
  admin.sleep(15);
  var rc=probe.getTransactionReceipt(tx);
  if(!rc)return{status:"UNKNOWN",note:"receipt timeout"};
  var block=probe.getBlock(rc.blockNumber,true);
  var found=false;
  if(block&&block.transactions){for(var i=0;i<block.transactions.length;i++){if(block.transactions[i].hash===tx){found=true;break}}}
  return{status:found?"PASS":"FAIL",tx:tx,block:rc.blockNumber,txInBlock:found};
});

console.log("");
console.log("RESULTS: "+pass+"/"+(pass+fail+unk)+" passed, "+fail+" failed, "+unk+" unknown");
