# ProbeChain 里德堡大升级 (Rydberg Upgrade)

## 网络拓扑

```
  Validator 1 (Genesis)          Validator 2              Validator 3
  192.168.110.142:30303    ←→    Machine B:30304    ←→    Machine C:30304
  0x63b356...1b982               (待创建)                  (待创建)
```

## 部署包内容

```
rydberg-deploy/
├── gprobe                  # 编译好的节点二进制 (macOS amd64)
├── genesis_pob.json        # PoB 创世配置
├── setup_validator.sh      # 一键部署脚本
└── RYDBERG_UPGRADE.md      # 本文档
```

## 操作步骤

### 第一步：传输部署包到另外两台电脑

```bash
# 在创世节点机器上，用 scp 或 AirDrop 传输
scp -r ~/Desktop/go-probe/rydberg-deploy/ user@MACHINE_B_IP:~/rydberg-deploy/
scp -r ~/Desktop/go-probe/rydberg-deploy/ user@MACHINE_C_IP:~/rydberg-deploy/
```

### 第二步：在每台新机器上运行部署脚本

```bash
cd ~/rydberg-deploy
./setup_validator.sh
```

脚本会自动：创建账户 → 初始化创世 → 启动节点 → 连接创世节点

**记下输出的验证者地址**，例如：`0xABCD...1234`

### 第三步：在创世节点上投票添加新验证者

回到创世节点机器 (192.168.110.142)，对每个新验证者投票：

```bash
# 投票添加 Validator 2
curl -X POST http://127.0.0.1:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"pob_propose","params":["0x新验证者2地址", true],"id":1}'

# 投票添加 Validator 3
curl -X POST http://127.0.0.1:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"pob_propose","params":["0x新验证者3地址", true],"id":1}'
```

### 第四步：验证升级成功

```bash
# 查看当前验证者列表
curl -X POST http://127.0.0.1:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"pob_getValidators","params":[],"id":1}'

# 查看行为分数
curl -X POST http://127.0.0.1:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"pob_getBehaviorScores","params":[],"id":1}'

# 查看连接的 peers
curl -X POST http://127.0.0.1:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"admin_peers","params":[],"id":1}'
```

## 注意事项

- 所有机器必须在同一局域网（或有公网 IP）
- 如果另外两台电脑不是 macOS amd64，需要交叉编译 gprobe
- 投票通过需要 >50% 现有验证者同意（目前只有 1 个，所以你的投票直接生效）
- 新验证者加入后，ValidatorWitnessCount 变为 3，将自动切换到完整 ACK 共识流程
