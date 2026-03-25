package input

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"ant-agent/internal/logs"
	"ant-agent/internal/skills"

	"github.com/chzyer/readline"
)

// InputHandler 处理用户输入
type InputHandler struct {
	skillCatalog *skills.SkillCatalog
	readline     *readline.Instance
}

// NewInputHandler 创建一个新的输入处理器
func NewInputHandler(skillCatalog *skills.SkillCatalog) *InputHandler {
	// 初始化 readline
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[34mYou »:\033[0m",
		HistoryFile:     ".ant-agent-history",
		AutoComplete:    NewCommandCompleter(skillCatalog),
		InterruptPrompt: "^C",
		EOFPrompt:       "^D",
	})
	if err != nil {
		logs.Warn("Failed to initialize readline: %v, falling back to basic input", err)
		return &InputHandler{
			skillCatalog: skillCatalog,
		}
	}

	return &InputHandler{
		skillCatalog: skillCatalog,
		readline:     rl,
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
}

// NewCommandCompleter 创建一个新的命令补全器
func NewCommandCompleter(skillCatalog *skills.SkillCatalog) *CommandCompleter {
	return &CommandCompleter{
		skillCatalog: skillCatalog,
	}
}

// Do 执行命令补全
func (c *CommandCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	// 获取当前输入
	input := string(line[:pos])

	// 基本命令列表
	commands := []string{
		"help", "exit", "q",
		"install-skill", "remove-skill",
		"show-skills",
	}

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
	// 检查是否为退出指令
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "exit" || input == "q" {
		logs.Info("Exiting...")
		os.Exit(0)
	}

	// 检查空输入
	if input == "" {
		logs.Info("Please enter something. If you want to exit, type 'exit' or 'q'.")
		return "", true, nil
	}

	// 检查是否为help指令
	if input == "help" || input == "?" {
		h.showHelp()
		return "", true, nil
	}

	// 检查是否为install-skill指令
	if strings.HasPrefix(input, "install-skill") {
		// 提取技能路径或URL
		skillPath := strings.TrimSpace(strings.TrimPrefix(input, "install-skill"))
		if skillPath == "" {
			logs.Error("Please provide a skill directory path or URL.")
			return "", true, nil
		}

		// 安装技能
		err := installSkill(skillPath, h.skillCatalog)
		if err != nil {
			logs.Error("Error installing skill: %v", err)
		} else {
			logs.Info("Skill installed successfully!")
		}
		return "", true, nil
	}

	// 检查是否为remove-skill指令
	if strings.HasPrefix(input, "remove-skill") {
		// 提取技能名称
		skillName := strings.TrimSpace(strings.TrimPrefix(input, "remove-skill"))
		if skillName == "" {
			logs.Error("Please provide a skill name.")
			return "", true, nil
		}

		// 删除技能
		err := removeSkill(skillName, h.skillCatalog)
		if err != nil {
			logs.Error("Error removing skill: %v", err)
		} else {
			logs.Info("Skill removed successfully!")
		}
		return "", true, nil
	}

	// 检查是否为show-skills指令
	if input == "show-skills" {
		h.showSkills()
		return "", true, nil
	}

	// 不是特殊指令，返回用户输入
	return input, false, nil
}

// showHelp 显示帮助信息
func (h *InputHandler) showHelp() {
	fmt.Println("=== Ant-Agent Help ===")
	fmt.Println("Available commands:")
	fmt.Println("  help, ?          - Show this help message")
	fmt.Println("  exit, q          - Exit the agent")
	fmt.Println("  install-skill <path> - Install a skill from a directory")
	fmt.Println("  remove-skill <name> - Remove a skill by name")
	fmt.Println("  show-skills      - Show all available skills")
	fmt.Println("\nType your message to interact with the agent.")
	fmt.Println("\nThe agent can use the following tools:")
	fmt.Println("  list_skills      - List all available skills")
	fmt.Println("  activate_skill <name> - Activate a skill")
	fmt.Println("  execute_shell    - Execute a shell command")
	fmt.Println("  read_file        - Read the content of a file")
	fmt.Println("  write_file       - Write content to a file")
	fmt.Println("  edit_file        - Edit a file by replacing a string")
}

// showSkills 显示所有可用的技能
func (h *InputHandler) showSkills() {
	fmt.Println("=== Available Skills ===")
	skills := h.skillCatalog.GetSkills()
	if len(skills) == 0 {
		fmt.Println("No skills available.")
		return
	}

	// 计算最长的技能名称长度，用于对齐显示
	maxNameLength := 0
	for name := range skills {
		if len(name) > maxNameLength {
			maxNameLength = len(name)
		}
	}

	// 确保最小长度为10，保证显示效果
	if maxNameLength < 10 {
		maxNameLength = 10
	}

	// 显示技能列表
	for name, skill := range skills {
		// 计算填充空格
		padding := maxNameLength - len(name)
		space := strings.Repeat(" ", padding)

		// 打印技能名称
		fmt.Printf("  %s%s - ", name, space)

		// 格式化技能描述，实现自动换行和缩进
		description := skill.Description
		lineLength := 100                 // 每行最大长度
		prefixLength := maxNameLength + 6 // 前缀长度（包括空格和连字符）
		remainingLength := lineLength - prefixLength

		words := strings.Fields(description)
		currentLine := ""

		for _, word := range words {
			if len(currentLine)+len(word)+1 <= remainingLength {
				if currentLine != "" {
					currentLine += " "
				}
				currentLine += word
			} else {
				// 打印当前行并开始新行
				fmt.Println(currentLine)
				// 新行添加缩进
				currentLine = strings.Repeat(" ", prefixLength) + word
			}
		}

		// 打印最后一行
		if currentLine != "" {
			fmt.Println(currentLine)
		}

		// 每个技能之间添加空行
		fmt.Println()
	}

	fmt.Printf("Total: %d skills available\n", len(skills))
}

// installSkill 安装技能并验证其结构是否符合规范
func installSkill(skillPath string, skillCatalog *skills.SkillCatalog) error {
	// 检查路径是否存在
	info, err := os.Stat(skillPath)
	if err != nil {
		return fmt.Errorf("skill path does not exist: %w", err)
	}

	// 确保是目录
	if !info.IsDir() {
		return fmt.Errorf("skill path must be a directory")
	}

	// 检查是否存在SKILL.md文件
	skillMdPath := strings.Join([]string{skillPath, "SKILL.md"}, "/")
	if _, err := os.Stat(skillMdPath); err != nil {
		return fmt.Errorf("SKILL.md file not found in skill directory")
	}

	// 解析SKILL.md文件，验证结构
	skill, err := skills.ParseSkill(skillMdPath)
	if err != nil {
		return fmt.Errorf("invalid SKILL.md file: %w", err)
	}

	// 验证技能名称和描述
	if skill.Name == "" {
		return fmt.Errorf("skill name is required")
	}
	if skill.Description == "" {
		return fmt.Errorf("skill description is required")
	}

	// 确定目标安装目录
	targetDir := strings.Join([]string{".agents", "skills", skill.Name}, "/")

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 复制技能文件
	if err := copySkillFiles(skillPath, targetDir); err != nil {
		return fmt.Errorf("failed to copy skill files: %w", err)
	}

	// 重新发现技能
	if err := skillCatalog.DiscoverSkills(); err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	return nil
}

// copySkillFiles 复制技能文件到目标目录
func copySkillFiles(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// 跳过隐藏文件和目录（除了源目录本身）
		if strings.HasPrefix(relPath, ".") && relPath != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 构建目标路径
		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			// 创建目录
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// 复制文件
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			dstFile, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer dstFile.Close()

			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return err
			}

			// 设置文件权限
			return os.Chmod(dstPath, info.Mode())
		}
	})
}

// removeSkill 删除指定名称的技能
func removeSkill(skillName string, skillCatalog *skills.SkillCatalog) error {
	// 检查技能是否存在
	skill, exists := skillCatalog.GetSkill(skillName)
	if !exists {
		return fmt.Errorf("skill '%s' not found", skillName)
	}

	// 确定技能目录路径
	skillDir := filepath.Dir(skill.Location)

	// 删除技能目录
	if err := os.RemoveAll(skillDir); err != nil {
		return fmt.Errorf("failed to remove skill directory: %w", err)
	}

	// 重新发现技能，更新技能目录
	if err := skillCatalog.DiscoverSkills(); err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	return nil
}
