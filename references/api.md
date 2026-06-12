# Parse API 参考文档

本文档按当前推荐的 v2 接口编写。v1 与 v2 的核心逻辑一致，主要差异是字段命名：v1 使用 snake_case，v2 使用 camelCase。新接入方优先使用 v2。

## 服务信息

- Base URL: `http://192.168.42.15:15216`
- API Key: 仅文档解析上传接口必填，字段名为 `apiKey`
- 默认图像解析模式: `cv`

## 1. 流式文档解析

`POST /service/document/parse/stream/v2`

请求格式: `multipart/form-data`

| 参数 | 必填 | 说明 |
|---|---|---|
| `file` | 是 | 要解析的文档文件，支持 PDF、DOC、DOCX、XLSX、PPT、图片等 |
| `apiKey` | 是 | API 密钥 |
| `userId` | 否 | 用户 ID，默认 `jklz` |
| `jobId` | 否 | 解析任务 ID，默认服务端生成 |
| `streamType` | 否 | `lz` 或 `sse`，默认 `lz` |
| `return` | 否 | 返回类型，多个用 `#` 连接 |
| `imageParseMode` | 否 | `cv` 高性能、`vl` 高精度、`auto` 自动 |
| `pageSelect` | 否 | PDF 页码选择，第一页为 `0`，如 `0,3-5,-1` |
| `splitNestedTable` | 否 | Word 嵌套表格拆分，`1` 开启 |
| `trace` | 否 | 溯源，`1` 开启 |
| `filterHfSupport` | 否 | 页眉页脚过滤，`1` 开启 |
| `forceMergeCrossPageTable` | 否 | 强制跨页表格合并，默认 `1` |
| `crossPageTableMergeSupport` | 否 | 跨页表格合并，默认 `1` |
| `splitType` | 否 | 切片方式：`toc`、`length`、`custom` |
| `splitCustomSeparators` | 否 | 自定义切片正则，多个用 `#` 连接 |
| `splitMaxLength` | 否 | 切片最大长度，默认 `512` |
| `overlap` | 否 | 是否使用切片重叠 |
| `overlapSize` | 否 | 重叠长度 |

示例:

```bash
curl --location 'http://192.168.42.15:15216/service/document/parse/stream/v2' \
  --form 'file=@"./test.docx"' \
  --form 'apiKey="YOUR_API_KEY"' \
  --form 'streamType="lz"' \
  --form 'return="content#toc#table"' \
  --form 'imageParseMode="cv"'
```

流式响应中的关键消息:

| `data.type` | 说明 |
|---|---|
| `agent` | 解析进度消息 |
| `parseReturn` | 解析结果，`data.value` 中包含 `userId`、`jobId`、`fileId` 和请求的返回字段 |
| `stop` | 解析结束 |

## 2. 非流式文档解析

`POST /service/document/parse/v2`

参数与流式接口基本一致，响应一次性返回 JSON。适合小文件或不需要进度流的场景。CLI 默认使用流式接口，智能体直接调用 API 时也优先使用流式接口。

## 3. 获取解析结果

`POST /service/document/parse/get/v2`

```json
{
  "userId": "jklz",
  "jobId": "qYFfQUwqJGEnovWTvxhJchT7",
  "fileId": "aaf058011aacba758122ba3b6c64f4a9",
  "returnTypeList": ["content", "toc", "html"]
}
```

`returnTypeList` 支持: `content`、`table`、`html`、`toc`、`slice`、`chunks`、`page`、`uloc`、`pkl`、`file`、`properties`。

## 4. 导出解析结果

`POST /service/document/export/v2`

用于把已解析结果导出为带样式/结构的文件，并返回下载链接。

```json
{
  "userId": "jklz",
  "jobId": "qYFfQUwqJGEnovWTvxhJchT7",
  "fileId": "aaf058011aacba758122ba3b6c64f4a9",
  "fileType": "html"
}
```

可选字段:

| 字段 | 说明 |
|---|---|
| `chunkId` | 只导出指定 chunk |
| `chunkType` | 当未指定 `chunkId` 时，按 chunk 类型导出 |

`fileType` 支持: `html`、`md`、`docx`、`xlsx`。

## 5. 关键词搜索

`POST /service/document/search/v2`

```json
{
  "userId": "jklz",
  "jobId": "qYFfQUwqJGEnovWTvxhJchT7",
  "fileId": "aaf058011aacba758122ba3b6c64f4a9",
  "keywords": ["合同", "协议", "条款"]
}
```

返回按关键词分组的命中位置、上下文和索引信息。

## 6. 修改 Chunk

`POST /service/document/parse/modify/v2`

```json
{
  "userId": "jklz",
  "jobId": "qYFfQUwqJGEnovWTvxhJchT7",
  "fileId": "aaf058011aacba758122ba3b6c64f4a9",
  "chunkId": "0",
  "text": "这是修改后的 chunk 内容"
}
```

## 7. 取消任务

`POST /service/document/parse/cancel/v2`

```json
{
  "userId": "jklz",
  "jobId": "qYFfQUwqJGEnovWTvxhJchT7"
}
```

## 8. 查询历史

`POST /service/document/parse/history/v2`

```json
{
  "userId": "jklz"
}
```

返回按日期分组的历史任务，包含 `jobId`、`status`、`fileName`、`fileId` 等信息。

## 9. 清理历史文件

`POST /service/document/parse/cleanup/v2`

```json
{
  "user_id": "jklz",
  "time": "7d"
}
```

`time` 支持 `m` 分钟、`h` 小时、`d` 天、`w` 周。

## 10. 健康检查

`GET /metrics`

```json
{
  "status": "success",
  "message": "Service is healthy"
}
```

## v1 兼容说明

v1 端点仍可用于历史兼容，命名示例:

| v1 | v2 |
|---|---|
| `api_key` | `apiKey` |
| `user_id` | `userId` |
| `job_id` | `jobId` |
| `file_id` | `fileId` |
| `return_type_list` | `returnTypeList` |
| `image_parse_mode` | `imageParseMode` |
| `page_selecte2parse` | `pageSelect` |
