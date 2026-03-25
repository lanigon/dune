# Dune API 参考

Base URL: `https://api.dune.com/api/v1/`

认证: `X-Dune-Api-Key` header 或 `api_key` query parameter

---

## 1. 查询执行与结果

### Execute Query

```
POST /query/{query_id}/execute
```

触发已保存查询的执行。

**参数：**
- `query_id` (path, integer, required)
- `performance` (body, string, optional): `medium` | `large`
- `query_parameters` (body, object, optional): JSON key-value 参数

**响应：**
```json
{
  "execution_id": "01HKZJ2683PHF9Q9PHHQ8FW4Q1",
  "state": "QUERY_STATE_PENDING"
}
```

### Execute SQL

```
POST /sql/execute
```

直接执行裸 SQL，无需保存查询。返回的 `query_id` 始终为 0。

**请求体：**
```json
{
  "sql": "SELECT * FROM dex.trades LIMIT 10",
  "performance": "medium"
}
```

### Get Execution Status

```
GET /execution/{execution_id}/status
```

**不消耗 credits。**

**执行状态：**

| State | 说明 |
|-------|------|
| `QUERY_STATE_PENDING` | 等待执行槽位 |
| `QUERY_STATE_EXECUTING` | 正在执行 |
| `QUERY_STATE_COMPLETED` | 成功完成 |
| `QUERY_STATE_FAILED` | 失败（含错误详情） |
| `QUERY_STATE_CANCELED` | 用户取消 |
| `QUERY_STATE_EXPIRED` | 结果已过期 |
| `QUERY_STATE_COMPLETED_PARTIAL` | 成功但结果被截断 |

**响应示例：**
```json
{
  "execution_id": "01HKZJ2683PHF9Q9PHHQ8FW4Q1",
  "query_id": 3493826,
  "is_execution_finished": true,
  "state": "QUERY_STATE_COMPLETED",
  "submitted_at": "2025-10-22T10:31:04.222464Z",
  "execution_started_at": "2025-10-22T10:31:05.123456Z",
  "execution_ended_at": "2025-10-22T10:31:15.789012Z",
  "execution_cost_credits": 10,
  "expires_at": "2026-01-20T10:31:04.36241Z"
}
```

### Get Execution Result

```
GET /execution/{execution_id}/results
```

结果数据存储 90 天。

### Get Latest Query Result

```
GET /query/{query_id}/results
```

获取最近一次执行的缓存结果。**不触发执行，但按结果大小消耗 credits。**

查询必须是公开的或你拥有的。

### Get Latest Query Result CSV

```
GET /query/{query_id}/results/csv
```

### Cancel Execution

```
POST /execution/{execution_id}/cancel
```

---

## 2. 结果过滤、排序、分页、采样

以下参数适用于所有结果获取端点：

### 过滤 (filters)

SQL WHERE 风格的服务端过滤。

**支持的运算符：**
- 比较: `>`, `>=`, `<`, `<=`, `=`
- 逻辑: `AND`, `OR`
- 特殊: `IN`, `LIKE`, `IS NOT NULL`
- **不支持**: `NOT IN`, `NOT LIKE`, SQL 函数（如 `now()`）、子查询

**语法规则：**
- 字符串值用单引号: `project = 'uniswap'`
- 数值不加引号: `amount_usd > 1000`
- 日期作为字符串: `block_time > '2024-03-01'`
- 特殊列名用双引号: `"special, column" = 'ABC'`

**示例：**
```
?filters=blockchain = 'ethereum' AND amount_usd > 1000
?filters=project IN ('uniswap', 'sushiswap', 'curve')
?filters=token_symbol LIKE 'WETH%'
```

### 排序 (sort_by)

SQL ORDER BY 风格。

```
?sort_by=amount_usd desc, block_time
```

### 分页 (limit/offset)

```
?limit=100&offset=200
```

**JSON 响应含：** `next_offset`, `next_uri`
**CSV 响应含：** `x-dune-next-offset`, `x-dune-next-uri` headers

**注意：** offset 超出范围返回空结果（不报错），metadata 含 `total_row_count`。

### 采样 (sample_count)

均匀随机采样。**与 offset/limit/filters 互斥。**

```
?sample_count=1000
```

### 列选择 (columns)

逗号分隔列名，减少数据量和 credit 消耗。

```
?columns=tx_from,tx_to,amount_usd
```

### 其他参数

- `allow_partial_results` (boolean): 允许获取超 32GB 的截断结果
- `ignore_max_credits_per_request` (boolean): 绕过默认单请求 credit 上限

---

## 3. 结果响应结构

```json
{
  "execution_id": "string",
  "query_id": 1234,
  "state": "QUERY_STATE_COMPLETED",
  "is_execution_finished": true,
  "submitted_at": "timestamp",
  "execution_started_at": "timestamp",
  "execution_ended_at": "timestamp",
  "expires_at": "timestamp",
  "next_offset": 100,
  "next_uri": "https://api.dune.com/...",
  "result": {
    "rows": [{"col1": "val1", "col2": 123}],
    "metadata": {
      "column_names": ["col1", "col2"],
      "column_types": ["varchar", "bigint"],
      "row_count": 100,
      "total_row_count": 5000,
      "result_set_bytes": 2042,
      "total_result_set_bytes": 56496,
      "datapoint_count": 200,
      "execution_time_millis": 29352,
      "pending_time_millis": 1185
    }
  },
  "error": {
    "type": "syntax_error",
    "message": "Error: Line 1:1: ...",
    "metadata": {"line": 1, "column": 1}
  }
}
```

