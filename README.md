# Hajimi King Go v2.0 - 多平台API密钥发现系统

## 🚀 项目概述

Hajimi King Go v2.0 是一个高性能的多平台API密钥发现系统，支持从GitHub代码中智能发现和验证多个平台的API密钥，包括Google Gemini、OpenRouter和SiliconFlow。

## ✨ 核心特性

### 🔧 多平台支持
- **Google Gemini API**: 支持AIza开头的API密钥发现和验证
- **OpenRouter API**: 支持sk-or-开头的API密钥发现和验证  
- **SiliconFlow API**: 支持sk-开头的API密钥发现和验证
- **可扩展架构**: 支持未来添加更多平台

### 🔍 智能验证系统
- **零配置验证**: 使用发现的API密钥本身进行验证，无需预先配置
- **实时验证**: 发现密钥后立即调用对应平台API验证
- **智能分类**: 自动区分有效、无效、限流、未授权等状态
- **安全可靠**: 不存储无效密钥，减少安全风险

### ⚡ 高性能优化
- **并发处理**: Worker Pool模式，支持多平台并发处理
- **多级缓存**: L1内存缓存 + L2文件缓存 + L3Redis缓存（可选）
- **智能检测**: 基于上下文的智能密钥识别，过滤占位符和测试密钥
- **增量扫描**: 支持断点续传和增量更新

### 🎛️ 灵活配置
- **平台开关**: 支持全局开关和单平台控制
- **执行模式**: 全部执行、单平台执行、选中平台执行
- **环境变量**: 完全基于环境变量的配置管理
- **热重载**: 配置变更无需重启

### 📊 监控和指标
- **实时指标**: 处理速度、缓存命中率、检测准确率
- **Web界面**: 现代化的管理界面，支持密钥查看和管理
- **日志系统**: 结构化日志，支持多级别日志记录
- **性能监控**: 详细的性能指标和系统状态

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        Hajimi King Go v2.0 - 优化架构                          │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                                配置层                                            │
├─────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐                  │
│  │ 环境变量配置     │  │ 平台开关配置     │  │ 缓存配置         │                  │
│  │ PLATFORM_*     │  │ PlatformSwitches│  │ CacheConfig     │                  │
│  │ CACHE_*        │  │ ExecutionMode   │  │ MultiLevelCache │                  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘                  │
└─────────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                应用核心层                                        │
├─────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │                        OptimizedHajimiKing                                │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │ │
│  │  │ WorkerPool  │  │CacheManager │  │SmartDetector│  │SystemMetrics│      │ │
│  │  │ 并发处理     │  │ 多级缓存     │  │ 智能识别     │  │ 系统指标     │      │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘      │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              平台处理层                                          │
├─────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │   Gemini    │  │ OpenRouter  │  │ SiliconFlow │  │   Future    │            │
│  │  Platform   │  │  Platform   │  │  Platform   │  │  Platforms  │            │
│  │ 查询+验证   │  │ 查询+验证   │  │ 查询+验证   │  │ 查询+验证   │            │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## 🚀 快速开始

### 1. 环境要求
- Go 1.21+
- GitHub API Token（必需）
- 目标平台的API密钥（可选，用于密钥验证）

### 2. 获取必要的API密钥

#### GitHub Personal Access Token（必需）

