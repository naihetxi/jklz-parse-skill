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

### 方式 1：使用 Python CLI 工具（推荐）

Python CLI 无需编译，开箱即用：

```bash
# 克隆仓库
git clone https://github.com/naihetxi/jklz-parse-skill.git
cd jklz-parse-skill

# 配置 API Key（联系管理员获取）
python3 cli/jklz-parse.py config --api-key YOUR_API_KEY --base-url http://YOUR_API_HOST:PORT

# 健康检查
python3 cli/jklz-parse.py health

# 解析文档
python3 cli/jklz-parse.py parse document.pdf

# 更多示例
python3 cli/jklz-parse.py parse document.pdf --return html -o output.html
python3 cli/jklz-parse.py parse document.pdf --page-range "1-5"
python3 cli/jklz-parse.py parse document.pdf --return content#toc#table
```

**注意**：Python CLI 需要 `requests` 库：
```bash
pip3 install requests
```

### 方式 2：使用 Go CLI（需要 Go 1.21+）

**已编译版本**：项目中已包含编译好的 Go CLI 二进制文件（`cli/jklz-parse`），可直接使用。

```bash
cd cli

# 直接使用预编译版本
./jklz-parse config --api-key YOUR_API_KEY --base-url http://YOUR_API_HOST:PORT

# 使用（功能与 Python CLI 完全相同）
./jklz-parse parse document.pdf
./jklz-parse parse document.pdf --return content#toc --image-mode cv -o output.json
```

**从源码编译**（可选）：

```bash
cd cli
go build -o jklz-parse main.go

# 配置和使用
./jklz-parse config --api-key YOUR_API_KEY
./jklz-parse parse document.pdf
```

**注意**：Go CLI 需要 Go 1.21+ 版本才能编译。如遇编译问题，建议直接使用 Python CLI 或预编译版本。

### 方式 2：安装为 AI Agent 技能

将本技能安装到支持的 AI Agent runtime，Agent 会自动识别文档解析需求并调用：

| Runtime | 安装路径 |
|---------|---------|
| Claude Code | `~/.claude/skills/jklz-parse-skill/` |
| Codex | `~/.codex/skills/jklz-parse-skill/` |
| Cursor | `~/.cursor/skills/jklz-parse-skill/` |
| OpenClaw | `~/.openclaw/skills/jklz-parse-skill/` |

**安装步骤：**

```bash
# 1. 克隆仓库
git clone https://github.com/naihetxi/jklz-parse-skill.git

# 2. 安装到 Claude Code（以 Claude Code 为例）
cp -r jklz-parse-skill ~/.claude/skills/

# 3. 配置 API Key（在技能目录下）
cd ~/.claude/skills/jklz-parse-skill
python3 cli/jklz-parse.py config --api-key YOUR_API_KEY
```

**Agent 自动触发条件：**

当你对 AI Agent 说以下内容时，会自动触发此技能：
- "解析这个 PDF 文件"
- "把 Word 文档转成 Markdown"
- "提取 Excel 中的所有表格"
- "分析这份合同的内容"
- "把这个文档切片用于知识库"

**Agent 使用示例：**

```
用户: 帮我解析 report.pdf 并提取其中的表格
Agent: 好的，我来帮你解析文档...
[自动调用 jklz-parse-skill]
[返回文本内容和表格数据]
```

**配置要求：**

Agent 运行时需要确保：
1. API Key 已配置（通过 CLI 工具配置）
2. Python 3 环境可用
3. 已安装 `requests` 库：`pip3 install requests`

## 功能特性

- 📄 **多格式支持**：PDF、DOC、DOCX、XLSX、PPT 等
- 📝 **内容提取**：文本、表格、目录、图片
- 🔪 **智能切片**：按目录或长度切分，适用于 RAG
- 📊 **表格处理**：自动识别、跨页合并、格式转换
- 🔍 **内容溯源**：追踪内容在原文中的位置
- 🚀 **CLI 工具**：Python（零依赖）+ Go（高性能）双实现

## Darwin 评分：86.6/100

经过专业 Skill 优化流程，7轮优化后达到：

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

## 使用示例

### CLI 工具

