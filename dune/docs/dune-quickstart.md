# Dune 快速入门

## 概述

Dune 是区块链数据分析平台，索引 130+ 条链的链上数据，通过 DuneSQL（基于 Trino）查询，支持 Dashboard 可视化和 API 编程访问。

## 账户与认证

1. 注册：https://dune.com/auth/register （新账户附带免费 credits）
2. 生成 API Key：https://dune.com/apis?tab=keys
3. 认证方式：
   - Header: `X-Dune-Api-Key: <your-key>`
   - Query Param: `?api_key=<your-key>`

**注意：** Dune 有 user 和 team 两种账户。API Key 绑定特定上下文（user 或 team），team 下创建的查询只能用该 team 的 key 管理。

## 定价

| 项目 | Free | Plus ($399/月) | Enterprise |
|------|------|----------------|------------|
| 月 credits | 2,500 | 25,000 | 定制 |
| 超额费用 | $5/100 credits | $5/100 credits | 协商 |
| 低限速率（写） | 15 rpm | 70 rpm | 350+ rpm |
| 高限速率（读） | 40 rpm | 200 rpm | 1000+ rpm |
| MV 存储 | 100MB (单个 1MB) | 15GB | 200GB+ |
| 私有查询 | 否 | 是 | 是 |
| CSV 导出 | 否 | 是 | 是 |

### Credits 消耗

- **执行查询**：基于实际计算资源消耗
- **导出数据**：Free 20 credits/MB, Plus 2 credits/MB
- **写入数据**：3 credits/GB（最低 1 credit）
- **失败执行也收费**
- `get_latest_result` 不触发执行但按结果大小收费

## IP 限制

全局硬限制：1,000 requests/second（基于 IP）

## 核心概念

### DuneSQL

基于 Trino 的 SQL 方言，增强了区块链特性：
- `varbinary` 类型存储地址/哈希（`0x` 前缀）
- `uint256` 原生大整数类型
- `http_get()` / `http_post()` LiveFetch 函数
- 支持跨链查询
- 执行超时：30 分钟

### Spellbook

开源 dbt 项目（github.com/duneanalytics/spellbook），提供标准化表：
- `dex.trades` — DEX 交易数据（50+ EVM 链）
- `dex_solana.trades` — Solana DEX 交易
- `prices.usd` / `prices.day` / `prices.hour` — 代币价格
- `tokens.erc20` — 代币元数据
- `nft.trades` — NFT 交易
- `labels.addresses` — 地址标签

### Materialized Views

- 将查询结果存为可查询的表
- 命名规范：`dune.<team>.result_<name>`
- 独立刷新计划（最少 15 分钟，最多每周）
- 每次刷新消耗 credits

### 数据类型

| 类别 | 说明 |
|------|------|
| Raw tables | blocks, transactions, logs, traces（每条链独立） |
| Decoded tables | 通过 ABI 解码的事件/函数表 |
| Spellbook tables | 社区维护的标准化表 |
| Community uploads | 用户上传的 CSV 数据 |

## SDK

| 语言 | 包 | 安装 |
|------|---|------|
| Python | `dune-client` | `pip install dune-client` |
| TypeScript | `@duneanalytics/client-sdk` | `npm install @duneanalytics/client-sdk` |
| Go | `duneapi-client-go` | `go get github.com/duneanalytics/duneapi-client-go` |

## 快速示例

### Go
```go
client := duneapi.New(os.Getenv("DUNE_API_KEY"))
results, err := client.ExecuteQuery(3493826)
```

### Python
```python
from dune_client import DuneClient
dune = DuneClient()  # 读取 DUNE_API_KEY 环境变量
results = dune.execute(query_id=3493826)
```

### cURL（执行裸 SQL）
```bash
curl -X POST "https://api.dune.com/api/v1/sql/execute" \
  -H "Content-Type: application/json" \
  -H "X-Dune-Api-Key: $DUNE_API_KEY" \
  -d '{"sql": "SELECT * FROM dex.trades WHERE block_time > now() - interval '\''1'\'' day LIMIT 10"}'
```

## 结果存储

- 执行结果存储 90 天（`expires_at` 字段）
- 最大结果大小 32GB，超出需 `allow_partial_results=true`

## 相关链接

- 文档：https://docs.dune.com/
- API 参考：https://docs.dune.com/api-reference/overview/introduction
- 数据目录：https://docs.dune.com/data-catalog/curated/overview
- Spellbook：https://github.com/duneanalytics/spellbook
- LLM 索引：https://docs.dune.com/llms.txt
