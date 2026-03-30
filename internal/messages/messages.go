package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ant-agent/internal/config"
	"ant-agent/internal/logs"
	"ant-agent/internal/skills"
	"ant-agent/internal/tools"
	"ant-agent/internal/weclaw"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/coder/acp-go-sdk"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	skillCatalog *skills.SkillCatalog
	toolRegistry *tools.ToolRegistry
	appConfig    *config.AppConfig
}

// NewMessageHandler 创建一个新的消息处理器
func NewMessageHandler(skillCatalog *skills.SkillCatalog, toolRegistry *tools.ToolRegistry, appConfig *config.AppConfig) *MessageHandler {
	return &MessageHandler{
		skillCatalog: skillCatalog,
		toolRegistry: toolRegistry,
		appConfig:    appConfig,
	}
}

// BuildSystemPrompt 构建系统提示
func (h *MessageHandler) BuildSystemPrompt() string {
	agentName := h.appConfig.AgentName
	systemPrompt := "Your name is " + agentName + ". You are a comprehensive, clear-thinking, and logically rigorous personal assistant. You can complete most work and life tasks, and for tasks in areas where you are not proficient, you can compensate by using skills.\n\n" +
		"## Available Skills\n" +
		"Here are the skills you can use:\n"

	// 添加已发现的技能
	skills := h.skillCatalog.GetSkills()
	if len(skills) > 0 {
		for name, skill := range skills {
			systemPrompt += "- **" + name + "**: " + skill.Description + "\n"
		}
	} else {
		systemPrompt += "- No skills available\n"
	}
	systemPrompt += "\nYou can view all skills using the `list_skills` tool, and activate specific skills using the `activate_skill` tool to get detailed instructions."

	return systemPrompt
}

// BuildSystemBlocks 构建系统提示块
func (h *MessageHandler) BuildSystemBlocks() []anthropic.TextBlockParam {
	systemPrompt := h.BuildSystemPrompt()
	return []anthropic.TextBlockParam{
		{
			Type: "text",
			Text: systemPrompt,
		},
	}
}

// ProcessMessage 处理消息并返回结果
// 支持 ACP 模式下的流式输出
func (h *MessageHandler) ProcessMessage(ctx context.Context, client anthropic.Client, userInput string, appConfig *config.AppConfig) (string, error) {
	// 构建消息历史
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userInput)),
	}

	// 循环处理工具调用
	var finalResult string
	for {
		// 生成工具模式
		toolSchemas := tools.GenerateToolSchemas(h.toolRegistry)

		// 构建系统提示
		systemBlocks := h.buildSystemBlocks(appConfig.SystemPrompt)

		if appConfig.Stream {
			// 使用流式响应
			result, hasToolCall, err := h.processStreamingResponse(ctx, client, &messages, toolSchemas, systemBlocks, appConfig)
			if err != nil {
				return "", err
			}
			finalResult += result

			// 如果没有工具调用，退出循环
			if !hasToolCall {
				break
			}
		} else {
			// 使用非流式响应
			result, hasToolCall, err := h.processNonStreamingResponse(ctx, client, &messages, toolSchemas, systemBlocks, appConfig)
			if err != nil {
				return "", err
			}
			finalResult += result

			// 如果没有工具调用，退出循环
			if !hasToolCall {
				break
			}
		}
	}

	return strings.TrimSpace(finalResult), nil
}

// buildSystemBlocks 构建系统提示块
func (h *MessageHandler) buildSystemBlocks(customPrompt string) []anthropic.TextBlockParam {
	if customPrompt != "" {
		return []anthropic.TextBlockParam{
			{
				Type: "text",
				Text: customPrompt,
			},
		}
	}
	return h.BuildSystemBlocks()
}

