package logs_test

import (
	"testing"

	"ant-agent/internal/logs"
)

func TestLogLevel(t *testing.T) {
	// 测试默认日志级别
	if logs.GetLevel() != logs.INFO {
		t.Errorf("Expected default log level to be INFO, got %v", logs.GetLevel())
	}

	// 测试设置日志级别
	logs.SetLevel(logs.DEBUG)
	if logs.GetLevel() != logs.DEBUG {
		t.Errorf("Expected log level to be DEBUG, got %v", logs.GetLevel())
	}

	// 测试重置日志级别
	logs.SetLevel(logs.INFO)
	if logs.GetLevel() != logs.INFO {
		t.Errorf("Expected log level to be INFO, got %v", logs.GetLevel())
	}
}

func TestLogFunctions(t *testing.T) {
	// 测试所有日志函数是否能正常调用
	logs.Debug("Debug test message")
	logs.Info("Info test message")
	logs.Warn("Warn test message")
	logs.Error("Error test message")
	// 不测试Fatal，因为它会退出程序
}
