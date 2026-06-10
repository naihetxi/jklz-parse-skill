# Parse API 参考文档

本文档包含 Parse API 的详细接口说明。

## 服务信息

- **Base URL**: `http://192.168.42.15:15216`
- **认证方式**: API Key (form data)

## 接口列表

### 1. 流式文档解析

**端点**: `POST /service/document/parse/stream/v1`

**请求格式**: `multipart/form-data`

**参数**:

| 参数 | 类型 | 必填 | 默认值 | 描述 |
|-----|------|-----|-------|------|
| file | File | 是 | - | 要解析的文档文件 |
| api_key | String | 是 | - | API 密钥 |
| user_id | String | 否 | "jklz" | 用户 ID |
| job_id | String | 否 | 随机24位 | 作业 ID |
| stream_type | String | 否 | "lz" | "lz"=自定义流式, "sse"=Server-Sent Events |
| return | String | 否 | "content" | 返回类型：content, table, toc, slice, html, chunks, page, uloc |
| image_parse_mode | String | 否 | "cv" | "cv"=高性能，"vl"=高精度 |
| split_nested_table | Integer | 否 | 0 | 是否拆分嵌套表格（Word） |
| trace | Integer | 否 | 0 | 是否启用溯源 |
| page_selecte2parse | String | 否 | "" | 选择页码（PDF） |
| table_format | String | 否 | "html" | "html" 或 "markdown" |
| filter_hf_support | Integer | 否 | 0 | 是否过滤页眉页脚 |
| cross_page_table_merge_support | Integer | 否 | 1 | 跨页表格合并 |
| split_type | String | 否 | "toc" | toc, length, custom |
| split_custom_separators | String | 否 | "" | 自定义切片分隔符 |
| split_max_length | Integer | 否 | 512 | 切片最大长度 |
| overlap | Boolean | 否 | false | 是否使用重叠 |
| overlap_size | Integer | 否 | 128 | 重叠长度 |
| save_table | Boolean | 否 | true | 是否保存表格 |

### 2. 获取解析结果

**端点**: `POST /service/document/parse/get/v1`

**请求体**:

```json
{
  "user_id": "jklz",
  "job_id": "xxx",
  "file_id": "xxx",
  "return_type_list": ["content", "toc", "html"]
}
```

### 3. 停止解析任务

**端点**: `POST /service/document/parse/cancel/v1`

**请求体**:

```json
{
  "user_id": "jklz",
  "job_id": "xxx"
}
```

### 4. 查询历史记录

**端点**: `POST /service/document/parse/history/v1`

**请求体**:

```json
{
  "user_id": "jklz"
}
```

### 5. 清理历史文件

**端点**: `POST /service/document/parse/cleanup/v1`

**请求体**:

```json
{
  "user_id": "jklz",
  "time": "7d"
}
```

时间格式: `数字 + 单位`，如 `12d` (天), `25h` (小时), `3w` (周), `20m` (分钟)

### 6. 健康检查

**端点**: `GET /metrics`

**响应**:

```json
{
  "status": "success",
  "message": "Service is healthy"
}
```

## 支持的文件格式

- PDF (`.pdf`)
- Word (`.doc`, `.docx`)
- Excel (`.xlsx`)
- PowerPoint (`.ppt`, `.pptx`)
- 纯文本 (`.txt`)
- 图片 (`.png`, `.jpg`, `.jpeg`)
