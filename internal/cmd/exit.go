package cmd

import (
	"os"
	"ant-agent/internal/logs"
	"ant-agent/internal/skills"
)

// ExitCommand 退出命令
type ExitCommand struct {}

// NewExitCommand 创建新的退出命令
func NewExitCommand() *ExitCommand {
	return &ExitCommand{}
}

// Name 返回命令名称
func (c *ExitCommand) Name() string {
	return "exit"
}

// Description 返回命令描述
func (c *ExitCommand) Description() string {
	return "Exit the agent"
}

// Execute 执行命令
func (c *ExitCommand) Execute(args []string, skillCatalog *skills.SkillCatalog) error {
	logs.Info("Exiting...")
	os.Exit(0)
	return nil
}
