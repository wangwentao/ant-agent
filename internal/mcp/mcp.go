package mcp

import (
	"ant-agent/internal/logs"
	"context"
	"fmt"
)

// Manager MCP管理器
type Manager struct {
	clientManager *ClientManager
	config        *Config
}

// NewManager 创建新的MCP管理器
func NewManager() (*Manager, error) {
	// 加载配置
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load mcp config: %w", err)
	}

	// 创建客户端管理器
	clientManager := NewClientManager(config)

	return &Manager{
		clientManager: clientManager,
		config:        config,
	}, nil
}

// GetTools 获取所有MCP服务的工具
func (m *Manager) GetTools(ctx context.Context) ([]*Tool, error) {
	var allTools []*Tool

	// 遍历所有配置的MCP服务
	for serverName := range m.config.MCPServers {
		logs.Debug("Connecting to MCP server: %s", serverName)

		// 获取会话
		session, err := m.clientManager.GetSession(ctx, serverName)
		if err != nil {
			logs.Warn("Failed to connect to MCP server %s: %v", serverName, err)
			// 跳过无法连接的服务
			continue
		}

		// 获取服务的工具
		tools, err := GetTools(ctx, session)
		if err != nil {
			logs.Warn("Failed to get tools from MCP server %s: %v", serverName, err)
			// 跳过无法获取工具的服务
			continue
		}

		logs.Debug("Successfully got %d tools from MCP server %s", len(tools), serverName)
		allTools = append(allTools, tools...)
	}

	logs.Debug("Total MCP tools registered: %d", len(allTools))
	return allTools, nil
}

// StopAll 停止所有MCP服务
func (m *Manager) StopAll() {
	logs.Debug("Stopping all MCP services...")
	m.clientManager.StopAll()
	logs.Debug("All MCP services stopped")
}
