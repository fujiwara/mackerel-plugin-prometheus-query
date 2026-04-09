package promq

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	Address             string
	Format              string
	Query               string
	Timeout             time.Duration
	EmitZero            bool
	AuthorizationHeader string
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

type authorizationRoundTripper struct {
	header string
	rt     http.RoundTripper
}

func (a *authorizationRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	r2.Header.Set("Authorization", a.header)
	return a.rt.RoundTrip(r2)
}

func (p Plugin) fetch(ctx context.Context) ([]*metric, error) {
	cfg := prometheus.Config{Address: p.Address}
	if p.AuthorizationHeader != "" {
		cfg.RoundTripper = &authorizationRoundTripper{
			header: p.AuthorizationHeader,
			rt:     http.DefaultTransport,
		}
	}
	client, err := prometheus.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to new client for prometheus API: %w", err)
	}
	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	result, warnings, err := api.Query(ctx, p.Query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query to prometheus API %s: %w", p.Address, err)
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
		return nil, fmt.Errorf("unexpected query response value type %s, vector required", result.Type())
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
			key:       p.Format,
			value:     0,
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
