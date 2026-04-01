package tools

import (
	"github.com/anthropics/anthropic-sdk-go"
)

// GenerateToolSchemas 生成工具的输入模式
func GenerateToolSchemas(registry *ToolRegistry) []anthropic.ToolUnionParam {
	tools := registry.GetAllTools()
	schemas := make([]anthropic.ToolUnionParam, 0, len(tools))

	for _, tool := range tools {
		// 为所有工具生成工具模式，包括内置工具和 MCP 工具
		inputSchema := anthropic.ToolInputSchemaParam{
			Type:       "object",
			Properties: tool.Schema(),
			Required:   tool.Required(),
		}

		// 创建ToolParam并设置Description字段
		toolParam := anthropic.ToolParam{
			InputSchema: inputSchema,
			Name:        tool.Name(),
			Description: anthropic.String(tool.Description()),
		}

		// 创建ToolUnionParam
		schemas = append(schemas, anthropic.ToolUnionParam{OfTool: &toolParam})
	}

	return schemas
}
