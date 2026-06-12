#!/usr/bin/env python3
"""
金科览智文档解析 CLI 工具 (Python 版本)
支持 PDF、Word、Excel、PPT 等格式的解析
"""
import sys
import json
import os
import argparse
import io
import zipfile
from pathlib import Path

try:
    import requests
except ImportError:
    print("错误: 需要安装 requests 库", file=sys.stderr)
    print("运行: pip3 install requests", file=sys.stderr)
    sys.exit(1)


CONFIG_DIR = Path.home() / ".config" / "jklz-parse"
CONFIG_FILE = CONFIG_DIR / "config.json"
DEFAULT_BASE_URL = "http://192.168.42.15:15216"


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


def mask_secret(value):
    """Mask secrets in user-facing output."""
    if not value:
        return "未配置"
    if len(value) <= 10:
        return value[:3] + "..."
    return value[:10] + "..."


def resolve_api_key(args=None, config=None):
    config = config if config is not None else load_config()
    arg_value = getattr(args, "api_key", None) if args else None
    return arg_value or os.getenv("JKLZ_PARSE_APIKEY") or config.get("api_key")


def resolve_base_url(args=None, config=None):
    config = config if config is not None else load_config()
    arg_value = getattr(args, "base_url", None) if args else None
    return arg_value or os.getenv("JKLZ_PARSE_BASEURL") or config.get("base_url") or DEFAULT_BASE_URL


def parse_json_stream(response, on_json_object):
    """
    Statefully parse a stream of concatenated JSON objects.
    Handles streams where multiple JSON objects appear concatenated
    on a single line, and strips any non-JSON prefixes like 'data: '.
    """
    json_buf = []
    in_string = False
    is_escaped = False
    depth = 0
    started = False

    for chunk in response.iter_content(chunk_size=4096, decode_unicode=True):
        if not chunk:
            continue
        for ch in chunk:
            if not started:
                if ch == '{':
                    started = True
                    depth = 1
                    json_buf.append(ch)
                continue

            json_buf.append(ch)

            if is_escaped:
                is_escaped = False
                continue

            if ch == '\\':
                is_escaped = True
                continue

            if ch == '"':
                in_string = not in_string
                continue

            if not in_string:
                if ch == '{':
                    depth += 1
                elif ch == '}':
                    depth -= 1
                    if depth == 0:
                        json_str = "".join(json_buf)
                        try:
                            obj = json.loads(json_str)
                            should_stop = on_json_object(obj)
                            if should_stop:
                                return
                        except json.JSONDecodeError as e:
                            print(f"[Debug] 跳过无效 JSON: {e}", file=sys.stderr)
                        json_buf = []
                        started = False


