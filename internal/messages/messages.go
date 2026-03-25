package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ant-agent/internal/logs"
	"ant-agent/internal/skills"
	"ant-agent/internal/tools"
	"ant-agent/internal/weclaw"

	"github.com/anthropics/anthropic-sdk-go"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	skillCatalog *skills.SkillCatalog
	toolRegistry *tools.ToolRegistry
}

// NewMessageHandler 创建一个新的消息处理器
func NewMessageHandler(skillCatalog *skills.SkillCatalog, toolRegistry *tools.ToolRegistry) *MessageHandler {
	return &MessageHandler{
		skillCatalog: skillCatalog,
		toolRegistry: toolRegistry,
	}
}

// BuildSystemPrompt 构建系统提示
func (h *MessageHandler) BuildSystemPrompt() string {
	systemPrompt := "You are a comprehensive, clear-thinking, and logically rigorous personal assistant. You can complete most work and life tasks, and for tasks in areas where you are not proficient, you can compensate by using skills.\n\n" +
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
func (h *MessageHandler) ProcessMessage(ctx context.Context, client anthropic.Client, userInput string, config map[string]interface{}) (string, error) {
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
		systemBlocks := h.BuildSystemBlocks()

		// 检查是否有自定义系统提示
		if systemPrompt, ok := config["system_prompt"].(string); ok && systemPrompt != "" {
			systemBlocks = []anthropic.TextBlockParam{
				{
					Type: "text",
					Text: systemPrompt,
				},
			}
		}

		// 检查是否使用流式响应
		streamEnabled := false
		if streamVal, ok := config["stream"].(bool); ok {
			streamEnabled = streamVal
		}

		// 检查是否是 weclaw 模式（需要输出 JSON 事件）
		outputFormat := ""
		if of, ok := config["output_format"].(string); ok {
			outputFormat = of
		}

		// 获取会话 ID（weclaw 模式下需要）
		sessionID := ""
		if sid, ok := config["session_id"].(string); ok {
			sessionID = sid
		}

		if streamEnabled {
			// 使用真正的流式响应
			stream := client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
				Model:     config["model"].(string),
				Messages:  messages,
				MaxTokens: int64(config["max_tokens"].(int)),
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
				id    string
				name  string
				input map[string]interface{}
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
					if event.Delta.Text != "" {
						// 检查是否有非空白内容
						if strings.TrimSpace(event.Delta.Text) != "" {
							hasContent = true
						}

						// 累积文本
						currentText += event.Delta.Text

						// 非 weclaw 模式，进行普通的流式输出
						if outputFormat != "stream-json" {
							// 如果还没有显示前缀且有实际内容，先显示前缀
							if !prefixShown && hasContent {
								fmt.Print("\033[32mAnt »:\033[0m ")
								prefixShown = true
							}
							// 实时输出文本内容，保留原始格式（包括换行符）
							if prefixShown {
								fmt.Print(event.Delta.Text)
							}
						}
					}
					// 检查是否是工具输入的增量
					if currentToolCall != nil && event.Delta.PartialJSON != "" {
						currentToolInput += event.Delta.PartialJSON
					}
				case "content_block_start":
					// 检查是否是工具调用
					if event.ContentBlock.Type == "tool_use" {
						hasToolCall = true
						// 转换为ToolUseBlock以正确获取字段
						toolUseBlock := event.ContentBlock.AsToolUse()

						// 初始化工具调用信息
						currentToolCall = &toolCallInfo{
							id:    toolUseBlock.ID,
							name:  toolUseBlock.Name,
							input: make(map[string]interface{}),
						}
						currentToolInput = ""
					} else if event.ContentBlock.Type == "text" {
						// 添加文本内容到助手内容
						assistantContent = append(assistantContent, anthropic.NewTextBlock(currentText))
					}
				case "content_block_stop":
					// 当内容块结束时，处理工具调用
					if currentToolCall != nil {
						// 解析工具输入
						var input map[string]interface{}
						if currentToolInput != "" {
							// 使用收集到的增量输入
							if err := json.Unmarshal([]byte(currentToolInput), &input); err != nil {
								input = make(map[string]interface{})
							}
						} else {
							// 使用原始输入
							toolUseBlock := event.ContentBlock.AsToolUse()
							if err := json.Unmarshal(toolUseBlock.Input, &input); err != nil {
								input = make(map[string]interface{})
							}
						}

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

			// 非 weclaw 模式下，只有当显示了前缀时才输出换行
			if !weclaw.ShouldUseStreamJSON(outputFormat) && prefixShown {
				// 确保输出以换行结束，保持格式一致性
				if !strings.HasSuffix(currentText, "\n") {
					fmt.Println()
				}
			} else if weclaw.ShouldUseStreamJSON(outputFormat) && sessionID != "" && currentText != "" {
				// 在 weclaw 模式下，只输出一个包含完整内容的 result 事件
				weclaw.OutputResultEvent(sessionID, currentText, false)
			}

			// 处理最终结果
			finalResult += currentText

			// 检查流是否有错误
			if stream.Err() != nil {
				logs.Error("Stream error: %v", stream.Err())
				return "", stream.Err()
			}

			// 如果有工具调用，添加工具结果到消息历史
			if hasToolCall {
				// 先添加assistant的响应消息（包含tool_use）
				if len(assistantContent) > 0 {
					messages = append(messages, anthropic.NewAssistantMessage(assistantContent...))
				}

				// 再添加user消息（包含tool_result）
				messages = append(messages, anthropic.NewUserMessage(toolResults...))
			}

			// 如果没有工具调用，退出循环
			if !hasToolCall {
				break
			}
		} else {
			// 使用非流式响应（原有方式）
			response, err := client.Messages.New(ctx, anthropic.MessageNewParams{
				Model:     config["model"].(string),
				Messages:  messages,
				MaxTokens: int64(config["max_tokens"].(int)),
				Tools:     toolSchemas,
				System:    systemBlocks,
			})
			if err != nil {
				logs.Error("%v", err)
				return "", err
			}

			// 检查是否有工具调用
			if len(response.Content) == 0 {
				logs.Error("Empty response from model")
				return "", fmt.Errorf("empty response from model")
			}

			// 处理响应内容
			hasToolCall := false
			var toolResults []anthropic.ContentBlockParamUnion

			for _, content := range response.Content {
				switch content.Type {
				case "text":
					// 非工具调用，输出结果
					textContent := strings.TrimSpace(content.Text)
					if textContent != "" {
						finalResult += textContent + "\n"
						// 同时输出到标准输出，保持原有行为
						fmt.Printf("\033[32mAnt »:\033[0m %s\n", textContent)
					}
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
				}
			}

			// 如果有工具调用，添加工具结果到消息历史
			if hasToolCall {
				// 先添加assistant的响应消息（包含tool_use）
				assistantContent := make([]anthropic.ContentBlockParamUnion, 0, len(response.Content))
				for _, content := range response.Content {
					switch content.Type {
					case "text":
						assistantContent = append(assistantContent, anthropic.NewTextBlock(content.Text))
					case "tool_use":
						var input map[string]interface{}
						json.Unmarshal(content.Input, &input)
						assistantContent = append(assistantContent, anthropic.NewToolUseBlock(content.ID, input, content.Name))
					}
				}
				messages = append(messages, anthropic.NewAssistantMessage(assistantContent...))

				// 再添加user消息（包含tool_result）
				messages = append(messages, anthropic.NewUserMessage(toolResults...))
			}

			// 如果没有工具调用，退出循环
			if !hasToolCall {
				break
			}
		}
	}

	return strings.TrimSpace(finalResult), nil
}
