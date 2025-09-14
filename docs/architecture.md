# 架构设计文档 (Architecture Design Document)

## Hajimi King Go v2.0 - 多平台API密钥发现系统架构设计文档

### 1. 架构概述

#### 1.1 系统架构图

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
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              数据存储层                                          │
├─────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │ 文件系统     │  │ 内存缓存     │  │ Redis缓存   │  │ 检查点文件   │            │
│  │ 密钥文件     │  │ L1 Cache    │  │ L3 Cache    │  │ 增量扫描     │            │
│  │ 日志文件     │  │ LRU算法     │  │ 可选启用    │  │ 状态保存     │            │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────────────────────┘
```

#### 1.2 核心设计原则

1. **模块化设计**: 各组件职责单一，松耦合
2. **可扩展性**: 支持新增平台和功能扩展
3. **高性能**: 并发处理，缓存优化
4. **可配置性**: 配置驱动，环境变量控制
5. **可观测性**: 完整的指标和日志系统

### 2. 核心组件设计

#### 2.1 并发处理组件 (WorkerPool)

**职责**: 管理并发任务执行，提供任务队列和结果处理

**核心接口**:
```go
type WorkerPool interface {
    SubmitTask(task Task) error
    GetResult() <-chan Result
    Start() error
    Stop() error
    GetMetrics() *PoolMetrics
}

type Task interface {
    Execute() Result
    GetID() string
    GetPriority() int
}
```

**设计特点**:
- 基于Goroutine的Worker Pool模式
- 支持任务优先级调度
- 提供完整的性能指标
- 支持优雅关闭

#### 2.2 多级缓存组件 (MultiLevelCache)

**职责**: 提供多级缓存服务，减少重复计算和网络请求

**缓存层级**:
- **L1 (内存缓存)**: LRU算法，5分钟TTL
- **L2 (文件缓存)**: 本地文件，1小时TTL  
- **L3 (Redis缓存)**: 分布式缓存，24小时TTL

**核心接口**:
```go
type MultiLevelCache interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration) error
    Delete(key string) error
    GetHitRate() float64
    GetMetrics() *CacheMetrics
}
```

**设计特点**:
- 自动回填机制
- 智能TTL管理
- 完整的缓存指标
- 支持缓存预热

#### 2.3 智能检测组件 (SmartKeyDetector)

**职责**: 智能识别和过滤API密钥，提供置信度评估

**核心功能**:
- 多平台正则表达式匹配
- 上下文分析和过滤
- 置信度评分系统
- 风险级别评估

**核心接口**:
```go
type SmartKeyDetector interface {
    DetectKeys(content string) []*KeyContext
    AnalyzeKeyContext(key, content string) *KeyContext
    GetDetectionRate() float64
    GetMetrics() *DetectionMetrics
}

type KeyContext struct {
    Key           string
    Confidence    float64
    IsPlaceholder bool
    IsTestKey     bool
    Platform      string
    RiskLevel     string
}
```

**设计特点**:
- 基于规则的智能过滤
- 上下文语义分析
- 可配置的置信度阈值
- 支持自定义模式扩展

### 3. 数据流设计

#### 3.1 主数据流

```
1. 配置加载 → 2. 平台选择 → 3. 并发任务分发 → 4. GitHub搜索
    ↓
5. 智能密钥检测 → 6. 并发验证 → 7. 结果缓存 → 8. 文件保存
    ↓
9. 外部同步 → 10. 指标更新 → 11. 日志记录
```

#### 3.2 平台处理流程

```
平台启动 → 查询配置加载 → 并发查询执行 → 结果聚合
    ↓
文件内容获取 → 智能密钥检测 → 并发验证 → 结果分类
    ↓
缓存更新 → 文件保存 → 外部同步 → 指标统计
```

### 4. 接口设计

#### 4.1 平台接口

```go
type Platform interface {
    GetName() string
    GetQueries() []string
    GetRegexPatterns() []*RegexPattern
    ValidateKey(key string) (*ValidationResult, error)
    GetConfig() *PlatformConfig
}
```

#### 4.2 缓存接口

```go
type Cache interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration) error
    Delete(key string) error
    Clear() error
    Size() int
}
```

#### 4.3 检测接口

```go
type Detector interface {
    DetectKeys(content string) []*KeyContext
    AnalyzeContext(key, content string) *KeyContext
    IsValidKey(context *KeyContext) bool
}
```

### 5. 配置设计

#### 5.1 平台配置

```json
{
  "platform_switches": {
    "global_enabled": true,
    "execution_mode": "all",
    "selected_platforms": ["gemini", "openrouter"],
    "platforms": {
      "gemini": {
        "enabled": true,
        "priority": 1,
        "description": "Google Gemini API平台"
      },
      "openrouter": {
        "enabled": false,
        "priority": 2,
        "description": "OpenRouter API平台"
      }
    }
  }
}
```

#### 5.2 缓存配置

```json
{
  "cache_config": {
    "l1_max_size": 1000,
    "l1_ttl": "5m",
    "l2_ttl": "1h",
    "l3_ttl": "24h",
    "enable_l3": false,
    "cleanup_interval": "10m"
  }
}
```

### 6. 性能设计

#### 6.1 并发策略

- **平台级并发**: 多个平台同时处理
- **查询级并发**: 单个平台内多个查询并发
- **文件级并发**: 多个文件并发处理
- **验证级并发**: 多个密钥并发验证

#### 6.2 缓存策略

- **查询结果缓存**: 5分钟TTL
- **验证结果缓存**: 1小时TTL
- **文件内容缓存**: 30分钟TTL
- **配置缓存**: 永久缓存

#### 6.3 内存管理

- **对象池**: 重用频繁创建的对象
- **流式处理**: 大文件流式读取
- **内存限制**: 设置最大内存使用量
- **垃圾回收**: 定期清理无用对象

### 7. 监控设计

#### 7.1 性能指标

```go
type SystemMetrics struct {
    // 处理指标
    ProcessedFiles   int64
    ProcessedKeys    int64
    ValidKeys        int64
    RateLimitedKeys  int64
    
    // 性能指标
    ThroughputKeysPerSecond float64
    AverageResponseTime     float64
    MemoryUsageMB          float64
    
    // 缓存指标
    CacheHitRate           float64
    CacheMissRate          float64
    
    // 检测指标
    DetectionRate          float64
    FalsePositiveRate      float64
}
```

#### 7.2 日志设计

- **结构化日志**: JSON格式
- **日志级别**: DEBUG, INFO, WARN, ERROR
- **日志轮转**: 按大小和时间轮转
- **敏感信息**: 自动脱敏处理

### 8. 安全设计

#### 8.1 数据安全

- **密钥加密**: 敏感信息加密存储
- **传输安全**: HTTPS/TLS加密传输
- **访问控制**: JWT认证和授权
- **审计日志**: 完整的操作记录

#### 8.2 系统安全

- **输入验证**: 严格的输入参数验证
- **错误处理**: 安全的错误信息返回
- **资源限制**: 防止资源耗尽攻击
- **定期更新**: 依赖包安全更新