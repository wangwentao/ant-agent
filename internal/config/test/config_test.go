package config_test

import (
	"os"
	"testing"

	"ant-agent/internal/config"
)

func TestLoadConfig(t *testing.T) {
	// 测试默认配置
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
	}

	// 检查默认值
	if cfg.Model == "" {
		t.Error("LoadConfig() returned empty model")
	}
	if cfg.BaseURL == "" {
		t.Error("LoadConfig() returned empty base URL")
	}
	if cfg.MaxTokens <= 0 {
		t.Error("LoadConfig() returned invalid max tokens")
	}
}

func TestGetAPIKey(t *testing.T) {
	// 测试从配置文件获取API Key
	cfg := config.Config{
		APIKey: "test-api-key",
	}

	apiKey := config.GetAPIKey(cfg)
	if apiKey != "test-api-key" {
		t.Errorf("GetAPIKey() = %v, want %v", apiKey, "test-api-key")
	}

	// 测试从环境变量获取API Key
	os.Setenv("ANTHROPIC_API_KEY", "env-api-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	cfg = config.Config{
		APIKey: "",
	}

	apiKey = config.GetAPIKey(cfg)
	if apiKey != "env-api-key" {
		t.Errorf("GetAPIKey() = %v, want %v", apiKey, "env-api-key")
	}
}
