package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	promq "github.com/fujiwara/mackerel-plugin-prometheus-query/lib"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	optAddress := flag.String("address", "http://localhost:9090", "Prometheus address")
	optFormat := flag.String("metric-key-format", "", "Metric key format")
	optQuery := flag.String("query", "", "PromQL query")
	optTimeout := flag.String("timeout", "10s", "timeout for query")
	optEmitZero := flag.Bool("emit-zero", false, "emit 0 when query returns no result")
	optAuthHeader := flag.String("authorization-header", "", "Authorization header value (e.g. \"Bearer TOKEN\")")
	flag.Parse()

	to, err := time.ParseDuration(*optTimeout)
	if err != nil {
		return fmt.Errorf("failed to parse timeout: %w", err)
	}

	p := promq.Plugin{
		Address:             *optAddress,
		Format:              *optFormat,
		Query:               *optQuery,
		Timeout:             to,
		EmitZero:            *optEmitZero,
		AuthorizationHeader: *optAuthHeader,
	}

	return p.Run(context.Background())
}
