# Dune 数据目录

## 概述

Dune 的数据分为四层：Raw → Decoded → Spellbook（Curated）→ Community

---

## 1. 支持的链

### EVM 链（50+）
Ethereum, Arbitrum, Optimism, Base, Polygon, zkSync, Scroll, Linea, Blast, Mantle, BNB Chain, Avalanche, Fantom, Gnosis, Ronin, Unichain, Abstract, Berachain, Flare, Sonic, X Layer, ApeChain, ...

### 非 EVM 链
- **Solana** — 完整 DeFi/NFT 覆盖
- **Bitcoin** — blocks, transactions, inputs, outputs
- **Aptos** — blocks, events, modules, resources, transactions
- **Tron, TON**

### 社区/第三方数据
- Farcaster (via Neynar) — casts, profiles, reactions, verifications
- Flashbots — mempool data
- Hyperliquid — market data
- Lens Protocol — profiles, publications, follows
- Reservoir — NFT 交易数据
- Snapshot — 治理投票数据

---

## 2. Raw Tables（每条链独立）

| 表 | 说明 |
|----|------|
| `<chain>.blocks` | 区块头信息 |
| `<chain>.transactions` | 交易记录 |
| `<chain>.logs` | 事件日志 |
| `<chain>.traces` | 内部调用追踪 |

示例：`ethereum.transactions`, `polygon.logs`, `arbitrum.blocks`

---

## 3. Decoded Tables

通过提交合约 ABI，Dune 自动解码事件和函数调用为可读表。

命名空间：`<project>_<chain>.<event_name>` 或 `<project>_<chain>.<function_name>`

示例：`uniswap_v3_ethereum.Swap`, `aave_v3_ethereum.Supply`

---

## 4. Curated / Spellbook 表（重点）

维护者：Dune 数据团队 | 刷新频率：~30 分钟 | 开源：github.com/duneanalytics/spellbook

### DEX 交易

| 表 | 覆盖 | 说明 |
|----|------|------|
| `dex.trades` | 50+ EVM 链 | 所有 DEX 原始交易（Uniswap, Curve, SushiSwap 等） |
| `dex_aggregator.trades` | EVM | 聚合器路由交易（1inch, 0x, ParaSwap） |
| `dex.sandwiches` | EVM | 三明治攻击的外部交易 |
| `dex.sandwiched` | EVM | 被三明治攻击的受害交易 |
| `dex_solana.trades` | Solana | Solana DEX 交易（Orca, Raydium 等） |
| `jupiter_solana.aggregator_swaps` | Solana | Jupiter 聚合器交易 |

#### dex.trades 表结构

| 列 | 类型 | 说明 |
|----|------|------|
| `blockchain` | VARCHAR | 链名 |
| `project` | VARCHAR | DEX 名称 |
| `version` | VARCHAR | 协议版本 |
| `block_month` | DATE | 分区键 |
| `block_date` | DATE | 日期 |
| `block_time` | TIMESTAMP | 时间戳 |
| `block_number` | BIGINT | 区块号 |
| `token_bought_symbol` | VARCHAR | 买入代币符号 |
| `token_sold_symbol` | VARCHAR | 卖出代币符号 |
| `token_pair` | VARCHAR | 交易对（字母序） |
| `token_bought_amount` | DOUBLE | 买入数量（display units） |
| `token_sold_amount` | DOUBLE | 卖出数量 |
| `token_bought_amount_raw` | UINT256 | 原始买入数量 |
| `token_sold_amount_raw` | UINT256 | 原始卖出数量 |
| `amount_usd` | DOUBLE | USD 价值 |
| `token_bought_address` | VARBINARY | 买入代币合约 |
| `token_sold_address` | VARBINARY | 卖出代币合约 |
| `taker` | VARBINARY | 买入方地址 |
| `maker` | VARBINARY | 卖出方地址 |
| `project_contract_address` | VARBINARY | 合约地址（pool/router） |
| `tx_hash` | VARBINARY | 交易哈希 |
| `tx_from` | VARBINARY | 发起 EOA |
| `tx_to` | VARBINARY | 首次调用地址 |
| `evt_index` | BIGINT | 事件索引 |

**查询优化**：分区键为 `blockchain`, `project`, `block_month`。始终包含这些过滤条件。

```sql
-- 好：使用分区键
SELECT * FROM dex.trades
WHERE blockchain = 'ethereum'
  AND block_month >= DATE '2025-01-01'
  AND project = 'uniswap'

-- 慢：没有分区过滤
SELECT * FROM dex.trades
WHERE token_bought_symbol = 'USDC'
```