**单个Token配置**：
1. 访问 [GitHub Settings > Personal Access Tokens](https://github.com/settings/tokens)
2. 点击 "Generate new token (classic)"
3. 选择权限：
   - `public_repo` - 访问公共仓库
   - `repo` - 访问私有仓库（如果需要）
4. 复制生成的token（格式：`ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`）

**多个Token配置（推荐）**：
- 支持配置多个Token实现负载均衡和故障转移
- 系统会自动轮换使用Token
- 当某个Token失效或限流时，自动切换到下一个Token
- 配置方式：`GITHUB_TOKENS=token1,token2,token3`

#### 其他平台API密钥（不需要预先配置）
**重要说明**: 系统现在使用**发现的API密钥本身**进行验证，无需预先配置任何API密钥！

- **Google Gemini**: 系统会自动发现 `AIzaSy` 开头的密钥并验证
- **OpenRouter**: 系统会自动发现 `sk-or-` 开头的密钥并验证  
- **SiliconFlow**: 系统会自动发现 `sk-` 开头的密钥并验证

**验证流程**:
1. 系统在GitHub代码中发现API密钥
2. 直接使用发现的密钥调用对应平台API
3. 根据API响应判断密钥有效性
4. 分类保存：有效、无效、限流、未授权等

### 3. 安装和配置

```bash
# 克隆项目
git clone <repository-url>
cd hajimi-king-go-v2

# 安装依赖
go mod tidy

# 复制配置文件
cp .env.example .env

# 编辑配置文件
vim .env
```

### 4. 配置说明

#### 🔑 必需配置

**GitHub API配置（必须）**
```bash
# GitHub Personal Access Token - 必需
# 获取地址: https://github.com/settings/tokens
# 权限要求: public_repo, repo (如果需要访问私有仓库)
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# GitHub API代理（可选，如果需要通过代理访问）
GITHUB_PROXY=http://proxy.company.com:8080

# GitHub API基础URL（可选，默认使用官方API）
GITHUB_BASE_URL=https://api.github.com
```

**平台开关配置（必须）**
```bash
# 全局平台开关
PLATFORM_GLOBAL_ENABLED=true

# 执行模式: all(全部平台) | single(单平台) | selected(选中平台)
PLATFORM_EXECUTION_MODE=all

# 各平台开关
PLATFORM_GEMINI_ENABLED=true
PLATFORM_OPENROUTER_ENABLED=false
PLATFORM_SILICONFLOW_ENABLED=false

# 选中平台列表（当EXECUTION_MODE=selected时生效）
PLATFORM_SELECTED=gemini,openrouter
```

#### 🔧 推荐配置

**注意**: 系统现在使用**发现的API密钥本身**进行验证，不再需要预先配置API密钥！

**验证方式说明**:
- 系统发现API密钥后，会直接使用该密钥调用对应平台的API进行验证
- 无需预先配置任何API密钥
- 验证结果包括：有效、无效、限流、未授权等状态

**性能优化配置**
```bash
# Worker Pool大小（建议根据CPU核心数调整）
WORKER_POOL_SIZE=8

# 最大并发文件处理数
MAX_CONCURRENT_FILES=10

# 重试配置
MAX_RETRIES=3
RETRY_DELAY=5s

# 扫描间隔
SCAN_INTERVAL=1m
```

**缓存配置**
```bash
# L1缓存（内存）配置
CACHE_L1_MAX_SIZE=1000
CACHE_L1_TTL=5m

# L2缓存（文件）配置
CACHE_L2_TTL=1h

# L3缓存（Redis）配置（可选）
CACHE_L3_TTL=24h
CACHE_ENABLE_L3=false
CACHE_CLEANUP_INTERVAL=10m
```

**数据存储配置**
```bash
# 数据存储路径
DATA_PATH=./data

# 搜索查询文件路径
QUERIES_PATH=./queries.txt

# 检查点文件路径
CHECKPOINT_PATH=./checkpoint.json

# 文件前缀配置
VALID_KEY_PREFIX=keys_valid
VALID_KEY_DETAIL_PREFIX=keys_valid_detail
RATE_LIMITED_PREFIX=key_429
```

#### 🌐 Web界面配置

```bash
# API服务器端口
API_SERVER_PORT=8080

# JWT密钥（生产环境请修改）
JWT_SECRET=hajimi-king-secret-key
```

#### 🔄 外部同步配置（可选）

```bash
# 同步到Gemini Balancer
SYNC_TO_GEMINI_BALANCER=false
SYNC_TO_GPT_LOAD=false

# 同步端点配置
SYNC_ENDPOINT=https://your-balancer.com/api
SYNC_TOKEN=your_sync_token_here
```

#### 📝 日志配置

```bash
# 日志级别: debug | info | warn | error
LOG_LEVEL=info

# 日志文件路径
LOG_FILE=logs/hajimi-king.log
```

### 5. 完整配置示例

**最小化配置（仅GitHub搜索）**
```bash
# .env 文件内容
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
PLATFORM_GLOBAL_ENABLED=true
PLATFORM_EXECUTION_MODE=all
PLATFORM_GEMINI_ENABLED=true
PLATFORM_OPENROUTER_ENABLED=false
PLATFORM_SILICONFLOW_ENABLED=false
```

**完整配置（包含所有平台）**
```bash
# .env 文件内容
# GitHub配置
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
GITHUB_PROXY=
GITHUB_BASE_URL=https://api.github.com

# 平台开关
PLATFORM_GLOBAL_ENABLED=true
PLATFORM_EXECUTION_MODE=all
PLATFORM_GEMINI_ENABLED=true
PLATFORM_OPENROUTER_ENABLED=true
PLATFORM_SILICONFLOW_ENABLED=true

# 注意：不再需要预先配置API密钥！
# 系统会使用发现的密钥直接进行验证

# 性能配置
WORKER_POOL_SIZE=8
MAX_CONCURRENT_FILES=10
MAX_RETRIES=3
RETRY_DELAY=5s
SCAN_INTERVAL=1m

# 缓存配置
CACHE_L1_MAX_SIZE=1000
CACHE_L1_TTL=5m
CACHE_L2_TTL=1h
CACHE_L3_TTL=24h
CACHE_ENABLE_L3=false

# 数据存储
DATA_PATH=./data
QUERIES_PATH=./queries.txt
CHECKPOINT_PATH=./checkpoint.json

# Web界面
API_SERVER_PORT=8080
JWT_SECRET=your-secure-jwt-secret-key

# 日志
LOG_LEVEL=info
LOG_FILE=logs/hajimi-king.log
```

### 6. 运行应用

```bash
# 编译
go build -o hajimi-king-v2 ./cmd/app

# 运行
./hajimi-king-v2
```

### 7. 访问Web界面

打开浏览器访问: http://localhost:8080

默认登录信息:
- 用户名: admin
- 密码: admin

### 8. 配置验证

**检查配置是否正确**
```bash
# 检查环境变量
echo $GITHUB_TOKEN

# 检查配置文件
cat .env

# 测试GitHub连接
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user
```

**快速测试脚本**
```bash
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
```

**常见配置问题**

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| 启动失败 | 缺少GITHUB_TOKEN | 设置有效的GitHub Personal Access Token |
| 无平台启用 | 平台开关配置错误 | 检查PLATFORM_*_ENABLED设置 |
| API验证失败 | API密钥无效 | 检查各平台的API密钥格式和有效性 |
| 内存不足 | 缓存配置过大 | 调整CACHE_L1_MAX_SIZE |
| 连接超时 | 网络问题 | 配置GITHUB_PROXY或检查网络连接 |
| 编译失败 | Go版本过低 | 升级到Go 1.21+ |
| 权限错误 | GitHub Token权限不足 | 检查Token权限设置 |

## 💡 使用示例

### 基本使用流程

1. **配置系统**
```bash
# 复制配置文件
cp .env.example .env

# 编辑配置（设置GitHub Token）
vim .env
```

2. **启动系统**
```bash
# 编译并运行
go build -o hajimi-king-v2 ./cmd/app
./hajimi-king-v2
```

3. **查看结果**
- 访问Web界面: http://localhost:8080
- 查看发现的密钥文件: `./data/keys_valid_*.txt`
- 查看详细日志: `./data/keys_valid_detail_*.log`

### 不同使用场景

**场景1: 仅搜索GitHub（不验证密钥）**
```bash
# .env 配置
GITHUB_TOKEN=your_github_token
PLATFORM_GLOBAL_ENABLED=true
PLATFORM_EXECUTION_MODE=all
PLATFORM_GEMINI_ENABLED=true
PLATFORM_OPENROUTER_ENABLED=false
PLATFORM_SILICONFLOW_ENABLED=false
```

**场景2: 搜索并验证所有平台**
```bash
# .env 配置
GITHUB_TOKEN=your_github_token
GEMINI_API_KEY=your_gemini_key
OPENROUTER_API_KEY=your_openrouter_key
SILICONFLOW_API_KEY=your_siliconflow_key
PLATFORM_GLOBAL_ENABLED=true
PLATFORM_EXECUTION_MODE=all
PLATFORM_GEMINI_ENABLED=true
PLATFORM_OPENROUTER_ENABLED=true
PLATFORM_SILICONFLOW_ENABLED=true
```

**场景3: 仅验证特定平台**
```bash
# .env 配置
GITHUB_TOKEN=your_github_token
GEMINI_API_KEY=your_gemini_key
PLATFORM_GLOBAL_ENABLED=true
PLATFORM_EXECUTION_MODE=selected
PLATFORM_SELECTED=gemini
PLATFORM_GEMINI_ENABLED=true
PLATFORM_OPENROUTER_ENABLED=false
PLATFORM_SILICONFLOW_ENABLED=false
```

### 监控和调试

**查看实时日志**
```bash
# 查看应用日志
tail -f logs/hajimi-king.log

# 查看特定平台日志
grep "gemini" logs/hajimi-king.log
```

**检查系统状态**
```bash
# 运行配置检查
./test_config.sh

# 检查API连接
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user
```

**性能调优**
```bash
# 调整并发数（根据CPU核心数）
WORKER_POOL_SIZE=16

# 调整缓存大小（根据内存大小）
CACHE_L1_MAX_SIZE=2000

# 调整扫描间隔（根据需求）
SCAN_INTERVAL=30s
```

## 📁 项目结构

```
hajimi-king-go-v2/
├── cmd/app/                    # 主程序入口
│   └── main.go
├── internal/                   # 内部包
│   ├── api/                   # API服务器
│   ├── cache/                 # 多级缓存系统
│   ├── config/                # 配置管理
│   ├── concurrent/            # 并发处理
│   ├── detection/             # 智能检测
│   ├── filemanager/           # 文件管理
│   ├── github/                # GitHub客户端
│   ├── logger/                # 日志系统
│   ├── metrics/               # 指标收集
│   ├── models/                # 数据模型
│   ├── platform/              # 平台抽象
│   └── syncutils/             # 外部同步
├── web/                       # Web界面
│   └── index.html
├── docs/                      # 文档
├── .env.example              # 配置示例
├── queries.txt               # 搜索查询
└── README.md
```

## 🔧 配置选项

### 平台开关配置

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| PLATFORM_GLOBAL_ENABLED | true | 全局平台开关 |
| PLATFORM_EXECUTION_MODE | all | 执行模式：all/single/selected |
| PLATFORM_GEMINI_ENABLED | true | Gemini平台开关 |
| PLATFORM_OPENROUTER_ENABLED | false | OpenRouter平台开关 |
| PLATFORM_SILICONFLOW_ENABLED | false | SiliconFlow平台开关 |

### 性能配置

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| WORKER_POOL_SIZE | 8 | Worker Pool大小 |
| MAX_CONCURRENT_FILES | 10 | 最大并发文件数 |
| CACHE_L1_MAX_SIZE | 1000 | L1缓存最大大小 |
| CACHE_L1_TTL | 5m | L1缓存TTL |
| CACHE_L2_TTL | 1h | L2缓存TTL |

## 📊 监控指标

### 系统指标
- 处理文件数
- 发现密钥数
- 有效密钥数
- 限流密钥数
- 处理速度（密钥/秒）
- 内存使用量

### 缓存指标
- 缓存命中率
- 缓存未命中率
- 缓存大小
- 缓存操作统计

### 检测指标
- 检测准确率
- 误报率
- 过滤统计

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/config
go test ./internal/cache
go test ./internal/detection

# 运行测试并显示覆盖率
go test -cover ./...
```

## 🚀 部署

### Docker部署

```bash
# 构建镜像
docker build -t hajimi-king-v2 .

# 运行容器
docker run -d \
  --name hajimi-king-v2 \
  -p 8080:8080 \
  -e GITHUB_TOKEN=your_token \
  -e PLATFORM_GEMINI_ENABLED=true \
  hajimi-king-v2
```

### 系统服务

```bash
# 创建systemd服务文件
sudo vim /etc/systemd/system/hajimi-king-v2.service

# 启动服务
sudo systemctl enable hajimi-king-v2
sudo systemctl start hajimi-king-v2
```

## 🔒 安全注意事项

1. **API密钥安全**: 确保API密钥安全存储，不要提交到版本控制
2. **访问控制**: 生产环境请修改默认的JWT密钥和登录凭据
3. **网络安全**: 建议在内网环境运行，或配置适当的防火墙规则
4. **日志安全**: 敏感信息会自动脱敏处理

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🙏 致谢

- [Go](https://golang.org/) - 编程语言
- [GitHub API](https://docs.github.com/en/rest) - 代码搜索
- [Google Gemini](https://ai.google.dev/) - AI平台
- [OpenRouter](https://openrouter.ai/) - API平台
- [SiliconFlow](https://siliconflow.cn/) - API平台

## 📞 支持

如有问题或建议，请通过以下方式联系：

- 创建 [Issue](https://github.com/your-repo/issues)
- 发送邮件到: your-email@example.com

---

**Hajimi King Go v2.0** - 让API密钥发现变得简单高效！ 🔑✨