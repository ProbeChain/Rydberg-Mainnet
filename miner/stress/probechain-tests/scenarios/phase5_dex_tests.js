// Phase 5: DEX Tests #86-#115 (30 tests) — TokenA + TokenB pair
var CB="0x4be266e7634b5fca22ec130b46ab7df422e02143", A1="0x3dff08a96aac3c923fd50f9d878ec73e8fa5f9a1", PW=process.env.NODE_PASSWORD||"";
var TOKEN_A="0x3a6575bdd09cd185868c55b47ad143b1b1cf66c1";
var SWAP="pro1d3ugls6ld8ddce3lgqfc8weu6u98s2fg3jvnvd";
var pass=0,fail=0,unk=0;
function t(id,name,fn){try{var r=fn();if(r.status==="PASS"){pass++;console.log("[P] #"+id+" "+name)}else if(r.status==="FAIL"){fail++;console.log("[F] #"+id+" "+name+" "+JSON.stringify(r).substring(0,120))}else{unk++;console.log("[?] #"+id+" "+name)}}catch(e){unk++;console.log("[?] #"+id+" "+name+" ERR:"+e.message.substring(0,80))}}
function waitTx(tx){var rc=null;for(var i=0;i<40;i++){rc=probe.getTransactionReceipt(tx);if(rc)break;admin.sleep(1)}return rc}
function txOk(rc){return rc&&(rc.status==1||rc.status===true||rc.status==="0x1")}
function txFail(rc){return rc&&(rc.status==0||rc.status===false||rc.status==="0x0")}
personal.unlockAccount(CB,PW,600);

// Deploy TokenB
console.log("Deploying TokenB...");
var tbTx=probe.sendTransaction({from:CB,data:"0x608060405234801561001057600080fd5b50600380546001600160a01b0319163390811790915569d3c21bcecceda1000000600281905560009182526020829052604090912055610492806100556000396000f3fe608060405234801561001057600080fd5b50600436106100a95760003560e01c806327e235e31161007157806327e235e31461016a57806340c10f19146101905780635c658165146101be57806370a08231146101ec578063a9059cbb14610212578063dd62ed3e1461023e576100a9565b8063047fc9aa146100ae57806307546172146100c8578063095ea7b3146100ec57806318160ddd1461012c57806323b872dd14610134575b600080fd5b6100b661026c565b60408051918252519081900360200190f35b6100d0610272565b604080516001600160a01b039092168252519081900360200190f35b6101186004803603604081101561010257600080fd5b506001600160a01b038135169060200135610281565b604080519115158252519081900360200190f35b6100b66102ab565b6101186004803603606081101561014a57600080fd5b506001600160a01b038135811691602081013590911690604001356102b1565b6100b66004803603602081101561018057600080fd5b50356001600160a01b0316610356565b6101bc600480360360408110156101a657600080fd5b506001600160a01b038135169060200135610368565b005b6100b6600480360360408110156101d457600080fd5b506001600160a01b03813581169160200135166103a9565b6100b66004803603602081101561020257600080fd5b50356001600160a01b03166103c6565b6101186004803603604081101561022857600080fd5b506001600160a01b0381351690602001356103e1565b6100b66004803603604081101561025457600080fd5b506001600160a01b0381358116916020013516610431565b60025481565b6003546001600160a01b031681565b3360009081526001602081815260408084206001600160a01b039690961684529490529290205590565b60025490565b6001600160a01b03831660009081526020819052604081205482118015906102fc57506001600160a01b03841660009081526001602090815260408083203384529091529020548211155b61030557600080fd5b506001600160a01b0392831660008181526020818152604080832080548690039055949095168152838120805484019055908152600180855283822033835290945291909120805491909103905590565b60006020819052908152604090205481565b6003546001600160a01b0316331461037f57600080fd5b60028054820190556001600160a01b03909116600090815260208190526040902080549091019055565b600160209081526000928352604080842090915290825290205481565b6001600160a01b031660009081526020819052604090205490565b336000908152602081905260408120548211156103fd57600080fd5b5033600090815260208190526040808220805484900390556001600160a01b03841682529020805482019055600192915050565b6001600160a01b0391821660009081526001602090815260408083209390941682529190915220549056fea26469706673582212202c40d4d86cfebc4c9b5cd7ebfa5b3d63563b92789a5f6467eb953fedc82bdd9964736f6c634300060c0033",gas:2000000});
var tbRc=waitTx(tbTx);
var TOKEN_B=tbRc.contractAddress;
console.log("TokenB: "+TOKEN_B+" gas:"+tbRc.gasUsed);

