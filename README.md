# jklz-parse-skill

金科览智文档解析技能 - 为 AI Agent 提供文档解析能力

[![Darwin Score](https://img.shields.io/badge/Darwin%20Score-86.6%2F100-brightgreen)]()
[![Skill-Creator](https://img.shields.io/badge/Skill--Creator-Standard-blue)]()

## 项目结构

```
jklz-parse-skill/
├── SKILL.md              # AI Agent 技能定义（328行）
├── agents/
│   └── openai.yaml       # UI 友好元数据
├── references/
│   └── api.md            # API 详细参考
├── scripts/
│   ├── call_api.sh       # API 调用脚本
│   └── parse-response.cjs # SSE 响应解析
└── cli/                  # 命令行工具
    ├── jklz-parse        # Python CLI（零依赖）
    ├── main.go           # Go CLI（高性能）
    └── README.md         # CLI 文档
```

## 快速开始

### 方式 1：使用 CLI 工具（推荐）

```bash
# 安装 CLI
curl -fsSL https://raw.githubusercontent.com/naihetxi/jklz-parse-skill/main/cli/install.sh | bash

# 配置 API Key
jklz-parse config --api-key YOUR_API_KEY

# 解析文档
jklz-parse parse document.pdf
```

### 方式 2：安装为 AI Agent 技能

将本技能安装到支持的 AI Agent runtime：

| Runtime | 安装路径 |
|---------|---------|
| Claude Code | `~/.claude/skills/jklz-parse-skill/` |
| Codex | `~/.codex/skills/jklz-parse-skill/` |
| Cursor | `~/.cursor/skills/jklz-parse-skill/` |
| OpenClaw | `~/.openclaw/skills/jklz-parse-skill/` |

```bash
# 示例：安装到 Claude Code
git clone https://github.com/naihetxi/jklz-parse-skill.git
cp -r jklz-parse-skill ~/.claude/skills/
```

## 功能特性

- 📄 **多格式支持**：PDF、DOC、DOCX、XLSX、PPT 等
- 📝 **内容提取**：文本、表格、目录、图片
- 🔪 **智能切片**：按目录或长度切分，适用于 RAG
- 📊 **表格处理**：自动识别、跨页合并、格式转换
- 🔍 **内容溯源**：追踪内容在原文中的位置
- 🚀 **CLI 工具**：Python（零依赖）+ Go（高性能）双实现

## Darwin 评分：86.6/100

经过 Darwin Skill 优化流程，7轮优化后达到：

| 维度 | 分数 | 说明 |
|------|------|------|
| Frontmatter 质量 | 9/10 | 简洁明确，触发词完整 |
| 工作流清晰度 | 9/10 | Phase 1-4 结构化流程 |
| 失败模式编码 | 9/10 | if-then-else 三段式 fallback |
| 检查点设计 | 8/10 | 🔴 CHECKPOINT 显性标记 |
| 可执行具体性 | 9/10 | 函数、示例、参数完整 |
| 资源整合度 | 10/10 | agents/ + references/ + scripts/ + cli/ |
| 整体架构 | 9/10 | 328行，渐进式披露 |
| 实测表现 | 8/10 | 工作流完整，失败处理健壮 |
| 反例与黑名单 | 9/10 | 5个禁止操作 + 4个不适用场景 |

详见：[优化报告](docs/optimization-report.md)

## 使用示例

### CLI 工具

```bash
# 提取 PDF 文本
jklz-parse parse report.pdf

# 提取 Excel 表格为 Markdown
jklz-parse parse data.xlsx --return table --table-format markdown

# 大文件分片用于 RAG
jklz-parse parse large.pdf --return slice --output chunks.json

# 选择页面范围
jklz-parse parse doc.pdf --page-range "1-5,10"
```

### AI Agent 使用

安装后，Agent 会自动触发，你可以这样说：

```
"解析这个 PDF 文件"
"把 Word 文档转成 Markdown"
"提取 Excel 中的所有表格"
"分析这份合同的内容"
"把这个文档切片用于知识库"
```

## 技术栈

- **Skill 定义**：Markdown + YAML frontmatter
- **CLI 工具**：Python 3（零依赖）+ Go 1.21+
- **API**：金科览智 Parse API（HTTP/SSE）

## 开发

### 构建 CLI 工具

```bash
cd cli

# Python 版本（直接可用）
./jklz-parse --help

# Go 版本（需要 Go 1.21+）
./build.sh          # 构建当前平台
./build.sh all      # 交叉编译全平台
```

### 优化 Skill

使用 Darwin Skill 进行质量评估和优化：

```bash
darwin-skill optimize jklz-parse-skill
```

## License

MIT

## 贡献

欢迎提交 Issue 和 Pull Request！
