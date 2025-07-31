package logger

import (
	"log"
	"os"
	"time"
)

// Logger 日志结构体
type Logger struct {
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
}

var globalLogger *Logger

// InitLogger 初始化日志
func InitLogger() *Logger {
	if globalLogger != nil {
		return globalLogger
	}

	logger := &Logger{
		infoLogger:    log.New(os.Stdout, "ℹ️  ", log.LstdFlags),
		warningLogger: log.New(os.Stdout, "⚠️  ", log.LstdFlags),
		errorLogger:   log.New(os.Stdout, "❌ ", log.LstdFlags),
	}

	globalLogger = logger
	return logger
}

// GetLogger 获取全局日志实例
func GetLogger() *Logger {
	if globalLogger == nil {
		return InitLogger()
	}
	return globalLogger
}

// Info 输出信息日志
func (l *Logger) Info(msg string) {
	l.infoLogger.Println(msg)
}

// Infof 输出格式化信息日志
func (l *Logger) Infof(format string, args ...interface{}) {
	l.infoLogger.Printf(format, args...)
}

// Warning 输出警告日志
func (l *Logger) Warning(msg string) {
	l.warningLogger.Println(msg)
}

// Warningf 输出格式化警告日志
func (l *Logger) Warningf(format string, args ...interface{}) {
	l.warningLogger.Printf(format, args...)
}

// Error 输出错误日志
func (l *Logger) Error(msg string) {
	l.errorLogger.Println(msg)
}

// Errorf 输出格式化错误日志
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.errorLogger.Printf(format, args...)
}

// LogSystemStartup 记录系统启动信息
func (l *Logger) LogSystemStartup() {
	l.Info("============================================================")
	l.Info("🚀 HAJIMI KING STARTING")
	l.Info("============================================================")
	l.Infof("⏰ Started at: %s", time.Now().Format("2006-01-02 15:04:05"))
}

// LogSystemReady 记录系统就绪信息
func (l *Logger) LogSystemReady() {
	l.Info("✅ System ready - Starting king")
	l.Info("============================================================")
}

// LogSystemShutdown 记录系统关闭信息
func (l *Logger) LogSystemShutdown(validKeys, rateLimitedKeys int) {
	l.Info("⛔ Interrupted by user")
	l.Infof("📊 Final stats - Valid keys: %d, Rate limited: %d", validKeys, rateLimitedKeys)
	l.Info("🔚 Shutting down...")
}

// LogLoopStart 记录循环开始
func (l *Logger) LogLoopStart(loopNumber int) {
	l.Infof("🔄 Loop #%d - %s", loopNumber, time.Now().Format("15:04:05"))
}

// LogLoopComplete 记录循环完成
func (l *Logger) LogLoopComplete(loopNumber, processedFiles, validKeys, rateLimitedKeys int) {
	l.Infof("🏁 Loop #%d complete - Processed %d files | Total valid: %d | Total rate limited: %d", 
		loopNumber, processedFiles, validKeys, rateLimitedKeys)
}

// LogQueryProgress 记录查询进度
func (l *Logger) LogQueryProgress(queryIndex, totalQueries int, processed, valid, rateLimited int) {
	l.Infof("✅ Query %d/%d complete - Processed: %d, Valid: +%d, Rate limited: +%d", 
		queryIndex, totalQueries, processed, valid, rateLimited)
}

// LogSkipStats 记录跳过统计
func (l *Logger) LogSkipStats(stats map[string]int) {
	totalSkipped := 0
	for _, count := range stats {
		totalSkipped += count
	}
	if totalSkipped > 0 {
		l.Infof("📊 Skipped %d items - Time: %d, Duplicate: %d, Age: %d, Docs: %d", 
			totalSkipped, stats["time_filter"], stats["sha_duplicate"], stats["age_filter"], stats["doc_filter"])
	}
}