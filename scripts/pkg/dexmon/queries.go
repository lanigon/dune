package dexmon

// DuneSQL queries for DEX monitoring.

// WhaleTradesSQL finds large trades in the last N minutes.
const WhaleTradesSQL = `
SELECT
  blockchain,
  project,
  version,
  block_time,
  token_bought_symbol,
  token_sold_symbol,
  token_pair,
  token_bought_amount,
  token_sold_amount,
  amount_usd,
  taker,
  maker,
  tx_hash,
  project_contract_address
FROM dex.trades
WHERE block_time > now() - interval '%d' minute
  AND amount_usd > %f
ORDER BY amount_usd DESC
LIMIT %d
`

// TopPairsSQL returns top trading pairs by volume.
const TopPairsSQL = `
SELECT
  blockchain,
  project,
  token_pair,
  token_bought_symbol,
  token_sold_symbol,
  COUNT(*) AS trade_count,
  SUM(amount_usd) AS total_volume_usd,
  AVG(amount_usd) AS avg_trade_usd,
  MAX(amount_usd) AS max_trade_usd
FROM dex.trades
WHERE blockchain = '%s'
  AND block_time > now() - interval '%d' hour
  AND block_month >= DATE_ADD('month', -1, CURRENT_DATE)
GROUP BY 1, 2, 3, 4, 5
ORDER BY total_volume_usd DESC
LIMIT %d
`

// VolumeByChainSQL returns hourly volume by chain.
const VolumeByChainSQL = `
SELECT
  blockchain,
  DATE_TRUNC('hour', block_time) AS hour,
  COUNT(*) AS trade_count,
  SUM(amount_usd) AS volume_usd
FROM dex.trades
WHERE block_time > now() - interval '%d' hour
  AND block_month >= DATE_ADD('month', -1, CURRENT_DATE)
GROUP BY 1, 2
ORDER BY hour DESC, volume_usd DESC
`

// SandwichDetectSQL finds recent sandwich attacks.
const SandwichDetectSQL = `
SELECT
  s.blockchain,
  s.block_time,
  s.project,
  s.token_pair,
  s.frontrun_tx_hash,
  s.backrun_tx_hash,
  v.tx_hash AS victim_tx_hash,
  v.amount_usd AS victim_amount_usd,
  s.profit_amount_usd
FROM dex.sandwiches s
JOIN dex.sandwiched v
  ON s.blockchain = v.blockchain
  AND s.tx_hash = v.frontrun_tx_hash
WHERE s.block_time > now() - interval '%d' hour
  AND s.blockchain = '%s'
ORDER BY s.profit_amount_usd DESC
LIMIT %d
`

// NewPoolsSQL detects recently created liquidity pools.
const NewPoolsSQL = `
SELECT
  blockchain,
  project,
  version,
  token_pair,
  token_bought_symbol,
  token_sold_symbol,
  MIN(block_time) AS first_trade_time,
  COUNT(*) AS trade_count,
  SUM(amount_usd) AS total_volume_usd
FROM dex.trades
WHERE block_time > now() - interval '%d' hour
  AND block_month >= DATE_ADD('month', -1, CURRENT_DATE)
  AND blockchain = '%s'
GROUP BY 1, 2, 3, 4, 5, 6
HAVING MIN(block_time) > now() - interval '%d' hour
ORDER BY total_volume_usd DESC
LIMIT %d
`

// TokenFlowSQL tracks buy/sell pressure for a specific token.
const TokenFlowSQL = `
SELECT
  DATE_TRUNC('hour', block_time) AS hour,
  SUM(CASE WHEN token_bought_symbol = '%s' THEN amount_usd ELSE 0 END) AS buy_volume_usd,
  SUM(CASE WHEN token_sold_symbol = '%s' THEN amount_usd ELSE 0 END) AS sell_volume_usd,
  SUM(CASE WHEN token_bought_symbol = '%s' THEN amount_usd ELSE 0 END)
    - SUM(CASE WHEN token_sold_symbol = '%s' THEN amount_usd ELSE 0 END) AS net_flow_usd,
  COUNT(*) AS trade_count
FROM dex.trades
WHERE block_time > now() - interval '%d' hour
  AND (token_bought_symbol = '%s' OR token_sold_symbol = '%s')
  AND block_month >= DATE_ADD('month', -1, CURRENT_DATE)
GROUP BY 1
ORDER BY 1 DESC
`
