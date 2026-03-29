package cmd

import (
	"fmt"
	"strings"
	"ant-agent/internal/skills"
)

// ShowSkillsCommand 显示技能命令
type ShowSkillsCommand struct {}

// NewShowSkillsCommand 创建新的显示技能命令
func NewShowSkillsCommand() *ShowSkillsCommand {
	return &ShowSkillsCommand{}
}

// Name 返回命令名称
func (c *ShowSkillsCommand) Name() string {
	return "show-skills"
}

// Description 返回命令描述
func (c *ShowSkillsCommand) Description() string {
	return "Show all available skills"
}

// Execute 执行命令
func (c *ShowSkillsCommand) Execute(args []string, skillCatalog *skills.SkillCatalog) error {
	fmt.Println("=== Available Skills ===")
	skills := skillCatalog.GetSkills()
	if len(skills) == 0 {
		fmt.Println("No skills available.")
		return nil
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
	return nil
}
