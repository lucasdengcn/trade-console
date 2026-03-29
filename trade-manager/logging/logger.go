package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger 日志记录器
type Logger struct {
	fileLogger *log.Logger
	stdLogger  *log.Logger
	level      LogLevel
	file       *os.File
	mu         sync.Mutex
}

// Config 日志配置
type Config struct {
	Level      string `json:"level"`       // 日志级别: debug, info, warn, error, fatal
	FilePath   string `json:"file_path"`  // 日志文件路径
	MaxSize    int64  `json:"max_size"`    // 最大文件大小 (MB)
	MaxBackups int    `json:"max_backups"` // 最大备份文件数
	MaxAge     int    `json:"max_age"`    // 最大保存天数
	Compress   bool   `json:"compress"`   // 是否压缩备份文件
}

// NewLogger 创建新的日志记录器
func NewLogger(config Config) (*Logger, error) {
	logger := &Logger{
		level: parseLogLevel(config.Level),
	}
	
	// 设置标准输出日志
	logger.stdLogger = log.New(os.Stdout, "", log.LstdFlags)
	
	// 设置文件日志
	if config.FilePath != "" {
		if err := logger.setupFileLogger(config); err != nil {
			return nil, err
		}
	}
	
	return logger, nil
}

// setupFileLogger 设置文件日志记录器
func (l *Logger) setupFileLogger(config Config) error {
	// 确保日志目录存在
	dir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}
	
	// 打开或创建日志文件
	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}
	
	l.file = file
	l.fileLogger = log.New(io.MultiWriter(file, os.Stdout), "", log.LstdFlags)
	
	return nil
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "warn":
		return WARN
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return INFO
	}
}

// shouldLog 检查是否应该记录该级别的日志
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// Debug 记录调试日志
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.shouldLog(DEBUG) {
		l.log("DEBUG", format, v...)
	}
}

// Info 记录信息日志
func (l *Logger) Info(format string, v ...interface{}) {
	if l.shouldLog(INFO) {
		l.log("INFO", format, v...)
	}
}

// Warn 记录警告日志
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.shouldLog(WARN) {
		l.log("WARN", format, v...)
	}
}

// Error 记录错误日志
func (l *Logger) Error(format string, v ...interface{}) {
	if l.shouldLog(ERROR) {
		l.log("ERROR", format, v...)
	}
}

// Fatal 记录致命错误日志并退出
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log("FATAL", format, v...)
	os.Exit(1)
}

// log 内部日志记录方法
func (l *Logger) log(level, format string, v ...interface{}) {
	message := fmt.Sprintf("[%s] %s", level, fmt.Sprintf(format, v...))
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.fileLogger != nil {
		l.fileLogger.Println(message)
	} else {
		l.stdLogger.Println(message)
	}
}

// WithFields 创建带字段的日志记录器（简化版）
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	return l // 简化实现，实际可以使用更复杂的结构化日志
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Rotate 轮转日志文件
func (l *Logger) Rotate() error {
	if l.file == nil {
		return nil
	}
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// 关闭当前文件
	if err := l.file.Close(); err != nil {
		return err
	}
	
	// 重命名当前日志文件
	oldPath := l.file.Name()
	newPath := fmt.Sprintf("%s.%s", oldPath, time.Now().Format("20060102_150405"))
	
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	
	// 创建新的日志文件
	file, err := os.OpenFile(oldPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	
	l.file = file
	l.fileLogger = log.New(io.MultiWriter(file, os.Stdout), "", log.LstdFlags)
	
	return nil
}

// Global logger instance
var (
	globalLogger *Logger
	once         sync.Once
)

// GetLogger 获取全局日志记录器
func GetLogger() *Logger {
	once.Do(func() {
		// 默认配置
		config := Config{
			Level:    "info",
			FilePath: "logs/trade-manager.log",
			MaxSize:  100, // 100MB
			MaxAge:   30,  // 30天
		}
		
		var err error
		globalLogger, err = NewLogger(config)
		if err != nil {
			// 如果文件日志失败，使用标准输出
			globalLogger = &Logger{
				stdLogger: log.New(os.Stdout, "", log.LstdFlags),
				level:     INFO,
			}
			globalLogger.Warn("文件日志初始化失败，使用标准输出: %v", err)
		}
	})
	return globalLogger
}

// SetGlobalLogger 设置全局日志记录器
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}