// processStreamingResponse 处理流式响应
func (h *MessageHandler) processStreamingResponse(ctx context.Context, client anthropic.Client, messages *[]anthropic.MessageParam, toolSchemas []anthropic.ToolUnionParam, systemBlocks []anthropic.TextBlockParam, appConfig *config.AppConfig) (string, bool, error) {
	// 使用真正的流式响应
	stream := client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     appConfig.Model,
		Messages:  *messages,
		MaxTokens: int64(appConfig.MaxTokens),
		Tools:     toolSchemas,
		System:    systemBlocks,
	})
	defer stream.Close()

	// 处理流式响应
	var hasToolCall bool
	var toolResults []anthropic.ContentBlockParamUnion
	var currentText string
	var assistantContent []anthropic.ContentBlockParamUnion

	// 用于存储工具调用的信息
	type toolCallInfo struct {
		id           string
		name         string
		input        map[string]interface{}
		toolUseBlock anthropic.ToolUseBlock
	}
	var currentToolCall *toolCallInfo
	var currentToolInput string

	// 用于跟踪是否已经显示了前缀和是否有实际内容
	var prefixShown bool
	var hasContent bool

	for stream.Next() {
		event := stream.Current()

		// 处理不同类型的事件
		switch event.Type {
		case "content_block_delta":
			// 处理内容块增量
			// 检查文本增量
			if event.Delta.Text != "" {
				// 检查是否有非空白内容
				if strings.TrimSpace(event.Delta.Text) != "" {
					hasContent = true
				}

				// 累积文本
				currentText += event.Delta.Text

				// 处理 ACP 模式或普通输出
				h.handleDeltaOutput(ctx, event.Delta.Text, appConfig, &prefixShown, hasContent)
			}

			// 检查工具输入的增量（PartialJSON）
			if currentToolCall != nil {
				// 捕获所有的 PartialJSON，即使为空
				if event.Delta.PartialJSON != "" {
					currentToolInput += event.Delta.PartialJSON
					logs.Debug("Received tool input delta: %s", event.Delta.PartialJSON)
				}
				// 记录当前累积的工具输入
				logs.Debug("Current tool input: %s", currentToolInput)
			}
		case "content_block_start":
			// 检查是否是工具调用
			if event.ContentBlock.Type == "tool_use" {
				hasToolCall = true
				// 转换为ToolUseBlock以正确获取字段
				toolUseBlock := event.ContentBlock.AsToolUse()

				// 初始化工具调用信息
				currentToolCall = &toolCallInfo{
					id:           toolUseBlock.ID,
					name:         toolUseBlock.Name,
					input:        make(map[string]interface{}),
					toolUseBlock: toolUseBlock,
				}
				currentToolInput = ""
			} else if event.ContentBlock.Type == "text" {
				// 添加文本内容到助手内容
				assistantContent = append(assistantContent, anthropic.NewTextBlock(currentText))
			}
		case "content_block_stop":
			// 当内容块结束时，处理工具调用
			if currentToolCall != nil {
				// 首先尝试从 event.ContentBlock 中获取工具输入
				toolUseBlock := event.ContentBlock.AsToolUse()

				// 解析工具输入
				input := h.parseToolInput(currentToolInput, toolUseBlock)
				logs.Debug("Tool input parsed: %v", input)

				// 添加工具调用到助手内容
				assistantContent = append(assistantContent, anthropic.NewToolUseBlock(currentToolCall.id, input, currentToolCall.name))

				// 执行工具调用
				toolResult := h.toolRegistry.RunTool(ctx, currentToolCall.name, input)

				// 创建工具结果块，使用工具调用的ID
				toolResultBlock := anthropic.NewToolResultBlock(currentToolCall.id, toolResult, false)
				toolResults = append(toolResults, toolResultBlock)

				// 重置当前工具调用
				currentToolCall = nil
				currentToolInput = ""
			}
		}
	}

	// 处理最终输出
	h.handleFinalOutput(currentText, appConfig, prefixShown)

	// 检查流是否有错误
	if stream.Err() != nil {
		logs.Error("Stream error: %v", stream.Err())
		return "", false, stream.Err()
	}

	// 如果有工具调用，添加工具结果到消息历史
	if hasToolCall {
		*messages = h.updateMessageHistory(*messages, assistantContent, toolResults)
	}

	return currentText, hasToolCall, nil
}

