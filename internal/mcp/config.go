package mcp

import (
	"encoding/json"
	"fmt"
	"os"
)

// ServerConfig MCP服务配置
type ServerConfig struct {
	Command  string   `json:"command,omitempty"`
	Args     []string `json:"args,omitempty"`
	Endpoint string   `json:"endpoint,omitempty"`
}

// Config MCP配置
type Config struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// LoadConfig 加载MCP配置
func LoadConfig() (*Config, error) {
	// 尝试从多个位置加载配置文件
	configPaths := []string{
		"./mcp.json",
		"$HOME/.ant-agent/mcp.json",
		"/etc/ant-agent/mcp.json",
	}

	for _, path := range configPaths {
		// 解析环境变量
		path = os.ExpandEnv(path)

		// 检查文件是否存在
		if _, err := os.Stat(path); err == nil {
			// 读取配置文件
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
			}

			// 解析配置文件
			var config Config
			if err := json.Unmarshal(content, &config); err != nil {
				return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
			}

			// 验证配置
			if err := config.Validate(); err != nil {
				return nil, fmt.Errorf("invalid config file %s: %w", path, err)
			}

			return &config, nil
		}
	}

	// 如果没有找到配置文件，返回默认配置
	return &Config{
		MCPServers: make(map[string]ServerConfig),
	}, nil
}

// GetServerConfig 获取指定服务的配置
func (c *Config) GetServerConfig(serverName string) (ServerConfig, bool) {
	config, ok := c.MCPServers[serverName]
	return config, ok
}

// Validate 验证配置
func (c *Config) Validate() error {
	for serverName, serverConfig := range c.MCPServers {
		// 检查是否使用stdio传输层（需要Command和Args）
		if serverConfig.Command != "" {
			if len(serverConfig.Args) == 0 {
				return fmt.Errorf("server %s: args is required when command is specified", serverName)
			}
		} else if serverConfig.Endpoint == "" {
			// 检查是否使用streamablehttp传输层（需要Endpoint）
			return fmt.Errorf("server %s: either command/args or endpoint is required", serverName)
		}
	}
	return nil
}
