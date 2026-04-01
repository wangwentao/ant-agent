package mcp

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Session MCP会话管理
type Session struct {
	client     *mcpsdk.Client
	session    *mcpsdk.ClientSession
	server     *exec.Cmd
	serverName string
	mutex      sync.Mutex
	isClosed   bool
}

// NewSession 创建新的MCP会话
func NewSession(serverName string, config ServerConfig) (*Session, error) {
	// 创建MCP客户端
	client := mcpsdk.NewClient(
		&mcpsdk.Implementation{
			Name:    "ant-agent",
			Version: "1.0.0",
		},
		nil,
	)

	var transport mcpsdk.Transport
	var server *exec.Cmd

	// 根据配置选择传输层
	if config.Endpoint != "" {
		// 使用StreamableHTTP传输层
		transport = &mcpsdk.StreamableClientTransport{
			Endpoint: config.Endpoint,
		}
	} else {
		// 使用CommandTransport（stdio传输层）
		cmd := exec.Command(config.Command, config.Args...)
		transport = &mcpsdk.CommandTransport{
			Command: cmd,
		}
		server = cmd
	}

	ctx := context.Background()
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mcp server: %w", err)
	}

	return &Session{
		client:     client,
		session:    session,
		server:     server,
		serverName: serverName,
		isClosed:   false,
	}, nil
}

// Close 关闭MCP会话
func (s *Session) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查会话是否已经关闭
	if s.isClosed {
		return nil
	}

	if s.session != nil {
		if err := s.session.Close(); err != nil {
			return err
		}
		s.session = nil
	}

	if s.server != nil {
		if err := s.server.Process.Kill(); err != nil {
			return err
		}
		if err := s.server.Wait(); err != nil {
			return err
		}
		s.server = nil
	}

	// 标记会话为已关闭
	s.isClosed = true

	return nil
}

// GetSession 获取MCP客户端会话
func (s *Session) GetSession() *mcpsdk.ClientSession {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查会话是否已经关闭
	if s.isClosed {
		return nil
	}

	return s.session
}

// GetServerName 获取服务器名称
func (s *Session) GetServerName() string {
	return s.serverName
}