// processNonStreamingResponse 处理非流式响应
func (h *MessageHandler) processNonStreamingResponse(ctx context.Context, client anthropic.Client, messages *[]anthropic.MessageParam, toolSchemas []anthropic.ToolUnionParam, systemBlocks []anthropic.TextBlockParam, appConfig *config.AppConfig) (string, bool, error) {
	// 使用非流式响应
	response, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     appConfig.Model,
		Messages:  *messages,
		MaxTokens: int64(appConfig.MaxTokens),
		Tools:     toolSchemas,
		System:    systemBlocks,
	})
	if err != nil {
		logs.Error("%v", err)
		return "", false, err
	}

	// 检查是否有工具调用
	if len(response.Content) == 0 {
		logs.Error("Empty response from model")
		return "", false, fmt.Errorf("empty response from model")
	}

	// 处理响应内容
	hasToolCall := false
	var toolResults []anthropic.ContentBlockParamUnion
	var currentText string
	var assistantContent []anthropic.ContentBlockParamUnion

	for _, content := range response.Content {
		switch content.Type {
		case "text":
			// 非工具调用，输出结果
			textContent := strings.TrimSpace(content.Text)
			if textContent != "" {
				currentText += textContent + "\n"
				// 同时输出到标准输出，保持原有行为
				fmt.Printf("\033[32m%s »:\033[0m %s\n", appConfig.AgentName, textContent)
			}
			assistantContent = append(assistantContent, anthropic.NewTextBlock(content.Text))
		case "tool_use":
			// 工具调用
			hasToolCall = true
			// 只在调试模式下显示工具调用日志
			logs.Debug("Tool Call: %s (ID: %s)", content.Name, content.ID)

			// 解析工具输入
			var input map[string]interface{}
			if err := json.Unmarshal(content.Input, &input); err != nil {
				logs.Debug("Error parsing tool input: %v", err)
				input = make(map[string]interface{})
			}

			// 执行工具调用
			toolResult := h.toolRegistry.RunTool(ctx, content.Name, input)

			// 创建工具结果块，使用工具调用的ID
			toolResultBlock := anthropic.NewToolResultBlock(content.ID, toolResult, false)
			toolResults = append(toolResults, toolResultBlock)

			// 添加工具调用到助手内容
			assistantContent = append(assistantContent, anthropic.NewToolUseBlock(content.ID, input, content.Name))
		}
	}

	// 如果有工具调用，添加工具结果到消息历史
	if hasToolCall {
		*messages = h.updateMessageHistory(*messages, assistantContent, toolResults)
	}

	return currentText, hasToolCall, nil
}

// handleDeltaOutput 处理增量输出
func (h *MessageHandler) handleDeltaOutput(ctx context.Context, deltaText string, appConfig *config.AppConfig, prefixShown *bool, hasContent bool) {
	// ACP 模式，发送会话更新通知
	if appConfig.ACPMode && appConfig.ACPConn != nil && appConfig.ACPSessionID != "" {
		h.sendACPSessionUpdate(ctx, appConfig.ACPConn, appConfig.ACPSessionID, deltaText)
	} else if appConfig.OutputFormat != "stream-json" {
		// 非 weclaw 模式，进行普通的流式输出
		// 如果还没有显示前缀且有实际内容，先显示前缀
		if !*prefixShown && hasContent {
			fmt.Printf("\033[32m%s »:\033[0m ", appConfig.AgentName)
			*prefixShown = true
		}
		// 实时输出文本内容，保留原始格式（包括换行符）
		if *prefixShown {
			fmt.Print(deltaText)
		}
	}
}

