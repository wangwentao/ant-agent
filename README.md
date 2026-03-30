# Ant Agent

## 产品概览

Ant Agent 是一款基于 Go 语言开发的智能代理工具，通过与大模型深度集成，为用户提供智能交互、工具调用和技能扩展能力。它采用模块化设计，具备高度的可扩展性和灵活性，可作为独立应用运行，也可与 WeClaw 集成，为微信 Clawbot 提供智能服务。

### 核心价值

- **智能化交互**: 与大模型实时对话，支持流式响应，提供自然流畅的用户体验
- **工具集成**: 内置多种实用工具，如文件操作、命令执行等，扩展大模型能力边界
- **技能生态**: 支持从本地或 GitHub 安装技能，构建个性化的功能生态
- **多模式集成**: 支持 CLI 和 ACP 协议，与 WeClaw 无缝集成，为微信提供智能服务
- **高度可配置**: 通过配置文件和环境变量灵活配置，适应不同场景需求

## 功能特性

### 基础功能

- **持续交互**：支持命令行交互，具备命令补全功能，提供友好的用户体验
- **大模型集成**：与多种大模型兼容，支持流式响应，实现实时交互
- **工具调用**：自动处理工具调用循环，执行工具并返回结果
- **配置管理**：通过配置文件管理大模型访问信息，支持环境变量优先级
- **自定义能力**：支持自定义 API 基础 URL，适应不同部署环境

### 技能系统

- **技能发现**：自动发现项目内（`.ant/skills`）和用户主目录下（`~/.agent/skills`）的技能
- **技能安装**：支持从本地目录或 GitHub 仓库安装技能，可指定单个、多个或所有技能
- **技能管理**：通过命令删除指定技能，维护技能生态
- **技能激活**：通过工具激活指定技能以获取详细说明和使用方法

### 工具系统

- `execute_shell`：执行 shell 命令并返回输出
- `read_file`：读取文件内容
- `write_file`：写入内容到文件
- `edit_file`：通过替换字符串编辑文件
- `list_skills`：列出所有可用技能
- `activate_skill`：激活指定技能以获取详细说明

### 集成能力

- **WeClaw 集成**：支持 CLI 模式和 ACP 模式与 WeClaw 集成
- **微信 Clawbot 支持**：为微信用户提供智能对话、工具调用、技能扩展等能力
- **流式响应**：实时向微信用户返回响应内容，提升用户体验

## 技术架构

Ant Agent 采用模块化设计，主要由以下核心组件构成：

- **配置管理**：负责加载和管理配置信息
- **输入处理**：处理用户输入，支持命令补全
- **消息处理**：与大模型交互，处理工具调用
- **工具系统**：提供各种实用工具
- **技能系统**：管理技能的发现、安装和激活
- **ACP 服务器**：支持 ACP 协议，与 WeClaw 集成

这种模块化设计使得 Ant Agent 具备高度的可扩展性和可维护性，便于后续功能扩展和技术升级。

## 安装部署

### 环境要求

- **Go 版本**：1.23+ 或更高版本
- **依赖管理**：Go Modules
- **操作系统**：支持 Linux, macOS, Windows

### 安装步骤

1. **克隆代码库**
   ```bash
   git clone https://github.com/wangwentao/ant-agent.git
   cd ant-agent
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **构建可执行文件**
   ```bash
   go build -o ant-agent main.go
   ```

4. **运行**
   ```bash
   ./ant-agent
   ```

## 配置管理

### 配置文件

Ant Agent 使用 `config.json` 文件管理大模型访问信息，配置文件格式如下：

```json
{
  "api_key": "your-api-key",
  "model": "qwen3.5-27b-claude-4.6-opus-reasoning-distilled",
  "base_url": "http://localhost:1234",
  "max_tokens": 4096,
  "name": "Ant"
}
```

### 配置项说明

| 配置项 | 描述 | 默认值 |
|-------|------|-------|
| `api_key` | LLM 密钥 | 无（必填） |
| `model` | 使用的模型名称 | `qwen3.5-27b-claude-4.6-opus-reasoning-distilled` |
| `base_url` | API 基础 URL | `http://localhost:1234` |
| `max_tokens` | 模型最大输出 tokens | 4096 |
| `name` | Agent 的名称 | `Ant` |

### 配置优先级

1. **配置文件**：首先使用配置文件中的值
2. **环境变量**：如果配置文件不存在或配置项为空，则使用环境变量 `ANTHROPIC_API_KEY`
3. **错误提示**：如果环境变量也不存在，则程序会退出并提示错误

## 命令使用

### 基本指令

- `help`, `?` - 显示帮助信息
- `exit`, `q` - 退出 agent
- `install-skill <path> [--skill <name>] [--skills <names>]` - 安装技能
  - `<path>`: 本地目录路径或 GitHub 仓库地址
  - `--skill <name>`: 安装指定的单个技能
  - `--skills <names>`: 安装多个技能，空格分隔
  - `--skills all`: 安装仓库中的所有技能