def parse_file(file_path, args):
    """解析文档"""
    config = load_config()

    api_key = resolve_api_key(args, config)
    base_url = resolve_base_url(args, config)

    if not api_key:
        print("错误: 未配置 API Key", file=sys.stderr)
        print("请运行: jklz-parse.py config --api-key YOUR_KEY", file=sys.stderr)
        sys.exit(1)

    if not os.path.exists(file_path):
        print(f"错误: 文件不存在: {file_path}", file=sys.stderr)
        sys.exit(1)

    url = f"{base_url}/service/document/parse/stream/v2"

    data = {
        "apiKey": api_key,
        "streamType": "lz",
        "return": args.return_type,
        "imageParseMode": args.image_mode,
    }

    if args.page_range:
        data["pageSelect"] = args.page_range

    print(f"正在解析 {os.path.basename(file_path)}...", file=sys.stderr)

    max_retries = 2
    response = None
    for attempt in range(1, max_retries + 1):
        try:
            with open(file_path, "rb") as file_obj:
                files = {"file": (os.path.basename(file_path), file_obj)}
                response = requests.post(url, files=files, data=data, stream=True, timeout=300)
            response.raise_for_status()
            break
        except requests.RequestException as e:
            status = getattr(getattr(e, "response", None), "status_code", None)
            if status in [502, 503] and attempt < max_retries:
                print(f"服务暂时不可用或触发限流 (502/503)，等待 5 秒后自动重试...", file=sys.stderr)
                import time
                time.sleep(5)
                response = None
                continue
            else:
                print(f"请求失败: {e}", file=sys.stderr)
                sys.exit(1)

    result = {}

    def handle_json_object(obj):
        nonlocal result
        if obj.get("code") == "200":
            data_obj = obj.get("data", {})
            data_type = data_obj.get("type")

            if data_type in ["parseReturn", "parse_return"]:
                result = data_obj.get("value", {})
            elif data_type in ["error", "fatal"]:
                error_msg = data_obj.get("value", {}).get("error", "未知错误")
                print(f"解析错误: {error_msg}", file=sys.stderr)
                sys.exit(1)
            elif data_type == "stop":
                return True  # signal to stop parsing
        return False  # continue parsing

    parse_json_stream(response, handle_json_object)

    if not result:
        print("未获取到解析结果", file=sys.stderr)
        sys.exit(1)

    # 输出结果
    output_path = args.output
    file_type = None
    if output_path:
        ext = os.path.splitext(output_path)[1].lower()
        if ext == ".docx":
            file_type = "docx"
        elif ext == ".xlsx":
            file_type = "xlsx"
        elif ext == ".html":
            file_type = "html"
        elif ext == ".md":
            file_type = "md"

    if file_type:
        print(f"检测到输出格式 {file_type}，正在请求服务导出格式化文件...", file=sys.stderr)
        export_and_save(base_url, result, file_type, output_path)
    else:
        output_content = format_output(result, args.return_type)
        if args.output:
            with open(args.output, "w", encoding="utf-8") as f:
                f.write(output_content)
            print(f"✓ 已保存到 {args.output}", file=sys.stderr)
        else:
            print(output_content)


def safe_zip_member_path(output_dir, name):
    normalized = os.path.normpath(name.replace("\\", "/"))
    if normalized in ["", "."] or normalized.startswith("../") or os.path.isabs(normalized):
        return None
    return os.path.join(output_dir, normalized)


def should_skip_zip_member(name):
    return (
        name.endswith("/")
        or name.startswith("__MACOSX/")
        or os.path.basename(name).startswith(".")
    )


def prepare_export_content(content, output_path, download_url, content_type=""):
    ext = os.path.splitext(output_path)[1].lower()
    url_path = download_url.split("?", 1)[0].lower()
    should_try_zip = (
        url_path.endswith(".zip")
        or ext in [".md", ".html"]
        or ("zip" in content_type.lower() and ext not in [".docx", ".xlsx"])
    )
    if not should_try_zip:
        return content

    try:
        with zipfile.ZipFile(io.BytesIO(content)) as archive:
            names = [name for name in archive.namelist() if not should_skip_zip_member(name)]
            if not names:
                return content
            preferred = [name for name in names if os.path.splitext(name)[1].lower() == ext]
            selected = preferred[0] if preferred else names[0]
            output_dir = os.path.dirname(output_path) or "."

            for name in names:
                if name == selected:
                    continue
                target_path = safe_zip_member_path(output_dir, name)
                if not target_path:
                    continue
                os.makedirs(os.path.dirname(target_path), exist_ok=True)
                with archive.open(name) as src, open(target_path, "wb") as dst:
                    dst.write(src.read())

            return archive.read(selected)
    except zipfile.BadZipFile:
        return content


def export_and_save(base_url, result, file_type, output_path):
    user_id = result.get("userId") or result.get("user_id")
    job_id = result.get("jobId") or result.get("job_id")
    file_id = result.get("fileId") or result.get("file_id")

    if not user_id or not job_id or not file_id:
        print(f"错误: 无法进行导出，未在结果中找到有效的 userId/jobId/fileId (userId: {user_id}, jobId: {job_id}, fileId: {file_id})", file=sys.stderr)
        sys.exit(1)

    export_result(base_url, user_id, job_id, file_id, file_type, output_path)