var tABI=[{constant:true,inputs:[{name:"",type:"address"}],name:"balanceOf",outputs:[{type:"uint256"}],type:"function"},{constant:false,inputs:[{name:"t",type:"address"},{name:"a",type:"uint256"}],name:"transfer",outputs:[{type:"bool"}],type:"function"},{constant:false,inputs:[{name:"s",type:"address"},{name:"a",type:"uint256"}],name:"approve",outputs:[{type:"bool"}],type:"function"},{constant:false,inputs:[{name:"t",type:"address"},{name:"a",type:"uint256"}],name:"mint",outputs:[],type:"function"},{constant:true,inputs:[],name:"totalSupply",outputs:[{type:"uint256"}],type:"function"}];
var sABI=[{constant:false,inputs:[{name:"a",type:"address"},{name:"b",type:"address"}],name:"createPair",outputs:[{type:"uint256"}],type:"function"},{constant:false,inputs:[{name:"id",type:"uint256"},{name:"a0",type:"uint256"},{name:"a1",type:"uint256"}],name:"addLiquidity",outputs:[{type:"uint256"}],type:"function"},{constant:false,inputs:[{name:"id",type:"uint256"},{name:"lp",type:"uint256"}],name:"removeLiquidity",outputs:[{type:"uint256"},{type:"uint256"}],type:"function"},{constant:false,inputs:[{name:"id",type:"uint256"},{name:"tin",type:"address"},{name:"ain",type:"uint256"},{name:"minOut",type:"uint256"}],name:"swap",outputs:[{type:"uint256"}],type:"function"},{constant:true,inputs:[{name:"id",type:"uint256"}],name:"getReserves",outputs:[{type:"uint256"},{type:"uint256"},{type:"uint256"}],type:"function"},{constant:true,inputs:[],name:"pairCount",outputs:[{type:"uint256"}],type:"function"},{constant:true,inputs:[{name:"a",type:"uint256"},{name:"ri",type:"uint256"},{name:"ro",type:"uint256"}],name:"getAmountOut",outputs:[{type:"uint256"}],type:"function"},{constant:true,inputs:[{name:"id",type:"uint256"},{name:"u",type:"address"}],name:"lpBalanceOf",outputs:[{type:"uint256"}],type:"function"},{constant:true,inputs:[{name:"",type:"address"},{name:"",type:"address"}],name:"getPairId",outputs:[{type:"uint256"}],type:"function"}];

var ta=web3.probe.contract(tABI).at(TOKEN_A);
var tb=web3.probe.contract(tABI).at(TOKEN_B);
var sw=web3.probe.contract(sABI).at(SWAP);

// Approve both tokens to SWAP contract
ta.approve.sendTransaction(SWAP,web3.toWei(999999,"probeer"),{from:CB,gas:100000});
tb.approve.sendTransaction(SWAP,web3.toWei(999999,"probeer"),{from:CB,gas:100000});
admin.sleep(15);

console.log("=== 第5类: ProSwap DEX (30测试 #86-#115) ===");

