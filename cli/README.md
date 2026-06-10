# jklz-parse-cli

金科览智文档解析命令行工具，提供 Python 和 Go 两种实现。

## 特性

- 🚀 **Python CLI**（推荐）：无需编译，开箱即用
- ⚡ **Go CLI**：单二进制文件，高性能
- 📄 支持 PDF、Word、Excel、PPT 等多种格式
- 🔍 提取文本、表格、目录等结构化信息
- 🌐 支持内网 API 部署
- 💾 灵活的配置管理

## 快速开始

### Python CLI（推荐）

**安装依赖**

```bash
pip3 install requests
```

**配置 API**

```bash
# 配置 API Key（联系金科览智管理员获取）
python3 jklz-parse.py config --api-key YOUR_API_KEY

# 配置 API 地址（可选，默认为内网地址）
python3 jklz-parse.py config --base-url http://192.168.42.15:15216

# 查看配置
python3 jklz-parse.py config --show

# 健康检查
python3 jklz-parse.py health
```

**使用示例**

```bash
# 解析 PDF 为 Markdown
python3 jklz-parse.py parse document.pdf

# 提取 HTML
python3 jklz-parse.py parse document.pdf --return html

# 提取表格（JSON 格式）
python3 jklz-parse.py parse data.xlsx --return table

# 保存到文件
python3 jklz-parse.py parse report.pdf --output result.md

# 选择页面范围
python3 jklz-parse.py parse large.pdf --page-range "1-5,10"

# 完整解析（文本+目录+表格）
python3 jklz-parse.py parse document.pdf --return content#toc#table

# 高性能模式
python3 jklz-parse.py parse large.pdf --image-mode cv
```

### Go CLI

**使用预编译版本**（推荐）：

项目中已包含编译好的 Go CLI 二进制文件，可直接使用：

```bash
# 直接运行预编译版本
./jklz-parse config --api-key YOUR_API_KEY
./jklz-parse health
./jklz-parse parse document.pdf
```

**从源码编译**（需要 Go 1.21+）：

```bash
go build -o jklz-parse main.go
```

**注意**：
- 预编译版本已经过完整测试，推荐直接使用
- 如需重新编译，确保 Go 版本 >= 1.21
- 编译遇到问题可直接使用 Python CLI，功能完全相同

## 命令参考

### parse - 解析文档

```bash
python3 jklz-parse.py parse <file> [flags]
# 或
./jklz-parse parse <file> [flags]
```

**参数：**
- `--return` - 返回类型：content/html/toc/table/slice (可用#分隔组合)
- `--image-mode` - 图像解析模式：
  - `vl` (高精度，默认) - 适合表格密集、复杂版式
  - `cv` (高性能) - 速度更快，适合大文件
- `--page-range` - 页面范围，如 "1-5,10"
- `--output, -o` - 输出文件路径
- `--api-key` - API Key（覆盖配置）
- `--base-url` - Base URL（覆盖配置）

**返回类型说明：**
- `content` - Markdown 文本（默认）
- `html` - HTML 格式
- `toc` - 目录结构（JSON）
- `table` - 表格数据（JSON）
- `slice` - 文档切片，用于 RAG（JSON）

### config - 配置管理

```bash
python3 jklz-parse.py config [flags]
```

**参数：**
- `--api-key` - 设置 API Key
- `--base-url` - 设置 Base URL
- `--show` - 显示当前配置

### health - 健康检查

```bash
python3 jklz-parse.py health
```

检查 API 服务是否可用。

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

配置保存在 `~/.config/jklz-parse/config.json`：

```json
{
  "api_key": "YOUR_API_KEY",
  "base_url": "http://192.168.42.15:15216"
}
```

也可以使用环境变量（会覆盖配置文件）：

```bash
export JKLZ_PARSE_APIKEY="YOUR_API_KEY"
export JKLZ_PARSE_BASEURL="http://192.168.42.15:15216"
```

## API Key 申请

API Key 需要向金科览智服务管理员申请。申请时请提供：
- 使用场景说明
- 预计调用量
- 使用期限

默认 API 地址为内网地址：`http://192.168.42.15:15216`

如使用其他环境，请通过 `--base-url` 或配置文件修改。

## License

MIT
