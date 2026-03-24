package main

import (
	"context"
	"fmt"

	"ant-agent/internal/config"
	"ant-agent/internal/input"
	"ant-agent/internal/logs"
	"ant-agent/internal/messages"
	"ant-agent/internal/skills"
	"ant-agent/internal/tools"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		logs.Warn("%v", err)
		// 继续执行，使用默认配置
	}

	// 初始化技能系统
	skillCatalog := skills.NewSkillCatalog()
	if err := skillCatalog.DiscoverSkills(); err != nil {
		logs.Warn("failed to discover skills: %v", err)
	}

	// 获取API Key
	apiKey := config.GetAPIKey(cfg)
	if apiKey == "" {
		logs.Fatal("API key not found in config file or environment variable")
	}

	// 初始化Anthropic客户端
	clientOptions := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	// 如果配置了BaseURL，使用它
	if cfg.BaseURL != "" {
		clientOptions = append(clientOptions, option.WithBaseURL(cfg.BaseURL))
	}

	client := anthropic.NewClient(clientOptions...)
	ctx := context.Background()

	// 初始化工具注册表
	toolRegistry := tools.NewToolRegistry(skillCatalog)

	// 初始化消息处理器
	messageHandler := messages.NewMessageHandler(skillCatalog, toolRegistry)

	// 初始化输入处理器
	inputHandler := input.NewInputHandler(skillCatalog)

	// 显示启动信息
	fmt.Println()
	fmt.Println("=== Ant Agent 启动成功 ===")
	fmt.Println()
	fmt.Println("📋 可用指令:")
	fmt.Println("  help, ?          - 显示帮助信息")
	fmt.Println("  exit, q          - 退出 agent")
	fmt.Println("  install-skill <path> - 从目录安装技能")
	fmt.Println("  remove-skill <name> - 删除指定技能")
	fmt.Println("  show-skills      - 显示所有可用技能")
	fmt.Println()

	// 显示已安装的技能
	skills := skillCatalog.GetSkills()
	if len(skills) > 0 {
		fmt.Println("🧩 已安装技能:")
		for name, skill := range skills {
			// 截断技能描述，保持显示简洁
			description := skill.Description
			maxLength := 80
			if len(description) > maxLength {
				description = description[:maxLength] + "..."
			}
			fmt.Printf("  - %s: %s\n", name, description)
		}
	} else {
		fmt.Println("🧩 暂无已安装技能")
	}
	fmt.Println()

	// 显示大模型访问地址
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	fmt.Printf("🌐 大模型访问地址: %s\n", baseURL)
	fmt.Println()
	fmt.Println("💡 提示:")
	fmt.Println("  - 输入 'help' 查看详细帮助")
	fmt.Println("  - 输入 'exit' 或 'q' 退出 agent")
	fmt.Println()
	for {
		// 读取用户输入
		userInput, err := inputHandler.ReadInput()
		if err != nil {
			logs.Error("Error reading input: %v", err)
			continue
		}

		// 处理用户输入
		processedInput, isSpecialCommand, err := inputHandler.ProcessInput(userInput)
		if err != nil {
			logs.Error("Error processing input: %v", err)
			continue
		}

		// 如果是特殊命令，继续下一轮循环
		if isSpecialCommand {
			continue
		}

		// 处理消息
		configMap := map[string]interface{}{
			"model":      cfg.Model,
			"max_tokens": cfg.MaxTokens,
		}
		if err := messageHandler.ProcessMessage(ctx, client, processedInput, configMap); err != nil {
			logs.Error("Error processing message: %v", err)
		}
	}
}