def export_result(base_url, user_id, job_id, file_id, file_type, output_path, chunk_id=None, chunk_type=None):
    url = f"{base_url}/service/document/export/v2"
    payload = {
        "userId": user_id,
        "jobId": job_id,
        "fileId": file_id,
        "fileType": file_type
    }
    if chunk_id:
        payload["chunkId"] = chunk_id
    if chunk_type:
        payload["chunkType"] = chunk_type

    try:
        response = requests.post(url, json=payload, timeout=60)
        response.raise_for_status()
        res_json = response.json()
    except Exception as e:
        print(f"错误: 请求导出接口失败: {e}", file=sys.stderr)
        sys.exit(1)

    code = str(res_json.get("code"))
    if code != "200":
        print(f"错误: 导出失败 (code={code}): {res_json.get('message')}", file=sys.stderr)
        sys.exit(1)

    download_url = res_json.get("data", {}).get("url")
    if not download_url:
        print("错误: 导出接口未返回下载链接", file=sys.stderr)
        sys.exit(1)

    print(f"正在下载文件: {download_url}", file=sys.stderr)
    try:
        file_response = requests.get(download_url, timeout=60)
        file_response.raise_for_status()

        out_dir = os.path.dirname(output_path)
        if out_dir:
            os.makedirs(out_dir, exist_ok=True)

        content = prepare_export_content(
            file_response.content,
            output_path,
            download_url,
            file_response.headers.get("Content-Type", ""),
        )
        with open(output_path, "wb") as f:
            f.write(content)
        print(f"✓ 已导出并保存到 {output_path}", file=sys.stderr)
    except Exception as e:
        print(f"错误: 下载/保存导出文件失败: {e}", file=sys.stderr)
        sys.exit(1)


def format_output(result, return_type):
    """格式化输出"""
    types = return_type.split("#")
    outputs = []

    for t in types:
        if t == "content" and "content" in result:
            outputs.append(result["content"])
        elif t == "html" and "html" in result:
            outputs.append(result["html"])
        elif t in ["toc", "table", "slice", "chunks", "page", "uloc", "file", "properties"] and t in result:
            outputs.append(json.dumps(result[t], indent=2, ensure_ascii=False))

    return "\n\n".join(outputs) if outputs else json.dumps(result, indent=2, ensure_ascii=False)


def config_command(args):
    """配置命令"""
    config = load_config()

    if args.show:
        print("当前配置:")
        api_key = resolve_api_key(args, config)
        base_url = resolve_base_url(args, config)
        print(f"  API Key: {mask_secret(api_key)}")
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
    base_url = resolve_base_url(args, config)

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

def call_json_api(endpoint, payload, args):
    config = load_config()
    base_url = resolve_base_url(args, config)
    url = f"{base_url}{endpoint}"

    try:
        response = requests.post(url, json=payload, timeout=30)
        response.raise_for_status()
        try:
            result = response.json()
            print(json.dumps(result, indent=2, ensure_ascii=False))
        except ValueError:
            print(response.text)
    except Exception as e:
        print(f"请求失败: {e}", file=sys.stderr)
        sys.exit(1)


def get_command(args):
    return_types = args.return_types.split(",") if hasattr(args, "return_types") and args.return_types else ["content"]
    payload = {"userId": args.userId, "jobId": args.jobId, "fileId": args.fileId, "returnTypeList": return_types}
    call_json_api("/service/document/parse/get/v2", payload, args)


def history_command(args):
    payload = {"userId": args.userId}
    call_json_api("/service/document/parse/history/v2", payload, args)


def cancel_command(args):
    payload = {"userId": args.userId, "jobId": args.jobId}
    call_json_api("/service/document/parse/cancel/v2", payload, args)


def modify_command(args):
    payload = {"userId": args.userId, "jobId": args.jobId, "fileId": args.fileId, "chunkId": args.chunk_id, "text": args.text}
    call_json_api("/service/document/parse/modify/v2", payload, args)


def cleanup_command(args):
    payload = {"user_id": args.userId, "time": args.time}
    call_json_api("/service/document/parse/cleanup/v2", payload, args)


def search_command(args):
    if not args.keywords:
        print("错误: 请提供搜索关键词", file=sys.stderr)
        sys.exit(1)
    keywords = [k.strip() for k in args.keywords.split(",")] if "," in args.keywords else args.keywords.split(" ")
    payload = {"userId": args.userId, "jobId": args.jobId, "fileId": args.fileId, "keywords": keywords}
    call_json_api("/service/document/search/v2", payload, args)


