package main

import (
	"context"
	"flag"
	"fmt"
	"os"

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
		// 集成 clawbot，添加 -p 参数支持
		pFlag string
		// 集成 WeClaw，添加以下参数
		outputFormat    string
		resumeSessionID string
		model           string
		systemPrompt    string
	)

	flag.BoolVar(&help, "help", false, "Show help information")
	flag.BoolVar(&version, "version", false, "Show version information")
	flag.BoolVar(&nonInteractive, "non-interactive", false, "Run in non-interactive mode")
	// 为了兼容 clawbot，添加 -p 参数支持
	flag.StringVar(&pFlag, "p", "", "Compatibility parameter for clawbot")
	// 为了兼容 WeClaw，添加以下参数
	flag.StringVar(&outputFormat, "output-format", "", "Output format (stream-json for WeClaw compatibility)")
	flag.StringVar(&resumeSessionID, "resume", "", "Resume existing session")
	flag.StringVar(&model, "model", "", "Specify model to use")
	flag.StringVar(&systemPrompt, "append-system-prompt", "", "Append system prompt")
	// 为了兼容 WeClaw，添加 --verbose 参数（忽略实际值）
	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "Verbose output (ignored)")
	// 添加流式响应参数
	var stream bool
	flag.BoolVar(&stream, "stream", true, "Use streaming response from model")
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
	cfg, err := config.LoadConfig()
	if err != nil {
		logs.Warn("%v", err)
		// 继续执行，使用默认配置
	}

	// 如果指定了模型，覆盖配置中的模型
	if model != "" {
		cfg.Model = model
	}

	// 初始化技能系统
	skillCatalog := skills.NewSkillCatalog()
	if err := skillCatalog.DiscoverSkills(); err != nil {
		logs.Warn("failed to discover skills: %v", err)
	}

	// 获取API Key
	apiKey := config.GetAPIKey(cfg)
	if apiKey == "" {
		// 当被 WeClaw 调用时，返回错误而不是直接退出
		if pFlag != "" {
			result := "Error: API key not found in config file or environment variable"
			if weclaw.ShouldUseStreamJSON(outputFormat) {
				// 生成会话ID
				sessionID := weclaw.GenerateSessionID(resumeSessionID)
				// 输出错误事件
				weclaw.OutputResultEvent(sessionID, result, true)
			} else {
				fmt.Println(result)
			}
			return
		}
		// 在交互式模式下，仍然直接退出
		logs.Fatal("API key not found in config file or environment variable")
	}

	// 初始化LLM客户端
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

	// 处理 -p 参数（WeClaw 兼容性）
	if pFlag != "" {
		// 生成会话ID
		sessionID := weclaw.GenerateSessionID(resumeSessionID)

		// 配置映射，传递给消息处理器
		configMap := map[string]interface{}{
			"model":         cfg.Model,
			"max_tokens":    cfg.MaxTokens,
			"stream":        stream,
			"output_format": outputFormat,
			"session_id":    sessionID,
		}

		// 处理系统提示
		if systemPrompt != "" {
			configMap["system_prompt"] = systemPrompt
		}

		// 处理消息并获取结果
		var result string
		var err error

		// 处理消息
		result, err = messageHandler.ProcessMessage(ctx, client, pFlag, configMap)
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
		configMap := map[string]interface{}{
			"model":      cfg.Model,
			"max_tokens": cfg.MaxTokens,
			"stream":     stream,
		}
		if _, err := messageHandler.ProcessMessage(ctx, client, processedInput, configMap); err != nil {
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
