# Hajimi King Go

🎪 **Hajimi King Go** - 人人都是哈基米大王 👑  

这是一个用于从GitHub搜索和验证Google Gemini API密钥的Go语言版本项目。  
基于原Python版本重构，提供更好的性能和并发处理能力。

⚠️ **注意**：本项目正处于beta期间，功能、结构、接口等都有可能变化，不保证稳定性，请自行承担风险。

## 🚀 核心功能

1. **🔍 GitHub搜索Gemini Key** - 基于自定义查询表达式搜索GitHub代码中的API密钥
2. **🌐 代理支持** - 支持多代理轮换，提高访问稳定性和成功率
3. **📊 增量扫描** - 支持断点续传，避免重复扫描已处理的文件
4. **🚫 智能过滤** - 自动过滤文档、示例、测试文件，专注有效代码
5. **🔄 外部同步** - 支持向Gemini-Balancer和GPT-Load同步发现的密钥

### 🔮 待开发功能 (TODO)

- [ ] **💾 数据库支持保存key** - 支持将发现的API密钥持久化存储到数据库中
- [ ] **📊 API、可视化展示抓取的key列表** - 提供API接口和可视化界面获取已抓取的密钥列表
- [ ] **💰 付费key检测** - 额外check下付费key

## 📋 项目结构 🗂️

```
hajimi-king-go/
├── cmd/app/                    # 应用程序入口
│   └── main.go                  # 主程序文件
├── internal/
│   ├── config/                 # 配置管理
│   │   └── config.go           # 配置加载和管理
│   ├── logger/                 # 日志管理
│   │   └── logger.go           # 日志记录器
│   ├── github/                 # GitHub客户端
│   │   └── client.go           # GitHub API客户端
│   ├── filemanager/            # 文件管理器
│   │   └── manager.go          # 文件操作和检查点管理
│   ├── syncutils/              # 同步工具
│   │   └── sync.go             # 外部服务同步
│   └── models/                 # 数据模型
│       └── models.go           # 数据结构定义
├── go.mod                      # Go模块文件
├── go.sum                      # 依赖校验文件
├── .env.example                # 环境变量示例
├── queries.example             # 查询配置示例
└── README.md                   # 项目文档
```

## 🖥️ 本地部署 🚀

### 1. 环境准备 🔧

```bash
# 确保已安装Go 1.21+
go version

# 克隆项目
git clone <repository-url>
cd hajimi-king-go

# 下载依赖
go mod tidy
```

### 2. 项目设置 📁

```bash
# 复制配置文件
cp .env.example .env

# 复制查询文件
cp queries.example queries.txt
```

### 3. 配置环境变量 🔑

编辑 `.env` 文件，**必须**配置GitHub Token：

```bash
# 必填：GitHub访问令牌
GITHUB_TOKENS=ghp1,ghp2,ghp3

# 可选：其他配置保持默认值即可
```

