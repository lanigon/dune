package stablemon

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

// TransferConfig configures large transfer detection.
type TransferConfig struct {
	MinAmountUSD float64 // minimum transfer size in USD
	LookbackMin  int     // lookback window in minutes
	Limit        int     // max results
}

func DefaultTransferConfig() TransferConfig {
	return TransferConfig{
		MinAmountUSD: 500000, // $500k+
		LookbackMin:  60,     // last hour
		Limit:        50,
	}
}

// FindLargeTransfers finds large USDC/USDT/USDT0 transfers.
func (m *Monitor) FindLargeTransfers(cfg TransferConfig) (*models.QueryResult, error) {
	sql := fmt.Sprintf(LargeTransfersSQL, cfg.LookbackMin, cfg.MinAmountUSD, cfg.Limit)
	return m.client.RunSQL(sql)
}

// CEXFlowConfig configures CEX flow monitoring.
type CEXFlowConfig struct {
	MinAmountUSD float64
	LookbackH    int
	Limit        int
}

func DefaultCEXFlowConfig() CEXFlowConfig {
	return CEXFlowConfig{
		MinAmountUSD: 1000000, // $1M+
		LookbackH:    24,
		Limit:        50,
	}
}

// GetCEXFlows tracks stablecoin flows to/from centralized exchanges.
func (m *Monitor) GetCEXFlows(cfg CEXFlowConfig) (*models.QueryResult, error) {
	sql := fmt.Sprintf(CEXFlowSQL, cfg.LookbackH, cfg.MinAmountUSD, cfg.Limit)
	return m.client.RunSQL(sql)
}

// BridgeFlowConfig configures bridge flow monitoring.
type BridgeFlowConfig struct {
	MinAmountUSD float64
	LookbackH    int
	Limit        int
}

func DefaultBridgeFlowConfig() BridgeFlowConfig {
	return BridgeFlowConfig{
		MinAmountUSD: 100000, // $100k+
		LookbackH:    24,
		Limit:        50,
	}
}

// GetBridgeFlows tracks stablecoin cross-chain bridge transfers.
func (m *Monitor) GetBridgeFlows(cfg BridgeFlowConfig) (*models.QueryResult, error) {
	sql := fmt.Sprintf(BridgeFlowSQL, cfg.LookbackH, cfg.MinAmountUSD, cfg.Limit)
	return m.client.RunSQL(sql)
}

// GetHourlyVolume returns hourly stablecoin transfer volume.
func (m *Monitor) GetHourlyVolume(lookbackH int) (*models.QueryResult, error) {
	sql := fmt.Sprintf(HourlyVolumeSQL, lookbackH)
	return m.client.RunSQL(sql)
}

// TopAddressConfig configures top sender/receiver queries.
type TopAddressConfig struct {
	LookbackH int
	Limit     int
}

func DefaultTopAddressConfig() TopAddressConfig {
	return TopAddressConfig{
		LookbackH: 24,
		Limit:     30,
	}
}

// GetTopSenders finds top stablecoin senders.
func (m *Monitor) GetTopSenders(cfg TopAddressConfig) (*models.QueryResult, error) {
	sql := fmt.Sprintf(TopSendersSQL, cfg.LookbackH, cfg.Limit)
	return m.client.RunSQL(sql)
}

// GetTopReceivers finds top stablecoin receivers.
func (m *Monitor) GetTopReceivers(cfg TopAddressConfig) (*models.QueryResult, error) {
	sql := fmt.Sprintf(TopReceiversSQL, cfg.LookbackH, cfg.Limit)
	return m.client.RunSQL(sql)
}

// StableTransfer is a parsed large transfer result.
type StableTransfer struct {
	Blockchain string  `json:"blockchain"`
	TokenSymbol string `json:"token_symbol"`
	AmountUSD  float64 `json:"amount_usd"`
	Amount     float64 `json:"amount"`
	From       string  `json:"from"`
	To         string  `json:"to"`
	FromLabel  string  `json:"from_label,omitempty"`
	ToLabel    string  `json:"to_label,omitempty"`
	TxHash     string  `json:"tx_hash"`
}

// ParseTransfers extracts typed transfers from raw results.
func ParseTransfers(result *models.QueryResult) []StableTransfer {
	if result == nil || result.Result == nil {
		return nil
	}
	var transfers []StableTransfer
	for _, row := range result.Result.Rows {
		t := StableTransfer{
			Blockchain:  strVal(row, "blockchain"),
			TokenSymbol: strVal(row, "token_symbol"),
			AmountUSD:   floatVal(row, "amount_usd"),
			Amount:      floatVal(row, "amount"),
			From:        strVal(row, "from"),
			To:          strVal(row, "to"),
			FromLabel:   strVal(row, "from_label"),
			ToLabel:     strVal(row, "to_label"),
			TxHash:      strVal(row, "tx_hash"),
		}
		transfers = append(transfers, t)
	}
	return transfers
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