- `remove-skill <name>` - 删除指定技能
- `show-skills` - 显示所有可用技能

### 命令示例

#### 安装技能

```bash
# 从本地目录安装技能
install-skill ./my-skill

# 从 GitHub 安装单个技能
install-skill vercel-labs/skills --skill find-skills

# 从 GitHub 安装多个技能
install-skill vercel-labs/skills --skills find-skills skill-creator

# 从 GitHub 安装所有技能
install-skill vercel-labs/skills --skills all
```

#### 删除技能

```bash
# 删除指定技能
remove-skill find-skills
```

#### 查看技能

```bash
# 显示所有可用技能
show-skills
```

## 技能生态

### 技能安装

Ant Agent 支持从本地目录或 GitHub 仓库安装技能，为用户提供丰富的功能扩展。技能安装后会被自动发现并集成到系统中，用户可以通过 `show-skills` 命令查看所有可用技能。

### 已安装技能

- **agent-browser**：浏览器自动化工具，用于与网站交互
- **skill-creator**：技能创建工具，用于创建和优化技能
- **xlsx**：电子表格处理工具，用于处理 Excel 文件
- **find-skills**：技能发现工具，帮助用户找到适合的技能
- **pdf**：PDF 处理工具，用于处理 PDF 文件
- **docx**：Word 文档处理工具，用于处理 Word 文件
- **pptx**：PowerPoint 处理工具，用于处理 PowerPoint 文件

## 用户界面

### 启动界面

Ant Agent 启动时会显示友好的界面，包括：

- 可用指令列表
- 已安装技能及其简要描述
- 大模型访问地址
- 退出提示

### 交互界面

在交互模式下，Ant Agent 提供清晰的输入输出界面：
- 用户输入提示符：`You »:`
- Agent 回复提示符：`<Agent 名称> »:`
- 工具执行结果：使用特殊格式显示

## 集成指南

### 与 WeClaw 集成

