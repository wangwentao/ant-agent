package cmd

import (
	"strings"
	"ant-agent/internal/skills"
)

// Command 命令接口
type Command interface {
	Name() string
	Description() string
	Execute(args []string, skillCatalog *skills.SkillCatalog) error
}

// CommandRegistry 命令注册表
type CommandRegistry struct {
	commands map[string]Command
}

// NewCommandRegistry 创建新的命令注册表
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]Command),
	}
}

// Register 注册命令
func (r *CommandRegistry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

// GetCommand 获取命令
func (r *CommandRegistry) GetCommand(name string) (Command, bool) {
	cmd, exists := r.commands[name]
	return cmd, exists
}

// GetCommands 获取所有命令
func (r *CommandRegistry) GetCommands() map[string]Command {
	return r.commands
}

// ParseCommand 解析命令
func (r *CommandRegistry) ParseCommand(input string) (Command, []string, bool) {
	// 去除首尾空格
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil, false
	}

	// 分割输入为命令名和参数
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, nil, false
	}

	// 获取命令名
	cmdName := parts[0]

	// 检查是否为帮助命令的别名
	if cmdName == "?" {
		cmdName = "help"
	}

	// 检查是否为退出命令
	if cmdName == "q" {
		cmdName = "exit"
	}

	// 查找命令
	cmd, exists := r.GetCommand(cmdName)
	if !exists {
		return nil, nil, false
	}

	// 获取参数
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	return cmd, args, true
}
