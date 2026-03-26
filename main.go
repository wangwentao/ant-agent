package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"ant-agent/internal/acp"
	"ant-agent/internal/config"
	"ant-agent/internal/input"
	"ant-agent/internal/logs"
	"ant-agent/internal/messages"
	"ant-agent/internal/skills"
	"ant-agent/internal/tools"
	"ant-agent/internal/weclaw"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func main() {
	// 解析命令行参数
	var (
		help           bool
		version        bool
		nonInteractive bool
		// 模拟 claude code 的行为，添加 -p 参数支持
		pFlag string
		// 模拟 claude code 的行为，添加以下参数
		outputFormat    string
		resumeSessionID string
		model           string
		systemPrompt    string
		// 模拟 claude code 的行为，添加 --verbose 参数（忽略实际值）
		verbose bool
		// 流式响应参数
		stream bool
		// ACP 模式参数
		acpMode bool
	)

	flag.BoolVar(&help, "help", false, "Show help information")
	flag.BoolVar(&version, "version", false, "Show version information")
	flag.BoolVar(&nonInteractive, "non-interactive", false, "Run in non-interactive mode")
	flag.StringVar(&pFlag, "p", "", "Compatibility parameter for clawbot")
	flag.StringVar(&outputFormat, "output-format", "", "Output format (stream-json for WeClaw compatibility)")
	flag.StringVar(&resumeSessionID, "resume", "", "Resume existing session")
	flag.StringVar(&model, "model", "", "Specify model to use")
	flag.StringVar(&systemPrompt, "append-system-prompt", "", "Append system prompt")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output (ignored)")
	flag.BoolVar(&stream, "stream", true, "Use streaming response from model")
	flag.BoolVar(&acpMode, "acp", false, "Run in ACP mode")
	flag.Parse()

	// 处理命令行参数
	if help {
		fmt.Println("Ant Agent - A smart agent based on Anthropic API")
		fmt.Println("Usage: ant-agent [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	}

	if version {
		fmt.Println("Ant Agent version 1.0.0")
		return
	}

	// 加载配置
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		logs.Warn("%v", err)
		// 继续执行，使用默认配置
		appConfig = config.DefaultConfig()
	}

	// 如果指定了模型，覆盖配置中的模型
	if model != "" {
		appConfig.Model = model
	}

	// 设置流式响应
	appConfig.Stream = stream

	// 初始化技能系统
	skillCatalog := skills.NewSkillCatalog()
	if err := skillCatalog.DiscoverSkills(); err != nil {
		logs.Warn("failed to discover skills: %v", err)
	}

	// 获取API Key
	apiKey := appConfig.GetAPIKey()
	if apiKey == "" {
		// 在交互式模式下，仍然直接退出
		logs.Fatal("API key not found in config file or environment variable")
	}

	// 初始化LLM客户端
	clientOptions := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	// 如果配置了BaseURL，使用它
	if appConfig.BaseURL != "" {
		clientOptions = append(clientOptions, option.WithBaseURL(appConfig.BaseURL))
	}

	client := anthropic.NewClient(clientOptions...)
	ctx := context.Background()

	// 初始化工具注册表
	toolRegistry := tools.NewToolRegistry(skillCatalog)

	// 初始化消息处理器
	messageHandler := messages.NewMessageHandler(skillCatalog, toolRegistry, appConfig)

	// 初始化输入处理器
	inputHandler := input.NewInputHandler(skillCatalog)

	// 处理 ACP 模式
	if acpMode {
		// 设置 ACP 模式
		appConfig.ACPMode = true

		// 创建 ACP 服务器
		server := acp.NewServer(client, skillCatalog, toolRegistry, appConfig)

		// 启动 ACP 服务器
		ctx := context.Background()
		if err := server.Start(ctx); err != nil {
			logs.Error("ACP server error: %v", err)
			return
		}

		return
	}

	// 处理 -p 参数（WeClaw 兼容性）
	if pFlag != "" {
		// 生成会话ID
		sessionID := weclaw.GenerateSessionID(resumeSessionID)

		// 设置 WeClaw 相关配置
		appConfig.OutputFormat = outputFormat
		appConfig.SessionID = sessionID

		// 处理系统提示
		if systemPrompt != "" {
			appConfig.SystemPrompt = systemPrompt
		}

		// 处理消息并获取结果
		var result string
		var err error

		// 处理消息
		result, err = messageHandler.ProcessMessage(ctx, client, pFlag, appConfig)
		if err != nil {
			result = fmt.Sprintf("Error: %v", err)
			if weclaw.ShouldUseStreamJSON(outputFormat) {
				// 输出错误事件
				weclaw.OutputResultEvent(sessionID, result, true)
			} else {
				fmt.Println(result)
			}
			return
		}

		return
	}

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
	baseURL := appConfig.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	fmt.Printf("🌐 大模型访问地址: %s\n", baseURL)
	fmt.Println()
	fmt.Println("💡 提示:")
	fmt.Println("  - 输入 'help' 查看详细帮助")
	fmt.Println("  - 输入 'exit' 或 'q' 退出 agent")
	fmt.Println()

	// 如果是非交互式模式，直接退出
	if nonInteractive {
		fmt.Println("Running in non-interactive mode. Exiting...")
		return
	}

	// 交互式模式：持续等待用户输入
	for {
		// 读取用户输入
		userInput, err := inputHandler.ReadInput()
		if err != nil {
			logs.Error("Error reading input: %v", err)
			// 在非交互式环境中，输入错误可能是正常的，直接退出
			if !isTerminal(os.Stdin) {
				logs.Info("Non-terminal environment detected, exiting...")
				return
			}
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
		if _, err := messageHandler.ProcessMessage(ctx, client, processedInput, appConfig); err != nil {
			logs.Error("Error processing message: %v", err)
		}
	}
}

// isTerminal 检查文件描述符是否是终端
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
