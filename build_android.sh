#!/bin/bash
# build_android.sh — 编译 SmartLight Go 代码为 Android .aar
# 用法: ./build_android.sh
set -e

export PATH="/opt/homebrew/Cellar/go/1.26.0/libexec/bin:$HOME/go/bin:$PATH"
cd "$(dirname "$0")"

echo "=== 1. 检查工具 ==="
go version
gomobile version

echo ""
echo "=== 2. 编译 Gprobe.aar ==="
echo "    包含: mobile/ (SmartLightNode + DilithiumKeyStore + LES client)"
echo "    目标: Android arm64 + arm"
echo ""

mkdir -p build

gomobile bind \
  -target android/arm64,android/arm \
  -androidapi 26 \
  -o ./build/gprobe.aar \
  -v \
  ./mobile/

echo ""
echo "=== 3. 完成! ==="
echo "生成文件: ./build/gprobe.aar"
echo ""
echo "下一步:"
echo "  1. 复制 build/gprobe.aar 到 ProbeSmartLight-Android/app/libs/"
echo "  2. 用 Android Studio 打开 ProbeSmartLight-Android/"
echo "  3. 连接 Android 手机 → Run"
