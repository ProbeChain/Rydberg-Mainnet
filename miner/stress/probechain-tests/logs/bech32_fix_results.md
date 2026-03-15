# Bech32/Hex 地址兼容性修复

## 问题
ProbeChain的gprobe console返回bech32 (pro1...) 地址，但sendTransaction等函数不接受它们，
用户必须手动转换为0x hex格式才能使用。

## 根因
web3.js的inputAddressFormatter只验证0x hex格式，不识别bech32。
web3.js通过go-bindata嵌入bindata.go，修改后必须重新运行go-bindata。

## 修复 (3处)
1. internal/jsre/deps/web3.js — inputAddressFormatter: 内联bech32解码器
2. internal/jsre/deps/web3.js — SolidityTypeAddress._inputFormatter: 合约ABI编码前转换
3. probe/filters/api.go — decodeAddress(): 支持bech32地址过滤

## 测试结果 (8/8 通过)
- sendTx from=bech32 to=bech32 ✓
- sendTx from=hex to=bech32 ✓
- sendTx from=bech32 to=hex ✓
- sendTx to=newAccount(bech32) ✓
- getBalance(bech32) ✓
- getTransactionCount(bech32) ✓
- contract.balanceOf(bech32) ✓
- contract.transfer(bech32) ✓

## 关键发现
- ProbeChain EVM不支持PUSH0 (Shanghai), 需用 --evm-version london 编译合约
- go-bindata必须在修改web3.js后重新生成bindata.go
- hrpExpand('pro') = [3,3,3,0,16,18,15] (不是[3,4,3,...])
