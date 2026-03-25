package aipay

import (
	"fmt"

	"github.com/bergtatt/morpheco/scripts/pkg/dune"
	"github.com/bergtatt/morpheco/scripts/pkg/models"
)

type Monitor struct {
	client *dune.Client
}

func New(client *dune.Client) *Monitor {
	return &Monitor{client: client}
}

// MicropaymentConfig configures micropayment pattern detection.
type MicropaymentConfig struct {
	LookbackH     int
	MinAmountUSD  float64 // minimum per-tx amount
	MaxAmountUSD  float64 // maximum per-tx amount
	MinTxCount    int     // minimum tx count to qualify as pattern
	Limit         int
}

func DefaultMicropaymentConfig() MicropaymentConfig {
	return MicropaymentConfig{
		LookbackH:    24,
		MinAmountUSD: 0.01,
		MaxAmountUSD: 100,    // AI payments typically < $100
		MinTxCount:   5,      // at least 5 txs
		Limit:        50,
	}
}

// FindMicropaymentPatterns detects high-frequency small payment patterns.
func (m *Monitor) FindMicropaymentPatterns(cfg MicropaymentConfig) (*models.QueryResult, error) {
	sql := fmt.Sprintf(MicropaymentPatternSQL,
		cfg.LookbackH, cfg.MinAmountUSD, cfg.MaxAmountUSD, cfg.MinTxCount, cfg.Limit)
	return m.client.RunSQL(sql)
}

// RecurringConfig configures recurring payment detection.
type RecurringConfig struct {
	LookbackH  int
	MinTxCount int // minimum payment count
	Limit      int
}

func DefaultRecurringConfig() RecurringConfig {
	return RecurringConfig{
		LookbackH:  48,
		MinTxCount: 3,
		Limit:      50,
	}
}

// FindRecurringPayments finds addresses with regular payment patterns.
// Low interval_stddev relative to avg_interval_min indicates regularity.
func (m *Monitor) FindRecurringPayments(cfg RecurringConfig) (*models.QueryResult, error) {
	sql := fmt.Sprintf(RecurringPaymentSQL, cfg.LookbackH, cfg.MinTxCount, cfg.Limit)
	return m.client.RunSQL(sql)
}

// ChannelConfig configures payment channel detection.
type ChannelConfig struct {
	LookbackH  int
	MinTxCount int
	Limit      int
}

func DefaultChannelConfig() ChannelConfig {
	return ChannelConfig{
		LookbackH:  24,
		MinTxCount: 10,
		Limit:      50,
	}
}

// FindPaymentChannels detects payment channel-like behavior.
func (m *Monitor) FindPaymentChannels(cfg ChannelConfig) (*models.QueryResult, error) {
	sql := fmt.Sprintf(PaymentChannelSQL, cfg.LookbackH, cfg.MinTxCount, cfg.Limit)
	return m.client.RunSQL(sql)
}

// ServicePaymentConfig configures service payment analysis.
type ServicePaymentConfig struct {
	LookbackH  int
	MinTxCount int
	Limit      int
}

func DefaultServicePaymentConfig() ServicePaymentConfig {
	return ServicePaymentConfig{
		LookbackH:  48,
		MinTxCount: 3,
		Limit:      50,
	}
}

// AnalyzeServicePayments tracks payments to known labeled services.
func (m *Monitor) AnalyzeServicePayments(cfg ServicePaymentConfig) (*models.QueryResult, error) {
	sql := fmt.Sprintf(ServicePaymentSQL, cfg.LookbackH, cfg.MinTxCount, cfg.Limit)
	return m.client.RunSQL(sql)
}

// MicropaymentPattern is a parsed micropayment pattern.
type MicropaymentPattern struct {
	Blockchain    string  `json:"blockchain"`
	Sender        string  `json:"sender"`
	Receiver      string  `json:"receiver"`
	TokenSymbol   string  `json:"token_symbol"`
	TxCount       int     `json:"tx_count"`
	TotalUSD      float64 `json:"total_usd"`
	AvgUSD        float64 `json:"avg_usd"`
	SenderLabel   string  `json:"sender_label,omitempty"`
	ReceiverLabel string  `json:"receiver_label,omitempty"`
}

// ParseMicropayments extracts typed micropayment patterns from raw results.
func ParseMicropayments(result *models.QueryResult) []MicropaymentPattern {
	if result == nil || result.Result == nil {
		return nil
	}
	var patterns []MicropaymentPattern
	for _, row := range result.Result.Rows {
		p := MicropaymentPattern{
			Blockchain:    strVal(row, "blockchain"),
			Sender:        strVal(row, "sender"),
			Receiver:      strVal(row, "receiver"),
			TokenSymbol:   strVal(row, "token_symbol"),
			TxCount:       intVal(row, "tx_count"),
			TotalUSD:      floatVal(row, "total_usd"),
			AvgUSD:        floatVal(row, "avg_usd"),
			SenderLabel:   strVal(row, "sender_label"),
			ReceiverLabel: strVal(row, "receiver_label"),
		}
		patterns = append(patterns, p)
	}
	return patterns
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

func intVal(row map[string]interface{}, key string) int {
	if v, ok := row[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		}
	}
	return 0
}
