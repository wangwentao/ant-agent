package acp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ant-agent/internal/config"
	"ant-agent/internal/messages"
	"ant-agent/internal/skills"
	"ant-agent/internal/tools"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/coder/acp-go-sdk"
)

// Session represents an ACP session
type Session struct {
	ID        string
	Mode      string
	StartTime int64
	State     map[string]interface{}
}

// Agent implements the acp.Agent interface
type Agent struct {
	messageHandler *messages.MessageHandler
	client         anthropic.Client
	appConfig      *config.AppConfig
	sessions       map[string]*Session
	conn           *acp.AgentSideConnection
}

// NewAgent creates a new ACP agent
func NewAgent(client anthropic.Client, skillCatalog *skills.SkillCatalog, toolRegistry *tools.ToolRegistry, appConfig *config.AppConfig) *Agent {
	return &Agent{
		messageHandler: messages.NewMessageHandler(skillCatalog, toolRegistry, appConfig),
		client:         client,
		appConfig:      appConfig,
		sessions:       make(map[string]*Session),
	}
}

// Initialize handles the Initialize request
func (a *Agent) Initialize(ctx context.Context, req acp.InitializeRequest) (acp.InitializeResponse, error) {
	// 从配置中读取 agent 信息，避免硬编码
	agentName := a.appConfig.AgentName
	agentVersion := a.appConfig.AgentVersion

	// 构建 AgentInfo
	agentInfo := &acp.Implementation{
		Name:    agentName,
		Version: agentVersion,
	}

	// 构建 AgentCapabilities
	agentCapabilities := acp.AgentCapabilities{
		LoadSession: true, // 支持会话管理
	}

	// 构建初始化响应
	response := acp.InitializeResponse{
		AgentInfo:         agentInfo,
		AgentCapabilities: agentCapabilities,
		AuthMethods:       []acp.AuthMethod{},
		ProtocolVersion:   acp.ProtocolVersionNumber,
	}

	return response, nil
}

// Authenticate handles the Authenticate request
func (a *Agent) Authenticate(ctx context.Context, req acp.AuthenticateRequest) (acp.AuthenticateResponse, error) {
	// TODO: Implement Authenticate
	return acp.AuthenticateResponse{}, nil
}

// NewSession handles the NewSession request
func (a *Agent) NewSession(ctx context.Context, req acp.NewSessionRequest) (acp.NewSessionResponse, error) {
	// 生成唯一会话 ID
	sessionID := "session-" + generateID()

	// 创建新会话
	session := &Session{
		ID:        sessionID,
		Mode:      "interactive", // 默认模式
		StartTime: getCurrentTimestamp(),
		State:     make(map[string]interface{}),
	}

	// 存储会话
	a.sessions[sessionID] = session

	// 构建响应
	response := acp.NewSessionResponse{
		SessionId: acp.SessionId(sessionID),
	}

	return response, nil
}

// Prompt handles the Prompt request
func (a *Agent) Prompt(ctx context.Context, req acp.PromptRequest) (acp.PromptResponse, error) {
	// 获取会话 ID
	sessionID := string(req.SessionId)

	// 查找会话
	_, ok := a.sessions[sessionID]
	if !ok {
		return acp.PromptResponse{}, fmt.Errorf("session not found: %s", sessionID)
	}

	// 处理用户输入
	userInput := ""
	for _, content := range req.Prompt {
		if content.Text != nil {
			userInput += content.Text.Text
		}
	}

	// 发送会话更新通知
	if a.conn != nil {
		if err := a.conn.SessionUpdate(ctx, acp.SessionNotification{
			SessionId: acp.SessionId(sessionID),
			Update:    acp.UpdateUserMessageText(userInput),
		}); err != nil {
			return acp.PromptResponse{}, err
		}
	}

	// 设置 ACP 相关信息
	a.appConfig.ACPMode = true
	a.appConfig.ACPConn = a.conn
	a.appConfig.ACPSessionID = sessionID

	// 处理消息
	_, err := a.messageHandler.ProcessMessage(ctx, a.client, userInput, a.appConfig)
	if err != nil {
		// 发送错误通知
		if a.conn != nil {
			_ = a.conn.SessionUpdate(ctx, acp.SessionNotification{
				SessionId: acp.SessionId(sessionID),
				Update:    acp.UpdateAgentMessageText(fmt.Sprintf("Error: %v", err)),
			})
		}
		return acp.PromptResponse{}, err
	}

	// 构建响应
	response := acp.PromptResponse{
		StopReason: acp.StopReasonEndTurn,
	}

	return response, nil
}

// HandleExtensionMethod handles extension methods
func (a *Agent) HandleExtensionMethod(ctx context.Context, method string, params json.RawMessage) (any, error) {
	// TODO: Implement HandleExtensionMethod
	return nil, acp.NewMethodNotFound(method)
}

// Cancel handles the Cancel request
func (a *Agent) Cancel(ctx context.Context, req acp.CancelNotification) error {
	// 获取会话 ID
	sessionID := string(req.SessionId)

	// 查找会话
	_, ok := a.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// 这里可以添加取消逻辑，例如取消正在进行的操作

	return nil
}

// SetSessionMode handles the SetSessionMode request
func (a *Agent) SetSessionMode(ctx context.Context, req acp.SetSessionModeRequest) (acp.SetSessionModeResponse, error) {
	// 构建响应
	response := acp.SetSessionModeResponse{}

	return response, nil
}

// SetAgentConnection implements acp.AgentConnAware to receive the connection after construction
func (a *Agent) SetAgentConnection(conn *acp.AgentSideConnection) {
	a.conn = conn
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// getCurrentTimestamp returns the current timestamp in seconds
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
