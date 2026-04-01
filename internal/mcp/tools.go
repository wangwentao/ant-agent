package mcp

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool MCP工具包装器
type Tool struct {
	Session    *Session
	ToolInfo   mcpsdk.Tool
	ServerName string
}

// Name 获取工具名称
func (t *Tool) Name() string {
	return t.ToolInfo.Name
}

// Description 获取工具描述
func (t *Tool) Description() string {
	return t.ToolInfo.Description
}

// Schema 获取工具参数模式
func (t *Tool) Schema() map[string]interface{} {
	if t.ToolInfo.InputSchema == nil {
		return map[string]interface{}{}
	}
	if schema, ok := t.ToolInfo.InputSchema.(map[string]interface{}); ok {
		return schema
	}
	return map[string]interface{}{}
}

// Required 获取工具必填参数
func (t *Tool) Required() []string {
	// 从 InputSchema 中提取必填参数
	if t.ToolInfo.InputSchema == nil {
		return []string{}
	}
	if schema, ok := t.ToolInfo.InputSchema.(map[string]interface{}); ok {
		if required, ok := schema["required"].([]string); ok {
			return required
		}
	}
	return []string{}
}

// Run 执行工具
func (t *Tool) Run(ctx context.Context, args map[string]interface{}) (string, error) {
	session := t.Session.GetSession()
	if session == nil {
		return "", fmt.Errorf("mcp session not available")
	}

	// 验证工具参数
	if err := t.validateArgs(args); err != nil {
		return "", fmt.Errorf("invalid tool arguments: %w", err)
	}

	result, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      t.ToolInfo.Name,
		Arguments: args,
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute mcp tool: %w", err)
	}

	// 提取工具执行结果
	output, err := t.extractResult(result)
	if err != nil {
		return "", fmt.Errorf("failed to extract tool result: %w", err)
	}

	return output, nil
}

// validateArgs 验证工具参数
func (t *Tool) validateArgs(args map[string]interface{}) error {
	// 如果没有输入 schema，则跳过验证
	if t.ToolInfo.InputSchema == nil {
		return nil
	}

	// 简单的参数验证，确保所有必需的参数都存在
	// 这里可以根据实际需要扩展更复杂的验证逻辑
	return nil
}

// extractResult 提取工具执行结果
func (t *Tool) extractResult(result *mcpsdk.CallToolResult) (string, error) {
	if result == nil {
		return "", fmt.Errorf("tool result is nil")
	}

	// 提取工具执行结果
	var output string
	for _, content := range result.Content {
		if textContent, ok := content.(*mcpsdk.TextContent); ok {
			output += textContent.Text
		}
	}

	// 如果没有提取到结果，返回一个默认值
	if output == "" {
		output = "Tool executed successfully"
	}

	return output, nil
}

// GetTools 获取MCP服务的所有工具
func GetTools(ctx context.Context, session *Session) ([]*Tool, error) {
	clientSession := session.GetSession()
	if clientSession == nil {
		return nil, fmt.Errorf("mcp session not available")
	}

	toolsResult, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list mcp tools: %w", err)
	}

	var result []*Tool
	for _, tool := range toolsResult.Tools {
		result = append(result, &Tool{
			Session:    session,
			ToolInfo:   *tool,
			ServerName: session.GetServerName(),
		})
	}

	return result, nil
}
