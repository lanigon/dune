//go:build stablemon || all

package main

import (
	"flag"

	"github.com/bergtatt/morpheco/scripts/pkg/stablemon"
)

func init() {
	Register(&Command{Name: "transfers", Desc: "Large USDC/USDT/USDT0 transfers", Run: runTransfers})
	Register(&Command{Name: "cex", Desc: "Stablecoin CEX inflow/outflow", Run: runCEX})
	Register(&Command{Name: "bridges", Desc: "Stablecoin bridge flows", Run: runBridges})
	Register(&Command{Name: "stvolume", Desc: "Hourly stablecoin volume", Run: runStVolume})
	Register(&Command{Name: "senders", Desc: "Top stablecoin senders", Run: runSenders})
	Register(&Command{Name: "receivers", Desc: "Top stablecoin receivers", Run: runReceivers})
}

func runTransfers(args []string) {
	fs := flag.NewFlagSet("transfers", flag.ExitOnError)
	min := fs.Float64("min", 500000, "Min USD")
	lookback := fs.Int("lookback", 60, "Lookback minutes")
	limit := fs.Int("limit", 50, "Max results")
	fs.Parse(args)

	mon := stablemon.New(mustClient())
	result, err := mon.FindLargeTransfers(stablemon.TransferConfig{
		MinAmountUSD: *min, LookbackMin: *lookback, Limit: *limit,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(stablemon.ParseTransfers(result))
}

func runCEX(args []string) {
	fs := flag.NewFlagSet("cex", flag.ExitOnError)
	min := fs.Float64("min", 1000000, "Min USD")
	hours := fs.Int("hours", 24, "Lookback hours")
	limit := fs.Int("limit", 50, "Max results")
	fs.Parse(args)

	mon := stablemon.New(mustClient())
	result, err := mon.GetCEXFlows(stablemon.CEXFlowConfig{
		MinAmountUSD: *min, LookbackH: *hours, Limit: *limit,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}

func runBridges(args []string) {
	fs := flag.NewFlagSet("bridges", flag.ExitOnError)
	min := fs.Float64("min", 100000, "Min USD")
	hours := fs.Int("hours", 24, "Lookback hours")
	limit := fs.Int("limit", 50, "Max results")
	fs.Parse(args)

	mon := stablemon.New(mustClient())
	result, err := mon.GetBridgeFlows(stablemon.BridgeFlowConfig{
		MinAmountUSD: *min, LookbackH: *hours, Limit: *limit,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}

func runStVolume(args []string) {
	fs := flag.NewFlagSet("stvolume", flag.ExitOnError)
	hours := fs.Int("hours", 24, "Lookback hours")
	fs.Parse(args)

	mon := stablemon.New(mustClient())
	result, err := mon.GetHourlyVolume(*hours)
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}

func runSenders(args []string) {
	fs := flag.NewFlagSet("senders", flag.ExitOnError)
	hours := fs.Int("hours", 24, "Lookback hours")
	limit := fs.Int("limit", 30, "Max results")
	fs.Parse(args)

	mon := stablemon.New(mustClient())
	result, err := mon.GetTopSenders(stablemon.TopAddressConfig{
		LookbackH: *hours, Limit: *limit,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}

func runReceivers(args []string) {
	fs := flag.NewFlagSet("receivers", flag.ExitOnError)
	hours := fs.Int("hours", 24, "Lookback hours")
	limit := fs.Int("limit", 30, "Max results")
	fs.Parse(args)

	mon := stablemon.New(mustClient())
	result, err := mon.GetTopReceivers(stablemon.TopAddressConfig{
		LookbackH: *hours, Limit: *limit,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}
