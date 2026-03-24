package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 配置结构体
type Config struct {
	APIKey    string `json:"api_key"`
	Model     string `json:"model"`
	BaseURL   string `json:"base_url"`
	MaxTokens int    `json:"max_tokens"`
}

// LoadConfig 读取配置文件
func LoadConfig() (Config, error) {
	var config Config

	// 尝试读取配置文件
	file, err := os.Open("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，返回默认配置
			return Config{
				Model:     "claude-3-opus-20240229",
				BaseURL:   "https://api.anthropic.com",
				MaxTokens: 1024,
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
		config.Model = "claude-3-opus-20240229"
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com"
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
