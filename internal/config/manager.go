package config

import (
	"os"
)

// AppConfig 应用配置结构体
type AppConfig struct {
	// 基础配置
	APIKey    string
	Model     string
	BaseURL   string
	MaxTokens int

	// 运行时配置
	Stream        bool
	OutputFormat  string
	SessionID     string
	SystemPrompt  string
	ACPMode       bool
	ACPConn       interface{}
	ACPSessionID  string

	// Agent 信息
	AgentName    string
	AgentVersion string
}

// DefaultConfig 返回默认配置
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Model:         "claude-3-opus-20240229",
		BaseURL:       "https://api.anthropic.com",
		MaxTokens:     1024,
		Stream:        true,
		AgentName:     "ant-agent",
		AgentVersion:  "1.0.0",
	}
}

// LoadAppConfig 加载应用配置
func LoadAppConfig() (*AppConfig, error) {
	// 加载基础配置
	baseConfig, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// 创建应用配置
	appConfig := DefaultConfig()
	appConfig.APIKey = baseConfig.APIKey
	appConfig.Model = baseConfig.Model
	appConfig.BaseURL = baseConfig.BaseURL
	appConfig.MaxTokens = baseConfig.MaxTokens

	return appConfig, nil
}

// GetAPIKey 获取API Key
func (c *AppConfig) GetAPIKey() string {
	if c.APIKey != "" {
		return c.APIKey
	}
	return os.Getenv("ANTHROPIC_API_KEY")
}

// ToMap 转换为 map[string]interface{}
func (c *AppConfig) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"api_key":         c.APIKey,
		"model":           c.Model,
		"base_url":        c.BaseURL,
		"max_tokens":      c.MaxTokens,
		"stream":          c.Stream,
		"output_format":   c.OutputFormat,
		"session_id":      c.SessionID,
		"system_prompt":   c.SystemPrompt,
		"acp_mode":        c.ACPMode,
		"acp_conn":        c.ACPConn,
		"acp_session_id":  c.ACPSessionID,
		"agent_name":      c.AgentName,
		"agent_version":   c.AgentVersion,
	}
}

// FromMap 从 map[string]interface{} 加载配置
func (c *AppConfig) FromMap(m map[string]interface{}) {
	if val, ok := m["api_key"].(string); ok {
		c.APIKey = val
	}
	if val, ok := m["model"].(string); ok {
		c.Model = val
	}
	if val, ok := m["base_url"].(string); ok {
		c.BaseURL = val
	}
	if val, ok := m["max_tokens"].(int); ok {
		c.MaxTokens = val
	}
	if val, ok := m["stream"].(bool); ok {
		c.Stream = val
	}
	if val, ok := m["output_format"].(string); ok {
		c.OutputFormat = val
	}
	if val, ok := m["session_id"].(string); ok {
		c.SessionID = val
	}
	if val, ok := m["system_prompt"].(string); ok {
		c.SystemPrompt = val
	}
	if val, ok := m["acp_mode"].(bool); ok {
		c.ACPMode = val
	}
	if val, ok := m["acp_conn"].(interface{}); ok {
		c.ACPConn = val
	}
	if val, ok := m["acp_session_id"].(string); ok {
		c.ACPSessionID = val
	}
	if val, ok := m["agent_name"].(string); ok {
		c.AgentName = val
	}
	if val, ok := m["agent_version"].(string); ok {
		c.AgentVersion = val
	}
}
