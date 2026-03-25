//go:build dexmon || all

package main

import (
	"flag"
	"fmt"

	"github.com/bergtatt/morpheco/scripts/pkg/dexmon"
)

func init() {
	Register(&Command{Name: "whales", Desc: "DEX whale trades", Run: runWhales})
	Register(&Command{Name: "pairs", Desc: "Top DEX trading pairs", Run: runPairs})
	Register(&Command{Name: "volume", Desc: "DEX volume by chain", Run: runVolume})
	Register(&Command{Name: "sandwich", Desc: "Sandwich attack detection", Run: runSandwich})
	Register(&Command{Name: "newpools", Desc: "New liquidity pools", Run: runNewPools})
	Register(&Command{Name: "flow", Desc: "Token buy/sell flow", Run: runTokenFlow})
}

func runWhales(args []string) {
	fs := flag.NewFlagSet("whales", flag.ExitOnError)
	min := fs.Float64("min", 100000, "Min trade USD")
	lookback := fs.Int("lookback", 60, "Lookback minutes")
	limit := fs.Int("limit", 50, "Max results")
	fs.Parse(args)

	mon := dexmon.New(mustClient())
	result, err := mon.FindWhaleTrades(dexmon.WhaleConfig{
		MinAmountUSD: *min, LookbackMin: *lookback, Limit: *limit,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(dexmon.ParseWhaleTrades(result))
}

func runPairs(args []string) {
	fs := flag.NewFlagSet("pairs", flag.ExitOnError)
	chain := fs.String("chain", "ethereum", "Blockchain")
	hours := fs.Int("hours", 24, "Lookback hours")
	limit := fs.Int("limit", 30, "Max results")
	fs.Parse(args)

	mon := dexmon.New(mustClient())
	result, err := mon.GetTopPairs(dexmon.TopPairsConfig{
		Blockchain: *chain, LookbackH: *hours, Limit: *limit,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}

func runVolume(args []string) {
	fs := flag.NewFlagSet("volume", flag.ExitOnError)
	hours := fs.Int("hours", 24, "Lookback hours")
	fs.Parse(args)

	mon := dexmon.New(mustClient())
	result, err := mon.GetVolumeByChain(*hours)
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}

func runSandwich(args []string) {
	fs := flag.NewFlagSet("sandwich", flag.ExitOnError)
	chain := fs.String("chain", "ethereum", "Blockchain")
	hours := fs.Int("hours", 24, "Lookback hours")
	limit := fs.Int("limit", 20, "Max results")
	fs.Parse(args)

	mon := dexmon.New(mustClient())
	result, err := mon.DetectSandwiches(dexmon.SandwichConfig{
		Blockchain: *chain, LookbackH: *hours, Limit: *limit,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}

func runNewPools(args []string) {
	fs := flag.NewFlagSet("newpools", flag.ExitOnError)
	chain := fs.String("chain", "ethereum", "Blockchain")
	hours := fs.Int("hours", 24, "Lookback hours")
	limit := fs.Int("limit", 20, "Max results")
	fs.Parse(args)

	mon := dexmon.New(mustClient())
	result, err := mon.FindNewPools(dexmon.NewPoolsConfig{
		Blockchain: *chain, LookbackH: *hours, Limit: *limit,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}

func runTokenFlow(args []string) {
	fs := flag.NewFlagSet("flow", flag.ExitOnError)
	token := fs.String("token", "", "Token symbol (required)")
	hours := fs.Int("hours", 24, "Lookback hours")
	fs.Parse(args)

	if *token == "" {
		fatal(fmt.Errorf("-token is required"))
	}

	mon := dexmon.New(mustClient())
	result, err := mon.GetTokenFlow(dexmon.TokenFlowConfig{
		TokenSymbol: *token, LookbackH: *hours,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}
