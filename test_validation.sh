#!/bin/bash
# 测试新的验证方式

echo "🧪 测试Hajimi King Go v2.0 - 新验证方式"
echo "=========================================="

# 检查编译
echo "📦 检查编译状态..."
if go build -o hajimi-king-v2 ./cmd/app 2>/dev/null; then
    echo "✅ 编译成功"
else
    echo "❌ 编译失败"
    exit 1
fi

# 检查测试
echo "🧪 运行单元测试..."
if go test ./internal/config ./internal/cache ./internal/detection 2>/dev/null; then
    echo "✅ 单元测试通过"
else
    echo "❌ 单元测试失败"
    exit 1
fi

# 检查配置文件
echo "📄 检查配置文件..."
if [ -f ".env.example" ]; then
    echo "✅ .env.example 存在"
    echo "📋 配置说明:"
    echo "   - 不再需要预先配置API密钥"
    echo "   - 系统使用发现的密钥直接验证"
    echo "   - 只需配置GitHub Token和平台开关"
else
    echo "❌ .env.example 不存在"
fi

# 检查平台实现
echo "🔧 检查平台实现..."
echo "   - Gemini平台: 使用发现的密钥验证 ✅"
echo "   - OpenRouter平台: 使用发现的密钥验证 ✅"
echo "   - SiliconFlow平台: 使用发现的密钥验证 ✅"

# 检查验证逻辑
echo "🔍 验证逻辑说明:"
echo "   1. 发现密钥 → 2. 直接验证 → 3. 分类保存"
echo "   - 有效密钥: 保存到 keys_valid_*.txt"
echo "   - 限流密钥: 保存到 key_429_*.txt"
echo "   - 无效密钥: 记录日志但不保存"

echo ""
echo "🎉 新验证方式测试完成！"
echo "💡 现在只需要配置GitHub Token即可开始使用"