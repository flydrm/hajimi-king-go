#!/bin/bash

# 检查密钥文件结构的脚本
echo "🔍 检查密钥文件结构..."
echo ""

# 检查data目录是否存在
if [ ! -d "data" ]; then
    echo "❌ data目录不存在，请先运行程序创建data目录"
    exit 1
fi

# 检查keys子目录
if [ ! -d "data/keys" ]; then
    echo "❌ data/keys目录不存在"
    echo "💡 请确保程序已经运行并创建了密钥文件"
    exit 1
fi

echo "📁 数据目录结构:"
echo "data/"
echo "├── keys/"

# 列出keys目录下的文件
echo ""
echo "📄 keys目录下的文件:"
ls -la data/keys/ 2>/dev/null || echo "❌ 无法读取keys目录"

# 检查有效密钥文件
echo ""
echo "🔑 检查有效密钥文件 (keys_valid_*):"
valid_files=$(find data/keys -name "keys_valid_*" 2>/dev/null)
if [ -z "$valid_files" ]; then
    echo "❌ 未找到有效密钥文件"
    echo "💡 有效密钥文件格式: keys_valid_YYYYMMDD_HHMMSS.txt"
else
    echo "✅ 找到有效密钥文件:"
    for file in $valid_files; do
        echo "   📄 $file"
        # 显示文件前几行
        echo "   📝 文件内容预览:"
        head -3 "$file" 2>/dev/null | sed 's/^/     /' || echo "     ❌ 无法读取文件内容"
        echo ""
    done
fi

# 检查限流密钥文件
echo "🚫 检查限流密钥文件 (key_429_*):"
rate_limited_files=$(find data/keys -name "key_429_*" 2>/dev/null)
if [ -z "$rate_limited_files" ]; then
    echo "❌ 未找到限流密钥文件"
    echo "💡 限流密钥文件格式: key_429_YYYYMMDD_HHMMSS.txt"
else
    echo "✅ 找到限流密钥文件:"
    for file in $rate_limited_files; do
        echo "   📄 $file"
        # 显示文件前几行
        echo "   📝 文件内容预览:"
        head -3 "$file" 2>/dev/null | sed 's/^/     /' || echo "     ❌ 无法读取文件内容"
        echo ""
    done
fi

# 检查文件格式
echo ""
echo "📋 检查文件格式示例:"
echo "密钥文件应该包含以下格式的行:"
echo "   AIzaSyxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx|user/repo|path/to/file|https://github.com/user/repo/blob/main/path/to/file"
echo ""

# 提供解决方案
echo "💡 如果没有找到密钥文件，请确保："
echo "   1. 程序已经成功运行并发现了密钥"
echo "   2. .env文件中的配置正确:"
echo "      - DATA_PATH=./data"
echo "      - VALID_KEY_PREFIX=keys/keys_valid_"
echo "      - RATE_LIMITED_KEY_PREFIX=keys/key_429_"
echo "   3. API服务器已启动 (API_ENABLED=true)"
echo ""

echo "🌐 启动API服务器后，可以通过以下方式调试:"
echo "   curl http://localhost:8080/api/debug/files -H \"Authorization: Bearer YOUR_API_KEY\""