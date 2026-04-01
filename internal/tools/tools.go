package tools

import (
	"context"
	"fmt"

	"ant-agent/internal/skills"
	"ant-agent/internal/tools/impl"
)

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	Run(ctx context.Context, input map[string]interface{}) (string, error)
	Schema() map[string]interface{}
	Required() []string
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry 创建一个新的工具注册表
func NewToolRegistry(skillCatalog *skills.SkillCatalog) *ToolRegistry {
	registry := &ToolRegistry{
		tools: make(map[string]Tool),
	}

	// 注册内置工具
	registry.Register(&impl.ShellCommandTool{})
	registry.Register(&impl.ReadFileTool{})
	registry.Register(&impl.WriteFileTool{})
	registry.Register(&impl.EditFileTool{})
	registry.Register(impl.NewListSkillsTool(skillCatalog))
	registry.Register(impl.NewActivateSkillTool(skillCatalog))

	return registry
}

// Register 注册工具
func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// GetTool 根据名称获取工具
func (r *ToolRegistry) GetTool(name string) (Tool, bool) {
	tool, exists := r.tools[name]
	return tool, exists
}

// GetAllTools 获取所有工具
func (r *ToolRegistry) GetAllTools() map[string]Tool {
	return r.tools
}

// RunTool 运行工具
func (r *ToolRegistry) RunTool(ctx context.Context, toolName string, input map[string]interface{}) string {
	tool, exists := r.GetTool(toolName)
	if !exists {
		return fmt.Sprintf("Tool %s not implemented", toolName)
	}

	result, err := tool.Run(ctx, input)
	if err != nil {
		return fmt.Sprintf("Error running tool %s: %v", toolName, err)
	}

	// 优化工具执行结果的展示格式
	formattedResult := fmt.Sprintf("=== Tool: %s ===\n%s\n=== End of Tool Result ===", toolName, result)
	return formattedResult
}
