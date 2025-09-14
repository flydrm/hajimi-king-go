# OpenRouter API密钥验证集成说明

## 🎯 概述

本项目已成功集成OpenRouter API密钥验证功能，在保持原有代码架构不变的情况下，将验证目标从Gemini API切换到OpenRouter API。现在项目默认使用OpenRouter进行API密钥验证，同时保留对Gemini API的向后兼容支持。

## 🔧 主要变更

### 1. 新增OpenRouter客户端模块
- **文件位置**: `internal/openrouter/client.go`
- **功能**: 提供OpenRouter API的完整客户端实现
- **特性**:
  - API密钥验证
  - 模型列表获取
  - 聊天完成测试
  - 代理支持
  - 错误处理

### 2. 更新配置系统
- **新增配置项**:
  - `VALIDATION_PROVIDER`: 验证提供商（openrouter/gemini）
  - `OPENROUTER_CHECK_MODEL`: OpenRouter验证模型
- **默认值**:
  - `VALIDATION_PROVIDER=openrouter`
  - `OPENROUTER_CHECK_MODEL=openai/gpt-3.5-turbo`

### 3. 更新验证逻辑
- **统一验证接口**: `validateAPIKey()` 方法
- **提供商选择**: 根据配置自动选择验证方式
- **向后兼容**: 保留Gemini验证方法

### 4. 更新搜索查询
- **新增查询文件**: `queries.openrouter.txt`
- **更新默认查询**: `queries.txt` 现在主要搜索OpenRouter密钥
- **正则表达式**: 支持 `sk-or-` 格式的OpenRouter密钥

## 🚀 使用方法

### 1. 基本配置
```bash
# 设置验证提供商为OpenRouter（默认）
VALIDATION_PROVIDER=openrouter

# 设置OpenRouter验证模型
OPENROUTER_CHECK_MODEL=openai/gpt-3.5-turbo

# 其他配置保持不变
GITHUB_TOKENS=ghp_your_token_here
API_ENABLED=true
API_AUTH_KEY=your_secure_key
```

### 2. 运行程序
```bash
# 编译并运行
go build -o hajimi-king cmd/app/main.go
./hajimi-king

# 或直接运行
go run cmd/app/main.go
```

### 3. 测试OpenRouter集成
```bash
# 使用测试脚本验证OpenRouter API密钥
go run test_openrouter.go sk-or-your-api-key-here
```

## 🔍 搜索功能

### OpenRouter密钥格式
- **标准格式**: `sk-or-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`
- **长度**: 48个字符（不包括前缀）
- **字符集**: 字母、数字、连字符、下划线

### 搜索查询示例
```bash
# 基础搜索
"sk-or-" in:file

# 环境变量搜索
filename:*.env "OPENROUTER_API_KEY"

# 配置文件搜索
extension:json "openrouter_api_key"

# 代码文件搜索
extension:py "openrouter" AND "api_key"
```

## ⚙️ 配置选项

### 验证提供商配置
| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `VALIDATION_PROVIDER` | `openrouter` | 验证提供商：openrouter 或 gemini |
| `OPENROUTER_CHECK_MODEL` | `openai/gpt-3.5-turbo` | OpenRouter验证模型 |
| `HAJIMI_CHECK_MODEL` | `gemini-2.5-flash` | Gemini验证模型（向后兼容） |

### 环境变量示例
```bash
# OpenRouter配置
VALIDATION_PROVIDER=openrouter
OPENROUTER_CHECK_MODEL=openai/gpt-3.5-turbo

# 或使用Gemini（向后兼容）
VALIDATION_PROVIDER=gemini
HAJIMI_CHECK_MODEL=gemini-2.5-flash
```

## 🔄 验证流程

### OpenRouter验证流程
1. **模型列表验证**: 调用 `/v1/models` 端点验证API密钥
2. **聊天测试**: 使用 `/v1/chat/completions` 进行完整测试
3. **错误处理**: 识别各种错误类型（未授权、限流、禁用等）

### 错误类型
- `ok`: 密钥有效
- `not_authorized_key`: 密钥无效或未授权
- `rate_limited`: 请求被限流
- `forbidden`: 访问被禁止
- `error:具体错误信息`: 其他错误

## 📊 性能优化

### 1. 验证策略
- **双重验证**: 先验证模型列表，再测试聊天功能
- **智能降级**: 如果聊天测试失败但模型列表验证成功，仍认为密钥有效
- **错误缓存**: 避免重复验证无效密钥

### 2. 网络优化
- **代理支持**: 支持HTTP/SOCKS5代理
- **超时控制**: 30秒请求超时
- **重试机制**: 自动重试失败的请求

## 🔒 安全特性

### 1. 密钥保护
- **本地验证**: 密钥不会发送到第三方服务
- **安全存储**: 验证结果本地存储
- **日志脱敏**: 敏感信息在日志中脱敏

### 2. 访问控制
- **API认证**: 支持JWT Token认证
- **权限验证**: 多级权限控制
- **安全传输**: 支持HTTPS代理

## 🐛 故障排除

### 常见问题

#### 1. OpenRouter API密钥无效
```bash
# 检查密钥格式
echo "sk-or-your-key" | grep -E "^sk-or-[A-Za-z0-9\-_]{48}$"

# 测试API密钥
go run test_openrouter.go sk-or-your-key
```

#### 2. 验证失败
```bash
# 检查网络连接
curl -H "Authorization: Bearer sk-or-your-key" https://openrouter.ai/api/v1/models

# 检查代理配置
PROXY=http://your-proxy:port
```

#### 3. 模型不可用
```bash
# 检查可用模型
go run test_openrouter.go sk-or-your-key

# 更换验证模型
OPENROUTER_CHECK_MODEL=anthropic/claude-3-haiku
```

## 📈 监控和统计

### 1. 验证统计
- **有效密钥数**: 成功验证的OpenRouter密钥数量
- **限流密钥数**: 被限流的密钥数量
- **无效密钥数**: 验证失败的密钥数量

### 2. 性能指标
- **验证速度**: 平均验证时间
- **成功率**: 验证成功率统计
- **错误率**: 各类错误的发生频率

## 🔮 未来扩展

### 1. 计划功能
- **多模型支持**: 支持更多OpenRouter模型
- **批量验证**: 提高验证效率
- **智能重试**: 更智能的重试策略

### 2. 集成扩展
- **更多提供商**: 支持其他AI服务提供商
- **统一接口**: 提供统一的验证接口
- **插件系统**: 支持自定义验证插件

## 📝 总结

OpenRouter集成已成功完成，项目现在可以：

1. **搜索OpenRouter API密钥**: 使用专门的查询模式
2. **验证OpenRouter密钥**: 通过OpenRouter API进行验证
3. **保持向后兼容**: 仍支持Gemini API验证
4. **提供统一接口**: 透明的验证提供商切换

项目架构保持不变，所有现有功能继续正常工作，同时新增了强大的OpenRouter支持。用户可以通过简单的配置切换验证提供商，享受更灵活的API密钥管理体验。