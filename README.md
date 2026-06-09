# jklz-parse-skill 安装与使用指南

## 简介

`jklz-parse-skill` 是一个 Agent Skills 标准技能，用于智能解析 PDF、DOC、DOCX、XLSX、PPT 等多种格式的文档。兼容 Claude Code、Codex、Cursor、OpenClaw 等多种 runtime。

### 主要功能

- 📄 **多格式支持**: PDF、Word、Excel、PowerPoint、TXT 等
- 📝 **内容提取**: 文本、表格、目录、图片
- 🔪 **智能切片**: 按目录或长度切分，适用于 RAG 知识库
- 📊 **表格处理**: 自动识别、跨页合并、格式转换
- 🔍 **内容溯源**: 追踪内容在原文中的位置

## 安装

### 一行命令（自动检测 runtime）

```bash
# 自动检测并安装到对应 runtime 的 skills 目录
curl -fsSL https://github.com/yourusername/jklz-parse-skill/raw/main/install.sh | bash
```

### 手动安装

根据你使用的 runtime，选择对应路径：

| Runtime | 安装路径 |
|---------|---------|
| Claude Code | `~/.claude/skills/jklz-parse-skill/` |
| Codex | `~/.codex/skills/jklz-parse-skill/` |
| Cursor | `~/.cursor/skills/jklz-parse-skill/` |
| OpenClaw | `~/.openclaw/skills/jklz-parse-skill/` |
| 其他 | 参考对应 runtime 的 skills 目录 |

```bash
# 示例：手动安装到 Claude Code
mkdir -p ~/.claude/skills/jklz-parse-skill
cp -r jklz-parse-skill/* ~/.claude/skills/jklz-parse-skill/
ls ~/.claude/skills/jklz-parse-skill/
# 应该看到: SKILL.md, references/, scripts/
```

## 配置

### 1. 获取 API Key

联系管理员获取 Parse API 服务的访问密钥。

### 2. 配置凭证

```bash
# 创建配置目录
mkdir -p ~/.config/jklz-parse

# 保存 API Key
echo "your_api_key_here" > ~/.config/jklz-parse/api_key

# 保存服务地址（内网地址）
echo "http://192.168.42.15:15216" > ~/.config/jklz-parse/base_url
```

### 3. 验证配置

```bash
# 测试服务连通性
curl -s "$(cat ~/.config/jklz-parse/base_url)/metrics"
# 预期输出: {"status":"success","message":"Service is healthy"}
```

## 使用方法

### 在支持的 runtime 中使用

安装并配置后，技能会自动触发。你可以这样说：

```
"解析这个 PDF 文件"
"把 Word 文档转成 Markdown"
"提取 Excel 中的所有表格"
"分析这份合同的内容"
"把这个文档切片用于知识库"
```

### 直接调用 API

```bash
# 设置环境变量
export PARSE_API_KEY=$(cat ~/.config/jklz-parse/api_key)
export PARSE_BASE_URL=$(cat ~/.config/jklz-parse/base_url)

# 解析文档
curl -X POST "${PARSE_BASE_URL}/service/document/parse/stream/v1" \
  -F "file=@document.pdf" \
  -F "api_key=${PARSE_API_KEY}" \
  -F "return=content#toc"
```

## 常见问题

### Q: 技能没有自动触发？

A: 确保技能文件在正确位置，并且尝试使用更明确的触发词，如：
- "解析文档"
- "提取 PDF 内容"
- "Word 转 Markdown"

### Q: VL 模式返回 502 错误？

A: 尝试切换到 CV 模式：
```bash
-F "image_parse_mode=cv"
```

### Q: 如何只解析特定页面？

A: 使用 `page_selecte2parse` 参数：
```bash
-F "page_selecte2parse=0,3-5,-1"  # 第1页、第4-6页、最后一页
```

### Q: 表格跨页断开了？

A: 启用跨页表格合并：
```bash
-F "cross_page_table_merge_support=1" \
-F "filter_hf_support=1"
```

## 文件结构

```
jklz-parse-skill/
├── SKILL.md           # 技能主文件
├── README.md          # 本说明文档
├── references/
│   └── api.md         # API 详细参考
└── scripts/
    └── parse-response.cjs  # SSE 响应解析脚本
```

## 技术支持

- API 文档: http://192.168.42.15:8086/docs/agent-api/
- 服务地址: http://192.168.42.15:15216

## 更新日志

### v1.0.0 (2026-03-24)

- 初始版本
- 支持 PDF、DOC、DOCX、XLSX、PPT 等格式
- 支持流式解析、表格提取、目录解析
- 支持 VL/CV 两种解析模式
- 支持文档溯源功能
