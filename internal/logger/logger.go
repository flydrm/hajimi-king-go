package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger represents the application logger
type Logger struct {
	logger *logrus.Logger
	file   *os.File
}

// LogLevel represents the log level
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Caller      string                 `json:"caller,omitempty"`
	Platform    string                 `json:"platform,omitempty"`
	Key         string                 `json:"key,omitempty"`
	Repository  string                 `json:"repository,omitempty"`
	FilePath    string                 `json:"file_path,omitempty"`
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel, logFile string) (*Logger, error) {
	logger := logrus.New()
	
	// Set log level
	switch level {
	case DebugLevel:
		logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
	
	// Set formatter
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})
	
	// Set output
	var output io.Writer = os.Stdout
	if logFile != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(logFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		
		// Open log file
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		
		// Set output to both file and stdout
		output = io.MultiWriter(os.Stdout, file)
		logger.SetOutput(output)
		
		return &Logger{
			logger: logger,
			file:   file,
		}, nil
	}
	
	logger.SetOutput(output)
	
	return &Logger{
		logger: logger,
		file:   nil,
	}, nil
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	l.logWithCaller(logrus.DebugLevel, message, fields...)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	l.logWithCaller(logrus.InfoLevel, message, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	l.logWithCaller(logrus.WarnLevel, message, fields...)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	l.logWithCaller(logrus.ErrorLevel, message, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, fields ...map[string]interface{}) {
	l.logWithCaller(logrus.FatalLevel, message, fields...)
	os.Exit(1)
}

// logWithCaller logs a message with caller information
func (l *Logger) logWithCaller(level logrus.Level, message string, fields ...map[string]interface{}) {
	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	caller := ""
	if ok {
		caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}
	
	// Merge fields
	allFields := make(map[string]interface{})
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			allFields[k] = v
		}
	}
	
	// Add caller information
	if caller != "" {
		allFields["caller"] = caller
	}
	
	// Log with fields
	l.logger.WithFields(allFields).Log(level, message)
}

// LogKeyDiscovery logs a key discovery event
func (l *Logger) LogKeyDiscovery(platform, key, repository, filePath string, isValid bool, confidence float64) {
	fields := map[string]interface{}{
		"platform":   platform,
		"key":        l.maskKey(key),
		"repository": repository,
		"file_path":  filePath,
		"is_valid":   isValid,
		"confidence": confidence,
		"event_type": "key_discovery",
	}
	
	if isValid {
		l.Info("Valid key discovered", fields)
	} else {
		l.Info("Invalid key discovered", fields)
	}
}

// LogKeyValidation logs a key validation event
func (l *Logger) LogKeyValidation(platform, key string, isValid bool, errorMsg string) {
	fields := map[string]interface{}{
		"platform":   platform,
		"key":        l.maskKey(key),
		"is_valid":   isValid,
		"error":      errorMsg,
		"event_type": "key_validation",
	}
	
	if isValid {
		l.Info("Key validation successful", fields)
	} else {
		l.Warn("Key validation failed", fields)
	}
}

// LogPlatformEvent logs a platform-specific event
func (l *Logger) LogPlatformEvent(platform, eventType, message string, fields ...map[string]interface{}) {
	allFields := make(map[string]interface{})
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			allFields[k] = v
		}
	}
	
	allFields["platform"] = platform
	allFields["event_type"] = eventType
	
	l.Info(message, allFields)
}

// LogSystemEvent logs a system event
func (l *Logger) LogSystemEvent(eventType, message string, fields ...map[string]interface{}) {
	allFields := make(map[string]interface{})
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			allFields[k] = v
		}
	}
	
	allFields["event_type"] = "system"
	allFields["sub_type"] = eventType
	
	l.Info(message, allFields)
}

// LogPerformance logs a performance metric
func (l *Logger) LogPerformance(metric string, value float64, unit string, fields ...map[string]interface{}) {
	allFields := make(map[string]interface{})
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			allFields[k] = v
		}
	}
	
	allFields["metric"] = metric
	allFields["value"] = value
	allFields["unit"] = unit
	allFields["event_type"] = "performance"
	
	l.Info("Performance metric", allFields)
}

// maskKey masks sensitive parts of a key for logging
func (l *Logger) maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "***" + key[len(key)-4:]
}

// Close closes the logger and any open files
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// GetLogLevel returns the current log level
func (l *Logger) GetLogLevel() LogLevel {
	switch l.logger.GetLevel() {
	case logrus.DebugLevel:
		return DebugLevel
	case logrus.InfoLevel:
		return InfoLevel
	case logrus.WarnLevel:
		return WarnLevel
	case logrus.ErrorLevel:
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// SetLogLevel sets the log level
func (l *Logger) SetLogLevel(level LogLevel) {
	switch level {
	case DebugLevel:
		l.logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		l.logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		l.logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		l.logger.SetLevel(logrus.ErrorLevel)
	}
}