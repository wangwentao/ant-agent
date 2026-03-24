package logs

import (
	"fmt"
	"os"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	// DEBUG 调试级别
	DEBUG LogLevel = iota
	// INFO 信息级别
	INFO
	// WARN 警告级别
	WARN
	// ERROR 错误级别
	ERROR
	// FATAL 致命级别
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

var currentLevel = INFO

// SetLevel 设置日志级别
func SetLevel(level LogLevel) {
	currentLevel = level
}

// GetLevel 获取当前日志级别
func GetLevel() LogLevel {
	return currentLevel
}

// log 通用日志函数
func log(level LogLevel, format string, args ...interface{}) {
	if level < currentLevel {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := levelNames[level]
	message := fmt.Sprintf(format, args...)

	fmt.Printf("[%s] [%s] %s\n", timestamp, levelName, message)

	if level == FATAL {
		os.Exit(1)
	}
}

// Debug 调试日志
func Debug(format string, args ...interface{}) {
	log(DEBUG, format, args...)
}

// Info 信息日志
func Info(format string, args ...interface{}) {
	log(INFO, format, args...)
}

// Warn 警告日志
func Warn(format string, args ...interface{}) {
	log(WARN, format, args...)
}

// Error 错误日志
func Error(format string, args ...interface{}) {
	log(ERROR, format, args...)
}

// Fatal 致命日志
func Fatal(format string, args ...interface{}) {
	log(FATAL, format, args...)
}
