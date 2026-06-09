---
name: jklz-parse-skill
description: |
  文档智能解析技能。解析 PDF、DOC、DOCX、XLSX、PPT 等文档，提取文本、表格、目录结构。
  触发词："解析文档"、"提取PDF内容"、"Word转Markdown"、"Excel提取表格"、"简历解析"、
  "合同分析"、"文档转HTML"、"获取文档目录"、"文档切片用于RAG"、"从文档提取"、"文档结构分析"。
  支持流式解析、页码选择、跨页表格合并、页眉页脚过滤、文档溯源等高级功能。
homepage: http://192.168.42.15:15216
metadata:
  tags: document-parse, pdf, docx, xlsx, table-extraction, ocr, rag, content-extraction
  platforms: openclaw, claude-code, codex
  openclaw:
    emoji: '📄'
    requires: { env: ['JKLZ_PARSE_APIKEY'] }
    primaryEnv: 'JKLZ_PARSE_APIKEY'
  security:
    credentials_usage: |
      This skill requires a Parse API key to authenticate with the Parse API service.
      Credentials are ONLY sent as HTTP form data to the configured Parse API endpoint.
      No credentials are logged, stored in files, or transmitted to any other destination.
    allowed_domains:
      - 192.168.42.15
---

# jklz-parse-skill

文档智能解析技能，支持 PDF、DOC、DOCX、XLSX、PPT 等多种格式的流式解析。

## Setup

> **重要**: 此技能依赖远程 Parse API 服务，必须先配置 API Key 才能使用。

### 1. 配置 API Key（二选一）

**方式 A — 配置文件（推荐）**:

```bash
mkdir -p ~/.config/jklz-parse
echo "your_api_key" > ~/.config/jklz-parse/api_key
echo "http://192.168.42.15:15216" > ~/.config/jklz-parse/base_url
```

**方式 B — 环境变量**:

```bash
export JKLZ_PARSE_APIKEY="your_api_key"
export JKLZ_PARSE_BASEURL="http://192.168.42.15:15216"
```

### 2. 验证配置

```bash
# 检查 API 是否可用
curl -s "$(cat ~/.config/jklz-parse/base_url 2>/dev/null)/metrics"
# 应返回: {"status":"success","message":"Service is healthy"}
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

**Step 1.2 — 检查 API 凭证**
```bash
PARSE_API_KEY="${JKLZ_PARSE_APIKEY:-$(cat ~/.config/jklz-parse/api_key 2>/dev/null)}"
PARSE_BASE_URL="${JKLZ_PARSE_BASEURL:-$(cat ~/.config/jklz-parse/base_url 2>/dev/null)}"

if [ -z "$PARSE_API_KEY" ]; then
  echo "错误：API Key 未配置"
  echo "请配置：echo 'your_key' > ~/.config/jklz-parse/api_key"
  exit 1
fi
```

**Step 1.3 — 检查文件大小和页数**
```bash
file_size=$(stat -f%z "$file_path" 2>/dev/null || stat -c%s "$file_path")
page_count=$(pdfinfo "$file_path" 2>/dev/null | grep Pages | awk '{print $2}')

# 🔴 CHECKPOINT: 大文件确认
if [ "$page_count" -gt 100 ]; then
  echo "⚠️  检测到大文件：$page_count 页"
  echo "建议使用 return=slice 分块处理，是否继续？(y/n)"
  read -r confirm
  if [ "$confirm" != "y" ]; then
    exit 0
  fi
fi
```

### Phase 2: 执行解析

**Step 2.1 — 构建参数**

根据用户意图选择参数组合（参考「常见场景参数组合」表）。

**Step 2.2 — 调用 API**

```bash
curl -s -X POST "${PARSE_BASE_URL}/service/document/parse/stream/v1" \
  -F "file=@${file_path}" \
  -F "api_key=${PARSE_API_KEY}" \
  -F "stream_type=lz" \
  -F "return=${return_types}" \
  -F "image_parse_mode=${image_mode}"
