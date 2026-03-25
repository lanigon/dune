package stablemon

// DuneSQL queries for USDC and USDT0 stablecoin transfer monitoring.
//
// USDC contract addresses:
//   Ethereum:  0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48
//   Polygon:   0x3c499c542cef5e3811e1192ce70d8cc03d5c3359
//   Arbitrum:  0xaf88d065e77c8cc2239327c5edb3a432268e5831
//   Base:      0x833589fcd6edb6e08f4c7c32d4f71b54bda02913
//   Optimism:  0x0b2c639c533813f4aa9d7837caf62653d097ff85
//
// USDT0 (OFT / LayerZero bridged USDT):
//   Multi-chain OFT standard token, varies by chain.

// LargeTransfersSQL finds large USDC/USDT transfers.
const LargeTransfersSQL = `
SELECT
  t.blockchain,
  t.block_time,
  t.token_symbol,
  t.amount_usd,
  t.amount_raw / POWER(10, t.token_decimals) AS amount,
  t.from,
  t.to,
  t.tx_hash,
  COALESCE(lf.name, '') AS from_label,
  COALESCE(lt.name, '') AS to_label
FROM tokens.transfers t
LEFT JOIN labels.addresses lf
  ON t.blockchain = lf.blockchain AND t.from = lf.address
LEFT JOIN labels.addresses lt
  ON t.blockchain = lt.blockchain AND t.to = lt.address
WHERE t.block_time > now() - interval '%d' minute
  AND t.token_symbol IN ('USDC', 'USDT', 'USDT0')
  AND t.amount_usd > %f
ORDER BY t.amount_usd DESC
LIMIT %d
`

// CEXFlowSQL tracks stablecoin flows to/from centralized exchanges.
const CEXFlowSQL = `
SELECT
  f.blockchain,
  f.block_time,
  f.token_symbol,
  f.amount_usd,
  f.direction,
  f.exchange,
  f.tx_hash
FROM cex.flows f
WHERE f.block_time > now() - interval '%d' hour
  AND f.token_symbol IN ('USDC', 'USDT', 'USDT0')
  AND f.amount_usd > %f
ORDER BY f.amount_usd DESC
LIMIT %d
`

// BridgeFlowSQL tracks stablecoin cross-chain bridge transfers.
const BridgeFlowSQL = `
SELECT
  d.blockchain AS source_chain,
  w.blockchain AS dest_chain,
  d.token_symbol,
  d.amount_usd,
  d.block_time AS deposit_time,
  w.block_time AS withdrawal_time,
  d.tx_hash AS deposit_tx,
  w.tx_hash AS withdrawal_tx
FROM bridges_evms.deposits d
JOIN bridges_evms.withdrawals w
  ON d.bridge = w.bridge
  AND d.transfer_id = w.transfer_id
WHERE d.block_time > now() - interval '%d' hour
  AND d.token_symbol IN ('USDC', 'USDT', 'USDT0')
  AND d.amount_usd > %f
ORDER BY d.amount_usd DESC
LIMIT %d
`

// HourlyVolumeSQL tracks hourly stablecoin transfer volume.
const HourlyVolumeSQL = `
SELECT
  blockchain,
  token_symbol,
  DATE_TRUNC('hour', block_time) AS hour,
  COUNT(*) AS transfer_count,
  SUM(amount_usd) AS volume_usd,
  AVG(amount_usd) AS avg_transfer_usd,
  MAX(amount_usd) AS max_transfer_usd
FROM tokens.transfers
WHERE block_time > now() - interval '%d' hour
  AND token_symbol IN ('USDC', 'USDT', 'USDT0')
  AND amount_usd > 0
GROUP BY 1, 2, 3
ORDER BY hour DESC, volume_usd DESC
`

// TopSendersSQL finds top stablecoin senders.
const TopSendersSQL = `
SELECT
  blockchain,
  "from" AS sender,
  COALESCE(l.name, '') AS sender_label,
  COUNT(*) AS tx_count,
  SUM(amount_usd) AS total_usd,
  AVG(amount_usd) AS avg_usd
FROM tokens.transfers t
LEFT JOIN labels.addresses l
  ON t.blockchain = l.blockchain AND t."from" = l.address
WHERE t.block_time > now() - interval '%d' hour
  AND t.token_symbol IN ('USDC', 'USDT', 'USDT0')
  AND t.amount_usd > 1000
GROUP BY 1, 2, 3
ORDER BY total_usd DESC
LIMIT %d
`

// TopReceiversSQL finds top stablecoin receivers.
const TopReceiversSQL = `
SELECT
  blockchain,
  "to" AS receiver,
  COALESCE(l.name, '') AS receiver_label,
  COUNT(*) AS tx_count,
  SUM(amount_usd) AS total_usd,
  AVG(amount_usd) AS avg_usd
FROM tokens.transfers t
LEFT JOIN labels.addresses l
  ON t.blockchain = l.blockchain AND t."to" = l.address
WHERE t.block_time > now() - interval '%d' hour
  AND t.token_symbol IN ('USDC', 'USDT', 'USDT0')
  AND t.amount_usd > 1000
GROUP BY 1, 2, 3
ORDER BY total_usd DESC
LIMIT %d
`
