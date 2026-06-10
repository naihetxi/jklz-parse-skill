# 🧠 jklz-parse-skill — 让 AI Agent 学会解析各种长文档

[![Darwin Score](https://img.shields.io/badge/Darwin%20Score-86.6%2F100-brightgreen)]()
[![Skill-Creator](https://img.shields.io/badge/Skill--Creator-Standard-blue)]()

> 安装这个 Skill，你的 AI Agent（如 Cursor, Claude Code, OpenClaw 等）就能理解如何使用 `jklz-parse` 命令行工具来解析 PDF、Word、Excel、PPT 等文件。对着 AI 说一句「帮我解析这份合同并提取表格」，剩下的交给它。

## 这是什么？

这是一份 AI Agent 的技能描述文件（`SKILL.md`），教会 AI Agent 如何调用内置的 `jklz-parse` 命令行工具来完成复杂文档的智能解析：

```
你说：「帮我解析 report.pdf 并提取其中的表格」

AI 自动执行：检测文件格式 → 选择高精度解析模式 → 提取 Markdown 文本和 JSON 表格 → 整理数据并返回
```

### CLI 和 Skill 的关系

| | CLI（命令行工具） | Skill（技能描述文件） |
|---|---|---|
| **是什么** | 位于 `cli/` 下的 Python / Go 命令行解析程序 | 一份教 AI 怎么用这些命令的详细说明书（`SKILL.md`） |
| **类比** | 一套锋利的解剖刀 | 一本解剖操作指南 |
| **单独能用吗** | **可以**。作为独立程序在终端中直接运行。 | **可以**。推荐配合本项目的 CLI 工具使用，但在受限沙盒环境中也能独立作为 API 调用指南使用。 |

简单说：**CLI 是手脚，Skill 是大脑。** 两者配合，AI Agent 才能完整地帮你阅读和解析长文档。

---

## 快速安装

### 第 1 步：配置解析工具 CLI

本项目内置了免编译的 Python CLI 工具和高性能跨平台编译包。你需要先为它配置 API Key。

**配置 API 密钥**（假设你已经克隆了本仓库）：
```bash
# Python CLI 配置方式（需安装 requests: pip install requests）
python3 cli/jklz-parse.py config --api-key YOUR_API_KEY --base-url http://192.168.42.15:15216

# 或使用 Go 预编译包配置方式（以 macOS ARM 为例）
./cli/build/jklz-parse-darwin-arm64 config --api-key YOUR_API_KEY --base-url http://192.168.42.15:15216
```
> 📧 没有 API Key？请联系金科览智服务管理员获取。

### 第 2 步：安装 Skill 到你的 AI Agent

Skill 由 `SKILL.md` 和 `references/` 目录共同组成。将整个仓库 clone 到 Agent 的技能目录即可：

**Windsurf / Claude Code：**
```bash
mkdir -p /path/to/your/project/.skills
git clone https://github.com/naihetxi/jklz-parse-skill.git \
  /path/to/your/project/.skills/jklz-parse-skill
```

**Cursor：**
```bash
mkdir -p /path/to/your/project/.cursor/rules
git clone https://github.com/naihetxi/jklz-parse-skill.git \
  /path/to/your/project/.cursor/rules/jklz-parse-skill
```

**小龙虾 OpenClaw：**
```bash
mkdir -p ~/.openclaw/skills
git clone https://github.com/naihetxi/jklz-parse-skill.git \
  ~/.openclaw/skills/jklz-parse-skill
```

**Codex：**
```bash
mkdir -p ~/.codex/skills
git clone https://github.com/naihetxi/jklz-parse-skill.git \
  ~/.codex/skills/jklz-parse-skill
```

### 第 3 步：开始对话！

安装完成后，直接用自然语言和你的 AI 交流：

- 「帮我解析这个 PDF 文件」
- 「把这个 Word 文档转成 Markdown」
- 「提取这份 Excel 中的所有表格数据」
- 「把这篇超长文档切片（slice），我需要用来做 RAG 知识库」

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

## Skill 能力范围

| 能力 | 说明 |
|------|------|
| 多格式支持 | PDF、Word（.doc/.docx）、Excel、PPT、图片等格式 |
| 多目标输出 | Markdown 文本（content）、HTML 源码、提取表格（table）、文档目录（toc）、切片（slice） |
| 性能模式切换 | `vl`（高精度，适合复杂版式和表格） / `cv`（高性能，适合快速处理大文件） |
| 精准页面控制 | 支持 `--page-range` 解析指定页码（如 `1-5,10`） |
| 完整容错处理 | 自动包含 API 检查、健康探测及网络连通性重试工作流 |

---

## 手动使用 CLI 工具（可选）

如果你不想通过 AI Agent，也可以直接在终端使用本项目内置的工具。详细使用说明请参考：
- [CLI 工具 README](cli/README.md)
- [CLI 快速使用指南](cli/快速使用指南.md)
- [API 参考手册](references/api.md)

**常用命令一览：**
```bash
# 解析 PDF 并保存为 Markdown
./cli/build/jklz-parse-darwin-arm64 parse document.pdf -o result.md

# 组合返回结构化数据（文本 + 目录 + 表格）
./cli/build/jklz-parse-darwin-arm64 parse document.pdf --return content#toc#table

# 使用 Python 版本提取表格
python3 cli/jklz-parse.py parse data.xlsx --return table
```

## 技术要求

- **Python**: 3.8+ (需安装 `requests` 库)
- **Go** (可选): 1.21+ (如果你需要自行编译)
- **预编译包**: 支持 MacOS(Intel/ARM)、Linux(AMD64)、Windows(x64)，均位于 `cli/build/` 目录下。

## 许可协议

MIT