def export_command(args):
    config = load_config()
    base_url = resolve_base_url(args, config)
    output_path = args.output or f"{args.fileId}.{args.file_type}"
    export_result(
        base_url,
        args.userId,
        args.jobId,
        args.fileId,
        args.file_type,
        output_path,
        chunk_id=args.chunk_id,
        chunk_type=args.chunk_type,
    )



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
  %(prog)s export userId jobId fileId --type md -o result.md
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
    parse_parser.add_argument("-o", "--output", help="输出文件路径（支持 .md/.html/.docx/.xlsx）")
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
    # get 命令
    get_parser = subparsers.add_parser("get", help="获取指定解析任务的结果")
    get_parser.add_argument("userId", help="用户ID")
    get_parser.add_argument("jobId", help="作业ID")
    get_parser.add_argument("fileId", help="文件ID")
    get_parser.add_argument("-r", "--return-types", default="content", help="返回类型列表，逗号分隔 (如 content,html,toc)")
    get_parser.add_argument("--base-url", help="Base URL（覆盖配置）")

    # history 命令
    history_parser = subparsers.add_parser("history", help="查询历史解析记录")
    history_parser.add_argument("userId", help="用户ID")
    history_parser.add_argument("--base-url", help="Base URL（覆盖配置）")

    # cancel 命令
    cancel_parser = subparsers.add_parser("cancel", help="停止正在运行的解析任务")
    cancel_parser.add_argument("userId", help="用户ID")
    cancel_parser.add_argument("jobId", help="作业ID")
    cancel_parser.add_argument("--base-url", help="Base URL（覆盖配置）")

    # modify 命令
    modify_parser = subparsers.add_parser("modify", help="修改解析后的Chunk内容")
    modify_parser.add_argument("userId", help="用户ID")
    modify_parser.add_argument("jobId", help="作业ID")
    modify_parser.add_argument("fileId", help="文件ID")
    modify_parser.add_argument("-c", "--chunk-id", required=True, help="Chunk ID")
    modify_parser.add_argument("-t", "--text", required=True, help="修改后的文本内容")
    modify_parser.add_argument("--base-url", help="Base URL（覆盖配置）")

    # cleanup 命令
    cleanup_parser = subparsers.add_parser("cleanup", help="清理历史解析文件")
    cleanup_parser.add_argument("userId", help="用户ID")
    cleanup_parser.add_argument("time", help="时间 (如: 7d, 24h)")
    cleanup_parser.add_argument("--base-url", help="Base URL（覆盖配置）")

    # search 命令
    search_parser = subparsers.add_parser("search", help="在解析结果中搜索关键词")
    search_parser.add_argument("userId", help="用户ID")
    search_parser.add_argument("jobId", help="作业ID")
    search_parser.add_argument("fileId", help="文件ID")
    search_parser.add_argument("-k", "--keywords", required=True, help="搜索关键词列表，用逗号分隔")
    search_parser.add_argument("--base-url", help="Base URL（覆盖配置）")

    # export 命令
    export_parser = subparsers.add_parser("export", help="导出已解析结果为文件")
    export_parser.add_argument("userId", help="用户ID")
    export_parser.add_argument("jobId", help="作业ID")
    export_parser.add_argument("fileId", help="文件ID")
    export_parser.add_argument("-t", "--type", dest="file_type", default="md",
                               choices=["md", "html", "docx", "xlsx"],
                               help="导出文件类型: md/html/docx/xlsx")
    export_parser.add_argument("-o", "--output", help="输出文件路径，默认 fileId.type")
    export_parser.add_argument("--chunk-id", help="只导出指定 chunkId")
    export_parser.add_argument("--chunk-type", help="只导出指定 chunkType")
    export_parser.add_argument("--base-url", help="Base URL（覆盖配置）")


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
    elif args.command == "get":
        get_command(args)
    elif args.command == "history":
        history_command(args)
    elif args.command == "cancel":
        cancel_command(args)
    elif args.command == "modify":
        modify_command(args)
    elif args.command == "cleanup":
        cleanup_command(args)
    elif args.command == "search":
        search_command(args)
    elif args.command == "export":
        export_command(args)


if __name__ == "__main__":
    main()
