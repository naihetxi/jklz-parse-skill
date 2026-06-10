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
  "合同分析"、"文档转HTML"、"获取文档目录"、"文档切片用于RAG"、"从文档提取"、"文档结构分析"。
  
  高级功能：流式解析、页码选择、页眉页脚过滤、性能模式切换（cv高性能/vl高精度）。
  
  配置要求：需要 API Key（联系金科览智管理员申请）。
homepage: https://github.com/naihetxi/jklz-parse-skill
metadata:
  tags: document-parse, pdf, docx, xlsx, table-extraction, ocr, rag, content-extraction, markdown
  platforms: openclaw, claude-code, codex, cursor
  version: 1.0.0
  openclaw:
    emoji: '📄'
    requires: { env: ['JKLZ_PARSE_APIKEY'] }
    primaryEnv: 'JKLZ_PARSE_APIKEY'
---

# jklz-parse-skill

文档智能解析技能，教导 AI 如何使用本项目内的 CLI 工具，实现对 PDF、DOC、DOCX、XLSX、PPT 等多种格式的解析。

## Setup

> **重要**: 此技能依赖本仓库自带的 `cli/jklz-parse.py` 或 `cli/build/` 下的各平台二进制命令行文件。同时需要调用远程 API，必须先配置 API Key。

### 1. 配置 API Key

可以通过 `cli` 工具直接配置（配置会自动写入 `~/.config/jklz-parse/config.json`），或通过环境变量。

```bash
# Python CLI（兼容所有平台）
python3 <runtime-skills-dir>/jklz-parse-skill/cli/jklz-parse.py config --api-key YOUR_API_KEY --base-url http://192.168.42.15:15216

# 或通过环境变量传递
export JKLZ_PARSE_APIKEY="YOUR_API_KEY"
export JKLZ_PARSE_BASEURL="http://192.168.42.15:15216"
```

### 2. 检查工具可用性

测试网络连通性和配置是否生效：
```bash
python3 <runtime-skills-dir>/jklz-parse-skill/cli/jklz-parse.py health
```

## 工作流程

### Phase 1: 预检查

**Step 1.1 — 验证文件存在**
```bash
if [ ! -f "$file_path" ]; then
  echo "错误：文件不存在 $file_path"
  exit 1
fi
```

**Step 1.2 — 检查参数边界**
如果用户没有明确要求提取其他结构（如 `toc`，`table`），默认使用 `return=content` 提取文本。
如果用户没有明确要求高精度视觉解析，默认使用 `image_mode=cv`（高性能模式）；仅在复杂版式、表格密集、普通模式结果不完整时切换为 `image_mode=vl`。

### Phase 2: 执行解析

**Step 2.1 — 调用 CLI 解析**

你可以选择调用 Python 版本或对应你所运行架构的预编译 Go 版本。
> 注意：Go 预编译版本通常位于 `<runtime-skills-dir>/jklz-parse-skill/cli/build/` 目录下（如 `jklz-parse-darwin-arm64`、`jklz-parse-windows-x64.exe`）。

在 macOS/Linux/GitBash 等类 Unix 环境下，执行如下 Bash 命令：

```bash
# 使用内置的 CLI 工具（以 Python 版本为例，安全稳定通用）
python3 <runtime-skills-dir>/jklz-parse-skill/cli/jklz-parse.py parse "${file_path}" \
  --return "${return_types}" \
  --image-mode "${image_mode:-cv}" \
  -o "output_result.md"
```

若处于原生 **Windows 平台 (PowerShell/CMD)**，使用 `.exe` 并避免 Bash 专属换行符即可：

```powershell
# 使用 Go 编译好的 Windows 版本（推荐）
& "<runtime-skills-dir>\jklz-parse-skill\cli\build\jklz-parse-windows-x64.exe" parse "${file_path}" --return "${return_types}" --image-mode "cv" -o "output_result.md"
```

**Step 2.1(a) — 脱离 CLI 的纯 API 调用（备用方案）**

如果你所处的沙盒环境无法运行本地脚本或二进制文件，你也可以作为后备方案直接通过 `curl` 调用 API（虽然丢失了 CLI 的健壮封装）：

在类 Unix 环境中：
```bash
curl -s -X POST "${JKLZ_PARSE_BASEURL:-http://192.168.42.15:15216}/service/document/parse/stream/v1" \
  -F "file=@${file_path}" -F "api_key=${PARSE_API_KEY}" -F "return=${return_types}"
```

在 Windows PowerShell 环境中：
```powershell
curl.exe -s -X POST "${env:JKLZ_PARSE_BASEURL:-http://192.168.42.15:15216}/service/document/parse/stream/v1" -F "file=@${file_path}" -F "api_key=${env:PARSE_API_KEY}" -F "return=${return_types}"
```


**Step 2.2 — 失败处理（if-then 分支）**

| 触发条件 | 建议对策 | 仍失败兜底 |
|---------|---------|----------|
| CLI 返回 502/503 | 等待 5s 后重试 1 次 | 尝试重试或等待服务端恢复 |
| CLI 返回 401 | 检查 API Key 配置 | 提示用户重新配置 |
| 解析结果异常 / 超时 | 检查文件大小 | 自动切换为 `--return slice` 模式切片 |

### Phase 3: 后处理与结果提取

**Step 3.1 — 读取提取结果**

CLI 执行完成后，文件内容会被保存至你指定的输出文件（如上面的 `output_result.md`）。读取此文件以完成你的任务。

## return 参数说明

| 值 | 描述 | 适用场景 |
|---|------|---------|
| `content` | Markdown 格式文本 | 提取文档正文内容 |
| `html` | HTML 格式（保留排版） | 保留原文格式样式 |
| `toc` | 目录结构 | 获取文档目录树 |
| `table` | 表格信息 | 提取所有表格（返回 JSON 结构） |
| `slice` | 文本切片结果 | RAG 场景，知识库构建 |

多个值用 `#` 分隔，如：`--return content#toc#table`。

## 常见场景参数组合

| 场景 | 参数 | 说明 |
|-----|------|------|
| 提取纯文本 | `--return content` | 获取 Markdown 格式正文 |
| 提取目录 | `--return toc` | 获取文档结构 |
| 提取表格 | `--return table` | 提取所有表格 |
| 完整解析 | `--return content#toc#table` | 文本+目录+表格 |
| 精度要求高 | `--image-mode vl` | 适合复杂版式、表格密集的场景 |
| 页面裁剪 | `--page-range "0,3-5,-1"` | 指定读取第1页+4-6页+最后一页 |

## ⚠️ 反例与黑名单（不要做的事）

### 🚫 禁止操作

1. **不要不带 `-o` 选项就处理极长文档**
   - 终端直接打印可能会刷屏，请将结果输出至文件后再读取所需部分。
3. **不要忽略配置报错**
   - 如果 CLI 提示缺少配置，必须引导用户先运行 `config` 设定 `api-key`。

### ⚠️ 不适用场景

此 skill **不应该**用于以下任务：
- **实时编辑文档**：只提取，不修改。
- **格式转换**：不做 PDF→Word 等同构转换操作。
