#!/usr/bin/env bash
# jklz-parse-skill - 文档智能解析 API 调用脚本
# 用法: ./call_api.sh <file_path> [return_type] [extra_form_args...]
# 示例: ./call_api.sh document.pdf content
#        ./call_api.sh document.xlsx table table_format=markdown
set -euo pipefail

FILE_PATH="${1:?用法: $0 <file_path> [return_type] [extra_form_args...]}"
RETURN_TYPE="${2:-content}"
DEFAULT_BASE_URL="http://192.168.42.15:15216"

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
URL="${URL:-$DEFAULT_BASE_URL}"
API_KEY="${JKLZ_PARSE_APIKEY:-$(config_value api_key)}"

if [ -z "$API_KEY" ]; then
  echo "错误: 未配置 API Key。请设置 JKLZ_PARSE_APIKEY 或运行 jklz-parse config --api-key YOUR_KEY" >&2
  exit 1
fi

if [ ! -f "$FILE_PATH" ]; then
  echo "错误: 文件不存在: $FILE_PATH" >&2
  exit 1
fi

TRACE_ID="parse-$(date +%s)"

# 收集额外参数
EXTRA_ARGS=()
for arg in "${@:3}"; do
  EXTRA_ARGS+=(-F "$arg")
done

curl -s -X POST "${URL}/service/document/parse/stream/v2" \
  -F "file=@${FILE_PATH}" \
  -F "apiKey=${API_KEY}" \
  -F "streamType=lz" \
  -F "return=${RETURN_TYPE}" \
  -F "imageParseMode=cv" \
  -F "traceId=${TRACE_ID}" \
  "${EXTRA_ARGS[@]}"
