package dexmon

import (
	"fmt"
	"time"

	"github.com/bergtatt/morpheco/scripts/pkg/dune"
	"github.com/bergtatt/morpheco/scripts/pkg/models"
)

type Monitor struct {
	client *dune.Client
}

func New(client *dune.Client) *Monitor {
	return &Monitor{client: client}
}

// WhaleConfig configures whale trade detection.
type WhaleConfig struct {
	MinAmountUSD float64 // minimum trade size in USD
	LookbackMin  int     // lookback window in minutes
	Limit        int     // max results
}

func DefaultWhaleConfig() WhaleConfig {
	return WhaleConfig{
		MinAmountUSD: 100000, // $100k+
		LookbackMin:  60,     // last hour
		Limit:        50,
	}
}

// FindWhaleTrades returns large trades in the recent window.
func (m *Monitor) FindWhaleTrades(cfg WhaleConfig) (*models.QueryResult, error) {
	sql := fmt.Sprintf(WhaleTradesSQL, cfg.LookbackMin, cfg.MinAmountUSD, cfg.Limit)
	return m.client.RunSQL(sql)
}

// TopPairsConfig configures top pairs query.
type TopPairsConfig struct {
	Blockchain  string
	LookbackH   int
	Limit       int
}

func DefaultTopPairsConfig() TopPairsConfig {
	return TopPairsConfig{
		Blockchain: "ethereum",
		LookbackH:  24,
		Limit:      30,
	}
}

// GetTopPairs returns top trading pairs by volume.
func (m *Monitor) GetTopPairs(cfg TopPairsConfig) (*models.QueryResult, error) {
	sql := fmt.Sprintf(TopPairsSQL, cfg.Blockchain, cfg.LookbackH, cfg.Limit)
	return m.client.RunSQL(sql)
}

// GetVolumeByChain returns hourly volume across chains.
func (m *Monitor) GetVolumeByChain(lookbackH int) (*models.QueryResult, error) {
	sql := fmt.Sprintf(VolumeByChainSQL, lookbackH)
	return m.client.RunSQL(sql)
}

// SandwichConfig configures sandwich attack detection.
type SandwichConfig struct {
	Blockchain string
	LookbackH  int
	Limit      int
}

// DetectSandwiches finds recent sandwich attacks.
func (m *Monitor) DetectSandwiches(cfg SandwichConfig) (*models.QueryResult, error) {
	sql := fmt.Sprintf(SandwichDetectSQL, cfg.LookbackH, cfg.Blockchain, cfg.Limit)
	return m.client.RunSQL(sql)
}

// NewPoolsConfig configures new pool detection.
type NewPoolsConfig struct {
	Blockchain string
	LookbackH  int
	Limit      int
}

// FindNewPools detects recently created liquidity pools.
func (m *Monitor) FindNewPools(cfg NewPoolsConfig) (*models.QueryResult, error) {
	sql := fmt.Sprintf(NewPoolsSQL, cfg.LookbackH, cfg.Blockchain, cfg.LookbackH, cfg.Limit)
	return m.client.RunSQL(sql)
}

// TokenFlowConfig configures token flow analysis.
type TokenFlowConfig struct {
	TokenSymbol string
	LookbackH   int
}

// GetTokenFlow tracks buy/sell pressure for a token.
func (m *Monitor) GetTokenFlow(cfg TokenFlowConfig) (*models.QueryResult, error) {
	t := cfg.TokenSymbol
	sql := fmt.Sprintf(TokenFlowSQL, t, t, t, t, cfg.LookbackH, t, t)
	return m.client.RunSQL(sql)
}

// WhaleTrade is a parsed whale trade result.
type WhaleTrade struct {
	Blockchain  string    `json:"blockchain"`
	Project     string    `json:"project"`
	BlockTime   time.Time `json:"block_time"`
	BuyToken    string    `json:"buy_token"`
	SellToken   string    `json:"sell_token"`
	AmountUSD   float64   `json:"amount_usd"`
	TxHash      string    `json:"tx_hash"`
}

// ParseWhaleTrades extracts typed whale trades from raw results.
func ParseWhaleTrades(result *models.QueryResult) []WhaleTrade {
	if result == nil || result.Result == nil {
		return nil
	}
	var trades []WhaleTrade
	for _, row := range result.Result.Rows {
		t := WhaleTrade{
			Blockchain: strVal(row, "blockchain"),
			Project:    strVal(row, "project"),
			BuyToken:   strVal(row, "token_bought_symbol"),
			SellToken:  strVal(row, "token_sold_symbol"),
			AmountUSD:  floatVal(row, "amount_usd"),
			TxHash:     strVal(row, "tx_hash"),
		}
		trades = append(trades, t)
	}
	return trades
}

func strVal(row map[string]interface{}, key string) string {
	if v, ok := row[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func floatVal(row map[string]interface{}, key string) float64 {
	if v, ok := row[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case string:
			var f float64
			fmt.Sscanf(val, "%f", &f)
			return f
		}
	}
	return 0
}
