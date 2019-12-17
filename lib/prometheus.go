package promq

import (
	"context"
	"fmt"
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
var normalizeRe = regexp.MustCompile(`[^\w-]`)
var normalizedStr = "_"
var unmachedLabel = "_"

// Plugin mackerel plugin for prometheus query
type Plugin struct {
	Address string
	Format  string
	Query   string
	Timeout time.Duration
	EmitZero bool
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

func (p Plugin) fetch(ctx context.Context) ([]*metric, error) {
	client, err := prometheus.NewClient(prometheus.Config{Address: p.Address})
	if err != nil {
		return nil, errors.Wrap(err, "failed to new client for prometheus API")
	}
	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	result, warnings, err := api.Query(ctx, p.Query, time.Now())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query to prometheous API %s", p.Address)
	}
	if len(warnings) > 0 {
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", w)
		}
	}

	var vector model.Vector
	switch result.Type() {
	case model.ValVector:
		vector = result.(model.Vector)
	default:
		return nil, errors.Errorf("unexpected query response value type %s, vector required", result.Type())
	}
	metrics := make([]*metric, 0, len(vector))
	for _, s := range vector {
		metrics = append(metrics, &metric{
			key:       formatKey(s.Metric, p.Format),
			value:     float64(s.Value),
			timestamp: s.Timestamp.Time(),
		})
	}
	if len(metrics) == 0 && p.EmitZero {
		metrics = append(metrics, &metric{
			key: p.Format,
			value: 0,
			timestamp: time.Now(),
		})
	}
	return metrics, nil
}

func formatKey(m model.Metric, format string) string {
	return formatRe.ReplaceAllStringFunc(format, func(match string) string {
		key := strings.Trim(match, "{}")
		if label, exists := m[model.LabelName(key)]; exists {
			return normalizeRe.ReplaceAllString(string(label), normalizedStr)
		}
		return unmachedLabel
	})
}

// Run runs plugin
func (p *Plugin) Run(ctx context.Context) error {
	metrics, err := p.fetch(ctx)
	if err != nil {
		return err
	}
	for _, m := range metrics {
		fmt.Fprintln(os.Stdout, m.String())
	}
	return nil
}
