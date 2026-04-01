package cmd

import (
	"ant-agent/internal/skills"
	"ant-agent/internal/tools"
	"fmt"
)

// HelpCommand 帮助命令
type HelpCommand struct {
	registry     *CommandRegistry
	agentName    string
	toolRegistry *tools.ToolRegistry
}

// NewHelpCommand 创建新的帮助命令
func NewHelpCommand(registry *CommandRegistry, agentName string, toolRegistry *tools.ToolRegistry) *HelpCommand {
	return &HelpCommand{
		registry:     registry,
		agentName:    agentName,
		toolRegistry: toolRegistry,
	}
}

// Name 返回命令名称
func (c *HelpCommand) Name() string {
	return "help"
}

// Description 返回命令描述
func (c *HelpCommand) Description() string {
	return "Show this help message"
}

// Execute 执行命令
func (c *HelpCommand) Execute(args []string, skillCatalog *skills.SkillCatalog) error {
	fmt.Printf("=== %s Help ===\n", c.agentName)
	fmt.Println("Available commands:")

	// 显示所有命令
	commands := c.registry.GetCommands()
	for name, cmd := range commands {
		fmt.Printf("  %-20s - %s\n", name, cmd.Description())
	}

	fmt.Println("\nOptions:")
	fmt.Println("  --skill <name>   - Install specific skill from skills directory")
	fmt.Println("  --skills <names> - Install multiple skills (space-separated)")

	fmt.Println("\nType your message to interact with the agent.")
	fmt.Println("\nThe agent can use the following tools:")
	if c.toolRegistry != nil {
		allTools := c.toolRegistry.GetAllTools()
		for name, tool := range allTools {
			fmt.Printf("  %-20s - %s\n", name, tool.Description())
		}
	} else {
		// 回退到默认工具列表
		fmt.Println("  list_skills      - List all available skills")
		fmt.Println("  activate_skill <name> - Activate a skill")
		fmt.Println("  execute_shell    - Execute a shell command")
		fmt.Println("  read_file        - Read the content of a file")
		fmt.Println("  write_file       - Write content to a file")
		fmt.Println("  edit_file        - Edit a file by replacing a string")
	}

	return nil
}
