//go:build aipay || all

package main

import (
	"flag"

	"github.com/bergtatt/morpheco/scripts/pkg/aipay"
)

func init() {
	Register(&Command{Name: "micro", Desc: "AI micropayment pattern detection", Run: runMicro})
	Register(&Command{Name: "recurring", Desc: "Recurring payment patterns", Run: runRecurring})
	Register(&Command{Name: "channels", Desc: "Payment channel detection", Run: runChannels})
	Register(&Command{Name: "services", Desc: "Service payment analysis", Run: runServices})
}

func runMicro(args []string) {
	fs := flag.NewFlagSet("micro", flag.ExitOnError)
	hours := fs.Int("hours", 24, "Lookback hours")
	min := fs.Float64("min", 0.01, "Min per-tx USD")
	max := fs.Float64("max", 100, "Max per-tx USD")
	count := fs.Int("count", 5, "Min tx count")
	limit := fs.Int("limit", 50, "Max results")
	fs.Parse(args)

	mon := aipay.New(mustClient())
	result, err := mon.FindMicropaymentPatterns(aipay.MicropaymentConfig{
		LookbackH: *hours, MinAmountUSD: *min, MaxAmountUSD: *max,
		MinTxCount: *count, Limit: *limit,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(aipay.ParseMicropayments(result))
}

func runRecurring(args []string) {
	fs := flag.NewFlagSet("recurring", flag.ExitOnError)
	hours := fs.Int("hours", 48, "Lookback hours")
	count := fs.Int("count", 3, "Min payment count")
	limit := fs.Int("limit", 50, "Max results")
	fs.Parse(args)

	mon := aipay.New(mustClient())
	result, err := mon.FindRecurringPayments(aipay.RecurringConfig{
		LookbackH: *hours, MinTxCount: *count, Limit: *limit,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}

func runChannels(args []string) {
	fs := flag.NewFlagSet("channels", flag.ExitOnError)
	hours := fs.Int("hours", 24, "Lookback hours")
	count := fs.Int("count", 10, "Min tx count")
	limit := fs.Int("limit", 50, "Max results")
	fs.Parse(args)

	mon := aipay.New(mustClient())
	result, err := mon.FindPaymentChannels(aipay.ChannelConfig{
		LookbackH: *hours, MinTxCount: *count, Limit: *limit,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}

func runServices(args []string) {
	fs := flag.NewFlagSet("services", flag.ExitOnError)
	hours := fs.Int("hours", 48, "Lookback hours")
	count := fs.Int("count", 3, "Min payment count")
	limit := fs.Int("limit", 50, "Max results")
	fs.Parse(args)

	mon := aipay.New(mustClient())
	result, err := mon.AnalyzeServicePayments(aipay.ServicePaymentConfig{
		LookbackH: *hours, MinTxCount: *count, Limit: *limit,
	})
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}
