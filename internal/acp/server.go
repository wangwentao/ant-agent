package acp

import (
	"context"
	"os"

	"ant-agent/internal/config"
	"ant-agent/internal/skills"
	"ant-agent/internal/tools"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/coder/acp-go-sdk"
)

// Server represents the ACP server
type Server struct {
	agent *Agent
}

// NewServer creates a new ACP server
func NewServer(client anthropic.Client, skillCatalog *skills.SkillCatalog, toolRegistry *tools.ToolRegistry, appConfig *config.AppConfig) *Server {
	agent := NewAgent(client, skillCatalog, toolRegistry, appConfig)
	return &Server{
		agent: agent,
	}
}

// Start starts the ACP server
func (s *Server) Start(ctx context.Context) error {
	// Create agent-side connection
	asc := acp.NewAgentSideConnection(s.agent, os.Stdout, os.Stdin)

	// Set the connection to the agent
	s.agent.SetAgentConnection(asc)

	// Block until the peer disconnects
	<-asc.Done()
	return nil
}