t(86,"TokenB部署",function(){return{status:probe.getCode(TOKEN_B).length>10?"PASS":"FAIL"}});
t(87,"TokenA余额",function(){var b=ta.balanceOf(CB);return{status:b.gt(0)?"PASS":"FAIL",bal:web3.fromWei(b,"probeer")}});
t(88,"TokenB余额",function(){var b=tb.balanceOf(CB);return{status:b.gt(0)?"PASS":"FAIL",bal:web3.fromWei(b,"probeer")}});
t(89,"MiniSwap合约存在",function(){return{status:probe.getCode(SWAP).length>10?"PASS":"FAIL"}});
t(90,"创建交易对",function(){var tx=sw.createPair.sendTransaction(TOKEN_A,TOKEN_B,{from:CB,gas:500000});var rc=waitTx(tx);return{status:txOk(rc)?"PASS":"FAIL",tx:tx}});
t(91,"重复创建revert",function(){var tx=sw.createPair.sendTransaction(TOKEN_A,TOKEN_B,{from:CB,gas:500000});var rc=waitTx(tx);return{status:txFail(rc)?"PASS":"FAIL"}});
t(92,"相同地址revert",function(){var tx=sw.createPair.sendTransaction(TOKEN_A,TOKEN_A,{from:CB,gas:500000});var rc=waitTx(tx);return{status:txFail(rc)?"PASS":"FAIL"}});
t(93,"添加流动性",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var tx=sw.addLiquidity.sendTransaction(pid,web3.toWei(10000,"probeer"),web3.toWei(10000,"probeer"),{from:CB,gas:500000});var rc=waitTx(tx);var res=sw.getReserves(pid);return{status:txOk(rc)&&res[0].gt(0)?"PASS":"FAIL",r0:res[0].toString(),r1:res[1].toString()}});
t(94,"LP余额>0",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var lp=sw.lpBalanceOf(pid,CB);return{status:lp.gt(0)?"PASS":"FAIL",lp:lp.toString()}});
t(95,"准备金查询",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var r=sw.getReserves(pid);return{status:r[0].gt(0)&&r[1].gt(0)?"PASS":"FAIL",r0:web3.fromWei(r[0],"probeer"),r1:web3.fromWei(r[1],"probeer")}});
t(96,"A→B swap",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var bb=tb.balanceOf(CB);var tx=sw.swap.sendTransaction(pid,TOKEN_A,web3.toWei(100,"probeer"),1,{from:CB,gas:500000});var rc=waitTx(tx);var ba=tb.balanceOf(CB);return{status:txOk(rc)&&ba.gt(bb)?"PASS":"FAIL",before:bb.toString(),after:ba.toString()}});
t(97,"B→A swap",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var bb=ta.balanceOf(CB);tb.approve.sendTransaction(SWAP,web3.toWei(999,"probeer"),{from:CB,gas:100000});admin.sleep(15);var tx=sw.swap.sendTransaction(pid,TOKEN_B,web3.toWei(100,"probeer"),1,{from:CB,gas:500000});var rc=waitTx(tx);var ba=ta.balanceOf(CB);return{status:txOk(rc)&&ba.gt(bb)?"PASS":"FAIL"}});
t(98,"滑点保护revert",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var tx=sw.swap.sendTransaction(pid,TOKEN_A,web3.toWei(1,"probeer"),web3.toWei(99999,"probeer"),{from:CB,gas:500000});var rc=waitTx(tx);return{status:txFail(rc)?"PASS":"FAIL"}});
t(99,"getAmountOut",function(){var o=sw.getAmountOut(web3.toWei(100,"probeer"),web3.toWei(10000,"probeer"),web3.toWei(10000,"probeer"));return{status:o.gt(0)?"PASS":"FAIL",out:o.toString()}});
t(100,"K值>0",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var r=sw.getReserves(pid);return{status:r[0].mul(r[1]).gt(0)?"PASS":"FAIL",k:r[0].mul(r[1]).toString()}});
t(101,"swap后准备金变化",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var rb=sw.getReserves(pid);var tx=sw.swap.sendTransaction(pid,TOKEN_A,web3.toWei(50,"probeer"),1,{from:CB,gas:500000});var rc=waitTx(tx);var ra=sw.getReserves(pid);return{status:txOk(rc)&&!ra[0].eq(rb[0])?"PASS":"FAIL"}});
t(102,"移除流动性",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var lp=sw.lpBalanceOf(pid,CB);var h=lp.div(4);var tx=sw.removeLiquidity.sendTransaction(pid,h,{from:CB,gas:500000});var rc=waitTx(tx);var la=sw.lpBalanceOf(pid,CB);return{status:txOk(rc)&&la.lt(lp)?"PASS":"FAIL",before:lp.toString(),after:la.toString()}});
t(103,"连续5次swap",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var ok=0;for(var i=0;i<5;i++){var tx=sw.swap.sendTransaction(pid,TOKEN_A,web3.toWei(10,"probeer"),1,{from:CB,gas:500000});var rc=waitTx(tx);if(txOk(rc))ok++}return{status:ok>=4?"PASS":"FAIL",ok:ok}});
t(104,"pairCount",function(){var c=sw.pairCount();return{status:c.gt(0)?"PASS":"FAIL",count:c.toString()}});
t(105,"getPairId",function(){var id=sw.getPairId(TOKEN_A,TOKEN_B);return{status:id.gt(0)?"PASS":"FAIL",id:id.toString()}});
t(106,"swap事件(logs)",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var tx=sw.swap.sendTransaction(pid,TOKEN_A,web3.toWei(5,"probeer"),1,{from:CB,gas:500000});var rc=waitTx(tx);return{status:txOk(rc)&&rc.logs.length>0?"PASS":"FAIL",logs:rc?rc.logs.length:0}});
t(107,"TokenA totalSupply",function(){return{status:ta.totalSupply().gt(0)?"PASS":"FAIL"}});
t(108,"TokenB totalSupply",function(){return{status:tb.totalSupply().gt(0)?"PASS":"FAIL"}});
t(109,"TokenA transfer",function(){var tx=ta.transfer.sendTransaction(A1,1000,{from:CB,gas:100000});var rc=waitTx(tx);return{status:txOk(rc)?"PASS":"FAIL"}});
t(110,"再次addLiquidity",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var tx=sw.addLiquidity.sendTransaction(pid,web3.toWei(1000,"probeer"),web3.toWei(1000,"probeer"),{from:CB,gas:500000});var rc=waitTx(tx);return{status:txOk(rc)?"PASS":"FAIL"}});
t(111,"价格影响(大单vs小单)",function(){var o1=sw.getAmountOut(web3.toWei(1,"probeer"),web3.toWei(10000,"probeer"),web3.toWei(10000,"probeer"));var o2=sw.getAmountOut(web3.toWei(1000,"probeer"),web3.toWei(10000,"probeer"),web3.toWei(10000,"probeer"));return{status:o1.mul(1000).gt(o2)?"PASS":"FAIL",small:o1.toString(),large:o2.toString()}});
t(112,"swap gas消耗",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var tx=sw.swap.sendTransaction(pid,TOKEN_A,web3.toWei(1,"probeer"),1,{from:CB,gas:500000});var rc=waitTx(tx);return{status:txOk(rc)&&rc.gasUsed>21000?"PASS":"FAIL",gas:rc?rc.gasUsed:0}});
t(113,"addLiquidity gas",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var tx=sw.addLiquidity.sendTransaction(pid,web3.toWei(100,"probeer"),web3.toWei(100,"probeer"),{from:CB,gas:500000});var rc=waitTx(tx);return{status:txOk(rc)&&rc.gasUsed>21000?"PASS":"FAIL",gas:rc?rc.gasUsed:0}});
t(114,"批量10次swap",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var ok=0;var hs=[];for(var i=0;i<10;i++){hs.push(sw.swap.sendTransaction(pid,TOKEN_A,web3.toWei(1,"probeer"),1,{from:CB,gas:500000}))}admin.sleep(30);for(var j=0;j<hs.length;j++){var r=probe.getTransactionReceipt(hs[j]);if(txOk(r))ok++}return{status:ok>=8?"PASS":"FAIL",ok:ok+"/10"}});
t(115,"全流程验证",function(){var pid=sw.getPairId(TOKEN_A,TOKEN_B);var lp=sw.lpBalanceOf(pid,CB);var r=sw.getReserves(pid);return{status:pid.gt(0)&&lp.gt(0)&&r[0].gt(0)?"PASS":"FAIL",pid:pid.toString(),lp:lp.toString(),r0:r[0].toString()}});

console.log("");
console.log("RESULTS: "+pass+"/"+(pass+fail+unk)+" passed, "+fail+" failed, "+unk+" unknown");
console.log(pass===(pass+fail+unk)?"ALL 30 PASSED":"SOME ISSUES");
