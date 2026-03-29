package input

import (
	"fmt"
	"strings"

	"ant-agent/internal/cmd"
	"ant-agent/internal/logs"
	"ant-agent/internal/skills"

	"github.com/chzyer/readline"
)

// InputHandler 处理用户输入
type InputHandler struct {
	skillCatalog *skills.SkillCatalog
	readline     *readline.Instance
	cmdRegistry  *cmd.CommandRegistry
}

// NewInputHandler 创建一个新的输入处理器
func NewInputHandler(skillCatalog *skills.SkillCatalog) *InputHandler {
	// 初始化命令注册表
	registry := cmd.NewCommandRegistry()

	// 注册命令
	exitCmd := cmd.NewExitCommand()
	helpCmd := cmd.NewHelpCommand(registry)
	showSkillsCmd := cmd.NewShowSkillsCommand()
	installSkillCmd := cmd.NewInstallSkillCommand()
	removeSkillCmd := cmd.NewRemoveSkillCommand()

	registry.Register(exitCmd)
	registry.Register(helpCmd)
	registry.Register(showSkillsCmd)
	registry.Register(installSkillCmd)
	registry.Register(removeSkillCmd)

	// 初始化 readline
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[34mYou »:\033[0m",
		HistoryFile:     ".ant-agent-history",
		AutoComplete:    NewCommandCompleter(skillCatalog, registry),
		InterruptPrompt: "^C",
		EOFPrompt:       "^D",
	})
	if err != nil {
		logs.Warn("Failed to initialize readline: %v, falling back to basic input", err)
		return &InputHandler{
			skillCatalog: skillCatalog,
			cmdRegistry:  registry,
		}
	}

	return &InputHandler{
		skillCatalog: skillCatalog,
		readline:     rl,
		cmdRegistry:  registry,
	}
}

// ReadInput 读取用户输入
func (h *InputHandler) ReadInput() (string, error) {
	if h.readline != nil {
		return h.readline.Readline()
	}

	// 回退到基本输入方式
	fmt.Print("\033[34mYou »:\033[0m")
	var input string
	_, err := fmt.Scanln(&input)
	return input, err
}

// CommandCompleter 命令补全器
type CommandCompleter struct {
	skillCatalog *skills.SkillCatalog
	cmdRegistry  *cmd.CommandRegistry
}

// NewCommandCompleter 创建一个新的命令补全器
func NewCommandCompleter(skillCatalog *skills.SkillCatalog, cmdRegistry *cmd.CommandRegistry) *CommandCompleter {
	return &CommandCompleter{
		skillCatalog: skillCatalog,
		cmdRegistry:  cmdRegistry,
	}
}

// Do 执行命令补全
func (c *CommandCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	// 获取当前输入
	input := string(line[:pos])

	// 从命令注册表获取命令列表
	commands := make([]string, 0)
	for name := range c.cmdRegistry.GetCommands() {
		commands = append(commands, name)
	}
	// 添加退出命令的别名
	commands = append(commands, "q")
	// 添加帮助命令的别名
	commands = append(commands, "?")

	// 获取技能列表
	skills := c.skillCatalog.GetSkills()
	skillNames := make([]string, 0, len(skills))
	for name := range skills {
		skillNames = append(skillNames, name)
	}

	// 补全逻辑
	var completions []string

	// 检查是否是命令前缀
	for _, cmd := range commands {
		if strings.HasPrefix(cmd, input) {
			// 如果输入与命令相同，不添加到补全列表
			if cmd != input {
				// 只返回要追加的部分，而不是完整的命令
				completion := cmd[len(input):]
				completions = append(completions, completion)
			}
		}
	}

	// 检查是否是技能名称前缀（用于 remove-skill 命令）
	if strings.HasPrefix(input, "remove-skill ") {
		prefix := strings.TrimPrefix(input, "remove-skill ")
		for _, skill := range skillNames {
			if strings.HasPrefix(skill, prefix) {
				// 只返回要追加的部分，而不是完整的命令
				completion := skill[len(prefix):]
				completions = append(completions, completion)
			}
		}
	}

	// 转换为 [][]rune
	for _, comp := range completions {
		newLine = append(newLine, []rune(comp))
	}

	// 始终返回原始输入的长度，因为 readline 会用补全结果替换从0到这个长度的部分
	return newLine, len(input)
}

// ProcessInput 处理用户输入
func (h *InputHandler) ProcessInput(input string) (string, bool, error) {
	// 检查空输入
	input = strings.TrimSpace(input)
	if input == "" {
		logs.Info("Please enter something. If you want to exit, type 'exit' or 'q'.")
		return "", true, nil
	}

	// 尝试解析命令
	cmd, args, found := h.cmdRegistry.ParseCommand(input)
	if found {
		// 执行命令
		err := cmd.Execute(args, h.skillCatalog)
		if err != nil {
			logs.Error("Error executing command: %v", err)
		} else if cmd.Name() != "exit" {
			logs.Info("Command executed successfully!")
		}
		return "", true, nil
	}

	// 不是特殊指令，返回用户输入
	return input, false, nil
}
