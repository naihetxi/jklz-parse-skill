---
name: jklz-parse-skill
description: |
  金科览智文档智能解析技能。支持 PDF、Word、Excel、PPT 等格式的高精度解析，提取文本、表格、目录结构。
  
  核心能力：
  - 文本提取：Markdown/HTML 格式输出，保留文档结构
  - 表格识别：跨页表格自动合并，支持复杂嵌套表格
  - 目录提取：自动识别章节层级，生成完整目录树
  - 文档切片：智能分段，优化 RAG 检索效果
  - 内容溯源：追踪内容在原文档中的位置
  
  触发词："解析文档"、"提取PDF内容"、"Word转Markdown"、"Excel提取表格"、"简历解析"、
  "合同分析"、"文档转HTML"、"获取文档目录"、"文档切片用于RAG"、"从文档提取"、"文档结构分析"、
  "查询解析历史"、"取消解析任务"、"搜索文档关键词"、"清理历史文件"、"导出解析结果"。
  
  高级功能：流式解析、页码选择、页眉页脚过滤、性能模式切换（cv高性能/vl高精度）。
  
  配置要求：需要 API Key（联系金科览智管理员申请）。
homepage: https://github.com/naihetxi/jklz-parse-skill
metadata:
  tags: document-parse, pdf, docx, xlsx, table-extraction, ocr, rag, content-extraction, markdown
  platforms: openclaw, claude-code, codex, cursor
  version: 1.1.0
  openclaw:
    emoji: '📄'
    requires: { env: ['JKLZ_PARSE_APIKEY'] }
    primaryEnv: 'JKLZ_PARSE_APIKEY'
---

## 架构说明

本技能提供 **三种使用路径**，相互独立、可自由组合：

### 路径 1：AI 智能体直接调用 API（无需 CLI）
智能体（Claude Code、Cursor、Codex 等）可直接通过 HTTP/curl 调用远程解析 API，**不依赖 CLI 工具**。技能 SKILL.md 本身即为智能体的"咒语书"，告诉 AI 如何构造请求、处理响应。
- 优势：零安装、跨平台、任何支持 HTTP 的沙箱都可使用
- 劣势：没有 CLI 的健壮封装（自动重试、进度提示、格式化输出等）
- **缺失 API Key 时**：智能体应主动提示用户提供 Key，而不是静默失败

### 路径 2：CLI 独立使用（无需技能/SKILL.md）
CLI 工具（`jklz-parse.py` / 各平台二进制）是独立可执行程序，用户可直接通过命令行操作，完全不依赖 AI 智能体或 SKILL.md。
- 支持命令：`parse`、`get`、`export`、`history`、`cancel`、`modify`、`cleanup`、`search`、`health`、`config`
- CLI 会自动处理重试、流解析、导出下载等逻辑

### 路径 3：智能体 + CLI 组合（推荐最优路径）
智能体调用本地 CLI 工具执行 API 请求，获得 CLI 的所有健壮特性。

---

## 快速开始：配置与验证

### API Key 配置（三选一）

#### 方式 1：CLI 配置（推荐）
API Key 和 Base URL 会保存到 `~/.config/jklz-parse/config.json`。

```bash
# Linux/macOS/Windows Git Bash
python3 <skill-dir>/cli/jklz-parse.py config --api-key YOUR_KEY --base-url http://192.168.42.15:15216

# Windows CMD / PowerShell (Go binary)
<skill-dir>\cli\build\jklz-parse-windows-x64.exe config --api-key YOUR_KEY --base-url http://192.168.42.15:15216
```

#### 方式 2：环境变量
```bash
# Linux/macOS
export JKLZ_PARSE_APIKEY="YOUR_API_KEY"
export JKLZ_PARSE_BASEURL="http://192.168.42.15:15216"
```
```cmd
REM Windows CMD
set JKLZ_PARSE_APIKEY=YOUR_API_KEY
set JKLZ_PARSE_BASEURL=http://192.168.42.15:15216
```
```powershell
# Windows PowerShell
$env:JKLZ_PARSE_APIKEY="YOUR_API_KEY"
$env:JKLZ_PARSE_BASEURL="http://192.168.42.15:15216"
```

## 工作流程

### Phase 1: 预检查

**Step 1.1 — 验证文件存在**
```bash
test -f "${file_path}" || { echo "错误：文件不存在 ${file_path}"; exit 1; }
```

