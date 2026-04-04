package impl

import (
	"ant-agent/internal/logs"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ShellCommandTool 执行shell命令的工具
type ShellCommandTool struct{}

func (t *ShellCommandTool) Name() string {
	return "execute_shell"
}

func (t *ShellCommandTool) Description() string {
	return "Executes a shell command and returns the output"
}

func (t *ShellCommandTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"command": map[string]interface{}{
			"type":        "string",
			"description": "The shell command to execute",
		},
	}
}

func (t *ShellCommandTool) Required() []string {
	return []string{"command"}
}

func (t *ShellCommandTool) Run(ctx context.Context, input map[string]interface{}) (string, error) {
	// 检查是否提供了command参数
	command, ok := input["command"].(string)
	if !ok {
		return "Error: command parameter is required", nil
	}

	// 安全控制：检查危险命令
	// 只保留真正危险的命令，放宽对常用命令的限制
	dangerousCommands := []string{
		"rm -rf", "rmdir -rf", "shutdown", "reboot", "halt", "poweroff",
		"dd", "mkfs", "format", "fdisk", "parted",
		"kill", "killall", "pkill",
		"sudo", "su", "login", "passwd",
		"wget", "ftp", "scp", "sftp",
	}

	// 检查命令是否包含危险操作
	cmdLower := strings.ToLower(command)
	for _, dangerousCmd := range dangerousCommands {
		// 确保我们匹配的是完整的命令，而不是命令的一部分
		// 例如，"rm" 应该匹配 "rm file" 但不应该匹配 "ls -lrm"
		if strings.Contains(cmdLower, " "+dangerousCmd+" ") || strings.HasPrefix(cmdLower, dangerousCmd+" ") || strings.HasSuffix(cmdLower, " "+dangerousCmd) {
			return fmt.Sprintf("Security Warning: Command '%s' is potentially dangerous and has been blocked.\nFor safety reasons, this operation is not allowed.", command), nil
		}
	}

	// 显示执行命令的提示
	logs.Debug("Executing command: %s\n", command)

	// 执行shell命令，添加超时机制
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second) // 30秒超时
	defer cancel()

	cmd := exec.CommandContext(ctxWithTimeout, "bash", "-c", command)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if ctxWithTimeout.Err() == context.DeadlineExceeded {
			return fmt.Sprintf("Error executing command: Command timed out after 30 seconds\nOutput: %s", string(output)), nil
		}
		return fmt.Sprintf("Error executing command: %v\nOutput: %s", err, string(output)), nil
	}

	// 显示命令执行完成的提示
	logs.Debug("Command executed successfully")
	return string(output), nil
}