// sendACPSessionUpdate 发送 ACP 会话更新
func (h *MessageHandler) sendACPSessionUpdate(ctx context.Context, acpConn interface{}, acpSessionID string, deltaText string) {
	// 检查 acpConn 是否实现了 SessionUpdate 方法
	if conn, ok := acpConn.(interface {
		SessionUpdate(context.Context, acp.SessionNotification) error
	}); ok {
		// 构建正确的 ACP SessionNotification
		acpSessionId := acp.SessionId(acpSessionID)
		update := acp.UpdateAgentMessageText(deltaText)
		notification := acp.SessionNotification{
			SessionId: acpSessionId,
			Update:    update,
		}
		// 发送会话更新
		_ = conn.SessionUpdate(ctx, notification)
	}
}

// handleFinalOutput 处理最终输出
func (h *MessageHandler) handleFinalOutput(currentText string, appConfig *config.AppConfig, prefixShown bool) {
	// 非 weclaw 模式下，只有当显示了前缀时才输出换行
	if !weclaw.ShouldUseStreamJSON(appConfig.OutputFormat) && prefixShown {
		// 确保输出以换行结束，保持格式一致性
		if !strings.HasSuffix(currentText, "\n") {
			fmt.Println()
		}
	} else if weclaw.ShouldUseStreamJSON(appConfig.OutputFormat) && appConfig.SessionID != "" && currentText != "" {
		// 在 weclaw 模式下，只输出一个包含完整内容的 result 事件
		weclaw.OutputResultEvent(appConfig.SessionID, currentText, false)
	}
}

// parseToolInput 解析工具输入
func (h *MessageHandler) parseToolInput(currentToolInput string, toolUseBlock anthropic.ToolUseBlock) map[string]interface{} {
	var input map[string]interface{}

	// 首先尝试使用收集到的增量输入
	if currentToolInput != "" {
		if err := json.Unmarshal([]byte(currentToolInput), &input); err == nil && len(input) > 0 {
			logs.Debug("Using incremental tool input: %v", input)
			return input
		}
	}

	// 然后尝试从 toolUseBlock 中获取工具输入
	if toolUseBlock.Input != nil && len(toolUseBlock.Input) > 0 {
		if err := json.Unmarshal(toolUseBlock.Input, &input); err == nil && len(input) > 0 {
			logs.Debug("Using tool input from ToolUseBlock: %v", input)
			return input
		}
	}

	// 最后，尝试从 toolUseBlock 的原始数据中提取
	// 这是一个安全的回退机制，确保我们能获取到工具输入
	rawJSON := toolUseBlock.RawJSON()
	if rawJSON != "" {
		var rawToolUse struct {
			Input json.RawMessage `json:"input"`
		}
		if err := json.Unmarshal([]byte(rawJSON), &rawToolUse); err == nil && rawToolUse.Input != nil && len(rawToolUse.Input) > 0 {
			if err := json.Unmarshal(rawToolUse.Input, &input); err == nil && len(input) > 0 {
				logs.Debug("Using tool input from raw JSON: %v", input)
				return input
			}
		}
	}

	// 如果所有尝试都失败，返回空 map
	logs.Debug("Failed to parse tool input, returning empty map")
	return make(map[string]interface{})
}

// updateMessageHistory 更新消息历史
func (h *MessageHandler) updateMessageHistory(messages []anthropic.MessageParam, assistantContent []anthropic.ContentBlockParamUnion, toolResults []anthropic.ContentBlockParamUnion) []anthropic.MessageParam {
	// 先添加assistant的响应消息（包含tool_use）
	if len(assistantContent) > 0 {
		messages = append(messages, anthropic.NewAssistantMessage(assistantContent...))
	}

	// 再添加user消息（包含tool_result）
	messages = append(messages, anthropic.NewUserMessage(toolResults...))

	return messages
}