```bash
# 提取 PDF 文本（Markdown 格式）
python3 cli/jklz-parse.py parse report.pdf

# 提取为 HTML
python3 cli/jklz-parse.py parse report.pdf --return html

# 提取 Excel 表格
python3 cli/jklz-parse.py parse data.xlsx --return table

# 大文件分片用于 RAG
python3 cli/jklz-parse.py parse large.pdf --return slice -o chunks.json

# 选择页面范围
python3 cli/jklz-parse.py parse doc.pdf --page-range "1-5,10"

# 组合多种格式
python3 cli/jklz-parse.py parse doc.pdf --return content#toc#table

# 高性能模式（更快，适合大文件）
python3 cli/jklz-parse.py parse large.pdf --image-mode cv
```

### AI Agent 使用

安装为技能后，Agent 会自动识别并处理文档解析需求：

```
"解析这个 PDF 文件"
"把 Word 文档转成 Markdown"
"提取 Excel 中的所有表格"
"分析这份合同的内容"
"把这个文档切片用于知识库"
```

Agent 会自动：
1. 检测文档类型和格式
2. 选择合适的解析模式
3. 提取所需的内容（文本/表格/目录）
4. 返回格式化的结果

## 技术栈

- **Skill 定义**：Markdown + YAML frontmatter
- **CLI 工具**：Python 3（推荐）+ Go 1.21+
- **API**：金科览智 Parse API（HTTP/SSE 流式响应）
- **配置存储**：`~/.config/jklz-parse/config.json`

## API 配置说明

### 获取 API Key

API Key 需要向金科览智服务管理员申请。申请时需提供：
- 使用场景说明
- 预计调用量

### 配置文件位置

配置保存在：`~/.config/jklz-parse/config.json`

```json
{
  "api_key": "your-api-key-here",
  "base_url": "http://192.168.42.15:15216"
}
```

也可以使用环境变量：
```bash
export JKLZ_PARSE_APIKEY="your-api-key"
export JKLZ_PARSE_BASEURL="http://your-api-host:port"
```

## 支持的输出格式

| 格式 | 说明 | 使用场景 |
|------|------|---------|
| `content` | Markdown 文本（默认） | 阅读、编辑、知识库 |
| `html` | HTML 格式 | 网页展示 |
| `toc` | 目录结构（JSON） | 文档导航 |
| `table` | 表格数据（JSON） | 数据分析 |
| `slice` | 文档切片（JSON） | RAG 知识库、向量搜索 |

可以用 `#` 组合多个格式，例如：`content#toc#table`

## 性能选项

- **vl 模式**（高精度）：适合表格密集、复杂版式的文档
- **cv 模式**（高性能）：速度更快，适合大文件和简单文档

```bash
# 高精度模式（默认）
python3 cli/jklz-parse.py parse doc.pdf --image-mode vl

# 高性能模式
python3 cli/jklz-parse.py parse doc.pdf --image-mode cv
```

## 开发

### 构建 Go CLI 工具

```bash
cd cli

# 需要 Go 1.21+
go build -o jklz-parse main.go

# 或使用构建脚本
./build.sh          # 构建当前平台
./build.sh all      # 交叉编译全平台
```

### Python CLI 开发

Python CLI 无需构建，直接运行：

```bash
python3 cli/jklz-parse.py --help
```

### Skill 定义

Skill 定义文件：`SKILL.md`，包含：
- Frontmatter（触发词、能力描述）
- Phase 1-4 工作流
- 失败处理和 fallback 策略
- API 调用示例

修改 Skill 定义后，重新安装到 Agent runtime 即可生效。

## 故障排除

### 1. 网络连接失败

```bash
# 检查服务状态
python3 cli/jklz-parse.py health

# 测试 API 可达性
curl http://YOUR_API_HOST:PORT/metrics
```

### 2. API Key 错误

```bash
# 重新配置
python3 cli/jklz-parse.py config --api-key YOUR_KEY

# 查看当前配置
python3 cli/jklz-parse.py config --show
```

### 3. 解析超时或卡住

- 尝试切换到高性能模式：`--image-mode cv`
- 减小处理范围：`--page-range "1-10"`
- 检查文件大小（建议 < 200MB）

### 4. Go CLI 编译失败

确保 Go 版本 >= 1.21：
```bash
go version
# 如果版本太低，升级 Go：brew install go
```

如果 Go 编译有问题，直接使用 Python CLI（功能完全相同）。

## 详细文档

- [快速使用指南](cli/快速使用指南.md) - 完整的使用示例和最佳实践
- [CLI README](cli/README.md) - CLI 工具详细说明
- [API 参考](references/api.md) - API 接口文档

## License

MIT

## 贡献

欢迎提交 Issue 和 Pull Request！

如需技术支持或申请 API Key，请联系金科览智服务管理员。
