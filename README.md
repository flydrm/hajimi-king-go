# Hajimi King Go v2.0 - 多平台API密钥发现系统

## 🚀 项目概述

Hajimi King Go v2.0 是一个高性能的多平台API密钥发现系统，支持从GitHub代码中智能发现和验证多个平台的API密钥，包括Google Gemini、OpenRouter和SiliconFlow。

## ✨ 核心特性

### 🔧 多平台支持
- **Google Gemini API**: 支持AIza开头的API密钥发现和验证
- **OpenRouter API**: 支持sk-or-开头的API密钥发现和验证  
- **SiliconFlow API**: 支持sk-开头的API密钥发现和验证
- **可扩展架构**: 支持未来添加更多平台

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
- GitHub API Token
- 目标平台的API密钥（可选）

### 2. 安装和配置

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

### 3. 配置说明

#### 必需配置
```bash
# GitHub配置
GITHUB_TOKEN=your_github_token_here

# 平台开关
PLATFORM_GLOBAL_ENABLED=true
PLATFORM_EXECUTION_MODE=all
PLATFORM_GEMINI_ENABLED=true
```

#### 可选配置
```bash
# API密钥（用于验证）
GEMINI_API_KEY=your_gemini_api_key_here
OPENROUTER_API_KEY=your_openrouter_api_key_here
SILICONFLOW_API_KEY=your_siliconflow_api_key_here

# 性能调优
WORKER_POOL_SIZE=8
MAX_CONCURRENT_FILES=10
CACHE_L1_MAX_SIZE=1000

# 外部同步
SYNC_TO_GEMINI_BALANCER=false
SYNC_TO_GPT_LOAD=false
```

### 4. 运行应用

```bash
# 编译
go build -o hajimi-king-v2 ./cmd/app

# 运行
./hajimi-king-v2
```

### 5. 访问Web界面

打开浏览器访问: http://localhost:8080

默认登录信息:
- 用户名: admin
- 密码: admin

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