---

## 4. 查询管理

**需要 Analyst 及以上计划。**

### Create Query
```
POST /query
```

**请求体：**
```json
{
  "name": "My Query",
  "query_sql": "SELECT * FROM dex.trades LIMIT 10",
  "description": "optional",
  "is_private": false,
  "parameters": [
    {"key": "address", "value": "0x...", "type": "text"},
    {"key": "limit", "value": "100", "type": "number"},
    {"key": "chain", "value": "ethereum", "type": "enum", "enumOptions": ["ethereum", "polygon"]}
  ]
}
```

### Read Query
```
GET /query/{queryId}
```

### Update Query
```
PATCH /query/{queryId}
```

### List Queries
```
GET /queries?limit=20&offset=0
```

### Archive / Unarchive
```
PATCH /query/{queryId}/archive
PATCH /query/{queryId}/unarchive
```

### Private / Unprivate
```
PATCH /query/{queryId}/private
PATCH /query/{queryId}/unprivate
```

### 查询参数类型

| 类型 | SQL 用法 |
|------|---------|
| text | `WHERE address = {{my_param}}` |
| number | `WHERE amount > {{threshold}}` |
| datetime | `WHERE block_time > timestamp '{{start_time}}'` |
| enum | `WHERE chain = {{chain}}` |

---

## 5. Materialized Views

### 命名规范
- 完整名: `dune.<team>.result_<name>`
- 创建时只传 `result_<name>` 部分

### Upsert (创建/更新)
```
POST /materialized-views
```

**请求体：**
```json
{
  "name": "result_daily_volume",
  "query_id": 1234,
  "cron_expression": "0 */4 * * *",
  "performance": "medium",
  "is_private": false
}
```

cron 间隔最少 15 分钟，最多每周。

### Get / List / Delete / Refresh
```
GET    /materialized-views/{name}
GET    /materialized-views
DELETE /materialized-views/{name}
POST   /materialized-views/{name}/refresh
```

---

## 6. 数据上传

### Create Table
```
POST /uploads
```

每次创建消耗 10 credits。

**请求体：**
```json
{
  "namespace": "my_user",
  "table_name": "interest_rates",
  "description": "...",
  "is_private": false,
  "schema": [
    {"name": "date", "type": "timestamp"},
    {"name": "value", "type": "double", "nullable": true}
  ]
}
```

### Insert Data
```
POST /uploads/{namespace}/{table_name}/insert
```

- Content-Type: `text/csv` 或 `application/x-ndjson`
- 最大请求大小 1.2GB
- 原子操作（全部成功或全部失败）
- 并发写建议不超过 5-10 个

**Varbinary 格式：**
- Base64: `{"col": "SGVsbG8gd29ybGQK"}`
- Hex: `{"col": "0x92b7d1031988c7af"}`

### Upload CSV
```
POST /uploads/csv
```

- 最大 500MB
- 自动推断 schema
- **覆盖已有表**（非追加）
- 追加只支持 create + insert 方式

### Clear / Delete / List
```
POST   /uploads/{namespace}/{table_name}/clear
DELETE /uploads/{namespace}/{table_name}
GET    /uploads
```

---

## 7. Dataset 搜索

### Search Datasets
```
POST /datasets/search
```

**请求体：**
```json
{
  "query": "dex trades",
  "blockchains": ["ethereum"],
  "categories": ["spell"],
  "include_schema": true,
  "limit": 10
}
```

### Search by Contract Address
```
POST /datasets/search-by-contract
```

**请求体：**
```json
{
  "contract_address": "0x...",
  "blockchains": ["ethereum"],
  "include_schema": true
}
```

---

## 8. Pipeline

### Execute Pipeline
```
POST /query/{query_id}/pipeline/execute
```

链式执行多个查询和 MV 刷新。

### Get Pipeline Status
```
GET /pipelines/{pipeline_execution_id}/status
```

---

## 9. Webhook

在 Dune 网页端 query 编辑器中配置 schedule + webhook URL。

**Webhook payload 结构：**
```json
{
  "message": "Query ... was successfully executed",
  "query_result": {
    "execution_id": "...",
    "query_id": 3106864,
    "state": "QUERY_STATE_COMPLETED",
    "result": {
      "data_uri": "https://api.dune.com/api/v1/execution/.../results",
      "metadata": {
        "column_names": ["week", "fee_usd"],
        "total_row_count": 25,
        "datapoint_count": 50
      }
    }
  },
  "visualizations": [
    {"title": "Chart", "image_url": "https://..."}
  ]
}
```

需用 API Key 请求 `data_uri` 获取实际数据。查询执行失败时不发送 webhook。

---

## 10. 用量追踪

```
GET /usage
```

---

## HTTP 错误码

| Code | 说明 |
|------|------|
| 200 | 成功 |
| 400 | 请求格式错误 |
| 401 | API Key 无效 |
| 402 | 超出 billing 限额 |
| 403 | 无权限（查询已归档/私有） |
| 404 | 资源不存在 |
| 409 | 请求冲突 |
| 429 | 限速 |
| 500 | 服务端错误 |
