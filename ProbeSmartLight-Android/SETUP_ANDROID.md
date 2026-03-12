# ProbeSmartLight Android — 部署指南

## 前置条件

- macOS / Linux / Windows
- Android Studio (Hedgehog 2023.1+)
- Go 1.22+
- gomobile
- Android 手机 (Android 8.0+ / API 26+), USB 连接

## 一、MacBook Pro 安装

```bash
# 1. 安装 Android Studio
brew install --cask android-studio
# 或从 https://developer.android.com/studio 下载（不需要 VPN）

# 2. 安装 Go（如果还没有）
brew install go

# 3. 安装 gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest
gomobile init

# 4. 设置 Android SDK（首次打开 Android Studio 会自动下载）
# 需要: Android SDK 34, NDK, CMake
```

## 二、编译 Go → Android .aar

```bash
cd /Volumes/probe/claude/go-probe
./build_android.sh
```

成功后生成: `build/gprobe.aar`

```bash
# 复制到 Android 项目
cp build/gprobe.aar ProbeSmartLight-Android/app/libs/
```

## 三、打开项目

1. 打开 **Android Studio**
2. **Open** → 选择 `ProbeSmartLight-Android/` 目录
3. 等待 Gradle Sync 完成
4. 如果提示安装 SDK/NDK → 点击安装

## 四、连接 Android 手机

1. 手机: **设置 → 开发者选项 → USB 调试** → 打开
   - (如果没有开发者选项: 设置 → 关于手机 → 连续点击"版本号" 7 次)
2. USB 连接手机到电脑
3. 手机弹出 "允许 USB 调试" → 确认
4. Android Studio 顶部选择你的手机
5. 点 ▶ Run

## 五、验证清单

| # | 验证项 | 如何验证 |
|---|--------|---------|
| 1 | App 安装启动 | Dashboard 显示 "Node Stopped" |
| 2 | 节点启动 | 点 "Start Node" → "Node Active" |
| 3 | 通知栏 | 显示 "ProbeSmartLight — Node syncing" |
| 4 | Header 同步 | Synced Block 数字增长 |
| 5 | Peer 连接 | Peers > 0 |
| 6 | 行为评分 | 初始 5000/10000 |
| 7 | GNSS 时间 | Settings 打开 GPS → 授权位置权限 |
| 8 | 电量自适应 | 拔充电 → 自动切 Eco 模式 |
| 9 | Dilithium 密钥 | Wallet → Generate Key |
| 10 | 后台保活 | 切到后台 → 通知栏仍显示运行 |

## Android vs iOS 对比

| 能力 | iOS (iPhone) | Android |
|------|-------------|---------|
| 密钥保护 | Secure Enclave + Face ID | Keystore StrongBox + 指纹 |
| GNSS | CoreLocation | LocationManager |
| 后台保活 | BGTaskScheduler | ForegroundService + WorkManager |
| 电量自适应 | UIDevice.batteryState | BatteryManager broadcast |
| 最低版本 | iOS 16+ | Android 8.0+ (API 26) |

## 资源使用

| 指标 | 预期值 |
|------|--------|
| RAM | 50-80 MB |
| 存储 | ~200 MB (headers only) |
| 电量 (充电) | ~5%/小时 |
| 电量 (Eco) | ~2%/小时 |
| 电量 (Sleep) | ~0.5%/小时 |
