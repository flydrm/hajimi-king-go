#!/bin/bash
# 快速测试脚本 - test_config.sh

echo "🔍 检查Hajimi King Go v2.0配置..."

# 检查Go版本
echo "📦 Go版本:"
go version

# 检查环境变量
echo "🔑 环境变量检查:"
if [ -z "$GITHUB_TOKEN" ]; then
    echo "❌ GITHUB_TOKEN 未设置"
else
    echo "✅ GITHUB_TOKEN 已设置"
fi

# 检查配置文件
echo "📄 配置文件检查:"
if [ -f ".env" ]; then
    echo "✅ .env 文件存在"
    echo "📋 当前配置:"
    grep -E "^(GITHUB_TOKEN|PLATFORM_|GEMINI_API_KEY|OPENROUTER_API_KEY|SILICONFLOW_API_KEY)" .env | sed 's/=.*/=***/'
else
    echo "❌ .env 文件不存在，请复制 .env.example 并配置"
fi

# 检查编译
echo "🔨 编译测试:"
if go build -o hajimi-king-v2 ./cmd/app 2>/dev/null; then
    echo "✅ 编译成功"
    rm -f hajimi-king-v2
else
    echo "❌ 编译失败"
fi

# 检查测试
echo "🧪 测试运行:"
if go test ./internal/config ./internal/cache ./internal/detection 2>/dev/null; then
    echo "✅ 测试通过"
else
    echo "❌ 测试失败"
fi

echo "🎉 配置检查完成！"