```

**Step 2.3 — 失败处理（if-then 分支）**

| 触发条件 | 一线修复 | 仍失败兜底 |
|---------|---------|----------|
| API 返回 502/503 | 等待 5s 后重试 1 次 | 切换 `image_parse_mode=cv` 再试 |
| API 返回 401 | 检查 API Key 是否过期 | 提示用户重新配置凭证 |
| 响应超时（>120s） | 检查文件大小，若 >200 页 | 自动切换为 `return=slice&split_max_length=512` |
| VL 模式解析失败 | 自动切换 `image_parse_mode=cv` | 提示用户该文档可能不支持 OCR |
| 跨页表格未合并 | 启用 `cross_page_table_merge_support=1&filter_hf_support=1` | 提示用户手动处理表格 |

### Phase 3: 解析响应

**Step 3.1 — 提取 parse_return 数据**

```bash
# 使用内置脚本解析 SSE 流
curl ... | node <runtime-skills-dir>/jklz-parse-skill/scripts/parse-response.cjs
```

**Step 3.2 — 验证结果**

检查关键字段是否存在：
- `content` 不为空
- `file_name` 匹配输入文件
- 无 `error` 字段

### Phase 4: 后处理

**Step 4.1 — 格式化输出**

根据用户需求展示结果（Markdown/HTML/表格）。

**Step 4.2 — 保存历史记录**

解析结果自动保存到服务端，可通过 `job_id` 和 `file_id` 回溯。

## API 调用模板

### 核心函数

```bash
# 加载凭证
PARSE_API_KEY="${JKLZ_PARSE_APIKEY:-$(cat ~/.config/jklz-parse/api_key 2>/dev/null)}"
PARSE_BASE_URL="${JKLZ_PARSE_BASEURL:-$(cat ~/.config/jklz-parse/base_url 2>/dev/null)}"

# 流式解析
parse_stream() {
  local file_path="$1"
  local return_types="${2:-content}"    # content, html, toc, table, slice, chunks
  local image_mode="${3:-vl}"            # vl (高精度) 或 cv (高性能)

  curl -s -X POST "${PARSE_BASE_URL}/service/document/parse/stream/v1" \
    -F "file=@${file_path}" \
    -F "api_key=${PARSE_API_KEY}" \
    -F "stream_type=lz" \
    -F "return=${return_types}" \
    -F "image_parse_mode=${image_mode}"
}

# 获取历史解析结果
parse_get() {
  local user_id="${1:-jklz}"
  local job_id="$2"
  local file_id="$3"

  curl -s -X POST "${PARSE_BASE_URL}/service/document/parse/get/v1" \
    -H "Content-Type: application/json" \
    -d "{\"user_id\":\"${user_id}\",\"job_id\":\"${job_id}\",\"file_id\":\"${file_id}\",\"return_type_list\":[\"content\",\"toc\"]}"
}

# 查询历史记录
parse_history() {
  curl -s -X POST "${PARSE_BASE_URL}/service/document/parse/history/v1" \
    -H "Content-Type: application/json" \
    -d "{\"user_id\":\"jklz\"}"
}

# 清理历史文件
parse_cleanup() {
  local time="${1:-7d}"  # 12d, 25h, 3w, 20m
  curl -s -X POST "${PARSE_BASE_URL}/service/document/parse/cleanup/v1" \
    -H "Content-Type: application/json" \
    -d "{\"user_id\":\"jklz\",\"time\":\"${time}\"}"
}
```

## return 参数说明

| 值 | 描述 | 适用场景 |
|---|------|---------|
| `content` | Markdown 格式文本 | 提取文档正文内容 |
| `html` | HTML 格式（保留排版） | 保留原文格式样式 |
| `toc` | 目录结构 `[(标题, 级别), ...]` | 获取文档目录结构 |
| `table` | 表格信息数组 | 提取所有表格 |
| `slice` | 文本切片结果 | RAG 场景，知识库构建 |
| `chunks` | 原始解析块数据 | 获取分块信息 |
| `page` | 页面级别数据 | 按页处理 |
| `uloc` | 统一位置信息（含溯源） | 需要内容溯源 |

多个值用 `#` 分隔，如：`return=content#toc#table`

## 常见场景参数组合

| 场景 | 参数 | 说明 |
|-----|------|------|
| 提取纯文本 | `return=content` | 获取 Markdown 格式正文 |
| 提取目录 | `return=toc` | 获取文档结构 |
| 提取表格 | `return=table&table_format=markdown` | 表格转 Markdown |
| 完整解析 | `return=content#toc#table` | 文本+目录+表格 |
| RAG 切片 | `return=slice&split_type=toc&split_max_length=512` | 按目录切片，512 token |
| 保留格式 | `return=html` | 输出 HTML 保留排版 |
| 选择页面 | `page_selecte2parse=0,3-5,-1` | 第1页+4-6页+最后页 |
| 过滤页眉页脚 | `filter_hf_support=1` | 去除页眉页脚 |
| 跨页表格合并 | `cross_page_table_merge_support=1` | 自动合并跨页表格 |

## 输出格式解析

