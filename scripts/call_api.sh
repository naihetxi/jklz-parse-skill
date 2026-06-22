#!/usr/bin/env bash
# jklz-parse-skill - 文档智能解析 API 调用脚本
# 用法: ./call_api.sh <file_path> [return_type] [extra_form_args...]
# 示例: ./call_api.sh document.pdf content
#        ./call_api.sh document.xlsx table table_format=markdown
set -euo pipefail

FILE_PATH="${1:?用法: $0 <file_path> [return_type] [extra_form_args...]}"
RETURN_TYPE="${2:-content}"

config_value() {
  python3 -c 'import json, pathlib, sys
p = pathlib.Path.home() / ".config" / "jklz-parse" / "config.json"
try:
    data = json.loads(p.read_text())
    print(data.get(sys.argv[1], ""))
except Exception:
    print("")' "$1"
}

URL="${JKLZ_PARSE_BASEURL:-$(config_value base_url)}"
API_KEY="${JKLZ_PARSE_APIKEY:-$(config_value api_key)}"

if [ -z "$URL" ]; then
  echo "错误: 未配置 Base URL。请设置 JKLZ_PARSE_BASEURL 或运行 jklz-parse config --base-url http://YOUR_HOST:PORT" >&2
  exit 1
fi

if [ -z "$API_KEY" ]; then
  echo "错误: 未配置 API Key。请设置 JKLZ_PARSE_APIKEY 或运行 jklz-parse config --api-key YOUR_KEY" >&2
  exit 1
fi

if [ ! -f "$FILE_PATH" ]; then
  echo "错误: 文件不存在: $FILE_PATH" >&2
  exit 1
fi

# 收集额外参数
EXTRA_ARGS=()
for arg in "${@:3}"; do
  EXTRA_ARGS+=(-F "$arg")
done

if [ ${#EXTRA_ARGS[@]} -gt 0 ]; then
  curl -s -X POST "${URL}/service/document/parse/stream/v2" \
    -F "file=@${FILE_PATH}" \
    -F "apiKey=${API_KEY}" \
    -F "streamType=lz" \
    -F "return=${RETURN_TYPE}" \
    -F "imageParseMode=cv" \
    "${EXTRA_ARGS[@]}"
else
  curl -s -X POST "${URL}/service/document/parse/stream/v2" \
    -F "file=@${FILE_PATH}" \
    -F "apiKey=${API_KEY}" \
    -F "streamType=lz" \
    -F "return=${RETURN_TYPE}" \
    -F "imageParseMode=cv"
fi
