# Ant Agent

一个基于 Anthropic Go SDK 开发的智能代理工具，能够持续接收用户输入，与大模型进行交互，并处理工具调用循环，直到获得最终结果。Ant Agent 具有强大的技能系统和工具系统，可以帮助用户完成各种任务，并支持与 WeClaw 项目集成，为微信 Clawbot 提供智能服务。

## 功能特性

- **持续交互**: 持续等待用户输入，支持命令补全
- **大模型集成**: 与 Anthropic 模型进行交互，支持流式响应
- **工具调用**: 处理工具调用循环，自动执行工具并返回结果
- **技能系统**: 发现、安装、删除和管理技能
- **工具系统**: 执行shell命令、文件操作等
- **友好界面**: 启动时显示可用指令和已安装技能
- **配置管理**: 使用配置文件管理大模型访问信息
- **自定义能力**: 支持自定义 API 基础 URL
- **WeClaw 集成**: 支持 CLI 模式和 ACP 模式与 WeClaw 集成

## 安装

### 编译环境

- **Go 版本**: 1.23+ 或更高版本
- **依赖管理**: Go Modules
- **操作系统**: 支持 Linux, macOS, Windows

### 构建步骤

1. 克隆代码库
   ```bash
   git clone https://github.com/wangwentao/ant-agent.git
   cd ant-agent
   ```

2. 安装依赖
   ```bash
   go mod download
   ```

3. 构建可执行文件
   ```bash
   go build -o ant-agent main.go
   ```

4. 运行
   ```bash
   ./ant-agent
   ```

## 配置

工具支持使用配置文件 `config.json` 来管理大模型访问信息。配置文件格式如下：

```json
{
  "api_key": "your-api-key",
  "model": "qwen3.5-27b-claude-4.6-opus-reasoning-distilled",
  "base_url": "http://localhost:1234",
  "max_tokens": 1024
}
```

### 配置项说明

- `api_key`: LLM 密钥
- `model`: 使用的模型名称，默认为 `qwen3.5-27b-claude-4.6-opus-reasoning-distilled`
- `base_url`: API 基础 URL，默认为 `http://localhost:1234`
- `max_tokens`: 模型最大输出 tokens，默认为 1024

### 配置优先级

1. 首先使用配置文件中的值
2. 如果配置文件不存在或配置项为空，则使用环境变量 `ANTHROPIC_API_KEY`
3. 如果环境变量也不存在，则程序会退出并提示错误

## 命令使用

### 基本指令

- `help`, `?` - 显示帮助信息
- `exit`, `q` - 退出 agent
- `install-skill <path> [--skill <name>] [--skills <names>]` - 安装技能
  - `<path>`: 本地目录路径或GitHub仓库地址
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

## 工具系统

Ant Agent 支持以下工具调用：

- `execute_shell`: 执行shell命令并返回输出
- `read_file`: 读取文件内容
- `write_file`: 写入内容到文件
- `edit_file`: 通过替换字符串编辑文件
- `list_skills`: 列出所有可用技能
- `activate_skill`: 激活指定技能以获取详细说明

## 技能扩展能力

Ant Agent 具有强大的技能系统，支持以下功能：

- **技能发现**: 自动发现项目内（`.ant/skills`）和用户主目录下（`~/.agent/skills`）的技能
- **技能安装**: 通过 `install-skill` 命令从本地目录或 GitHub 仓库安装技能
- **技能删除**: 通过 `remove-skill` 命令删除指定技能
- **技能激活**: 通过 `activate_skill` 工具激活指定技能以获取详细说明

### 技能安装示例

#### 从本地目录安装技能
```bash
# 从本地目录安装技能
install-skill ./my-skill
```

#### 从 GitHub 安装技能
```bash
# 安装单个技能
install-skill vercel-labs/skills --skill find-skills

# 安装多个技能
install-skill vercel-labs/skills --skills find-skills skill-creator

# 安装所有技能
install-skill vercel-labs/skills --skills all
```

### 技能卸载示例

```bash
# 卸载指定技能
remove-skill find-skills
```

### 已安装技能

- **agent-browser**: 浏览器自动化工具，用于与网站交互
- **skill-creator**: 技能创建工具，用于创建和优化技能
- **xlsx**: 电子表格处理工具，用于处理 Excel 文件
- **find-skills**: 技能发现工具，帮助用户找到适合的技能
- **pdf**: PDF 处理工具，用于处理 PDF 文件
- **docx**: Word 文档处理工具，用于处理 Word 文件
- **pptx**: PowerPoint 处理工具，用于处理 PowerPoint 文件
- **golang-patterns**: Go 语言模式工具，提供 Go 最佳实践
- **golang-pro**: Go 高级编程工具，提供并发编程等高级特性
- **golang-testing**: Go 测试工具，提供测试最佳实践

## 启动界面

Ant Agent 启动时会显示友好的界面，包括：

