package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 配置结构体
type Config struct {
	APIKey    string `json:"api_key"`
	Model     string `json:"model"`
	BaseURL   string `json:"base_url"`
	MaxTokens int    `json:"max_tokens"`
	Name      string `json:"name"`
	LogLevel  string `json:"log_level"`
}

// LoadConfig 读取配置文件
func LoadConfig() (Config, error) {
	var config Config

	// 尝试在可执行文件所在目录读取配置文件
	execPath, err := os.Executable()
	if err != nil {
		// 如果无法获取可执行文件路径，尝试当前目录
		execPath = "."
	}
	execDir := filepath.Dir(execPath)
	configPath := filepath.Join(execDir, "config.json")

	// 尝试读取配置文件
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，返回默认配置
			return Config{
				Model:     "qwen3.5-27b-claude-4.6-opus-reasoning-distilled",
				BaseURL:   "http://127.0.0.1:1234",
				MaxTokens: 4096,
				Name:      "Ant",
			}, nil
		}
		return Config{}, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	// 解析配置文件
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, fmt.Errorf("error decoding config file: %w", err)
	}

	// 设置默认值
	if config.Model == "" {
		config.Model = "qwen3.5-27b-claude-4.6-opus-reasoning-distilled"
	}
	if config.BaseURL == "" {
		config.BaseURL = "http://127.0.0.1:1234"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}
	if config.Name == "" {
		config.Name = "Ant"
	}
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	return config, nil
}

// GetAPIKey 获取API Key，优先使用配置文件中的值
func GetAPIKey(config Config) string {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	return apiKey
}
