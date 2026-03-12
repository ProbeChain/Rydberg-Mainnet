# ProbeSmartLight — iPhone 部署指南

## 前置条件

- macOS + Apple Silicon
- Xcode 15+ (从 App Store 安装)
- Go 1.22+ (已有: /opt/homebrew/Cellar/go/1.26.0/)
- 二手 iPhone (iOS 16+), USB 连接 Mac
- Apple Developer 账号 (免费即可，用于真机调试)

## 一、安装工具

```bash
# 1. 设置 Xcode
sudo xcode-select -s /Applications/Xcode.app/Contents/Developer
sudo xcodebuild -license accept

# 2. 安装 gomobile
export PATH="/opt/homebrew/Cellar/go/1.26.0/libexec/bin:$PATH"
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest
~/go/bin/gomobile init
```

## 二、编译 Go → iOS Framework

```bash
cd /Volumes/probe/claude/go-probe
./build_ios.sh
```

成功后生成: `build/Gprobe.xcframework`

## 三、创建 Xcode 项目

1. 打开 Xcode → **Create New Project** → **iOS App**
2. Product Name: `ProbeSmartLight`
3. Interface: **SwiftUI**
4. Language: **Swift**
5. Bundle ID: `com.probechain.smartlight`
6. 保存位置: 随意（或直接用 ProbeSmartLight/ 目录）

## 四、导入文件

1. **拖入 Framework**: 把 `build/Gprobe.xcframework` 拖入 Xcode 项目导航
   - 勾选 "Copy items if needed"
   - 在 Target → General → Frameworks 确认已添加

2. **复制 Swift 文件**: 把以下文件拖入 Xcode 项目:
   ```
   ProbeSmartLight/App/ProbeSmartLightApp.swift    → 替换默认 App 入口
   ProbeSmartLight/App/ContentView.swift
   ProbeSmartLight/Core/GoNodeBridge.swift
   ProbeSmartLight/Core/SecureKeyWrapper.swift
   ProbeSmartLight/Core/GNSSTimeProvider.swift
   ProbeSmartLight/Views/DashboardView.swift
   ProbeSmartLight/Views/WalletView.swift
   ProbeSmartLight/Views/AgentView.swift
   ProbeSmartLight/Views/SettingsView.swift
   ProbeSmartLight/Services/NodeService.swift
   ProbeSmartLight/Services/ScoreService.swift
   ProbeSmartLight/Services/RewardService.swift
   ```

3. **Info.plist**: 用 `ProbeSmartLight/Info.plist` 的内容更新项目的 Info.plist
   (后台模式、GPS权限、Face ID 权限)

## 五、Xcode 配置

1. **Signing**: Target → Signing & Capabilities
   - Team: 选择你的 Apple ID
   - Bundle ID: `com.probechain.smartlight`

2. **Capabilities**: 添加:
   - Background Modes: ✅ Background fetch, ✅ Background processing, ✅ Location updates
   - Keychain Sharing (用于 Secure Enclave)

3. **Deployment Target**: iOS 16.0+

## 六、连接 iPhone 并运行

1. USB 连接 iPhone 到 Mac
2. iPhone: 设置 → 隐私与安全性 → 开发者模式 → 打开
3. Xcode 顶部选择你的 iPhone 作为目标设备
4. ⌘R 编译运行

### 首次运行注意:
- iPhone 上会提示 "不受信任的开发者" → 去 设置 → 通用 → VPN与设备管理 → 信任
- 允许 GPS 权限 (选择 "始终允许")
- 允许 Face ID 权限

## 七、验证清单

| # | 验证项 | 如何验证 |
|---|--------|---------|
| 1 | App 启动 | Dashboard 页面显示 "Node Stopped" |
| 2 | 节点启动 | 点 "Start Node" → 状态变 "Node Active" |
| 3 | Header 同步 | Synced Block 数字开始增长 |
| 4 | Peer 连接 | Peers 数量 > 0 |
| 5 | 行为评分 | Score 卡片显示 5000/10000 初始分 |
| 6 | ACK 发送 | (需连接 PoB 全节点) Dashboard 显示 ACK 活动 |
| 7 | 心跳证明 | 每 100 块后 heartbeat 计数增加 |
| 8 | GNSS 时间 | Settings 打开 GPS → 时间样本开始提交 |
| 9 | 电量自适应 | 拔掉充电 → 自动切换 Eco 模式 |
| 10 | Dilithium 密钥 | Wallet → Generate Key → 显示地址 |

## 八、连接本地全节点（可选）

如果你有一台运行 PoB 全节点的机器:

```bash
# 在全节点机器上
cd /Volumes/probe/claude/go-probe
make gprobe
./build/bin/gprobe --networkid 1205 --http --http.api probe,pob,net

# 获取全节点 enode URL
./build/bin/gprobe attach --exec admin.nodeInfo.enode
```

然后在 SmartLight 配置中添加这个 enode 作为 bootstrap node。

## 资源使用

| 指标 | 预期值 |
|------|--------|
| RAM | 50-80 MB |
| 存储 | ~200 MB (headers only) |
| 电量 (充电) | ~5%/小时 |
| 电量 (Eco) | ~2%/小时 |
| 电量 (Sleep) | ~0.5%/小时 |
