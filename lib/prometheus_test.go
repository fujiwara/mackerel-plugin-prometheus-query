package promq

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	model "github.com/prometheus/common/model"
)

func TestMetric(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2019-12-10T11:22:33+09:00")
	m := &metric{
		key:       "foo.bar",
		value:     123.456,
		timestamp: ts,
	}
	if m.String() != "foo.bar\t123.456\t1575944553" {
		t.Error("unexpected metric string", m.String())
	}
}

func TestAuthorizationHeader(t *testing.T) {
	var gotAuth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		// Return a valid empty vector response
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`))
	}))
	defer ts.Close()

	p := Plugin{
		Address:             ts.URL,
		Format:              "test",
		Query:               "up",
		Timeout:             10 * time.Second,
		AuthorizationHeader: "Bearer my-secret-token",
	}
	_, err := p.fetch(context.Background())
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if gotAuth != "Bearer my-secret-token" {
		t.Errorf("expected Authorization header 'Bearer my-secret-token', got '%s'", gotAuth)
	}
}

func TestNoAuthorizationHeader(t *testing.T) {
	var gotAuth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`))
	}))
	defer ts.Close()

	p := Plugin{
		Address: ts.URL,
		Format:  "test",
		Query:   "up",
		Timeout: 10 * time.Second,
	}
	_, err := p.fetch(context.Background())
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if gotAuth != "" {
		t.Errorf("expected no Authorization header, got '%s'", gotAuth)
	}
}

func TestAuthorizationHeaderNotInErrorMessage(t *testing.T) {
	p := Plugin{
		Address:             "http://localhost:1", // connection refused
		Format:              "test",
		Query:               "up",
		Timeout:             1 * time.Second,
		AuthorizationHeader: "Bearer my-secret-token",
	}
	_, err := p.fetch(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if strings.Contains(err.Error(), "my-secret-token") {
		t.Error("error message should not contain the token")
	}
}

func TestFormatKey(t *testing.T) {
	mm := make(model.Metric)
	mm[model.LabelName("foo")] = model.LabelValue("FOO")
	mm[model.LabelName("bar")] = model.LabelValue("BAR")
	if s := formatKey(mm, "xxx.{foo}.{foo}.{bar}.{baz}.zzz"); s != "xxx.FOO.FOO.BAR._.zzz" {
		t.Error("unexpected formated string", s)
	}
}
