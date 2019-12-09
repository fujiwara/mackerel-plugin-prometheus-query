package promq

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	prometheus "github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	model "github.com/prometheus/common/model"
)

var formatRe = regexp.MustCompile(`\{.*?\}`)
var digitRe = regexp.MustCompile(`[^\w-]`)

// Plugin mackerel plugin for gunfish
type Plugin struct {
	Address string
	Format  string
	Query   string
}

type metric struct {
	key       string
	value     float64
	timestamp time.Time
}

func (m *metric) String() string {
	value := strconv.FormatFloat(m.value, 'f', -1, 64)
	ts := strconv.FormatInt(m.timestamp.Unix(), 10)
	return strings.Join([]string{m.key, value, ts}, "\t")
}

func (p Plugin) fetchMetrics() ([]*metric, error) {
	client, err := prometheus.NewClient(prometheus.Config{Address: p.Address})
	if err != nil {
		return nil, errors.Wrap(err, "failed to new client to prometheus API")
	}
	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := api.Query(ctx, p.Query, time.Now())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query to prometheous API %s", p.Address)
	}
	if len(warnings) > 0 {
		log.Printf("Warnings: %v\n", warnings)
	}

	var vector model.Vector
	switch result.Type() {
	case model.ValVector:
		vector = result.(model.Vector)
	default:
		return nil, errors.Errorf("unexpected value type %s", result.Type())
	}
	ret := make([]*metric, 0, len(vector))
	for _, s := range vector {
		ret = append(ret, &metric{
			key:       p.formatKey(s.Metric),
			value:     float64(s.Value),
			timestamp: s.Timestamp.Time(),
		})
	}
	return ret, nil
}

func (p *Plugin) formatKey(m model.Metric) string {
	return formatRe.ReplaceAllStringFunc(p.Format, func(match string) string {
		key := strings.Trim(match, "{}")
		if label := m[model.LabelName(key)]; label == "" {
			return "__unmatched__"
		} else {
			return digitRe.ReplaceAllString(string(label), "_")
		}
	})
}

// Do the plugin
func Do() error {
	optAddress := flag.String("address", "http://localhost:9090", "Prometheus address")
	optFormat := flag.String("metric-key-format", "", "Metric key format")
	optQuery := flag.String("query", "", "PromQL query")
	flag.Parse()

	promq := Plugin{
		Address: *optAddress,
		Format:  *optFormat,
		Query:   *optQuery,
	}

	metrics, err := promq.fetchMetrics()
	if err != nil {
		return err
	}
	for _, m := range metrics {
		fmt.Fprintln(os.Stdout, m.String())
	}
	return nil
}
