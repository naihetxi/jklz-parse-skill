# 🧠 jklz-parse-skill — 让 AI Agent 学会解析各种长文档

[![Darwin Score](https://img.shields.io/badge/Darwin%20Score-86.6%2F100-brightgreen)]()
[![Skill-Creator](https://img.shields.io/badge/Skill--Creator-Standard-blue)]()

> 安装这个 Skill，你的 AI Agent（如 Cursor, Claude Code, OpenClaw 等）就能理解如何使用 `jklz-parse` 工具解析 PDF、Word、Excel、PPT 等文件。对着 AI 说一句「帮我解析这份合同并提取表格」，剩下的交给它。

## 这是什么？

这是一份 AI Agent 的技能描述文件（`SKILL.md`），教会 AI Agent 如何调用项目内置的 `jklz-parse` 命令行工具或直接通过 HTTP/curl 调用远程 API 来完成复杂文档的智能解析：

```
你说：「帮我解析 report.pdf 并提取其中的表格」

AI 自动执行：检测文件格式 → 选择高精度解析模式 → 提取 Markdown 文本和 JSON 表格 → 整理数据并返回
```

### CLI 和 Skill 的关系

| | CLI（命令行工具） | Skill（技能描述文件） |
|---|---|---|
| **是什么** | 位于 `cli/` 下的 Python / Go 命令行解析程序 | 一份教 AI 怎么用命令或 API 的说明书（`SKILL.md`） |
| **单独能用吗** | **可以**。独立运行，无需 AI Agent | **可以**。Agent 可通过 HTTP 直接调用 API，无需 CLI |
| **最佳组合** | AI Agent + CLI 配合使用，享有各类健壮特性（重试、导出、流解析） | |

简单说：**CLI 是工具，Skill 是说明书，API 是引擎。** 三者各司其职，也可自由组合。

---

## 快速安装

### 第 1 步：一键安装 CLI

**macOS / Linux**：
```bash
curl -fsSL https://raw.githubusercontent.com/naihetxi/jklz-parse-skill/main/install.sh | bash
```

**Windows PowerShell**：
```powershell
powershell -ExecutionPolicy Bypass -Command "iwr https://raw.githubusercontent.com/naihetxi/jklz-parse-skill/main/install.ps1 -UseBasicParsing | iex"
```

安装完成后配置 API Key：

```bash
jklz-parse config --api-key YOUR_API_KEY --base-url http://192.168.42.15:15216
jklz-parse health
```
> 📧 没有 API Key？请联系金科览智服务管理员获取。

安装脚本默认从 GitHub Releases 下载二进制文件；如果当前版本还没有发布 Release，会回退到仓库内 `cli/build/` 的同名二进制。

也可前往 [Releases 页面](https://github.com/naihetxi/jklz-parse-skill/releases) 手动下载对应平台的二进制文件，或直接用 Python CLI：

```bash
python3 cli/jklz-parse.py config --api-key YOUR_API_KEY --base-url http://192.168.42.15:15216
```

内网分发时，把二进制文件放到自己的静态文件服务并指定下载地址：

```bash
curl -fsSL https://raw.githubusercontent.com/naihetxi/jklz-parse-skill/main/install.sh | \
  JKLZ_INSTALL_BASE_URL="https://your-internal-host/jklz-parse/cli/build" bash
```

### 第 2 步：安装 Skill 到你的 AI Agent

将整个仓库 clone 到 Agent 的技能目录即可：

**Claude Code：**
```bash
git clone https://github.com/naihetxi/jklz-parse-skill.git \
  /path/to/project/.skills/jklz-parse-skill
```

**Cursor：**
```bash
git clone https://github.com/naihetxi/jklz-parse-skill.git \
  /path/to/project/.cursor/rules/jklz-parse-skill
```

**Codex：**
```bash
git clone https://github.com/naihetxi/jklz-parse-skill.git ~/.codex/skills/jklz-parse-skill
```

### 第 3 步：开始对话！

直接用自然语言和你的 AI 交流：
- 「帮我解析这个 PDF 文件」
- 「把这个 Word 文档转成 Markdown」
- 「提取这份 Excel 中的所有表格数据」
- 「把这篇超长文档切片，我需要用来做 RAG 知识库」
- 「查询我的解析历史记录」

---

## 已测试平台

| 平台 | 安装方式 | 状态 |
|------|---------|------|
| **Claude Code** | `git clone` 到项目 `.skills` 目录 | ✅ 已验证 |
| **Cursor** | `git clone` 到 `.cursor/rules` 目录 | ✅ 已验证 |
| **Windsurf** | `git clone` 到 `.skills` 目录 | ✅ 已验证 |
| **WorkBuddy (腾讯)** | 上传 `SKILL.md` + 其他所需文件 | ✅ 已验证 |
| **QClaw (腾讯)** | 上传 `SKILL.md` + 其他所需文件 | ✅ 已验证 |
| **有道龙虾 (Youdao)** | `git clone` 到技能目录 | ✅ 已验证 |
| **元气 AI** | `git clone` 到技能目录 | ✅ 已验证 |
| **Codex** | `git clone` 到 `~/.codex/skills` | ✅ 已验证 |
| **小龙虾 OpenClaw** | `git clone` 到 `~/.openclaw/skills` | ✅ 已验证 |
| 其他支持 Markdown Skill 的 Agent | `git clone` 后指向 SKILL.md | ✅ 兼容 |

---

## CLI 完整命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `parse` | 解析文档 | `jklz-parse parse doc.pdf --return content#toc --output result.md` |
| `get` | 获取解析结果 | `jklz-parse get userId jobId fileId -r content,html` |
| `export` | 导出解析结果文件 | `jklz-parse export userId jobId fileId --type md -o result.md` |
| `history` | 查询历史记录 | `jklz-parse history userId` |
| `search` | 关键词搜索 | `jklz-parse search userId jobId fileId -k "合同,条款"` |
| `cancel` | 取消任务 | `jklz-parse cancel userId jobId` |
| `modify` | 修改 chunk | `jklz-parse modify userId jobId fileId -c chunkId -t "新内容"` |
| `cleanup` | 清理历史 | `jklz-parse cleanup userId 7d` |
| `config` | 配置管理 | `jklz-parse config --api-key KEY --base-url URL` |
| `health` | 健康检查 | `jklz-parse health` |

### 常用示例

```bash
# 解析 PDF 并导出格式化的 Markdown
jklz-parse parse document.pdf -o result.md

# 组合返回多种结构化数据（文本 + 目录 + 表格）
jklz-parse parse document.pdf --return content#toc#table

# 提取 Excel 表格
jklz-parse parse data.xlsx --return table

# 高精度图像模式（复杂版式）
jklz-parse parse document.pdf --image-mode vl

# 指定页面范围
jklz-parse parse document.pdf --page-range "0,3-5,-1"

# 已有 userId/jobId/fileId 时直接导出，不重新解析
jklz-parse export userId jobId fileId --type html -o result.html
```

### 启用/配置

**Linux/macOS：**
```bash
export JKLZ_PARSE_APIKEY="your_key"
export JKLZ_PARSE_BASEURL="http://192.168.42.15:15216"
```

**Windows CMD：**
```cmd
set JKLZ_PARSE_APIKEY=your_key
set JKLZ_PARSE_BASEURL=http://192.168.42.15:15216
```

**Windows PowerShell：**
```powershell
$env:JKLZ_PARSE_APIKEY="your_key"
$env:JKLZ_PARSE_BASEURL="http://192.168.42.15:15216"
```

---

## Skill 能力范围

| 能力 | 说明 |
|------|------|
| 多格式支持 | PDF、Word（.doc/.docx）、Excel（.xlsx）、PPT、图片等 |
| 多目标输出 | Markdown 文本（content）、HTML、表格（table）、目录（toc）、切片（slice）、chunks |
| 性能模式 | `vl`（高精度） / `cv`（高性能，默认） |
| 精准页码 | `--page-range` 指定解析页（如 `1-5,10`） |
| 智能导出 | `-o result.md/.html/.docx/.xlsx` 自动调用服务端导出接口 |
| 独立导出 | `export` 命令可基于 `userId/jobId/fileId` 导出已有解析结果 |
| 完整容错 | 502/503 自动重试，流式 JSON 解析器支持异常格式 |
| 多 API 支持 | 流式解析、非流式获取、搜索、导出、历史管理等 |

---

## 跨平台构建

预编译二进制发布在 GitHub Releases；当前仓库也保留 `cli/build/` 目录中的同名构建产物，便于内网或离线分发。

| 平台 | 文件 |
|------|------|
| macOS ARM64 (Apple Silicon) | `jklz-parse-darwin-arm64` |
| macOS AMD64 (Intel) | `jklz-parse-darwin-amd64` |
| Linux AMD64 | `jklz-parse-linux-amd64` |
| Linux ARM64 | `jklz-parse-linux-arm64` |
| Windows x64 | `jklz-parse-windows-x64.exe` |
| Windows x86 | `jklz-parse-windows-x86.exe` |

自行编译：
```bash
cd cli
./build.sh all
```

## 技术要求

- **Python**: 3.8+ (需安装 `requests` 库)
- **Go**: 1.21+ (如需自行编译)

## 许可协议

MIT
