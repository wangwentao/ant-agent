# Ant Agent

一个基于 Anthropic Go SDK 开发的智能代理工具，能够持续接收用户输入，与大模型进行交互，并处理工具调用循环，直到获得最终结果。Ant Agent 具有强大的技能系统和工具系统，可以帮助用户完成各种任务。

## 功能特性

- 持续等待用户输入，支持命令补全
- 与 Anthropic 模型进行交互
- 处理工具调用循环，自动执行工具并返回结果
- 输出最终结果到控制台
- 使用配置文件管理大模型访问信息
- 技能系统：发现、安装、删除和管理技能
- 工具系统：执行shell命令、文件操作等
- 友好的启动界面，显示可用指令和已安装技能
- 支持自定义 API 基础 URL

## 安装

1. 确保安装了 Go 1.23+ 版本
2. 克隆代码库
3. 运行 `go build -o ant-agent main.go` 构建可执行文件

## 配置

工具支持使用配置文件 `config.json` 来管理大模型访问信息。配置文件格式如下：

```json
{
  "api_key": "your-anthropic-api-key",
  "model": "claude-3-opus-20240229",
  "base_url": "https://api.anthropic.com",
  "max_tokens": 1024
}
```

### 配置项说明

- `api_key`: Anthropic API 密钥
- `model`: 使用的模型名称，默认为 `claude-3-opus-20240229`
- `base_url`: API 基础 URL，默认为 `https://api.anthropic.com`
- `max_tokens`: 模型最大输出 tokens，默认为 1024

### 配置优先级

1. 首先使用配置文件中的值
2. 如果配置文件不存在或配置项为空，则使用环境变量 `ANTHROPIC_API_KEY`
3. 如果环境变量也不存在，则程序会退出并提示错误

## 使用方法

1. 复制 `config.json.example` 为 `config.json`
2. 编辑 `config.json` 文件，填入你的 API 密钥
3. 运行 `./ant-agent` 启动工具
4. 输入你的问题或请求，按回车发送
5. 工具会自动处理与模型的交互，包括工具调用
6. 最终结果会显示在控制台
7. 输入 `exit` 或 `q` 退出工具

## 可用指令

- `help`, `?` - 显示帮助信息
- `exit`, `q` - 退出 agent
- `install-skill <path>` - 从目录安装技能
- `remove-skill <name>` - 删除指定技能
- `show-skills` - 显示所有可用技能

## 工具系统

Ant Agent 支持以下工具调用：

- `execute_shell`: 执行shell命令并返回输出
- `read_file`: 读取文件内容
- `write_file`: 写入内容到文件
- `edit_file`: 通过替换字符串编辑文件
- `list_skills`: 列出所有可用技能
- `activate_skill`: 激活指定技能

## 技能系统

Ant Agent 具有强大的技能系统，支持以下功能：

- **技能发现**: 自动发现项目内和用户主目录下的技能
- **技能安装**: 通过 `install-skill` 命令从目录安装技能
- **技能删除**: 通过 `remove-skill` 命令删除指定技能
- **技能激活**: 通过 `activate_skill` 工具激活指定技能以获取详细说明

### 已安装技能

- **agent-browser**: 浏览器自动化工具，用于与网站交互
- **skill-creator**: 技能创建工具，用于创建和优化技能
- **xlsx**: 电子表格处理工具，用于处理 Excel 文件
- **find-skills**: 技能发现工具，帮助用户找到适合的技能

## 启动界面

Ant Agent 启动时会显示友好的界面，包括：

- 可用指令列表
- 已安装技能及其简要描述
- 大模型访问地址
- 退出提示

## 示例

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

## 注意事项

- 确保你的 API 密钥是有效的
- 配置文件中的 `base_url` 不需要包含 `/v1` 路径，SDK 会自动添加
- 如果配置文件格式错误，程序会显示警告并使用默认配置
- 命令补全功能支持基本命令和技能名称的补全
- 工具执行有安全限制，某些危险命令会被阻止
- 技能描述在启动时会被截断显示，使用 `show-skills` 命令查看完整描述

## 与 WeClaw 集成

Ant Agent 可以与 [WeClaw](https://github.com/fastclaw-ai/weclaw) 项目集成，作为 WeChat 的 AI 代理。

### WeClaw 配置

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

### 支持的参数

Ant Agent 支持 WeClaw 需要的以下参数：

- `-p <message>`: 接收消息内容
- `--output-format stream-json`: 输出 JSON 格式的事件
- `--resume <session_id>`: 恢复现有会话
- `--model <model_name>`: 指定使用的模型
- `--append-system-prompt <prompt>`: 添加系统提示

### 集成原理

WeClaw 通过以下方式与 Ant Agent 交互：

1. WeClaw 发送消息到 Ant Agent：`ant-agent -p "Hello" --output-format stream-json`
2. Ant Agent 处理消息并返回 JSON 格式的事件：
   - 会话事件：包含会话 ID
   - 结果事件：包含处理结果
3. WeClaw 解析 JSON 事件并将结果发送到 WeChat

### 测试集成

```bash
# 测试 Ant Agent 是否能正确处理 WeClaw 格式的请求
./ant-agent -p "Hello, world!" --output-format stream-json
```

预期输出：

```json
{"type":"session","session_id":"session_123456789","result":"","is_error":false}
{"type":"result","session_id":"session_123456789","result":"Hello! I'm Ant Agent...","is_error":false}
```
