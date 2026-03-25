package aipay

// DuneSQL queries for AI payment monitoring and analysis.
//
// AI Payment 场景：
// 1. AI Agent 自动用稳定币支付（API 调用费、数据费、计算费）
// 2. 链上 AI 服务的收付款追踪
// 3. AI 相关协议的交易分析（如 Skyfire, Payman, Morpheus, Autonolas 等）
// 4. 微支付通道分析（高频小额交易模式检测）

// MicropaymentPatternSQL detects micropayment patterns (high frequency, small amounts).
// AI agents typically make many small payments — this identifies such patterns.
const MicropaymentPatternSQL = `
SELECT
  blockchain,
  "from" AS sender,
  "to" AS receiver,
  token_symbol,
  COUNT(*) AS tx_count,
  SUM(amount_usd) AS total_usd,
  AVG(amount_usd) AS avg_usd,
  MIN(amount_usd) AS min_usd,
  MAX(amount_usd) AS max_usd,
  MIN(block_time) AS first_tx,
  MAX(block_time) AS last_tx,
  COALESCE(ls.name, '') AS sender_label,
  COALESCE(lr.name, '') AS receiver_label
FROM tokens.transfers t
LEFT JOIN labels.addresses ls
  ON t.blockchain = ls.blockchain AND t."from" = ls.address
LEFT JOIN labels.addresses lr
  ON t.blockchain = lr.blockchain AND t."to" = lr.address
WHERE t.block_time > now() - interval '%d' hour
  AND t.token_symbol IN ('USDC', 'USDT', 'USDT0', 'DAI')
  AND t.amount_usd BETWEEN %f AND %f
GROUP BY 1, 2, 3, 4, ls.name, lr.name
HAVING COUNT(*) >= %d
ORDER BY tx_count DESC
LIMIT %d
`

// RecurringPaymentSQL finds addresses with regular payment patterns.
const RecurringPaymentSQL = `
WITH payment_intervals AS (
  SELECT
    blockchain,
    "from" AS sender,
    "to" AS receiver,
    token_symbol,
    block_time,
    amount_usd,
    LAG(block_time) OVER (PARTITION BY "from", "to", token_symbol ORDER BY block_time) AS prev_time,
    DATE_DIFF('minute', LAG(block_time) OVER (PARTITION BY "from", "to", token_symbol ORDER BY block_time), block_time) AS interval_min
  FROM tokens.transfers
  WHERE block_time > now() - interval '%d' hour
    AND token_symbol IN ('USDC', 'USDT', 'USDT0')
    AND amount_usd BETWEEN 0.01 AND 1000
)
SELECT
  blockchain,
  sender,
  receiver,
  token_symbol,
  COUNT(*) AS payment_count,
  SUM(amount_usd) AS total_usd,
  AVG(amount_usd) AS avg_amount_usd,
  AVG(interval_min) AS avg_interval_min,
  STDDEV(interval_min) AS interval_stddev,
  MIN(block_time) AS first_payment,
  MAX(block_time) AS last_payment
FROM payment_intervals
WHERE interval_min IS NOT NULL
GROUP BY 1, 2, 3, 4
HAVING COUNT(*) >= %d
  AND STDDEV(interval_min) < AVG(interval_min) * 0.5
ORDER BY payment_count DESC
LIMIT %d
`

// PaymentChannelSQL tracks payment channel-like behavior
// (same sender-receiver pair with frequent transfers).
const PaymentChannelSQL = `
SELECT
  blockchain,
  "from" AS sender,
  "to" AS receiver,
  token_symbol,
  COUNT(*) AS tx_count,
  SUM(amount_usd) AS total_usd,
  AVG(amount_usd) AS avg_usd,
  APPROX_PERCENTILE(amount_usd, 0.5) AS median_usd,
  MIN(block_time) AS first_tx,
  MAX(block_time) AS last_tx,
  DATE_DIFF('minute', MIN(block_time), MAX(block_time)) AS duration_min
FROM tokens.transfers
WHERE block_time > now() - interval '%d' hour
  AND token_symbol IN ('USDC', 'USDT', 'USDT0')
  AND amount_usd > 0
GROUP BY 1, 2, 3, 4
HAVING COUNT(*) >= %d
ORDER BY tx_count DESC
LIMIT %d
`

// ServicePaymentSQL analyzes payments to known AI/compute service contracts.
// Tracks which addresses are paying for on-chain services.
const ServicePaymentSQL = `
SELECT
  t.blockchain,
  t."from" AS payer,
  t."to" AS service,
  t.token_symbol,
  COALESCE(l.name, '') AS service_label,
  COUNT(*) AS payment_count,
  SUM(t.amount_usd) AS total_paid_usd,
  AVG(t.amount_usd) AS avg_payment_usd,
  MIN(t.block_time) AS first_payment,
  MAX(t.block_time) AS last_payment
FROM tokens.transfers t
LEFT JOIN labels.addresses l
  ON t.blockchain = l.blockchain AND t."to" = l.address
WHERE t.block_time > now() - interval '%d' hour
  AND t.token_symbol IN ('USDC', 'USDT', 'USDT0')
  AND t.amount_usd BETWEEN 0.001 AND 500
  AND l.name IS NOT NULL
GROUP BY 1, 2, 3, 4, l.name
HAVING COUNT(*) >= %d
ORDER BY total_paid_usd DESC
LIMIT %d
`
