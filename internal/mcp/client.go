package mcp

import (
	"context"
	"sync"

	"ant-agent/internal/logs"
)

// ClientManager MCP客户端管理器
type ClientManager struct {
	config   *Config
	sessions map[string]*Session
	mutex    sync.Mutex
}

// NewClientManager 创建新的MCP客户端管理器
func NewClientManager(config *Config) *ClientManager {
	return &ClientManager{
		config:   config,
		sessions: make(map[string]*Session),
	}
}

// GetSession 获取指定服务的MCP会话
func (cm *ClientManager) GetSession(ctx context.Context, serverName string) (*Session, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 检查会话是否已经存在
	if session, ok := cm.sessions[serverName]; ok {
		logs.Debug("Using existing session for MCP server: %s", serverName)
		return session, nil
	}

	// 获取服务配置
	config, ok := cm.config.GetServerConfig(serverName)
	if !ok {
		logs.Debug("No configuration found for MCP server: %s", serverName)
		return nil, nil
	}

	// 创建新会话
	if config.Endpoint != "" {
		logs.Debug("Creating new session for MCP server: %s with endpoint: %s", serverName, config.Endpoint)
	} else {
		logs.Debug("Creating new session for MCP server: %s with command: %s %v", serverName, config.Command, config.Args)
	}
	session, err := NewSession(serverName, config)
	if err != nil {
		logs.Warn("Failed to create session for MCP server %s: %v", serverName, err)
		return nil, err
	}

	// 存储会话
	cm.sessions[serverName] = session
	logs.Debug("Successfully created session for MCP server: %s", serverName)

	return session, nil
}

// StopAll 停止所有MCP服务
func (cm *ClientManager) StopAll() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if len(cm.sessions) == 0 {
		logs.Debug("No MCP sessions to stop")
		return
	}

	logs.Debug("Stopping %d MCP sessions...", len(cm.sessions))
	for serverName, session := range cm.sessions {
		logs.Debug("Stopping session for MCP server: %s", serverName)
		if err := session.Close(); err != nil {
			logs.Warn("Failed to stop session for MCP server %s: %v", serverName, err)
		} else {
			logs.Debug("Successfully stopped session for MCP server: %s", serverName)
		}
	}

	// 清空会话列表
	cm.sessions = make(map[string]*Session)
	logs.Debug("All MCP sessions stopped")
}
