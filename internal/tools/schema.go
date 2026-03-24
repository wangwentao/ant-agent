package tools

import (
	"github.com/anthropics/anthropic-sdk-go"
)

// GenerateToolSchemas 生成工具的输入模式
func GenerateToolSchemas(registry *ToolRegistry) []anthropic.ToolUnionParam {
	tools := registry.GetAllTools()
	schemas := make([]anthropic.ToolUnionParam, 0, len(tools))

	for _, tool := range tools {
		switch tool.Name() {
		case "execute_shell":
			schemas = append(schemas, anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "The shell command to execute",
					},
				},
				Required: []string{"command"},
			}, tool.Name()))

		case "read_file":
			schemas = append(schemas, anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]any{
					"file_path": map[string]any{
						"type":        "string",
						"description": "The path to the file to read",
					},
				},
				Required: []string{"file_path"},
			}, tool.Name()))

		case "write_file":
			schemas = append(schemas, anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]any{
					"file_path": map[string]any{
						"type":        "string",
						"description": "The path to the file to write",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "The content to write to the file",
					},
				},
				Required: []string{"file_path", "content"},
			}, tool.Name()))

		case "edit_file":
			schemas = append(schemas, anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]any{
					"file_path": map[string]any{
						"type":        "string",
						"description": "The path to the file to edit",
					},
					"old_string": map[string]any{
						"type":        "string",
						"description": "The string to replace in the file",
					},
					"new_string": map[string]any{
						"type":        "string",
						"description": "The new string to replace with",
					},
				},
				Required: []string{"file_path", "old_string", "new_string"},
			}, tool.Name()))

		case "list_skills":
			schemas = append(schemas, anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
				Type:       "object",
				Properties: map[string]any{},
				Required:   []string{},
			}, tool.Name()))

		case "activate_skill":
			schemas = append(schemas, anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]any{
					"skill_name": map[string]any{
						"type":        "string",
						"description": "The name of the skill to activate",
					},
				},
				Required: []string{"skill_name"},
			}, tool.Name()))
		}
	}

	return schemas
}
