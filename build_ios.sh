#!/bin/bash
# build_ios.sh — 编译 SmartLight Go 代码为 iOS .xcframework
# 用法: ./build_ios.sh
set -e

export PATH="/usr/local/Cellar/go/1.26.1/libexec/bin:$HOME/go/bin:$PATH"
cd "$(dirname "$0")"

echo "=== 1. 检查工具 ==="
go version
gomobile version || echo "(gomobile version 报告不影响编译)"

echo ""
echo "=== 2. 编译 Gprobe.xcframework ==="
echo "    包含: mobile/ (SmartLightNode + DilithiumKeyStore + LES client)"
echo "    目标: iOS arm64 (真机)"
echo ""

# gomobile bind 会生成 Gprobe.xcframework
# -target ios 只编译真机 arm64
# -o 指定输出路径
gomobile bind \
  -target ios \
  -o ./build/Gprobe.xcframework \
  -v \
  ./mobile/

echo ""
echo "=== 3. 完成! ==="
echo "生成文件: ./build/Gprobe.xcframework"
echo ""
echo "下一步:"
echo "  1. 打开 Xcode"
echo "  2. 创建新 iOS App 项目 'ProbeSmartLight'"
echo "  3. 把 build/Gprobe.xcframework 拖入项目"
echo "  4. 把 ProbeSmartLight/ 下的 Swift 文件复制到项目"
echo "  5. 连接 iPhone → Build & Run"
