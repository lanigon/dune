//go:build dune || all

package main

import (
	"flag"
	"fmt"

	"github.com/bergtatt/morpheco/scripts/pkg/models"
)

func init() {
	Register(&Command{
		Name: "query",
		Desc: "Execute a saved Dune query",
		Run:  runDuneQuery,
	})
	Register(&Command{
		Name: "sql",
		Desc: "Execute raw DuneSQL",
		Run:  runDuneSQL,
	})
	Register(&Command{
		Name: "result",
		Desc: "Get latest cached query result",
		Run:  runDuneResult,
	})
	Register(&Command{
		Name: "status",
		Desc: "Check execution status",
		Run:  runDuneStatus,
	})
}

func runDuneQuery(args []string) {
	fs := flag.NewFlagSet("query", flag.ExitOnError)
	queryID := fs.Int("id", 0, "Query ID")
	wait := fs.Bool("wait", false, "Wait for results")
	limit := fs.Int("limit", 0, "Limit rows")
	fs.Parse(args)

	if *queryID == 0 {
		fatal(fmt.Errorf("-id is required"))
	}

	client := mustClient()
	if *wait {
		var opts []models.ResultOption
		if *limit > 0 {
			opts = append(opts, models.WithLimit(*limit))
		}
		result, err := client.RunQuery(*queryID, opts...)
		if err != nil {
			fatal(err)
		}
		outputJSON(result)
	} else {
		resp, err := client.ExecuteQuery(*queryID, nil, "medium")
		if err != nil {
			fatal(err)
		}
		outputJSON(resp)
	}
}

func runDuneSQL(args []string) {
	fs := flag.NewFlagSet("sql", flag.ExitOnError)
	query := fs.String("q", "", "SQL query")
	wait := fs.Bool("wait", false, "Wait for results")
	limit := fs.Int("limit", 0, "Limit rows")
	fs.Parse(args)

	if *query == "" {
		fatal(fmt.Errorf("-q is required"))
	}

	client := mustClient()
	if *wait {
		var opts []models.ResultOption
		if *limit > 0 {
			opts = append(opts, models.WithLimit(*limit))
		}
		result, err := client.RunSQL(*query, opts...)
		if err != nil {
			fatal(err)
		}
		outputJSON(result)
	} else {
		resp, err := client.ExecuteSQL(*query, "medium")
		if err != nil {
			fatal(err)
		}
		outputJSON(resp)
	}
}

func runDuneResult(args []string) {
	fs := flag.NewFlagSet("result", flag.ExitOnError)
	queryID := fs.Int("id", 0, "Query ID")
	limit := fs.Int("limit", 100, "Max rows")
	offset := fs.Int("offset", 0, "Row offset")
	filters := fs.String("filters", "", "Filter expression")
	columns := fs.String("columns", "", "Column names")
	sortBy := fs.String("sort", "", "ORDER BY")
	fs.Parse(args)

	if *queryID == 0 {
		fatal(fmt.Errorf("-id is required"))
	}

	var opts []models.ResultOption
	if *limit > 0 {
		opts = append(opts, models.WithLimit(*limit))
	}
	if *offset > 0 {
		opts = append(opts, models.WithOffset(*offset))
	}
	if *filters != "" {
		opts = append(opts, models.WithFilters(*filters))
	}
	if *columns != "" {
		opts = append(opts, models.WithColumns(*columns))
	}
	if *sortBy != "" {
		opts = append(opts, models.WithSortBy(*sortBy))
	}

	client := mustClient()
	result, err := client.GetLatestResult(*queryID, opts...)
	if err != nil {
		fatal(err)
	}
	outputJSON(result)
}

func runDuneStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	execID := fs.String("exec", "", "Execution ID")
	fs.Parse(args)

	if *execID == "" {
		fatal(fmt.Errorf("-exec is required"))
	}

	client := mustClient()
	status, err := client.GetExecutionStatus(*execID)
	if err != nil {
		fatal(err)
	}
	outputJSON(status)
}

