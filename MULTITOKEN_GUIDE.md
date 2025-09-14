# Hajimi King Go v2.0 - 多Token配置指南

## 🚀 概述

Hajimi King Go v2.0 现在支持配置多个GitHub Personal Access Token，实现负载均衡、故障转移和自动重试功能。

## 🔧 配置方式

### 1. 单个Token配置（向后兼容）

```bash
# .env 文件
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### 2. 多个Token配置（推荐）

```bash
# .env 文件
GITHUB_TOKENS=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx,ghp_yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy,ghp_zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz
```

### 3. 混合配置（系统会自动合并）

```bash
# .env 文件
GITHUB_TOKEN=ghp_primary_token
GITHUB_TOKENS=ghp_backup1,ghp_backup2,ghp_backup3
```

## ⚡ 核心功能

### 🔄 自动轮换
- 系统会自动轮换使用配置的Token
- 每次API调用都会使用下一个可用的Token
- 实现负载均衡，避免单个Token过载

### 🛡️ 故障转移
- 当Token失效（401/403错误）时，自动切换到下一个Token
- 当Token被限流（429错误）时，自动切换到下一个Token
- 支持最多3次重试

### 🚫 黑名单机制
- 失效的Token会被加入黑名单
- 黑名单TTL为5分钟，过期后自动恢复
- 避免重复使用已知失效的Token

### 📊 状态监控
- 实时监控Token使用状态
- 显示可用Token数量和黑名单数量
- 支持通过API查看Token状态

## 🎯 使用场景

### 高并发场景
```bash
# 配置多个Token分散负载
GITHUB_TOKENS=token1,token2,token3,token4,token5
```

### 故障容错场景
```bash
# 主Token + 备用Token
GITHUB_TOKEN=ghp_primary_token
GITHUB_TOKENS=ghp_backup1,ghp_backup2
```

### 团队协作场景
```bash
# 每个团队成员一个Token
GITHUB_TOKENS=ghp_member1,token,ghp_member2,token,ghp_member3,token
```

## 📋 配置示例

### 最小配置
```bash
# 单个Token
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
PLATFORM_GEMINI_ENABLED=true
```

### 推荐配置
```bash
# 多个Token + 所有平台
GITHUB_TOKENS=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx,ghp_yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy,ghp_zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz
PLATFORM_GEMINI_ENABLED=true
PLATFORM_OPENROUTER_ENABLED=true
PLATFORM_SILICONFLOW_ENABLED=true
```

### 完整配置
```bash
# GitHub配置
GITHUB_TOKENS=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx,ghp_yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy,ghp_zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz
GITHUB_PROXY=
GITHUB_BASE_URL=https://api.github.com

# 平台配置
PLATFORM_GLOBAL_ENABLED=true
PLATFORM_EXECUTION_MODE=all
PLATFORM_GEMINI_ENABLED=true
PLATFORM_OPENROUTER_ENABLED=true
PLATFORM_SILICONFLOW_ENABLED=true

# 性能配置
WORKER_POOL_SIZE=8
MAX_CONCURRENT_FILES=10
```

## 🔍 监控和调试

### Token状态查看
```bash
# 通过API查看Token状态
curl http://localhost:8080/api/token-status
```

### 日志监控
```bash
# 查看Token切换日志
tail -f logs/hajimi-king.log | grep "Token"
```

### 状态指标
- `total_tokens`: 总Token数量
- `available_tokens`: 可用Token数量
- `blacklisted`: 黑名单Token数量
- `current_index`: 当前Token索引

## ⚠️ 注意事项

### Token权限
确保所有Token都有相同的权限：
- `public_repo` - 访问公共仓库
- `repo` - 访问私有仓库（如果需要）

### Token格式
所有Token必须符合GitHub格式：
- 以 `ghp_` 开头
- 长度为40个字符
- 用逗号分隔多个Token

### 性能考虑
- 建议配置3-5个Token
- 过多Token可能影响性能
- 定期检查Token有效性

## 🚀 最佳实践

### 1. Token管理
- 定期轮换Token
- 监控Token使用情况
- 及时移除失效Token

### 2. 配置优化
- 根据并发量调整Token数量
- 使用代理分散请求
- 配置合适的重试策略

### 3. 监控告警
- 设置Token失效告警
- 监控API调用频率
- 跟踪错误率变化

## 🎉 总结

多Token功能为Hajimi King Go v2.0提供了：
- ✅ **高可用性**: 故障自动转移
- ✅ **负载均衡**: 自动轮换使用
- ✅ **容错能力**: 黑名单和重试机制
- ✅ **监控能力**: 实时状态查看
- ✅ **向后兼容**: 支持单Token配置

现在您可以配置多个GitHub Token，让系统更加稳定可靠！

---
**Hajimi King Go v2.0** - 让API密钥发现变得简单高效！ 🔑✨