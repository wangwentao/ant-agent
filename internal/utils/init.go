package utils

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ant-agent/internal/config"
	"ant-agent/internal/logs"
	"ant-agent/internal/mcp"
	"ant-agent/internal/messages"
	"ant-agent/internal/skills"
	"ant-agent/internal/tools"
	"ant-agent/internal/weclaw"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// CommandLineArgs 命令行参数
type CommandLineArgs struct {
	Help            bool
	Version         bool
	NonInteractive  bool
	PFlag           string
	OutputFormat    string
	ResumeSessionID string
	Model           string
	SystemPrompt    string
	Verbose         bool
	Stream          bool
	ACPmode         bool
}

// ParseCommandLineArgs 解析命令行参数
func ParseCommandLineArgs() *CommandLineArgs {
	var args CommandLineArgs

	flag.BoolVar(&args.Help, "help", false, "Show help information")
	flag.BoolVar(&args.Version, "version", false, "Show version information")
	flag.BoolVar(&args.NonInteractive, "non-interactive", false, "Run in non-interactive mode")
	flag.StringVar(&args.PFlag, "p", "", "Compatibility parameter for clawbot")
	flag.StringVar(&args.OutputFormat, "output-format", "", "Output format (stream-json for WeClaw compatibility)")
	flag.StringVar(&args.ResumeSessionID, "resume", "", "Resume existing session")
	flag.StringVar(&args.Model, "model", "", "Specify model to use")
	flag.StringVar(&args.SystemPrompt, "append-system-prompt", "", "Append system prompt")
	flag.BoolVar(&args.Verbose, "verbose", false, "Verbose output (ignored)")
	flag.BoolVar(&args.Stream, "stream", true, "Use streaming response from model")
	flag.BoolVar(&args.ACPmode, "acp", false, "Run in ACP mode")
	flag.Parse()

	return &args
}

// HandleCommandLineArgs 处理命令行参数
func HandleCommandLineArgs(args *CommandLineArgs) bool {
	if args.Help {
		fmt.Println("Ant Agent - A smart agent based on Anthropic API")
		fmt.Println("Usage: ant-agent [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
		return true
	}

	if args.Version {
		fmt.Println("Ant Agent version 1.0.0")
		return true
	}

	return false
}

// LoadConfig 加载配置
func LoadConfig() (*config.AppConfig, error) {
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		logs.Warn("%v", err)
		// 继续执行，使用默认配置
		appConfig = config.DefaultConfig()
	}

	// 设置日志级别
	switch appConfig.LogLevel {
	case "debug":
		logs.SetLevel(logs.DEBUG)
	case "info":
		logs.SetLevel(logs.INFO)
	case "warn":
		logs.SetLevel(logs.WARN)
	case "error":
		logs.SetLevel(logs.ERROR)
	case "fatal":
		logs.SetLevel(logs.FATAL)
	default:
		logs.SetLevel(logs.INFO)
		logs.Warn("Invalid log level: %s, using info level instead", appConfig.LogLevel)
	}

	return appConfig, nil
}

// InitLLMClient 初始化LLM客户端
func InitLLMClient(appConfig *config.AppConfig) (anthropic.Client, error) {
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
	return client, nil
}

// InitMCPManager 初始化MCP管理器
func InitMCPManager(ctx context.Context, toolRegistry *tools.ToolRegistry) *mcp.Manager {
	mcpManager, err := mcp.NewManager()
	if err != nil {
		logs.Warn("Failed to initialize MCP manager: %v", err)
		return nil
	}

	// 获取 MCP 工具并注册到工具注册表
	mcpTools, err := mcpManager.GetTools(ctx)
	if err != nil {
		logs.Warn("Failed to get MCP tools: %v", err)
	} else {
		for _, tool := range mcpTools {
			toolRegistry.Register(tool)
			logs.Debug("Registered MCP tool: %s from server %s", tool.Name(), tool.ServerName)
		}
	}

	return mcpManager
}

// IsTerminal 检查文件描述符是否是终端
func IsTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// SetupSignalHandler 设置信号处理器
func SetupSignalHandler(mcpManager *mcp.Manager) {
	if mcpManager == nil {
		return
	}

	// 处理信号，确保 MCP 服务能够正确停止
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		mcpManager.StopAll()
		os.Exit(0)
	}()
}

// DisplayStartupInfo 显示启动信息
func DisplayStartupInfo(appConfig *config.AppConfig, skillCatalog *skills.SkillCatalog) {
	fmt.Println()
	fmt.Printf("=== %s 启动成功 ===\n", appConfig.AgentName)
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
}

// HandleWeClawCompatibility 处理WeClaw兼容性
func HandleWeClawCompatibility(args *CommandLineArgs, appConfig *config.AppConfig, messageHandler *messages.MessageHandler, client anthropic.Client) bool {
	if args.PFlag != "" {
		// 生成会话ID
		sessionID := weclaw.GenerateSessionID(args.ResumeSessionID)

		// 设置 WeClaw 相关配置
		appConfig.OutputFormat = args.OutputFormat
		appConfig.SessionID = sessionID

		// 处理系统提示
		if args.SystemPrompt != "" {
			appConfig.SystemPrompt = args.SystemPrompt
		}

		// 处理消息并获取结果
		ctx := context.Background()
		result, err := messageHandler.ProcessMessage(ctx, client, args.PFlag, appConfig)
		if err != nil {
			result = fmt.Sprintf("Error: %v", err)
			if weclaw.ShouldUseStreamJSON(args.OutputFormat) {
				// 输出错误事件
				weclaw.OutputResultEvent(sessionID, result, true)
			} else {
				fmt.Println(result)
			}
			return true
		}

		return true
	}

	return false
}
