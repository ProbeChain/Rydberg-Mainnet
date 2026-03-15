// Phase 9: Wallet Operations #171-#185 (15 tests)
var CB="0x4be266e7634b5fca22ec130b46ab7df422e02143",A1="0x3dff08a96aac3c923fd50f9d878ec73e8fa5f9a1",A2="0x4aa863aff32bed671177f1d7a26258f6db248179",PW=process.env.NODE_PASSWORD||"";
var pass=0,fail=0,unk=0;
function t(id,name,fn){try{var r=fn();if(r.status==="PASS"){pass++;console.log("[P] #"+id+" "+name)}else if(r.status==="FAIL"){fail++;console.log("[F] #"+id+" "+name+" "+JSON.stringify(r).substring(0,120))}else{unk++;console.log("[?] #"+id+" "+name)}}catch(e){unk++;console.log("[?] #"+id+" "+name+" ERR:"+e.message.substring(0,80))}}
function txOk(rc){return rc&&(rc.status==1||rc.status===true||rc.status==="0x1")}

personal.unlockAccount(CB,PW,600);
console.log("=== 第9类: 钱包操作 (15测试 #171-#185) ===");

t(171,"创建新账户",function(){var a=personal.newAccount("wallet_test");return{status:a&&a.length>10?"PASS":"FAIL",addr:a}});
t(172,"解锁账户",function(){var ok=personal.unlockAccount(CB,PW,60);return{status:ok?"PASS":"FAIL"}});
t(173,"账户列表",function(){var list=probe.accounts;return{status:list&&list.length>0?"PASS":"FAIL",count:list.length}});
t(174,"余额查询",function(){var b=probe.getBalance(CB);return{status:b.gt(0)?"PASS":"FAIL",bal:web3.fromWei(b,"probeer")}});
t(175,"区块查询",function(){var n=probe.blockNumber;var b=probe.getBlock(n);return{status:b&&b.number===n?"PASS":"FAIL",block:n,hash:b?b.hash:"null",txCount:b?b.transactions.length:0}});
t(176,"交易详情",function(){
  var tx=probe.sendTransaction({from:CB,to:A1,value:web3.toWei(0.01,"probeer"),gas:21000});
  admin.sleep(15);
  var detail=probe.getTransaction(tx);
  return{status:detail&&detail.hash===tx?"PASS":"FAIL",hash:detail?detail.hash:"null",from:detail?detail.from:"null",value:detail?detail.value.toString():"null"};
});
t(177,"交易回执",function(){
  var tx=probe.sendTransaction({from:CB,to:A1,value:"0x1",gas:21000});
  admin.sleep(15);
  var rc=probe.getTransactionReceipt(tx);
  return{status:rc&&rc.transactionHash===tx?"PASS":"FAIL",status_field:rc?rc.status:"null",gas:rc?rc.gasUsed:0};
});
t(178,"签名验证",function(){
  var msg="0x68656c6c6f"; // "hello" in hex
  var sig=probe.sign(CB,msg);
  return{status:sig&&sig.length>10?"PASS":"FAIL",sigLen:sig?sig.length:0};
});
t(179,"gas估算",function(){
  var gas=probe.estimateGas({from:CB,to:A1,value:web3.toWei(1,"probeer")});
  return{status:gas>=21000?"PASS":"FAIL",estimate:gas};
});
t(180,"多账户管理(创建10个)",function(){
  var ok=0;
  for(var i=0;i<10;i++){var a=personal.newAccount("multi"+i);if(a&&a.length>5)ok++}
  return{status:ok>=9?"PASS":"FAIL",created:ok};
});
t(181,"gasPrice查询",function(){var p=probe.gasPrice;return{status:p&&p.gt(0)?"PASS":"FAIL",gasPrice:p?p.toString():"null"}});
t(182,"区块头查询",function(){var b=probe.getBlock("latest");return{status:b&&b.number>0?"PASS":"FAIL",number:b?b.number:0,miner:b?b.miner:"null",gasLimit:b?b.gasLimit:0}});
t(183,"pending交易数",function(){var n=probe.getTransactionCount(CB,"pending");var l=probe.getTransactionCount(CB,"latest");return{status:n>=l?"PASS":"FAIL",pending:n,latest:l}});
t(184,"txpool状态",function(){var s=txpool.status;return{status:s.pending!==undefined?"PASS":"FAIL",pending:s.pending,queued:s.queued}});
t(185,"客户端版本",function(){var v=web3.version.node;return{status:v&&v.indexOf("Gprobe")>=0?"PASS":"FAIL",version:v}});

console.log("");
console.log("RESULTS: "+pass+"/"+(pass+fail+unk)+" passed, "+fail+" failed, "+unk+" unknown");