### 代币价格

| 表 | 说明 |
|----|------|
| `prices.usd` | 综合价格（分钟级） |
| `prices.day` | 日级价格 |
| `prices.hour` | 小时级价格 |
| `prices.minute` | 分钟级价格 |

### 代币与余额

| 表 | 说明 |
|----|------|
| `tokens.erc20` | ERC20 代币元数据（symbol, decimals, address） |
| `balances` 系列 | EVM + Solana 地址余额 |
| `solana_utils.latest_balances` | Solana 当前余额 |
| `solana_utils.daily_balances` | Solana 历史日余额 |

### NFT

| 表 | 说明 |
|----|------|
| `nft.trades` | NFT 交易（所有 marketplace） |
| `nft.mints` | NFT 铸造事件 |
| `nft.wash_trades` | 洗盘交易检测 |

### DeFi 借贷

| 表 | 说明 |
|----|------|
| `lending.borrow` | 借款/还款事件 |
| `lending.supply` | 存款/取款事件 |
| `lending.flashloans` | 闪电贷 |
| `lending.info` | 市场快照（利率、TVL） |

### 跨链桥

| 表 | 说明 |
|----|------|
| `bridges_evms.deposits` | 桥存款（源链锁定/销毁） |
| `bridges_evms.withdrawals` | 桥提款（目标链铸造/释放） |
| `bridges_evms.flows` | 匹配的完整跨链流 |

### 稳定币

覆盖 EVM、Solana、Tron 的稳定币转账、余额、活动分类。

### CEX 流量

| 表 | 说明 |
|----|------|
| `cex.addresses` | 已知 CEX 地址目录 |
| `cex.deposit_addresses` | CEX 充值地址 |
| `cex.flows` | CEX 进出流量（29 条链） |

### Gas 费

| 表 | 说明 |
|----|------|
| `gas.fees` | EVM 交易级 gas 数据 |
| `gas_solana.fees` | Solana 交易费数据 |

### 地址标签

| 表 | 说明 |
|----|------|
| `labels.addresses` | 地址标签和归属 |
| `labels.ens` | ENS 域名解析 |
| `labels.owner_addresses` | 实体地址映射 |
| `labels.owner_details` | 实体详情 |

### 预测市场

| 表 | 说明 |
|----|------|
| `polymarket_polygon.market_details` | Polymarket 市场元数据 |
| `kalshi.market_report` | Kalshi 市场数据 |
| `kalshi.trade_report` | Kalshi 交易数据 |

### 其他

- **Rollup Economics** — L2 收入 vs L1 发布成本
- **Ethereum Staking** — 信标链质押数据
- **Utilities** — 时间序列辅助表

---

## 5. 常用查询示例

### 每周 DEX 总交易量
```sql
SELECT
  blockchain,
  DATE_TRUNC('week', block_time) AS week,
  SUM(CAST(amount_usd AS DOUBLE)) AS usd_volume
FROM dex.trades
WHERE block_time > NOW() - INTERVAL '365' day
GROUP BY 1, 2
```

### Ethereum 各 DEX 日交易量（30 天）
```sql
SELECT
  DATE_TRUNC('day', block_time) AS day,
  project,
  SUM(amount_usd) AS volume_usd,
  COUNT(*) AS num_trades
FROM dex.trades
WHERE blockchain = 'ethereum'
  AND block_month >= DATE '2025-01-01'
  AND block_time >= NOW() - INTERVAL '30' DAY
GROUP BY 1, 2
ORDER BY 1 DESC, 3 DESC
```

### Uniswap V3 Top 交易对（7 天）
```sql
SELECT
  token_bought_symbol,
  token_sold_symbol,
  SUM(amount_usd) AS volume_usd,
  COUNT(*) AS num_trades
FROM dex.trades
WHERE blockchain = 'ethereum'
  AND project = 'uniswap'
  AND version = '3'
  AND block_time >= NOW() - INTERVAL '7' DAY
GROUP BY 1, 2
ORDER BY 3 DESC
LIMIT 20
```

### 跨链 DEX 交易量对比
```sql
SELECT
  blockchain,
  DATE_TRUNC('day', block_time) AS date,
  SUM(amount_usd) AS volume_usd
FROM dex.trades
WHERE block_time >= NOW() - INTERVAL '30' day
GROUP BY 1, 2
ORDER BY date DESC, volume_usd DESC
```
