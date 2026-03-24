package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ant-agent/internal/logs"
	"ant-agent/internal/skills"
	"ant-agent/internal/tools"

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

// ProcessMessage 处理消息
func (h *MessageHandler) ProcessMessage(ctx context.Context, client anthropic.Client, userInput string, config map[string]interface{}) error {
	// 构建消息历史
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userInput)),
	}

	// 循环处理工具调用
	for {
		// 生成工具模式
		toolSchemas := tools.GenerateToolSchemas(h.toolRegistry)

		// 构建系统提示
		systemBlocks := h.BuildSystemBlocks()

		// 发送消息到模型
		response, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     config["model"].(string),
			Messages:  messages,
			MaxTokens: int64(config["max_tokens"].(int)),
			Tools:     toolSchemas,
			System:    systemBlocks,
		})
		if err != nil {
			logs.Error("%v", err)
			break
		}

		// 检查是否有工具调用
		if len(response.Content) == 0 {
			logs.Error("Empty response from model")
			break
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
					fmt.Printf("Ant agent: %s\n", textContent)
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

	return nil
}