Ant Agent 可以与 [WeClaw](https://github.com/fastclaw-ai/weclaw) 项目集成，作为 WeChat 的 AI 代理，支持两种集成方式：

#### CLI 模式配置

在 WeClaw 的配置文件 `~/.weclaw/config.json` 中添加 Ant Agent 配置：

```json
{
  "default_agent": "ant",
  "agents": {
    "ant": {
      "type": "cli",
      "command": "/path/to/ant-agent",
      "model": "qwen3.5-27b-claude-4.6-opus-reasoning-distilled"
    }
  }
}
```

#### ACP 模式配置

ACP (Agent Client Protocol) 是一种标准化的代理-编辑器通信协议，提供更可靠的集成方式，推荐使用：

```json
{
  "default_agent": "ant-acp",
  "agents": {
    "ant-acp": {
      "type": "acp",
      "command": "/path/to/ant-agent",
      "model": "qwen3.5-27b-claude-4.6-opus-reasoning-distilled"
    }
  }
}
```

### 支持的参数

Ant Agent 支持 WeClaw 需要的以下参数：

- `-p <message>`: 接收消息内容（CLI 模式）
- `--output-format stream-json`: 输出 JSON 格式的事件（CLI 模式）
- `--resume <session_id>`: 恢复现有会话
- `--model <model_name>`: 指定使用的模型
- `--append-system-prompt <prompt>`: 添加系统提示
- `-acp`: 以 ACP 模式运行（ACP 模式）

### 微信 Clawbot 使用

1. **配置 WeClaw**：在 `~/.weclaw/config.json` 中配置 Ant Agent
2. **启动 WeClaw**：运行 `weclaw` 命令启动服务
3. **添加 Clawbot**：按照 WeClaw 文档添加微信机器人
4. **开始对话**：在微信中向 Clawbot 发送消息，如 "你好"
5. **接收响应**：Clawbot 会通过 Ant Agent 处理消息并返回响应

#### 示例对话

```
用户: 你好
Clawbot: 你好！我是 Ant，你的个人助手。有什么我可以帮助你的吗？

用户: 列出当前目录的文件
Clawbot: 好的，我将执行 `ls` 命令来列出当前目录的文件。

（Ant Agent 执行命令并返回结果）

Clawbot: 以下是当前目录的文件：
- README.md
- ant-agent
- config.json
- config.json.example
- go.mod
- go.sum
- internal
- skills-lock.json
- vendor
```

## 使用示例

### 基本交互

```bash
$ ./ant-agent

=== Ant 启动成功 ===

📋 可用指令:
  help, ?          - 显示帮助信息
  exit, q          - 退出 agent
  install-skill <path> - 从目录安装技能
  remove-skill <name> - 删除指定技能
  show-skills      - 显示所有可用技能

🧩 已安装技能:
  - agent-browser: Browser automation CLI for AI agents...
  - skill-creator: Create new skills, modify and improve existing skills...
  - xlsx: Use this skill any time a spreadsheet file is the primary input or output...
  - find-skills: Helps users discover and install agent skills...

🌐 大模型访问地址: http://localhost:1234

💡 提示:
  - 输入 'help' 查看详细帮助
  - 输入 'exit' 或 'q' 退出 agent

You »: 你好
Ant »: 你好！我是 Ant，你的个人助手。有什么我可以帮助你的吗？
You »: 列出当前目录的文件
Ant »: 好的，我将执行 `ls` 命令来列出当前目录的文件。
=== Tool: execute_shell ===
README.md
ant-agent
config.json
config.json.example
go.mod
go.sum
internal
skills-lock.json
vendor
=== End of Tool Result ===
Ant »: 以下是当前目录的文件：

- README.md
- ant-agent (可执行文件)
- config.json (配置文件)
- config.json.example (配置文件示例)
- go.mod (Go 模块文件)
- go.sum (Go 依赖校验文件)
- internal (内部代码目录)
- skills-lock.json (技能锁定文件)
- vendor (依赖库目录)
You »: exit
[2026-03-30 21:18:40] [INFO] Exiting...
```

### 技能安装

```bash
$ ./ant-agent

=== Ant 启动成功 ===

You »: install-skill vercel-labs/skills --skill find-skills
[INFO] Cloning GitHub repository: vercel-labs/skills
[INFO] Installing skill: find-skills
[INFO] Skill find-skills installed successfully

You »: show-skills
=== 可用技能 ===

1. agent-browser
   描述: Browser automation CLI for AI agents...

2. skill-creator
   描述: Create new skills, modify and improve existing skills...

3. find-skills
   描述: Helps users discover and install agent skills...

You »: exit
```

## 最佳实践

### 配置管理

- **安全存储 API 密钥**：避免在代码或配置文件中硬编码 API 密钥，优先使用环境变量
- **合理设置 max_tokens**：根据实际需求调整最大输出 tokens，平衡响应质量和速度
- **选择合适的模型**：根据任务类型和性能需求选择合适的模型

### 技能管理

- **按需安装技能**：只安装必要的技能，避免过多技能影响性能
- **定期更新技能**：保持技能的最新版本，获取最新功能和 bug 修复
- **创建自定义技能**：根据特定需求创建自定义技能，扩展 Ant Agent 能力

### 性能优化

- **使用流式响应**：对于长对话，启用流式响应以提升用户体验
- **合理使用工具**：避免频繁调用工具，减少 API 调用次数
- **优化系统提示**：根据具体任务优化系统提示，提高模型响应质量

## 注意事项

- 确保你的 API 密钥是有效的
- 配置文件中的 `base_url` 不需要包含 `/v1` 路径，SDK 会自动添加
- 如果配置文件格式错误，程序会显示警告并使用默认配置
- 命令补全功能支持基本命令和技能名称的补全
- 工具执行有安全限制，某些危险命令会被阻止
- 技能描述在启动时会被截断显示，使用 `show-skills` 命令查看完整描述
- 技能存储目录为 `.ant/skills`，位于可执行文件目录或用户主目录

## 项目规划

Ant Agent 团队致力于持续提升产品能力，为用户提供更智能、更强大的 AI 代理服务。以下是我们的后续规划：

### 近期规划

1. **MCP 协议支持**：集成 MCP (Model Context Protocol) 协议，扩展与外部服务的交互能力，支持更多类型的工具和服务集成。

2. **记忆系统**：实现会话记忆功能，使 Ant Agent 能够记住之前的对话内容，提供更连贯的交互体验。

3. **知识库检索 (RAG)**：集成知识库检索能力，通过检索增强生成 (RAG) 技术，提升 Ant Agent 的知识储备和回答准确性。

### 中长期规划

4. **多模态内容交互**：支持处理和生成图像、音频等多模态内容，扩展交互维度，提供更丰富的用户体验。

5. **多 Agent 通讯与协作**：实现多个 Ant Agent 之间的通讯和协作机制，解决复杂任务分解和多步骤处理问题。

6. **安全管控**：加强安全措施，包括输入输出过滤、权限控制、安全审计等，确保系统安全稳定运行。

我们欢迎社区贡献和反馈，共同推动 Ant Agent 的发展和完善。

## 社区与支持

### 贡献指南

我们欢迎社区贡献，包括但不限于：
- 提交 bug 报告和功能请求
- 贡献代码和技能
- 改进文档和示例
- 分享使用经验和最佳实践

### 联系方式

- **GitHub 仓库**：[https://github.com/wangwentao/ant-agent](https://github.com/wangwentao/ant-agent)
- **问题反馈**：在 GitHub 仓库提交 Issue
- **讨论交流**：通过 GitHub Discussions 进行交流

## 许可证

Ant Agent 采用 MIT 许可证，详见 LICENSE 文件。