### SSE 流式输出

```
data: {"code":"200","data":{"type":"agent","value":"..."}}
data: {"code":"200","data":{"type":"parse_return","value":{"content":"...","toc":[...]}}}
data: {"code":"200","data":{"type":"stop","value":"parse files done"}}
```

### parse_return 结果结构

```json
{
  "user_id": "jklz",
  "job_id": "xxx",
  "file_id": "xxx",
  "file_name": "document.pdf",
  "content": "# Markdown内容",
  "html": "<!DOCTYPE html>...",
  "toc": [["第一章", 0], ["1.1节", 1]],
  "table": [{...}],
  "slice": [{...}],
  "chunks": [{...}]
}
```

### 使用脚本解析响应

```bash
# 使用内置脚本解析 SSE 响应
# 根据你的 runtime 调整路径
curl ... | node <runtime-skills-dir>/jklz-parse-skill/scripts/parse-response.cjs

# 示例路径：
# Claude Code: ~/.claude/skills/jklz-parse-skill/scripts/parse-response.cjs
# Codex: ~/.codex/skills/jklz-parse-skill/scripts/parse-response.cjs
```

## 高级参数

| 参数 | 默认值 | 说明 |
|-----|-------|------|
| `user_id` | "jklz" | 用户标识 |
| `job_id` | 随机 | 任务 ID |
| `image_parse_mode` | "vl" | "vl"=视觉语言模型，"cv"=计算机视觉 |
| `split_nested_table` | 0 | Word 嵌套表格拆分 |
| `trace` | 0 | 启用内容溯源 |
| `page_selecte2parse` | "" | PDF 页面选择 |
| `table_format` | "html" | "html" 或 "markdown" |
| `filter_hf_support` | 0 | 过滤页眉页脚 |
| `cross_page_table_merge_support` | 1 | 跨页表格合并 |
| `split_type` | "toc" | 切片方式：toc/length/custom |
| `split_max_length` | 512 | 最大切片长度 |
| `overlap` | false | 切片重叠 |
| `overlap_size` | 128 | 重叠长度 |
| `save_table` | true | 保存表格 |

## 错误处理

```bash
# 检查 API 健康状态
curl -s "${PARSE_BASE_URL}/metrics"

# 如果 VL 模式失败，尝试 CV 模式
curl -F "image_parse_mode=cv" ...

# 检查响应中的错误
data: {"code":"200","data":{"type":"error","value":{"error":"..."}}}
```

## 注意事项

1. **文件路径**: 使用绝对路径或确保文件在当前目录
2. **大文件**: 超过 100 页建议使用 `return=slice` 分块处理
3. **扫描版 PDF**: 使用 `image_parse_mode=vl` 获得最佳 OCR 效果
4. **跨页表格**: 启用 `cross_page_table_merge_support=1` 和 `filter_hf_support=1`
5. **溯源功能**: 设置 `trace=1&return=uloc` 获取内容在原文的位置
6. **流式输出**: 使用 `stream_type=lz` 获取自定义格式，`stream_type=sse` 获取标准 SSE

## ⚠️ 反例与黑名单（不要做的事）

### 🚫 禁止操作

1. **不要解析敏感文档而不经用户确认**
   - 含密级标识的文档（机密、秘密、绝密）
   - 个人隐私文档（身份证、护照、医疗记录）
   - 财务敏感文件（银行对账单、税务文件）
   - → 必须先询问用户是否继续

2. **不要忽略 API Key 缺失直接调用**
   - 检查 `JKLZ_PARSE_APIKEY` 或 `~/.config/jklz-parse/api_key`
   - 缺失时明确提示配置方法，不要静默失败

3. **不要对超大文件（>200页）使用 `return=content` 全量解析**
   - 会导致响应超时或内存溢出
   - 应该使用 `return=slice` 分块处理

4. **不要在生产环境使用默认 user_id**
   - 默认 `user_id=jklz` 仅用于测试
   - 生产环境应使用实际用户标识

5. **不要忽略 VL 模式失败后继续重试**
   - VL 模式失败时应自动切换到 CV 模式
   - 不要连续 3 次以上重试同一模式

### ⚠️ 不适用场景

此 skill **不应该**用于以下任务：

- **实时编辑文档**：此 skill 只读取解析，不修改原文件
- **格式转换**：不做 PDF→Word、Excel→CSV 等格式转换，只提取内容
- **OCR 图片**：虽然支持扫描版 PDF，但不是独立的 OCR 工具
- **文档生成**：不生成新文档，只解析现有文档