> 💡 **获取GitHub Token**：访问 [GitHub Settings > Tokens](https://github.com/settings/tokens)，创建具有 `public_repo` 权限的访问令牌 🎫

### 4. 运行程序 ⚡

```bash
# 创建数据目录
mkdir -p data

# 运行程序
go run cmd/app/main.go

# 或者编译后运行
go build -o hajimi-king cmd/app/main.go
./hajimi-king
```

### 5. 本地运行管理 🎮

```bash
# 查看日志文件
tail -f data/logs/keys_valid_detail_*.log

# 查看找到的有效密钥
cat data/keys/keys_valid_*.txt

# 停止程序
Ctrl + C
```

## ⚙️ 配置变量说明 📖

以下是所有可配置的环境变量，在 `.env` 文件中设置：

### 🔴 必填配置 ⚠️

| 变量名 | 说明 | 示例值 |
|--------|------|--------|
| `GITHUB_TOKENS` | GitHub API访问令牌，多个用逗号分隔 🎫 | `ghp_token1,ghp_token2` |

### 🟡 重要配置（建议了解）🤓

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PROXY` | 空 | 代理服务器地址，支持多个（逗号分隔）和账密认证 🌐 |
| `DATA_PATH` | `./data` | 数据存储目录路径 📂 |
| `DATE_RANGE_DAYS` | `730` | 仓库年龄过滤（天数），只扫描指定天数内的仓库 📅 |
| `QUERIES_FILE` | `queries.txt` | 搜索查询配置文件路径 🎯 |
| `HAJIMI_CHECK_MODEL` | `gemini-2.5-flash` | 用于验证key有效的模型 🤖 |
| `GEMINI_BALANCER_SYNC_ENABLED` | `false` | 是否启用Gemini Balancer同步 🔗 |
| `GEMINI_BALANCER_URL` | 空 | Gemini Balancer服务地址 🌐 |
| `GEMINI_BALANCER_AUTH` | 空 | Gemini Balancer认证信息 🔐 |
| `GPT_LOAD_SYNC_ENABLED` | `false` | 是否启用GPT Load Balancer同步 🔗 |
| `GPT_LOAD_URL` | 空 | GPT Load 服务地址 🌐 |
| `GPT_LOAD_AUTH` | 空 | GPT Load 认证Token 🔐 |
| `GPT_LOAD_GROUP_NAME` | 空 | GPT Load 组名，多个用逗号分隔 👥 |

### 🟢 可选配置（不懂就别动）😅

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `VALID_KEY_PREFIX` | `keys/keys_valid_` | 有效密钥文件名前缀 🗝️ |
| `RATE_LIMITED_KEY_PREFIX` | `keys/key_429_` | 频率限制密钥文件名前缀 ⏰ |
| `KEYS_SEND_PREFIX` | `keys/keys_send_` | 发送到外部应用的密钥文件名前缀 🚀 |
| `VALID_KEY_DETAIL_PREFIX` | `logs/keys_valid_detail_` | 详细日志文件名前缀 📝 |
| `RATE_LIMITED_KEY_DETAIL_PREFIX` | `logs/key_429_detail_` | 频率限制详细日志文件名前缀 📊 |
| `SCANNED_SHAS_FILE` | `scanned_shas.txt` | 已扫描文件SHA记录文件名 📋 |
| `FILE_PATH_BLACKLIST` | `readme,docs,...` | 文件路径黑名单，逗号分隔 🚫 |

### 配置文件示例 💫

完整的 `.env` 文件示例：

```bash
# 必填配置
GITHUB_TOKENS=ghp_your_token_here_1,ghp_your_token_here_2

# 重要配置（可选修改）
DATA_PATH=./data
DATE_RANGE_DAYS=730
QUERIES_FILE=queries.txt
HAJIMI_CHECK_MODEL=gemini-2.5-flash
PROXY=

# Gemini Balancer同步配置
GEMINI_BALANCER_SYNC_ENABLED=false
GEMINI_BALANCER_URL=
GEMINI_BALANCER_AUTH=

# GPT Load Balancer同步配置
GPT_LOAD_SYNC_ENABLED=false
GPT_LOAD_URL=
GPT_LOAD_AUTH=
GPT_LOAD_GROUP_NAME=group1,group2,group3

# 高级配置（建议保持默认）
VALID_KEY_PREFIX=keys/keys_valid_
RATE_LIMITED_KEY_PREFIX=keys/key_429_
KEYS_SEND_PREFIX=keys/keys_send_
VALID_KEY_DETAIL_PREFIX=logs/keys_valid_detail_
RATE_LIMITED_KEY_DETAIL_PREFIX=logs/key_429_detail_
KEYS_SEND_DETAIL_PREFIX=logs/keys_send_detail_
SCANNED_SHAS_FILE=scanned_shas.txt
FILE_PATH_BLACKLIST=readme,docs,doc/,.md,example,sample,tutorial,test,spec,demo,mock
```

### 查询配置文件 🔍

编辑 `queries.txt` 文件自定义搜索规则：

⚠️ **重要提醒**：query 是本项目的核心！好的表达式可以让搜索更高效，需要发挥自己的想象力！🧠💡

```bash
# GitHub搜索查询配置文件
# 每行一个查询语句，支持GitHub搜索语法
# 以#开头的行为注释，空行会被忽略

# 基础搜索
AIzaSy in:file
AizaSy in:file filename:.env
```

> 📖 **搜索语法参考**：[GitHub Code Search Syntax](https://docs.github.com/en/search-github/searching-on-github/searching-code) 📚  
> 🎯 **核心提示**：创造性的查询表达式是成功的关键，多尝试不同的组合！

## 🔒 安全注意事项 🛡️

- ✅ GitHub Token权限最小化（只需`public_repo`读取权限）🔐
- ✅ 定期轮换GitHub Token 🔄
- ✅ 不要将真实的API密钥提交到版本控制 🙈
- ✅ 定期检查和清理发现的密钥文件 🧹

## 🐳 Docker部署 🌊

### 方式一：使用环境变量

```yaml
version: '3.8'
services:
  hajimi-king-go:
    build: .
    container_name: hajimi-king-go
    restart: unless-stopped
    environment:
      # 必填：GitHub访问令牌
      - GITHUB_TOKENS=ghp_your_token_here_1,ghp_your_token_here_2
      # 可选配置
      - HAJIMI_CHECK_MODEL=gemini-2.5-flash
      - QUERIES_FILE=queries.txt
    volumes:
      - ./data:/app/data
    working_dir: /app
```

### 方式二：使用.env文件

```yaml
version: '3.8'
services:
  hajimi-king-go:
    build: .
    container_name: hajimi-king-go
    restart: unless-stopped
    env_file:
      - .env
    volumes:
      - ./data:/app/data
    working_dir: /app
```

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o hajimi-king cmd/app/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/hajimi-king .
COPY --from=builder /app/queries.example .
COPY --from=builder /app/.env.example .

EXPOSE 8080
CMD ["./hajimi-king"]
```

## 📊 性能优势

相比Python版本，Go版本具有以下优势：

- **⚡ 更高的性能** - Go的编译型语言特性提供更好的执行效率
- **🔄 原生并发** - Go的goroutine提供更高效的并发处理
- **📦 单文件部署** - 编译后的单个二进制文件，无需依赖
- **🧠 内存管理** - 更好的内存使用效率
- **🌐 跨平台** - 轻松编译到不同平台

## 🤝 贡献

欢迎提交Issue和Pull Request来帮助改进项目！

## 📄 许可证

本项目采用MIT许可证，详见LICENSE文件。

---

💖 **享受使用 Hajimi King Go 的快乐时光！** 🎉✨🎊