**Step 1.2 — 检查参数边界**
- 未指定提取类型时，默认 `--return content`
- 图像解析模式默认使用 `--image-mode cv`（高性能）
- 不需要 `toc`，`table` 等结构时无需在 `return` 中指定
- 如果 `JKLZ_PARSE_APIKEY` 和 CLI 配置都不存在，先提示用户配置 API Key，不要继续调用接口

**🔴 CHECKPOINT / STOP — 执行边界**
- **STOP**：文件不存在、API Key 缺失、Base URL 不可达时，停止调用接口并把缺失项明确告诉用户。
- **CHECKPOINT**：执行 `cancel`、`modify`、`cleanup` 前，必须向用户复述 `userId/jobId/fileId/time` 和影响范围，获得明确确认后再执行。
- **STOP**：用户未确认状态变更命令时，不执行 `cancel`、`modify`、`cleanup`。

### Phase 2: 执行解析

**Step 2.1 — 调用 CLI 解析**

> Go 预编译版本位于 `<skill-dir>/cli/build/`，命名规则：`jklz-parse-{os}-{arch}[.exe]`

```bash
# macOS / Linux / Git Bash
python3 <skill-dir>/cli/jklz-parse.py parse "${file_path}" \
  --return "${return_types:-content}" \
  --image-mode "${image_mode:-cv}" \
  --output "${output_file:-result.md}"

# Windows CMD / PowerShell
& "<skill-dir>\cli\build\jklz-parse-windows-x64.exe" parse "${file_path}" --return "content#toc#table" --image-mode "cv" --output "result.md"
```

**Step 2.1(a) — 直接 API 调用（不使用 CLI）**

如果所处沙盒环境无法运行本地脚本或二进制文件，也可以作为后备方案直接通过 `curl` 调用 API（虽然丢失了 CLI 的健壮封装）：

类 Unix 环境：
```bash
curl -s -X POST "${JKLZ_PARSE_BASEURL:-http://192.168.42.15:15216}/service/document/parse/stream/v2" \
  -F "file=@${file_path}" \
  -F "apiKey=${JKLZ_PARSE_APIKEY}" \
  -F "streamType=lz" \
  -F "return=${return_types:-content}" \
  -F "imageParseMode=${image_mode:-cv}"
```

Windows PowerShell 环境：
```powershell
$baseUrl = if ($env:JKLZ_PARSE_BASEURL) { $env:JKLZ_PARSE_BASEURL } else { "http://192.168.42.15:15216" }
$returnTypes = if ($return_types) { $return_types } else { "content" }
curl.exe -s -X POST "$baseUrl/service/document/parse/stream/v2" `
  -F "file=@$file_path" `
  -F "apiKey=$env:JKLZ_PARSE_APIKEY" `
  -F "streamType=lz" `
  -F "return=$returnTypes" `
  -F "imageParseMode=cv"
```

**Step 2.2 — 失败处理（if-then 分支）**

| 触发条件 | 处理方式 | 仍失败兜底 |
|---------|---------|----------|
| CLI 返回 502/503 | 等待 5s 后重试 1 次 | 尝试重试或等待服务端恢复 |
| CLI 返回 401 | 检查 API Key 配置 | 提示用户重新配置 |
| 解析结果异常 / 超时 | 检查文件大小 | 自动切换为 `--return slice` 模式切片 |

### Phase 3: 后处理与结果提取

**Step 3.1 — 读取提取结果**

CLI 执行完成后，文件内容会被保存至你指定的输出文件（如 `output_result.md`）。读取此文件以完成你的任务。

## return 参数说明

`return` 参数指定解析结果类型，多个类型用 `#` 分隔。

| 值 | 描述 | 适用场景 |
|---|------|---------|
| `content` | Markdown 格式文本 | 提取文档正文内容 |
| `html` | HTML 格式（保留排版） | 保留原文格式样式 |
| `toc` | 目录结构 | 获取文档目录树 |
| `table` | 表格信息 | 提取所有表格（返回 JSON 结构） |
| `slice` | 文本切片结果 | RAG 场景，知识库构建 |
| `chunks` | 原始解析块 | 获取原始 chunk 数据 |
| `page` | 页码信息 | 按页抽取内容 |
| `properties` | 文档属性 | 文档元信息（作者、页数等） |
| `file` | 文件下载链接 | 获取原始或中间文件链接 |
| `uloc` | 溯源定位信息 | 内容定位和审计 |
| `pkl` | 原始 pkl 数据 | 调试和高级处理 |

