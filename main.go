package main

import (
	"flag"
	"context"
	"fmt"
	"os"
	"time"

	promq "github.com/fujiwara/mackerel-plugin-prometheus-query/lib"
	"github.com/pkg/errors"
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
	flag.Parse()

	to, err := time.ParseDuration(*optTimeout)
	if err != nil {
		return errors.Wrap(err, "failed to parse timeout")
	}

	p := promq.Plugin{
		Address: *optAddress,
		Format:  *optFormat,
		Query:   *optQuery,
		Timeout: to,
	}

	return p.Run(context.Background())
}
