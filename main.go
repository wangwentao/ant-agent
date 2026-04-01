package main

import (
	"context"
	"fmt"
	"os"

	"ant-agent/internal/acp"
	"ant-agent/internal/config"
	"ant-agent/internal/input"
	"ant-agent/internal/messages"
	"ant-agent/internal/skills"
	"ant-agent/internal/tools"
	"ant-agent/internal/utils"
)

func main() {
	// 解析命令行参数
	args := utils.ParseCommandLineArgs()

	// 处理命令行参数
	if utils.HandleCommandLineArgs(args) {
		return
	}

	// 加载配置
	appConfig, err := utils.LoadConfig()
	if err != nil {
		appConfig = config.DefaultConfig()
	}

	// 如果指定了模型，覆盖配置中的模型
	if args.Model != "" {
		appConfig.Model = args.Model
	}

	// 设置流式响应
	appConfig.Stream = args.Stream

	// 初始化技能系统
	skillCatalog := skills.NewSkillCatalog()
	if err := skillCatalog.DiscoverSkills(); err != nil {
		fmt.Printf("failed to discover skills: %v\n", err)
	}

	// 初始化LLM客户端
	client, err := utils.InitLLMClient(appConfig)
	if err != nil {
		return
	}
	ctx := context.Background()

	// 初始化工具注册表
	toolRegistry := tools.NewToolRegistry(skillCatalog)

	// 初始化 MCP 模块
	mcpManager := utils.InitMCPManager(ctx, toolRegistry)
	if mcpManager != nil {
		// 处理程序退出时停止 MCP 服务
		defer mcpManager.StopAll()
		// 设置信号处理器
		utils.SetupSignalHandler(mcpManager)
	}

	// 初始化消息处理器
	messageHandler := messages.NewMessageHandler(skillCatalog, toolRegistry, appConfig)

	// 初始化输入处理器
	inputHandler := input.NewInputHandler(skillCatalog, toolRegistry, appConfig.AgentName)

	// 处理 ACP 模式
	if args.ACPmode {
		// 设置 ACP 模式
		appConfig.ACPMode = true

		// 创建 ACP 服务器
		server := acp.NewServer(client, skillCatalog, toolRegistry, appConfig)

		// 启动 ACP 服务器
		if err := server.Start(ctx); err != nil {
			fmt.Printf("ACP server error: %v\n", err)
			return
		}

		return
	}

	// 处理 WeClaw 兼容性
	if utils.HandleWeClawCompatibility(args, appConfig, messageHandler, client) {
		return
	}

	// 显示启动信息
	utils.DisplayStartupInfo(appConfig, skillCatalog)

	// 如果是非交互式模式，直接退出
	if args.NonInteractive {
		fmt.Println("Running in non-interactive mode. Exiting...")
		return
	}

	// 交互式模式：持续等待用户输入
	for {
		// 读取用户输入
		userInput, err := inputHandler.ReadInput()
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			// 在非交互式环境中，输入错误可能是正常的，直接退出
			if !utils.IsTerminal(os.Stdin) {
				fmt.Println("Non-terminal environment detected, exiting...")
				return
			}
			continue
		}

		// 处理用户输入
		processedInput, isSpecialCommand, err := inputHandler.ProcessInput(userInput)
		if err != nil {
			fmt.Printf("Error processing input: %v\n", err)
			continue
		}

		// 如果是特殊命令，继续下一轮循环
		if isSpecialCommand {
			continue
		}

		// 处理消息
		if _, err := messageHandler.ProcessMessage(ctx, client, processedInput, appConfig); err != nil {
			fmt.Printf("Error processing message: %v\n", err)
		}
	}
}