- 可用指令列表
- 已安装技能及其简要描述
- 大模型访问地址
- 退出提示

## 示例

### 基本交互示例

```bash
$ ./ant-agent

=== Ant Agent 启动成功 ===

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

🌐 大模型访问地址: https://api.anthropic.com

💡 提示:
  - 输入 'help' 查看详细帮助
  - 输入 'exit' 或 'q' 退出 agent

You »: 你好
Ant agent: 你好！我是 Claude，由 Anthropic 开发的 AI 助手。有什么我可以帮助你的吗？
You »: 列出当前目录的文件
Ant agent: 好的，我将执行 `ls` 命令来列出当前目录的文件。
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
Ant agent: 以下是当前目录的文件：

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
[2026-03-25 00:00:00] [INFO] Exiting...
```

### 技能安装示例

```bash
$ ./ant-agent

=== Ant Agent 启动成功 ===

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

## 注意事项

- 确保你的 API 密钥是有效的
- 配置文件中的 `base_url` 不需要包含 `/v1` 路径，SDK 会自动添加
- 如果配置文件格式错误，程序会显示警告并使用默认配置
- 命令补全功能支持基本命令和技能名称的补全
- 工具执行有安全限制，某些危险命令会被阻止
- 技能描述在启动时会被截断显示，使用 `show-skills` 命令查看完整描述
- 技能存储目录为 `.ant/skills`，位于可执行文件目录或用户主目录

## 与 WeClaw 集成

Ant Agent 可以与 [WeClaw](https://github.com/fastclaw-ai/weclaw) 项目集成，作为 WeChat 的 AI 代理，支持两种集成方式：CLI 模式和 ACP 模式。

### CLI 模式配置

在 WeClaw 的配置文件 `~/.weclaw/config.json` 中添加 Ant Agent 配置：

```json
{
  "default_agent": "ant",
  "agents": {
    "ant": {
      "type": "cli",
      "command": "/path/to/ant-agent",
      "model": "claude-3-opus-20240229"
    }
  }
}
```

### ACP 模式配置

ACP (Agent Client Protocol) 是一种标准化的代理-编辑器通信协议，提供更可靠的集成方式，推荐使用：

```json
{
  "default_agent": "ant-acp",
  "agents": {
    "ant-acp": {
      "type": "acp",
      "command": "/path/to/ant-agent",
      "model": "claude-3-opus-20240229"
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

### 集成原理

#### CLI 模式

1. WeClaw 发送消息到 Ant Agent：`ant-agent -p "Hello" --output-format stream-json`
2. Ant Agent 处理消息并返回 JSON 格式的事件：
   - 会话事件：包含会话 ID
   - 结果事件：包含处理结果
3. WeClaw 解析 JSON 事件并将结果发送到 WeChat

#### ACP 模式

1. WeClaw 启动 Ant Agent 并建立 ACP 连接：`ant-agent -acp`
2. WeClaw 通过 ACP 协议发送初始化请求
3. Ant Agent 响应初始化请求，建立会话
4. WeClaw 发送消息到 Ant Agent
5. Ant Agent 处理消息并通过 ACP 协议返回结果
6. WeClaw 解析 ACP 消息并将结果发送到 WeChat

### 微信 Clawbot 支持

通过与 WeClaw 集成，Ant Agent 可以为微信 Clawbot 提供以下能力：

- **智能对话**: 与微信用户进行自然语言对话
- **工具调用**: 执行各种工具操作，如文件处理、命令执行等
- **技能扩展**: 利用已安装的技能提供更专业的服务
- **流式响应**: 实时向微信用户返回响应内容

### 测试集成

#### 测试 CLI 模式

```bash
# 测试 Ant Agent 是否能正确处理 WeClaw 格式的请求
./ant-agent -p "Hello, world!" --output-format stream-json
```

预期输出：

```json
{"type":"session","session_id":"session_123456789","result":"","is_error":false}
{"type":"result","session_id":"session_123456789","result":"Hello! I'm Ant Agent...","is_error":false}
```

#### 测试 ACP 模式

```bash
# 启动 Ant Agent 以 ACP 模式运行
./ant-agent -acp
```

Ant Agent 会等待 ACP 客户端连接并处理请求。

### 微信 Clawbot 使用示例

1. **配置 WeClaw**：在 `~/.weclaw/config.json` 中配置 Ant Agent
2. **启动 WeClaw**：运行 `weclaw` 命令启动服务
3. **添加 Clawbot**：按照 WeClaw 文档添加微信机器人
4. **开始对话**：在微信中向 Clawbot 发送消息，如 "你好"
5. **接收响应**：Clawbot 会通过 Ant Agent 处理消息并返回响应

#### 示例对话

```
用户: 你好
Clawbot: 你好！我是 Ant Agent，由 Anthropic 提供支持的智能助手。有什么我可以帮助你的吗？

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
