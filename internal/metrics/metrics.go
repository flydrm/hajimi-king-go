package metrics

import (
	"sync"
	"time"
)

// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	// Processing metrics
	ProcessedFiles   int64
	ProcessedKeys    int64
	ValidKeys        int64
	RateLimitedKeys  int64
	
	// Performance metrics
	ThroughputKeysPerSecond float64
	AverageResponseTime     float64
	MemoryUsageMB          float64
	
	// Cache metrics
	CacheHitRate           float64
	CacheMissRate          float64
	
	// Detection metrics
	DetectionRate          float64
	FalsePositiveRate      float64
	
	// Worker pool metrics
	ActiveWorkers          int
	QueueSize              int
	TasksCompleted         int64
	TasksFailed            int64
	
	// Platform metrics
	PlatformMetrics        map[string]PlatformMetrics
	
	// Timestamps
	StartTime              time.Time
	LastUpdateTime         time.Time
	
	mutex                  sync.RWMutex
}

// PlatformMetrics represents metrics for a specific platform
type PlatformMetrics struct {
	PlatformName    string
	KeysFound       int64
	ValidKeys       int64
	InvalidKeys     int64
	RateLimitedKeys int64
	AverageResponseTime float64
	SuccessRate     float64
	LastProcessed   time.Time
}

// NewSystemMetrics creates a new system metrics instance
func NewSystemMetrics() *SystemMetrics {
	return &SystemMetrics{
		PlatformMetrics: make(map[string]PlatformMetrics),
		StartTime:       time.Now(),
		LastUpdateTime:  time.Now(),
	}
}

// IncrementProcessedFiles increments the processed files counter
func (sm *SystemMetrics) IncrementProcessedFiles() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.ProcessedFiles++
	sm.LastUpdateTime = time.Now()
}

// IncrementProcessedKeys increments the processed keys counter
func (sm *SystemMetrics) IncrementProcessedKeys() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.ProcessedKeys++
	sm.LastUpdateTime = time.Now()
}

// IncrementValidKeys increments the valid keys counter
func (sm *SystemMetrics) IncrementValidKeys() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.ValidKeys++
	sm.LastUpdateTime = time.Now()
}

// IncrementRateLimitedKeys increments the rate limited keys counter
func (sm *SystemMetrics) IncrementRateLimitedKeys() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.RateLimitedKeys++
	sm.LastUpdateTime = time.Now()
}

// UpdateThroughput updates the throughput metric
func (sm *SystemMetrics) UpdateThroughput(keysPerSecond float64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.ThroughputKeysPerSecond = keysPerSecond
	sm.LastUpdateTime = time.Now()
}

// UpdateAverageResponseTime updates the average response time
func (sm *SystemMetrics) UpdateAverageResponseTime(responseTime float64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.AverageResponseTime = responseTime
	sm.LastUpdateTime = time.Now()
}

// UpdateMemoryUsage updates the memory usage metric
func (sm *SystemMetrics) UpdateMemoryUsage(memoryMB float64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.MemoryUsageMB = memoryMB
	sm.LastUpdateTime = time.Now()
}

// UpdateCacheMetrics updates cache metrics
func (sm *SystemMetrics) UpdateCacheMetrics(hitRate, missRate float64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.CacheHitRate = hitRate
	sm.CacheMissRate = missRate
	sm.LastUpdateTime = time.Now()
}

// UpdateDetectionMetrics updates detection metrics
func (sm *SystemMetrics) UpdateDetectionMetrics(detectionRate, falsePositiveRate float64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.DetectionRate = detectionRate
	sm.FalsePositiveRate = falsePositiveRate
	sm.LastUpdateTime = time.Now()
}

// UpdateWorkerPoolMetrics updates worker pool metrics
func (sm *SystemMetrics) UpdateWorkerPoolMetrics(activeWorkers, queueSize int, tasksCompleted, tasksFailed int64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.ActiveWorkers = activeWorkers
	sm.QueueSize = queueSize
	sm.TasksCompleted = tasksCompleted
	sm.TasksFailed = tasksFailed
	sm.LastUpdateTime = time.Now()
}

// UpdatePlatformMetrics updates platform-specific metrics
func (sm *SystemMetrics) UpdatePlatformMetrics(platform string, keysFound, validKeys, invalidKeys, rateLimitedKeys int64, avgResponseTime, successRate float64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sm.PlatformMetrics[platform] = PlatformMetrics{
		PlatformName:        platform,
		KeysFound:           keysFound,
		ValidKeys:           validKeys,
		InvalidKeys:         invalidKeys,
		RateLimitedKeys:     rateLimitedKeys,
		AverageResponseTime: avgResponseTime,
		SuccessRate:         successRate,
		LastProcessed:       time.Now(),
	}
	sm.LastUpdateTime = time.Now()
}

// GetMetrics returns a copy of current metrics
func (sm *SystemMetrics) GetMetrics() SystemMetrics {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	// Create a deep copy
	platformMetricsCopy := make(map[string]PlatformMetrics)
	for k, v := range sm.PlatformMetrics {
		platformMetricsCopy[k] = v
	}
	
	return SystemMetrics{
		ProcessedFiles:         sm.ProcessedFiles,
		ProcessedKeys:          sm.ProcessedKeys,
		ValidKeys:              sm.ValidKeys,
		RateLimitedKeys:        sm.RateLimitedKeys,
		ThroughputKeysPerSecond: sm.ThroughputKeysPerSecond,
		AverageResponseTime:    sm.AverageResponseTime,
		MemoryUsageMB:          sm.MemoryUsageMB,
		CacheHitRate:           sm.CacheHitRate,
		CacheMissRate:          sm.CacheMissRate,
		DetectionRate:          sm.DetectionRate,
		FalsePositiveRate:      sm.FalsePositiveRate,
		ActiveWorkers:          sm.ActiveWorkers,
		QueueSize:              sm.QueueSize,
		TasksCompleted:         sm.TasksCompleted,
		TasksFailed:            sm.TasksFailed,
		PlatformMetrics:        platformMetricsCopy,
		StartTime:              sm.StartTime,
		LastUpdateTime:         sm.LastUpdateTime,
	}
}

// GetUptime returns the system uptime
func (sm *SystemMetrics) GetUptime() time.Duration {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return time.Since(sm.StartTime)
}

// GetTotalThroughput returns the total throughput since start
func (sm *SystemMetrics) GetTotalThroughput() float64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	uptime := time.Since(sm.StartTime).Seconds()
	if uptime == 0 {
		return 0
	}
	return float64(sm.ProcessedKeys) / uptime
}

// Reset resets all metrics
func (sm *SystemMetrics) Reset() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sm.ProcessedFiles = 0
	sm.ProcessedKeys = 0
	sm.ValidKeys = 0
	sm.RateLimitedKeys = 0
	sm.ThroughputKeysPerSecond = 0
	sm.AverageResponseTime = 0
	sm.MemoryUsageMB = 0
	sm.CacheHitRate = 0
	sm.CacheMissRate = 0
	sm.DetectionRate = 0
	sm.FalsePositiveRate = 0
	sm.ActiveWorkers = 0
	sm.QueueSize = 0
	sm.TasksCompleted = 0
	sm.TasksFailed = 0
	sm.PlatformMetrics = make(map[string]PlatformMetrics)
	sm.StartTime = time.Now()
	sm.LastUpdateTime = time.Now()
}