#!/usr/bin/env python3
"""
金科览智文档解析 CLI 工具 (Python 版本)
支持 PDF、Word、Excel、PPT 等格式的解析
"""
import sys
import json
import os
import argparse
from pathlib import Path

try:
    import requests
except ImportError:
    print("错误: 需要安装 requests 库", file=sys.stderr)
    print("运行: pip3 install requests", file=sys.stderr)
    sys.exit(1)


CONFIG_DIR = Path.home() / ".config" / "jklz-parse"
CONFIG_FILE = CONFIG_DIR / "config.json"


def load_config():
    """加载配置"""
    if CONFIG_FILE.exists():
        with open(CONFIG_FILE) as f:
            return json.load(f)
    return {}


def save_config(config):
    """保存配置"""
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    with open(CONFIG_FILE, "w") as f:
        json.dump(config, f, indent=2)


def parse_file(file_path, args):
    """解析文档"""
    config = load_config()

    api_key = args.api_key or config.get("api_key")
    base_url = args.base_url or config.get("base_url", "http://192.168.42.15:15216")

    if not api_key:
        print("错误: 未配置 API Key", file=sys.stderr)
        print("请运行: jklz-parse.py config --api-key YOUR_KEY", file=sys.stderr)
        sys.exit(1)

    if not os.path.exists(file_path):
        print(f"错误: 文件不存在: {file_path}", file=sys.stderr)
        sys.exit(1)

    url = f"{base_url}/service/document/parse/stream/v1"

    files = {"file": open(file_path, "rb")}
    data = {
        "api_key": api_key,
        "stream_type": "lz",
        "return": args.return_type,
        "image_parse_mode": args.image_mode,
    }

    if args.page_range:
        data["page_selecte2parse"] = args.page_range

    print(f"正在解析 {os.path.basename(file_path)}...", file=sys.stderr)

    try:
        response = requests.post(url, files=files, data=data, stream=True, timeout=300)
        response.raise_for_status()
    except requests.RequestException as e:
        print(f"请求失败: {e}", file=sys.stderr)
        sys.exit(1)

    result = None
    for line in response.iter_lines():
        if not line:
            continue

        try:
            data = json.loads(line.decode('utf-8'))
            if data.get("code") == "200":
                data_obj = data.get("data", {})
                data_type = data_obj.get("type")

                if data_type == "parse_return":
                    result = data_obj.get("value", )
                elif data_type in ["error", "fatal"]:
                    error_msg = data_obj.get("value", {}).get("error", "未知错误")
                    print(f"解析错误: {error_msg}", file=sys.stderr)
                    sys.exit(1)
                elif data_type == "stop":
                    break
        except json.JSONDecodeError:
            continue

    if not result:
        print("未获取到解析结果", file=sys.stderr)
        sys.exit(1)

    # 输出结果
    output_content = format_output(result, args.return_type)

    if args.output:
        with open(args.output, "w", encoding="utf-8") as f:
            f.write(output_content)
        print(f"✓ 已保存到 {args.output}", file=sys.stderr)
    else:
        print(output_content)


def format_output(result, return_type):
    """格式化输出"""
    types = return_type.split("#")
    outputs = []

    for t in types:
        if t == "content" and "content" in result:
            outputs.append(result["content"])
        elif t == "html" and "html" in result:
            outputs.append(result["html"])
        elif t in ["toc", "table", "slice"] and t in result:
            outputs.append(json.dumps(result[t], indent=2, ensure_ascii=False))

    return "\n\n".join(outputs) if outputs else json.dumps(result, indent=2, ensure_ascii=False)


def config_command(args):
    """配置命令"""
    config = load_config()

    if args.show:
        print("当前配置:")
        api_key = config.get("api_key", "")
        if api_key:
            print(f"  API Key: {api_key[:10]}...")
        else:
            print("  API Key: 未配置")

        base_url = config.get("base_url", "http://192.168.42.15:15216")
        print(f"  Base URL: {base_url}")
        print(f"\n配置文件: {CONFIG_FILE}")
        return

    if args.api_key:
        config["api_key"] = args.api_key
        print("✓ API Key 已保存")

    if args.base_url:
        config["base_url"] = args.base_url
        print("✓ Base URL 已保存")

    if args.api_key or args.base_url:
        save_config(config)


def health_command(args):
    """健康检查"""
    config = load_config()
    base_url = args.base_url or config.get("base_url", "http://192.168.42.15:15216")

    try:
        response = requests.get(f"{base_url}/metrics", timeout=5)
        response.raise_for_status()
        data = response.json()

        if data.get("status") == "success":
            print(f"✓ 服务正常: {data.get('message')}")
            print(f"  Base URL: {base_url}")
        else:
            print("✗ 服务异常", file=sys.stderr)
            sys.exit(1)
    except Exception as e:
        print(f"✗ 连接失败: {e}", file=sys.stderr)
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(
        description="金科览智文档解析 CLI 工具",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  %(prog)s parse document.pdf
  %(prog)s parse report.pdf --page-range "1-5"
  %(prog)s parse data.xlsx --return table --output result.json
  %(prog)s parse doc.pdf --return content#toc#table
  %(prog)s config --api-key YOUR_KEY
  %(prog)s health
        """
    )

    subparsers = parser.add_subparsers(dest="command", help="子命令")

    # parse 命令
    parse_parser = subparsers.add_parser("parse", help="解析文档")
    parse_parser.add_argument("file", help="要解析的文件路径")
    parse_parser.add_argument("--return", dest="return_type", default="content",
                            help="返回类型: content/html/toc/table/slice (可用#分隔)")
    parse_parser.add_argument("--image-mode", dest="image_mode", default="cv",
                            choices=["vl", "cv"],
                            help="图像解析模式: vl(高精度) 或 cv(高性能，默认)")
    parse_parser.add_argument("--page-range", help="页面范围，如 '1-5,10'")
    parse_parser.add_argument("-o", "--output", help="输出文件路径")
    parse_parser.add_argument("--api-key", help="API Key（覆盖配置）")
    parse_parser.add_argument("--base-url", help="Base URL（覆盖配置）")

    # config 命令
    config_parser = subparsers.add_parser("config", help="配置管理")
    config_parser.add_argument("--api-key", help="设置 API Key")
    config_parser.add_argument("--base-url", help="设置 Base URL")
    config_parser.add_argument("--show", action="store_true", help="显示当前配置")

    # health 命令
    health_parser = subparsers.add_parser("health", help="健康检查")
    health_parser.add_argument("--base-url", help="Base URL（覆盖配置）")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    if args.command == "parse":
        parse_file(args.file, args)
    elif args.command == "config":
        config_command(args)
    elif args.command == "health":
        health_command(args)


if __name__ == "__main__":
    main()
