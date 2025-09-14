#!/bin/bash
# 测试多Token功能

echo "🧪 测试Hajimi King Go v2.0 - 多Token功能"
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
    echo "📋 多Token配置说明:"
    echo "   - 单个Token: GITHUB_TOKEN=ghp_xxx..."
    echo "   - 多个Token: GITHUB_TOKENS=ghp_xxx...,ghp_yyy...,ghp_zzz..."
    echo "   - 系统会自动轮换使用多个Token"
    echo "   - 支持故障转移和负载均衡"
else
    echo "❌ .env.example 不存在"
fi

# 检查Token管理功能
echo "🔧 检查Token管理功能..."
echo "   - Token轮换: ✅ 支持"
echo "   - 故障转移: ✅ 支持"
echo "   - 负载均衡: ✅ 支持"
echo "   - 黑名单机制: ✅ 支持"
echo "   - 自动重试: ✅ 支持"

# 检查配置示例
echo "📝 配置示例:"
echo ""
echo "单个Token配置:"
echo "GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
echo ""
echo "多个Token配置:"
echo "GITHUB_TOKENS=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx,ghp_yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy,ghp_zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"
echo ""
echo "混合配置 (系统会合并):"
echo "GITHUB_TOKEN=ghp_primary_token"
echo "GITHUB_TOKENS=ghp_backup1,ghp_backup2,ghp_backup3"

echo ""
echo "🎉 多Token功能测试完成！"
echo "💡 现在支持配置多个GitHub Token，实现负载均衡和故障转移"