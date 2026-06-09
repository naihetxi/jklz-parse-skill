# jklz-parse-cli

金科览智文档解析命令行工具，基于 Go 语言实现。

## 特性

- 🚀 单二进制文件，无需依赖
- 📄 支持 PDF、Word、Excel、PPT 等多种格式
- 🔍 提取文本、表格、目录等结构化信息
- 🌐 零配置使用免费 API
- ⚡ 高性能、跨平台

## 快速开始

### 安装

**一键安装（推荐）**

```bash
curl -fsSL https://raw.githubusercontent.com/naihetxi/jklz-parse-skill/main/cli/install.sh | bash
```

**从源码构建**

```bash
git clone https://github.com/naihetxi/jklz-parse-skill.git
cd jklz-parse-skill/cli
./build.sh
```

### 配置

```bash
# 配置 API Key
jklz-parse config --api-key YOUR_API_KEY

# 查看配置
jklz-parse config --show

# 健康检查
jklz-parse health
```

### 使用

```bash
# 解析 PDF 为 Markdown
jklz-parse parse document.pdf

# 提取表格
jklz-parse parse data.xlsx --return table --table-format markdown

# 保存到文件
jklz-parse parse report.pdf --output result.md

# 选择页面范围
jklz-parse parse large.pdf --page-range "1-5,10"

# 完整解析（文本+目录+表格）
jklz-parse parse document.pdf --return content#toc#table
```

## 命令参考

### parse - 解析文档

```bash
jklz-parse parse <file> [flags]
```

**参数：**
- `--return` - 返回类型：content/html/toc/table/slice (可用#分隔)
- `--image-mode` - 图像解析模式：vl(高精度) 或 cv(高性能)
- `--page-range` - 页面范围，如 "1-5,10"
- `--output, -o` - 输出文件路径
- `--table-format` - 表格格式：html 或 markdown
- `--api-key` - API Key（覆盖配置）
- `--base-url` - Base URL（覆盖配置）

### config - 配置管理

```bash
jklz-parse config [flags]
```

**参数：**
- `--api-key` - 设置 API Key
- `--base-url` - 设置 Base URL
- `--show` - 显示当前配置

### health - 健康检查

```bash
jklz-parse health
```

## 构建

### 构建当前平台

```bash
./build.sh
```

### 构建全平台

```bash
./build.sh all
```

产物位于 `dist/` 目录：

| 平台 | 文件 |
|------|------|
| Linux x86_64 | `jklz-parse-linux-amd64` |
| Linux ARM64 | `jklz-parse-linux-arm64` |
| macOS Intel | `jklz-parse-darwin-amd64` |
| macOS Apple Silicon | `jklz-parse-darwin-arm64` |
| Windows x86_64 | `jklz-parse-windows-amd64.exe` |
| Windows ARM64 | `jklz-parse-windows-arm64.exe` |

## 配置文件

配置保存在 `~/.config/jklz-parse/config.yaml`：

```yaml
api_key: YOUR_API_KEY
base_url: http://192.168.42.15:15216
```

也可以使用环境变量：

```bash
export JKLZ_PARSE_APIKEY="YOUR_API_KEY"
export JKLZ_PARSE_BASEURL="http://192.168.42.15:15216"
```

## License

MIT