示例：`--return content#toc#table`

## 常见场景参数组合

| 场景 | 参数 | 说明 |
|-----|------|------|
| 提取纯文本 | `--return content` | 获取 Markdown 格式正文 |
| 提取目录 | `--return toc` | 获取文档结构 |
| 提取表格 | `--return table` | 提取所有表格 |
| 完整解析 | `--return content#toc#table` | 文本+目录+表格 |
| 精度要求高 | `--image-mode vl` | 适合复杂版式、表格密集的场景 |
| 页面裁剪 | `--page-range "0,3-5,-1"` | 指定读取第1页+4-6页+最后一页 |
| 导出有样式文件 | `-o result.html` 或 `export ... --type html` | 调用服务端导出接口 |

## CLI 完整命令参考

### `parse` — 解析文档
```
parse <file> [--return content#toc#table] [--image-mode cv|vl] [--page-range "0,3-5"] [--output result.md]
```

### `get` — 获取已解析结果
```
get <userId> <jobId> <fileId> [-r content,html,toc]
```

### `export` — 导出已解析结果
```
export <userId> <jobId> <fileId> --type md|html|docx|xlsx -o result.md
export <userId> <jobId> <fileId> --type html --chunk-id <chunkId> -o chunk.html
```

### `history` — 查询历史记录
```
history <userId>
```

### `search` — 关键词检索
```
search <userId> <jobId> <fileId> -k "关键词1,关键词2"
```

### `cancel` — 取消解析任务
```
cancel <userId> <jobId>
```

### `modify` — 修改 chunk 内容
```
modify <userId> <jobId> <fileId> -c <chunkId> -t "新文本"
```

### `cleanup` — 清理历史文件
```
cleanup <userId> <time>   # e.g., 7d, 24h, 30m
```

### `config` — API 配置
```
config --api-key YOUR_KEY [--base-url http://...]
config --show    # 查看当前配置
```

### `health` — 健康检查
```
health [--base-url http://...]
```

## 输出文件格式与导出

当使用 `--output result.md` （`.docx`、`.html`、`.xlsx`）时，CLI 会自动检测文件扩展名：
- `.md` / `.html` / `.docx` / `.xlsx`：调用服务端导出接口（`/service/document/export/v2`），下载格式化文件
- 服务端返回 zip 包时，CLI 会自动提取目标文件，并把图片等资源目录保存到输出文件同级目录
- 其他扩展名（如 `.json`）：直接保存解析结果原文本

不使用 `-o` 时，结果直接输出到 `stdout`。

已存在 `userId/jobId/fileId` 时，不需要重新解析，直接导出：

```bash
jklz-parse export <userId> <jobId> <fileId> --type md -o result.md
jklz-parse export <userId> <jobId> <fileId> --type html --chunk-type table -o tables.html
```

## 达尔文评分机制说明

本技能遵循达尔文技能评分体系，评分项包括：
- **安全性**：API Key 通过配置文件或环境变量管理，不在代码中硬编码
- **鲁棒性**：CLI 自动处理 502/503 重试、连接超时、流解析异常
- **跨平台**：提供 Python 通用版本 + Go 编译的 Linux/macOS/Windows 多架构二进制
- **清晰性**：SKILL.md 明确说明 CLI 与 Skill 的独立性与组合方式
- **可追溯性**：所有参数说明与 API 文档对齐，支持透明调试

## ⚠️ 反例与黑名单（不要做的事）

### 🚫 禁止操作

1. **不要不带 `-o` 选项就处理极长文档**
   - 终端直接打印可能会刷屏，请将结果输出至文件后再读取所需部分。
2. **CLI 与 Skill 是独立模块，不要相互强制依赖**
   - 智能体可在没有 CLI 的情况下通过 API 直接调用 parse 服务
   - CLI 可在没有 Skill 的情况下独立提供给用户使用
3. **不要忽略配置报错**
   - 如果 CLI 提示缺少配置，必须引导用户先运行 `config` 设定 `api-key`。

### ⚠️ 不适用场景

此 skill **不应该**用于以下任务：
- **实时编辑文档**：只提取，不修改。
- **同构格式转换**：不做 PDF→Word 这类保真版式转换；只导出解析结果为 `md/html/docx/xlsx`。
